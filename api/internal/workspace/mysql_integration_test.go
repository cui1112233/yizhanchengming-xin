package workspace

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// This regression uses an existing migrated staging database read-only. It
// catches collation conflicts that sqlmock cannot execute across UNION arms.
func TestMySQLWorkspaceProjectionsExecuteAcrossMixedMigrationCollations(t *testing.T) {
	dsn := os.Getenv("WORKSPACE_MYSQL_DSN")
	if dsn == "" {
		t.Skip("WORKSPACE_MYSQL_DSN not configured")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}

	store := NewMySQLStore(db)
	if _, err := store.ListRecentElevated(ctx, 20); err != nil {
		t.Fatalf("recent projection: %v", err)
	}
	if _, err := store.ListHistory(ctx, HistoryQuery{Elevated: true, Page: 1, Limit: 20, Archived: "all"}); err != nil {
		t.Fatalf("history projection: %v", err)
	}
}
