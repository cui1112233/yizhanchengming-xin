package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type intakeStoreTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type beginIntakeStoreTx func(context.Context) (intakeStoreTx, error)

type SQLIntakeStore struct {
	begin beginIntakeStoreTx
}

func NewSQLIntakeStore(db *sql.DB) *SQLIntakeStore {
	if db == nil {
		return &SQLIntakeStore{}
	}
	return &SQLIntakeStore{begin: func(ctx context.Context) (intakeStoreTx, error) {
		return db.BeginTx(ctx, nil)
	}}
}

func newSQLIntakeStoreWithBegin(begin beginIntakeStoreTx) *SQLIntakeStore {
	return &SQLIntakeStore{begin: begin}
}

func (s *SQLIntakeStore) CreateIntake(ctx context.Context, record batchfactory.IntakeRecord) error {
	if s == nil || s.begin == nil {
		return errors.New("sql intake store database is required")
	}
	if strings.TrimSpace(record.ID) == "" || strings.TrimSpace(record.Owner) == "" || len(record.Books) == 0 {
		return errors.New("intake id, owner and books are required")
	}
	tx, err := s.begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO intakes (id, owner, title, status) VALUES (?, ?, ?, 'created')`,
		record.ID, record.Owner, record.Title,
	); err != nil {
		return err
	}

	for _, book := range record.Books {
		resolvedGender := ""
		genderSource := string(novel.GenderSourceUnresolved)
		if book.ManualGender == novel.GenderMale || book.ManualGender == novel.GenderFemale {
			resolvedGender = string(book.ManualGender)
			genderSource = string(novel.GenderSourceManual)
		}
		styleSource := "unresolved"
		if strings.TrimSpace(book.Style) != "" {
			styleSource = "manual"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO intake_books (intake_id, source_platform_id, source_platform_name, max_txt, book_id, manual_gender, resolved_gender, gender_source, style, style_source, fetch_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending')`,
			record.ID,
			book.PlatformID,
			book.PlatformName,
			book.MaxTxt,
			book.BookID,
			string(book.ManualGender),
			resolvedGender,
			genderSource,
			book.Style,
			styleSource,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}
