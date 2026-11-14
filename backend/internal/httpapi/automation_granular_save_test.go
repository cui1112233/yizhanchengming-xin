package httpapi

import (
	"context"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type recordingAutomationStore struct {
	memoryAutomationStateStore
	savedJobs []map[string]map[string]any
}

func (s *recordingAutomationStore) SaveJobs(_ context.Context, jobs map[string]map[string]any) error {
	copied := make(map[string]map[string]any, len(jobs))
	for jobID, job := range jobs {
		copied[jobID] = job
	}
	s.savedJobs = append(s.savedJobs, copied)
	if s.state.Jobs == nil {
		s.state.Jobs = map[string]map[string]any{}
	}
	for jobID, job := range jobs {
		s.state.Jobs[jobID] = job
	}
	s.found = true
	return nil
}

func newRecordingCoordinator(now time.Time, jobs map[string]map[string]any) (*BatchAutomationCoordinator, *recordingAutomationStore) {
	store := &recordingAutomationStore{memoryAutomationStateStore: memoryAutomationStateStore{
		found: true,
		state: automationStateFile{Version: 1, Jobs: jobs},
	}}
	return &BatchAutomationCoordinator{store: batchfactoryv11.NewMemoryStore(), automationStore: store, now: func() time.Time { return now }}, store
}

func TestReapZombieJobsSavesOnlyReclaimedJobs(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	coordinator, store := newRecordingCoordinator(now, map[string]map[string]any{
		"job-zombie": {
			"id": "job-zombie", "owner": "alice", "batchId": "batch-1", "state": "running",
			"startedAt": now.Add(-2 * time.Hour), "updatedAt": now.Add(-2 * time.Hour),
			"books": map[string]any{
				"book-1": map[string]any{"status": "running", "stage": "novel_fetch", "updatedAt": now.Add(-2 * time.Hour)},
			},
		},
		"job-done": {
			"id": "job-done", "owner": "alice", "batchId": "batch-1", "state": "completed",
			"updatedAt": now.Add(-time.Hour),
		},
	})
	if err := coordinator.ReapZombieJobs(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.savedJobs) != 1 {
		t.Fatalf("save batches=%d, want 1", len(store.savedJobs))
	}
	saved := store.savedJobs[0]
	if len(saved) != 1 {
		t.Fatalf("saved jobs=%v, want only job-zombie", saved)
	}
	if _, ok := saved["job-zombie"]; !ok {
		t.Fatalf("saved jobs=%v, want job-zombie", saved)
	}
}

func TestRecoverInterruptedJobsSavesOnlyRunningJobs(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	coordinator, store := newRecordingCoordinator(now, map[string]map[string]any{
		"job-running": {
			"id": "job-running", "owner": "alice", "batchId": "batch-1", "state": "running",
			"books": map[string]any{
				"book-1": map[string]any{"status": "running", "stage": "comic_images"},
			},
		},
		"job-done": {
			"id": "job-done", "owner": "alice", "batchId": "batch-1", "state": "completed",
			"books": map[string]any{
				"book-2": map[string]any{"status": "completed"},
			},
		},
	})
	if err := coordinator.RecoverInterruptedJobs(); err != nil {
		t.Fatal(err)
	}
	if len(store.savedJobs) != 1 {
		t.Fatalf("save batches=%d, want 1", len(store.savedJobs))
	}
	saved := store.savedJobs[0]
	if len(saved) != 1 {
		t.Fatalf("saved jobs=%v, want only job-running", saved)
	}
	if _, ok := saved["job-running"]; !ok {
		t.Fatalf("saved jobs=%v, want job-running", saved)
	}
}

func TestStatusRepairSavesOnlyRepairedJobs(t *testing.T) {
	now := time.Now().UTC()
	store := &recordingAutomationStore{memoryAutomationStateStore: memoryAutomationStateStore{
		found: true,
		state: automationStateFile{Version: 1, Jobs: map[string]map[string]any{
			"job-bad": {
				"id": "job-bad", "owner": "alice", "batchId": "batch-1", "state": "failed",
				"books": map[string]any{
					"book-1": map[string]any{"status": "running", "stage": "comic_images", "updatedAt": now.Add(-time.Hour)},
				},
			},
			"job-ok": {
				"id": "job-ok", "owner": "alice", "batchId": "batch-1", "state": "completed",
				"books": map[string]any{
					"book-2": map[string]any{"status": "completed"},
				},
			},
		}},
	}}
	reader := NewFileAutomationStatusReader("")
	reader.SetStateStore(store)
	if _, err := reader.Status("alice", "batch-1"); err != nil {
		t.Fatal(err)
	}
	if len(store.savedJobs) != 1 {
		t.Fatalf("save batches=%d, want 1", len(store.savedJobs))
	}
	saved := store.savedJobs[0]
	if len(saved) != 1 {
		t.Fatalf("saved jobs=%v, want only job-bad", saved)
	}
	if _, ok := saved["job-bad"]; !ok {
		t.Fatalf("saved jobs=%v, want job-bad", saved)
	}
}
