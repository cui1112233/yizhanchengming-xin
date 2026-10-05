package taskruntime

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestRuntimeInterfacesRemainDomainAgnostic(t *testing.T) {
	var _ Queue = (*RedisQueue)(nil)
	var _ LeaseStore = (*RedisLeaseStore)(nil)
}

func TestStaleOwnerCannotReleaseNewOwnerLease(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" {
		t.Skip("TASK9_REDIS_ADDR is required for Redis integration")
	}
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
