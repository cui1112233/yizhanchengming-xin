package aicut

import (
	"context"
	"errors"
	"testing"
)

func TestNamedEditPlanApplyIsAtomicIdempotentAndUndoable(t *testing.T) {
	base, plan, now := createEditPlanStoreFixture(t)
	collection := NewMemoryTimelineCollectionStore(base)
	a, _ := collection.Create(t.Context(), "alice", "cut-1", "A")
	_, _ = collection.Save(t.Context(), "alice", "cut-1", a.ID, Timeline{Tracks: []Track{{ID: "subtitle-1", Kind: TrackSubtitle, Clips: []Clip{}}}}, 0, now)
	plan.TimelineID = a.ID
	plan.BaseRevision = 1
	if _, e := base.CreateEditPlan(t.Context(), plan); e != nil {
		t.Fatal(e)
	}
	if _, _, _, e := base.ApplyStoredEditPlan(t.Context(), "alice", "cut-1", plan.ID, nil, now); !errors.Is(e, ErrInvalidEditPlan) {
		t.Fatalf("legacy apply named=%v", e)
	}
	stored, v, _, e := collection.ApplyStoredEditPlan(t.Context(), base, "alice", "cut-1", plan.ID, a.ID, nil, now)
	if e != nil || stored.State != EditPlanApplied || v.Revision != 2 || len(v.Tracks) != 1 {
		t.Fatalf("applied=%+v timeline=%+v %v", stored, v, e)
	}
	_, again, _, e := collection.ApplyStoredEditPlan(t.Context(), base, "alice", "cut-1", plan.ID, a.ID, nil, now)
	if e != nil || again.Revision != 2 {
		t.Fatalf("again=%+v %v", again, e)
	}
	d, _ := base.Timeline(context.Background(), "alice", "cut-1")
	if d.Revision != 1 || len(d.Tracks) != 4 {
		t.Fatalf("default changed=%+v", d)
	}
	undo, _, e := collection.Undo(t.Context(), "alice", "cut-1", a.ID, 2, now)
	if e != nil || len(undo.Tracks) != 1 || len(undo.Tracks[0].Clips) != 0 {
		t.Fatalf("undo=%+v %v", undo, e)
	}
	if _, _, _, e = collection.ApplyStoredEditPlan(t.Context(), base, "alice", "cut-1", plan.ID, DefaultTimelineID, nil, now); !errors.Is(e, ErrEditPlanNotFound) {
		t.Fatalf("wrong scope=%v", e)
	}
}
func TestEditPlanTargetIsImmutable(t *testing.T) {
	base, plan, _ := createEditPlanStoreFixture(t)
	plan.TimelineID = "target"
	stored, e := base.CreateEditPlan(t.Context(), plan)
	if e != nil {
		t.Fatal(e)
	}
	stored.TimelineID = "other"
	stored.Request.TimelineID = "other"
	if _, e = base.UpdateEditPlan(t.Context(), stored); !errors.Is(e, ErrInvalidEditPlan) {
		t.Fatalf("retarget=%v", e)
	}
}
