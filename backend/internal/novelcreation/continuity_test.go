package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func continuityFirstChapter(t *testing.T, f serviceFixture) {
	t.Helper()
	ctx := context.Background()
	run, err := f.service.SendMessage(ctx, "alice", f.sendInput("first"))
	if err != nil {
		t.Fatal(err)
	}
	b := serviceTestBundle(RunContext{Project: f.project, Run: run}, RunSucceeded)
	b.Foundation = Foundation{ID: "foundation", ProjectID: f.project.ID, Snapshot: json.RawMessage(`{"plan":{"title":"蓝图"},"truthSettlement":{"facts":["钥匙已交出"]}}`)}
	b.Memories = []Memory{{ID: "memory", ProjectID: f.project.ID, ChapterID: b.Chapter.ID, Snapshot: json.RawMessage(`{"summary":"记者拿到匿名信"}`)}}
	b.RuntimeSnapshot = json.RawMessage(`{"truthSettlement":{"facts":["钥匙已交出"]},"memoryIndex":{"summary":"记者拿到匿名信"}}`)
	if err = f.store.SaveChapterBundle(ctx, "alice", b); err != nil {
		t.Fatal(err)
	}
	if _, err = f.store.UpdateRun(ctx, "alice", RunUpdate{ID: run.ID, Status: RunSucceeded, RuntimeSnapshot: b.RuntimeSnapshot}); err != nil {
		t.Fatal(err)
	}
}
func TestContinuitySecondChapterUsesPersistedFirstChapter(t *testing.T) {
	runner := &stubPipelineRunner{run: func(_ context.Context, input RunContext, _ EventSink) (ChapterBundle, error) {
		bundle := serviceTestBundle(input, RunSucceeded)
		// This case writes chapter two; the generic fixture intentionally writes chapter one.
		bundle.Chapter.ID = input.Project.ID + "-chapter-2"
		bundle.Chapter.Number = 2
		bundle.ChapterVersion.ID = bundle.Chapter.ID + "-version-" + input.Run.ID
		bundle.ChapterVersion.ChapterID = bundle.Chapter.ID
		return bundle, nil
	}}
	f := newServiceFixture(t, runner)
	continuityFirstChapter(t, f)
	r, e := f.service.SendMessage(context.Background(), "alice", f.sendInput("second"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.service.Run(context.Background(), "alice", r.ID); e != nil {
		t.Fatal(e)
	}
	g := f.runner.lastCall()
	if string(g.PreviousMemory) != `{"summary":"记者拿到匿名信"}` {
		t.Fatalf("lost persisted chapter memory: %s", g.PreviousMemory)
	}
	if string(g.PreviousTruth) != `{"facts":["钥匙已交出"]}` {
		t.Fatalf("lost persisted facts: %s", g.PreviousTruth)
	}
	var p struct {
		Chapters []struct {
			Version int
			Ending  string
		} `json:"previous_chapters"`
	}
	json.Unmarshal(r.PromptSnapshot, &p)
	if len(p.Chapters) != 1 || p.Chapters[0].Version != 1 || p.Chapters[0].Ending != "雨水敲着窗。" {
		t.Fatalf("missing version/ending: %s", r.PromptSnapshot)
	}
}

type continuityFailStore struct{ Store }

func (s continuityFailStore) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, owner, project, func(tx Store) error { return fn(continuityFailStore{tx}) })
}

