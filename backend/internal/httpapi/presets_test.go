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
