package novelcreation

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAutomationScheduleBusyRetainsPlannedDue(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, e := f.service.CreateSchedule(c, "alice", f.project.ID, ScheduleInput{SessionID: f.session.ID, IntervalMinutes: 30, Enabled: true, Prompt: "继续"})
	if e != nil {
		t.Fatal(e)
	}
	due := time.Now().UTC().Add(-time.Hour)
	p.NextRunAt = due
	f.store.UpdateSchedule(c, "alice", p)
	_, e = f.service.SendMessage(c, "alice", SendMessageInput{SessionID: f.session.ID, ProjectID: f.project.ID, ModelID: f.project.ModelID, Content: "占用", IdempotencyKey: "busy"})
	if e != nil {
		t.Fatal(e)
	}
	f.service.DispatchDueSchedules(c, time.Now().UTC())
	runs, _ := f.store.ListRuns(c, "alice", f.project.ID)
	messages, _ := f.store.ListMessages(c, "alice", f.session.ID, 0)
	if len(runs) != 1 || len(messages) != 1 {
		t.Fatalf("busy created extra runs/messages: %d/%d", len(runs), len(messages))
	}
	got, _ := f.store.GetSchedule(c, "alice", p.ID)
	if !got.NextRunAt.Equal(due) {
		t.Fatalf("busy advanced due: %v -> %v", due, got.NextRunAt)
	}
}
func TestAutomationScheduleImmediateRetainsPlannedDue(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, e := f.service.CreateSchedule(c, "alice", f.project.ID, ScheduleInput{SessionID: f.session.ID, IntervalMinutes: 30, Enabled: true, Prompt: "继续"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.service.RunSchedule(c, "alice", f.project.ID, p.ID); e != nil {
		t.Fatal(e)
	}
	got, _ := f.store.GetSchedule(c, "alice", p.ID)
	if !got.NextRunAt.Equal(p.NextRunAt) {
		t.Fatal("immediate execution changed normal due")
	}
}
func TestAutomationBatchOnlySavedSuccessCounts(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, e := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 2})
	if e != nil {
		t.Fatal(e)
	}
	runs, e := f.service.AdvanceBatches(c)
	if e != nil || len(runs) != 1 {
		t.Fatal(e)
	}
	f.store.UpdateRun(c, "alice", RunUpdate{ID: runs[0].Run.ID, Status: RunSucceeded})
	f.service.AdvanceBatches(c)
	got, _ := f.store.GetBatch(c, "alice", p.ID)
	if got.CompletedChapters != 0 {
		t.Fatal("status-only success counted without saved chapter")
	}
}

