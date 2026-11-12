package aicut

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestMySQLASRJobRetainsNamedTimeline(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	m.ExpectQuery("SELECT .* FROM aicut_asr_jobs WHERE id").WithArgs("asr", "p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "owner_username", "project_id", "asset_id", "model_id", "language", "expected_revision", "state", "progress", "error_message", "imported_count", "track_id", "created_at", "updated_at", "timeline_id"}).AddRow("asr", "alice", "p", "asset", "model", "zh", 4, ASRQueued, 0, "", 0, "", now, now, "tl_named"))
	job, e := NewMySQLStore(db).ASRJob(context.Background(), "alice", "p", "asr")
	if e != nil || job.TimelineID != "tl_named" || job.ExpectedRevision != 4 {
		t.Fatalf("job=%+v %v", job, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestMySQLASRJobInsertPersistsTarget(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	m.ExpectQuery("SELECT id, title, aspect_ratio").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "aspect_ratio", "background_mode", "background_color", "created_at", "updated_at"}).AddRow("p", "P", "9:16", "color", "#000000", now, now))
	m.ExpectExec("INSERT INTO aicut_asr_jobs .* timeline_id").WithArgs("job", "p", "alice", "audio", "model", "zh", int64(2), ASRQueued, 0, "", 0, "", now, now, "named").WillReturnResult(sqlmock.NewResult(1, 1))
	e = NewMySQLStore(db).CreateASRJob(t.Context(), "alice", ASRJob{ID: "job", ProjectID: "p", TimelineID: "named", AssetID: "audio", ModelID: "model", Language: "zh", ExpectedRevision: 2, State: ASRQueued, CreatedAt: now, UpdatedAt: now})
	if e != nil {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
