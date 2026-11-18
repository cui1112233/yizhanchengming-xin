package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func shortInput(key string) ShortStoryInput {
	return ShortStoryInput{IdempotencyKey: key, Mode: "short_imitation", ModelID: "fake-only", ReferenceText: strings.Repeat("甲", 100)}
}

func TestShortStoryCancelAndFailedRetryAreRealStates(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{}
	failResolve := false
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		if failResolve {
			failResolve = false
			return nil, &DomainError{Code: ErrModelUnavailable, Message: "模型不可用，尚未发送"}
		}
		return provider, nil
	}))
	out, e := s.StartShortStory(ctx, "alice", shortInput("cancel"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.CancelRun(ctx, "alice", out.Run.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Run(ctx, "alice", out.Run.ID); e != nil {
		t.Fatal(e)
	}
	if len(provider.requests) != 0 {
		t.Fatal("cancelled run called model")
	}
	story, e := s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || story.Status != "cancelled" {
		t.Fatal("fake generating after cancellation", e)
	}
	failResolve = true
	out, e = s.StartShortStory(ctx, "alice", shortInput("failure"))
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e == nil || run.Status != RunFailed {
		t.Fatal("upstream failure not surfaced", run.Status, e)
	}
	story, e = s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || story.Status != "failed" {
		t.Fatal("fake generating after failure", story, e)
	}
	retry, e := s.RetryRun(ctx, "alice", run.ID)
	if e != nil {
		t.Fatal("retry unavailable", e)
	}
	if retry.ID == run.ID {
		t.Fatal("no retry run")
	}
	// Empty responses fail validation, but adopting the retry must not loop on a
	// stale short-story run pointer or silently claim completion.
	retried, _ := s.Run(ctx, "alice", retry.ID)
	if retried.Status != RunFailed {
		t.Fatal("retry did not run actual stages", retried.Status)
	}
	story, e = s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || story.RunID != retry.ID || story.Status != "failed" {
		t.Fatal("retry pointer/status not persisted", story, e)
	}
}

func TestShortStoryUncertainProviderDoesNotPermitRebilling(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{errors: []error{errors.New("connection lost after send")}}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	out, e := s.StartShortStory(ctx, "alice", shortInput("uncertain"))
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if !IsCode(e, ErrResultUnknown) || run.Status != RunPaused {
		t.Fatal("unknown result was not paused", run.Status, e)
	}
	if _, e = s.RetryRun(ctx, "alice", run.ID); e == nil {
		t.Fatal("unknown result allowed automatic rebilling")
	}
	if len(provider.requests) != 1 {
		t.Fatal("duplicate paid attempt")
	}
}

func TestShortStoryRejectsReferencesAndUnsupportedSizeBeforeDispatch(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), nil)
	for _, in := range []ShortStoryInput{func() ShortStoryInput { v := shortInput("size"); v.TargetWordCount = 500001; return v }(), func() ShortStoryInput { v := shortInput("style"); v.KnowledgeBookIDs = []string{"not-owned"}; return v }(), func() ShortStoryInput {
		v := shortInput("memo")
		v.MemoReferences = []ImitationMemoReference{{ProjectID: "other", NoteID: "other"}}
		return v
	}()} {
		if _, e := s.StartShortStory(ctx, "alice", in); e == nil {
			t.Fatal("invalid request dispatched")
		}
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 0 {
		t.Fatal("invalid input left hidden project")
	}
}
func TestShortStoryOriginalBoundariesAndReplay(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), nil)
	in := shortInput("short-one")
	out, e := s.StartShortStory(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if out.Story.Input.TargetWordCount != 13000 || out.Story.Input.TargetChapters != 1 || out.Story.ID == "" {
		t.Fatalf("wrong defaults: %+v", out)
	}
	again, e := s.StartShortStory(ctx, "alice", in)
	if e != nil || again.Run.ID != out.Run.ID {
		t.Fatal("replay duplicated", e)
	}
	in.ReaderPreferences = "changed"
	if _, e = s.StartShortStory(ctx, "alice", in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal("mutable request accepted", e)
	}
	if _, e = s.GetShortStory(ctx, "bob", out.Project.ID); e == nil {
		t.Fatal("cross account read")
	}
	for _, text := range []string{"", strings.Repeat("甲", 99), strings.Repeat(" \n", 100)} {
		in = shortInput("bad")
		in.ReferenceText = text
		if _, e = normalizeShortStory(in); e == nil {
			t.Fatal("invalid reference accepted")
		}
	}
	in = shortInput("emoji")
	in.ReferenceText = strings.Repeat("😀", 50)
	if _, e = normalizeShortStory(in); e != nil {
		t.Fatal(e)
	}
	chapters, _ := st.ListChapters(ctx, "alice", out.Project.ID)
	if len(chapters) != 0 {
		t.Fatal("creation made fake chapter")
	}
}
func TestShortStoryIndependentResultCheckpointEditAndReadback(t *testing.T) {
	body, _ := json.Marshal(shortText{Content: strings.Repeat("完整正文。", 200)})
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: []string{
		`{"content":"情绪递进和核心价值分析"}`,
		`{"title":"新故事","summary":"寻找真相","chapters":[{"number":1,"title":"雨夜","summary":"发现线索"}]}`,
		`{"score":95,"issues":[]}`,
		string(body),
		`{"score":96,"issues":[]}`,
		`{"content":"一场雨夜揭开的真相。"}`,
	}}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	in := shortInput("flow")
	in.PauseAfterOutline = true
	in.TargetWordCount = 1000
	out, e := s.StartShortStory(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation {
		t.Fatal("not checkpointed", run, e)
	}
	story, e := s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || story.Outline == nil || len(story.Chapters) != 0 {
		t.Fatal("checkpoint not durable", e)
	}
	story.Outline.Summary = "人工修改的大纲"
	updated, e := s.EditShortStoryOutline(ctx, "alice", out.Project.ID, story.Version, *story.Outline)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.EditShortStoryOutline(ctx, "alice", out.Project.ID, story.Version, *story.Outline); e == nil {
		t.Fatal("stale edit accepted")
	}
	continued, e := s.ConfirmShortStory(ctx, "alice", out.Run.ID, updated.Version)
	if e != nil {
		t.Fatal(e)
	}
	replay, e := s.ConfirmShortStory(ctx, "alice", out.Run.ID, updated.Version)
	if e != nil || replay.ID != continued.ID {
		t.Fatal("double confirmation", e)
	}
	run, e = s.Run(ctx, "alice", continued.ID)
	if e != nil || run.Status != RunSucceeded {
		t.Fatal("not complete", run, e)
	}
	got, e := s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || len(got.Chapters) != 1 || len(got.Artifacts) < 3 || got.Recommendation == "" {
		t.Fatal("missing independent results", got, e)
	}
	b, _ := json.Marshal(provider.requests)
	if !strings.Contains(string(b), "人工修改的大纲") {
		t.Fatal("ignored edited outline")
	}
	chapters, _ := st.ListChapters(ctx, "alice", out.Project.ID)
	if len(chapters) != 0 {
		t.Fatal("short saved as ordinary chapter")
	}
	before := len(provider.requests)
	if _, e = s.Run(ctx, "alice", continued.ID); e != nil || len(provider.requests) != before {
		t.Fatal("repeat run charged again", e)
	}
}