func automationSchedule(t *testing.T, f serviceFixture, now *time.Time, cap int) Schedule {
	t.Helper()
	f.service.automationClock = func() time.Time { return *now }
	p, e := f.service.CreateSchedule(context.Background(), "alice", f.project.ID, ScheduleInput{SessionID: f.session.ID, Enabled: true, Prompt: "继续", IntervalMinutes: 30, DailyChapterCap: cap})
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func automationFinish(t *testing.T, f serviceFixture, run Run, status RunStatus, code string, now time.Time) {
	t.Helper()
	finishQueuedRun(t, f.store, "alice", run, status, code)
	got, e := f.store.GetRun(context.Background(), "alice", run.ID)
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.store.UpdateRun(context.Background(), "alice", RunUpdate{ID: run.ID, Status: status, RuntimeSnapshot: got.RuntimeSnapshot, ErrorCode: code, ErrorMessage: code, FinishedAt: &now})
	if e != nil {
		t.Fatal(e)
	}
}
func TestAutomationDefaultsAndBoundaries(t *testing.T) {
	p, e := NormalizeScheduleInput(ScheduleInput{SessionID: "s", Prompt: "写"})
	if e != nil || p.IntervalMinutes != 30 || p.DailyChapterCap != 10 || p.MaxConsecutiveFailures != 3 {
		t.Fatalf("defaults=%+v %v", p, e)
	}
	b, e := NormalizeBatchInput(BatchInput{SessionID: "s"})
	if e != nil || b.TotalChapters != 5 || b.MaxConsecutiveFailures != 3 {
		t.Fatalf("defaults=%+v %v", b, e)
	}
	for _, n := range []int{1, 1440} {
		if _, e := NormalizeScheduleInput(ScheduleInput{SessionID: "s", Prompt: "写", IntervalMinutes: n, DailyChapterCap: 100, MaxConsecutiveFailures: 10}); e != nil {
			t.Fatal(e)
		}
	}
	for _, n := range []int{-1, 1441} {
		if _, e := NormalizeScheduleInput(ScheduleInput{SessionID: "s", Prompt: "写", IntervalMinutes: n}); e == nil {
			t.Fatalf("accepted interval %d", n)
		}
	}
	for _, n := range []int{-1, 101} {
		if _, e := NormalizeScheduleInput(ScheduleInput{SessionID: "s", Prompt: "写", DailyChapterCap: n}); e == nil {
			t.Fatal("accepted cap")
		}
	}
	for _, n := range []int{-1, 11} {
		if _, e := NormalizeBatchInput(BatchInput{SessionID: "s", MaxConsecutiveFailures: n}); e == nil {
			t.Fatal("accepted threshold")
		}
	}
}
func TestAutomationScheduleShanghaiDayCapAndPause(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	now := time.Date(2026, 11, 17, 1, 0, 0, 0, time.UTC)
	p := automationSchedule(t, f, &now, 10)
	for i := 0; i < 10; i++ {
		now = now.Add(31 * time.Minute)
		runs, e := f.service.DispatchDueSchedules(ctx, now)
		if e != nil || len(runs) != 1 {
			t.Fatalf("round %d %v %v", i, runs, e)
		}
		automationFinish(t, f, runs[0].Run, RunSucceeded, "", now)
		if r, e := f.service.DispatchDueSchedules(ctx, now); e != nil || len(r) != 0 {
			t.Fatal(e)
		}
	}
	got, _ := f.store.GetSchedule(ctx, "alice", p.ID)
	if got.TotalSavedChapters != 10 || got.DailySavedChapters != 10 || got.SuccessDay != "2026-11-17" {
		t.Fatalf("counts=%+v", got)
	}
	now = now.Add(31 * time.Minute)
	if r, e := f.service.DispatchDueSchedules(ctx, now); e != nil || len(r) != 0 {
		t.Fatal("cap dispatched", e)
	}
	got, _ = f.store.GetSchedule(ctx, "alice", p.ID)
	want := time.Date(2026, 11, 17, 16, 0, 0, 0, time.UTC)
	if got.EligibleAt == nil || !got.EligibleAt.Equal(want) {
		t.Fatalf("eligibility=%v", got.EligibleAt)
	}
	if _, e := f.service.RunSchedule(ctx, "alice", f.project.ID, p.ID); e == nil {
		t.Fatal("manual bypassed cap")
	}
	f.service.ControlSchedule(ctx, "alice", f.project.ID, p.ID, "pause")
	now = want.Add(time.Minute)
	if r, e := f.service.DispatchDueSchedules(ctx, now); e != nil || len(r) != 0 {
		t.Fatal("user pause resumed at midnight")
	}
	got, _ = f.store.GetSchedule(ctx, "alice", p.ID)
	if got.Enabled || got.Status != "paused" {
		t.Fatal(got.Status)
	}
	f.service.ControlSchedule(ctx, "alice", f.project.ID, p.ID, "resume")
	r, e := f.service.DispatchDueSchedules(ctx, now)
	if e != nil || len(r) != 1 {
		t.Fatalf("day recovery %v %v", r, e)
	}
	got, _ = f.store.GetSchedule(ctx, "alice", p.ID)
	if got.DailySavedChapters != 0 || got.TotalSavedChapters != 10 || got.SuccessDay != "2026-11-18" {
		t.Fatal("Shanghai day did not reset")
	}
}
func TestAutomationScheduleFailureThresholdAndCompletionTime(t *testing.T) {
	for _, code := range []string{ErrModelRateLimited, ErrModelAuthFailed, ErrResultUnknown} {
		t.Run(code, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			ctx := context.Background()
			now := time.Date(2026, 11, 17, 1, 0, 0, 0, time.UTC)
			p := automationSchedule(t, f, &now, 10)
			limit := 1
			if code == ErrModelRateLimited {
				limit = 3
			}
			for i := 0; i < limit; i++ {
				now = now.Add(2 * time.Hour)
				r, e := f.service.DispatchDueSchedules(ctx, now)
				if e != nil || len(r) != 1 {
					t.Fatalf("dispatch %v %v", r, e)
				}
				now = now.Add(45 * time.Minute)
				status := RunFailed
				if code == ErrResultUnknown {
					status = RunPaused
				}
				automationFinish(t, f, r[0].Run, status, code, now)
				f.service.DispatchDueSchedules(ctx, now)
				got, _ := f.store.GetSchedule(ctx, "alice", p.ID)
				if code != ErrResultUnknown && !got.NextRunAt.Equal(now.Add(30*time.Minute)) {
					t.Fatal("not completion + interval")
				}
			}
			got, _ := f.store.GetSchedule(ctx, "alice", p.ID)
			if got.Enabled || got.Status != "paused" || got.TotalSavedChapters != 0 {
				t.Fatalf("not paused %+v", got)
			}
			now = now.Add(48 * time.Hour)
			r, e := f.service.DispatchDueSchedules(ctx, now)
			if e != nil || len(r) != 0 {
				t.Fatal("paused dispatched")
			}
			if code == ErrResultUnknown {
				if _, e = f.service.ControlSchedule(ctx, "alice", f.project.ID, p.ID, "resume"); !IsCode(e, ErrResultUnknown) {
					t.Fatal("unknown result resumed", e)
				}
			}
		})
	}
}
func TestAutomationScheduleImmediateCompletionLeavesNormalDue(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	now := time.Date(2026, 11, 17, 1, 0, 0, 0, time.UTC)
	p := automationSchedule(t, f, &now, 10)
	run, e := f.service.RunSchedule(ctx, "alice", f.project.ID, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	now = now.Add(2 * time.Hour)
	automationFinish(t, f, run, RunSucceededWithWarnings, "", now)
	f.service.DispatchDueSchedules(ctx, now)
	got, _ := f.store.GetSchedule(ctx, "alice", p.ID)
	if !got.NextRunAt.Equal(p.NextRunAt) || got.TotalSavedChapters != 1 || !got.Rounds[0].Immediate {
		t.Fatalf("manual changed due %+v", got)
	}
	r, e := f.service.DispatchDueSchedules(ctx, now)
	if e != nil || len(r) != 1 {
		t.Fatalf("overdue normal round %v %v", r, e)
	}
	if len(got.Rounds) != 1 {
		t.Fatal("backlog storm")
	}
}
func TestAutomationBatchCheckpointAndSavePending(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, e := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 2})
	if e != nil {
		t.Fatal(e)
	}
	r, e := f.service.AdvanceBatches(c)
	if e != nil || len(r) != 1 {
		t.Fatal(e)
	}
	source := r[0].Run
	f.store.UpdateRun(c, "alice", RunUpdate{ID: source.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: []byte(`{"foundationPlan":{"title":"蓝图"}}`)})
	if r, e = f.service.AdvanceBatches(c); e != nil || len(r) != 0 {
		t.Fatal("checkpoint advanced")
	}
	next, e := f.service.ConfirmRun(c, "alice", source.ID, []byte(foundationJSON))
	if e != nil {
		t.Fatal(e)
	}
	exec, _, _ := f.store.GetExecution(c, "alice", next.ID)
	exec.State = "save_pending"
	f.store.PutExecution(c, exec)
	f.store.UpdateRun(c, "alice", RunUpdate{ID: next.ID, Status: RunFailed, ErrorCode: ErrPersistenceFailed})
	if r, e = f.service.AdvanceBatches(c); e != nil || len(r) != 0 {
		t.Fatal("save pending spawned run")
	}
	got, _ := f.store.GetBatch(c, "alice", p.ID)
	if got.CurrentRunID != next.ID || got.CompletedChapters != 0 {
		t.Fatalf("wrong successor %+v", got)
	}
	automationFinish(t, f, next, RunSucceeded, "", time.Now().UTC())
	r, e = f.service.AdvanceBatches(c)
	if e != nil || len(r) != 1 {
		t.Fatalf("next chapter %v %v", r, e)
	}
	got, _ = f.store.GetBatch(c, "alice", p.ID)
	if got.CompletedChapters != 1 || got.Rounds[0].SourceRunID != source.ID || got.Rounds[0].RunID != next.ID || got.Rounds[1].PreviousRunID != next.ID {
		t.Fatalf("source linkage %+v", got)
	}
}
func TestAutomationBatchFailureDoesNotCompleteAndBounded(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, _ := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 2})
	r, e := f.service.AdvanceBatches(c)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		automationFinish(t, f, r[0].Run, RunFailed, ErrModelRateLimited, time.Now().UTC())
		r, e = f.service.AdvanceBatches(c)
		if e != nil {
			t.Fatal(e)
		}
		got, _ := f.store.GetBatch(c, "alice", p.ID)
		if got.CompletedChapters != 0 || got.ConsecutiveFailures != i+1 {
			t.Fatalf("failure counted %+v", got)
		}
		if i < 2 && len(r) != 1 {
			t.Fatal("temporary failure not retried")
		}
	}
	got, _ := f.store.GetBatch(c, "alice", p.ID)
	if got.Status != BatchFailed || len(r) != 0 || got.StoppedReason == "" {
		t.Fatalf("unbounded %+v", got)
	}
	f.service.ResumeBatch(c, "alice", f.project.ID, p.ID)
	r, e = f.service.AdvanceBatches(c)
	if e != nil || len(r) != 1 {
		t.Fatal("resume failed", e)
	}
	if len(got.Rounds) != 3 {
		t.Fatal("lost failed attempts")
	}
}
func TestAutomationTwoServicesClaimAndAccountOnce(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			c := context.Background()
			now := time.Now().UTC()
			second := NewService(f.store, NewMemoryRuntime(), f.runner)
			var id string
			if kind == "schedule" {
				p := automationSchedule(t, f, &now, 10)
				now = p.NextRunAt
				id = p.ID
			} else {
				p, e := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 1})
				if e != nil {
					t.Fatal(e)
				}
				id = p.ID
			}
			scan := func(s *Service) {
				if kind == "schedule" {
					_, e := s.DispatchDueSchedules(c, now)
					if e != nil {
						t.Error(e)
					}
				} else {
					_, e := s.AdvanceBatches(c)
					if e != nil {
						t.Error(e)
					}
				}
			}
			var wg sync.WaitGroup
			for _, s := range []*Service{f.service, second} {
				wg.Add(1)
				go func(s *Service) { defer wg.Done(); scan(s) }(s)
			}
			wg.Wait()
			runs, _ := f.store.ListRuns(c, "alice", f.project.ID)
			if len(runs) != 1 {
				t.Fatalf("duplicate runs %d", len(runs))
			}
			automationFinish(t, f, runs[0], RunSucceeded, "", now)
			for _, s := range []*Service{f.service, second} {
				wg.Add(1)
				go func(s *Service) { defer wg.Done(); scan(s) }(s)
			}
			wg.Wait()
			if kind == "schedule" {
				p, _ := f.store.GetSchedule(c, "alice", id)
				if p.TotalSavedChapters != 1 || len(p.Rounds) != 1 {
					t.Fatalf("duplicate accounting %+v", p)
				}
			} else {
				p, _ := f.store.GetBatch(c, "alice", id)
				if p.CompletedChapters != 1 || len(p.Rounds) != 1 {
					t.Fatalf("duplicate accounting %+v", p)
				}
			}
		})
	}
}

