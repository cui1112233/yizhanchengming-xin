package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayStoryboardAttemptStatusOwnerScopeFiltersAndNoRawLeak(t *testing.T) {
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	_, _ = projects.CreateProject(ctx, screenplay.Project{ID: "attempt-p", OwnerUsername: "alice"})
	_, e := projects.MutateStoryboardGenerations(ctx, "alice", "attempt-p", func(st *screenplay.StoryboardGenerationState) error {
		for i := 1; i <= 2; i++ {
			key := []string{"old-key", "latest-key"}[i-1]
			st.Attempts[key] = &screenplay.StoryboardGenerationAttempt{RequestKey: key, Status: "succeeded", Engine: screenplay.EngineXiaoluoV7, EpisodeNumbers: []int{i}, Episodes: map[int]*screenplay.StoryboardGenerationEpisode{i: {EpisodeNo: i, Status: "succeeded", Calls: map[string]screenplay.StoryboardGenerationCall{"storyboard": {Status: "received", RawResponse: "raw-provider-private"}}, Storyboard: &screenplay.Storyboard{ID: "board", EpisodeNo: i, Shots: []screenplay.StoryboardShot{{VideoDesc: "保存的可见分镜"}}}}}, CreatedAt: time.Now().Add(time.Duration(i) * time.Second), Roles: "private-source-asset-text"}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayStoryboardGenerationRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, nil)
	base := "/api/comic-screenplays/projects/attempt-p/storyboard-generation-attempts"
	own := signedJSONRequest(t, mux, now, "alice", "GET", base+"/old-key", nil)
	if own.Code != 200 || !strings.Contains(own.Body.String(), "保存的可见分镜") || strings.Contains(own.Body.String(), "raw-provider-private") || strings.Contains(own.Body.String(), "private-source-asset-text") || strings.Contains(own.Body.String(), "raw_response") {
		t.Fatal(own.Code, own.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", "GET", base+"/old-key", nil)
	if foreign.Code == 200 {
		t.Fatal("owner attempt leak")
	}
	list := signedJSONRequest(t, mux, now, "alice", "GET", base+"?episode=1&engine="+string(screenplay.EngineXiaoluoV7), nil)
	var out struct {
		Attempts []struct {
			Key string `json:"request_key"`
		} `json:"attempts"`
	}
	if e = json.Unmarshal(list.Body.Bytes(), &out); e != nil || len(out.Attempts) != 1 || out.Attempts[0].Key != "old-key" {
		t.Fatal(list.Code, list.Body.String(), e)
	}
}
