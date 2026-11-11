package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
)

func TestAICutTTSRouteCreatesDurableProjectAudioAsset(t *testing.T) {
	now := time.Date(2026, 11, 11, 8, 0, 0, 0, time.UTC)
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		var body ttsRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Input != "这是需要生成的旁白" || body.Voice != "zh-CN-YunxiNeural" || body.Speed != 1.2 || body.Style != "friendly" {
			t.Fatalf("unexpected TTS input: %+v", body)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("ID3-real-aicut-tts"))
	}))
	defer upstream.Close()
	previous := defaultTTSEndpoint
	defaultTTSEndpoint = upstream.URL
	defer func() { defaultTTSEndpoint = previous }()

	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-tts", Title: "配音工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	mux := http.NewServeMux()
	RegisterAICutTTSRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, func() time.Time { return now })

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-tts/tts", map[string]any{
		"text": "这是需要生成的旁白", "name": "第一段旁白", "voice": "zh-CN-YunxiNeural", "speed": 1.2, "pitch": 4, "style": "friendly",
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Asset aicut.Asset `json:"asset"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Asset.Kind != "audio" || payload.Asset.Name != "第一段旁白.mp3" || payload.Asset.ContentType != "audio/mpeg" || payload.Asset.ByteSize != int64(len("ID3-real-aicut-tts")) || payload.Asset.URL == "" {
		t.Fatalf("asset=%+v", payload.Asset)
	}
	stored, err := store.Asset(t.Context(), "alice", "cut-tts", payload.Asset.ID)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := objects.Open(t.Context(), aicutObjectAsset(stored))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	if string(data) != "ID3-real-aicut-tts" || calls.Load() != 1 {
		t.Fatalf("stored=%q calls=%d", data, calls.Load())
	}
}

func TestAICutTTSRouteValidatesBeforeCallingProvider(t *testing.T) {
	now := time.Date(2026, 11, 11, 8, 5, 0, 0, time.UTC)
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("ID3"))
	}))
	defer upstream.Close()
	previous := defaultTTSEndpoint
	defaultTTSEndpoint = upstream.URL
	defer func() { defaultTTSEndpoint = previous }()

	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-tts", Title: "配音工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutTTSRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, agentworkspace.NewLocalAssetObjectStore(t.TempDir()), nil, func() time.Time { return now })

	for _, test := range []struct {
		name string
		path string
		body map[string]any
		want string
	}{
		{name: "empty", path: "/api/ai-cut/projects/cut-tts/tts", body: map[string]any{"text": "  "}, want: "配音文本不能为空"},
		{name: "speed", path: "/api/ai-cut/projects/cut-tts/tts", body: map[string]any{"text": "测试", "speed": 9}, want: "语速必须在 0.5 到 2 之间"},
		{name: "pitch", path: "/api/ai-cut/projects/cut-tts/tts", body: map[string]any{"text": "测试", "pitch": 80}, want: "音调必须在 -50 到 50 之间"},
		{name: "missing project", path: "/api/ai-cut/projects/missing/tts", body: map[string]any{"text": "测试"}, want: "项目不存在"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, test.path, test.body)
			if response.Code != http.StatusBadRequest && response.Code != http.StatusNotFound {
				t.Fatalf("status=%d %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), test.want) {
				t.Fatalf("body=%s want %q", response.Body.String(), test.want)
			}
		})
	}
	if calls.Load() != 0 {
		t.Fatalf("provider calls=%d want 0", calls.Load())
	}
}

func TestAICutTTSRouteDoesNotCreateAssetWhenProviderFails(t *testing.T) {
	now := time.Date(2026, 11, 11, 8, 10, 0, 0, time.UTC)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
	}))
	defer upstream.Close()
	previous := defaultTTSEndpoint
	defaultTTSEndpoint = upstream.URL
	defer func() { defaultTTSEndpoint = previous }()

	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-tts", Title: "配音工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutTTSRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, agentworkspace.NewLocalAssetObjectStore(t.TempDir()), nil, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-tts/tts", map[string]any{"text": "失败测试"})
	if response.Code != http.StatusBadGateway || !strings.Contains(response.Body.String(), "配音生成失败") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assets, err := store.ListAssets(t.Context(), "alice", "cut-tts")
	if err != nil || len(assets) != 0 {
		t.Fatalf("assets=%+v err=%v", assets, err)
	}
}
