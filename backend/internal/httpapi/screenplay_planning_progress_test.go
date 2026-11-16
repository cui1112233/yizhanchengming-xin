package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
	"sync"
	"testing"
	"time"
)

type planningProgressTestRepo struct {
	mu   sync.Mutex
	row  screenplay.PlanningProgress
	fail bool
}

func (p *planningProgressTestRepo) Put(_ context.Context, _, _ string, v screenplay.PlanningProgress) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail {
		return errors.New("controlled Redis-down")
	}
	p.row = v
	return nil
}
func (p *planningProgressTestRepo) Get(_ context.Context, _, _, key string) (screenplay.PlanningProgress, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail {
		return p.row, screenplay.ErrPlanningProgressUnavailable
	}
	if p.row.RequestKey != key {
		return p.row, screenplay.ErrNotFound
	}
	return p.row, nil
}
func (p *planningProgressTestRepo) setFail(v bool) { p.mu.Lock(); defer p.mu.Unlock(); p.fail = v }

func TestScreenplayPlanningProgressDuringRequestAndCancel(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(fmt.Sprintf("cancel=%t", cancelled), func(t *testing.T) {
			ctx := context.Background()
			store := screenplay.NewMemoryStore()
			_, err := store.CreateProject(ctx, screenplay.Project{ID: "progress-p", OwnerUsername: "alice", CreativeConfig: &screenplay.CreativeConfig{StoryboardPromptID: "storyboard_xiaoluo_v7_15s"}})
			if err != nil {
				t.Fatal(err)
			}
			repo := &planningProgressTestRepo{}
			authors := &screenplay.AuthoringService{Store: store, Repository: screenplay.NewMemoryAuthoringRepository(), PlanningProgress: repo}
			state, err := authors.Save(ctx, "alice", "progress-p", screenplay.AuthoringState{Chapters: []screenplay.Chapter{{ID: "c", Title: "本集", ScriptText: "自造剧本", PlanningText: "OLD_SAVED_PLAN"}}}, 0)
			if err != nil {
				t.Fatal(err)
			}
			blocked, release := make(chan struct{}), make(chan struct{})
			calls := 0
			provider := screenplayProviderFunc(func(c context.Context, req batchfactoryv11.TextCompletionRequest) (string, error) {
				calls++
				if calls == 2 {
					close(blocked)
					if len(req.Messages) != 4 {
						t.Error("continuation history missing")
					}
					select {
					case <-release:
					case <-c.Done():
						return "", c.Err()
					}
				}
				return "## 一、总览\nOWN_PARTIAL", nil
			})
			mux := http.NewServeMux()
			now := time.Now()
			RegisterScreenplayStoryboardPlanRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, authors, func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil })
			// The signed request helper uses a background request. This adapter attaches
			// a cancellable context to the real handler, without changing its body/auth.
			runCtx, stop := context.WithCancel(ctx)
			defer stop()
			entry := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { mux.ServeHTTP(w, r.WithContext(runCtx)) })
			done := make(chan int, 1)
			go func() {
				resp := signedJSONRequest(t, entry, now, "alice", "POST", "/api/comic-screenplays/projects/progress-p/storyboard-plan", map[string]any{"chapter_id": "c", "model": "own-model", "expected_revision": state.Revision, "idempotencyKey": "one", "max_rounds": 2})
				done <- resp.Code
			}()
			select {
			case <-blocked:
			case <-time.After(5 * time.Second):
				t.Fatal("second round not reached")
			}
			path := "/api/comic-screenplays/projects/progress-p/storyboard-plan-progress?chapter_id=c&request_key=one"
			resp := signedJSONRequest(t, mux, now, "alice", "GET", path, nil)
			if resp.Code != 200 {
				t.Fatal(resp.Code, resp.Body.String())
			}
			var read struct {
				Status   string
				Progress screenplay.PlanningProgress
				Revision int64
			}
			if err = json.Unmarshal(resp.Body.Bytes(), &read); err != nil {
				t.Fatal(err)
			}
			if read.Status != "running" || read.Progress.Round != 2 || read.Progress.Chars != len([]rune("## 一、总览\nOWN_PARTIAL")) || read.Progress.RequestKey != "one" || read.Progress.ChapterID != "c" || read.Revision != state.Revision+1 {
				t.Fatal("live progress is not actual round", read)
			}
			existing, _ := authors.Load(ctx, "alice", "progress-p")
			if existing.Chapters[0].PlanningText != "OLD_SAVED_PLAN" {
				t.Fatal("progress overwrote old plan")
			}
			if bad := signedJSONRequest(t, mux, now, "bob", "GET", path, nil); bad.Code != 403 {
				t.Fatal("other user read progress", bad.Code)
			}
			if bad := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects/progress-p/storyboard-plan-progress?chapter_id=c&request_key=old", nil); bad.Code != 409 {
				t.Fatal("old request accepted", bad.Code)
			}
			repo.setFail(true)
			if bad := signedJSONRequest(t, mux, now, "alice", "GET", path, nil); bad.Code != 503 {
				t.Fatal("Redis failure hidden", bad.Code)
			}
			repo.setFail(false)
			if cancelled {
				stop()
			} else {
				close(release)
			}
			select {
			case code := <-done:
				want := 200
				if cancelled {
					want = 499
				}
				if code != want {
					t.Fatal("unexpected completion", code, want)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("handler did not stop")
			}
			fresh, _ := authors.Load(ctx, "alice", "progress-p")
			if cancelled && fresh.Chapters[0].PlanningText != "OLD_SAVED_PLAN" {
				t.Fatal("cancel lost old plan")
			}
			if calls != 2 {
				t.Fatal("unexpected model call", calls)
			}
			repo.setFail(true)
			final := signedJSONRequest(t, mux, now, "alice", "GET", path, nil)
			if final.Code != 200 {
				t.Fatal("canonical result unavailable after Redis failure", final.Code)
			}
		})
	}
}
