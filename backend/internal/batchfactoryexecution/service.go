package batchfactoryexecution

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrInvalidCommand = errors.New("invalid execution command")

type EnsureCommand struct {
	Owner           string
	BatchID         string
	BookIDs         []string
	Mode            string
	Target          string
	ClientRequestID string
	Config          map[string]any
}

type EnsureResponse struct {
	ExecutionID      string   `json:"executionId"`
	Created          bool     `json:"created"`
	Attached         bool     `json:"attached"`
	ReusedBooks      []string `json:"reusedBooks"`
	QueuedBooks      []string `json:"queuedBooks"`
	ReconcilingBooks []string `json:"reconcilingBooks"`
	BlockedBooks     []string `json:"blockedBooks"`
}

type Service struct {
	Repository   Repository
	Legacy       LegacyAutomationGateway
	ValidateBook func(string) Failure
}

func NewService(repository Repository, legacy LegacyAutomationGateway) *Service {
	return &Service{Repository: repository, Legacy: legacy}
}

func validateEnsureCommand(command EnsureCommand) ([]string, error) {
	command.Owner = strings.TrimSpace(command.Owner)
	command.BatchID = strings.TrimSpace(command.BatchID)
	command.Mode = strings.TrimSpace(command.Mode)
	command.Target = strings.TrimSpace(command.Target)
	command.ClientRequestID = strings.TrimSpace(command.ClientRequestID)
	if command.Owner == "" || command.BatchID == "" || command.ClientRequestID == "" {
		return nil, fmt.Errorf("%w: owner, batch and client request id are required", ErrInvalidCommand)
	}
	if command.Mode != "video" && command.Mode != "comic" {
		return nil, fmt.Errorf("%w: unsupported mode", ErrInvalidCommand)
	}
	if command.Target == "" {
		return nil, fmt.Errorf("%w: target is required", ErrInvalidCommand)
	}
	seen := map[string]struct{}{}
	ids := make([]string, 0, len(command.BookIDs))
	for _, raw := range command.BookIDs {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("%w: at least one book is required", ErrInvalidCommand)
	}
	return ids, nil
}

func (s *Service) EnsureExecution(ctx context.Context, command EnsureCommand) (EnsureResponse, error) {
	var response EnsureResponse
	if s == nil || s.Repository == nil {
		return response, fmt.Errorf("execution service unavailable")
	}
	bookIDs, err := validateEnsureCommand(command)
	if err != nil {
		return response, err
	}
	canonical, err := CanonicalJSON(command.Config)
	if err != nil {
		return response, fmt.Errorf("freeze execution config: %w", err)
	}
	digest, err := ConfigDigest(command.Config)
	if err != nil {
		return response, fmt.Errorf("digest execution config: %w", err)
	}

	var active LegacyTask
	var hasActive bool
	if s.Legacy != nil {
		active, hasActive, err = s.Legacy.PeekActive(ctx, command.Owner, command.BatchID)
		if err != nil {
			return response, err
		}
	}
	blocked := map[string]Failure{}
	books := make([]EnsureBook, 0, len(bookIDs))
	titles, _ := command.Config["bookTitles"].(map[string]any)
	for _, id := range bookIDs {
		books = append(books, EnsureBook{BookID: id, Title: strings.TrimSpace(fmt.Sprint(titles[id]))})
		if s.ValidateBook != nil {
			if failure := s.ValidateBook(id); strings.TrimSpace(failure.Code) != "" {
				blocked[id] = failure
				response.BlockedBooks = append(response.BlockedBooks, id)
			}
		}
	}
	result, err := s.Repository.Ensure(ctx, EnsureRecord{Owner: command.Owner, BatchID: command.BatchID, Mode: command.Mode, Target: command.Target, ClientRequestID: command.ClientRequestID, ConfigJSON: canonical, ConfigDigest: digest, Books: books})
	if err != nil {
		return response, err
	}
	response.ExecutionID = result.Execution.ID
	response.Created = result.Created
	response.ReusedBooks = append([]string(nil), result.ReusedBooks...)
	if !result.Created {
		response.QueuedBooks = append([]string(nil), result.QueuedBooks...)
		return response, nil
	}

	for id, failure := range blocked {
		if err := s.Repository.UpdateBookStatus(ctx, command.Owner, id, BookNeedsAttention, failure); err != nil {
			return response, err
		}
	}
	runnable := make([]string, 0, len(result.QueuedBooks))
	for _, id := range result.QueuedBooks {
		if _, isBlocked := blocked[id]; !isBlocked {
			runnable = append(runnable, id)
		}
	}
	response.QueuedBooks = append(response.QueuedBooks, runnable...)
	if hasActive {
		if err := s.Repository.AttachLegacyJob(ctx, command.Owner, result.Execution.ID, active.JobID); err != nil {
			return response, err
		}
		response.Attached = true
		for id, projected := range ProjectLegacyTask(active) {
			if containsString(bookIDs, id) {
				_ = s.Repository.UpdateBookStatus(ctx, command.Owner, id, projected.Status, Failure{Code: projected.FailureCode, Message: projected.FailureMessage})
			}
		}
		return response, nil
	}
	if len(runnable) == 0 || s.Legacy == nil {
		return response, nil
	}
	legacy, startErr := s.Legacy.Start(ctx, command.Owner, command.BatchID, LegacyStartInput{BookIDs: runnable, Mode: command.Mode, Target: command.Target, Config: command.Config, ClientRequestID: command.ClientRequestID})
	if startErr != nil {
		response.QueuedBooks = nil
		response.ReconcilingBooks = append(response.ReconcilingBooks, runnable...)
		for _, id := range runnable {
			if updateErr := s.Repository.UpdateBookStatus(ctx, command.Owner, id, BookWaitingExternal, Failure{Code: "AMBIGUOUS_LEGACY_START", Message: startErr.Error()}); updateErr != nil {
				return response, updateErr
			}
		}
		return response, nil
	}
	if strings.TrimSpace(legacy.JobID) != "" {
		if err := s.Repository.AttachLegacyJob(ctx, command.Owner, result.Execution.ID, legacy.JobID); err != nil {
			return response, err
		}
	}
	return response, nil
}

