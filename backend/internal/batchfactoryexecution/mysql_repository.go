package batchfactoryexecution

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var executionColumns = []string{"id", "owner_username", "batch_id", "mode", "target", "client_request_id", "status", "legacy_job_id", "created_at", "updated_at"}

const selectExecutionByRequestForUpdateSQL = `SELECT id, owner_username, batch_id, mode, target, client_request_id, status, legacy_job_id, created_at, updated_at FROM batch_executions WHERE owner_username = ? AND batch_id = ? AND client_request_id = ? FOR UPDATE`

type MySQLRepository struct {
	db    *sql.DB
	now   func() time.Time
	newID func(string) string
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db, now: func() time.Time { return time.Now().UTC() }, newID: randomExecutionID}
}

func randomExecutionID(prefix string) string {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	return prefix + "_" + hex.EncodeToString(raw[:])
}

func validateEnsureRecord(record EnsureRecord) error {
	if strings.TrimSpace(record.Owner) == "" || strings.TrimSpace(record.BatchID) == "" || strings.TrimSpace(record.ClientRequestID) == "" {
		return fmt.Errorf("owner, batch and client request id are required")
	}
	if len(record.Books) == 0 {
		return fmt.Errorf("at least one book is required")
	}
	seen := map[string]struct{}{}
	for _, book := range record.Books {
		id := strings.TrimSpace(book.BookID)
		if id == "" {
			return fmt.Errorf("book id is required")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate book id %s", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func scanExecution(scanner interface{ Scan(...any) error }) (Execution, error) {
	var execution Execution
	var legacy sql.NullString
	err := scanner.Scan(&execution.ID, &execution.OwnerUsername, &execution.BatchID, &execution.Mode, &execution.Target, &execution.ClientRequestID, &execution.Status, &legacy, &execution.CreatedAt, &execution.UpdatedAt)
	if legacy.Valid {
		execution.LegacyJobID = legacy.String
	}
	return execution, err
}

func (r *MySQLRepository) Ensure(ctx context.Context, record EnsureRecord) (result EnsureResult, err error) {
	if r == nil || r.db == nil {
		return result, fmt.Errorf("execution repository unavailable")
	}
	if err := validateEnsureRecord(record); err != nil {
		return result, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	existing, scanErr := scanExecution(tx.QueryRowContext(ctx, selectExecutionByRequestForUpdateSQL, record.Owner, record.BatchID, record.ClientRequestID))
	if scanErr == nil {
		rows, queryErr := tx.QueryContext(ctx, `SELECT book_id FROM batch_execution_books WHERE execution_id = ? ORDER BY book_id`, existing.ID)
		if queryErr != nil {
			err = queryErr
			return result, err
		}
		for rows.Next() {
			var id string
			if queryErr = rows.Scan(&id); queryErr != nil {
				_ = rows.Close()
				err = queryErr
				return result, err
			}
			result.QueuedBooks = append(result.QueuedBooks, id)
		}
		if queryErr = rows.Close(); queryErr != nil {
			err = queryErr
			return result, err
		}
		if err = tx.Commit(); err != nil {
			return result, err
		}
		result.Execution = existing
		return result, nil
	}
	if scanErr != sql.ErrNoRows {
		err = scanErr
		return result, err
	}

	args := []any{record.Owner, record.BatchID}
	placeholders := make([]string, len(record.Books))
	for index, book := range record.Books {
		placeholders[index] = "?"
		args = append(args, strings.TrimSpace(book.BookID))
	}
	activeRows, queryErr := tx.QueryContext(ctx, `SELECT eb.book_id FROM batch_execution_books eb JOIN batch_executions e ON e.id = eb.execution_id WHERE e.owner_username = ? AND e.batch_id = ? AND e.status IN ('planned','running','needs_attention') AND eb.status IN ('pending','running','waiting_external','retry_wait') AND eb.book_id IN (`+strings.Join(placeholders, ",")+`) FOR UPDATE`, args...)
	if queryErr != nil {
		err = queryErr
		return result, err
	}
	active := map[string]struct{}{}
	for activeRows.Next() {
		var id string
		if queryErr = activeRows.Scan(&id); queryErr != nil {
			_ = activeRows.Close()
			err = queryErr
			return result, err
		}
		active[id] = struct{}{}
		result.ReusedBooks = append(result.ReusedBooks, id)
	}
	if queryErr = activeRows.Close(); queryErr != nil {
		err = queryErr
		return result, err
	}

	now := r.now().UTC()
	executionID := r.newID("execution")
	if _, err = tx.ExecContext(ctx, `INSERT INTO batch_executions (id, owner_username, batch_id, mode, target, client_request_id, status, legacy_job_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, executionID, record.Owner, record.BatchID, record.Mode, record.Target, record.ClientRequestID, ExecutionPlanned, nil, now, now); err != nil {
		return result, err
	}
	if len(record.ConfigJSON) == 0 {
		record.ConfigJSON = []byte(`{}`)
	}
	if record.ConfigDigest == "" {
		record.ConfigDigest, err = ConfigDigest(json.RawMessage(record.ConfigJSON))
		if err != nil {
			return result, err
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO batch_execution_config_snapshots (execution_id, config_json, config_digest, created_at) VALUES (?, ?, ?, ?)`, executionID, string(record.ConfigJSON), record.ConfigDigest, now); err != nil {
		return result, err
	}
	for _, book := range record.Books {
		bookID := strings.TrimSpace(book.BookID)
		if _, exists := active[bookID]; exists {
			continue
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO batch_execution_books (id, execution_id, book_id, title, status, current_stage, failure_code, failure_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, r.newID("execution-book"), executionID, bookID, book.Title, BookPending, nil, nil, nil, now, now); err != nil {
			return result, err
		}
		result.QueuedBooks = append(result.QueuedBooks, bookID)
	}
	payload, _ := json.Marshal(map[string]any{"executionId": executionID, "bookIds": result.QueuedBooks})
	if _, err = tx.ExecContext(ctx, `INSERT INTO batch_execution_outbox (id, aggregate_type, aggregate_id, event_type, payload_json, available_at, claimed_at, delivered_at, attempts, last_error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, r.newID("outbox"), "execution", executionID, "execution.ensure", payload, now, nil, nil, 0, nil, now); err != nil {
		return result, err
	}
	if err = tx.Commit(); err != nil {
		return result, err
	}
	result.Execution = Execution{ID: executionID, OwnerUsername: record.Owner, BatchID: record.BatchID, Mode: record.Mode, Target: record.Target, ClientRequestID: record.ClientRequestID, Status: ExecutionPlanned, CreatedAt: now, UpdatedAt: now}
	result.Created = true
	return result, nil
}

func (r *MySQLRepository) Get(ctx context.Context, owner, executionID string) (Execution, error) {
	execution, err := scanExecution(r.db.QueryRowContext(ctx, `SELECT id, owner_username, batch_id, mode, target, client_request_id, status, legacy_job_id, created_at, updated_at FROM batch_executions WHERE owner_username = ? AND id = ?`, owner, executionID))
	if err == sql.ErrNoRows {
		return Execution{}, ErrNotFound
	}
	return execution, err
}

func (r *MySQLRepository) GetByClientRequest(ctx context.Context, owner, batchID, requestID string) (Execution, bool, error) {
	execution, err := scanExecution(r.db.QueryRowContext(ctx, `SELECT id, owner_username, batch_id, mode, target, client_request_id, status, legacy_job_id, created_at, updated_at FROM batch_executions WHERE owner_username = ? AND batch_id = ? AND client_request_id = ?`, owner, batchID, requestID))
	if err == sql.ErrNoRows {
		return Execution{}, false, nil
	}
	return execution, err == nil, err
}

func (r *MySQLRepository) ListBookTimeline(ctx context.Context, owner, executionID, bookID string) ([]TimelineEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT sr.stage_key, sr.status, COALESCE(sa.attempt_no,0), sr.input_fingerprint, COALESCE(sa.error_code,''), COALESCE(sa.error_message,''), sr.created_at FROM batch_execution_stage_runs sr JOIN batch_execution_books eb ON eb.id = sr.execution_book_id JOIN batch_executions e ON e.id = eb.execution_id LEFT JOIN batch_execution_stage_attempts sa ON sa.stage_run_id = sr.id WHERE e.owner_username = ? AND e.id = ? AND eb.book_id = ? ORDER BY sr.created_at, sa.attempt_no`, owner, executionID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []TimelineEvent
	for rows.Next() {
		var event TimelineEvent
		if err := rows.Scan(&event.StageKey, &event.Status, &event.AttemptNo, &event.InputFingerprint, &event.ErrorCode, &event.ErrorMessage, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *MySQLRepository) AttachLegacyJob(ctx context.Context, owner, executionID, legacyJobID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE batch_executions SET legacy_job_id = ?, status = ?, updated_at = ? WHERE owner_username = ? AND id = ?`, legacyJobID, ExecutionRunning, r.now().UTC(), owner, executionID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MySQLRepository) UpdateBookStatus(ctx context.Context, owner, bookID string, status BookStatus, failure Failure) error {
	result, err := r.db.ExecContext(ctx, `UPDATE batch_execution_books eb JOIN batch_executions e ON e.id = eb.execution_id SET eb.status = ?, eb.failure_code = ?, eb.failure_message = ?, eb.updated_at = ? WHERE e.owner_username = ? AND eb.book_id = ? AND e.status IN ('planned','running','needs_attention')`, status, nullableText(failure.Code), nullableText(failure.Message), r.now().UTC(), owner, bookID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MySQLRepository) AppendOutbox(ctx context.Context, event OutboxEvent) error {
	if event.ID == "" {
		event.ID = r.newID("outbox")
	}
	if event.AvailableAt.IsZero() {
		event.AvailableAt = r.now().UTC()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = r.now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO batch_execution_outbox (id, aggregate_type, aggregate_id, event_type, payload_json, available_at, claimed_at, delivered_at, attempts, last_error, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, event.ID, event.AggregateType, event.AggregateID, event.EventType, event.PayloadJSON, event.AvailableAt, event.ClaimedAt, event.DeliveredAt, event.Attempts, nullableText(event.LastError), event.CreatedAt)
	return err
}

func nullableText(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
