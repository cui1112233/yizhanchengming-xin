package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// contextAwareMetadataStore 模拟真实数据库对请求取消的响应：客户端断开后
// r.Context() 已取消，写回必须失败——除非 handler 使用摘除取消信号的 ctx。
type contextAwareMetadataStore struct {
	batchfactoryv11.Store
}

func (s *contextAwareMetadataStore) UpdateBookMetadata(ctx context.Context, owner, batchID, bookID string, input batchfactoryv11.UpdateBookMetadataInput) (batchfactoryv11.Book, error) {
	if err := ctx.Err(); err != nil {
		return batchfactoryv11.Book{}, err
	}
	return s.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, input)
}

// 生产实况：用户点「AI 判断发布信息」后模型响应慢，用户等不及刷新页面，
// nginx 记录 499（客户端断开）。分类必须在服务端独立跑完并落库，否则用户
// 重开页面还是「待 AI 判断」，形成点了又失败的循环。
func TestV12ClassifyPublishMetadataPersistsAfterClientDisconnect(t *testing.T) {
	now := time.Unix(1700000000, 0)
	base := batchfactoryv11.NewMemoryStore()
	batch, err := base.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		Title: "主母的账本", Platform: "15", SourceText: "现代都市里，女主发现丈夫背叛后冷静反击。",
		SourceMetadata: map[string]any{"gender": "女频", "genderSource": "bookinfo"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	store := &contextAwareMetadataStore{Store: base}
	slowJudge := func(ctx context.Context, info novelfetchworkshop.BookInfo) (string, error) {
		time.Sleep(300 * time.Millisecond)
		return `{"gender":"女频","confidence":0.9,"reason":"正文情感关系线突出"}`, nil
	}
	slowRewrite := singleFetchRewriteProvider(func(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
		time.Sleep(300 * time.Millisecond)
		return `{"style":"现代甜文","confidence":0.9,"reason":"人物关系以轻松甜感为主"}`, nil
	})
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: store, Slice: 1,
		NovelFetchAIGenderJudge: slowJudge,
		NovelFetchRewrite:       slowRewrite,
	})
	server := httptest.NewServer(api)
	defer server.Close()

	// 客户端在 150ms 后主动断开——远早于慢速 AI 判断完成（约 600ms）。
	go func() {
		request, _ := http.NewRequest(http.MethodPost,
			server.URL+"/api/batch-factory/v12/batches/"+batch.ID+"/books/"+batch.Books[0].ID+"/classify-publish-metadata",
			bytes.NewReader([]byte(`{"force":true}`)))
		request.Header.Set("Content-Type", "application/json")
		SignBridgeRequest(request, "alice", false, now, "secret")
		ctx, cancel := context.WithCancel(context.Background())
		request = request.WithContext(ctx)
		go func() {
			time.Sleep(150 * time.Millisecond)
			cancel()
		}()
		_, _ = http.DefaultClient.Do(request)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		current, err := store.GetBatch(context.Background(), "alice", batch.ID)
		if err == nil && current.Books[0].SourceMetadata["classifyStatus"] == "classified" {
			metadata := current.Books[0].SourceMetadata
			if metadata["style"] != "现代甜文" || metadata["gender"] != "女频" {
				t.Fatalf("metadata=%#v, want classification persisted", metadata)
			}
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("classification was not persisted after client disconnect")
}
