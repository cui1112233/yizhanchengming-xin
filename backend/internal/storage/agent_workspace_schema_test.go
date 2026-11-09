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

func TestAppMigrationsIncludeLatestAgentWorkspaceMigration(t *testing.T) {
	migrations := AppMigrations()
	found := false
	for _, migration := range migrations {
		if migration.Version == 8800017 && migration.CallbackChecksum == "v88-agent-script-shot-merge-v1" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing agent script shot merge migration 8800017")
	}
}

func TestAgentWorkspaceGroupMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	if len(migrations) != 16 {
		t.Fatalf("AgentWorkspaceMigrations length=%d, want 16", len(migrations))
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
	if config.Adopt == nil {
		t.Fatal("node config migration must adopt an already-applied config_json column into the migration ledger")
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
	viewState := migrations[6]
	if viewState.Version != 8800007 || viewState.CallbackChecksum != "v88-agent-workspace-view-state-v1" {
		t.Fatalf("workspace view state migration=%+v", viewState)
	}
	if !strings.Contains(strings.ToLower(strings.Join(viewState.SQL, "\n")), "add column view_state_json json") {
		t.Fatalf("missing view_state_json migration: %v", viewState.SQL)
	}
	generation := migrations[7]
	if generation.Version != 8800008 || generation.CallbackChecksum != "v88-agent-generation-tasks-v1" {
		t.Fatalf("generation task migration=%+v", generation)
	}
	generationSchema := strings.ToLower(strings.Join(generation.SQL, "\n"))
	for _, clause := range []string{
		"create table if not exists agent_generation_tasks",
		"create table if not exists agent_generation_results",
		"uq_agent_generation_tasks_owner_project_idempotency",
		"idx_agent_generation_tasks_owner_project_updated",
		"idx_agent_generation_results_task",
	} {
		if !strings.Contains(generationSchema, clause) {
			t.Fatalf("missing generation schema clause %q: %v", clause, generation.SQL)
		}
	}
	snapshot := migrations[8]
	if snapshot.Version != 8800009 || snapshot.CallbackChecksum != "v88-agent-canvas-json-snapshot-v1" {
		t.Fatalf("canvas snapshot migration=%+v", snapshot)
	}
	if !strings.Contains(strings.ToLower(strings.Join(snapshot.SQL, "\n")), "add column canvas_json json") {
		t.Fatalf("missing canvas_json migration: %v", snapshot.SQL)
	}
	backfill := migrations[9]
	if backfill.Version != 8800010 || backfill.CallbackChecksum != "v88-agent-canvas-json-backfill-v1" {
		t.Fatalf("canvas backfill migration=%+v", backfill)
	}
	backfillSQL := strings.ToLower(strings.Join(backfill.SQL, "\n"))
	for _, clause := range []string{"update agent_canvases", "json_arrayagg", "agent_canvas_nodes", "agent_canvas_edges", "where c.canvas_json is null"} {
		if !strings.Contains(backfillSQL, clause) {
			t.Fatalf("missing canvas snapshot backfill clause %q: %v", clause, backfill.SQL)
		}
	}
	shots := migrations[10]
	if shots.Version != 8800011 || shots.CallbackChecksum != "v88-agent-script-shots-v1" {
		t.Fatalf("script shots migration=%+v", shots)
	}
	shotSchema := strings.ToLower(strings.Join(shots.SQL, "\n"))
	for _, clause := range []string{
		"create table if not exists agent_script_shots",
		"primary key (project_id, id)",
		"unique key uq_agent_script_shots_sequence (project_id, script_node_id, sequence)",
		"create table if not exists agent_shot_generation_links",
		"unique key uq_agent_shot_generation_links (project_id, shot_id, node_id, generation_id)",
		"owner_username",
	} {
		if !strings.Contains(shotSchema, clause) {
			t.Fatalf("missing script shot schema clause %q: %v", clause, shots.SQL)
		}
	}
}

func TestAgentWorkspaceSchemaIncludesScriptShots(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	shots := migrations[10]
	if shots.Version != 8800011 || shots.CallbackChecksum != "v88-agent-script-shots-v1" {
		t.Fatalf("script shots migration=%+v", shots)
	}
	schema := strings.ToLower(strings.Join(shots.SQL, "\n"))
	for _, clause := range []string{
		"create table if not exists agent_script_shots",
		"primary key (project_id, id)",
		"unique key uq_agent_script_shots_sequence (project_id, script_node_id, sequence)",
		"create table if not exists agent_shot_generation_links",
		"unique key uq_agent_shot_generation_links (project_id, shot_id, node_id, generation_id)",
		"owner_username",
	} {
		if !strings.Contains(schema, clause) {
			t.Fatalf("missing script shot schema clause %q: %v", clause, shots.SQL)
		}
	}
}

func TestAgentWorkspaceBatchFactoryBindingMigrationIsAppendOnlyAndSafe(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	binding := migrations[len(migrations)-5]
	if binding.Version != 8800012 || binding.CallbackChecksum != "v88-agent-batch-factory-bindings-v1" {
		t.Fatalf("batch factory binding migration=%+v", binding)
	}
	schema := strings.ToLower(strings.Join(binding.SQL, "\n"))
	for _, clause := range []string{
		"create table if not exists agent_batch_factory_bindings",
		"unique key uq_agent_batch_factory_bindings_owner_batch (owner_username, batch_id)",
		"unique key uq_agent_batch_factory_bindings_owner_project (owner_username, project_id)",
		"create table if not exists agent_batch_factory_node_origins",
		"primary key (project_id, node_id)",
		"idx_agent_batch_factory_node_origins_owner_book",
	} {
		if !strings.Contains(schema, clause) {
			t.Fatalf("missing batch factory binding schema clause %q: %v", clause, binding.SQL)
		}
	}
	for _, forbidden := range []string{"media_bytes", "object_key", "provider_credential"} {
		if strings.Contains(schema, forbidden) {
			t.Fatalf("batch factory binding schema contains forbidden column %q", forbidden)
		}
	}
}

func TestAgentWorkspaceGenerationHistoryIndexMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	history := migrations[len(migrations)-4]
	if history.Version != 8800013 || history.CallbackChecksum != "v88-agent-generation-history-index-v1" {
		t.Fatalf("generation history index migration=%+v", history)
	}
	schema := strings.ToLower(strings.Join(history.SQL, "\n"))
	if !strings.Contains(schema, "idx_agent_generation_tasks_owner_history") {
		t.Fatalf("missing generation history index: %v", history.SQL)
	}
	if strings.Contains(schema, "drop ") || strings.Contains(schema, "delete ") {
		t.Fatalf("generation history index migration must be additive: %v", history.SQL)
	}
}

