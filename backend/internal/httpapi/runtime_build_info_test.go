package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/runtimeaudit"
)

func TestRuntimeBuildInfoReportsInjectedReleaseSHA(t *testing.T) {
	handler := NewRouter(RouterOptions{
		BridgeSecret: "test-secret",
		ReleaseSHA:   "abc123",
	})
	req := httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["service"] != "go-api" || body["git_sha"] != "abc123" || body["runtime"] != "go" {
		t.Fatalf("body = %#v, want Go runtime with injected SHA", body)
	}
	if body["bridge_hits"] != float64(0) {
		t.Fatalf("bridge_hits = %#v, want 0", body["bridge_hits"])
	}
}

func TestRuntimeBuildInfoReportsMediaEngine(t *testing.T) {
	handler := NewRouter(RouterOptions{ReleaseSHA: "abc123"})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil))

	var body struct {
		MediaEngine struct {
			Available     bool   `json:"available"`
			FFmpegPath    string `json:"ffmpeg_path"`
			FFmpegVersion string `json:"ffmpeg_version"`
			FFprobePath   string `json:"ffprobe_path"`
		} `json:"media_engine"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.MediaEngine.Available && (body.MediaEngine.FFmpegPath == "" || body.MediaEngine.FFmpegVersion == "") {
		t.Fatalf("media engine must expose path and version when available: %#v", body.MediaEngine)
	}
}

func TestRuntimeBuildInfoReadsRuntimeManifest(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "runtime-manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"gitSha":"manifest123","frontendEntry":"/assets/index-current.js"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{RuntimeManifestPath: manifestPath})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil))

	var body struct {
		FrontendEntry      string `json:"frontend_entry"`
		RuntimeManifestSHA string `json:"runtime_manifest_sha"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.FrontendEntry != "/assets/index-current.js" || body.RuntimeManifestSHA != "manifest123" {
		t.Fatalf("body = %#v", body)
	}
}

func TestRuntimeBuildInfoReadsBOMRuntimeManifest(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "runtime-manifest.json")
	contents := append([]byte{0xef, 0xbb, 0xbf}, []byte(`{"gitSha":"bom-manifest123","frontendEntry":"/assets/bom-index.js"}`)...)
	if err := os.WriteFile(manifestPath, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{RuntimeManifestPath: manifestPath})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil))

	var body struct {
		FrontendEntry      string `json:"frontend_entry"`
		RuntimeManifestSHA string `json:"runtime_manifest_sha"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.FrontendEntry != "/assets/bom-index.js" || body.RuntimeManifestSHA != "bom-manifest123" {
		t.Fatalf("body = %#v", body)
	}
}

func TestRuntimeBuildInfoIgnoresInvalidRuntimeManifest(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "runtime-manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`invalid`), 0o600); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{RuntimeManifestPath: manifestPath})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil))

	var body struct {
		FrontendEntry      string `json:"frontend_entry"`
		RuntimeManifestSHA string `json:"runtime_manifest_sha"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.FrontendEntry != "" || body.RuntimeManifestSHA != "" {
		t.Fatalf("body = %#v", body)
	}
}

func TestRuntimeBuildInfoReportsBridgeHits(t *testing.T) {
	audit := runtimeaudit.NewRegistry()
	audit.RecordBridge("/api/login")
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", RuntimeAudit: audit})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil))
	var body struct {
		BridgeHits  int            `json:"bridge_hits"`
		BridgePaths map[string]int `json:"bridge_paths"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.BridgeHits != 1 || body.BridgePaths["/api/login"] != 1 {
		t.Fatalf("body = %#v", body)
	}
}
