package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
)

type memoryAudioObjects struct{ contents map[string][]byte }

type audioGeneratorFunc func(context.Context, string, agentworkspace.AudioGenerateInput) ([]agentworkspace.AudioOutput, error)

func (fn audioGeneratorFunc) Generate(ctx context.Context, owner string, input agentworkspace.AudioGenerateInput) ([]agentworkspace.AudioOutput, error) {
	return fn(ctx, owner, input)
}

func (s *memoryAudioObjects) Put(_ context.Context, asset agentworkspace.Asset, body io.Reader) error {
	if s.contents == nil {
		s.contents = map[string][]byte{}
	}
	s.contents[asset.ObjectKey], _ = io.ReadAll(body)
	return nil
}

func TestConfiguredAgentAudioGeneratorUsesExactSelectedModel(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer audio-secret" {
			t.Fatalf("authorization=%q", req.Header.Get("Authorization"))
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body["text"] != "雨夜环境声" || body["style"] != "sound_effect" {
			t.Fatalf("body=%+v err=%v", body, err)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("generated-audio"))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "audio-first", "kind": "audio", "enabled": true, "endpoint": "http://127.0.0.1:1", "credential": "wrong", "requestTemplate": `{"body":{"text":"{{input}}"}}`, "responseMapping": `{"resultUrl":"data.url"}`},
		map[string]any{"id": "audio-selected", "kind": "audio", "enabled": true, "endpoint": upstream.URL, "credential": "audio-secret", "requestTemplate": `{"body":{"text":"{{input}}","style":"{{style}}"}}`, "responseMapping": `{"resultUrl":"data.url"}`},
	}}); err != nil {
		t.Fatal(err)
	}
	generator := &configuredAgentAudioGenerator{catalog: store}
	outputs, err := generator.Generate(context.Background(), "alice", agentworkspace.AudioGenerateInput{Operation: agentworkspace.AudioSoundEffect, Prompt: "雨夜环境声", ModelID: "audio-selected", Settings: map[string]any{"speed": float64(1)}, Advanced: map[string]any{"emotion": "calm"}})
	if err != nil || len(outputs) != 1 || string(outputs[0].Bytes) != "generated-audio" {
		t.Fatalf("outputs=%+v err=%v", outputs, err)
	}
}

func TestConfiguredAgentAudioGeneratorUsesBuiltInTTS(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body["input"] != "画布 TTS 验收" {
			t.Fatalf("body=%+v err=%v", body, err)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("default-tts-audio"))
	}))
	defer upstream.Close()
	previous := defaultTTSEndpoint
	defaultTTSEndpoint = upstream.URL
	defer func() { defaultTTSEndpoint = previous }()

	generator := &configuredAgentAudioGenerator{catalog: novelfetchworkshop.NewMemoryStore()}
	outputs, err := generator.Generate(context.Background(), "alice", agentworkspace.AudioGenerateInput{
		Operation: agentworkspace.AudioSpeech, Prompt: "画布 TTS 验收", ModelID: builtInTTSModelID,
		Settings: map[string]any{"speed": float64(1)},
	})
	if err != nil || len(outputs) != 1 || string(outputs[0].Bytes) != "default-tts-audio" {
		t.Fatalf("outputs=%+v err=%v", outputs, err)
	}
}
func (s *memoryAudioObjects) Open(_ context.Context, asset agentworkspace.Asset) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.contents[asset.ObjectKey])), nil
}
func (s *memoryAudioObjects) Delete(_ context.Context, asset agentworkspace.Asset) error {
	delete(s.contents, asset.ObjectKey)
	return nil
}

type fixedAudioProcessor struct{}

func (fixedAudioProcessor) Process(_ context.Context, input agentworkspace.AudioProcessInput) ([]agentworkspace.AudioOutput, error) {
	return []agentworkspace.AudioOutput{{Name: "trimmed.mp3", ContentType: "audio/mpeg", Extension: "mp3", Bytes: append([]byte("trimmed:"), input.SourceBytes...)}}, nil
}

type countingAudioProcessor struct{ calls int }

