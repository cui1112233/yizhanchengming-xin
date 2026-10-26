package novelfetchworkshop

import (
	"fmt"
	"strings"
)

const (
	TaskStatusInputReady    = "input_ready"
	TaskStatusQueued        = "queued"
	TaskStatusProcessing    = "processing"
	TaskStatusWaitingRetry  = "waiting_retry"
	TaskStatusWaitingConfig = "waiting_configuration"
	TaskStatusWaitingLogin  = "waiting_login"
	TaskStatusPendingSubmit = "pending_submit"
	TaskStatusCompleted     = "completed"
	TaskStatusFailed        = "failed"
)

// DeriveTaskStatus is the single task-state rule shared by the Go API and UI.
func DeriveTaskStatus(document Document) string {
	meta := document.Meta
	status := statusValue(meta, "status", "taskStatus", "task_status", "queueState", "queue_state")
	original := statusValue(meta, "originalStatus", "original_status")
	ai := statusValue(meta, "aiStatus", "ai_status")
	submit := statusValue(meta, "siteSubmitStatus", "site_submit_status")

	if hasState(status, "failed", "error") || hasState(original, "failed", "error") || hasState(ai, "failed", "error") || hasState(submit, "failed", "error") {
		return TaskStatusFailed
	}
	if hasState(status, TaskStatusWaitingConfig) || hasState(submit, TaskStatusWaitingConfig) {
		return TaskStatusWaitingConfig
	}
	if hasState(status, TaskStatusWaitingLogin) || hasState(submit, TaskStatusWaitingLogin) {
		return TaskStatusWaitingLogin
	}
	if hasState(status, "queued") {
		return TaskStatusQueued
	}
	if hasState(status, TaskStatusWaitingRetry) || hasState(original, TaskStatusWaitingRetry) || hasState(ai, TaskStatusWaitingRetry) {
		return TaskStatusWaitingRetry
	}
	if hasState(status, "running", "processing") || hasState(original, "running", "processing") || hasState(ai, "running", "processing") {
		return TaskStatusProcessing
	}
	if originalDone(document, original) &&
		(hasState(ai, "ai_done", "done", "completed", "success", "succeeded") || len(document.Versions) > 0) &&
		!hasState(submit, "submitted", "done", "completed", "success", "succeeded") {
		return TaskStatusPendingSubmit
	}
	if hasState(status, "done", "completed", "success", "succeeded") {
		return TaskStatusCompleted
	}
	if hasState(submit, "submitted", "done", "completed", "success", "succeeded") {
		return TaskStatusCompleted
	}
	return TaskStatusInputReady
}

func statusValue(meta map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.ToLower(strings.TrimSpace(fmt.Sprint(meta[key]))); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func hasState(value string, states ...string) bool {
	for _, state := range states {
		if value == state {
			return true
		}
	}
	return false
}

func originalDone(document Document, original string) bool {
	return document.OriginalRawChars > 0 || hasState(original, "done", "completed", "success", "succeeded", "restored")
}
