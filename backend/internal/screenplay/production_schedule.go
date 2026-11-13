package screenplay

import (
	"context"
	"fmt"
	"time"
)

type ScheduledProductionRepository interface {
	SetProductionSchedule(context.Context, string, string, time.Time) error
	ClaimDueProduction(context.Context, string, string, time.Time) (bool, error)
	CancelProductionSchedule(context.Context, string, string) error
}

func (s *MySQLStore) SetProductionSchedule(ctx context.Context, o, p string, at time.Time) error {
	_, e := s.db.ExecContext(ctx, `UPDATE screenplay_production_jobs SET status='scheduled',run_at=?,error='',updated_at=? WHERE owner_username=? AND project_id=?`, at.UTC(), time.Now().UTC(), o, p)
	return e
}
func (s *MySQLStore) ClaimDueProduction(ctx context.Context, o, p string, now time.Time) (bool, error) {
	result, e := s.db.ExecContext(ctx, `UPDATE screenplay_production_jobs SET status='running',updated_at=? WHERE owner_username=? AND project_id=? AND status='scheduled' AND run_at<=?`, now.UTC(), o, p, now.UTC())
	if e != nil {
		return false, e
	}
	n, e := result.RowsAffected()
	return n == 1, e
}
func (s *MySQLStore) CancelProductionSchedule(ctx context.Context, o, p string) error {
	result, e := s.db.ExecContext(ctx, `UPDATE screenplay_production_jobs SET status='paused',run_at=NULL,updated_at=? WHERE owner_username=? AND project_id=? AND status='scheduled'`, time.Now().UTC(), o, p)
	if e != nil {
		return e
	}
	n, e := result.RowsAffected()
	if e == nil && n == 0 {
		return fmt.Errorf("%w: 定时任务已启动或已取消，请刷新状态", ErrInvalidState)
	}
	return e
}
func (r *MemoryProductionRepository) SetProductionSchedule(_ context.Context, o, p string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := productionKey(o, p)
	j := r.jobs[k]
	j.Owner = o
	j.ProjectID = p
	j.Status = "scheduled"
	at = at.UTC()
	j.RunAt = &at
	j.UpdatedAt = time.Now().UTC()
	r.jobs[k] = j
	return nil
}
func (r *MemoryProductionRepository) ClaimDueProduction(_ context.Context, o, p string, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := productionKey(o, p)
	j := r.jobs[k]
	if j.Status != "scheduled" || j.RunAt == nil || j.RunAt.After(now) {
		return false, nil
	}
	j.Status = "running"
	j.UpdatedAt = now
	r.jobs[k] = j
	return true, nil
}
func (r *MemoryProductionRepository) CancelProductionSchedule(_ context.Context, o, p string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := productionKey(o, p)
	j := r.jobs[k]
	if j.Status != "scheduled" {
		return ErrInvalidState
	}
	j.Status = "paused"
	j.RunAt = nil
	j.UpdatedAt = time.Now().UTC()
	r.jobs[k] = j
	return nil
}
func (r *ProductionRuntime) Schedule(ctx context.Context, owner, id string, at time.Time, rebuild bool) (ProductionJob, error) {
	if !at.After(time.Now()) {
		return ProductionJob{}, fmt.Errorf("%w: 定时制作必须选择未来时间", ErrInvalidArgument)
	}
	if _, ok := r.Repository.(ScheduledProductionRepository); !ok {
		return ProductionJob{}, fmt.Errorf("%w: 定时制作存储未配置", ErrInvalidState)
	}
	return r.prepareProduction(ctx, owner, id, rebuild, &at)
}
func (r *ProductionRuntime) CancelSchedule(ctx context.Context, owner, id string) (ProductionJob, error) {
	if _, e := r.Store.GetProject(ctx, owner, id); e != nil {
		return ProductionJob{}, e
	}
	repo, ok := r.Repository.(ScheduledProductionRepository)
	if !ok {
		return ProductionJob{}, ErrInvalidState
	}
	token := newScreenplayID("unschedule")
	key := "screenplay-production-control:" + owner + ":" + id
	if e := r.Store.AcquireLock(ctx, key, token, 30*time.Second); e != nil {
		return ProductionJob{}, e
	}
	defer r.Store.ReleaseLock(context.Background(), key, token)
	if e := repo.CancelProductionSchedule(ctx, owner, id); e != nil {
		return ProductionJob{}, e
	}
	return r.Status(ctx, owner, id)
}
