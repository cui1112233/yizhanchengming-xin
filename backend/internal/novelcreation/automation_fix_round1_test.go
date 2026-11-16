package novelcreation

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"sync"
	"testing"
	"time"
)

func TestAutomationFixRound1ScheduleWaitingDoesNotStarveLaterProject(t *testing.T) {
	for _, wait := range []string{"busy", "day_cap", "unknown"} {
		for _, paging := range []string{"memory", "sql_pages"} {
			t.Run(wait+"/"+paging, func(t *testing.T) {
				f := newServiceFixture(t, nil)
				ctx := context.Background()
				now := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)
				var later Schedule
				var plans []Schedule
				oldDue := map[string]time.Time{}
				for i := 0; i < 21; i++ {
					project, e := f.service.CreateProject(ctx, "alice", CreateProjectInput{Title: fmt.Sprintf("作品%d", i), Genre: "都市", ModelID: "text-1"})
					if e != nil {
						t.Fatal(e)
					}
					session, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: project.ID, Mode: "long_form_original"})
					if e != nil {
						t.Fatal(e)
					}
					p := Schedule{ID: fmt.Sprintf("plan-%02d", i), ProjectID: project.ID, SessionID: session.ID, Enabled: true, IntervalMinutes: 30, Prompt: "写下一章", ModelID: "text-1", NextRunAt: now.Add(time.Duration(i-30) * time.Minute), AutomationState: AutomationState{DailyChapterCap: 10, MaxConsecutiveFailures: 3}}
					if i < 20 {
						switch wait {
						case "busy":
							_, e = f.service.SendMessage(ctx, "alice", SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "先写", ModelID: "text-1", IdempotencyKey: fmt.Sprintf("busy-%d", i)})
							if e != nil {
								t.Fatal(e)
							}
						case "day_cap":
							p.DailySavedChapters = 10
							p.SuccessDay = now.In(shanghai).Format("2006-01-02")
						case "unknown":
							r, e := f.service.SendMessage(ctx, "alice", SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "先写", ModelID: "text-1", IdempotencyKey: fmt.Sprintf("unknown-%d", i)})
							if e != nil {
								t.Fatal(e)
							}
							f.store.UpdateRun(ctx, "alice", RunUpdate{ID: r.ID, Status: RunPaused, ErrorCode: ErrResultUnknown})
							execution, _, _ := f.store.GetExecution(ctx, "alice", r.ID)
							execution.State = "paused"
							f.store.PutExecution(ctx, execution)
							p.Enabled = false
							p.Status = "paused"
							p.ActiveRound = "due"
							p.Rounds = []AutomationRound{{Key: "due", RunID: r.ID, PlannedAt: p.NextRunAt}}
						}
						oldDue[p.ID] = p.NextRunAt
					}
					p, e = f.store.CreateSchedule(ctx, "alice", p)
					if e != nil {
						t.Fatal(e)
					}
					plans = append(plans, p)
					if i == 20 {
						later = p
					}
				}
				service := f.service
				var mock sqlmock.Sqlmock
				if paging == "sql_pages" {
					sqlStore, m := newNovelCreationMockStore(t)
					mock = m
					service = NewService(automationSQLPageStore{Store: f.store, pages: sqlStore}, f.runtime, f.runner)
				}
				for scan := 0; scan < 2; scan++ {
					if mock != nil {
						expectAutomationSchedulePages(mock, now, plans)
					}
					r, e := service.DispatchDueSchedules(ctx, now)
					if e != nil {
						t.Fatal(e)
					}
					if scan == 0 && (len(r) != 1 || r[0].Schedule.ID != later.ID) {
						t.Fatalf("21st eligible project starved behind %s: dispatched=%d", wait, len(r))
					}
					if scan == 1 && len(r) != 0 {
						t.Fatal("repeated scan dispatched again")
					}
				}
				if mock != nil {
					if e := mock.ExpectationsWereMet(); e != nil {
						t.Fatal(e)
					}
				}
				for id, due := range oldDue {
					p, _ := f.store.GetSchedule(ctx, "alice", id)
					if !p.NextRunAt.Equal(due) {
						t.Fatal("fairness changed business due", id)
					}
				}
			})
		}
	}
}

