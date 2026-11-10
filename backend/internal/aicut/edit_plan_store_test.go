package aicut

import (
	"context"
	"errors"
	"testing"
	"time"
)

func createEditPlanStoreFixture(t *testing.T) (*MemoryStore, EditPlan, time.Time) {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-1", Title: "测试", AspectRatio: "9:16", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	timeline := editPlanBaseTimeline()
	timeline.Revision = 0
	saved, err := store.SaveTimeline(ctx, "alice", "cut-1", timeline, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	plan := editPlanForApply(EditOperation{ID: "op-caption", Type: EditAddCaption, TrackID: "subtitle-1", StartMS: 100, DurationMS: 800, Text: "自动字幕", Reason: "补充字幕", Enabled: true})
	plan.ID, plan.ClientRequestID, plan.Owner, plan.ProjectID = "plan-1", "request-1", "alice", "cut-1"
	plan.BaseRevision, plan.State, plan.CreatedAt, plan.UpdatedAt = saved.Revision, EditPlanReady, now, now
	return store, plan, now
}

func TestMemoryEditPlanStoreIsIdempotentByClientRequest(t *testing.T) {
	store, plan, _ := createEditPlanStoreFixture(t)
	first, err := store.CreateEditPlan(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := plan
	duplicate.ID = "plan-duplicate"
	second, err := store.CreateEditPlan(context.Background(), duplicate)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate created another plan: first=%s second=%s", first.ID, second.ID)
	}
	byRequest, err := store.EditPlanByRequest(context.Background(), "alice", "cut-1", "request-1")
	if err != nil || byRequest.ID != first.ID {
		t.Fatalf("byRequest=%#v err=%v", byRequest, err)
	}
}

func TestMemoryEditPlanStoreCancelIsIdempotent(t *testing.T) {
	store, plan, now := createEditPlanStoreFixture(t)
	if _, err := store.CreateEditPlan(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	first, err := store.UpdateEditPlanState(context.Background(), "alice", "cut-1", plan.ID, EditPlanCanceled, "", now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.UpdateEditPlanState(context.Background(), "alice", "cut-1", plan.ID, EditPlanCanceled, "", now.Add(2*time.Second))
	if err != nil || second.State != EditPlanCanceled || first.ID != second.ID {
		t.Fatalf("first=%#v second=%#v err=%v", first, second, err)
	}
}

func TestMemoryEditPlanStoreAppliesExactlyOnceAndUndoRestoresTimeline(t *testing.T) {
	store, plan, now := createEditPlanStoreFixture(t)
	ctx := context.Background()
	if _, err := store.CreateEditPlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	stored, err := store.EditPlan(ctx, "alice", "cut-1", plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeEditPlan(enabledEditPlan(stored, []string{"op-caption"})); err != nil {
		t.Fatalf("stored plan became invalid before apply: %#v: %v", stored, err)
	}
	applied, timeline, summary, err := store.ApplyStoredEditPlan(ctx, "alice", "cut-1", plan.ID, []string{"op-caption"}, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if applied.State != EditPlanApplied || timeline.Revision != plan.BaseRevision+1 || summary.Applied != 1 {
		t.Fatalf("applied=%#v timeline=%#v summary=%#v", applied, timeline, summary)
	}
	again, againTimeline, againSummary, err := store.ApplyStoredEditPlan(ctx, "alice", "cut-1", plan.ID, []string{"op-caption"}, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if again.AppliedRevision != timeline.Revision || againTimeline.Revision != timeline.Revision || againSummary.Applied != summary.Applied {
		t.Fatalf("duplicate apply changed result: plan=%#v timeline=%#v summary=%#v", again, againTimeline, againSummary)
	}
	undone, _, err := store.UndoTimeline(ctx, "alice", "cut-1", timeline.Revision, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if undone.Revision != timeline.Revision+1 {
		t.Fatalf("undo revision=%d", undone.Revision)
	}
	for _, track := range undone.Tracks {
		for _, clip := range track.Clips {
			if clip.ID == "plan-op-caption" {
				t.Fatal("undo left the applied caption behind")
			}
		}
	}
}

func TestMemoryEditPlanStoreMarksChangedProjectStale(t *testing.T) {
	store, plan, now := createEditPlanStoreFixture(t)
	ctx := context.Background()
	if _, err := store.CreateEditPlan(ctx, plan); err != nil {
		t.Fatal(err)
	}
	current, err := store.Timeline(ctx, "alice", "cut-1")
	if err != nil {
		t.Fatal(err)
	}
	current.Script = "用户刚刚修改了工程"
	if _, err := store.SaveTimeline(ctx, "alice", "cut-1", current, current.Revision, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	stale, _, _, err := store.ApplyStoredEditPlan(ctx, "alice", "cut-1", plan.ID, []string{"op-caption"}, now.Add(2*time.Second))
	if !errors.Is(err, ErrRevisionConflict) || stale.State != EditPlanStale {
		t.Fatalf("stale=%#v err=%v", stale, err)
	}
	after, _ := store.Timeline(ctx, "alice", "cut-1")
	if after.Script != "用户刚刚修改了工程" || after.Revision != plan.BaseRevision+1 {
		t.Fatalf("timeline changed by stale apply: %#v", after)
	}
}
