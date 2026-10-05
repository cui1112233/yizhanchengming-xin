package authn

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrBootstrapRequired = errors.New("auth bootstrap credentials are required for an empty database")

// EnsureInitialAdmin prevents a fresh deployment from starting with auth enabled
// but no login-capable user. Deployment credentials are only consumed when the
// auth_users table is empty; once any user exists they are ignored.
func (s *MySQLStore) EnsureInitialAdmin(ctx context.Context, username, password string) (bool, error) {
	if s == nil || s.db == nil {
		return false, errors.New("auth user store is unavailable")
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM auth_users`).Scan(&count); err != nil {
		return false, fmt.Errorf("count auth users: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	username = strings.TrimSpace(username)
	if username == "" || len(strings.TrimSpace(password)) < 12 {
		return false, ErrBootstrapRequired
	}
	hash, err := HashPassword(password)
	if err != nil {
		return false, fmt.Errorf("hash bootstrap admin password: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO auth_users (username, display_name, password_hash, role, active) VALUES (?, ?, ?, 'admin', TRUE)`, username, username, hash)
	if err != nil {
		// A second process may have won the first-start race. If a user now
		// exists, treat bootstrap as already completed rather than overwriting it.
		if qerr := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM auth_users`).Scan(&count); qerr == nil && count > 0 {
			return false, nil
		}
		return false, fmt.Errorf("create bootstrap admin: %w", err)
	}
	return true, nil
}

func (s *MySQLStore) SetUserPassword(ctx context.Context, username, password string) error {
	if s == nil || s.db == nil {
		return errors.New("auth user store is unavailable")
	}
	username = strings.TrimSpace(username)
	if username == "" || len(strings.TrimSpace(password)) < 12 {
		return errors.New("username and password of at least 12 characters are required")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `UPDATE auth_users SET password_hash = ? WHERE username = ?`, hash, username)
	if err != nil {
		return fmt.Errorf("update auth password: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verify auth password update: %w", err)
	}
	if rows != 1 {
		return errors.New("auth user not found")
	}
	return nil
}

func (s *MySQLStore) SetUserActive(ctx context.Context, username string, active bool) error {
	if s == nil || s.db == nil {
		return errors.New("auth user store is unavailable")
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("username is required")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE auth_users SET active = ? WHERE username = ?`, active, username)
	if err != nil {
		return fmt.Errorf("update auth user active state: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verify auth user active update: %w", err)
	}
	if rows != 1 {
		return errors.New("auth user not found")
	}
	return nil
}
