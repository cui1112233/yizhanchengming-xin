package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestConfiguredTTSRequestRendersTemplateWithoutLeakingShape(t *testing.T) {
	method, headers, body, err := configuredTTSRequest(`{"method":"POST","headers":{"X-Test":"{{voice}}"},"body":{"text":"{{prompt}}","rate":"{{speech_rate}}"}}`, ttsRequest{Input: "你好", Voice: "女声", Speed: 1.5, Pitch: "2", Style: "general"}, "secret")
	if err != nil || method != http.MethodPost || headers["X-Test"] != "女声" {
		t.Fatalf("method=%s headers=%v err=%v", method, headers, err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil || decoded["text"] != "你好" || decoded["rate"] != "1.5" {
		t.Fatalf("body=%s err=%v", body, err)
	}
}

func TestConfiguredTTSModelUsesEnabledManagedAudioModel(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tts-owner", map[string]any{"modelCatalog": []any{
		map[string]any{"kind": "audio", "enabled": false, "endpoint": "https://disabled.example", "credential": "secret", "requestTemplate": `{}`},
		map[string]any{"kind": "audio", "enabled": true, "endpoint": "https://enabled.example", "credential": "secret", "requestTemplate": `{}`},
	}}); err != nil {
		t.Fatal(err)
	}
	model, ok := configuredTTSModel(context.Background(), store, "tts-owner")
	if !ok || model["endpoint"] != "https://enabled.example" {
		t.Fatalf("model=%#v ok=%v", model, ok)
	}
}

func TestTTSRouteValidatesInputBeforeUpstream(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }})
	rec := signedJSONRequest(t, api, now, "tts-user", http.MethodPost, "/api/tts", map[string]any{"input": "  "})
	if rec.Code != http.StatusBadRequest || rec.Body.String() != "{\"error\":\"输入文本不能为空\"}\n" {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
