package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIResponderSendsRecentConversationAndMediaContext(t *testing.T) {
	var received chatCompletionRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" { t.Fatalf("authorization=%q", got) }
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil { t.Fatal(err) }
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": `{"content":"第二张已经作为当前参考图。"}`}}},
		})
	}))
	defer server.Close()

	responder, err := NewAIResponder(AIResponderConfig{BaseURL: server.URL, APIKey: "secret", Model: "model-x", HTTPClient: server.Client()})
	if err != nil { t.Fatal(err) }
	result, err := responder.Respond(context.Background(), ResponseContext{
		Owner: "owner",
		ThreadID: "thread",
		Input: "用第二张继续",
		MediaAssetIDs: []string{"asset_2"},
		Messages: []Message{
			{Role: RoleUser, Content: "先生成两张"},
			{Role: RoleAssistant, Content: "已经准备好两张结果"},
		},
	})
	if err != nil { t.Fatal(err) }
	if result.Content != "第二张已经作为当前参考图。" { t.Fatalf("content=%q", result.Content) }
	if received.Model != "model-x" { t.Fatalf("model=%q", received.Model) }
	joined := ""
	for _, message := range received.Messages { joined += message.Content + "\n" }
	for _, want := range []string{"先生成两张", "已经准备好两张结果", "用第二张继续", "asset_2"} {
		if !strings.Contains(joined, want) { t.Fatalf("request missing %q: %s", want, joined) }
	}
}

func TestAIResponderAcceptsOnlyAllowlistedNavigation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []any{map[string]any{"message": map[string]any{"content": `{"content":"我来打开。","navigate":{"path":"https://evil.example"}}`}}},
		})
	}))
	defer server.Close()
	responder, err := NewAIResponder(AIResponderConfig{BaseURL: server.URL, APIKey: "secret", Model: "model-x", HTTPClient: server.Client()})
	if err != nil { t.Fatal(err) }
	result, err := responder.Respond(context.Background(), ResponseContext{Input: "打开那个网站"})
	if err != nil { t.Fatal(err) }
	if result.Tool != nil { t.Fatalf("unexpected tool=%#v", result.Tool) }
}

func TestHybridResponderKeepsDeterministicNavigation(t *testing.T) {
	hybrid := NewHybridResponder(NewResponder(), nil)
	result, err := hybrid.Respond(context.Background(), ResponseContext{Input: "帮我打开小说获取"})
	if err != nil { t.Fatal(err) }
	if result.Tool == nil || result.Tool.Name != ToolNavigate { t.Fatalf("tool=%#v", result.Tool) }
}
