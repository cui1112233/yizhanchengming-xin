package publishing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type MySQLStore struct { db *sql.DB }
func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) CreateAccountWithCredential(ctx context.Context, account Account, credential EncryptedCredential) (Account, error) {
	if s == nil || s.db == nil { return Account{}, ErrUnavailable }
	tx, err := s.db.BeginTx(ctx, nil); if err != nil { return Account{}, err }; defer tx.Rollback()
	var team any; if credential.TeamID > 0 { team = credential.TeamID }
	_, err = tx.ExecContext(ctx, `INSERT INTO publishing_credentials (ref, owner_user_id, team_id, platform, name, key_id, nonce, ciphertext, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, credential.Ref.ID, credential.OwnerUserID, team, credential.Ref.Platform, credential.Ref.Name, credential.KeyID, credential.Nonce, credential.Ciphertext, credential.Ref.CreatedAt, credential.Ref.UpdatedAt)
	if err != nil { return Account{}, fmt.Errorf("insert publishing credential: %w", err) }
	if account.TeamID > 0 { team = account.TeamID } else { team = nil }
	result, err := tx.ExecContext(ctx, `INSERT INTO publishing_accounts (owner_user_id, team_id, platform, display_name, credential_ref, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, account.OwnerUserID, team, account.Platform, account.DisplayName, account.CredentialRefID, account.Active, account.CreatedAt, account.UpdatedAt)
	if err != nil { return Account{}, fmt.Errorf("insert publishing account: %w", err) }
	account.ID, err = result.LastInsertId(); if err != nil { return Account{}, err }
	if err := tx.Commit(); err != nil { return Account{}, err }
	return account, nil
}

func (s *MySQLStore) ListAccountsVisible(ctx context.Context, userID, teamID int64, elevated bool) ([]Account, error) {
	query := `SELECT id, owner_user_id, COALESCE(team_id,0), platform, display_name, credential_ref, active, created_at, updated_at FROM publishing_accounts`
	args := []any{}
	if !elevated { query += ` WHERE owner_user_id = ? OR (team_id IS NOT NULL AND team_id = ?)`; args = append(args, userID, teamID) }
	query += ` ORDER BY id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...); if err != nil { return nil, err }; defer rows.Close()
	var values []Account
	for rows.Next() { var a Account; if err := rows.Scan(&a.ID,&a.OwnerUserID,&a.TeamID,&a.Platform,&a.DisplayName,&a.CredentialRefID,&a.Active,&a.CreatedAt,&a.UpdatedAt); err != nil { return nil, err }; values = append(values,a) }
	return values, rows.Err()
}

func (s *MySQLStore) GetAccount(ctx context.Context, id int64) (Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx, `SELECT id, owner_user_id, COALESCE(team_id,0), platform, display_name, credential_ref, active, created_at, updated_at FROM publishing_accounts WHERE id = ?`, id).Scan(&a.ID,&a.OwnerUserID,&a.TeamID,&a.Platform,&a.DisplayName,&a.CredentialRefID,&a.Active,&a.CreatedAt,&a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) { return Account{}, ErrNotFound }; return a, err
}

func (s *MySQLStore) ClaimBatchProject(ctx context.Context, projectID, ownerUserID, teamID int64) error {
	if s == nil || s.db == nil { return ErrUnavailable }
	var team any
	if teamID > 0 { team = teamID }
	_, err := s.db.ExecContext(ctx, `INSERT INTO auth_batch_project_ownership (batch_project_id, owner_user_id, team_id) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE batch_project_id = VALUES(batch_project_id)`, projectID, ownerUserID, team)
	return err
}

func (s *MySQLStore) CanAccessBatchProject(ctx context.Context, projectID, userID, teamID int64, elevated bool) (bool, error) {
	if s == nil || s.db == nil { return false, ErrUnavailable }
	var allowed bool
	if elevated {
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM batch_projects WHERE id = ?)`, projectID).Scan(&allowed)
		return allowed, err
	}
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM auth_batch_project_ownership WHERE batch_project_id = ? AND (owner_user_id = ? OR (? > 0 AND team_id IS NOT NULL AND team_id = ?)))`, projectID, userID, teamID, teamID).Scan(&allowed)
	return allowed, err
}

func (s *MySQLStore) IsBatchProjectArchived(ctx context.Context, projectID int64) (bool, error) {
	if s == nil || s.db == nil { return false, ErrUnavailable }
	var archived bool
	err := s.db.QueryRowContext(ctx, `SELECT archived_at IS NOT NULL FROM batch_projects WHERE id = ?`, projectID).Scan(&archived)
	if errors.Is(err, sql.ErrNoRows) { return false, ErrNotFound }
	return archived, err
}

func (s *MySQLStore) BookBelongsToBatchProject(ctx context.Context, projectID, bookID int64) (bool, error) {
	if s == nil || s.db == nil { return false, ErrUnavailable }
	var allowed bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM batch_projects p JOIN books b ON b.intake_id = p.intake_id WHERE p.id = ? AND b.id = ?)`, projectID, bookID).Scan(&allowed)
	return allowed, err
}

