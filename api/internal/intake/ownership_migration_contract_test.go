package intake

import (
	"os"
	"strings"
	"testing"
)

func TestIntakeOwnershipMigrationBackfillsOnlyDeterministicProjectOwnership(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00019_intake_ownership.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE auth_intake_ownership",
		"PRIMARY KEY (intake_id)",
		"FOREIGN KEY (intake_id) REFERENCES intakes(id) ON DELETE CASCADE",
		"FOREIGN KEY (owner_user_id) REFERENCES auth_users(id) ON DELETE CASCADE",
		"JOIN auth_batch_project_ownership",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	if strings.Contains(sql, "FROM intakes i") && !strings.Contains(sql, "JOIN batch_projects") {
		t.Fatal("migration must not blanket-claim legacy intakes without project ownership evidence")
	}

	downParts := strings.Split(sql, "-- +goose Down")
	if len(downParts) != 2 {
		t.Fatalf("migration must contain exactly one Goose Down section")
	}
	down := downParts[1]
	if strings.Contains(strings.ToUpper(down), "DROP ") {
		t.Fatal("irreversible Down must preserve the ownership table and facts")
	}
	for _, required := range []string{
		"-- +goose StatementBegin",
		"SIGNAL SQLSTATE '45000'",
		"intake ownership migration is irreversible",
		"-- +goose StatementEnd",
	} {
		if !strings.Contains(down, required) {
			t.Fatalf("irreversible Down missing %q", required)
		}
	}
}
