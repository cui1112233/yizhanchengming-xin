package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"qiantie/backend/internal/auth"
)

func TestNovelFetchAccountScopeEnforcesTeamVisibility(t *testing.T) {
	accounts := &applicationTestStore{accounts: map[string]auth.Account{
		"developer": {Username: "developer", IsOwner: true, Active: true},
		"manager-a": {Username: "manager-a", Active: true},
		"manager-b": {Username: "manager-b", Active: true},
		"member-a":  {Username: "member-a", Active: true},
		"member-b":  {Username: "member-b", Active: true},
	}}
	profiles := NewMemberProfileStore(t.TempDir())
	if err := profiles.writeArray("member-profiles.json", []map[string]any{
		{"username": "manager-a", "role": "manager"},
		{"username": "manager-b", "role": "manager"},
		{"username": "member-a", "role": "member", "boundTo": "manager-a"},
		{"username": "member-b", "role": "member", "boundTo": "manager-b"},
	}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		actor  string
		target string
		want   string
		ok     bool
	}{
		{name: "manager sees own member", actor: "manager-a", target: "member-a", want: "member-a", ok: true},
		{name: "manager cannot see another team", actor: "manager-a", target: "member-b", ok: false},
		{name: "developer sees every account", actor: "developer", target: "member-b", want: "member-b", ok: true},
		{name: "member cannot switch account", actor: "member-a", target: "manager-a", ok: false},
		{name: "self remains available", actor: "member-a", target: "member-a", want: "member-a", ok: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveNovelFetchAccountScope(context.Background(), BridgeIdentity{Username: test.actor}, test.target, accounts, profiles)
			if test.ok && (err != nil || got.Username != test.want) {
				t.Fatalf("resolve scope = %#v, %v; want %q", got, err, test.want)
			}
			if !test.ok && err == nil {
				t.Fatalf("resolve scope unexpectedly allowed %#v", got)
			}
		})
	}
}

func TestNovelFetchAccountScopeMiddlewareUsesRequestedOwner(t *testing.T) {
	accounts := &applicationTestStore{accounts: map[string]auth.Account{
		"manager": {Username: "manager", Active: true},
		"member":  {Username: "member", Active: true},
	}}
	profiles := NewMemberProfileStore(t.TempDir())
	if err := profiles.writeArray("member-profiles.json", []map[string]any{
		{"username": "manager", "role": "manager"},
		{"username": "member", "role": "member", "boundTo": "manager"},
	}); err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		writeJSON(w, http.StatusOK, map[string]string{"owner": identity.Username})
	})
	handler := novelFetchAccountScopeMiddleware(accounts, profiles, next)
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	req.Header.Set(HeaderNovelFetchAccount, "member")
	req = req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, BridgeIdentity{Username: "manager"}))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || res.Body.String() != "{\"owner\":\"member\"}\n" {
		t.Fatalf("scope response=%d %s", res.Code, res.Body.String())
	}
}
