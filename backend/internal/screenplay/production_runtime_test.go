package screenplay

import (
	"context"
	"testing"
)

func TestProductionOwnerIsolationAndPause(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "alice"})
	r := &ProductionRuntime{Store: s, Repository: NewMemoryProductionRepository(), RedisDo: func(context.Context, ...string) (any, error) { return "PONG", nil }, Executor: func(string) BatchExecutor { return BatchExecutor{} }}
	if _, e := r.Start(ctx, "mallory", "p"); e == nil {
		t.Fatal("foreign project accepted")
	}
	if _, e := r.Start(ctx, "alice", "p"); e != nil {
		t.Fatal(e)
	}
	j, e := r.Pause(ctx, "alice", "p")
	if e != nil || j.Status != "paused" {
		t.Fatalf("%+v %v", j, e)
	}
}
func TestProductionProgressCannotOverwritePause(t *testing.T) {
	ctx := context.Background()
	r := NewMemoryProductionRepository()
	_ = r.SetProductionState(ctx, "a", "p", "running", "")
	_ = r.SetProductionState(ctx, "a", "p", "paused", "")
	_ = r.SaveProductionProgress(ctx, ProductionJob{Owner: "a", ProjectID: "p", Status: "completed", Units: 90})
	j, _ := r.GetProductionJob(ctx, "a", "p")
	if j.Status != "paused" {
		t.Fatal(j)
	}
}
