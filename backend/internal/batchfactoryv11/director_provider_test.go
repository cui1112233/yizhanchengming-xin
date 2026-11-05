package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleHTTPClientUsesHTTP2ForGGWKEndpoint(t *testing.T) {
	client := openAICompatibleHTTPClient("https://www.ggwk1.online/v1/chat/completions")
	transport, ok := client.Transport.(*http.Transport)
	if !ok || !transport.ForceAttemptHTTP2 || transport.TLSNextProto != nil {
		t.Fatalf("transport=%#v, want HTTP/2-capable transport", client.Transport)
	}
}

func TestOpenAICompatibleHTTPClientDefaultsToHTTP2(t *testing.T) {
	for _, endpoint := range []string{
		"https://nvewvip.preview.tencent-zeabur.cn/v1/chat/completions",
		"https://abc.preview.tencent-zeabur.cn/v1/chat/completions",
		"https://text.example.test/v1/chat/completions",
	} {
		client := openAICompatibleHTTPClient(endpoint)
		transport, ok := client.Transport.(*http.Transport)
		if !ok || !transport.ForceAttemptHTTP2 || transport.TLSNextProto != nil {
			t.Fatalf("endpoint %s transport=%#v, want HTTP/2-capable transport", endpoint, client.Transport)
		}
	}
}

func TestOpenAICompatibleHTTPClientKeepsLegacyGatewayOnHTTP1(t *testing.T) {
	legacy := map[string]struct{}{"legacy.example.test": {}}
	client := openAICompatibleHTTPClientForHosts("https://legacy.example.test/v1/chat/completions", legacy)
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.ForceAttemptHTTP2 || transport.TLSNextProto == nil {
		t.Fatalf("transport=%#v, want HTTP/1.1-only transport for the legacy exception host", client.Transport)
	}
}

type directorStubTransport struct {
	calls int
	fail  bool
	body  string
}

func (s *directorStubTransport) RoundTrip(*http.Request) (*http.Response, error) {
	s.calls++
	if s.fail {
		return nil, fmt.Errorf("net/http: HTTP/1.x transport connection broken: malformed HTTP response \"\\x00\\x00\\x1e\\x04\"")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

func TestOpenAICompatibleProviderRetriesOnceWithHTTP2OnProtocolMismatch(t *testing.T) {
	first := &directorStubTransport{fail: true}
	second := &directorStubTransport{body: `{"choices":[{"message":{"content":"{}"}}]}`}
	provider := OpenAICompatibleProvider{
		Endpoint:    "https://nvewvip.preview.tencent-zeabur.cn/v1/chat/completions",
		APIKey:      "test-key",
		Model:       "test-model",
		Client:      &http.Client{Transport: first},
		retryClient: &http.Client{Transport: second},
	}
	content, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if content != "{}" {
		t.Fatalf("content=%q, want {}", content)
	}
	if first.calls != 1 || second.calls != 1 {
		t.Fatalf("calls first=%d second=%d, want one attempt plus exactly one HTTP/2 retry", first.calls, second.calls)
	}
}

func TestOpenAICompatibleProviderReportsOriginalErrorWhenHTTP2RetryFails(t *testing.T) {
	first := &directorStubTransport{fail: true}
	second := &directorStubTransport{fail: true}
	provider := OpenAICompatibleProvider{
		Endpoint:    "https://nvewvip.preview.tencent-zeabur.cn/v1/chat/completions",
		APIKey:      "test-key",
		Model:       "test-model",
		Client:      &http.Client{Transport: first},
		retryClient: &http.Client{Transport: second},
	}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "malformed HTTP response") {
		t.Fatalf("error=%v, want the original malformed-response error", err)
	}
	if first.calls != 1 || second.calls != 1 {
		t.Fatalf("calls first=%d second=%d, want exactly one retry", first.calls, second.calls)
	}
}

func TestOpenAICompatibleProviderRequestsJSONObjectResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		format, ok := request["response_format"].(map[string]any)
		if !ok || format["type"] != "json_object" {
			t.Fatalf("response_format=%#v, want json_object", request["response_format"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		Endpoint: server.URL,
		APIKey:   "test-key",
		Model:    "test-model",
		Client:   server.Client(),
	}
	if _, err := provider.Complete(context.Background(), TextCompletionRequest{}); err != nil {
		t.Fatalf("complete: %v", err)
	}
}

func TestOpenAICompatibleProviderDoesNotForceJSONObjectForPlainTextProtocol(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if _, exists := request["response_format"]; exists {
			t.Fatalf("plain-text request must not force JSON response: %#v", request["response_format"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"===VIDEO 01===\n时长：10秒\n资产引用：人物=无；场景=无；道具=无\n正文"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", Client: server.Client()}
	if _, err := provider.Complete(context.Background(), TextCompletionRequest{DisableJSONResponse: true}); err != nil {
		t.Fatalf("complete: %v", err)
	}
}

func TestOpenAICompatibleProviderReportsTruncatedStructuredOutput(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"length","message":{"content":"{\"director_cards\":["}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		Endpoint: server.URL,
		APIKey:   "test-key",
		Model:    "test-model",
		Client:   server.Client(),
	}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "finish_reason=length") {
		t.Fatalf("error=%v, want explicit truncation diagnostic", err)
	}
}
