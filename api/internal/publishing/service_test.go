package publishing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

type fakeStore struct {
	accounts      map[int64]Account
	createdAccount Account
	createdIntent  Intent
	createdAudit   Audit
}

func (s *fakeStore) CreateAccount(_ context.Context, account Account) (Account, error) {
	account.ID = 41
	s.createdAccount = account
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

func (s *fakeStore) CreateIntentWithAudit(_ context.Context, intent Intent, audit Audit) (Intent, error) {
	intent.ID = 73
	audit.ID = 91
	audit.IntentID = intent.ID
	s.createdIntent = intent
	s.createdAudit = audit
	return intent, nil
}

func (s *fakeStore) GetIntent(context.Context, int64) (Intent, error) { return s.createdIntent, nil }
func (s *fakeStore) ListAuditsVisible(context.Context, int64, int64, int64, bool) ([]Audit, error) {
	return []Audit{s.createdAudit}, nil
}

func TestCreatePublishingAccountPinsOwnershipAndNeverSerializesCredentialRef(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, func() time.Time { return time.Date(2026, 10, 5, 11, 0, 0, 0, time.UTC) })
	actor := authn.User{ID: 7, Username: "alice", TeamID: 3}

	account, err := service.CreateAccount(context.Background(), actor, CreateAccountInput{
		Platform: "douyin", DisplayName: "Alice 抖音", CredentialRef: "vault://publisher/alice/douyin",
	})
	if err != nil { t.Fatalf("create account: %v", err) }
	if account.OwnerUserID != actor.ID || account.TeamID != actor.TeamID {
		t.Fatalf("ownership = user:%d team:%d, want user:%d team:%d", account.OwnerUserID, account.TeamID, actor.ID, actor.TeamID)
	}
	if account.CredentialRef != "vault://publisher/alice/douyin" { t.Fatalf("credential ref was not persisted server-side: %#v", account) }
	if !account.Active { t.Fatal("new publishing account must be active") }

	encoded := account.Public()
	if encoded.CredentialConfigured != true { t.Fatal("public account should expose only configured=true") }
	if encoded.ID != 41 || encoded.Platform != "douyin" { t.Fatalf("public account = %#v", encoded) }
}

func TestCreatePublishIntentAllowsSameTeamAndDerivesPlatformFromAccount(t *testing.T) {
	store := &fakeStore{accounts: map[int64]Account{
		11: {ID: 11, OwnerUserID: 99, TeamID: 3, Platform: "douyin", CredentialRef: "vault://secret", Active: true},
	}}
	service := NewService(store, func() time.Time { return time.Date(2026, 10, 5, 11, 5, 0, 0, time.UTC) })
	actor := authn.User{ID: 7, TeamID: 3}

	intent, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{
		BatchProjectID: 21, BookID: 22, PublishingAccountID: 11, OutputRef: "tos://video/final.mp4",
	})
	if err != nil { t.Fatalf("create intent: %v", err) }
	if intent.RequestedByUserID != actor.ID || intent.Platform != "douyin" || intent.Status != IntentStatusPending {
		t.Fatalf("intent = %#v", intent)
	}
	if store.createdAudit.ActorUserID != actor.ID || store.createdAudit.Result != AuditResultAccepted || store.createdAudit.AccountID != 11 {
		t.Fatalf("audit = %#v", store.createdAudit)
	}
	if store.createdAudit.ErrorSummary != "" { t.Fatalf("unexpected audit error: %#v", store.createdAudit) }
}

func TestCreatePublishIntentRejectsForeignOwnershipWithoutPersistingIntent(t *testing.T) {
	store := &fakeStore{accounts: map[int64]Account{
		11: {ID: 11, OwnerUserID: 99, TeamID: 8, Platform: "douyin", CredentialRef: "vault://secret", Active: true},
	}}
	service := NewService(store, time.Now)
	actor := authn.User{ID: 7, TeamID: 3}

	_, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{BatchProjectID: 21, PublishingAccountID: 11, OutputRef: "tos://video/final.mp4"})
	if !errors.Is(err, ErrForbidden) { t.Fatalf("err=%v, want ErrForbidden", err) }
	if store.createdIntent.ID != 0 || store.createdAudit.ID != 0 { t.Fatalf("foreign account must not persist publish work: intent=%#v audit=%#v", store.createdIntent, store.createdAudit) }
}

func TestAdminCanUseAccountAcrossOwnershipBoundary(t *testing.T) {
	store := &fakeStore{accounts: map[int64]Account{
		11: {ID: 11, OwnerUserID: 99, TeamID: 8, Platform: "douyin", CredentialRef: "vault://secret", Active: true},
	}}
	service := NewService(store, time.Now)
	actor := authn.User{ID: 1, Role: "admin"}
	if _, err := service.CreateIntent(context.Background(), actor, CreateIntentInput{BatchProjectID: 21, PublishingAccountID: 11, OutputRef: "tos://video/final.mp4"}); err != nil {
		t.Fatalf("admin publish authorization: %v", err)
	}
}
