package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

// staleCredentialConnector 模拟目标站会话已过期：CheckSession 与 RefreshStyles
// 都返回登录页，用于验证状态接口与同步接口使用同一个凭证判定。
type staleCredentialConnector struct{}

func (c *staleCredentialConnector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, nil
}
func (c *staleCredentialConnector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return videomanagementconnector.ErrSessionExpired
}
func (c *staleCredentialConnector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}
func (c *staleCredentialConnector) Submit(context.Context, videomanagementconnector.Session, videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{}, nil
}
func (c *staleCredentialConnector) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{}, nil
}
func (c *staleCredentialConnector) RefreshStyles(context.Context, videomanagementconnector.Session) (string, error) {
	return `<html><title>管理员登录</title><input name="username"><input name="password"></html>`, nil
}

func TestWebSubmitStatusAndSyncShareExpiredCredential(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	usersDir := filepath.Join(t.TempDir(), "users")
	userDir := filepath.Join(usersDir, "tester")
	if err := os.MkdirAll(userDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userDir, "upload-target.json"), []byte(`{"cookie":"PHPSESSID=stale","targetUsername":"operator"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	store := novelfetchworkshop.NewMemoryStore()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}
	mux := http.NewServeMux()
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, usersDir, &staleCredentialConnector{})

	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/web-submit/environment", nil)
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("environment status=%d body=%s", res.Code, res.Body.String())
	}
	var env struct {
		OK         bool `json:"ok"`
		Credential struct {
			State   string `json:"state"`
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"credential"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.OK || env.Credential.State != TargetCredentialExpired || env.Credential.Code != ErrorCodeTargetSessionExpired {
		t.Fatalf("environment credential=%#v ok=%v body=%s", env.Credential, env.OK, res.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/sync-styles", strings.NewReader(`{"persist":false}`))
	req.Header.Set("Authorization", "Bearer test-session")
	res = httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("sync-styles status=%d body=%s", res.Code, res.Body.String())
	}
	var syncErr struct {
		Code       string `json:"code"`
		Error      string `json:"error"`
		Credential struct {
			State string `json:"state"`
		} `json:"credential"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &syncErr); err != nil {
		t.Fatal(err)
	}
	if syncErr.Code != ErrorCodeTargetSessionExpired || syncErr.Credential.State != TargetCredentialExpired || !strings.Contains(syncErr.Error, "过期") {
		t.Fatalf("sync-styles body=%s", res.Body.String())
	}
}

