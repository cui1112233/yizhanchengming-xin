package intake

import (
	"context"
	"database/sql"
	"fmt"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) CreateIntake(ctx context.Context, name string) (Intake, error) {
	result, err := s.db.ExecContext(ctx, "INSERT INTO intakes (name, status) VALUES (?, ?)", name, StatusPending)
	if err != nil {
		return Intake{}, fmt.Errorf("create intake: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Intake{}, fmt.Errorf("read intake id: %w", err)
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

func (s *MySQLStore) UpsertBook(ctx context.Context, book Book) (Book, error) {
	const query = "INSERT INTO books (intake_id, source, external_book_id, title, body_ref, category, genre, gender, gender_source, style, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), title = VALUES(title), body_ref = VALUES(body_ref), category = VALUES(category), genre = VALUES(genre), gender = VALUES(gender), gender_source = VALUES(gender_source), style = VALUES(style), status = VALUES(status), error_message = VALUES(error_message)"
	result, err := s.db.ExecContext(ctx, query,
		book.IntakeID,
		book.Source,
		book.ExternalBookID,
		book.Title,
		book.BodyRef,
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
