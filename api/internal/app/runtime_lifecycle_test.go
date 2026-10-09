package app

import (
	"context"
	"errors"
	"io"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type lifecycleRecovery struct {
	calls    atomic.Int32
	failures atomic.Int32
	restored atomic.Bool
}

func (r *lifecycleRecovery) Rebuild(context.Context, time.Time, int) (int, error) {
	r.calls.Add(1)
	if r.failures.Add(-1) >= 0 {
		return 0, errors.New("redis password=must-not-leak")
	}
	r.restored.Store(true)
	return 1, nil
}

type lifecycleWorker struct {
	recovery *lifecycleRecovery
	started  chan bool
	stopped  chan struct{}
}

func (w *lifecycleWorker) RunSupervised(ctx context.Context, _ taskruntime.Queue, _ time.Duration, _ func(task9runtime.RetryEvent)) error {
	w.started <- w.recovery.restored.Load()
	<-ctx.Done()
	close(w.stopped)
	return ctx.Err()
}

type lifecycleScheduler struct{ ticks atomic.Int32 }

func (s *lifecycleScheduler) Tick(context.Context, int) (int, error) {
	s.ticks.Add(1)
	return 0, nil
}

type lifecycleCloser struct {
	mu    sync.Mutex
	calls int
}

func (c *lifecycleCloser) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return nil
}

