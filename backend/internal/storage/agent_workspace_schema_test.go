package storage

import (
	"strings"
	"testing"
)

func TestAgentWorkspaceMigrationIsAppendOnlyAndComplete(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	if len(migrations) < 1 {
		t.Fatal("AgentWorkspaceMigrations is empty")
	}
	migration := migrations[0]
	if migration.Version != 8800001 {
		t.Fatalf("version=%d, want 8800001", migration.Version)
	}
	if migration.CallbackChecksum != "v88-agent-workspace-foundation-v1" {
		t.Fatalf("checksum marker=%q", migration.CallbackChecksum)
	}

	schema := strings.ToLower(strings.Join(migration.SQL, "\n"))
	for _, table := range []string{
		"agent_projects",
		"agent_messages",
		"agent_canvases",
		"agent_canvas_nodes",
		"agent_canvas_edges",
		"agent_node_versions",
		"agent_executions",
		"agent_legacy_imports",
	} {
		if !strings.Contains(schema, "create table if not exists "+table) {
			t.Errorf("missing idempotent table %s", table)
		}
	}
}

func TestAppMigrationsEndsWithAgentWorkspaceMigration(t *testing.T) {
	migrations := AppMigrations()
	if len(migrations) == 0 || migrations[len(migrations)-1].Version != 8800003 {
		t.Fatalf("last migration=%v, want 8800003", migrations[len(migrations)-1].Version)
	}
	for _, migration := range migrations[:len(migrations)-1] {
		if migration.Version >= 8800003 {
			t.Fatalf("agent migration must be greater than existing version %d", migration.Version)
		}
	}
}

func TestAgentWorkspaceGroupMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	if len(migrations) != 3 {
		t.Fatalf("AgentWorkspaceMigrations length=%d, want 3", len(migrations))
	}
	if migrations[0].Version != 8800001 || migrations[0].CallbackChecksum != "v88-agent-workspace-foundation-v1" {
		t.Fatalf("historical migration changed: %+v", migrations[0])
	}
	group := migrations[1]
	if group.Version != 8800002 || group.CallbackChecksum != "v88-agent-canvas-groups-v1" {
		t.Fatalf("group migration=%+v", group)
	}
	if !strings.Contains(strings.ToLower(strings.Join(group.SQL, "\n")), "add column group_id") {
		t.Fatalf("missing group_id migration: %v", group.SQL)
	}
	appearance := migrations[2]
	if appearance.Version != 8800003 || appearance.CallbackChecksum != "v88-agent-canvas-appearance-v1" {
		t.Fatalf("appearance migration=%+v", appearance)
	}
	if !strings.Contains(strings.ToLower(strings.Join(appearance.SQL, "\n")), "add column appearance_json") {
		t.Fatalf("missing appearance_json migration: %v", appearance.SQL)
	}
}
