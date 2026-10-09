package task9runtime

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
	"github.com/redis/go-redis/v9"
)

func TestRealRedisFlushDBRecoveryPreservesLiveExecution(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" {
		t.Skip("TASK9_REDIS_ADDR not configured")
	}
	f := newIntegrationFixture(t, 3)
	ctx := context.Background()
	now := time.Now()
	r, _, err := f.createGenerationRun(ctx, f.projectID, "real-flushdb", now.Add(-time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.store.ClaimDueRun(ctx, now); err != nil || !ok {
		t.Fatalf("claim run ok=%v err=%v", ok, err)
	}
	items, err := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
	if err != nil || len(items) != 3 {
		t.Fatalf("items=%v err=%v", items, err)
	}

	if _, err := f.db.Exec(`UPDATE book_runs SET status='queued' WHERE id=?`, items[0].BookRunID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET status='running', retryable=1, execution_token=11, execution_owner='dead', running_since=?, lease_deadline=? WHERE id=?`, now.Add(-time.Minute), now.Add(-time.Second), items[1].BookRunID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET status='running', retryable=1, execution_token=12, execution_owner='live', running_since=?, lease_deadline=? WHERE id=?`, now.Add(-time.Second), now.Add(time.Minute), items[2].BookRunID); err != nil {
		t.Fatal(err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: addr})
	defer redisClient.Close()
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatal(err)
	}

	q, err := taskruntime.NewRedisQueue(addr, "task9-real-wipe-"+fmt.Sprint(time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	recovery := NewRecovery(NewMySQLStore(f.db), NewQueueCoordinator(q))
	requeued, err := recovery.Rebuild(ctx, now, 100)
	if err != nil {
		t.Fatal(err)
	}
	if requeued != 2 {
		t.Fatalf("requeued=%d want 2 (queued + expired only)", requeued)
	}

	var liveStatus, liveOwner string
	var liveToken uint64
	var liveDeadline time.Time
	if err := f.db.QueryRow(`SELECT status,execution_owner,execution_token,lease_deadline FROM book_runs WHERE id=?`, items[2].BookRunID).Scan(&liveStatus, &liveOwner, &liveToken, &liveDeadline); err != nil {
		t.Fatal(err)
	}
	if liveStatus != "running" || liveOwner != "live" || liveToken != 12 || !liveDeadline.After(now) {
		t.Fatalf("live execution changed status=%s owner=%s token=%d deadline=%v", liveStatus, liveOwner, liveToken, liveDeadline)
	}
}

func TestMySQLStoreReloadRecoveryUsesDurableFacts(t *testing.T) {
	addr := os.Getenv("TASK9_REDIS_ADDR")
	if addr == "" {
		t.Skip("TASK9_REDIS_ADDR not configured")
	}
	f := newIntegrationFixture(t, 1)
	ctx := context.Background()
	now := time.Now()
	r, _, err := f.createGenerationRun(ctx, f.projectID, "reload-recovery", now.Add(-time.Second), 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, err := f.store.ClaimDueRun(ctx, now); err != nil || !ok {
		t.Fatalf("claim run ok=%v err=%v", ok, err)
	}
	items, err := f.store.EnsureQueuedBooks(ctx, RunClaim{RunID: r.ID})
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%v err=%v", items, err)
	}
	if _, err := f.db.Exec(`UPDATE book_runs SET status='running', retryable=1, execution_token=31, execution_owner='old-process', running_since=?, lease_deadline=? WHERE id=?`, now.Add(-time.Minute), now.Add(-time.Second), items[0].BookRunID); err != nil {
		t.Fatal(err)
	}

	// Simulate a Go process restart: all in-memory runtime objects are discarded;
	// recovery is reconstructed only from MySQL durable facts and a fresh Redis queue client.
	reloadedStore := NewMySQLStore(f.db)
	q, err := taskruntime.NewRedisQueue(addr, "task9-reload-"+fmt.Sprint(time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	recovery := NewRecovery(reloadedStore, NewQueueCoordinator(q))
	requeued, err := recovery.Rebuild(ctx, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if requeued != 1 {
		t.Fatalf("requeued=%d want 1", requeued)
	}

	var oldStatus string
	if err := f.db.QueryRow(`SELECT status FROM book_runs WHERE id=?`, items[0].BookRunID).Scan(&oldStatus); err != nil {
		t.Fatal(err)
	}
	if oldStatus != "failed" {
		t.Fatalf("expired attempt status=%s want failed", oldStatus)
	}
	var nextCount int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM book_runs WHERE run_id=? AND book_id=? AND attempt=2 AND status='queued'`, r.ID, f.bookIDs[0]).Scan(&nextCount); err != nil {
		t.Fatal(err)
	}
	if nextCount != 1 {
		t.Fatalf("attempt2 queued count=%d want 1", nextCount)
	}
}
