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
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO intakes (name, status) VALUES (?, ?)")).
		WithArgs("知乎+黑岩", StatusPending).
		WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?)")).
		WithArgs(int64(11), int64(7), nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	created, err := store.CreateOwnedIntake(context.Background(), "知乎+黑岩", ActorScope{UserID: 7})
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

func TestMySQLStoreUpsertBookIsIdempotentWithinIntakeAndPersistsExecutionState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	book := Book{
		IntakeID:       11,
		Source:         "知乎",
		PlatformID:     "15",
		ExternalBookID: "1001",
		Title:          "林子深处有声音",
		BodyRef:        "",
		OriginalText:   "第一章\n正文",
		Category:       "女生言情",
		Genre:          "8",
		Gender:         "女频",
		GenderSource:   "121_category",
		Style:          "情感",
		Status:         BookStatusFetched,
	}

	query := "INSERT INTO books (intake_id, source, platform_id, external_book_id, title, body_ref, original_text, category, genre, gender, gender_source, style, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), platform_id = VALUES(platform_id), title = VALUES(title), body_ref = VALUES(body_ref), original_text = VALUES(original_text), category = VALUES(category), genre = VALUES(genre), gender = VALUES(gender), gender_source = VALUES(gender_source), style = VALUES(style), status = VALUES(status), error_message = VALUES(error_message)"
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(book.IntakeID, book.Source, book.PlatformID, book.ExternalBookID, book.Title, book.BodyRef, book.OriginalText, book.Category, book.Genre, book.Gender, book.GenderSource, book.Style, book.Status, "").
		WillReturnResult(sqlmock.NewResult(31, 1))

	got, err := store.UpsertBook(context.Background(), book)
	if err != nil {
		t.Fatalf("UpsertBook: %v", err)
	}
	if got.ID != 31 || got.PlatformID != "15" || got.OriginalText == "" || got.GenderSource != "121_category" {
		t.Fatalf("got = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListBooksAndUpdateIntakeStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	query := "SELECT id, intake_id, source, platform_id, external_book_id, title, body_ref, original_text, category, genre, gender, gender_source, style, status, error_message, created_at, updated_at FROM books WHERE intake_id = ? ORDER BY id ASC"
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "source", "platform_id", "external_book_id", "title", "body_ref", "original_text", "category", "genre", "gender", "gender_source", "style", "status", "error_message", "created_at", "updated_at"}).
			AddRow(31, 11, "番茄付费", "2", "2001", "测试书", "", "正文", "男生生活", "8", "男频", "121_category", "现代通用", BookStatusFetched, "", now, now))

	books, err := store.ListBooks(context.Background(), 11)
	if err != nil {
		t.Fatalf("ListBooks: %v", err)
	}
	if len(books) != 1 || books[0].PlatformID != "2" || books[0].OriginalText != "正文" {
		t.Fatalf("books = %+v", books)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE intakes SET status = ? WHERE id = ?")).
		WithArgs(StatusCompleted, int64(11)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.UpdateIntakeStatus(context.Background(), 11, StatusCompleted); err != nil {
		t.Fatalf("UpdateIntakeStatus: %v", err)
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
	projectQuery := "INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = IF(archived_at IS NULL, VALUES(name), name)"
	mock.ExpectExec(regexp.QuoteMeta(projectQuery)).
		WithArgs(project.IntakeID, project.Name).
		WillReturnResult(sqlmock.NewResult(51, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ?")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))

	createdProject, err := store.CreateBatchProject(context.Background(), project)
	if err != nil {
		t.Fatalf("CreateBatchProject: %v", err)
	}
	if createdProject.ID != 51 {
		t.Fatalf("project id = %d, want 51", createdProject.ID)
	}

	runAt := time.Date(2026, 10, 6, 8, 30, 0, 0, time.UTC)
	run := Run{BatchProjectID: 51, RunAt: runAt, Status: RunStatusPending}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO runs (batch_project_id, run_at, status) VALUES (?, ?, ?)")).
		WithArgs(run.BatchProjectID, run.RunAt, run.Status).
		WillReturnResult(sqlmock.NewResult(71, 1))
	mock.ExpectCommit()

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

type batchProjectLister interface {
	ListBatchProjects(context.Context, BatchProjectListQuery) (BatchProjectPage, error)
}

func TestMySQLStoreListsBatchProjectsFromDatabase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	store := NewMySQLStore(db)
	lister, ok := any(store).(batchProjectLister)
	if !ok {
		t.Fatalf("MySQLStore must implement ListBatchProjects")
	}

	now := time.Date(2026, 10, 5, 5, 0, 0, 0, time.UTC)
	mock.ExpectQuery("^SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("^SELECT bp.id").WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}).
			AddRow(52, 12, "点众批次", "点众", 2, "女频", "情感", RunStatusPending, 0, now, now, nil).
			AddRow(51, 11, "知乎批次", "知乎", 1, "男频", "悬疑", RunStatusRunning, 0, now, now, nil))

	page, err := lister.ListBatchProjects(context.Background(), BatchProjectListQuery{Elevated: true, Archived: BatchProjectArchivedAll, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("ListBatchProjects: %v", err)
	}
	if len(page.Projects) != 2 || page.Projects[0].ID != 52 || page.Projects[0].Name != "点众批次" || page.Projects[1].ID != 51 {
		t.Fatalf("projects = %+v", page.Projects)
	}
	if page.Projects[0].BookCount != 2 || page.Projects[0].RunStatus != RunStatusPending || len(page.Projects[0].Sources) != 1 || page.Projects[0].Sources[0] != "点众" {
		t.Fatalf("project summary = %+v", page.Projects[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
