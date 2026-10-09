package task9runtime

import (
	"os"
	"strings"
	"testing"
)

func TestGenerationMigrationPreservesLegacyAndRejectsLossyDown(t *testing.T) {
	b, err := os.ReadFile("../../db/migrations/00021_generation_runtime.sql")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(strings.ToLower(string(b)), "-- +goose down")
	if len(parts) != 2 {
		t.Fatal("missing explicit down boundary")
	}
	for _, field := range []string{"run_kind", "default 'legacy'", "target_book_id", "request_schema_version", "request_snapshot", "request_hash", "requested_by_user_id", "cancel_requested_at", "cancelled_at", "(run_kind, status, run_at, id)"} {
		if !strings.Contains(parts[0], field) {
			t.Fatalf("missing generation metadata %s", field)
		}
	}
	if !strings.Contains(parts[1], "signal sqlstate '45000'") || strings.Contains(parts[1], "drop ") {
		t.Fatal("down must reject metadata loss before any drop")
	}
	if strings.Contains(parts[0], "create table") {
		t.Fatal("generation must reuse existing facts")
	}
}
