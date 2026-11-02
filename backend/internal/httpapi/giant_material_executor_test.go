package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/giantmaterialexecutor"
)

func TestGiantMaterialExecutorRoutesKeepControlPlaneSeparateAndHideVideoURL(t *testing.T) {
	now := time.Date(2026, 9, 29, 13, 0, 0, 0, time.UTC)
	service := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, GiantMaterialExecutor: service})

	pairingResponse := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/giant-material-executor/pairings", map[string]any{"platform": giantmaterialexecutor.PlatformGiantMaterial})
	if pairingResponse.Code != http.StatusCreated {
		t.Fatalf("pairing=%d body=%s", pairingResponse.Code, pairingResponse.Body.String())
	}
	pairing := decodeBody[giantmaterialexecutor.PairingSecret](t, pairingResponse)

	pairedRequest := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/giant-material-executor/v1/pair", map[string]any{"code": pairing.Code, "platform": giantmaterialexecutor.PlatformGiantMaterial, "deviceName": "win-box", "os": "windows", "version": "0.1.0"})
	if pairedRequest.Code != http.StatusOK {
		t.Fatalf("pair=%d body=%s", pairedRequest.Code, pairedRequest.Body.String())
	}
	paired := decodeBody[giantmaterialexecutor.PairResult](t, pairedRequest)
	heartbeat := bearerJSONRequest(t, api, paired.Token, http.MethodPost, "/api/giant-material-executor/v1/heartbeat", map[string]any{"deviceName": "win-box", "os": "windows", "version": "0.1.0"})
	if heartbeat.Code != http.StatusOK {
		t.Fatalf("heartbeat=%d body=%s", heartbeat.Code, heartbeat.Body.String())
	}

	jobRequest := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/giant-material-jobs", map[string]any{"platform": giantmaterialexecutor.PlatformGiantMaterial, "materialId": "7689285397448523826", "platformBookId": "book-1", "title": "滚屏测试", "videoUrl": "https://material.hnqingyuwen.top/n8_videos/sample.mp4", "durationSeconds": 1, "modelVersion": "windows-paddleocr-v1"})
	if jobRequest.Code != http.StatusCreated {
		t.Fatalf("job=%d body=%s", jobRequest.Code, jobRequest.Body.String())
	}
	job := decodeBody[map[string]giantmaterialexecutor.JobView](t, jobRequest)["job"]
	if job.ID == "" || job.State != giantmaterialexecutor.JobQueued {
		t.Fatalf("job=%+v", job)
	}
	if body := jobRequest.Body.String(); body == "" || containsAny(body, "n8_videos/sample.mp4", "leaseToken") {
		t.Fatalf("public response leaked private fields: %s", body)
	}

	claim := bearerJSONRequest(t, api, paired.Token, http.MethodPost, "/api/giant-material-executor/v1/jobs/claim", nil)
	if claim.Code != http.StatusOK {
		t.Fatalf("claim=%d body=%s", claim.Code, claim.Body.String())
	}
	claimValue := decodeBody[giantmaterialexecutor.ClaimResult](t, claim)
	if claimValue.Job.VideoURL != "https://material.hnqingyuwen.top/n8_videos/sample.mp4" || claimValue.Job.DurationSeconds != 1 {
		t.Fatalf("executor payload=%+v", claimValue.Job)
	}
	progressBody := map[string]any{"leaseToken": claimValue.LeaseToken, "leaseGeneration": claimValue.LeaseGeneration, "state": giantmaterialexecutor.JobRunning, "completed": 1, "total": 1, "percent": 100}
	if progress := bearerJSONRequest(t, api, paired.Token, http.MethodPost, "/api/giant-material-executor/v1/jobs/"+job.ID+"/progress", progressBody); progress.Code != http.StatusOK {
		t.Fatalf("progress=%d body=%s", progress.Code, progress.Body.String())
	}
	get := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/giant-material-jobs/"+job.ID, nil)
	if get.Code != http.StatusOK {
		t.Fatalf("get=%d body=%s", get.Code, get.Body.String())
	}
}

func TestGiantMaterialJobRouteQueuesWhenNoExecutorIsOnline(t *testing.T) {
	now := time.Date(2026, 10, 4, 13, 0, 0, 0, time.UTC)
	service := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, GiantMaterialExecutor: service})

	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/giant-material-jobs", map[string]any{"platform": giantmaterialexecutor.PlatformGiantMaterial, "materialId": "offline-material", "platformBookId": "offline-book", "title": "离线测试", "videoUrl": "https://material.hnqingyuwen.top/n8_videos/sample.mp4", "durationSeconds": 1, "modelVersion": "windows-paddleocr-v1"})
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	job := decodeBody[map[string]giantmaterialexecutor.JobView](t, response)["job"]
	if job.State != giantmaterialexecutor.JobQueued || job.TargetExecutorID != "" || !job.WaitingForExecutor {
		t.Fatalf("job should wait in the public queue when no executor is online: %+v", job)
	}
}

