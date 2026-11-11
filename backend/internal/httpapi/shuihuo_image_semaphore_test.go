package httpapi

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// 验证任务级图片名额从“固定 4+4 切分”改为“共享 8 路 + 阶段保底”后：
// 1) 单一阶段活跃时能用满全部名额；
// 2) 两个阶段同时有排队时，彼此都不会被饿死。
func TestShuihuoGlobalImageSemaphoresUseIdleStageCapacity(t *testing.T) {
	semaphores := newShuihuoGlobalImageSemaphores(8)
	var peak atomic.Int32
	var current atomic.Int32
	stop := make(chan struct{})
	var watcher sync.WaitGroup
	watcher.Add(1)
	go func() {
		defer watcher.Done()
		for {
			select {
			case <-stop:
				return
			case <-time.After(2 * time.Millisecond):
			}
			if value := current.Load(); value > peak.Load() {
				peak.Store(value)
			}
		}
	}()

	var workers sync.WaitGroup
	start := time.Now()
	for index := 0; index < 24; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			release, err := semaphores.acquire(context.Background(), shuihuoImageStagePage)
			if err != nil {
				t.Errorf("acquire failed: %v", err)
				return
			}
			current.Add(1)
			time.Sleep(50 * time.Millisecond)
			current.Add(-1)
			release()
		}()
	}
	workers.Wait()
	close(stop)
	watcher.Wait()
	if got := peak.Load(); got != 8 {
		t.Fatalf("expected page stage to use all 8 slots when asset stage is idle, peak=%d", got)
	}
	// 24 个 50ms 请求在 8 路并发下应约 3 轮完成（远快于旧实现固定 4 路）。
	if elapsed := time.Since(start); elapsed > 400*time.Millisecond {
		t.Fatalf("expected shared capacity to finish in ~3 rounds, elapsed=%v", elapsed)
	}
}

func TestShuihuoGlobalImageSemaphoresKeepReserveForOtherStage(t *testing.T) {
	semaphores := newShuihuoGlobalImageSemaphores(8)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 先占满页图阶段。
	var pageReleases []func()
	for index := 0; index < 8; index++ {
		release, err := semaphores.acquire(ctx, shuihuoImageStagePage)
		if err != nil {
			t.Fatalf("page acquire %d failed: %v", index, err)
		}
		pageReleases = append(pageReleases, release)
	}
	// 资产图阶段开始排队：页图已满 8 路时资产必须先等到名额，不能饿死。
	assetAcquired := make(chan struct{})
	go func() {
		release, err := semaphores.acquire(ctx, shuihuoImageStageAsset)
		if err != nil {
			return
		}
		close(assetAcquired)
		release()
	}()
	select {
	case <-assetAcquired:
		t.Fatal("asset stage acquired a slot while page stage held the whole budget")
	case <-time.After(150 * time.Millisecond):
	}
	for _, release := range pageReleases {
		release()
	}
	select {
	case <-assetAcquired:
	case <-time.After(2 * time.Second):
		t.Fatal("asset stage was starved after page stage released slots")
	}
}