func (s *MySQLStore) CreateIntentWithAudit(ctx context.Context, intent Intent, audit Audit) (Intent, error) {
	tx, err := s.db.BeginTx(ctx, nil); if err != nil { return Intent{}, err }; defer tx.Rollback()
	var archivedAt sql.NullTime; err = tx.QueryRowContext(ctx, `SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE`, intent.BatchProjectID).Scan(&archivedAt); if errors.Is(err, sql.ErrNoRows) { return Intent{}, ErrNotFound }; if err != nil { return Intent{}, err }; if archivedAt.Valid { return Intent{}, ErrProjectArchived }
	var book any; if intent.BookID > 0 { book = intent.BookID }
	result, err := tx.ExecContext(ctx, `INSERT INTO publish_intents (batch_project_id, book_id, publishing_account_id, requested_by_user_id, platform, status, requested_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, intent.BatchProjectID, book, intent.PublishingAccountID, intent.RequestedByUserID, intent.Platform, intent.Status, intent.RequestedAt, intent.UpdatedAt)
	if err != nil { return Intent{}, err }
	intent.ID, err = result.LastInsertId(); if err != nil { return Intent{}, err }
	_, err = tx.ExecContext(ctx, `INSERT INTO publish_audits (intent_id, batch_project_id, publishing_account_id, actor_user_id, platform, action, result, error_summary, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, intent.ID, audit.BatchProjectID, audit.AccountID, audit.ActorUserID, audit.Platform, audit.Action, audit.Result, audit.ErrorSummary, audit.CreatedAt)
	if err != nil { return Intent{}, err }
	if err := tx.Commit(); err != nil { return Intent{}, err }
	return intent, nil
}

func (s *MySQLStore) GetIntent(ctx context.Context, id int64) (Intent, error) {
	var v Intent; var book sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT id, batch_project_id, book_id, publishing_account_id, requested_by_user_id, platform, status, requested_at, updated_at FROM publish_intents WHERE id = ?`, id).Scan(&v.ID,&v.BatchProjectID,&book,&v.PublishingAccountID,&v.RequestedByUserID,&v.Platform,&v.Status,&v.RequestedAt,&v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) { return Intent{}, ErrNotFound }; if err != nil { return Intent{}, err }; if book.Valid { v.BookID = book.Int64 }; return v,nil
}

func (s *MySQLStore) ListAuditsVisible(ctx context.Context, userID, teamID, batchProjectID int64, elevated bool) ([]Audit, error) {
	query := `SELECT a.id, a.intent_id, a.batch_project_id, a.publishing_account_id, a.actor_user_id, a.platform, a.action, a.result, a.error_summary, a.created_at FROM publish_audits a JOIN publishing_accounts p ON p.id = a.publishing_account_id WHERE 1=1`
	args := []any{}
	if batchProjectID > 0 { query += ` AND a.batch_project_id = ?`; args = append(args,batchProjectID) }
	if !elevated { query += ` AND (p.owner_user_id = ? OR (p.team_id IS NOT NULL AND p.team_id = ?))`; args = append(args,userID,teamID) }
	query += ` ORDER BY a.id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...); if err != nil { return nil, err }; defer rows.Close()
	var values []Audit
	for rows.Next() { var a Audit; if err := rows.Scan(&a.ID,&a.IntentID,&a.BatchProjectID,&a.AccountID,&a.ActorUserID,&a.Platform,&a.Action,&a.Result,&a.ErrorSummary,&a.CreatedAt); err != nil { return nil,err }; values=append(values,a) }
	return values, rows.Err()
}
