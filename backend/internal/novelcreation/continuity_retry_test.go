package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func retryContinuityBundle(input RunContext, status RunStatus) ChapterBundle {
	b := serviceTestBundle(input, status)
	b.Memories = []Memory{{ID: input.Run.ID + "-memory", ProjectID: input.Project.ID, ChapterID: b.Chapter.ID, Snapshot: json.RawMessage(`{"summary":"最新记忆"}`)}}
	b.RuntimeSnapshot = json.RawMessage(`{"truthSettlement":{"facts":["最新事实"]},"memoryIndex":{"summary":"最新记忆"}}`)
	return b
}
func TestContinuityRetryCreatesNewVersionAndNextChapter(t *testing.T) {
	calls := 0
	runner := &stubPipelineRunner{run: func(_ context.Context, in RunContext, _ EventSink) (ChapterBundle, error) {
		calls++
		status := RunSucceededWithWarnings
		if calls > 1 {
			status = RunSucceeded
		}
		return retryContinuityBundle(in, status), nil
	}}
	f := newServiceFixture(t, runner)
	ctx := context.Background()
	first, err := f.service.SendMessage(ctx, "alice", f.sendInput("first"))
	if err != nil {
		t.Fatal(err)
	}
	first, err = f.service.Run(ctx, "alice", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := f.service.RetryRun(ctx, "alice", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Run(ctx, "alice", retry.ID); err != nil {
		t.Fatal(err)
	}
	b := f.store.chapterBundles[memoryKey("alice", f.project.ID, f.project.ID+"-chapter-1")]
	if b.Chapter.Version != 2 || b.ChapterVersion.Version != 2 || b.ChapterVersion.RunID != retry.ID {
		t.Fatalf("retry reused existing version: chapter=%d version=%d run=%s", b.Chapter.Version, b.ChapterVersion.Version, b.ChapterVersion.RunID)
	}
	firstVersion, exists := f.store.chapterVersions[memoryKey("alice", f.project.ID, b.Chapter.ID, "1")]
	if !exists || firstVersion.RunID != first.ID {
		t.Fatal("retry discarded the immutable original version")
	}
	next, err := f.service.SendMessage(ctx, "alice", f.sendInput("second"))
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Chapters []PreviousChapter `json:"previous_chapters"`
		Truth    json.RawMessage   `json:"previous_truth"`
	}
	if err = json.Unmarshal(next.PromptSnapshot, &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Chapters) != 1 || snapshot.Chapters[0].Version != 2 || string(snapshot.Truth) != `{"facts":["最新事实"]}` {
		t.Fatalf("next chapter lost retry continuity: %s", next.PromptSnapshot)
	}
}
func TestContinuityRetryExpectedAbsentRejectsNewChapter(t *testing.T) {
	f := newServiceFixture(t, &stubPipelineRunner{run: func(context.Context, RunContext, EventSink) (ChapterBundle, error) {
		return ChapterBundle{}, errors.New("failed")
	}})
	ctx := context.Background()
	first, err := f.service.SendMessage(ctx, "alice", f.sendInput("first"))
	if err != nil {
		t.Fatal(err)
	}
	first, err = f.service.Run(ctx, "alice", first.ID)
	if err == nil {
		t.Fatal("wanted initial failure")
	}
	retry, err := f.service.RetryRun(ctx, "alice", first.ID)
	if err != nil {
		t.Fatal(err)
	}
	other, err := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_form_original"})
	if err != nil {
		t.Fatal(err)
	}
	newer := Run{ID: "another-session-run", ProjectID: f.project.ID, SessionID: other.ID}
	b := retryContinuityBundle(RunContext{Project: f.project, Run: newer}, RunSucceeded)
	b.Chapter.Content = "另一会话已经写好的第一章"
	b.ChapterVersion.Content = b.Chapter.Content
	b.RunID = "" // Simulate the other session's already committed chapter without dispatching a model.
	if err = f.store.SaveChapterBundle(ctx, "alice", b); err != nil {
		t.Fatal(err)
	}
	before := f.runner.callCount()
	f.runner.run = func(_ context.Context, in RunContext, _ EventSink) (ChapterBundle, error) {
		return serviceTestBundle(in, RunSucceeded), nil
	}
	if _, err = f.service.Run(ctx, "alice", retry.ID); err == nil {
		t.Fatal("retry accepted a target expected absent")
	}
	if f.runner.callCount() != before {
		t.Fatal("retry called pipeline after another session saved target")
	}
	chapters, _ := f.store.ListChapters(ctx, "alice", f.project.ID)
	if len(chapters) != 1 || chapters[0].Content != b.Chapter.Content || chapters[0].Version != b.Chapter.Version {
		t.Fatal("new chapter overwritten")
	}
}
func TestContinuityMemoryStoreRejectsHistoricalVersionOverwrite(t *testing.T) {
	f := newServiceFixture(t, nil)
	continuityFirstChapter(t, f)
	key := memoryKey("alice", f.project.ID, f.project.ID+"-chapter-1")
	old := cloneChapterBundle(f.store.chapterBundles[key])
	changed := cloneChapterBundle(old)
	changed.Chapter.Content = "不应覆盖旧版"
	changed.ChapterVersion.Content = changed.Chapter.Content
	changed.ChapterVersion.RunID = "another-run"
	if err := f.store.SaveChapterBundle(context.Background(), "alice", changed); err == nil {
		t.Fatal("historical chapter version overwritten")
	}
	if f.store.chapterBundles[key].Chapter.Content != old.Chapter.Content {
		t.Fatal("rejected save changed current content")
	}
}
