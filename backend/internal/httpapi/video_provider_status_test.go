package httpapi

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVideoProviderStatusDoesNotExposeCredentials(t *testing.T) {
	status := videoProviderStatus(map[string]any{
		"video": map[string]any{"ydApiKey": "secret-key"},
	}, "personal_api")
	if status["configured"] != true || status["model"] != "yd2.0-mini" {
		t.Fatalf("status=%#v", status)
	}
	if strings.Contains(strings.ToLower(strings.Join([]string{status["provider"].(string), status["model"].(string)}, " ")), "secret") {
		t.Fatal("video provider status leaked a credential")
	}
}

func TestV12VideoProviderStatusUsesGoUserConfig(t *testing.T) {
	now := time.Unix(1700000000, 0)
	config := NewUserConfigStore(filepath.Join(t.TempDir(), "users"))
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, UserConfigStore: config})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/video-provider/status?provider=personal_api", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"provider":"personal_api"`) || !strings.Contains(rec.Body.String(), `"configured":false`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
