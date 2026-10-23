package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDirectorPlanRejectsForbiddenAndUnknownCommands(t *testing.T) {
	project := Project{ID: "project-1", Canvas: Canvas{Version: 1, Nodes: []Node{{ID: "text-1", Type: "text", Version: 1}}, Edges: []Edge{}}}
	for _, command := range []DirectorCommand{
		{Type: "delete_asset", NodeID: "asset-1"},
		{Type: "launch_magic", NodeID: "text-1"},
	} {
		err := ValidateDirectorPlan(project, DirectorPlan{ID: "plan-1", Commands: []DirectorCommand{command}})
		if !errors.Is(err, ErrDirectorCommandForbidden) && !errors.Is(err, ErrDirectorCommandUnknown) {
			t.Fatalf("command=%s err=%v", command.Type, err)
		}
	}
}

func TestDirectorPlanRejectsGenerationWithoutConfiguredModel(t *testing.T) {
	project := Project{ID: "project-1", Canvas: Canvas{Version: 1, Nodes: []Node{{ID: "image-1", Type: "image", Version: 1, Config: map[string]any{"prompt": "雨夜"}}}, Edges: []Edge{}}}
	err := ValidateDirectorPlan(project, DirectorPlan{ID: "plan-1", Commands: []DirectorCommand{{Type: DirectorCommandStartGeneration, NodeID: "image-1"}}})
	if !errors.Is(err, ErrDirectorMissingConfiguration) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyDirectorPlanAppliesStructuralCommandsOnce(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 21, 4, 30, 0, 0, time.UTC)
	store := NewMemoryStore()
	project := Project{
		ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now,
		Canvas: Canvas{Version: 1, Nodes: []Node{{ID: "text-1", Type: "text", Title: "旧标题", Version: 1, Status: "draft", Config: map[string]any{"prompt": "旧提示"}}}, Edges: []Edge{}},
	}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	plan := DirectorPlan{ID: "plan-1", Summary: "建立图片参考链", Commands: []DirectorCommand{
		{Type: DirectorCommandUpdateNode, NodeID: "text-1", Patch: map[string]any{"title": "开场文案", "summary": "雨夜开场"}},
		{Type: DirectorCommandCreateNode, Node: &Node{ID: "image-1", Type: "image", Position: Position{X: 420, Y: 120}, Title: "开场图", Version: 1, Status: "draft", Config: map[string]any{"prompt": "雨夜街道", "modelId": "image-model"}}},
		{Type: DirectorCommandConnectNodes, SourceNodeID: "image-1", TargetNodeID: "text-1", Relation: "reference"},
		{Type: DirectorCommandMoveNodes, NodeID: "text-1", Position: &Position{X: 120, Y: 80}},
	}}

	canvas, err := ApplyDirectorPlan(ctx, store, "alice", project.ID, 0, plan, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if canvas.Revision != 1 || len(canvas.Nodes) != 2 || len(canvas.Edges) != 1 {
		t.Fatalf("canvas=%+v", canvas)
	}
	if canvas.Nodes[0].Title != "开场文案" || canvas.Nodes[0].Summary != "雨夜开场" || canvas.Nodes[0].Version != 2 || canvas.Nodes[0].Position != (Position{X: 120, Y: 80}) {
		t.Fatalf("updated node=%+v", canvas.Nodes[0])
	}
	if canvas.Edges[0].Source != "image-1" || canvas.Edges[0].Target != "text-1" || canvas.Edges[0].Relation != "reference" {
		t.Fatalf("edge=%+v", canvas.Edges[0])
	}
	if _, err := ApplyDirectorPlan(ctx, store, "alice", project.ID, 0, plan, now.Add(2*time.Second)); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("second apply err=%v", err)
	}
}

func TestApplyDirectorPlanDoesNotPretendToRunOperationalCommands(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 21, 4, 40, 0, 0, time.UTC)
	store := NewMemoryStore()
	project := Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: Canvas{Version: 1, Nodes: []Node{{ID: "image-1", Type: "image", Version: 1, Config: map[string]any{"modelId": "image-model"}}}, Edges: []Edge{}}}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	_, err := ApplyDirectorPlan(ctx, store, "alice", project.ID, 0, DirectorPlan{ID: "plan-1", Commands: []DirectorCommand{{Type: DirectorCommandStartGeneration, NodeID: "image-1"}}}, now)
	if !errors.Is(err, ErrDirectorCommandNeedsExecution) {
		t.Fatalf("err=%v", err)
	}
}
