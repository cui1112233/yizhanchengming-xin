package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestSpinoffCreationValidatesAndFreezesParent(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := chapterImportService(st)
	parent, e := s.ImportChapters(ctx, "alice", chapterImportInput("parent"))
	if e != nil {
		t.Fatal(e)
	}
	in := SpinoffInput{IdempotencyKey: "spin", ParentBookID: parent.Book.ID, Title: "归来之后", ModelID: "own", SpinoffType: "sequel", StoryIdea: "故人的婚礼"}
	out, e := s.StartSpinoff(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if out.Book.ID == parent.Book.ID || out.Session.Mode != "fanfiction" || out.Run.Status != RunQueued {
		t.Fatalf("bad result %+v", out)
	}
	if len(out.Creation.ParentChapters) != 2 || !strings.Contains(out.Creation.ParentChapters[0].Content, "船靠岸") {
		t.Fatal("parent not frozen")
	}
	again, e := s.StartSpinoff(ctx, "alice", in)
	if e != nil || again.Run.ID != out.Run.ID {
		t.Fatal("duplicate created work", e)
	}
	in.Title = "改了输入"
	if _, e = s.StartSpinoff(ctx, "alice", in); e == nil {
		t.Fatal("changed replay accepted")
	}
	in.Title = "归来之后"
	for _, bad := range []SpinoffInput{{IdempotencyKey: "empty", Title: "书", ModelID: "own"}, {IdempotencyKey: "badtype", Title: "书", ParentBookID: parent.Book.ID, SpinoffType: "fake", ModelID: "own"}} {
		if _, e = s.StartSpinoff(ctx, "alice", bad); e == nil {
			t.Fatal("invalid input accepted")
		}
	}
	if _, e = s.StartSpinoff(ctx, "bob", in); e == nil {
		t.Fatal("foreign parent accepted")
	}
	var snap map[string]any
	_ = json.Unmarshal(out.Run.PromptSnapshot, &snap)
	if snap["spinoff_init"] == nil {
		t.Fatal("not an initialization run")
	}
	books, _ := st.ListProjects(ctx, "alice")
	if len(books) != 2 {
		t.Fatalf("failed requests leaked books: %d", len(books))
	}
}

func TestSpinoffInitializesCanonAndFoundationWithoutWritingChapter(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: []string{"# 正传正典\n角色不能离开渡口。", foundationJSON}}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	parent, e := s.ImportChapters(ctx, "alice", chapterImportInput("source"))
	if e != nil {
		t.Fatal(e)
	}
	out, e := s.StartSpinoff(ctx, "alice", SpinoffInput{Title: "渡口番外", ParentBookID: parent.Book.ID, SpinoffType: "if", StoryIdea: "如果船没有来", IdempotencyKey: "initialize", ModelID: "own"})
	if e != nil {
		t.Fatal(e)
	}
	done, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || done.Status != RunSucceeded {
		t.Fatalf("initialization %+v %v", done, e)
	}
	if len(provider.requests) != 2 || !strings.Contains(provider.requests[0].UserPrompt, "船靠岸") || !strings.Contains(provider.requests[1].UserPrompt, "角色不能离开渡口") {
		t.Fatalf("wrong chain %+v", provider.requests)
	}
	chapters, _ := st.ListChapters(ctx, "alice", out.Book.ID)
	if len(chapters) != 0 {
		t.Fatal("initialization wrote chapter")
	}
	creation, e := s.GetSpinoff(ctx, "alice", out.Book.ID)
	if e != nil || creation.ParentCanon == "" || len(creation.Foundation) == 0 {
		t.Fatal("results not saved", e)
	}
	selected, e := s.prepareGenerationInput(ctx, "alice", out.Book.ID, out.Session.ID, AutomationInput{ModeInputs: map[string]string{"existingSetting": "伪造覆盖"}})
	if e != nil || !strings.Contains(selected.ModeInputs["existingSetting"], "角色不能离开渡口") || selected.ModeInputs["spinoffType"] != "if" {
		t.Fatalf("not used in writing %+v %v", selected, e)
	}
	if _, e = s.Run(ctx, "alice", out.Run.ID); e != nil || len(provider.requests) != 2 {
		t.Fatal("duplicate generated again", e)
	}
}

func TestSpinoffQueuedCancelMakesNoModelCalls(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	p := &queuedPipelineProvider{}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return p, nil }))
	parent, _ := s.ImportChapters(ctx, "alice", chapterImportInput("source"))
	out, e := s.StartSpinoff(ctx, "alice", SpinoffInput{Title: "番外", ParentBookID: parent.Book.ID, ModelID: "own", IdempotencyKey: "cancel"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.CancelRun(ctx, "alice", out.Run.ID); e != nil {
		t.Fatal(e)
	}
	done, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || done.Status != RunCancelled || len(p.requests) != 0 {
		t.Fatalf("cancel %+v %v", done, e)
	}
	if _, e = s.prepareGenerationInput(ctx, "alice", out.Book.ID, out.Session.ID, AutomationInput{}); e == nil {
		t.Fatal("unfinished canon accepted")
	}
}

func TestSpinoffUnknownReplyPausesAndBlocksWritingFromIncompleteFoundation(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	p := &queuedPipelineProvider{responses: []string{"# 正传正典\n不可离开渡口。"}, errors: []error{nil, errors.New("provider refused")}}
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return p, nil }))
	parent, _ := s.ImportChapters(ctx, "alice", chapterImportInput("source"))
	out, e := s.StartSpinoff(ctx, "alice", SpinoffInput{Title: "番外", ParentBookID: parent.Book.ID, ModelID: "own", IdempotencyKey: "error"})
	if e != nil {
		t.Fatal(e)
	}
	done, e := s.Run(ctx, "alice", out.Run.ID)
	if !IsCode(e, ErrResultUnknown) || done.Status != RunPaused {
		t.Fatalf("unexpected unknown-reply state %s %v", done.Status, e)
	}
	c, _ := s.GetSpinoff(ctx, "alice", out.Book.ID)
	if (len(c.Foundation) != 0 && string(c.Foundation) != "null") || len(c.ParentChapters) != 2 {
		t.Fatal("failure fabricated result or lost source")
	}
	if _, e = s.prepareGenerationInput(ctx, "alice", out.Book.ID, out.Session.ID, AutomationInput{}); e == nil {
		t.Fatal("failed initialization allowed writing")
	}
}
