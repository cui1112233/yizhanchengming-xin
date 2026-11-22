package batchfactoryexecution

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type memoryExecutionRepository struct {
	byRequest  map[string]EnsureResult
	executions map[string]Execution
	active     map[string]string
	updates    map[string]BookStatus
	legacy     map[string]string
	next       int
}

func newMemoryExecutionRepository() *memoryExecutionRepository {
	return &memoryExecutionRepository{byRequest: map[string]EnsureResult{}, executions: map[string]Execution{}, active: map[string]string{}, updates: map[string]BookStatus{}, legacy: map[string]string{}}
}
func (r *memoryExecutionRepository) Ensure(_ context.Context, record EnsureRecord) (EnsureResult, error) {
	key := record.Owner + "/" + record.BatchID + "/" + record.ClientRequestID
	if result, ok := r.byRequest[key]; ok {
		result.Created = false
		return result, nil
	}
	r.next++
	execution := Execution{ID: "execution-" + string(rune('0'+r.next)), OwnerUsername: record.Owner, BatchID: record.BatchID, Mode: record.Mode, Target: record.Target, ClientRequestID: record.ClientRequestID, Status: ExecutionPlanned}
	result := EnsureResult{Execution: execution, Created: true}
	for _, book := range record.Books {
		if owner, ok := r.active[book.BookID]; ok {
			result.ReusedBooks = append(result.ReusedBooks, book.BookID)
			_ = owner
		} else {
			result.QueuedBooks = append(result.QueuedBooks, book.BookID)
			r.active[book.BookID] = execution.ID
		}
	}
	r.byRequest[key] = result
	r.executions[execution.ID] = execution
	return result, nil
}
func (r *memoryExecutionRepository) Get(_ context.Context, owner, id string) (Execution, error) {
	execution, ok := r.executions[id]
	if !ok || execution.OwnerUsername != owner {
		return Execution{}, ErrNotFound
	}
	return execution, nil
}
func (r *memoryExecutionRepository) GetByClientRequest(_ context.Context, owner, batchID, requestID string) (Execution, bool, error) {
	result, ok := r.byRequest[owner+"/"+batchID+"/"+requestID]
	if !ok {
		return Execution{}, false, nil
	}
	execution := r.executions[result.Execution.ID]
	return execution, true, nil
}
func (r *memoryExecutionRepository) ListBookTimeline(context.Context, string, string, string) ([]TimelineEvent, error) {
	return nil, nil
}
func (r *memoryExecutionRepository) AttachLegacyJob(_ context.Context, owner, executionID, jobID string) error {
	r.legacy[owner+"/"+executionID] = jobID
	execution := r.executions[executionID]
	execution.LegacyJobID = jobID
	r.executions[executionID] = execution
	return nil
}
func (r *memoryExecutionRepository) UpdateBookStatus(_ context.Context, owner, bookID string, status BookStatus, _ Failure) error {
	r.updates[owner+"/"+bookID] = status
	return nil
}

func (r *memoryExecutionRepository) UpdateExecutionStatus(_ context.Context, owner, id string, status ExecutionStatus) error {
	execution, ok := r.executions[id]
	if !ok || execution.OwnerUsername != owner {
		return ErrNotFound
	}
	execution.Status = status
	r.executions[id] = execution
	return nil
}
func (r *memoryExecutionRepository) AppendOutbox(context.Context, OutboxEvent) error { return nil }

type fakeLegacyGateway struct {
	active    LegacyTask
	hasActive bool
	starts    int
	startErr  error
	started   LegacyStartInput
	readTask  LegacyTask
	readFound bool
}

func (g *fakeLegacyGateway) PeekActive(context.Context, string, string) (LegacyTask, bool, error) {
	return g.active, g.hasActive, nil
}
func (g *fakeLegacyGateway) Start(_ context.Context, _ string, _ string, input LegacyStartInput) (LegacyTask, error) {
	g.starts++
	g.started = input
	if g.startErr != nil {
		return LegacyTask{}, g.startErr
	}
	return LegacyTask{JobID: "job-new", State: "running"}, nil
}
func (g *fakeLegacyGateway) Read(context.Context, string, string) (LegacyTask, bool, error) {
	return g.readTask, g.readFound, nil
}

func executionCommand(requestID string, books ...string) EnsureCommand {
	return EnsureCommand{Owner: "alice", BatchID: "batch-1", BookIDs: books, Mode: "video", Target: "published_confirmed", ClientRequestID: requestID, Config: map[string]any{"presetId": "preset-1"}}
}

func TestServiceEnsureExecutionFirstClickQueuesBooks(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{}
	service := NewService(repo, legacy)
	got, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1", "book-2"))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Created || got.ExecutionID == "" || legacy.starts != 0 || len(got.QueuedBooks) != 2 {
		t.Fatalf("response=%#v starts=%d", got, legacy.starts)
	}
}

