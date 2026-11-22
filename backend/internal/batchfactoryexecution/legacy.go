package batchfactoryexecution

import (
	"context"
	"strings"
	"time"
)

type LegacyAutomationGateway interface {
	PeekActive(context.Context, string, string) (LegacyTask, bool, error)
	Start(context.Context, string, string, LegacyStartInput) (LegacyTask, error)
	Read(context.Context, string, string) (LegacyTask, bool, error)
}

type LegacyStartInput struct {
	BookIDs         []string
	Mode            string
	Target          string
	Config          map[string]any
	ClientRequestID string
}

type LegacyBook struct {
	BookID  string
	State   string
	Message string
}

type LegacyTask struct {
	JobID string
	State string
	Books map[string]LegacyBook
	Since time.Time
}

type LegacyBookProjection struct {
	Status         BookStatus
	FailureCode    string
	FailureMessage string
}

func LegacyTaskActive(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "running", "paused", "retry_wait", "scheduled", "queued", "pending":
		return true
	default:
		return false
	}
}

func ProjectLegacyTask(task LegacyTask) map[string]LegacyBookProjection {
	projected := make(map[string]LegacyBookProjection, len(task.Books))
	for key, book := range task.Books {
		id := strings.TrimSpace(book.BookID)
		if id == "" {
			id = key
		}
		state := strings.ToLower(strings.TrimSpace(book.State))
		projection := LegacyBookProjection{}
		switch state {
		case "completed", "published", "uploaded", "succeeded", "success":
			projection.Status = BookCompleted
		case "pending", "queued", "running", "generating", "processing", "scheduled", "paused", "retry_wait":
			projection.Status = BookRunning
		case "cancelled", "canceled", "stopped", "skipped":
			projection.Status = BookCancelled
		default:
			projection.Status = BookNeedsAttention
			projection.FailureCode = "LEGACY_STATE_UNKNOWN"
			projection.FailureMessage = strings.TrimSpace(book.Message)
			if projection.FailureMessage == "" {
				projection.FailureMessage = "旧任务状态无法自动确认：" + state
			}
		}
		projected[id] = projection
	}
	return projected
}
