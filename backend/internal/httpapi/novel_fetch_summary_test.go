package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
	"strings"
	"testing"
	"time"
)

type summaryGuardStore struct {
	novelfetchworkshop.LifecycleStore
	bodyReads, writes int
}

func (s *summaryGuardStore) GetBody(ctx context.Context, owner, id, version string) (novelfetchworkshop.BodyRecord, error) {
	s.bodyReads++
	return s.LifecycleStore.GetBody(ctx, owner, id, version)
}
func (s *summaryGuardStore) PutDocument(ctx context.Context, owner string, d novelfetchworkshop.Document) error {
	s.writes++
	return s.LifecycleStore.PutDocument(ctx, owner, d)
}

func TestNovelFetchSummaryDoesNotLoadBodiesOrWriteAndKeepsOwnerScope(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "summary-test"
	base := novelfetchworkshop.NewMemoryStore()
	base.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "12345", Original: strings.Repeat("正文秘密", 1000), Versions: map[string]any{"ai1": "改文秘密"}, Meta: map[string]any{"bookName": "摘要书"}})
	base.PutDocument(context.Background(), "other", novelfetchworkshop.Document{BookID: "67890", Original: "其他用户秘密"})
	store := &summaryGuardStore{LifecycleStore: base}
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: store})
	for _, path := range []string{"/api/novel-fetch-workshop/tasks/12345/summary", "/api/tasks/12345/summary"} {
		rr := doNovelFetchRequest(t, handler, http.MethodGet, path, nil, now, secret)
		if rr.Code != 200 {
			t.Fatalf("%s: %d %s", path, rr.Code, rr.Body.String())
		}
		var data map[string]any
		json.Unmarshal(rr.Body.Bytes(), &data)
		if data["content_loaded"] != false || strings.Contains(rr.Body.String(), "正文秘密") || strings.Contains(rr.Body.String(), "改文秘密") {
			t.Fatalf("body leaked: %.200s", rr.Body.String())
		}
	}
	rr := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/67890/summary", nil, now, secret)
	if rr.Code != 404 {
		t.Fatalf("foreign task status %d", rr.Code)
	}
	if store.bodyReads != 0 || store.writes != 0 {
		t.Fatalf("reads=%d writes=%d", store.bodyReads, store.writes)
	}
}
