package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/batchfactoryv11"
	"sync/atomic"
	"testing"
)

func TestNovelIntegrationProviderSinglePOSTOnRedirect(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/" {
			w.Header().Set("Location", "/target")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
		fmt.Fprint(w, `{"choices":[{"message":{"content":"{}"}}]}`)
	}))
	defer server.Close()
	client := server.Client()
	p := &integrationTextProvider{Endpoint: server.URL, Model: "local-test", APIKey: "local-test", Client: client}
	if _, err := p.Complete(batchfactoryv11.WithSingleAttempt(context.Background()), batchfactoryv11.TextCompletionRequest{}); err == nil {
		t.Fatal("redirect must stop")
	}
	if calls.Load() != 1 {
		t.Fatalf("novel sent %d requests", calls.Load())
	}
	calls.Store(0)
	if _, err := p.Complete(context.Background(), batchfactoryv11.TextCompletionRequest{}); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || client.CheckRedirect != nil {
		t.Fatal("other modules redirect behavior changed")
	}
}
