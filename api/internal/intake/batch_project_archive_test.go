package intake

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const activeBatchProjectSQL = "SELECT EXISTS(SELECT 1 FROM runs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM book_runs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM stage_runs sr JOIN book_runs br ON br.id = sr.book_run_id WHERE br.batch_project_id = ? AND sr.status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM shuihuo_media_tasks WHERE batch_project_id = ? AND status IN ('pending_executor','pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_production_jobs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_production_tasks vt JOIN video_production_jobs vj ON vj.id = vt.production_job_id WHERE vj.batch_project_id = ? AND vt.status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_merge_jobs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_merge_attempts va JOIN video_merge_jobs vj ON vj.id = va.merge_job_id WHERE vj.batch_project_id = ? AND va.status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM intakes i JOIN batch_projects bp ON bp.intake_id = i.id WHERE bp.id = ? AND i.status = 'running')"

func TestMySQLStoreArchivesProjectAtomicallyWithoutDeletingFacts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectQuery(regexp.QuoteMeta(activeBatchProjectSQL)).WithArgs(int64(51), int64(51), int64(51), int64(51), int64(51), int64(51), int64(51), int64(51), int64(51)).WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE batch_projects SET archived_at = UTC_TIMESTAMP(6), archived_by_user_id = ? WHERE id = ? AND archived_at IS NULL")).WithArgs(int64(7), int64(51)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.ArchiveBatchProject(context.Background(), 51, 7); err != nil {
		t.Fatalf("ArchiveBatchProject: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("archive must only update the project row: %v", err)
	}
}

func TestMySQLStoreRejectsArchiveWhileProjectHasActiveWork(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectQuery(regexp.QuoteMeta(activeBatchProjectSQL)).WithArgs(int64(51), int64(51), int64(51), int64(51), int64(51), int64(51), int64(51), int64(51), int64(51)).WillReturnRows(sqlmock.NewRows([]string{"active"}).AddRow(true))
	mock.ExpectRollback()

	if err := store.ArchiveBatchProject(context.Background(), 51, 7); !errors.Is(err, ErrBatchProjectActive) {
		t.Fatalf("err=%v, want ErrBatchProjectActive", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreArchiveAndRestoreAreIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectCommit()
	if err := store.ArchiveBatchProject(context.Background(), 51, 7); err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectCommit()
	if err := store.RestoreBatchProject(context.Background(), 51); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreRestoresArchivedProjectWithoutRestartingWork(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE batch_projects SET archived_at = NULL, archived_by_user_id = NULL WHERE id = ? AND archived_at IS NOT NULL")).WithArgs(int64(51)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.RestoreBatchProject(context.Background(), 51); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("restore must not restart runs or media: %v", err)
	}
}

func TestMySQLStoreArchivedStateDistinguishesMissingProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at IS NOT NULL FROM batch_projects WHERE id = ?")).WithArgs(int64(51)).WillReturnError(sql.ErrNoRows)

	_, err = store.IsBatchProjectArchived(context.Background(), 51)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err=%v, want sql.ErrNoRows", err)
	}
}

func TestMySQLStoreReadsArchiveStateThroughIntakeAlias(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM batch_projects WHERE intake_id = ? AND archived_at IS NOT NULL)")).WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"archived"}).AddRow(true))

	archived, err := store.IsIntakeBatchProjectArchived(context.Background(), 11)
	if err != nil || !archived {
		t.Fatalf("archived=%v err=%v", archived, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreDoesNotRenameArchivedProjectDuringPipelineUpsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 11, 0, 0, 0, time.UTC)
	query := "INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = IF(archived_at IS NULL, VALUES(name), name)"
	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(int64(11), "new name").WillReturnResult(sqlmock.NewResult(51, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ?")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))

	_, err = store.CreateBatchProject(context.Background(), BatchProject{IntakeID: 11, Name: "new name"})
	if !errors.Is(err, ErrBatchProjectArchived) {
		t.Fatalf("err=%v, want ErrBatchProjectArchived", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreDoesNotCreateRunAfterProjectIsArchived(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	runAt := time.Date(2026, 10, 9, 11, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(runAt))
	mock.ExpectRollback()

	_, err = store.CreateRun(context.Background(), Run{BatchProjectID: 51, RunAt: runAt, Status: RunStatusPending})
	if !errors.Is(err, ErrBatchProjectArchived) {
		t.Fatalf("err=%v, want ErrBatchProjectArchived", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreDoesNotCreateIdempotentRunAfterProjectIsArchived(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	runAt := time.Date(2026, 10, 9, 11, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(runAt))
	mock.ExpectRollback()

	_, err = store.CreateRunIdempotent(context.Background(), Run{BatchProjectID: 51, RunAt: runAt, Status: RunStatusPending}, "key")
	if !errors.Is(err, ErrBatchProjectArchived) {
		t.Fatalf("err=%v, want ErrBatchProjectArchived", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreDoesNotStartIntakeForArchivedProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT bp.archived_at FROM batch_projects bp WHERE bp.intake_id = ? FOR UPDATE")).WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
	mock.ExpectRollback()
	if err := store.UpdateIntakeStatus(context.Background(), 11, StatusRunning); !errors.Is(err, ErrBatchProjectArchived) {
		t.Fatalf("err=%v, want ErrBatchProjectArchived", err)
	}
}
