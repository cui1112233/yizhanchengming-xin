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

func TestWorkshopOperationsRejectUnsupportedAndHideOtherOwners(t *testing.T) {
	mux := http.NewServeMux()
	registerNovelFetchLegacyWorkflowRoutes(mux, novelfetchworkshop.NewMemoryStore(), novelfetchworkshop.NewMemoryQueue(), "", nil)
	call := func(owner, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, BridgeIdentity{Username: owner}))
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		return res
	}
	base := "/api/novel-fetch-workshop/web-submit/operations"
	res := call("alice", "POST", base, `{"kind":"not-implemented"}`)
	if res.Code != 400 {
		t.Fatalf("unsupported operation claimed success: %d %s", res.Code, res.Body.String())
	}
	res = call("alice", "POST", base, `{"kind":"configs"}`)
	var out map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &out)
	id, _ := out["id"].(string)
	if id == "" {
		t.Fatalf("missing operation id: %s", res.Body.String())
	}
	if got := call("alice", "GET", base+"/"+id, ""); got.Code != 200 {
		t.Fatalf("owner lost operation: %s", got.Body.String())
	}
	if got := call("bob", "GET", base+"/"+id, ""); got.Code != 404 {
		t.Fatalf("other owner's operation exposed: %s", got.Body.String())
	}
}

func TestWorkshopRulePreviewUsesSavedAccountRules(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	settings := map[string]any{"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": false}}}
	_ = store.PutConfig(context.Background(), "tester", settings)
	now := time.Unix(1800000000, 0)
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", Now: func() time.Time { return now }, NovelFetchStore: store})
	res := doNovelFetchRequest(t, handler, "POST", "/api/novel-fetch-workshop/rules/preview", map[string]any{"text": "你好！！！", "scope": "original"}, now, "test-secret")
	var out map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &out)
	if res.Code != 200 || out["text"] != "你好！！！" {
		t.Fatalf("saved disable rule ignored: %d %s", res.Code, res.Body.String())
	}
	res = doNovelFetchRequest(t, handler, "POST", "/api/novel-fetch-workshop/rules/preview", map[string]any{"text": "你好！！！", "scope": "original", "layoutConfig": map[string]any{"apply_to_original": true}}, now, "test-secret")
	_ = json.Unmarshal(res.Body.Bytes(), &out)
	if res.Code != 200 || out["text"] != "你好！" {
		t.Fatalf("draft override ignored: %d %s", res.Code, res.Body.String())
	}
	saved, _ := store.GetConfig(context.Background(), "tester")
	if mapOrEmpty(mapOrEmpty(saved["appConfig"])["layout"])["apply_to_original"] != false {
		t.Fatal("preview saved an unsaved override")
	}
}

func TestWorkshopFlatConfigSaveUpdatesActualRuntimeView(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true}, "rewrite": map[string]any{"default_ai_count": 1}}})
	now := time.Unix(1800000000, 0)
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", Now: func() time.Time { return now }, NovelFetchStore: store})
	res := doNovelFetchRequest(t, handler, "POST", "/api/novel-fetch-workshop/config", map[string]any{"layout": map[string]any{"apply_to_original": false}, "rewrite": map[string]any{"default_ai_count": 3}, "web_submit": map[string]any{"submit_versions": []string{"ai3"}}}, now, "test-secret")
	if res.Code != 200 {
		t.Fatalf("save failed: %d %s", res.Code, res.Body.String())
	}
	saved, _ := store.GetConfig(context.Background(), "tester")
	result, _, _ := novelfetchworkshop.ProcessConfiguredText("你好！！！", "original", saved)
	if result != "你好！！！" || workshopAIRequestedCount(novelfetchworkshop.Document{}, saved, 0) != 3 || len(anyStrings(mapOrEmpty(mapOrEmpty(saved["appConfig"])["web_submit"])["submit_versions"])) != 1 {
		t.Fatal("saved form differs from execution config")
	}
}
