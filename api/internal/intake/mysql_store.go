package intake

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrBatchProjectActive   = errors.New("batch project has active work")
	ErrBatchProjectArchived = errors.New("batch project is archived")
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

// CreateOwnedIntake commits the intake row and its access boundary atomically.
// Books deliberately remain outside this transaction so their existing
// partial-failure/status semantics are preserved.
func (s *MySQLStore) CreateOwnedIntake(ctx context.Context, name string, actor ActorScope) (Intake, error) {
	if s == nil || s.db == nil || actor.UserID <= 0 {
		return Intake{}, fmt.Errorf("create owned intake: authenticated owner is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Intake{}, fmt.Errorf("begin create intake: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, "INSERT INTO intakes (name, status) VALUES (?, ?)", name, StatusPending)
	if err != nil {
		return Intake{}, fmt.Errorf("create intake: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Intake{}, fmt.Errorf("read intake id: %w", err)
	}
	var teamID any
	if actor.TeamID > 0 {
		teamID = actor.TeamID
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?)", id, actor.UserID, teamID); err != nil {
		return Intake{}, fmt.Errorf("create intake ownership: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Intake{}, fmt.Errorf("commit create intake: %w", err)
	}
	return Intake{ID: id, Name: name, Status: StatusPending}, nil
}

func (s *MySQLStore) GetIntake(ctx context.Context, id int64) (Intake, error) {
	var value Intake
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, status, created_at, updated_at FROM intakes WHERE id = ?",
		id,
	).Scan(&value.ID, &value.Name, &value.Status, &value.CreatedAt, &value.UpdatedAt)
	if err != nil {
		return Intake{}, fmt.Errorf("get intake: %w", err)
	}
	return value, nil
}

func (s *MySQLStore) ListIntakes(ctx context.Context) ([]Intake, error) {
	const query = "SELECT id, name, status, created_at, updated_at FROM intakes ORDER BY id DESC LIMIT 100"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list intakes: %w", err)
	}
	defer rows.Close()

	result := make([]Intake, 0)
	for rows.Next() {
		var value Intake
		if err := rows.Scan(&value.ID, &value.Name, &value.Status, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan intake: %w", err)
		}
		result = append(result, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate intakes: %w", err)
	}
	return result, nil
}

// ListVisibleIntakes filters at the SQL boundary. Elevated callers have
// already passed the route capability check and may inspect legacy-unowned
// rows; ordinary users only receive owner/team rows.
func (s *MySQLStore) ListVisibleIntakes(ctx context.Context, userID, teamID int64, elevated bool) ([]Intake, error) {
	if elevated {
		return s.ListIntakes(ctx)
	}
	const query = "SELECT i.id, i.name, i.status, i.created_at, i.updated_at FROM intakes i JOIN auth_intake_ownership o ON o.intake_id = i.id WHERE o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?) ORDER BY i.id DESC LIMIT 100"
	rows, err := s.db.QueryContext(ctx, query, userID, teamID, teamID)
	if err != nil {
		return nil, fmt.Errorf("list visible intakes: %w", err)
	}
	defer rows.Close()
	result := make([]Intake, 0)
	for rows.Next() {
		var value Intake
		if err := rows.Scan(&value.ID, &value.Name, &value.Status, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan visible intake: %w", err)
		}
		result = append(result, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate visible intakes: %w", err)
	}
	return result, nil
}

func (s *MySQLStore) CanAccessIntake(ctx context.Context, intakeID, userID, teamID int64, elevated bool) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("intake access store unavailable")
	}
	var allowed bool
	if elevated {
		err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM intakes WHERE id = ?)", intakeID).Scan(&allowed)
		return allowed, err
	}
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM auth_intake_ownership WHERE intake_id = ? AND (owner_user_id = ? OR (? > 0 AND team_id IS NOT NULL AND team_id = ?)))", intakeID, userID, teamID, teamID).Scan(&allowed)
	return allowed, err
}

