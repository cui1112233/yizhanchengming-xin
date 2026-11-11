package novelcreation

import "fmt"

type RunStatus string

const (
	RunDraft                 RunStatus = "draft"
	RunClarifying            RunStatus = "clarifying"
	RunAwaitingConfirmation  RunStatus = "awaiting_confirmation"
	RunQueued                RunStatus = "queued"
	RunPlanning              RunStatus = "planning"
	RunWriting               RunStatus = "writing"
	RunNormalizing           RunStatus = "normalizing"
	RunAuditing              RunStatus = "auditing"
	RunRevising              RunStatus = "revising"
	RunSettling              RunStatus = "settling"
	RunIndexing              RunStatus = "indexing"
	RunSucceeded             RunStatus = "succeeded"
	RunSucceededWithWarnings RunStatus = "succeeded_with_warnings"
	RunFailed                RunStatus = "failed"
	RunCancelled             RunStatus = "cancelled"
)

const (
	ErrInputRequired     = "novel_input_required"
	ErrProjectNotFound   = "novel_project_not_found"
	ErrSessionNotFound   = "novel_session_not_found"
	ErrRunInProgress     = "novel_run_in_progress"
	ErrProjectLocked     = "novel_project_locked"
	ErrInvalidTransition = "novel_invalid_transition"
	ErrModelUnavailable  = "novel_model_unavailable"
	ErrGenerationTimeout = "novel_generation_timeout"
	ErrPersistenceFailed = "novel_persistence_failed"
)

type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code
	}
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type CreateProjectInput struct {
	Title              string
	Genre              string
	World              string
	Protagonist        string
	CoreConflict       string
	ModelID            string
	TargetChapterWords int
}

type SendMessageInput struct {
	SessionID      string
	ProjectID      string
	Content        string
	ModelID        string
	ThinkingLevel  string
	IdempotencyKey string
}
