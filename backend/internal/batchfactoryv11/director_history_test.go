package batchfactoryv11

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestTextCompletionHistorySurvivesProviderAndParameterFallback(t *testing.T) {
	messages := []TextCompletionMessage{{"system", "system"}, {"user", "first"}, {"assistant", "saved partial"}, {"user", "continue"}}
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body struct {
			Messages []TextCompletionMessage `json:"messages"`
		}
		if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
			t.Error(e)
		}
		if !reflect.DeepEqual(body.Messages, messages) {
			t.Error("history lost", body.Messages)
		}
		if calls == 1 {
			w.WriteHeader(400)
			w.Write([]byte(`{"error":{"message":"请求参数有误，请检查后重试"}}`))
			return
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"next"}}]}`))
	}))
	defer server.Close()
	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "OWN_FAKE", Model: "own-model", Client: server.Client()}
	result, e := provider.Complete(context.Background(), TextCompletionRequest{SystemPrompt: "unused", UserPrompt: "unused", Messages: messages, MaxTokens: 24000})
	if e != nil || result != "next" || calls != 2 {
		t.Fatal(result, e, calls)
	}
}
