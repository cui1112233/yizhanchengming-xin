package giantmaterialexecutor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPairingIsOneTimeAndExpiredCodesAreRejected(t *testing.T) {
	now := time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	clock := &testClock{now: now}
	service := NewService(NewMemoryStore(), clock.Now)

	pairing, err := service.CreatePairing(context.Background(), "alice", PlatformGiantMaterial)
	if err != nil {
		t.Fatal(err)
	}
	input := PairInput{Code: pairing.Code, Platform: PlatformGiantMaterial, DeviceName: "win-box", OS: "windows", Version: "0.1.0"}
	if _, err := service.Pair(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Pair(context.Background(), input); !errors.Is(err, ErrPairingInvalid) {
		t.Fatalf("second pairing error=%v", err)
	}

	expired, err := service.CreatePairing(context.Background(), "alice", PlatformGiantMaterial)
	if err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(PairingTTL + time.Second)
	if _, err := service.Pair(context.Background(), PairInput{Code: expired.Code, Platform: PlatformGiantMaterial, DeviceName: "win-box", OS: "windows", Version: "0.1.0"}); !errors.Is(err, ErrPairingInvalid) {
		t.Fatalf("expired pairing error=%v", err)
	}
}

func TestGiantExecutorRejectsDoubaoJobsAndClaimsOnlyItsPlatform(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	pairTestExecutor(t, service, "alice")
	if _, err := service.CreateJob(context.Background(), "alice", CreateJobInput{Platform: "doubao", MaterialID: "doubao-1", PlatformBookID: "book-1", Title: "wrong platform", VideoURL: "https://material.hnqingyuwen.top/video.mp4", DurationSeconds: 1, ModelVersion: "v1"}); !errors.Is(err, ErrInvalidPlatform) {
		t.Fatalf("doubao job error=%v", err)
	}
	if _, err := service.CreateJob(context.Background(), "alice", CreateJobInput{Platform: PlatformGiantMaterial, MaterialID: "giant-1", PlatformBookID: "book-1", Title: "giant", VideoURL: "https://material.hnqingyuwen.top/video.mp4", DurationSeconds: 1, ModelVersion: "v1"}); err != nil {
		t.Fatal(err)
	}
}

func TestQueuedJobReportsWhenNoExecutorIsOnline(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	created, err := service.CreateJob(context.Background(), "alice", testJobInput("offline-material", "offline-book"))
	if err != nil {
		t.Fatal(err)
	}
	job, err := service.GetJob(context.Background(), "alice", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if job.State != JobQueued || !job.WaitingForExecutor {
		t.Fatalf("job=%+v", job)
	}
}

func TestGetJobKeepsQueuedJobWaitingWhenItsExecutorGoesOffline(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	pairTestExecutor(t, service, "alice")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("offline-after-queue", "offline-book"))
	if err != nil {
		t.Fatal(err)
	}

	clock.now = clock.now.Add(OnlineThreshold + time.Second)
	updated, err := service.GetJob(context.Background(), "alice", job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.State != JobQueued || !updated.WaitingForExecutor || updated.ErrorCode != "" {
		t.Fatalf("queued job should wait for a returning executor: %+v", updated)
	}
}

func TestOneActiveLeaseAndStaleLeaseRejection(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 9, 29, 11, 0, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	token := pairTestExecutor(t, service, "alice")
	first, err := service.CreateJob(context.Background(), "alice", testJobInput("material-1", "book-1"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.CreateJob(context.Background(), "alice", testJobInput("material-2", "book-2"))
	if err != nil {
		t.Fatal(err)
	}
	firstLease, err := service.Claim(context.Background(), token)
	if err != nil || (firstLease.Job.ID != first.ID && firstLease.Job.ID != second.ID) {
		t.Fatalf("first claim=%+v err=%v", firstLease, err)
	}
	secondLease, err := service.Claim(context.Background(), token)
	if err != nil || secondLease.Job.ID == firstLease.Job.ID || (secondLease.Job.ID != first.ID && secondLease.Job.ID != second.ID) {
		t.Fatalf("second claim=%+v err=%v", secondLease, err)
	}
	if _, err := service.Claim(context.Background(), token); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("third claim error=%v", err)
	}
	clock.now = clock.now.Add(JobLeaseTTL + time.Second)
	if _, err := service.Renew(context.Background(), token, first.ID, LeaseCredential{Token: firstLease.LeaseToken, Generation: firstLease.LeaseGeneration}); !errors.Is(err, ErrStaleLease) {
		t.Fatalf("stale renew error=%v", err)
	}
	if recovered, err := service.Claim(context.Background(), token); err != nil || (recovered.Job.ID != first.ID && recovered.Job.ID != second.ID) || recovered.LeaseGeneration != 2 {
		t.Fatalf("recovered claim=%+v err=%v", recovered, err)
	}
}

func TestCancellationIsDurableAndNotClaimable(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	pairTestExecutor(t, service, "alice")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("material-cancel", "book-cancel"))
	if err != nil {
		t.Fatal(err)
	}
	cancelled, err := service.CancelJob(context.Background(), "alice", job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.State != JobCancelled {
		t.Fatalf("state=%s", cancelled.State)
	}
	if _, err := service.Claim(context.Background(), "unused"); !errors.Is(err, ErrExecutorUnauthorized) {
		t.Fatalf("unauthorized claim error=%v", err)
	}
}

func TestNewJobTargetsMostRecentlyOnlineExecutor(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 10, 1, 2, 0, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	oldToken := pairTestExecutor(t, service, "alice")
	if err := service.Heartbeat(context.Background(), oldToken, HeartbeatInput{DeviceName: "old-windows", OS: "windows", Version: "0.4.9"}); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(time.Second)
	newToken := pairTestExecutor(t, service, "alice")
	if err := service.Heartbeat(context.Background(), newToken, HeartbeatInput{DeviceName: "new-mac", OS: "darwin", Version: "0.5.3"}); err != nil {
		t.Fatal(err)
	}
	executors, err := service.ListExecutors(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	var macID string
	for _, executor := range executors {
		if executor.Name == "new-mac" {
			macID = executor.ID
		}
	}
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("targeted-material", "targeted-book"))
	if err != nil {
		t.Fatal(err)
	}
	if job.TargetExecutorID != macID {
		t.Fatalf("target=%q want newest online mac=%q", job.TargetExecutorID, macID)
	}
	if _, err := service.Claim(context.Background(), oldToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("old executor claim error=%v", err)
	}
	claim, err := service.Claim(context.Background(), newToken)
	if err != nil || claim.Job.ID != job.ID {
		t.Fatalf("mac claim=%+v err=%v", claim, err)
	}
}

func TestRetryCancelledJobRetargetsNewestOnlineExecutor(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 10, 1, 3, 0, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	windowsToken := pairTestExecutor(t, service, "alice")
	if err := service.Heartbeat(context.Background(), windowsToken, HeartbeatInput{DeviceName: "old-windows", OS: "windows", Version: "0.4.9"}); err != nil {
		t.Fatal(err)
	}
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("retry-material", "retry-book"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CancelJob(context.Background(), "alice", job.ID); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(time.Second)
	macToken := pairTestExecutor(t, service, "alice")
	if err := service.Heartbeat(context.Background(), macToken, HeartbeatInput{DeviceName: "new-mac", OS: "darwin", Version: "0.5.3"}); err != nil {
		t.Fatal(err)
	}

	retried, err := service.RetryJob(context.Background(), "alice", job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retried.ID != job.ID || retried.State != JobQueued || retried.TargetExecutorID == "" {
		t.Fatalf("retried=%+v", retried)
	}
	if _, err := service.Claim(context.Background(), windowsToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("old executor claim error=%v", err)
	}
	claim, err := service.Claim(context.Background(), macToken)
	if err != nil || claim.Job.ID != job.ID {
		t.Fatalf("mac claim=%+v err=%v", claim, err)
	}
}

func TestResultIsIdempotentByMaterialBookAndModel(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	token := pairTestExecutor(t, service, "alice")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("material-idempotent", "book-idempotent"))
	if err != nil {
		t.Fatal(err)
	}
	claim, err := service.Claim(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	lease := LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	for _, state := range []JobState{JobRunning, JobCleaning, JobUploading} {
		if err := service.Progress(context.Background(), token, job.ID, lease, state, ProgressInput{Completed: 1, Total: 1, Percent: 100}); err != nil {
			t.Fatal(err)
		}
	}
	completed, err := service.Complete(context.Background(), token, job.ID, lease, ResultInput{Text: "正文", WordCount: 2})
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != JobSucceeded || completed.Result.Text != "正文" {
		t.Fatalf("completed=%+v", completed)
	}
	reused, err := service.CreateJob(context.Background(), "alice", testJobInput("material-idempotent", "book-idempotent"))
	if err != nil {
		t.Fatal(err)
	}
	if reused.ID != job.ID || reused.State != JobSucceeded || reused.Result.Text != "正文" {
		t.Fatalf("reused=%+v", reused)
	}
}

func TestExecutorStaysOnlineAfterCompletion(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	token := pairTestExecutor(t, service, "alice")
	if err := service.Heartbeat(context.Background(), token, HeartbeatInput{DeviceName: "win-box", OS: "windows", Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	executors, err := service.ListExecutors(context.Background(), "alice")
	if err != nil || len(executors) != 1 || !executors[0].Online {
		t.Fatalf("executors=%+v err=%v", executors, err)
	}
}

func pairTestExecutor(t *testing.T, service *Service, owner string) string {
	t.Helper()
	pairing, err := service.CreatePairing(context.Background(), owner, PlatformGiantMaterial)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), PairInput{Code: pairing.Code, Platform: PlatformGiantMaterial, DeviceName: "win-box", OS: "windows", Version: "0.1.0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Heartbeat(context.Background(), paired.Token, HeartbeatInput{DeviceName: "win-box", OS: "windows", Version: "0.1.0"}); err != nil {
		t.Fatal(err)
	}
	return paired.Token
}

func testJobInput(materialID, bookID string) CreateJobInput {
	return CreateJobInput{Platform: PlatformGiantMaterial, MaterialID: materialID, PlatformBookID: bookID, Title: "测试书", VideoURL: "https://material.hnqingyuwen.top/video.mp4", DurationSeconds: 1, ModelVersion: "v1"}
}

type testClock struct{ now time.Time }

func (c *testClock) Now() time.Time { return c.now }
