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
