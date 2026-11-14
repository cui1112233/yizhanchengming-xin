package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestPlanningSelectedBodyAndChangedContext(t *testing.T) {
	ctx := context.Background()
	_, store, _ := newStoryboardService(t, PermissionBypass)
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
	p, err := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	if err != nil {
		t.Fatal(err)
	}
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, err := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "章节", SourceText: "全文原文"}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	in := StoryboardPlanRequest{ChapterID: "c", Model: "own-text", UseOriginal: true, ExpectedRevision: &state.Revision, IdempotencyKey: "one"}
	gen := &fakeGenerator{responses: []string{"完整规划"}}
	out, err := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen)
	if err != nil {
		t.Fatal(err)
	}
	selected, _, _, _ := storyboardPlanningContext(p)
	if !strings.Contains(gen.requests[0].SystemPrompt, selected.Body) || out.Chapter.PlanningPromptID != selected.ID || out.Chapter.PlanningPromptDigest != contentDigest(selected.Body) {
		t.Fatal("selected body omitted", out)
	}
	if _, err = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); err != nil || gen.callCount() != 1 {
		t.Fatal("repeat billed", err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancelRequest := in
	cancelRequest.IdempotencyKey = "cancel-after-good"
	cancelRequest.ExpectedRevision = &out.Revision
	cancelled, cancelErr := authors.GenerateStoryboardPlan(cancelCtx, "alice", "p1", cancelRequest, lateCanceledGenerator{cancel: cancel, content: "晚到回包"})
	if !errors.Is(cancelErr, context.Canceled) || cancelled.Chapter.PlanningText != out.Chapter.PlanningText || cancelled.Chapter.PlanningPromptDigest != out.Chapter.PlanningPromptDigest || cancelled.Chapter.PlanningConfigDigest != out.Chapter.PlanningConfigDigest {
		t.Fatal("cancel overwrote good result", cancelled, cancelErr)
	}
	state, _ = authors.Load(ctx, "alice", "p1")
	use := true
	if _, err = storyboardPlanningText(StoryboardRequest{ChapterID: "c", UsePlanning: &use}, &state, 1, p); err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{"body", "style", "density", "selected"} {
		altered := p
		changed := EffectiveCreativeConfig(p)
		altered.CreativeConfig = &changed
		switch change {
		case "body":
			frozen := selected
			frozen.Body += "正文版本更新"
			changed.PromptSnapshots = map[string]CreativePromptSnapshot{"storyboard": frozen}
		case "style":
			changed.Style = &CreativeStyleSnapshot{ID: "new", Prompt: "新画风"}
		case "density":
			changed.PlanDensity = "more"
		case "selected":
			changed.StoryboardPromptID = "storyboard_xiaoyunque_v3"
		}
		if _, err = storyboardPlanningText(StoryboardRequest{ChapterID: "c", UsePlanning: &use}, &state, 1, altered); !errors.Is(err, ErrInvalidState) {
			t.Fatal(change, err)
		}
	}
	cfg = EffectiveCreativeConfig(p)
	cfg.StoryboardPromptID = "storyboard_xiaoyunque_v3"
	_, _ = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	in.IdempotencyKey = "cancel-after-good"
	in.ExpectedRevision = &state.Revision
	if _, err = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); !errors.Is(err, ErrInvalidArgument) || gen.callCount() != 1 {
		t.Fatal("changed prompt rebilled", err)
	}
}
func TestPlanningEmptySelectionDoesNotGenerate(t *testing.T) {
	ctx := context.Background()
	_, store, _ := newStoryboardService(t, PermissionBypass)
	authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, _ := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "章节", SourceText: "原文"}}}, 0)
	gen := &fakeGenerator{}
	_, err := authors.GenerateStoryboardPlan(ctx, "alice", "p1", StoryboardPlanRequest{ChapterID: "c", UseOriginal: true, Model: "own", IdempotencyKey: "empty", ExpectedRevision: &state.Revision}, gen)
	if err == nil || gen.callCount() != 0 {
		t.Fatal(err, gen.callCount())
	}
}