func (processor *countingAudioProcessor) Process(_ context.Context, input agentworkspace.AudioProcessInput) ([]agentworkspace.AudioOutput, error) {
	processor.calls++
	return []agentworkspace.AudioOutput{{Name: "trimmed.mp3", ContentType: "audio/mpeg", Extension: "mp3", Bytes: append([]byte("trimmed:"), input.SourceBytes...)}}, nil
}

func TestAudioInputStringReturnsEmptyForMissingOrNilValue(t *testing.T) {
	if got := audioInputString(nil, "voice"); got != "" {
		t.Fatalf("nil map value=%q, want empty", got)
	}
	if got := audioInputString(map[string]any{"voice": nil}, "voice"); got != "" {
		t.Fatalf("nil field value=%q, want empty", got)
	}
}

func TestAgentAudioIdempotencyReplayDoesNotExecuteQueuedTaskAgain(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 30, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Title: "旁白", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryAudioObjects{contents: map[string][]byte{}}
	source := agentworkspace.Asset{ID: "audio-source", ProjectID: "project-1", Kind: "audio", Name: "source.mp3", ContentType: "audio/mpeg", ObjectKey: "agent-assets/alice/project-1/audio-source.mp3", ByteSize: 6, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", source); err != nil {
		t.Fatal(err)
	}
	objects.contents[source.ObjectKey] = []byte("source")

	existing, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "generation-existing", ProjectID: "project-1", NodeID: "audio-1", NodeVersion: 1,
		Kind: "audio", Provider: "local_ffmpeg", ModelID: "ffmpeg", Mode: "trim",
		Settings:          map[string]any{"parameters": map[string]any{"startMs": 1000, "endMs": 3000}, "sourceAssetId": source.ID},
		ReferenceAssetIDs: []string{source.ID}, IdempotencyKey: "audio-trim-replay", Now: now,
	})
	if err != nil || existing.State != agentworkspace.GenerationQueued {
		t.Fatalf("existing=%+v err=%v", existing, err)
	}

	processor := &countingAudioProcessor{}
	mux := http.NewServeMux()
	RegisterAgentAudioJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentAudioServices{
		Projects: store, Assets: assets, Objects: objects, Processor: processor, Now: func() time.Time { return now },
	})
	body := map[string]any{"nodeId": "audio-1", "nodeVersion": 1, "operation": "trim", "sourceAssetId": source.ID, "settings": map[string]any{"startMs": 1000, "endMs": 3000}, "idempotencyKey": "audio-trim-replay"}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/audio-generations", body)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if processor.calls != 0 {
		t.Fatalf("processor calls=%d, want 0 for idempotency replay", processor.calls)
	}
	if !containsText(response.Body.String(), "generation-existing") || !containsText(response.Body.String(), "queued") {
		t.Fatalf("replay body=%s", response.Body.String())
	}
}

func TestAgentAudioGenerationStoresResolvedCanvasInputSnapshot(t *testing.T) {
	now := time.Date(2026, 10, 23, 17, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Title: "旁白", Version: 1, Status: "draft"})
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes = append([]agentworkspace.Node{{ID: "text-1", Type: "text", Title: "文案", Summary: "上游旁白文案", Version: 1, Status: "ready"}}, canvas.Nodes...)
	canvas.Edges = []agentworkspace.Edge{{ID: "text-to-audio", Source: "text-1", Target: "audio-1"}}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, 0, now); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryAudioObjects{contents: map[string][]byte{}}
	mux := http.NewServeMux()
	RegisterAgentAudioJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentAudioServices{Projects: store, Assets: assets, Objects: objects, QueueAudio: true, Now: func() time.Time { return now }})

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/audio-generations", map[string]any{
		"nodeId": "audio-1", "nodeVersion": 1, "operation": "speech", "prompt": "本次 TTS 指令", "modelId": "built-in-tts", "settings": map[string]any{}, "idempotencyKey": "audio-input-snapshot",
	})
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	tasks, err := store.ListGenerationTasks(context.Background(), "alice", "project-1")
	if err != nil || len(tasks) != 1 {
		t.Fatalf("tasks=%+v err=%v", tasks, err)
	}
	snapshot, ok := tasks[0].Settings["resolvedInputs"].(map[string]any)
	if !ok || snapshot["prompt"] != "本次 TTS 指令" || snapshot["resolvedPrompt"] != "上游旁白文案" {
		t.Fatalf("snapshot=%+v", tasks[0].Settings["resolvedInputs"])
	}
}

