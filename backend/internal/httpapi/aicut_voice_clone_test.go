package httpapi

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestAICutVoiceCloneCreatesDurableAudioAsset(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	store, objects, reference := seedAICutVoiceReference(t, now, "audio")
	var calls atomic.Int32
	cloner := func(ctx context.Context, owner string, asset aicut.Asset, referenceBytes []byte, input aicutVoiceCloneRequest) ([]byte, string, error) {
		calls.Add(1)
		if owner != "alice" || asset.ID != reference.ID || string(referenceBytes) != "ID3-reference" || input.Text != "这是克隆音色测试" {
			t.Fatalf("unexpected clone input owner=%q asset=%+v ref=%q input=%+v", owner, asset, referenceBytes, input)
		}
		return []byte("ID3-cloned-voice"), "audio/mpeg", nil
	}
	mux := http.NewServeMux()
	RegisterAICutVoiceCloneRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, cloner, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-voice/voice-clone", map[string]any{"assetId": "voice-reference", "text": "这是克隆音色测试", "name": "角色克隆音色"})
	if response.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Asset aicut.Asset `json:"asset"`
	}
	decodeTestJSON(t, response.Body.Bytes(), &payload)
	if payload.Asset.Kind != "audio" || payload.Asset.Name != "角色克隆音色.mp3" || payload.Asset.URL == "" || calls.Load() != 1 {
		t.Fatalf("asset=%+v calls=%d", payload.Asset, calls.Load())
	}
	stored, err := store.Asset(t.Context(), "alice", "cut-voice", payload.Asset.ID)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := objects.Open(t.Context(), aicutObjectAsset(stored))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	if string(data) != "ID3-cloned-voice" {
		t.Fatalf("stored=%q", data)
	}
}

func TestAICutVoiceCloneValidatesBeforeProvider(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 5, 0, 0, time.UTC)
	store, objects, _ := seedAICutVoiceReference(t, now, "video")
	var calls atomic.Int32
	cloner := func(context.Context, string, aicut.Asset, []byte, aicutVoiceCloneRequest) ([]byte, string, error) {
		calls.Add(1)
		return []byte("ID3"), "audio/mpeg", nil
	}
	mux := http.NewServeMux()
	RegisterAICutVoiceCloneRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, cloner, func() time.Time { return now })
	for _, tc := range []struct {
		body map[string]any
		want string
	}{
		{body: map[string]any{"assetId": "voice-reference", "text": "测试"}, want: "仅支持音频素材"},
		{body: map[string]any{"assetId": "missing", "text": "测试"}, want: "素材不存在"},
		{body: map[string]any{"assetId": "voice-reference", "text": " "}, want: "试听文本不能为空"},
	} {
		response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-voice/voice-clone", tc.body)
		if response.Code != http.StatusBadRequest && response.Code != http.StatusNotFound {
			t.Fatalf("status=%d %s", response.Code, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), tc.want) {
			t.Fatalf("body=%s want=%q", response.Body.String(), tc.want)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("provider calls=%d", calls.Load())
	}
}

func TestAICutVoiceCloneUsesConfiguredTemplateWithReferenceAudio(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 10, 0, 0, time.UTC)
	store, objects, _ := seedAICutVoiceReference(t, now, "audio")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		raw, _ := io.ReadAll(req.Body)
		body := string(raw)
		if req.Header.Get("Authorization") != "Bearer clone-secret" || !strings.Contains(body, `"text":"克隆测试"`) || !strings.Contains(body, `"reference":"SUQzLXJlZmVyZW5jZQ=="`) {
			t.Fatalf("auth=%q body=%s", req.Header.Get("Authorization"), body)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("ID3-configured-clone"))
	}))
	defer upstream.Close()
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(t.Context(), "alice", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "clone-audio", "kind": "audio", "enabled": true, "endpoint": upstream.URL, "credential": "clone-secret",
		"capabilities":              map[string]any{"supportsVoiceClone": true},
		"voiceCloneRequestTemplate": `{"method":"POST","headers":{"Authorization":"Bearer {{credential}}"},"body":{"text":"{{input}}","reference":"{{reference_audio_base64}}","mediaType":"{{reference_audio_media_type}}"}}`,
		"voiceCloneResponseMapping": `{}`,
	}}}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutVoiceCloneRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, catalog, nil, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-voice/voice-clone", map[string]any{"assetId": "voice-reference", "text": "克隆测试"})
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAICutVoiceCloneCanceledRequestDoesNotCreateAsset(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 15, 0, 0, time.UTC)
	store, objects, _ := seedAICutVoiceReference(t, now, "audio")
	cloner := func(context.Context, string, aicut.Asset, []byte, aicutVoiceCloneRequest) ([]byte, string, error) {
		return nil, "", context.Canceled
	}
	mux := http.NewServeMux()
	RegisterAICutVoiceCloneRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, cloner, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-voice/voice-clone", map[string]any{"assetId": "voice-reference", "text": "取消测试"})
	if response.Code != http.StatusRequestTimeout || !strings.Contains(response.Body.String(), "已取消") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assets, err := store.ListAssets(t.Context(), "alice", "cut-voice")
	if err != nil || len(assets) != 1 {
		t.Fatalf("assets=%+v err=%v", assets, err)
	}
}

func TestAICutVoiceCloneRepeatedCompletedRequestsCreateDistinctAssets(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 20, 0, 0, time.UTC)
	store, objects, _ := seedAICutVoiceReference(t, now, "audio")
	var calls atomic.Int32
	cloner := func(context.Context, string, aicut.Asset, []byte, aicutVoiceCloneRequest) ([]byte, string, error) {
		calls.Add(1)
		return []byte("ID3-clone"), "audio/mpeg", nil
	}
	mux := http.NewServeMux()
	RegisterAICutVoiceCloneRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, cloner, func() time.Time { return now })
	ids := map[string]bool{}
	for range 2 {
		response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-voice/voice-clone", map[string]any{"assetId": "voice-reference", "text": "重复完成测试"})
		if response.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		var payload struct {
			Asset aicut.Asset `json:"asset"`
		}
		decodeTestJSON(t, response.Body.Bytes(), &payload)
		ids[payload.Asset.ID] = true
	}
	if calls.Load() != 2 || len(ids) != 2 {
		t.Fatalf("calls=%d ids=%v", calls.Load(), ids)
	}
}

func seedAICutVoiceReference(t *testing.T, now time.Time, kind string) (*aicut.MemoryStore, *agentworkspace.LocalAssetObjectStore, aicut.Asset) {
	t.Helper()
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-voice", Title: "声音克隆", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	contentType := "audio/mpeg"
	if kind == "video" {
		contentType = "video/mp4"
	}
	asset := aicut.Asset{ID: "voice-reference", ProjectID: "cut-voice", Kind: kind, Name: "参考音色.mp3", ContentType: contentType, ObjectKey: "aicut-assets/alice/cut-voice/voice-reference.mp3", ByteSize: int64(len("ID3-reference")), DurationMS: 3000, CreatedAt: now}
	if err := store.CreateAsset(t.Context(), "alice", asset); err != nil {
		t.Fatal(err)
	}
	if err := objects.Put(t.Context(), aicutObjectAsset(asset), bytes.NewReader([]byte("ID3-reference"))); err != nil {
		t.Fatal(err)
	}
	return store, objects, asset
}