func TestAutomationFixRound1LegacyStopRestoresAndCancels(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		for _, confirmed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/confirmed=%v", kind, confirmed), func(t *testing.T) {
				f := newServiceFixture(t, nil)
				ctx := context.Background()
				source, e := f.service.SendMessage(ctx, "alice", f.sendInput("legacy"))
				if e != nil {
					t.Fatal(e)
				}
				active := source
				if confirmed {
					f.store.UpdateRun(ctx, "alice", RunUpdate{ID: source.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: []byte(`{"foundationPlan":{"title":"蓝图"}}`)})
					active, e = f.service.ConfirmRun(ctx, "alice", source.ID, []byte(foundationJSON))
					if e != nil {
						t.Fatal(e)
					}
				}
				if kind == "schedule" {
					p, e := f.store.CreateSchedule(ctx, "alice", Schedule{ProjectID: f.project.ID, SessionID: f.session.ID, Enabled: true, IntervalMinutes: 30, Prompt: "继续", ModelID: "text-1", LastRunID: source.ID, NextRunAt: time.Now().UTC().Add(time.Hour)})
					if e != nil {
						t.Fatal(e)
					}
					p, e = f.service.ControlSchedule(ctx, "alice", f.project.ID, p.ID, "stop")
					if e != nil {
						t.Fatal(e)
					}
					if p.ActiveRound == "" {
						t.Error("legacy schedule round not restored")
					}
				} else {
					p, e := f.store.CreateBatch(ctx, "alice", Batch{ProjectID: f.project.ID, SessionID: f.session.ID, Status: BatchRunning, TotalChapters: 2, CurrentRunID: source.ID, Prompt: "继续", ModelID: "text-1"})
					if e != nil {
						t.Fatal(e)
					}
					p, e = f.service.StopBatch(ctx, "alice", f.project.ID, p.ID)
					if e != nil {
						t.Fatal(e)
					}
					if p.ActiveRound == "" || p.CurrentRunID != active.ID {
						t.Error("legacy batch successor not restored")
					}
				}
				r, _ := f.store.GetRun(ctx, "alice", active.ID)
				execution, _, _ := f.store.GetExecution(ctx, "alice", active.ID)
				if r.Status != RunCancelled || !execution.CancelRequested {
					t.Fatalf("first control stop missed current run: %s %v", r.Status, execution.CancelRequested)
				}
			})
		}
	}
}

func TestAutomationFixRound1StopCompletedBatchReturnsExistingValue(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	p, e := f.store.CreateBatch(ctx, "alice", Batch{ProjectID: f.project.ID, SessionID: f.session.ID, Status: BatchCompleted, TotalChapters: 2, CompletedChapters: 2})
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := f.service.StopBatch(ctx, "alice", f.project.ID, p.ID)
			if e != nil || got.ID != p.ID || got.Status != BatchCompleted || got.CompletedChapters != 2 {
				t.Errorf("completed stop returned zero/stale batch: %+v %v", got, e)
			}
		}()
	}
	wg.Wait()
}

// SQL executes the pagination queries; project transactions and domain behavior
// use the same MemoryStore fixture. This verifies the service does not stop at
// a first MySQL page of 20 waiting plans, without claiming live MySQL coverage.
type automationSQLPageStore struct {
	Store
	pages Store
}

func (s automationSQLPageStore) ListDueSchedulePage(ctx context.Context, now time.Time, c ScheduleCursor, n int) ([]Schedule, error) {
	return s.pages.ListDueSchedulePage(ctx, now, c, n)
}
func automationScheduleRows(plans []Schedule) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id", "owner", "project", "session", "interval", "enabled", "prompt", "model", "thinking", "next", "last", "run", "created", "updated", "input"})
	for _, p := range plans {
		rows.AddRow(p.ID, p.OwnerUsername, p.ProjectID, p.SessionID, p.IntervalMinutes, p.Enabled, p.Prompt, p.ModelID, p.ThinkingLevel, p.NextRunAt, p.LastRunAt, p.LastRunID, p.CreatedAt, p.UpdatedAt, inputJSON(scheduleStoredInput(p)))
	}
	return rows
}
func expectAutomationSchedulePages(m sqlmock.Sqlmock, now time.Time, plans []Schedule) {
	order := regexp.QuoteMeta("ORDER BY s.next_run_at, BINARY s.owner_username, BINARY s.schedule_id LIMIT ?")
	m.ExpectQuery("SELECT .*"+order).WithArgs(now, 20).WillReturnRows(automationScheduleRows(plans[:20]))
	last := plans[19]
	m.ExpectQuery("SELECT .*"+regexp.QuoteMeta("AND (s.next_run_at > ? OR (s.next_run_at = ? AND (BINARY s.owner_username > ? OR (BINARY s.owner_username = ? AND BINARY s.schedule_id > ?))))")+".*"+order).WithArgs(now, last.NextRunAt, last.NextRunAt, last.OwnerUsername, last.OwnerUsername, last.ID, 20).WillReturnRows(automationScheduleRows(plans[20:]))
}