func TestAgentAudioRetryKeepsOriginalResolvedCanvasInputSnapshot(t *testing.T) {
	now := time.Date(2026, 10, 23, 17, 10, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Title: "旁白", Version: 1, Status: "draft"})
	canvas, err := store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes = append([]agentworkspace.Node{{ID: "text-1", Type: "text", Title: "文案", Summary: "提交时的文案", Version: 1, Status: "ready"}}, canvas.Nodes...)
	canvas.Edges = []agentworkspace.Edge{{ID: "text-to-audio", Source: "text-1", Target: "audio-1"}}
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, 0, now); err != nil {
		t.Fatal(err)
	}
	original, err := store.CreateGenerationTask(context.Background(), "alice", agentworkspace.CreateGenerationTaskInput{
		ID: "audio-original", ProjectID: "project-1", NodeID: "audio-1", NodeVersion: 1, Kind: "audio", Provider: "personal_api", ModelID: "built-in-tts", Mode: "speech",
		Settings: map[string]any{"resolvedInputs": map[string]any{"prompt": "提交时的指令", "resolvedPrompt": "提交时的文案"}}, IdempotencyKey: "audio-original", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.TransitionGenerationTask(context.Background(), "alice", "project-1", original.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: "provider failed", Now: now.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	canvas, err = store.Canvas(context.Background(), "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	canvas.Nodes[0].Summary = "画布后来的文案"
	if _, err := store.SaveCanvas(context.Background(), "alice", "project-1", canvas, int64(canvas.Version), now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	retry, err := store.RetryGenerationTask(context.Background(), "alice", "project-1", original.ID, agentworkspace.RetryGenerationTaskInput{ID: "audio-retry", IdempotencyKey: "audio-retry", Now: now.Add(3 * time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, ok := retry.Settings["resolvedInputs"].(map[string]any)
	if !ok || snapshot["prompt"] != "提交时的指令" || snapshot["resolvedPrompt"] != "提交时的文案" {
		t.Fatalf("retry snapshot=%+v", retry.Settings["resolvedInputs"])
	}
}

func TestAgentAudioCoordinatorRunsQueuedTTS(t *testing.T) {
	now := time.Date(2026, 10, 23, 16, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Title: "旁白", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryAudioObjects{contents: map[string][]byte{}}
	called := 0
	services := AgentAudioServices{Projects: store, Assets: assets, Objects: objects, QueueAudio: true, Now: func() time.Time { return now }, Generator: audioGeneratorFunc(func(_ context.Context, _ string, input agentworkspace.AudioGenerateInput) ([]agentworkspace.AudioOutput, error) {
		called++
		if input.Operation != agentworkspace.AudioSpeech || input.Prompt != "画布后台 TTS" {
			t.Fatalf("input=%+v", input)
		}
		return []agentworkspace.AudioOutput{{Name: "tts.mp3", ContentType: "audio/mpeg", Extension: "mp3", Bytes: []byte("tts")}}, nil
	})}
	mux := http.NewServeMux()
	RegisterAgentAudioJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, services)
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/audio-generations", map[string]any{"nodeId": "audio-1", "nodeVersion": 1, "operation": "speech", "prompt": "画布后台 TTS", "modelId": "built-in-tts", "settings": map[string]any{}, "idempotencyKey": "queued-tts"})
	if response.Code != http.StatusAccepted || called != 0 || !containsText(response.Body.String(), "queued") {
		t.Fatalf("status=%d calls=%d body=%s", response.Code, called, response.Body.String())
	}
	coordinator := newAgentAudioCoordinator(services)
	if err := coordinator.pollPendingOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		tasks, _ := store.ListGenerationTasks(context.Background(), "alice", "project-1")
		if called == 1 && len(tasks) == 1 && tasks[0].State == agentworkspace.GenerationAwaitingConfirmation {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("audio task did not finish, calls=%d", called)
}

func TestAgentAudioTrimCreatesNewAssetWithoutChangingSource(t *testing.T) {
	now := time.Date(2026, 10, 21, 19, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Title: "旁白", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryAudioObjects{contents: map[string][]byte{}}
	source := agentworkspace.Asset{ID: "audio-source", ProjectID: "project-1", Kind: "audio", Name: "source.mp3", ContentType: "audio/mpeg", ObjectKey: "agent-assets/alice/project-1/audio-source.mp3", ByteSize: 6, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", source); err != nil {
		t.Fatal(err)
	}
	objects.contents[source.ObjectKey] = []byte("source")

	mux := http.NewServeMux()
	RegisterAgentAudioJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentAudioServices{
		Projects: store, Assets: assets, Objects: objects, Processor: fixedAudioProcessor{}, Now: func() time.Time { return now },
	})
	body := map[string]any{"nodeId": "audio-1", "nodeVersion": 1, "operation": "trim", "sourceAssetId": source.ID, "settings": map[string]any{"startMs": 1000, "endMs": 3000}, "idempotencyKey": "audio-trim-1"}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/project-1/audio-generations", body)
	if response.Code != http.StatusOK || !containsText(response.Body.String(), "awaiting_confirmation") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	items, err := assets.ListAssets(context.Background(), "alice", "project-1", "audio")
	if err != nil || len(items) != 2 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	var processed agentworkspace.Asset
	for _, item := range items {
		if item.ID != source.ID {
			processed = item
		}
	}
	if processed.ID == "" {
		t.Fatalf("new processed asset missing from %+v", items)
	}
	if got := string(objects.contents[source.ObjectKey]); got != "source" {
		t.Fatalf("source bytes changed to %q", got)
	}
	if got := string(objects.contents[processed.ObjectKey]); got != "trimmed:source" {
		t.Fatalf("processed bytes=%q", got)
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/agent/tasks/project-1/audio-generations", body)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign status=%d body=%s", foreign.Code, foreign.Body.String())
	}
}

func TestAgentAudioWaveformReturnsBoundedSummaryAndRejectsNonAudioAsset(t *testing.T) {
	now := time.Date(2026, 10, 24, 3, 0, 0, 0, time.UTC)
	store := newAgentNodeGenerationTestStore(t, now, agentworkspace.Node{ID: "audio-1", Type: "audio", Title: "旁白", Version: 1, Status: "draft"})
	assets := agentworkspace.NewMemoryAssetStore()
	objects := &memoryAudioObjects{contents: map[string][]byte{}}
	audio := agentworkspace.Asset{ID: "wave-audio", ProjectID: "project-1", Kind: "audio", Name: "voice.mp3", ContentType: "audio/mpeg", ObjectKey: "agent-assets/alice/project-1/wave-audio.mp3", ByteSize: 6, CreatedAt: now}
	image := agentworkspace.Asset{ID: "wave-image", ProjectID: "project-1", Kind: "image", Name: "cover.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/project-1/wave-image.png", ByteSize: 5, CreatedAt: now}
	for _, asset := range []agentworkspace.Asset{audio, image} {
		if err := assets.CreateAsset(context.Background(), "alice", asset); err != nil {
			t.Fatal(err)
		}
	}
	objects.contents[audio.ObjectKey] = []byte("audio!")
	objects.contents[image.ObjectKey] = []byte("image")
	mux := http.NewServeMux()
	RegisterAgentAudioJobRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, AgentAudioServices{
		Projects: store, Assets: assets, Objects: objects, Now: func() time.Time { return now },
		Waveform: agentAudioWaveformFunc(func(_ context.Context, _ agentworkspace.Asset, contents []byte) (AgentAudioWaveformSummary, error) {
			if string(contents) != "audio!" {
				t.Fatalf("contents=%q", contents)
			}
			return AgentAudioWaveformSummary{SampleRate: 8000, DurationMs: 1200, Peaks: []float64{0.1, 0.5, 0.2}}, nil
		}),
	})
	response := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/tasks/project-1/audio-assets/wave-audio/waveform", nil)
	if response.Code != http.StatusOK || containsText(response.Body.String(), "audio!") || !containsText(response.Body.String(), `"durationMs":1200`) || !containsText(response.Body.String(), `"peaks":[0.1,0.5,0.2]`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	nonAudio := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/tasks/project-1/audio-assets/wave-image/waveform", nil)
	if nonAudio.Code != http.StatusUnprocessableEntity {
		t.Fatalf("non-audio status=%d body=%s", nonAudio.Code, nonAudio.Body.String())
	}
}