func TestWebSubmitStatusDistinguishesIncompleteCredential(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"username": "operator"}}}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/web-submit/environment", nil)
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	var env struct {
		Credential struct {
			State string `json:"state"`
			Code  string `json:"code"`
		} `json:"credential"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Credential.State != TargetCredentialIncomplete || env.Credential.Code != ErrorCodeTargetCredentialIncomplete {
		t.Fatalf("credential=%#v body=%s", env.Credential, res.Body.String())
	}
}

func TestStyleNameDiffReportsRealAddsAndRemovals(t *testing.T) {
	old := []string{"古风虐文", "现代甜文", "已下线风格"}
	fresh := []string{"古风虐文", "现代甜文", "新增风格"}
	if added := styleNameDiff(fresh, old); !reflect.DeepEqual(added, []string{"新增风格"}) {
		t.Fatalf("added=%#v", added)
	}
	if removed := styleNameDiff(old, fresh); !reflect.DeepEqual(removed, []string{"已下线风格"}) {
		t.Fatalf("removed=%#v", removed)
	}
	if unchanged := styleNameDiff(old, old); len(unchanged) != 0 {
		t.Fatalf("unchanged diff=%#v", unchanged)
	}
}

func TestTargetConfigItemsAcceptsNestedListResponse(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(`{"success":true,"data":{"list":[{"id":"1"}]}}`), &payload); err != nil {
		t.Fatal(err)
	}
	items, ok := targetConfigItems(payload)
	if !ok || len(items) != 1 {
		t.Fatalf("items=%#v ok=%v", items, ok)
	}
}

func TestTargetConfigItemsAcceptsResultArrayResponse(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(`{"code":0,"result":[{"id":"1"}]}`), &payload); err != nil {
		t.Fatal(err)
	}
	items, ok := targetConfigItems(payload)
	if !ok || len(items) != 1 {
		t.Fatalf("items=%#v ok=%v", items, ok)
	}
}

func TestLegacyBatchRewriteTaskIncludesStoredSubmitReceipt(t *testing.T) {
	task := legacyBatchRewriteTask(novelfetchworkshop.Document{
		BookID:    "book-submit",
		UpdatedAt: "2026-10-12T05:19:27Z",
		Meta: map[string]any{
			"bookName":               "已提交的书",
			"siteSubmitStatus":       "submitted",
			"siteSubmitDoneVersions": []any{"original", "ai1"},
		},
	})
	if task["site_submit_status"] != "submitted" {
		t.Fatalf("site_submit_status=%#v", task["site_submit_status"])
	}
	confirmed, ok := task["site_submit_confirmed_versions"].([]string)
	if !ok || len(confirmed) != 2 {
		t.Fatalf("confirmed versions=%#v", task["site_submit_confirmed_versions"])
	}
	if task["created_at"] != "2026-10-12T05:19:27Z" {
		t.Fatalf("created_at=%#v", task["created_at"])
	}
}

func TestBatchRewriteWebSubmitHistoryReadsGoDocumentLogs(t *testing.T) {
	now := time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-1",
		Meta:   map[string]any{"bookName": "测试书", "platformName": "番茄"},
		Logs: []any{map[string]any{
			"event": "site_submit", "status": "submitted", "version": "ai1", "group_id": "group-1",
			"time": now.Format(time.RFC3339), "remote_receipt": map[string]any{"verified": true, "upload_success_count": 2},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	storeConfig := map[string]any{"site_submit_groups": []any{map[string]any{"group_id": "group-1", "updated_at": now.Format(time.RFC3339)}}}
	if err := store.PutConfig(context.Background(), "tester", storeConfig); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, filepath.Join(t.TempDir(), "users"))
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/web-submit/history", nil)
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	records, ok := payload["records"].([]any)
	if !ok || len(records) != 1 {
		t.Fatalf("records=%#v", payload["records"])
	}
	if records[0].(map[string]any)["result"] != "视频管理系统已确认接收 2 个文件" {
		t.Fatalf("record=%#v", records[0])
	}
	groups, ok := payload["groups"].([]any)
	if !ok || len(groups) != 1 {
		t.Fatalf("groups=%#v", payload["groups"])
	}
}

func TestCanonicalNovelFetchWebSubmitConfigUsesTheSharedGoHandler(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true}}}); err != nil {
		t.Fatal(err)
	}
	accounts := &authTestStore{
		account: auth.Account{Username: "tester", Active: true},
		session: auth.Session{Username: "tester", Token: "canonical-session", ExpiresAt: now.Add(time.Hour)},
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, AuthStore: accounts})
	// The canonical web-submit alias only accepts a real Bearer session; signed
	// headers are no longer honoured on that hop.
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/web-submit/config", nil)
	req.Header.Set("Authorization", "Bearer canonical-session")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("canonical config status=%d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	settings, ok := payload["settings"].(map[string]any)
	if !ok || settings["enabled"] != true {
		t.Fatalf("canonical config payload=%#v", payload)
	}
}

func TestPublicBatchWebSubmitCompletesEveryVersionProfileBinding(t *testing.T) {
	public := publicBatchWebSubmit(map[string]any{"profile_bindings": map[string]any{"ai1": "profile-ai1"}})
	bindings := mapOrEmpty(public["profile_bindings"])
	for _, version := range []string{"original", "ai1", "ai2", "ai3", "ai4", "ai5"} {
		if _, ok := bindings[version]; !ok {
			t.Fatalf("missing %s binding: %#v", version, bindings)
		}
	}
	if bindings["ai1"] != "profile-ai1" || bindings["ai2"] != "" {
		t.Fatalf("unexpected bindings: %#v", bindings)
	}
}

func TestLegacyNovelFetchEndpointIsNoLongerMounted(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	response := doLegacyBatchRewriteRequest(t, handler, http.MethodPost, "/api/novel-fetch", map[string]any{"platform": 15, "bookIds": []string{"2091926811672429633"}, "maxTxt": 4000}, now, "test-secret")
	if response.Code != http.StatusNotFound {
		t.Fatalf("legacy novel-fetch endpoint is still mounted: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestLegacyBatchRewriteConfigIncludesTextModelAndParserDefaults(t *testing.T) {
	payload := legacyBatchRewriteConfig(map[string]any{
		"appConfig": map[string]any{
			"modelCatalog":  []any{map[string]any{"id": "managed-text", "kind": "text", "modelId": "gemini", "credential": "secret", "enabled": true}},
			"text_model_id": "managed-text",
		},
	}, false)
	app := payload["app_config"].(map[string]any)
	if app["text_model_id"] != "managed-text" {
		t.Fatalf("text model not preserved: %#v", app["text_model_id"])
	}
	models := app["modelCatalog"].([]any)
	if len(models) != 1 || models[0].(map[string]any)["credential"] != nil || models[0].(map[string]any)["hasCredential"] != true {
		t.Fatalf("model catalog was not safely exposed: %#v", models)
	}
	if got := len(payload["parse_modes"].([]any)); got != 7 {
		t.Fatalf("parse modes=%d", got)
	}
	if got := len(payload["column_presets"].([]any)); got != 5 {
		t.Fatalf("column presets=%d", got)
	}
	webSubmit := payload["web_submit"].(map[string]any)
	if webSubmit["submit_mode"] != "version" || webSubmit["target"] != "two.121w.com" {
		t.Fatalf("web submit defaults=%#v", webSubmit)
	}
}

func TestLegacyBatchRewriteConfigRestoresKnowledgeCatalog(t *testing.T) {
	payload := legacyBatchRewriteConfig(map[string]any{
		"appConfig": map[string]any{
			"knowledge": map[string]any{
				"opening_phrases":   map[string]any{"items": []any{}},
				"high_imitation":    map[string]any{"prompts": []any{}, "references": []any{}},
				"rewrite_templates": map[string]any{"profiles": []any{}, "temporary_instructions": []any{}},
			},
		},
	}, false)
	summary := payload["knowledge_summary"].(map[string]any)
	if summary["opening_phrases"] != 174 || summary["high_imitation_prompts"] != 1 || summary["rewrite_profiles"] != 8 {
		t.Fatalf("knowledge catalog was not restored: %#v", summary)
	}
}

func TestLegacyBatchRewriteConfigExposesSensitiveGroupsFromAppConfig(t *testing.T) {
	payload := legacyBatchRewriteConfig(map[string]any{
		"appConfig": map[string]any{
			"sensitive": map[string]any{
				"groups": []any{map[string]any{
					"name": "实际词库", "enabled": true,
					"rules": []any{map[string]any{"find": "玩", "replace": "", "enabled": true}},
				}},
			},
		},
		"sensitive": map[string]any{},
	}, false)
	sensitive := payload["sensitive"].(map[string]any)
	groups := sensitive["groups"].([]any)
	if len(groups) != 1 {
		t.Fatalf("sensitive groups=%#v", sensitive)
	}
	group := groups[0].(map[string]any)
	if group["name"] != "实际词库" {
		t.Fatalf("sensitive group was not exposed from appConfig: %#v", group)
	}
}

func TestBatchRewriteTaskDiagnosticsReadGoDocument(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-1", Meta: map[string]any{"bookName": "测试书"},
		Logs: []any{map[string]any{"event": "sensitive", "status": "done"}},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, t.TempDir())
	for _, path := range []string{"sensitive-log", "rules-trace", "site-submit-log"} {
		req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/tasks/book-1/"+path, nil)
		SignBridgeRequest(req, "tester", false, now, "secret")
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", path, res.Code, res.Body.String())
		}
	}
}

func TestBatchRewriteSensitiveDiagnosticUsesLegacyFieldsAndCounts(t *testing.T) {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-sensitive", Gender: "女频", ProcessedChars: 3919, OriginalRawChars: 4334,
		Meta: map[string]any{
			"bookName": "敏感测试书", "platformId": "15", "sensitiveStatus": "done",
			"sensitiveReplaced": 44, "updatedAt": now.Format(time.RFC3339),
		},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/tasks/book-sensitive/sensitive-log", nil)
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("diagnostic status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	meta := payload["meta"].(map[string]any)
	if meta["book_id"] != "book-sensitive" || meta["book_name"] != "敏感测试书" {
		t.Fatalf("legacy meta missing: %#v", meta)
	}
	if meta["sensitive_hit_count"] != float64(44) || meta["sensitive_fixed_count"] != float64(44) {
		t.Fatalf("legacy sensitive counts missing: %#v", meta)
	}
	fixed := payload["sensitive_fixed"].(map[string]any)
	if fixed["fixed_count"] != float64(44) {
		t.Fatalf("sensitive_fixed=%#v", fixed)
	}
}

func TestBatchRewriteTaskListFlattensGoDocumentFields(t *testing.T) {
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "2010030448764536200", Gender: "女频", BookInfo: novelfetchworkshop.BookInfo{Genre: "现代女主"},
		ProcessedChars: 3919, OriginalRawChars: 4334, UpdatedAt: now.Format(time.RFC3339),
		// Old in-flight metadata can remain at ai_processing even after the
		// finished AI body has been persisted. The compatibility response must
		// report the stored version as completed so the processing page does not
		// disagree with the task page.
		Meta:     map[string]any{"bookName": "非要给我打青霉素的实习护士", "platformId": "15", "originalStatus": "done", "sensitiveStatus": "done", "sensitiveReplaced": 44, "aiStatus": "ai_processing", "aiGeneratedCount": 1},
		Versions: map[string]any{"ai1": "已生成的改文"},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/tasks", nil)
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Tasks) != 1 {
		t.Fatalf("tasks=%#v", payload.Tasks)
	}
	task := payload.Tasks[0]
	for key, want := range map[string]any{"book_name": "非要给我打青霉素的实习护士", "platform_id": "15", "platform_name": "知乎付费", "gender": "女频", "style": "现代女主", "original_status": "done", "ai_status": "ai_done"} {
		if task[key] != want {
			t.Fatalf("%s=%#v want=%#v task=%#v", key, task[key], want, task)
		}
	}
	if task["sensitive_hit_count"] != float64(44) {
		t.Fatalf("sensitive_hit_count=%#v", task["sensitive_hit_count"])
	}
	if got, want := task["ai_files"], []any{"ai1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ai_files=%#v want=%#v", got, want)
	}
}

func TestBatchRewriteTaskListPreservesSensitiveAIProvenance(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-ai-sensitive",
		Meta: map[string]any{
			"bookName":       "AI 敏感测试",
			"sensitiveHits":  map[string]any{"mode": "ai_group", "hit_count": 2, "hits": []any{}},
			"sensitiveFixed": map[string]any{"mode": "ai_group", "fixed_count": 1, "failed_count": 1, "items": []any{}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", AuthStore: accountStore}, store, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/tasks", nil)
	SignBridgeRequest(req, "tester", false, time.Now(), "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Tasks) != 1 {
		t.Fatalf("tasks=%#v", payload.Tasks)
	}
	task := payload.Tasks[0]
	if task["sensitive_mode"] != "ai_group" || task["sensitive_hit_count"] != float64(2) || task["sensitive_fixed_count"] != float64(1) || task["sensitive_failed_count"] != float64(1) {
		t.Fatalf("sensitive provenance overwritten: %#v", task)
	}
}

func TestBatchRewriteIssuesReadsGoDocumentLogs(t *testing.T) {
	now := time.Date(2026, 10, 11, 11, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:    "book-issue",
		UpdatedAt: now.Format(time.RFC3339),
		Meta:      map[string]any{"bookName": "问题书"},
		Logs: []any{map[string]any{
			"event": "ai_generate_failed",
			"time":  now.Format(time.RFC3339),
			"data":  map[string]any{"status": "failed", "error": "文本模型未登录"},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/issues", nil)
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("issues status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Entries []map[string]any `json:"entries"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Entries) != 1 {
		t.Fatalf("entries=%#v", payload.Entries)
	}
	entry := payload.Entries[0]
	if entry["kind"] != "batch-rewrite.ai" || entry["status"] != float64(401) || entry["path"] != "/novel-fetch" {
		t.Fatalf("entry=%#v", entry)
	}
	legacyAlias := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/issues", nil)
	SignBridgeRequest(legacyAlias, "tester", false, now, "secret")
	aliasResponse := httptest.NewRecorder()
	mux.ServeHTTP(aliasResponse, legacyAlias)
	if aliasResponse.Code != http.StatusOK {
		t.Fatalf("novel-fetch-workshop issues alias status=%d body=%s", aliasResponse.Code, aliasResponse.Body.String())
	}
}