func TestAutomationScheduleMidnightEligibilityDoesNotForceFutureDue(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	now := time.Date(2026, 11, 17, 15, 40, 0, 0, time.UTC)
	p := automationSchedule(t, f, &now, 1)
	r, e := f.service.RunSchedule(c, "alice", f.project.ID, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	automationFinish(t, f, r, RunSucceeded, "", now)
	f.service.DispatchDueSchedules(c, now)
	now = time.Date(2026, 11, 17, 16, 0, 0, 0, time.UTC)
	runs, e := f.service.DispatchDueSchedules(c, now)
	if e != nil || len(runs) != 0 {
		t.Fatal("midnight forced future due", e)
	}
	now = p.NextRunAt
	runs, e = f.service.DispatchDueSchedules(c, now)
	if e != nil || len(runs) != 1 {
		t.Fatal("future due did not dispatch", e)
	}
}
func TestAutomationScheduleLegacyInvalidValuesPreserved(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, e := f.store.CreateSchedule(c, "alice", Schedule{ProjectID: f.project.ID, SessionID: f.session.ID, Enabled: true, IntervalMinutes: 10080, Prompt: "继续", NextRunAt: time.Now().UTC().Add(-time.Hour), ModelID: "text-1"})
	if e != nil {
		t.Fatal(e)
	}
	r, e := f.service.DispatchDueSchedules(c, time.Now().UTC())
	if e != nil || len(r) != 0 {
		t.Fatal(e)
	}
	got, _ := f.store.GetSchedule(c, "alice", p.ID)
	if got.IntervalMinutes != 10080 || got.Enabled || got.ConfigurationError == "" {
		t.Fatalf("legacy clamped %+v", got)
	}
}
func TestAutomationStopCancelsConfirmedSuccessor(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			c := context.Background()
			now := time.Now().UTC()
			var id string
			var source Run
			if kind == "schedule" {
				p := automationSchedule(t, f, &now, 10)
				id = p.ID
				var e error
				source, e = f.service.RunSchedule(c, "alice", f.project.ID, id)
				if e != nil {
					t.Fatal(e)
				}
			} else {
				p, e := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID})
				if e != nil {
					t.Fatal(e)
				}
				id = p.ID
				r, e := f.service.AdvanceBatches(c)
				if e != nil {
					t.Fatal(e)
				}
				source = r[0].Run
			}
			f.store.UpdateRun(c, "alice", RunUpdate{ID: source.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: []byte(`{"foundationPlan":{"title":"蓝图"}}`)})
			next, e := f.service.ConfirmRun(c, "alice", source.ID, []byte(foundationJSON))
			if e != nil {
				t.Fatal(e)
			}
			if kind == "schedule" {
				_, e = f.service.ControlSchedule(c, "alice", f.project.ID, id, "stop")
			} else {
				_, e = f.service.StopBatch(c, "alice", f.project.ID, id)
			}
			if e != nil {
				t.Fatal(e)
			}
			run, _ := f.store.GetRun(c, "alice", next.ID)
			execution, _, _ := f.store.GetExecution(c, "alice", next.ID)
			if run.Status != RunCancelled || !execution.CancelRequested {
				t.Fatal("confirmed successor not cancelled")
			}
		})
	}
}
func TestAutomationBatchExecutionPausedRequiresConfirmation(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, _ := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID})
	r, e := f.service.AdvanceBatches(c)
	if e != nil {
		t.Fatal(e)
	}
	execution, _, _ := f.store.GetExecution(c, "alice", r[0].Run.ID)
	execution.State = "paused"
	f.store.PutExecution(c, execution)
	runs, e := f.service.AdvanceBatches(c)
	if e != nil || len(runs) != 0 {
		t.Fatal(e)
	}
	got, _ := f.store.GetBatch(c, "alice", p.ID)
	if got.Status != BatchFailed || got.StoppedReason == "" {
		t.Fatal("execution paused not explained")
	}
	if _, e = f.service.ResumeBatch(c, "alice", f.project.ID, p.ID); !IsCode(e, ErrResultUnknown) {
		t.Fatal("unknown ordinary resumed", e)
	}
}