func (s *MySQLStore) UpdateIntakeStatus(ctx context.Context, id int64, status Status) error {
	if status != StatusRunning {
		result, err := s.db.ExecContext(ctx, "UPDATE intakes SET status = ? WHERE id = ?", status, id)
		if err != nil {
			return fmt.Errorf("update intake status: %w", err)
		}
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update intake status: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	err = tx.QueryRowContext(ctx, "SELECT bp.archived_at FROM batch_projects bp WHERE bp.intake_id = ? FOR UPDATE", id).Scan(&archivedAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("lock intake batch project: %w", err)
	}
	if archivedAt.Valid {
		return ErrBatchProjectArchived
	}
	result, err := tx.ExecContext(ctx, "UPDATE intakes SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("update intake status: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (s *MySQLStore) UpsertBook(ctx context.Context, book Book) (Book, error) {
	const query = "INSERT INTO books (intake_id, source, platform_id, external_book_id, title, body_ref, original_text, category, genre, gender, gender_source, style, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), platform_id = VALUES(platform_id), title = VALUES(title), body_ref = VALUES(body_ref), original_text = VALUES(original_text), category = VALUES(category), genre = VALUES(genre), gender = VALUES(gender), gender_source = VALUES(gender_source), style = VALUES(style), status = VALUES(status), error_message = VALUES(error_message)"
	result, err := s.db.ExecContext(ctx, query,
		book.IntakeID,
		book.Source,
		book.PlatformID,
		book.ExternalBookID,
		book.Title,
		book.BodyRef,
		book.OriginalText,
		book.Category,
		book.Genre,
		book.Gender,
		book.GenderSource,
		book.Style,
		book.Status,
		book.ErrorMessage,
	)
	if err != nil {
		return Book{}, fmt.Errorf("upsert book: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Book{}, fmt.Errorf("read book id: %w", err)
	}
	book.ID = id
	return book, nil
}

func (s *MySQLStore) ListBooks(ctx context.Context, intakeID int64) ([]Book, error) {
	const query = "SELECT id, intake_id, source, platform_id, external_book_id, title, body_ref, original_text, category, genre, gender, gender_source, style, status, error_message, created_at, updated_at FROM books WHERE intake_id = ? ORDER BY id ASC"
	rows, err := s.db.QueryContext(ctx, query, intakeID)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()
	books := make([]Book, 0)
	for rows.Next() {
		var book Book
		if err := rows.Scan(
			&book.ID,
			&book.IntakeID,
			&book.Source,
			&book.PlatformID,
			&book.ExternalBookID,
			&book.Title,
			&book.BodyRef,
			&book.OriginalText,
			&book.Category,
			&book.Genre,
			&book.Gender,
			&book.GenderSource,
			&book.Style,
			&book.Status,
			&book.ErrorMessage,
			&book.CreatedAt,
			&book.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate books: %w", err)
	}
	return books, nil
}

// UpdateBookOriginalText keeps Script Workspace edits in the canonical Intake
// book record instead of creating a second script document store.
func (s *MySQLStore) UpdateBookOriginalText(ctx context.Context, intakeID, bookID int64, text string) (Book, error) {
	result, err := s.db.ExecContext(ctx, "UPDATE books SET original_text = ? WHERE id = ? AND intake_id = ?", text, bookID, intakeID)
	if err != nil {
		return Book{}, fmt.Errorf("update book original text: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return Book{}, sql.ErrNoRows
	}
	rows, err := s.ListBooks(ctx, intakeID)
	if err != nil {
		return Book{}, err
	}
	for _, book := range rows {
		if book.ID == bookID {
			return book, nil
		}
	}
	return Book{}, sql.ErrNoRows
}

func (s *MySQLStore) CreateBatchProject(ctx context.Context, project BatchProject) (BatchProject, error) {
	const query = "INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = IF(archived_at IS NULL, VALUES(name), name)"
	result, err := s.db.ExecContext(ctx, query, project.IntakeID, project.Name)
	if err != nil {
		return BatchProject{}, fmt.Errorf("create batch project: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return BatchProject{}, fmt.Errorf("read batch project id: %w", err)
	}
	project.ID = id
	var archivedAt sql.NullTime
	if err := s.db.QueryRowContext(ctx, "SELECT archived_at FROM batch_projects WHERE id = ?", id).Scan(&archivedAt); err != nil {
		return BatchProject{}, fmt.Errorf("read batch project archive state: %w", err)
	}
	if archivedAt.Valid {
		return BatchProject{}, ErrBatchProjectArchived
	}
	return project, nil
}

func (s *MySQLStore) GetBatchProject(ctx context.Context, id int64) (BatchProject, error) {
	var project BatchProject
	err := s.db.QueryRowContext(ctx,
		"SELECT id, intake_id, name, created_at, updated_at, archived_at FROM batch_projects WHERE id = ?",
		id,
	).Scan(&project.ID, &project.IntakeID, &project.Name, &project.CreatedAt, &project.UpdatedAt, &project.ArchivedAt)
	if err != nil {
		return BatchProject{}, fmt.Errorf("get batch project: %w", err)
	}
	return project, nil
}

func (s *MySQLStore) ListBatchProjects(ctx context.Context, filter BatchProjectListQuery) (BatchProjectPage, error) {
	page := BatchProjectPage{Page: filter.Page, Limit: filter.Limit, Projects: []BatchProject{}}
	where, args := batchProjectListWhere(filter)
	countQuery := "SELECT COUNT(*) FROM batch_projects bp" + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&page.Total); err != nil {
		return BatchProjectPage{}, fmt.Errorf("count batch projects: %w", err)
	}
	order := " ORDER BY bp.updated_at DESC, bp.id DESC"
	if filter.Sort == BatchProjectSortNameAsc {
		order = " ORDER BY bp.name ASC, bp.id ASC"
	}
	const selectColumns = "SELECT bp.id, bp.intake_id, bp.name, COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|') FROM books b WHERE b.intake_id = bp.intake_id), ''), (SELECT COUNT(*) FROM books bc WHERE bc.intake_id = bp.intake_id), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bg.gender), '') ORDER BY bg.gender SEPARATOR '|') FROM books bg WHERE bg.intake_id = bp.intake_id), ''), COALESCE((SELECT GROUP_CONCAT(DISTINCT NULLIF(TRIM(bs.style), '') ORDER BY bs.style SEPARATOR '|') FROM books bs WHERE bs.intake_id = bp.intake_id), ''), COALESCE((SELECT lr.status FROM runs lr WHERE lr.batch_project_id = bp.id ORDER BY lr.run_at DESC, lr.id DESC LIMIT 1), ''), COALESCE((SELECT COUNT(*) FROM book_runs br WHERE br.run_id = (SELECT fr.id FROM runs fr WHERE fr.batch_project_id = bp.id ORDER BY fr.run_at DESC, fr.id DESC LIMIT 1) AND br.attempt = (SELECT MAX(br_latest.attempt) FROM book_runs br_latest WHERE br_latest.run_id = br.run_id AND br_latest.book_id = br.book_id) AND br.status IN ('failed','retryable_failed')), 0), bp.created_at, bp.updated_at, bp.archived_at FROM batch_projects bp"
	query := selectColumns + where + order + " LIMIT ? OFFSET ?"
	queryArgs := append(append([]any(nil), args...), filter.Limit, (filter.Page-1)*filter.Limit)
	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return BatchProjectPage{}, fmt.Errorf("list batch projects: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var project BatchProject
		var sources, genders, styles, runStatus string
		if err := rows.Scan(
			&project.ID,
			&project.IntakeID,
			&project.Name,
			&sources,
			&project.BookCount,
			&genders,
			&styles,
			&runStatus,
			&project.FailureCount,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.ArchivedAt,
		); err != nil {
			return BatchProjectPage{}, fmt.Errorf("scan batch project: %w", err)
		}
		project.Sources = splitBatchProjectSummary(sources)
		project.Genders = splitBatchProjectSummary(genders)
		project.Styles = splitBatchProjectSummary(styles)
		project.RunStatus = RunStatus(strings.TrimSpace(runStatus))
		page.Projects = append(page.Projects, project)
	}
	if err := rows.Err(); err != nil {
		return BatchProjectPage{}, fmt.Errorf("iterate batch projects: %w", err)
	}
	return page, nil
}

func batchProjectListWhere(filter BatchProjectListQuery) (string, []any) {
	conditions := make([]string, 0, 5)
	args := make([]any, 0, 10)
	switch filter.Archived {
	case BatchProjectArchivedActive:
		conditions = append(conditions, "bp.archived_at IS NULL")
	case BatchProjectArchivedArchived:
		conditions = append(conditions, "bp.archived_at IS NOT NULL")
	}
	if !filter.Elevated {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM auth_batch_project_ownership o WHERE o.batch_project_id = bp.id AND (o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?)))")
		args = append(args, filter.UserID, filter.TeamID, filter.TeamID)
	}
	if query := strings.ToLower(strings.TrimSpace(filter.Query)); query != "" {
		conditions = append(conditions, "(LOWER(bp.name) LIKE ? OR EXISTS (SELECT 1 FROM books qb WHERE qb.intake_id = bp.intake_id AND (LOWER(qb.title) LIKE ? OR LOWER(qb.external_book_id) LIKE ?)))")
		like := "%" + query + "%"
		args = append(args, like, like, like)
	}
	if source := strings.TrimSpace(filter.Source); source != "" {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM books sb WHERE sb.intake_id = bp.intake_id AND sb.source = ?)")
		args = append(args, source)
	}
	if status := strings.TrimSpace(string(filter.Status)); status != "" {
		conditions = append(conditions, "COALESCE((SELECT sr.status FROM runs sr WHERE sr.batch_project_id = bp.id ORDER BY sr.run_at DESC, sr.id DESC LIMIT 1), '') = ?")
		args = append(args, status)
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

const activeBatchProjectQuery = "SELECT EXISTS(SELECT 1 FROM runs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM book_runs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM stage_runs sr JOIN book_runs br ON br.id = sr.book_run_id WHERE br.batch_project_id = ? AND sr.status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM shuihuo_media_tasks WHERE batch_project_id = ? AND status IN ('pending_executor','pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_production_jobs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_production_tasks vt JOIN video_production_jobs vj ON vj.id = vt.production_job_id WHERE vj.batch_project_id = ? AND vt.status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_merge_jobs WHERE batch_project_id = ? AND status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM video_merge_attempts va JOIN video_merge_jobs vj ON vj.id = va.merge_job_id WHERE vj.batch_project_id = ? AND va.status IN ('pending','queued','scheduled','running') UNION ALL SELECT 1 FROM intakes i JOIN batch_projects bp ON bp.intake_id = i.id WHERE bp.id = ? AND i.status = 'running')"

func (s *MySQLStore) ArchiveBatchProject(ctx context.Context, projectID, actorUserID int64) error {
	if s == nil || s.db == nil || projectID <= 0 || actorUserID <= 0 {
		return fmt.Errorf("archive batch project: invalid request")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin archive batch project: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, "SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE", projectID).Scan(&archivedAt); err != nil {
		return fmt.Errorf("lock batch project: %w", err)
	}
	if archivedAt.Valid {
		return tx.Commit()
	}
	var active bool
	if err := tx.QueryRowContext(ctx, activeBatchProjectQuery, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID).Scan(&active); err != nil {
		return fmt.Errorf("check active batch project work: %w", err)
	}
	if active {
		return ErrBatchProjectActive
	}
	result, err := tx.ExecContext(ctx, "UPDATE batch_projects SET archived_at = UTC_TIMESTAMP(6), archived_by_user_id = ? WHERE id = ? AND archived_at IS NULL", actorUserID, projectID)
	if err != nil {
		return fmt.Errorf("archive batch project: %w", err)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		if affectedErr != nil {
			return fmt.Errorf("archive batch project rows: %w", affectedErr)
		}
		return fmt.Errorf("archive batch project: %w", sql.ErrNoRows)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit archive batch project: %w", err)
	}
	return nil
}

func (s *MySQLStore) RestoreBatchProject(ctx context.Context, projectID int64) error {
	if s == nil || s.db == nil || projectID <= 0 {
		return fmt.Errorf("restore batch project: invalid request")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin restore batch project: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, "SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE", projectID).Scan(&archivedAt); err != nil {
		return fmt.Errorf("lock batch project: %w", err)
	}
	if !archivedAt.Valid {
		return tx.Commit()
	}
	result, err := tx.ExecContext(ctx, "UPDATE batch_projects SET archived_at = NULL, archived_by_user_id = NULL WHERE id = ? AND archived_at IS NOT NULL", projectID)
	if err != nil {
		return fmt.Errorf("restore batch project: %w", err)
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		if affectedErr != nil {
			return fmt.Errorf("restore batch project rows: %w", affectedErr)
		}
		return fmt.Errorf("restore batch project: %w", sql.ErrNoRows)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit restore batch project: %w", err)
	}
	return nil
}

func (s *MySQLStore) IsBatchProjectArchived(ctx context.Context, projectID int64) (bool, error) {
	if s == nil || s.db == nil || projectID <= 0 {
		return false, fmt.Errorf("batch project archive reader unavailable")
	}
	var archived bool
	if err := s.db.QueryRowContext(ctx, "SELECT archived_at IS NOT NULL FROM batch_projects WHERE id = ?", projectID).Scan(&archived); err != nil {
		return false, fmt.Errorf("read batch project archive state: %w", err)
	}
	return archived, nil
}

func (s *MySQLStore) IsIntakeBatchProjectArchived(ctx context.Context, intakeID int64) (bool, error) {
	if s == nil || s.db == nil || intakeID <= 0 {
		return false, fmt.Errorf("batch project archive reader unavailable")
	}
	var archived bool
	if err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM batch_projects WHERE intake_id = ? AND archived_at IS NOT NULL)", intakeID).Scan(&archived); err != nil {
		return false, fmt.Errorf("read intake batch project archive state: %w", err)
	}
	return archived, nil
}

func splitBatchProjectSummary(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	result := make([]string, 0)
	for _, item := range strings.Split(value, "|") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func (s *MySQLStore) CreateRun(ctx context.Context, run Run) (Run, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, fmt.Errorf("begin create run: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, "SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE", run.BatchProjectID).Scan(&archivedAt); err != nil {
		return Run{}, fmt.Errorf("lock batch project for run: %w", err)
	}
	if archivedAt.Valid {
		return Run{}, ErrBatchProjectArchived
	}
	result, err := tx.ExecContext(ctx,
		"INSERT INTO runs (batch_project_id, run_at, status) VALUES (?, ?, ?)",
		run.BatchProjectID,
		run.RunAt,
		run.Status,
	)
	if err != nil {
		return Run{}, fmt.Errorf("create run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Run{}, fmt.Errorf("read run id: %w", err)
	}
	run.ID = id
	if err := tx.Commit(); err != nil {
		return Run{}, fmt.Errorf("commit create run: %w", err)
	}
	return run, nil
}
