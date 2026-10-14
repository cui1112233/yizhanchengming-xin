package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
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
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
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
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
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
