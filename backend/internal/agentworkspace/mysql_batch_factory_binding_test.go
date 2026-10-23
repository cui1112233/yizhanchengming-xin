package agentworkspace

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLBatchFactoryBindingPersistsOwnerScopedOrigins(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	binding := BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: "project-1", CreatedAt: now, UpdatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT owner_username, batch_id, project_id, created_at, updated_at FROM agent_batch_factory_bindings WHERE owner_username = ? AND batch_id = ?")).
		WithArgs("alice", "batch-1").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_batch_factory_bindings (owner_username, batch_id, project_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")).
		WithArgs("alice", "batch-1", "project-1", now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := store.BindBatchFactoryProject(context.Background(), binding); err != nil {
		t.Fatal(err)
	}

	origin := BatchFactoryNodeOrigin{ProjectID: "project-1", NodeID: "image-1", BatchID: "batch-1", BookID: "book-1", ResourceType: "book_asset", ResourceID: "asset-1", VersionID: "version-1"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_batch_factory_bindings WHERE owner_username = ? AND batch_id = ? AND project_id = ?")).
		WithArgs("alice", "batch-1", "project-1").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_batch_factory_node_origins (project_id, node_id, owner_username, batch_id, book_id, resource_type, resource_id, version_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE batch_id = VALUES(batch_id), book_id = VALUES(book_id), resource_type = VALUES(resource_type), resource_id = VALUES(resource_id), version_id = VALUES(version_id)")).
		WithArgs("project-1", "image-1", "alice", "batch-1", "book-1", "book_asset", "asset-1", "version-1").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.BindBatchFactoryNode(context.Background(), "alice", origin); err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, node_id, batch_id, book_id, resource_type, resource_id, version_id FROM agent_batch_factory_node_origins WHERE owner_username = ? AND project_id = ? AND book_id = ? ORDER BY node_id")).
		WithArgs("alice", "project-1", "book-1").WillReturnRows(sqlmock.NewRows([]string{"project_id", "node_id", "batch_id", "book_id", "resource_type", "resource_id", "version_id"}).
		AddRow("project-1", "image-1", "batch-1", "book-1", "book_asset", "asset-1", "version-1"))
	origins, err := store.BatchFactoryNodeOrigins(context.Background(), "alice", "project-1", "book-1")
	if err != nil || len(origins) != 1 || origins[0] != origin {
		t.Fatalf("origins=%+v err=%v", origins, err)
	}

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM agent_batch_factory_node_origins WHERE owner_username = ? AND project_id = ? AND node_id = ?")).
		WithArgs("alice", "project-1", "image-1").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.RemoveBatchFactoryNodeBinding(context.Background(), "alice", "project-1", "image-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
