package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
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
		hardDeadline: now.Add(2 * comicGenerationIdleExtend),
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

func TestAutomationRetryResumesFromFailedStage(t *testing.T) {
	now := time.Now().UTC()
	for _, stage := range []string{"assets", "director", "opening", "video"} {
		entry := map[string]any{"status": "failed", "stage": stage, "batchRetryPending": true, "batchRetryCount": 1}
		if !prepareAutomationBookRetry(entry, 1, now) {
			t.Fatalf("stage %s must be retryable", stage)
		}
		if got := automationStringValue(entry["resumeStage"], ""); got != stage {
			t.Fatalf("resumeStage = %q, want %q: retry must not restart from assets", got, stage)
		}
		if got := automationStringValue(entry["stage"], ""); got != stage {
			t.Fatalf("stage = %q, want %q", got, stage)
		}
	}
	// 合成阶段失败说明视频已完成，只需回到 video 重新回读，不能重建视频记录。
	entry := map[string]any{"status": "failed", "stage": "merge", "batchRetryPending": true, "batchRetryCount": 1}
	if !prepareAutomationBookRetry(entry, 1, now) {
		t.Fatal("merge failure must be retryable")
	}
	if got := automationStringValue(entry["resumeStage"], ""); got != "video" {
		t.Fatalf("resumeStage = %q, want video", got)
	}
}

func TestAutomationResumeRunsOnlyTheFailedStageAndWhatFollows(t *testing.T) {
	tests := []struct {
		resume string
		stage  string
		want   bool
	}{
		{resume: "", stage: "assets", want: true},
		{resume: "director", stage: "assets", want: false},
		{resume: "director", stage: "director", want: true},
		{resume: "video", stage: "assets", want: false},
		{resume: "video", stage: "director", want: false},
		{resume: "video", stage: "opening", want: false},
		{resume: "video", stage: "video", want: true},
	}
	for _, tc := range tests {
		if got := automationResumeShouldRunStage(tc.resume, tc.stage); got != tc.want {
			t.Fatalf("resume=%q stage=%q got=%t want=%t", tc.resume, tc.stage, got, tc.want)
		}
	}
}

func TestProductionInFlightDoesNotFailTheBookOrConsumeRetryBudget(t *testing.T) {
	now := time.Now().UTC()
	entry := map[string]any{"status": "failed", "stage": "video", "batchRetryCount": 1, "error": "视频生成超时或已停止"}
	job := batchfactoryv11.ProductionJob{
		BookID:    "book_1",
		RequestID: "request_1",
		Status:    batchfactoryv11.ProductionRunning,
		Tasks:     []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionRunning}},
	}
	fetch := func(context.Context) (batchfactoryv11.BatchStatus, error) {
		return batchfactoryv11.BatchStatus{Jobs: []batchfactoryv11.ProductionJob{job}}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	waited, err := waitProductionStatus(ctx, fetch, "book_1", "request_1", 20*time.Millisecond)
	if !errors.Is(err, errProductionInFlight) {
		t.Fatalf("err = %v, want errProductionInFlight: in-flight tasks must not be reported as failure", err)
	}
	if len(waited.Tasks) != len(job.Tasks) {
		t.Fatalf("waited job lost tasks: %+v", waited)
	}
	if scheduleAutomationBatchEndRetry(entry, "video", err, "视频仍在生成", 1, now) {
		t.Fatal("in-flight production must not schedule a whole-book retry")
	}
	if automationNumberValue(entry["batchRetryCount"], 0) != 1 {
		t.Fatalf("batchRetryCount = %v, must stay unchanged", entry["batchRetryCount"])
	}
	if automationBoolValue(entry["batchRetryPending"]) {
		t.Fatal("in-flight production must not mark the book as retry-pending")
	}
}

func TestProductionWaitFinalReadAcceptsAJobThatJustCompleted(t *testing.T) {
	inFlight := batchfactoryv11.ProductionJob{
		BookID: "book_1", RequestID: "request_1", Status: batchfactoryv11.ProductionRunning,
		Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionRunning}},
	}
	done := batchfactoryv11.ProductionJob{
		BookID: "book_1", RequestID: "request_1", Status: batchfactoryv11.ProductionSucceeded,
		Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded, MediaURL: "https://cdn/v.mp4"}},
	}
	calls := 0
	fetch := func(context.Context) (batchfactoryv11.BatchStatus, error) {
		calls++
		if calls > 1 {
			// 终态回读时看到任务已经完成。
			return batchfactoryv11.BatchStatus{Jobs: []batchfactoryv11.ProductionJob{done}}, nil
		}
		return batchfactoryv11.BatchStatus{Jobs: []batchfactoryv11.ProductionJob{inFlight}}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	job, err := waitProductionStatus(ctx, fetch, "book_1", "request_1", 20*time.Millisecond)
	if err != nil {
		t.Fatalf("err = %v, want nil: the last task finished before the deadline check", err)
	}
	if len(job.Tasks) != 1 || job.Tasks[0].MediaURL == "" {
		t.Fatalf("job = %+v, want the completed production", job)
	}
}
