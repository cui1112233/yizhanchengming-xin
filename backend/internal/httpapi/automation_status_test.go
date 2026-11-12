package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestFileAutomationStatusReaderMatchesNodePublicShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"version":1,"jobs":{"owner%3Aalice":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"running","runMode":"video_no_submit","counts":{"total":1,"running":1},"books":{"book-1":{"bookId":"book-1","title":"第一本","status":"running","stage":"video","message":"执行中","error":"","attempts":1,"retryCount":0,"retryRequested":false},"secret":{"apiKey":"must-not-leak"}}}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	reader := NewFileAutomationStatusReader(path)
	status, err := reader.Status("alice", "batch-1")
	if err != nil || status["id"] != "job-1" {
		t.Fatalf("status=%+v err=%v", status, err)
	}
	books, ok := status["books"].([]any)
	if !ok || len(books) != 2 {
		t.Fatalf("books=%#v", status["books"])
	}
	if strings.Contains(string(mustJSON(status)), "apiKey") {
		t.Fatal("private runtime fields leaked into public automation status")
	}
}

func TestFileAutomationStatusReaderUsesNewestMatchingRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"jobs":{"old":{"owner":"alice","batchId":"batch-1","id":"old-run","state":"failed","updatedAt":"2026-10-14T12:10:33Z","books":{}},"new":{"owner":"alice","batchId":"batch-1","id":"new-run","state":"completed","updatedAt":"2026-10-14T12:44:22Z","books":{}}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := NewFileAutomationStatusReader(path).Status("alice", "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	if status["id"] != "new-run" || status["state"] != "completed" {
		t.Fatalf("reader returned stale run: %#v", status)
	}
}