// 防回归：任务 A 和任务 B 虽然各自都有独立的任务级名额池，但它们发往同一
// Go 服务的图片请求必须再经过同一个进程级总闸门。若第二个任务能立即拿到
// 名额，说明“全局 8 路”仍然只是“每任务 8 路”，多个批次会叠加压垮上游。
func TestShuihuoImageSlotsShareProcessLimitAcrossAutomationTasks(t *testing.T) {
	original := shuihuoProcessImageGenerationSemaphore
	shuihuoProcessImageGenerationSemaphore = newShuihuoImageSemaphoreRaw(1)
	t.Cleanup(func() { shuihuoProcessImageGenerationSemaphore = original })

	firstCtx := withShuihuoGlobalImageConcurrency(context.Background(), 8)
	firstCtx = withShuihuoImageConcurrency(firstCtx, 4)
	firstRelease, err := acquireShuihuoGPTImageGenerationSlot(firstCtx, "gpt-image-2")
	if err != nil {
		t.Fatalf("first task acquire failed: %v", err)
	}

	secondCtx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	secondCtx = withShuihuoGlobalImageConcurrency(secondCtx, 8)
	secondCtx = withShuihuoImageConcurrency(secondCtx, 4)
	if _, err := acquireShuihuoGPTImageGenerationSlot(secondCtx, "gpt-image-2"); err == nil {
		firstRelease()
		t.Fatal("second automation task bypassed the shared process image limit")
	}

	firstRelease()
	thirdCtx, thirdCancel := context.WithTimeout(context.Background(), time.Second)
	defer thirdCancel()
	thirdCtx = withShuihuoGlobalImageConcurrency(thirdCtx, 8)
	thirdCtx = withShuihuoImageConcurrency(thirdCtx, 4)
	thirdRelease, err := acquireShuihuoGPTImageGenerationSlot(thirdCtx, "gpt-image-2")
	if err != nil {
		t.Fatalf("process slot was not released for the next task: %v", err)
	}
	thirdRelease()
}

func TestShuihuoProcessImageLimitMatchesSafeGatewayMaximum(t *testing.T) {
	if got := shuihuoProcessImageGenerationSemaphore.limit; got != shuihuoImageConcurrencyMax {
		t.Fatalf("process image limit=%d, want safe gateway maximum=%d", got, shuihuoImageConcurrencyMax)
	}
}

// 多批次同时排队时必须先到先执行。旧实现每 50ms 一起抢锁，新来的请求可能
// 连续插队，导致已经等了十几分钟的漫画页直到上下文过期也拿不到名额。
func TestShuihuoImageSemaphoreServesWaitersInFIFOOrder(t *testing.T) {
	sem := newShuihuoImageSemaphoreRaw(1)
	releaseFirst, err := sem.acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	order := make(chan int, 2)
	started := make(chan struct{}, 2)
	for _, id := range []int{1, 2} {
		id := id
		go func() {
			started <- struct{}{}
			release, acquireErr := sem.acquire(context.Background())
			if acquireErr != nil {
				return
			}
			order <- id
			release()
		}()
		<-started
		time.Sleep(20 * time.Millisecond)
	}

	releaseFirst()
	if got := <-order; got != 1 {
		t.Fatalf("first waiter=%d, want oldest waiter 1", got)
	}
	if got := <-order; got != 2 {
		t.Fatalf("second waiter=%d, want waiter 2", got)
	}
}

func TestShuihuoImageQueueBusyDoesNotRetryImmediately(t *testing.T) {
	err := shuihuoImageQueueError(context.Background(), context.DeadlineExceeded)
	if !errors.Is(err, errShuihuoImageQueueBusy) {
		t.Fatalf("queue error=%v, want queue busy marker", err)
	}
	if retryableShuihuoAssetImageTransportError(err) {
		t.Fatalf("queue busy must wait for page-level continuation, not retry immediately: %v", err)
	}
}

func TestShuihuoImageQueueBusyHeadlineIsChineseOnly(t *testing.T) {
	err := shuihuoImageQueueError(context.Background(), context.DeadlineExceeded)
	headline, raw, found := strings.Cut(err.Error(), "\n上游原文：")
	if !found {
		t.Fatalf("raw upstream must be on its own line: %q", err.Error())
	}
	want := "图片队列繁忙，等待 5 分钟后仍没有可用名额，本轮已停止，稍后会自动补齐"
	if headline != want {
		t.Fatalf("headline=%q, want %q", headline, want)
	}
	for _, english := range []string{"temporary upstream failure", "context deadline exceeded", "5m0s"} {
		if strings.Contains(headline, english) {
			t.Fatalf("headline leaks %q: %q", english, headline)
		}
	}
	if !strings.Contains(raw, "context deadline exceeded") {
		t.Fatalf("raw upstream text missing: %q", raw)
	}
}