func (s *Service) GetExecution(ctx context.Context, owner, executionID string) (Execution, error) {
	if s == nil || s.Repository == nil {
		return Execution{}, fmt.Errorf("execution service unavailable")
	}
	return s.Repository.Get(ctx, strings.TrimSpace(owner), strings.TrimSpace(executionID))
}

func (s *Service) Timeline(ctx context.Context, owner, executionID, bookID string) ([]TimelineEvent, error) {
	if s == nil || s.Repository == nil {
		return nil, fmt.Errorf("execution service unavailable")
	}
	return s.Repository.ListBookTimeline(ctx, strings.TrimSpace(owner), strings.TrimSpace(executionID), strings.TrimSpace(bookID))
}

func (s *Service) Cancel(ctx context.Context, owner, executionID string) (Execution, error) {
	execution, err := s.GetExecution(ctx, owner, executionID)
	if err != nil {
		return Execution{}, err
	}
	if execution.Status == ExecutionCompleted || execution.Status == ExecutionCancelled {
		return execution, nil
	}
	if err = s.Repository.UpdateExecutionStatus(ctx, owner, executionID, ExecutionCancelled); err != nil {
		return Execution{}, err
	}
	execution.Status = ExecutionCancelled
	return execution, nil
}

func (s *Service) Reconcile(ctx context.Context, owner, executionID string) (Execution, error) {
	execution, err := s.GetExecution(ctx, owner, executionID)
	if err != nil {
		return Execution{}, err
	}
	if s.Legacy == nil || strings.TrimSpace(execution.LegacyJobID) == "" {
		return execution, nil
	}
	legacy, found, err := s.Legacy.Read(ctx, owner, execution.LegacyJobID)
	if err != nil {
		return Execution{}, err
	}
	if !found {
		return execution, nil
	}
	projected := ProjectLegacyTask(legacy)
	needsAttention := false
	for id, book := range projected {
		if book.Status == BookCompleted && !CanCompleteBook(execution.Target, nil) {
			// The legacy coordinator can prove local generation/upload work, but
			// not the remote publication receipt required by this target.
			book.Status = BookWaitingExternal
		}
		if book.Status == BookNeedsAttention {
			needsAttention = true
		}
		if err := s.Repository.UpdateBookStatus(ctx, owner, id, book.Status, Failure{Code: book.FailureCode, Message: book.FailureMessage}); err != nil && !errors.Is(err, ErrNotFound) {
			return Execution{}, err
		}
	}
	// A legacy "completed" state is not enough to claim published_confirmed;
	// publication receipts are reconciled separately. Keep it running unless a
	// visible legacy failure requires attention.
	status := ExecutionRunning
	if needsAttention || (!LegacyTaskActive(legacy.State) && strings.ToLower(strings.TrimSpace(legacy.State)) != "completed") {
		status = ExecutionNeedsAttention
	}
	if execution.Status != status {
		if err := s.Repository.UpdateExecutionStatus(ctx, owner, executionID, status); err != nil {
			return Execution{}, err
		}
		execution.Status = status
	}
	return execution, nil
}

func (s *Service) DispatchOutboxEvent(ctx context.Context, event OutboxEvent) error {
	if event.EventType != "execution.ensure" {
		return nil
	}
	var payload struct {
		ExecutionID     string         `json:"executionId"`
		Owner           string         `json:"owner"`
		BatchID         string         `json:"batchId"`
		BookIDs         []string       `json:"bookIds"`
		Mode            string         `json:"mode"`
		Target          string         `json:"target"`
		ClientRequestID string         `json:"clientRequestId"`
		Config          map[string]any `json:"config"`
	}
	if err := json.Unmarshal(event.PayloadJSON, &payload); err != nil {
		return fmt.Errorf("decode execution outbox event: %w", err)
	}
	if s.Legacy == nil {
		return fmt.Errorf("legacy execution gateway unavailable")
	}
	execution, found, err := s.Repository.GetByClientRequest(ctx, payload.Owner, payload.BatchID, payload.ClientRequestID)
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	if strings.TrimSpace(execution.LegacyJobID) != "" {
		return nil
	}
	active, hasActive, err := s.Legacy.PeekActive(ctx, payload.Owner, payload.BatchID)
	if err != nil {
		return err
	}
	if hasActive {
		return s.Repository.AttachLegacyJob(ctx, payload.Owner, execution.ID, active.JobID)
	}
	if len(payload.BookIDs) == 0 {
		return nil
	}
	legacy, err := s.Legacy.Start(ctx, payload.Owner, payload.BatchID, LegacyStartInput{BookIDs: payload.BookIDs, Mode: payload.Mode, Target: payload.Target, Config: payload.Config, ClientRequestID: payload.ClientRequestID})
	if err != nil {
		return err
	}
	if strings.TrimSpace(legacy.JobID) == "" {
		return fmt.Errorf("legacy start returned no job id")
	}
	return s.Repository.AttachLegacyJob(ctx, payload.Owner, execution.ID, legacy.JobID)
}

func containsString(values []string, expected string) bool {
	index := sort.SearchStrings(sortedCopy(values), expected)
	sorted := sortedCopy(values)
	return index < len(sorted) && sorted[index] == expected
}
func sortedCopy(values []string) []string {
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return copied
}
