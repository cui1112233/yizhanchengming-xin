package intake

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	result, err := s.db.ExecContext(ctx, "UPDATE intakes SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("update intake status: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return nil
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
	const query = "INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = VALUES(name)"
	result, err := s.db.ExecContext(ctx, query, project.IntakeID, project.Name)
	if err != nil {
		return BatchProject{}, fmt.Errorf("create batch project: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return BatchProject{}, fmt.Errorf("read batch project id: %w", err)
	}
	project.ID = id
	return project, nil
}

func (s *MySQLStore) GetBatchProject(ctx context.Context, id int64) (BatchProject, error) {
	var project BatchProject
	err := s.db.QueryRowContext(ctx,
		"SELECT id, intake_id, name, created_at, updated_at FROM batch_projects WHERE id = ?",
		id,
	).Scan(&project.ID, &project.IntakeID, &project.Name, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return BatchProject{}, fmt.Errorf("get batch project: %w", err)
	}
	return project, nil
}

func (s *MySQLStore) ListBatchProjects(ctx context.Context) ([]BatchProject, error) {
	const query = "SELECT bp.id, bp.intake_id, bp.name, COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.source), '') ORDER BY b.source SEPARATOR '|'), ''), COUNT(b.id), COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.gender), '') ORDER BY b.gender SEPARATOR '|'), ''), COALESCE(GROUP_CONCAT(DISTINCT NULLIF(TRIM(b.style), '') ORDER BY b.style SEPARATOR '|'), ''), COALESCE((SELECT r.status FROM runs r WHERE r.batch_project_id = bp.id ORDER BY r.run_at DESC, r.id DESC LIMIT 1), ''), bp.created_at, bp.updated_at FROM batch_projects bp LEFT JOIN books b ON b.intake_id = bp.intake_id GROUP BY bp.id, bp.intake_id, bp.name, bp.created_at, bp.updated_at ORDER BY bp.updated_at DESC, bp.id DESC LIMIT 100"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list batch projects: %w", err)
	}
	defer rows.Close()

	projects := make([]BatchProject, 0)
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
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan batch project: %w", err)
		}
		project.Sources = splitBatchProjectSummary(sources)
		project.Genders = splitBatchProjectSummary(genders)
		project.Styles = splitBatchProjectSummary(styles)
		project.RunStatus = RunStatus(strings.TrimSpace(runStatus))
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch projects: %w", err)
	}
	return projects, nil
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
	result, err := s.db.ExecContext(ctx,
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
	return run, nil
}
