package screenplay

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type planningFunction func(context.Context, TextRequest) (TextResponse, error)

func (f planningFunction) Generate(c context.Context, r TextRequest) (TextResponse, error) {
	return f(c, r)
}

func TestPlanningCancellationDuringContinuationPreservesPreviousPlan(t *testing.T) {
	ctx := context.Background()
	_, store, _ := newStoryboardService(t, PermissionBypass)
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
	_, e := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	if e != nil {
		t.Fatal(e)
	}
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, e := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "本集", ScriptText: "自造剧本"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	in := StoryboardPlanRequest{ChapterID: "c", Model: "own", ExpectedRevision: &state.Revision, IdempotencyKey: "complete"}
	old, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, &fakeGenerator{responses: []string{originalCompletePlanningFixture(t)}})
	if e != nil {
		t.Fatal(e)
	}
	cancelled, stop := context.WithCancel(ctx)
	calls := 0
	gen := planningFunction(func(c context.Context, r TextRequest) (TextResponse, error) {
		calls++
		if calls == 2 {
			stop()
		}
		return TextResponse{Content: "## 一、总览\n新部分稿"}, nil
	})
	in.IdempotencyKey = "cancel-next-round"
	in.ExpectedRevision = &old.Revision
	result, e := authors.GenerateStoryboardPlan(cancelled, "alice", "p1", in, gen)
	if !errors.Is(e, context.Canceled) || calls != 2 || result.Chapter.PlanningStatus != "cancelled" || result.Chapter.PlanningText != old.Chapter.PlanningText || !reflect.DeepEqual(result.Chapter.PlanningCoverage, old.Chapter.PlanningCoverage) {
		t.Fatal("continuation cancel damaged old plan", e, calls)
	}
	if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); !errors.Is(e, ErrInvalidState) || calls != 2 {
		t.Fatal("cancel automatically retried", e, calls)
	}
	limit := 1
	in.MaxRounds = &limit
	if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); !errors.Is(e, ErrInvalidArgument) || calls != 2 {
		t.Fatal("changed limits reused paid request", e, calls)
	}
}

func TestPlanningPreviousEpisodeChangeRejectsSameRequest(t *testing.T) {
	ctx := context.Background()
	_, store, _ := newStoryboardService(t, PermissionBypass)
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
	_, e := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	if e != nil {
		t.Fatal(e)
	}
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, e := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "prev", Title: "上集", PlanningText: "上集末尾"}, {ID: "current", Title: "本集", ScriptText: "自造剧本"}}}, 0)
	if e != nil {
		t.Fatal(e)
	}
	auto := false
	in := StoryboardPlanRequest{AutoContinue: &auto, ChapterID: "current", Model: "own", ExpectedRevision: &state.Revision, IdempotencyKey: "once"}
	gen := &fakeGenerator{responses: []string{"部分稿"}}
	result, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen)
	if e != nil {
		t.Fatal(e)
	}
	state, e = authors.Load(ctx, "alice", "p1")
	if e != nil {
		t.Fatal(e)
	}
	state.Chapters[0].PlanningText = "已修改的上集结尾"
	_, e = authors.Save(ctx, "alice", "p1", state, result.Revision)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); !errors.Is(e, ErrInvalidArgument) || gen.callCount() != 1 {
		t.Fatal("changed previous plan billed", e, gen.callCount())
	}
}
