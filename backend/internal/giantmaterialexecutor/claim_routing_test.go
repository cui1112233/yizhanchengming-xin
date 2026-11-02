package giantmaterialexecutor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func pairExecutorVersion(t *testing.T, service *Service, owner, osName, device, version string) string {
	t.Helper()
	pairing, err := service.CreatePairing(context.Background(), owner, PlatformGiantMaterial)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), PairInput{
		Code: pairing.Code, Platform: PlatformGiantMaterial,
		DeviceName: device, OS: osName, Version: version,
	})
	if err != nil {
		t.Fatal(err)
	}
	return paired.Token
}

func heartbeatExecutorVersion(t *testing.T, service *Service, token, osName, device, version string) {
	t.Helper()
	err := service.Heartbeat(context.Background(), token, HeartbeatInput{
		DeviceName: device, OS: osName, Version: version,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPreferredDarwinBothOnlineOnlyDarwinClaims(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	winToken := pairExecutorVersion(t, service, "alice", "windows", "win-box", "0.4.9")
	macToken := pairExecutorVersion(t, service, "alice", "darwin", "mac-box", "0.5.0")
	if _, err := service.SavePreference(context.Background(), "alice", "darwin"); err != nil {
		t.Fatal(err)
	}
	heartbeatExecutorVersion(t, service, winToken, "windows", "win-box", "0.4.9")
	heartbeatExecutorVersion(t, service, macToken, "darwin", "mac-box", "0.5.0")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("m1", "b1"))
	if err != nil {
		t.Fatal(err)
	}
	macClaim, err := service.Claim(context.Background(), macToken)
	if err != nil || macClaim.Job.ID != job.ID {
		t.Fatalf("mac claim=%+v err=%v", macClaim, err)
	}
	if _, err := service.Claim(context.Background(), winToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("windows claim error=%v", err)
	}
}

func TestPreferredDarwinOfflineWindowsClaims(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 10, 1, 10, 10, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	winToken := pairExecutorVersion(t, service, "alice", "windows", "win-box", "0.4.9")
	pairExecutorVersion(t, service, "alice", "darwin", "mac-box", "0.5.0")
	if _, err := service.SavePreference(context.Background(), "alice", "darwin"); err != nil {
		t.Fatal(err)
	}
	heartbeatExecutorVersion(t, service, winToken, "windows", "win-box", "0.4.9")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("m2", "b2"))
	if err != nil {
		t.Fatal(err)
	}
	winClaim, err := service.Claim(context.Background(), winToken)
	if err != nil || winClaim.Job.ID != job.ID {
		t.Fatalf("windows claim=%+v err=%v", winClaim, err)
	}
}

func TestFailureCooldownFallbackAndTrialRestore(t *testing.T) {
	clock := &testClock{now: time.Date(2026, 10, 1, 10, 20, 0, 0, time.UTC)}
	service := NewService(NewMemoryStore(), clock.Now)
	winToken := pairExecutorVersion(t, service, "alice", "windows", "win-box", "0.4.9")
	macToken := pairExecutorVersion(t, service, "alice", "darwin", "mac-box", "0.5.0")
	if _, err := service.SavePreference(context.Background(), "alice", "darwin"); err != nil {
		t.Fatal(err)
	}
	heartbeatExecutorVersion(t, service, winToken, "windows", "win-box", "0.4.9")
	heartbeatExecutorVersion(t, service, macToken, "darwin", "mac-box", "0.5.0")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("m3", "b3"))
	if err != nil {
		t.Fatal(err)
	}
	macClaim, err := service.Claim(context.Background(), macToken)
	if err != nil {
		t.Fatal(err)
	}
	macLease := LeaseCredential{Token: macClaim.LeaseToken, Generation: macClaim.LeaseGeneration}
	if err := service.Progress(context.Background(), macToken, job.ID, macLease, JobRunning,
		ProgressInput{Completed: 1, Total: 10, Percent: 10}); err != nil {
		t.Fatal(err)
	}
	if err := service.Fail(context.Background(), macToken, job.ID, macLease,
		FailureInput{Code: "ocr_crash", Message: "worker died"}); err != nil {
		t.Fatal(err)
	}
	// 同 key 重新创建 = 失败任务重新排队
	if _, err := service.CreateJob(context.Background(), "alice", testJobInput("m3", "b3")); err != nil {
		t.Fatal(err)
	}
	winClaim, err := service.Claim(context.Background(), winToken)
	if err != nil || winClaim.Job.ID != job.ID || winClaim.LeaseGeneration != 3 {
		t.Fatalf("windows fallback=%+v err=%v", winClaim, err)
	}
	if _, err := service.Claim(context.Background(), macToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("mac in cooldown error=%v", err)
	}
	// Windows 也干砸
	winLease := LeaseCredential{Token: winClaim.LeaseToken, Generation: winClaim.LeaseGeneration}
	if err := service.Progress(context.Background(), winToken, job.ID, winLease, JobRunning,
		ProgressInput{Completed: 2, Total: 10, Percent: 20}); err != nil {
		t.Fatal(err)
	}
	if err := service.Fail(context.Background(), winToken, job.ID, winLease,
		FailureInput{Code: "ocr_crash", Message: "also died"}); err != nil {
		t.Fatal(err)
	}
	queued, err := service.CreateJob(context.Background(), "alice", testJobInput("m3", "b3"))
	if err != nil || queued.State != JobQueued || !queued.WaitingForExecutor {
		t.Fatalf("requeue while all executors are cooling down=%+v error=%v", queued, err)
	}
	// 两台都在冷却：都领不到
	if _, err := service.Claim(context.Background(), macToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("both cooling mac error=%v", err)
	}
	if _, err := service.Claim(context.Background(), winToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("both cooling windows error=%v", err)
	}
	clock.now = clock.now.Add(FailureCooldown + time.Second)
	heartbeatExecutorVersion(t, service, winToken, "windows", "win-box", "0.4.9")
	heartbeatExecutorVersion(t, service, macToken, "darwin", "mac-box", "0.5.0")
	if _, err := service.CreateJob(context.Background(), "alice", testJobInput("m3", "b3")); err != nil {
		t.Fatal(err)
	}
	trial, err := service.Claim(context.Background(), macToken)
	if err != nil || trial.Job.ID != job.ID || trial.LeaseGeneration != 5 {
		t.Fatalf("trial restore=%+v err=%v", trial, err)
	}
}

