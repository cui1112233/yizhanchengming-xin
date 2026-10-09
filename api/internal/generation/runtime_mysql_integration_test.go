package generation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	_ "github.com/go-sql-driver/mysql"
)

func TestRuntimeMySQLSourceFencingAttemptsAndFinalizeAggregate(t *testing.T) {
	dsn := os.Getenv("TASK9_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TASK9_MYSQL_DSN not configured")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	name := fmt.Sprintf("task3-runtime-%d", time.Now().UnixNano())
	intakeResult, err := db.ExecContext(ctx, `INSERT INTO intakes(name,status) VALUES(?,'completed')`, name)
	if err != nil {
		t.Fatal(err)
	}
	intakeID, _ := intakeResult.LastInsertId()
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), `DELETE FROM intakes WHERE id=?`, intakeID) })
	projectResult, err := db.ExecContext(ctx, `INSERT INTO batch_projects(intake_id,name) VALUES(?,?)`, intakeID, name)
	if err != nil {
		t.Fatal(err)
	}
	projectID, _ := projectResult.LastInsertId()
	bookResult, err := db.ExecContext(ctx, `INSERT INTO books(intake_id,source,platform_id,external_book_id,title,original_text) VALUES(?,'test','1',?,?,?)`, intakeID, name, name, "body")
	if err != nil {
		t.Fatal(err)
	}
	bookID, _ := bookResult.LastInsertId()
	runtimeStore := task9runtime.NewMySQLStore(db)
	source, err := runtimeStore.AdmitGeneration(ctx, task9runtime.GenerationRequest{BatchProjectID: projectID, BookIDs: []int64{bookID}, RequestID: name + "-source", RequestedByUserID: 17})
	if err != nil || len(source.Items) != 1 {
		t.Fatalf("source admission=%+v err=%v", source, err)
	}
	sourceBookRunID := source.Items[0].BookRunID
	if _, err := db.ExecContext(ctx, `UPDATE book_runs SET status='failed',retryable=0,finished_at=UTC_TIMESTAMP(6) WHERE id=?`, sourceBookRunID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE runs SET status='failed',finished_at=UTC_TIMESTAMP(6) WHERE id=?`, source.Run.ID); err != nil {
		t.Fatal(err)
	}
	for attempt, row := range []struct {
		stage  Stage
		status Status
		output string
	}{{StageScript, StatusCompleted, "script"}, {StageHook, StatusSkipped, ""}, {StageDirector, StatusFailed, ""}} {
		if _, err := db.ExecContext(ctx, `INSERT INTO stage_runs(book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, sourceBookRunID, bookID, row.stage, row.status, attempt+1, name, "test.prompt", 1, `{}`, row.output, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	retry, err := runtimeStore.AdmitGeneration(ctx, task9runtime.GenerationRequest{BatchProjectID: projectID, BookIDs: []int64{bookID}, RequestID: name + "-retry", RequestedByUserID: 17, Action: task9runtime.GenerationActionStageRetry, RetryStage: string(StageDirector), SourceBookRunID: sourceBookRunID})
	if err != nil || len(retry.Items) != 1 {
		t.Fatalf("retry admission=%+v err=%v", retry, err)
	}
	execution, ok, err := runtimeStore.Claim(ctx, retry.Items[0], "task3-worker", time.Now().Add(time.Minute))
	if err != nil || !ok {
		t.Fatalf("claim=%+v ok=%v err=%v", execution, ok, err)
	}
	store := NewMySQLStore(db)
	run, input, err := store.RuntimeBookRun(ctx, execution)
	if err != nil || run.ID != execution.BookRunID || input.SourceBookRunID != sourceBookRunID || input.RetryStage != StageDirector {
		t.Fatalf("runtime run=%+v input=%+v err=%v", run, input, err)
	}
	first, err := store.CreateStageRunFenced(ctx, execution, StageRun{Stage: StageDirector, Status: StatusRunning, RequestID: name, InputSnapshot: `{}`})
	if err != nil || first.Attempt != 1 || first.BookID != bookID {
		t.Fatalf("first stage=%+v err=%v", first, err)
	}
	second, err := store.CreateStageRunFenced(ctx, execution, StageRun{Stage: StageDirector, Status: StatusRunning, RequestID: name, InputSnapshot: `{}`})
	if err != nil || second.Attempt != 2 {
		t.Fatalf("second stage=%+v err=%v", second, err)
	}
	stale := execution
	stale.FencingToken++
	if _, err := store.UpdateStageRunFenced(ctx, stale, second); !errors.Is(err, task9runtime.ErrStaleExecution) {
		t.Fatalf("stale update err=%v", err)
	}
	second.Status = StatusCompleted
	if _, err := store.UpdateStageRunFenced(ctx, execution, second); err != nil {
		t.Fatal(err)
	}
	durable, err := runtimeStore.Complete(ctx, execution)
	if err != nil || !durable {
		t.Fatalf("complete durable=%v err=%v", durable, err)
	}
	var bookStatus, runStatus string
	if err := db.QueryRowContext(ctx, `SELECT br.status,r.status FROM book_runs br JOIN runs r ON r.id=br.run_id WHERE br.id=?`, execution.BookRunID).Scan(&bookStatus, &runStatus); err != nil {
		t.Fatal(err)
	}
	if bookStatus != "succeeded" || runStatus != "succeeded" {
		t.Fatalf("book=%s run=%s", bookStatus, runStatus)
	}
}
