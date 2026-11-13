package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplayProductionRoutesSharedReadAndExplicitGenerate(t *testing.T) {
	ctx := context.Background()
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice"})
	a := &screenplay.AuthoringService{Store: s, Repository: screenplay.NewMemoryAuthoringRepository()}
	state, _ := a.Load(ctx, "alice", "p")
	state.Members = []screenplay.MemberGrant{{Username: "bob", Read: true, Edit: true}}
	if _, e := a.Save(ctx, "alice", "p", state, state.Revision); e != nil {
		t.Fatal(e)
	}
	repo := screenplay.NewMemoryProductionRepository()
	r := &screenplay.ProductionRuntime{Store: s, Repository: repo, RedisDo: func(context.Context, ...string) (any, error) { return "PONG", nil }, Executor: func(string) screenplay.BatchExecutor { return screenplay.BatchExecutor{Store: s} }}
	now := time.Date(2026, 11, 14, 1, 0, 0, 0, time.UTC)
	mux := http.NewServeMux()
	RegisterScreenplayProductionRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, r, a)
	read := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/comic-screenplays/projects/p/production/status", nil)
	if read.Code != 200 {
		t.Fatalf("shared read: %d %s", read.Code, read.Body.String())
	}
	blocked := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/start", map[string]any{})
	if blocked.Code < 400 {
		t.Fatal("read member spent owner model")
	}
	scheduledDenied := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/schedule", map[string]any{"run_at": time.Now().Add(time.Hour).Format(time.RFC3339)})
	if scheduledDenied.Code < 400 {
		t.Fatal("shared reader scheduled paid production")
	}
	state, _ = a.Load(ctx, "alice", "p")
	state.Members[0].Generate = true
	if _, e := a.Save(ctx, "alice", "p", state, state.Revision); e != nil {
		t.Fatal(e)
	}
	invalidSchedule := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/schedule", map[string]any{"run_at": "2000-01-01T00:00:00Z"})
	if invalidSchedule.Code != http.StatusBadRequest {
		t.Fatalf("past schedule %d %s", invalidSchedule.Code, invalidSchedule.Body.String())
	}
	scheduled := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/schedule", map[string]any{"run_at": time.Now().Add(time.Hour).Format(time.RFC3339)})
	if scheduled.Code != 200 || !containsText(scheduled.Body.String(), `"status":"scheduled"`) {
		t.Fatalf("schedule %d %s", scheduled.Code, scheduled.Body.String())
	}
	canceled := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/cancel-schedule", map[string]any{})
	if canceled.Code != 200 {
		t.Fatalf("cancel %d %s", canceled.Code, canceled.Body.String())
	}

	started := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/start", map[string]any{})
	if started.Code != 200 {
		t.Fatalf("shared start: %d %s", started.Code, started.Body.String())
	}
	paused := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/comic-screenplays/projects/p/production/pause", map[string]any{})
	if paused.Code != 200 {
		t.Fatalf("pause: %d %s", paused.Code, paused.Body.String())
	}
}
