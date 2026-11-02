package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type fakeAgentVideoProcessor struct{}

func (fakeAgentVideoProcessor) Process(_ context.Context, input AgentVideoProcessInput) ([]AgentVideoProcessOutput, error) {
	return []AgentVideoProcessOutput{{Kind: input.Operation.ResultKind(), Name: "trimmed.mp4", ContentType: "video/mp4", Extension: "mp4", Bytes: []byte("new-video")}}, nil
}

func TestAgentVideoToolQueuesCandidateAndConfirmsIntoNewVideoNode(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "source-video-node", Type: "video", Version: 1, Status: "confirmed", AssetRef: map[string]any{"id": "source-video", "kind": "video"}})
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes = append(canvas.Nodes, agentworkspace.Node{ID: "result-video-node", Type: "video", Version: 1, Status: "draft"})
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, canvas.Revision, now); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryImageToolObjects{files: map[string][]byte{}}
	sourceBytes := []byte("source-video")
	source := agentworkspace.Asset{ID: "source-video", ProjectID: "project-1", Kind: "video", Name: "source.mp4", ContentType: "video/mp4", ObjectKey: "agent-assets/alice/project-1/source.mp4", ByteSize: int64(len(sourceBytes)), CreatedAt: now}
	if err := objects.Put(context.Background(), source, bytes.NewReader(sourceBytes)); err != nil {
		t.Fatal(err)
	}
	if err := assets.CreateAsset(context.Background(), "alice", source); err != nil {
		t.Fatal(err)
	}
	services := AgentVideoToolServices{Projects: store, Assets: assets, Objects: objects, Processor: fakeAgentVideoProcessor{}, Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentGenerationTaskRoutes(mux, auth, store, func() time.Time { return now })
	RegisterAgentVideoToolRoutes(mux, auth, services)

	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/video-tools", `{"nodeId":"result-video-node","nodeVersion":1,"operation":"trim","sourceAssetId":"source-video","settings":{"startMs":0,"endMs":1000},"idempotencyKey":"trim-source-video"}`)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	coordinator := newAgentVideoToolCoordinator(services)
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
		if err != nil || len(tasks) != 1 || tasks[0].State != agentworkspace.GenerationAwaitingConfirmation {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		if len(tasks[0].Results) != 1 || tasks[0].Results[0].Kind != "video" {
			t.Fatalf("task=%+v", tasks[0])
		}
		if !bytes.Equal(objects.files[source.ObjectKey], sourceBytes) {
			t.Fatal("source video bytes changed")
		}
		confirmed := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/generations/"+tasks[0].ID+"/confirm", `{"resultId":"`+tasks[0].Results[0].ID+`"}`)
		if confirmed.Code != http.StatusOK {
			t.Fatalf("confirm status=%d body=%s", confirmed.Code, confirmed.Body.String())
		}
		canvas, err := store.Canvas(context.Background(), "alice", "project-1")
		if err != nil {
			t.Fatal(err)
		}
		if canvas.Nodes[0].AssetRef["id"] != "source-video" || canvas.Nodes[1].AssetRef["id"] != tasks[0].Results[0].AssetID {
			t.Fatalf("canvas=%+v", canvas.Nodes)
		}
		return
	}
	t.Fatal("video tool task did not reach awaiting confirmation")
}