func TestBatchRewriteDeleteSkipsRunningTasks(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	for bookID, status := range map[string]string{"book-running": "running", "book-done": "done"} {
		if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
			BookID: bookID,
			Meta:   map[string]any{"bookName": bookID, "status": status},
		}); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, t.TempDir())
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/tasks/batch-delete", strings.NewReader(`{"ids":["book-running","book-done"]}`))
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Deleted int      `json:"deleted"`
		Skipped []string `json:"skipped"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Deleted != 1 || !reflect.DeepEqual(payload.Skipped, []string{"book-running"}) {
		t.Fatalf("payload=%#v body=%s", payload, res.Body.String())
	}
	if _, err := store.GetDocument(context.Background(), "tester", "book-running"); err != nil {
		t.Fatalf("running task must be kept: %v", err)
	}
	if _, err := store.GetDocument(context.Background(), "tester", "book-done"); err == nil {
		t.Fatal("finished task must be deleted")
	}
}

func TestIsTargetLoginPage(t *testing.T) {
	loginPage := `<!DOCTYPE html><html><head><title>管理员登录</title></head><body>
<form method="post" action="login.php"><input name="username"><input type="password" name="password"></form></body></html>`
	if !isTargetLoginPage(loginPage) {
		t.Fatal("管理员登录页未被识别为登录页")
	}
	stylePage := `<select id="style"><option value="1">古风虐文</option></select>`
	if isTargetLoginPage(stylePage) {
		t.Fatal("风格页面被误判为登录页")
	}
}
