package export

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"yizhanchengming/magic-cut-workbench/internal/material"
	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

func TestSQLRepositoryCreatesExportAndTaskAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 8, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertExportTaskSQL)).WithArgs("mct_1", "mcp_1", "alice", "export_video", worktask.StatusPending, float64(0), sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertExportSQL)).WithArgs("mce_1", "mcp_1", "mct_1", "alice", "成片", "out.mp4", nil, nil, now).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	repository := NewSQLRepository(db)
	err = repository.Create(context.Background(), Export{ID: "mce_1", ProjectID: "mcp_1", TaskID: "mct_1", Owner: "alice", OutputName: "成片", StorageKey: "out.mp4", CreatedAt: now}, worktask.Task{ID: "mct_1", ProjectID: "mcp_1", Owner: "alice", Type: "export_video", Status: worktask.StatusPending, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryFinishesAttemptAndRecordsOutput(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 8, 1, 0, 0, time.UTC)
	metadata := material.Metadata{DurationMs: 1000, Width: 640, Height: 360, VideoCodec: "h264", AudioCodec: "aac"}
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(updateExportSuccessSQL)).WithArgs("out.mp4", sqlmock.AnyArg(), "abc123", "mce_1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(updateTaskSuccessSQL)).WithArgs(worktask.StatusSucceeded, float64(1), sqlmock.AnyArg(), now, "mct_1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(finishAttemptSQL)).WithArgs(worktask.StatusSucceeded, now, nil, nil, "mct_1", worktask.StatusRunning).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = NewSQLRepository(db).Succeed(context.Background(), "mce_1", "mct_1", Result{OutputPath: "out.mp4", SizeBytes: 42, SHA256: "abc123"}, metadata, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryGetsOnlyOwnedExportFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 8, 1, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(selectOwnedExportSQL)).WithArgs("mce_1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "task_id", "owner_username", "output_name", "storage_key", "media_metadata_json", "sha256", "created_at"}).AddRow("mce_1", "mcp_1", "mct_1", "alice", "成片", "out.mp4", []byte(`{"durationMs":1000}`), "abc123", now))
	item, err := NewSQLRepository(db).Get(context.Background(), "alice", "mce_1")
	if err != nil {
		t.Fatal(err)
	}
	if item.StorageKey != "out.mp4" || item.SHA256 != "abc123" || item.Metadata.DurationMs != 1000 {
		t.Fatalf("item=%+v", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
