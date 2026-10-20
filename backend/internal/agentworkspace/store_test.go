package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStoreProjectMessagesAndOwnerIsolation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 18, 1, 2, 3, 0, time.UTC)
	project := Project{ID: "project-1", Title: "第一本书", TitleMode: "manual", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Project(ctx, "bob", project.ID); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("cross-owner read err=%v", err)
	}

	user := Message{Role: "user", Content: "做这本书", CreatedAt: now.Add(time.Second)}
	assistant := Message{Role: "assistant", Content: "先分析小说", CreatedAt: now.Add(2 * time.Second)}
	updated, err := store.AppendExchange(ctx, "alice", project.ID, user, assistant)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Messages) != 2 || updated.Preview != assistant.Content {
		t.Fatalf("updated=%+v", updated)
	}

	projects, err := store.ListProjects(ctx, "alice")
	if err != nil || len(projects) != 1 || projects[0].MessageCount != 2 {
		t.Fatalf("projects=%+v err=%v", projects, err)
	}
	if err := store.ClearMessages(ctx, "alice", project.ID, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	cleared, err := store.Project(ctx, "alice", project.ID)
	if err != nil || len(cleared.Messages) != 0 || cleared.Preview != "" {
		t.Fatalf("cleared=%+v err=%v", cleared, err)
	}
}

func TestMemoryStoreCanvasRevisionPreventsLostUpdates(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 18, 1, 2, 3, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}

	input := Canvas{
		Version:  1,
		Nodes:    []Node{{ID: "node-1", Type: "input", Title: "小说原文", Summary: "第一章", Position: Position{X: 120, Y: 80}, Version: 1, Status: "confirmed", GroupID: "group-1"}},
		Edges:    []Edge{},
		Viewport: map[string]any{"x": float64(10), "y": float64(20), "zoom": float64(1)},
	}
	saved, err := store.SaveCanvas(ctx, "alice", "project-1", input, 0, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != 1 || len(saved.Nodes) != 1 || saved.Viewport["zoom"] != float64(1) {
		t.Fatalf("saved=%+v", saved)
	}
	if saved.Nodes[0].GroupID != "group-1" {
		t.Fatalf("saved groupId=%q, want group-1", saved.Nodes[0].GroupID)
	}
	if _, err := store.SaveCanvas(ctx, "alice", "project-1", input, 0, now.Add(2*time.Second)); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale save err=%v", err)
	}
	loaded, err := store.Canvas(ctx, "alice", "project-1")
	if err != nil || loaded.Revision != 1 || loaded.Nodes[0].Title != "小说原文" || loaded.Nodes[0].GroupID != "group-1" {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
}

func TestMemoryStoreCanvasPreservesBusinessNodeConfiguration(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 19, 2, 3, 4, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-config", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}

	input := Canvas{Nodes: []Node{{
		ID: "video-1", Type: "video", Title: "分镜视频", Position: Position{X: 12, Y: 24}, Version: 1, Status: "draft",
		Config: map[string]any{"prompt": "雨夜追逐", "model": "seedance-2.0", "ratio": "9:16", "durationSeconds": float64(5)},
	}}, Edges: []Edge{}}
	if _, err := store.SaveCanvas(ctx, "alice", "project-config", input, 0, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Canvas(ctx, "alice", "project-config")
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Nodes[0].Config; got["prompt"] != "雨夜追逐" || got["model"] != "seedance-2.0" || got["ratio"] != "9:16" || got["durationSeconds"] != float64(5) {
		t.Fatalf("node config=%#v", got)
	}
}

func TestEmptyCanvasUsesPersistentAppearanceDefaults(t *testing.T) {
	appearance := EmptyCanvas().Appearance
	if appearance.EdgeColor != "coral" || !appearance.EdgeAnimated || appearance.EdgeWidth != "default" || !appearance.BackgroundDots || !appearance.AlignmentGuides {
		t.Fatalf("appearance=%+v", appearance)
	}
}

func TestMemoryStoreNormalizesInvalidCanvasAppearance(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 19, 1, 2, 3, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-appearance", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	saved, err := store.SaveCanvas(ctx, "alice", "project-appearance", Canvas{
		Appearance: Appearance{EdgeColor: "unknown", EdgeWidth: "huge"},
		Nodes:      []Node{},
		Edges:      []Edge{},
	}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Appearance.EdgeColor != "coral" || saved.Appearance.EdgeWidth != "default" || !saved.Appearance.EdgeAnimated || !saved.Appearance.BackgroundDots || !saved.Appearance.AlignmentGuides {
		t.Fatalf("saved appearance=%+v", saved.Appearance)
	}
}

func TestMemoryStoreDeleteIsOwnerScoped(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 18, 1, 2, 3, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteProject(ctx, "bob", "project-1", now); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("cross-owner delete err=%v", err)
	}
	if err := store.DeleteProject(ctx, "alice", "project-1", now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Project(ctx, "alice", "project-1"); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("deleted project err=%v", err)
	}
}
