package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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
	if payload.Local.Pairing != nil || payload.Giant.Pairing != nil {
		t.Fatalf("unexpected pairing: local=%+v giant=%+v", payload.Local.Pairing, payload.Giant.Pairing)
	}
}

func TestExecutorStatusReportsPendingPairingWithoutPlaintext(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	local := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	giant := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now },
		LocalExecutors: local, GiantMaterialExecutor: giant,
	})
	localCode, err := local.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	giantCode, err := giant.CreatePairing(context.Background(), "alice", giantmaterialexecutor.PlatformGiantMaterial)
	if err != nil {
		t.Fatal(err)
	}

	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/executors/status", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if strings.Contains(body, localCode.Code) || strings.Contains(body, giantCode.Code) {
		t.Fatalf("status response must not leak plaintext pairing code: %s", body)
	}
	var payload executorStatusResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if payload.Local.State != ExecutorStateNeedsPairing || payload.Giant.State != ExecutorStateNeedsPairing {
		t.Fatalf("unexpected states: local=%q giant=%q", payload.Local.State, payload.Giant.State)
	}
	for _, pairing := range []*executorPairingStatus{payload.Local.Pairing, payload.Giant.Pairing} {
		if pairing == nil {
			t.Fatal("pairing must be present after creating a code")
		}
		if len(pairing.Fingerprint) != 8 || pairing.MaskedCode != executorPairingMaskedCode || pairing.Expired {
			t.Fatalf("unexpected pairing view: %+v", pairing)
		}
	}
	if payload.Local.Pairing.Kind != "local" || payload.Giant.Pairing.Kind != "giant" {
		t.Fatalf("unexpected kinds: local=%q giant=%q", payload.Local.Pairing.Kind, payload.Giant.Pairing.Kind)
	}

	// 设备完成配对后，待配对记录被消费，状态应回到未安装而不是继续 needs_pairing。
	if _, err := local.Pair(context.Background(), localexecutor.PairInput{
		Code: localCode.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "PC-1", OS: "windows", Version: "1.0.0",
	}); err != nil {
		t.Fatal(err)
	}
	response = signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/executors/status", nil)
	var after executorStatusResponse
	if err := json.Unmarshal(response.Body.Bytes(), &after); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if after.Local.Pairing != nil || after.Local.State != ExecutorStateOffline {
		t.Fatalf("consumed pairing must not stay pending: state=%q pairing=%+v", after.Local.State, after.Local.Pairing)
	}
}

func TestExecutorStatusPairingExpiresWithoutLeakingState(t *testing.T) {
	base := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	current := base
	local := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return current })
	giant := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return current })
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return current },
		LocalExecutors: local, GiantMaterialExecutor: giant,
	})
	if _, err := local.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao); err != nil {
		t.Fatal(err)
	}
	if _, err := giant.CreatePairing(context.Background(), "alice", giantmaterialexecutor.PlatformGiantMaterial); err != nil {
		t.Fatal(err)
	}

	current = base.Add(localexecutor.PairingTTL + time.Second)
	response := signedJSONRequest(t, api, current, "alice", http.MethodGet, "/api/shuihuo-production/executors/status", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload executorStatusResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if payload.Local.State != ExecutorStateNotInstalled || payload.Giant.State != ExecutorStateNotInstalled {
		t.Fatalf("expired pairing must not keep needs_pairing: local=%q giant=%q", payload.Local.State, payload.Giant.State)
	}
	if payload.Local.Pairing == nil || !payload.Local.Pairing.Expired {
		t.Fatalf("expired pairing must be reported as expired: %+v", payload.Local.Pairing)
	}
}
