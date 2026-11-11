package httpapi

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

func TestAICutProxyMediaGenerateReuseAndServe(t *testing.T) {
	now := time.Date(2026, 11, 11, 6, 30, 0, 0, time.UTC)
	temp := t.TempDir()
	sourcePath := filepath.Join(temp, "source.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=1920x1080:r=30:d=0.35", "-c:v", "libx264", "-pix_fmt", "yuv420p", sourcePath); err != nil {
		t.Skipf("ffmpeg unavailable: %v", err)
	}
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "代理测试", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	asset := aicut.Asset{ID: "video-1", ProjectID: "cut-1", Kind: "video", Name: "source.mp4", ContentType: "video/mp4", ObjectKey: "aicut-assets/alice/cut-1/video-1.mp4", ByteSize: info.Size(), DurationMS: 350, CreatedAt: now}
	if err := store.CreateAsset(t.Context(), "alice", asset); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(temp, "objects"))
	file, err := os.Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := objects.Put(t.Context(), aicutObjectAsset(asset), file); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	_ = file.Close()

	mux := http.NewServeMux()
	RegisterAICutAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, func() time.Time { return now })
	endpoint := "/api/ai-cut/projects/cut-1/assets/video-1/proxy"
	first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, endpoint, nil)
	if first.Code != http.StatusCreated || !containsText(first.Body.String(), `"proxyUrl":"/api/ai-cut/projects/cut-1/assets/video-1/proxy/content?access=`) || !containsText(first.Body.String(), `"reused":false`) {
		t.Fatalf("first=%d %s", first.Code, first.Body.String())
	}
	second := signedJSONRequest(t, mux, now, "alice", http.MethodPost, endpoint, nil)
	if second.Code != http.StatusOK || !containsText(second.Body.String(), `"reused":true`) {
		t.Fatalf("second=%d %s", second.Code, second.Body.String())
	}
	listed := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/ai-cut/projects/cut-1/assets", nil)
	if listed.Code != http.StatusOK || !containsText(listed.Body.String(), `"proxyUrl":"/api/ai-cut/projects/cut-1/assets/video-1/proxy/content?access=`) {
		t.Fatalf("listed=%d %s", listed.Code, listed.Body.String())
	}
	served := signedJSONRequest(t, mux, now, "alice", http.MethodGet, endpoint+"/content", nil)
	if served.Code != http.StatusOK || served.Header().Get("Content-Type") != "video/mp4" || served.Body.Len() == 0 {
		t.Fatalf("served=%d type=%s bytes=%d body=%s", served.Code, served.Header().Get("Content-Type"), served.Body.Len(), served.Body.String())
	}
	proxyPath := filepath.Join(temp, "proxy.mp4")
	if err := os.WriteFile(proxyPath, served.Body.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	command, _, err := ffmpegbin.Command(t.Context(), "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", proxyPath)
	if err != nil {
		t.Fatal(err)
	}
	probe, err := command.Output()
	if err != nil || strings.TrimSpace(string(probe)) != "960x540" {
		t.Fatalf("proxy dimensions output=%q err=%v", probe, err)
	}
}

func TestAICutProxyMediaRejectsNonVideoAndMissingProxy(t *testing.T) {
	now := time.Date(2026, 11, 11, 6, 45, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "代理边界", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	asset := aicut.Asset{ID: "image-1", ProjectID: "cut-1", Kind: "image", Name: "cover.png", ContentType: "image/png", ObjectKey: "aicut-assets/alice/cut-1/image-1.png", ByteSize: 4, CreatedAt: now}
	if err := store.CreateAsset(t.Context(), "alice", asset); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(t.TempDir())
	mux := http.NewServeMux()
	RegisterAICutAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, func() time.Time { return now })
	invalid := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/assets/image-1/proxy", nil)
	if invalid.Code != http.StatusBadRequest || !containsText(invalid.Body.String(), "仅视频素材可以生成代理") {
		t.Fatalf("invalid=%d %s", invalid.Code, invalid.Body.String())
	}
	missing := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/ai-cut/projects/cut-1/assets/image-1/proxy/content", nil)
	if missing.Code != http.StatusBadRequest && missing.Code != http.StatusNotFound {
		body, _ := io.ReadAll(missing.Body)
		t.Fatalf("missing=%d %s", missing.Code, body)
	}
}

func TestAICutExportSourceNeverUsesProxyObject(t *testing.T) {
	asset := aicut.Asset{ID: "video-1", ProjectID: "cut-1", Kind: "video", Name: "source.mp4", ContentType: "video/mp4", ObjectKey: "aicut-assets/alice/cut-1/video-1.mp4", ByteSize: 100}
	original := aicutExportSourceObject(asset)
	proxy := aicutProxyObjectAsset(asset)
	if original.ObjectKey != asset.ObjectKey {
		t.Fatalf("export source=%q, want original %q", original.ObjectKey, asset.ObjectKey)
	}
	if original.ObjectKey == proxy.ObjectKey || original.Name == proxy.Name {
		t.Fatalf("export source must differ from proxy: original=%+v proxy=%+v", original, proxy)
	}
}
