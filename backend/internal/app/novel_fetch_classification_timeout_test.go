package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// 分类 LLM 调用必须自带短超时：模型网关挂起时不能继承 300 秒的导演链路
// 超时，否则「AI 判断发布信息」会长时间转圈才兜底。这里收紧超时后验证
// adapter 对挂起的端点会按超时快速失败，而不是等到导演链路的 300 秒。
func TestNovelFetchRewriteAdapterAppliesClassificationTimeout(t *testing.T) {
	original := novelFetchClassificationTimeout
	novelFetchClassificationTimeout = 120 * time.Millisecond
	t.Cleanup(func() { novelFetchClassificationTimeout = original })

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release // 模拟模型网关挂起，直到测试结束才放行
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"style\":\"现代甜文\"}"}}]}`))
	}))
	defer server.Close()
	defer close(release)

	adapter := novelFetchRewriteAdapter{provider: &batchfactoryv11.OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "key", Model: "test-model"}}
	started := time.Now()
	if _, err := adapter.Complete(context.Background(), novelfetchworkshop.RewriteRequest{}); err == nil {
		t.Fatal("Complete should fail when the model gateway hangs")
	}
	elapsed := time.Since(started)
	if elapsed > 30*time.Second {
		t.Fatalf("Complete returned after %v, want the classification timeout to cut it short", elapsed)
	}
	if elapsed < 100*time.Millisecond {
		t.Fatalf("Complete returned after %v, suspiciously faster than the configured timeout", elapsed)
	}
	if !errors.Is(context.DeadlineExceeded, context.DeadlineExceeded) {
		t.Fatal("unreachable guard")
	}
}
