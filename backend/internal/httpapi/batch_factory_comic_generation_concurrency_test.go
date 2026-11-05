package httpapi

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestShuihuoImageGenerationConcurrencyDefaultsToTwo(t *testing.T) {
	t.Setenv(shuihuoImageConcurrencyEnv, "")
	if got := shuihuoImageGenerationConcurrency(); got != 2 {
		t.Fatalf("default image concurrency=%d, want 2", got)
	}
}

// 页级隔离：一页失败后，排队页必须继续执行，不能被 stopQueue 掐断。
func TestRunComicPageWorkContinuesQueuedPagesAfterFailure(t *testing.T) {
	work := []comicPageWorkItem{{index: 0}, {index: 1}, {index: 2}}
	startedTwo := make(chan struct{})
	releaseFirst := make(chan struct{})
	var calls atomic.Int32
	var completed atomic.Int32
	result := make(chan error, 1)
	go func() {
		result <- runComicPageWork(context.Background(), work, 2, func(ctx context.Context, _ comicPageWorkItem) error {
			call := calls.Add(1)
			if call == 1 {
				<-startedTwo
				<-releaseFirst
				if ctx.Err() != nil {
					return ctx.Err()
				}
				completed.Add(1)
				return nil
			}
			if call == 2 {
				close(startedTwo)
				return errors.New("provider failed")
			}
			return nil
		})
	}()

	select {
	case <-startedTwo:
	case <-time.After(time.Second):
		t.Fatal("two pages did not start")
	}
	deadline := time.Now().Add(time.Second)
	for calls.Load() != 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("started pages=%d, want queued page to keep running after failure", got)
	}
	close(releaseFirst)
	select {
	case err := <-result:
		var pageErr *comicPageWorkError
		if !errors.As(err, &pageErr) || len(pageErr.Failures) != 1 || pageErr.Failures[0].Err.Error() != "provider failed" {
			t.Fatalf("result error=%v, want aggregated provider failure", err)
		}
	case <-time.After(time.Second):
		t.Fatal("in-flight page did not finish")
	}
	if got := completed.Load(); got != 1 {
		t.Fatalf("completed in-flight pages=%d, want 1", got)
	}
}

func TestRunComicPageWorkContinuesAfterContentPolicyRejection(t *testing.T) {
	work := []comicPageWorkItem{{index: 0}, {index: 1}}
	var calls atomic.Int32
	err := runComicPageWork(context.Background(), work, 1, func(_ context.Context, item comicPageWorkItem) error {
		calls.Add(1)
		if item.index == 0 {
			return errShuihuoImageContentPolicy
		}
		return nil
	})
	var pageErr *comicPageWorkError
	if !errors.As(err, &pageErr) || len(pageErr.Failures) != 1 {
		t.Fatalf("content policy should be recorded as one page failure, got %v", err)
	}
	if pageErr.Failures[0].Code != automationCodeContentPolicy {
		t.Fatalf("failure code=%s, want %s", pageErr.Failures[0].Code, automationCodeContentPolicy)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("generated pages=%d, want both pages attempted", got)
	}
}
