package novelcreation

import (
	"context"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestContinuityRefreshDurableNoProseAndFakeCalls(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	r, _ := st.CreateRun(ctx, "alice", Run{ProjectID: p.ID, SessionID: se.ID, IdempotencyKey: "seed", Status: RunQueued})
	b := serviceTestBundle(RunContext{Run: r, Project: p, Session: se, ChapterNumber: 1}, RunSucceeded)
	if e := st.SaveChapterBundle(ctx, "alice", b); e != nil {
		t.Fatal(e)
	}
	ch, _ := st.ListChapters(ctx, "alice", p.ID)
	edited, e := s.EditChapter(ctx, "alice", p.ID, ch[0].ID, ChapterEdit{ExpectedVersion: ch[0].Version, Content: "改写后的新正文"})
	if e != nil {
		t.Fatal(e)
	}
	provider := &queuedPipelineProvider{responses: []string{settlementJSON, memoryJSON}}
	s.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil })
	run, e := s.QueueMaintenance(ctx, "alice", p.ID, MaintenanceInput{Kind: "continuity_refresh", SessionID: se.ID, IdempotencyKey: "refresh"})
	if e != nil {
		t.Fatal(e)
	}
	done, e := s.Run(ctx, "alice", run.ID)
	if e != nil || done.Status != RunSucceeded {
		t.Fatalf("%+v %v", done, e)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("fake calls=%d", len(provider.requests))
	}
	if !strings.Contains(provider.requests[0].UserPrompt, "改写后的新正文") {
		t.Fatal("old source used")
	}
	current, _ := st.ListChapters(ctx, "alice", p.ID)
	if len(current) != 1 || current[0].Version != edited.Version || current[0].Content != edited.Content {
		t.Fatal("refresh regenerated prose")
	}
	if _, e = s.loadContinuity(ctx, "alice", p.ID, current); e != nil {
		t.Fatal(e)
	}
	saved, e := st.HasSavedRun(ctx, "alice", run.ID)
	if e != nil || saved {
		t.Fatalf("maintenance counted chapter: %v %v", saved, e)
	}
	s.Run(ctx, "alice", run.ID)
	if len(provider.requests) != 2 {
		t.Fatal("repeated model call")
	}
}
