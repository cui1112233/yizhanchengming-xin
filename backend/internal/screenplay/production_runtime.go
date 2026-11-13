package screenplay

import (
	"context"
	"errors"
	"fmt"
	"qiantie/backend/internal/mergeworker"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ProductionJob struct {
	RunAt        *time.Time `json:"run_at,omitempty"`
	Rebuild      bool       `json:"rebuild"`
	Owner        string     `json:"owner"`
	ProjectID    string     `json:"project_id"`
	Status       string     `json:"status"`
	Error        string     `json:"error,omitempty"`
	Units        int        `json:"units"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Batch        *Batch     `json:"batch,omitempty"`
	Shots        []ShotTask `json:"shots,omitempty"`
	Logs         []BatchLog `json:"logs,omitempty"`
	Cancellation string     `json:"cancellation,omitempty"`
}
type ShotTask struct {
	Owner        string    `json:"owner"`
	ProjectID    string    `json:"project_id"`
	StoryboardID string    `json:"storyboard_id"`
	ShotID       string    `json:"shot_id"`
	Episode      int       `json:"episode"`
	Seq          int       `json:"seq"`
	TaskID       string    `json:"task_id"`
	Status       string    `json:"status"`
	Artifact     string    `json:"artifact,omitempty"`
	Error        string    `json:"error,omitempty"`
	RequestHash  string    `json:"request_hash"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type ProductionRepository interface {
	SetProductionState(context.Context, string, string, string, string) error
	GetProductionJob(context.Context, string, string) (ProductionJob, error)
	ListRunningProductionJobs(context.Context) ([]ProductionJob, error)
	SaveProductionProgress(context.Context, ProductionJob) error
	GetShotTask(context.Context, string, string, string, int) (ShotTask, error)
	SaveShotTask(context.Context, ShotTask) error
	ListShotTasks(context.Context, string, string) ([]ShotTask, error)
}

var ErrClipsReady = errors.New("镜头视频已就绪，尚未生成可交付成片")

type ProductionFinalizer interface {
	Finalize(context.Context, Project) ([]string, error)
}
type ProductionRuntime struct {
	Store      Store
	Repository ProductionRepository
	RedisDo    mergeworker.RedisDoFunc
	Executor   func(string) BatchExecutor
	Interval   time.Duration
}

func (r *ProductionRuntime) Start(ctx context.Context, owner, id string) (ProductionJob, error) {
	return r.StartWithOptions(ctx, owner, id, false)
}
func (r *ProductionRuntime) StartWithOptions(ctx context.Context, owner, id string, rebuild bool) (ProductionJob, error) {
	return r.prepareProduction(ctx, owner, id, rebuild, nil)
}
func (r *ProductionRuntime) prepareProduction(ctx context.Context, owner, id string, rebuild bool, runAt *time.Time) (ProductionJob, error) {
	if r == nil || r.Store == nil || r.Repository == nil {
		return ProductionJob{}, fmt.Errorf("%w: 后台生产存储未配置", ErrInvalidState)
	}
	project, e := r.Store.GetProject(ctx, owner, id)
	if e != nil {
		return ProductionJob{}, e
	}
	if r.RedisDo == nil || r.Executor == nil {
		return ProductionJob{}, fmt.Errorf("%w: 后台生产需要可用的 Redis 和执行器", ErrInvalidState)
	}
	if _, err := r.RedisDo(ctx, "PING"); err != nil {
		return ProductionJob{}, fmt.Errorf("%w: Redis 连接失败，任务尚未启动", ErrInvalidState)
	}
	token := newScreenplayID("start")
	lockKey := "screenplay-production-control:" + owner + ":" + id
	if err := r.Store.AcquireLock(ctx, lockKey, token, 30*time.Second); err != nil {
		return ProductionJob{}, err
	}
	defer func() { _ = r.Store.ReleaseLock(context.Background(), lockKey, token) }()
	if j, e := r.Repository.GetProductionJob(ctx, owner, id); e == nil && j.Status == "running" {
		if runAt != nil {
			return ProductionJob{}, fmt.Errorf("%w: 任务正在执行，请先暂停再安排定时", ErrInvalidState)
		}
		return r.Status(ctx, owner, id)
	}
	b, e := r.Store.GetBatch(ctx, owner, id)
	if errors.Is(e, ErrNotFound) {
		e = r.Store.CreateBatch(ctx, NewBatch(newScreenplayID("batch"), id, owner, "production", project.PermissionMode))
	} else if e == nil {
		if rebuild {
			b, e = r.Store.MutateBatch(ctx, owner, id, func(b *Batch) error {
				MarkStale(b, []string{"7_video", "8_finalize"})
				for _, key := range []string{"7_video", "8_finalize"} {
					st := b.Step(key)
					if st == nil {
						continue
					}
					st.Artifacts = nil
					st.Error = ""
					for _, ep := range st.Episodes {
						ep.Status = StepStale
						ep.Artifact = ""
						ep.Reason = ""
					}
				}
				b.Status = BatchRunning
				return nil
			})
			if e != nil {
				return ProductionJob{}, e
			}
		}
		if b.Status == BatchCompleted {
			final := b.Step("8_finalize")
			valid := final != nil && len(final.Artifacts) > 0
			if valid {
				for _, a := range final.Artifacts {
					if !(strings.HasPrefix(a, "https://") || strings.HasPrefix(a, "/api/")) {
						valid = false
					}
				}
			}
			if valid {
				if runAt != nil {
					return ProductionJob{}, fmt.Errorf("%w: 已有交付结果，请先修改内容或选择重建", ErrInvalidState)
				}
				if err := r.Repository.SetProductionState(ctx, owner, id, "completed", ""); err != nil {
					return ProductionJob{}, err
				}
				return r.Status(ctx, owner, id)
			}
			_, e = r.Store.MutateBatch(ctx, owner, id, func(b *Batch) error {
				b.Step("8_finalize").Status = StepPending
				b.Step("8_finalize").Artifacts = nil
				b.Status = BatchRunning
				return nil
			})
			if e != nil {
				return ProductionJob{}, e
			}
		}
		_, e = r.Store.MutateBatch(ctx, owner, id, func(b *Batch) error {
			if len(BlockedSteps(b)) > 0 {
				return fmt.Errorf("%w: 质检阻断尚未解除", ErrInvalidState)
			}
			b.Status = BatchRunning
			return nil
		})
	}
	if e != nil {
		return ProductionJob{}, e
	}
	if versions, ok := r.Repository.(VersionedProductionRepository); ok {
		if e = versions.SetProductionRebuild(ctx, owner, id, rebuild); e != nil {
			return ProductionJob{}, e
		}
	} else if rebuild {
		return ProductionJob{}, fmt.Errorf("%w: 当前存储不支持保留历史的重建", ErrInvalidState)
	}
	if runAt != nil {
		if _, e = r.Store.MutateBatch(ctx, owner, id, func(b *Batch) error { b.Status = BatchPaused; return nil }); e != nil {
			return ProductionJob{}, e
		}
		if e = r.Repository.(ScheduledProductionRepository).SetProductionSchedule(ctx, owner, id, *runAt); e != nil {
			return ProductionJob{}, e
		}
		return r.Status(ctx, owner, id)
	}
	if e = r.Repository.SetProductionState(ctx, owner, id, "running", ""); e != nil {
		return ProductionJob{}, e
	}
	return r.Status(ctx, owner, id)
}
func (r *ProductionRuntime) Status(ctx context.Context, owner, id string) (ProductionJob, error) {
	if _, e := r.Store.GetProject(ctx, owner, id); e != nil {
		return ProductionJob{}, e
	}
	j, e := r.Repository.GetProductionJob(ctx, owner, id)
	if errors.Is(e, ErrNotFound) {
		j = ProductionJob{Owner: owner, ProjectID: id, Status: "idle"}
	} else if e != nil {
		return j, e
	}
	j.Batch, _ = r.Store.GetBatch(ctx, owner, id)
	j.Shots, _ = r.Repository.ListShotTasks(ctx, owner, id)
	if j.Batch != nil {
		j.Logs, _ = r.Store.ListBatchLogs(ctx, j.Batch.ID, owner, 100)
	}
	if j.Status == "paused" {
		j.Cancellation = "已停止后续提交；供应商不支持取消，已提交任务会在继续时查询原任务"
	}
	return j, nil
}
func (r *ProductionRuntime) Pause(ctx context.Context, owner, id string) (ProductionJob, error) {
	token := newScreenplayID("pause")
	lockKey := "screenplay-production-control:" + owner + ":" + id
	if err := r.Store.AcquireLock(ctx, lockKey, token, 30*time.Second); err != nil {
		return ProductionJob{}, err
	}
	defer func() { _ = r.Store.ReleaseLock(context.Background(), lockKey, token) }()
	if _, e := r.Store.GetProject(ctx, owner, id); e != nil {
		return ProductionJob{}, e
	}
	if e := r.Repository.SetProductionState(ctx, owner, id, "paused", ""); e != nil {
		return ProductionJob{}, e
	}
	_, e := r.Store.MutateBatch(ctx, owner, id, func(b *Batch) error { b.Status = BatchPaused; return nil })
	if e != nil {
		return ProductionJob{}, e
	}
	return r.Status(ctx, owner, id)
}

// Run belongs to the server lifetime, never the start HTTP request. MySQL is the
// recovery journal; Redis schedules and leases one project across processes.
func (r *ProductionRuntime) Run(ctx context.Context) error {
	if r.Repository == nil || r.RedisDo == nil || r.Executor == nil {
		return fmt.Errorf("后台生产需要 MySQL、Redis 和执行器")
	}
	interval := r.Interval
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var wg sync.WaitGroup
	defer wg.Wait()
	slots := make(chan struct{}, 4)
	for {
		jobs, e := r.Repository.ListRunningProductionJobs(ctx)
		if e != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

		} else {
			scheduled := map[string]ProductionJob{}
			for _, j := range jobs {
				member := j.Owner + ":" + j.ProjectID
				scheduled[member] = j
				if j.Status == "scheduled" {
					_, _ = r.RedisDo(ctx, "ZADD", "qiantie:screenplay:production:due", productionDueScore(j), member)
				} else {
					_, _ = r.RedisDo(ctx, "EVAL", `local old=redis.call('ZSCORE',KEYS[1],ARGV[2]); if not old or tonumber(old)>tonumber(ARGV[1]) then return redis.call('ZADD',KEYS[1],ARGV[1],ARGV[2]) else return 0 end`, "1", "qiantie:screenplay:production:due", productionDueScore(j), member)
				}
			}
			due, err := r.RedisDo(ctx, "ZRANGEBYSCORE", "qiantie:screenplay:production:due", "-inf", strconv.FormatInt(time.Now().UnixMilli(), 10), "LIMIT", "0", "256")
			if err == nil {
				if members, ok := due.([]any); ok {
					for _, member := range members {
						key := fmt.Sprint(member)
						j, ok := scheduled[key]
						if !ok {
							_, _ = r.RedisDo(ctx, "ZREM", "qiantie:screenplay:production:due", key)
							continue
						}
						select {
						case slots <- struct{}{}:
							wg.Add(1)
							go func(j ProductionJob) { defer wg.Done(); defer func() { <-slots }(); _ = r.RunOne(ctx, j) }(j)
						default:
						}
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
func (r *ProductionRuntime) RunOne(ctx context.Context, j ProductionJob) error {
	key := "qiantie:screenplay:production:lease:" + j.Owner + ":" + j.ProjectID
	token := newScreenplayID("lease")
	v, e := r.RedisDo(ctx, "SET", key, token, "NX", "PX", "30000")
	if e != nil {
		return e
	}
	if v == nil {
		return nil
	}
	defer func() {
		_, _ = r.RedisDo(context.Background(), "EVAL", `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',KEYS[1]) else return 0 end`, "1", key, token)
	}()
	_, _ = r.RedisDo(ctx, "ZREM", "qiantie:screenplay:production:due", j.Owner+":"+j.ProjectID)
	work, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{})
	defer close(done)

	current, e := r.Repository.GetProductionJob(work, j.Owner, j.ProjectID)
	if e == nil && current.Status == "scheduled" {
		repo, ok := r.Repository.(ScheduledProductionRepository)
		if !ok {
			return ErrInvalidState
		}
		control := "screenplay-production-control:" + j.Owner + ":" + j.ProjectID
		if err := r.Store.AcquireLock(work, control, token, 30*time.Second); err != nil {
			return err
		}
		claimed, claimErr := repo.ClaimDueProduction(work, j.Owner, j.ProjectID, time.Now())
		_ = r.Store.ReleaseLock(context.Background(), control, token)
		if claimErr != nil {
			return claimErr
		}
		if !claimed {
			return nil
		}
		current, e = r.Repository.GetProductionJob(work, j.Owner, j.ProjectID)
	}
	if e != nil || current.Status != "running" {
		return e
	}
	// Claiming a due job may wait on MySQL. Renew and verify our Redis token
	// before starting any work, so an expired claimant cannot submit media.
	lease, leaseErr := r.RedisDo(work, "EVAL", `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('PEXPIRE',KEYS[1],ARGV[2]) else return 0 end`, "1", key, token, "30000")
	if leaseErr != nil {
		return leaseErr
	}
	if fmt.Sprint(lease) != "1" {
		return ErrLockBusy
	}
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-work.Done():
				return
			case <-t.C:
				current, err := r.Repository.GetProductionJob(work, j.Owner, j.ProjectID)
				if err != nil || current.Status != "running" {
					cancel()
					return
				}
				v, err := r.RedisDo(work, "EVAL", `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('PEXPIRE',KEYS[1],ARGV[2]) else return 0 end`, "1", key, token, "30000")
				if err != nil || fmt.Sprint(v) != "1" {
					cancel()
					return
				}
			}
		}
	}()
	// A server shutdown may pause the batch while the durable job stays running.
	// Only a durable running job is allowed to recover that paused unit.
	if b, err := r.Store.GetBatch(work, j.Owner, j.ProjectID); err == nil && b.Status == BatchPaused {
		_, e = r.Store.MutateBatch(work, j.Owner, j.ProjectID, func(b *Batch) error { b.Status = BatchRunning; return nil })
		if e != nil {
			return e
		}
	}
	executor := r.Executor(j.Owner)
	if executor.Generator != nil {
		executor.Generator = productionGuardGenerator{generator: executor.Generator, repository: r.Repository, owner: j.Owner, projectID: j.ProjectID}
	}
	result, e := executor.RunNext(work, j.Owner, j.ProjectID)
	if work.Err() != nil {
		return work.Err()
	}
	current.Units++
	current.Error = ""
	switch {
	case errors.Is(e, ErrClipsReady):
		current.Status = "clips_ready"
		current.Error = e.Error()
	case e != nil:
		current.Status = "failed"
		current.Error = e.Error()
	case result.Batch != nil && result.Batch.Status == BatchCompleted:
		current.Status = "completed"
	case result.Batch != nil && result.Batch.Status == BatchWaitingUser:
		current.Status = "blocked"
	}
	if persist := r.Repository.SaveProductionProgress(work, current); persist != nil {
		return persist
	}
	return e
}

type productionGuardGenerator struct {
	generator        TextGenerator
	repository       ProductionRepository
	owner, projectID string
}

func (g productionGuardGenerator) Generate(ctx context.Context, request TextRequest) (TextResponse, error) {
	if err := ctx.Err(); err != nil {
		return TextResponse{}, err
	}
	job, err := g.repository.GetProductionJob(ctx, g.owner, g.projectID)
	if err != nil {
		return TextResponse{}, err
	}
	if job.Status != "running" {
		return TextResponse{}, ErrUserCancelled
	}
	return g.generator.Generate(ctx, request)
}

func productionDueScore(j ProductionJob) string {
	at := time.Now()
	if j.Status == "scheduled" && j.RunAt != nil {
		at = *j.RunAt
	}
	return strconv.FormatInt(at.UnixMilli(), 10)
}
