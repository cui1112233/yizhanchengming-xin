package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"testing"
	"time"
)

func TestBlueprintPauseAndCallConfirmationAreIndependent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pause  *bool
		legacy bool
		want   bool
	}{{"legacy", nil, true, true}, {"content-on", boolPointer(true), false, true}, {"content-off-call-on", boolPointer(false), true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			ctx := context.Background()
			session, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_original", Title: "内容确认"})
			if e != nil {
				t.Fatal(e)
			}
			settings, _ := f.store.GetProjectSettings(ctx, "alice", f.project.ID)
			settings.Agent.PauseAfterBlueprint = tc.pause
			settings.Agent.RequireSecondaryConfirm = tc.legacy
			if _, e = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, settings); e != nil {
				t.Fatal(e)
			}
			in := f.sendInput(tc.name)
			in.SessionID = session.ID
			run, e := f.service.SendMessage(ctx, "alice", in)
			if e != nil {
				t.Fatal(e)
			}
			var snap struct {
				Checkpoint bool            `json:"checkpoint"`
				Settings   ProjectSettings `json:"project_settings"`
			}
			if e = json.Unmarshal(run.PromptSnapshot, &snap); e != nil {
				t.Fatal(e)
			}
			if snap.Checkpoint != tc.want || snap.Settings.Agent.RequireSecondaryConfirm != tc.legacy {
				t.Fatal("confirmation rules conflated", snap)
			}
			settings.Agent.PauseAfterBlueprint = boolPointer(!tc.want)
			f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, settings)
			stored, _ := f.store.GetRun(ctx, "alice", run.ID)
			var frozen map[string]any
			json.Unmarshal(stored.PromptSnapshot, &frozen)
			if frozen["checkpoint"] != tc.want {
				t.Fatal("old run was rewritten")
			}
		})
	}
}
func boolPointer(v bool) *bool { return &v }
func TestSchedulerStartContractReceiptAndControls(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	session, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_original", Title: "调度"})
	if e != nil {
		t.Fatal(e)
	}
	in := SchedulerStartInput{IdempotencyKey: "scheduler-1", SessionID: session.ID, ModelID: f.project.ModelID, IntervalSeconds: 1800, DailyChapterCap: 10, MaxConsecutiveFailures: 3}
	plan, e := f.service.StartScheduler(ctx, "alice", f.project.ID, in)
	if e != nil {
		t.Fatal(e)
	}
	if plan.IntervalMinutes != 30 || !plan.Enabled || plan.Prompt == "" || plan.NextRunAt.Before(time.Now()) {
		t.Fatal("invalid adapter", plan)
	}
	runs, _ := f.store.ListRuns(ctx, "alice", f.project.ID)
	if len(runs) != 0 {
		t.Fatal("start generated immediately")
	}
	again, e := f.service.StartScheduler(ctx, "alice", f.project.ID, in)
	if e != nil || again.ID != plan.ID {
		t.Fatal("replay duplicated", e)
	}
	other := in
	other.IdempotencyKey = "second-active-scheduler"
	if _, e = f.service.StartScheduler(ctx, "alice", f.project.ID, other); e == nil {
		t.Fatal("duplicate active scheduler allowed")
	}
	in.IntervalSeconds = 60
	if _, e = f.service.StartScheduler(ctx, "alice", f.project.ID, in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal("changed request accepted", e)
	}
	for _, action := range []string{"pause", "resume", "stop"} {
		v, e := f.service.ControlSchedule(ctx, "alice", f.project.ID, plan.ID, action)
		if e != nil {
			t.Fatal(action, e)
		}
		if action == "pause" && v.Status != "paused" || action == "stop" && v.Status != "stopped" {
			t.Fatal("fake control", v.Status)
		}
	}
	if _, e = f.service.StartScheduler(ctx, "bob", f.project.ID, in); e == nil {
		t.Fatal("owner leak")
	}
	for _, seconds := range []int{-1, 59, 86401} {
		bad := in
		bad.IdempotencyKey = "bad"
		bad.IntervalSeconds = seconds
		if _, e = f.service.StartScheduler(ctx, "alice", f.project.ID, bad); e == nil {
			t.Fatal("bad interval accepted", seconds)
		}
	}
	short, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "short_imitation", Title: "短篇"})
	if e != nil {
		t.Fatal(e)
	}
	in.IdempotencyKey = "short"
	in.SessionID = short.ID
	in.IntervalSeconds = 1800
	if _, e = f.service.StartScheduler(ctx, "alice", f.project.ID, in); e == nil {
		t.Fatal("short flow sent through chapter scheduler")
	}
	plans, _ := f.service.ListSchedules(ctx, "alice", f.project.ID)
	if len(plans) != 1 {
		t.Fatal("invalid requests left hidden plans", len(plans))
	}
}

