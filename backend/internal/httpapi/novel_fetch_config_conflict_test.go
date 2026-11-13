package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/novelfetchworkshop"
	"strings"
	"testing"
	"time"
)

func TestWorkshopStaleFormAndLibraryCannotOverwriteNewSettings(t *testing.T) {
	s := novelfetchworkshop.NewMemoryStore()
	ctx := context.Background()
	old, _ := s.GetConfig(ctx, "tester")
	revision := old[novelfetchworkshop.ConfigRevisionKey]
	old["label"] = "new value"
	if err := s.PutConfig(ctx, "tester", old); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1800000000, 0)
	h := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchStore: s})
	res := doNovelFetchRequest(t, h, "POST", "/api/novel-fetch-workshop/config", map[string]any{"configRevision": revision, "label": "stale"}, now, "secret")
	if res.Code != http.StatusConflict || !strings.Contains(res.Body.String(), "刷新配置") {
		t.Fatalf("stale form %d %s", res.Code, res.Body.String())
	}
	res = doNovelFetchRequest(t, h, "POST", "/api/novel-fetch-workshop/knowledge/opening_phrases", map[string]any{"configRevision": revision, "item": map[string]any{"text": "stale draft"}}, now, "secret")
	if res.Code != http.StatusConflict {
		t.Fatalf("stale library %d %s", res.Code, res.Body.String())
	}
	current, _ := s.GetConfig(ctx, "tester")
	if current["label"] != "new value" {
		t.Fatal("new settings overwritten")
	}
}
