package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type memoryImageToolObjects struct {
	mu    sync.Mutex
	files map[string][]byte
}

func (s *memoryImageToolObjects) Put(_ context.Context, asset agentworkspace.Asset, body io.Reader) error {
	contents, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.files == nil {
		s.files = map[string][]byte{}
	}
	s.files[asset.ObjectKey] = append([]byte(nil), contents...)
	return nil
}

func (s *memoryImageToolObjects) Open(_ context.Context, asset agentworkspace.Asset) (io.ReadCloser, error) {
	s.mu.Lock()
	contents := append([]byte(nil), s.files[asset.ObjectKey]...)
	s.mu.Unlock()
	return io.NopCloser(bytes.NewReader(contents)), nil
}

func (s *memoryImageToolObjects) Delete(_ context.Context, asset agentworkspace.Asset) error {
	s.mu.Lock()
	delete(s.files, asset.ObjectKey)
	s.mu.Unlock()
	return nil
}

func testImageToolPNG(t *testing.T) []byte {
	t.Helper()
	canvas := image.NewRGBA(image.Rect(0, 0, 3, 2))
	canvas.Set(0, 0, color.RGBA{R: 255, A: 255})
	canvas.Set(2, 1, color.RGBA{B: 255, A: 255})
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func seedImageToolSource(t *testing.T, now time.Time, assets agentworkspace.AssetStore, objects *memoryImageToolObjects) agentworkspace.Asset {
	t.Helper()
	contents := testImageToolPNG(t)
	source := agentworkspace.Asset{ID: "source-image", ProjectID: "project-1", Kind: "image", Name: "source.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/project-1/source-image.png", ByteSize: int64(len(contents)), CreatedAt: now}
	if err := objects.Put(context.Background(), source, bytes.NewReader(contents)); err != nil {
		t.Fatal(err)
	}
	if err := assets.CreateAsset(context.Background(), "alice", source); err != nil {
		t.Fatal(err)
	}
	return source
}

func TestAgentImageToolRotateCreatesNewCandidateWithoutChangingSource(t *testing.T) {
	now := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Title: "画面", Version: 1, Status: "confirmed"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryImageToolObjects{}
	source := seedImageToolSource(t, now, assets, objects)
	sourceBytes := append([]byte(nil), objects.files[source.ObjectKey]...)
	services := AgentNodeGenerationServices{Projects: store, Assets: assets, Objects: objects, QueueImages: true, ImageProcess: NewAgentImageProcessor(), Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentGenerationTaskRoutes(mux, auth, store, func() time.Time { return now })
	RegisterAgentNodeGenerationRoutes(mux, auth, services)

	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-tools", `{"nodeId":"image-1","nodeVersion":1,"operation":"rotate","sourceAssetId":"source-image","settings":{"degrees":90},"idempotencyKey":"rotate-source-image"}`)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	coordinator := newAgentImageCoordinator(services)
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
		if err == nil && len(tasks) == 1 && tasks[0].State == agentworkspace.GenerationAwaitingConfirmation && len(tasks[0].Results) == 1 {
			result, err := assets.Asset(context.Background(), "alice", "project-1", tasks[0].Results[0].AssetID)
			if err != nil {
				t.Fatal(err)
			}
			body, err := objects.Open(context.Background(), result)
			if err != nil {
				t.Fatal(err)
			}
			decoded, _, err := image.Decode(body)
			body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if got := decoded.Bounds().Size(); got.X != 2 || got.Y != 3 {
				t.Fatalf("rotated size=%v want 2x3", got)
			}
			if result.ObjectKey == source.ObjectKey {
				t.Fatal("tool result overwrote source object key")
			}
			if got := objects.files[source.ObjectKey]; !bytes.Equal(got, sourceBytes) {
				t.Fatal("source bytes changed")
			}
			confirmed := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/generations/"+tasks[0].ID+"/confirm", `{"resultId":"`+tasks[0].Results[0].ID+`"}`)
			if confirmed.Code != http.StatusOK {
				t.Fatalf("confirm status=%d body=%s", confirmed.Code, confirmed.Body.String())
			}
			canvas, err := store.Canvas(context.Background(), "alice", "project-1")
			if err != nil || canvas.Nodes[0].AssetRef["assetId"] != result.ID || canvas.Nodes[0].Version != 2 {
				t.Fatalf("confirmed canvas=%+v err=%v", canvas, err)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	tasks, _ := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	encoded, _ := json.Marshal(tasks)
	t.Fatalf("image tool task did not finish: %s", encoded)
}

func TestAgentImageProcessorCropsMirrorsAndSplitsWithoutCallingModels(t *testing.T) {
	processor := NewAgentImageProcessor()
	source := testImageToolPNG(t)
	asset := agentworkspace.Asset{ID: "source-image", ProjectID: "project-1", Kind: "image", Name: "source.png", ContentType: "image/png"}
	for _, scenario := range []struct {
		name      string
		operation agentworkspace.ImageToolOperation
		settings  map[string]any
		count     int
		width     int
		height    int
	}{
		{name: "crop", operation: agentworkspace.ImageCrop, settings: map[string]any{"x": 1, "y": 0, "width": 2, "height": 1}, count: 1, width: 2, height: 1},
		{name: "mirror", operation: agentworkspace.ImageMirror, settings: map[string]any{"direction": "horizontal"}, count: 1, width: 3, height: 2},
		{name: "grid split", operation: agentworkspace.ImageGridSplit, settings: map[string]any{"columns": 2, "rows": 2}, count: 4, width: 1, height: 1},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			outputs, err := processor.Process(context.Background(), AgentImageProcessInput{Operation: scenario.operation, Source: asset, SourceBytes: source, Settings: scenario.settings})
			if err != nil || len(outputs) != scenario.count {
				t.Fatalf("outputs=%d err=%v", len(outputs), err)
			}
			decoded, _, err := image.Decode(bytes.NewReader(outputs[0].Bytes))
			if err != nil {
				t.Fatal(err)
			}
			if got := decoded.Bounds().Size(); got.X != scenario.width || got.Y != scenario.height {
				t.Fatalf("size=%v want=%dx%d", got, scenario.width, scenario.height)
			}
		})
	}
}

func TestAgentImageToolRejectsInvalidCropBeforeQueueing(t *testing.T) {
	now := time.Date(2026, 10, 24, 10, 10, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "image-1", Type: "image", Version: 1, Status: "confirmed"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryImageToolObjects{}
	seedImageToolSource(t, now, assets, objects)
	mux := http.NewServeMux()
	RegisterAgentNodeGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentNodeGenerationServices{Projects: store, Assets: assets, Objects: objects, QueueImages: true, ImageProcess: NewAgentImageProcessor(), Now: func() time.Time { return now }})

	recorder := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/project-1/image-tools", `{"nodeId":"image-1","nodeVersion":1,"operation":"crop","sourceAssetId":"source-image","settings":{"x":0,"y":0,"width":0,"height":2},"idempotencyKey":"invalid-crop"}`)
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(tasks) != 0 {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}
}
