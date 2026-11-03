package httpapi

import (
	"context"
	"testing"
	"time"
)

func TestComicGenerationBudgetScalesWithPages(t *testing.T) {
	small := comicGenerationBudget(10, 4)
	if small < comicGenerationMinBudget {
		t.Fatalf("small budget = %v, want >= %v", small, comicGenerationMinBudget)
	}
	large := comicGenerationBudget(35, 4)
	if large < 150*time.Minute {
		t.Fatalf("35-page budget = %v, want >= 150m so normal books are not cut off", large)
	}
	if large <= small {
		t.Fatalf("budget must grow with page count: 10 pages=%v 35 pages=%v", small, large)
	}
	huge := comicGenerationBudget(200, 1)
	if huge != 200*comicGenerationPerPageBudget+comicGenerationCushion {
		t.Fatalf("unexpected huge budget %v", huge)
	}
}

func TestComicGenerationLeaseTouchExtendsAndClampsToHardCap(t *testing.T) {
	now := time.Now()
	lease := &comicGenerationLease{
		deadline:     now.Add(time.Second),
		hardDeadline: now.Add(30 * time.Minute),
	}
	lease.touch(comicGenerationIdleExtend)
	if lease.expired(now) {
		t.Fatal("lease must not expire right after a progress heartbeat")
	}
	lease.mu.Lock()
	deadline := lease.deadline
	lease.mu.Unlock()
	if deadline.Before(now.Add(comicGenerationIdleExtend - time.Second)) {
		t.Fatalf("deadline = %v, want about now+%v", deadline, comicGenerationIdleExtend)
	}

	// 心跳不能突破硬上限：换一个有 1 分钟硬上限的租约，续期 10 分钟仍被夹住。
	clampedLease := &comicGenerationLease{
		deadline:     now.Add(time.Second),
		hardDeadline: now.Add(time.Minute),
	}
	clampedLease.touch(10 * time.Minute)
	clampedLease.mu.Lock()
	clamped := clampedLease.deadline
	clampedLease.mu.Unlock()
	if clamped.After(clampedLease.hardDeadline) {
		t.Fatalf("deadline %v escaped hard cap %v", clamped, clampedLease.hardDeadline)
	}
	if !clampedLease.expired(clampedLease.hardDeadline.Add(time.Second)) {
		t.Fatal("lease must expire after the hard cap even with heartbeats")
	}
}

func TestComicGenerationLeaseContextCancelsWhenIdle(t *testing.T) {
	previous := comicGenerationLeaseTick
	comicGenerationLeaseTick = 10 * time.Millisecond
	t.Cleanup(func() { comicGenerationLeaseTick = previous })

	ctx, lease, cancel := comicGenerationLeaseContext(context.Background(), 40*time.Millisecond, time.Second)
	defer cancel()
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("idle generation context was not cancelled within the lease")
	}
	if !lease.expired(time.Now()) {
		t.Fatal("lease should report expired when the idle deadline is passed")
	}
}

func TestPrepareAutomationBookRetryOnlyAcceptsPendingBooks(t *testing.T) {
	now := time.Now().UTC()
	entry := map[string]any{"status": "failed", "stage": "comic_images", "batchRetryPending": true, "batchRetryCount": 1}
	if !prepareAutomationBookRetry(entry, 1, now) {
		t.Fatal("pending book must be scheduled for retry")
	}
	if automationBoolValue(entry["batchRetryPending"]) {
		t.Fatal("retry schedule must clear batchRetryPending")
	}
	if automationStringValue(entry["status"], "") != "running" {
		t.Fatalf("status = %v, want running", entry["status"])
	}

	publishEntry := map[string]any{"status": "failed", "stage": "comic_publish", "comicCollectionId": "collection-1", "batchRetryPending": true}
	if !prepareAutomationBookRetry(publishEntry, 1, now) {
		t.Fatal("pending publish book must be retried")
	}
	if automationStringValue(publishEntry["stage"], "") != "comic_publish" {
		t.Fatalf("stage = %v, want comic_publish so only the collection is rebuilt", publishEntry["stage"])
	}

	if prepareAutomationBookRetry(map[string]any{"status": "failed"}, 1, now) {
		t.Fatal("book without batchRetryPending must not be scheduled")
	}
}
