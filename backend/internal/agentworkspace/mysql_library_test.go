package agentworkspace

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLLibrarySavesReferenceOnlyAndListsByOwnerFolder(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 24, 12, 30, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_library_folders (id, owner_username, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")).
		WithArgs("folder-1", "alice", "角色参考", now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	folder, err := store.CreateLibraryFolder(context.Background(), "alice", CreateLibraryFolderInput{ID: "folder-1", Name: "角色参考", CreatedAt: now})
	if err != nil || folder.ID != "folder-1" {
		t.Fatalf("folder=%+v err=%v", folder, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM agent_library_folders WHERE id = ? AND owner_username = ?")).
		WithArgs("folder-1", "alice").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_library_items (id, folder_id, owner_username, source_project_id, source_asset_id, kind, name, tags_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")).
		WithArgs("item-1", "folder-1", "alice", "project-source", "asset-source", "image", "女主参考.png", `["女主"]`, now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	item, err := store.SaveLibraryItem(context.Background(), "alice", SaveLibraryItemInput{ID: "item-1", FolderID: "folder-1", SourceProjectID: "project-source", SourceAssetID: "asset-source", Kind: "image", Name: "女主参考.png", Tags: []string{"女主"}, CreatedAt: now})
	if err != nil || item.SourceAssetID != "asset-source" {
		t.Fatalf("item=%+v err=%v", item, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, folder_id, source_project_id, source_asset_id, kind, name, tags_json, created_at, updated_at FROM agent_library_items WHERE owner_username = ? AND folder_id = ? AND kind = ? AND (LOWER(name) LIKE ? OR LOWER(tags_json) LIKE ?) ORDER BY updated_at DESC, id DESC")).
		WithArgs("alice", "folder-1", "image", "%女主%", "%女主%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "folder_id", "source_project_id", "source_asset_id", "kind", "name", "tags_json", "created_at", "updated_at"}).AddRow("item-1", "folder-1", "project-source", "asset-source", "image", "女主参考.png", `["女主"]`, now, now))
	items, err := store.ListLibraryItems(context.Background(), "alice", "folder-1", LibraryItemFilter{Kind: "image", Query: "女主"})
	if err != nil || len(items) != 1 || items[0].SourceProjectID != "project-source" {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
