package generation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPProvider struct {
	BaseURL string
	APIKey string
	Model string
	Client *http.Client
}

func NewHTTPProvider(baseURL, apiKey, model string) *HTTPProvider {
	return &HTTPProvider{BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"), APIKey: strings.TrimSpace(apiKey), Model: strings.TrimSpace(model), Client: &http.Client{Timeout: 90 * time.Second}}
}

func (p *HTTPProvider) Complete(ctx context.Context, req TextRequest) (string, error) {
	if p == nil || p.BaseURL == "" || p.APIKey == "" || p.Model == "" { return "", ErrUnavailable }
	client := p.Client; if client == nil { client = &http.Client{Timeout: 90 * time.Second} }
	payload := map[string]any{"model": p.Model, "messages": []map[string]string{{"role":"system","content":req.SystemPrompt},{"role":"user","content":req.UserPrompt}}, "temperature": 0.2}
	encoded, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(encoded)); if err != nil { return "", err }
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := client.Do(httpReq); if err != nil { return "", fmt.Errorf("generation provider request failed: %w", err) }
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20)); if err != nil { return "", err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return "", fmt.Errorf("generation provider returned status %d", resp.StatusCode) }
	var decoded struct { Choices []struct { Message struct { Content string `json:"content"` } `json:"message"` } `json:"choices"` }
	if err := json.Unmarshal(body, &decoded); err != nil { return "", fmt.Errorf("generation provider response invalid: %w", err) }
	if len(decoded.Choices)==0 || strings.TrimSpace(decoded.Choices[0].Message.Content)=="" { return "", fmt.Errorf("generation provider returned empty content") }
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}

type UnavailableProvider struct{}
func (UnavailableProvider) Complete(context.Context, TextRequest) (string,error){ return "", ErrUnavailable }
