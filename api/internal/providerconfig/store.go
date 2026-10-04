package providerconfig

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	KindImage = "image"
	KindVideo = "video"
)

var ErrNotFound = errors.New("generation provider config not found")

type PutInput struct {
	Owner      string
	MediaKind  string
	Provider   string
	Model      string
	APIKey     string
	CreateURL  string
	TasksURL   string
	ResultURL  string
	Settings   json.RawMessage
	Enabled    bool
}

type Record struct {
	ID         string          `json:"id"`
	Owner      string          `json:"-"`
	MediaKind  string          `json:"media_kind"`
	Provider   string          `json:"provider"`
	Model      string          `json:"model"`
	CreateURL  string          `json:"create_url,omitempty"`
	TasksURL   string          `json:"tasks_url,omitempty"`
	ResultURL  string          `json:"result_url,omitempty"`
	Settings   json.RawMessage `json:"settings,omitempty"`
	Enabled    bool            `json:"enabled"`
	Configured bool            `json:"configured"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Resolved struct {
	Record
	APIKey string `json:"-"`
}

type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type rowScanner interface { Scan(...any) error }
type queryRowFunc func(context.Context, string, ...any) rowScanner

type SQLStore struct {
	exec  sqlExecutor
	query queryRowFunc
	cipher *Cipher
}

func NewSQLStore(db *sql.DB, cipher *Cipher) *SQLStore {
	return newSQLStore(db, func(ctx context.Context, query string, args ...any) rowScanner {
		return db.QueryRowContext(ctx, query, args...)
	}, cipher)
}

func newSQLStore(exec sqlExecutor, query queryRowFunc, cipher *Cipher) *SQLStore {
	return &SQLStore{exec: exec, query: query, cipher: cipher}
}

func (s *SQLStore) Put(ctx context.Context, input PutInput) (Record, error) {
	if s == nil || s.exec == nil || s.cipher == nil { return Record{}, errors.New("provider config store unavailable") }
	input.Owner = strings.TrimSpace(input.Owner)
	input.MediaKind = strings.ToLower(strings.TrimSpace(input.MediaKind))
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.Model = strings.TrimSpace(input.Model)
	input.APIKey = strings.TrimSpace(input.APIKey)
	input.CreateURL = strings.TrimSpace(input.CreateURL)
	input.TasksURL = strings.TrimSpace(input.TasksURL)
	input.ResultURL = strings.TrimSpace(input.ResultURL)
	if input.Owner == "" || input.Provider == "" { return Record{}, errors.New("owner and provider are required") }
	if input.MediaKind != KindImage && input.MediaKind != KindVideo { return Record{}, errors.New("media_kind must be image or video") }
	if len(input.Settings) > 0 && !json.Valid(input.Settings) { return Record{}, errors.New("provider settings must be valid json") }

	var nonce, ciphertext []byte
	var err error
	if input.APIKey != "" {
		nonce, ciphertext, err = s.cipher.Encrypt(input.Owner, input.MediaKind, input.Provider, []byte(input.APIKey))
		if err != nil { return Record{}, err }
	}
	id, err := newConfigID()
	if err != nil { return Record{}, err }
	var settings any
	if len(input.Settings) > 0 { settings = []byte(input.Settings) }
	_, err = s.exec.ExecContext(ctx, `INSERT INTO generation_provider_configs
(id, owner, media_kind, provider, model, create_url, tasks_url, result_url, credential_nonce, credential_ciphertext, settings_json, enabled)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE model=VALUES(model), create_url=VALUES(create_url), tasks_url=VALUES(tasks_url), result_url=VALUES(result_url), credential_nonce=COALESCE(VALUES(credential_nonce), credential_nonce), credential_ciphertext=COALESCE(VALUES(credential_ciphertext), credential_ciphertext), settings_json=VALUES(settings_json), enabled=VALUES(enabled), updated_at=CURRENT_TIMESTAMP(6)`,
		id, input.Owner, input.MediaKind, input.Provider, input.Model, input.CreateURL, input.TasksURL, input.ResultURL, nullableBytes(nonce), nullableBytes(ciphertext), settings, input.Enabled)
	if err != nil { return Record{}, err }
	now := time.Now().UTC()
	return Record{ID:id, Owner:input.Owner, MediaKind:input.MediaKind, Provider:input.Provider, Model:input.Model, CreateURL:input.CreateURL, TasksURL:input.TasksURL, ResultURL:input.ResultURL, Settings:append(json.RawMessage(nil), input.Settings...), Enabled:input.Enabled, Configured:input.APIKey!="", CreatedAt:now, UpdatedAt:now}, nil
}

func (s *SQLStore) Resolve(ctx context.Context, owner, mediaKind, provider string) (Resolved, error) {
	if s == nil || s.query == nil || s.cipher == nil { return Resolved{}, errors.New("provider config store unavailable") }
	owner = strings.TrimSpace(owner)
	mediaKind = strings.ToLower(strings.TrimSpace(mediaKind))
	provider = strings.ToLower(strings.TrimSpace(provider))
	if owner == "" || provider == "" { return Resolved{}, errors.New("owner and provider are required") }
	var record Record
	var nonce, ciphertext, settings []byte
	err := s.query(ctx, `SELECT id, owner, media_kind, provider, model, create_url, tasks_url, result_url, credential_nonce, credential_ciphertext, settings_json, enabled, created_at, updated_at
FROM generation_provider_configs WHERE owner=? AND media_kind=? AND provider=? LIMIT 1`, owner, mediaKind, provider).Scan(
		&record.ID, &record.Owner, &record.MediaKind, &record.Provider, &record.Model, &record.CreateURL, &record.TasksURL, &record.ResultURL,
		&nonce, &ciphertext, &settings, &record.Enabled, &record.CreatedAt, &record.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) { return Resolved{}, ErrNotFound }
	if err != nil { return Resolved{}, err }
	if len(settings) > 0 { record.Settings = append(json.RawMessage(nil), settings...) }
	resolved := Resolved{Record: record}
	if len(ciphertext) > 0 {
		plain, err := s.cipher.Decrypt(record.Owner, record.MediaKind, record.Provider, nonce, ciphertext)
		if err != nil { return Resolved{}, err }
		resolved.APIKey = string(plain)
		resolved.Record.Configured = true
	}
	return resolved, nil
}

func newConfigID() (string,error) {
	buf := make([]byte,16)
	if _,err := rand.Read(buf); err != nil { return "",err }
	return "cfg_"+hex.EncodeToString(buf),nil
}

func nullableBytes(value []byte) any {
	if len(value)==0 { return nil }
	return append([]byte(nil), value...)
}
