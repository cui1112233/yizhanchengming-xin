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

// Domain tests establish how the conversation is built. This checks the actual
// configured-channel HTTP transport, without contacting a network model.
func TestNovelShortWriterConversationReachesConfiguredChannel(t *testing.T) {
	for _, repair := range []bool{false, true} {
		t.Run(map[bool]string{false: "revision", true: "format-repair"}[repair], func(t *testing.T) {
			messages := []batchfactoryv11.TextCompletionMessage{{Role: "system", Content: "frozen writer + addendum"}, {Role: "user", Content: "original draft request"}, {Role: "assistant", Content: `{"content":"original draft"}`}, {Role: "user", Content: "frozen followup + review"}}
			if repair {
				messages = append(messages, batchfactoryv11.TextCompletionMessage{Role: "assistant", Content: "invalid reply"}, batchfactoryv11.TextCompletionMessage{Role: "user", Content: "format correction"})
			}
			var received struct {
				Model    string                                  `json:"model"`
				Messages []batchfactoryv11.TextCompletionMessage `json:"messages"`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer fixture-only" {
					t.Error("configured transport contract changed")
				}
				if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"fixture result"}}]}`))
			}))
			defer server.Close()
			provider := &integrationTextProvider{Endpoint: server.URL, Client: server.Client(), APIKey: "fixture-only", Model: "configured-channel-model"}
			result, err := provider.Complete(context.Background(), batchfactoryv11.TextCompletionRequest{Messages: messages, SystemPrompt: messages[0].Content, UserPrompt: messages[len(messages)-1].Content, MaxTokens: 4000})
			if err != nil || result != "fixture result" {
				t.Fatal(result, err)
			}
			if received.Model != "configured-channel-model" || !reflect.DeepEqual(received.Messages, messages) {
				t.Fatalf("wire lost conversation: %#v", received)
			}
		})
	}
}
