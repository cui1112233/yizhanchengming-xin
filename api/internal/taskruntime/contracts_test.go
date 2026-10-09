package taskruntime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func newRealRedisQueue(t *testing.T) *RedisQueue {
	t.Helper()
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" {
		t.Skip("TASK9_REDIS_ADDR is required for real Redis integration")
	}
	q, err := NewRedisQueue(addr, fmt.Sprintf("task1-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = q.Flush(context.Background()); _ = q.Close() })
	return q
}

func TestRedisQueueReclaimsExpiredProcessingReceipt(t *testing.T) {
	q := newRealRedisQueue(t)
	ctx := context.Background()
	msg := Message{TaskKey: "bookrun:41:book:9", Attempt: 1}
	if err := q.Enqueue(ctx, msg); err != nil {
		t.Fatal(err)
	}
	first, err := q.Claim(ctx, "worker-a", 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := q.ReclaimExpired(ctx, time.Now(), 10); err != nil || n != 0 {
		t.Fatalf("live reclaim=%d err=%v", n, err)
	}
	if n, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 10); err != nil || n != 1 {
		t.Fatalf("expired reclaim=%d err=%v", n, err)
	}
	if n, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 10); err != nil || n != 0 {
		t.Fatalf("duplicate reclaim=%d err=%v", n, err)
	}
	second, err := q.Claim(ctx, "worker-b", 10*time.Millisecond)
	if err != nil || second.Message != msg {
		t.Fatalf("redelivery=%+v err=%v", second, err)
	}
	if first.Receipt == second.Receipt {
		t.Fatal("redelivery reused a stale receipt")
	}
	if err := q.Ack(ctx, first); err != nil {
		t.Fatal(err)
	}
	if n, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 10); err != nil || n != 1 {
		t.Fatalf("stale ACK removed new receipt: n=%d err=%v", n, err)
	}
}

func TestRedisQueueAckAndNackRemoveProcessingMetadata(t *testing.T) {
	for _, delay := range []time.Duration{-1, 0, 50 * time.Millisecond} {
		t.Run(delay.String(), func(t *testing.T) {
			q := newRealRedisQueue(t)
			ctx := context.Background()
			if err := q.Enqueue(ctx, Message{TaskKey: "bookrun:1:book:2", Attempt: 1}); err != nil {
				t.Fatal(err)
			}
			d, err := q.Claim(ctx, "worker", time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if got := q.client.HGet(ctx, q.prefix+":queue:processing:envelope", d.Receipt).Val(); got == "" {
				t.Fatal("missing processing envelope")
			}
			if delay < 0 {
				err = q.Ack(ctx, d)
			} else {
				err = q.Nack(ctx, d, delay)
			}
			if err != nil {
				t.Fatal(err)
			}
			if q.client.HLen(ctx, q.prefix+":queue:processing:envelope").Val() != 0 || q.client.ZCard(ctx, q.prefix+":queue:processing:deadline").Val() != 0 || q.client.LLen(ctx, q.processingKey()).Val() != 0 {
				t.Fatal("processing metadata was not removed")
			}
			if n, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 10); err != nil || n != 0 {
				t.Fatalf("settled receipt reclaimed n=%d err=%v", n, err)
			}
			if delay < 0 {
				if q.client.LLen(ctx, q.readyKey()).Val() != 0 {
					t.Fatal("ACK requeued message")
				}
				return
			}
			if delay > 0 && q.client.ZCard(ctx, q.delayedKey()).Val() != 1 {
				t.Fatal("NACK lost delayed envelope")
			}
			next, err := q.Claim(ctx, "next", time.Second)
			if err != nil || next.Message != d.Message {
				t.Fatalf("NACK redelivery=%+v err=%v", next, err)
			}
		})
	}
}

