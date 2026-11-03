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
		return "", err
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

func openAICompatibleHTTPClient(endpoint string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	host := ""
	if err == nil {
		host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), "."))
	}
	if host == "api.vveai.com" || host == "ggwk1.online" || host == "www.ggwk1.online" {
		transport.ForceAttemptHTTP2 = true
		transport.TLSNextProto = nil
	} else {
		transport.ForceAttemptHTTP2 = false
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	return &http.Client{Timeout: directorProviderRequestTimeout, Transport: transport}
}
