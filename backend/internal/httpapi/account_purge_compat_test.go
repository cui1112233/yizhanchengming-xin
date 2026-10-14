package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestAccountPurgeCompatRequiresArchiveAndDeletesMember(t *testing.T) {
	devHash, devAlgorithm, err := auth.HashPassword("dev-password")
	if err != nil {
		t.Fatal(err)
	}
	accounts := &applicationTestStore{accounts: map[string]auth.Account{
		"owner":  {Username: "owner", PasswordHash: devHash, PasswordAlgorithm: devAlgorithm, IsOwner: true, Active: true},
		"member": {Username: "member", PasswordHash: devHash, PasswordAlgorithm: devAlgorithm, Active: false},
	}}
	profiles := NewMemberProfileStore(t.TempDir())
	if err := profiles.writeArray("member-profiles.json", []map[string]any{{"username": "member", "role": "member", "displayName": "成员"}}); err != nil {
		t.Fatal(err)
	}
	if err := profiles.writeArray("member-archives.json", []map[string]any{{"username": "member", "archivedAt": time.Now().UTC().Format(time.RFC3339)}}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "test-secret", AuthStore: accounts, Now: func() time.Time { return time.Unix(100, 0) }}
	RegisterAccountPurgeCompat(mux, middleware, accounts, profiles, "")
	req := httptest.NewRequest(http.MethodPost, "/api/account-recovery/purge/member", strings.NewReader(`{"currentPassword":"dev-password","confirmation":"member","reason":"测试"}`))
	req.Header.Set("Content-Type", "application/json")
	SignBridgeRequest(req, "owner", true, time.Unix(100, 0), "test-secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"deleted":true`) {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	if _, ok := accounts.accounts["member"]; ok {
		t.Fatal("member account still exists")
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil || payload["auditRetained"] != true {
		t.Fatalf("payload=%v err=%v", payload, err)
	}
}
