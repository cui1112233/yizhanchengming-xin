package httpapi

import (
	"net/http"
	"testing"
	"time"
)

func TestTTSRouteValidatesInputBeforeUpstream(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }})
	rec := signedJSONRequest(t, api, now, "tts-user", http.MethodPost, "/api/tts", map[string]any{"input": "  "})
	if rec.Code != http.StatusBadRequest || rec.Body.String() != "{\"error\":\"输入文本不能为空\"}\n" {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
