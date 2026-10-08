package authn

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnsureInitialAdminCreatesFirstUserFromDeploymentCredentials(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM auth_users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_users (username, display_name, password_hash, role, active) VALUES (?, ?, ?, 'owner', TRUE)")).
		WithArgs("bootstrap-admin", "bootstrap-admin", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	created, err := store.EnsureInitialAdmin(context.Background(), "bootstrap-admin", "a-strong-bootstrap-password")
	if err != nil { t.Fatalf("ensure initial admin: %v", err) }
	if !created { t.Fatal("expected first deployment to create the bootstrap admin") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestEnsureInitialAdminRefusesToStartEmptyDatabaseWithoutCredentials(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM auth_users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	created, err := store.EnsureInitialAdmin(context.Background(), "", "")
	if err == nil { t.Fatal("empty auth database must require deployment bootstrap credentials") }
	if created { t.Fatal("must not create a user without explicit credentials") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestEnsureInitialAdminDoesNotReapplyBootstrapCredentials(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM auth_users")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	created, err := store.EnsureInitialAdmin(context.Background(), "", "")
	if err != nil { t.Fatalf("existing installation should not require bootstrap env: %v", err) }
	if created { t.Fatal("bootstrap must be one-time only") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestAdminCredentialsCanBeChangedAndUserRevokedAfterBootstrap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE auth_users SET password_hash = ? WHERE username = ?")).
		WithArgs(sqlmock.AnyArg(), "bootstrap-admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.SetUserPassword(context.Background(), "bootstrap-admin", "new-strong-password"); err != nil {
		t.Fatalf("set password: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE auth_users SET active = ? WHERE username = ?")).
		WithArgs(false, "bootstrap-admin").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.SetUserActive(context.Background(), "bootstrap-admin", false); err != nil {
		t.Fatalf("disable bootstrap admin: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
