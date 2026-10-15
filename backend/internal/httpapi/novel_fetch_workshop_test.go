package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

func signNovelFetchRequest(req *http.Request, username string, isOwner bool, now time.Time, secret string) {
	issuedAt := strconv.FormatInt(now.Unix(), 10)
	owner := strconv.FormatBool(isOwner)
	payload := username + "\n" + issuedAt + "\n" + owner + "\n" + req.Method + "\n" + req.URL.Path
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	req.Header.Set(HeaderUsername, username)
	req.Header.Set(HeaderIsOwner, owner)
	req.Header.Set(HeaderIssuedAt, issuedAt)
	req.Header.Set(HeaderSignature, hex.EncodeToString(mac.Sum(nil)))
}

func doNovelFetchRequest(t *testing.T, handler http.Handler, method, path string, body any, now time.Time, secret string) *httptest.ResponseRecorder {
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
	signNovelFetchRequest(req, "tester", true, now, secret)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestNovelFetchWorkshopRealtimeStatusIsAvailableOnCanonicalRoute(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-realtime-status",
		Meta:   map[string]any{"aiStatus": "running"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/realtime/status", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("canonical realtime status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Counts map[string]int `json:"counts"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Counts["running"] != 1 {
		t.Fatalf("counts=%#v, want one running task", payload.Counts)
	}
}

func TestNovelFetchWorkshopTaskListBackfillsPlatformAndGender(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:   "2093734568914334892",
		Gender:   "待确认",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "都市"},
		Meta:     map[string]any{"platformId": "1", "originalStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("tasks status=%d body=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"platform_name":"黑岩付费"`)) {
		t.Fatalf("platform name was not backfilled: %s", response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"gender":"男频"`)) || !bytes.Contains(response.Body.Bytes(), []byte(`"classify_status":"classified"`)) {
		t.Fatalf("gender was not backfilled: %s", response.Body.String())
	}
}

func TestNovelFetchWorkshopTaskListBackfillsDeterministicStyle(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:   "book-style-backfill",
		Gender:   "男频",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "古风通用"},
		Meta:     map[string]any{"platformId": "1", "originalStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("tasks status=%d body=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(`"style":"古风通用"`)) {
		t.Fatalf("deterministic style was not backfilled: %s", response.Body.String())
	}
}

func TestNovelFetchWorkshopSparsePutPreservesExistingGenderSource(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-existing-gender", Gender: "女频", GenderSource: novelfetchworkshop.GenderAI, GenderReason: "AI 判断",
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-existing-gender", map[string]any{"bookInfo": map[string]string{"category": "女频"}}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("sparse put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-existing-gender", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.Gender != "女频" || document.GenderSource != novelfetchworkshop.GenderAI {
		t.Fatalf("sparse put changed existing gender source: %#v", document)
	}
}

func TestNovelFetchWorkshopPutResolvesPendingGenderFromBookInfo(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-pending-gender", Gender: "待确认",
		BookInfo: novelfetchworkshop.BookInfo{Category: "男频", Genre: "都市"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-pending-gender", map[string]any{
		"bookInfo": map[string]string{"category": "男频", "genre": "都市"},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	stored, err := store.GetDocument(context.Background(), "tester", "book-pending-gender")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Gender != "男频" || stored.GenderSource != novelfetchworkshop.GenderBookInfo {
		t.Fatalf("pending gender was not resolved during put: %#v", stored)
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-pending-gender", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.Gender != "男频" || document.GenderSource != novelfetchworkshop.GenderBookInfo {
		t.Fatalf("pending gender was not retained from book info: %#v", document)
	}
}

func TestNovelFetchWorkshopBatchRetryUsesGoStore(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID:        "already-done",
		Original:      "已有正文",
		ProcessedText: "已有正文",
		Meta:          map[string]any{"originalStatus": "done", "aiStatus": "done"},
	}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	response := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/batch-retry", map[string]any{"ids": []string{"already-done"}}, now, secret)
	if response.Code != http.StatusOK {
		t.Fatalf("batch retry status=%d body=%s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true || payload["skipped"] != float64(1) || payload["retried"] != float64(0) {
		t.Fatalf("unexpected batch retry result: %#v", payload)
	}
}

func TestNovelFetchWorkshopPutPreservesExistingPersistentFields(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-preserve", map[string]any{
		"originalRaw": "完整原文", "processedText": "处理正文", "bookInfo": map[string]string{"category": "男频", "genre": "都市"},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("initial put status=%d body=%s", put.Code, put.Body.String())
	}
	put = doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-preserve", map[string]any{"gender": "女频"}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("sparse put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-preserve", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "完整原文" || document.ProcessedText != "处理正文" || document.BookInfo.Category != "男频" || document.Gender != "女频" {
		t.Fatalf("sparse put lost existing fields: %#v", document)
	}
}

func TestNovelFetchWorkshopBridgeRoundTrip(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{
		BridgeSecret:    secret,
		Now:             func() time.Time { return now },
		NovelFetchStore: store,
	})

	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-1", map[string]any{
		"meta":        map[string]any{"bookName": "测试小说", "targetVersions": []any{"original", "ai1", "ai3"}},
		"original":    "正文",
		"originalRaw": "原始正文",
		"versions":    map[string]any{"ai1": "AI1", "ai3": "AI3"},
		"logs":        []any{map[string]any{"event": "saved"}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}

	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-1", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document["bookId"] != "book-1" || document["original"] != "正文" {
		t.Fatalf("unexpected document: %#v", document)
	}
	versions := document["versions"].(map[string]any)
	if versions["ai1"] != "AI1" || versions["ai3"] != "AI3" {
		t.Fatalf("sparse versions lost: %#v", versions)
	}

	list := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var listed struct {
		Tasks []novelfetchworkshop.Document `json:"tasks"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 1 || listed.Tasks[0].BookID != "book-1" {
		t.Fatalf("unexpected tasks: %#v", listed.Tasks)
	}

	configPut := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"settings": map[string]any{"fetch": map[string]any{"timeout_seconds": 9}},
	}, now, secret)
	if configPut.Code != http.StatusOK {
		t.Fatalf("config put status=%d body=%s", configPut.Code, configPut.Body.String())
	}
	configGet := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if configGet.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", configGet.Code, configGet.Body.String())
	}
	var config struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(configGet.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	if config.Settings["fetch"].(map[string]any)["timeout_seconds"] != float64(9) {
		t.Fatalf("config did not round trip: %#v", config.Settings)
	}

	deleted := doNovelFetchRequest(t, handler, http.MethodDelete, "/api/novel-fetch-workshop/tasks", map[string]any{"ids": []string{"book-1"}}, now, secret)
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	missing := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-1", nil, now, secret)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", missing.Code, missing.Body.String())
	}
}

