package batchfactoryexecution

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrNotFound = errors.New("execution not found")

type EnsureBook struct {
	BookID string
	Title  string
}

type EnsureRecord struct {
	Owner           string
	BatchID         string
	Mode            string
	Target          string
	ClientRequestID string
	ConfigJSON      []byte
	ConfigDigest    string
	Books           []EnsureBook
}

type EnsureResult struct {
	Execution   Execution
	Created     bool
	ReusedBooks []string
	QueuedBooks []string
}

type Failure struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type TimelineEvent struct {
	StageKey         string      `json:"stageKey"`
	Status           StageStatus `json:"status"`
	AttemptNo        uint32      `json:"attemptNo,omitempty"`
	InputFingerprint string      `json:"inputFingerprint,omitempty"`
	ErrorCode        string      `json:"errorCode,omitempty"`
	ErrorMessage     string      `json:"errorMessage,omitempty"`
	CreatedAt        time.Time   `json:"createdAt"`
}

type OutboxEvent struct {
	ID            string          `json:"id"`
	AggregateType string          `json:"aggregateType"`
	AggregateID   string          `json:"aggregateId"`
	EventType     string          `json:"eventType"`
	PayloadJSON   json.RawMessage `json:"payload"`
	AvailableAt   time.Time       `json:"availableAt"`
	ClaimedAt     *time.Time      `json:"claimedAt,omitempty"`
	DeliveredAt   *time.Time      `json:"deliveredAt,omitempty"`
	Attempts      uint32          `json:"attempts"`
	LastError     string          `json:"lastError,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type Repository interface {
	Ensure(context.Context, EnsureRecord) (EnsureResult, error)
	Get(context.Context, string, string) (Execution, error)
	GetByClientRequest(context.Context, string, string, string) (Execution, bool, error)
	ListBookTimeline(context.Context, string, string, string) ([]TimelineEvent, error)
	AttachLegacyJob(context.Context, string, string, string) error
	UpdateBookStatus(context.Context, string, string, BookStatus, Failure) error
	UpdateExecutionStatus(context.Context, string, string, ExecutionStatus) error
	AppendOutbox(context.Context, OutboxEvent) error
}
