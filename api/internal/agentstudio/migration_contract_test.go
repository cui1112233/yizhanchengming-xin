package agentstudio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentStudioMigrationKeepsAttachmentBytesOutOfMySQL(t *testing.T) {
	path := filepath.Join("..", "..", "db", "migrations", "00014_task16_agent_studio.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sql := strings.ToLower(string(contents))
	for _, required := range []string{
		"create table agent_attachments", "project_id", "owner_user_id", "bucket", "object_key", "filename", "content_type", "byte_size", "sha256",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	if strings.Contains(sql, "attachment_blob") || strings.Contains(sql, "attachment_data") {
		t.Fatal("attachment migration must persist object metadata, not bytes")
	}
}