func TestRunAuditPersistsSelectedAndResolvedModelWithoutCredential(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})

	put := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/runs", map[string]any{
		"runId": "run-1", "bookId": "book-1", "stage": "rewrite", "status": "succeeded", "attempts": 1,
		"textModelId": "gemini-3", "modelId": "gemini-3", "modelDisplayName": "Gemini 3", "api_key": "must-not-persist",
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}

	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/runs?bookId=book-1", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	if bytes.Contains(get.Body.Bytes(), []byte("must-not-persist")) || bytes.Contains(bytes.ToLower(get.Body.Bytes()), []byte("api_key")) {
		t.Fatalf("audit leaked secret: %s", get.Body.String())
	}
	var response struct {
		Runs []novelfetchworkshop.RunRecord `json:"runs"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Runs) != 1 || response.Runs[0].TextModelID != "gemini-3" || response.Runs[0].ModelID != "gemini-3" {
		t.Fatalf("unexpected audit records: %#v", response.Runs)
	}
}

func TestNovelFetchWorkshopRejectsUnsignedRequest(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", NovelFetchStore: store})
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestNovelFetchWorkshopAcceptsBrowserSessionToken(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := novelfetchworkshop.NewMemoryStore()
	authStore := &authTestStore{
		account: auth.Account{Username: "alice", Active: true, IsOwner: true},
		session: auth.Session{Token: "browser-token", Username: "alice", ExpiresAt: now.Add(time.Hour)},
	}
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "unused-bridge-secret",
		Now:             func() time.Time { return now },
		AuthStore:       authStore,
		NovelFetchStore: store,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	req.Header.Set("Authorization", "Bearer browser-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("browser session status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestNovelFetchWorkshopConfigMatchesFrontendContract(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"appConfig": map[string]any{"workflow": map[string]any{"auto_fetch_original": true}},
		"platforms": []any{map[string]any{"id": "2", "name": "起点"}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("config put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", get.Code, get.Body.String())
	}
	var response struct {
		AppConfig map[string]any   `json:"appConfig"`
		Platforms []map[string]any `json:"platforms"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	workflow, ok := response.AppConfig["workflow"].(map[string]any)
	if !ok || workflow["auto_fetch_original"] != true || len(response.Platforms) != 1 || response.Platforms[0]["name"] != "起点" {
		t.Fatalf("frontend config contract mismatch: %#v", response)
	}
	platforms := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/platforms", nil, now, secret)
	if platforms.Code != http.StatusOK || !bytes.Contains(platforms.Body.Bytes(), []byte(`"name":"起点"`)) {
		t.Fatalf("platforms status=%d body=%s", platforms.Code, platforms.Body.String())
	}
}

