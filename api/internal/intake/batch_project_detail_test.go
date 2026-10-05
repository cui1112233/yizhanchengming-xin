package intake

import (
	"context"
	"database/sql"
	"testing"
)

type batchProjectDetailReader interface {
	GetBatchProject(context.Context, int64) (BatchProject, error)
}

func TestMySQLStoreProvidesBatchProjectDetailReader(t *testing.T) {
	var db *sql.DB
	store := NewMySQLStore(db)
	if _, ok := any(store).(batchProjectDetailReader); !ok {
		t.Fatalf("MySQLStore must implement GetBatchProject for the Batch Factory detail workbench")
	}
}
