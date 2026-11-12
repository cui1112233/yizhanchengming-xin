package aicut

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestAssetFolderOwnerProjectIsolationAndSafeDelete(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	now := time.Now().UTC()
	for _, owner := range []string{"alice", "bob"} {
		for _, project := range []string{"project-a", "project-b"} {
			if err := s.CreateProject(ctx, owner, Project{ID: project, Title: project, CreatedAt: now, UpdatedAt: now}); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, owner := range []string{"alice", "bob"} {
		if err := s.CreateAssetFolder(ctx, owner, AssetFolder{ID: "folder-1", ProjectID: "project-a", Name: "素材", CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		if err := s.CreateAsset(ctx, owner, Asset{ID: "asset-1", ProjectID: "project-a", Kind: "image", Name: "a.png", ContentType: "image/png", ObjectKey: "a.png", ByteSize: 1, CreatedAt: now}); err != nil {
			t.Fatal(err)
		}
		if err := s.MoveAssetToFolder(ctx, owner, "project-a", "asset-1", "folder-1"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CreateAssetFolder(ctx, "alice", AssetFolder{ID: "other-folder", ProjectID: "project-b", Name: "另一个工程", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := s.MoveAssetToFolder(ctx, "alice", "project-a", "asset-1", "other-folder"); !errors.Is(err, ErrAssetFolderNotFound) {
		t.Fatalf("cross project move=%v", err)
	}
	if err := s.RenameAssetFolder(ctx, "alice", "project-a", "folder-1", "改名"); err != nil {
		t.Fatal(err)
	}
	bob, _ := s.ListAssetFolders(ctx, "bob", "project-a")
	if len(bob) != 1 || bob[0].Name != "素材" {
		t.Fatalf("other owner changed=%+v", bob)
	}
	if err := s.DeleteAssetFolder(ctx, "alice", "project-a", "folder-1"); err != nil {
		t.Fatal(err)
	}
	assignments, _ := s.AssetFolderAssignments(ctx, "alice", "project-a")
	if len(assignments) != 0 {
		t.Fatalf("deleted folder assignments=%v", assignments)
	}
	if assets, err := s.ListAssets(ctx, "alice", "project-a"); err != nil || len(assets) != 1 {
		t.Fatalf("delete removed media: %+v %v", assets, err)
	}
	assignments, _ = s.AssetFolderAssignments(ctx, "bob", "project-a")
	if assignments["asset-1"] != "folder-1" {
		t.Fatalf("other owner lost folder=%v", assignments)
	}
	if _, err := s.ListAssetFolders(ctx, "mallory", "project-a"); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("owner leak=%v", err)
	}
	if err := s.MoveAssetToFolder(ctx, "bob", "project-a", "asset-1", ""); err != nil {
		t.Fatal(err)
	}
	assignments, _ = s.AssetFolderAssignments(ctx, "bob", "project-a")
	if len(assignments) != 0 {
		t.Fatal(assignments)
	}
}
func TestAssetFolderRejectsInvalidNames(t *testing.T) {
	s := NewMemoryStore()
	for _, name := range []string{"", " \t", "bad\nname"} {
		if err := s.CreateAssetFolder(t.Context(), "alice", AssetFolder{ID: "f", ProjectID: "p", Name: name}); !errors.Is(err, ErrInvalidAssetFolder) {
			t.Fatalf("name %q=%v", name, err)
		}
	}
}

func expectFolderProject(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery("SELECT id, title, aspect_ratio, background_mode, background_color, created_at, updated_at FROM aicut_projects").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "ratio", "mode", "color", "created", "updated"}).AddRow("p", "工程", "9:16", "solid", "#000000", now, now))
}

func TestMySQLAssetFolderPersistenceAcrossStoreInstances(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := t.Context()
	now := time.Now().UTC()
	s := NewMySQLStore(db)
	expectFolderProject(mock, now)
	mock.ExpectExec("INSERT INTO aicut_asset_folders").WithArgs("f", "p", "alice", "片头", now).WillReturnResult(sqlmock.NewResult(1, 1))
	if err = s.CreateAssetFolder(ctx, "alice", AssetFolder{ID: "f", ProjectID: "p", Name: "片头", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	expectFolderProject(mock, now)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM aicut_assets").WithArgs("a", "p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("a"))
	mock.ExpectQuery("SELECT id FROM aicut_asset_folders").WithArgs("f", "p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("f"))
	mock.ExpectExec("INSERT INTO aicut_asset_folder_members").WithArgs("a", "p", "alice", "f").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err = s.MoveAssetToFolder(ctx, "alice", "p", "a", "f"); err != nil {
		t.Fatal(err)
	}
	// A fresh store reads persisted folder metadata and membership from SQL.
	reopened := NewMySQLStore(db)
	expectFolderProject(mock, now)
	mock.ExpectQuery("SELECT id, name, created_at FROM aicut_asset_folders").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created"}).AddRow("f", "片头", now))
	folders, err := reopened.ListAssetFolders(ctx, "alice", "p")
	if err != nil || len(folders) != 1 || folders[0].Name != "片头" {
		t.Fatalf("folders=%+v %v", folders, err)
	}
	expectFolderProject(mock, now)
	mock.ExpectQuery("SELECT asset_id, folder_id FROM aicut_asset_folder_members").WithArgs("p", "alice").WillReturnRows(sqlmock.NewRows([]string{"asset", "folder"}).AddRow("a", "f"))
	assignments, err := reopened.AssetFolderAssignments(ctx, "alice", "p")
	if err != nil || assignments["a"] != "f" {
		t.Fatalf("membership=%v %v", assignments, err)
	}
	expectFolderProject(mock, now)
	mock.ExpectExec("DELETE FROM aicut_asset_folders").WithArgs("f", "p", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	if err = reopened.DeleteAssetFolder(ctx, "alice", "p", "f"); err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
