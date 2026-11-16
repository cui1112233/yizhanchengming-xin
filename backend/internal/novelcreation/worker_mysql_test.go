package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWorkerSQLAtomicRollbackAndNestedTransaction(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			store, mock := newNovelCreationMockStore(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT 1 FROM novel_projects.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
			mock.ExpectQuery("SELECT s.message_cursor.*FOR UPDATE").WithArgs("alice", "s", "p").WillReturnRows(sqlmock.NewRows([]string{"message_cursor"}).AddRow(4))
			mock.ExpectExec("INSERT INTO novel_messages").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("UPDATE novel_sessions SET message_cursor").WillReturnResult(sqlmock.NewResult(0, 1))
			if fail {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			sentinel := errors.New("outbox insert failed")
			err := store.Atomic(context.Background(), "alice", "p", func(tx Store) error {
				m, e := tx.AppendMessage(context.Background(), "alice", Message{ProjectID: "p", SessionID: "s", Role: "user", Content: "hello"})
				if e != nil {
					return e
				}
				if m.Sequence != 5 {
					t.Fatalf("cursor=%d", m.Sequence)
				}
				if fail {
					return sentinel
				}
				return nil
			})
			if fail && !errors.Is(err, sentinel) || !fail && err != nil {
				t.Fatal(err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWorkerSQLJournalAndOwnerScopedRecovery(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	ctx := context.Background()
	state := Execution{Owner: "alice", ProjectID: "p", RunID: "r", State: "leased", Fence: 7, LeaseUntil: time.Now().UTC().Add(time.Minute), Calls: map[string]CallRecord{"audit:1:repair": {Fingerprint: "f", State: "completed", Response: "raw response"}}}
	mock.ExpectExec("INSERT INTO novel_run_dispatch").WithArgs("alice", "r", "p", "leased", state.LeaseUntil, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_run_steps").WithArgs(sqlmock.AnyArg(), "alice", "r", "p", "audit:1:repair", "completed", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.PutExecution(ctx, state); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(state)
	mock.ExpectQuery("SELECT execution_json FROM novel_run_dispatch WHERE owner_username = \\? AND run_id = \\?").WithArgs("alice", "r").WillReturnRows(sqlmock.NewRows([]string{"execution_json"}).AddRow(raw))
	got, ok, err := store.GetExecution(ctx, "alice", "r")
	if err != nil || !ok || got.Calls["audit:1:repair"].Response != "raw response" || got.Fence != 7 {
		t.Fatalf("recovery=%+v err=%v", got, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerSQLExpiredFenceRollsBackBeforeMutation(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	now := time.Now().UTC()
	e := Execution{Owner: "alice", RunID: "r", ProjectID: "p", Fence: 2, LeaseUntil: now.Add(-time.Second)}
	raw, _ := json.Marshal(e)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM novel_projects.*FOR UPDATE").WithArgs("alice", "p").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectQuery("SELECT execution_json").WithArgs("alice", "r").WillReturnRows(sqlmock.NewRows([]string{"execution_json"}).AddRow(raw))
	mock.ExpectQuery("SELECT UTC_TIMESTAMP").WillReturnRows(sqlmock.NewRows([]string{"now"}).AddRow(now))
	mock.ExpectRollback()
	g := &guardedStore{Store: store, lease: RunLease{Owner: "alice", ProjectID: "p", RunID: "r", Fence: 2}}
	_, err := g.UpdateRun(context.Background(), "alice", RunUpdate{ID: "r", Status: RunSucceeded})
	if !IsCode(err, ErrLeaseLost) {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerFixR4SQLFiltersBusyProjectBeforeLimit(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	mock.ExpectQuery(`SELECT .*NOT EXISTS \(SELECT 1 FROM novel_project_slots slot WHERE slot.owner_username=r.owner_username AND slot.project_id=r.project_id AND slot.lease_until > UTC_TIMESTAMP\(6\)\).*ORDER BY r.created_at LIMIT \?`).WithArgs(64).WillReturnRows(sqlmock.NewRows([]string{"owner", "run", "project"}).AddRow("alice", "free-run", "free-project"))
	entries, err := store.ListDispatch(context.Background(), 64)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].RunID != "free-run" {
		t.Fatalf("dispatch=%+v", entries)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
