package generation

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE book_runs SET status=?,request_id=?,error_message=?,started_at=?,finished_at=? WHERE id=?`)).WithArgs(StatusFailed, "execution-old", outcomeFailureMessage, now, now, int64(17)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id,batch_project_id,book_id,status,request_id,error_message,started_at,finished_at,created_at,updated_at FROM book_runs WHERE id=?`)).WithArgs(int64(17)).WillReturnRows(sqlmock.NewRows([]string{"id", "batch_project_id", "book_id", "status", "request_id", "error_message", "started_at", "finished_at", "created_at", "updated_at"}).AddRow(17, 3, 11, StatusFailed, "execution-old", outcomeFailureMessage, now, now, now, now))
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
