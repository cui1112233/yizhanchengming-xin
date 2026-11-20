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

func TestNovelShortArtifactRouteValidationOwnershipAndReadback(t *testing.T) {
	ctx := context.Background()
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	out, err := svc.StartShortStory(ctx, "alice", novelcreation.ShortStoryInput{IdempotencyKey: "artifact-http", Mode: "short_original", ModelID: "fake-only", Requirements: "短篇测试", TargetWordCount: 1000})
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/novel-creation/projects/" + out.Project.ID + "/short-story/artifact"
	input := map[string]any{"version": out.Story.Version, "key": "recommendation.txt", "content": "修改后的推荐文案"}
	r := signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, input)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("active task must reject editing: %d %s", r.Code, r.Body.String())
	}
	r = signedJSONRequest(t, mux, now, "bob", http.MethodPut, path, input)
	if r.Code != http.StatusNotFound {
		t.Fatalf("cross-account edit: %d %s", r.Code, r.Body.String())
	}
	input["systemPrompt"] = "frontend cannot set prompts"
	r = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, input)
	if r.Code != http.StatusBadRequest {
		t.Fatalf("unknown field: %d", r.Code)
	}
	delete(input, "systemPrompt")
	// Seed only an isolated in-memory finished story: no real generation is used.
	out.Story.Status = "completed"
	out.Story.Recommendation = "原推荐"
	out.Story.Artifacts = []novelcreation.ShortArtifact{{Key: "recommendation.txt", Title: "推荐", Content: "原推荐"}}
	raw, _ := json.Marshal(out.Story)
	if err = st.PutContentRecord(ctx, novelcreation.ContentRecord{Owner: "alice", ProjectID: out.Project.ID, Kind: "short_story", ID: "story", Status: "completed", Data: raw}); err != nil {
		t.Fatal(err)
	}
	if _, err = st.UpdateRun(ctx, "alice", novelcreation.RunUpdate{ID: out.Run.ID, Status: novelcreation.RunSucceeded}); err != nil {
		t.Fatal(err)
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, input)
	if r.Code != http.StatusOK {
		t.Fatalf("save: %d %s", r.Code, r.Body.String())
	}
	var saved struct {
		Story novelcreation.ShortStory `json:"story"`
	}
	if err = json.Unmarshal(r.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Story.Recommendation != input["content"] || saved.Story.Version <= out.Story.Version {
		t.Fatal("response did not reflect saved version", r.Body.String())
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, input)
	if r.Code != http.StatusConflict {
		t.Fatalf("stale version: %d %s", r.Code, r.Body.String())
	}
	input["version"] = saved.Story.Version
	for _, key := range []string{"../../config.json", "reviews/editorial.md", "missing.md"} {
		input["key"] = key
		r = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, input)
		if r.Code == http.StatusOK {
			t.Fatalf("unallowed key accepted: %s", key)
		}
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodGet, strings.TrimSuffix(path, "/artifact"), nil)
	if r.Code != http.StatusOK || !strings.Contains(r.Body.String(), "修改后的推荐文案") {
		t.Fatal("saved content not readable", r.Code, r.Body.String())
	}
}
