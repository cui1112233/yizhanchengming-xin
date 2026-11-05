package batchfactoryv11

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DirectorHTTPError keeps the upstream response body so a 400/403 is never
// reduced to a bare status code again. The body is what tells operators
// whether the model rejected the request shape, the model name, or the API key.
type DirectorHTTPError struct {
	StatusCode int
	Model      string
	Body       string
}

func (e *DirectorHTTPError) Error() string {
	body := directorHTTPBodySnippet([]byte(e.Body), 400)
	model := strings.TrimSpace(e.Model)
	if body == "" {
		body = http.StatusText(e.StatusCode)
	}
	if model != "" {
		return fmt.Sprintf("文本模型请求失败（%s，HTTP %d，model=%s）：%s", directorHTTPFailureKind(e.StatusCode, e.Body), e.StatusCode, model, body)
	}
	return fmt.Sprintf("文本模型请求失败（%s，HTTP %d）：%s", directorHTTPFailureKind(e.StatusCode, e.Body), e.StatusCode, body)
}

// directorHTTPFailureKind turns the upstream body into a human category so a
// 400 is not reported as a generic outage. The raw body is still appended by
// Error(), which keeps the exact model-side wording available for diagnosis.
func directorHTTPFailureKind(status int, body string) string {
	lower := strings.ToLower(body)
	switch {
	case status == http.StatusTooManyRequests:
		return "上游限流"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "鉴权失败"
	case status == http.StatusRequestEntityTooLarge ||
		strings.Contains(lower, "context length") || strings.Contains(lower, "context_length") ||
		strings.Contains(lower, "maximum context") || strings.Contains(lower, "max_tokens") ||
		strings.Contains(lower, "token limit") || strings.Contains(lower, "too long") ||
		strings.Contains(lower, "超长") || strings.Contains(lower, "过长"):
		return "请求超长"
	case strings.Contains(lower, "content_policy") || strings.Contains(lower, "content policy") ||
		strings.Contains(lower, "content filter") || strings.Contains(lower, "moderation") ||
		strings.Contains(lower, "safety") || strings.Contains(lower, "审核") || strings.Contains(lower, "违规"):
		return "内容审核拒绝"
	case status == http.StatusNotFound || strings.Contains(lower, "model not found") ||
		strings.Contains(lower, "unknown model") || strings.Contains(lower, "invalid model"):
		return "模型不存在或不支持"
	case status >= 500:
		return "上游通道错误"
	default:
		return "参数或请求格式不合法"
	}
}

// Unwrap separates retryable channel failures from user-fixable model/key
// configuration problems. 400/401/403/404/422 must not be advertised as a
// transient "capability unavailable" outage.
func (e *DirectorHTTPError) Unwrap() error {
	if directorProviderRetryableStatus(e.StatusCode) {
		return ErrUpstream
	}
	return ErrConfigRequired
}

func directorProviderRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return true
	}
	return status >= 500
}

func directorHTTPBodySnippet(raw []byte, limit int) string {
	message := strings.Join(strings.Fields(string(raw)), " ")
	if message == "" {
		return ""
	}
	runes := []rune(message)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return message
}

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
	// A 429/5xx is a channel hiccup, not a broken book. Retry a bounded number
	// of times; a 400/401/403/404/422 is a configuration/model problem and is
	// returned immediately with the upstream body so it can be diagnosed.
	const maxAttempts = 3
	var raw []byte
	var statusCode int
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, reqErr := p.do(ctx, body)
		if reqErr != nil {
			if !directorProviderRetryableTransportError(reqErr) || attempt == maxAttempts {
				return "", reqErr
			}
			if waitErr := directorProviderWaitRetry(ctx, attempt); waitErr != nil {
				return "", reqErr
			}
			continue
		}
		raw, err = io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		statusCode = resp.StatusCode
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || attempt == maxAttempts {
				return "", err
			}
			if waitErr := directorProviderWaitRetry(ctx, attempt); waitErr != nil {
				return "", err
			}
			continue
		}
		if statusCode >= 200 && statusCode < 300 {
			break
		}
		httpErr := &DirectorHTTPError{StatusCode: statusCode, Model: strings.TrimSpace(p.Model), Body: string(raw)}
		log.Printf("director provider error model=%s status=%d endpoint=%s body=%s", p.Model, statusCode, directorProviderEndpointHost(p.Endpoint), directorHTTPBodySnippet(raw, 400))
		if !directorProviderRetryableStatus(statusCode) || attempt == maxAttempts {
			return "", httpErr
		}
		if waitErr := directorProviderWaitRetry(ctx, attempt); waitErr != nil {
			return "", httpErr
		}
	}
	if statusCode < 200 || statusCode >= 300 {
		// Defensive: the loop only exits with a 2xx status or an explicit return.
		return "", &DirectorHTTPError{StatusCode: statusCode, Model: strings.TrimSpace(p.Model), Body: string(raw)}
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

// do sends one logical request. It keeps the HTTP/2 protocol-mismatch retry
// exactly once, separate from the bounded 429/5xx attempt loop.
func (p *OpenAICompatibleProvider) do(ctx context.Context, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	client := p.Client
	if client == nil {
		client = openAICompatibleHTTPClient(p.Endpoint)
	}
	resp, err := client.Do(req)
	if err == nil {
		return resp, nil
	}
	// 服务端按 HTTP/2 回帧、客户端按 HTTP/1.1 解析时会产生 malformed HTTP
	// response。这里只切换协议重发一次；重试仍失败则保留原始错误。
	if !directorProtocolMismatchError(err) {
		return nil, err
	}
	retryClient := p.retryClient
	if retryClient == nil {
		retryClient = openAICompatibleHTTP2Client(p.Endpoint)
	}
	retryReq, retryReqErr := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if retryReqErr != nil {
		return nil, err
	}
	retryReq.Header.Set("Content-Type", "application/json")
	retryReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	retryResp, retryErr := retryClient.Do(retryReq)
	if retryErr != nil {
		return nil, err
	}
	return retryResp, nil
}

func directorProviderRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// The protocol switch already retried once; do not multiply it.
	return !directorProtocolMismatchError(err)
}

func directorProviderWaitRetry(ctx context.Context, attempt int) error {
	delay := time.Duration(attempt) * 200 * time.Millisecond
	if attempt > 1 {
		delay = 500 * time.Millisecond
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func directorProviderEndpointHost(endpoint string) string {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return ""
	}
	return parsed.Host
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
