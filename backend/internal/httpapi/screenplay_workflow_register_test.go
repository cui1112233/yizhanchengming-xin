package httpapi

import (
	"net/http"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplayWorkflowRegistrationSharesAuthoringStateAndRefusesUnavailableProduction(t *testing.T) {
	now := time.Date(2026, 11, 14, 10, 0, 0, 0, time.UTC)
	mux := http.NewServeMux()
	store := screenplay.NewMemoryStore()
	registerScreenplayWorkflow(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, RouterOptions{Screenplay: store, Now: func() time.Time { return now }}, nil, nil)
	call := func(method, path string, body any) int {
		return signedJSONRequest(t, mux, now, "alice", method, path, body).Code
	}
	base := "/api/comic-screenplays/projects/integration"
	if code := call("POST", "/api/comic-screenplays/projects", map[string]any{"id": "integration", "title": "集成验收"}); code != 201 {
		t.Fatalf("create %d", code)
	}
	if code := call("PUT", base+"/chapters", map[string]any{"expected_revision": 0, "chapters": []map[string]any{{"id": "chapter-one", "title": "开场", "source_text": "一段故事"}}}); code != 200 {
		t.Fatalf("chapters %d", code)
	}
	response := signedJSONRequest(t, mux, now, "alice", "GET", base+"/display", nil)
	if response.Code != 200 || !containsText(response.Body.String(), "chapter-one") {
		t.Fatalf("shared state %d %s", response.Code, response.Body.String())
	}
	if code := call("POST", base+"/production/start", map[string]any{}); code < 400 {
		t.Fatalf("missing durable/redis backend must fail, got %d", code)
	}
	workspace := signedJSONRequest(t, mux, now, "alice", "GET", base+"/workspace", nil)
	if workspace.Code != 200 || !containsText(workspace.Body.String(), `"edit":true`) {
		t.Fatalf("workspace ACL %d %s", workspace.Code, workspace.Body.String())
	}
	outsider := signedJSONRequest(t, mux, now, "bob", "GET", base+"/display", nil)
	if outsider.Code < 400 {
		t.Fatalf("foreign state readable %d", outsider.Code)
	}
}
