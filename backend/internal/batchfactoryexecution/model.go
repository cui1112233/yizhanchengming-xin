// Package batchfactoryexecution defines the records and state transitions for
// unified batch execution. Model timestamps use UTC. Nil timestamps represent
// events that have yet to happen or retries that have yet to be scheduled.
package batchfactoryexecution

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ExecutionStatus string

const (
	ExecutionPlanned        ExecutionStatus = "planned"
	ExecutionRunning        ExecutionStatus = "running"
	ExecutionNeedsAttention ExecutionStatus = "needs_attention"
	ExecutionCompleted      ExecutionStatus = "completed"
	ExecutionCancelled      ExecutionStatus = "cancelled"
)

type BookStatus string

const (
	BookPending         BookStatus = "pending"
	BookRunning         BookStatus = "running"
	BookWaitingExternal BookStatus = "waiting_external"
	BookCompleted       BookStatus = "completed"
	BookRetryWait       BookStatus = "retry_wait"
	BookNeedsAttention  BookStatus = "needs_attention"
	BookCancelled       BookStatus = "cancelled"
)

type StageStatus string

const (
	StagePending        StageStatus = "pending"
	StageQueued         StageStatus = "queued"
	StageRunning        StageStatus = "running"
	StageSucceeded      StageStatus = "succeeded"
	StageReconciling    StageStatus = "reconciling"
	StageRetryWait      StageStatus = "retry_wait"
	StageTerminalFailed StageStatus = "terminal_failed"
	StageCancelled      StageStatus = "cancelled"
)

// PublicationStatus is the status reported by the publication platform.
type PublicationStatus string

type Execution struct {
	ID              string
	OwnerUsername   string
	BatchID         string
	Mode            string
	Target          string
	ClientRequestID string
	Status          ExecutionStatus
	LegacyJobID     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Book struct {
	ID             string
	ExecutionID    string
	BookID         string
	Title          string
	Status         BookStatus
	CurrentStage   string
	FailureCode    string
	FailureMessage string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StageRun struct {
	ID               string
	ExecutionBookID  string
	StageKey         string
	InputFingerprint string
	Status           StageStatus
	NextRetryAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type StageAttempt struct {
	ID                string
	StageRunID        string
	AttemptNo         uint32
	Status            StageStatus
	ProviderRequestID string
	ErrorCode         string
	ErrorMessage      string
	StartedAt         time.Time
	FinishedAt        *time.Time
}

type PublicationReceipt struct {
	ID              string
	ExecutionBookID string
	Platform        string
	RemoteWorkID    string
	RemoteStatus    PublicationStatus
	PayloadJSON     json.RawMessage
	ConfirmedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

var ErrInvalidTransition = errors.New("invalid state transition")

func (s ExecutionStatus) CanTransition(to ExecutionStatus) bool {
	switch s {
	case ExecutionPlanned:
		return to == ExecutionRunning
	case ExecutionRunning:
		return to == ExecutionCompleted || to == ExecutionNeedsAttention || to == ExecutionCancelled
	default:
		return false
	}
}

func (s ExecutionStatus) ValidateTransition(to ExecutionStatus) error {
	if s.CanTransition(to) {
		return nil
	}
	return fmt.Errorf("%w for execution: %q -> %q", ErrInvalidTransition, s, to)
}

func (s BookStatus) CanTransition(to BookStatus) bool {
	switch s {
	case BookPending:
		return to == BookRunning
	case BookRunning:
		return to == BookWaitingExternal || to == BookCompleted || to == BookRetryWait ||
			to == BookNeedsAttention || to == BookCancelled
	case BookWaitingExternal:
		return to == BookCompleted || to == BookRetryWait || to == BookNeedsAttention || to == BookCancelled
	case BookRetryWait:
		return to == BookRunning || to == BookNeedsAttention || to == BookCancelled
	default:
		return false
	}
}

func (s BookStatus) ValidateTransition(to BookStatus) error {
	if s.CanTransition(to) {
		return nil
	}
	return fmt.Errorf("%w for book: %q -> %q", ErrInvalidTransition, s, to)
}

func (s StageStatus) CanTransition(to StageStatus) bool {
	switch s {
	case StagePending:
		return to == StageQueued
	case StageQueued:
		return to == StageRunning
	case StageRunning:
		return to == StageSucceeded || to == StageReconciling || to == StageRetryWait ||
			to == StageTerminalFailed || to == StageCancelled
	case StageReconciling:
		return to == StageSucceeded || to == StageRetryWait || to == StageTerminalFailed || to == StageCancelled
	case StageRetryWait:
		return to == StageQueued || to == StageRunning || to == StageTerminalFailed || to == StageCancelled
	default:
		return false
	}
}

func (s StageStatus) ValidateTransition(to StageStatus) error {
	if s.CanTransition(to) {
		return nil
	}
	return fmt.Errorf("%w for stage: %q -> %q", ErrInvalidTransition, s, to)
}

// DeriveExecutionStatus gives active books priority over books that need attention.
func DeriveExecutionStatus(books []Book) ExecutionStatus {
	var pending int
	var completed, needsAttention bool
	for _, book := range books {
		switch book.Status {
		case BookPending:
			pending++
		case BookRunning, BookWaitingExternal, BookRetryWait:
			return ExecutionRunning
		case BookCompleted:
			completed = true
		case BookCancelled:
		default:
			needsAttention = true
		}
	}

	switch {
	case pending == len(books):
		return ExecutionPlanned
	case pending > 0:
		return ExecutionRunning
	case needsAttention:
		return ExecutionNeedsAttention
	case completed:
		return ExecutionCompleted
	default:
		return ExecutionCancelled
	}
}
