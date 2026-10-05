package intake

import (
	"context"
	"testing"
	"time"
)

type bookRunWorkerStore interface {
	ListPendingBookRuns(context.Context, int64, int) ([]BookRun, error)
	ClaimBookRun(context.Context, int64, time.Time) (bool, error)
	CompleteBookRun(context.Context, int64) error
	FailBookRun(context.Context, int64, string) error
	FinalizeRun(context.Context, int64) error
}

func TestMySQLStoreProvidesBookRunWorkerOperations(t *testing.T) {
	store := &MySQLStore{}
	if _, ok := any(store).(bookRunWorkerStore); !ok {
		t.Fatal("MySQLStore must provide per-book worker operations")
	}
}
