package authbridge

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolverGetsTheAuthenticatedOwnerFrom4001(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer account-token" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"owner-a","active":true}`))
	}))
	defer upstream.Close()

	resolver, err := NewResolver(upstream.URL, upstream.Client())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/magic-cut/projects", nil)
	request.AddCookie(&http.Cookie{Name: accountCookieName, Value: "account-token"})

	owner, err := resolver.Resolve(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if owner != "owner-a" {
		t.Fatalf("owner = %q", owner)
	}
}

func TestResolverRejectsRequestsWithoutTheBridgeCookie(t *testing.T) {
	resolver, err := NewResolver("http://127.0.0.1:4001", http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}

	_, err = resolver.Resolve(context.Background(), httptest.NewRequest(http.MethodGet, "/", nil))
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("err = %v", err)
	}
}
