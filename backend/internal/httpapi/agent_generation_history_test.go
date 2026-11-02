package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func TestAgentGenerationHistoryListsOnlyOwnersMatchingResults(t *testing.T) {
	now := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	for _, project := range []agentworkspace.Project{
		generationHistoryProject("project-image", "image-1", "image", now),
		generationHistoryProject("project-image-second", "image-3", "image", now),
		generationHistoryProject("project-audio", "audio-1", "audio", now),
	} {
		if err := store.CreateProject(context.Background(), "alice", project); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.CreateProject(context.Background(), "bob", generationHistoryProject("project-other", "image-2", "image", now)); err != nil {
		t.Fatal(err)
	}
	seedGenerationHistoryTask(t, store, "alice", "project-image", "image-1", "image", "image-model", "image-result", now)
	seedGenerationHistoryTask(t, store, "alice", "project-image-second", "image-3", "image", "image-model", "image-archive", now.Add(30*time.Second))
	seedGenerationHistoryTask(t, store, "alice", "project-audio", "audio-1", "audio", "voice-model", "audio-result", now.Add(time.Minute))
	seedGenerationHistoryTask(t, store, "bob", "project-other", "image-2", "image", "foreign-model", "foreign-result", now.Add(2*time.Minute))

	mux := http.NewServeMux()
	RegisterAgentGenerationHistoryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })

	response := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/generation-history?kind=image&modelId=image-model&state=awaiting_confirmation&q=image-result", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("history=%d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Tasks []agentworkspace.GenerationTask `json:"tasks"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Tasks) != 1 || payload.Tasks[0].ID != "image-result" || payload.Tasks[0].ProjectID != "project-image" {
		t.Fatalf("tasks=%+v", payload.Tasks)
	}
}

func TestAgentGenerationHistoryIsMountedByApplicationRouter(t *testing.T) {
	now := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	project := generationHistoryProject("project-image", "image-1", "image", now)
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	seedGenerationHistoryTask(t, store, "alice", project.ID, "image-1", "image", "image-model", "image-result", now)

	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AgentWorkspace: store})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/agent/generation-history", nil)
	if response.Code != http.StatusOK || !containsText(response.Body.String(), "image-result") {
		t.Fatalf("history=%d %s", response.Code, response.Body.String())
	}
}

func TestAgentGenerationHistoryInsertsResultIntoCurrentCanvasWithLineage(t *testing.T) {
	now := time.Date(2026, 10, 24, 11, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	source := generationHistoryProject("project-source", "image-1", "image", now)
	target := generationHistoryProject("project-target", "text-1", "text", now)
	for _, project := range []agentworkspace.Project{source, target} {
		if err := store.CreateProject(context.Background(), "alice", project); err != nil {
			t.Fatal(err)
		}
	}
	seedGenerationHistoryTask(t, store, "alice", source.ID, "image-1", "image", "image-model", "source-task", now)
	assets := agentworkspace.NewMemoryAssetStore()
	if err := assets.CreateAsset(context.Background(), "alice", agentworkspace.Asset{
		ID: "source-task-asset", ProjectID: source.ID, Kind: "image", Name: "source.png", ContentType: "image/png",
		ObjectKey: "agent-assets/alice/project-source/source-task-asset.png", ByteSize: 4, CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterAgentGenerationHistoryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, assets, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-target/generation-history/source-task/results/source-task-candidate/insert", nil)
	if response.Code != http.StatusCreated {
		t.Fatalf("insert=%d %s", response.Code, response.Body.String())
	}
	canvas, err := store.Canvas(context.Background(), "alice", target.ID)
	if err != nil || len(canvas.Nodes) != 2 {
		t.Fatalf("canvas=%+v err=%v", canvas, err)
	}
	inserted := canvas.Nodes[1]
	if inserted.Type != "image" || inserted.AssetRef["originProjectId"] != source.ID || inserted.AssetRef["generationTaskId"] != "source-task" || inserted.AssetRef["generationResultId"] != "source-task-candidate" {
		t.Fatalf("node=%+v", inserted)
	}
	assetID, _ := inserted.AssetRef["id"].(string)
	imported, err := assets.Asset(context.Background(), "alice", target.ID, assetID)
	if err != nil || imported.ObjectKey != "agent-assets/alice/project-source/source-task-asset.png" {
		t.Fatalf("imported=%+v err=%v", imported, err)
	}
}

func generationHistoryProject(projectID, nodeID, nodeType string, now time.Time) agentworkspace.Project {
	canvas := agentworkspace.EmptyCanvas()
	canvas.Nodes = []agentworkspace.Node{{ID: nodeID, Type: nodeType, Title: nodeType, Version: 1, Status: "draft"}}
	return agentworkspace.Project{ID: projectID, Title: projectID, CreatedAt: now, UpdatedAt: now, Canvas: canvas}
}

func seedGenerationHistoryTask(t *testing.T, store *agentworkspace.MemoryStore, owner, projectID, nodeID, kind, modelID, taskID string, now time.Time) {
	t.Helper()
	task, err := store.CreateGenerationTask(context.Background(), owner, agentworkspace.CreateGenerationTaskInput{
		ID: taskID, ProjectID: projectID, NodeID: nodeID, NodeVersion: 1, Kind: kind, Provider: "fixture", ModelID: modelID,
		IdempotencyKey: taskID, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	running, err := store.TransitionGenerationTask(context.Background(), owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, Now: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.TransitionGenerationTask(context.Background(), owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{
		State: agentworkspace.GenerationAwaitingConfirmation, ExpectedUpdatedAt: running.UpdatedAt, Now: now.Add(2 * time.Second),
		Results: []agentworkspace.GenerationResult{{ID: taskID + "-candidate", AssetID: taskID + "-asset", Kind: kind}},
	})
	if err != nil {
		t.Fatal(err)
	}
}
