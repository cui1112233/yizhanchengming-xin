package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
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
