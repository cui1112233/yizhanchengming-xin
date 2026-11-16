package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"sync"
	"testing"
	"time"
)

type uploadGateObjects struct {
	*fakeDocumentObjects
	entered, release chan struct{}
	once             sync.Once
}

func (g *uploadGateObjects) unblock() { g.once.Do(func() { close(g.release) }) }
func (g *uploadGateObjects) Put(ctx context.Context, o DocumentObject, r io.Reader) error {
	close(g.entered)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-g.release:
	}
	return g.fakeDocumentObjects.Put(ctx, o, r)
}
func TestActiveUploadCannotLoseArchiveProtection(t *testing.T) {
	for _, action := range []string{"retry", "delete"} {
		t.Run(action, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s, st, p := materialService(t)
			objects := &uploadGateObjects{fakeDocumentObjects: &fakeDocumentObjects{objects: map[string][]byte{}}, entered: make(chan struct{}), release: make(chan struct{})}
			defer objects.unblock()
			s.ConfigureDocuments(objects, nil)
			done := make(chan error, 1)
			go func() { _, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("正文")); done <- e }()
			select {
			case <-objects.entered:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			docs, e := s.ListDocuments(ctx, "alice", p.ID)
			if e != nil || len(docs) != 1 {
				t.Fatalf("jobs=%v err=%v", docs, e)
			}
			id := docs[0].ID
			before, _, _ := st.GetContentRecord(ctx, "alice", p.ID, "document", id)
			if action == "retry" {
				_, e = s.RetryDocument(ctx, "alice", p.ID, id)
			} else {
				e = s.DeleteDocument(ctx, "alice", p.ID, id)
			}
			if e == nil {
				t.Fatalf("%s accepted during live upload", action)
			}
			after, _, _ := st.GetContentRecord(ctx, "alice", p.ID, "document", id)
			if after.Status != "uploading" || !after.LeaseUntil.Equal(before.LeaseUntil) {
				t.Fatal("upload protection changed")
			}
			if e = s.ArchiveProject(ctx, "alice", p.ID); e == nil {
				t.Fatal("active upload archived")
			}
			objects.unblock()
			if e = <-done; e != nil {
				t.Fatal(e)
			}
			if _, e = s.ParseDocument(ctx, "alice", p.ID, id); e != nil {
				t.Fatal(e)
			}
			if e = s.DeleteDocument(ctx, "alice", p.ID, id); e != nil {
				t.Fatalf("completed upload cannot delete: %v", e)
			}
			if e = s.ArchiveProject(ctx, "alice", p.ID); e != nil {
				t.Fatalf("completed/deleted document blocks archive: %v", e)
			}
		})
	}
}

type documentClockStore struct {
	Store
	now time.Time
}

func (s documentClockStore) Clock(context.Context) (time.Time, error) { return s.now, nil }
func (s documentClockStore) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(documentClockStore{tx, s.now}) })
}
func TestAbandonedUploadRetryUsesTransactionalClock(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	s.ConfigureDocuments(&fakeDocumentObjects{objects: map[string][]byte{}}, nil)
	fail := true
	s.store = failDocumentSave{st, &fail}
	j, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("正文"))
	if e == nil {
		t.Fatal("missing save failure")
	}
	rec, _, _ := st.GetContentRecord(ctx, "alice", p.ID, "document", j.ID)
	s.store = documentClockStore{st, rec.LeaseUntil.Add(-time.Nanosecond)}
	if _, e = s.RetryDocument(ctx, "alice", p.ID, j.ID); e == nil {
		t.Fatal("transaction clock still inside lease")
	}
	s.store = documentClockStore{st, rec.LeaseUntil}
	if _, e = s.RetryDocument(ctx, "alice", p.ID, j.ID); e != nil {
		t.Fatalf("expired upload cannot recover: %v", e)
	}
	if j, e = s.ParseDocument(ctx, "alice", p.ID, j.ID); e != nil || j.Status != "ready" {
		t.Fatalf("recover: %s %v", j.Status, e)
	}
}

