package novelcreation

import (
	"context"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func refreshFixture(t *testing.T, secondary bool) (*Service, *MemoryStore, Project, Session, Chapter, *queuedPipelineProvider) {
	t.Helper()
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	r, _ := st.CreateRun(ctx, "alice", Run{ProjectID: p.ID, SessionID: se.ID, IdempotencyKey: "seed", Status: RunQueued})
	b := serviceTestBundle(RunContext{Run: r, Project: p, Session: se, ChapterNumber: 1}, RunSucceeded)
	if e := st.SaveChapterBundle(ctx, "alice", b); e != nil {
		t.Fatal(e)
	}
	ch, _ := st.ListChapters(ctx, "alice", p.ID)
	edited, e := s.EditChapter(ctx, "alice", p.ID, ch[0].ID, ChapterEdit{ExpectedVersion: ch[0].Version, Content: "新正文事实"})
	if e != nil {
		t.Fatal(e)
	}
	settings, _ := st.GetProjectSettings(ctx, "alice", p.ID)
	settings.Agent.RequireSecondaryConfirm = secondary
	st.UpdateProjectSettings(ctx, "alice", p.ID, settings)
	provider := &queuedPipelineProvider{responses: []string{settlementJSON, memoryJSON}}
	s.pipeline = pipelineWithProvider(provider)
	return s, st, p, se, edited, provider
}
func queueRefresh(t *testing.T, s *Service, p Project, se Session) Run {
	t.Helper()
	r, e := s.QueueMaintenance(context.Background(), "alice", p.ID, MaintenanceInput{Kind: "continuity_refresh", SessionID: se.ID, IdempotencyKey: "refresh"})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestMaintenanceCompletedRawRestartAndSecondarySameRun(t *testing.T) {
	ctx := context.Background()
	s, st, p, se, _, provider := refreshFixture(t, true)
	s.store = faultJournalStore{st, &onceWriteFault{kind: "event"}}
	r := queueRefresh(t, s, p, se)
	if _, e := s.Run(ctx, "alice", r.ID); e == nil {
		t.Fatal("expected local event failure")
	}
	if len(provider.requests) != 1 {
		t.Fatal("first call count")
	}
	restarted := NewService(st, NewMemoryRuntime(), s.pipeline)
	pending, e := restarted.Run(ctx, "alice", r.ID)
	if e != nil || pending.Status != RunAwaitingConfirmation || len(provider.requests) != 1 {
		t.Fatalf("pending=%s calls=%d err=%v", pending.Status, len(provider.requests), e)
	}
	ex, _, _ := st.GetExecution(ctx, "alice", r.ID)
	if ex.PendingConfirmation == nil {
		t.Fatal("missing token")
	}
	confirmed, e := restarted.ConfirmModelCall(ctx, "alice", r.ID, ex.PendingConfirmation.ID)
	if e != nil || confirmed.ID != r.ID {
		t.Fatalf("confirm %v", e)
	}
	done, e := restarted.Run(ctx, "alice", r.ID)
	if e != nil || done.Status != RunSucceeded || len(provider.requests) != 2 {
		t.Fatalf("done=%s calls=%d err=%v", done.Status, len(provider.requests), e)
	}
}

type maintenanceSaveFault struct {
	Store
	fail *bool
}

func (s maintenanceSaveFault) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(maintenanceSaveFault{tx, s.fail}) })
}
func (s maintenanceSaveFault) PutContentRecord(ctx context.Context, r ContentRecord) error {
	if r.Kind == "continuity" && r.Status == "fresh" && *s.fail {
		*s.fail = false
		return errors.New("save unavailable")
	}
	return s.Store.PutContentRecord(ctx, r)
}
func TestMaintenanceSavePendingAndSourceConflict(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "save_retry", true: "source_conflict"}[changed], func(t *testing.T) {
			ctx := context.Background()
			s, st, p, se, ch, provider := refreshFixture(t, false)
			fail := true
			s.store = maintenanceSaveFault{st, &fail}
			r := queueRefresh(t, s, p, se)
			if _, e := s.Run(ctx, "alice", r.ID); e == nil {
				t.Fatal("expected save fault")
			}
			ex, _, _ := st.GetExecution(ctx, "alice", r.ID)
			if ex.State != "save_pending" || ex.Maintenance == nil || len(provider.requests) != 2 {
				t.Fatalf("state=%s calls=%d", ex.State, len(provider.requests))
			}
			next := NewService(st, NewMemoryRuntime(), s.pipeline)
			if e := next.ArchiveProject(ctx, "alice", p.ID); e == nil {
				t.Fatal("save_pending writer archived")
			}
			if changed {
				if _, e := next.EditChapter(ctx, "alice", p.ID, ch.ID, ChapterEdit{ExpectedVersion: ch.Version, Content: "再改版"}); e != nil {
					t.Fatal(e)
				}
			}
			done, e := next.Run(ctx, "alice", r.ID)
			if changed {
				if e == nil || done.Status != RunFailed {
					t.Fatalf("source conflict=%s %v", done.Status, e)
				}
				state, _ := next.GetContinuityState(ctx, "alice", p.ID)
				if state.DirtyFrom != 1 {
					t.Fatal("stale result became fresh")
				}
			} else if e != nil || done.Status != RunSucceeded {
				t.Fatal(e)
			}
			if len(provider.requests) != 2 {
				t.Fatal("paid result regenerated")
			}
		})
	}
}
func TestMaintenanceQueuedSourceChangedBeforePaidCall(t *testing.T) {
	ctx := context.Background()
	s, _, p, se, ch, provider := refreshFixture(t, false)
	r := queueRefresh(t, s, p, se)
	s.EditChapter(ctx, "alice", p.ID, ch.ID, ChapterEdit{ExpectedVersion: ch.Version, Content: "queued后改版"})
	done, e := s.Run(ctx, "alice", r.ID)
	if e == nil || done.Status != RunFailed || len(provider.requests) != 0 {
		t.Fatalf("status=%s calls=%d err=%v", done.Status, len(provider.requests), e)
	}
}