func TestFileAutomationStatusReaderReadsSplitAutomationLayout(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "batch-factory-v11-automation.json")
	jobsDir := filepath.Join(root, "automation", "jobs")
	if err := os.MkdirAll(jobsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "automation", "index.json"), []byte(`{"jobs":[{"id":"job-split"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	job := `{"owner":"alice","batchId":"batch-1","id":"job-split","state":"running","runMode":"full_submit","counts":{"total":1,"running":1},"books":{"book-1":{"bookId":"book-1","title":"第一本","status":"running","stage":"comic_images","message":"出图中","logs":[]}}}`
	if err := os.WriteFile(filepath.Join(jobsDir, "job-split.json"), []byte(job), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := NewFileAutomationStatusReader(path).Status("alice", "batch-1")
	if err != nil || status["id"] != "job-split" {
		t.Fatalf("split status=%+v err=%v", status, err)
	}
	books, ok := status["books"].([]any)
	if !ok || len(books) != 1 {
		t.Fatalf("split books=%#v", status["books"])
	}
}

func TestFileAutomationStatusReaderHidesRetryAtAfterTerminalFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"jobs":{"job":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"failed","books":{"book-1":{"bookId":"book-1","status":"failed","stage":"comic_images","retryAt":"2026-10-30T13:30:00Z"}}}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := NewFileAutomationStatusReader(path).Status("alice", "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	books, ok := status["books"].([]any)
	if !ok || len(books) != 1 {
		t.Fatalf("books=%#v", status["books"])
	}
	book, ok := books[0].(map[string]any)
	if !ok || book["retryAt"] != nil {
		t.Fatalf("terminal failure exposed stale retryAt: %#v", books[0])
	}
}

func TestReconcileTerminalAutomationJobClosesFalseRunningState(t *testing.T) {
	now := time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC)
	job := map[string]any{
		"state": "running",
		"books": map[string]any{
			"book-1": map[string]any{"status": "failed", "stage": "comic_images", "needsManual": true},
		},
		"counts": map[string]any{"total": 1, "running": 0, "failed": 1},
	}
	if !reconcileTerminalAutomationJob(job, now) {
		t.Fatal("expected impossible running state to be reconciled")
	}
	if got := automationStringValue(job["state"], ""); got != "failed" {
		t.Fatalf("state = %q, want failed", got)
	}
	if _, ok := job["completedAt"]; !ok {
		t.Fatal("terminal job did not receive completedAt")
	}
}

func TestReconcileTerminalAutomationJobKeepsDueRetryRunning(t *testing.T) {
	now := time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC)
	job := map[string]any{
		"state": "running",
		"books": map[string]any{
			"book-1": map[string]any{"status": "failed", "retryRequested": true, "retryAt": now.Add(time.Minute)},
		},
	}
	if reconcileTerminalAutomationJob(job, now) {
		t.Fatal("job with an automatic continuation was closed")
	}
	if got := automationStringValue(job["state"], ""); got != "running" {
		t.Fatalf("state = %q, want running", got)
	}
}

func TestRepairStaleRunningBooksMakesInterruptedBookRetryable(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	job := map[string]any{
		"state": "paused",
		"books": map[string]any{
			"book-stale":   map[string]any{"status": "running", "stage": "comic_images", "updatedAt": now.Add(-2 * time.Hour).Format(time.RFC3339Nano)},
			"book-ready":   map[string]any{"status": "completed", "stage": "published"},
			"book-running": map[string]any{"status": "running", "stage": "comic_images", "updatedAt": now.Add(-time.Minute).Format(time.RFC3339Nano)},
		},
		"counts": map[string]any{"total": 3, "running": 2, "failed": 0, "ready": 1},
	}
	if !repairStaleRunningAutomationJob(job, now) {
		t.Fatal("expected stale running book to be repaired")
	}
	books := job["books"].(map[string]any)
	stale := books["book-stale"].(map[string]any)
	if got := automationStringValue(stale["status"], ""); got != "failed" {
		t.Fatalf("stale book status = %q, want failed", got)
	}
	if !automationBookRetryable(stale) {
		t.Fatal("repaired book must be selectable by retry")
	}
	// 刚有过心跳的在途书不能误判失败。
	if got := automationStringValue(books["book-running"].(map[string]any)["status"], ""); got != "running" {
		t.Fatalf("in-flight book status = %q, want running", got)
	}
	counts := job["counts"].(map[string]any)
	if got := automationNumberValue(counts["failed"], 0); got != 1 {
		t.Fatalf("counts.failed = %v, want 1", got)
	}
	// 用户主动暂停的状态必须保留，不能被计数反推成 failed。
	if got := automationStringValue(job["state"], ""); got != "paused" {
		t.Fatalf("job state = %q, want paused", got)
	}
}

func TestRepairInterruptedRunningJobBecomesRetryableAfterRestart(t *testing.T) {
	now := time.Now().UTC()
	job := map[string]any{
		"state":     "running",
		"updatedAt": now.Add(-time.Hour).Format(time.RFC3339Nano),
		"books": map[string]any{
			"book-1": map[string]any{"status": "running", "stage": "comic_images", "updatedAt": now.Add(-time.Hour).Format(time.RFC3339Nano)},
			"book-2": map[string]any{"status": "queued", "stage": "preflight"},
			"book-3": map[string]any{"status": "completed", "stage": "published"},
		},
	}
	if !repairStaleRunningAutomationBooks(&automationStateFile{Jobs: map[string]map[string]any{"job-1": job}}, now) {
		t.Fatal("job running from a previous process must be reclaimed")
	}
	if got := automationStringValue(job["state"], ""); got != "failed" {
		t.Fatalf("job state = %q, want failed", got)
	}
	books := job["books"].(map[string]any)
	for _, id := range []string{"book-1", "book-2"} {
		entry := books[id].(map[string]any)
		if got := automationStringValue(entry["status"], ""); got != "failed" {
			t.Fatalf("%s status = %q, want failed", id, got)
		}
		if !automationBookRetryable(entry) {
			t.Fatalf("%s must be retryable after restart", id)
		}
	}
	if got := automationStringValue(books["book-3"].(map[string]any)["status"], ""); got != "completed" {
		t.Fatalf("completed book must stay completed, got %q", got)
	}
}

func TestRepairStaleRunningBooksSkipsActiveRun(t *testing.T) {
	now := time.Now().UTC()
	job := map[string]any{
		"state":     "running",
		"updatedAt": now.Format(time.RFC3339Nano),
		"books": map[string]any{
			"book-1": map[string]any{"status": "running", "stage": "comic_images", "updatedAt": now.Format(time.RFC3339Nano)},
		},
	}
	if repairStaleRunningAutomationJob(job, now) {
		t.Fatal("active run must be left to the stall watchdog")
	}
	if repairInterruptedRunningAutomationJob(job, now) {
		t.Fatal("job heartbeating in this process must not be reclaimed")
	}
}

func TestFileAutomationStatusReaderRepairsN8AlreadyUploadedImages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"jobs":{"job":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"failed","books":{"book-1":{"bookId":"book-1","status":"failed","stage":"comic_publish","error":"N8接口返回错误：IMAGE_ALREADY_IN_CAROUSEL","retryRequested":true,"retryAt":"2026-10-30T13:30:00Z"}}}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := NewFileAutomationStatusReader(path).Status("alice", "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	books, ok := status["books"].([]any)
	if !ok || len(books) != 1 {
		t.Fatalf("books=%#v", status["books"])
	}
	book, ok := books[0].(map[string]any)
	if !ok || book["status"] != "completed" || book["stage"] != "uploaded" || book["error"] != "" || book["retryAt"] != nil {
		t.Fatalf("old N8 upload failure was not repaired: %#v", book)
	}
	if status["state"] != "completed" {
		t.Fatalf("state=%v, want completed", status["state"])
	}
	persisted, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(persisted), `"status":"failed"`) {
		t.Fatalf("repaired state was not persisted: %s (err=%v)", persisted, err)
	}
}

func TestFileAutomationStatusReaderClearsStaleRetryRequestAfterBookCompletes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "automation.json")
	raw := `{"jobs":{"job":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"completed","books":{"book-1":{"bookId":"book-1","status":"completed","stage":"published","retryRequested":true,"retryAt":"2026-10-30T13:30:00Z","retryReason":"content_policy"}}}}}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err := NewFileAutomationStatusReader(path).Status("alice", "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	books, ok := status["books"].([]any)
	if !ok || len(books) != 1 {
		t.Fatalf("books=%#v", status["books"])
	}
	book, ok := books[0].(map[string]any)
	if !ok || automationBoolValue(book["retryRequested"]) || book["retryAt"] != nil {
		t.Fatalf("completed book still appears retryable: %#v", books[0])
	}
	persisted, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(persisted), `"retryRequested":true`) || strings.Contains(string(persisted), `"retryReason"`) {
		t.Fatalf("completed book retry marker was not removed: %s", persisted)
	}
}

