package httpapi

import (
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"testing"
	"time"
)

func TestAICutCreativeBindingUsesExistingCanvasAndOwner(t *testing.T) {
	now := time.Now().UTC()
	cuts := aicut.NewMemoryStore()
	projects := agentworkspace.NewMemoryStore()
	if err := cuts.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-one", Title: "剪辑工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutCreativeCanvasRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, cuts, projects, nil, nil, func() time.Time { return now })
	created := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/projects/cut-one/creative-canvas", nil)
	var payload struct {
		TaskID string `json:"taskId"`
	}
	_ = json.Unmarshal(created.Body.Bytes(), &payload)
	response := signedJSONRequest(t, mux, now, "alice", "GET", "/api/agent/tasks/"+payload.TaskID+"/cut-project", nil)
	if response.Code != 200 {
		t.Fatalf("binding=%d %s", response.Code, response.Body.String())
	}
	var result struct {
		Binding struct {
			ProjectID string `json:"projectId"`
			TaskID    string `json:"taskId"`
		} `json:"binding"`
	}
	_ = json.Unmarshal(response.Body.Bytes(), &result)
	if result.Binding.ProjectID != "cut-one" || result.Binding.TaskID != payload.TaskID {
		t.Fatalf("wrong binding: %s", response.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", "GET", "/api/agent/tasks/"+payload.TaskID+"/cut-project", nil)
	if foreign.Code != 404 {
		t.Fatalf("foreign=%d", foreign.Code)
	}
	if err := projects.CreateProject(t.Context(), "alice", agentworkspace.Project{ID: "independent", Title: "独立创作", Canvas: aicutCreativeSeed(), CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	unbound := signedJSONRequest(t, mux, now, "alice", "GET", "/api/agent/tasks/independent/cut-project", nil)
	var empty map[string]any
	_ = json.Unmarshal(unbound.Body.Bytes(), &empty)
	if unbound.Code != 200 || empty["binding"] != nil {
		t.Fatalf("unbound=%d %s", unbound.Code, unbound.Body.String())
	}
}
