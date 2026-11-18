package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelcreation"
	"strings"
	"testing"
	"time"
)

func TestNovelShortStoryRoutesOwnershipStrictInputsAndReplay(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	path := "/api/novel-creation/short-stories"
	input := map[string]any{"idempotencyKey": "short-ui", "mode": "short_imitation", "modelId": "fake-only", "referenceText": strings.Repeat("甲", 100), "readerPreferences": "智斗权谋", "inspirationTags": []string{"悬疑", "重生"}}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, input)
	if response.Code != 202 {
		t.Fatal(response.Code, response.Body.String())
	}
	var out novelcreation.ShortStoryResult
	if e := json.Unmarshal(response.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	replay := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, input)
	if replay.Code != 202 || !strings.Contains(replay.Body.String(), out.Run.ID) {
		t.Fatal("lost reply duplicated", replay.Body.String())
	}
	for _, p := range []string{path + "/requests/short-ui", "/api/novel-creation/projects/" + out.Project.ID + "/short-story"} {
		r := signedJSONRequest(t, mux, now, "alice", http.MethodGet, p, nil)
		if r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		r = signedJSONRequest(t, mux, now, "bob", http.MethodGet, p, nil)
		if r.Code != 404 {
			t.Fatal("owner boundary", r.Code, r.Body.String())
		}
	}
	input["inventedPrompt"] = "frontend must not set system prompt"
	if r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, input); r.Code != 400 {
		t.Fatal("unknown field accepted")
	}
	if r := signedJSONRequest(t, mux, now, "alice", http.MethodGet, path+"/requests/not-created", nil); r.Code != 404 {
		t.Fatal("missing request must be recoverable404", r.Code)
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	if len(projects) != 1 {
		t.Fatal("duplicate hidden project")
	}
}
