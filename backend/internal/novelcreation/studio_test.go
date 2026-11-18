package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"sync"
	"testing"
)

type studioTestProvider struct {
	mu       sync.Mutex
	calls    int
	reply    string
	err      error
	started  chan struct{}
	release  chan struct{}
	requests []batchfactoryv11.TextCompletionRequest
}

func (p *studioTestProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	p.requests = append(p.requests, request)
	if p.started != nil && p.calls == 1 {
		close(p.started)
	}
	if p.release != nil {
		<-p.release
	}
	return p.reply, p.err
}

type studioTestPipeline struct {
	PipelineRunner
	provider *studioTestProvider
}

func (p studioTestPipeline) ResolveStudioProvider(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
	return p.provider, nil
}
func studioFixture(t *testing.T) (*Service, *MemoryStore, *studioTestProvider, StudioSession) {
	t.Helper()
	st := NewMemoryStore()
	p := &studioTestProvider{reply: `{"message":"建议如下，请确认","action":{"type":"create_book","book":{"title":"试书","genre":"幻想","world":"山海","protagonist":"小明","coreConflict":"寻找故乡","targetChapterWords":2000}}}`}
	s := NewService(st, NewMemoryRuntime(), studioTestPipeline{provider: p})
	ss, e := s.CreateStudioSession(context.Background(), "alice", StudioCreateSessionInput{IdempotencyKey: "new", Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	return s, st, p, ss
}
func TestStudioNoHiddenProjectAndDurableConfirmation(t *testing.T) {
	s, st, p, ss := studioFixture(t)
	ctx := context.Background()
	in := StudioSendInput{IdempotencyKey: "turn1", Content: "写一个幻想故事", ModelID: "own"}
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, in)
	if e != nil {
		t.Fatal(e)
	}
	if snap.PendingAction == nil || snap.PendingAction.Type != "create_book" {
		t.Fatalf("action missing: %+v", snap)
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 0 {
		t.Fatal("hidden project")
	}
	s.SendStudioMessage(ctx, "alice", ss.ID, in)
	if p.calls != 1 {
		t.Fatal("replayed paid call")
	}
	if _, e = s.GetStudioSnapshot(ctx, "bob", ss.ID); e == nil {
		t.Fatal("cross owner read")
	}
	a := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"}
	out, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, a)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, a)
	if e != nil || again.Project.ID != out.Project.ID {
		t.Fatalf("duplicate: %+v %v", again, e)
	}
	projects, _ = st.ListProjects(ctx, "alice")
	if len(projects) != 1 {
		t.Fatal("duplicate project")
	}
	ex, ok, e := st.GetExecution(ctx, "alice", out.Run.ID)
	if e != nil || !ok || ex.State != "ready" {
		t.Fatalf("missing durable dispatch %+v %v", ex, e)
	}
	if out.ProductionSession.Mode != "long_original" {
		t.Fatal("wrong mode")
	}
	var frozen map[string]any
	if e = json.Unmarshal(out.Run.PromptSnapshot, &frozen); e != nil || frozen["checkpoint"] != true {
		t.Fatal("initial run must pause after blueprint", e)
	}
	messages, e := st.ListMessages(ctx, "alice", out.ProductionSession.ID, 0)
	if e != nil || len(messages) != 1 || !strings.Contains(messages[0].Content, in.Content) {
		t.Fatal("planning discussion not carried into production", e)
	}
}
func TestStudioCancelAndUnknownDoNotCreateOrRebill(t *testing.T) {
	s, st, p, ss := studioFixture(t)
	ctx := context.Background()
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "想法", ModelID: "own"})
	if e != nil {
		t.Fatal(e)
	}
	a := StudioActionInput{ActionID: snap.PendingAction.ID}
	_, e = s.CancelStudioAction(ctx, "alice", ss.ID, a)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ConfirmStudioAction(ctx, "alice", ss.ID, StudioActionInput{ActionID: a.ActionID, ModelID: "own"}); e == nil {
		t.Fatal("cancelled action executed")
	}
	p.err = errors.New("lost connection")
	in := StudioSendInput{IdempotencyKey: "b", Content: "再想想", ModelID: "own"}
	s.SendStudioMessage(ctx, "alice", ss.ID, in)
	s.SendStudioMessage(ctx, "alice", ss.ID, in)
	if p.calls != 2 {
		t.Fatal("unknown call resent")
	}
	snap, e = s.GetStudioSnapshot(ctx, "alice", ss.ID)
	if e != nil || snap.Turn.State != "unknown" {
		t.Fatalf("unknown lost %+v %v", snap, e)
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 0 {
		t.Fatal("hidden project")
	}
}
func TestStudioIdempotencyFingerprintAndInvalidOutput(t *testing.T) {
	s, _, p, ss := studioFixture(t)
	ctx := context.Background()
	p.reply = "not JSON"
	in := StudioSendInput{IdempotencyKey: "same", Content: "hello", ModelID: "own"}
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, in)
	if e != nil {
		t.Fatal(e)
	}
	if snap.Turn.State != "invalid_output" || snap.PendingAction != nil {
		t.Fatal("invalid output accepted")
	}
	in.Content = "different"
	if _, e = s.SendStudioMessage(ctx, "alice", ss.ID, in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal("different content reused key", e)
	}
	if p.calls != 1 {
		t.Fatal("invalid output rebilled")
	}
}
