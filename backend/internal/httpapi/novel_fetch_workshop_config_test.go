package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func doLegacyBatchRewriteRequest(t *testing.T, handler http.Handler, method, path string, body any, now time.Time, secret string) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	SignBridgeRequest(req, "tester", true, now, secret)
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, req)
	return result
}

func TestApplyBatchRewriteDefaultsEnablesPlatformDetectionUnlessExplicitlyDisabled(t *testing.T) {
	defaultConfig := map[string]any{}
	applyBatchRewriteDefaults(defaultConfig)
	defaultFetch := mapOrEmpty(defaultConfig["fetch"])
	if defaultFetch["auto_detect_platform"] != true {
		t.Fatalf("expected platform detection enabled by default, got %#v", defaultFetch["auto_detect_platform"])
	}

	disabledConfig := map[string]any{"fetch": map[string]any{"auto_detect_platform": false}}
	applyBatchRewriteDefaults(disabledConfig)
	disabledFetch := mapOrEmpty(disabledConfig["fetch"])
	if disabledFetch["auto_detect_platform"] != false {
		t.Fatalf("expected explicit platform detection disable to be preserved, got %#v", disabledFetch["auto_detect_platform"])
	}
}

func TestLegacyBatchRewriteConfigPostSavesAndReturnsConfig(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	for _, path := range []string{"/api/batch-rewrite/bootstrap", "/api/batch-rewrite/config"} {
		get := doLegacyBatchRewriteRequest(t, handler, http.MethodGet, path, nil, now, secret)
		if get.Code != http.StatusOK {
			t.Fatalf("legacy config GET path=%s status=%d body=%s", path, get.Code, get.Body.String())
		}
	}
	post := doLegacyBatchRewriteRequest(t, handler, http.MethodPost, "/api/batch-rewrite/config", map[string]any{
		"appConfig": map[string]any{"fetch": map[string]any{"default_max_txt": 4000}},
	}, now, secret)
	if post.Code == http.StatusNotFound {
		t.Fatalf("legacy config POST was not routed: status=%d body=%s", post.Code, post.Body.String())
	}
	if post.Code != http.StatusOK {
		t.Fatalf("legacy config POST status=%d body=%s", post.Code, post.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(post.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	config, ok := payload["config"].(map[string]any)
	if !ok || config["app_config"] == nil {
		t.Fatalf("legacy config response=%#v", payload)
	}
	saved, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	if saved["appConfig"] == nil {
		t.Fatalf("legacy config was not saved: %#v", saved)
	}
	get := doLegacyBatchRewriteRequest(t, handler, http.MethodGet, "/api/batch-rewrite/config", nil, now, secret)
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), "default_max_txt") {
		t.Fatalf("legacy config GET did not reflect saved config: status=%d body=%s", get.Code, get.Body.String())
	}
}

func TestLegacyBatchRewriteConfigPostRequiresAuthentication(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", Now: func() time.Time { return now }, NovelFetchStore: store})
	before, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	post := doLegacyBatchRewriteRequest(t, handler, http.MethodPost, "/api/batch-rewrite/config", map[string]any{"appConfig": map[string]any{"fetch": map[string]any{"default_max_txt": 4000}}}, now, "")
	if post.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated legacy config POST status=%d body=%s", post.Code, post.Body.String())
	}
	if after, err := store.GetConfig(context.Background(), "tester"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(before, after) {
		t.Fatalf("unauthenticated request changed config")
	}
}

func TestNovelFetchWorkshopConfigHidesCredentialsAndPreservesOnEdit(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"appConfig": map[string]any{"ai": map[string]any{"api_key": "secret-value", "model": "old-model"}},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if get.Code != http.StatusOK || strings.Contains(get.Body.String(), "secret-value") {
		t.Fatalf("config leaked credential: status=%d body=%s", get.Code, get.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	app := payload["appConfig"].(map[string]any)
	ai := app["ai"].(map[string]any)
	if ai["api_key"] != nil || ai["api_key_configured"] != true {
		t.Fatalf("unexpected public ai config: %#v", ai)
	}
	put := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/config", map[string]any{
		"appConfig": map[string]any{"ai": map[string]any{"model": "new-model", "api_key": ""}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	saved, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	savedAI := saved["appConfig"].(map[string]any)["ai"].(map[string]any)
	if savedAI["api_key"] != "secret-value" || savedAI["model"] != "new-model" {
		t.Fatalf("edit should preserve credential: %#v", savedAI)
	}
	put = doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"appConfig": map[string]any{"ai": map[string]any{"model": "put-model", "api_key": ""}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
}

func TestNovelFetchWorkshopConfigPersistsSensitiveAIEnabledSwitch(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"sensitive_ai": map[string]any{"enabled": false},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/config", map[string]any{
		"sensitive_ai_enabled": true,
	}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("save status=%d body=%s", response.Code, response.Body.String())
	}
	saved, err := store.GetConfig(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	if mapOrEmpty(saved["sensitive_ai"])["enabled"] != true {
		t.Fatalf("sensitive AI switch was not persisted canonically: %#v", saved["sensitive_ai"])
	}
}

func TestNovelFetchWorkshopConfigUnwrapsMigratedSettings(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"appConfig": map[string]any{"workflow": map[string]any{"auto_classify_missing": true}},
		"settings": map[string]any{
			"appConfig": map[string]any{
				"workflow": map[string]any{"auto_fetch_original": true},
				"fetch":    map[string]any{"default_max_txt": 4000},
				"rewrite":  map[string]any{"max_ai_count": 5},
			},
			"platforms": []any{map[string]any{"id": "15", "name": "知乎付费"}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", get.Code, get.Body.String())
	}
	var response struct {
		AppConfig map[string]any   `json:"appConfig"`
		Platforms []map[string]any `json:"platforms"`
		Knowledge map[string]any   `json:"knowledge"`
		Summary   map[string]any   `json:"knowledge_summary"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	workflow := response.AppConfig["workflow"].(map[string]any)
	fetch := response.AppConfig["fetch"].(map[string]any)
	if workflow["auto_fetch_original"] != true || fetch["default_max_txt"] != float64(4000) || len(response.Platforms) != 1 {
		t.Fatalf("migrated settings were not unwrapped: %#v", response)
	}
	if len(response.Knowledge) == 0 || response.Summary["opening_phrases"] != float64(174) || response.Summary["rewrite_profiles"] != float64(8) || response.Summary["high_imitation_prompts"] != float64(1) {
		t.Fatalf("canonical config omitted the bundled knowledge catalog: %#v", response)
	}
}

func TestNovelFetchWorkshopConfigMergesRootSectionsIntoAppConfig(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{
		"workflow": map[string]any{"auto_fetch_original": true, "auto_rewrite_after_fetch": true},
		"fetch":    map[string]any{"default_max_txt": 4000},
		"rewrite":  map[string]any{"max_ai_count": 5},
		"appConfig": map[string]any{
			"workflow":  map[string]any{"auto_classify_missing": true},
			"work_form": map[string]any{"platform_id": "15"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", get.Code, get.Body.String())
	}
	var response struct {
		AppConfig map[string]any `json:"appConfig"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	workflow := response.AppConfig["workflow"].(map[string]any)
	fetch := response.AppConfig["fetch"].(map[string]any)
	rewrite := response.AppConfig["rewrite"].(map[string]any)
	if workflow["auto_fetch_original"] != true || workflow["auto_classify_missing"] != true || fetch["default_max_txt"] != float64(4000) || rewrite["max_ai_count"] != float64(5) {
		t.Fatalf("root config sections were dropped: %#v", response.AppConfig)
	}
}
