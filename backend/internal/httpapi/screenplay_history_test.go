package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"reflect"
	"testing"
)

func TestScreenplayTextHistoryReachesConfiguredChannel(t *testing.T) {
	messages := []screenplay.TextMessage{{Role: "system", Content: "system"}, {Role: "user", Content: "first"}, {Role: "assistant", Content: "partial"}, {Role: "user", Content: "continue"}}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body struct {
			Model    string                   `json:"model"`
			Messages []screenplay.TextMessage `json:"messages"`
		}
		if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
			t.Error(e)
		}
		if body.Model != "own-configured-model" || !reflect.DeepEqual(body.Messages, messages) {
			t.Error("configured channel changed history", body)
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"next"}}]}`))
	}))
	defer server.Close()
	g := screenplayTextGenerator{owner: "own-user", resolver: func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "own-user" || model != "own-config-entry" {
			t.Error(owner, model)
		}
		return &integrationTextProvider{Endpoint: server.URL, APIKey: "OWN_FAKE", Model: "own-configured-model", Client: server.Client()}, nil
	}}
	out, e := g.Generate(context.Background(), screenplay.TextRequest{Model: "own-config-entry", Messages: messages, MaxOutputTokens: 16000})
	if e != nil || out.Content != "next" || calls != 1 {
		t.Fatal(out, e, calls)
	}
}
