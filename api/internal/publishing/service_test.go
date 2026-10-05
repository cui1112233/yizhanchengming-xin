package publishing

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

type fakeStore struct {
	accounts          map[int64]Account
	createdAccount    Account
	credential        EncryptedCredential
	createdIntent     Intent
	createdAudit      Audit
	claimedProjectID  int64
	claimedOwnerID    int64
	claimedTeamID     int64
	denyProjectAccess bool
}

func (s *fakeStore) CreateAccountWithCredential(_ context.Context, account Account, credential EncryptedCredential) (Account, error) {
	account.ID = 41
	s.createdAccount = account
	s.credential = credential
	if s.accounts == nil { s.accounts = map[int64]Account{} }
	s.accounts[account.ID] = account
	return account, nil
}
func (s *fakeStore) ListAccountsVisible(context.Context, int64, int64, bool) ([]Account, error) {
	result := make([]Account, 0, len(s.accounts))
	for _, account := range s.accounts { result = append(result, account) }
	return result, nil
}
func (s *fakeStore) GetAccount(_ context.Context, id int64) (Account, error) {
	account, ok := s.accounts[id]
	if !ok { return Account{}, ErrNotFound }
	return account, nil
}
func (s *fakeStore) ClaimBatchProject(_ context.Context, projectID, ownerUserID, teamID int64) error {
	s.claimedProjectID = projectID
	s.claimedOwnerID = ownerUserID
	s.claimedTeamID = teamID
	return nil
}
func (s *fakeStore) CanAccessBatchProject(context.Context, int64, int64, int64, bool) (bool, error) {
	return !s.denyProjectAccess, nil
}
func (s *fakeStore) CreateIntentWithAudit(_ context.Context, intent Intent, audit Audit) (Intent, error) {
	intent.ID = 73
	audit.ID = 91
	audit.IntentID = intent.ID
	s.createdIntent = intent
	s.createdAudit = audit
	return intent, nil
}
func (s *fakeStore) GetIntent(context.Context, int64) (Intent, error) { return s.createdIntent, nil }
func (s *fakeStore) ListAuditsVisible(context.Context, int64, int64, int64, bool) ([]Audit, error) { return []Audit{s.createdAudit}, nil }

func publishingTestService(store Store, now func() time.Time) *Service {
	return NewService(store, Options{CredentialKey: bytes.Repeat([]byte{0x24}, 32), Now: now})
}

func TestCreatePublishingAccountPinsOwnershipAndEncryptsCredential(t *testing.T) {
	store := &fakeStore{}
	service := publishingTestService(store, func() time.Time { return time.Date(2026, 10, 5, 11, 0, 0, 0, time.UTC) })
	actor := authn.User{ID: 7, Username: "alice", TeamID: 3}
	secret := "platform-cookie-secret"

	account, err := service.CreateAccount(context.Background(), actor, CreateAccountInput{
		Platform: "douyin", DisplayName: "Alice 抖音", Credential: CredentialInput{Name: "alice-douyin", Secret: secret},
	})
	if err != nil { t.Fatalf("create account: %v", err) }
	if account.OwnerUserID != actor.ID || account.TeamID != actor.TeamID { t.Fatalf("ownership = user:%d team:%d", account.OwnerUserID, account.TeamID) }
	if account.CredentialRefID == "" || store.credential.Ref.ID != account.CredentialRefID { t.Fatalf("credential ref not linked: account=%#v credential=%#v", account, store.credential) }
	if bytes.Contains(store.credential.Ciphertext, []byte(secret)) || bytes.Equal(store.credential.Ciphertext, []byte(secret)) { t.Fatal("plaintext publishing secret reached persistent credential record") }
	if len(store.credential.Nonce) == 0 || store.credential.KeyID == "" { t.Fatalf("encrypted credential metadata missing: %#v", store.credential) }
	if !account.Active { t.Fatal("new publishing account must be active") }

	public := account.Public()
	if !public.CredentialConfigured || public.ID != 41 || public.Platform != "douyin" { t.Fatalf("public account = %#v", public) }
}

