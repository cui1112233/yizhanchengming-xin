package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelcreation"
	"testing"
	"time"
)

func TestNovelSchedulerAdapterStrictInputsReplayAndScope(t *testing.T) {
	ctx := context.Background()
	st := novelcreation.NewMemoryStore()
	s := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
	p, e := s.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "调度测试", Genre: "悬疑", ModelID: "own-model", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: p.ID, Mode: "long_original", Title: "已有会话"})
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
	path := "/api/novel-creation/projects/" + p.ID + "/scheduler/start"
	in := map[string]any{"idempotencyKey": "scheduler-http", "sessionId": session.ID, "modelId": "own-model", "intervalSeconds": 1800, "dailyChapterCap": 10, "maxConsecutiveFailures": 3}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, in)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var first struct {
		Schedule novelcreation.Schedule `json:"schedule"`
	}
	json.Unmarshal(response.Body.Bytes(), &first)
	response = signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, in)
	var again struct {
		Schedule novelcreation.Schedule `json:"schedule"`
	}
	json.Unmarshal(response.Body.Bytes(), &again)
	if response.Code != 201 || again.Schedule.ID != first.Schedule.ID {
		t.Fatal("duplicate scheduler", response.Body.String())
	}
	if response = signedJSONRequest(t, mux, now, "bob", http.MethodPost, path, in); response.Code != 404 {
		t.Fatal("scope", response.Code)
	}
	in["prompt"] = "frontend must not supply this"
	if response = signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, in); response.Code != 400 {
		t.Fatal("unknown prompt accepted", response.Code)
	}
	runs, _ := st.ListRuns(ctx, "alice", p.ID)
	if len(runs) != 0 {
		t.Fatal("start generated immediately")
	}
}
