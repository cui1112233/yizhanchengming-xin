package material

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

func TestStoreSavesMaterialAndProbeTaskInOneTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	item := Material{ID: "mcm_1", ProjectID: "mcp_1", Owner: "alice", OriginalName: "clip.mp4", MediaType: "video", StorageKey: `C:\data\mcm_1.mp4`, SizeBytes: 12, Metadata: Metadata{DurationMs: 1000}, Status: "ready", CreatedAt: now, UpdatedAt: now}
	probeTask := worktask.Task{ID: "mct_1", ProjectID: "mcp_1", Owner: "alice", Type: "material_probe", Status: worktask.StatusSucceeded, Progress: 1, CreatedAt: now, UpdatedAt: now}
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(insertMaterialSQL)).WithArgs(item.ID, item.ProjectID, item.Owner, item.OriginalName, item.MediaType, item.StorageKey, item.SizeBytes, sqlmock.AnyArg(), item.Status, item.CreatedAt, item.UpdatedAt).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(insertTaskSQL)).WithArgs(probeTask.ID, probeTask.ProjectID, probeTask.Owner, probeTask.Type, probeTask.Status, probeTask.Progress, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, nil, probeTask.CreatedAt, probeTask.UpdatedAt).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := NewStore(db).SaveImport(context.Background(), item, probeTask); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStoreListsOnlyMaterialsOwnedByRequestedAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(listMaterialsSQL)).WithArgs("alice", "mcp_1").WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "original_name", "media_type", "storage_key", "size_bytes", "metadata_json", "status", "created_at", "updated_at"}).AddRow("mcm_1", "mcp_1", "alice", "clip.mp4", "video", `C:\data\mcm_1.mp4`, 12, `{"durationMs":1000}`, "ready", now, now))
	items, err := NewStore(db).List(context.Background(), "alice", "mcp_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Owner != "alice" || items[0].Metadata.DurationMs != 1000 {
		t.Fatalf("items=%+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
