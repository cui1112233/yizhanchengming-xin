package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

// Use the real configured-channel transport and an isolated local server. The
// domain tests separately verify the native builders, routes and frozen text.
func TestNovelShortFullMarkdownReachesConfiguredChannel(t *testing.T) {
	const markdown = "=== SHORT_FICTION_TITLE ===\n渡口\n=== CHAPTER 1 CONTENT ===\n正文。"
	for _, revision := range []bool{false, true} {
		t.Run(map[bool]string{false: "first-draft", true: "four-message-revision"}[revision], func(t *testing.T) {
			messages := []batchfactoryv11.TextCompletionMessage{{Role: "system", Content: "frozen full writer"}, {Role: "user", Content: "full draft request"}}
			if revision {
				messages = append(messages, batchfactoryv11.TextCompletionMessage{Role: "assistant", Content: markdown}, batchfactoryv11.TextCompletionMessage{Role: "user", Content: "frozen revision followup + natural review"})
			}
			var received struct {
				Model          string                                  `json:"model"`
				Messages       []batchfactoryv11.TextCompletionMessage `json:"messages"`
				ResponseFormat json.RawMessage                         `json:"response_format"`
				Temperature    float64                                 `json:"temperature"`
				MaxTokens      int                                     `json:"max_tokens"`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": markdown}}}})
			}))
			defer server.Close()
			provider := &integrationTextProvider{Endpoint: server.URL, Client: server.Client(), APIKey: "local-fixture-only", Model: "our-configured-model"}
			request := batchfactoryv11.TextCompletionRequest{Messages: messages, SystemPrompt: messages[0].Content, UserPrompt: messages[len(messages)-1].Content, DisableJSONResponse: true, Temperature: .45, MaxTokens: 12288}
			result, err := provider.Complete(context.Background(), request)
			if err != nil || result != markdown {
				t.Fatal(result, err)
			}
			if received.Model != "our-configured-model" || !reflect.DeepEqual(received.Messages, messages) || len(received.ResponseFormat) != 0 || received.Temperature != .45 || received.MaxTokens != 12288 {
				t.Fatalf("full Markdown wire contract changed: %#v", received)
			}
		})
	}
}
