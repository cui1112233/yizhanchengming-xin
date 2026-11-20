package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"sync"
	"testing"
)

func TestStudioShortRunCompletesPersistedOutputWithFakeProvider(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	body, _ := json.Marshal(shortText{Content: strings.Repeat("完整故事。", 200)})
	provider := &queuedPipelineProvider{responses: []string{
		`{"message":"请确认短篇","action":{"type":"short_run","title":"旧信","direction":"写寻找故乡的悬疑短篇","targetWordCount":1000,"targetChapters":1}}`,
		`{"title":"旧信","summary":"寻找故乡","chapters":[{"number":1,"title":"雨夜","summary":"发现来信"}]}`,
		`{"score":95,"issues":[]}`, string(body), `{"score":96,"issues":[]}`, `{"content":"一封旧信带来的真相。"}`,
	}}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	ss, e := s.CreateStudioSession(ctx, "alice", StudioCreateSessionInput{IdempotencyKey: "studio-short-flow", Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "proposal", Content: "创作短篇", ModelID: "our-model"})
	if e != nil || snap.PendingAction == nil {
		t.Fatalf("no proposal %+v %v", snap, e)
	}
	result, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "our-model"})
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", result.Run.ID)
	if e != nil || run.Status != RunSucceeded {
		t.Fatalf("pipeline failed %+v %v", run, e)
	}
	saved, e := s.GetShortStory(ctx, "alice", result.Project.ID)
	if e != nil || len(saved.Chapters) != 1 || saved.Chapters[0].Content == "" || saved.Recommendation == "" || len(saved.Artifacts) < 3 {
		t.Fatalf("missing durable output %+v %v", saved, e)
	}
	chapters, _ := st.ListChapters(ctx, "alice", result.Project.ID)
	if len(chapters) != 0 {
		t.Fatal("short saved as long chapter")
	}
	count := len(provider.requests)
	if count != 6 {
		t.Fatalf("unexpected stage count %d", count)
	}
	if _, e = s.Run(ctx, "alice", result.Run.ID); e != nil || len(provider.requests) != count {
		t.Fatal("repeat charged", e)
	}
}

func TestStudioShortRunConfirmationUsesIndependentShortTask(t *testing.T) {
	s, st, p, ss := studioFixture(t)
	ctx := context.Background()
	p.reply = `{"message":"请确认短篇方向","action":{"type":"short_run","title":"来信","direction":"小镇青年寻找神秘来信者","targetWordCount":13000,"targetChapters":2}}`
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "short-propose", Content: "写一篇悬疑短篇", ModelID: "own"})
	if e != nil || snap.PendingAction == nil || snap.PendingAction.Type != "short_run" {
		t.Fatalf("short proposal rejected: %+v %v", snap, e)
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 0 {
		t.Fatal("proposal created project")
	}
	in := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"}
	out, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, in)
	if e != nil {
		t.Fatal(e)
	}
	if out.ProductionSession.Mode != "short_original" {
		t.Fatalf("wrong production mode: %s", out.ProductionSession.Mode)
	}
	story, e := s.GetShortStory(ctx, "alice", out.Project.ID)
	if e != nil || story.Input.Requirements != "小镇青年寻找神秘来信者" || story.Input.TargetChapters != 2 || story.Input.TargetWordCount != 13000 || story.RunID != out.Run.ID {
		t.Fatalf("short input lost: %+v %v", story, e)
	}
	var frozen map[string]json.RawMessage
	if e = json.Unmarshal(out.Run.PromptSnapshot, &frozen); e != nil || frozen["short_story"] == nil {
		t.Fatal("not a short task", e)
	}
	again, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, in)
	if e != nil || again.Run.ID != out.Run.ID {
		t.Fatalf("duplicate confirmation: %+v %v", again, e)
	}
	projects, _ = st.ListProjects(ctx, "alice")
	if len(projects) != 1 || p.calls != 1 {
		t.Fatal("duplicate project or provider call")
	}
	if _, e = s.ConfirmStudioAction(ctx, "bob", ss.ID, in); e == nil {
		t.Fatal("cross-account confirmation")
	}
}

