package aicut

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestTimelineCollectionMySQLOwnerLockRejectsMissingParent(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	b := NewMemoryStore()
	ctx := context.Background()
	_ = b.CreateProject(ctx, "alice", Project{ID: "p", Title: "P"})
	s := NewMySQLTimelineCollectionStore(db, b)
	m.ExpectBegin()
	m.ExpectQuery("SELECT id FROM aicut_projects").WithArgs("p", "alice").WillReturnError(sql.ErrNoRows)
	m.ExpectRollback()
	if _, e = s.Create(ctx, "alice", "p", "New"); !errors.Is(e, ErrProjectNotFound) {
		t.Fatalf("missing parent=%v", e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestTimelineCollectionMySQLDoesNotLoadOtherSnapshots(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	b := NewMemoryStore()
	ctx := context.Background()
	_ = b.CreateProject(ctx, "alice", Project{ID: "p", Title: "P"})
	s := NewMySQLTimelineCollectionStore(db, b)
	m.ExpectBegin()
	m.ExpectQuery("SELECT id FROM aicut_projects").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p"))
	m.ExpectQuery("SELECT id, title, sort_index, created_at, updated_at FROM aicut_timelines").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "sort_index", "created_at", "updated_at"}))
	m.ExpectRollback()
	if _, e = s.Timeline(ctx, "alice", "p", "wrong"); !errors.Is(e, ErrTimelineNotFound) {
		t.Fatalf("missing timeline=%v", e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestTimelineCollectionMySQLSaveAtomicRevisionAndHistory(t *testing.T) {
	for _, stale := range []bool{false, true} {
		t.Run(map[bool]string{false: "save", true: "conflict"}[stale], func(t *testing.T) {
			db, m, e := sqlmock.New()
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			ctx := context.Background()
			base := NewMemoryStore()
			_ = base.CreateProject(ctx, "alice", Project{ID: "p", Title: "P"})
			s := NewMySQLTimelineCollectionStore(db, base)
			now := time.Date(2026, 11, 12, 1, 0, 0, 0, time.UTC)
			m.ExpectBegin()
			m.ExpectQuery("SELECT id FROM aicut_projects").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p"))
			m.ExpectQuery("SELECT id, title, sort_index, created_at, updated_at FROM aicut_timelines").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "sort_index", "created_at", "updated_at"}).AddRow("default", "Default", 0, now, now).AddRow("tl", "A", 1, now, now))
			revision := int64(0)
			if stale {
				revision = 7
			}
			m.ExpectQuery("SELECT revision, cursor_no, timeline_json FROM aicut_timelines").WithArgs("p", "alice", "tl").WillReturnRows(sqlmock.NewRows([]string{"revision", "cursor_no", "timeline_json"}).AddRow(revision, 0, `{"revision":0,"tracks":[]}`))
			m.ExpectQuery("SELECT sequence_no, timeline_json FROM aicut_named_timeline_history").WithArgs("p", "alice", "tl").WillReturnRows(sqlmock.NewRows([]string{"sequence_no", "timeline_json"}).AddRow(0, `{"revision":0,"tracks":[]}`))
			if stale {
				m.ExpectRollback()
			} else {
				m.ExpectExec("UPDATE aicut_timelines SET revision").WithArgs(int64(1), 1, `{"revision":1,"tracks":[],"script":"changed"}`, now, "p", "alice", "tl").WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectExec("DELETE FROM aicut_named_timeline_history").WithArgs("p", "alice", "tl", 2).WillReturnResult(sqlmock.NewResult(0, 0))
				m.ExpectExec("INSERT INTO aicut_named_timeline_history").WithArgs("p", "alice", "tl", 1, `{"revision":1,"tracks":[],"script":"changed"}`).WillReturnResult(sqlmock.NewResult(0, 1))
				m.ExpectCommit()
			}
			got, e := s.Save(ctx, "alice", "p", "tl", Timeline{Tracks: []Track{}, Script: "changed"}, 0, now)
			if stale {
				if !errors.Is(e, ErrRevisionConflict) {
					t.Fatalf("stale=%v", e)
				}
			} else if e != nil || got.Revision != 1 {
				t.Fatalf("saved=%+v %v", got, e)
			}
			if e = m.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}