func TestBlueprintAndProviderConfirmationExecutionAreSeparate(t *testing.T) {
	for _, tc := range []struct {
		name            string
		blueprint, call bool
		kind            string
		calls           int
	}{{"content-only", true, false, "", 1}, {"call-only", false, true, "next_model_call", 1}, {"neither", false, false, "", 6}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			f := newServiceFixture(t, nil)
			provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
			f.service.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil })
			session, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_original", Title: "执行确认"})
			if e != nil {
				t.Fatal(e)
			}
			settings, _ := f.store.GetProjectSettings(ctx, "alice", f.project.ID)
			settings.Agent = AgentSettings{Version: 1, PassScore: 80, TargetChapterWords: 2000, PauseAfterBlueprint: boolPointer(tc.blueprint), RequireSecondaryConfirm: tc.call}
			if _, e = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, settings); e != nil {
				t.Fatal(e)
			}
			in := f.sendInput(tc.name)
			in.SessionID = session.ID
			run, e := f.service.SendMessage(ctx, "alice", in)
			if e != nil {
				t.Fatal(e)
			}
			run, e = f.service.Run(ctx, "alice", run.ID)
			if e != nil {
				t.Fatal(e)
			}
			var snap map[string]any
			json.Unmarshal(run.RuntimeSnapshot, &snap)
			if len(provider.requests) != tc.calls {
				t.Fatal("wrong actual provider call count", len(provider.requests), tc.calls)
			}
			if tc.blueprint || tc.call {
				if run.Status != RunAwaitingConfirmation {
					t.Fatal("did not pause", run.Status)
				}
				if tc.call && snap["confirmationKind"] != tc.kind {
					t.Fatal("provider confirmation missing", snap)
				}
				if tc.blueprint && snap["confirmationKind"] == "next_model_call" {
					t.Fatal("content pause became provider approval")
				}
			} else if run.Status != RunSucceeded {
				t.Fatal("disabled content pause still blocked", run.Status)
			}
		})
	}
}

func TestDedicatedSchedulerDueDispatchSavesAndCountsExactlyOnce(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	f.service.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil })
	session, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_original", Title: "自动生产"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, PassScore: 80, TargetChapterWords: 2000, PauseAfterBlueprint: boolPointer(false)}})
	if e != nil {
		t.Fatal(e)
	}
	plan, e := f.service.StartScheduler(ctx, "alice", f.project.ID, SchedulerStartInput{IdempotencyKey: "due-flow", SessionID: session.ID, ModelID: f.project.ModelID, IntervalSeconds: 60})
	if e != nil {
		t.Fatal(e)
	}
	plan.NextRunAt = time.Now().UTC().Add(-time.Second)
	if _, e = f.store.UpdateSchedule(ctx, "alice", plan); e != nil {
		t.Fatal(e)
	}
	dispatched, e := f.service.DispatchDueSchedules(ctx, time.Now().UTC())
	if e != nil || len(dispatched) != 1 {
		t.Fatal("adapter did not dispatch", len(dispatched), e)
	}
	run, e := f.service.Run(ctx, "alice", dispatched[0].Run.ID)
	if e != nil || run.Status != RunSucceeded {
		t.Fatal("due model chain failed", run.Status, e)
	}
	again, e := f.service.DispatchDueSchedules(ctx, time.Now().UTC())
	if e != nil || len(again) != 0 {
		t.Fatal("same round dispatched twice", e)
	}
	saved, e := f.store.GetSchedule(ctx, "alice", plan.ID)
	if e != nil || saved.TotalSavedChapters != 1 || saved.DailySavedChapters != 1 {
		t.Fatal("saved output not accounted", saved, e)
	}
	chapters, e := f.store.ListChapters(ctx, "alice", f.project.ID)
	if e != nil || len(chapters) != 1 || len(provider.requests) != 6 {
		t.Fatal("missing or duplicate content", len(chapters), len(provider.requests), e)
	}
}
