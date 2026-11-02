package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func generationTestProject(now time.Time) Project {
	canvas := EmptyCanvas()
	canvas.Nodes = []Node{{
		ID: "image-1", Type: "image", Title: "图片", Version: 1, Status: "draft",
	}}
	return Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: canvas}
}

func generationTestInput(now time.Time) CreateGenerationTaskInput {
	return CreateGenerationTaskInput{
		ID: "generation-1", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1,
		Kind: "image", Provider: "personal_api", ModelID: "image-model", Mode: "text_to_image",
		Settings: map[string]any{"ratio": "16:9", "count": 2}, ReferenceAssetIDs: []string{"asset-ref-1"},
		IdempotencyKey: "image-1-v1", Now: now,
	}
}

func TestMemoryStoreListsRunnableGenerationTasksByModesOldestFirst(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 23, 14, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", generationTestProject(now)); err != nil {
		t.Fatal(err)
	}
	for _, input := range []CreateGenerationTaskInput{
		{ID: "audio-old", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "audio", Provider: "tts", ModelID: "voice", Mode: "text_to_speech", IdempotencyKey: "audio-old", Now: now},
		{ID: "image-new", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "image", Provider: "image", ModelID: "model", Mode: "text_to_image", IdempotencyKey: "image-new", Now: now.Add(time.Second)},
		{ID: "video", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "video", Provider: "video", ModelID: "model", Mode: "text_to_video", IdempotencyKey: "video", Now: now.Add(2 * time.Second)},
	} {
		if _, err := store.CreateGenerationTask(ctx, "alice", input); err != nil {
			t.Fatal(err)
		}
	}
	running, err := store.GenerationTask(ctx, "alice", "project-1", "image-new")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.TransitionGenerationTask(ctx, "alice", "project-1", running.ID, GenerationTaskTransition{State: GenerationRunning, Now: now.Add(3 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err = store.CancelGenerationTask(ctx, "alice", "project-1", "audio-old", now.Add(4*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err = store.TransitionGenerationTask(ctx, "alice", "project-1", "image-new", GenerationTaskTransition{State: GenerationFailed, Now: now.Add(4*time.Second)}); err != nil {
		t.Fatal(err)
	}
	// Add another queued task after cancellation, preserving both runnable states.
	if _, err = store.CreateGenerationTask(ctx, "alice", CreateGenerationTaskInput{ID: "audio-queued", ProjectID: "project-1", NodeID: "image-1", NodeVersion: 1, Kind: "audio", Provider: "tts", ModelID: "voice", Mode: "text_to_speech", IdempotencyKey: "audio-queued", Now: now.Add(5 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	tasks, err := store.ListRunnableGenerationTasksByModes(ctx, []string{"text_to_speech", "text_to_image"}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != "audio-queued" || tasks[0].OwnerUsername != "alice" {
		t.Fatalf("tasks=%+v", tasks)
	}
}

func TestGenerationTaskTransitionsKeepMultipleDurableResults(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", generationTestProject(now)); err != nil {
		t.Fatal(err)
	}

	task, err := store.CreateGenerationTask(ctx, "alice", generationTestInput(now))
	if err != nil {
		t.Fatal(err)
	}
	if task.State != GenerationQueued {
		t.Fatalf("created state=%q, want queued", task.State)
	}
	if _, err := store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{State: GenerationCompleted, Now: now.Add(time.Second)}); !errors.Is(err, ErrGenerationState) {
		t.Fatalf("queued -> completed err=%v, want ErrGenerationState", err)
	}
	running, err := store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{State: GenerationRunning, Now: now.Add(2 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	results := []GenerationResult{
		{ID: "result-1", AssetID: "asset-1", Kind: "image", CreatedAt: now.Add(3 * time.Second)},
		{ID: "result-2", AssetID: "asset-2", Kind: "image", CreatedAt: now.Add(3 * time.Second)},
	}
	pending, err := store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{
		State: GenerationAwaitingConfirmation, Results: results, SelectedResultID: "result-2", ExpectedUpdatedAt: running.UpdatedAt, Now: now.Add(3 * time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending.Results) != 2 || pending.SelectedResultID != "result-2" || pending.Results[1].AssetID != "asset-2" {
		t.Fatalf("pending task lost result set: %+v", pending)
	}
}

func TestGenerationTaskConfirmRejectsChangedNodeAndAppliesSelectedResult(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 21, 11, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	project := generationTestProject(now)
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	task, err := store.CreateGenerationTask(ctx, "alice", generationTestInput(now))
	if err != nil {
		t.Fatal(err)
	}
	running, err := store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{State: GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{
		State:            GenerationAwaitingConfirmation,
		Results:          []GenerationResult{{ID: "result-1", AssetID: "asset-1", Kind: "image", Asset: map[string]any{"id": "asset-1", "kind": "image", "name": "第一张.png"}}, {ID: "result-2", AssetID: "asset-2", Kind: "image", Asset: map[string]any{"id": "asset-2", "kind": "image", "name": "第二张.png"}}},
		SelectedResultID: "result-2", ExpectedUpdatedAt: running.UpdatedAt, Now: now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}

	changed := project.Canvas
	changed.Nodes[0].Version = 2
	if _, err := store.SaveCanvas(ctx, "alice", project.ID, changed, 0, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ConfirmGenerationTask(ctx, "alice", project.ID, task.ID, "result-2", now.Add(4*time.Second)); !errors.Is(err, ErrGenerationNodeChanged) {
		t.Fatalf("confirm stale node err=%v, want ErrGenerationNodeChanged", err)
	}

	freshStore := NewMemoryStore()
	if err := freshStore.CreateProject(ctx, "alice", generationTestProject(now)); err != nil {
		t.Fatal(err)
	}
	freshTask, _ := freshStore.CreateGenerationTask(ctx, "alice", generationTestInput(now))
	freshRunning, _ := freshStore.TransitionGenerationTask(ctx, "alice", freshTask.ProjectID, freshTask.ID, GenerationTaskTransition{State: GenerationRunning, Now: now.Add(time.Second)})
	_, _ = freshStore.TransitionGenerationTask(ctx, "alice", freshTask.ProjectID, freshTask.ID, GenerationTaskTransition{
		State:            GenerationAwaitingConfirmation,
		Results:          []GenerationResult{{ID: "result-1", AssetID: "asset-1", Kind: "image", Asset: map[string]any{"id": "asset-1", "kind": "image", "name": "第一张.png"}}, {ID: "result-2", AssetID: "asset-2", Kind: "image", Asset: map[string]any{"id": "asset-2", "kind": "image", "name": "第二张.png"}}},
		SelectedResultID: "result-2", ExpectedUpdatedAt: freshRunning.UpdatedAt, Now: now.Add(2 * time.Second),
	})
	canvas, completed, err := freshStore.ConfirmGenerationTask(ctx, "alice", project.ID, freshTask.ID, "result-2", now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != GenerationCompleted || completed.SelectedResultID != "result-2" || len(completed.Results) != 2 {
		t.Fatalf("completed=%+v", completed)
	}
	if got := canvas.Nodes[0].AssetRef["id"]; got != "asset-2" {
		t.Fatalf("applied asset=%v, want asset-2", got)
	}
}

func TestGenerationTaskActivatesHistoricalResultWithExpectedNodeVersion(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 22, 8, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	project := generationTestProject(now)
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	task, err := store.CreateGenerationTask(ctx, "alice", generationTestInput(now))
	if err != nil {
		t.Fatal(err)
	}
	running, err := store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{State: GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.TransitionGenerationTask(ctx, "alice", task.ProjectID, task.ID, GenerationTaskTransition{
		State: GenerationAwaitingConfirmation, Results: []GenerationResult{
			{ID: "result-1", AssetID: "asset-1", Kind: "image", Asset: map[string]any{"id": "asset-1", "kind": "image"}},
			{ID: "result-2", AssetID: "asset-2", Kind: "image", Asset: map[string]any{"id": "asset-2", "kind": "image"}},
		}, SelectedResultID: "result-2", ExpectedUpdatedAt: running.UpdatedAt, Now: now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ConfirmGenerationTask(ctx, "alice", project.ID, task.ID, "result-2", now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}

	canvas, activated, err := store.ActivateGenerationResult(ctx, "alice", project.ID, task.ID, "result-1", 2, now.Add(4*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if canvas.Nodes[0].Version != 3 || canvas.Nodes[0].AssetRef["id"] != "asset-1" || activated.SelectedResultID != "result-1" {
		t.Fatalf("historical result not activated: canvas=%+v task=%+v", canvas, activated)
	}
	if _, _, err := store.ActivateGenerationResult(ctx, "alice", project.ID, task.ID, "result-2", 2, now.Add(5*time.Second)); !errors.Is(err, ErrGenerationNodeChanged) {
		t.Fatalf("stale activation err=%v, want ErrGenerationNodeChanged", err)
	}
}

func TestGenerationTaskIsOwnerScopedIdempotentCancelableAndRetryable(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", generationTestProject(now)); err != nil {
		t.Fatal(err)
	}
	first, err := store.CreateGenerationTask(ctx, "alice", generationTestInput(now))
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := store.CreateGenerationTask(ctx, "alice", generationTestInput(now))
	if err != nil || duplicate.ID != first.ID {
		t.Fatalf("duplicate=%+v err=%v, want idempotent task", duplicate, err)
	}
	if _, err := store.GenerationTask(ctx, "bob", first.ProjectID, first.ID); !errors.Is(err, ErrGenerationNotFound) {
		t.Fatalf("cross-owner get err=%v", err)
	}
	cancelled, err := store.CancelGenerationTask(ctx, "alice", first.ProjectID, first.ID, now.Add(time.Second))
	if err != nil || cancelled.State != GenerationCancelled {
		t.Fatalf("cancelled=%+v err=%v", cancelled, err)
	}
	retryInput := RetryGenerationTaskInput{ID: "generation-2", IdempotencyKey: "image-1-v1-retry", Now: now.Add(2 * time.Second)}
	retry, err := store.RetryGenerationTask(ctx, "alice", first.ProjectID, first.ID, retryInput)
	if err != nil {
		t.Fatal(err)
	}
	if retry.State != GenerationQueued || retry.RetryOfTaskID != first.ID || retry.ID == first.ID {
		t.Fatalf("retry=%+v", retry)
	}
}