func TestV12AutomationStatusRouteUsesPersistedState(t *testing.T) {
	now := time.Unix(1700000000, 0)
	path := filepath.Join(t.TempDir(), "automation.json")
	if err := os.WriteFile(path, []byte(`{"jobs":{"job-key":{"owner":"alice","batchId":"batch-1","id":"job-1","state":"completed","books":{}}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AutomationStatusReader: NewFileAutomationStatusReader(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/batch-1/automation", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"job-1"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12RuntimeSummaryUsesGoReadModels(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "运行总览", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "automation.json")
	if err := os.WriteFile(path, []byte(`{"jobs":{"job-key":{"owner":"alice","batchId":"`+batch.ID+`","id":"job-1","state":"running","books":{}}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, AutomationStatusReader: NewFileAutomationStatusReader(path)})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/runtime-summary", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"automation"`) || !strings.Contains(rec.Body.String(), `"stageSummaries"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

type runtimeSummaryTextProvider struct{}

func (runtimeSummaryTextProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	return "{}", nil
}

func TestV12RuntimeSummaryUsesConfiguredTextProvider(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "文本模型状态", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret",
		Now:          func() time.Time { return now },
		Slice:        2,
		Store:        store,
		Director:     &batchfactoryv11.DirectorService{Store: store, Provider: runtimeSummaryTextProvider{}},
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/runtime-summary", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"stage":"assets","available":true`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12RuntimeSummaryUsesConfiguredImageModel(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "图片模型状态", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "image-demo", "kind": "image", "enabled": true, "modelId": "image-model", "baseUrl": "https://images.example/v1", "credential": "image-secret",
	}}}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, NovelFetchStore: catalog})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/runtime-summary", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"stage":"image","available":true`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func mustJSON(value any) []byte {
	encoded, _ := json.Marshal(value)
	return encoded
}

