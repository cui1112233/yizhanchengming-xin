package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLegacySessionResolverAcceptsActiveNodeSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/login/session" || req.Header.Get("Authorization") != "Bearer node-token" {
			t.Fatalf("unexpected legacy session request: %s %s", req.URL.Path, req.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"alice","active":true,"isOwner":true}`))
	}))
	defer server.Close()
	identity, err := NewLegacySessionResolver(server.URL)(context.Background(), "node-token")
	if err != nil {
		t.Fatal(err)
	}
	if identity.Username != "alice" || !identity.IsOwner {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}
