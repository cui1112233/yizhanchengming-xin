package intake

import (
	"os"
	"strings"
	"testing"
)

func TestBatchProjectArchiveMigrationPreservesFactsAndIsIrreversible(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00020_batch_project_archive.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"ADD COLUMN archived_at DATETIME(6) NULL",
		"ADD COLUMN archived_by_user_id BIGINT UNSIGNED NULL",
		"FOREIGN KEY (archived_by_user_id) REFERENCES auth_users(id) ON DELETE SET NULL",
		"KEY idx_batch_projects_archive_updated (archived_at, updated_at, id)",
		"KEY idx_batch_projects_archived_by (archived_by_user_id)",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	downParts := strings.Split(sql, "-- +goose Down")
	if len(downParts) != 2 {
		t.Fatalf("migration must contain exactly one Goose Down section")
	}
	down := downParts[1]
	if strings.Contains(strings.ToUpper(down), "DROP ") {
		t.Fatal("irreversible Down must preserve archived business facts")
	}
	for _, required := range []string{
		"-- +goose StatementBegin",
		"SIGNAL SQLSTATE '45000'",
		"batch project archive migration is irreversible",
		"-- +goose StatementEnd",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("irreversible Down missing %q", required)
		}
	}
}
