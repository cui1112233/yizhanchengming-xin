package authn

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreFindLoginCredentialLoadsCapabilities(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, display_name, password_hash, role, COALESCE(team_id, 0), active FROM auth_users WHERE username = ? LIMIT 1")).
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "display_name", "password_hash", "role", "team_id", "active"}).AddRow(int64(7), "alice", "Alice", "$2a$10$hash", "member", int64(3), true))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT capability FROM auth_user_capabilities WHERE user_id = ? ORDER BY capability")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"capability"}).AddRow("batch.read").AddRow("settings.edit"))

	credential, err := store.FindLoginCredential(context.Background(), "alice")
	if err != nil { t.Fatal(err) }
	if credential.User.ID != 7 || credential.User.TeamID != 3 || credential.PasswordHash == "" || !credential.Active {
		t.Fatalf("credential=%#v", credential)
	}
	if len(credential.User.Capabilities) != 2 || credential.User.Capabilities[0] != "batch.read" {
		t.Fatalf("capabilities=%v", credential.User.Capabilities)
	}
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreCreateSessionPersistsOnlyHashes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	record := SessionRecord{UserID: 7, AccessTokenHash: "access-hash", RefreshTokenHash: "refresh-hash", AccessExpiresAt: now.Add(15*time.Minute), RefreshExpiresAt: now.Add(30*24*time.Hour)}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(record.UserID, record.AccessTokenHash, record.RefreshTokenHash, record.AccessExpiresAt, record.RefreshExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.CreateSession(context.Background(), record); err != nil { t.Fatal(err) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreResolveAccessRestoresUserAfterProcessRestart(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.id, u.username, u.display_name, u.role, COALESCE(u.team_id, 0) FROM auth_sessions s JOIN auth_users u ON u.id = s.user_id WHERE s.access_token_hash = ? AND s.revoked_at IS NULL AND s.access_expires_at > ? AND u.active = TRUE LIMIT 1")).
		WithArgs("access-hash", now).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "display_name", "role", "team_id"}).AddRow(int64(7), "alice", "Alice", "member", int64(3)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT capability FROM auth_user_capabilities WHERE user_id = ? ORDER BY capability")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"capability"}).AddRow("batch.read"))

	user, err := store.ResolveAccess(context.Background(), "access-hash", now)
	if err != nil { t.Fatal(err) }
	if user.ID != 7 || user.Username != "alice" || len(user.Capabilities) != 1 { t.Fatalf("user=%#v", user) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLStoreRotateRefreshIsTransactional(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	next := SessionRecord{AccessTokenHash: "next-access", RefreshTokenHash: "next-refresh", AccessExpiresAt: now.Add(15*time.Minute), RefreshExpiresAt: now.Add(30*24*time.Hour)}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id FROM auth_sessions WHERE refresh_token_hash = ? AND revoked_at IS NULL AND refresh_expires_at > ? FOR UPDATE")).
		WithArgs("old-refresh", now).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(7)))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE auth_sessions SET revoked_at = ?, updated_at = CURRENT_TIMESTAMP(6) WHERE refresh_token_hash = ? AND revoked_at IS NULL")).
		WithArgs(now, "old-refresh").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_sessions (user_id, access_token_hash, refresh_token_hash, access_expires_at, refresh_expires_at) VALUES (?, ?, ?, ?, ?)")).
		WithArgs(int64(7), next.AccessTokenHash, next.RefreshTokenHash, next.AccessExpiresAt, next.RefreshExpiresAt).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, display_name, role, COALESCE(team_id, 0) FROM auth_users WHERE id = ? AND active = TRUE LIMIT 1")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "display_name", "role", "team_id"}).AddRow(int64(7), "alice", "Alice", "member", int64(3)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT capability FROM auth_user_capabilities WHERE user_id = ? ORDER BY capability")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"capability"}).AddRow("batch.read"))
	mock.ExpectCommit()

	user, err := store.RotateByRefresh(context.Background(), "old-refresh", next, now)
	if err != nil { t.Fatal(err) }
	if user.ID != 7 || user.Username != "alice" { t.Fatalf("user=%#v", user) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
