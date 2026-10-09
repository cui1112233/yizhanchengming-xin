package task9runtime

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrRunNotFound         = errors.New("task9 run not found")
	ErrBookRunNotRetryable = errors.New("task9 book run not retryable")
)

type RunRecord struct {
	ID                   int64
	BatchProjectID       int64
	IdempotencyKey       string
	RunAt                time.Time
	Status               RunState
	MaxAttempts          int
	RunKind              string
	TargetBookID         *int64
	RequestSchemaVersion int
	RequestSnapshot      json.RawMessage
	RequestHash          string
	RequestedByUserID    int64
}

type MySQLStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func publicRunState(state RunState) (string, bool) {
	switch state {
	case RunSucceeded:
		return "completed", true
	case RunPartialFailed:
		return "partial_failed", true
	case RunFailed:
		return "failed", true
	case RunState("cancelled"):
		return "cancelled", true
	case RunPending:
		return "queued", false
	default:
		return "running", false
	}
}

func publicBookState(state BookState) string {
	switch state {
	case BookSucceeded:
		return "completed"
	case BookPending, BookQueued:
		return "queued"
	case BookRunning:
		return "running"
	default:
		return "failed"
	}
}

// GenerationRun returns lifecycle facts only. It deliberately does not load
// request_snapshot or any generation output, and performs no queue/provider work.
func (s *MySQLStore) GenerationRun(ctx context.Context, projectID, runID int64) (GenerationRunStatus, error) {
	if s == nil || s.db == nil || projectID <= 0 || runID <= 0 {
		return GenerationRunStatus{}, ErrRunNotFound
	}
	var raw string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM runs WHERE id=? AND batch_project_id=? AND BINARY run_kind='generation'`, runID, projectID).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GenerationRunStatus{}, ErrRunNotFound
		}
		return GenerationRunStatus{}, err
	}
	status, terminal := publicRunState(RunState(raw))
	out := GenerationRunStatus{RunID: runID, BatchProjectID: projectID, Status: status, Terminal: terminal, Tasks: []GenerationTaskStatus{}}
	rows, err := s.db.QueryContext(ctx, `SELECT br.id,br.book_id,br.attempt,br.status FROM book_runs br JOIN (SELECT book_id,MAX(attempt) max_attempt FROM book_runs WHERE run_id=? GROUP BY book_id) latest ON latest.book_id=br.book_id AND latest.max_attempt=br.attempt WHERE br.run_id=? ORDER BY br.book_id`, runID, runID)
	if err != nil {
		return GenerationRunStatus{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var task GenerationTaskStatus
		var state string
		if err := rows.Scan(&task.TaskID, &task.BookID, &task.Attempt, &state); err != nil {
			return GenerationRunStatus{}, err
		}
		task.Status = publicBookState(BookState(state))
		out.Tasks = append(out.Tasks, task)
		out.Counts.Total++
		switch task.Status {
		case "queued":
			out.Counts.Pending++
		case "running":
			out.Counts.Running++
		case "completed":
			out.Counts.Completed++
		default:
			out.Counts.Failed++
		}
	}
	if err := rows.Err(); err != nil {
		return GenerationRunStatus{}, err
	}
	return out, nil
}

func (s *MySQLStore) CreateRun(ctx context.Context, projectID int64, key string, runAt time.Time, maxAttempts int) (RunRecord, bool, error) {
	if key == "" {
		return RunRecord{}, false, errors.New("idempotency key required")
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO runs (batch_project_id,idempotency_key,run_at,status,max_attempts) VALUES (?,?,?,'pending',?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`, projectID, key, runAt, maxAttempts)
	if err != nil {
		return RunRecord{}, false, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return RunRecord{}, false, err
	}
	affected, _ := res.RowsAffected()
	created := affected == 1
	var r RunRecord
	var status string
	if err := s.db.QueryRowContext(ctx, `SELECT id,batch_project_id,idempotency_key,run_at,status,max_attempts FROM runs WHERE id=?`, id).Scan(&r.ID, &r.BatchProjectID, &r.IdempotencyKey, &r.RunAt, &status, &r.MaxAttempts); err != nil {
		return RunRecord{}, false, err
	}
	r.Status = RunState(status)
	return r, created, nil
}

