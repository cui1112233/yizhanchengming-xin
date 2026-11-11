package httpapi

import (
	"testing"
	"time"
)

func TestAutomationBatchRunLockDoesNotBlockDifferentBatch(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{}
	unlockFirst := coordinator.lockBatchRun("member", "batch-a")
	defer unlockFirst()

	acquired := make(chan func(), 1)
	go func() {
		acquired <- coordinator.lockBatchRun("member", "batch-b")
	}()

	select {
	case unlockSecond := <-acquired:
		unlockSecond()
	case <-time.After(250 * time.Millisecond):
		t.Fatal("an unrelated batch was blocked by another batch lifecycle transition")
	}
}

func TestAutomationBatchRunLockSerializesSameBatch(t *testing.T) {
	coordinator := &BatchAutomationCoordinator{}
	unlockFirst := coordinator.lockBatchRun("member", "batch-a")

	acquired := make(chan func(), 1)
	go func() {
		acquired <- coordinator.lockBatchRun("member", "batch-a")
	}()

	select {
	case unlockSecond := <-acquired:
		unlockSecond()
		unlockFirst()
		t.Fatal("the same batch acquired its lifecycle lock twice")
	case <-time.After(50 * time.Millisecond):
	}

	unlockFirst()
	select {
	case unlockSecond := <-acquired:
		unlockSecond()
	case <-time.After(250 * time.Millisecond):
		t.Fatal("the same batch did not continue after the lifecycle lock was released")
	}

	coordinator.batchRunMu.Lock()
	defer coordinator.batchRunMu.Unlock()
	if len(coordinator.batchRunLocks) != 0 {
		t.Fatalf("idle batch locks leaked: %d", len(coordinator.batchRunLocks))
	}
}
