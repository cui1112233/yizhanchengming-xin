package storage

import (
	"strings"
	"testing"
)

func TestAutomationMigrationsAppendUnifiedExecutionKernel(t *testing.T) {
	got := AutomationMigrations()
	if len(got) == 0 || got[len(got)-1].Version != 8800035 {
		t.Fatalf("last automation migration must be 8800035, got %v", got)
	}
	sql := strings.Join(got[len(got)-1].SQL, "\n")
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS batch_executions (",
		"CREATE TABLE IF NOT EXISTS batch_execution_config_snapshots (",
		"CREATE TABLE IF NOT EXISTS batch_execution_books (",
		"CREATE TABLE IF NOT EXISTS batch_execution_stage_runs (",
		"CREATE TABLE IF NOT EXISTS batch_execution_stage_attempts (",
		"CREATE TABLE IF NOT EXISTS batch_execution_publication_receipts (",
		"CREATE TABLE IF NOT EXISTS batch_execution_outbox (",
		"UNIQUE KEY uq_batch_execution_request (owner_username, batch_id, client_request_id)",
		"UNIQUE KEY uq_batch_execution_book (execution_id, book_id)",
		"UNIQUE KEY uq_batch_execution_stage_fingerprint (execution_book_id, stage_key, input_fingerprint)",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	registered := 0
	for _, migration := range AppMigrations() {
		if migration.Version == 8800035 {
			registered++
		}
	}
	if registered != 1 {
		t.Fatalf("migration 8800035 registered %d times, want once", registered)
	}
}
