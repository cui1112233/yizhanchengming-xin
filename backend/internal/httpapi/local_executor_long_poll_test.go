package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/localexecutor"
)

func longPollTestAPI(t *testing.T) (http.Handler, *localexecutor.Service, localexecutor.PairResult) {
	t.Helper()
	service := localexecutor.NewService(localexecutor.NewMemoryStore(), time.Now)
	pairing, err := service.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), localexecutor.PairInput{
		Code: pairing.Code, Platform: localexecutor.PlatformDoubao,
		DeviceName: "长等待测试", OS: "windows", Version: "1.2.0",
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewRouter(RouterOptions{BridgeSecret: "secret", LocalExecutors: service}), service, paired
}

func executorClaimRequest(token, query string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/local-executor/v1/jobs/claim"+query, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func TestLocalExecutorClaimWithoutWaitKeepsImmediate204(t *testing.T) {
	api, _, paired := longPollTestAPI(t)
	recorder := httptest.NewRecorder()
	started := time.Now()
	api.ServeHTTP(recorder, executorClaimRequest(paired.Token, ""))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("legacy claim blocked: %v", elapsed)
	}
}

func TestLocalExecutorLongPollWakesForNewJob(t *testing.T) {
	api, service, paired := longPollTestAPI(t)
	recorder := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		api.ServeHTTP(recorder, executorClaimRequest(paired.Token, "?waitSeconds=1"))
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	_, err := service.CreateJob(context.Background(), "alice", localexecutor.CreateJobInput{
		SourceTaskID: "http-long-poll", Platform: localexecutor.PlatformDoubao,
		Payload: map[string]any{"type": "image", "prompt": "测试"},
	})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("long-poll request did not wake")
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var claim localexecutor.ClaimResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &claim); err != nil {
		t.Fatal(err)
	}
	if claim.Job.SourceTaskID != "http-long-poll" {
		t.Fatalf("claim=%+v", claim)
	}
}

func TestLocalExecutorLongPollRejectsInvalidWait(t *testing.T) {
	api, _, paired := longPollTestAPI(t)
	for _, query := range []string{"?waitSeconds=0", "?waitSeconds=31", "?waitSeconds=abc"} {
		recorder := httptest.NewRecorder()
		api.ServeHTTP(recorder, executorClaimRequest(paired.Token, query))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("query=%s status=%d body=%s", query, recorder.Code, recorder.Body.String())
		}
	}
}