func TestStuckLeaseReclaimedWithoutWaitingForExpiry(t *testing.T) {
	start := time.Date(2026, 10, 1, 11, 0, 0, 0, time.UTC)
	clock := &testClock{now: start}
	service := NewService(NewMemoryStore(), clock.Now)
	winToken := pairExecutorVersion(t, service, "alice", "windows", "win-box", "0.4.9")
	heartbeatExecutorVersion(t, service, winToken, "windows", "win-box", "0.4.9")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("m4", "b4"))
	if err != nil {
		t.Fatal(err)
	}
	claim, err := service.Claim(context.Background(), winToken)
	if err != nil {
		t.Fatal(err)
	}
	lease := LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	for clock.now.Before(start.Add(10*time.Minute + 30*time.Second)) {
		clock.now = clock.now.Add(30 * time.Second)
		if _, err := service.Renew(context.Background(), winToken, job.ID, lease); err != nil {
			t.Fatal(err)
		}
	}
	reclaim, err := service.Claim(context.Background(), winToken)
	if err != nil {
		t.Fatal(err)
	}
	if reclaim.Job.ID != job.ID || reclaim.LeaseGeneration != 2 || !reclaim.LeaseExpiresAt.After(clock.now) {
		t.Fatalf("reclaim=%+v", reclaim)
	}
}

func TestFreshProgressIsNotReclaimedAsStuck(t *testing.T) {
	start := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	clock := &testClock{now: start}
	service := NewService(NewMemoryStore(), clock.Now)
	winToken := pairExecutorVersion(t, service, "alice", "windows", "win-box", "0.4.9")
	heartbeatExecutorVersion(t, service, winToken, "windows", "win-box", "0.4.9")
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("m5", "b5"))
	if err != nil {
		t.Fatal(err)
	}
	claim, err := service.Claim(context.Background(), winToken)
	if err != nil {
		t.Fatal(err)
	}
	lease := LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	for clock.now.Before(start.Add(8 * time.Minute)) {
		clock.now = clock.now.Add(30 * time.Second)
		if _, err := service.Renew(context.Background(), winToken, job.ID, lease); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.Progress(context.Background(), winToken, job.ID, lease, JobRunning,
		ProgressInput{Completed: 100, Total: 281, Percent: 35}); err != nil {
		t.Fatal(err)
	}
	for clock.now.Before(start.Add(10*time.Minute + 30*time.Second)) {
		clock.now = clock.now.Add(30 * time.Second)
		if _, err := service.Renew(context.Background(), winToken, job.ID, lease); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.Claim(context.Background(), winToken); !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("fresh progress reclaimed error=%v", err)
	}
}
