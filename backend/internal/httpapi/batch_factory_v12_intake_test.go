package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestV12LegacyImageGenerationUsesManagedImageModel(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "secret"
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer image-secret" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"url":"https://images.example/generated.png"}]}`))
	}))
	defer provider.Close()
	store := batchfactoryv11.NewMemoryStore()
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "tester", map[string]any{"modelCatalog": []any{map[string]any{
		"id": "image-demo", "kind": "image", "enabled": true, "modelId": "image-model", "baseUrl": provider.URL, "credential": "image-secret",
	}}}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, Slice: 1, Store: store, NovelFetchStore: catalog})
	rec := signedJSONRequest(t, api, now, "tester", http.MethodPost, "/api/batch-factory/v12/image-generation", map[string]any{"imageModelId": "image-demo", "prompt": "一间明亮的客厅"})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"imageUrl":"https://images.example/generated.png"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12NovelFetchIntakeUsesGoBatchStore(t *testing.T) {
	const secret = "secret"
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, Slice: 1, Store: store})
	input := map[string]any{
		"books":    []map[string]any{{"id": "book-1", "bookId": "book-1", "sourceTaskId": "task-1", "title": "测试小说", "sourceText": "正文"}},
		"metadata": map[string]any{"source": "novel-fetch"},
	}
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/intakes/novel-fetch", input)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var payload struct {
		Intake batchfactoryv11.Intake `json:"intake"`
	}
	if err := json.NewDecoder(bytes.NewReader(created.Body.Bytes())).Decode(&payload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if payload.Intake.ID == "" {
		t.Fatal("Go V12 intake response did not include an id")
	}
	if !strings.Contains(created.Body.String(), `"redirectTo":"/shuihuo-production?intake=`) {
		t.Fatalf("expected batch factory redirect, got %s", created.Body.String())
	}
	read := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/intakes/"+payload.Intake.ID, nil)
	if read.Code != http.StatusOK {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body.String())
	}
}

func TestV12BatchListUsesGoBatchStore(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	if _, err := store.CreateBatch(nil, "alice", batchfactoryv11.CreateBatchInput{Title: "列表测试", Books: []batchfactoryv11.CreateBookInput{{BookID: "book-1", Title: "一本书", SourceText: "不应出现在列表"}}}); err != nil {
		t.Fatalf("create test batch: %v", err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12BatchSummaryOmitsSourceText(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	if _, err := store.CreateBatch(nil, "alice", batchfactoryv11.CreateBatchInput{Title: "摘要测试", Books: []batchfactoryv11.CreateBookInput{{BookID: "book-1", Title: "一本书", SourceText: "不应出现在摘要"}}}); err != nil {
		t.Fatalf("create test batch: %v", err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/summary", nil)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "不应出现在摘要") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12NovelFetchIntakeRejectsEmptyBookList(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: batchfactoryv11.NewMemoryStore()})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/intakes/novel-fetch", map[string]any{"books": []any{}})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
