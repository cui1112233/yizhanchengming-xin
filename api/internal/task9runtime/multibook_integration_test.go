package task9runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type selectiveExecutor struct {
	mu    sync.Mutex
	fail  map[int64]bool
	calls map[int64]int
}

func (e *selectiveExecutor) Execute(_ context.Context, x Execution) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls[x.BookRunID]++
	if e.fail[x.BookRunID] {
		return errors.New("fixture book execution failed")
	}
	return nil
}

func TestMultiBookFailureIsolationRetryAndReaggregate(t *testing.T) {
	f := newIntegrationFixture(t, 3)
	ctx := context.Background()
	run, _, err := f.createGenerationRun(ctx, f.projectID, "multibook-isolation", time.Now().Add(-time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	claim, ok, err := f.store.ClaimDueRun(ctx, time.Now())
	if err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	items, err := f.store.EnsureQueuedBooks(ctx, claim)
	if err != nil || len(items) != 3 {
		t.Fatalf("items=%v err=%v", items, err)
	}

	exec := &selectiveExecutor{fail: map[int64]bool{items[1].BookRunID: true}, calls: map[int64]int{}}
	worker := NewWorker(f.store, nil, exec, "worker-isolation", time.Second, time.Now)
	for _, item := range items {
		if err := worker.Process(ctx, item); err != nil {
			t.Fatalf("process %d: %v", item.BookRunID, err)
		}
	}

	for _, idx := range []int{0, 2} {
		var status string
		if err := f.db.QueryRow(`SELECT status FROM book_runs WHERE id=?`, items[idx].BookRunID).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status != "succeeded" {
			t.Fatalf("book %d status=%s want succeeded", idx, status)
		}
	}
	var failedStatus string
	if err := f.db.QueryRow(`SELECT status FROM book_runs WHERE id=?`, items[1].BookRunID).Scan(&failedStatus); err != nil {
		t.Fatal(err)
	}
	if failedStatus != "failed" {
		t.Fatalf("B status=%s", failedStatus)
	}
	state, err := f.store.AggregateRunStatus(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state != RunPartialFailed {
		t.Fatalf("aggregate after A/C success B failure=%s want partial_failed", state)
	}

	retry, created, err := f.store.RetryBookRun(ctx, items[1].BookRunID)
	if err != nil || !created {
		t.Fatalf("retry created=%v err=%v", created, err)
	}
	if retry.BookID != items[1].BookID || retry.Attempt != 2 {
		t.Fatalf("retry=%+v", retry)
	}
	if err := worker.Process(ctx, retry); err != nil {
		t.Fatal(err)
	}
	state, err = f.store.AggregateRunStatus(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state != RunSucceeded {
		t.Fatalf("aggregate after retry=%s want succeeded", state)
	}

	if exec.calls[items[0].BookRunID] != 1 || exec.calls[items[2].BookRunID] != 1 {
		t.Fatalf("successful books rerun: calls=%v", exec.calls)
	}
	var unexpected int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM book_runs WHERE run_id=? AND book_id IN (?,?) AND attempt>1`, run.ID, items[0].BookID, items[2].BookID).Scan(&unexpected); err != nil {
		t.Fatal(err)
	}
	if unexpected != 0 {
		t.Fatalf("A/C next attempts=%d want 0", unexpected)
	}
}
