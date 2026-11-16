package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelcreation"
	"testing"
	"time"
)

func TestNovelCreationAutomationControlsCapsAndLogsHTTP(t *testing.T) {
	ctx := context.Background()
	store := novelcreation.NewMemoryStore()
	s := novelcreation.NewService(store, novelcreation.NewMemoryRuntime(), nil)
	p, e := s.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "作品", Genre: "都市", ModelID: "m"})
	if e != nil {
		t.Fatal(e)
	}
	session, _ := s.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
	base := "/api/novel-creation/projects/" + p.ID
	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/schedules", map[string]any{"sessionId": session.ID, "prompt": "写", "intervalMinutes": 1, "dailyChapterCap": 7, "maxConsecutiveFailures": 2})
	if r.Code != 201 {
		t.Fatal(r.Code, r.Body.String())
	}
	var out struct {
		Schedule novelcreation.Schedule `json:"schedule"`
	}
	json.Unmarshal(r.Body.Bytes(), &out)
	if out.Schedule.DailyChapterCap != 7 || out.Schedule.MaxConsecutiveFailures != 2 {
		t.Fatal("caps lost")
	}
	path := base + "/schedules/" + out.Schedule.ID
	for _, action := range []string{"pause", "resume", "stop"} {
		r = signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/"+action, nil)
		if r.Code != 200 {
			t.Fatal(action, r.Code, r.Body.String())
		}
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodGet, path+"/logs", nil)
	if r.Code != 200 {
		t.Fatal(r.Code)
	}
	r = signedJSONRequest(t, mux, now, "bob", http.MethodGet, path+"/logs", nil)
	if r.Code == 200 {
		t.Fatal("foreign logs exposed")
	}
	r = signedJSONRequest(t, mux, now, "bob", http.MethodPost, path+"/resume", nil)
	if r.Code == 200 {
		t.Fatal("foreign control allowed")
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/batches", map[string]any{"sessionId": session.ID, "totalChapters": 5, "maxConsecutiveFailures": 4})
	if r.Code != 201 {
		t.Fatal(r.Code, r.Body.String())
	}
	var b struct {
		Batch novelcreation.Batch `json:"batch"`
	}
	json.Unmarshal(r.Body.Bytes(), &b)
	if b.Batch.MaxConsecutiveFailures != 4 {
		t.Fatal("batch cap lost")
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodGet, base+"/batches/"+b.Batch.ID+"/logs", nil)
	if r.Code != 200 {
		t.Fatal(r.Code)
	}
	r = signedJSONRequest(t, mux, now, "bob", http.MethodGet, base+"/batches/"+b.Batch.ID+"/logs", nil)
	if r.Code == 200 {
		t.Fatal("foreign batch logs exposed")
	}
}
