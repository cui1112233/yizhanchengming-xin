package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestModelTestReusesSavedCredentialWhenEditing(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer saved-secret" {
			t.Errorf("authorization = %q", req.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "text-saved", "kind": "text", "enabled": true, "modelId": "saved-model", "baseUrl": upstream.URL, "credential": "saved-secret",
	}}}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/config/models/test", map[string]any{
		"kind": "text", "existingModelId": "text-saved", "baseUrl": upstream.URL, "modelId": "saved-model",
	}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestModelCatalogCRUDHidesCredential(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	create := doNovelFetchRequest(t, handler, http.MethodPost, "/api/config/models", map[string]any{
		"id": "text-demo", "kind": "text", "displayName": "测试文本", "modelId": "demo", "baseUrl": "https://example.com/v1/chat/completions", "credential": "secret-value", "enabled": true,
	}, now, secret)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	if strings.Contains(create.Body.String(), "secret-value") {
		t.Fatal("model credential leaked in response")
	}
	list := doNovelFetchRequest(t, handler, http.MethodGet, "/api/models?kind=text", nil, now, secret)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		Models []map[string]any `json:"models"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Models) != 1 || payload.Models[0]["hasCredential"] != true {
		t.Fatalf("unexpected public catalog: %#v", payload.Models)
	}
}

func TestLegacyModelTestRoutesAreServedByGo(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	for _, path := range []string{"/api/test/text", "/api/test/image"} {
		response := doNovelFetchRequest(t, handler, http.MethodPost, path, map[string]any{"baseUrl": "ftp://invalid", "model": "demo", "apiKey": "secret"}, now, secret)
		if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "Base URL 格式不正确") {
			t.Fatalf("%s should be Go-owned validation, status=%d body=%s", path, response.Code, response.Body.String())
		}
	}
}

func TestModelCatalogPatchPreservesCredentialAndValidatesWorkflow(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "minimax-h3-video", "kind": "video", "displayName": "H3", "workflowId": "saved-workflow", "credential": "saved-secret", "enabled": false,
	}}}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPatch, "/api/config/models/minimax-h3-video", map[string]any{"enabled": true, "workflowId": "new-workflow"}, now, secret)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "saved-secret") {
		t.Fatalf("patch status=%d body=%s", response.Code, response.Body.String())
	}
	settings, err := store.GetConfig(context.Background(), "tester")
	if err != nil { t.Fatal(err) }
	model := modelCatalog(settings)[0]
	if model["credential"] != "saved-secret" || model["workflowId"] != "new-workflow" { t.Fatalf("patch lost protected fields: %#v", model) }
}

func TestModelCatalogH3ConfigRequiresKeyAndSavesWorkflowTogether(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})

	missingKey := doNovelFetchRequest(t, handler, http.MethodPost, "/api/config/models", map[string]any{
		"id": "minimax-h3-video", "kind": "video", "displayName": "H3", "workflowId": "saved-workflow", "enabled": false,
	}, now, secret)
	if missingKey.Code != http.StatusBadRequest || !strings.Contains(missingKey.Body.String(), "API Key") {
		t.Fatalf("H3 save without key must be rejected, status=%d body=%s", missingKey.Code, missingKey.Body.String())
	}

	withKey := doNovelFetchRequest(t, handler, http.MethodPost, "/api/config/models", map[string]any{
		"id": "minimax-h3-video", "kind": "video", "displayName": "H3", "workflowId": "saved-workflow", "credential": "h3-secret", "enabled": false,
	}, now, secret)
	if withKey.Code != http.StatusCreated || strings.Contains(withKey.Body.String(), "h3-secret") {
		t.Fatalf("H3 save with key should succeed without leaking it, status=%d body=%s", withKey.Code, withKey.Body.String())
	}
	settings, err := store.GetConfig(context.Background(), "tester")
	if err != nil { t.Fatal(err) }
	model := modelCatalog(settings)[0]
	if model["credential"] != "h3-secret" || model["workflowId"] != "saved-workflow" {
		t.Fatalf("H3 key and workflow must be saved together: %#v", model)
	}
}

func TestModelCatalogRejectsEmptyH3Workflow(t *testing.T) {
	model, err := normalizeManagedModel(map[string]any{"id": "minimax-h3-video", "kind": "video", "workflowId": ""})
	if err == nil || model != nil { t.Fatalf("empty H3 workflow must be rejected: %#v, %v", model, err) }
}

func TestModelCatalogAcceptsManagedAudioFields(t *testing.T) {
	model, err := normalizeManagedModel(map[string]any{
		"id": "tts-demo", "kind": "audio", "displayName": "测试配音", "enabled": false,
		"endpoint": "https://audio.example.test/speech", "requestTemplate": `{"method":"POST","body":{"text":"{{input}}"}}`,
		"responseMapping": `{"resultUrl":"data.url"}`, "credential": "secret-value",
	})
	if err != nil {
		t.Fatalf("audio model should be accepted: %v", err)
	}
	for key, want := range map[string]string{"endpoint": "https://audio.example.test/speech", "requestTemplate": `{"method":"POST","body":{"text":"{{input}}"}}`, "responseMapping": `{"resultUrl":"data.url"}`} {
		if model[key] != want {
			t.Fatalf("%s=%#v, want %q", key, model[key], want)
		}
	}
}

func TestManagedAudioModelConnectionTestAcceptsAudioResponse(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer audio-secret" {
			t.Errorf("authorization = %q", req.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("audio-probe"))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/config/models/test-audio", map[string]any{
		"kind": "audio", "endpoint": upstream.URL, "credential": "audio-secret",
		"requestTemplate": `{"method":"POST","body":{"text":"{{input}}"}}`, "responseMapping": `{}`,
	}, now, secret)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "音频模型连接成功") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