func (s continuityFailStore) GetChapterMemory(context.Context, string, string, string) (Memory, bool, error) {
	return Memory{}, false, errors.New("read failed")
}
func TestContinuityReadFailureStopsBeforePipeline(t *testing.T) {
	f := newServiceFixture(t, nil)
	continuityFirstChapter(t, f)
	f.service.store = continuityFailStore{f.store}
	if _, e := f.service.SendMessage(context.Background(), "alice", f.sendInput("second")); e == nil {
		t.Fatal("memory read failure accepted")
	}
	if f.runner.callCount() != 0 {
		t.Fatal("pipeline called")
	}
	m, _ := f.store.ListMessages(context.Background(), "alice", f.session.ID, 0)
	if len(m) != 1 {
		t.Fatal("failed read appended message")
	}
}
func TestContinuityMissingOrStaleMemoryStops(t *testing.T) {
	for _, kind := range []string{"missing", "edited", "wrong-memory"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			continuityFirstChapter(t, f)
			key := memoryKey("alice", f.project.ID, f.project.ID+"-chapter-1")
			b := f.store.chapterBundles[key]
			switch kind {
			case "missing":
				b.Memories = nil
			case "edited":
				b.Chapter.Version = 2
				b.Chapter.Content = "手动编辑"
			case "wrong-memory":
				b.Memories[0].Snapshot = json.RawMessage(`{"summary":"另一版"}`)
			}
			f.store.chapterBundles[key] = b
			if _, e := f.service.SendMessage(context.Background(), "alice", f.sendInput("second")); e == nil {
				t.Fatal("missing/stale continuity accepted")
			}
		})
	}
}
func TestSessionSnapshotEmptyAndAccountIsolation(t *testing.T) {
	f := newServiceFixture(t, nil)
	g, ok := any(f.service).(interface {
		GetSessionSnapshot(context.Context, string, string) (Snapshot, error)
	})
	if !ok {
		t.Fatal("GetSessionSnapshot missing")
	}
	s, e := g.GetSessionSnapshot(context.Background(), "alice", f.session.ID)
	if e != nil {
		t.Fatal(e)
	}
	if s.Session.ID != f.session.ID || s.Run.ID != "" || len(s.Messages) != 0 {
		t.Fatal("bad empty snapshot")
	}
	if _, e = g.GetSessionSnapshot(context.Background(), "bob", f.session.ID); !IsCode(e, ErrSessionNotFound) {
		t.Fatalf("foreign: %v", e)
	}
	continuityFirstChapter(t, f)
	s, e = g.GetSessionSnapshot(context.Background(), "alice", f.session.ID)
	if e != nil || s.Run.ID == "" || len(s.Messages) != 1 || len(s.Chapters) != 1 {
		t.Fatalf("lost saved state: %+v %v", s, e)
	}
	latest, e := f.service.SendMessage(context.Background(), "alice", f.sendInput("latest"))
	if e != nil {
		t.Fatal(e)
	}
	s, e = g.GetSessionSnapshot(context.Background(), "alice", f.session.ID)
	if e != nil || s.Run.ID != latest.ID || len(s.Messages) != 2 {
		t.Fatalf("latest session run not restored: %+v %v", s, e)
	}
	other, e := f.service.CreateSession(context.Background(), "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_form_original"})
	if e != nil {
		t.Fatal(e)
	}
	s, e = g.GetSessionSnapshot(context.Background(), "alice", other.ID)
	if e != nil || s.Run.ID != "" || len(s.Messages) != 0 {
		t.Fatal("session leaked")
	}
}
func TestContinuityRetryTargetVersionChangeStopsBeforePipeline(t *testing.T) {
	f := newServiceFixture(t, nil)
	source, err := f.service.SendMessage(context.Background(), "alice", f.sendInput("first"))
	if err != nil {
		t.Fatal(err)
	}
	f.runner.run = func(_ context.Context, input RunContext, _ EventSink) (ChapterBundle, error) {
		return serviceTestBundle(input, RunSucceededWithWarnings), nil
	}
	source, err = f.service.Run(context.Background(), "alice", source.ID)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := f.service.RetryRun(context.Background(), "alice", source.ID)
	if err != nil {
		t.Fatal(err)
	}
	key := memoryKey("alice", f.project.ID, f.project.ID+"-chapter-1")
	b := f.store.chapterBundles[key]
	b.Chapter.Version = 2
	f.store.chapterBundles[key] = b
	if _, err = f.service.Run(context.Background(), "alice", retry.ID); err == nil {
		t.Fatal("retry overwrote changed target version")
	}
	if f.runner.callCount() != 1 {
		t.Fatal("retry called model after target version changed")
	}
}
func TestContinuityPipelinePassesEndingAndFactsToDraft(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.PreviousTruth = json.RawMessage(`{"facts":["不能重新取回钥匙"]}`)
	input.PreviousChapters = []PreviousChapter{{ChapterID: "previous", Version: 3, Number: 1, Ending: "他关上了门。"}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(provider.requests[2])
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "他关上了门。") || !strings.Contains(text, "不能重新取回钥匙") {
		t.Fatalf("draft omitted continuity: %s", raw)
	}
	var version struct {
		Truth  json.RawMessage `json:"truthSettlement"`
		Memory json.RawMessage `json:"memoryIndex"`
	}
	if err = json.Unmarshal(bundle.ChapterVersion.Snapshot, &version); err != nil || len(version.Truth) == 0 || len(version.Memory) == 0 {
		t.Fatal("version omitted continuity")
	}
}
