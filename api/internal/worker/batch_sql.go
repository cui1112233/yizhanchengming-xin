package worker

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
)

type batchTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

type beginBatchTx func(context.Context) (batchTx, error)

type SQLBatchCreator struct {
	begin beginBatchTx
}

func NewSQLBatchCreator(db *sql.DB) *SQLBatchCreator {
	if db == nil {
		return &SQLBatchCreator{}
	}
	return &SQLBatchCreator{begin: func(ctx context.Context) (batchTx, error) {
		return db.BeginTx(ctx, nil)
	}}
}

func newSQLBatchCreatorWithBegin(begin beginBatchTx) *SQLBatchCreator {
	return &SQLBatchCreator{begin: begin}
}

func (c *SQLBatchCreator) CreateFromIntake(ctx context.Context, intakeID, jobID string) (string, error) {
	if c == nil || c.begin == nil {
		return "", errors.New("batch creator database is required")
	}
	if intakeID == "" || jobID == "" {
		return "", errors.New("intake id and job id are required")
	}
	batchID, err := newBatchID()
	if err != nil {
		return "", err
	}
	tx, err := c.begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO batches (id, owner, intake_id, title, status) SELECT ?, owner, id, title, 'created' FROM intakes WHERE id = ?`,
		batchID, intakeID,
	)
	if err != nil {
		return "", err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if rows != 1 {
		return "", errors.New("intake not found")
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO batch_books (batch_id, intake_book_id, book_id, title, source_platform_id, gender, gender_source, style, source_text)
SELECT ?, id, book_id, book_name, source_platform_id, resolved_gender, gender_source, style, source_text
FROM intake_books WHERE intake_id = ? ORDER BY id ASC`,
		batchID, intakeID,
	); err != nil {
		return "", err
	}

	res, err = tx.ExecContext(ctx,
		`UPDATE pipeline_jobs SET batch_id = ? WHERE id = ? AND intake_id = ?`,
		batchID, jobID, intakeID,
	)
	if err != nil {
		return "", err
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return "", err
	}
	if rows != 1 {
		return "", errors.New("pipeline job not found")
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return batchID, nil
}

func newBatchID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
