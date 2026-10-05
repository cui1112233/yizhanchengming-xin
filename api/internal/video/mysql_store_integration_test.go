package video

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"os"
	"testing"
	"time"

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

	job, created, err := store1.CreateOrGetProductionJob(ctx, ProductionJob{BatchProjectID: projectID, BookID: bookID, Status: JobRunning, InputRevision: "rev-restart", FinalPromptStageRunID: stageRunID, FinalPromptVersion: 1, FinalPromptText: "compiled prompt", Provider: ProviderPersonalAPI, Model: ModelYD20Mini, IdempotencyKey: "restart-contract-integration"})
	if err != nil { t.Fatal(err) }
	if !created { t.Fatal("expected a new production job") }
	task, err := store1.CreateProductionTask(ctx, ProductionTask{ProductionJobID: job.ID, Attempt: 1, Provider: ProviderPersonalAPI, Model: ModelYD20Mini, ProviderJobID: "remote-restart", Status: TaskRunning, OutputURL: "https://tos.example/segment.mp4"})
	if err != nil { t.Fatal(err) }

	now := time.Now().UTC().Truncate(time.Microsecond)
	tokenHash := sha256.Sum256([]byte("restart-local-executor-token"))
	executor := LocalExecutorRecord{
		ID: "lex_restart", Name: "restart executor", ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance,
		Capabilities: []string{"text_to_video"}, TokenHash: tokenHash, LastSeenAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := store1.CreateLocalExecutor(ctx, executor); err != nil { t.Fatal(err) }
	defer db.ExecContext(context.Background(), `DELETE FROM video_local_executor_tasks WHERE executor_id=? OR id=?`, executor.ID, "let_restart")
	defer db.ExecContext(context.Background(), `DELETE FROM video_local_executors WHERE id=?`, executor.ID)
	localTask := LocalExecutorTask{
		ID: "let_restart", SourceTaskID: "video-task-restart", ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance,
		Prompt: "durable prompt", RequestID: "local-restart", Status: TaskQueued, CreatedAt: now, UpdatedAt: now,
	}
	if err := store1.CreateLocalExecutorTask(ctx, localTask); err != nil { t.Fatal(err) }

	mergeJob, err := store1.CreateMergeJob(ctx, MergeJob{BatchProjectID: projectID, BookID: bookID, Status: MergeQueued, CurrentAttempt: 1, CreatedAt: now, UpdatedAt: now})
	if err != nil { t.Fatal(err) }
	mergeAttempt, err := store1.CreateMergeAttempt(ctx, MergeAttempt{
		MergeJobID: mergeJob.ID, Attempt: 1, Status: MergeQueued, AspectRatio: "9:16", Speed: 1,
		Inputs: []MergeInputAsset{{ProductionTaskID: task.ID, URL: "https://tos.example/segment.mp4", Order: 1}}, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil { t.Fatal(err) }

	store2 := NewMySQLStore(db)
	recoveredExecutor, err := store2.GetLocalExecutorByTokenHash(ctx, tokenHash)
	if err != nil { t.Fatal(err) }
	if recoveredExecutor.ID != executor.ID || recoveredExecutor.ProviderKey != ProviderDoubaoLocalExecutor || recoveredExecutor.Model != ModelDoubaoSeedance {
		t.Fatalf("new store instance did not recover executor: %+v", recoveredExecutor)
	}
	recoveredLocalTask, err := store2.GetLocalExecutorTask(ctx, localTask.ID)
	if err != nil { t.Fatal(err) }
	if recoveredLocalTask.ID != localTask.ID || recoveredLocalTask.Status != TaskQueued || recoveredLocalTask.Prompt != localTask.Prompt {
		t.Fatalf("new store instance did not recover local task: %+v", recoveredLocalTask)
	}
	recoveredMerge, err := store2.GetMergeAttempt(ctx, mergeAttempt.ID)
	if err != nil { t.Fatal(err) }
	if recoveredMerge.MergeJobID != mergeJob.ID || recoveredMerge.Status != MergeQueued || len(recoveredMerge.Inputs) != 1 || recoveredMerge.Inputs[0].ProductionTaskID != task.ID {
		t.Fatalf("new store instance did not recover merge attempt: %+v", recoveredMerge)
	}

	projectJobs, err := store2.ListLatestProductionJobsByProject(ctx, projectID)
	if err != nil { t.Fatal(err) }
	if len(projectJobs) != 1 || projectJobs[0].ID != job.ID {
		t.Fatalf("project video status jobs = %+v", projectJobs)
	}
	projectTasks, err := store2.ListProductionTasksByJobIDs(ctx, []int64{job.ID})
	if err != nil { t.Fatal(err) }
	if len(projectTasks[job.ID]) != 1 || projectTasks[job.ID][0].ID != task.ID {
		t.Fatalf("project video status tasks = %+v", projectTasks)
	}

	recoverable, err := store2.ListRecoverableTasks(ctx, 100)
	if err != nil { t.Fatal(err) }
	for _, got := range recoverable {
		if got.ID == task.ID && got.ProviderJobID == "remote-restart" && got.Status == TaskRunning {
			return
		}
	}
	t.Fatalf("new store instance did not recover task %d: %+v", task.ID, recoverable)
}