package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

const runLeaseTTL = 90 * time.Second
const runLeaseRenew = 30 * time.Second

type workerControl struct {
	mu     sync.Mutex
	cancel map[string]activeExecution
	wake   chan struct{}
}
type activeExecution struct {
	fence  int64
	cancel context.CancelFunc
}

func newWorkerControl() *workerControl {
	return &workerControl{cancel: map[string]activeExecution{}, wake: make(chan struct{}, 1)}
}
func (s *Service) Notify(_ context.Context, _, _ string) {
	select {
	case s.control.wake <- struct{}{}:
	default:
	}
}

// StartWorker scans MySQL even if every Redis notification is lost. Every
// process may run a scanner: the project transaction decides the sole owner.
func StartWorker(ctx context.Context, s *Service, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			entries, err := s.store.ListDispatch(ctx, 64)
			if err == nil {
				for _, e := range entries {
					e := e
					go func() {
						runCtx, cancel := context.WithTimeout(ctx, 45*time.Minute)
						defer cancel()
						_, _ = s.Run(runCtx, e.Owner, e.RunID)
					}()
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			case <-s.control.wake:
			}
		}
	}()
}

func (s *Service) claim(ctx context.Context, owner, id string) (RunLease, Execution, error) {
	run, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return RunLease{}, Execution{}, err
	}
	var lease RunLease
	var execution Execution
	var paused bool
	err = s.store.Atomic(ctx, owner, run.ProjectID, func(tx Store) error {
		current, e := tx.GetRun(ctx, owner, id)
		if e != nil {
			return e
		}
		if isTerminalRunStatus(current.Status) || current.Status == RunAwaitingConfirmation || current.Status == RunPaused {
			return &DomainError{Code: ErrInvalidTransition, Message: "任务当前不可领取"}
		}
		execution, _, e = tx.GetExecution(ctx, owner, id)
		if e != nil {
			return e
		}
		// No journal means an old process could already have sent a paid request.
		// Only a legacy queued row is known not to have started.
		if execution.RunID == "" {
			execution = Execution{Owner: owner, RunID: id, ProjectID: run.ProjectID, State: "ready", Calls: map[string]CallRecord{}}
			if current.Status != RunQueued {
				paused = true
			}
		}
		if execution.CancelRequested {
			return leaseLost()
		}
		if execution.State == "paused" {
			return unknownResult()
		}
		slot, e := tx.GetSlot(ctx, owner, run.ProjectID)
		if e != nil {
			return e
		}
		now, clockErr := tx.Clock(ctx)
		if clockErr != nil {
			return clockErr
		}
		if slot.LeaseUntil.After(now) {
			return &DomainError{Code: ErrProjectLocked, Message: "作品已有正在执行的任务"}
		}
		for _, call := range execution.Calls {
			if call.State == "may_have_sent" {
				paused = true
			}
		}
		if paused {
			execution.State = "paused"
			execution.LeaseUntil = time.Time{}
			if _, e = tx.UpdateRun(ctx, owner, RunUpdate{ID: id, Status: RunPaused, ErrorCode: ErrResultUnknown, ErrorMessage: unknownResult().Error()}); e != nil {
				return e
			}
			if _, e = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: id, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.paused", Status: RunPaused, Message: unknownResult().Error()}); e != nil {
				return e
			}
			return tx.PutExecution(ctx, execution)
		}
		slot = ProjectSlot{RunID: id, Fence: slot.Fence + 1, LeaseUntil: now.Add(runLeaseTTL)}
		execution.Fence, execution.LeaseUntil = slot.Fence, slot.LeaseUntil
		execution.State = "leased"
		if execution.Bundle != nil {
			execution.State = "save_pending"
		}
		if e = tx.PutSlot(ctx, owner, run.ProjectID, slot); e != nil {
			return e
		}
		if e = tx.PutExecution(ctx, execution); e != nil {
			return e
		}
		lease = RunLease{owner, run.ProjectID, id, slot.Fence}
		return nil
	})
	if err == nil && paused {
		err = unknownResult()
	}
	return lease, execution, err
}

