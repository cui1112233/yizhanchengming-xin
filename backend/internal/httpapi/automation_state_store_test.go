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
	mock.ExpectQuery("SELECT id FROM batch_automation_jobs").WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow("job-1"),
	)
	mock.ExpectQuery("SELECT id, job_json FROM batch_automation_jobs WHERE id IN").WillReturnRows(
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

func TestMySQLAutomationStateStoreSaveJobsWritesOnlyChangedRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLAutomationStateStore(db).(*mysqlAutomationStateStore)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO batch_automation_jobs").WithArgs(
		"job-1", "alice", "batch-1", "running", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	jobs := map[string]map[string]any{
		"job-1": {"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "running"},
	}
	if err := store.SaveJobs(context.Background(), jobs); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type recordingPartialAutomationStore struct {
	memoryAutomationStateStore
	saveJobs int
	lastJobs map[string]map[string]any
}

func (s *recordingPartialAutomationStore) SaveJobs(_ context.Context, jobs map[string]map[string]any) error {
	s.saveJobs++
	s.lastJobs = jobs
	if s.state.Jobs == nil {
		s.state.Jobs = map[string]map[string]any{}
	}
	for id, job := range jobs {
		s.state.Jobs[id] = job
	}
	s.found = true
	return nil
}

func TestPersistAutomationBookEntryUpsertsOnlyThatJob(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	store := &recordingPartialAutomationStore{memoryAutomationStateStore: memoryAutomationStateStore{
		found: true,
		state: automationStateFile{Version: 1, Jobs: map[string]map[string]any{
			"job-1": {
				"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "running", "updatedAt": now,
				"books": map[string]any{"book-1": map[string]any{"bookId": "book-1", "status": "running"}},
			},
			"job-2": {"id": "job-2", "owner": "bob", "batchId": "batch-2", "state": "running", "updatedAt": now},
		}},
	}}
	coordinator := &BatchAutomationCoordinator{automationStore: store, now: func() time.Time { return now }}

	written, err := coordinator.persistAutomationBookEntry(
		"alice", "batch-1", "job-1", "book-1", nil,
		map[string]any{"status": "completed", "message": "已完成"}, 0, nil,
	)
	if err != nil || !written {
		t.Fatalf("written=%v err=%v", written, err)
	}
	if store.saveJobs != 1 || store.saves != 0 {
		t.Fatalf("saveJobs=%d saves=%d, want one partial save and no full save", store.saveJobs, store.saves)
	}
	if len(store.lastJobs) != 1 || store.lastJobs["job-1"] == nil {
		t.Fatalf("partial save payload=%#v, want only job-1", store.lastJobs)
	}
	if store.lastJobs["job-2"] != nil {
		t.Fatal("unchanged job must not be rewritten")
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
