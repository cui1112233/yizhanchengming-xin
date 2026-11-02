package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestAgentCanvasInputPreviewReturnsResolvedPromptAndReferences(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "text-1", Type: "text", Version: 1, Status: "confirmed", Summary: "雨夜的街道"})
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes = append(canvas.Nodes,
		agentworkspace.Node{ID: "image-source", Type: "image", Version: 1, Status: "confirmed", AssetRef: map[string]any{"id": "asset-image-1", "kind": "image"}},
		agentworkspace.Node{ID: "image-target", Type: "image", Version: 1, Status: "draft", Config: map[string]any{"prompt": "生成电影海报"}},
	)
	canvas.Edges = []agentworkspace.Edge{
		{ID: "edge-text", Source: "text-1", Target: "image-target"},
		{ID: "edge-image", Source: "image-source", Target: "image-target"},
	}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, canvas.Revision, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{Projects: store, Now: func() time.Time { return now }})
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/canvas-inputs/preview", `{"nodeId":"image-target"}`)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Inputs agentworkspace.ResolvedCanvasInputs `json:"inputs"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if got, want := body.Inputs.Prompt, "雨夜的街道\n生成电影海报"; got != want {
		t.Fatalf("prompt=%q want=%q", got, want)
	}
	if got, want := body.Inputs.Images, []agentworkspace.ResolvedCanvasAsset{{NodeID: "image-source", AssetID: "asset-image-1", Kind: "image"}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("images=%#v want=%#v", got, want)
	}
	if got, want := body.Inputs.SourceNodeIDs, []string{"text-1", "image-source", "image-target"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("sources=%#v want=%#v", got, want)
	}
}

func TestResolveAgentCanvasInputCapabilityRejectsDisabledSelectedModel(t *testing.T) {
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "image-disabled", "kind": "image", "enabled": false},
	}}); err != nil {
		t.Fatal(err)
	}
	_, err := resolveAgentCanvasInputCapability(context.Background(), catalog, "alice", "image", "personal_api", "image-disabled")
	if err == nil || !strings.Contains(err.Error(), "当前模型不可用") {
		t.Fatalf("err=%v", err)
	}
}

func TestAgentCanvasInputPreviewRejectsUnsupportedAudioVideoLink(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Version: 1, Status: "confirmed", AssetRef: map[string]any{"id": "asset-audio-1", "kind": "audio"}})
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes = append(canvas.Nodes, agentworkspace.Node{ID: "video-1", Type: "video", Version: 1, Status: "draft", Config: map[string]any{"prompt": "做成视频"}})
	canvas.Edges = []agentworkspace.Edge{{ID: "edge-audio", Source: "audio-1", Target: "video-1"}}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, canvas.Revision, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{Projects: store, Now: func() time.Time { return now }})
	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/canvas-inputs/preview", `{"nodeId":"video-1"}`)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "音频参考素材") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
