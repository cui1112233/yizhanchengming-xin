package project

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var testProjectColumns = []string{"id", "owner_username", "name", "width", "height", "frame_rate", "idempotency_key", "created_at", "updated_at"}

func TestCreateReturnsExistingProjectForTheSameOwnerAndIdempotencyKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 11, 20, 7, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT id, owner_username, name, width, height, frame_rate, idempotency_key, created_at, updated_at FROM magic_cut_projects WHERE owner_username = \? AND idempotency_key = \? LIMIT 1`).
		WithArgs("owner-a", "create-1").
		WillReturnRows(sqlmock.NewRows(testProjectColumns).AddRow("project-existing", "owner-a", "我的工程", 1920, 1080, 25, "create-1", createdAt, createdAt))

	store := NewStore(db, func() time.Time { return createdAt }, func() string { return "must-not-be-used" })
	got, err := store.Create(context.Background(), "owner-a", CreateInput{Name: "重复提交", Width: 1080, Height: 1920, FrameRate: 30, IdempotencyKey: "create-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "project-existing" || got.Name != "我的工程" {
		t.Fatalf("project = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreatePersistsANewProjectInsideTheOwnerBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 11, 20, 7, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT id, owner_username, name, width, height, frame_rate, idempotency_key, created_at, updated_at FROM magic_cut_projects WHERE owner_username = \? AND idempotency_key = \? LIMIT 1`).
		WithArgs("owner-a", "create-2").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO magic_cut_projects`).
		WithArgs("project-new", "owner-a", "竖屏工程", 1080, 1920, 30, "create-2", createdAt, createdAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	store := NewStore(db, func() time.Time { return createdAt }, func() string { return "project-new" })
	got, err := store.Create(context.Background(), "owner-a", CreateInput{Name: "竖屏工程", Width: 1080, Height: 1920, FrameRate: 30, IdempotencyKey: "create-2"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "project-new" || got.Owner != "owner-a" {
		t.Fatalf("project = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListAlwaysScopesTheQueryToTheAuthenticatedOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT id, owner_username, name, width, height, frame_rate, idempotency_key, created_at, updated_at FROM magic_cut_projects WHERE owner_username = \? ORDER BY updated_at DESC`).
		WithArgs("owner-b").
		WillReturnRows(sqlmock.NewRows(testProjectColumns))

	store := NewStore(db, time.Now, func() string { return "unused" })
	projects, err := store.List(context.Background(), "owner-b")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Fatalf("projects = %+v", projects)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNewIDProducesOpaqueProjectIdentifiers(t *testing.T) {
	first := NewID()
	second := NewID()
	if !strings.HasPrefix(first, "mcp_") || len(first) != 36 || first == second {
		t.Fatalf("first=%q second=%q", first, second)
	}
}
