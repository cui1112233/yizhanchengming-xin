package task9runtime

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
	"github.com/go-sql-driver/mysql"
)

type integrationFixture struct {
	db        *sql.DB
	store     *MySQLStore
	intakeID  int64
	projectID int64
	bookIDs   []int64
}

// Existing worker/lease tests need scheduled Generation fixtures. Legacy
// CreateRun intentionally remains legacy and must never reach an executor.
func (f integrationFixture) createGenerationRun(ctx context.Context, projectID int64, key string, at time.Time, maxAttempts int) (RunRecord, bool, error) {
	result, err := f.store.AdmitGeneration(ctx, GenerationRequest{BatchProjectID: projectID, RequestID: key, RequestedByUserID: 1})
	if err != nil {
		return RunRecord{}, false, err
	}
	if result.Created {
		tx, err := f.db.BeginTx(ctx, nil)
		if err != nil {
			return RunRecord{}, false, err
		}
		defer tx.Rollback()
		if _, err := tx.ExecContext(ctx, `DELETE FROM book_runs WHERE run_id=?`, result.Run.ID); err != nil {
			return RunRecord{}, false, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE runs SET status='pending',run_at=?,max_attempts=?,started_at=NULL WHERE id=?`, at, maxAttempts, result.Run.ID); err != nil {
			return RunRecord{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return RunRecord{}, false, err
		}
	}
	result.Run.RunAt = at
	result.Run.Status = RunPending
	result.Run.MaxAttempts = maxAttempts
	return result.Run, result.Created, nil
}

func newIntegrationFixture(t *testing.T, books int) integrationFixture {
	t.Helper()
	dsn := os.Getenv("TASK9_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TASK9_MYSQL_DSN not configured")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatal(err)
	}
	ctx := context.Background()
	name := fmt.Sprintf("task9-%d", time.Now().UnixNano())
	res, err := db.ExecContext(ctx, `INSERT INTO intakes (name,status) VALUES (?, 'completed')`, name)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	intakeID, _ := res.LastInsertId()
	t.Cleanup(func() { db.ExecContext(context.Background(), `DELETE FROM intakes WHERE id=?`, intakeID); db.Close() })
	res, err = db.ExecContext(ctx, `INSERT INTO batch_projects (intake_id,name) VALUES (?,?)`, intakeID, name)
	if err != nil {
		t.Fatal(err)
	}
	projectID, _ := res.LastInsertId()
	ids := make([]int64, 0, books)
	for i := 0; i < books; i++ {
		res, err = db.ExecContext(ctx, `INSERT INTO books (intake_id,source,platform_id,external_book_id,title,original_text) VALUES (?,?,?,?,?,?)`, intakeID, "test", "1", fmt.Sprintf("%s-%d", name, i), fmt.Sprintf("book-%d", i), "body")
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		ids = append(ids, id)
	}
	return integrationFixture{db: db, store: NewMySQLStore(db), intakeID: intakeID, projectID: projectID, bookIDs: ids}
}

func TestConcurrentIdempotentRunCreationCreatesOneLogicalRun(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	const n = 20
	ids := make(chan int64, n)
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, _, err := f.store.CreateRun(ctx, f.projectID, "same-key", time.Now().Add(time.Hour), 3)
			if err != nil {
				errCh <- err
				return
			}
			ids <- r.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errCh)
	for err := range errCh {
		t.Fatalf("CreateRun: %v", err)
	}
	var first int64
	for id := range ids {
		if first == 0 {
			first = id
		}
		if id != first {
			t.Fatalf("different logical run ids: %d vs %d", first, id)
		}
	}
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE batch_project_id=? AND idempotency_key='same-key'`, f.projectID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("run count=%d want 1", count)
	}
}

func TestTwoSchedulersOnlyOneClaimsDueRun(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	r, _, err := f.createGenerationRun(ctx, f.projectID, "due", time.Now().Add(-time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claim, ok, err := NewMySQLStore(f.db).ClaimDueRun(ctx, time.Now())
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			if ok {
				if claim.RunID != r.ID {
					t.Errorf("claimed run=%d want %d", claim.RunID, r.ID)
				}
				wins.Add(1)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("scheduler wins=%d want 1", wins.Load())
	}
}

func TestConcurrentRetryCreatesOnlyOneNextAttempt(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	r, _, err := f.createGenerationRun(ctx, f.projectID, "retry", time.Now().Add(-time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.store.ClaimDueRun(ctx, time.Now()); err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	items, err := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET status='failed', retryable=1 WHERE id=?`, items[0].BookRunID); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _, err := f.store.RetryBookRun(ctx, items[0].BookRunID); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("retry: %v", err)
		}
	}
	var a2, a3 int
	if err := f.db.QueryRow(`SELECT SUM(attempt=2), SUM(attempt=3) FROM book_runs WHERE run_id=? AND book_id=?`, r.ID, f.bookIDs[0]).Scan(&a2, &a3); err != nil {
		t.Fatal(err)
	}
	if a2 != 1 || a3 != 0 {
		t.Fatalf("attempt2=%d attempt3=%d", a2, a3)
	}
}

func TestRunAggregationUsesDurableLatestAttempts(t *testing.T) {
	cases := []struct {
		name   string
		states []string
		want   RunState
	}{
		{"all-success", []string{"succeeded", "succeeded", "succeeded"}, RunSucceeded},
		{"partial", []string{"succeeded", "succeeded", "failed"}, RunPartialFailed},
		{"all-failed", []string{"failed", "failed", "failed"}, RunFailed},
		{"still-running", []string{"succeeded", "succeeded", "running"}, RunRunning},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIntegrationFixture(t, 3)
			ctx := context.Background()
			r, _, err := f.createGenerationRun(ctx, f.projectID, tc.name, time.Now().Add(-time.Second), 1)
			if err != nil {
				t.Fatal(err)
			}
			if _, ok, err := f.store.ClaimDueRun(ctx, time.Now()); err != nil || !ok {
				t.Fatalf("claim ok=%v err=%v", ok, err)
			}
			items, err := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
			if err != nil || len(items) != 3 {
				t.Fatalf("items=%v err=%v", items, err)
			}
			for i, item := range items {
				if _, err := f.db.Exec(`UPDATE book_runs SET status=?, retryable=0, max_attempts=1 WHERE id=?`, tc.states[i], item.BookRunID); err != nil {
					t.Fatal(err)
				}
			}
			got, err := f.store.AggregateRunStatus(ctx, r.ID)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestStaleCompleteAndFailRejectedByMySQLFencing(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	r, _, _ := f.createGenerationRun(ctx, f.projectID, "fence", time.Now().Add(-time.Second), 3)
	f.store.ClaimDueRun(ctx, time.Now())
	items, _ := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
	exec, ok, err := f.store.Claim(ctx, items[0], "worker-a", time.Now().Add(time.Second))
	if err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET execution_token=execution_token+1, execution_owner='worker-b' WHERE id=?`, exec.BookRunID); err != nil {
		t.Fatal(err)
	}
	if ok, err := f.store.Complete(ctx, exec); err != nil || ok {
		t.Fatalf("stale complete ok=%v err=%v", ok, err)
	}
	if ok, err := f.store.Fail(ctx, exec, Failure{Code: "x", Message: "safe", Retryable: true}); err != nil || ok {
		t.Fatalf("stale fail ok=%v err=%v", ok, err)
	}
}

func TestRedisWipeRecoveryDoesNotDuplicateLiveExecution(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" {
		t.Skip("TASK9_REDIS_ADDR not configured")
	}
	f := newIntegrationFixture(t, 3)
	ctx := context.Background()
	r, _, _ := f.createGenerationRun(ctx, f.projectID, "wipe", time.Now().Add(-time.Second), 3)
	f.store.ClaimDueRun(ctx, time.Now())
	items, _ := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
	now := time.Now()
	if _, err := f.db.Exec(`UPDATE book_runs SET status='queued' WHERE id=?`, items[0].BookRunID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET status='running', execution_token=1, execution_owner='dead', running_since=?, lease_deadline=? WHERE id=?`, now.Add(-time.Minute), now.Add(-time.Second), items[1].BookRunID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET status='running', execution_token=2, execution_owner='live', running_since=?, lease_deadline=? WHERE id=?`, now.Add(-time.Second), now.Add(time.Minute), items[2].BookRunID); err != nil {
		t.Fatal(err)
	}
	q, err := taskruntime.NewRedisQueue(addr, "task9-wipe-"+fmt.Sprint(time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	if err := q.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	recovery := NewRecovery(f.store, NewQueueCoordinator(q))
	n, err := recovery.Rebuild(ctx, now, 100)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("requeued=%d want queued+expired=2", n)
	}
	var liveStatus string
	var liveToken uint64
	if err := f.db.QueryRow(`SELECT status,execution_token FROM book_runs WHERE id=?`, items[2].BookRunID).Scan(&liveStatus, &liveToken); err != nil {
		t.Fatal(err)
	}
	if liveStatus != "running" || liveToken != 2 {
		t.Fatalf("live execution changed status=%s token=%d", liveStatus, liveToken)
	}
}

type countingExecutor struct{ calls atomic.Int32 }

func (e *countingExecutor) Execute(context.Context, Execution) error { e.calls.Add(1); return nil }

func TestWorkerDuplicateQueueMessageExecutesOnce(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" {
		t.Skip("TASK9_REDIS_ADDR not configured")
	}
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	r, _, _ := f.createGenerationRun(ctx, f.projectID, "dup", time.Now().Add(-time.Second), 3)
	f.store.ClaimDueRun(ctx, time.Now())
	items, _ := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
	q, _ := taskruntime.NewRedisQueue(addr, "task9-dup-"+fmt.Sprint(time.Now().UnixNano()))
	defer q.Close()
	c := NewQueueCoordinator(q)
	c.Enqueue(ctx, items[0])
	c.Enqueue(ctx, items[0])
	leases, _ := taskruntime.NewRedisLeaseStore(addr, "task9-dup-lease-"+fmt.Sprint(time.Now().UnixNano()))
	defer leases.Close()
	exec := &countingExecutor{}
	w := NewWorker(f.store, leases, exec, "worker-a", time.Second, time.Now)
	for i := 0; i < 2; i++ {
		if err := w.RunOnce(ctx, q, 100*time.Millisecond); err != nil && !errors.Is(err, taskruntime.ErrQueueEmpty) {
			t.Fatal(err)
		}
	}
	if exec.calls.Load() != 1 {
		t.Fatalf("executor calls=%d want 1", exec.calls.Load())
	}
}

func TestWorkerAndSchedulerStopOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q := &blockingQueue{}
	w := NewWorker(&noopStore{}, &noopLease{}, &countingExecutor{}, "x", time.Second, time.Now)
	if err := w.Run(ctx, q, 10*time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("worker err=%v", err)
	}
	s := NewScheduler(&noopSchedulerStore{}, nil, time.Now)
	if err := s.Run(ctx, 10*time.Millisecond, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("scheduler err=%v", err)
	}
}

type blockingQueue struct{}

func (*blockingQueue) Enqueue(context.Context, taskruntime.Message) error { return nil }
func (*blockingQueue) Claim(ctx context.Context, _ string, _ time.Duration) (taskruntime.Delivery, error) {
	<-ctx.Done()
	return taskruntime.Delivery{}, ctx.Err()
}
func (*blockingQueue) Ack(context.Context, taskruntime.Delivery) error                 { return nil }
func (*blockingQueue) Nack(context.Context, taskruntime.Delivery, time.Duration) error { return nil }

type noopStore struct{}

func (*noopStore) Claim(context.Context, WorkItem, string, time.Time) (Execution, bool, error) {
	return Execution{}, false, nil
}
func (*noopStore) Renew(context.Context, Execution, time.Time) (bool, error) { return true, nil }
func (*noopStore) Complete(context.Context, Execution) (bool, error)         { return true, nil }
func (*noopStore) Fail(context.Context, Execution, Failure) (bool, error)    { return true, nil }

type noopLease struct{}

func (*noopLease) Claim(context.Context, string, string, uint64, time.Duration) (taskruntime.Lease, bool, error) {
	return taskruntime.Lease{}, true, nil
}
func (*noopLease) Renew(context.Context, taskruntime.Lease, time.Duration) (bool, error) {
	return true, nil
}
func (*noopLease) Release(context.Context, taskruntime.Lease) (bool, error) { return true, nil }
func (*noopLease) RequeueExpired(context.Context, time.Time, int) ([]taskruntime.ExpiredLease, error) {
	return nil, nil
}

type noopSchedulerStore struct{}

func (*noopSchedulerStore) ClaimAndMaterializeDueRun(context.Context, time.Time) (RunClaim, []WorkItem, bool, error) {
	return RunClaim{}, nil, false, nil
}

func TestGenerationAdmissionConcurrentFullMaterialization(t *testing.T) {
	f := newIntegrationFixture(t, 3)
	ctx := context.Background()
	const n = 20
	results := make(chan AdmissionResult, n)
	errs := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := f.store.AdmitGeneration(ctx, GenerationRequest{BatchProjectID: f.projectID, BookIDs: f.bookIDs, RequestID: "full-admission", RequestedByUserID: 1})
			if err != nil {
				errs <- err
				return
			}
			results <- r
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	var id int64
	created := 0
	for r := range results {
		if id == 0 {
			id = r.Run.ID
		}
		if r.Run.ID != id || len(r.Items) != 3 {
			t.Fatalf("incomplete admission: %+v", r)
		}
		if r.Created {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("created=%d", created)
	}
	var runs, books, nulls int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE batch_project_id=?`, f.projectID).Scan(&runs); err != nil {
		t.Fatal(err)
	}
	if err := f.db.QueryRow(`SELECT COUNT(*),SUM(run_id IS NULL) FROM book_runs WHERE batch_project_id=?`, f.projectID).Scan(&books, &nulls); err != nil {
		t.Fatal(err)
	}
	if runs != 1 || books != 3 || nulls != 0 {
		t.Fatalf("runs=%d books=%d nulls=%d", runs, books, nulls)
	}
}

func TestGenerationAdmissionFrozenAllSelectionAndConflicts(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	req := GenerationRequest{BatchProjectID: f.projectID, RequestID: "frozen", RequestedByUserID: 1}
	first, err := f.store.AdmitGeneration(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`INSERT INTO books (intake_id,source,platform_id,external_book_id,original_text) VALUES (?,'test','1','added','private original')`, f.intakeID); err != nil {
		t.Fatal(err)
	}
	second, err := f.store.AdmitGeneration(ctx, req)
	if err != nil || second.Created || second.Run.ID != first.Run.ID || len(second.Items) != 1 {
		t.Fatalf("replay=%+v err=%v", second, err)
	}
	req.HookEnabled = true
	if _, err := f.store.AdmitGeneration(ctx, req); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("different hash=%v", err)
	}
	req.HookEnabled = false
	req.RequestID = "FROZEN"
	if _, err := f.store.AdmitGeneration(ctx, req); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("different key bytes=%v", err)
	}
}

func TestGenerationAdmissionForeignAndMissingBooksAreAtomic(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	other := newIntegrationFixture(t, 1)
	ctx := context.Background()
	for _, id := range []int64{other.bookIDs[0], 9223372036854775807} {
		_, err := f.store.AdmitGeneration(ctx, GenerationRequest{BatchProjectID: f.projectID, BookIDs: []int64{f.bookIDs[0], id}, RequestID: fmt.Sprint(id), RequestedByUserID: 1})
		if !errors.Is(err, ErrInvalidGenerationRequest) {
			t.Fatalf("foreign/missing id=%d err=%v", id, err)
		}
	}
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE batch_project_id=?`, f.projectID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("partial runs=%d err=%v", count, err)
	}
	if _, err := f.db.Exec(`UPDATE batch_projects SET archived_at=NOW(6) WHERE id=?`, f.projectID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.AdmitGeneration(ctx, GenerationRequest{BatchProjectID: f.projectID, RequestID: "archived", RequestedByUserID: 1}); !errors.Is(err, ErrProjectArchived) {
		t.Fatalf("archived=%v", err)
	}
}

func TestGenerationAdmissionClientFoundRowsPreservesCreated(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	cfg, err := mysql.ParseDSN(os.Getenv("TASK9_MYSQL_DSN"))
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	cfg.ClientFoundRows = true
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	req := GenerationRequest{BatchProjectID: f.projectID, BookIDs: f.bookIDs, RequestID: "found-rows", RequestedByUserID: 1}
	first, err := store.AdmitGeneration(context.Background(), req)
	if err != nil || !first.Created {
		t.Fatalf("first Created=%v err=%v", first.Created, err)
	}
	replay, err := store.AdmitGeneration(context.Background(), req)
	if err != nil || replay.Created || replay.Run.ID != first.Run.ID {
		t.Fatalf("replay Created=%v err=%v", replay.Created, err)
	}
}

func TestGenerationMigrationLiveLegacyDefaultAndDownRejects(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	r, _, err := f.store.CreateRun(context.Background(), f.projectID, "legacy-default", time.Now(), 3)
	if err != nil {
		t.Fatal(err)
	}
	var kind string
	if err := f.db.QueryRow(`SELECT run_kind FROM runs WHERE id=?`, r.ID).Scan(&kind); err != nil || kind != "legacy" {
		t.Fatalf("kind=%s err=%v", kind, err)
	}
	source, err := os.ReadFile("../../db/migrations/00021_generation_runtime.sql")
	if err != nil {
		t.Fatal(err)
	}
	down := strings.Split(string(source), "-- +goose Down")[1]
	_, err = f.db.Exec(down)
	var dbErr *mysql.MySQLError
	if !errors.As(err, &dbErr) || dbErr.Number != 1644 {
		t.Fatalf("Down must SIGNAL, err=%v", err)
	}
	if _, _, ok, err := f.store.ClaimAndMaterializeDueRun(context.Background(), time.Now()); err != nil || ok {
		t.Fatalf("legacy scheduler claimed=%v err=%v", ok, err)
	}
}

func TestGenerationSchedulerSkipsArchivedInvalidAndCaseVariantBeforeValid(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	archived := newIntegrationFixture(t, 1)
	ctx := context.Background()
	now := time.Now()
	bad, _, err := f.createGenerationRun(ctx, f.projectID, "bad-first", now.Add(-3*time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE runs SET request_schema_version=99 WHERE id=?`, bad.ID); err != nil {
		t.Fatal(err)
	}
	archiveRun, _, err := archived.createGenerationRun(ctx, archived.projectID, "archived-before", now.Add(-2*time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE batch_projects SET archived_at=NOW(6) WHERE id=?`, archived.projectID); err != nil {
		t.Fatal(err)
	}
	caseRun, _, err := f.createGenerationRun(ctx, f.projectID, "case-before", now.Add(-time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE runs SET run_kind='Generation' WHERE id=?`, caseRun.ID); err != nil {
		t.Fatal(err)
	}
	valid, _, err := f.createGenerationRun(ctx, f.projectID, "valid-after", now, 3)
	if err != nil {
		t.Fatal(err)
	}
	claim, items, ok, err := f.store.ClaimAndMaterializeDueRun(ctx, now)
	if err != nil || !ok || claim.RunID != valid.ID || len(items) != 1 {
		t.Fatalf("claim=%+v items=%+v ok=%v err=%v", claim, items, ok, err)
	}
	for _, id := range []int64{bad.ID, archiveRun.ID, caseRun.ID} {
		var status string
		var count int
		if err := f.db.QueryRow(`SELECT status,(SELECT COUNT(*) FROM book_runs WHERE run_id=runs.id) FROM runs WHERE id=?`, id).Scan(&status, &count); err != nil || status != "pending" || count != 0 {
			t.Fatalf("invalid run=%d status=%s count=%d err=%v", id, status, count, err)
		}
	}
}

func TestGenerationRecoveryRebuildOnlyResetsValidExactZeroBookRun(t *testing.T) {
	f := newIntegrationFixture(t, 1)
	archived := newIntegrationFixture(t, 1)
	ctx := context.Background()
	now := time.Now()
	var ids []int64
	for _, key := range []string{"unsupported", "wrong-hash", "empty", "case", "legacy", "valid"} {
		r, _, err := f.createGenerationRun(ctx, f.projectID, key, now, 3)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, r.ID)
	}
	archiveRun, _, err := archived.createGenerationRun(ctx, archived.projectID, "archived", now, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE runs SET status='running' WHERE batch_project_id IN (?,?)`, f.projectID, archived.projectID); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct {
		id        int64
		statement string
	}{{ids[0], `request_schema_version=99`}, {ids[1], `request_hash='bad'`}, {ids[2], `request_snapshot=JSON_SET(request_snapshot,'$.book_ids',JSON_ARRAY())`}, {ids[3], `run_kind='Generation'`}, {ids[4], `run_kind='legacy'`}} {
		if _, err := f.db.Exec(`UPDATE runs SET `+mutation.statement+` WHERE id=?`, mutation.id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.db.Exec(`UPDATE batch_projects SET archived_at=NOW(6) WHERE id=?`, archived.projectID); err != nil {
		t.Fatal(err)
	}
	if n, err := NewRecovery(f.store, nil).Rebuild(ctx, now, 1); err != nil || n != 0 {
		t.Fatalf("rebuild=%d err=%v", n, err)
	}
	for _, id := range append(ids, archiveRun.ID) {
		var status string
		if err := f.db.QueryRow(`SELECT status FROM runs WHERE id=?`, id).Scan(&status); err != nil {
			t.Fatal(err)
		}
		want := "running"
		if id == ids[5] {
			want = "pending"
		}
		if status != want {
			t.Fatalf("run=%d status=%s want=%s", id, status, want)
		}
	}
}
