package storage

import (
	"strings"
	"testing"
)

func TestV11ProductionMigrationAddsDurableJobTaskAndEventTables(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11ProductionStatements(), "\n"))
	for _, table := range []string{
		"batch_factory_v11_production_jobs",
		"batch_factory_v11_production_tasks",
		"batch_factory_v11_production_events",
	} {
		if !strings.Contains(joined, table) {
			t.Fatalf("missing production table %s", table)
		}
	}
	if !strings.Contains(joined, "final_prompt_hash") || !strings.Contains(joined, "unique key") {
		t.Fatalf("production migration must persist prompt hashes and idempotency: %s", joined)
	}
}

func TestV11MigrationsRegisterProductionAfterDirector(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 6 || migrations[4].Version != 1100005 {
		t.Fatalf("production migration missing: %+v", migrations)
	}
}

func TestV11BookStageRunsMigrationIsAdditiveAndDurable(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 16 || migrations[14].Version != 1100015 || migrations[14].CallbackChecksum != "batch-factory-v11-book-stage-runs-v1" {
		t.Fatalf("stage migration=%+v", migrations)
	}
	last := migrations[len(migrations)-1]
	if last.Version != 1100023 || last.CallbackChecksum != "batch-factory-v11-shuihuo-config-v1" {
		t.Fatalf("last migration=%+v", last)
	}
	if joined := strings.ToLower(strings.Join(V11ShuihuoConfigStatements(), "\n")); !strings.Contains(joined, "batch_factory_v11_shuihuo_configs") || !strings.Contains(joined, "config_json") {
		t.Fatalf("shuihuo config migration=%s", joined)
	}
	if joined := strings.ToLower(strings.Join(V11LocalArtifactPurgeStatements(), "\n")); !strings.Contains(joined, "batch_factory_v11_local_artifact_purges") || !strings.Contains(joined, "purge_after") || !strings.Contains(joined, "storage_ref") {
		t.Fatalf("local artifact purge migration=%s", joined)
	}
	if joined := strings.ToLower(strings.Join(V11ProductionDurationStatements(), "\n")); !strings.Contains(joined, "target_duration_seconds") || !strings.Contains(joined, "requested_duration_seconds") || !strings.Contains(joined, "actual_duration_seconds") {
		t.Fatalf("duration migration=%s", joined)
	}
	if joined := strings.ToLower(strings.Join(V11BookMergeStatements(), "\n")); !strings.Contains(joined, "batch_factory_v11_merge_jobs") || !strings.Contains(joined, "book_id") {
		t.Fatalf("book merge migration=%s", joined)
	}
	if joined := strings.ToLower(strings.Join(V11MergeSourceDurationStatements(), "\n")); !strings.Contains(joined, "actual_duration_seconds") {
		t.Fatalf("merge source duration migration=%s", joined)
	}
	if joined := strings.ToLower(strings.Join(V11ProductionLibraryStatements(), "\n")); !strings.Contains(joined, "batch_factory_v11_hidden_production_tasks") || !strings.Contains(joined, "deleted_at") {
		t.Fatalf("production library migration=%s", joined)
	}
	if joined := strings.ToLower(strings.Join(V11MergeProgressStatements(), "\n")); !strings.Contains(joined, "progress_phase") || !strings.Contains(joined, "speed") {
		t.Fatalf("merge progress migration=%s", joined)
	}
	joined := strings.ToLower(strings.Join(V11BookStageRunsStatements(), "\n"))
	for _, required := range []string{"batch_factory_v11_book_stage_runs", "book_id", "stage", "attempt", "error_message", "idx_bfv11_stage_owner_book_updated"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("stage migration misses %q", required)
		}
	}
}