func TestRedisQueueProcessingUpgradeIsBoundedAndProgresses(t *testing.T) {
	q := newRealRedisQueue(t)
	ctx := context.Background()
	// A live entry at the tail must not permanently hide old upgrade entries.
	for _, raw := range []string{`{"receipt":"old-1","message":{"task_key":"bookrun:1:book:1","attempt":1}}`, `{"receipt":"old-2","message":{"task_key":"bookrun:2:book:2","attempt":1}}`} {
		if err := q.client.LPush(ctx, q.processingKey(), raw).Err(); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.Enqueue(ctx, Message{TaskKey: "bookrun:3:book:3", Attempt: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Claim(ctx, "live", time.Second); err != nil {
		t.Fatal(err)
	}
	count := 0
	for i := 0; i < 4; i++ {
		n, err := q.ReclaimExpired(ctx, time.Now(), 1)
		if err != nil || n > 1 {
			t.Fatalf("reclaim n=%d err=%v", n, err)
		}
		count += n
	}
	if count != 2 || q.client.LLen(ctx, q.processingKey()).Val() != 1 || q.client.LLen(ctx, q.readyKey()).Val() != 2 {
		t.Fatalf("upgrade count=%d", count)
	}
}

func TestRedisQueueConcurrentReclaimDoesNotDuplicate(t *testing.T) {
	q := newRealRedisQueue(t)
	ctx := context.Background()
	if err := q.Enqueue(ctx, Message{TaskKey: "bookrun:4:book:4", Attempt: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Claim(ctx, "worker", time.Second); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	result := make(chan int, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 1)
			result <- n
			errs <- err
		}()
	}
	wg.Wait()
	close(result)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	total := 0
	for n := range result {
		total += n
	}
	if total != 1 || q.client.LLen(ctx, q.readyKey()).Val() != 1 {
		t.Fatalf("concurrent reclaim total=%d", total)
	}
}

func TestRedisQueueClaimCancellationAndMalformedEnvelope(t *testing.T) {
	q := newRealRedisQueue(t)
	ctx := context.Background()
	if err := q.client.LPush(ctx, q.readyKey(), "not-json").Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Claim(ctx, "worker", time.Second); err == nil {
		t.Fatal("malformed envelope accepted")
	}
	if q.client.LLen(ctx, q.processingKey()).Val() != 0 || q.client.LLen(ctx, q.readyKey()).Val() != 0 {
		t.Fatal("poison envelope retained")
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	time.AfterFunc(20*time.Millisecond, cancel)
	started := time.Now()
	if _, err := q.Claim(cancelCtx, "worker", 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel err=%v", err)
	}
	if time.Since(started) > 500*time.Millisecond {
		t.Fatal("empty claim ignored cancellation")
	}
}

func TestRedisQueueNackFailureDoesNotLoseProcessingReceipt(t *testing.T) {
	q := newRealRedisQueue(t)
	ctx := context.Background()
	if err := q.Enqueue(ctx, Message{TaskKey: "bookrun:5:book:5", Attempt: 1}); err != nil {
		t.Fatal(err)
	}
	d, err := q.Claim(ctx, "worker", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	// Redis scripts do not roll back prior writes on WRONGTYPE errors.
	if err := q.client.Set(ctx, q.delayedKey(), "wrong type", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := q.Nack(ctx, d, time.Second); !errors.Is(err, ErrQueueUnavailable) {
		t.Fatalf("NACK err=%v", err)
	}
	if q.client.LLen(ctx, q.processingKey()).Val() != 1 || q.client.HLen(ctx, q.prefix+":queue:processing:envelope").Val() != 1 || q.client.ZCard(ctx, q.prefix+":queue:processing:deadline").Val() != 1 {
		t.Fatal("failed NACK lost its recoverable receipt")
	}
	if err := q.client.Del(ctx, q.delayedKey()).Err(); err != nil {
		t.Fatal(err)
	}
	if n, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 10); err != nil || n != 1 {
		t.Fatalf("failed NACK recovery n=%d err=%v", n, err)
	}
}

func TestRedisQueueClaimMetadataFailureKeepsReadyEnvelope(t *testing.T) {
	q := newRealRedisQueue(t)
	ctx := context.Background()
	if err := q.Enqueue(ctx, Message{TaskKey: "bookrun:6:book:6", Attempt: 1}); err != nil {
		t.Fatal(err)
	}
	if err := q.client.Set(ctx, q.prefix+":queue:processing:envelope", "wrong type", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Claim(ctx, "worker", time.Second); !errors.Is(err, ErrQueueUnavailable) {
		t.Fatalf("claim err=%v", err)
	}
	if q.client.LLen(ctx, q.readyKey()).Val() != 1 || q.client.LLen(ctx, q.processingKey()).Val() != 0 {
		t.Fatal("failed claim changed ready/processing before metadata validation")
	}
}

func TestRuntimeInterfacesRemainDomainAgnostic(t *testing.T) {
	var _ Queue = (*RedisQueue)(nil)
	var _ LeaseStore = (*RedisLeaseStore)(nil)
}

func TestStaleOwnerCannotReleaseNewOwnerLease(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" { t.Skip("TASK9_REDIS_ADDR is required for Redis integration") }
	ctx := context.Background()
	leases, err := NewRedisLeaseStore(addr, "task9-test-stale-release-"+time.Now().Format("150405.000000000"))
	if err != nil { t.Fatal(err) }
	defer leases.Close()

	first, ok, err := leases.Claim(ctx, "book:42", "worker-a", 11, 80*time.Millisecond)
	if err != nil || !ok { t.Fatalf("first claim ok=%v err=%v", ok, err) }
	time.Sleep(120 * time.Millisecond)
	second, ok, err := leases.Claim(ctx, "book:42", "worker-b", 12, time.Second)
	if err != nil || !ok { t.Fatalf("second claim ok=%v err=%v", ok, err) }
	if second.FencingToken == first.FencingToken { t.Fatalf("fencing token did not advance: %d", second.FencingToken) }

	if released, err := leases.Release(ctx, first); !errors.Is(err, ErrLeaseNotOwner) || released {
		t.Fatalf("stale release released=%v err=%v", released, err)
	}
	if renewed, err := leases.Renew(ctx, second, time.Second); err != nil || !renewed {
		t.Fatalf("new owner lease was removed: renewed=%v err=%v", renewed, err)
	}
}

func TestStaleOwnerCannotRenewNewOwnerLease(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" { t.Skip("TASK9_REDIS_ADDR is required for Redis integration") }
	ctx := context.Background()
	leases, err := NewRedisLeaseStore(addr, "task9-test-stale-renew-"+time.Now().Format("150405.000000000"))
	if err != nil { t.Fatal(err) }
	defer leases.Close()

	oldLease, ok, err := leases.Claim(ctx, "book:7", "worker-a", 1, 80*time.Millisecond)
	if err != nil || !ok { t.Fatalf("claim old ok=%v err=%v", ok, err) }
	time.Sleep(120 * time.Millisecond)
	_, ok, err = leases.Claim(ctx, "book:7", "worker-b", 2, time.Second)
	if err != nil || !ok { t.Fatalf("claim new ok=%v err=%v", ok, err) }
	if renewed, err := leases.Renew(ctx, oldLease, time.Second); !errors.Is(err, ErrLeaseNotOwner) || renewed {
		t.Fatalf("stale renew renewed=%v err=%v", renewed, err)
	}
}

func TestLeaseRequeueExpiredUsesRealRedis(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" { t.Skip("TASK9_REDIS_ADDR is required for Redis integration") }
	ctx := context.Background()
	leases, err := NewRedisLeaseStore(addr, "task9-test-expired-"+time.Now().Format("150405.000000000"))
	if err != nil { t.Fatal(err) }
	defer leases.Close()

	lease, ok, err := leases.Claim(ctx, "book:expired", "worker-a", 41, 80*time.Millisecond)
	if err != nil || !ok { t.Fatalf("claim ok=%v err=%v", ok, err) }
	live, ok, err := leases.Claim(ctx, "book:live", "worker-b", 42, time.Second)
	if err != nil || !ok { t.Fatalf("live claim ok=%v err=%v", ok, err) }
	time.Sleep(120 * time.Millisecond)

	expired, err := leases.RequeueExpired(ctx, time.Now(), 10)
	if err != nil { t.Fatal(err) }
	if len(expired) != 1 || expired[0].TaskKey != lease.TaskKey || expired[0].FencingToken != lease.FencingToken {
		t.Fatalf("expired=%+v want task=%s token=%d", expired, lease.TaskKey, lease.FencingToken)
	}
	if renewed, err := leases.Renew(ctx, live, time.Second); err != nil || !renewed {
		t.Fatalf("live lease was disturbed: renewed=%v err=%v", renewed, err)
	}
}

func TestQueueClaimAckNackContract(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" { t.Skip("TASK9_REDIS_ADDR is required for Redis integration") }
	ctx := context.Background()
	q, err := NewRedisQueue(addr, "task9-test-queue-"+time.Now().Format("150405.000000000"))
	if err != nil { t.Fatal(err) }
	defer q.Close()

	msg := Message{TaskKey: "book:9", Attempt: 2}
	if err := q.Enqueue(ctx, msg); err != nil { t.Fatal(err) }
	delivery, err := q.Claim(ctx, "worker-a", time.Second)
	if err != nil { t.Fatal(err) }
	if delivery.Message != msg { t.Fatalf("delivery=%+v", delivery) }
	if err := q.Nack(ctx, delivery, 0); err != nil { t.Fatal(err) }
	delivery, err = q.Claim(ctx, "worker-b", time.Second)
	if err != nil { t.Fatal(err) }
	if err := q.Ack(ctx, delivery); err != nil { t.Fatal(err) }
}
