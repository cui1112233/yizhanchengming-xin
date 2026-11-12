package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/novelcreation"
)

func TestNovelCreationRoutesExposeSixModesAndProjectRunChain(t *testing.T) {
	store := novelcreation.NewMemoryStore()
	runtime := novelcreation.NewMemoryRuntime()
	service := novelcreation.NewService(store, runtime, nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{
		BridgeSecret:  "secret",
		Now:           func() time.Time { return now },
		NovelCreation: service,
	})

	modes := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/novel-creation/modes", nil)
	if modes.Code != http.StatusOK {
		t.Fatalf("modes status=%d body=%s", modes.Code, modes.Body.String())
	}
	for _, want := range []string{"长篇原创", "短篇原创", "仿写长篇", "仿写短篇", "同人创作", "续写"} {
		if !containsText(modes.Body.String(), want) {
			t.Fatalf("modes body missing %q: %s", want, modes.Body.String())
		}
	}

	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/projects", map[string]any{
		"title": "雨夜归人", "genre": "都市", "modelId": "text-main", "targetChapterWords": 2200,
	})
	if created.Code != http.StatusCreated || !containsText(created.Body.String(), `"title":"雨夜归人"`) {
		t.Fatalf("create project status=%d body=%s", created.Code, created.Body.String())
	}
	var projectPayload struct {
		Project novelcreation.Project `json:"project"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &projectPayload); err != nil {
		t.Fatalf("decode project: %v", err)
	}

	session := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/projects/"+projectPayload.Project.ID+"/sessions", map[string]any{
		"mode": "long_original", "title": "第一卷",
	})
	if session.Code != http.StatusCreated {
		t.Fatalf("create session status=%d body=%s", session.Code, session.Body.String())
	}
	var sessionPayload struct {
		Session novelcreation.Session `json:"session"`
	}
	if err := json.Unmarshal(session.Body.Bytes(), &sessionPayload); err != nil {
		t.Fatalf("decode session: %v", err)
	}

	runResponse := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/sessions/"+sessionPayload.Session.ID+"/messages", map[string]any{
		"content": "写第一章，开场在暴雨夜。", "modelId": "text-main", "thinkingLevel": "standard", "idempotencyKey": "client-1",
	})
	if runResponse.Code != http.StatusAccepted {
		t.Fatalf("send status=%d body=%s", runResponse.Code, runResponse.Body.String())
	}
	var runPayload struct {
		Run novelcreation.Run `json:"run"`
	}
	if err := json.Unmarshal(runResponse.Body.Bytes(), &runPayload); err != nil {
		t.Fatalf("decode run: %v", err)
	}

	snapshot := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/novel-creation/runs/"+runPayload.Run.ID, nil)
	if snapshot.Code != http.StatusOK || !containsText(snapshot.Body.String(), "写第一章") || !containsText(snapshot.Body.String(), "已加入小说创作队列") {
		t.Fatalf("snapshot status=%d body=%s", snapshot.Code, snapshot.Body.String())
	}

	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/novel-creation/runs/"+runPayload.Run.ID, nil)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign status=%d body=%s", foreign.Code, foreign.Body.String())
	}

}
