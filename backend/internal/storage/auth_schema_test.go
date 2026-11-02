package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAuthSchemaMigrationIsAdditiveAndComplete(t *testing.T) {
	migrations := AppMigrations()
	var auth *Migration
	maxVersion := 0
	for i := range migrations {
		if migrations[i].CallbackChecksum == "go-auth-mysql-schema-v1" {
			auth = &migrations[i]
			break
		}
		if migrations[i].Version > maxVersion {
			maxVersion = migrations[i].Version
		}
	}
	if auth == nil {
		t.Fatal("auth schema migration is missing")
	}
	if auth.Version <= 7802002 || auth.Version <= maxVersion {
		t.Fatalf("auth migration version=%d must be newer than existing version=%d", auth.Version, maxVersion)
	}
	joined := strings.ToLower(strings.Join(auth.SQL, "\n"))
	for _, required := range []string{
		"create table accounts", "username varchar(32) primary key", "create table auth_sessions",
		"token_hash char(64) primary key", "auth_sessions_username_expires",
		"foreign key (username) references accounts(username)", "create table auth_migrations", "source_summary json not null",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("auth schema migration misses %q: %s", required, joined)
		}
	}
	if auth.Adopt == nil {
		t.Fatal("auth migration must adopt a known partial DDL application")
	}
}

