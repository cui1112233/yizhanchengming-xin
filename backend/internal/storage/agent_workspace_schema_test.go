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
	if len(migrations) == 0 || migrations[len(migrations)-1].Version != 8800006 {
		t.Fatalf("last migration=%v, want 8800006", migrations[len(migrations)-1].Version)
	}
	for _, migration := range migrations[:len(migrations)-1] {
		if migration.Version >= 8800006 {
			t.Fatalf("agent migration must be greater than existing version %d", migration.Version)
		}
	}
}

func TestAgentWorkspaceGroupMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	if len(migrations) != 6 {
		t.Fatalf("AgentWorkspaceMigrations length=%d, want 6", len(migrations))
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
	assets := migrations[3]
	if assets.Version != 8800004 || assets.CallbackChecksum != "v88-agent-project-assets-v1" {
		t.Fatalf("asset migration=%+v", assets)
	}
	assetSchema := strings.ToLower(strings.Join(assets.SQL, "\n"))
	if !strings.Contains(assetSchema, "create table if not exists agent_assets") || !strings.Contains(assetSchema, "idx_agent_assets_owner_project_created") {
		t.Fatalf("missing project asset table or owner index: %v", assets.SQL)
	}
	config := migrations[4]
	if config.Version != 8800005 || config.CallbackChecksum != "v88-agent-canvas-node-config-v1" {
		t.Fatalf("node config migration=%+v", config)
	}
	if !strings.Contains(strings.ToLower(strings.Join(config.SQL, "\n")), "add column config_json json") {
		t.Fatalf("missing config_json migration: %v", config.SQL)
	}
	jobs := migrations[5]
	if jobs.Version != 8800006 || jobs.CallbackChecksum != "v88-agent-video-jobs-v1" {
		t.Fatalf("video jobs migration=%+v", jobs)
	}
	jobSchema := strings.ToLower(strings.Join(jobs.SQL, "\n"))
	for _, clause := range []string{"create table if not exists agent_video_jobs", "uq_agent_video_jobs_owner_idempotency", "idx_agent_video_jobs_project_node_updated"} {
		if !strings.Contains(jobSchema, clause) {
			t.Fatalf("missing video job schema clause %q: %v", clause, jobs.SQL)
		}
	}
}
