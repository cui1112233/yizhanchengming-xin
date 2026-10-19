//go:build mysql84

package storage

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQL84EmptyToLatest(t *testing.T) {
	dsn := os.Getenv("QIANTIE_TEST_MYSQL84_DSN")
	if dsn == "" {
		t.Skip("QIANTIE_TEST_MYSQL84_DSN not configured; run against a disposable MySQL 8.4 database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := RunMigrations(context.Background(), db, V11Migrations()); err != nil {
		t.Fatal(err)
	}
	if err := RunMigrations(context.Background(), db, V11Migrations()); err != nil {
		t.Fatalf("repeat startup: %v", err)
	}
}

func TestMySQL84AgentWorkspaceGroupMigrationEmptyAndRepeat(t *testing.T) {
	dsn := os.Getenv("QIANTIE_TEST_MYSQL84_DSN")
	if dsn == "" {
		t.Skip("QIANTIE_TEST_MYSQL84_DSN not configured; run against a disposable MySQL 8.4 database")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := RunMigrations(ctx, db, AgentWorkspaceMigrations()); err != nil {
		t.Fatalf("empty agent workspace migration: %v", err)
	}
	if err := RunMigrations(ctx, db, AgentWorkspaceMigrations()); err != nil {
		t.Fatalf("repeat agent workspace migration: %v", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns
WHERE table_schema = DATABASE() AND table_name = 'agent_canvas_nodes' AND column_name = 'group_id'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("agent_canvas_nodes.group_id count=%d, want 1", count)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.columns
WHERE table_schema = DATABASE() AND table_name = 'agent_canvases' AND column_name = 'appearance_json' AND data_type = 'json'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("agent_canvases.appearance_json count=%d, want 1", count)
	}
}
