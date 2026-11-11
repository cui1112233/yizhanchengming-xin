package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func AuthMigrations() []Migration {
	return []Migration{{
		Version:          7802004,
		SQL:              AuthSchemaStatements(),
		CallbackChecksum: "go-auth-mysql-schema-v1",
		// The public database already contains this verified auth schema from
		// the first Go-compatible release, which used the SQL-era checksum.
		LegacyChecksums: []string{"bb76069aedb05e6ee9ff9e887edda99b349fdecc5bc2fbdc56400b94cb8375e1"},
		Adopt:           adoptAuthSchemaLedger,
	}}
}

func AuthSchemaStatements() []string {
	return []string{
		`CREATE TABLE accounts (
			username VARCHAR(32) PRIMARY KEY,
			password_hash TEXT NOT NULL,
			password_algorithm VARCHAR(32) NOT NULL,
			is_owner BOOLEAN NOT NULL,
			active BOOLEAN NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL
		) ENGINE=InnoDB`,
		`CREATE TABLE auth_sessions (
			token_hash CHAR(64) PRIMARY KEY,
			username VARCHAR(32) NOT NULL,
			expires_at DATETIME(6) NOT NULL,
			issued_at DATETIME(6) NOT NULL,
			browser VARCHAR(80) NULL,
			os VARCHAR(80) NULL,
			ip_hint VARCHAR(64) NULL,
			INDEX auth_sessions_username_expires(username, expires_at),
			FOREIGN KEY (username) REFERENCES accounts(username)
		) ENGINE=InnoDB`,
		`CREATE TABLE auth_migrations (
			migration_key VARCHAR(128) PRIMARY KEY,
			completed_at DATETIME(6) NOT NULL,
			source_summary JSON NOT NULL
		) ENGINE=InnoDB`,
	}
}

func adoptAuthSchemaLedger(ctx context.Context, tx *sql.Tx) (bool, error) {
	var accountsExists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = DATABASE() AND table_name = 'accounts'
)`).Scan(&accountsExists); err != nil || !accountsExists {
		return accountsExists, err
	}
	if err := validateAuthTable(ctx, tx, "accounts", []string{"username", "password_hash", "password_algorithm", "is_owner", "active", "created_at", "updated_at"}, "username", ""); err != nil {
		return false, err
	}
	existing := make(map[string]bool, 2)
	for _, table := range []string{"auth_sessions", "auth_migrations"} {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)`, table).Scan(&exists); err != nil {
			return false, err
		}
		existing[table] = exists
	}
	for _, table := range []string{"auth_sessions", "auth_migrations"} {
		if existing[table] {
			if err := validateAuthTable(ctx, tx, table, authRequiredColumns(table), authPrimaryColumn(table), authForeignKeyTarget(table)); err != nil {
				return false, err
			}
		}
	}
	for i, statement := range AuthSchemaStatements()[1:] {
		table := []string{"auth_sessions", "auth_migrations"}[i]
		if existing[table] {
			continue
		}
		statement = strings.Replace(statement, "CREATE TABLE ", "CREATE TABLE IF NOT EXISTS ", 1)
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return false, err
		}
	}
	return true, nil
}

func authRequiredColumns(table string) []string {
	switch table {
	case "accounts":
		return []string{"username", "password_hash", "password_algorithm", "is_owner", "active", "created_at", "updated_at"}
	case "auth_sessions":
		return []string{"token_hash", "username", "expires_at", "issued_at", "browser", "os", "ip_hint"}
	default:
		return []string{"migration_key", "completed_at", "source_summary"}
	}
}

func authPrimaryColumn(table string) string {
	if table == "accounts" {
		return "username"
	}
	if table == "auth_sessions" {
		return "token_hash"
	}
	return "migration_key"
}

func authForeignKeyTarget(table string) string {
	if table == "auth_sessions" {
		return "accounts(username)"
	}
	return ""
}

type authColumnDefinition struct {
	dataType  string
	length    int64
	nullable  string
	precision int64
}

