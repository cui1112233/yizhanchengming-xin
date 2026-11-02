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
