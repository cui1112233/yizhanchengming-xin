package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/localexecutor"
)

func TestExecutorStateForNormalizesFourStates(t *testing.T) {
	if got := executorStateFor(nil); got != ExecutorStateNotInstalled {
		t.Fatalf("empty state=%q", got)
	}
	if got := executorStateFor([]executorStatusDevice{{ID: "a"}}); got != ExecutorStateOffline {
		t.Fatalf("offline state=%q", got)
	}
	if got := executorStateFor([]executorStatusDevice{{ID: "a"}, {ID: "b", Online: true}}); got != ExecutorStateOnline {
		t.Fatalf("online state=%q", got)
	}
}

func TestExecutorStatusRouteAggregatesLocalAndGiant(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	local := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	giant := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now },
		LocalExecutors: local, GiantMaterialExecutor: giant,
	})

	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/executors/status", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload executorStatusResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if payload.Local.State != ExecutorStateNotInstalled || payload.Giant.State != ExecutorStateNotInstalled {
		t.Fatalf("unexpected states: local=%q giant=%q", payload.Local.State, payload.Giant.State)
	}
	if payload.Local.Devices == nil || payload.Giant.Devices == nil || payload.Devices == nil {
		t.Fatal("devices must be an empty array, not null")
	}
}
