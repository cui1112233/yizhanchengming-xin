package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

func TestAdminGovernanceRoutesListApplicationsAndGrantsWithoutLeakingPassword(t *testing.T) {
	now := time.Unix(1700000000, 0)
	profiles := NewMemberProfileStore(t.TempDir())
	if err := os.WriteFile(filepath.Join(profiles.SystemDir, "applications.json"), []byte(`[{"id":"app-1","username":"new-user","passwordHash":"secret-hash","passwordAlgorithm":"scrypt","reason":"测试","status":"pending"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	accounts := &authTestStore{account: auth.Account{Username: "alice", Active: true, IsOwner: true}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AuthStore: accounts, MemberProfileStore: profiles})

	applications := signedJSONRequest(t, api, now, "alice", "GET", "/api/admin/applications", nil)
	if applications.Code != 200 || !strings.Contains(applications.Body.String(), "app-1") || strings.Contains(applications.Body.String(), "secret-hash") {
		t.Fatalf("applications response = %d %s", applications.Code, applications.Body.String())
	}
	grants := signedJSONRequest(t, api, now, "alice", "GET", "/api/admin/grants", nil)
	if grants.Code != 200 || !strings.Contains(grants.Body.String(), `"grants":[]`) {
		t.Fatalf("grants response = %d %s", grants.Code, grants.Body.String())
	}
}
