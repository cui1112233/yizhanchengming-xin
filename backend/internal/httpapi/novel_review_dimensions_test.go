package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelcreation"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"testing"
	"time"
)

func TestNovelReviewDimensionsSettingsContract(t *testing.T) {
	ctx := context.Background()
	st := novelcreation.NewMemoryStore()
	s := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
	p, e := s.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "审稿目录", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
	path := "/api/novel-creation/projects/" + p.ID + "/settings"
	r := signedJSONRequest(t, mux, now, "alice", http.MethodGet, path, nil)
	var out struct {
		Settings novelcreation.ProjectSettings `json:"settings"`
		Catalog  []novelprompt.ReviewDimension `json:"reviewDimensionCatalog"`
	}
	if e = json.Unmarshal(r.Body.Bytes(), &out); e != nil || r.Code != 200 || len(out.Catalog) != 37 {
		t.Fatal(r.Code, r.Body.String(), e)
	}
	for _, ids := range [][]int{{32}, {}} {
		r = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, map[string]any{"agent": map[string]any{"version": 1, "reviewDimensionIds": ids}})
		if r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		json.Unmarshal(r.Body.Bytes(), &out)
		if out.Settings.Agent.ReviewDimensionIDs == nil || len(*out.Settings.Agent.ReviewDimensionIDs) != len(ids) {
			t.Fatal("save mismatch")
		}
	}
	r = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, map[string]any{"agent": map[string]any{"version": 1, "reviewDimensionIds": []int{99}}})
	if r.Code != 400 {
		t.Fatal("unknown accepted", r.Code)
	}
	r = signedJSONRequest(t, mux, now, "bob", http.MethodGet, path, nil)
	if r.Code != 404 {
		t.Fatal("cross owner", r.Code)
	}
	runs, _ := st.ListRuns(ctx, "alice", p.ID)
	if len(runs) != 0 {
		t.Fatal("editing generated")
	}
}