func TestShuihuoImageQueueCanceledHeadlineIsChineseOnly(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	expiredDeadline := time.Now().Add(-time.Second)
	expiredCtx, cancelExpired := context.WithDeadline(context.Background(), expiredDeadline)
	defer cancelExpired()
	<-expiredCtx.Done()

	cases := []struct {
		name     string
		parent   context.Context
		cause    error
		headline string
		raw      string
	}{
		{"canceled", canceledCtx, context.Canceled, "本轮已取消，图片名额等待已中止", "context canceled"},
		{"deadline", expiredCtx, context.DeadlineExceeded, "本轮已超时，图片名额等待已中止", "context deadline exceeded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := shuihuoImageQueueError(tc.parent, tc.cause)
			headline, raw, found := strings.Cut(err.Error(), "\n上游原文：")
			if !found {
				t.Fatalf("raw upstream must be on its own line: %q", err.Error())
			}
			if headline != tc.headline {
				t.Fatalf("headline=%q, want %q", headline, tc.headline)
			}
			for _, english := range []string{"temporary upstream failure", "context canceled", "context deadline exceeded", "5m0s"} {
				if strings.Contains(headline, english) {
					t.Fatalf("headline leaks %q: %q", english, headline)
				}
			}
			if !strings.Contains(raw, tc.raw) {
				t.Fatalf("raw upstream text missing %q: %q", tc.raw, raw)
			}
			if !errors.Is(err, tc.cause) {
				t.Fatalf("canceled queue error must expose %v: %v", tc.cause, err)
			}
			if !errors.Is(err, batchfactoryv11.ErrUpstream) {
				t.Fatalf("canceled queue error must keep upstream classification: %v", err)
			}
			if retryableShuihuoAssetImageTransportError(err) {
				t.Fatalf("canceled queue wait must not retry immediately: %v", err)
			}
		})
	}
}

// 自动化整本出图即使有一小时页级截止时间，也不能把一次排队拖成几十分钟。
// 5 分钟拿不到名额就应释放本轮，由页级续跑机制稍后只补缺页。
func TestShuihuoImageQueueBudgetStaysBoundedByQueueTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	got := shuihuoImageQueueBudget(ctx)
	if got != shuihuoImageQueueWaitTimeout {
		t.Fatalf("automation queue budget=%s, want bounded=%s", got, shuihuoImageQueueWaitTimeout)
	}
}

func TestShuihuoImageQueueContextEmitsHeartbeatWhileWaiting(t *testing.T) {
	original := shuihuoImageQueueHeartbeatInterval
	shuihuoImageQueueHeartbeatInterval = 10 * time.Millisecond
	t.Cleanup(func() { shuihuoImageQueueHeartbeatInterval = original })

	heartbeats := make(chan struct{}, 1)
	ctx := withComicGenerationHeartbeat(context.Background(), func() {
		select {
		case heartbeats <- struct{}{}:
		default:
		}
	})
	_, cancel := shuihuoImageQueueContext(ctx)
	defer cancel()
	select {
	case <-heartbeats:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("queue wait did not refresh the automation heartbeat")
	}
}

func TestBeginShuihuoImageResponseTransferReleasesGenerationSlot(t *testing.T) {
	original := shuihuoImageResponseTransferSemaphore
	shuihuoImageResponseTransferSemaphore = newShuihuoImageSemaphoreRaw(1)
	t.Cleanup(func() { shuihuoImageResponseTransferSemaphore = original })

	var generationReleased atomic.Bool
	releaseTransfer, err := beginShuihuoImageResponseTransfer(context.Background(), func() {
		generationReleased.Store(true)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !generationReleased.Load() {
		t.Fatal("generation slot must be released after the response headers arrive")
	}
	releaseTransfer()
}

func TestShuihuoImageQueueBudgetKeepsShortFallbackWithoutDeadline(t *testing.T) {
	if got := shuihuoImageQueueBudget(context.Background()); got != shuihuoImageQueueWaitTimeout {
		t.Fatalf("queue budget=%s, want fallback=%s", got, shuihuoImageQueueWaitTimeout)
	}
}
