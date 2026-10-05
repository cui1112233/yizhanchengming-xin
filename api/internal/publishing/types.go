package publishing

import (
	"errors"
	"time"
)

var (
	ErrInvalid     = errors.New("invalid publishing input")
	ErrForbidden   = errors.New("publishing access forbidden")
	ErrNotFound    = errors.New("publishing resource not found")
	ErrUnavailable = errors.New("publishing service unavailable")
)

const (
	IntentStatusPending = "pending"
	AuditResultAccepted = "accepted"
)

type CredentialInput struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

type CredentialRef struct {
	ID        string
	Platform  string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EncryptedCredential struct {
	Ref         CredentialRef
	OwnerUserID int64
	TeamID      int64
	KeyID       string
	Nonce       []byte
	Ciphertext  []byte
}

type Account struct {
	ID              int64     `json:"id"`
	OwnerUserID     int64     `json:"-"`
	TeamID          int64     `json:"-"`
	Platform        string    `json:"platform"`
	DisplayName     string    `json:"displayName"`
	CredentialRefID string    `json:"-"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type PublicAccount struct {
	ID                   int64     `json:"id"`
	Platform             string    `json:"platform"`
	DisplayName          string    `json:"displayName"`
	CredentialConfigured bool      `json:"credentialConfigured"`
	Active               bool      `json:"active"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

func (a Account) Public() PublicAccount {
	return PublicAccount{ID: a.ID, Platform: a.Platform, DisplayName: a.DisplayName, CredentialConfigured: a.CredentialRefID != "", Active: a.Active, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
}

type CreateAccountInput struct {
	Platform    string          `json:"platform"`
	DisplayName string          `json:"displayName"`
	Credential  CredentialInput `json:"credential"`
}

type Intent struct {
	ID                  int64     `json:"id"`
	BatchProjectID      int64     `json:"batchProjectId"`
	BookID              int64     `json:"bookId,omitempty"`
	PublishingAccountID int64     `json:"publishingAccountId"`
	RequestedByUserID   int64     `json:"requestedByUserId"`
	Platform            string    `json:"platform"`
	Status              string    `json:"status"`
	RequestedAt         time.Time `json:"requestedAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type CreateIntentInput struct {
	BatchProjectID      int64 `json:"batchProjectId"`
	BookID              int64 `json:"bookId,omitempty"`
	PublishingAccountID int64 `json:"publishingAccountId"`
}

type Audit struct {
	ID             int64     `json:"id"`
	IntentID       int64     `json:"intentId"`
	BatchProjectID int64     `json:"batchProjectId"`
	AccountID      int64     `json:"publishingAccountId"`
	ActorUserID    int64     `json:"actorUserId"`
	Platform       string    `json:"platform"`
	Action         string    `json:"action"`
	Result         string    `json:"result"`
	ErrorSummary   string    `json:"errorSummary,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}