func guard(ctx context.Context, tx Store, l RunLease) (Execution, error) {
	if err := ctx.Err(); err != nil {
		return Execution{}, err
	}
	e, ok, err := tx.GetExecution(ctx, l.Owner, l.RunID)
	if err != nil {
		return e, err
	}
	now, clockErr := tx.Clock(ctx)
	if clockErr != nil {
		return e, clockErr
	}
	if !ok || e.CancelRequested || e.Fence != l.Fence || !e.LeaseUntil.After(now) {
		return e, leaseLost()
	}
	slot, err := tx.GetSlot(ctx, l.Owner, l.ProjectID)
	if err != nil {
		return e, err
	}
	if slot.RunID != l.RunID || slot.Fence != l.Fence || !slot.LeaseUntil.After(now) {
		return e, leaseLost()
	}
	return e, nil
}
func guarded(ctx context.Context, store Store, l RunLease, fn func(Store, *Execution) error) error {
	err := store.Atomic(ctx, l.Owner, l.ProjectID, func(tx Store) error {
		e, err := guard(ctx, tx, l)
		if err != nil {
			return err
		}
		if err = fn(tx, &e); err != nil {
			return err
		}
		if !e.LeaseUntil.IsZero() {
			if _, err = guard(ctx, tx, l); err != nil {
				return err
			}
		}
		return tx.PutExecution(ctx, e)
	})
	if err != nil && !IsCode(err, ErrLeaseLost) && !IsCode(err, ErrResultUnknown) && !IsCode(err, ErrIdempotencyConflict) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return &durableWriteError{err}
	}
	return err
}

// Local journal/event failures are not model failures. Preserve this marker
// through DomainError wrappers so a paid response is never retried as new work.
type durableWriteError struct{ cause error }

func (e *durableWriteError) Error() string { return e.cause.Error() }
func (e *durableWriteError) Unwrap() error { return e.cause }
func isDurableWriteFailure(err error) bool {
	var local *durableWriteError
	return errors.As(err, &local)
}
func hasUncertainCall(e Execution) bool {
	for _, call := range e.Calls {
		if call.State == "may_have_sent" {
			return true
		}
	}
	return false
}

func (s *Service) recoverDurableWriteFailure(ctx context.Context, owner string, run Run, cause error) (Run, error) {
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, cause
	}
	resultErr := persistenceError(cause)
	err := guarded(ctx, g.Store, g.lease, func(tx Store, e *Execution) error {
		var err error
		if hasUncertainCall(*e) {
			resultErr = unknownResult()
			e.State = "paused"
			run, err = tx.UpdateRun(ctx, owner, RunUpdate{ID: run.ID, Status: RunPaused, ErrorCode: ErrResultUnknown, ErrorMessage: resultErr.Error()})
			if err != nil {
				return err
			}
			_, err = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.paused", Status: RunPaused, Message: resultErr.Error()})
			return err
		}
		e.State = "ready"
		if e.Bundle != nil {
			e.State = "save_pending"
		}
		run, err = tx.GetRun(ctx, owner, run.ID)
		return err
	})
	// If MySQL is still unavailable, leave the prior leased record untouched:
	// the scanner will reconsider it after expiry. Never invent a done state.
	if err != nil {
		return run, err
	}
	return run, resultErr
}
func (s *Service) renew(ctx context.Context, l RunLease) error {
	return guarded(ctx, s.store, l, func(tx Store, e *Execution) error {
		now, err := tx.Clock(ctx)
		if err != nil {
			return err
		}
		e.LeaseUntil = now.Add(runLeaseTTL)
		return tx.PutSlot(ctx, l.Owner, l.ProjectID, ProjectSlot{l.RunID, l.Fence, e.LeaseUntil})
	})
}
func (s *Service) release(ctx context.Context, l RunLease) {
	_ = guarded(ctx, s.store, l, func(tx Store, e *Execution) error {
		e.LeaseUntil = time.Time{}
		if e.State == "leased" {
			e.State = "ready"
		}
		return tx.PutSlot(ctx, l.Owner, l.ProjectID, ProjectSlot{RunID: l.RunID, Fence: l.Fence})
	})
}

