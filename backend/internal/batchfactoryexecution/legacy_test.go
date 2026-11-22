package batchfactoryexecution

import (
	"testing"
	"time"
)

func TestProjectLegacyTaskSeparatesCompletedAndActiveBooks(t *testing.T) {
	task := LegacyTask{JobID: "job-1", State: "running", Since: time.Now(), Books: map[string]LegacyBook{
		"done":    {BookID: "done", State: "completed"},
		"active":  {BookID: "active", State: "generating"},
		"mystery": {BookID: "mystery", State: "provider_unknown", Message: "上游状态未知"},
	}}
	projection := ProjectLegacyTask(task)
	if projection["done"].Status != BookCompleted {
		t.Fatalf("done=%#v", projection["done"])
	}
	if projection["active"].Status != BookRunning {
		t.Fatalf("active=%#v", projection["active"])
	}
	if projection["mystery"].Status != BookNeedsAttention || projection["mystery"].FailureMessage == "" {
		t.Fatalf("mystery=%#v", projection["mystery"])
	}
}

func TestLegacyTaskActiveStates(t *testing.T) {
	for _, state := range []string{"running", "paused", "retry_wait", "scheduled", "queued"} {
		if !LegacyTaskActive(state) {
			t.Errorf("%s must be active", state)
		}
	}
	for _, state := range []string{"completed", "failed", "cancelled", ""} {
		if LegacyTaskActive(state) {
			t.Errorf("%s must be terminal", state)
		}
	}
}
