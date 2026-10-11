package httpapi

import (
	"net/http"
	"testing"
	"time"
)

func TestNovelPanelReferenceAssetUploadSupportsMountedWorkbenchPath(t *testing.T) {
	mux := http.NewServeMux()
	RegisterNovelPanelAssetCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return time.Unix(100, 0) }}, t.TempDir())
	recorder := signedJSONRequest(t, mux, time.Unix(100, 0), "alice", http.MethodPost, "/api/novel-panel/reference-assets/upload", map[string]any{
		"asset_type": "character", "asset_id": "hero_1", "variant": "source",
		"data_url": "data:image/png;base64,iVBORw0KGgo=",
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	promote := signedJSONRequest(t, mux, time.Unix(100, 0), "alice", http.MethodPost, "/api/novel-panel/reference-assets/use-source-as-main", map[string]any{
		"asset_type": "character", "asset_id": "hero_1",
	})
	if promote.Code != http.StatusOK {
		t.Fatalf("promote status=%d body=%s", promote.Code, promote.Body.String())
	}
	read := signedJSONRequest(t, mux, time.Unix(100, 0), "alice", http.MethodGet, "/api/novel-panel/reference-assets/file/character/hero_1/main", nil)
	if read.Code != http.StatusOK || read.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("read status=%d content-type=%q body=%s", read.Code, read.Header().Get("Content-Type"), read.Body.String())
	}
}
