package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestMemberProfileIncludesDetailsAndOwnerTeamShape(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "member-profile-details.json"), []byte(`[{"username":"alice","bio":"个人简介","email":"alice@example.com","teamTitle":"负责人"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewMemberProfileStore(dir)
	accounts := &authTestStore{account: auth.Account{Username: "alice", IsOwner: true, Active: true, CreatedAt: time.Unix(100, 0)}}
	member := store.member("alice", accounts)
	if member["bio"] != "个人简介" || member["email"] != "alice@example.com" || member["teamTitle"] != "负责人" {
		t.Fatalf("member details = %#v", member)
	}
	if member["apiEnabled"] != false {
		t.Fatalf("owner apiEnabled should remain account preference: %#v", member["apiEnabled"])
	}
	team := goTeam(member)
	if team["managerUsername"] != "alice" || team["name"] != "alice的团队" {
		t.Fatalf("team = %#v", team)
	}
	usage := emptyMemberUsage()
	if _, ok := usage["month"].(map[string]any); !ok {
		t.Fatalf("usage = %#v", usage)
	}
	if _, err := json.Marshal(map[string]any{"member": member, "team": team, "usage": usage}); err != nil {
		t.Fatal(err)
	}
}

func TestMemberSecuritySeparatesRememberedSessions(t *testing.T) {
	now := time.Date(2026, 10, 11, 3, 0, 0, 0, time.UTC)
	accounts := &authTestStore{
		account: auth.Account{Username: "alice", IsOwner: true, Active: true, CreatedAt: now},
		sessions: []auth.Session{
			{Token: "runtime-token-hash", Username: "alice", IssuedAt: now, ExpiresAt: now.Add(24 * time.Hour), IPHint: "127.0.0.1"},
			{Token: "remembered-token-hash", Username: "alice", IssuedAt: now, ExpiresAt: now.Add(30 * 24 * time.Hour), IPHint: "127.0.0.1"},
		},
	}
	mux := http.NewServeMux()
	RegisterMemberProfileRoutes(mux, BridgeAuth{Secret: "test-secret", Now: func() time.Time { return now }, AuthStore: accounts}, NewMemberProfileStore(t.TempDir()), accounts, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/member/security", nil)
	SignBridgeRequest(req, "alice", true, now, "test-secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("security status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(res.Body.String())).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if got := len(payload["sessions"].([]any)); got != 1 {
		t.Fatalf("runtime sessions=%d, want 1: %#v", got, payload["sessions"])
	}
	persistent, ok := payload["persistentSessions"].([]any)
	if !ok || len(persistent) != 1 {
		t.Fatalf("persistent sessions=%#v, want one remembered session", payload["persistentSessions"])
	}
	if _, ok := persistent[0].(map[string]any)["expiresAt"]; !ok {
		t.Fatalf("remembered session must expose expiry: %#v", persistent[0])
	}
}

var _ interface {
	ListSessionsForUser(context.Context, string) ([]auth.Session, error)
} = (*authTestStore)(nil)