// Run is also a public entry point and must obtain the same durable lease.
func (s *Service) Run(ctx context.Context, owner, id string) (Run, error) {
	current, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return Run{}, err
	}
	if isTerminalRunStatus(current.Status) || current.Status == RunAwaitingConfirmation {
		return current, nil
	}
	l, e, err := s.claim(ctx, owner, id)
	if err != nil {
		current, _ = s.store.GetRun(ctx, owner, id)
		return current, err
	}
	runCtx, cancel := context.WithCancel(ctx)
	runCtx = context.WithValue(runCtx, journalKey{}, &callJournal{store: s.store, lease: l})
	defer cancel()
	defer s.release(context.Background(), l)
	key := memoryKey(owner, id)
	s.control.mu.Lock()
	s.control.cancel[key] = activeExecution{l.Fence, cancel}
	s.control.mu.Unlock()
	defer func() {
		s.control.mu.Lock()
		if entry := s.control.cancel[key]; entry.fence == l.Fence {
			delete(s.control.cancel, key)
		}
		s.control.mu.Unlock()
	}()
	done := make(chan struct{})
	defer close(done)
	go func() {
		renewTick := time.NewTicker(runLeaseRenew)
		defer renewTick.Stop()
		cancelTick := time.NewTicker(time.Second)
		defer cancelTick.Stop()
		for {
			select {
			case <-done:
				return
			case <-runCtx.Done():
				return
			case <-renewTick.C:
				if err := s.renew(runCtx, l); err != nil {
					cancel()
					return
				}
			case <-cancelTick.C:
				state, _, err := s.store.GetExecution(runCtx, owner, id)
				if err != nil || state.CancelRequested {
					cancel()
					return
				}
			}
		}
	}()
	gs := &guardedStore{Store: s.store, lease: l}
	child := &Service{store: gs, runtime: executionRuntime{Runtime: s.runtime, store: s.store, owner: owner}, pipeline: s.pipeline, control: s.control}
	if e.Bundle != nil {
		return child.savePending(runCtx, owner, current, *e.Bundle)
	}
	return child.runPipeline(runCtx, owner, id)
}

type guardedStore struct {
	Store
	lease RunLease
}

func (g *guardedStore) UpdateRun(ctx context.Context, owner string, u RunUpdate) (out Run, err error) {
	if owner != g.lease.Owner || u.ID != g.lease.RunID {
		return out, leaseLost()
	}
	err = guarded(ctx, g.Store, g.lease, func(tx Store, e *Execution) error {
		var er error
		out, er = tx.UpdateRun(ctx, owner, u)
		if er == nil {
			if isTerminalRunStatus(out.Status) || out.Status == RunAwaitingConfirmation {
				e.State = "done"
			}
			if out.Status == RunPaused {
				e.State = "paused"
			}
		}
		return er
	})
	return
}
func (g *guardedStore) AppendRunEvent(ctx context.Context, owner string, v RunEvent) (out RunEvent, err error) {
	if owner != g.lease.Owner || v.RunID != g.lease.RunID {
		return out, leaseLost()
	}
	err = guarded(ctx, g.Store, g.lease, func(tx Store, e *Execution) error {
		var er error
		out, er = tx.AppendRunEvent(ctx, owner, v)
		return er
	})
	return
}
func (g *guardedStore) SaveFoundation(ctx context.Context, owner string, v Foundation) error {
	if owner != g.lease.Owner || v.ProjectID != g.lease.ProjectID {
		return leaseLost()
	}
	return guarded(ctx, g.Store, g.lease, func(tx Store, e *Execution) error { return tx.SaveFoundation(ctx, owner, v) })
}

