package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/auth"
	"testing"
	"time"
)

func TestMemberUsageReadsAuthenticatedLedgerAndReportsFailure(t *testing.T) {
	now := time.Now()
	accounts := &authTestStore{account: auth.Account{Username: "alice", Active: true}}
	profiles := NewMemberProfileStore(t.TempDir())
	failed := false
	profiles.UsageReader = func(ctx context.Context, owner string) (map[string]any, error) {
		if owner != "alice" {
			t.Fatal("wrong ledger owner", owner)
		}
		if failed {
			return nil, errors.New("database unavailable")
		}
		return map[string]any{"month": map[string]any{"totalTokens": int64(31), "actualCost": .049}, "recent": []any{}, "source": "yzcm_billing_log"}, nil
	}
	mux := http.NewServeMux()
	RegisterMemberProfileRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: accounts}, profiles, accounts, t.TempDir())
	request := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", "/api/member/me?username=bob", nil)
		SignBridgeRequest(r, "alice", false, now, "test-secret")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	w := request()
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var body struct {
		Usage map[string]any `json:"usage"`
		Quota map[string]any `json:"memberQuota"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Usage["source"] != "yzcm_billing_log" || body.Quota["used"] != float64(31) {
		t.Fatal(body)
	}
	failed = true
	w = request()
	if w.Code != 503 {
		t.Fatal("ledger failure was hidden", w.Code, w.Body.String())
	}
}

type usageAccounts struct{ *authTestStore }

func (s usageAccounts) FindAccount(ctx context.Context, name string) (auth.Account, error) {
	if name == "alice" || name == "bob" {
		return auth.Account{Username: name, Active: true}, nil
	}
	return auth.Account{}, auth.ErrNotFound
}
func TestMemberUsageCannotReadAnotherMembersSharedKeyLedger(t *testing.T) {
	now := time.Now()
	accounts := usageAccounts{&authTestStore{account: auth.Account{Username: "alice", Active: true}}}
	profiles := NewMemberProfileStore(t.TempDir())
	profiles.UsageReader = func(context.Context, string) (map[string]any, error) {
		t.Fatal("unauthorized ledger was read")
		return nil, nil
	}
	mux := http.NewServeMux()
	RegisterMemberTeamReadCompat(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: accounts}, accounts, profiles)
	r := httptest.NewRequest("GET", "/api/member/team/members/bob/usage", nil)
	SignBridgeRequest(r, "alice", false, now, "test-secret")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatal(w.Code, w.Body.String())
	}
}