func TestNovelFetchWorkshopProcessCreatesTasksFromPastedList(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	process := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{
		"inputText": "书籍ID\t书名\t男女频\n12345678901\t第一本\t男频\n12345678901\t第一本\t男频\n22345678901\t第二本\t女频",
	}, now, secret)
	if process.Code != http.StatusOK {
		t.Fatalf("process status=%d body=%s", process.Code, process.Body.String())
	}
	var response struct {
		UniqueTasks    int                           `json:"uniqueTasks"`
		DuplicateCount int                           `json:"duplicateCount"`
		Tasks          []novelfetchworkshop.Document `json:"tasks"`
	}
	if err := json.Unmarshal(process.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.UniqueTasks != 2 || response.DuplicateCount != 1 || len(response.Tasks) != 2 {
		t.Fatalf("unexpected process response: %#v", response)
	}
}

func TestNovelFetchWorkshopCurrentBatchOnlyContainsLatestInput(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	first := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{"inputText": "12345678901\t旧批次"}, now, secret)
	if first.Code != http.StatusOK {
		t.Fatalf("first process status=%d body=%s", first.Code, first.Body.String())
	}
	second := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{"inputText": "22345678901\t当前批次"}, now, secret)
	if second.Code != http.StatusOK {
		t.Fatalf("second process status=%d body=%s", second.Code, second.Body.String())
	}
	current := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/batches/current", nil, now, secret)
	if current.Code != http.StatusOK {
		t.Fatalf("current status=%d body=%s", current.Code, current.Body.String())
	}
	var response struct {
		Batch struct {
			TaskIDs []string `json:"taskIds"`
		} `json:"batch"`
	}
	if err := json.Unmarshal(current.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Batch.TaskIDs) != 1 || response.Batch.TaskIDs[0] != "22345678901" {
		t.Fatalf("current batch leaked historical tasks: %#v", response.Batch.TaskIDs)
	}
}

