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
		"DROP TABLE IF EXISTS auth_intake_ownership",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	if strings.Contains(sql, "FROM intakes i") && !strings.Contains(sql, "JOIN batch_projects") {
		t.Fatal("migration must not blanket-claim legacy intakes without project ownership evidence")
	}
}
