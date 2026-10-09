package task9runtime

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAdmissionArchivedProjectRollsBackBeforeRunOrBookWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects WHERE id=\\? FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, time.Now()))
	mock.ExpectRollback()
	_, err = NewMySQLStore(db).AdmitGeneration(context.Background(), GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: "key", RequestedByUserID: 4})
	if !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdmissionOriginalKeyBytesConflictAfterCollationLookup(t *testing.T) {
	for _, key := range []string{"Key", "café"} {
		t.Run(key, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
			stored := "key"
			if key == "café" {
				stored = "cafe"
			}
			mock.ExpectQuery("SELECT .* FROM runs WHERE batch_project_id=\\? AND idempotency_key=\\? FOR UPDATE").WithArgs(int64(7), key).WillReturnRows(generationRunRows().AddRow(11, 7, stored, time.Now(), "running", 3, "generation", nil, 1, []byte(`{}`), "hash", 4))
			mock.ExpectRollback()
			_, err = NewMySQLStore(db).AdmitGeneration(context.Background(), GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: key, RequestedByUserID: 4})
			if !errors.Is(err, ErrIdempotencyConflict) {
				t.Fatalf("err=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func generationRunRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "batch_project_id", "idempotency_key", "run_at", "status", "max_attempts", "run_kind", "target_book_id", "request_schema_version", "request_snapshot", "request_hash", "requested_by_user_id"})
}

func TestAdmissionMissingBookRollsBackWithoutPartialRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE batch_project_id").WithArgs(int64(7), "key").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT id FROM books WHERE intake_id=\\? ORDER BY id").WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectRollback()
	_, err = NewMySQLStore(db).AdmitGeneration(context.Background(), GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2, 9}, RequestID: "key", RequestedByUserID: 4})
	if !errors.Is(err, ErrInvalidGenerationRequest) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdmissionInsertFailureRollsBackRunAndEarlierBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE batch_project_id").WithArgs(int64(7), "atomic").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT id FROM books WHERE intake_id").WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2).AddRow(9))
	mock.ExpectQuery("SELECT COALESCE\\(ps.settings_json,JSON_OBJECT\\(\\)\\),vp.id,vp.profile_name,vp.version,COALESCE\\(vp.settings_json,JSON_OBJECT\\(\\)\\)").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"project", "profile_id", "profile_name", "profile_version", "profile"}).AddRow(`{}`, nil, nil, nil, `{}`))
	mock.ExpectExec("INSERT INTO runs").WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectExec("INSERT INTO book_runs").WithArgs(int64(11), int64(7), int64(2), 3, "atomic").WillReturnResult(sqlmock.NewResult(12, 1))
	injected := errors.New("materialization unavailable")
	mock.ExpectExec("INSERT INTO book_runs").WithArgs(int64(11), int64(7), int64(9), 3, "atomic").WillReturnError(injected)
	mock.ExpectRollback()
	_, err = NewMySQLStore(db).AdmitGeneration(context.Background(), GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2, 9}, RequestID: "atomic", RequestedByUserID: 4})
	if !errors.Is(err, injected) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAdmissionReplayCreatedFalseWithoutRowsAffected(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	body := []byte(`{"schema_version":1,"batch_project_id":7,"book_ids":[2],"selection_all":false,"hook_enabled":false,"plot_mode":false,"director_mode":"normal","match_audio":false,"shot_duration_limit_sec":15,"requested_by_user_id":4}`)
	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE batch_project_id").WithArgs(int64(7), "replay").WillReturnRows(generationRunRows().AddRow(11, 7, "replay", time.Now(), "running", 3, "generation", 2, 1, body, hash, 4))
	mock.ExpectQuery("SELECT id FROM books WHERE intake_id").WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT id,book_id,attempt FROM book_runs").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"id", "book_id", "attempt"}).AddRow(12, 2, 1))
	mock.ExpectCommit()
	result, err := NewMySQLStore(db).AdmitGeneration(context.Background(), GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: "replay", RequestedByUserID: 4})
	if err != nil || result.Created || result.Run.ID != 11 || len(result.Items) != 1 {
		t.Fatalf("replay=%+v err=%v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationWorkerClaimFiltersLegacyRuns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT run_id,batch_project_id FROM book_runs WHERE id=\\?").WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id"}).AddRow(11, 7))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects WHERE id=\\? FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE id=\\? FOR UPDATE").WithArgs(int64(11)).WillReturnRows(generationRunRows().AddRow(11, 7, "legacy", time.Now(), "running", 3, "legacy", nil, 0, nil, "", nil))
	mock.ExpectRollback()
	_, ok, err := NewMySQLStore(db).Claim(context.Background(), WorkItem{BookRunID: 1, BookID: 2, Attempt: 1}, "worker", time.Now())
	if err != nil || ok {
		t.Fatalf("legacy claim ok=%v err=%v", ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationWorkerClaimLocksProjectRunThenBookAndReturnsNewFence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	body := []byte(`{"schema_version":1,"batch_project_id":7,"book_ids":[2],"selection_all":false,"hook_enabled":false,"plot_mode":false,"director_mode":"normal","match_audio":false,"shot_duration_limit_sec":15,"requested_by_user_id":4}`)
	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	mock.ExpectQuery("SELECT run_id,batch_project_id FROM book_runs").WithArgs(int64(12)).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id"}).AddRow(11, 7))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE id=\\? FOR UPDATE").WithArgs(int64(11)).WillReturnRows(generationRunRows().AddRow(11, 7, "claim", time.Now(), "running", 3, "generation", 2, 1, body, hash, 4))
	mock.ExpectExec("UPDATE book_runs SET status='running'").WithArgs("worker", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(12), int64(11), int64(7), int64(2), 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT attempt,execution_token FROM book_runs").WithArgs(int64(12)).WillReturnRows(sqlmock.NewRows([]string{"attempt", "execution_token"}).AddRow(1, 8))
	mock.ExpectCommit()
	execution, ok, err := NewMySQLStore(db).Claim(context.Background(), WorkItem{BookRunID: 12, BookID: 2, Attempt: 1}, "worker", time.Now())
	if err != nil || !ok || execution.FencingToken != 8 || execution.Owner != "worker" {
		t.Fatalf("execution=%+v ok=%v err=%v", execution, ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationRecoveryQueriesFilterLegacyAndArchivedWork(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT br.id,br.book_id,br.attempt FROM book_runs br JOIN runs r .*BINARY r.run_kind='generation'.*p.archived_at IS NULL").WithArgs(int64(0)).WillReturnRows(sqlmock.NewRows([]string{"id", "book_id", "attempt"}))
	if _, err := NewMySQLStore(db).ListQueuedBookRuns(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT br.id,br.lease_deadline FROM book_runs br JOIN runs r .*BINARY r.run_kind='generation'.*p.archived_at IS NULL").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id", "lease_deadline"}))
	if _, err := NewMySQLStore(db).RecoverStaleBookRuns(context.Background(), time.Now(), 5); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationLegacyClaimWrapperUsesAtomicScheduler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id,batch_project_id,run_at FROM runs WHERE BINARY run_kind='generation'").WillReturnError(sql.ErrNoRows)
	if _, ok, err := NewMySQLStore(db).ClaimDueRun(context.Background(), time.Now()); err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationSchedulerSkipsInvalidEarlierCandidateAndCommitsBooksWithRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Now()
	old := now.Add(-time.Second)
	body := []byte(`{"schema_version":1,"batch_project_id":7,"book_ids":[2],"selection_all":false,"hook_enabled":false,"plot_mode":false,"director_mode":"normal","match_audio":false,"shot_duration_limit_sec":15,"requested_by_user_id":4}`)
	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	mock.ExpectQuery("SELECT id,batch_project_id,run_at FROM runs WHERE BINARY run_kind='generation'").WithArgs(now).WillReturnRows(sqlmock.NewRows([]string{"id", "batch_project_id", "run_at"}).AddRow(10, 7, old))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE SKIP LOCKED").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE id=\\? FOR UPDATE SKIP LOCKED").WithArgs(int64(10)).WillReturnRows(generationRunRows().AddRow(10, 7, "bad", old, "pending", 3, "generation", 2, 99, body, hash, 4))
	mock.ExpectRollback()
	mock.ExpectQuery("SELECT id,batch_project_id,run_at FROM runs .*id>\\?").WithArgs(now, old, old, int64(10)).WillReturnRows(sqlmock.NewRows([]string{"id", "batch_project_id", "run_at"}).AddRow(11, 7, old))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE SKIP LOCKED").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT .* FROM runs WHERE id=\\? FOR UPDATE SKIP LOCKED").WithArgs(int64(11)).WillReturnRows(generationRunRows().AddRow(11, 7, "valid", old, "pending", 3, "generation", 2, 1, body, hash, 4))
	mock.ExpectQuery("SELECT id FROM books WHERE intake_id").WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectExec("INSERT INTO book_runs").WithArgs(int64(11), int64(7), int64(2), 3, "valid").WillReturnResult(sqlmock.NewResult(12, 1))
	mock.ExpectQuery("SELECT id,book_id,attempt FROM book_runs.*FOR UPDATE").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"id", "book_id", "attempt"}).AddRow(12, 2, 1))
	mock.ExpectExec("UPDATE runs SET status='running'").WithArgs(now, int64(11)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	claim, items, ok, err := NewMySQLStore(db).ClaimAndMaterializeDueRun(context.Background(), now)
	if err != nil || !ok || claim.RunID != 11 || len(items) != 1 || items[0].BookRunID != 12 {
		t.Fatalf("claim=%+v items=%+v ok=%v err=%v", claim, items, ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationRetryAndStaleRecoveryLockRunBeforeBookRun(t *testing.T) {
	for _, retry := range []bool{false, true} {
		t.Run(fmt.Sprint(retry), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery("SELECT run_id,batch_project_id FROM book_runs WHERE id=\\?").WithArgs(int64(12)).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id"}).AddRow(11, 7))
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
			mock.ExpectQuery("SELECT .* FROM runs WHERE id=\\? FOR UPDATE").WithArgs(int64(11)).WillReturnRows(generationRunRows().AddRow(11, 7, "legacy", time.Now(), "running", 3, "legacy", nil, 0, nil, "", nil))
			mock.ExpectRollback()
			store := NewMySQLStore(db)
			if retry {
				_, ok, err := store.RetryBookRun(context.Background(), 12)
				if ok || !errors.Is(err, ErrBookRunNotRetryable) {
					t.Fatalf("ok=%v err=%v", ok, err)
				}
			} else {
				_, ok, err := store.recoverOne(context.Background(), 12, time.Now())
				if ok || err != nil {
					t.Fatalf("ok=%v err=%v", ok, err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGenerationRecoveryLimitOneSkipsInvalidBeforeValid(t *testing.T) {
	for _, stale := range []bool{false, true} {
		t.Run(fmt.Sprint(stale), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			now := time.Now()
			expired := now.Add(-time.Minute)
			body := []byte(`{"schema_version":1,"batch_project_id":7,"book_ids":[2],"selection_all":false,"hook_enabled":false,"plot_mode":false,"director_mode":"normal","match_audio":false,"shot_duration_limit_sec":15,"requested_by_user_id":4}`)
			sum := sha256.Sum256(body)
			hash := hex.EncodeToString(sum[:])
			for _, id := range []int64{12, 13} {
				if stale {
					query := mock.ExpectQuery("SELECT br.id,br.lease_deadline FROM book_runs.*ORDER BY br.lease_deadline,br.id LIMIT 1")
					if id == 12 {
						query.WithArgs(now)
					} else {
						query.WithArgs(now, expired, expired, int64(12))
					}
					query.WillReturnRows(sqlmock.NewRows([]string{"id", "lease_deadline"}).AddRow(id, expired))
				} else {
					cursor := int64(0)
					if id == 13 {
						cursor = 12
					}
					mock.ExpectQuery("SELECT br.id,br.book_id,br.attempt FROM book_runs.*br.id>\\?.*LIMIT 1").WithArgs(cursor).WillReturnRows(sqlmock.NewRows([]string{"id", "book_id", "attempt"}).AddRow(id, 2, 1))
				}
				mock.ExpectQuery("SELECT run_id,batch_project_id FROM book_runs").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id"}).AddRow(id-2, 7))
				mock.ExpectBegin()
				mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
				version := 1
				if id == 12 {
					version = 99
				}
				mock.ExpectQuery("SELECT .* FROM runs WHERE id=\\? FOR UPDATE").WithArgs(id - 2).WillReturnRows(generationRunRows().AddRow(id-2, 7, "recovery", now, "running", 3, "generation", 2, version, body, hash, 4))
				if id == 12 {
					mock.ExpectRollback()
					continue
				}
				mock.ExpectQuery("SELECT id FROM books WHERE intake_id").WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				if stale {
					mock.ExpectQuery("SELECT run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status,lease_deadline FROM book_runs.*FOR UPDATE").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id", "book_id", "attempt", "max_attempts", "retryable", "status", "lease_deadline"}).AddRow(11, 7, 2, 1, 3, true, "running", expired))
					mock.ExpectExec("UPDATE book_runs SET status='failed'").WithArgs(now, id).WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectQuery("SELECT id FROM book_runs WHERE run_id").WithArgs(int64(11), int64(2), 2).WillReturnError(sql.ErrNoRows)
					mock.ExpectExec("INSERT INTO book_runs").WithArgs(int64(11), int64(7), int64(2), 2, 3, "recovery").WillReturnResult(sqlmock.NewResult(14, 1))
				} else {
					mock.ExpectQuery("SELECT run_id,batch_project_id,book_id,attempt,status FROM book_runs.*FOR UPDATE").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id", "book_id", "attempt", "status"}).AddRow(11, 7, 2, 1, "queued"))
				}
				mock.ExpectCommit()
			}
			var items []WorkItem
			if stale {
				items, err = NewMySQLStore(db).RecoverStaleBookRuns(context.Background(), now, 1)
			} else {
				items, err = NewMySQLStore(db).ListQueuedBookRuns(context.Background(), 1)
			}
			wantID := int64(13)
			if stale {
				wantID = 14
			}
			if err != nil || len(items) != 1 || items[0].BookRunID != wantID {
				t.Fatalf("items=%+v err=%v", items, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGenerationAggregateLocksProjectRunBeforeLatestAttempt(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT batch_project_id FROM runs WHERE id=\\?").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"batch_project_id"}).AddRow(7))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT batch_project_id,status,run_kind,cancel_requested_at,cancelled_at FROM runs.*FOR UPDATE").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"batch_project_id", "status", "run_kind", "cancel_requested_at", "cancelled_at"}).AddRow(7, "running", "generation", nil, nil))
	mock.ExpectQuery("SELECT br.status FROM book_runs.*FOR UPDATE").WithArgs(int64(11), int64(11)).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("queued"))
	mock.ExpectExec("UPDATE runs SET status='running',finished_at=NULL").WithArgs(int64(11)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	state, err := NewMySQLStore(db).AggregateRunStatus(context.Background(), 11)
	if err != nil || state != RunRunning {
		t.Fatalf("state=%s err=%v", state, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationCompleteFinalizesAndAggregatesInOneTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	execution := Execution{BookRunID: 12, Attempt: 1, FencingToken: 8, Owner: "worker-a"}
	mock.ExpectQuery("SELECT run_id,batch_project_id FROM book_runs").WithArgs(int64(12)).WillReturnRows(sqlmock.NewRows([]string{"run_id", "batch_project_id"}).AddRow(11, 7))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
	mock.ExpectQuery("SELECT batch_project_id,status,run_kind,cancel_requested_at,cancelled_at FROM runs.*FOR UPDATE").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"batch_project_id", "status", "run_kind", "cancel_requested_at", "cancelled_at"}).AddRow(7, "running", "generation", nil, nil))
	mock.ExpectQuery("SELECT book_id,attempt,execution_token,execution_owner,status FROM book_runs.*FOR UPDATE").WithArgs(int64(12), int64(11), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"book_id", "attempt", "execution_token", "execution_owner", "status"}).AddRow(2, 1, 8, "worker-a", "running"))
	mock.ExpectExec("UPDATE book_runs SET status='succeeded'").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), int64(12), 1, uint64(8), "worker-a").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT br.status FROM book_runs.*FOR UPDATE").WithArgs(int64(11), int64(11)).WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("succeeded").AddRow("failed"))
	mock.ExpectExec("UPDATE runs SET status=\\?,finished_at=\\?").WithArgs(string(RunPartialFailed), sqlmock.AnyArg(), int64(11)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	ok, err := NewMySQLStore(db).Complete(context.Background(), execution)
	if err != nil || !ok {
		t.Fatalf("complete ok=%v err=%v", ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerationAggregatePreservesTerminalAndCancellation(t *testing.T) {
	for _, status := range []string{"succeeded", "failed", "partial_failed", "cancelled", "cancelling", "running"} {
		t.Run(status, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var cancel any
			if status == "running" {
				cancel = time.Now()
			}
			mock.ExpectQuery("SELECT batch_project_id FROM runs").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"batch_project_id"}).AddRow(7))
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT intake_id,archived_at FROM batch_projects.*FOR UPDATE").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"intake_id", "archived_at"}).AddRow(3, nil))
			mock.ExpectQuery("SELECT batch_project_id,status,run_kind,cancel_requested_at,cancelled_at FROM runs.*FOR UPDATE").WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"batch_project_id", "status", "run_kind", "cancel_requested_at", "cancelled_at"}).AddRow(7, status, "generation", cancel, nil))
			mock.ExpectRollback()
			got, err := NewMySQLStore(db).AggregateRunStatus(context.Background(), 11)
			if err != nil || string(got) != status {
				t.Fatalf("got=%s err=%v", got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
