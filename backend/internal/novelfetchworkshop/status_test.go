package novelfetchworkshop

import "testing"

func TestDeriveTaskStatusTreatsSubmittedAIResultAsCompleted(t *testing.T) {
	document := Document{
		BookID: "12345678901",
		Meta: map[string]any{
			"originalStatus":   "done",
			"aiStatus":         "ai_done",
			"siteSubmitStatus": "submitted",
		},
	}
	if got := DeriveTaskStatus(document); got != TaskStatusCompleted {
		t.Fatalf("DeriveTaskStatus() = %q, want %q", got, TaskStatusCompleted)
	}
}

func TestDeriveTaskStatusTreatsProcessedButUnsubmittedTaskAsPendingSubmit(t *testing.T) {
	document := Document{
		BookID:           "12345678901",
		OriginalRawChars: 27831,
		Meta: map[string]any{
			"status":         "completed",
			"originalStatus": "done",
			"aiStatus":       "ai_done",
		},
	}
	if got := DeriveTaskStatus(document); got != TaskStatusPendingSubmit {
		t.Fatalf("DeriveTaskStatus() = %q, want %q", got, TaskStatusPendingSubmit)
	}
}

func TestDeriveTaskStatusKeepsSubmissionConfigurationBlockActionable(t *testing.T) {
	document := Document{Meta: map[string]any{
		"status":           "waiting_configuration",
		"siteSubmitStatus": "waiting_configuration",
		"originalStatus":   "done",
		"aiStatus":         "waiting_ai_config",
	}}
	if got := DeriveTaskStatus(document); got != "waiting_configuration" {
		t.Fatalf("DeriveTaskStatus() = %q, want waiting_configuration", got)
	}
}

func TestDeriveTaskStatusKeepsSubmissionLoginBlockActionable(t *testing.T) {
	document := Document{Meta: map[string]any{
		"status":           "waiting_login",
		"siteSubmitStatus": "waiting_login",
		"originalStatus":   "done",
		"aiStatus":         "ai_done",
	}}
	if got := DeriveTaskStatus(document); got != "waiting_login" {
		t.Fatalf("DeriveTaskStatus() = %q, want waiting_login", got)
	}
}

func TestDeriveTaskStatusKeepsWaitingRetrySeparateFromProcessing(t *testing.T) {
	document := Document{Meta: map[string]any{
		"status": "waiting_retry",
	}}
	if got := DeriveTaskStatus(document); got != TaskStatusWaitingRetry {
		t.Fatalf("DeriveTaskStatus() = %q, want %q", got, TaskStatusWaitingRetry)
	}
}
