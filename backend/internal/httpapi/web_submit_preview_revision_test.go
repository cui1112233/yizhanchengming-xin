package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestWebSubmitIgnoresChangedPreviewAndProceeds(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutConfig(ctx, "tester", map[string]any{})
	_ = store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: "book1"})
	put := func(text string) {
		_, err := store.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{BookID: "book1", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: text})
		if err != nil {
			t.Fatal(err)
		}
	}
	put("原来的内容")
	revision, err := webSubmitPreviewRevision(ctx, store, "tester", []string{"book1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	put("新的内容")
	now := time.Now()
	mux := http.NewServeMux()
	connector := &configuredSubmitConnector{}
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}}
	RegisterWebSubmitSubmitCompat(mux, middleware, t.TempDir(), store, nil, nil, connector)
	req := httptest.NewRequest("POST", "/api/batch-rewrite/web-submit/submit", strings.NewReader(`{"taskIds":["book1"],"previewRevision":"`+revision+`"}`))
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	// 参考实现的提交不依赖预览令牌；预览过期不能再返回 409，
	// 请求应继续进入提交执行，由真实前置条件（这里未启用网站提交）决定结果。
	if res.Code == http.StatusConflict || strings.Contains(res.Body.String(), "重新生成提交预览") || !strings.Contains(res.Body.String(), "尚未启用") {
		t.Fatalf("stale preview still blocked submit: %d %s calls=%d", res.Code, res.Body.String(), connector.submitCalls)
	}
}
