package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

type governanceAuthStore struct {
	accounts map[string]auth.Account
}

func (s *governanceAuthStore) FindAccount(_ context.Context, username string) (auth.Account, error) {
	account, ok := s.accounts[username]
	if !ok {
		return auth.Account{}, auth.ErrNotFound
	}
	return account, nil
}
func (s *governanceAuthStore) CreateAccount(_ context.Context, account auth.Account) error {
	if _, exists := s.accounts[account.Username]; exists {
		return auth.ErrInvalid
	}
	s.accounts[account.Username] = account
	return nil
}
func (s *governanceAuthStore) UpdatePassword(_ context.Context, username, hash, algorithm string) error {
	account, ok := s.accounts[username]
	if !ok {
		return auth.ErrNotFound
	}
	account.PasswordHash, account.PasswordAlgorithm, account.UpdatedAt = hash, algorithm, time.Now().UTC()
	s.accounts[username] = account
	return nil
}
func (s *governanceAuthStore) SetActive(_ context.Context, username string, active bool) error {
	account, ok := s.accounts[username]
	if !ok {
		return auth.ErrNotFound
	}
	account.Active, account.UpdatedAt = active, time.Now().UTC()
	s.accounts[username] = account
	return nil
}
func (s *governanceAuthStore) CreateSession(context.Context, auth.Session) error { return nil }
func (s *governanceAuthStore) FindSession(context.Context, string, time.Time) (auth.Session, error) {
	return auth.Session{}, auth.ErrNotFound
}
func (s *governanceAuthStore) RevokeSession(context.Context, string) error { return nil }

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

func TestAdminAccountOperationsPersistInGoStore(t *testing.T) {
	profiles := NewMemberProfileStore(t.TempDir())
	store := &governanceAuthStore{accounts: map[string]auth.Account{
		"owner": {Username: "owner", Active: true, IsOwner: true},
	}}
	ctx := context.Background()
	created, err := createAdminAccount(ctx, profiles, store, "owner", map[string]any{"username": "member1", "password": "initial-password"})
	if err != nil || created["username"] != "member1" {
		t.Fatalf("create account = %#v, %v", created, err)
	}
	if _, err := setAdminAccountStatus(ctx, profiles, store, "owner", "member1", false); err != nil {
		t.Fatalf("disable account: %v", err)
	}
	if store.accounts["member1"].Active {
		t.Fatal("account should be inactive after status update")
	}
	oldHash := store.accounts["member1"].PasswordHash
	if _, err := resetAdminAccountPassword(ctx, profiles, store, "owner", "member1", "replacement-password"); err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if store.accounts["member1"].PasswordHash == oldHash {
		t.Fatal("password hash did not change")
	}
}
