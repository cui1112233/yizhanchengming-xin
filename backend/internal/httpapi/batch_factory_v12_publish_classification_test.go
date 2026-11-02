package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestV12ClassifyBookPublishMetadataPersistsGenderAndStyle(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "book", Platform: "15", SourceText: "现代都市里，女主发现丈夫背叛后冷静反击。",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, Slice: 1,
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
	if metadata["gender"] != "女频" || metadata["style"] != "现代甜文" || metadata["classifyStatus"] != "classified" {
		t.Fatalf("metadata=%#v, want persisted publish classification", metadata)
	}
}

func TestV12ClassifyBookPublishMetadataPersistsThemesLabelsAndTags(t *testing.T) {
	metadata := classifyFetchedBatchBook(context.Background(), "alice", novelfetchworkshop.BookInfo{
		Themes: "言情、婚恋",
		Labels: "现代",
	}, "正文", nil, batchFetchClassification{})

	if metadata["themes"] != "言情、婚恋" || metadata["labels"] != "现代" || metadata["tags"] != "言情、婚恋、现代" {
		t.Fatalf("metadata=%#v, want upstream facets copied to display metadata", metadata)
	}
}
