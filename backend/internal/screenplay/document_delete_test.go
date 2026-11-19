package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func checkSeedVersions(t *testing.T, store Store) (string, string, string) {
	t.Helper()
	ctx := context.Background()
	owner, projectID := "alice", "proj-check"
	if _, err := store.CreateProject(ctx, Project{ID: projectID, OwnerUsername: owner, Title: "版本择优"}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	first, err := SaveManualDocument(ctx, store, owner, projectID, "# 第1集\n\n第一场\n林川走进巷口。")
	if err != nil {
		t.Fatalf("save v1: %v", err)
	}
	_ = first
	second, err := SaveManualDocument(ctx, store, owner, projectID, "# 第1集\n\n第一场\n林川走进巷口，雨下起来了。")
	if err != nil {
		t.Fatalf("save v2: %v", err)
	}
	return owner, projectID, second.ID
}

func TestDeleteDocumentRejectsActiveVersion(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	owner, projectID, _ := checkSeedVersions(t, store)
	documents, err := store.ListDocuments(ctx, owner, projectID)
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	var activeID string
	for _, doc := range documents {
		if doc.Status == "active" {
			activeID = doc.ID
		}
	}
	if activeID == "" {
		t.Fatalf("expected an active document")
	}
	err = DeleteDocument(ctx, store, owner, projectID, activeID)
	if err == nil || !strings.Contains(err.Error(), "当前版本不可删除") {
		t.Fatalf("expected active version rejection, got %v", err)
	}
}

func TestDeleteDocumentRejectsEmptyID(t *testing.T) {
	store := NewMemoryStore()
	if err := DeleteDocument(context.Background(), store, "alice", "proj", "  "); err == nil || !strings.Contains(err.Error(), "剧本版本不能为空") {
		t.Fatalf("expected empty id rejection, got %v", err)
	}
}

func TestDeleteDocumentUnknownIDReturnsNotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	owner, projectID, _ := checkSeedVersions(t, store)
	err := DeleteDocument(ctx, store, owner, projectID, "doc-missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteDocumentRemovesHistoricalVersion(t *testing.T) {
	db, expect, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := &MySQLStore{db: db}
	ctx := context.Background()

	expect.ExpectBegin()
	for _, table := range []string{"screenplay_shots", "screenplay_characters", "screenplay_scenes"} {
		expect.ExpectExec("DELETE FROM "+table+" WHERE document_id = .*").WithArgs("doc-old", "proj-check", "alice").WillReturnResult(sqlmock.NewResult(0, 2))
	}
	expect.ExpectExec("DELETE FROM screenplay_documents WHERE id = .*").WithArgs("doc-old", "proj-check", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	expect.ExpectCommit()

	if err := store.DeleteDocument(ctx, "alice", "proj-check", "doc-old"); err != nil {
		t.Fatalf("mysql delete: %v", err)
	}
	if err := expect.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteDocumentMissingRowRollsBack(t *testing.T) {
	db, expect, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := &MySQLStore{db: db}
	ctx := context.Background()

	expect.ExpectBegin()
	for _, table := range []string{"screenplay_shots", "screenplay_characters", "screenplay_scenes"} {
		expect.ExpectExec("DELETE FROM "+table+" WHERE document_id = .*").WithArgs("doc-missing", "proj-check", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	}
	expect.ExpectExec("DELETE FROM screenplay_documents WHERE id = .*").WithArgs("doc-missing", "proj-check", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	expect.ExpectRollback()

	err = store.DeleteDocument(ctx, "alice", "proj-check", "doc-missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on missing row, got %v", err)
	}
	if err := expect.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}
