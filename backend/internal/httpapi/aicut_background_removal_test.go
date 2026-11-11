package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
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

func TestAICutBackgroundRemovalCreatesDurableTransparentImageAsset(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	store, objects, source := seedAICutBackgroundRemovalSource(t, now, "image")
	var calls atomic.Int32
	remover := func(ctx context.Context, owner string, asset aicut.Asset, payload []byte) ([]byte, string, error) {
		calls.Add(1)
		if owner != "alice" || asset.ID != source.ID || !bytes.Equal(payload, opaquePNG(t)) {
			t.Fatalf("unexpected removal input owner=%q asset=%+v bytes=%d", owner, asset, len(payload))
		}
		return transparentPNG(t), "image/png", nil
	}
	mux := http.NewServeMux()
	RegisterAICutBackgroundRemovalRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, remover, func() time.Time { return now })

	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-bg/assets/source-image/remove-background", map[string]any{"name": "人物透明图"})
	if response.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", response.Code, response.Body.String())
	}
	var payload struct {
		Asset aicut.Asset `json:"asset"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Asset.Kind != "image" || payload.Asset.Name != "人物透明图.png" || payload.Asset.ContentType != "image/png" || payload.Asset.URL == "" || calls.Load() != 1 {
		t.Fatalf("asset=%+v calls=%d", payload.Asset, calls.Load())
	}
	stored, err := store.Asset(t.Context(), "alice", "cut-bg", payload.Asset.ID)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := objects.Open(t.Context(), aicutObjectAsset(stored))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	if !aicutImageHasTransparency(data) {
		t.Fatal("stored image is not transparent")
	}
}

func TestAICutBackgroundRemovalValidatesBeforeProvider(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 5, 0, 0, time.UTC)
	store, objects, _ := seedAICutBackgroundRemovalSource(t, now, "video")
	var calls atomic.Int32
	remover := func(context.Context, string, aicut.Asset, []byte) ([]byte, string, error) {
		calls.Add(1)
		return transparentPNG(t), "image/png", nil
	}
	mux := http.NewServeMux()
	RegisterAICutBackgroundRemovalRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, remover, func() time.Time { return now })

	for _, tc := range []struct{ path, want string }{
		{path: "/api/ai-cut/projects/cut-bg/assets/source-image/remove-background", want: "仅支持图片素材"},
		{path: "/api/ai-cut/projects/cut-bg/assets/missing/remove-background", want: "素材不存在"},
		{path: "/api/ai-cut/projects/missing/assets/source-image/remove-background", want: "项目不存在"},
	} {
		response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, tc.path, map[string]any{})
		if response.Code != http.StatusBadRequest && response.Code != http.StatusNotFound {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), tc.want) {
			t.Fatalf("body=%s want %q", response.Body.String(), tc.want)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("provider calls=%d want 0", calls.Load())
	}
}

func TestAICutBackgroundRemovalRejectsOpaqueProviderResultWithoutCreatingAsset(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 10, 0, 0, time.UTC)
	store, objects, _ := seedAICutBackgroundRemovalSource(t, now, "image")
	remover := func(context.Context, string, aicut.Asset, []byte) ([]byte, string, error) {
		return opaquePNG(t), "image/png", nil
	}
	mux := http.NewServeMux()
	RegisterAICutBackgroundRemovalRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, nil, remover, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-bg/assets/source-image/remove-background", map[string]any{})
	if response.Code != http.StatusBadGateway || !strings.Contains(response.Body.String(), "没有返回透明背景") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assets, err := store.ListAssets(t.Context(), "alice", "cut-bg")
	if err != nil || len(assets) != 1 {
		t.Fatalf("assets=%+v err=%v", assets, err)
	}
}

func TestAICutBackgroundRemovalUsesConfiguredImageEditModel(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 15, 0, 0, time.UTC)
	store, objects, _ := seedAICutBackgroundRemovalSource(t, now, "image")
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls.Add(1)
		if req.URL.Path != "/v1/images/edits" || req.Header.Get("Authorization") != "Bearer image-secret" {
			t.Fatalf("request path=%q auth=%q", req.URL.Path, req.Header.Get("Authorization"))
		}
		if err := req.ParseMultipartForm(4 << 20); err != nil {
			t.Fatal(err)
		}
		if req.FormValue("model") != "edit-model" || !strings.Contains(req.FormValue("prompt"), "transparent background") {
			t.Fatalf("model=%q prompt=%q", req.FormValue("model"), req.FormValue("prompt"))
		}
		files := req.MultipartForm.File["image[]"]
		if len(files) != 1 {
			t.Fatalf("image parts=%d", len(files))
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": []any{map[string]any{"b64_json": base64.StdEncoding.EncodeToString(transparentPNG(t))}}})
	}))
	defer upstream.Close()
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(t.Context(), "alice", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "image-edit", "kind": "image", "enabled": true, "supportsImageEdits": true,
		"modelId": "edit-model", "endpoint": upstream.URL + "/v1/images/generations", "credential": "image-secret",
	}}}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutBackgroundRemovalRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, catalog, nil, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-bg/assets/source-image/remove-background", map[string]any{})
	if response.Code != http.StatusCreated || calls.Load() != 1 {
		t.Fatalf("status=%d body=%s calls=%d", response.Code, response.Body.String(), calls.Load())
	}
}

func seedAICutBackgroundRemovalSource(t *testing.T, now time.Time, kind string) (*aicut.MemoryStore, *agentworkspace.LocalAssetObjectStore, aicut.Asset) {
	t.Helper()
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-bg", Title: "去背工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	contentType := "image/png"
	if kind == "video" {
		contentType = "video/mp4"
	}
	source := aicut.Asset{ID: "source-image", ProjectID: "cut-bg", Kind: kind, Name: "人物.png", ContentType: contentType, ObjectKey: "aicut-assets/alice/cut-bg/source-image.png", ByteSize: int64(len(opaquePNG(t))), CreatedAt: now}
	if err := objects.Put(t.Context(), aicutObjectAsset(source), bytes.NewReader(opaquePNG(t))); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(t.Context(), "alice", source); err != nil {
		t.Fatal(err)
	}
	return store, objects, source
}

func opaquePNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 220, G: 80, B: 60, A: 255})
		}
	}
	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func transparentPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 220, G: 80, B: 60, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{A: 0})
	img.SetNRGBA(0, 1, color.NRGBA{A: 0})
	img.SetNRGBA(1, 1, color.NRGBA{A: 0})
	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
