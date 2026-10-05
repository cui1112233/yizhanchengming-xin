package video

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLRestartRecoveryIntegration(t *testing.T) {
	dsn := os.Getenv("TASK14_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("TASK14_INTEGRATION_DSN is set by the Goose/MySQL CI job")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	intakeResult, err := db.ExecContext(ctx, `INSERT INTO intakes(name,status) VALUES('task14-recovery-integration','pending')`)
	if err != nil { t.Fatal(err) }
	intakeID, _ := intakeResult.LastInsertId()
	defer db.ExecContext(context.Background(), `DELETE FROM intakes WHERE id=?`, intakeID)

	bookResult, err := db.ExecContext(ctx, `INSERT INTO books(intake_id,source,platform_id,external_book_id,title,original_text,status) VALUES(?,?,?,?,?,?,?)`, intakeID, "task14", "task14", "restart-recovery", "Task14 Recovery", "source", "ready")
	if err != nil { t.Fatal(err) }
	bookID, _ := bookResult.LastInsertId()

	projectResult, err := db.ExecContext(ctx, `INSERT INTO batch_projects(intake_id,name) VALUES(?,?)`, intakeID, "Task14 Recovery Project")
	if err != nil { t.Fatal(err) }
	projectID, _ := projectResult.LastInsertId()

	bookRunResult, err := db.ExecContext(ctx, `INSERT INTO book_runs(batch_project_id,book_id,status,request_id) VALUES(?,?,?,?)`, projectID, bookID, "completed", "task14-recovery")
	if err != nil { t.Fatal(err) }
	bookRunID, _ := bookRunResult.LastInsertId()

	stageResult, err := db.ExecContext(ctx, `INSERT INTO stage_runs(book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text) VALUES(?,?,?,?,?,?,?,?,?,?)`, bookRunID, bookID, "FINAL_PROMPT", "completed", 1, "task14-recovery", "final_prompt.default", 1, `{"revision":"restart"}`, "compiled prompt")
	if err != nil { t.Fatal(err) }
	stageRunID, _ := stageResult.LastInsertId()

	store1 := NewMySQLStore(db)
	config := ProviderConfig{ProviderKey: ProviderPersonalAPI, Model: ModelYD20Mini, EncryptedSecret: []byte("cipher"), SecretNonce: []byte("nonce"), Enabled: true}
	if err := store1.UpsertProviderConfig(ctx, config); err != nil { t.Fatal(err) }
	defer db.ExecContext(context.Background(), `DELETE FROM video_provider_configs WHERE provider_key=? AND model=?`, ProviderPersonalAPI, ModelYD20Mini)

	job, created, err := store1.CreateOrGetProductionJob(ctx, ProductionJob{BatchProjectID: projectID, BookID: bookID, Status: JobRunning, InputRevision: "rev-restart", FinalPromptStageRunID: stageRunID, FinalPromptVersion: 1, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, IdempotencyKey: "restart-contract-integration"})
	if err != nil { t.Fatal(err) }
	if !created { t.Fatal("expected a new production job") }
	task, err := store1.CreateProductionTask(ctx, ProductionTask{ProductionJobID: job.ID, Attempt: 1, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, ProviderJobID: "remote-restart", Status: TaskRunning})
	if err != nil { t.Fatal(err) }

	store2 := NewMySQLStore(db)
	recoverable, err := store2.ListRecoverableTasks(ctx, 100)
	if err != nil { t.Fatal(err) }
	for _, got := range recoverable {
		if got.ID == task.ID && got.ProviderJobID == "remote-restart" && got.Status == TaskRunning {
			return
		}
	}
	t.Fatalf("new store instance did not recover task %d: %+v", task.ID, recoverable)
}
