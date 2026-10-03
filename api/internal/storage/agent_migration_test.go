package storage

import (
	"os"
	"strings"
	"testing"
)

func TestAgentWorkspaceMigrationContainsDurableAgentTables(t *testing.T) {
	body, err := os.ReadFile("../../migrations/00003_agent_workspace.sql")
	if err != nil { t.Fatal(err) }
	sql := string(body)
	for _, table := range []string{
		"agent_threads",
		"agent_messages",
		"agent_tasks",
		"agent_tool_calls",
		"agent_message_media",
	} {
		if !strings.Contains(sql, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Fatalf("migration missing table %s", table)
		}
	}
	for _, required := range []string{"media_asset_id", "needs_decision", "arguments_json", "result_json", "owner"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing required contract %s", required)
		}
	}
	if strings.Contains(strings.ToLower(sql), "local_path") || strings.Contains(strings.ToLower(sql), "provider_url") {
		t.Fatal("agent schema must not make local/provider media paths authoritative")
	}
}
