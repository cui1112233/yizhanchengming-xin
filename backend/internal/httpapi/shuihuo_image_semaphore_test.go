package httpapi

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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
