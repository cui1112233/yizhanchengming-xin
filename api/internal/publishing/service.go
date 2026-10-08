package publishing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

type Store interface {
	CreateAccountWithCredential(context.Context, Account, EncryptedCredential) (Account, error)
	ListAccountsVisible(context.Context, int64, int64, bool) ([]Account, error)
	GetAccount(context.Context, int64) (Account, error)
	ClaimBatchProject(context.Context, int64, int64, int64) error
	CanAccessBatchProject(context.Context, int64, int64, int64, bool) (bool, error)
	BookBelongsToBatchProject(context.Context, int64, int64) (bool, error)
	CreateIntentWithAudit(context.Context, Intent, Audit) (Intent, error)
	GetIntent(context.Context, int64) (Intent, error)
	ListAuditsVisible(context.Context, int64, int64, int64, bool) ([]Audit, error)
}

type Options struct {
	CredentialKey []byte
	Now           func() time.Time
}

type Service struct { store Store; key []byte; now func() time.Time }

func NewService(store Store, options Options) *Service {
	now := options.Now; if now == nil { now = time.Now }
	return &Service{store: store, key: append([]byte(nil), options.CredentialKey...), now: now}
}

func (s *Service) CreateAccount(ctx context.Context, actor authn.User, input CreateAccountInput) (Account, error) {
	if s == nil || s.store == nil || len(s.key) != 32 { return Account{}, ErrUnavailable }
	platform := strings.ToLower(strings.TrimSpace(input.Platform)); display := strings.TrimSpace(input.DisplayName); name := strings.TrimSpace(input.Credential.Name)
	if actor.ID <= 0 || platform == "" || display == "" || name == "" || input.Credential.Secret == "" { return Account{}, ErrInvalid }
	plain := []byte(input.Credential.Secret); defer zeroBytes(plain)
	keyID, nonce, ciphertext, err := EncryptCredential(s.key, plain); if err != nil { return Account{}, fmt.Errorf("%w: encrypt credential", ErrUnavailable) }
	refID, err := randomCredentialRef(); if err != nil { return Account{}, fmt.Errorf("%w: create credential reference", ErrUnavailable) }
	now := s.now().UTC()
	credential := EncryptedCredential{Ref: CredentialRef{ID: refID, Platform: platform, Name: name, CreatedAt: now, UpdatedAt: now}, OwnerUserID: actor.ID, TeamID: actor.TeamID, KeyID: keyID, Nonce: nonce, Ciphertext: ciphertext}
	account := Account{OwnerUserID: actor.ID, TeamID: actor.TeamID, Platform: platform, DisplayName: display, CredentialRefID: refID, Active: true, CreatedAt: now, UpdatedAt: now}
	return s.store.CreateAccountWithCredential(ctx, account, credential)
}

func (s *Service) ListAccounts(ctx context.Context, actor authn.User) ([]Account, error) {
	if s == nil || s.store == nil || actor.ID <= 0 { return nil, ErrForbidden }
	return s.store.ListAccountsVisible(ctx, actor.ID, actor.TeamID, elevated(actor))
}

func (s *Service) ClaimBatchProject(ctx context.Context, actor authn.User, projectID int64) error {
	if s == nil || s.store == nil { return ErrUnavailable }
	if actor.ID <= 0 || projectID <= 0 { return ErrInvalid }
	if err := s.store.ClaimBatchProject(ctx, projectID, actor.ID, actor.TeamID); err != nil { return err }
	allowed, err := s.store.CanAccessBatchProject(ctx, projectID, actor.ID, actor.TeamID, elevated(actor))
	if err != nil { return err }
	if !allowed { return ErrForbidden }
	return nil
}

func (s *Service) CreateIntent(ctx context.Context, actor authn.User, input CreateIntentInput) (Intent, error) {
	if s == nil || s.store == nil { return Intent{}, ErrUnavailable }
	if actor.ID <= 0 || input.BatchProjectID <= 0 || input.PublishingAccountID <= 0 || input.BookID < 0 { return Intent{}, ErrInvalid }
	projectAllowed, err := s.store.CanAccessBatchProject(ctx, input.BatchProjectID, actor.ID, actor.TeamID, elevated(actor))
	if err != nil { return Intent{}, err }
	if !projectAllowed { return Intent{}, ErrForbidden }
	if input.BookID > 0 {
		bookAllowed, err := s.store.BookBelongsToBatchProject(ctx, input.BatchProjectID, input.BookID)
		if err != nil { return Intent{}, err }
		if !bookAllowed { return Intent{}, ErrNotFound }
	}
	account, err := s.store.GetAccount(ctx, input.PublishingAccountID); if err != nil { return Intent{}, err }
	if !account.Active { return Intent{}, ErrForbidden }
	if !ownsAccount(actor, account) { return Intent{}, ErrForbidden }
	now := s.now().UTC()
	intent := Intent{BatchProjectID: input.BatchProjectID, BookID: input.BookID, PublishingAccountID: account.ID, RequestedByUserID: actor.ID, Platform: account.Platform, Status: IntentStatusPending, RequestedAt: now, UpdatedAt: now}
	audit := Audit{BatchProjectID: input.BatchProjectID, AccountID: account.ID, ActorUserID: actor.ID, Platform: account.Platform, Action: "intent.created", Result: AuditResultAccepted, CreatedAt: now}
	return s.store.CreateIntentWithAudit(ctx, intent, audit)
}

func (s *Service) GetIntent(ctx context.Context, actor authn.User, id int64) (Intent, error) {
	if s == nil || s.store == nil || id <= 0 { return Intent{}, ErrInvalid }
	intent, err := s.store.GetIntent(ctx, id); if err != nil { return Intent{}, err }
	projectAllowed, err := s.store.CanAccessBatchProject(ctx, intent.BatchProjectID, actor.ID, actor.TeamID, elevated(actor))
	if err != nil { return Intent{}, err }
	if !projectAllowed { return Intent{}, ErrForbidden }
	account, err := s.store.GetAccount(ctx, intent.PublishingAccountID); if err != nil { return Intent{}, err }
	if !ownsAccount(actor, account) { return Intent{}, ErrForbidden }
	return intent, nil
}

func (s *Service) ListAudits(ctx context.Context, actor authn.User, batchProjectID int64) ([]Audit, error) {
	if s == nil || s.store == nil || actor.ID <= 0 || batchProjectID < 0 { return nil, ErrInvalid }
	if batchProjectID > 0 {
		projectAllowed, err := s.store.CanAccessBatchProject(ctx, batchProjectID, actor.ID, actor.TeamID, elevated(actor))
		if err != nil { return nil, err }
		if !projectAllowed { return nil, ErrForbidden }
	}
	return s.store.ListAuditsVisible(ctx, actor.ID, actor.TeamID, batchProjectID, elevated(actor))
}

func ownsAccount(actor authn.User, account Account) bool {
	if elevated(actor) { return true }
	if actor.ID > 0 && account.OwnerUserID == actor.ID { return true }
	return actor.TeamID > 0 && account.TeamID > 0 && actor.TeamID == account.TeamID
}
func elevated(actor authn.User) bool { role := strings.ToLower(strings.TrimSpace(actor.Role)); return role == "admin" || role == "owner" }
func zeroBytes(value []byte) { for i := range value { value[i] = 0 } }
func randomCredentialRef() (string, error) { value := make([]byte, 16); if _, err := rand.Read(value); err != nil { return "", err }; return "cred_" + hex.EncodeToString(value), nil }
