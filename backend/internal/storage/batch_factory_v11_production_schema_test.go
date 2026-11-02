package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestV11ProductionTaskAttemptIndexMigrationLeavesOriginalProductionMigrationUnchanged(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11ProductionStatements(), "\n"))
	if strings.Contains(joined, "opening_variant_index") {
		t.Fatalf("already-ledgered production migration must not contain opening_variant_index: %s", joined)
	}
	if !strings.Contains(joined, "unique key uq_bfv11_production_task_attempt (job_id, video_id, attempt)") {
		t.Fatalf("original production task unique key is missing: %s", joined)
	}
	var migration *Migration
	for i := range V11Migrations() {
		if V11Migrations()[i].Version == 1100026 {
			migration = &V11Migrations()[i]
			break
		}
	}
	if migration == nil || migration.Reconcile == nil || migration.Adopt == nil {
		t.Fatal("production task unique-key migration must reconcile existing and unledgered databases")
	}
}

func TestV11ProductionTaskAttemptIndexMigrationDropsLegacyIndexAndRecreatesIt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COALESCE\\( \\(SELECT table_name").WithArgs("batch_factory_v11_production_tasks", "batch_factory_v12_production_tasks").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("batch_factory_v11_production_tasks"))
	mock.ExpectQuery("SELECT COALESCE\\(GROUP_CONCAT").WithArgs("batch_factory_v11_production_tasks").WillReturnRows(sqlmock.NewRows([]string{"columns"}).AddRow("job_id,video_id,attempt"))
	mock.ExpectQuery("SELECT COALESCE\\( \\(SELECT table_name").WithArgs("batch_factory_v11_production_tasks", "batch_factory_v12_production_tasks").WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("batch_factory_v11_production_tasks"))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM information_schema.statistics").WithArgs("batch_factory_v11_production_tasks", "idx_bfv11_production_tasks_job_id").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("ALTER TABLE batch_factory_v11_production_tasks ADD KEY idx_bfv11_production_tasks_job_id \\(job_id\\)").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("ALTER TABLE batch_factory_v11_production_tasks DROP INDEX uq_bfv11_production_task_attempt").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("ALTER TABLE batch_factory_v11_production_tasks ADD UNIQUE KEY uq_bfv11_production_task_attempt").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureV11ProductionTaskAttemptIndex(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestV11MergeProgressMigrationReconcilesRecordedPartialSchemas(t *testing.T) {
	for _, migration := range V11Migrations() {
		if migration.Version == 1100020 {
			if migration.Reconcile == nil {
				t.Fatal("merge progress migration must reconcile a historically recorded partial schema")
			}
			return
		}
	}
	t.Fatal("merge progress migration is missing")
}

func TestV11OpeningVariantMigrationReconcilesRecordedPartialSchemas(t *testing.T) {
	for _, migration := range V11Migrations() {
		if migration.Version == 1100021 {
			if migration.Reconcile == nil {
				t.Fatal("opening variant migration must reconcile a historically recorded partial schema")
			}
			return
		}
	}
	t.Fatal("opening variant migration is missing")
}

func TestV11BookStageRunsMigrationIsAdditiveAndDurable(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 16 || migrations[14].Version != 1100015 || migrations[14].CallbackChecksum != "batch-factory-v11-book-stage-runs-v1" {
		t.Fatalf("stage migration=%+v", migrations)
	}
	var detailsMigration Migration
	for _, candidate := range migrations {
		if candidate.Version == 1100024 {
			detailsMigration = candidate
			break
		}
	}
	if detailsMigration.Version != 1100024 || detailsMigration.CallbackChecksum != "batch-factory-v11-book-asset-details-v1" {
		t.Fatalf("book asset details migration=%+v", detailsMigration)
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
