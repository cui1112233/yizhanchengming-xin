package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestScriptShotStoreKeepsOrderedOwnerScopedShotsAndLineage(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "project", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(ctx, "alice", "project-1", "script-1", []ScriptShot{
		{ID: "shot-2", Sequence: 2},
		{ID: "shot-1", Sequence: 1},
	}); err != nil {
		t.Fatal(err)
	}
	otherOwner, err := store.ListShots(ctx, "bob", "project-1", "script-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(otherOwner) != 0 {
		t.Fatalf("other owner shots=%+v, want empty", otherOwner)
	}
	shots, err := store.ListShots(ctx, "alice", "project-1", "script-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(shots) != 2 || shots[0].ID != "shot-1" || shots[1].ID != "shot-2" {
		t.Fatalf("shots=%+v, want sequence order", shots)
	}
	if err := store.AttachGeneration(ctx, "alice", ShotGenerationLink{
		ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "shot-1", NodeID: "image-1", GenerationID: "gen-1", Attempt: 1,
	}); err != nil {
		t.Fatal(err)
	}
	links, err := store.ListGenerationLinks(ctx, "alice", "project-1", "shot-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].GenerationID != "gen-1" {
		t.Fatalf("links=%+v, want generation lineage", links)
	}
}

func TestScriptShotStoreUpdatesOnlyTheCurrentRevision(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "project", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(ctx, "alice", "project-1", "script-1", []ScriptShot{{
		ID: "shot-1", Sequence: 1, Scene: "雨夜街道", ImagePrompt: "wide shot",
	}}); err != nil {
		t.Fatal(err)
	}
	shot, err := store.ListShots(ctx, "alice", "project-1", "script-1")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateShot(ctx, "alice", ScriptShot{
		ID: "shot-1", ProjectID: "project-1", ScriptNodeID: "script-1", Sequence: 1, Scene: "雨夜小巷", ImagePrompt: "close shot", Revision: shot[0].Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision != shot[0].Revision+1 || updated.Scene != "雨夜小巷" {
		t.Fatalf("updated=%+v, want revised shot", updated)
	}
	_, err = store.UpdateShot(ctx, "alice", ScriptShot{
		ID: "shot-1", ProjectID: "project-1", ScriptNodeID: "script-1", Sequence: 1, Scene: "stale update", Revision: shot[0].Revision,
	})
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale update err=%v, want ErrRevisionConflict", err)
	}
}

func TestScriptShotStoreRejectsForeignProjectWritesAndMissingShotLineage(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "project", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(ctx, "bob", "project-1", "script-1", []ScriptShot{{ID: "shot-1", Sequence: 1}}); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("foreign replace err=%v, want ErrProjectNotFound", err)
	}
	if err := store.AttachGeneration(ctx, "alice", ShotGenerationLink{ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "missing", NodeID: "image-1", GenerationID: "gen-1", Attempt: 1}); !errors.Is(err, ErrScriptShotNotFound) {
		t.Fatalf("missing shot lineage err=%v, want ErrScriptShotNotFound", err)
	}
}

func TestScriptShotStoreCreatesNodeTaskAndLinkAtomically(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 22, 13, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	canvas := EmptyCanvas()
	canvas.Nodes = []Node{{ID: "script-1", Type: "script", Version: 1}}
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceShots(ctx, "alice", "project-1", "script-1", []ScriptShot{{ID: "shot-1", Sequence: 1}}); err != nil {
		t.Fatal(err)
	}
	input := CreateShotGenerationInput{
		ExpectedCanvasRevision: 0,
		Node:                   Node{ID: "image-1", Type: "image", Version: 1, Status: "queued", SourceNodeIDs: []string{"script-1"}},
		Edge:                   Edge{ID: "edge-1", Source: "script-1", Target: "image-1", Relation: "shot-generation"},
		Task:                   CreateGenerationTaskInput{ID: "generation-1", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "personal_api", ModelID: "image-model", Mode: "text_to_image", IdempotencyKey: "shot:shot-1:image:1", Now: now},
		Link:                   ShotGenerationLink{ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "shot-1", NodeID: "image-1", GenerationID: "generation-1", Attempt: 1, Active: true},
		Now:                    now,
	}
	createdCanvas, task, link, err := store.CreateShotGeneration(ctx, "alice", input)
	if err != nil || len(createdCanvas.Nodes) != 2 || task.ID != "generation-1" || link.GenerationID != task.ID {
		t.Fatalf("canvas=%+v task=%+v link=%+v err=%v", createdCanvas, task, link, err)
	}

	bad := input
	bad.Node.ID, bad.Edge.ID, bad.Edge.Target, bad.Task.ID, bad.Task.NodeID, bad.Task.IdempotencyKey = "image-2", "edge-2", "image-2", "generation-2", "image-2", "shot:missing:image:1"
	bad.Link = ShotGenerationLink{ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "missing", NodeID: "image-2", GenerationID: "generation-2", Attempt: 1}
	bad.ExpectedCanvasRevision = createdCanvas.Revision
	if _, _, _, err := store.CreateShotGeneration(ctx, "alice", bad); !errors.Is(err, ErrScriptShotNotFound) {
		t.Fatalf("bad err=%v", err)
	}
	after, _ := store.Canvas(ctx, "alice", "project-1")
	if len(after.Nodes) != 2 {
		t.Fatalf("failed atomic create persisted node: %+v", after.Nodes)
	}
	if _, err := store.GenerationTask(ctx, "alice", "project-1", "generation-2"); !errors.Is(err, ErrGenerationNotFound) {
		t.Fatalf("failed atomic create persisted task: %v", err)
	}
}