func (s *MySQLStore) ClaimDueRun(ctx context.Context, now time.Time) (RunClaim, bool, error) {
	// Compatibility for existing callers: claiming already materializes, so
	// there is no longer a committed running Run with an uncommitted book list.
	claim, _, ok, err := s.ClaimAndMaterializeDueRun(ctx, now)
	return claim, ok, err
}

func (s *MySQLStore) EnsureQueuedBooks(ctx context.Context, claim RunClaim) ([]WorkItem, error) {
	// Compatibility read only. The atomic admission/scheduler owns creation.
	var projectID int64
	if err := s.db.QueryRowContext(ctx, `SELECT batch_project_id FROM runs WHERE id=?`, claim.RunID).Scan(&projectID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := lockGenerationProject(ctx, tx, projectID, false); err != nil {
		return nil, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE`, claim.RunID))
	if err != nil {
		return nil, err
	}
	if r.RunKind != "generation" || r.Status != RunRunning || r.BatchProjectID != projectID {
		return nil, fmt.Errorf("run %d not active generation", claim.RunID)
	}
	if _, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID); err != nil {
		return nil, err
	}
	items, err := queuedGenerationItems(ctx, tx, claim.RunID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *MySQLStore) Claim(ctx context.Context, item WorkItem, owner string, deadline time.Time) (Execution, bool, error) {
	runID, projectID, err := s.bookRunParents(ctx, item.BookRunID)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrBookRunNotRetryable) {
		return Execution{}, false, nil
	}
	if err != nil {
		return Execution{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Execution{}, false, err
	}
	defer tx.Rollback()
	if _, err := lockGenerationProject(ctx, tx, projectID, false); err != nil {
		if errors.Is(err, ErrProjectArchived) || errors.Is(err, sql.ErrNoRows) {
			return Execution{}, false, nil
		}
		return Execution{}, false, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE`, runID))
	if errors.Is(err, sql.ErrNoRows) {
		return Execution{}, false, nil
	}
	if err != nil {
		return Execution{}, false, err
	}
	if r.RunKind != "generation" || r.Status != RunRunning || r.BatchProjectID != projectID {
		return Execution{}, false, nil
	}
	snapshot, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
	if errors.Is(err, ErrInvalidGenerationRequest) {
		return Execution{}, false, nil
	}
	if err != nil {
		return Execution{}, false, err
	}
	selected := false
	for _, id := range snapshot.BookIDs {
		if id == item.BookID {
			selected = true
			break
		}
	}
	if !selected {
		return Execution{}, false, nil
	}
	now := time.Now()
	res, err := tx.ExecContext(ctx, `UPDATE book_runs SET status='running',execution_token=execution_token+1,execution_owner=?,running_since=?,lease_deadline=?,heartbeat_at=? WHERE id=? AND run_id=? AND batch_project_id=? AND book_id=? AND attempt=? AND status='queued'`, owner, now, deadline, now, item.BookRunID, runID, projectID, item.BookID, item.Attempt)
	if err != nil {
		return Execution{}, false, err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return Execution{}, false, nil
	}
	var token uint64
	var attempt int
	if err := tx.QueryRowContext(ctx, `SELECT attempt,execution_token FROM book_runs WHERE id=?`, item.BookRunID).Scan(&attempt, &token); err != nil {
		return Execution{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, false, err
	}
	return Execution{BookRunID: item.BookRunID, Attempt: attempt, FencingToken: token, Owner: owner}, true, nil
}

func (s *MySQLStore) Renew(ctx context.Context, e Execution, deadline time.Time) (bool, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE book_runs SET lease_deadline=?,heartbeat_at=? WHERE id=? AND attempt=? AND execution_token=? AND execution_owner=? AND status='running'`, deadline, time.Now(), e.BookRunID, e.Attempt, e.FencingToken, e.Owner)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (s *MySQLStore) Complete(ctx context.Context, e Execution) (bool, error) {
	return s.finalizeExecution(ctx, e, nil)
}

func safeFailure(code, message string) (string, string) {
	if message == "" {
		if code == "" {
			code = "execution_error"
		}
		return code, ""
	}
	safeCode, safeMessage := SafeError(errors.New(message))
	if safeCode == "internal_error" {
		return safeCode, safeMessage
	}
	if code == "" {
		code = safeCode
	}
	return code, safeMessage
}
func (s *MySQLStore) Fail(ctx context.Context, e Execution, f Failure) (bool, error) {
	code, msg := safeFailure(f.Code, f.Message)
	return s.finalizeExecution(ctx, e, &Failure{Code: code, Message: msg, Retryable: f.Retryable})
}

func (s *MySQLStore) finalizeExecution(ctx context.Context, e Execution, failure *Failure) (bool, error) {
	runID, projectID, err := s.bookRunParents(ctx, e.BookRunID)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrBookRunNotRetryable) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if _, err := lockGenerationProject(ctx, tx, projectID, false); err != nil {
		return false, err
	}
	var lockedProjectID int64
	var runStatus, runKind string
	var cancelRequested, cancelled sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT batch_project_id,status,run_kind,cancel_requested_at,cancelled_at FROM runs WHERE id=? FOR UPDATE`, runID).Scan(&lockedProjectID, &runStatus, &runKind, &cancelRequested, &cancelled); err != nil {
		return false, err
	}
	if lockedProjectID != projectID || runKind != "generation" || RunState(runStatus) != RunRunning || cancelRequested.Valid || cancelled.Valid {
		return false, nil
	}
	var bookID int64
	var attempt int
	var token uint64
	var owner, status string
	if err := tx.QueryRowContext(ctx, `SELECT book_id,attempt,execution_token,execution_owner,status FROM book_runs WHERE id=? AND run_id=? AND batch_project_id=? FOR UPDATE`, e.BookRunID, runID, projectID).Scan(&bookID, &attempt, &token, &owner, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if attempt != e.Attempt || token != e.FencingToken || owner != e.Owner || status != string(BookRunning) {
		return false, nil
	}
	now := time.Now()
	var result sql.Result
	if failure == nil {
		result, err = tx.ExecContext(ctx, `UPDATE book_runs SET status='succeeded',retryable=0,error_code='',error_message='',finished_at=?,lease_deadline=NULL,heartbeat_at=? WHERE id=? AND attempt=? AND execution_token=? AND execution_owner=? AND status='running'`, now, now, e.BookRunID, e.Attempt, e.FencingToken, e.Owner)
	} else {
		retryable := 0
		if failure.Retryable {
			retryable = 1
		}
		result, err = tx.ExecContext(ctx, `UPDATE book_runs SET status='failed',retryable=?,error_code=?,error_message=?,finished_at=?,lease_deadline=NULL,heartbeat_at=? WHERE id=? AND attempt=? AND execution_token=? AND execution_owner=? AND status='running'`, retryable, failure.Code, failure.Message, now, now, e.BookRunID, e.Attempt, e.FencingToken, e.Owner)
	}
	if err != nil {
		return false, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		if err != nil {
			return false, err
		}
		return false, nil
	}
	if _, err := aggregateRunStatusTx(ctx, tx, runID, now); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	_ = bookID
	return true, nil
}

func aggregateRunStatusTx(ctx context.Context, tx *sql.Tx, runID int64, now time.Time) (RunState, error) {
	rows, err := tx.QueryContext(ctx, `SELECT br.status FROM book_runs br JOIN (SELECT book_id,MAX(attempt) max_attempt FROM book_runs WHERE run_id=? GROUP BY book_id) latest ON latest.book_id=br.book_id AND latest.max_attempt=br.attempt WHERE br.run_id=? FOR UPDATE`, runID, runID)
	if err != nil {
		return RunRunning, err
	}
	total, succeeded, failed := 0, 0, 0
	active := false
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			rows.Close()
			return RunRunning, err
		}
		total++
		switch status {
		case string(BookSucceeded):
			succeeded++
		case string(BookFailed):
			failed++
		default:
			active = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return RunRunning, err
	}
	if err := rows.Close(); err != nil {
		return RunRunning, err
	}
	state := RunRunning
	if total > 0 && !active {
		switch {
		case succeeded == total:
			state = RunSucceeded
		case failed == total:
			state = RunFailed
		case succeeded > 0 && failed > 0:
			state = RunPartialFailed
		}
	}
	if state == RunSucceeded || state == RunPartialFailed || state == RunFailed {
		_, err = tx.ExecContext(ctx, `UPDATE runs SET status=?,finished_at=? WHERE id=? AND status='running' AND cancel_requested_at IS NULL AND cancelled_at IS NULL`, string(state), now, runID)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE runs SET status='running',finished_at=NULL WHERE id=? AND status='running' AND cancel_requested_at IS NULL AND cancelled_at IS NULL`, runID)
	}
	return state, err
}

func (s *MySQLStore) ProjectIDForBookRun(ctx context.Context, bookRunID int64) (int64, error) {
	var projectID int64
	if err := s.db.QueryRowContext(ctx, `SELECT batch_project_id FROM book_runs WHERE id=?`, bookRunID).Scan(&projectID); err != nil {
		return 0, err
	}
	return projectID, nil
}

func (s *MySQLStore) RetryBookRun(ctx context.Context, failedBookRunID int64) (WorkItem, bool, error) {
	parentRunID, parentProjectID, err := s.bookRunParents(ctx, failedBookRunID)
	if err != nil {
		return WorkItem{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return WorkItem{}, false, err
	}
	defer tx.Rollback()
	intakeID, err := lockGenerationProject(ctx, tx, parentProjectID, false)
	if err != nil {
		return WorkItem{}, false, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE`, parentRunID))
	if err != nil {
		return WorkItem{}, false, err
	}
	if r.RunKind != "generation" || r.BatchProjectID != parentProjectID {
		return WorkItem{}, false, ErrBookRunNotRetryable
	}
	snapshot, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
	if err != nil {
		return WorkItem{}, false, ErrBookRunNotRetryable
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		return WorkItem{}, false, err
	}
	var runID sql.NullInt64
	var projectID, bookID int64
	var attempt, maxAttempts int
	var retryable bool
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status FROM book_runs WHERE id=? FOR UPDATE`, failedBookRunID).Scan(&runID, &projectID, &bookID, &attempt, &maxAttempts, &retryable, &status); err != nil {
		return WorkItem{}, false, err
	}
	if !runID.Valid || runID.Int64 != parentRunID || projectID != parentProjectID || status != "failed" || !retryable || attempt >= maxAttempts {
		return WorkItem{}, false, ErrBookRunNotRetryable
	}
	next := attempt + 1
	item, created, err := insertGenerationAttempt(ctx, tx, r, bookID, next, maxAttempts)
	if err != nil {
		return WorkItem{}, false, err
	}
	if created {
		if _, err := tx.ExecContext(ctx, `UPDATE runs SET status='running',finished_at=NULL WHERE id=?`, runID.Int64); err != nil {
			return WorkItem{}, false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return WorkItem{}, false, err
	}
	return item, created, nil
}

func (s *MySQLStore) AggregateRunStatus(ctx context.Context, runID int64) (RunState, error) {
	var projectID int64
	if err := s.db.QueryRowContext(ctx, `SELECT batch_project_id FROM runs WHERE id=?`, runID).Scan(&projectID); err != nil {
		return RunRunning, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return RunRunning, err
	}
	defer tx.Rollback()
	if _, err := lockGenerationProject(ctx, tx, projectID, false); err != nil {
		return RunRunning, err
	}
	var lockedProjectID int64
	var status, kind string
	var cancelRequested, cancelled sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT batch_project_id,status,run_kind,cancel_requested_at,cancelled_at FROM runs WHERE id=? FOR UPDATE`, runID).Scan(&lockedProjectID, &status, &kind, &cancelRequested, &cancelled); err != nil {
		return RunRunning, err
	}
	if lockedProjectID != projectID {
		return RunRunning, ErrRunNotFound
	}
	// Only a live generation execution may aggregate. A Retry explicitly
	// reopens a terminal Run under this same lock; cancellation/terminal state
	// cannot be overwritten by a stale aggregation snapshot.
	if kind != "generation" || RunState(status) != RunRunning || cancelRequested.Valid || cancelled.Valid {
		return RunState(status), nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT br.status FROM book_runs br JOIN (SELECT book_id,MAX(attempt) max_attempt FROM book_runs WHERE run_id=? GROUP BY book_id) latest ON latest.book_id=br.book_id AND latest.max_attempt=br.attempt WHERE br.run_id=? FOR UPDATE`, runID, runID)
	if err != nil {
		return RunRunning, err
	}
	total, succeeded, failed := 0, 0, 0
	active := false
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			rows.Close()
			return RunRunning, err
		}
		total++
		switch status {
		case "succeeded":
			succeeded++
		case "failed":
			failed++
		default:
			active = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return RunRunning, err
	}
	if err := rows.Close(); err != nil {
		return RunRunning, err
	}
	state := RunRunning
	if total == 0 {
		state = RunRunning
	} else if active {
		state = RunRunning
	} else if succeeded == total {
		state = RunSucceeded
	} else if failed == total {
		state = RunFailed
	} else if succeeded > 0 && failed > 0 {
		state = RunPartialFailed
	}
	if state == RunSucceeded || state == RunPartialFailed || state == RunFailed {
		_, err = tx.ExecContext(ctx, `UPDATE runs SET status=?,finished_at=? WHERE id=? AND status='running' AND cancel_requested_at IS NULL AND cancelled_at IS NULL`, string(state), time.Now(), runID)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE runs SET status='running',finished_at=NULL WHERE id=? AND status='running' AND cancel_requested_at IS NULL AND cancelled_at IS NULL`, runID)
	}
	if err != nil {
		return state, err
	}
	if err := tx.Commit(); err != nil {
		return state, err
	}
	return state, nil
}