func TestGiantMaterialJobReadKeepsQueuedJobAfterExecutorGoesOffline(t *testing.T) {
	now := time.Date(2026, 10, 4, 13, 0, 0, 0, time.UTC)
	service := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, GiantMaterialExecutor: service})
	pairingResponse := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/giant-material-executor/pairings", map[string]any{"platform": giantmaterialexecutor.PlatformGiantMaterial})
	if pairingResponse.Code != http.StatusCreated {
		t.Fatalf("pairing=%d body=%s", pairingResponse.Code, pairingResponse.Body.String())
	}
	pairing := decodeBody[giantmaterialexecutor.PairingSecret](t, pairingResponse)
	pairedResponse := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/giant-material-executor/v1/pair", map[string]any{"code": pairing.Code, "platform": giantmaterialexecutor.PlatformGiantMaterial, "deviceName": "win-box", "os": "windows", "version": "0.1.0"})
	if pairedResponse.Code != http.StatusOK {
		t.Fatalf("pair=%d body=%s", pairedResponse.Code, pairedResponse.Body.String())
	}
	paired := decodeBody[giantmaterialexecutor.PairResult](t, pairedResponse)
	heartbeat := bearerJSONRequest(t, api, paired.Token, http.MethodPost, "/api/giant-material-executor/v1/heartbeat", map[string]any{"deviceName": "win-box", "os": "windows", "version": "0.1.0"})
	if heartbeat.Code != http.StatusOK {
		t.Fatalf("heartbeat=%d body=%s", heartbeat.Code, heartbeat.Body.String())
	}
	jobResponse := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/giant-material-jobs", map[string]any{"platform": giantmaterialexecutor.PlatformGiantMaterial, "materialId": "stale-material", "platformBookId": "stale-book", "title": "离线后状态读取", "videoUrl": "https://material.hnqingyuwen.top/n8_videos/sample.mp4", "durationSeconds": 1, "modelVersion": "windows-paddleocr-v1"})
	if jobResponse.Code != http.StatusCreated {
		t.Fatalf("job=%d body=%s", jobResponse.Code, jobResponse.Body.String())
	}
	job := decodeBody[map[string]giantmaterialexecutor.JobView](t, jobResponse)["job"]

	now = now.Add(giantmaterialexecutor.OnlineThreshold + time.Second)
	get := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/shuihuo-production/giant-material-jobs/"+job.ID, nil)
	if get.Code != http.StatusOK {
		t.Fatalf("get=%d body=%s", get.Code, get.Body.String())
	}
	updated := decodeBody[map[string]giantmaterialexecutor.JobView](t, get)["job"]
	if updated.State != giantmaterialexecutor.JobQueued || !updated.WaitingForExecutor || updated.ErrorCode != "" {
		t.Fatalf("queued job should wait for a returning executor: %+v", updated)
	}
}

func TestGiantMaterialExecutorClaimLongPollsAndReturns204WhenIdle(t *testing.T) {
	original := giantClaimLongPollWait
	giantClaimLongPollWait = 30 * time.Millisecond
	t.Cleanup(func() { giantClaimLongPollWait = original })

	now := time.Date(2026, 9, 29, 14, 0, 0, 0, time.UTC)
	service := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMemoryStore(), func() time.Time { return now })
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, GiantMaterialExecutor: service})

	pairingResponse := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/shuihuo-production/giant-material-executor/pairings", map[string]any{"platform": giantmaterialexecutor.PlatformGiantMaterial})
	pairing := decodeBody[giantmaterialexecutor.PairingSecret](t, pairingResponse)
	pairedRequest := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/giant-material-executor/v1/pair", map[string]any{"code": pairing.Code, "platform": giantmaterialexecutor.PlatformGiantMaterial, "deviceName": "win-box", "os": "windows", "version": "0.5.3"})
	paired := decodeBody[giantmaterialexecutor.PairResult](t, pairedRequest)

	start := time.Now()
	claim := bearerJSONRequest(t, api, paired.Token, http.MethodPost, "/api/giant-material-executor/v1/jobs/claim", nil)
	if claim.Code != http.StatusNoContent {
		t.Fatalf("claim=%d body=%s", claim.Code, claim.Body.String())
	}
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Fatalf("claim returned in %v, long poll should wait briefly", elapsed)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("long poll claim took too long: %v", elapsed)
	}
}

func bearerJSONRequest(t *testing.T, api http.Handler, token, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	return rec
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