func TestServiceEnsureExecutionSameRequestReusesExecution(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{}
	service := NewService(repo, legacy)
	first, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1"))
	if err != nil {
		t.Fatal(err)
	}
	if first.ExecutionID != second.ExecutionID || legacy.starts != 0 {
		t.Fatalf("first=%#v second=%#v starts=%d", first, second, legacy.starts)
	}
}

func TestServiceEnsureExecutionDifferentClickOnlyQueuesMissingBooks(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{}
	service := NewService(repo, legacy)
	if _, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1")); err != nil {
		t.Fatal(err)
	}
	got, err := service.EnsureExecution(context.Background(), executionCommand("request-2", "book-1", "book-2"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.ReusedBooks) != 1 || got.ReusedBooks[0] != "book-1" || len(got.QueuedBooks) != 1 || got.QueuedBooks[0] != "book-2" {
		t.Fatalf("response=%#v", got)
	}
}

func TestServiceEnsureExecutionAttachesExistingLegacyTask(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{hasActive: true, active: LegacyTask{JobID: "job-old", State: "running", Books: map[string]LegacyBook{"book-1": {BookID: "book-1", State: "running"}}}}
	service := NewService(repo, legacy)
	got, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1"))
	if err != nil {
		t.Fatal(err)
	}
	if !got.Attached || legacy.starts != 0 || repo.legacy["alice/"+got.ExecutionID] != "job-old" {
		t.Fatalf("response=%#v starts=%d legacy=%#v", got, legacy.starts, repo.legacy)
	}
}

func TestServiceEnsureExecutionBlockedBookDoesNotStopRunnableBook(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{}
	service := NewService(repo, legacy)
	service.ValidateBook = func(id string) Failure {
		if id == "blocked" {
			return Failure{Code: "MISSING_MEDIA", Message: "缺少成片"}
		}
		return Failure{}
	}
	got, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "blocked", "ready"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.BlockedBooks) != 1 || got.BlockedBooks[0] != "blocked" || len(got.QueuedBooks) != 1 || got.QueuedBooks[0] != "ready" || legacy.starts != 0 {
		t.Fatalf("response=%#v starts=%d", got, legacy.starts)
	}
}

func TestServiceEnsureExecutionAmbiguousStartReconcilesWithoutResubmit(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{startErr: errors.New("timeout after submit")}
	service := NewService(repo, legacy)
	got, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1"))
	if err != nil {
		t.Fatal(err)
	}
	payload := json.RawMessage(`{"executionId":"` + got.ExecutionID + `","owner":"alice","batchId":"batch-1","bookIds":["book-1"],"mode":"video","target":"published_confirmed","clientRequestId":"request-1","config":{}}`)
	if err := service.DispatchOutboxEvent(context.Background(), OutboxEvent{ID: "event-1", EventType: "execution.ensure", PayloadJSON: payload, Attempts: 1}); err == nil {
		t.Fatal("ambiguous first dispatch should stay recoverable")
	}
	if repo.updates["alice/book-1"] != BookWaitingExternal || repo.executions[got.ExecutionID].Status != ExecutionReconciling {
		t.Fatalf("updates=%#v execution=%#v", repo.updates, repo.executions[got.ExecutionID])
	}
	if err := service.DispatchOutboxEvent(context.Background(), OutboxEvent{ID: "event-1", EventType: "execution.ensure", PayloadJSON: payload, Attempts: 2}); err == nil {
		t.Fatal("unconfirmed retry should keep reconciling")
	}
	if legacy.starts != 1 {
		t.Fatalf("ambiguous request was resubmitted: %d", legacy.starts)
	}
}

func TestServiceReconcileDoesNotCompletePublishedBookWithoutReceipt(t *testing.T) {
	repo := newMemoryExecutionRepository()
	legacy := &fakeLegacyGateway{readFound: true, readTask: LegacyTask{
		JobID: "job-1", State: "completed",
		Books: map[string]LegacyBook{"book-1": {BookID: "book-1", State: "completed"}},
	}}
	service := NewService(repo, legacy)
	ensured, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1"))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.AttachLegacyJob(context.Background(), "alice", ensured.ExecutionID, "job-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Reconcile(context.Background(), "alice", ensured.ExecutionID); err != nil {
		t.Fatal(err)
	}
	if got := repo.updates["alice/book-1"]; got != BookWaitingExternal {
		t.Fatalf("published book status=%q, want waiting_external until publication readback", got)
	}
	if got := repo.executions[ensured.ExecutionID].Status; got != ExecutionRunning {
		t.Fatalf("execution status=%q, want running until publication readback", got)
	}
}
