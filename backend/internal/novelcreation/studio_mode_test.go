package novelcreation

import (
	"context"
	"testing"
)

func TestShortOriginalStartsAsDiscussionBeforeIndependentStory(t *testing.T) {
	ctx := context.Background()
	s, st, p, _ := studioFixture(t)
	ss, e := s.CreateStudioSession(ctx, "alice", StudioCreateSessionInput{IdempotencyKey: "short-discussion", Mode: "short_original"})
	if e != nil || ss.Mode != "short_original" || ss.Title != "短篇原创讨论" {
		t.Fatalf("short discussion entry missing: %+v %v", ss, e)
	}
	p.reply = `{"message":"先讨论结尾和主人公","action":null}`
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "discuss", Content: "想写一篇悬疑短篇", ModelID: "own"})
	if e != nil || snap.Turn.State != "completed" || snap.PendingAction != nil {
		t.Fatalf("discussion generated unexpectedly %+v %v", snap, e)
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 0 {
		t.Fatal("discussion created hidden work")
	}
	p.reply = `{"message":"请确认短篇","action":{"type":"short_run","title":"旧信","direction":"寻找故乡的悬疑故事","targetWordCount":13000,"targetChapters":1}}`
	snap, e = s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "propose", Content: "确认这个方向，准备创作", ModelID: "own"})
	if e != nil || snap.PendingAction == nil {
		t.Fatalf("no short proposal %+v %v", snap, e)
	}
	result, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"})
	if e != nil || result.ProductionSession.Mode != "short_original" || result.Story == nil {
		t.Fatalf("wrong output %+v %v", result, e)
	}
}

func TestShortDiscussionCannotConfirmLongBookProposal(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := studioFixture(t)
	ss, e := s.CreateStudioSession(ctx, "alice", StudioCreateSessionInput{IdempotencyKey: "short", Mode: "short_original"})
	if e != nil {
		t.Fatal(e)
	}
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "wrong-mode", Content: "讨论短篇", ModelID: "own"})
	if e != nil || snap.Turn.State != "invalid_output" || snap.PendingAction != nil {
		t.Fatalf("short discussion accepted long book action %+v %v", snap, e)
	}
}