func styleRetryFixture(t *testing.T) (*Service, *MemoryStore, Project, Session, StyleSource, Run) {
	t.Helper()
	ctx := context.Background()
	s, st, p := materialService(t)
	se, e := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_imitation"})
	if e != nil {
		t.Fatal(e)
	}
	j, e := s.PasteDocument(ctx, "alice", p.ID, "sample", "简短句子。人物说话。")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SelectDocument(ctx, "alice", p.ID, j.ID, DocumentSelection{j.Version, []string{j.Chapters[0].ID}}); e != nil {
		t.Fatal(e)
	}
	style, e := s.CreateStyleFromDocument(ctx, "alice", p.ID, j.ID, "style", "对白")
	if e != nil {
		t.Fatal(e)
	}
	settings, _ := st.GetProjectSettings(ctx, "alice", p.ID)
	settings.Agent.RequireSecondaryConfirm = true
	st.UpdateProjectSettings(ctx, "alice", p.ID, settings)
	run, e := s.QueueMaintenance(ctx, "alice", p.ID, MaintenanceInput{Kind: "style_analysis", SessionID: se.ID, StyleID: style.Profile.ID, IdempotencyKey: "first-style"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.CancelRun(ctx, "alice", run.ID); e != nil {
		t.Fatal(e)
	}
	return s, st, p, se, style, run
}
func storeQueueCounts(st *MemoryStore) [4]int {
	st.mu.RLock()
	defer st.mu.RUnlock()
	msgs, events := 0, 0
	for _, v := range st.messages {
		msgs += len(v)
	}
	for _, v := range st.runEvents {
		events += len(v)
	}
	return [4]int{len(st.runs), len(st.executions), msgs, events}
}
func TestArchivedSessionCannotQueueRetry(t *testing.T) {
	for _, kind := range []string{"ordinary", "style_analysis", "continuity_refresh"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			var s *Service
			var st *MemoryStore
			var p Project
			var se Session
			var run Run
			var e error
			if kind == "style_analysis" {
				s, st, p, se, _, run = styleRetryFixture(t)
			} else if kind == "continuity_refresh" {
				s, st, p, se, _, _ = refreshFixture(t, false)
				run = queueRefresh(t, s, p, se)
				s.CancelRun(ctx, "alice", run.ID)
			} else {
				s, st, p = materialService(t)
				se, _ = s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
				run, e = s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "story", IdempotencyKey: "ordinary"})
				if e != nil {
					t.Fatal(e)
				}
				s.CancelRun(ctx, "alice", run.ID)
			}
			if _, e = s.ManageSession(ctx, "alice", p.ID, se.ID, "archive", ""); e != nil {
				t.Fatal(e)
			}
			before := storeQueueCounts(st)
			if _, e = s.RetryRun(ctx, "alice", run.ID); e == nil {
				t.Fatal("archived session accepted retry")
			}
			if got := storeQueueCounts(st); got != before {
				t.Fatalf("created rows %v -> %v", before, got)
			}
			if _, e = s.ManageSession(ctx, "alice", p.ID, se.ID, "restore", ""); e != nil {
				t.Fatal(e)
			}
			retry, e := s.RetryRun(ctx, "alice", run.ID)
			if e != nil || retry.Status != RunQueued {
				t.Fatalf("restored retry: %+v %v", retry, e)
			}
			if _, e = s.CancelRun(ctx, "alice", retry.ID); e != nil {
				t.Fatal(e)
			}
			if e = s.ArchiveProject(ctx, "alice", p.ID); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestStyleRetryTracksNewAnalysisRun(t *testing.T) {
	ctx := context.Background()
	s, st, p, _, style, old := styleRetryFixture(t)
	before, _ := s.GetStyleSource(ctx, "alice", p.ID, style.Profile.ID)
	retry, e := s.RetryRun(ctx, "alice", old.ID)
	if e != nil {
		t.Fatal(e)
	}
	assertStyle := func(want RunStatus) {
		t.Helper()
		got, e := s.GetStyleSource(ctx, "alice", p.ID, style.Profile.ID)
		if e != nil || got.AnalysisRunID != retry.ID || got.Status != string(want) {
			t.Fatalf("style=%+v retry=%s err=%v", got, retry.ID, e)
		}
		if fingerprint(got.Source) != fingerprint(before.Source) {
			t.Fatal("frozen source changed")
		}
	}
	assertStyle(RunQueued)
	again, e := s.RetryRun(ctx, "alice", old.ID)
	if e != nil || again.ID != retry.ID {
		t.Fatalf("idempotence: %s %v", again.ID, e)
	}
	// An actual local fake provider blocks after the running event, then returns
	// invalid JSON so the existing repair gate creates a secondary confirmation.
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	provider := &styleRetryGateProvider{entered: entered, release: release}
	s.pipeline = pipelineWithProvider(provider)
	done := make(chan error, 1)
	go func() { _, e := s.Run(ctx, "alice", retry.ID); done <- e }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("model not reached")
	}
	assertStyle(RunAnalyzing)
	once.Do(func() { close(release) })
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	assertStyle(RunAwaitingConfirmation)
	ex, _, _ := st.GetExecution(ctx, "alice", retry.ID)
	if ex.PendingConfirmation == nil {
		t.Fatal("missing secondary checkpoint")
	}
	confirmed, e := s.ConfirmModelCall(ctx, "alice", retry.ID, ex.PendingConfirmation.ID)
	if e != nil || confirmed.ID != retry.ID {
		t.Fatal(e)
	}
	if _, e = s.Run(ctx, "alice", retry.ID); e != nil {
		t.Fatal(e)
	}
	assertStyle(RunSucceeded)
	if provider.calls != 2 {
		t.Fatalf("fake calls=%d", provider.calls)
	}
}
func TestStyleRetryDoesNotOverwriteNewerPointerOrChangedSource(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "newer_pointer", true: "changed_source"}[changed], func(t *testing.T) {
			ctx := context.Background()
			s, st, p, se, style, old := styleRetryFixture(t)
			newer, e := s.QueueMaintenance(ctx, "alice", p.ID, MaintenanceInput{Kind: "style_analysis", SessionID: se.ID, StyleID: style.Profile.ID, IdempotencyKey: "newer"})
			if e != nil {
				t.Fatal(e)
			}
			s.CancelRun(ctx, "alice", newer.ID)
			if changed {
				j, _ := s.GetDocument(ctx, "alice", p.ID, strings.TrimPrefix(style.Source.ID, "docref-"))
				if _, e = s.ResegmentDocument(ctx, "alice", p.ID, j.ID, DocumentSegments{j.Version, []MaterialChapter{{Title: "new", Content: "新来源正文"}}}); e != nil {
					t.Fatal(e)
				}
			}
			counts := storeQueueCounts(st)
			retry, e := s.RetryRun(ctx, "alice", old.ID)
			if changed {
				if e == nil {
					t.Fatal("changed source retry accepted")
				}
				if storeQueueCounts(st) != counts {
					t.Fatal("changed source left queued rows")
				}
			} else if e != nil || retry.ID == "" {
				t.Fatal(e)
			}
			current, e := s.GetStyleSource(ctx, "alice", p.ID, style.Profile.ID)
			if e != nil || current.AnalysisRunID != newer.ID {
				t.Fatalf("newer pointer overwritten: %+v %v", current, e)
			}
			var frozen maintenanceSnapshot
			json.Unmarshal(old.PromptSnapshot, &frozen)
			if fingerprint(current.Source) != fingerprint(frozen.Style.Source) {
				t.Fatal("source snapshot overwritten")
			}
		})
	}
}

