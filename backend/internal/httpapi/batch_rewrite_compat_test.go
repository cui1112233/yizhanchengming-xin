package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestBatchRewriteWebSubmitHistoryReadsGoDocumentLogs(t *testing.T) {
	now := time.Date(2026, 10, 11, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-1",
		Meta:   map[string]any{"bookName": "测试书", "platformName": "番茄"},
		Logs: []any{map[string]any{
			"event": "site_submit", "status": "submitted", "version": "ai1", "group_id": "group-1",
			"time": now.Format(time.RFC3339), "remote_receipt": map[string]any{"verified": true, "upload_success_count": 2},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	storeConfig := map[string]any{"site_submit_groups": []any{map[string]any{"group_id": "group-1", "updated_at": now.Format(time.RFC3339)}}}
	if err := store.PutConfig(context.Background(), "tester", storeConfig); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	accountStore := &authTestStore{account: auth.Account{Username: "tester", Active: true}}
	RegisterBatchRewriteConfigCompat(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: accountStore}, store, filepath.Join(t.TempDir(), "users"))
	req := httptest.NewRequest(http.MethodGet, "/api/batch-rewrite/web-submit/history", nil)
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", res.Code, res.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	records, ok := payload["records"].([]any)
	if !ok || len(records) != 1 {
		t.Fatalf("records=%#v", payload["records"])
	}
	if records[0].(map[string]any)["result"] != "视频管理系统已确认接收 2 个文件" {
		t.Fatalf("record=%#v", records[0])
	}
	groups, ok := payload["groups"].([]any)
	if !ok || len(groups) != 1 {
		t.Fatalf("groups=%#v", payload["groups"])
	}
}
