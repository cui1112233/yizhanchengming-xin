package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestStudioContextBudgetBoundsActualRequestKeepsDurableHistory(t *testing.T) {
	s, st, p, ss := studioFixture(t)
	ctx := context.Background()
	p.reply = `{"message":"旧回复内容很长","action":null}`
	if _, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "old", Content: "旧消息内容很长", ModelID: "own"}); e != nil {
		t.Fatal(e)
	}
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "new", Content: "新😀", ModelID: "own", ContextBudget: 2})
	if e != nil {
		t.Fatal(e)
	}
	if len(snap.Messages) != 4 {
		t.Fatal("durable history was cut", snap.Messages)
	}
	var payload struct {
		Messages      []StudioMessage `json:"messages"`
		ContextBudget int             `json:"contextBudget"`
	}
	if e = json.Unmarshal([]byte(p.requests[1].UserPrompt), &payload); e != nil {
		t.Fatal(e)
	}
	if len(payload.Messages) != 1 || payload.Messages[0].Content != "新😀" || payload.ContextBudget != 2 {
		t.Fatalf("budget ignored in real supplied request: %+v", payload)
	}
	raw, e := st.LoadStudio(ctx, "alice", ss.ID, false)
	if e != nil {
		t.Fatal(e)
	}
	var saved studioRecord
	if e = json.Unmarshal(raw, &saved); e != nil {
		t.Fatal(e)
	}
	if saved.Turns[1].Request.UserPrompt != p.requests[1].UserPrompt {
		t.Fatal("actual bounded request was not frozen")
	}
	_, e = s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "too-long", Content: "三个字", ModelID: "own", ContextBudget: 2})
	if !IsCode(e, ErrInputRequired) || p.calls != 2 {
		t.Fatal("oversize latest message reached provider", e)
	}
	snap, e = s.GetStudioSnapshot(ctx, "alice", ss.ID)
	if e != nil || len(snap.Messages) != 4 {
		t.Fatal("rejected message modified durable history", e)
	}
}

func TestStudioProductionHandoffUsesFrozenBoundedDiscussion(t *testing.T) {
	s, st, _, ss := studioFixture(t)
	ctx := context.Background()
	excluded := "旧的长讨论不应回注" + strings.Repeat("旧背景", 200)
	if _, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "old", Content: excluded, ModelID: "own"}); e != nil {
		t.Fatal(e)
	}
	latest := "只改主角为小红"
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "latest", Content: latest, ModelID: "own", ContextBudget: 100})
	if e != nil {
		t.Fatal(e)
	}
	out, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"})
	if e != nil {
		t.Fatal(e)
	}
	var frozen map[string]any
	if e = json.Unmarshal(out.Run.PromptSnapshot, &frozen); e != nil {
		t.Fatal(e)
	}
	content, ok := frozen["user_message"].(string)
	if !ok {
		t.Fatal("missing queued request")
	}
	if strings.Contains(content, excluded) || !strings.Contains(content, latest) || !strings.Contains(content, ss.ID) || !strings.Contains(content, "confirmedBook") {
		t.Fatal("handoff lost bounded discussion or reinjected old history")
	}
	if len(content) > 3000 {
		t.Fatal("handoff unexpectedly unbounded", len(content))
	}
	saved, e := st.ListMessages(ctx, "alice", out.ProductionSession.ID, 0)
	if e != nil || len(saved) != 1 || saved[0].Content != content {
		t.Fatal("queued handoff differs from durable production message", e)
	}
}
func TestStudioContextBudgetKeepsWholeNewestPairs(t *testing.T) {
	messages := []StudioMessage{{Role: "user", Content: "旧问题"}, {Role: "assistant", Content: "旧答复"}, {Role: "user", Content: "近问题"}, {Role: "assistant", Content: "近答复"}, {Role: "user", Content: "新😀"}}
	got, e := boundedStudioMessages(messages, 8)
	if e != nil || len(got) != 3 || got[0].Content != "近问题" || got[2].Content != "新😀" {
		t.Fatalf("whole pair selection: %+v %v", got, e)
	}
	got, e = boundedStudioMessages(messages, 7)
	if e != nil || len(got) != 1 || got[0].Content != "新😀" {
		t.Fatalf("must not include half pair: %+v %v", got, e)
	}
}

func TestStudioRefinementFreezesPriorStructuredProposalAndRejectsInsufficientBudget(t *testing.T) {
	s, _, p, ss := studioFixture(t)
	ctx := context.Background()
	first, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "first", Content: "创作幻想故事", ModelID: "own"})
	if e != nil {
		t.Fatal(e)
	}
	oldID := first.PendingAction.ID
	_, e = s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "too-small", Content: "改主角", ModelID: "own", ContextBudget: 3})
	if !IsCode(e, ErrInputRequired) || p.calls != 1 {
		t.Fatal("proposal overflow reached provider", e)
	}
	unchanged, e := s.GetStudioSnapshot(ctx, "alice", ss.ID)
	if e != nil || unchanged.PendingAction.ID != oldID || unchanged.PendingAction.Status != "pending" || len(unchanged.Messages) != 2 {
		t.Fatal("budget rejection superseded proposal", e)
	}
	refined, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "refine", Content: "只把主角改为小红，其他不变", ModelID: "own", ContextBudget: 64000})
	if e != nil {
		t.Fatal(e)
	}
	var payload struct {
		Messages         []StudioMessage `json:"messages"`
		PreviousProposal *StudioAction   `json:"previousProposal"`
	}
	if e = json.Unmarshal([]byte(p.requests[1].UserPrompt), &payload); e != nil {
		t.Fatal(e)
	}
	if payload.PreviousProposal == nil || payload.PreviousProposal.Book.Title != first.PendingAction.Book.Title || payload.PreviousProposal.Book.World != "山海" || payload.PreviousProposal.Book.Protagonist != "小明" || payload.PreviousProposal.Status != "pending" || payload.Messages[len(payload.Messages)-1].Content != "只把主角改为小红，其他不变" {
		t.Fatalf("refinement context incomplete: %+v", payload)
	}
	if refined.PendingAction.ID == oldID {
		t.Fatal("proposal identity not replaced")
	}
	if _, e = s.ConfirmStudioAction(ctx, "alice", ss.ID, StudioActionInput{ActionID: oldID, ModelID: "own"}); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal("old proposal remained executable", e)
	}
	if p.calls != 2 {
		t.Fatal("refinement made extra provider calls")
	}
}
