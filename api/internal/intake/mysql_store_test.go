package intake

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreCreateAndReadIntake(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO intakes (name, status) VALUES (?, ?)")).
		WithArgs("知乎+黑岩", StatusPending).
		WillReturnResult(sqlmock.NewResult(11, 1))

	created, err := store.CreateIntake(context.Background(), "知乎+黑岩")
	if err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	if created.ID != 11 || created.Status != StatusPending {
		t.Fatalf("created = %+v", created)
	}

	now := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, status, created_at, updated_at FROM intakes WHERE id = ?")).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at"}).
			AddRow(11, "知乎+黑岩", StatusPending, now, now))

	got, err := store.GetIntake(context.Background(), 11)
	if err != nil {
		t.Fatalf("GetIntake: %v", err)
	}
	if got.ID != 11 || got.Name != "知乎+黑岩" || got.Status != StatusPending {
		t.Fatalf("got = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreUpsertBookIsIdempotentWithinIntakeAndPersistsMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	book := Book{
		IntakeID:       11,
		Source:         "知乎",
		ExternalBookID: "book-1001",
		Title:          "林子深处有声音",
		BodyRef:        "tos://novels/book-1001.txt",
		Category:       "女频",
		Genre:          "现代言情",
		Gender:         "女频",
		GenderSource:   "category",
		Style:          "情感",
		Status:         BookStatusFetched,
	}

	query := "INSERT INTO books (intake_id, source, external_book_id, title, body_ref, category, genre, gender, gender_source, style, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), title = VALUES(title), body_ref = VALUES(body_ref), category = VALUES(category), genre = VALUES(genre), gender = VALUES(gender), gender_source = VALUES(gender_source), style = VALUES(style), status = VALUES(status), error_message = VALUES(error_message)"
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(book.IntakeID, book.Source, book.ExternalBookID, book.Title, book.BodyRef, book.Category, book.Genre, book.Gender, book.GenderSource, book.Style, book.Status, "").
		WillReturnResult(sqlmock.NewResult(31, 1))

	got, err := store.UpsertBook(context.Background(), book)
	if err != nil {
		t.Fatalf("UpsertBook: %v", err)
	}
	if got.ID != 31 || got.Gender != "女频" || got.GenderSource != "category" || got.Style != "情感" {
		t.Fatalf("got = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreCreateBatchProjectAndScheduledRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	project := BatchProject{IntakeID: 11, Name: "批量项目-知乎黑岩"}
	projectQuery := "INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = VALUES(name)"
	mock.ExpectExec(regexp.QuoteMeta(projectQuery)).
		WithArgs(project.IntakeID, project.Name).
		WillReturnResult(sqlmock.NewResult(51, 1))

	createdProject, err := store.CreateBatchProject(context.Background(), project)
	if err != nil {
		t.Fatalf("CreateBatchProject: %v", err)
	}
	if createdProject.ID != 51 {
		t.Fatalf("project id = %d, want 51", createdProject.ID)
	}

	runAt := time.Date(2026, 10, 6, 8, 30, 0, 0, time.UTC)
	run := Run{BatchProjectID: 51, RunAt: runAt, Status: RunStatusPending}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO runs (batch_project_id, run_at, status) VALUES (?, ?, ?)")).
		WithArgs(run.BatchProjectID, run.RunAt, run.Status).
		WillReturnResult(sqlmock.NewResult(71, 1))

	createdRun, err := store.CreateRun(context.Background(), run)
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if createdRun.ID != 71 || !createdRun.RunAt.Equal(runAt) || createdRun.Status != RunStatusPending {
		t.Fatalf("run = %+v", createdRun)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
