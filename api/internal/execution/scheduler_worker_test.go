package execution

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type fakeDueRunStore struct { ids []int64 }
func (f *fakeDueRunStore) ListRunnableRunIDs(context.Context, time.Time, time.Time, int) ([]int64, error) { return f.ids, nil }

type fakeQueue struct { enqueued []int64; next int64 }
func (f *fakeQueue) Enqueue(_ context.Context, runID int64) error { f.enqueued = append(f.enqueued, runID); return nil }
func (f *fakeQueue) Dequeue(context.Context) (int64, error) { return f.next, nil }

type fakeLocker struct { acquired bool; keys []string; released int }
func (f *fakeLocker) Acquire(_ context.Context, key string, _ time.Duration) (func(context.Context) error, bool, error) {
	f.keys = append(f.keys, key)
	if !f.acquired { return func(context.Context) error { return nil }, false, nil }
	return func(context.Context) error { f.released++; return nil }, true, nil
}

type fakeRunner struct { runs []int64 }
func (f *fakeRunner) ExecuteRun(_ context.Context, runID int64) error { f.runs = append(f.runs, runID); return nil }

func TestSchedulerEnqueuesDueAndStaleRunsFromDurableStore(t *testing.T) {
	now := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	store := &fakeDueRunStore{ids: []int64{71, 72}}
	queue := &fakeQueue{}
	scheduler := NewScheduler(store, queue, func() time.Time { return now })
	if err := scheduler.Tick(context.Background()); err != nil { t.Fatalf("Tick: %v", err) }
	if !reflect.DeepEqual(queue.enqueued, []int64{71, 72}) { t.Fatalf("enqueued=%v", queue.enqueued) }
}

func TestWorkerUsesDistributedRunLockBeforeExecution(t *testing.T) {
	queue := &fakeQueue{next: 71}
	locker := &fakeLocker{acquired: true}
	runner := &fakeRunner{}
	worker := NewWorker(queue, locker, runner)
	if err := worker.RunOnce(context.Background()); err != nil { t.Fatalf("RunOnce: %v", err) }
	if !reflect.DeepEqual(locker.keys, []string{"batch-factory:run:71"}) { t.Fatalf("keys=%v", locker.keys) }
	if !reflect.DeepEqual(runner.runs, []int64{71}) { t.Fatalf("runs=%v", runner.runs) }
	if locker.released != 1 { t.Fatalf("released=%d", locker.released) }
}

func TestWorkerSkipsDuplicateWhenRunLockIsHeldElsewhere(t *testing.T) {
	queue := &fakeQueue{next: 71}
	locker := &fakeLocker{acquired: false}
	runner := &fakeRunner{}
	worker := NewWorker(queue, locker, runner)
	if err := worker.RunOnce(context.Background()); err != nil { t.Fatalf("RunOnce: %v", err) }
	if len(runner.runs) != 0 { t.Fatalf("duplicate execution: %v", runner.runs) }
}
