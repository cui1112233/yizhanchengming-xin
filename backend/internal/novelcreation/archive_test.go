package novelcreation

import (
	"context"
	"testing"
)

func TestArchiveGuardsMutationsClaimsAndRestoresPaused(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	run, e := s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "写故事", IdempotencyKey: "work"})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.ArchiveProject(ctx, "alice", p.ID); e == nil {
		t.Fatal("queued writer archived")
	}
	s.CancelRun(ctx, "alice", run.ID)
	if e = s.ArchiveProject(ctx, "alice", p.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.CreateNote(ctx, "alice", p.ID, ProjectNoteInput{Content: "late"}); e == nil {
		t.Fatal("archived note write")
	}
	if _, e = s.RenameProject(ctx, "alice", p.ID, "late"); e == nil {
		t.Fatal("archived rename")
	}
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "late", IdempotencyKey: "new"}); e == nil {
		t.Fatal("archived queue")
	}
	if _, _, e = s.claim(ctx, "alice", run.ID); e == nil {
		t.Fatal("archived claim")
	}
	if e = s.RestoreProject(ctx, "alice", p.ID); e != nil {
		t.Fatal(e)
	}
	restored, _ := st.GetProject(ctx, "alice", p.ID)
	if restored.Status == "archived" || restored.UpdatedAt.Before(p.UpdatedAt) {
		t.Fatal("clock not reset")
	}
}

func TestSessionArchiveScopeAndScheduleRemainsPausedAfterRestore(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	plan, e := s.CreateSchedule(ctx, "alice", p.ID, ScheduleInput{SessionID: se.ID, IntervalMinutes: 60, Enabled: true, Prompt: "写作", ModelID: "fake"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ManageSession(ctx, "bob", p.ID, se.ID, "archive", ""); e == nil {
		t.Fatal("foreign session")
	}
	if _, e = s.ManageSession(ctx, "alice", p.ID, se.ID, "rename", "新名称"); e != nil {
		t.Fatal(e)
	}
	archived, e := s.ManageSession(ctx, "alice", p.ID, se.ID, "archive", "")
	if e != nil || archived.ArchivedAt == nil {
		t.Fatal(e)
	}
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, Content: "late", ModelID: "fake", IdempotencyKey: "late"}); e == nil {
		t.Fatal("archived session accepted input")
	}
	if _, e = s.ManageSession(ctx, "alice", p.ID, se.ID, "restore", ""); e != nil {
		t.Fatal(e)
	}
	saved, e := st.GetSchedule(ctx, "alice", plan.ID)
	if e != nil || saved.Enabled || saved.AutomationState.Status != "paused" {
		t.Fatalf("%+v %v", saved, e)
	}
}

func TestPurgingGuardsMutationsRestoreAndClaims(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	run, e := s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "body", IdempotencyKey: "beforepurge"})
	if e != nil {
		t.Fatal(e)
	}
	p.Status = "purging"
	st.mu.Lock()
	st.projects[memoryKey("alice", p.ID)] = p
	st.mu.Unlock()
	if _, e = s.CreateNote(ctx, "alice", p.ID, ProjectNoteInput{Content: "late"}); e == nil {
		t.Fatal("purging mutation accepted")
	}
	if e = s.RestoreProject(ctx, "alice", p.ID); e == nil {
		t.Fatal("purging restored")
	}
	if _, _, e = s.claim(ctx, "alice", run.ID); e == nil {
		t.Fatal("purging claimed")
	}
}
