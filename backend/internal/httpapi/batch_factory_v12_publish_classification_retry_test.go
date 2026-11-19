package httpapi

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// conflictOnceMetadataStore 模拟「AI 判断期间自动化推进书档」的并发场景：
// 第一次分类写回前先以并发链路更新一次书档（revision+1），随后返回
// ErrConflict；之后的写回放行给底层 store。
type conflictOnceMetadataStore struct {
	batchfactoryv11.Store
	conflictsLeft   int
	concurrentPatch func(ctx context.Context, store batchfactoryv11.Store)
}

func (s *conflictOnceMetadataStore) UpdateBookMetadata(ctx context.Context, owner, batchID, bookID string, input batchfactoryv11.UpdateBookMetadataInput) (batchfactoryv11.Book, error) {
	if s.conflictsLeft > 0 {
		s.conflictsLeft--
		if s.concurrentPatch != nil {
			s.concurrentPatch(ctx, s.Store)
		}
		return batchfactoryv11.Book{}, batchfactoryv11.ErrConflict
	}
	return s.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, input)
}

// 生产实况：书档已有书城自带的男女频（genderSource=bookinfo）但风格为空，
// classifyStatus 从未落盘；此时人工点击「AI 判断发布信息」必须能在自动化
// 并发推进 revision 的情况下成功，且不能覆盖并发链路写入的字段。
func TestV12ClassifyPublishMetadataRetriesAfterRevisionConflict(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "主母的账本", Platform: "15", SourceText: "现代都市里，女主发现丈夫背叛后冷静反击。",
		SourceMetadata: map[string]any{"gender": "女频", "genderSource": "bookinfo", "queueStatus": "working"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	wrapped := &conflictOnceMetadataStore{
		Store:         store,
		conflictsLeft: 1,
		concurrentPatch: func(ctx context.Context, store batchfactoryv11.Store) {
			current, err := store.GetBatch(ctx, "alice", batch.ID)
			if err != nil {
				t.Errorf("concurrent GetBatch: %v", err)
				return
			}
			book, ok := findBatchBookByID(current, batch.Books[0].ID)
			if !ok {
				t.Fatal("concurrent book missing")
			}
			patch := map[string]any{}
			for key, value := range book.SourceMetadata {
				patch[key] = value
			}
			patch["queueStatus"] = "fetching"
			if _, err := store.UpdateBookMetadata(ctx, "alice", batch.ID, book.ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: patch, ExpectedRevision: book.Revision}); err != nil {
				t.Errorf("concurrent metadata update: %v", err)
			}
		},
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: wrapped, Slice: 1,
		NovelFetchAIGenderJudge: func(context.Context, novelfetchworkshop.BookInfo) (string, error) { return `{"gender":"女频","confidence":0.9,"reason":"正文情感关系线突出"}`, nil },
		NovelFetchRewrite: singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
			return `{"style":"现代甜文","confidence":0.9,"reason":"人物关系以轻松甜感为主"}`, nil
		}),
	})

	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/classify-publish-metadata", map[string]any{"force": true})

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	metadata := updated.Books[0].SourceMetadata
	if metadata["classifyStatus"] != "classified" || metadata["style"] != "现代甜文" || metadata["gender"] != "女频" {
		t.Fatalf("metadata=%#v, want classification persisted", metadata)
	}
	if metadata["queueStatus"] != "fetching" {
		t.Fatalf("metadata=%#v, want concurrent queueStatus preserved by merge retry", metadata)
	}
}

func TestV12ClassifyPublishMetadataFailsAfterExhaustedRetries(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "book", Platform: "15", SourceText: "正文。",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	wrapped := &conflictOnceMetadataStore{Store: store, conflictsLeft: automationBookWriteAttempts + 1}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: wrapped, Slice: 1,
		NovelFetchAIGenderJudge: func(context.Context, novelfetchworkshop.BookInfo) (string, error) { return `{"gender":"女频"}`, nil },
		NovelFetchRewrite:       singleFetchRewriteProvider(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) { return `{"style":"现代甜文"}`, nil }),
	})

	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/classify-publish-metadata", map[string]any{"force": true})

	if rec.Code != http.StatusConflict && rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s, want conflict-class failure", rec.Code, rec.Body.String())
	}
	if !errors.Is(batchfactoryv11.ErrConflict, batchfactoryv11.ErrConflict) {
		t.Fatal("unreachable guard")
	}
}