func TestAgentWorkspaceMaterialLibraryMigrationIsAppendOnlyAndStoresReferencesOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	library := migrations[len(migrations)-3]
	if library.Version != 8800014 || library.CallbackChecksum != "v88-agent-material-library-v1" {
		t.Fatalf("material library migration=%+v", library)
	}
	schema := strings.ToLower(strings.Join(library.SQL, "\n"))
	for _, clause := range []string{
		"create table if not exists agent_library_folders",
		"create table if not exists agent_library_items",
		"unique key uq_agent_library_items_owner_folder_source (owner_username, folder_id, source_project_id, source_asset_id)",
		"idx_agent_library_items_owner_folder_updated",
	} {
		if !strings.Contains(schema, clause) {
			t.Fatalf("missing material library schema clause %q: %v", clause, library.SQL)
		}
	}
	for _, forbidden := range []string{"object_key", "media_bytes", "credential"} {
		if strings.Contains(schema, forbidden) {
			t.Fatalf("material library schema contains forbidden column %q", forbidden)
		}
	}
}

func TestAgentWorkspaceProjectTypeMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	typeMigration := migrations[len(migrations)-2]
	if typeMigration.Version != 8800015 || typeMigration.CallbackChecksum != "v88-agent-project-type-v1" {
		t.Fatalf("project type migration=%+v", typeMigration)
	}
	schema := strings.ToLower(strings.Join(typeMigration.SQL, "\n"))
	if !strings.Contains(schema, "add column project_type") || strings.Contains(schema, "drop ") || strings.Contains(schema, "delete ") {
		t.Fatalf("project type migration must be additive: %v", typeMigration.SQL)
	}
}

func TestAgentWorkspaceShotMergeMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	merge := migrations[len(migrations)-1]
	if merge.Version != 8800017 || merge.CallbackChecksum != "v88-agent-script-shot-merge-v1" {
		t.Fatalf("shot merge migration=%+v", merge)
	}
	schema := strings.ToLower(strings.Join(merge.SQL, "\n"))
	for _, clause := range []string{"add column duration_seconds", "add column source_shot_ids_json"} {
		if !strings.Contains(schema, clause) {
			t.Fatalf("missing %q in %v", clause, merge.SQL)
		}
	}
	if strings.Contains(schema, "drop ") || strings.Contains(schema, "delete ") {
		t.Fatalf("shot merge migration must be additive: %v", merge.SQL)
	}
}
