package video

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPerformancePersonalAPIControlledFailures(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		delay      time.Duration
		clientTO   time.Duration
		wantCode   ErrorCode
		wantFailed bool
	}{
		{name: "rate_limit_429", status: http.StatusTooManyRequests, wantCode: ErrorProviderRequestFailed, wantFailed: true},
		{name: "server_500", status: http.StatusInternalServerError, wantCode: ErrorProviderRequestFailed, wantFailed: true},
		{name: "auth_failed", status: http.StatusUnauthorized, wantCode: ErrorProviderAuthFailed, wantFailed: true},
		{name: "timeout", status: http.StatusOK, delay: 80 * time.Millisecond, clientTO: 15 * time.Millisecond, wantCode: ErrorProviderUnavailable, wantFailed: true},
		{name: "slow_success", status: http.StatusOK, delay: 20 * time.Millisecond, clientTO: 500 * time.Millisecond, wantFailed: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if tc.delay > 0 {
					time.Sleep(tc.delay)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status >= 200 && tc.status < 300 {
					_, _ = fmt.Fprint(w, `{"task_id":"controlled-job"}`)
				} else {
					_, _ = fmt.Fprint(w, `{"error":"controlled"}`)
				}
			}))
			defer server.Close()

			client := server.Client()
			if tc.clientTO > 0 {
				client.Timeout = tc.clientTO
			}
			provider, err := NewPersonalAPIProvider(ProviderConfig{
				ProviderKey: ProviderPersonalAPI,
				Model:       ModelYD20Mini,
				CreateURL:   server.URL + "/create",
				TasksURL:    server.URL + "/tasks",
			}, "test-secret", client)
			if err != nil {
				t.Fatal(err)
			}

			const concurrency = 20
			start := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(concurrency)
			var failed atomic.Int64
			var wrongCode atomic.Int64
			started := time.Now()
			for i := 0; i < concurrency; i++ {
				go func() {
					defer wg.Done()
					<-start
					_, callErr := provider.Submit(context.Background(), SubmitRequest{Model: ModelYD20Mini, Prompt: "controlled failure probe"})
					if callErr != nil {
						failed.Add(1)
						var providerErr *ProviderError
						if !errors.As(callErr, &providerErr) || providerErr.Code != tc.wantCode {
							wrongCode.Add(1)
						}
					} else if tc.wantFailed {
						wrongCode.Add(1)
					}
				}()
			}
			close(start)
			wg.Wait()
			duration := time.Since(started)
			requestCount := requests.Load()
			retries := requestCount - concurrency
			if retries < 0 {
				retries = 0
			}
			t.Logf("PERF_PROVIDER_FAILURE case=%s concurrency=%d requests=%d retries=%d failed=%d wrong_code=%d duration=%s expected_code=%s",
				tc.name, concurrency, requestCount, retries, failed.Load(), wrongCode.Load(), duration, tc.wantCode)

			if wrongCode.Load() != 0 {
				t.Fatalf("wrong error classification/count=%d", wrongCode.Load())
			}
			if requestCount != concurrency {
				t.Fatalf("adapter request count=%d want=%d; unexpected internal retry/storm", requestCount, concurrency)
			}
			if tc.wantFailed && failed.Load() != concurrency {
				t.Fatalf("failed=%d want=%d", failed.Load(), concurrency)
			}
			if !tc.wantFailed && failed.Load() != 0 {
				t.Fatalf("slow success failed=%d", failed.Load())
			}
		})
	}
}

func TestPerformancePersonalAPIUnavailableNoRetryStorm(t *testing.T) {
	provider, err := NewPersonalAPIProvider(ProviderConfig{
		ProviderKey: ProviderPersonalAPI,
		Model:       ModelYD20Mini,
		CreateURL:   "http://127.0.0.1:1/create",
		TasksURL:    "http://127.0.0.1:1/tasks",
	}, "test-secret", &http.Client{Timeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, callErr := provider.Submit(context.Background(), SubmitRequest{Model: ModelYD20Mini, Prompt: "unavailable"})
	duration := time.Since(started)
	var providerErr *ProviderError
	if !errors.As(callErr, &providerErr) || providerErr.Code != ErrorProviderUnavailable {
		t.Fatalf("error=%v want provider_unavailable", callErr)
	}
	t.Logf("PERF_PROVIDER_FAILURE case=unavailable concurrency=1 requests=1 retries=0 failed=1 duration=%s expected_code=%s", duration, ErrorProviderUnavailable)
}
