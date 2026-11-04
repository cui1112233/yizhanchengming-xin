package httpapi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type memoryAutomationStateStore struct {
	state automationStateFile
	found bool
	loads int
	saves int
}

func TestMySQLAutomationStateStoreLoadsAndSavesJobs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLAutomationStateStore(db)
	mock.ExpectQuery("SELECT id, job_json FROM batch_automation_jobs").WillReturnRows(
		sqlmock.NewRows([]string{"id", "job_json"}).AddRow("job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1","state":"completed"}`),
	)
	state, found, err := store.Load(context.Background())
	if err != nil || !found || state.Jobs["job-1"]["owner"] != "alice" {
		t.Fatalf("load state=%#v found=%v err=%v", state, found, err)
	}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO batch_automation_jobs").WithArgs(
		"job-1", "alice", "batch-1", "completed", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := store.Save(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func (s *memoryAutomationStateStore) Load(context.Context) (automationStateFile, bool, error) {
	s.loads++
	return s.state, s.found, nil
}

func (s *memoryAutomationStateStore) Save(_ context.Context, state automationStateFile) error {
	s.saves++
	s.state = state
	s.found = true
	return nil
}

func writeAutomationTestState(t *testing.T, path string) {
	t.Helper()
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {
			"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "completed",
			"updatedAt": time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC),
		},
	}}
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestBatchAutomationCoordinatorLoadsMySQLStateOnce(t *testing.T) {
	store := &memoryAutomationStateStore{found: true, state: automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "completed"},
	}}}
	coordinator := &BatchAutomationCoordinator{automationStore: store}

	first, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	second, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if store.loads != 1 {
		t.Fatalf("store loads=%d, want one load", store.loads)
	}
	if first.Jobs["job-1"]["batchId"] != second.Jobs["job-1"]["batchId"] {
		t.Fatalf("cached state changed: first=%#v second=%#v", first, second)
	}
}

func TestBatchAutomationCoordinatorMigratesLegacyJSONIntoMySQLStateStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "automation.json")
	writeAutomationTestState(t, path)
	store := &memoryAutomationStateStore{}
	coordinator := &BatchAutomationCoordinator{path: path, automationStore: store}

	state, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	if store.saves != 1 {
		t.Fatalf("store saves=%d, want one legacy import", store.saves)
	}
	if !store.found || state.Jobs["job-1"]["batchId"] != "batch-1" {
		t.Fatalf("migrated state=%#v store=%#v", state, store.state)
	}
}
