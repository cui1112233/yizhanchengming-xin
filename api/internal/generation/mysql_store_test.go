package generation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestMySQLStoreResolvePromptUsesLatestEnabledVersion(t *testing.T) {
	db, mock := newSQLMock(t)
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	query := regexp.QuoteMeta(`SELECT id,prompt_key,version,content,enabled,created_at,updated_at FROM generation_prompts WHERE prompt_key=? AND enabled=1 AND lifecycle='published' ORDER BY version DESC LIMIT 1`)
	mock.ExpectQuery(query).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "prompt_key", "version", "content", "enabled", "created_at", "updated_at"}).AddRow(7, PromptScript, 3, "v3", true, now, now))
	store := NewMySQLStore(db)
	prompt, err := store.ResolvePrompt(context.Background(), PromptScript)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.Version != 3 || prompt.Content != "v3" || !prompt.Enabled {
		t.Fatalf("prompt=%#v", prompt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreResolvePromptIgnoresDraftLifecycle(t *testing.T) {
	db, mock := newSQLMock(t)
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	query := regexp.QuoteMeta(`SELECT id,prompt_key,version,content,enabled,created_at,updated_at FROM generation_prompts WHERE prompt_key=? AND enabled=1 AND lifecycle='published' ORDER BY version DESC LIMIT 1`)
	mock.ExpectQuery(query).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "prompt_key", "version", "content", "enabled", "created_at", "updated_at"}).AddRow(9, PromptScript, 3, "published", true, now, now))

	prompt, err := NewMySQLStore(db).ResolvePrompt(context.Background(), PromptScript)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.Content != "published" || prompt.Version != 3 {
		t.Fatalf("prompt=%#v", prompt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreMapsRuntimeSucceededToPublicCompleted(t *testing.T) {
	db, mock := newSQLMock(t)
	now := time.Now()
	mock.ExpectQuery("SELECT id,run_id,batch_project_id,book_id,status.*FROM book_runs WHERE batch_project_id").WithArgs(int64(3), int64(11)).WillReturnRows(sqlmock.NewRows([]string{"id", "run_id", "batch_project_id", "book_id", "status", "request_id", "error_message", "started_at", "finished_at", "created_at", "updated_at"}).AddRow(90, 80, 3, 11, "succeeded", "req", "", now, now, now, now))
	run, err := NewMySQLStore(db).LatestBookRun(context.Background(), 3, 11)
	if err != nil || run.RunID != 80 || run.Status != StatusCompleted {
		t.Fatalf("run=%+v err=%v", run, err)
	}
}

func TestMySQLStoreLegacyUpdateRejectsRuntimeBookRun(t *testing.T) {
	db, mock := newSQLMock(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT run_id FROM book_runs WHERE id=?`)).WithArgs(int64(90)).WillReturnRows(sqlmock.NewRows([]string{"run_id"}).AddRow(80))
	if _, err := NewMySQLStore(db).UpdateBookRun(context.Background(), BookRun{ID: 90, RunID: 80, BatchProjectID: 3, BookID: 11, Status: StatusFailed}); !errors.Is(err, ErrConflict) {
		t.Fatalf("err=%v", err)
	}
}

func TestMySQLStoreRuntimeBookRunLoadsExactFencedSnapshotAndConfig(t *testing.T) {
	db, mock := newSQLMock(t)
	snapshot := task9runtime.GenerationSnapshot{SchemaVersion: 2, BatchProjectID: 3, BookIDs: []int64{11}, DirectorMode: "normal", ShotDurationLimitSec: 15, RequestedByUserID: 7, Action: task9runtime.GenerationActionFull, Config: task9runtime.GenerationConfigSnapshot{ProcessingRules: "RULES", ModelConfig: "MODEL"}}
	body, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	now := time.Now()
	mock.ExpectQuery("SELECT br.id,br.run_id.*FROM book_runs br JOIN runs r").WithArgs(int64(90)).WillReturnRows(sqlmock.NewRows([]string{"id", "run_id", "project_id", "book_id", "book_status", "request_id", "started_at", "created_at", "updated_at", "attempt", "token", "owner", "run_status", "run_kind", "schema", "snapshot", "hash", "actor", "archived_at"}).AddRow(90, 80, 3, 11, "running", "req", now, now, now, 2, 17, "worker-a", "running", "generation", 2, body, hash, 7, nil))
	execution := task9runtime.Execution{BookRunID: 90, Attempt: 2, FencingToken: 17, Owner: "worker-a"}
	run, input, err := NewMySQLStore(db).RuntimeBookRun(context.Background(), execution)
	if err != nil || run.RunID != 80 || input.Action != task9runtime.GenerationActionFull || input.Request.ProcessingRules != "RULES" || input.Request.ModelConfig != "MODEL" {
		t.Fatalf("run=%+v input=%+v err=%v", run, input, err)
	}
}

func TestMySQLStoreCreateStageRunFencedPropagatesAttemptReadFailure(t *testing.T) {
	db, mock := newSQLMock(t)
	execution := task9runtime.Execution{BookRunID: 90, Attempt: 2, FencingToken: 17, Owner: "worker-a"}
	mock.ExpectQuery("SELECT run_id,batch_project_id FROM book_runs").WithArgs(int64(90)).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id"}).AddRow(80, 3))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectQuery("SELECT batch_project_id,status,run_kind FROM runs.*FOR UPDATE").WithArgs(int64(80)).WillReturnRows(sqlmock.NewRows([]string{"batch_project_id", "status", "run_kind"}).AddRow(3, "running", "generation"))
	mock.ExpectQuery("SELECT book_id,attempt,execution_token,execution_owner,status FROM book_runs.*FOR UPDATE").WithArgs(int64(90), int64(80), int64(3)).WillReturnRows(sqlmock.NewRows([]string{"book_id", "attempt", "execution_token", "execution_owner", "status"}).AddRow(11, 2, 17, "worker-a", "running"))
	injected := errors.New("attempt read failed")
	mock.ExpectQuery("SELECT MAX\\(attempt\\) FROM stage_runs.*FOR UPDATE").WithArgs(int64(90), StageScript).WillReturnError(injected)
	mock.ExpectRollback()
	if _, err := NewMySQLStore(db).CreateStageRunFenced(context.Background(), execution, StageRun{Stage: StageScript}); !errors.Is(err, injected) {
		t.Fatalf("err=%v", err)
	}
}

func TestGenerationSafeOutcomeMySQLStageWrite(t *testing.T) {
	db, mock := newSQLMock(t)
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	validation := `{"error":"生成阶段执行失败，请稍后重试","matchAudio":true,"valid":false}`
	insert := regexp.QuoteMeta(`INSERT INTO stage_runs(book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	mock.ExpectExec(insert).WithArgs(int64(2), int64(11), StageDirector, StatusFailed, 1, "req", PromptDirector, 4, "{}", "", outcomeFailureMessage, validation, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(9, 1))
	selectQ := regexp.QuoteMeta(`SELECT id,book_run_id,book_id,stage,status,attempt,request_id,prompt_key,prompt_version,input_snapshot,output_text,error_message,validation_result,started_at,finished_at,created_at,updated_at FROM stage_runs WHERE id=?`)
	mock.ExpectQuery(selectQ).WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"id", "book_run_id", "book_id", "stage", "status", "attempt", "request_id", "prompt_key", "prompt_version", "input_snapshot", "output_text", "error_message", "validation_result", "started_at", "finished_at", "created_at", "updated_at"}).AddRow(9, 2, 11, StageDirector, StatusFailed, 1, "req", PromptDirector, 4, "{}", "", outcomeFailureMessage, validation, now, now, now, now))
	store := NewMySQLStore(db)
	started, finished := now, now
	value, err := store.CreateStageRun(context.Background(), StageRun{BookRunID: 2, BookID: 11, Stage: StageDirector, Status: StatusFailed, Attempt: 1, RequestID: "req", PromptKey: PromptDirector, PromptVersion: 4, InputSnapshot: "{}", ErrorMessage: safeError(errors.New("provider-canary-short")), ValidationResult: validation, StartedAt: &started, FinishedAt: &finished})
	if err != nil {
		t.Fatal(err)
	}
	if value.PromptVersion != 4 || value.ErrorMessage != outcomeFailureMessage || value.ValidationResult != validation {
		t.Fatalf("stage=%#v", value)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationSafeOutcomeMySQLFailRunWrite(t *testing.T) {
	db, mock := newSQLMock(t)
	now := time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT run_id FROM book_runs WHERE id=?`)).WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"run_id"}).AddRow(nil))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE book_runs SET status=?,request_id=?,error_message=?,started_at=?,finished_at=? WHERE id=?`)).WithArgs(StatusFailed, "execution-old", outcomeFailureMessage, now, now, int64(17)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,run_id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE id=?`)).WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"id", "run_id", "batch_project_id", "book_id", "status", "request_id", "error_message", "started_at", "finished_at", "created_at", "updated_at"}).AddRow(17, nil, 3, 11, StatusFailed, "execution-old", outcomeFailureMessage, now, now, now, now))
	mock.ExpectQuery("SELECT .* FROM stage_runs WHERE book_run_id=").WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"id", "book_run_id", "book_id", "stage", "status", "attempt", "request_id", "prompt_key", "prompt_version", "input_snapshot", "output_text", "error_message", "validation_result", "started_at", "finished_at", "created_at", "updated_at"}))
	s := NewService(NewMySQLStore(db), nil, func() time.Time { return now })
	cause := errors.New("provider-canary-short")
	result, err := s.failRun(context.Background(), BookRun{ID: 17, BatchProjectID: 3, BookID: 11, RequestID: "execution-old", StartedAt: &now}, cause)
	if err != cause || result.Error != outcomeFailureMessage || result.Run.ErrorMessage != outcomeFailureMessage {
		t.Fatalf("safe SQL write/internal cause: %#v %v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
