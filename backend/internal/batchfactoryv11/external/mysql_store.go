package external

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) SaveCredential(ctx context.Context, owner string, provider Provider, value encryptedCredential) (CredentialRef, error) {
	if s == nil || s.db == nil {
		return CredentialRef{}, ErrUnavailable
	}
	if value.ID == "" {
		value.ID = fmt.Sprintf("credential-%d", time.Now().UnixNano())
	}
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_external_credentials(id,owner_username,provider,credential_name,key_id,nonce,ciphertext,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id=VALUES(id),credential_name=VALUES(credential_name),key_id=VALUES(key_id),nonce=VALUES(nonce),ciphertext=VALUES(ciphertext),updated_at=VALUES(updated_at)`, value.ID, owner, provider, value.Name, value.KeyID, value.Nonce, value.Ciphertext, now, now)
	if err != nil {
		return CredentialRef{}, err
	}
	return CredentialRef{ID: value.ID, Provider: provider, Name: value.Name, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *MySQLStore) GetCredential(ctx context.Context, owner string, provider Provider) (encryptedCredential, error) {
	if s == nil || s.db == nil {
		return encryptedCredential{}, ErrUnavailable
	}
	var value encryptedCredential
	var storedProvider string
	err := s.db.QueryRowContext(ctx, `SELECT id,provider,credential_name,key_id,nonce,ciphertext,created_at,updated_at FROM batch_factory_v12_external_credentials WHERE owner_username=? AND provider=?`, owner, provider).Scan(&value.ID, &storedProvider, &value.Name, &value.KeyID, &value.Nonce, &value.Ciphertext, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return encryptedCredential{}, ErrNotFound
	}
	if err != nil {
		return encryptedCredential{}, err
	}
	value.Owner, value.Provider = owner, Provider(storedProvider)
	return value, nil
}

func (s *MySQLStore) CreateIntent(ctx context.Context, value SubmissionIntent) (SubmissionIntent, error) {
	if s == nil || s.db == nil {
		return SubmissionIntent{}, ErrUnavailable
	}
	if value.ID == "" {
		value.ID = fmt.Sprintf("intent-%d", time.Now().UnixNano())
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_external_intents(id,owner_username,provider,batch_id,book_id,payload_digest,payload_json,expires_at,confirmed_at,submitted_at,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.Owner, value.Provider, value.BatchID, value.BookID, value.PayloadDigest, string(value.Payload), value.ExpiresAt, value.ConfirmedAt, value.SubmittedAt, time.Now().UTC())
	if err != nil {
		return SubmissionIntent{}, err
	}
	return value, nil
}

func (s *MySQLStore) GetIntent(ctx context.Context, owner, id string) (SubmissionIntent, error) {
	if s == nil || s.db == nil {
		return SubmissionIntent{}, ErrUnavailable
	}
	var value SubmissionIntent
	var provider string
	var bookID string
	var payload []byte
	var confirmed, submitted sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id,provider,batch_id,book_id,payload_digest,payload_json,expires_at,confirmed_at,submitted_at FROM batch_factory_v12_external_intents WHERE id=? AND owner_username=?`, id, owner).Scan(&value.ID, &provider, &value.BatchID, &bookID, &value.PayloadDigest, &payload, &value.ExpiresAt, &confirmed, &submitted)
	if errors.Is(err, sql.ErrNoRows) {
		return SubmissionIntent{}, ErrNotFound
	}
	if err != nil {
		return SubmissionIntent{}, err
	}
	value.Owner, value.Provider, value.BookID = owner, Provider(provider), bookID
	value.Payload = append(value.Payload[:0], payload...)
	if confirmed.Valid {
		value.ConfirmedAt = &confirmed.Time
	}
	if submitted.Valid {
		value.SubmittedAt = &submitted.Time
	}
	return value, nil
}

func (s *MySQLStore) ConfirmIntent(ctx context.Context, owner, id string) (SubmissionIntent, error) {
	if s == nil || s.db == nil {
		return SubmissionIntent{}, ErrUnavailable
	}
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v12_external_intents SET confirmed_at=COALESCE(confirmed_at,?) WHERE id=? AND owner_username=? AND confirmed_at IS NULL AND expires_at>?`, now, id, owner, now)
	if err != nil {
		return SubmissionIntent{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		value, getErr := s.GetIntent(ctx, owner, id)
		if getErr != nil {
			return SubmissionIntent{}, getErr
		}
		if value.ConfirmedAt != nil {
			return value, nil
		}
		if !value.ExpiresAt.After(now) {
			return SubmissionIntent{}, ErrIntentExpired
		}
		return SubmissionIntent{}, ErrConflict
	}
	return s.GetIntent(ctx, owner, id)
}

func (s *MySQLStore) MarkSubmitted(ctx context.Context, owner, id string, at time.Time) (SubmissionIntent, error) {
	if s == nil || s.db == nil {
		return SubmissionIntent{}, ErrUnavailable
	}
	result, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v12_external_intents SET submitted_at=? WHERE id=? AND owner_username=? AND confirmed_at IS NOT NULL AND submitted_at IS NULL`, at, id, owner)
	if err != nil {
		return SubmissionIntent{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return SubmissionIntent{}, ErrConflict
	}
	return s.GetIntent(ctx, owner, id)
}

func (s *MySQLStore) AppendAudit(ctx context.Context, value AuditRecord) (AuditRecord, error) {
	if s == nil || s.db == nil {
		return AuditRecord{}, ErrUnavailable
	}
	if value.ID == "" {
		value.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}
	if len(value.Message) > 255 {
		value.Message = value.Message[:255]
	}
	if len(value.Reference) > 255 {
		value.Reference = value.Reference[:255]
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_external_audits(id,owner_username,provider,intent_id,action,outcome,reference_value,message,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, value.ID, value.Owner, value.Provider, value.IntentID, value.Action, value.Outcome, value.Reference, value.Message, value.CreatedAt)
	if err != nil {
		return AuditRecord{}, err
	}
	return value, nil
}

func (s *MySQLStore) ListAudits(ctx context.Context, owner, intentID string) ([]AuditRecord, error) {
	if s == nil || s.db == nil {
		return nil, ErrUnavailable
	}
	query := `SELECT id,provider,intent_id,action,outcome,reference_value,message,created_at FROM batch_factory_v12_external_audits WHERE owner_username=?`
	args := []any{owner}
	if intentID != "" {
		query += ` AND intent_id=?`
		args = append(args, intentID)
	}
	query += ` ORDER BY created_at,id`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditRecord{}
	for rows.Next() {
		var value AuditRecord
		var provider string
		if err := rows.Scan(&value.ID, &provider, &value.IntentID, &value.Action, &value.Outcome, &value.Reference, &value.Message, &value.CreatedAt); err != nil {
			return nil, err
		}
		value.Owner, value.Provider = owner, Provider(provider)
		out = append(out, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
