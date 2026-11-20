package httpapi

import (
	"net/http"
	"qiantie/backend/internal/novelcreation"
	"testing"
	"time"
)

func TestNovelAgentDefaultsWithoutProject(t *testing.T) {
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: novelcreation.NewService(novelcreation.NewMemoryStore(), novelcreation.NewMemoryRuntime(), nil)})
	r := signedJSONRequest(t, mux, now, "alice", http.MethodPut, "/api/novel-creation/agent-settings", map[string]any{"agent": map[string]any{"version": 1, "passScore": 92, "maxRevisionRounds": 0, "reviewDimensionIds": []int{}}})
	if r.Code != 200 {
		t.Fatalf("无作品保存失败: %d %s", r.Code, r.Body.String())
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/novel-creation/agent-settings", nil)
	if r.Code != 200 || !containsText(r.Body.String(), `"passScore":92`) || !containsText(r.Body.String(), `"reviewDimensionIds":[]`) {
		t.Fatal(r.Code, r.Body.String())
	}
	r = signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/novel-creation/agent-settings", nil)
	if r.Code != 200 || containsText(r.Body.String(), `"passScore":92`) {
		t.Fatal("账号隔离失效", r.Code, r.Body.String())
	}
}
