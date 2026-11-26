package imagecapacity

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueuedImagesBoundRedisPollingAndCancelPromptly(t *testing.T) {
	var attempts atomic.Int32
	c := NewController(&testRepo{Policy{AccountLimit: 1, ModelLimit: 1, InitialLimit: 1}}, func(_ context.Context, args ...string) (any, error) {
		if args[1] == configReadScript {
			return "0", nil
		}
		if args[7] == "acquire" {
			attempts.Add(1)
		}
		return `{"granted":0}`, nil
	}, "load-test")
	ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
	defer cancel()
	started := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Acquire(ctx, Request{Endpoint: "https://test", Credential: "test", Model: "test"})
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("queue cancellation: %v", err)
			}
		}()
	}
	wg.Wait()
	if n := attempts.Load(); n != 32 {
		t.Fatalf("32 blocked images polled Redis %d times in 700ms; waiting work must back off", n)
	}
	if time.Since(started) > time.Second {
		t.Fatal("queue cancellation waited for polling timer")
	}
}
