package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/prompt"
)

func TestComicStylePreviewRouteServesEmbeddedJPEG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, PresetCatalog: prompt.NewCatalog(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/presets/comic-style/batch-comic-style-01/preview", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("content type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("cache control = %q", got)
	}
	if body := rec.Body.Bytes(); len(body) < 2 || body[0] != 0xff || body[1] != 0xd8 {
		t.Fatalf("preview is not a JPEG: %x", body[:min(len(body), 8)])
	}
}

func TestPresetSlotLabelUsesChineseNames(t *testing.T) {
	if got := presetSlotLabel("batch.comic-style"); got != "漫画画风库" {
		t.Fatalf("comic style label = %q", got)
	}
	if got := presetSlotLabel("batch.comic-reference-scene"); got != "漫画场景参考图库" {
		t.Fatalf("comic scene label = %q", got)
	}
}

func TestComicPresetDefaultsDoNotExposeInternalBootstrapProtocol(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	seed := `[{"id":"batch-comic-bootstrap","module":"batch-factory","name":"CM·漫画首次统一生成","version":1,"status":"published","body":"old","createdBy":"system"},{"id":"admin-custom","module":"batch-factory","name":"管理员自定义","version":1,"status":"published","body":"keep","createdBy":"admin"}]`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	catalog := prompt.NewCatalog(path)
	if _, found, err := catalog.PublishedBody("batch-factory", "batch-comic-bootstrap"); err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatal("internal comic bootstrap protocol must not be exposed as an editable preset")
	}
	if _, found, err := catalog.PublishedBody("batch-factory", "admin-custom"); err != nil {
		t.Fatal(err)
	} else if !found {
		t.Fatal("unrelated administrator preset must be preserved")
	}
}

func TestComicReferenceScenePreviewRouteServesEmbeddedJPEG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, PresetCatalog: prompt.NewCatalog(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/presets/comic-reference-scene/batch-comic-reference-scene-campus-b/preview", nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("status=%d type=%q body=%s", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
	if body := rec.Body.Bytes(); len(body) < 2 || body[0] != 0xff || body[1] != 0xd8 {
		t.Fatalf("scene preview is not a JPEG: %x", body[:min(len(body), 8)])
	}
}
