package intake

import (
	"context"
	"testing"
	"time"
)

type runExecutionStore interface {
	StartRun(context.Context, int64) (bool, error)
	EnsureBookRuns(context.Context, int64) error
	RecoverExpiredBookRuns(context.Context, int64, time.Time) (int64, error)
	RetryBookRun(context.Context, int64) (bool, error)
}

func TestMySQLStoreProvidesRunExecutionMethods(t *testing.T) {
	store := &MySQLStore{}
	if _, ok := any(store).(runExecutionStore); !ok {
		t.Fatal("MySQLStore must provide run execution methods")
	}
}
