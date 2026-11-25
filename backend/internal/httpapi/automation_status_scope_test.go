package httpapi

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type scopedAutomationStatusStore struct {
	memoryAutomationStateStore
	owner, batch string
	err          error
}

func (s *scopedAutomationStatusStore) LoadBatch(_ context.Context, owner, batch string) (automationStateFile, bool, error) {
	s.owner, s.batch = owner, batch
	return s.state, s.found, s.err
}

func (s *scopedAutomationStatusStore) SaveJobs(_ context.Context, jobs map[string]map[string]any) error {
	s.saves++
	return nil
}

func TestAutomationStatusUsesScopedBatchRead(t *testing.T) {
	store := &scopedAutomationStatusStore{memoryAutomationStateStore: memoryAutomationStateStore{
		found: true, state: automationStateFile{Version: 1, Jobs: map[string]map[string]any{
			"old": {"id": "old", "owner": "alice", "batchId": "batch-1", "state": "completed", "updatedAt": time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano)},
			"new": {"id": "new", "owner": "alice", "batchId": "batch-1", "state": "completed", "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)},
		}},
	}}
	reader := NewFileAutomationStatusReader("")
	reader.SetStateStore(store)
	status, err := reader.Status("alice", "batch-1")
	if err != nil || status["id"] != "new" {
		t.Fatalf("status=%v err=%v", status, err)
	}
	if store.loads != 0 || store.owner != "alice" || store.batch != "batch-1" {
		t.Fatalf("single batch status loaded all history: full=%d owner=%q batch=%q", store.loads, store.owner, store.batch)
	}
	store.err = errors.New("database read failed")
	if _, err := reader.Status("alice", "batch-1"); err == nil {
		t.Fatal("scoped read failure was hidden")
	}
	if store.loads != 0 {
		t.Fatal("failed scoped read fell back to loading all users")
	}
}

func TestMySQLAutomationStatusScopesPayloadQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT id FROM batch_automation_jobs WHERE owner_username = \? AND batch_id = \? ORDER BY id LIMIT \? OFFSET \?`).
		WithArgs("alice", "batch-1", automationJobPageSize, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("job-1"))
	mock.ExpectQuery(`SELECT id, job_json FROM batch_automation_jobs WHERE id IN`).WithArgs("job-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "job_json"}).AddRow("job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1","state":"completed"}`))
	reader := NewFileAutomationStatusReader("")
	reader.SetStateStore(NewMySQLAutomationStateStore(db))
	status, err := reader.Status("alice", "batch-1")
	if err != nil || status["id"] != "job-1" {
		t.Fatalf("status=%v err=%v", status, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLBatchAutomationReadKeepsAllPages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for page, count := range []int{automationJobPageSize, 1} {
		ids := sqlmock.NewRows([]string{"id"})
		payloads := sqlmock.NewRows([]string{"id", "job_json"})
		for i := 0; i < count; i++ {
			id := fmt.Sprintf("job-%04d", page*automationJobPageSize+i)
			ids.AddRow(id)
			payloads.AddRow(id, `{"owner":"alice","batchId":"batch-1","state":"completed"}`)
		}
		mock.ExpectQuery(`SELECT id FROM batch_automation_jobs WHERE owner_username = \? AND batch_id = \? ORDER BY id LIMIT \? OFFSET \?`).
			WithArgs("alice", "batch-1", automationJobPageSize, page*automationJobPageSize).WillReturnRows(ids)
		mock.ExpectQuery(`SELECT id, job_json FROM batch_automation_jobs WHERE id IN`).WillReturnRows(payloads)
	}
	store := NewMySQLAutomationStateStore(db).(AutomationBatchStateReader)
	state, found, err := store.LoadBatch(context.Background(), "alice", "batch-1")
	if err != nil || !found || len(state.Jobs) != automationJobPageSize+1 {
		t.Fatalf("batch history lost: jobs=%d found=%v err=%v", len(state.Jobs), found, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