func TestAuthMigrationDoesNotAdoptIncompatibleSchemas(t *testing.T) {
	cases := []struct {
		name    string
		table   string
		mutate  func([]authColumn) []authColumn
		primary []string
		foreign authForeignKey
	}{
		{
			name: "wrong accounts column type", table: "accounts",
			mutate: func(columns []authColumn) []authColumn {
				return replaceAuthColumn(columns, "password_hash", authColumn{"password_hash", "varchar", "varchar(255)", "NO", 255, nil})
			},
			primary: []string{"username"},
		},
		{
			name: "composite accounts primary key", table: "accounts",
			mutate:  func(columns []authColumn) []authColumn { return columns },
			primary: []string{"username", "active"},
		},
		{
			name: "wrong auth sessions foreign key target", table: "auth_sessions",
			mutate:  func(columns []authColumn) []authColumn { return columns },
			primary: []string{"token_hash"}, foreign: authForeignKey{"username", "other_accounts", "username"},
		},
		{
			name: "wrong auth sessions foreign key column", table: "auth_sessions",
			mutate:  func(columns []authColumn) []authColumn { return columns },
			primary: []string{"token_hash"}, foreign: authForeignKey{"token_hash", "accounts", "username"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newAuthMock(t)
			defer db.Close()
			migration := AuthMigrations()[0]
			expectAuthMigrationStart(mock, migration)
			expectAuthTableExists(mock, "accounts", true)
			accounts := compatibleAuthColumns("accounts")
			accountsPrimary := []string{"username"}
			if tc.table == "accounts" {
				accounts = tc.mutate(accounts)
				accountsPrimary = tc.primary
				if tc.name == "wrong accounts column type" {
					accountsPrimary = nil
				}
			}
			expectAuthTableDefinition(mock, "accounts", accounts, accountsPrimary, authForeignKey{})
			if tc.table == "accounts" {
				mock.ExpectRollback()
			} else {
				expectAuthTableExists(mock, "auth_sessions", true)
				expectAuthTableExists(mock, "auth_migrations", true)
				sessions := compatibleAuthColumns("auth_sessions")
				if tc.table == "auth_sessions" {
					sessions = tc.mutate(sessions)
				}
				expectAuthTableDefinition(mock, "auth_sessions", sessions, tc.primary, tc.foreign)
				mock.ExpectRollback()
			}

			err := RunMigrations(context.Background(), db, []Migration{migration})
			if err == nil || !strings.Contains(err.Error(), tc.table) {
				t.Fatalf("incompatible %s must reject adoption, err=%v", tc.table, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAuthMigrationCreatesMissingAuthTablesAndRecordsLedger(t *testing.T) {
	db, mock := newAuthMock(t)
	defer db.Close()
	migration := AuthMigrations()[0]
	expectAuthMigrationStart(mock, migration)
	expectAuthTableExists(mock, "accounts", true)
	expectAuthTableDefinition(mock, "accounts", compatibleAuthColumns("accounts"), []string{"username"}, authForeignKey{})
	expectAuthTableExists(mock, "auth_sessions", false)
	expectAuthTableExists(mock, "auth_migrations", false)
	for _, statement := range AuthSchemaStatements()[1:] {
		mock.ExpectExec(regexp.QuoteMeta(strings.Replace(statement, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", 1))).WillReturnResult(sqlmock.NewResult(0, 0))
	}
	expectAuthLedgerAndCommit(mock, migration)

	if err := RunMigrations(context.Background(), db, []Migration{migration}); err != nil {
		t.Fatalf("compatible accounts should create missing auth tables: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAuthMigrationRecordsLedgerForCompatibleSchema(t *testing.T) {
	db, mock := newAuthMock(t)
	defer db.Close()
	migration := AuthMigrations()[0]
	expectAuthMigrationStart(mock, migration)
	expectAuthTableExists(mock, "accounts", true)
	expectAuthTableDefinition(mock, "accounts", compatibleAuthColumns("accounts"), []string{"username"}, authForeignKey{})
	for _, table := range []string{"auth_sessions", "auth_migrations"} {
		expectAuthTableExists(mock, table, true)
	}
	expectAuthTableDefinition(mock, "auth_sessions", compatibleAuthColumns("auth_sessions"), []string{"token_hash"}, authForeignKey{"username", "accounts", "username"})
	expectAuthTableDefinition(mock, "auth_migrations", compatibleAuthColumns("auth_migrations"), []string{"migration_key"}, authForeignKey{})
	expectAuthLedgerAndCommit(mock, migration)

	if err := RunMigrations(context.Background(), db, []Migration{migration}); err != nil {
		t.Fatalf("compatible auth schema should adopt: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type authColumn struct {
	name, dataType, columnType, nullable string
	length                               int64
	precision                            *int64
}
type authForeignKey struct{ column, referencedTable, referencedColumn string }

func newAuthMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return db, mock
}
func expectAuthMigrationStart(mock sqlmock.Sqlmock, migration Migration) {
	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT NOT NULL PRIMARY KEY, checksum CHAR(64) NOT NULL, applied_at DATETIME(6) NOT NULL) ENGINE=InnoDB`)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT checksum FROM schema_migrations WHERE version = ?`)).WithArgs(migration.Version).WillReturnError(sql.ErrNoRows)
	mock.ExpectBegin()
}
func expectAuthTableExists(mock sqlmock.Sqlmock, table string, exists bool) {
	if table == "accounts" {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = DATABASE() AND table_name = 'accounts'
)`)).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
		return
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)`)).WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
}
func expectAuthTableDefinition(mock sqlmock.Sqlmock, table string, columns []authColumn, primary []string, foreign authForeignKey) {
	required := authRequiredColumns(table)
	args := []driver.Value{table}
	for _, column := range required {
		args = append(args, column)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(required)), ",")
	rows := sqlmock.NewRows([]string{"column_name", "data_type", "column_type", "is_nullable", "character_maximum_length", "datetime_precision"})
	for _, column := range columns {
		var precision any
		if column.precision != nil {
			precision = *column.precision
		}
		rows.AddRow(column.name, column.dataType, column.columnType, column.nullable, nullableLength(column.length), precision)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT column_name, data_type, column_type, is_nullable, character_maximum_length, datetime_precision FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name IN (` + placeholders + `)`)).WithArgs(args...).WillReturnRows(rows)
	if primary != nil {
		primaryRows := sqlmock.NewRows([]string{"column_name"})
		for _, column := range primary {
			primaryRows.AddRow(column)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT column_name FROM information_schema.key_column_usage WHERE constraint_schema = DATABASE() AND table_name = ? AND constraint_name = 'PRIMARY' ORDER BY ordinal_position`)).WithArgs(table).WillReturnRows(primaryRows)
	}
	if table == "auth_sessions" && primary != nil {
		foreignRows := sqlmock.NewRows([]string{"column_name", "referenced_table_name", "referenced_column_name"})
		if foreign.column != "" {
			foreignRows.AddRow(foreign.column, foreign.referencedTable, foreign.referencedColumn)
		}
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT column_name, referenced_table_name, referenced_column_name FROM information_schema.key_column_usage WHERE constraint_schema = DATABASE() AND table_name = ? AND referenced_table_name IS NOT NULL ORDER BY constraint_name, ordinal_position`)).WithArgs(table).WillReturnRows(foreignRows)
	}
}
func expectAuthLedgerAndCommit(mock sqlmock.Sqlmock, migration Migration) {
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO schema_migrations(version, checksum, applied_at) VALUES (?, ?, CURRENT_TIMESTAMP(6))`)).WithArgs(migration.Version, ChecksumFor(migration)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}
func compatibleAuthColumns(table string) []authColumn {
	six := int64(6)
	switch table {
	case "accounts":
		return []authColumn{{"username", "varchar", "varchar(32)", "NO", 32, nil}, {"password_hash", "text", "text", "NO", 0, nil}, {"password_algorithm", "varchar", "varchar(32)", "NO", 32, nil}, {"is_owner", "tinyint", "tinyint(1)", "NO", 0, nil}, {"active", "tinyint", "tinyint(1)", "NO", 0, nil}, {"created_at", "datetime", "datetime(6)", "NO", 0, &six}, {"updated_at", "datetime", "datetime(6)", "NO", 0, &six}}
	case "auth_sessions":
		return []authColumn{{"token_hash", "char", "char(64)", "NO", 64, nil}, {"username", "varchar", "varchar(32)", "NO", 32, nil}, {"expires_at", "datetime", "datetime(6)", "NO", 0, &six}, {"issued_at", "datetime", "datetime(6)", "NO", 0, &six}, {"browser", "varchar", "varchar(80)", "YES", 80, nil}, {"os", "varchar", "varchar(80)", "YES", 80, nil}, {"ip_hint", "varchar", "varchar(64)", "YES", 64, nil}}
	default:
		return []authColumn{{"migration_key", "varchar", "varchar(128)", "NO", 128, nil}, {"completed_at", "datetime", "datetime(6)", "NO", 0, &six}, {"source_summary", "json", "json", "NO", 0, nil}}
	}
}
func replaceAuthColumn(columns []authColumn, name string, replacement authColumn) []authColumn {
	result := append([]authColumn(nil), columns...)
	for i := range result {
		if result[i].name == name {
			result[i] = replacement
		}
	}
	return result
}
func nullableLength(length int64) any {
	if length == 0 {
		return nil
	}
	return length
}
