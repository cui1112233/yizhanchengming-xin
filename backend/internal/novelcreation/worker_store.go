package novelcreation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"time"
)

const (
	RunPaused              RunStatus = "paused"
	ErrLeaseLost                     = "novel_lease_lost"
	ErrResultUnknown                 = "novel_upstream_result_unknown"
	ErrIdempotencyConflict           = "novel_idempotency_conflict"
)

type CallRecord struct {
	Fingerprint  string
	State        string
	Response     string
	ErrorCode    string
	ErrorMessage string
}
type Execution struct {
	Owner           string
	RunID           string
	ProjectID       string
	State           string
	Fingerprint     string
	Fence           int64
	CancelRequested bool
	LeaseUntil      time.Time
	Calls           map[string]CallRecord
	Bundle          *ChapterBundle
	CommittedHash   string
}
type ProjectSlot struct {
	RunID      string
	Fence      int64
	LeaseUntil time.Time
}
type RunLease struct {
	Owner, ProjectID, RunID string
	Fence                   int64
}

func fingerprint(value any) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func leaseLost() error {
	return &DomainError{Code: ErrLeaseLost, Message: "任务执行权已失效，已停止写入"}
}
func unknownResult() error {
	return &DomainError{Code: ErrResultUnknown, Message: "上游结果待确认，已暂停自动重试"}
}
func conflict() error {
	return &DomainError{Code: ErrIdempotencyConflict, Message: "同一请求编号对应的内容不同"}
}
func cloneExecution(e Execution) Execution {
	raw, _ := json.Marshal(e)
	var out Execution
	_ = json.Unmarshal(raw, &out)
	return out
}

func (s *MemoryStore) Clock(context.Context) (time.Time, error) { return time.Now().UTC(), nil }
func (s *MySQLStore) Clock(ctx context.Context) (time.Time, error) {
	var now time.Time
	err := s.db.QueryRowContext(ctx, `SELECT UTC_TIMESTAMP(6)`).Scan(&now)
	return now, persistenceError(err)
}