func (s *MySQLStore) ListQueuedBookRuns(ctx context.Context, limit int) ([]WorkItem, error) {
	if limit <= 0 {
		return nil, nil
	}
	var out []WorkItem
	var cursorID int64
	for len(out) < limit {
		var candidate WorkItem
		err := s.db.QueryRowContext(ctx, `SELECT br.id,br.book_id,br.attempt FROM book_runs br JOIN runs r ON r.id=br.run_id JOIN batch_projects p ON p.id=r.batch_project_id WHERE br.status='queued' AND BINARY r.run_kind='generation' AND r.status='running' AND p.archived_at IS NULL AND br.id>? ORDER BY br.id LIMIT 1`, cursorID).Scan(&candidate.BookRunID, &candidate.BookID, &candidate.Attempt)
		if errors.Is(err, sql.ErrNoRows) {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		cursorID = candidate.BookRunID
		item, valid, err := s.validQueuedGenerationItem(ctx, candidate.BookRunID)
		if err != nil {
			return out, err
		}
		if valid {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *MySQLStore) RecoverStaleBookRuns(ctx context.Context, now time.Time, limit int) ([]WorkItem, error) {
	if limit <= 0 {
		return nil, nil
	}
	out := make([]WorkItem, 0, limit)
	var cursorDeadline time.Time
	var cursorID int64
	for len(out) < limit {
		query := `SELECT br.id,br.lease_deadline FROM book_runs br JOIN runs r ON r.id=br.run_id JOIN batch_projects p ON p.id=r.batch_project_id WHERE br.status='running' AND br.lease_deadline IS NOT NULL AND br.lease_deadline<=? AND BINARY r.run_kind='generation' AND r.status='running' AND p.archived_at IS NULL`
		args := []any{now}
		if cursorID > 0 {
			query += ` AND (br.lease_deadline>? OR (br.lease_deadline=? AND br.id>?))`
			args = append(args, cursorDeadline, cursorDeadline, cursorID)
		}
		query += ` ORDER BY br.lease_deadline,br.id LIMIT 1`
		var id int64
		var deadline time.Time
		err := s.db.QueryRowContext(ctx, query, args...).Scan(&id, &deadline)
		if errors.Is(err, sql.ErrNoRows) {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		cursorDeadline, cursorID = deadline, id
		item, created, err := s.recoverOne(ctx, id, now)
		if err != nil {
			return out, err
		}
		if created {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *MySQLStore) validQueuedGenerationItem(ctx context.Context, id int64) (WorkItem, bool, error) {
	runID, projectID, err := s.bookRunParents(ctx, id)
	if errors.Is(err, ErrBookRunNotRetryable) || errors.Is(err, sql.ErrNoRows) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return WorkItem{}, false, err
	}
	defer tx.Rollback()
	intakeID, err := lockGenerationProject(ctx, tx, projectID, false)
	if errors.Is(err, ErrProjectArchived) || errors.Is(err, sql.ErrNoRows) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE`, runID))
	if errors.Is(err, sql.ErrNoRows) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	if r.RunKind != "generation" || r.BatchProjectID != projectID || r.Status != RunRunning {
		return WorkItem{}, false, nil
	}
	snapshot, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
	if errors.Is(err, ErrInvalidGenerationRequest) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		if errors.Is(err, ErrInvalidGenerationRequest) {
			return WorkItem{}, false, nil
		}
		return WorkItem{}, false, err
	}
	var actualRunID, actualProjectID int64
	var item WorkItem
	var status string
	item.BookRunID = id
	if err := tx.QueryRowContext(ctx, `SELECT run_id,batch_project_id,book_id,attempt,status FROM book_runs WHERE id=? FOR UPDATE`, id).Scan(&actualRunID, &actualProjectID, &item.BookID, &item.Attempt, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WorkItem{}, false, nil
		}
		return WorkItem{}, false, err
	}
	selected := false
	for _, bookID := range snapshot.BookIDs {
		if bookID == item.BookID {
			selected = true
			break
		}
	}
	if actualRunID != runID || actualProjectID != projectID || status != "queued" || !selected {
		return WorkItem{}, false, nil
	}
	if err := tx.Commit(); err != nil {
		return WorkItem{}, false, err
	}
	return item, true, nil
}
func (s *MySQLStore) recoverOne(ctx context.Context, id int64, now time.Time) (WorkItem, bool, error) {
	parentRunID, parentProjectID, err := s.bookRunParents(ctx, id)
	if errors.Is(err, ErrBookRunNotRetryable) || errors.Is(err, sql.ErrNoRows) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return WorkItem{}, false, err
	}
	defer tx.Rollback()
	intakeID, err := lockGenerationProject(ctx, tx, parentProjectID, false)
	if errors.Is(err, ErrProjectArchived) || errors.Is(err, sql.ErrNoRows) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	r, err := scanGenerationRun(tx.QueryRowContext(ctx, `SELECT `+generationRunColumns+` FROM runs WHERE id=? FOR UPDATE`, parentRunID))
	if errors.Is(err, sql.ErrNoRows) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	if r.RunKind != "generation" || r.BatchProjectID != parentProjectID || r.Status != RunRunning {
		return WorkItem{}, false, nil
	}
	snapshot, err := decodeGenerationSnapshot(r.RequestSchemaVersion, r.RequestSnapshot, r.RequestHash, r.BatchProjectID, r.RequestedByUserID)
	if errors.Is(err, ErrInvalidGenerationRequest) {
		return WorkItem{}, false, nil
	}
	if err != nil {
		return WorkItem{}, false, err
	}
	if err := validateFrozenBooks(ctx, tx, intakeID, snapshot.BookIDs); err != nil {
		if errors.Is(err, ErrInvalidGenerationRequest) {
			return WorkItem{}, false, nil
		}
		return WorkItem{}, false, err
	}
	var runID sql.NullInt64
	var projectID, bookID int64
	var attempt, maxAttempts int
	var retryable bool
	var status string
	var deadline sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status,lease_deadline FROM book_runs WHERE id=? FOR UPDATE`, id).Scan(&runID, &projectID, &bookID, &attempt, &maxAttempts, &retryable, &status, &deadline); err != nil {
		return WorkItem{}, false, err
	}
	if !runID.Valid || runID.Int64 != parentRunID || projectID != parentProjectID || status != "running" || !deadline.Valid || deadline.Time.After(now) {
		return WorkItem{}, false, nil
	}
	if !retryable || attempt >= maxAttempts {
		_, err := tx.ExecContext(ctx, `UPDATE book_runs SET status='failed',retryable=0,error_code='worker_lease_expired',error_message='worker lease expired',finished_at=?,lease_deadline=NULL WHERE id=? AND status='running'`, now, id)
		if err != nil {
			return WorkItem{}, false, err
		}
		if _, err := aggregateRunStatusTx(ctx, tx, parentRunID, now); err != nil {
			return WorkItem{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return WorkItem{}, false, err
		}
		return WorkItem{}, false, nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE book_runs SET status='failed',error_code='worker_lease_expired',error_message='worker lease expired',finished_at=?,lease_deadline=NULL WHERE id=? AND status='running'`, now, id); err != nil {
		return WorkItem{}, false, err
	}
	next := attempt + 1
	item, created, err := insertGenerationAttempt(ctx, tx, r, bookID, next, maxAttempts)
	if err != nil {
		return WorkItem{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return WorkItem{}, false, err
	}
	return item, created, nil
}

func (s *MySQLStore) bookRunParents(ctx context.Context, id int64) (int64, int64, error) {
	var runID sql.NullInt64
	var projectID int64
	if err := s.db.QueryRowContext(ctx, `SELECT run_id,batch_project_id FROM book_runs WHERE id=?`, id).Scan(&runID, &projectID); err != nil {
		return 0, 0, err
	}
	if !runID.Valid {
		return 0, 0, ErrBookRunNotRetryable
	}
	return runID.Int64, projectID, nil
}
func insertGenerationAttempt(ctx context.Context, tx *sql.Tx, r RunRecord, bookID int64, attempt, maxAttempts int) (WorkItem, bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM book_runs WHERE run_id=? AND book_id=? AND attempt=? FOR UPDATE`, r.ID, bookID, attempt).Scan(&id)
	created := errors.Is(err, sql.ErrNoRows)
	if err != nil && !created {
		return WorkItem{}, false, err
	}
	if created {
		res, err := tx.ExecContext(ctx, `INSERT INTO book_runs (run_id,batch_project_id,book_id,attempt,max_attempts,retryable,status,request_id) VALUES (?,?,?,?,?,1,'queued',?)`, r.ID, r.BatchProjectID, bookID, attempt, maxAttempts, r.IdempotencyKey)
		if err != nil {
			return WorkItem{}, false, err
		}
		id, err = res.LastInsertId()
		if err != nil {
			return WorkItem{}, false, err
		}
	}
	return WorkItem{BookRunID: id, BookID: bookID, Attempt: attempt}, created, nil
}
