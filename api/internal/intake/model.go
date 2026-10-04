package intake

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusPartial   Status = "partial_failed"
	StatusFailed    Status = "failed"
)

type BookStatus string

const (
	BookStatusPending         BookStatus = "pending"
	BookStatusFetched         BookStatus = "fetched"
	BookStatusRetryableFailed BookStatus = "retryable_failed"
)

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

type Intake struct {
	ID        int64
	Name      string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Book struct {
	ID             int64
	IntakeID       int64
	Source         string
	PlatformID     string
	ExternalBookID string
	Title          string
	BodyRef        string
	OriginalText   string
	Category       string
	Genre          string
	Gender         string
	GenderSource   string
	Style          string
	Status         BookStatus
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BatchProject struct {
	ID        int64
	IntakeID  int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Run struct {
	ID             int64
	BatchProjectID int64
	RunAt          time.Time
	Status         RunStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
