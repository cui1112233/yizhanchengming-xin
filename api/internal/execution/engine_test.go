package execution

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type fakeStore struct {
	items      []BookRun
	claimed    []int64
	completed  []int64
	failed     map[int64]string
	finalized  []int64
}

func (f *fakeStore) StartRun(context.Context, int64) (bool, error) { return true, nil }
func (f *fakeStore) EnsureBookRuns(context.Context, int64) error { return nil }
func (f *fakeStore) RecoverExpiredBookRuns(context.Context, int64, time.Time) (int64, error) { return 0, nil }
func (f *fakeStore) ListPendingBookRuns(context.Context, int64, int) ([]BookRun, error) { return f.items, nil }
func (f *fakeStore) ClaimBookRun(_ context.Context, id int64, _ time.Time) (bool, error) { f.claimed = append(f.claimed, id); return true, nil }
func (f *fakeStore) CompleteBookRun(_ context.Context, id int64) error { f.completed = append(f.completed, id); return nil }
func (f *fakeStore) FailBookRun(_ context.Context, id int64, message string) error {
	if f.failed == nil { f.failed = map[int64]string{} }
	f.failed[id] = message
	return nil
}
func (f *fakeStore) FinalizeRun(_ context.Context, id int64) error { f.finalized = append(f.finalized, id); return nil }

type fakeExecutor struct { calls []int64 }
func (f *fakeExecutor) ExecuteBook(_ context.Context, bookID int64) error {
	f.calls = append(f.calls, bookID)
	if bookID == 102 { return errors.New("provider timeout") }
	return nil
}

func TestEngineIsolatesBookFailureAndContinuesBatch(t *testing.T) {
	store := &fakeStore{items: []BookRun{{ID: 1, BookID: 101}, {ID: 2, BookID: 102}, {ID: 3, BookID: 103}}}
	executor := &fakeExecutor{}
	engine := NewEngine(store, executor, func() time.Time { return time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC) })

	if err := engine.ExecuteRun(context.Background(), 71); err != nil { t.Fatalf("ExecuteRun: %v", err) }
	if !reflect.DeepEqual(executor.calls, []int64{101, 102, 103}) { t.Fatalf("executor calls = %v", executor.calls) }
	if !reflect.DeepEqual(store.completed, []int64{1, 3}) { t.Fatalf("completed = %v", store.completed) }
	if got := store.failed[2]; got != "provider timeout" { t.Fatalf("failed error = %q", got) }
	if !reflect.DeepEqual(store.finalized, []int64{71}) { t.Fatalf("finalized = %v", store.finalized) }
}
