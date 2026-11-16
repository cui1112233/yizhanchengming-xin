package screenplay

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
)

type testPlanningProgress struct {
	mu   sync.Mutex
	rows map[string]PlanningProgress
	fail bool
}

func (s *testPlanningProgress) Put(_ context.Context, o, p string, v PlanningProgress) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		return errors.New("controlled Redis unavailable")
	}
	if s.rows == nil {
		s.rows = map[string]PlanningProgress{}
	}
	s.rows[PlanningProgressRedisKey(o, p, v.RequestKey)] = v
	return nil
}
func (s *testPlanningProgress) Get(_ context.Context, o, p, k string) (PlanningProgress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		return PlanningProgress{}, ErrPlanningProgressUnavailable
	}
	v, ok := s.rows[PlanningProgressRedisKey(o, p, k)]
	if !ok {
		return v, ErrNotFound
	}
	return v, nil
}

func TestPlanningProgressOriginalEmissionGolden(t *testing.T) {
	f, e := os.Open("testdata/screenplay_plan_progress_original.json.gz")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var d struct {
		Cases []struct {
			Label  string
			Events []struct {
				Round, Total int
				Message      string
				Detail       struct {
					Type         string
					ChapterIndex int `json:"chapter_index"`
					Round, Chars int
					Complete     bool
					Missing      []string
				}
			}
		}
	}
	if e = json.NewDecoder(z).Decode(&d); e != nil {
		t.Fatal(e)
	}
	for _, row := range d.Cases {
		t.Run(row.Label, func(t *testing.T) {
			gen := &planningRoundGenerator{mode: row.Label, complete: originalCompletePlanningFixture(t)}
			frames := []PlanningProgress{}
			limit := 8
			if row.Label == "two-rounds" {
				limit = 2
			}
			_, _ = generateStoryboardPlanRounds(context.Background(), gen, "OWN_FAKE_MODEL", "system", "user", "standard", limit, func(p PlanningProgress) { frames = append(frames, p) })
			if len(frames) != len(row.Events) {
				t.Fatalf("events%d/%d", len(frames), len(row.Events))
			}
			for i, p := range frames {
				e := row.Events[i]
				if p.Type != e.Detail.Type || p.Round != e.Round || p.Total != e.Total || p.Chars != e.Detail.Chars || p.Complete != e.Detail.Complete || !reflect.DeepEqual(p.Missing, e.Detail.Missing) || p.Message != e.Message {
					t.Fatalf("frame%d differs: %+v", i, p)
				}
			}
		})
	}
}

func TestPlanningProgressDoesNotChangeAuthoringRevisionOrBlockOnRedisFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{true: "Redis-down", false: "Redis-up"}[fail], func(t *testing.T) {
			ctx := context.Background()
			_, store, _ := newStoryboardService(t, PermissionBypass)
			p, _ := store.GetProject(ctx, "alice", "p1")
			cfg := EffectiveCreativeConfig(p)
			cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
			_, err := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
			if err != nil {
				t.Fatal(err)
			}
			repo := &testPlanningProgress{fail: fail}
			authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository(), PlanningProgress: repo}
			state, err := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "本集", ScriptText: "自造剧本"}}}, 0)
			if err != nil {
				t.Fatal(err)
			}
			result, err := authors.GenerateStoryboardPlan(ctx, "alice", "p1", StoryboardPlanRequest{ChapterID: "c", Model: "own", ExpectedRevision: &state.Revision, IdempotencyKey: "one"}, &fakeGenerator{responses: []string{"## 一、总览\n部分稿"}})
			if err != nil || result.Chapter.PlanningRounds != 8 || result.Revision != state.Revision+2 {
				t.Fatal("progress changed result or revision", err, result.Revision)
			}
			if !fail {
				p, err := repo.Get(ctx, "alice", "p1", "one")
				if err != nil || p.Sequence != 17 || p.Round != 8 || p.Status != "partial" || p.RequestKey != "one" || p.ChapterID != "c" {
					t.Fatal("final progress missing", p, err)
				}
				if _, err = repo.Get(ctx, "other", "p1", "one"); !errors.Is(err, ErrNotFound) {
					t.Fatal("cross-owner key collision")
				}
			}
		})
	}
}