func TestPublicAutomationJobDerivesCountsFromBooks(t *testing.T) {
	job := map[string]any{
		"id": "job-1", "batchId": "batch-1", "state": "running",
		"counts": map[string]any{"total": 3, "ready": 0, "running": 3, "failed": 0, "blocked": 0},
		"books": map[string]any{
			"a": map[string]any{"status": "completed"},
			"b": map[string]any{"status": "failed"},
			"c": map[string]any{"status": "queued"},
		},
	}
	public := publicAutomationJob(job)
	counts, _ := public["counts"].(map[string]any)
	if counts["ready"] != 1 || counts["failed"] != 1 || counts["queued"] != 1 || counts["running"] != 0 {
		t.Fatalf("derived counts = %#v, want ready=1 failed=1 queued=1 running=0", counts)
	}
}

func TestPublicAutomationJobHidesHistoricalFailureAfterPublishCompleted(t *testing.T) {
	job := map[string]any{
		"id": "job-1", "batchId": "batch-1", "state": "failed",
		"books": map[string]any{
			"book-1": map[string]any{
				"bookId": "book-1", "status": "completed", "stage": "published",
				"message": "漫画合集已发布", "error": "已生成 18/28 页，10 页失败",
				"retryRequested": true, "retryAt": time.Now(), "batchRetryCount": 2,
			},
		},
	}

	public := publicAutomationJob(job)
	books := public["books"].([]any)
	book := books[0].(map[string]any)
	if book["error"] != "" || book["publishError"] != "" || book["retryRequested"] != nil || book["retryAt"] != nil || book["batchRetryCount"] != nil {
		t.Fatalf("completed published book leaked stale failure state: %#v", book)
	}
}

func TestAutomationFailureCategoryClassifiesKnownReasons(t *testing.T) {
	cases := []struct {
		name    string
		status  string
		message string
		want    string
	}{
		{"completed has no category", "completed", "已生成 29/30 页", ""},
		{"billed page", "failed", "已生成 29/30 页，1 页需要人工处理；需人工处理（IMAGE_BILLED_NOT_RETRIEVED）", "billed_not_retrieved"},
		{"upstream timeout", "failed", "图片通道响应读取中断（HTTP 200）：context deadline exceeded", "upstream_timeout"},
		{"slot stalled", "failed", "出图超过 10m0s 没有任何进展，已释放出图槽位，本书可直接重试", "upstream_timeout"},
		{"missing work", "failed", "漫画合集里没有同名作品《半生离恨》，请先创建作品", "external_work_missing"},
		{"missing config", "failed", "图片模型或图片生成服务未配置", "config_missing"},
		{"audio config", "failed", "当前组织的“现代悬疑”分类没有可用音频", "config_missing"},
	}
	for _, tc := range cases {
		got, _ := automationFailureCategory(tc.status, tc.message, "", "")
		if got != tc.want {
			t.Fatalf("%s: category=%q want %q", tc.name, got, tc.want)
		}
	}
}
