package authn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) FindLoginCredential(ctx context.Context, username string) (LoginCredential, error) {
	if s == nil || s.db == nil {
		return LoginCredential{}, ErrUnauthenticated
	}
	var credential LoginCredential
	err := s.db.QueryRowContext(ctx, `SELECT id, username, display_name, password_hash, role, COALESCE(team_id, 0), active FROM auth_users WHERE username = ? LIMIT 1`, username).Scan(
		&credential.User.ID,
		&credential.User.Username,
		&credential.User.DisplayName,
		&credential.PasswordHash,
		&credential.User.Role,
		&credential.User.TeamID,
		&credential.Active,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return LoginCredential{}, ErrUnauthenticated
	}
	if err != nil {
		return LoginCredential{}, fmt.Errorf("find login credential: %w", err)
	}
	capabilities, err := loadCapabilities(ctx, s.db, credential.User.ID)
	if err != nil {
		return LoginCredential{}, err
	}
	credential.User.Capabilities = capabilities
	return credential, nil
}

func (s *MySQLStore) CreateSession(ctx context.Context, record SessionRecord) error {
	if s == nil || s.db == nil {
		return errors.New("auth session store is unavailable")
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO auth_sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at) VALUES (?, ?, ?, ?, ?)`,
		record.UserID,
		record.AccessTokenHash,
		record.RefreshTokenHash,
		record.AccessExpiresAt,
		record.RefreshExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create auth session: %w", err)
	}
	return nil
}

func (s *MySQLStore) ResolveAccess(ctx context.Context, accessHash string, now time.Time) (User, error) {
	if s == nil || s.db == nil || accessHash == "" {
		return User{}, ErrUnauthenticated
	}
	var user User
	err := s.db.QueryRowContext(ctx, `SELECT u.id, u.username, u.display_name, u.role, COALESCE(u.team_id, 0) FROM auth_sessions s JOIN auth_users u ON u.id = s.user_id WHERE s.access_token_hash = ? AND s.revoked_at IS NULL AND s.access_expires_at > ? AND u.active = TRUE LIMIT 1`, accessHash, now).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Role,
		&user.TeamID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthenticated
	}
	if err != nil {
		return User{}, fmt.Errorf("resolve auth session: %w", err)
	}
	capabilities, err := loadCapabilities(ctx, s.db, user.ID)
	if err != nil {
		return User{}, err
	}
	user.Capabilities = capabilities
	return user, nil
}

func (s *MySQLStore) RotateByRefresh(ctx context.Context, oldRefreshHash string, next SessionRecord, now time.Time) (User, error) {
	if s == nil || s.db == nil || oldRefreshHash == "" {
		return User{}, ErrUnauthenticated
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin refresh transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var userID int64
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM auth_sessions WHERE refresh_token_hash = ? AND revoked_at IS NULL AND refresh_expires_at > ? FOR UPDATE`, oldRefreshHash, now).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthenticated
	}
	if err != nil {
		return User{}, fmt.Errorf("lock refresh session: %w", err)
	}

	result, err := tx.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at = ?, updated_at = CURRENT_TIMESTAMP(6) WHERE refresh_token_hash = ? AND revoked_at IS NULL`, now, oldRefreshHash)
	if err != nil {
		return User{}, fmt.Errorf("revoke refresh session: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("verify refresh revocation: %w", err)
	}
	if rows != 1 {
		return User{}, ErrUnauthenticated
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO auth_sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at) VALUES (?, ?, ?, ?, ?)`,
		userID,
		next.AccessTokenHash,
		next.RefreshTokenHash,
		next.AccessExpiresAt,
		next.RefreshExpiresAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("insert rotated auth session: %w", err)
	}

	user, err := loadActiveUserByID(ctx, tx, userID)
	if err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit refresh transaction: %w", err)
	}
	return user, nil
}

func (s *MySQLStore) RevokeByAccess(ctx context.Context, accessHash string, now time.Time) error {
	if s == nil || s.db == nil || accessHash == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at = COALESCE(revoked_at, ?), updated_at = CURRENT_TIMESTAMP(6) WHERE access_token_hash = ?`, now, accessHash)
	if err != nil {
		return fmt.Errorf("revoke access session: %w", err)
	}
	return nil
}

func (s *MySQLStore) RevokeByRefresh(ctx context.Context, refreshHash string, now time.Time) error {
	if s == nil || s.db == nil || refreshHash == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at = COALESCE(revoked_at, ?), updated_at = CURRENT_TIMESTAMP(6) WHERE refresh_token_hash = ?`, now, refreshHash)
	if err != nil {
		return fmt.Errorf("revoke refresh session: %w", err)
	}
	return nil
}

type rowQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadCapabilities(ctx context.Context, queryer rowQuerier, userID int64) ([]string, error) {
	rows, err := queryer.QueryContext(ctx, `SELECT capability FROM auth_user_capabilities WHERE user_id = ? ORDER BY capability`, userID)
	if err != nil {
		return nil, fmt.Errorf("load user capabilities: %w", err)
	}
	defer rows.Close()
	capabilities := make([]string, 0)
	for rows.Next() {
		var capability string
		if err := rows.Scan(&capability); err != nil {
			return nil, fmt.Errorf("scan user capability: %w", err)
		}
		capabilities = append(capabilities, capability)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user capabilities: %w", err)
	}
	return capabilities, nil
}

func loadActiveUserByID(ctx context.Context, queryer rowQuerier, userID int64) (User, error) {
	var user User
	err := queryer.QueryRowContext(ctx, `SELECT id, username, display_name, role, COALESCE(team_id, 0) FROM auth_users WHERE id = ? AND active = TRUE LIMIT 1`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.Role,
		&user.TeamID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthenticated
	}
	if err != nil {
		return User{}, fmt.Errorf("load active user: %w", err)
	}
	capabilities, err := loadCapabilities(ctx, queryer, user.ID)
	if err != nil {
		return User{}, err
	}
	user.Capabilities = capabilities
	return user, nil
}