func TestNovelFetchWorkshopProcessPreviewDoesNotPersistTasks(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	preview := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/preview", map[string]any{
		"input_text": "书籍ID\t书名\t男女频\n12345678901\t预览一本\t男频\n12345678901\t预览一本\t男频\n22345678901\t预览二本\t女频",
		"max_txt":    4000,
	}, now, secret)
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	var response struct {
		Parsed         int              `json:"parsed"`
		UniqueTasks    int              `json:"unique_tasks"`
		DuplicateCount int              `json:"duplicate_count"`
		Tasks          []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(preview.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Parsed != 3 || response.UniqueTasks != 2 || response.DuplicateCount != 1 || len(response.Tasks) != 2 {
		t.Fatalf("unexpected preview response: %#v", response)
	}
	documents, err := store.ListDocuments(context.Background(), "tester")
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 0 {
		t.Fatalf("preview must not persist tasks: %d", len(documents))
	}
}

func TestNovelFetchWorkshopProcessStartCanBePolled(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	start := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process/start", map[string]any{"inputText": "12345678901\t可轮询任务"}, now, secret)
	if start.Code != http.StatusOK {
		t.Fatalf("start status=%d body=%s", start.Code, start.Body.String())
	}
	var job struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	if job.ID == "" || job.Status != "done" {
		t.Fatalf("unexpected job=%+v", job)
	}
	polled := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/process/jobs/"+job.ID, nil, now, secret)
	if polled.Code != http.StatusOK || !bytes.Contains(polled.Body.Bytes(), []byte(`"status":"done"`)) {
		t.Fatalf("poll status=%d body=%s", polled.Code, polled.Body.String())
	}
	latest := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/process/jobs/latest?limit=5", nil, now, secret)
	if latest.Code != http.StatusOK || !bytes.Contains(latest.Body.Bytes(), []byte(job.ID)) {
		t.Fatalf("latest status=%d body=%s", latest.Code, latest.Body.String())
	}
}

func TestNovelFetchWorkshopFetchOriginalPersistsRawAndProcessedText(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("bookid") != "12345678901" || req.URL.Query().Get("platform") != "15" {
			t.Fatalf("unexpected upstream query: %s", req.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"第一行\n第二行\n第三行","bookinfo":{"category":"男频","genre":"都市"}}`))
	}))
	defer upstream.Close()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345678901", MaxTxt: 4, Meta: map[string]any{"platformId": "15"}}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}})
	fetch := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345678901/fetch", map[string]any{"maxTxt": 4}, now, secret)
	if fetch.Code != http.StatusOK {
		t.Fatalf("fetch status=%d body=%s", fetch.Code, fetch.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "第一行\n第二行\n第三行" || document.ProcessedText != "第一行\n" || document.OriginalRawChars <= document.ProcessedChars || document.Gender != "男频" || document.Meta["sensitiveStatus"] != "done" {
		t.Fatalf("fetch persistence mismatch: %#v", document)
	}
}

func TestNovelFetchWorkshopRestoreOriginalReprocessesStoredRawText(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "完整原文内容", ProcessedText: "旧版本", Original: "旧版本", MaxTxt: 4}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	restore := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345678901/restore-original", nil, now, secret)
	if restore.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", restore.Code, restore.Body.String())
	}
	document, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil {
		t.Fatal(err)
	}
	if document.OriginalRaw != "完整原文内容" || document.ProcessedText != "完整原文" || document.Meta["originalStatus"] != "restored" {
		t.Fatalf("restore mismatch: %#v", document)
	}
}

func TestNovelFetchMemoryStoreIsOwnerScoped(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "alice", novelfetchworkshop.Document{BookID: "same", Original: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "bob", novelfetchworkshop.Document{BookID: "same", Original: "bob"}); err != nil {
		t.Fatal(err)
	}
	a, _ := store.GetDocument(context.Background(), "alice", "same")
	b, _ := store.GetDocument(context.Background(), "bob", "same")
	if a.Original != "alice" || b.Original != "bob" {
		t.Fatalf("owner isolation broken: alice=%q bob=%q", a.Original, b.Original)
	}
}
