package giantmaterialexecutor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeleteExecutorQueuesReleasedJobWhenNoExecutorRemains(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	pairing, err := service.CreatePairing(context.Background(), "alice", PlatformGiantMaterial)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), PairInput{
		Code: pairing.Code, Platform: PlatformGiantMaterial,
		DeviceName: "win-box", OS: "windows", Version: "0.4.9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.Heartbeat(context.Background(), paired.Token, HeartbeatInput{DeviceName: "win-box", OS: "windows", Version: "0.4.9"}); err != nil {
		t.Fatal(err)
	}
	job, err := service.CreateJob(context.Background(), "alice", testJobInput("m-del", "b-del"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Claim(context.Background(), paired.Token); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteExecutor(context.Background(), "alice", paired.ExecutorID); err != nil {
		t.Fatal(err)
	}
	record, err := service.GetJob(context.Background(), "alice", job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if record.State != JobQueued || !record.WaitingForExecutor || record.LeaseExecutorID != "" || record.ErrorCode != "" {
		t.Fatalf("record state=%s lease=%s error=%s", record.State, record.LeaseExecutorID, record.ErrorCode)
	}
	if err := service.DeleteExecutor(context.Background(), "alice", paired.ExecutorID); !errors.Is(err, ErrExecutorNotFound) {
		t.Fatalf("second delete error=%v", err)
	}
}

func TestDeleteExecutorRejectsOtherOwner(t *testing.T) {
	service := NewService(NewMemoryStore(), time.Now)
	pairing, _ := service.CreatePairing(context.Background(), "alice", PlatformGiantMaterial)
	paired, _ := service.Pair(context.Background(), PairInput{
		Code: pairing.Code, Platform: PlatformGiantMaterial,
		DeviceName: "win-box", OS: "windows", Version: "0.4.9",
	})
	if err := service.DeleteExecutor(context.Background(), "bob", paired.ExecutorID); !errors.Is(err, ErrExecutorNotFound) {
		t.Fatalf("other owner error=%v", err)
	}
}
