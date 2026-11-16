package projectretention

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNovelArchiveUsesProjectLockDefersWriterAndKeepsManualMarker(t *testing.T) {
	for _, active := range []int{0, 1} {
		for _, automatic := range []bool{false, true} {
			t.Run(strings.Join([]string{map[int]string{0: "idle", 1: "active"}[active], map[bool]string{false: "manual", true: "auto"}[automatic]}, "_"), func(t *testing.T) {
				db, m, e := sqlmock.New()
				if e != nil {
					t.Fatal(e)
				}
				defer db.Close()
				s := NewService(db, nil, nil, nil, nil)
				m.ExpectBegin()
				m.ExpectQuery("SELECT title FROM novel_projects.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("作品"))
				if automatic {
					m.ExpectQuery("SELECT COUNT.*project_retention").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
					m.ExpectQuery("SELECT p.project_id,.*novel_content_records.*p.project_id=\\?").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "activity"}).AddRow("p", "作品", time.Now().AddDate(0, 0, -30)))
				}
				m.ExpectQuery("SELECT .*novel_runs.*novel_content_records").WithArgs("alice", "p", "alice", "p").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(active))
				if active == 1 {
					m.ExpectRollback()
				} else {
					if automatic {
						m.ExpectExec("UPDATE project_retention SET archived_at").WillReturnResult(sqlmock.NewResult(0, 1))
					} else {
						m.ExpectExec("INSERT INTO project_retention .*manual_archived.*manual_archived=1").WillReturnResult(sqlmock.NewResult(0, 1))
					}
					m.ExpectExec("UPDATE novel_projects SET status='archived'").WithArgs("alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
					m.ExpectExec("UPDATE novel_schedules SET enabled=0").WillReturnResult(sqlmock.NewResult(0, 1))
					m.ExpectExec("UPDATE novel_batches SET status='paused'").WillReturnResult(sqlmock.NewResult(0, 1))
					m.ExpectCommit()
				}
				e = s.archiveNovel(context.Background(), "alice", "p", automatic)
				if (active == 1 && !automatic) != (e != nil) {
					t.Fatalf("active=%d automatic=%v err=%v", active, automatic, e)
				}
				if e = m.ExpectationsWereMet(); e != nil {
					t.Fatal(e)
				}
			})
		}
	}
}
func TestNovelRestoreResetsClockAndDoesNotResumeAutomation(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewService(db, nil, nil, nil, func(string) int { return 7 })
	now := time.Date(2026, 11, 17, 1, 0, 0, 0, time.UTC)
	s.Now = func() time.Time { return now }
	m.ExpectBegin()
	m.ExpectQuery("SELECT status FROM novel_projects.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("archived"))
	m.ExpectExec("UPDATE project_retention SET last_activity_at=.*manual_archived=0").WithArgs(now, now.AddDate(0, 0, 7), now, "alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE novel_projects SET status='draft',updated_at=").WithArgs(now, "alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE novel_schedules SET enabled=0").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE novel_batches SET status='paused'").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	if e := s.Restore(context.Background(), "alice", ModuleNovelCreation, "p"); e != nil {
		t.Fatal(e)
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestNovelNewContentAfterRecomputeDefersExpiry(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewService(db, nil, nil, nil, nil)
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	s.Now = func() time.Time { return now }
	m.ExpectBegin()
	m.ExpectQuery("SELECT title FROM novel_projects.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow("作品"))
	m.ExpectQuery("SELECT COUNT.*project_retention").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	m.ExpectQuery("SELECT p.project_id,.*novel_content_records.*novel_materials.*novel_project_notes").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "activity"}).AddRow("p", "作品", now))
	m.ExpectExec("UPDATE project_retention SET last_activity_at=.*archived_at IS NULL").WithArgs(now, s.expiresAt(now, 7), now, "alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	if e := s.archiveNovel(context.Background(), "alice", "p", true); e != nil {
		t.Fatal(e)
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestNovelPurgeObjectsBeforeRecordsAndCoversWorkerTables(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewService(db, nil, nil, nil, nil)
	objects := 0
	s.NovelObjects = func(context.Context, string, string) error { objects++; return nil }
	m.ExpectQuery("SELECT archived_at FROM project_retention").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	expectNovelPurgeClaim(m)
	expectNovelPurgeFinish(m)
	if e := s.Purge(context.Background(), "alice", ModuleNovelCreation, "p"); e != nil {
		t.Fatal(e)
	}
	if objects != 1 {
		t.Fatal("objects omitted")
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
	raw, _ := os.ReadFile("novel.go")
	if !strings.Contains(string(raw), "'paused','awaiting_confirmation'") {
		t.Fatal("paused tasks must not block forever")
	}
}

func expectNovelPurgeClaim(m sqlmock.Sqlmock) {
	m.ExpectBegin()
	m.ExpectQuery("SELECT 1 FROM novel_projects.*status IN.*archived.*purging.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	m.ExpectQuery("SELECT archived_at.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	m.ExpectExec("UPDATE novel_projects SET status='purging'").WithArgs("alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
}
func TestNovelPurgeBlockedObjectsRejectRestoreAndFailureRetainsRetry(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	db.SetMaxOpenConns(1)
	s := NewService(db, nil, nil, nil, nil)
	entered, release := make(chan struct{}), make(chan struct{})
	calls := 0
	s.NovelObjects = func(context.Context, string, string) error {
		calls++
		if calls == 1 {
			close(entered)
			<-release
			return errors.New("fake TOS unavailable")
		}
		return nil
	}
	m.ExpectQuery("SELECT archived_at FROM project_retention").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	expectNovelPurgeClaim(m)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Purge(ctx, "alice", ModuleNovelCreation, "p") }()
	select {
	case <-entered:
	case err := <-done:
		t.Fatalf("purge did not reach objects: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	m.ExpectBegin()
	m.ExpectQuery("SELECT status FROM novel_projects.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("purging"))
	m.ExpectRollback()
	if err := s.Restore(ctx, "alice", ModuleNovelCreation, "p"); !errors.Is(err, ErrPurging) {
		t.Fatalf("restore accepted or opaque error during cleanup: %v", err)
	}
	close(release)
	if err := <-done; err == nil {
		t.Fatal("object failure swallowed")
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	m.ExpectQuery("SELECT archived_at FROM project_retention").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	expectNovelPurgeClaim(m)
	expectNovelPurgeFinish(m)
	if err := s.Purge(ctx, "alice", ModuleNovelCreation, "p"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("retry never revisited preserved object list: %d", calls)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectNovelPurgeFinish(m sqlmock.Sqlmock) {
	m.ExpectBegin()
	m.ExpectQuery("SELECT 1 FROM novel_projects.*status='purging'.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	m.ExpectQuery("SELECT archived_at.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	m.ExpectExec("DELETE FROM novel_automation_rounds.*SELECT run_id FROM novel_runs").WithArgs("alice", "alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
	for _, table := range []string{"novel_content_records", "novel_run_dispatch", "novel_project_slots", "novel_run_events", "novel_run_steps", "novel_runs", "novel_schedules", "novel_batches", "novel_messages", "novel_chapter_versions", "novel_audits", "novel_revisions", "novel_chapters", "novel_volumes", "novel_memories", "novel_foundations", "novel_project_settings", "novel_materials", "novel_project_notes", "novel_sessions", "novel_projects"} {
		m.ExpectExec("DELETE FROM "+table+" WHERE owner_username=\\? AND project_id=\\?").WithArgs("alice", "p").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	m.ExpectExec("DELETE FROM project_retention").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
}

func TestNovelArchivedListIdentifiesPurging(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewService(db, nil, nil, nil, nil)
	now := time.Now()
	m.ExpectQuery("SELECT module, project_id, title.*FROM project_retention").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"module", "project_id", "title", "last_activity_at", "retention_expires_at", "archived_at"}).AddRow("novel_creation", "p", "book", now, now, now))
	m.ExpectQuery("SELECT status FROM novel_projects").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("purging"))
	rows, e := s.Archived(context.Background(), "alice")
	if e != nil || len(rows) != 1 || !rows[0].Purging {
		t.Fatalf("rows=%+v err=%v", rows, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestNovelPurgeDatabaseFailureKeepsObjectListForRetry(t *testing.T) {
	db, m, _ := sqlmock.New()
	defer db.Close()
	s := NewService(db, nil, nil, nil, nil)
	calls := 0
	s.NovelObjects = func(context.Context, string, string) error { calls++; return nil }
	m.ExpectQuery("SELECT archived_at FROM project_retention").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	expectNovelPurgeClaim(m)
	m.ExpectBegin()
	m.ExpectQuery("SELECT 1 FROM novel_projects.*status='purging'.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	m.ExpectQuery("SELECT archived_at.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	m.ExpectExec("DELETE FROM novel_automation_rounds").WillReturnError(errors.New("fake database failure"))
	m.ExpectRollback()
	if e := s.Purge(context.Background(), "alice", ModuleNovelCreation, "p"); e == nil {
		t.Fatal("database failure swallowed")
	}
	m.ExpectQuery("SELECT archived_at FROM project_retention").WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(time.Now()))
	expectNovelPurgeClaim(m)
	expectNovelPurgeFinish(m)
	if e := s.Purge(context.Background(), "alice", ModuleNovelCreation, "p"); e != nil {
		t.Fatal(e)
	}
	if calls != 2 {
		t.Fatalf("lost retry metadata: calls=%d", calls)
	}
	if e := m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
