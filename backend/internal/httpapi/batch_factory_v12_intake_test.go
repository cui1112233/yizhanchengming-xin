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

type v12SummaryIndexStore struct {
	*batchfactoryv11.MemoryStore
	summaryCalls int
}

func (s *v12SummaryIndexStore) ListBatchSummaryIndex(_ context.Context, _ string) ([]batchfactoryv11.Batch, error) {
	s.summaryCalls++
	return []batchfactoryv11.Batch{{
		ID:                     "batch-1",
		Title:                  "封面测试",
		ProjectCoverArtifactID: "merge_cover",
		ProjectCoverMediaURL:   "https://media.example/first-shot.mp4",
		Books:                  []batchfactoryv11.Book{{ID: "book-1", Title: "测试书"}},
	}}, nil
}

func TestV12LegacyImageGenerationUsesManagedImageModel(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "secret"
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer image-secret" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode provider payload: %v", err)
		}
		if _, exists := payload["quality"]; exists {
			t.Errorf("legacy image payload unexpectedly included quality: %#v", payload)
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
	if !strings.Contains(created.Body.String(), `"redirectTo":"/batch-factory?intake=`) {
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

type v12BulkShuihuoStore struct {
	*batchfactoryv11.MemoryStore
	hidden map[string]bool
	calls  int
}

func (s *v12BulkShuihuoStore) ListShuihuoProjectBatchIDs(_ context.Context, _ string) (map[string]bool, error) {
	s.calls++
	return s.hidden, nil
}

func TestV12BatchListUsesBulkShuihuoFlagLookup(t *testing.T) {
	now := time.Unix(1700000000, 0)
	memory := batchfactoryv11.NewMemoryStore()
	hidden, err := memory.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "水货项目"})
	if err != nil {
		t.Fatalf("create hidden batch: %v", err)
	}
	if _, err := memory.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "正常项目"}); err != nil {
		t.Fatalf("create visible batch: %v", err)
	}
	store := &v12BulkShuihuoStore{MemoryStore: memory, hidden: map[string]bool{hidden.ID: true}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches", nil)
	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, body)
	}
	if store.calls != 1 {
		t.Fatalf("bulk flag lookups=%d, want 1", store.calls)
	}
	if strings.Contains(body, "水货项目") || !strings.Contains(body, "正常项目") {
		t.Fatalf("unexpected batch list body=%s", body)
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

type v12SummaryFallbackStore struct {
	*batchfactoryv11.MemoryStore
}

func (s *v12SummaryFallbackStore) ListBatchSummaryIndex(_ context.Context, _ string) ([]batchfactoryv11.Batch, error) {
	return []batchfactoryv11.Batch{{
		ID:                   "batch-2",
		Title:                "漫画封面",
		ProjectCoverImageURL: "/api/batch-factory/v12/batches/batch-2/books/book-1/comic-pages/page-1/images/image-1/content",
		ProductionMode:       "comic",
		Books:                []batchfactoryv11.Book{{ID: "book-1", Title: "漫画书", Mode: "comic"}},
	}}, nil
}

func TestV12BatchSummaryFallsBackToComicImageCoverAndMode(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := &v12SummaryFallbackStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/summary", nil)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "comic-pages/page-1/images/image-1/content") || !strings.Contains(body, `"productionMode":"comic"`) {
		t.Fatalf("status=%d body=%s", rec.Code, body)
	}
}

func TestV12BatchSummaryUsesLightweightCoverIndex(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := &v12SummaryIndexStore{MemoryStore: batchfactoryv11.NewMemoryStore()}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/summary", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"kind":"image"`) || !strings.Contains(rec.Body.String(), `/api/batch-factory/v12/batches/batch-1/merge-cover/merge_cover`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if store.summaryCalls != 1 {
		t.Fatalf("summary index calls=%d, want 1", store.summaryCalls)
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

func TestV12NovelFetchIntakeCanAppendBooksToBatch(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "追加测试"})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/intakes/novel-fetch", map[string]any{
		"books": []map[string]any{{"id": "book-append", "bookId": "book-append", "title": "追加小说"}},
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create intake status=%d body=%s", created.Code, created.Body.String())
	}
	var payload struct {
		Intake batchfactoryv11.Intake `json:"intake"`
	}
	if err := json.NewDecoder(bytes.NewReader(created.Body.Bytes())).Decode(&payload); err != nil {
		t.Fatalf("decode intake: %v", err)
	}
	appended := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/intakes/"+payload.Intake.ID+"/books", map[string]any{})
	if appended.Code != http.StatusOK {
		t.Fatalf("append status=%d body=%s", appended.Code, appended.Body.String())
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || len(updated.Books) != 1 || updated.Books[0].BookID != "book-append" {
		t.Fatalf("batch books after append = %#v, err=%v", updated.Books, err)
	}
}
