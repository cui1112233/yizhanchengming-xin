package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
)

func TestAICutAssetListReturnsExpiringContentURLUsableByMediaElements(t *testing.T) {
	current := time.Date(2026, 11, 11, 11, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-access", Title: "媒体访问", CreatedAt: current, UpdatedAt: current}); err != nil {
		t.Fatal(err)
	}
	asset := aicut.Asset{ID: "image-1", ProjectID: "cut-access", Kind: "image", Name: "人物.png", ContentType: "image/png", ObjectKey: "aicut-assets/alice/cut-access/image-1.png", ByteSize: int64(len(opaquePNG(t))), CreatedAt: current}
	if err := store.CreateAsset(t.Context(), "alice", asset); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	if err := objects.Put(t.Context(), aicutObjectAsset(asset), bytes.NewReader(opaquePNG(t))); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return current }}, store, objects, func() time.Time { return current })

	listed := signedJSONRequest(t, mux, current, "alice", http.MethodGet, "/api/ai-cut/projects/cut-access/assets", nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list=%d %s", listed.Code, listed.Body.String())
	}
	var payload struct {
		Assets []aicut.Asset `json:"assets"`
	}
	decodeTestJSON(t, listed.Body.Bytes(), &payload)
	if len(payload.Assets) != 1 {
		t.Fatalf("assets=%+v", payload.Assets)
	}
	parsed, err := url.Parse(payload.Assets[0].URL)
	if err != nil || parsed.Query().Get("access") == "" || parsed.Query().Get("expires") == "" || parsed.Query().Get("owner") != "alice" {
		t.Fatalf("signed url=%q err=%v", payload.Assets[0].URL, err)
	}

	request := httptest.NewRequest(http.MethodGet, payload.Assets[0].URL, nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), opaquePNG(t)) {
		t.Fatalf("content=%d bytes=%d", response.Code, response.Body.Len())
	}

	current = current.Add(25 * time.Hour)
	expired := httptest.NewRecorder()
	mux.ServeHTTP(expired, httptest.NewRequest(http.MethodGet, payload.Assets[0].URL, nil))
	if expired.Code != http.StatusUnauthorized {
		t.Fatalf("expired=%d body=%s", expired.Code, expired.Body.String())
	}
}

func decodeTestJSON(t *testing.T, raw []byte, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatal(err)
	}
}
