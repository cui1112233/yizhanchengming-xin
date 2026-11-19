package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func existingPlanningService(t *testing.T) (*AuthoringService, AuthoringState) {
	t.Helper()
	ctx := context.Background()
	_, store, _ := newStoryboardService(t, PermissionBypass)
	p, e := store.GetProject(ctx, "alice", "p1")
	if e != nil {
		t.Fatal(e)
	}
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
	if _, e = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision); e != nil {
		t.Fatal(e)
	}
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	st, e := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "本集", ScriptText: "自造完整剧本", PlanningText: "OLD_SAVED_PLAN"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	return authors, st
}

func TestPlanningExistingServiceReplayAndInputIdentity(t *testing.T) {
	ctx := context.Background()
	authors, st := existingPlanningService(t)
	calls := 0
	gen := planningFunction(func(_ context.Context, r TextRequest) (TextResponse, error) {
		calls++
		if len(r.Messages) != 4 || r.Messages[2].Content != "OLD_SAVED_PLAN" {
			t.Fatal("saved seed not sent")
		}
		return TextResponse{Content: "NEW_APPEND"}, nil
	})
	limit, auto := 8, true
	in := StoryboardPlanRequest{Continue: true, ChapterID: "c", Model: "own", IdempotencyKey: "one", ExpectedRevision: &st.Revision, MaxRounds: &limit, AutoContinue: &auto}
	first, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen)
	if e != nil {
		t.Fatal(e)
	}
	if calls != 1 || first.Chapter.PlanningRounds != 1 || first.Chapter.PlanningText != "OLD_SAVED_PLAN\nNEW_APPEND" {
		t.Fatal("explicit continuation must be one appended round", calls, first)
	}
	replay, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen)
	if e != nil || calls != 1 || replay.Revision != first.Revision || replay.Chapter.PlanningText != first.Chapter.PlanningText {
		t.Fatal("saved fallback changed replay digest", e, calls)
	}
	in.Continue = false
	in.IsContinue = true
	if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); e != nil || calls != 1 {
		t.Fatal("alias not equivalent", e)
	}
	for _, change := range []func(*StoryboardPlanRequest){func(r *StoryboardPlanRequest) { r.ExistingPlan = "DIFFERENT" }, func(r *StoryboardPlanRequest) { r.IsContinue = false }} {
		changed := in
		change(&changed)
		if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", changed, gen); !errors.Is(e, ErrInvalidArgument) || calls != 1 {
			t.Fatal("changed request billed", e, calls)
		}
	}
}

func TestPlanningExistingServiceEmptyFailureCancelAndLongSeed(t *testing.T) {
	for _, kind := range []string{"empty", "error", "cancel", "long", "ignored", "whitespace"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			authors, st := existingPlanningService(t)
			calls := 0
			in := StoryboardPlanRequest{Continue: true, ExistingPlan: "UNSAVED_SEED", ChapterID: "c", Model: "own", IdempotencyKey: "one", ExpectedRevision: &st.Revision}
			if kind == "long" {
				in.ExistingPlan = strings.Repeat("长", 100001)
			}
			if kind == "ignored" {
				in.Continue = false
			}
			if kind == "whitespace" {
				in.ExistingPlan = "   "
			}
			gen := planningFunction(func(_ context.Context, r TextRequest) (TextResponse, error) {
				calls++
				if kind == "ignored" || kind == "whitespace" {
					if len(r.Messages) != 2 {
						t.Fatal("ignored/blank seed sent")
					}
				} else if kind == "long" {
					if len([]rune(r.Messages[2].Content)) != 32000 {
						t.Fatal("native assistant tail differs")
					}
				}
				if kind == "cancel" {
					cancel()
				}
				if kind == "error" {
					return TextResponse{}, errors.New("OWN_FAILURE")
				}
				if kind == "empty" {
					return TextResponse{}, nil
				}
				return TextResponse{Content: originalCompletePlanningFixture(t)}, nil
			})
			out, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen)
			if calls != 1 {
				t.Fatal("unexpected calls", calls)
			}
			saved, loadErr := authors.Load(context.Background(), "alice", "p1")
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			switch kind {
			case "error", "cancel":
				if e == nil || saved.Chapters[0].PlanningText != "OLD_SAVED_PLAN" {
					t.Fatal("failure damaged saved plan", e)
				}
			case "empty":
				if e != nil || out.Chapter.PlanningText != "OLD_SAVED_PLAN" || out.Chapter.PlanningRounds != 1 {
					t.Fatal("empty response saved unsaved seed", e)
				}
			case "long":
				if e != nil || len([]rune(out.Chapter.PlanningText)) != 102501 {
					t.Fatal("long existing plan incorrectly capped", e, len([]rune(out.Chapter.PlanningText)))
				}
			default:
				if e != nil || out.Chapter.PlanningText != originalCompletePlanningFixture(t) {
					t.Fatal("fresh result incorrect", e)
				}
			}
		})
	}
}