func (c *lifecycleCloser) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func TestGenerationRuntimeLifecycleRecoversBeforeReadyAndJoinsOnClose(t *testing.T) {
	recovery := &lifecycleRecovery{}
	worker := &lifecycleWorker{recovery: recovery, started: make(chan bool, 1), stopped: make(chan struct{})}
	scheduler := &lifecycleScheduler{}
	closer := &lifecycleCloser{}
	runtime := newGenerationRuntimeLifecycle(runtimeLifecycleOptions{
		Configured:        true,
		Queue:             noopRuntimeQueue{},
		Worker:            worker,
		Scheduler:         scheduler,
		Recovery:          recovery,
		Closers:           []io.Closer{closer},
		SchedulerInterval: time.Millisecond,
		RecoveryInterval:  time.Hour,
	})

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case restored := <-worker.started:
		if !restored {
			t.Fatal("worker started before durable recovery")
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	if !runtime.Ready() || runtime.Status().State != "available" {
		t.Fatalf("status=%+v want available", runtime.Status())
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-worker.stopped:
	case <-time.After(time.Second):
		t.Fatal("close did not join worker")
	}
	if closer.count() != 1 {
		t.Fatalf("closer calls=%d want 1", closer.count())
	}
	if err := runtime.Wait(); err != nil {
		t.Fatalf("wait after close: %v", err)
	}
}

func TestGenerationRuntimeLifecycleRetriesInitialRecoveryBeforeBecomingReady(t *testing.T) {
	recovery := &lifecycleRecovery{}
	recovery.failures.Store(1)
	worker := &lifecycleWorker{recovery: recovery, started: make(chan bool, 1), stopped: make(chan struct{})}
	runtime := newGenerationRuntimeLifecycle(runtimeLifecycleOptions{
		Configured:        true,
		Queue:             noopRuntimeQueue{},
		Worker:            worker,
		Scheduler:         &lifecycleScheduler{},
		Recovery:          recovery,
		SchedulerInterval: time.Hour,
		RecoveryInterval:  time.Millisecond,
	})
	t.Cleanup(func() { _ = runtime.Close() })

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if status := runtime.Status(); status.State != "degraded" || status.ReasonCode != "recovery_failed" || runtime.Ready() {
		t.Fatalf("initial status=%+v ready=%v", status, runtime.Ready())
	}
	deadline := time.Now().Add(time.Second)
	for !runtime.Ready() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !runtime.Ready() || runtime.Status().State != "available" || recovery.calls.Load() < 2 {
		t.Fatalf("recovered status=%+v calls=%d", runtime.Status(), recovery.calls.Load())
	}
}

func TestGenerationRuntimeLifecycleDoesNotBecomeReadyUntilRedisHealthRecovers(t *testing.T) {
	recovery := &lifecycleRecovery{}
	worker := &lifecycleWorker{recovery: recovery, started: make(chan bool, 1), stopped: make(chan struct{})}
	var redisUnavailable atomic.Bool
	redisUnavailable.Store(true)
	runtime := newGenerationRuntimeLifecycle(runtimeLifecycleOptions{
		Configured:      true,
		RedisConfigured: true,
		Queue:           noopRuntimeQueue{},
		Worker:          worker,
		Scheduler:       &lifecycleScheduler{},
		Recovery:        recovery,
		HealthCheck: func(context.Context) error {
			if redisUnavailable.Load() {
				return taskruntime.ErrQueueUnavailable
			}
			return nil
		},
		SchedulerInterval: time.Millisecond,
		RecoveryInterval:  time.Millisecond,
	})
	t.Cleanup(func() { _ = runtime.Close() })

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runtime.Ready() || runtime.Status().ReasonCode != "redis_unavailable" {
		t.Fatalf("status=%+v", runtime.Status())
	}
	select {
	case <-worker.started:
		t.Fatal("worker started before Redis was healthy")
	default:
	}
	redisUnavailable.Store(false)
	deadline := time.Now().Add(time.Second)
	for !runtime.Ready() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !runtime.Ready() {
		t.Fatalf("status=%+v", runtime.Status())
	}
	select {
	case restored := <-worker.started:
		if !restored {
			t.Fatal("worker started without durable recovery")
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not start after Redis recovered")
	}
}

func TestGenerationRuntimeLifecycleWithoutConfigurationStaysUnavailable(t *testing.T) {
	runtime := newGenerationRuntimeLifecycle(runtimeLifecycleOptions{})
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runtime.Ready() {
		t.Fatal("unconfigured runtime reported ready")
	}
	if status := runtime.Status(); status.State != "unavailable" || status.ReasonCode != "not_configured" {
		t.Fatalf("status=%+v", status)
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewApplicationKeepsGenerationRuntimeUnavailableWithoutConfiguration(t *testing.T) {
	for _, key := range []string{"REDIS_ADDR", "QIANTIE_REDIS_ADDR", "TASK9_REDIS_ADDR", "QIANTIE_TEXT_API_BASE_URL", "QIANTIE_TEXT_API_KEY", "QIANTIE_TEXT_MODEL"} {
		t.Setenv(key, "")
	}
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	application, err := NewApplication(db, fakeFetcher{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if application.Handler == nil || application.Runtime == nil {
		t.Fatalf("application=%+v", application)
	}
	if err := application.Runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if application.Runtime.Ready() {
		t.Fatal("unconfigured production runtime reported ready")
	}
	if status := application.Runtime.Status(); status.State != "unavailable" || status.ReasonCode != "not_configured" {
		t.Fatalf("status=%+v", status)
	}
	if err := application.Runtime.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeOwnerIncludesInstanceAndProcessIdentity(t *testing.T) {
	if owner := runtimeOwner(); !regexp.MustCompile(`^api-.+-pid[1-9][0-9]*$`).MatchString(owner) {
		t.Fatalf("owner=%q", owner)
	}
}

func TestHandlerDoesNotCreateFallbackRuntimeCoordinator(t *testing.T) {
	t.Setenv("REDIS_ADDR", "127.0.0.1:6379")
	wiring := runtimeWiring{}
	if wiring.coordinator != nil {
		t.Fatalf("coordinator=%T want nil even when Redis environment exists", wiring.coordinator)
	}
}

type stoppedRuntimeWorker struct{}

func (stoppedRuntimeWorker) RunSupervised(context.Context, taskruntime.Queue, time.Duration, func(task9runtime.RetryEvent)) error {
	return errors.New("worker stopped")
}

func TestGenerationRuntimeLifecycleDoesNotReportReadyAfterWorkerStops(t *testing.T) {
	runtime := newGenerationRuntimeLifecycle(runtimeLifecycleOptions{
		Configured:        true,
		Queue:             noopRuntimeQueue{},
		Worker:            stoppedRuntimeWorker{},
		Scheduler:         &lifecycleScheduler{},
		Recovery:          &lifecycleRecovery{},
		SchedulerInterval: time.Millisecond,
		RecoveryInterval:  time.Hour,
	})
	t.Cleanup(func() { _ = runtime.Close() })

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for runtime.Status().ReasonCode != "worker_stopped" && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if status := runtime.Status(); status.State != "degraded" || status.ReasonCode != "worker_stopped" {
		t.Fatalf("status=%+v", status)
	}
	time.Sleep(10 * time.Millisecond)
	if status := runtime.Status(); status.State != "degraded" || status.ReasonCode != "worker_stopped" || runtime.Ready() {
		t.Fatalf("scheduler masked stopped worker: status=%+v ready=%v", status, runtime.Ready())
	}
}

type noopRuntimeQueue struct{}

func (noopRuntimeQueue) Enqueue(context.Context, taskruntime.Message) error { return nil }
func (noopRuntimeQueue) Claim(context.Context, string, time.Duration) (taskruntime.Delivery, error) {
	return taskruntime.Delivery{}, taskruntime.ErrQueueEmpty
}
func (noopRuntimeQueue) Ack(context.Context, taskruntime.Delivery) error { return nil }
func (noopRuntimeQueue) Nack(context.Context, taskruntime.Delivery, time.Duration) error {
	return nil
}
