package auth

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockStore(t *testing.T) (*MySQLStore, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewMySQLStore(db), mock
}

func TestCreateSessionRejectsInactiveOrMissingAccount(t *testing.T) {
	store, mock := newMockStore(t)
	now := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	session := Session{Token: "raw-secret-token", Username: "alice", IssuedAt: now, ExpiresAt: now.Add(time.Hour)}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_sessions(token_hash, username, expires_at, issued_at, browser, os, ip_hint) SELECT ?, ?, ?, ?, ?, ?, ? FROM accounts WHERE username = ? AND active = TRUE")).WithArgs(sessionTokenHash(session.Token), session.Username, session.ExpiresAt, session.IssuedAt, nil, nil, nil, session.Username).WillReturnResult(sqlmock.NewResult(0, 0))

	if err := store.CreateSession(context.Background(), session); err != ErrNotFound {
		t.Fatalf("CreateSession() error = %v, want %v", err, ErrNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSessionLookupUsesHashAndRejectsExpiryAtNowAndInactiveAccount(t *testing.T) {
	store, mock := newMockStore(t)
	now := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	rawToken := "raw-secret-token"
	hash := sessionTokenHash(rawToken)
	if hash == rawToken {
		t.Fatal("session token must be hashed")
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT s.token_hash, s.username, s.expires_at, s.issued_at, s.browser, s.os, s.ip_hint, a.active FROM auth_sessions s JOIN accounts a ON a.username = s.username WHERE s.token_hash = ? AND s.expires_at > ? AND a.active = TRUE")).WithArgs(hash, now).WillReturnError(sql.ErrNoRows)

	if _, err := store.FindSession(context.Background(), rawToken, now); err != ErrNotFound {
		t.Fatalf("FindSession() error = %v, want %v", err, ErrNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindSessionReturnsValidActiveSession(t *testing.T) {
	store, mock := newMockStore(t)
	now := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	rawToken := "raw-secret-token"
	hash := sessionTokenHash(rawToken)
	expiresAt := now.Add(time.Hour)
	issuedAt := now.Add(-time.Minute)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT s.token_hash, s.username, s.expires_at, s.issued_at, s.browser, s.os, s.ip_hint, a.active FROM auth_sessions s JOIN accounts a ON a.username = s.username WHERE s.token_hash = ? AND s.expires_at > ? AND a.active = TRUE")).WithArgs(hash, now).WillReturnRows(sqlmock.NewRows([]string{"token_hash", "username", "expires_at", "issued_at", "browser", "os", "ip_hint", "active"}).AddRow(hash, "alice", expiresAt, issuedAt, "Chrome", "Windows", "127.0.0.1", true))

	session, err := store.FindSession(context.Background(), rawToken, now)
	if err != nil {
		t.Fatal(err)
	}
	if session.Token != rawToken || session.Username != "alice" || !session.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("FindSession() = %+v", session)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRevokeSessionUsesTokenHash(t *testing.T) {
	store, mock := newMockStore(t)
	rawToken := "raw-secret-token"
	hash := sessionTokenHash(rawToken)
	if hash == rawToken {
		t.Fatal("session token must be hashed")
	}
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM auth_sessions WHERE token_hash = ?")).WithArgs(hash).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.RevokeSession(context.Background(), rawToken); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupExpiredSessionsOnlyDeletesExpired(t *testing.T) {
	store, mock := newMockStore(t)
	now := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM auth_sessions WHERE expires_at <= ?")).WithArgs(now).WillReturnResult(sqlmock.NewResult(0, 2))

	if err := store.CleanupExpiredSessions(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPurgeAccountDeletesSessionsAndNonOwner(t *testing.T) {
	store, mock := newMockStore(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM auth_sessions WHERE username = ?")).WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM accounts WHERE username = ? AND is_owner = FALSE")).WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	removed, err := store.PurgeAccount(context.Background(), "alice")
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