func TestAutomationFixRound1LegacySavedResultSurvivesStop(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			ctx := context.Background()
			r, e := f.service.SendMessage(ctx, "alice", f.sendInput("saved"))
			if e != nil {
				t.Fatal(e)
			}
			automationFinish(t, f, r, RunSucceeded, "", time.Now().UTC())
			if kind == "schedule" {
				p, e := f.store.CreateSchedule(ctx, "alice", Schedule{ProjectID: f.project.ID, SessionID: f.session.ID, Enabled: true, IntervalMinutes: 30, Prompt: "继续", ModelID: "text-1", LastRunID: r.ID, NextRunAt: time.Now().UTC().Add(time.Hour)})
				if e != nil {
					t.Fatal(e)
				}
				if _, e = f.service.ControlSchedule(ctx, "alice", f.project.ID, p.ID, "stop"); e != nil {
					t.Fatal(e)
				}
				for i := 0; i < 2; i++ {
					if _, e = f.service.DispatchDueSchedules(ctx, time.Now().UTC()); e != nil {
						t.Fatal(e)
					}
				}
				p, _ = f.store.GetSchedule(ctx, "alice", p.ID)
				if p.TotalSavedChapters != 1 || p.Status != "stopped" {
					t.Fatalf("saved schedule result lost/doubled %+v", p)
				}
			} else {
				p, e := f.store.CreateBatch(ctx, "alice", Batch{ProjectID: f.project.ID, SessionID: f.session.ID, Status: BatchRunning, TotalChapters: 2, CurrentRunID: r.ID, Prompt: "继续", ModelID: "text-1"})
				if e != nil {
					t.Fatal(e)
				}
				if _, e = f.service.StopBatch(ctx, "alice", f.project.ID, p.ID); e != nil {
					t.Fatal(e)
				}
				for i := 0; i < 2; i++ {
					if _, e = f.service.AdvanceBatches(ctx); e != nil {
						t.Fatal(e)
					}
				}
				p, _ = f.store.GetBatch(ctx, "alice", p.ID)
				if p.CompletedChapters != 1 || p.Status != BatchStopped {
					t.Fatalf("saved batch result lost/doubled %+v", p)
				}
			}
			run, _ := f.store.GetRun(ctx, "alice", r.ID)
			if run.Status != RunSucceeded {
				t.Fatal("saved run cancelled")
			}
		})
	}
}

type automationCompletionRaceStore struct {
	Store
	completed, release chan struct{}
}

func (s automationCompletionRaceStore) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, owner, project, func(tx Store) error {
		return fn(automationCompletionRaceStore{Store: tx, completed: s.completed, release: s.release})
	})
}
func (s automationCompletionRaceStore) UpdateBatch(ctx context.Context, owner string, p Batch) (Batch, error) {
	out, err := s.Store.UpdateBatch(ctx, owner, p)
	if err == nil && p.Status == BatchCompleted {
		close(s.completed)
		<-s.release
	}
	return out, err
}
func TestAutomationFixRound1CompletionWinsConcurrentStopReturnsSavedCount(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	p, e := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 1})
	if e != nil {
		t.Fatal(e)
	}
	runs, e := f.service.AdvanceBatches(ctx)
	if e != nil {
		t.Fatal(e)
	}
	automationFinish(t, f, runs[0].Run, RunSucceeded, "", time.Now().UTC())
	gate := automationCompletionRaceStore{Store: f.store, completed: make(chan struct{}), release: make(chan struct{})}
	s := NewService(gate, f.runtime, f.runner)
	advanced := make(chan error, 1)
	go func() { _, e := s.AdvanceBatches(ctx); advanced <- e }()
	<-gate.completed
	type stoppedResult struct {
		p Batch
		e error
	}
	stopped := make(chan stoppedResult, 1)
	go func() { p, e := s.StopBatch(ctx, "alice", f.project.ID, p.ID); stopped <- stoppedResult{p, e} }()
	close(gate.release)
	if e = <-advanced; e != nil {
		t.Fatal(e)
	}
	result := <-stopped
	if result.e != nil || result.p.ID != p.ID || result.p.Status != BatchCompleted || result.p.CompletedChapters != 1 {
		t.Fatalf("completion/stop lost persisted batch: %+v %v", result.p, result.e)
	}
}
