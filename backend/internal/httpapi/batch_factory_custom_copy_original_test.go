package httpapi

import (
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

type customCopyConcurrentStore struct {
	batchfactoryv11.Store
	changed bool
}

func (s *customCopyConcurrentStore) CaptureBookSource(ctx context.Context, owner, batchID, bookID string, input batchfactoryv11.CaptureBookSourceInput) (batchfactoryv11.Book, error) {
	if !s.changed {
		s.changed = true
		batch, _ := s.Store.GetBatch(ctx, owner, batchID)
		for _, book := range batch.Books {
			if book.ID == bookID {
				meta := copyAnyMap(book.SourceMetadata)
				meta["concurrentClassification"] = "kept"
				s.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: meta, ExpectedRevision: book.Revision})
			}
		}
	}
	return s.Store.CaptureBookSource(ctx, owner, batchID, bookID, input)
}
func TestCustomCopyOriginalRetriesConcurrentMetadataSave(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "copy", Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "打码版", TxtText: "原文坏词保持", SourceMetadata: map[string]any{"sourceMode": "novel_fetch", "customCopyRecordId": "280893", "processingStatus": "failed"}}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := reuseCustomCopyOriginal(ctx, &customCopyConcurrentStore{Store: store}, "alice", batch.ID, batch.Books[0])
	if err != nil || got.SourceText != "原文坏词保持" || got.SourceMetadata["sourceMode"] != "custom_copy" || got.SourceMetadata["processingStatus"] != "done" || got.SourceMetadata["concurrentClassification"] != "kept" {
		t.Fatalf("restore lost race: %+v %v", got, err)
	}
}

func TestCustomCopyRetryKeepsOriginalWithoutSensitiveProcessing(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	const original = "这是坏词原文\n第二行保持不变。"
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "文案获取", InitialBatchSettings: batchfactoryv11.SettingsPatch{"sensitiveAIEnabled": json.RawMessage(`true`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "书", SourceText: "旧打码文案", TxtText: original, SourceMetadata: map[string]any{"sourceMode": "custom_copy", "customCopyRecordId": "280893", "processingStatus": "failed", "sourceContentVersion": "sensitive", "sensitiveReviewRequired": true, "gender": "女频", "style": "现代悬疑"}}}})
	if err != nil {
		t.Fatal(err)
	}
	workshop := novelfetchworkshop.NewMemoryStore()
	workshop.PutConfig(ctx, "alice", map[string]any{"appConfig": map[string]any{"layout": map[string]any{"apply_to_original": true, "apply_sensitive": true}}, "knowledgeLibraries": map[string]any{"layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"enabled": true, "find": "坏词", "replace": "替换内容", "apply_to_original": true}}}}})
	coordinator := NewBatchAutomationCoordinator(store, nil, nil, nil, nil, workshop, nil, nil, nil)
	coordinator.SetSourceClassification(batchFetchClassification{RewriteProvider: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
		t.Error("copy must not call sensitive repair")
		return "", nil
	})})
	got, err := coordinator.ensureAutomationSource(ctx, "alice", batch.ID, batch.Books[0])
	if err != nil || got.SourceText != original || got.TxtText != original || got.SourceMetadata["sourceContentVersion"] != "original" || got.SourceMetadata["sensitiveReviewRequired"] == true {
		t.Fatalf("original changed: %+v %v", got, err)
	}
}

func TestCustomCopyForcedFetchReusesSelectedOriginal(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	const original = "不替换的坏词文案。"
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "copy", Books: []batchfactoryv11.CreateBookInput{{Title: "书", BookID: "2100242549395006289", Platform: "15", SourceText: original, TxtText: original, SourceMetadata: map[string]any{"sourceMode": "custom_copy", "customCopyRecordId": "280893"}}}})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"code":200,"data":"其他版本"}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL}})
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/fetch-original", map[string]any{"force": true})
	if response.Code != 200 || calls != 0 || !strings.Contains(response.Body.String(), original) {
		t.Fatalf("copy was fetched/processed again: status=%d calls=%d body=%s", response.Code, calls, response.Body.String())
	}
}
