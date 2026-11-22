package batchfactoryexecution

import (
	"context"
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