func TestAutomationScheduleLateAccountingUsesSavedShanghaiDay(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	now := time.Date(2026, 11, 17, 15, 50, 0, 0, time.UTC)
	p := automationSchedule(t, f, &now, 10)
	run, e := f.service.RunSchedule(c, "alice", f.project.ID, p.ID)
	if e != nil {
		t.Fatal(e)
	}
	automationFinish(t, f, run, RunSucceeded, "", now)
	now = now.Add(20 * time.Minute)
	f.service.DispatchDueSchedules(c, now)
	got, _ := f.store.GetSchedule(c, "alice", p.ID)
	if got.TotalSavedChapters != 1 || got.DailySavedChapters != 0 || got.SuccessDay != "2026-11-18" {
		t.Fatalf("yesterday counted today %+v", got)
	}
}
func TestAutomationBatchSuccessResetsFailuresAndPreservesSavedOnStop(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, _ := f.service.CreateBatch(c, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 2})
	r, e := f.service.AdvanceBatches(c)
	if e != nil {
		t.Fatal(e)
	}
	automationFinish(t, f, r[0].Run, RunFailed, ErrModelRateLimited, time.Now().UTC())
	r, e = f.service.AdvanceBatches(c)
	if e != nil || len(r) != 1 {
		t.Fatal(e)
	}
	automationFinish(t, f, r[0].Run, RunSucceeded, "", time.Now().UTC())
	if _, e = f.service.StopBatch(c, "alice", f.project.ID, p.ID); e != nil {
		t.Fatal(e)
	}
	f.service.AdvanceBatches(c)
	got, _ := f.store.GetBatch(c, "alice", p.ID)
	if got.CompletedChapters != 1 || got.ConsecutiveFailures != 0 || got.Status != BatchStopped {
		t.Fatalf("saved success lost %+v", got)
	}
}
func TestAutomationBatchLegacyInvalidValuesPreserved(t *testing.T) {
	f := newServiceFixture(t, nil)
	c := context.Background()
	p, e := f.store.CreateBatch(c, "alice", Batch{ProjectID: f.project.ID, SessionID: f.session.ID, Status: BatchRunning, TotalChapters: 100, Prompt: "写", ModelID: "text-1"})
	if e != nil {
		t.Fatal(e)
	}
	r, e := f.service.AdvanceBatches(c)
	if e != nil || len(r) != 0 {
		t.Fatal(e)
	}
	got, _ := f.store.GetBatch(c, "alice", p.ID)
	if got.TotalChapters != 100 || got.Status != BatchPaused || got.StoppedReason == "" {
		t.Fatalf("invalid clamped %+v", got)
	}
}
