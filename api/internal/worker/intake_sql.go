package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type intakeBookRows interface {
	Next() bool
	Scan(...any) error
	Close() error
	Err() error
}

type intakeQuery func(context.Context, string, ...any) (intakeBookRows, error)
type intakeExec func(context.Context, string, ...any) (sql.Result, error)

type SQLIntakeRepository struct {
	query intakeQuery
	exec  intakeExec
}

func NewSQLIntakeRepository(db *sql.DB) *SQLIntakeRepository {
	if db == nil {
		return &SQLIntakeRepository{}
	}
	return &SQLIntakeRepository{
		query: func(ctx context.Context, query string, args ...any) (intakeBookRows, error) { return db.QueryContext(ctx, query, args...) },
		exec:  db.ExecContext,
	}
}

func newSQLIntakeRepository(query intakeQuery, exec intakeExec) *SQLIntakeRepository {
	return &SQLIntakeRepository{query: query, exec: exec}
}

func (r *SQLIntakeRepository) ListBooks(ctx context.Context, intakeID string) ([]IntakeBook, error) {
	if r == nil || r.query == nil {
		return nil, errors.New("intake repository database is required")
	}
	rows, err := r.query(ctx, `SELECT id, book_id, source_platform_id, manual_gender, category, genre, style FROM intake_books WHERE intake_id = ? ORDER BY id ASC`, intakeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var books []IntakeBook
	for rows.Next() {
		var book IntakeBook
		var manualGender string
		var genre sql.NullInt64
		if err := rows.Scan(&book.ID, &book.BookID, &book.PlatformID, &manualGender, &book.Category, &genre, &book.Style); err != nil {
			return nil, err
		}
		book.ManualGender = novel.Gender(manualGender)
		if genre.Valid {
			book.Genre = int(genre.Int64)
		}
		book.MaxTxt = 4000
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *SQLIntakeRepository) SaveFetched(ctx context.Context, id int64, book FetchedBook) error {
	if r == nil || r.exec == nil {
		return errors.New("intake repository database is required")
	}
	metadata, err := json.Marshal(book.BookInfo)
	if err != nil {
		return err
	}
	_, err = r.exec(ctx, `UPDATE intake_books SET book_name = ?, author = ?, source_text = ?, category = ?, genre = ?, source_metadata = ?, fetch_status = 'ready', error_message = '' WHERE id = ?`,
		book.BookName, book.Author, book.SourceText, book.Category, book.Genre, metadata, id)
	return err
}

func (r *SQLIntakeRepository) SaveResolved(ctx context.Context, id int64, meta ResolvedMetadata) error {
	if r == nil || r.exec == nil {
		return errors.New("intake repository database is required")
	}
	styleSource := "existing"
	if meta.Style == "" {
		styleSource = "unresolved"
	}
	_, err := r.exec(ctx, `UPDATE intake_books SET resolved_gender = ?, gender_source = ?, style = ?, style_source = ? WHERE id = ?`,
		string(meta.Gender.Gender), string(meta.Gender.Source), meta.Style, styleSource, id)
	return err
}