func TestMaintenanceUncertainPaidResponsePausesWithoutRetry(t *testing.T) {
	ctx := context.Background()
	s, st, p, se, _, provider := refreshFixture(t, false)
	s.store = faultJournalStore{st, &onceWriteFault{kind: "completed"}}
	r := queueRefresh(t, s, p, se)
	s.Run(ctx, "alice", r.ID)
	saved, _ := st.GetRun(ctx, "alice", r.ID)
	if saved.Status != RunPaused || saved.ErrorCode != ErrResultUnknown || len(provider.requests) != 1 {
		t.Fatalf("status=%s code=%s fakecalls=%d", saved.Status, saved.ErrorCode, len(provider.requests))
	}
	if _, e := s.RetryRun(ctx, "alice", r.ID); !IsCode(e, ErrResultUnknown) {
		t.Fatal("uncertain request allowed retry")
	}
	s.store = st
	if e := s.ArchiveProject(ctx, "alice", p.ID); e != nil {
		t.Fatalf("paused unknown blocked archive: %v", e)
	}
}
func TestStyleStatisticsAndExplicitAnalysisFakeOnly(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_imitation"})
	doc, e := s.PasteDocument(ctx, "alice", p.ID, "样本", "简短句子。\n人物说：“你好。”")
	if e != nil {
		t.Fatal(e)
	}
	s.SelectDocument(ctx, "alice", p.ID, doc.ID, DocumentSelection{doc.Version, []string{doc.Chapters[0].ID}})
	calls := 0
	provider := &queuedPipelineProvider{responses: []string{`{"content":"节奏明快，以简短对白推进情节。"}`}}
	s.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		calls++
		return provider, nil
	})
	style, e := s.CreateStyleFromDocument(ctx, "alice", p.ID, doc.ID, "简洁", "保留对白")
	if e != nil || calls != 0 || style.Statistics.Characters == 0 {
		t.Fatal(e)
	}
	r, e := s.QueueMaintenance(ctx, "alice", p.ID, MaintenanceInput{Kind: "style_analysis", SessionID: se.ID, StyleID: style.Profile.ID, IdempotencyKey: "style"})
	if e != nil {
		t.Fatal(e)
	}
	done, e := s.Run(ctx, "alice", r.ID)
	if e != nil || done.Status != RunSucceeded || calls != 1 {
		t.Fatalf("%s %d %v", done.Status, calls, e)
	}
	profile, _ := st.GetStyleProfile(ctx, "alice", style.Profile.ID)
	if !strings.Contains(profile.Content, "简短对白") {
		t.Fatal("analysis not saved")
	}
	saved, _ := st.HasSavedRun(ctx, "alice", r.ID)
	if saved {
		t.Fatal("style counted chapter")
	}
}
