package aicut

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAICompatibleASRProviderSendsMediaAndParsesSegments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("authorization=%q", r.Header.Get("Authorization"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("model") != "whisper-large-v3" || r.FormValue("response_format") != "verbose_json" || r.FormValue("language") != "zh" {
			t.Fatalf("unexpected form: %#v", r.MultipartForm.Value)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		body, _ := io.ReadAll(file)
		if header.Filename != "dialogue.mp3" || string(body) != "audio-bytes" {
			t.Fatalf("file=%q body=%q", header.Filename, body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"segments": []map[string]any{
			{"start": 0.25, "end": 1.5, "text": "你好"},
			{"start": 1.5, "end": 3.0, "text": "欢迎回来"},
		}})
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleASRProvider(ASRProviderConfig{Endpoint: server.URL, APIKey: "secret", Model: "whisper-large-v3"}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	cues, err := provider.Transcribe(context.Background(), ASRInput{Filename: "dialogue.mp3", ContentType: "audio/mpeg", Language: "zh", Body: bytes.NewBufferString("audio-bytes")})
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 2 || cues[0].StartMS != 250 || cues[0].EndMS != 1500 || cues[1].Text != "欢迎回来" {
		t.Fatalf("cues=%#v", cues)
	}
}

func TestOpenAICompatibleASRProviderRejectsEmptyOrInvalidResponse(t *testing.T) {
	if _, err := NewOpenAICompatibleASRProvider(ASRProviderConfig{}, nil); err == nil {
		t.Fatal("empty provider config must fail")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "quota exhausted", http.StatusTooManyRequests)
	}))
	defer server.Close()
	provider, err := NewOpenAICompatibleASRProvider(ASRProviderConfig{Endpoint: server.URL, APIKey: "secret", Model: "whisper"}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	_, err = provider.Transcribe(context.Background(), ASRInput{Filename: "a.mp3", Body: strings.NewReader("x")})
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("err=%v", err)
	}
}

func TestASRProviderFromCatalogUsesEnabledSelectedModel(t *testing.T) {
	provider, err := ASRProviderFromCatalog(map[string]any{"modelCatalog": []any{
		map[string]any{"id": "off", "kind": "asr", "enabled": false},
		map[string]any{"id": "asr-main", "kind": "asr", "enabled": true, "baseUrl": "https://api.example.test/v1", "modelId": "whisper-large-v3", "credential": "secret"},
	}}, "asr-main", nil)
	if err != nil || provider == nil {
		t.Fatalf("provider=%#v err=%v", provider, err)
	}
	if provider.config.Endpoint != "https://api.example.test/v1/audio/transcriptions" {
		t.Fatalf("endpoint=%s", provider.config.Endpoint)
	}
	if _, err := ASRProviderFromCatalog(map[string]any{"modelCatalog": []any{}}, "missing", nil); err == nil {
		t.Fatal("missing selected model must fail")
	}
}

func TestMemoryStoreASRJobLifecycleAndCancellation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 11, 9, 10, 0, 0, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-1", Title: "测试", AspectRatio: "9:16", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job := ASRJob{ID: "asr-1", ProjectID: "cut-1", AssetID: "asset-1", ModelID: "asr-main", Language: "zh", ExpectedRevision: 0, State: ASRQueued, Progress: 0, CreatedAt: now, UpdatedAt: now}
	if err := store.CreateASRJob(ctx, "alice", job); err != nil {
		t.Fatal(err)
	}
	claimed, err := store.ClaimNextASRJob(ctx, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if claimed.State != ASRRunning || claimed.Progress != 5 {
		t.Fatalf("claimed=%#v", claimed)
	}
	if _, err := store.RequestASRCancel(ctx, "bob", "cut-1", "asr-1", now); err == nil {
		t.Fatal("other owner must not cancel")
	}
	canceled, err := store.RequestASRCancel(ctx, "alice", "cut-1", "asr-1", now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if canceled.State != ASRCancelRequested {
		t.Fatalf("state=%s", canceled.State)
	}
	if err := store.UpdateASRJob(ctx, canceled.WithState(ASRCanceled, 100, "", now.Add(3*time.Second))); err != nil {
		t.Fatal(err)
	}
	got, err := store.ASRJob(ctx, "alice", "cut-1", "asr-1")
	if err != nil || got.State != ASRCanceled {
		t.Fatalf("got=%#v err=%v", got, err)
	}
}
