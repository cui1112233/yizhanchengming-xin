package localexecutor

import (
	"context"
	"testing"
	"time"
)

func TestPairingSameMachineRefreshesExistingExecutor(t *testing.T) {
	now := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	service := NewService(store, func() time.Time { return now })

	firstCode, err := service.CreatePairing(context.Background(), "alice", PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.Pair(context.Background(), PairInput{Code: firstCode.Code, Platform: PlatformDoubao, DeviceName: "PC-1", OS: "windows", Version: "1.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	secondCode, err := service.CreatePairing(context.Background(), "alice", PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Pair(context.Background(), PairInput{Code: secondCode.Code, Platform: PlatformDoubao, DeviceName: "PC-1", OS: "windows", Version: "1.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	if second.ExecutorID != first.ExecutorID {
		t.Fatalf("executor id changed: first=%q second=%q", first.ExecutorID, second.ExecutorID)
	}
	if _, err := service.List(context.Background(), "alice"); err != nil {
		t.Fatal(err)
	}
	if err := service.Heartbeat(context.Background(), first.Token, HeartbeatInput{DeviceName: "PC-1", OS: "windows", Version: "1.2.0", Accounts: AccountStats{Total: 1, Available: 1}}); err == nil {
		t.Fatal("old token remained authorized after re-pair")
	}
	if err := service.Heartbeat(context.Background(), second.Token, HeartbeatInput{DeviceName: "PC-1", OS: "windows", Version: "1.2.0", Accounts: AccountStats{Total: 1, Available: 1}}); err != nil {
		t.Fatal(err)
	}
	views, err := service.List(context.Background(), "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || !views[0].Online {
		t.Fatalf("views=%+v", views)
	}
}
