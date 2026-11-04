package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func newYFAITestAdapter(server *httptest.Server) *YFAISeedanceAdapter {
	return &YFAISeedanceAdapter{
		BaseURL:     server.URL,
		APIKey:      "seedance-test-key",
		Model:       "seedance-2-0-official",
		Client:      server.Client(),
		ValidateURL: func(raw string) (*url.URL, error) { return url.Parse(raw) },
	}
}

func TestYFAISeedancePollPrefersIsFinalTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "running", "is_final": true, "output_url": "https://media.example/final.mp4",
		})
	}))
	defer server.Close()
	ref, err := newYFAITestAdapter(server).Poll(context.Background(), FrozenVideoModel{}, ProviderTaskRef{ProviderTaskID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.State != ProductionSucceeded || ref.MediaURL != "https://media.example/final.mp4" {
		t.Fatalf("ref=%+v", ref)
	}
}

func TestYFAISeedancePollFallsBackToCompletedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "completed", "is_final": false, "output_url": "https://media.example/done.mp4",
		})
	}))
	defer server.Close()
	ref, err := newYFAITestAdapter(server).Poll(context.Background(), FrozenVideoModel{}, ProviderTaskRef{ProviderTaskID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.State != ProductionSucceeded || ref.MediaURL != "https://media.example/done.mp4" {
		t.Fatalf("ref=%+v", ref)
	}
}

func TestYFAISeedancePollTreatsNonTerminalErrorAsFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"processing","is_final":false,"error":"content rejected"}`))
	}))
	defer server.Close()
	ref, err := newYFAITestAdapter(server).Poll(context.Background(), FrozenVideoModel{}, ProviderTaskRef{ProviderTaskID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.State != ProductionFailed || ref.ErrorMessage != "content rejected" {
		t.Fatalf("ref=%+v", ref)
	}
}

func TestYFAISeedancePollKeepsDownloadingStageRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "processing", "stage": "downloading", "is_final": false,
		})
	}))
	defer server.Close()
	ref, err := newYFAITestAdapter(server).Poll(context.Background(), FrozenVideoModel{}, ProviderTaskRef{ProviderTaskID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.State != ProductionRunning {
		t.Fatalf("ref=%+v", ref)
	}
}

func TestYFAISeedancePollDoesNotSucceedWithoutOutputURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"completed","is_final":true}`))
	}))
	defer server.Close()
	ref, err := newYFAITestAdapter(server).Poll(context.Background(), FrozenVideoModel{}, ProviderTaskRef{ProviderTaskID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if ref.State != ProductionRunning {
		t.Fatalf("ref=%+v", ref)
	}
}

func TestYFAISeedanceSubmitReadsErrorField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":400,"msg":"","error":"invalid prompt"}`))
	}))
	defer server.Close()
	_, err := newYFAITestAdapter(server).Submit(context.Background(), FrozenVideoModel{ID: "seedance-2-0-official", MaxDuration: 15}, FinalPrompt{CompiledPrompt: "test", DurationSeconds: 5})
	if err == nil || !strings.Contains(err.Error(), "invalid prompt") {
		t.Fatalf("err=%v", err)
	}
}

func TestYFAISeedanceClassifiesHTTPStatusCodes(t *testing.T) {
	cases := []struct {
		name         string
		status       int
		retryAfter   string
		wantRedirect time.Duration
		wantText     string
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, wantText: "HTTP 401"},
		{name: "payment required", status: http.StatusPaymentRequired, wantText: "HTTP 402"},
		{name: "rate limited with header", status: http.StatusTooManyRequests, retryAfter: "7", wantRedirect: 7 * time.Second, wantText: "HTTP 429"},
		{name: "rate limited without header", status: http.StatusTooManyRequests, wantRedirect: 30 * time.Second, wantText: "HTTP 429"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.retryAfter != "" {
					w.Header().Set("Retry-After", tc.retryAfter)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error":"upstream rejected"}`))
			}))
			defer server.Close()
			adapter := newYFAITestAdapter(server)
			_, err := adapter.do(context.Background(), http.MethodGet, server.URL+"/v1/tasks/t1", nil)
			var upstream *YFAIUpstreamError
			if !errors.As(err, &upstream) {
				t.Fatalf("err type=%T err=%v", err, err)
			}
			if upstream.StatusCode != tc.status || !strings.Contains(upstream.Message, tc.wantText) {
				t.Fatalf("upstream=%+v", upstream)
			}
			if tc.wantRedirect != 0 && upstream.RetryAfter != tc.wantRedirect {
				t.Fatalf("retryAfter=%v want=%v", upstream.RetryAfter, tc.wantRedirect)
			}
		})
	}
}

func TestYFAISeedanceHTTPDateRetryAfterFallsBack(t *testing.T) {
	if got := parseYFAIRetryAfter("Wed, 21 Oct 2015 07:28:00 GMT"); got != yfaiDefaultRetryAfter {
		t.Fatalf("retryAfter=%v", got)
	}
	if got := parseYFAIRetryAfter(""); got != yfaiDefaultRetryAfter {
		t.Fatalf("retryAfter=%v", got)
	}
}
