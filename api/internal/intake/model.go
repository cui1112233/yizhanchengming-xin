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
	RunStatusPending       RunStatus = "pending"
	RunStatusQueued        RunStatus = "queued"
	RunStatusScheduled     RunStatus = "scheduled"
	RunStatusRunning       RunStatus = "running"
	RunStatusCompleted     RunStatus = "completed"
	RunStatusSucceeded     RunStatus = "succeeded"
	RunStatusPartialFailed RunStatus = "partial_failed"
	RunStatusFailed        RunStatus = "failed"
)

type Intake struct {
	ID        int64
	Name      string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ActorScope is the authenticated ownership context attached to newly created
// intake facts. TeamID zero means that only the creating user owns the intake.
type ActorScope struct {
	UserID int64
	TeamID int64
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
	ID           int64
	IntakeID     int64
	Name         string
	Sources      []string
	BookCount    int
	Genders      []string
	Styles       []string
	RunStatus    RunStatus
	FailureCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ArchivedAt   *time.Time
}

type BatchProjectArchivedFilter string

const (
	BatchProjectArchivedActive   BatchProjectArchivedFilter = "active"
	BatchProjectArchivedArchived BatchProjectArchivedFilter = "archived"
	BatchProjectArchivedAll      BatchProjectArchivedFilter = "all"
)

type BatchProjectSort string

const (
	BatchProjectSortUpdatedDesc BatchProjectSort = "updated_desc"
	BatchProjectSortNameAsc     BatchProjectSort = "name_asc"
)

type BatchProjectListQuery struct {
	UserID   int64
	TeamID   int64
	Elevated bool
	Query    string
	Source   string
	Status   RunStatus
	Archived BatchProjectArchivedFilter
	Page     int
	Limit    int
	Sort     BatchProjectSort
}

type BatchProjectPage struct {
	Projects []BatchProject
	Page     int
	Limit    int
	Total    int
}

type Run struct {
	ID             int64
	BatchProjectID int64
	RunAt          time.Time
	Status         RunStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