func TestShortStoryTwentyThousandWordsKeepsOneChapterAndBoundedRequests(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: []string{`{"content":"参考写法分析"}`, `{"title":"完整短篇","summary":"追查线索直到真相大白","chapters":[{"number":1,"title":"雨夜","summary":"发现线索，揭开真相"}]}`, `{"score":95,"issues":[]}`}}
	body, _ := json.Marshal(shortText{Content: strings.Repeat("甲", 4000)})
	for i := 0; i < 5; i++ {
		provider.responses = append(provider.responses, string(body), `{"score":95,"issues":[]}`)
		if i < 4 {
			provider.responses = append(provider.responses, `{"content":"人物已查明上一段的线索，仍需寻找失踪者。"}`)
		}
	}
	provider.responses = append(provider.responses, `{"content":"雨夜中的追寻与真相。"}`)
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	in := shortInput("twenty-thousand")
	in.TargetWordCount = 20000
	out, e := s.StartShortStory(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunSucceeded {
		t.Fatal(run.Status, e)
	}
	got, e := s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || len(got.Chapters) != 1 || len([]rune(got.Chapters[0].Content)) < 20000 {
		t.Fatal("lost selected chapter count or content", e)
	}
	for _, req := range provider.requests {
		b, _ := json.Marshal(req)
		if len(b) > 60000 {
			t.Fatal("unbounded model request", len(b))
		}
	}
}
