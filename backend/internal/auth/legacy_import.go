package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const legacyImportKey = "node-accounts-v1"

type LegacyAccountFile struct {
	Accounts []LegacyAccount `json:"accounts"`
}

type LegacyAccount struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	IsOwner      bool   `json:"isOwner"`
	Active       bool   `json:"active"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// ImportLegacyAccountsOnce reads the old account file only when the MySQL
// completion marker is absent. It never imports old sessions or tokens.
func ImportLegacyAccountsOnce(ctx context.Context, store *MySQLStore, sourcePath string) error {
	if store == nil || store.db == nil || strings.TrimSpace(sourcePath) == "" {
		return nil
	}
	var completed bool
	if err := store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM auth_migrations WHERE migration_key = ?)`, legacyImportKey).Scan(&completed); err != nil {
		return err
	}
	if completed {
		return nil
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read legacy account file: %w", err)
	}
	accounts, err := decodeLegacyAccounts(data)
	if err != nil {
		return err
	}
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, account := range accounts {
		if err := validateLegacyAccount(account); err != nil {
			return err
		}
		created, err := time.Parse(time.RFC3339Nano, account.CreatedAt)
		if err != nil {
			return fmt.Errorf("invalid legacy account time: %w", err)
		}
		updated, err := time.Parse(time.RFC3339Nano, account.UpdatedAt)
		if err != nil {
			return fmt.Errorf("invalid legacy account time: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO accounts(username, password_hash, password_algorithm, is_owner, active, created_at, updated_at) VALUES (?, ?, 'scrypt', ?, ?, ?, ?) ON DUPLICATE KEY UPDATE username = username`, account.Username, account.PasswordHash, account.IsOwner, account.Active, created.UTC(), updated.UTC()); err != nil {
			return err
		}
	}
	summary, _ := json.Marshal(map[string]any{"source": filepath.Base(sourcePath), "accounts": len(accounts)})
	if _, err := tx.ExecContext(ctx, `INSERT INTO auth_migrations(migration_key, completed_at, source_summary) VALUES (?, ?, ?)`, legacyImportKey, time.Now().UTC(), string(summary)); err != nil {
		return err
	}
	return tx.Commit()
}

func decodeLegacyAccounts(data []byte) ([]LegacyAccount, error) {
	var list []LegacyAccount
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil
	}
	var wrapper LegacyAccountFile
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("invalid legacy account file: %w", err)
	}
	return wrapper.Accounts, nil
}

func validateLegacyAccount(account LegacyAccount) error {
	if err := validateUsername(account.Username); err != nil || account.PasswordHash == "" || !strings.HasPrefix(account.PasswordHash, "scrypt$") {
		return ErrInvalid
	}
	if account.CreatedAt == "" || account.UpdatedAt == "" {
		return ErrInvalid
	}
	return nil
}
