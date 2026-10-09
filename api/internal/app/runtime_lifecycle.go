package app

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type RuntimeStatus struct {
	State      string
	ReasonCode string
}

type runtimeWorker interface {
	RunSupervised(context.Context, taskruntime.Queue, time.Duration, func(task9runtime.RetryEvent)) error
}

type runtimeScheduler interface {
	Tick(context.Context, int) (int, error)
}

type runtimeRecovery interface {
	Rebuild(context.Context, time.Time, int) (int, error)
}

type runtimeLifecycleOptions struct {
	Configured        bool
	RedisConfigured   bool
	UnavailableReason string
	Queue             taskruntime.Queue
	Worker            runtimeWorker
	Scheduler         runtimeScheduler
	Recovery          runtimeRecovery
	HealthCheck       func(context.Context) error
	Closers           []io.Closer
	WorkerPoll        time.Duration
	SchedulerInterval time.Duration
	RecoveryInterval  time.Duration
	BatchLimit        int
}

type generationRuntimeLifecycle struct {
	options runtimeLifecycleOptions

	mu            sync.RWMutex
	status        RuntimeStatus
	cancel        context.CancelFunc
	done          chan struct{}
	start         sync.Once
	workers       sync.Once
	close         sync.Once
	wg            sync.WaitGroup
	closeErr      error
	workerUp      bool
	workerStopped bool
}

func newGenerationRuntimeLifecycle(options runtimeLifecycleOptions) *generationRuntimeLifecycle {
	if options.WorkerPoll <= 0 {
		options.WorkerPoll = time.Second
	}
	if options.SchedulerInterval <= 0 {
		options.SchedulerInterval = time.Second
	}
	if options.RecoveryInterval <= 0 {
		options.RecoveryInterval = 30 * time.Second
	}
	if options.BatchLimit <= 0 {
		options.BatchLimit = 100
	}
	if options.UnavailableReason == "" {
		options.UnavailableReason = "not_configured"
	}
	return &generationRuntimeLifecycle{
		options: options,
		status:  RuntimeStatus{State: "unavailable", ReasonCode: options.UnavailableReason},
		done:    make(chan struct{}),
	}
}

func (r *generationRuntimeLifecycle) RuntimeDiagnostics() httpapi.RuntimeDiagnosticsStatus {
	status := r.Status()
	redisStatus := "not_configured"
	if r.options.RedisConfigured {
		redisStatus = "configured"
	}
	if status.ReasonCode == "redis_unavailable" {
		redisStatus = "unavailable"
	}
	return httpapi.RuntimeDiagnosticsStatus{Ready: r.Ready(), Status: status.State, Redis: redisStatus, ReasonCode: status.ReasonCode}
}

func (r *generationRuntimeLifecycle) Ready() bool { return r.Status().State == "available" }

func (r *generationRuntimeLifecycle) Status() RuntimeStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

func (r *generationRuntimeLifecycle) setStatus(state, reason string) {
	r.mu.Lock()
	r.status = RuntimeStatus{State: state, ReasonCode: reason}
	r.mu.Unlock()
}

func (r *generationRuntimeLifecycle) Start(parent context.Context) error {
	r.start.Do(func() {
		if !r.options.Configured {
			return
		}
		if r.options.Queue == nil || r.options.Worker == nil || r.options.Scheduler == nil || r.options.Recovery == nil {
			r.setStatus("degraded", "runtime_setup_failed")
			return
		}
		ctx, cancel := context.WithCancel(parent)
		r.mu.Lock()
		r.cancel = cancel
		r.mu.Unlock()

		if _, err := r.options.Recovery.Rebuild(ctx, time.Now(), r.options.BatchLimit); err != nil {
			r.setStatus("degraded", "recovery_failed")
		} else if !r.redisHealthy(ctx) {
			r.setStatus("degraded", "redis_unavailable")
		} else {
			r.startWorkers(ctx)
			r.setAvailableIfWorkerRunning()
		}
		r.wg.Add(1)
		go r.runRecovery(ctx)
	})
	return nil
}

func (r *generationRuntimeLifecycle) startWorkers(ctx context.Context) {
	r.workers.Do(func() {
		r.mu.Lock()
		r.workerUp = true
		r.workerStopped = false
		r.mu.Unlock()
		r.wg.Add(2)
		go func() {
			defer r.wg.Done()
			_ = r.options.Worker.RunSupervised(ctx, r.options.Queue, r.options.WorkerPoll, func(task9runtime.RetryEvent) {
				r.setStatus("degraded", "worker_retrying")
			})
			if ctx.Err() == nil {
				r.mu.Lock()
				r.workerUp = false
				r.workerStopped = true
				r.status = RuntimeStatus{State: "degraded", ReasonCode: "worker_stopped"}
				r.mu.Unlock()
			}
		}()
		go r.runScheduler(ctx)
	})
}

func (r *generationRuntimeLifecycle) runScheduler(ctx context.Context) {
	defer r.wg.Done()
	ticker := time.NewTicker(r.options.SchedulerInterval)
	defer ticker.Stop()
	for {
		if _, err := r.options.Scheduler.Tick(ctx, r.options.BatchLimit); err != nil {
			if ctx.Err() == nil {
				r.setStatus("degraded", "scheduler_failed")
			}
		} else if ctx.Err() == nil && r.redisHealthy(ctx) {
			r.setAvailableIfWorkerRunning()
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *generationRuntimeLifecycle) runRecovery(ctx context.Context) {
	defer r.wg.Done()
	ticker := time.NewTicker(r.options.RecoveryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := r.options.Recovery.Rebuild(ctx, time.Now(), r.options.BatchLimit); err != nil {
				if ctx.Err() == nil {
					r.setStatus("degraded", "recovery_failed")
				}
				continue
			}
			if r.redisHealthy(ctx) {
				r.startWorkers(ctx)
				r.setAvailableIfWorkerRunning()
			} else {
				r.setStatus("degraded", "redis_unavailable")
			}
		}
	}
}

func (r *generationRuntimeLifecycle) setAvailableIfWorkerRunning() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.workerUp && !r.workerStopped {
		r.status = RuntimeStatus{State: "available", ReasonCode: "ready"}
	}
}

func (r *generationRuntimeLifecycle) redisHealthy(ctx context.Context) bool {
	return r.options.HealthCheck == nil || r.options.HealthCheck(ctx) == nil
}

func (r *generationRuntimeLifecycle) Wait() error {
	<-r.done
	return r.closeErr
}

func (r *generationRuntimeLifecycle) Close() error {
	r.close.Do(func() {
		r.mu.RLock()
		cancel := r.cancel
		r.mu.RUnlock()
		if cancel != nil {
			cancel()
		}
		r.wg.Wait()
		for _, closer := range r.options.Closers {
			if closer != nil {
				if err := closer.Close(); err != nil && r.closeErr == nil {
					r.closeErr = err
				}
			}
		}
		r.setStatus("unavailable", "stopped")
		close(r.done)
	})
	return r.closeErr
}