func TestCreatePublishingAccountRequiresEncryptionKey(t *testing.T) {
	service := NewService(&fakeStore{}, Options{})
	_, err := service.CreateAccount(context.Background(), authn.User{ID: 7}, CreateAccountInput{Platform: "douyin", DisplayName: "x", Credential: CredentialInput{Name: "x", Secret: "secret"}})
	if !errors.Is(err, ErrUnavailable) { t.Fatalf("err=%v want ErrUnavailable", err) }
}

func TestClaimBatchProjectPinsCreatorOwnership(t *testing.T) {
	store := &fakeStore{}
	service := publishingTestService(store, time.Now)
	actor := authn.User{ID: 7, TeamID: 3}
	if err := service.ClaimBatchProject(context.Background(), actor, 21); err != nil { t.Fatalf("claim project: %v", err) }
	if store.claimedProjectID != 21 || store.claimedOwnerID != 7 || store.claimedTeamID != 3 {
		t.Fatalf("claim = project:%d owner:%d team:%d", store.claimedProjectID, store.claimedOwnerID, store.claimedTeamID)
	}
}

func TestCreatePublishIntentAllowsSameTeamAndDerivesPlatformFromAccount(t *testing.T) {
	store := &fakeStore{accounts: map[int64]Account{11: {ID: 11, OwnerUserID: 99, TeamID: 3, Platform: "douyin", CredentialRefID: "cred-11", Active: true}}}
	service := publishingTestService(store, func() time.Time { return time.Date(2026, 10, 5, 11, 5, 0, 0, time.UTC) })
	actor := authn.User{ID: 7, TeamID: 3}

	intent, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{BatchProjectID: 21, BookID: 22, PublishingAccountID: 11})
	if err != nil { t.Fatalf("create intent: %v", err) }
	if intent.RequestedByUserID != actor.ID || intent.Platform != "douyin" || intent.Status != IntentStatusPending { t.Fatalf("intent = %#v", intent) }
	if store.createdAudit.ActorUserID != actor.ID || store.createdAudit.Result != AuditResultAccepted || store.createdAudit.AccountID != 11 { t.Fatalf("audit = %#v", store.createdAudit) }
	if store.createdAudit.ErrorSummary != "" { t.Fatalf("unexpected audit error: %#v", store.createdAudit) }
}

func TestCreatePublishIntentRejectsForeignProjectWithoutPersistingIntent(t *testing.T) {
	store := &fakeStore{
		accounts: map[int64]Account{11: {ID: 11, OwnerUserID: 7, TeamID: 3, Platform: "douyin", CredentialRefID: "cred-11", Active: true}},
		denyProjectAccess: true,
	}
	service := publishingTestService(store, time.Now)
	actor := authn.User{ID: 7, TeamID: 3}
	_, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{BatchProjectID: 999, PublishingAccountID: 11})
	if !errors.Is(err, ErrForbidden) { t.Fatalf("err=%v, want ErrForbidden", err) }
	if store.createdIntent.ID != 0 || store.createdAudit.ID != 0 { t.Fatalf("foreign project persisted work: intent=%#v audit=%#v", store.createdIntent, store.createdAudit) }
}

func TestCreatePublishIntentRejectsForeignOwnershipWithoutPersistingIntent(t *testing.T) {
	store := &fakeStore{accounts: map[int64]Account{11: {ID: 11, OwnerUserID: 99, TeamID: 8, Platform: "douyin", CredentialRefID: "cred-11", Active: true}}}
	service := publishingTestService(store, time.Now)
	actor := authn.User{ID: 7, TeamID: 3}
	_, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{BatchProjectID: 21, PublishingAccountID: 11})
	if !errors.Is(err, ErrForbidden) { t.Fatalf("err=%v, want ErrForbidden", err) }
	if store.createdIntent.ID != 0 || store.createdAudit.ID != 0 { t.Fatalf("foreign account persisted work: intent=%#v audit=%#v", store.createdIntent, store.createdAudit) }
}

func TestAdminCanUseAccountAcrossOwnershipBoundary(t *testing.T) {
	store := &fakeStore{accounts: map[int64]Account{11: {ID: 11, OwnerUserID: 99, TeamID: 8, Platform: "douyin", CredentialRefID: "cred-11", Active: true}}}
	service := publishingTestService(store, time.Now)
	actor := authn.User{ID: 1, Role: "admin"}
	if _, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{BatchProjectID: 21, PublishingAccountID: 11}); err != nil { t.Fatalf("admin publish authorization: %v", err) }
}
