package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func TestSettingsCanCreateImageAndVideoExecutorTests(t *testing.T) {
	now := time.Date(2026, 10, 19, 14, 0, 0, 0, time.UTC)
	memory := localexecutor.NewMemoryStore()
	service := localexecutor.NewService(memory, func() time.Time { return now })
	pairing, err := service.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "test", OS: "windows", Version: "1.1.0"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, LocalExecutors: service, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1<<20)})

	for _, test := range []struct {
		kind     string
		duration int
	}{{"image", 0}, {"video", 5}} {
		response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/local-executor-tests", map[string]any{
			"type": test.kind, "prompt": "测试提示词", "ratio": "1:1", "duration": test.duration,
		})
		if response.Code != http.StatusAccepted {
			t.Fatalf("%s status=%d body=%s", test.kind, response.Code, response.Body.String())
		}
		var created struct {
			JobID string `json:"jobId"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
			t.Fatal(err)
		}
		claim, err := service.ClaimJob(context.Background(), paired.Token)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if err := json.Unmarshal(claim.Job.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		if created.JobID != claim.Job.ID || payload["type"] != test.kind {
			t.Fatalf("created=%+v payload=%v", created, payload)
		}
	}
}

func TestSettingsExecutorTestRejectsUnsupportedType(t *testing.T) {
	now := time.Date(2026, 10, 19, 14, 0, 0, 0, time.UTC)
	service := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, LocalExecutors: service, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1<<20)})
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/local-executor-tests", map[string]any{"type": "audio", "prompt": "x"})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