func (g *guardedStore) finishRun(ctx context.Context, owner string, u RunUpdate, f *Foundation, event RunEvent) (out Run, persisted RunEvent, err error) {
	if owner != g.lease.Owner || u.ID != g.lease.RunID || event.RunID != g.lease.RunID {
		return out, persisted, leaseLost()
	}
	err = guarded(ctx, g.Store, g.lease, func(tx Store, e *Execution) error {
		if u.Status == RunFailed && hasUncertainCall(*e) {
			u.Status = RunPaused
			u.ErrorCode = ErrResultUnknown
			u.ErrorMessage = unknownResult().Error()
			u.FinishedAt = nil
			event.Status = RunPaused
			event.Type = "run.paused"
			event.Message = u.ErrorMessage
		}
		if f != nil {
			if f.ProjectID != g.lease.ProjectID {
				return leaseLost()
			}
			if er := tx.SaveFoundation(ctx, owner, *f); er != nil {
				return er
			}
		}
		var er error
		out, er = tx.UpdateRun(ctx, owner, u)
		if er != nil {
			return er
		}
		persisted, er = tx.AppendRunEvent(ctx, owner, event)
		if er != nil {
			return er
		}
		e.State = "done"
		if out.Status == RunPaused {
			e.State = "paused"
		}
		return nil
	})
	return
}
func (g *guardedStore) SaveChapterBundle(ctx context.Context, owner string, b ChapterBundle) error {
	if owner != g.lease.Owner || b.RunID != g.lease.RunID || b.ProjectID != g.lease.ProjectID {
		return leaseLost()
	}
	return guarded(ctx, g.Store, g.lease, func(tx Store, e *Execution) error {
		hash := fingerprint(b)
		if e.CommittedHash != "" {
			if e.CommittedHash == hash {
				return nil
			}
			return conflict()
		}
		run, err := tx.GetRun(ctx, owner, b.RunID)
		if err != nil {
			return err
		}
		if isTerminalRunStatus(run.Status) {
			return leaseLost()
		}
		check := &Service{store: tx}
		if err = check.validateRunSnapshot(ctx, owner, run); err != nil {
			return err
		}
		if err = tx.SaveChapterBundle(ctx, owner, b); err != nil {
			return err
		}
		if _, err = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: b.RunID, ProjectID: b.ProjectID, SessionID: run.SessionID, Type: "run.succeeded", Status: b.FinalRunStatus, Progress: 100, Message: "章节已保存"}); err != nil {
			return err
		}
		e.CommittedHash, e.State = hash, "done"
		return nil
	})
}

func (s *Service) savePending(ctx context.Context, owner string, run Run, b ChapterBundle) (Run, error) {
	if err := s.store.SaveChapterBundle(ctx, owner, b); err != nil {
		return run, err
	}
	updated, err := s.store.GetRun(ctx, owner, run.ID)
	if err == nil {
		_ = s.runtime.Publish(ctx, RunEvent{RunID: run.ID, Status: updated.Status, Type: "run.succeeded"})
	}
	return updated, err
}

type journalKey struct{}
type callJournal struct {
	store Store
	lease RunLease
}

