package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestAutomationSQLRoundUniqueClaimCommitsOrRollsBack(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		t.Run(map[bool]string{false: "new", true: "duplicate"}[duplicate], func(t *testing.T) {
			s, m := newNovelCreationMockStore(t)
			now := time.Date(2026, 11, 17, 1, 0, 0, 0, time.UTC)
			m.ExpectBegin()
			m.ExpectQuery("SELECT 1 FROM novel_projects.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
			expect := m.ExpectExec("INSERT INTO novel_automation_rounds").WithArgs("alice", "schedule", "plan", "2026-11-17T01:00:00Z", "r", now)
			if duplicate {
				expect.WillReturnError(&mysql.MySQLError{Number: 1062, Message: "duplicate planned round"})
				m.ExpectRollback()
			} else {
				expect.WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectCommit()
			}
			e := s.Atomic(context.Background(), "alice", "p", func(tx Store) error {
				return tx.InsertAutomationRound(context.Background(), "alice", "schedule", "plan", "2026-11-17T01:00:00Z", "r", now)
			})
			if duplicate && !IsCode(e, ErrPersistenceFailed) || !duplicate && e != nil {
				t.Fatal(e)
			}
			if e = m.ExpectationsWereMet(); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestAutomationSQLSavedChapterTruthAndStateRoundTrip(t *testing.T) {
	s, m := newNovelCreationMockStore(t)
	ctx := context.Background()
	m.ExpectQuery("SELECT 1 FROM novel_chapter_versions.*owner_username = \\? AND run_id = \\?").WithArgs("alice", "r").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	ok, e := s.HasSavedRun(ctx, "alice", "r")
	if e != nil || !ok {
		t.Fatal(e)
	}
	a := AutomationState{Status: "paused", DailyChapterCap: 10, MaxConsecutiveFailures: 3, ConsecutiveFailures: 2, TotalSavedChapters: 9, DailySavedChapters: 4, SuccessDay: "2026-11-17", StoppedReason: "用户暂停", Rounds: []AutomationRound{{Key: "due", RunID: "r", State: "saved"}}}
	input := AutomationInput{State: &a}
	m.ExpectQuery("SELECT .*FROM novel_schedules").WithArgs("alice", "plan").WillReturnRows(scheduleSQLRows(inputJSON(input)))
	p, e := s.GetSchedule(ctx, "alice", "plan")
	if e != nil || p.Status != "paused" || p.TotalSavedChapters != 9 || p.Rounds[0].RunID != "r" {
		t.Fatalf("state lost %+v %v", p, e)
	}
	raw, _ := json.Marshal(p.AutomationState)
	want, _ := json.Marshal(a)
	if string(raw) != string(want) {
		t.Fatal("state changed")
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

type automationFailUpdate struct {
	Store
	failure error
}

func (s automationFailUpdate) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(automationFailUpdate{tx, s.failure}) })
}
func (s automationFailUpdate) UpdateSchedule(context.Context, string, Schedule) (Schedule, error) {
	return Schedule{}, s.failure
}
func TestAutomationRoundAndOutboxRollbackTogether(t *testing.T) {
	f := newServiceFixture(t, nil)
	now := time.Now().UTC()
	p := automationSchedule(t, f, &now, 10)
	now = p.NextRunAt
	fail := errors.New("state update unavailable")
	s := NewService(automationFailUpdate{f.store, fail}, f.runtime, f.runner)
	if _, e := s.DispatchDueSchedules(context.Background(), now); !errors.Is(e, fail) {
		t.Fatal(e)
	}
	runs, _ := f.store.ListRuns(context.Background(), "alice", f.project.ID)
	messages, _ := f.store.ListMessages(context.Background(), "alice", f.session.ID, 0)
	if len(runs) != 0 || len(messages) != 0 || len(f.store.automationRounds) != 0 {
		t.Fatal("partial run/round escaped rollback")
	}
}
