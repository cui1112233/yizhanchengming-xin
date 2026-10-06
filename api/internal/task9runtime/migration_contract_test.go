package task9runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTask9MigrationIs00008AndCarriesDurableRuntimeFacts(t *testing.T) {
	path := filepath.Join("..", "..", "db", "migrations", "00008_task9_runtime.sql")
	b, err := os.ReadFile(path)
	if err != nil { t.Fatalf("read %s: %v", path, err) }
	sql := strings.ToLower(string(b))
	for _, required := range []string{
		"idempotency_key", "unique", "run_id", "attempt", "execution_token", "lease_deadline", "running_since", "max_attempts", "error_code", "error_message",
	} {
		if !strings.Contains(sql, required) { t.Fatalf("migration missing %q", required) }
	}
	if strings.Contains(sql, "runtime_book_runs") || strings.Contains(sql, "runtime_jobs") { t.Fatal("migration created forbidden second fact model") }
}
