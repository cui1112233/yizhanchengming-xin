package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/localexecutor"
)

func TestReferenceAssetImageUsesSelectedLocalExecutorModel(t *testing.T) {
	now := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	service := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	pairing, err := service.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "test", OS: "windows", Version: "1.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerReferenceAssetImageGenerateRoute(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, NewUserConfigStore(t.TempDir()), nil, service)
	req := httptest.NewRequest(http.MethodPost, "/api/reference-assets/generate", strings.NewReader(`{"asset_type":"character","asset_id":"hero","description":"红衣英雄","imageModelId":"local-doubao-executor-image","aspect_ratio":"16:9"}`))
	SignBridgeRequest(req, "alice", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var created struct {
		TaskID string `json:"taskId"`
	}
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(created.TaskID, localReferenceAssetImageTaskPrefix) {
		t.Fatalf("task id=%q", created.TaskID)
	}
	claim, err := service.ClaimJob(context.Background(), paired.Token)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(claim.Job.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["type"] != "image" || payload["model"] != "Seedream 4.5" || payload["ratio"] != "16:9" {
		t.Fatalf("payload=%v", payload)
	}
}
