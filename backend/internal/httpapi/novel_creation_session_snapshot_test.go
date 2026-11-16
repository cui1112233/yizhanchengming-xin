package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelcreation"
	"testing"
	"time"
)

func TestNovelCreationSessionSnapshotRestoresEmptyAndLatestRun(t *testing.T) {
	ctx := context.Background()
	service := novelcreation.NewService(novelcreation.NewMemoryStore(), novelcreation.NewMemoryRuntime(), nil)
	p, e := service.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "测试", Genre: "悬疑", ModelID: "text"})
	if e != nil {
		t.Fatal(e)
	}
	session, e := service.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: p.ID, Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: service})
	path := "/api/novel-creation/sessions/" + session.ID + "/snapshot"
	response := signedJSONRequest(t, mux, now, "alice", http.MethodGet, path, nil)
	var snapshot novelcreation.Snapshot
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &snapshot) != nil || snapshot.Session.ID != session.ID || snapshot.Run.ID != "" {
		t.Fatalf("empty status=%d body=%s", response.Code, response.Body)
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, path, nil)
	if foreign.Code != 404 {
		t.Fatalf("foreign status=%d body=%s", foreign.Code, foreign.Body)
	}
	run, e := service.SendMessage(ctx, "alice", novelcreation.SendMessageInput{SessionID: session.ID, Content: "第一章", ModelID: "text", IdempotencyKey: "first"})
	if e != nil {
		t.Fatal(e)
	}
	response = signedJSONRequest(t, mux, now, "alice", http.MethodGet, path, nil)
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &snapshot) != nil || snapshot.Run.ID != run.ID || len(snapshot.Messages) != 1 || len(snapshot.Events) != 1 {
		t.Fatalf("restored status=%d body=%s", response.Code, response.Body)
	}
}
