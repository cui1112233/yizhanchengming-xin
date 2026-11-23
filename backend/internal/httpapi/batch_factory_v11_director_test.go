package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
)

type directorHTTPProvider struct {
	output string
	calls  int
}

func (p *directorHTTPProvider) Complete(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	return p.output, nil
}

const directorHTTPJSON = `{"characters":[{"name":"林晚","prompt":"黑色长发"}],"scenes":[{"name":"客厅","prompt":"现代客厅"}],"props":[],"storyboard":[{"duration_sec":9,"characters":["林晚"],"props":[],"scene":"客厅","prefix_key":"modern_conflict","shots":[{"start_sec":0,"end_sec":9,"description":"林晚推门进入客厅"}],"video_desc":"林晚推门进入客厅"}],"source_coverage":{"source_complete":true,"has_remaining_source":false}}`

func TestSliceTwoDirectorRoutePersistsRevision(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。", TxtText: "林晚推门进入客厅。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
	provider := &directorHTTPProvider{output: directorHTTPJSON}
	service := &batchfactoryv11.DirectorService{Store: store, Provider: provider}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, Director: service})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/director"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "directorRevision") || provider.calls != 1 {
		t.Fatalf("body=%s calls=%d", rec.Body.String(), provider.calls)
	}
	loaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Books[0].DirectorRevision == nil || len(loaded.Books[0].Videos) != 1 {
		t.Fatalf("book=%+v", loaded.Books[0])
	}
}

func TestDirectorForRequestUsesGoConfiguredProviderAndIgnoresLegacyRequestCredentials(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	provider := &directorHTTPProvider{output: directorHTTPJSON}
	shared := &batchfactoryv11.DirectorService{Store: store, Provider: provider}
	scoped, err := directorForRequest(shared, &textProviderInput{
		Endpoint: "https://example.com/v1/chat/completions",
		APIKey:   "test-key",
		Model:    "test-model",
	})
	if err != nil {
		t.Fatalf("unexpected scoped provider error: %v", err)
	}
	if scoped != shared {
		t.Fatal("request must use the Go-configured director service")
	}
	if shared.Provider != provider {
		t.Fatal("request credentials must not replace the Go-configured provider")
	}
}

func TestSliceTwoWorkingFrontViralRouteReturnsReviewableCandidate(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "captured source"}}})
	provider := &directorHTTPProvider{output: "candidate text"}
	service := &batchfactoryv11.DirectorService{Store: store, Provider: provider}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, Director: service})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/working-front/viral"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"content": "editable front"})
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), "candidate text") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls=%d", provider.calls)
	}
	loaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || loaded.Books[0].SourceText != "captured source" {
		t.Fatalf("captured source mutated: %+v err=%v", loaded.Books[0], err)
	}
}

func TestSliceOneDoesNotRegisterDirectorMutation(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v11/batches/x/books/y/director", map[string]any{})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSliceTwoBookStageRouteRecordsDirectorAndImageAvailability(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。", TxtText: "林晚推门进入客厅。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}})
	provider := &directorHTTPProvider{output: directorHTTPJSON}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider}})
	base := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/stages"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, base+"/director", map[string]any{"mode": "missing", "requestId": "director-1"})
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"stage":"director"`) {
		t.Fatalf("director status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = signedJSONRequest(t, api, now, "alice", http.MethodPost, base+"/image", map[string]any{"mode": "missing", "requestId": "image-1"})
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "图片模型或图片生成服务未配置") {
		t.Fatalf("image status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = signedJSONRequest(t, api, now, "alice", http.MethodGet, base, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"lastFailed"`) || !strings.Contains(rec.Body.String(), `"stage":"image"`) {
		t.Fatalf("summary status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBatchFactorySensitiveStageStopsManualDirectorBeforeProvider(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "历史书", Books: []batchfactoryv11.CreateBookInput{{Title: "片段", SourceText: "只保留了截取稿", Videos: []batchfactoryv11.CreateVideoInput{{Label: "旧页"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	provider := &directorHTTPProvider{output: directorHTTPJSON}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024), NovelFetchStore: novelfetchworkshop.NewMemoryStore(), Store: store, Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider}})
	base := "/api/batch-factory/v12/batches/" + batch.ID
	bookPath := base + "/books/" + batch.Books[0].ID
	for _, path := range []string{bookPath + "/stages/director", bookPath + "/director", base + "/director", bookPath + "/comic-source-segments/segment", bookPath + "/director-prompt-candidate", bookPath + "/h3/director", bookPath + "/hook", bookPath + "/opening-variants", bookPath + "/comic-pages/" + batch.Books[0].Videos[0].ID + "/generate", bookPath + "/assets/images/generate"} {
		latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
		if err != nil {
			t.Fatal(err)
		}
		rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"mode": "missing", "requestId": "old-book", "expectedRevision": latest.Books[0].Revision})
		if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "缺少完整原文") || provider.calls != 0 {
			t.Fatalf("%s 手动分镜未在收费前停止: calls=%d status=%d body=%s", path, provider.calls, rec.Code, rec.Body.String())
		}
	}
}

func TestRegenerationChecksCurrentProductionEvidenceWithoutCallingAI(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	now := time.Unix(1700000000, 0)
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画", InitialBatchSettings: batchfactoryv11.SettingsPatch{"sensitiveMode": comicTestRaw(t, "ai"), "textModelId": comicTestRaw(t, "model")}, Books: []batchfactoryv11.CreateBookInput{{Title: "书", TxtText: "全文。", SourceText: "全文。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}, Videos: []batchfactoryv11.CreateVideoInput{{Label: "页"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		return sensitiveTestProvider(func(context.Context, batchfactoryv11.TextCompletionRequest) (string, error) {
			calls++
			return `{"changes":[]}`, nil
		}), nil
	}
	c := &BatchAutomationCoordinator{store: store, stages: &batchfactoryv11.BookStageService{TextProviderResolver: resolver}}
	book, err := c.runBatchSensitiveStage(ctx, "alice", batch.ID, batch.Books[0].ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.batchSensitiveStage(ctx, "alice", batch.ID, book.ID, false, true); err != nil {
		t.Fatal(err)
	}
	if _, err = store.SaveDraft(ctx, "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Scope: batch.ID, Kind: "working-front-content", Content: "更换了段落。"}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store, Director: &batchfactoryv11.DirectorService{Store: store}, TextProviderResolver: resolver, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024), NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	base := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID
	for _, suffix := range []string{"/assets/images/generate", "/comic-pages/" + book.Videos[0].ID + "/generate"} {
		rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, base+suffix, map[string]any{})
		if rec.Code != 409 || !strings.Contains(rec.Body.String(), "生产文本版本") || calls != 1 {
			t.Fatalf("%s status=%d body=%s calls=%d", suffix, rec.Code, rec.Body.String(), calls)
		}
	}
}
