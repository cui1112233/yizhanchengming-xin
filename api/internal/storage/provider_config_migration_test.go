package storage

import (
	"os"
	"strings"
	"testing"
)

func TestProviderConfigMigrationIsSharedAndEncrypted(t *testing.T) {
	body, err := os.ReadFile("../../migrations/00005_generation_provider_configs.sql")
	if err != nil { t.Fatal(err) }
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS generation_provider_configs",
		"owner",
		"media_kind",
		"provider",
		"model",
		"create_url",
		"tasks_url",
		"result_url",
		"credential_nonce",
		"credential_ciphertext",
		"settings_json",
		"UNIQUE KEY uk_generation_provider_owner_kind_provider",
	} {
		if !strings.Contains(sql, required) { t.Fatalf("migration missing %s", required) }
	}
	for _, forbidden := range []string{"api_key VARCHAR", "api_key TEXT", "secret_key VARCHAR", "secret_key TEXT"} {
		if strings.Contains(strings.ToLower(sql), strings.ToLower(forbidden)) { t.Fatalf("plaintext credential column forbidden: %s", forbidden) }
	}
}
