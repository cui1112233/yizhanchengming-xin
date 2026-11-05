package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func TestOpenAICompatibleProviderRetainsUpstreamBodyAndClassifies400(t *testing.T) {
	var calls int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"model not found: [anti-2]gemini-3.8-flash","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "[anti-2]gemini-3.8-flash", Client: server.Client()}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil {
		t.Fatal("a 400 must fail the request")
	}
	if !strings.Contains(err.Error(), "model not found") || !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("error must retain the upstream body, got %v", err)
	}
	if !errors.Is(err, ErrConfigRequired) {
		t.Fatalf("400 is a model/config error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("400 calls=%d, want exactly 1 (no retry)", got)
	}
}

func TestOpenAICompatibleProviderRetriesRateLimitOnce(t *testing.T) {
	var calls int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if attempt == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", Client: server.Client()}
	content, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err != nil {
		t.Fatalf("a 429 must be retried, got %v", err)
	}
	if content != "{}" {
		t.Fatalf("content=%q, want {}", content)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("429 calls=%d, want one retry", got)
	}
}

func TestOpenAICompatibleProviderStopsAfterBoundedServerErrorRetries(t *testing.T) {
	var calls int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"upstream exploded"}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", Client: server.Client()}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "upstream exploded") {
		t.Fatalf("error must retain the upstream body, got %v", err)
	}
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("5xx must remain a retryable upstream error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("5xx calls=%d, want exactly 3 attempts", got)
	}
}
