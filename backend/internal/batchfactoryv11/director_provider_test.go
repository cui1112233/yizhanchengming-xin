package batchfactoryv11

import (
	"context"
	"encoding/json"
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