type styleRetryGateProvider struct {
	entered, release chan struct{}
	calls            int
}

func (p *styleRetryGateProvider) Complete(ctx context.Context, _ batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	if p.calls == 1 {
		close(p.entered)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-p.release:
		}
		return "not json", nil
	}
	return `{"content":"简洁文风"}`, nil
}

func TestArchivedSessionRejectsSendAndBlueprintConfirm(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, e := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "original"})
	if e != nil {
		t.Fatal(e)
	}
	run, e := st.CreateRun(ctx, "alice", Run{ProjectID: p.ID, SessionID: se.ID, Status: RunAwaitingConfirmation, IdempotencyKey: "blueprint", PromptSnapshot: json.RawMessage(`{"chapter_number":1,"checkpoint":true}`)})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ManageSession(ctx, "alice", p.ID, se.ID, "archive", ""); e != nil {
		t.Fatal(e)
	}
	before := storeQueueCounts(st)
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "new", IdempotencyKey: "send"}); e == nil {
		t.Fatal("archived send accepted")
	}
	if _, e = s.ConfirmRun(ctx, "alice", run.ID, json.RawMessage(foundationJSON)); e == nil {
		t.Fatal("archived blueprint confirmed")
	}
	if storeQueueCounts(st) != before {
		t.Fatal("archived operation created dispatch/message/event")
	}
	if _, found, e := st.GetFoundation(ctx, "alice", p.ID); e != nil || found {
		t.Fatal("archived blueprint persisted")
	}
	if _, e = s.ManageSession(ctx, "alice", p.ID, se.ID, "restore", ""); e != nil {
		t.Fatal(e)
	}
	if storeQueueCounts(st) != before {
		t.Fatal("restore implicitly dispatched work")
	}
	next, e := s.ConfirmRun(ctx, "alice", run.ID, json.RawMessage(foundationJSON))
	if e != nil || next.Status != RunQueued {
		t.Fatalf("explicit confirm after restore: %s %v", next.Status, e)
	}
	s.CancelRun(ctx, "alice", next.ID)
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "new", IdempotencyKey: "send"}); e != nil {
		t.Fatal(e)
	}
}
func TestArchivedSessionSecondaryConfirmationStaysReadOnly(t *testing.T) {
	ctx := context.Background()
	f, provider, run := confirmationGenerationFixture(t)
	oldToken := pendingGenerationID(t, f, run.ID)
	if _, e := f.service.ConfirmModelCall(ctx, "alice", run.ID, oldToken); e != nil {
		t.Fatal(e)
	}
	if _, e := f.service.Run(ctx, "alice", run.ID); e != nil {
		t.Fatal(e)
	}
	pending := pendingGenerationID(t, f, run.ID)
	if _, e := f.service.ManageSession(ctx, "alice", f.project.ID, f.session.ID, "archive", ""); e != nil {
		t.Fatal(e)
	}
	before, _, _ := f.store.GetExecution(ctx, "alice", run.ID)
	counts := storeQueueCounts(f.store)
	calls := len(provider.requests)
	if _, e := f.service.ConfirmModelCall(ctx, "alice", run.ID, pending); e == nil {
		t.Fatal("archived secondary approval redispatched run")
	}
	historical, e := f.service.ConfirmModelCall(ctx, "alice", run.ID, oldToken)
	if e != nil || historical.Status != RunAwaitingConfirmation {
		t.Fatalf("history lookup: %s %v", historical.Status, e)
	}
	after, _, _ := f.store.GetExecution(ctx, "alice", run.ID)
	if fingerprint(after) != fingerprint(before) || storeQueueCounts(f.store) != counts || len(provider.requests) != calls {
		t.Fatal("archived confirmation changed checkpoint or called model")
	}
	if _, e = f.service.ManageSession(ctx, "alice", f.project.ID, f.session.ID, "restore", ""); e != nil {
		t.Fatal(e)
	}
	after, _, _ = f.store.GetExecution(ctx, "alice", run.ID)
	if fingerprint(after) != fingerprint(before) {
		t.Fatal("restore implicitly dispatched work")
	}
	resumed, e := f.service.ConfirmModelCall(ctx, "alice", run.ID, pending)
	if e != nil || resumed.ID != run.ID || resumed.Status != RunQueued {
		t.Fatalf("explicit resume: %s %v", resumed.Status, e)
	}
	if len(provider.requests) != calls {
		t.Fatal("confirmation itself called provider")
	}
}

