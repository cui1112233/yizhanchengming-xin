package screenplay

import (
	"context"
	"testing"
	"time"
)

func TestProductionScheduleFutureValidationCancelAndRestart(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	repo := NewMemoryProductionRepository()
	r := &ProductionRuntime{Store: s, Repository: repo, RedisDo: newProductionTestRedis(), Executor: func(string) BatchExecutor { t.Fatal("scheduled future task executed"); return BatchExecutor{} }}
	if _, e := r.Schedule(ctx, "a", "p", time.Now().Add(-time.Minute), false); e == nil {
		t.Fatal("past time accepted")
	}
	if _, e := s.GetBatch(ctx, "a", "p"); e == nil {
		t.Fatal("invalid time created batch")
	}
	when := time.Now().Add(time.Hour).UTC()
	j, e := r.Schedule(ctx, "a", "p", when, false)
	if e != nil || j.Status != "scheduled" || j.RunAt == nil {
		t.Fatalf("%+v %v", j, e)
	}
	r = &ProductionRuntime{Store: s, Repository: repo, RedisDo: newProductionTestRedis(), Executor: r.Executor}
	if e = r.RunOne(ctx, j); e != nil {
		t.Fatal(e)
	}
	j, e = r.CancelSchedule(ctx, "a", "p")
	if e != nil || j.Status != "paused" || j.RunAt != nil {
		t.Fatalf("cancel=%+v %v", j, e)
	}
}
func TestProductionScheduleClaimIsDueAndAtomic(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryProductionRepository()
	future := time.Now().Add(time.Hour)
	_ = repo.SetProductionSchedule(ctx, "a", "p", future)
	if claimed, e := repo.ClaimDueProduction(ctx, "a", "p", time.Now()); e != nil || claimed {
		t.Fatalf("earlyclaim=%v %v", claimed, e)
	}
	now := future.Add(time.Second)
	if claimed, e := repo.ClaimDueProduction(ctx, "a", "p", now); e != nil || !claimed {
		t.Fatalf("dueclaim=%v %v", claimed, e)
	}
	if claimed, _ := repo.ClaimDueProduction(ctx, "a", "p", now); claimed {
		t.Fatal("duplicate claim")
	}
	if e := repo.CancelProductionSchedule(ctx, "a", "p"); e == nil {
		t.Fatal("cancel schedule paused alreadyrunning task")
	}
}
func TestProductionScheduledWorkerRecoversAtDueTimeAfterRestart(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	_, _ = s.SaveProjectSource(ctx, "a", "p", "真实内容")
	b := NewBatch("b", "p", "a", "s", PermissionBypass)
	for _, st := range b.Steps {
		st.Status = StepDone
	}
	b.Step("0_ingest").Status = StepPending
	b.Status = BatchPaused
	_ = s.CreateBatch(ctx, b)
	repo := NewMemoryProductionRepository()
	r := &ProductionRuntime{Store: s, Repository: repo, RedisDo: newProductionTestRedis(), Executor: func(string) BatchExecutor { return BatchExecutor{Store: s} }}
	at := time.Now().Add(250 * time.Millisecond)
	if _, e := r.Schedule(ctx, "a", "p", at, false); e != nil {
		t.Fatal(e)
	}
	// Replace both the runtime and its Redis due set; MySQL journal rehydrates it.
	r = &ProductionRuntime{Store: s, Repository: repo, RedisDo: newProductionTestRedis(), Interval: 5 * time.Millisecond, Executor: r.Executor}
	worker, stop := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- r.Run(worker) }()
	defer func() { stop(); <-done }()
	time.Sleep(30 * time.Millisecond)
	j, _ := repo.GetProductionJob(ctx, "a", "p")
	if j.Status != "scheduled" || j.Units != 0 {
		t.Fatalf("early execution %+v", j)
	}
	deadline := time.After(2 * time.Second)
	for {
		j, _ = repo.GetProductionJob(ctx, "a", "p")
		if j.Status == "completed" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("not resumed when due %+v", j)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

type slowClaimProductionRepository struct {
	*MemoryProductionRepository
	release chan struct{}
}

func (r slowClaimProductionRepository) ClaimDueProduction(ctx context.Context, o, p string, now time.Time) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-r.release:
		return r.MemoryProductionRepository.ClaimDueProduction(ctx, o, p, now)
	}
}
func TestProductionSlowDueClaimDoesNotCancelBeforeWatchdogStarts(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
	_, _ = s.SaveProjectSource(ctx, "a", "p", "正文")
	b := NewBatch("b", "p", "a", "", PermissionBypass)
	for _, st := range b.Steps {
		st.Status = StepDone
	}
	b.Step("0_ingest").Status = StepPending
	b.Status = BatchPaused
	_ = s.CreateBatch(ctx, b)
	repo := slowClaimProductionRepository{MemoryProductionRepository: NewMemoryProductionRepository(), release: make(chan struct{})}
	_ = repo.SetProductionSchedule(ctx, "a", "p", time.Now().Add(-time.Minute))
	r := &ProductionRuntime{Store: s, Repository: repo, RedisDo: newProductionTestRedis(), Executor: func(string) BatchExecutor { return BatchExecutor{Store: s} }}
	j, _ := repo.GetProductionJob(ctx, "a", "p")
	done := make(chan error, 1)
	go func() { done <- r.RunOne(ctx, j) }()
	time.Sleep(1100 * time.Millisecond)
	close(repo.release)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	j, _ = repo.GetProductionJob(ctx, "a", "p")
	if j.Status != "completed" {
		t.Fatal(j)
	}
}