func validateAuthTable(ctx context.Context, tx *sql.Tx, table string, requiredColumns []string, primaryColumn, foreignKeyTarget string) error {
	args := []any{table}
	for _, column := range requiredColumns {
		args = append(args, column)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(requiredColumns)), ",")
	query := `SELECT column_name, data_type, column_type, is_nullable, character_maximum_length, datetime_precision FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name IN (` + placeholders + `)`
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	expected := authColumnDefinitions(table)
	found := make(map[string]bool, len(expected))
	for rows.Next() {
		var name, dataType, columnType, nullable string
		var length, precision sql.NullInt64
		if err := rows.Scan(&name, &dataType, &columnType, &nullable, &length, &precision); err != nil {
			return err
		}
		definition, ok := expected[name]
		if !ok || strings.ToLower(dataType) != definition.dataType || nullable != definition.nullable || (definition.length > 0 && (!length.Valid || length.Int64 != definition.length)) || (definition.precision > 0 && (!precision.Valid || precision.Int64 != definition.precision)) {
			return fmt.Errorf("incompatible %s schema: column %s definition differs", table, name)
		}
		found[name] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(found) != len(expected) {
		return fmt.Errorf("incompatible %s schema: required columns missing", table)
	}
	primaryRows, err := tx.QueryContext(ctx, `SELECT column_name FROM information_schema.key_column_usage WHERE constraint_schema = DATABASE() AND table_name = ? AND constraint_name = 'PRIMARY' ORDER BY ordinal_position`, table)
	if err != nil {
		return err
	}
	defer primaryRows.Close()
	var primaryColumns []string
	for primaryRows.Next() {
		var column string
		if err := primaryRows.Scan(&column); err != nil {
			return err
		}
		primaryColumns = append(primaryColumns, column)
	}
	if err := primaryRows.Err(); err != nil {
		return err
	}
	if len(primaryColumns) != 1 || primaryColumns[0] != primaryColumn {
		return fmt.Errorf("incompatible %s schema: primary key differs", table)
	}
	if foreignKeyTarget != "" {
		foreignRows, err := tx.QueryContext(ctx, `SELECT column_name, referenced_table_name, referenced_column_name FROM information_schema.key_column_usage WHERE constraint_schema = DATABASE() AND table_name = ? AND referenced_table_name IS NOT NULL ORDER BY constraint_name, ordinal_position`, table)
		if err != nil {
			return err
		}
		defer foreignRows.Close()
		var foreignKeys [][3]string
		for foreignRows.Next() {
			var foreignKey [3]string
			if err := foreignRows.Scan(&foreignKey[0], &foreignKey[1], &foreignKey[2]); err != nil {
				return err
			}
			foreignKeys = append(foreignKeys, foreignKey)
		}
		if err := foreignRows.Err(); err != nil {
			return err
		}
		if len(foreignKeys) != 1 || foreignKeys[0] != [3]string{"username", "accounts", "username"} {
			return fmt.Errorf("incompatible auth_sessions schema: foreign key differs")
		}
	}
	return nil
}

func authColumnDefinitions(table string) map[string]authColumnDefinition {
	datetime := authColumnDefinition{dataType: "datetime", nullable: "NO", precision: 6}
	switch table {
	case "accounts":
		return map[string]authColumnDefinition{"username": {"varchar", 32, "NO", 0}, "password_hash": {"text", 0, "NO", 0}, "password_algorithm": {"varchar", 32, "NO", 0}, "is_owner": {"tinyint", 0, "NO", 0}, "active": {"tinyint", 0, "NO", 0}, "created_at": datetime, "updated_at": datetime}
	case "auth_sessions":
		return map[string]authColumnDefinition{"token_hash": {"char", 64, "NO", 0}, "username": {"varchar", 32, "NO", 0}, "expires_at": datetime, "issued_at": datetime, "browser": {"varchar", 80, "YES", 0}, "os": {"varchar", 80, "YES", 0}, "ip_hint": {"varchar", 64, "YES", 0}}
	default:
		return map[string]authColumnDefinition{"migration_key": {"varchar", 128, "NO", 0}, "completed_at": datetime, "source_summary": {"json", 0, "NO", 0}}
	}
}