func TestStudioShortRunCancellationAndInvalidDirection(t *testing.T) {
	for _, body := range []string{
		`{"type":"short_run"}`,
		`{"type":"short_run","direction":"   "}`,
		`{"type":"short_run","direction":"方向","targetChapters":51}`,
		`{"type":"unsupported","direction":"方向"}`,
	} {
		t.Run(body, func(t *testing.T) {
			s, st, p, ss := studioFixture(t)
			p.reply = `{"message":"建议","action":` + body + `}`
			snap, e := s.SendStudioMessage(context.Background(), "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "输入", ModelID: "own"})
			if e != nil || snap.Turn.State != "invalid_output" || snap.PendingAction != nil {
				t.Fatalf("invalid action accepted %+v %v", snap, e)
			}
			projects, _ := st.ListProjects(context.Background(), "alice")
			if len(projects) != 0 {
				t.Fatal("invalid action created project")
			}
		})
	}
	s, st, p, ss := studioFixture(t)
	p.reply = `{"message":"建议","action":{"type":"short_run","instruction":"写一个完整悬疑故事"}}`
	snap, e := s.SendStudioMessage(context.Background(), "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "输入", ModelID: "own"})
	if e != nil || snap.PendingAction == nil {
		t.Fatalf("missing action %+v %v", snap, e)
	}
	in := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"}
	if _, e = s.CancelStudioAction(context.Background(), "alice", ss.ID, in); e != nil {
		t.Fatal(e)
	}
	if _, e = s.ConfirmStudioAction(context.Background(), "alice", ss.ID, in); e == nil {
		t.Fatal("cancelled action started")
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	if len(projects) != 0 {
		t.Fatal("cancel created project")
	}
}

func TestStudioShortDirectionFallbackAndAtomicConfirmation(t *testing.T) {
	for _, field := range []string{"direction", "instruction", "summary", "title"} {
		t.Run(field, func(t *testing.T) {
			s, st, p, ss := studioFixture(t)
			ctx := context.Background()
			p.reply = `{"message":"请确认","action":{"type":"short_run","` + field + `":"完整故事方向"}}`
			snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "短篇", ModelID: "own"})
			if e != nil || snap.PendingAction == nil {
				t.Fatalf("proposal missing %+v %v", snap, e)
			}
			in := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"}
			fail := true
			faulted := NewService(studioFaultStore{Store: st, state: "confirmed", fail: &fail}, NewMemoryRuntime(), studioTestPipeline{provider: p})
			if _, e = faulted.ConfirmStudioAction(ctx, "alice", ss.ID, in); e == nil {
				t.Fatal("injected failure ignored")
			}
			projects, _ := st.ListProjects(ctx, "alice")
			if len(projects) != 0 {
				t.Fatal("failed receipt left orphan short task")
			}
			var wg sync.WaitGroup
			results := make(chan StudioTransition, 4)
			errs := make(chan error, 4)
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					o, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, in)
					results <- o
					errs <- e
				}()
			}
			wg.Wait()
			close(results)
			close(errs)
			for e := range errs {
				if e != nil {
					t.Fatal(e)
				}
			}
			id := ""
			for o := range results {
				if id != "" && id != o.Run.ID {
					t.Fatal("duplicate run")
				}
				id = o.Run.ID
			}
			projects, _ = st.ListProjects(ctx, "alice")
			if len(projects) != 1 || p.calls != 1 {
				t.Fatal("duplicate project/call")
			}
			story, e := s.GetShortStory(ctx, "alice", projects[0].ID)
			if e != nil || story.Input.Requirements != "完整故事方向" || story.Input.TargetWordCount != 13000 || story.Input.TargetChapters != 1 {
				t.Fatalf("fallback/default lost %+v %v", story, e)
			}
		})
	}
}
