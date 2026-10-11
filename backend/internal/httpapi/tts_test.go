package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestRequestConfiguredTTSBytesAcceptsAudioAndRejectsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/audio" {
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write([]byte("fake-audio"))
			return
		}
		if r.URL.Path == "/json-audio" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"url":"` + serverURLForTest(r, "/audio") + `"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":"bad model"}`))
	}))
	defer server.Close()
	model := map[string]any{
		"endpoint":        server.URL + "/audio",
		"credential":      "secret",
		"requestTemplate": `{"method":"POST","body":{"text":"{{input}}"}}`,
	}
	payload, contentType, err := requestConfiguredTTSBytes(context.Background(), ttsRequest{Input: "测试"}, model)
	if err != nil || string(payload) != "fake-audio" || contentType != "audio/mpeg" {
		t.Fatalf("payload=%q contentType=%q err=%v", payload, contentType, err)
	}
	model["endpoint"] = server.URL + "/json"
	if _, _, err := requestConfiguredTTSBytes(context.Background(), ttsRequest{Input: "测试"}, model); err == nil {
		t.Fatal("JSON error response must not be accepted as audio")
	}
	model["endpoint"] = server.URL + "/json-audio"
	model["responseMapping"] = `{"resultUrl":"data.url"}`
	payload, contentType, err = requestConfiguredTTSBytes(context.Background(), ttsRequest{Input: "测试"}, model)
	if err != nil || string(payload) != "fake-audio" || contentType != "audio/mpeg" {
		t.Fatalf("mapped payload=%q contentType=%q err=%v", payload, contentType, err)
	}
}

func serverURLForTest(r *http.Request, path string) string {
	return "http://" + r.Host + path
}

func TestTTSRouteValidatesInputBeforeUpstream(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }})
	rec := signedJSONRequest(t, api, now, "tts-user", http.MethodPost, "/api/tts", map[string]any{"input": "  "})
	if rec.Code != http.StatusBadRequest || rec.Body.String() != "{\"error\":\"输入文本不能为空\"}\n" {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTTSRouteUsesConfiguredAudioModelEndToEnd(t *testing.T) {
	now := time.Unix(1700000000, 0)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer audio-secret" {
			t.Fatalf("unexpected upstream request: method=%s authorization=%q", r.Method, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("configured-audio"))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tts-user", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "tts-configured", "kind": "audio", "enabled": true, "endpoint": upstream.URL,
		"credential": "audio-secret", "requestTemplate": `{"method":"POST","body":{"text":"{{input}}"}}`,
	}}}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchStore: store})
	rec := signedJSONRequest(t, api, now, "tts-user", http.MethodPost, "/api/tts", map[string]any{"input": "你好"})
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "audio/mpeg" || rec.Body.String() != "configured-audio" {
		t.Fatalf("status=%d content-type=%q body=%q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
}