type stylePointerFaultStore struct{ Store }

func (s stylePointerFaultStore) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(stylePointerFaultStore{tx}) })
}
func (s stylePointerFaultStore) PutContentRecord(ctx context.Context, r ContentRecord) error {
	if r.Kind == "style" && r.Status == "queued" {
		return errors.New("fake style pointer write failure")
	}
	return s.Store.PutContentRecord(ctx, r)
}
func TestStyleRetryPointerFailureRollsBackQueue(t *testing.T) {
	ctx := context.Background()
	s, st, p, _, style, source := styleRetryFixture(t)
	counts := storeQueueCounts(st)
	s.store = stylePointerFaultStore{st}
	if _, e := s.RetryRun(ctx, "alice", source.ID); e == nil {
		t.Fatal("missing style pointer failure")
	}
	if storeQueueCounts(st) != counts {
		t.Fatal("pointer failure left new run/outbox/messages")
	}
	current, e := s.GetStyleSource(ctx, "alice", p.ID, style.Profile.ID)
	if e != nil || current.AnalysisRunID != source.ID {
		t.Fatal("failed transaction changed pointer")
	}
	s.store = st
	if _, e = s.RetryRun(ctx, "alice", source.ID); e != nil {
		t.Fatal(e)
	}
}

type beforeProjectTransactionStore struct {
	Store
	before func()
	once   *sync.Once
}

