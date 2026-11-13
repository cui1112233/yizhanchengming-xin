package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type VersionedProductionRepository interface {
	SetProductionRebuild(context.Context, string, string, bool) error
	GetShotTaskVersion(context.Context, string, string, string, int, string) (ShotTask, error)
}

func (s *MySQLStore) SetProductionRebuild(ctx context.Context, o, p string, rebuild bool) error {
	if err := s.SetProductionState(ctx, o, p, "paused", ""); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `UPDATE screenplay_production_jobs SET rebuild=? WHERE owner_username=? AND project_id=?`, rebuild, o, p)
	return err
}
func (s *MySQLStore) GetShotTaskVersion(ctx context.Context, o, p, b string, seq int, hash string) (ShotTask, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM screenplay_production_shots WHERE owner_username=? AND project_id=? AND storyboard_id=? AND shot_seq=? AND request_hash=?`, o, p, b, seq, hash).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return ShotTask{}, ErrNotFound
	}
	var task ShotTask
	if err == nil {
		err = json.Unmarshal(raw, &task)
	}
	return task, err
}
func (s *MySQLStore) SetProductionState(ctx context.Context, owner, id, status, message string) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO screenplay_production_jobs(owner_username,project_id,status,error,units,updated_at) VALUES(?,?,?,?,0,?) ON DUPLICATE KEY UPDATE status=VALUES(status),error=VALUES(error),updated_at=VALUES(updated_at),run_at=NULL`, owner, id, status, message, time.Now().UTC())
	return e
}
func (s *MySQLStore) GetProductionJob(ctx context.Context, owner, id string) (ProductionJob, error) {
	var j ProductionJob
	e := s.db.QueryRowContext(ctx, `SELECT owner_username,project_id,status,error,units,updated_at,rebuild,run_at FROM screenplay_production_jobs WHERE owner_username=? AND project_id=?`, owner, id).Scan(&j.Owner, &j.ProjectID, &j.Status, &j.Error, &j.Units, &j.UpdatedAt, &j.Rebuild, &j.RunAt)
	if errors.Is(e, sql.ErrNoRows) {
		e = ErrNotFound
	}
	return j, e
}
func (s *MySQLStore) ListRunningProductionJobs(ctx context.Context) ([]ProductionJob, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT owner_username,project_id,status,error,units,updated_at,rebuild,run_at FROM screenplay_production_jobs WHERE status IN ('running','scheduled') ORDER BY CASE WHEN status='running' OR run_at<=UTC_TIMESTAMP(6) THEN 0 ELSE 1 END,COALESCE(run_at,updated_at),updated_at LIMIT 256`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ProductionJob{}
	for rows.Next() {
		var j ProductionJob
		if e = rows.Scan(&j.Owner, &j.ProjectID, &j.Status, &j.Error, &j.Units, &j.UpdatedAt, &j.Rebuild, &j.RunAt); e != nil {
			return nil, e
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
func (s *MySQLStore) SaveProductionProgress(ctx context.Context, j ProductionJob) error {
	_, e := s.db.ExecContext(ctx, `UPDATE screenplay_production_jobs SET status=?,error=?,units=?,updated_at=? WHERE owner_username=? AND project_id=? AND status='running'`, j.Status, j.Error, j.Units, time.Now().UTC(), j.Owner, j.ProjectID)
	return e
}
func (s *MySQLStore) GetShotTask(ctx context.Context, owner, id, board string, seq int) (ShotTask, error) {
	var raw []byte
	e := s.db.QueryRowContext(ctx, `SELECT payload FROM screenplay_production_shots WHERE owner_username=? AND project_id=? AND storyboard_id=? AND shot_seq=? ORDER BY updated_at DESC LIMIT 1`, owner, id, board, seq).Scan(&raw)
	if errors.Is(e, sql.ErrNoRows) {
		return ShotTask{}, ErrNotFound
	}
	var t ShotTask
	if e == nil {
		e = json.Unmarshal(raw, &t)
	}
	return t, e
}
func (s *MySQLStore) SaveShotTask(ctx context.Context, t ShotTask) error {
	t.UpdatedAt = time.Now().UTC()
	raw, e := json.Marshal(t)
	if e != nil {
		return e
	}
	_, e = s.db.ExecContext(ctx, `INSERT INTO screenplay_production_shots(owner_username,project_id,storyboard_id,shot_seq,request_hash,payload,updated_at) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE payload=VALUES(payload),updated_at=VALUES(updated_at)`, t.Owner, t.ProjectID, t.StoryboardID, t.Seq, t.RequestHash, raw, t.UpdatedAt)
	return e
}
func (s *MySQLStore) ListShotTasks(ctx context.Context, owner, id string) ([]ShotTask, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT payload FROM screenplay_production_shots WHERE owner_username=? AND project_id=? ORDER BY storyboard_id,shot_seq`, owner, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ShotTask{}
	for rows.Next() {
		var raw []byte
		var t ShotTask
		if e = rows.Scan(&raw); e != nil {
			return nil, e
		}
		if e = json.Unmarshal(raw, &t); e != nil {
			return nil, e
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// MemoryProductionRepository is for deterministic tests and local fixtures. The
// server explicitly requires the MySQL extension; it never falls back to this.
type MemoryProductionRepository struct {
	mu    sync.Mutex
	jobs  map[string]ProductionJob
	shots map[string]ShotTask
}

func NewMemoryProductionRepository() *MemoryProductionRepository {
	return &MemoryProductionRepository{jobs: map[string]ProductionJob{}, shots: map[string]ShotTask{}}
}
func productionKey(o, p string) string { return o + "\x00" + p }
func (r *MemoryProductionRepository) SetProductionState(_ context.Context, o, p, s, e string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j := r.jobs[productionKey(o, p)]
	j.Owner = o
	j.ProjectID = p
	j.Status = s
	j.RunAt = nil
	j.Error = e
	j.UpdatedAt = time.Now().UTC()
	r.jobs[productionKey(o, p)] = j
	return nil
}
func (r *MemoryProductionRepository) GetProductionJob(_ context.Context, o, p string) (ProductionJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[productionKey(o, p)]
	if !ok {
		return j, ErrNotFound
	}
	return j, nil
}
func (r *MemoryProductionRepository) ListRunningProductionJobs(_ context.Context) ([]ProductionJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []ProductionJob{}
	for _, j := range r.jobs {
		if j.Status == "running" || j.Status == "scheduled" {
			out = append(out, j)
		}
	}
	return out, nil
}
func (r *MemoryProductionRepository) SaveProductionProgress(_ context.Context, j ProductionJob) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := productionKey(j.Owner, j.ProjectID)
	if r.jobs[k].Status == "running" {
		j.UpdatedAt = time.Now().UTC()
		r.jobs[k] = j
	}
	return nil
}
func shotTaskKey(o, p, b string, s int) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%d", o, p, b, s)
}
func (r *MemoryProductionRepository) SetProductionRebuild(ctx context.Context, o, p string, rebuild bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	j := r.jobs[productionKey(o, p)]
	j.Owner = o
	j.ProjectID = p
	j.Rebuild = rebuild
	r.jobs[productionKey(o, p)] = j
	return nil
}
func (r *MemoryProductionRepository) GetShotTaskVersion(_ context.Context, o, p, b string, seq int, hash string) (ShotTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.shots[shotTaskKey(o, p, b, seq)+"\x00"+hash]
	if !ok {
		return t, ErrNotFound
	}
	return t, nil
}
func (r *MemoryProductionRepository) GetShotTask(_ context.Context, o, p, b string, seq int) (ShotTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var found ShotTask
	ok := false
	for _, t := range r.shots {
		if t.Owner == o && t.ProjectID == p && t.StoryboardID == b && t.Seq == seq && (!ok || t.UpdatedAt.After(found.UpdatedAt)) {
			found = t
			ok = true
		}
	}
	if !ok {
		return found, ErrNotFound
	}
	return found, nil
}
func (r *MemoryProductionRepository) SaveShotTask(_ context.Context, t ShotTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t.UpdatedAt = time.Now().UTC()
	r.shots[shotTaskKey(t.Owner, t.ProjectID, t.StoryboardID, t.Seq)+"\x00"+t.RequestHash] = t
	return nil
}
func (r *MemoryProductionRepository) ListShotTasks(_ context.Context, o, p string) ([]ShotTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []ShotTask{}
	for _, t := range r.shots {
		if t.Owner == o && t.ProjectID == p {
			out = append(out, t)
		}
	}
	return out, nil
}
