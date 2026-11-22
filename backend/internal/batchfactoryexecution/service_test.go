package batchfactoryexecution

import (
	"context"
	"errors"
	"testing"
)

type memoryExecutionRepository struct {
	byRequest map[string]EnsureResult
	active    map[string]string
	updates   map[string]BookStatus
	legacy    map[string]string
	next      int
}

func newMemoryExecutionRepository() *memoryExecutionRepository {
	return &memoryExecutionRepository{byRequest: map[string]EnsureResult{}, active: map[string]string{}, updates: map[string]BookStatus{}, legacy: map[string]string{}}
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
	return result, nil
}
func (r *memoryExecutionRepository) Get(context.Context, string, string) (Execution, error) {
	return Execution{}, ErrNotFound
}
func (r *memoryExecutionRepository) GetByClientRequest(context.Context, string, string, string) (Execution, bool, error) {
	return Execution{}, false, nil
}
func (r *memoryExecutionRepository) ListBookTimeline(context.Context, string, string, string) ([]TimelineEvent, error) {
	return nil, nil
}
func (r *memoryExecutionRepository) AttachLegacyJob(_ context.Context, owner, executionID, jobID string) error {
	r.legacy[owner+"/"+executionID] = jobID
	return nil
}
func (r *memoryExecutionRepository) UpdateBookStatus(_ context.Context, owner, bookID string, status BookStatus, _ Failure) error {
	r.updates[owner+"/"+bookID] = status
	return nil
}
func (r *memoryExecutionRepository) UpdateExecutionStatus(context.Context, string, string, ExecutionStatus) error {
	return nil
}
func (r *memoryExecutionRepository) AppendOutbox(context.Context, OutboxEvent) error { return nil }

type fakeLegacyGateway struct {
	active    LegacyTask
	hasActive bool
	starts    int
	startErr  error
	started   LegacyStartInput
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
	return LegacyTask{}, false, nil
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
	if !got.Created || got.ExecutionID == "" || legacy.starts != 1 || len(got.QueuedBooks) != 2 {
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
	if first.ExecutionID != second.ExecutionID || legacy.starts != 1 {
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
	if len(got.BlockedBooks) != 1 || got.BlockedBooks[0] != "blocked" || len(got.QueuedBooks) != 1 || got.QueuedBooks[0] != "ready" || legacy.starts != 1 {
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
	if len(got.ReconcilingBooks) != 1 || got.ReconcilingBooks[0] != "book-1" || repo.updates["alice/book-1"] != BookWaitingExternal {
		t.Fatalf("response=%#v updates=%#v", got, repo.updates)
	}
	if _, err := service.EnsureExecution(context.Background(), executionCommand("request-1", "book-1")); err != nil {
		t.Fatal(err)
	}
	if legacy.starts != 1 {
		t.Fatalf("ambiguous request was resubmitted: %d", legacy.starts)
	}
}
