package batchfactoryv11

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DirectorProvider interface {
	Complete(context.Context, TextCompletionRequest) (string, error)
}

type OpenAICompatibleProvider struct {
	Endpoint string
	APIKey   string
	Model    string
	Client   *http.Client
	// retryClient 用于协议不匹配时的一次性 HTTP/2 重试；生产为 nil，按 Endpoint 自动构造。
	retryClient *http.Client
}

// Director planning is intentionally granted more time than a short asset or
// video-provider request. The provider returns one non-streaming JSON result,
// so it may not send headers until the complete H3 director plan is ready.
const directorProviderRequestTimeout = 300 * time.Second

func (p *OpenAICompatibleProvider) Validate() error {
	u, err := url.Parse(strings.TrimSpace(p.Endpoint))
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return fmt.Errorf("director provider endpoint must be an absolute https URL")
	}
	if strings.TrimSpace(p.APIKey) == "" || strings.TrimSpace(p.Model) == "" {
		return fmt.Errorf("director provider API key and model are required")
	}
	return nil
}

func (p *OpenAICompatibleProvider) Complete(ctx context.Context, input TextCompletionRequest) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	payload := map[string]any{
		"model": p.Model,
		"messages": []map[string]string{
			{"role": "system", "content": input.SystemPrompt},
			{"role": "user", "content": input.UserPrompt},
		},
		"temperature": input.Temperature,
		"max_tokens":  input.MaxTokens,
	}
	// SD director cards and opening variants have a deliberately plain-text
	// wire protocol. Sending OpenAI's JSON-only response_format for those calls
	// contradicts their required ===VIDEO/时长 lines and causes false failures.
	if !input.DisableJSONResponse {
		payload["response_format"] = map[string]string{"type": "json_object"}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	client := p.Client
	if client == nil {
		client = openAICompatibleHTTPClient(p.Endpoint)
	}
	resp, err := client.Do(req)
	if err != nil {
		// 服务端按 HTTP/2 回帧、客户端按 HTTP/1.1 解析时会产生 malformed HTTP
		// response。这里只切换协议重发一次；重试仍失败则保留原始错误。
		if !directorProtocolMismatchError(err) {
			return "", err
		}
		retryClient := p.retryClient
		if retryClient == nil {
			retryClient = openAICompatibleHTTP2Client(p.Endpoint)
		}
		retryReq, retryReqErr := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
		if retryReqErr != nil {
			return "", err
		}
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("Authorization", "Bearer "+p.APIKey)
		retryResp, retryErr := retryClient.Do(retryReq)
		if retryErr != nil {
			return "", err
		}
		resp = retryResp
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("director provider returned HTTP %d", resp.StatusCode)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text         string `json:"text"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil || len(decoded.Choices) == 0 {
		return "", fmt.Errorf("director provider returned an invalid response")
	}
	choice := decoded.Choices[0]
	// A partial JSON response must never reach the H3 parser as a generic
	// "invalid input".  OpenAI-compatible services expose this condition as
	// finish_reason=length; make it retryable and tell the caller to use a
	// model/output budget that can complete the frozen director contract.
	if strings.EqualFold(strings.TrimSpace(choice.FinishReason), "length") {
		return "", fmt.Errorf("director provider output truncated (finish_reason=length)")
	}
	content := strings.TrimSpace(choice.Message.Content)
	if content == "" {
		content = strings.TrimSpace(choice.Text)
	}
	if content == "" {
		return "", fmt.Errorf("director provider returned empty content")
	}
	return content, nil
}

// openAICompatibleLegacyHTTP1Hosts 是已知会重置长连接的旧网关反向例外表。
// 默认所有 https 端点都走 HTTP/2；只有列在这里的 host 继续强制 HTTP/1.1。
var openAICompatibleLegacyHTTP1Hosts = map[string]struct{}{}

func openAICompatibleHTTPClient(endpoint string) *http.Client {
	return openAICompatibleHTTPClientForHosts(endpoint, openAICompatibleLegacyHTTP1Hosts)
}

func openAICompatibleHTTPClientForHosts(endpoint string, legacyHTTP1 map[string]struct{}) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	host := ""
	if err == nil {
		host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), "."))
	}
	if openAICompatibleHostPrefersHTTP1(host, legacyHTTP1) {
		transport.ForceAttemptHTTP2 = false
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	} else {
		transport.ForceAttemptHTTP2 = true
		transport.TLSNextProto = nil
	}
	return &http.Client{Timeout: directorProviderRequestTimeout, Transport: transport}
}

// openAICompatibleHostPrefersHTTP1 只对反向例外表内的旧网关返回 true。
// Tencent Zeabur preview 域名可能只提供 HTTP/2，明确保留在允许 h2 的范围。
func openAICompatibleHostPrefersHTTP1(host string, legacyHTTP1 map[string]struct{}) bool {
	if host == "" {
		return false
	}
	if host == "nvewvip.preview.tencent-zeabur.cn" || strings.HasSuffix(host, ".preview.tencent-zeabur.cn") {
		return false
	}
	_, ok := legacyHTTP1[host]
	return ok
}

// openAICompatibleHTTP2Client 强制 HTTP/2，用于协议不匹配时的一次性重试。
func openAICompatibleHTTP2Client(endpoint string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = true
	transport.TLSNextProto = nil
	return &http.Client{Timeout: directorProviderRequestTimeout, Transport: transport}
}

// directorProtocolMismatchError 识别“服务端按 HTTP/2 回帧、客户端却按 H1 解析”
// 的错误。这类错误只允许一次协议切换重试。
func directorProtocolMismatchError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "malformed HTTP response") || strings.Contains(message, "\x00\x00\x1e\x04")
}
