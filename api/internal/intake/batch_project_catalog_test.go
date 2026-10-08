package intake

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreListsOwnedActiveBatchProjectsWithSQLScopedCountAndPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	countSQL := "SELECT COUNT(*) FROM batch_projects bp WHERE bp.archived_at IS NULL AND EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id = bp.id AND (o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?))) AND (LOWER(bp.name) LIKE ? OR EXISTS (SELECT 1 FROM books qb WHERE qb.intake_id = bp.intake_id AND (LOWER(qb.title) LIKE ? OR LOWER(qb.external_book_id) LIKE ?))) AND EXISTS (SELECT 1 FROM books sb WHERE sb.intake_id = bp.intake_id AND sb.source = ?) AND COALESCE((SELECT sr.status FROM runs sr WHERE sr.batch_project_id = bp.id ORDER BY sr.run_at DESC, sr.id DESC LIMIT 1), '') = ?"
	mock.ExpectQuery(regexp.QuoteMeta(countSQL)).WithArgs(int64(7), int64(12), int64(12), "%hero%", "%hero%", "%hero%", "知乎", "failed").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	pageSQL := "SELECT bp.id, bp.intake_id, bp.name, COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|') FROM books b WHERE b.intake_id = bp.intake_id), ''), (SELECT COUNT(*) FROM books bc WHERE bc.intake_id = bp.intake_id), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bg.gender), '') ORDER BY bg.gender SEPARATOR '|') FROM books bg WHERE bg.intake_id = bp.intake_id), ''), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bs.style), '') ORDER BY bs.style SEPARATOR '|') FROM books bs WHERE bs.intake_id = bp.intake_id), ''), COALESCE((SELECT lr.status FROM runs lr WHERE lr.batch_project_id = bp.id ORDER BY lr.run_at DESC, lr.id DESC LIMIT 1), ''), COALESCE((SELECT COUNT(*) FROM book_runs br WHERE br.run_id = (SELECT fr.id FROM runs fr WHERE fr.batch_project_id = bp.id ORDER BY fr.run_at DESC, fr.id DESC LIMIT 1) AND br.attempt = (SELECT MAX(br_latest.attempt) FROM book_runs br_latest WHERE br_latest.run_id = br.run_id AND br_latest.book_id = br.book_id) AND br.status IN ('failed','retryable_failed')), 0), bp.created_at, bp.updated_at, bp.archived_at FROM batch_projects bp WHERE bp.archived_at IS NULL AND EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id = bp.id AND (o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?))) AND (LOWER(bp.name) LIKE ? OR EXISTS (SELECT 1 FROM books qb WHERE qb.intake_id = bp.intake_id AND (LOWER(qb.title) LIKE ? OR LOWER(qb.external_book_id) LIKE ?))) AND EXISTS (SELECT 1 FROM books sb WHERE sb.intake_id = bp.intake_id AND sb.source = ?) AND COALESCE((SELECT sr.status FROM runs sr WHERE sr.batch_project_id = bp.id ORDER BY sr.run_at DESC, sr.id DESC LIMIT 1), '') = ? ORDER BY bp.updated_at DESC, bp.id DESC LIMIT ? OFFSET ?"
	now := time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(pageSQL)).WithArgs(int64(7), int64(12), int64(12), "%hero%", "%hero%", "%hero%", "知乎", "failed", 2, 2).WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}).
		AddRow(51, 11, "Hero project", "知乎", 3, "女频", "情感", "failed", 2, now, now, nil))

	page, err := store.ListBatchProjects(context.Background(), BatchProjectListQuery{
		UserID: 7, TeamID: 12, Query: "Hero", Source: "知乎", Status: RunStatusFailed,
		Archived: BatchProjectArchivedActive, Page: 2, Limit: 2, Sort: BatchProjectSortUpdatedDesc,
	})
	if err != nil {
		t.Fatalf("ListBatchProjects: %v", err)
	}
	if page.Total != 3 || page.Page != 2 || page.Limit != 2 || len(page.Projects) != 1 {
		t.Fatalf("page = %+v", page)
	}
	project := page.Projects[0]
	if project.ID != 51 || project.FailureCount != 2 || project.RunStatus != RunStatusFailed || project.ArchivedAt != nil {
		t.Fatalf("project = %+v", project)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreElevatedListDoesNotUseOwnershipAndNameSortIsStable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	countSQL := "SELECT COUNT(*) FROM batch_projects bp"
	mock.ExpectQuery(regexp.QuoteMeta(countSQL)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	pageSQL := "SELECT bp.id, bp.intake_id, bp.name, COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|') FROM books b WHERE b.intake_id = bp.intake_id), ''), (SELECT COUNT(*) FROM books bc WHERE bc.intake_id = bp.intake_id), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bg.gender), '') ORDER BY bg.gender SEPARATOR '|') FROM books bg WHERE bg.intake_id = bp.intake_id), ''), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bs.style), '') ORDER BY bs.style SEPARATOR '|') FROM books bs WHERE bs.intake_id = bp.intake_id), ''), COALESCE((SELECT lr.status FROM runs lr WHERE lr.batch_project_id = bp.id ORDER BY lr.run_at DESC, lr.id DESC LIMIT 1), ''), COALESCE((SELECT COUNT(*) FROM book_runs br WHERE br.run_id = (SELECT fr.id FROM runs fr WHERE fr.batch_project_id = bp.id ORDER BY fr.run_at DESC, fr.id DESC LIMIT 1) AND br.attempt = (SELECT MAX(br_latest.attempt) FROM book_runs br_latest WHERE br_latest.run_id = br.run_id AND br_latest.book_id = br.book_id) AND br.status IN ('failed','retryable_failed')), 0), bp.created_at, bp.updated_at, bp.archived_at FROM batch_projects bp ORDER BY bp.name ASC, bp.id ASC LIMIT ? OFFSET ?"
	mock.ExpectQuery(regexp.QuoteMeta(pageSQL)).WithArgs(20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}))

	_, err = store.ListBatchProjects(context.Background(), BatchProjectListQuery{Elevated: true, Archived: BatchProjectArchivedAll, Page: 1, Limit: 20, Sort: BatchProjectSortNameAsc})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreTeamZeroCannotMatchTeamOwnership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	countSQL := "SELECT COUNT(*) FROM batch_projects bp WHERE bp.archived_at IS NOT NULL AND EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id = bp.id AND (o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?)))"
	mock.ExpectQuery(regexp.QuoteMeta(countSQL)).WithArgs(int64(7), int64(0), int64(0)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	pageSQL := "SELECT bp.id, bp.intake_id, bp.name, COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|') FROM books b WHERE b.intake_id = bp.intake_id), ''), (SELECT COUNT(*) FROM books bc WHERE bc.intake_id = bp.intake_id), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bg.gender), '') ORDER BY bg.gender SEPARATOR '|') FROM books bg WHERE bg.intake_id = bp.intake_id), ''), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bs.style), '') ORDER BY bs.style SEPARATOR '|') FROM books bs WHERE bs.intake_id = bp.intake_id), ''), COALESCE((SELECT lr.status FROM runs lr WHERE lr.batch_project_id = bp.id ORDER BY lr.run_at DESC, lr.id DESC LIMIT 1), ''), COALESCE((SELECT COUNT(*) FROM book_runs br WHERE br.run_id = (SELECT fr.id FROM runs fr WHERE fr.batch_project_id = bp.id ORDER BY fr.run_at DESC, fr.id DESC LIMIT 1) AND br.attempt = (SELECT MAX(br_latest.attempt) FROM book_runs br_latest WHERE br_latest.run_id = br.run_id AND br_latest.book_id = br.book_id) AND br.status IN ('failed','retryable_failed')), 0), bp.created_at, bp.updated_at, bp.archived_at FROM batch_projects bp WHERE bp.archived_at IS NOT NULL AND EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id = bp.id AND (o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?))) ORDER BY bp.updated_at DESC, bp.id DESC LIMIT ? OFFSET ?"
	mock.ExpectQuery(regexp.QuoteMeta(pageSQL)).WithArgs(int64(7), int64(0), int64(0), 20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}))

	_, err = store.ListBatchProjects(context.Background(), BatchProjectListQuery{UserID: 7, Archived: BatchProjectArchivedArchived, Page: 1, Limit: 20, Sort: BatchProjectSortUpdatedDesc})
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
