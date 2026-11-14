package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayRecoveredBoardVisibleInWorkspaceListingAndProjectExport(t *testing.T) {
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	_, e := store.CreateProject(ctx, screenplay.Project{ID: "recovered-view", OwnerUsername: "alice"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = store.SaveStoryboard(ctx, screenplay.Storyboard{ProjectID: "recovered-view", OwnerUsername: "alice", EpisodeNo: 1, Engine: "prompt_137", Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 12, VideoDesc: "已保存新模板", ImagePrompt: "静帧图片"}}})
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	for _, suffix := range []string{"/workspace", "/storyboards", "/export"} {
		response := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects/recovered-view"+suffix, nil)
		if response.Code != 200 || !strings.Contains(response.Body.String(), "prompt_137") || !strings.Contains(response.Body.String(), "静帧图片") {
			t.Fatal(suffix, response.Code, response.Body.String())
		}
	}
}
