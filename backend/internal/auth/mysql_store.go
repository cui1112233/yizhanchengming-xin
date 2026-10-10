package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func (s *MySQLStore) available() error {
	if s == nil || s.db == nil {
		return errors.New("auth store unavailable")
	}
	return nil
}

func validateUsername(username string) error {
	if username == "" || len(username) > 32 {
		return ErrInvalid
	}
	return nil
}

func (s *MySQLStore) FindAccount(ctx context.Context, username string) (Account, error) {
	if err := s.available(); err != nil {
		return Account{}, err
	}
	if err := validateUsername(username); err != nil {
		return Account{}, err
	}
	var account Account
	err := s.db.QueryRowContext(ctx, `SELECT username, password_hash, password_algorithm, is_owner, active, created_at, updated_at FROM accounts WHERE username = ?`, username).Scan(&account.Username, &account.PasswordHash, &account.PasswordAlgorithm, &account.IsOwner, &account.Active, &account.CreatedAt, &account.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func (s *MySQLStore) CreateAccount(ctx context.Context, account Account) error {
	if err := s.available(); err != nil {
		return err
	}
	if err := validateUsername(account.Username); err != nil || account.PasswordHash == "" || account.PasswordAlgorithm == "" {
		return ErrInvalid
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO accounts(username, password_hash, password_algorithm, is_owner, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, account.Username, account.PasswordHash, account.PasswordAlgorithm, account.IsOwner, account.Active, account.CreatedAt, account.UpdatedAt)
	return err
}

func (s *MySQLStore) UpdatePassword(ctx context.Context, username, hash, algorithm string) error {
	if err := s.available(); err != nil {
		return err
	}
	if err := validateUsername(username); err != nil || hash == "" || algorithm == "" {
		return ErrInvalid
	}
	_, err := s.db.ExecContext(ctx, `UPDATE accounts SET password_hash = ?, password_algorithm = ?, updated_at = ? WHERE username = ?`, hash, algorithm, time.Now().UTC(), username)
	return err
}

func (s *MySQLStore) SetActive(ctx context.Context, username string, active bool) error {
	if err := s.available(); err != nil {
		return err
	}
	if err := validateUsername(username); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `UPDATE accounts SET active = ?, updated_at = ? WHERE username = ?`, active, time.Now().UTC(), username)
	return err
}

func (s *MySQLStore) CreateSession(ctx context.Context, session Session) error {
	if err := s.available(); err != nil {
		return err
	}
	if session.Token == "" || validateUsername(session.Username) != nil || !session.ExpiresAt.After(session.IssuedAt) {
		return ErrInvalid
	}
	result, err := s.db.ExecContext(ctx, `INSERT INTO auth_sessions(token_hash, username, expires_at, issued_at, browser, os, ip_hint) SELECT ?, ?, ?, ?, ?, ?, ? FROM accounts WHERE username = ? AND active = TRUE`, sessionTokenHash(session.Token), session.Username, session.ExpiresAt, session.IssuedAt, nullable(session.Browser), nullable(session.OS), nullable(session.IPHint), session.Username)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (s *MySQLStore) FindSession(ctx context.Context, rawToken string, now time.Time) (Session, error) {
	if err := s.available(); err != nil {
		return Session{}, err
	}
	if rawToken == "" {
		return Session{}, ErrNotFound
	}
	var session Session
	var browser, os, ipHint sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT s.token_hash, s.username, s.expires_at, s.issued_at, s.browser, s.os, s.ip_hint, a.active FROM auth_sessions s JOIN accounts a ON a.username = s.username WHERE s.token_hash = ? AND s.expires_at > ? AND a.active = TRUE`, sessionTokenHash(rawToken), now).Scan(new(string), &session.Username, &session.ExpiresAt, &session.IssuedAt, &browser, &os, &ipHint, new(bool))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, err
	}
	session.Token = rawToken
	session.Browser, session.OS, session.IPHint = browser.String, os.String, ipHint.String
	return session, nil
}

func (s *MySQLStore) CleanupExpiredSessions(ctx context.Context, now time.Time) error {
	if err := s.available(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE expires_at <= ?`, now)
	return err
}

func (s *MySQLStore) RevokeSession(ctx context.Context, rawToken string) error {
	if err := s.available(); err != nil {
		return err
	}
	if rawToken == "" {
		return ErrInvalid
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE token_hash = ?`, sessionTokenHash(rawToken))
	return err
}

// RevokeOtherSessions removes every login session for username except the
// current token. It is used after a password change so other devices are
// forced to sign in again.
func (s *MySQLStore) RevokeOtherSessions(ctx context.Context, username, currentToken string) (int64, error) {
	if err := s.available(); err != nil {
		return 0, err
	}
	if err := validateUsername(username); err != nil {
		return 0, err
	}
	if strings.TrimSpace(currentToken) == "" {
		result, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE username = ?`, username)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM auth_sessions WHERE username = ? AND token_hash <> ?`, username, sessionTokenHash(currentToken))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ListSessionsForUser returns the persisted login sessions. Token contains the
// stored hash rather than the raw bearer token and is only used for display
// and current-session comparison.
func (s *MySQLStore) ListSessionsForUser(ctx context.Context, username string) ([]Session, error) {
	if err := s.available(); err != nil {
		return nil, err
	}
	if err := validateUsername(username); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT token_hash, expires_at, issued_at, browser, os, ip_hint FROM auth_sessions WHERE username = ? AND expires_at > ? ORDER BY issued_at DESC`, username, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Session, 0)
	for rows.Next() {
		var session Session
		var browser, osName, ipHint sql.NullString
		if err := rows.Scan(&session.Token, &session.ExpiresAt, &session.IssuedAt, &browser, &osName, &ipHint); err != nil {
			return nil, err
		}
		session.Username = username
		session.Browser, session.OS, session.IPHint = browser.String, osName.String, ipHint.String
		result = append(result, session)
	}
	return result, rows.Err()
}
