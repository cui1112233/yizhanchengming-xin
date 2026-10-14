package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"qiantie/backend/internal/localexecutorclient"
)

func TestRetryOperationStopsAfterSuccess(t *testing.T) {
	attempts := 0
	got, err := retryOperation(context.Background(), retryConfig{Attempts: 3}, func(context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary")
		}
		return "ok", nil
	})
	if err != nil || got != "ok" || attempts != 3 {
		t.Fatalf("retry result=%q err=%v attempts=%d", got, err, attempts)
	}
}

func TestLeaseRenewerRenewsUntilStopped(t *testing.T) {
	calls := make(chan localexecutorclient.Lease, 2)
	client := &renewClient{calls: calls}
	ctx, cancel := context.WithCancel(context.Background())
	stop := startLeaseRenewer(ctx, client, "token", "job", localexecutorclient.Lease{Token: "lease", Generation: 1}, 5*time.Millisecond)
	select {
	case <-calls:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("lease was not renewed")
	}
	cancel()
	stop()
}

func TestAccountPoolDoesNotReuseBusyAccount(t *testing.T) {
	pool := newAccountPool([]string{"a", "b"})
	first, ok := pool.Acquire()
	if !ok {
		t.Fatal("first account should be available")
	}
	second, ok := pool.Acquire()
	if !ok || second == first {
		t.Fatalf("second account=%q ok=%v", second, ok)
	}
	if _, ok := pool.Acquire(); ok {
		t.Fatal("busy accounts must not be reused")
	}
	pool.Release(first)
	if got, ok := pool.Acquire(); !ok || got != first {
		t.Fatalf("released account=%q ok=%v", got, ok)
	}
}

type renewClient struct {
	calls chan localexecutorclient.Lease
}

func (c *renewClient) Renew(_ context.Context, _ string, _ string, lease localexecutorclient.Lease) error {
	c.calls <- lease
	return nil
}