func journalComplete(ctx context.Context, key string, p batchfactoryv11.DirectorProvider, request batchfactoryv11.TextCompletionRequest) (string, error) {
	j, _ := ctx.Value(journalKey{}).(*callJournal)
	if j == nil {
		return p.Complete(batchfactoryv11.WithSingleAttempt(ctx), request)
	}
	hash := fingerprint(request)
	var record CallRecord
	err := guarded(ctx, j.store, j.lease, func(tx Store, e *Execution) error {
		if e.Calls == nil {
			e.Calls = map[string]CallRecord{}
		}
		if old, ok := e.Calls[key]; ok {
			record = old
			if old.Fingerprint != hash {
				return conflict()
			}
			if old.State == "may_have_sent" {
				return unknownResult()
			}
			return nil
		}
		record = CallRecord{Fingerprint: hash, State: "prepared"}
		e.Calls[key] = record
		return nil
	})
	if err != nil {
		return "", err
	}
	if record.State == "completed" {
		return record.Response, nil
	}
	if record.State == "rejected" {
		return "", &DomainError{Code: record.ErrorCode, Message: record.ErrorMessage}
	}
	err = guarded(ctx, j.store, j.lease, func(tx Store, e *Execution) error { record.State = "may_have_sent"; e.Calls[key] = record; return nil })
	if err != nil {
		return "", err
	}
	response, err := p.Complete(batchfactoryv11.WithSingleAttempt(ctx), request)
	if err != nil {
		var httpErr *batchfactoryv11.DirectorHTTPError
		if errors.As(err, &httpErr) && (httpErr.StatusCode == 400 || httpErr.StatusCode == 401 || httpErr.StatusCode == 403 || httpErr.StatusCode == 404 || httpErr.StatusCode == 422 || httpErr.StatusCode == 429) {
			mapped := mapModelError(err)
			record.ErrorCode, record.ErrorMessage = domainFailure(mapped)
			record.State = "rejected"
			record.Response = httpErr.Body
			if saveErr := guarded(ctx, j.store, j.lease, func(tx Store, e *Execution) error { e.Calls[key] = record; return nil }); saveErr != nil {
				return "", saveErr
			}
			return "", mapped
		}
		return "", unknownResult()
	}
	err = guarded(ctx, j.store, j.lease, func(tx Store, e *Execution) error {
		record.State = "completed"
		record.Response = response
		e.Calls[key] = record
		return nil
	})
	if err != nil {
		return "", err
	}
	return response, nil
}

func persistBundle(ctx context.Context, b ChapterBundle) error {
	j, _ := ctx.Value(journalKey{}).(*callJournal)
	if j == nil {
		return &DomainError{Code: ErrPersistenceFailed, Message: "缺少执行租约"}
	}
	return guarded(ctx, j.store, j.lease, func(tx Store, e *Execution) error { e.Bundle = &b; e.State = "save_pending"; return nil })
}

type quietRuntime struct{ Runtime }

func (quietRuntime) Enqueue(context.Context, string) error   { return nil }
func (quietRuntime) Publish(context.Context, RunEvent) error { return nil }

type executionRuntime struct {
	Runtime
	store Store
	owner string
}

func (r executionRuntime) IsCancelled(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}
	e, ok, err := r.store.GetExecution(ctx, r.owner, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, leaseLost()
	}
	if e.CancelRequested {
		return true, nil
	}
	// Redis may shorten notification latency, but failure cannot override MySQL.
	cancelled, _ := r.Runtime.IsCancelled(ctx, id)
	return cancelled, nil
}

