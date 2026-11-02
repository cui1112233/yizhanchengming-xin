package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"
)

var _ BookStageRunRepository = (*MySQLStore)(nil)

func (s *MySQLStore) CreateBookStageRun(ctx context.Context, value BookStageRun) (BookStageRun, error) {
	value, err := normalizeBookStageRun(value)
	if err != nil {
		return BookStageRun{}, err
	}
	if err := ensureBookOwnership(ctx, s.db, value.Owner, value.BatchID, value.BookID); err != nil {
		return BookStageRun{}, err
	}
	id, err := newID("book-stage")
	if err != nil {
		return BookStageRun{}, err
	}
	now := time.Now().UTC()
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = value.CreatedAt
	}
	value.ID = id
	_, err = s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_book_stage_runs(id,owner_username,batch_id,book_id,stage,status,attempt,request_id,input_revision,error_message,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, value.ID, value.Owner, value.BatchID, value.BookID, value.Stage, value.Status, value.Attempt, nullableString(value.RequestID), nullableString(value.InputRevision), nullableString(value.ErrorMessage), value.CreatedAt, value.UpdatedAt)
	if err != nil {
		return BookStageRun{}, err
	}
	return value, nil
}

func (s *MySQLStore) UpdateBookStageRun(ctx context.Context, owner, id string, value BookStageRun) (BookStageRun, error) {
	if value.ID != "" && value.ID != id {
		return BookStageRun{}, ErrInvalid
	}
	var existing BookStageRun
	var stage, status string
	err := s.db.QueryRowContext(ctx, `SELECT batch_id,book_id,stage,status,attempt,COALESCE(request_id,''),COALESCE(input_revision,''),COALESCE(error_message,''),created_at FROM batch_factory_v12_book_stage_runs WHERE id=? AND owner_username=?`, id, owner).Scan(&existing.BatchID, &existing.BookID, &stage, &status, &existing.Attempt, &existing.RequestID, &existing.InputRevision, &existing.ErrorMessage, &existing.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BookStageRun{}, ErrNotFound
	}
	if err != nil {
		return BookStageRun{}, err
	}
	existing.ID, existing.Owner, existing.Stage, existing.Status = id, owner, BookStage(stage), ProductionState(status)
	if value.Status != "" {
		existing.Status = value.Status
	}
	if value.ErrorMessage != "" || existing.Status == ProductionSucceeded {
		existing.ErrorMessage = value.ErrorMessage
	}
	if value.InputRevision != "" {
		existing.InputRevision = value.InputRevision
	}
	existing.UpdatedAt = time.Now().UTC()
	existing, err = normalizeBookStageRun(existing)
	if err != nil {
		return BookStageRun{}, err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE batch_factory_v12_book_stage_runs SET status=?,error_message=?,input_revision=?,updated_at=? WHERE id=? AND owner_username=?`, existing.Status, nullableString(existing.ErrorMessage), nullableString(existing.InputRevision), existing.UpdatedAt, id, owner)
	if err != nil {
		return BookStageRun{}, err
	}
	return existing, nil
}

func (s *MySQLStore) ListBookStageRuns(ctx context.Context, owner, batchID, bookID string) ([]BookStageRun, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,stage,status,attempt,COALESCE(request_id,''),COALESCE(input_revision,''),COALESCE(error_message,''),created_at,updated_at FROM batch_factory_v12_book_stage_runs WHERE owner_username=? AND batch_id=? AND book_id=? ORDER BY updated_at,id`, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BookStageRun{}
	for rows.Next() {
		var value BookStageRun
		var stage, status string
		if err := rows.Scan(&value.ID, &stage, &status, &value.Attempt, &value.RequestID, &value.InputRevision, &value.ErrorMessage, &value.CreatedAt, &value.UpdatedAt); err != nil {
			return nil, err
		}
		value.Owner, value.BatchID, value.BookID, value.Stage, value.Status = owner, batchID, bookID, BookStage(stage), ProductionState(status)
		out = append(out, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt.Before(out[j].UpdatedAt)
	})
	return out, nil
}