// Atomic serializes every decision for a project. The callback must not perform
// network I/O. Its Store is bound to this transaction; nested calls share it.
func (s *MemoryStore) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[memoryKey(owner, project)]
	if !ok || p.DeletedAt != nil {
		return &DomainError{Code: ErrProjectNotFound, Message: "project not found"}
	}
	c := NewMemoryStore()
	c.nextID, c.nextEventID = s.nextID, s.nextEventID
	c.projects = maps.Clone(s.projects)
	c.sessions = maps.Clone(s.sessions)
	c.messages = maps.Clone(s.messages)
	for k, v := range c.messages {
		c.messages[k] = append([]Message(nil), v...)
	}
	c.runs = maps.Clone(s.runs)
	c.runByKey = maps.Clone(s.runByKey)
	c.runEvents = maps.Clone(s.runEvents)
	for k, v := range c.runEvents {
		c.runEvents[k] = append([]RunEvent(nil), v...)
	}
	c.chapterBundles = maps.Clone(s.chapterBundles)
	c.chapterVersions = maps.Clone(s.chapterVersions)
	c.settings = maps.Clone(s.settings)
	c.foundations = maps.Clone(s.foundations)
	c.materials = maps.Clone(s.materials)
	c.notes = maps.Clone(s.notes)
	c.styles = maps.Clone(s.styles)
	c.schedules = maps.Clone(s.schedules)
	c.batches = maps.Clone(s.batches)
	c.executions = maps.Clone(s.executions)
	c.slots = maps.Clone(s.slots)
	if err := fn(c); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.nextID, s.nextEventID = c.nextID, c.nextEventID
	s.projects, s.sessions, s.messages, s.runs, s.runByKey, s.runEvents = c.projects, c.sessions, c.messages, c.runs, c.runByKey, c.runEvents
	s.chapterBundles, s.chapterVersions, s.settings, s.foundations = c.chapterBundles, c.chapterVersions, c.settings, c.foundations
	s.materials, s.notes, s.styles, s.schedules, s.batches = c.materials, c.notes, c.styles, c.schedules, c.batches
	s.executions, s.slots = c.executions, c.slots
	return nil
}
func (s *MemoryStore) GetExecution(_ context.Context, owner, id string) (Execution, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.executions[memoryKey(owner, id)]
	return cloneExecution(e), ok, nil
}
func (s *MemoryStore) PutExecution(_ context.Context, e Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[memoryKey(e.Owner, e.RunID)] = cloneExecution(e)
	return nil
}
func (s *MemoryStore) GetSlot(_ context.Context, owner, project string) (ProjectSlot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.slots[memoryKey(owner, project)], nil
}
func (s *MemoryStore) PutSlot(_ context.Context, owner, project string, slot ProjectSlot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slots[memoryKey(owner, project)] = slot
	return nil
}
func (s *MemoryStore) ListDispatch(_ context.Context, limit int) ([]Execution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Execution{}
	for key, r := range s.runs {
		if isTerminalRunStatus(r.Status) || r.Status == RunAwaitingConfirmation || r.Status == RunPaused {
			continue
		}
		owner := strings.SplitN(key, "\x00", 2)[0]
		e, ok := s.executions[key]
		if !ok {
			e = Execution{Owner: owner, RunID: r.ID, ProjectID: r.ProjectID, State: "legacy"}
		}
		if e.State == "paused" || e.State == "done" || e.LeaseUntil.After(time.Now().UTC()) {
			continue
		}
		out = append(out, cloneExecution(e))
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// sqlExecutor permits existing persistence methods to participate in Atomic.
type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	BeginTx(context.Context, *sql.TxOptions) (sqlTransaction, error)
}
type sqlTransaction interface {
	sqlExecutor
	Commit() error
	Rollback() error
}
type sqlDatabase struct{ *sql.DB }

func (d sqlDatabase) BeginTx(ctx context.Context, o *sql.TxOptions) (sqlTransaction, error) {
	tx, e := d.DB.BeginTx(ctx, o)
	if e != nil {
		return nil, e
	}
	return sqlBoundTx{Tx: tx, owns: true}, nil
}

type sqlBoundTx struct {
	*sql.Tx
	owns bool
}

func (t sqlBoundTx) BeginTx(context.Context, *sql.TxOptions) (sqlTransaction, error) {
	return sqlBoundTx{Tx: t.Tx}, nil
}
func (t sqlBoundTx) Commit() error {
	if t.owns {
		return t.Tx.Commit()
	}
	return nil
}
func (t sqlBoundTx) Rollback() error {
	if t.owns {
		return t.Tx.Rollback()
	}
	return nil
}
func (s *MySQLStore) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return persistenceError(err)
	}
	defer tx.Rollback()
	var one int
	if err = tx.QueryRowContext(ctx, `SELECT 1 FROM novel_projects WHERE owner_username = ? AND project_id = ? AND deleted_at IS NULL FOR UPDATE`, owner, project).Scan(&one); err != nil {
		return persistenceError(err)
	}
	if err = fn(&MySQLStore{db: tx}); err != nil {
		return err
	}
	return persistenceError(tx.Commit())
}
func (s *MySQLStore) GetExecution(ctx context.Context, owner, id string) (Execution, bool, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT execution_json FROM novel_run_dispatch WHERE owner_username = ? AND run_id = ?`, owner, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return Execution{}, false, nil
	}
	if err != nil {
		return Execution{}, false, persistenceError(err)
	}
	var e Execution
	if err = json.Unmarshal(raw, &e); err != nil {
		return e, false, persistenceError(err)
	}
	return e, true, nil
}
func (s *MySQLStore) PutExecution(ctx context.Context, e Execution) error {
	raw, err := json.Marshal(e)
	if err != nil {
		return persistenceError(err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO novel_run_dispatch (owner_username, run_id, project_id, state, lease_until, execution_json) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE state=VALUES(state), lease_until=VALUES(lease_until), execution_json=VALUES(execution_json)`, e.Owner, e.RunID, e.ProjectID, e.State, nullableLease(e.LeaseUntil), raw)
	if err != nil {
		return persistenceError(err)
	}
	for key, call := range e.Calls {
		snapshot, _ := json.Marshal(call)
		id := fingerprint([]string{e.RunID, key})
		_, err = s.db.ExecContext(ctx, `INSERT INTO novel_run_steps (step_id, owner_username, run_id, project_id, step_name, status, snapshot_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, UTC_TIMESTAMP(6), UTC_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE status=VALUES(status), snapshot_json=VALUES(snapshot_json), updated_at=VALUES(updated_at)`, id, e.Owner, e.RunID, e.ProjectID, key, call.State, snapshot)
		if err != nil {
			return persistenceError(err)
		}
	}
	return nil
}
func nullableLease(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
func (s *MySQLStore) GetSlot(ctx context.Context, owner, project string) (ProjectSlot, error) {
	var slot ProjectSlot
	var until sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT active_run_id, fence, lease_until FROM novel_project_slots WHERE owner_username = ? AND project_id = ?`, owner, project).Scan(&slot.RunID, &slot.Fence, &until)
	if errors.Is(err, sql.ErrNoRows) {
		return slot, nil
	}
	slot.LeaseUntil = until.Time
	return slot, persistenceError(err)
}
func (s *MySQLStore) PutSlot(ctx context.Context, owner, project string, slot ProjectSlot) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO novel_project_slots (owner_username, project_id, active_run_id, fence, lease_until) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE active_run_id=VALUES(active_run_id), fence=VALUES(fence), lease_until=VALUES(lease_until)`, owner, project, slot.RunID, slot.Fence, nullableLease(slot.LeaseUntil))
	return persistenceError(err)
}
func (s *MySQLStore) ListDispatch(ctx context.Context, limit int) ([]Execution, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.owner_username, r.run_id, r.project_id FROM novel_runs r JOIN novel_projects p ON p.owner_username=r.owner_username AND p.project_id=r.project_id LEFT JOIN novel_run_dispatch d ON d.owner_username=r.owner_username AND d.run_id=r.run_id WHERE p.deleted_at IS NULL AND r.status NOT IN ('succeeded','succeeded_with_warnings','failed','cancelled','awaiting_confirmation','paused') AND (d.run_id IS NULL OR (d.state IN ('ready','leased','save_pending') AND (d.lease_until IS NULL OR d.lease_until <= UTC_TIMESTAMP(6)))) ORDER BY r.created_at LIMIT ?`, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	out := []Execution{}
	for rows.Next() {
		var e Execution
		if err = rows.Scan(&e.Owner, &e.RunID, &e.ProjectID); err != nil {
			return nil, persistenceError(err)
		}
		out = append(out, e)
	}
	return out, persistenceError(rows.Err())
}