func (s beforeProjectTransactionStore) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	s.once.Do(s.before)
	return s.Store.Atomic(ctx, o, p, fn)
}
func TestRetryRechecksSessionInsideProjectTransaction(t *testing.T) {
	ctx := context.Background()
	s, st, p, se, _, run := styleRetryFixture(t)
	before := storeQueueCounts(st)
	var once sync.Once
	s.store = beforeProjectTransactionStore{Store: st, once: &once, before: func() {
		archiver := NewService(st, NewMemoryRuntime(), nil)
		if _, e := archiver.ManageSession(ctx, "alice", p.ID, se.ID, "archive", ""); e != nil {
			t.Fatal(e)
		}
	}}
	if _, e := s.RetryRun(ctx, "alice", run.ID); e == nil {
		t.Fatal("retry did not recheck archived session under project lock")
	}
	if storeQueueCounts(st) != before {
		t.Fatal("transaction-race left new rows")
	}
}
func TestRetryRejectsSessionInDifferentProject(t *testing.T) {
	ctx := context.Background()
	s, st, _, se, _, run := styleRetryFixture(t)
	other, e := s.CreateProject(ctx, "alice", CreateProjectInput{Title: "Other", Genre: "幻想", ModelID: "fake"})
	if e != nil {
		t.Fatal(e)
	}
	// Simulate a migrated/corrupt source's session link; it must not create work.
	st.mu.Lock()
	session := st.sessions[memoryKey("alice", se.ID)]
	session.ProjectID = other.ID
	st.sessions[memoryKey("alice", se.ID)] = session
	st.mu.Unlock()
	before := storeQueueCounts(st)
	if _, e = s.RetryRun(ctx, "alice", run.ID); e == nil {
		t.Fatal("cross-project session retry accepted")
	}
	if storeQueueCounts(st) != before {
		t.Fatal("cross-project retry wrote rows")
	}
}
