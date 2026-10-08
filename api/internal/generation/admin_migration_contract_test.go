package generation

import (
	"os"
	"strings"
	"testing"
)

func TestAdminGovernanceMigrationDefinesPromptLifecycleAndAudit(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00015_task7_admin_governance.sql")
	if err != nil { t.Fatal(err) }
	for _, required := range []string{"lifecycle ENUM('draft','published','archived')", "uq_generation_prompts_one_published", "CREATE TABLE admin_audit_logs", "content_sha256", "request_id"} {
		if !strings.Contains(string(body), required) { t.Fatalf("migration missing %q", required) }
	}
	if strings.Contains(string(body), "你是") { t.Fatal("migration must not contain system prompt body") }
}