func (s *Service) queueAtomic(ctx context.Context, owner, project, hash string, fn func(*Service) (Run, error)) (Run, error) {
	var out Run
	err := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		child := &Service{store: tx, runtime: quietRuntime{s.runtime}, pipeline: s.pipeline, control: s.control}
		var err error
		out, err = fn(child)
		if err != nil {
			return err
		}
		e, ok, err := tx.GetExecution(ctx, owner, out.ID)
		if err != nil {
			return err
		}
		if ok {
			if e.Fingerprint != hash {
				return conflict()
			}
			return nil
		}
		return tx.PutExecution(ctx, Execution{Owner: owner, ProjectID: project, RunID: out.ID, State: "ready", Fingerprint: hash, Calls: map[string]CallRecord{}})
	})
	if err == nil {
		s.Notify(ctx, owner, out.ID)
	}
	return out, err
}
func (s *Service) SendMessage(ctx context.Context, owner string, input SendMessageInput) (Run, error) {
	normalized, err := NormalizeSendMessageInput(input)
	if err != nil {
		return Run{}, err
	}
	session, err := s.store.GetSession(ctx, owner, normalized.SessionID)
	if err != nil {
		return Run{}, err
	}
	if normalized.ProjectID == "" {
		normalized.ProjectID = session.ProjectID
	}
	return s.queueAtomic(ctx, owner, session.ProjectID, fingerprint(normalized), func(child *Service) (Run, error) { return child.sendMessage(ctx, owner, normalized) })
}
func (s *Service) RetryRun(ctx context.Context, owner, id string) (Run, error) {
	run, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return Run{}, err
	}
	e, ok, err := s.store.GetExecution(ctx, owner, id)
	if err != nil {
		return Run{}, err
	}
	if ok {
		if e.State == "paused" {
			return run, unknownResult()
		}
		for _, c := range e.Calls {
			if c.State == "may_have_sent" {
				return run, unknownResult()
			}
		}
		if e.Bundle != nil && e.CommittedHash == "" && !e.CancelRequested {
			s.Notify(ctx, owner, id)
			return run, nil
		}
		if e.State == "ready" && !isTerminalRunStatus(run.Status) && !e.CancelRequested {
			s.Notify(ctx, owner, id)
			return run, nil
		}
	}
	return s.queueAtomic(ctx, owner, run.ProjectID, fingerprint("retry:"+id), func(child *Service) (Run, error) {
		state, found, err := child.store.GetExecution(ctx, owner, id)
		if err != nil {
			return Run{}, err
		}
		if found {
			if state.State == "paused" {
				return run, unknownResult()
			}
			for _, call := range state.Calls {
				if call.State == "may_have_sent" {
					return run, unknownResult()
				}
			}
		}
		return child.retryRun(ctx, owner, id)
	})
}
func (s *Service) ConfirmRun(ctx context.Context, owner, id string, blueprint json.RawMessage) (Run, error) {
	run, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return Run{}, err
	}
	return s.queueAtomic(ctx, owner, run.ProjectID, fingerprint(blueprint), func(child *Service) (Run, error) { return child.confirmRun(ctx, owner, id, blueprint) })
}
func (s *Service) CancelRun(ctx context.Context, owner, id string) (Run, error) {
	run, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return Run{}, err
	}
	err = s.store.Atomic(ctx, owner, run.ProjectID, func(tx Store) error {
		latest, e := tx.GetRun(ctx, owner, id)
		if e != nil {
			return e
		}
		if isTerminalRunStatus(latest.Status) {
			run = latest
			return nil
		}
		state, _, e := tx.GetExecution(ctx, owner, id)
		if e != nil {
			return e
		}
		state.Owner, state.RunID, state.ProjectID = owner, id, run.ProjectID
		slot, slotErr := tx.GetSlot(ctx, owner, run.ProjectID)
		if slotErr != nil {
			return slotErr
		}
		if slot.RunID == id {
			slot.LeaseUntil = time.Time{}
			if slotErr = tx.PutSlot(ctx, owner, run.ProjectID, slot); slotErr != nil {
				return slotErr
			}
		}
		state.CancelRequested = true
		state.State = "done"
		if e = tx.PutExecution(ctx, state); e != nil {
			return e
		}
		now := time.Now().UTC()
		run, e = tx.UpdateRun(ctx, owner, RunUpdate{ID: id, Status: RunCancelled, ErrorCode: ErrRunCancelled, ErrorMessage: "已停止后续步骤，已发送请求可能仍由上游处理", FinishedAt: &now})
		if e != nil {
			return e
		}
		_, e = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: id, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.cancelled", Status: RunCancelled, Message: run.ErrorMessage})
		return e
	})
	if err != nil {
		return run, err
	}
	s.control.mu.Lock()
	cancel := s.control.cancel[memoryKey(owner, id)].cancel
	s.control.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	_ = s.runtime.Cancel(ctx, id)
	return run, nil
}

func (s *Service) validateRunSnapshot(ctx context.Context, owner string, run Run) error {
	var p struct {
		ExpectedTarget   *PreviousChapter  `json:"expected_target"`
		PreviousChapters []PreviousChapter `json:"previous_chapters"`
		ChapterNumber    int               `json:"chapter_number"`
		RetryOfRunID     string            `json:"retry_of_run_id"`
	}
	if err := json.Unmarshal(run.PromptSnapshot, &p); err != nil {
		return persistenceError(err)
	}
	if err := s.validateTargetVersion(ctx, owner, run.ProjectID, p.ExpectedTarget); err != nil {
		return err
	}
	return s.validateContinuity(ctx, owner, run.ProjectID, p.PreviousChapters, p.ChapterNumber, p.RetryOfRunID != "")
}
