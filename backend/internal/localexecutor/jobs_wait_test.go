package localexecutor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func pairedWaitTestExecutor(t *testing.T, service *Service, owner string) PairResult {
	t.Helper()
	pairing, err := service.CreatePairing(context.Background(), owner, PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), PairInput{
		Code: pairing.Code, Platform: PlatformDoubao,
		DeviceName: "等待测试执行器", OS: "windows", Version: "1.2.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	return paired
}

func TestClaimJobWaitWakesWhenJobIsCreated(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	paired := pairedWaitTestExecutor(t, service, "alice")
	result := make(chan ClaimResult, 1)
	errorsCh := make(chan error, 1)

	go func() {
		claim, err := service.ClaimJobWait(context.Background(), paired.Token, time.Second)
		if err != nil {
			errorsCh <- err
			return
		}
		result <- claim
	}()

	time.Sleep(20 * time.Millisecond)
	_, err := service.CreateJob(context.Background(), "alice", CreateJobInput{
		SourceTaskID: "wait-test", Platform: PlatformDoubao,
		Payload: map[string]any{"type": "image", "prompt": "测试"},
	})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errorsCh:
		t.Fatal(err)
	case claim := <-result:
		if claim.Job.SourceTaskID != "wait-test" {
			t.Fatalf("claim=%+v", claim)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("long wait did not wake")
	}
}

func TestClaimJobWaitTimesOutWithoutBusyPolling(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	paired := pairedWaitTestExecutor(t, service, "alice")
	started := time.Now()
	_, err := service.ClaimJobWait(context.Background(), paired.Token, 40*time.Millisecond)
	if !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("err=%v", err)
	}
	if elapsed := time.Since(started); elapsed < 30*time.Millisecond {
		t.Fatalf("returned too early: %v", elapsed)
	}
}

func TestClaimJobWaitZeroKeepsImmediateCompatibility(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	paired := pairedWaitTestExecutor(t, service, "alice")
	started := time.Now()
	_, err := service.ClaimJobWait(context.Background(), paired.Token, 0)
	if !errors.Is(err, ErrNoClaimableJob) {
		t.Fatalf("err=%v", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("compatibility claim blocked: %v", elapsed)
	}
}
