package localexecutorclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientClaimAndLeaseFlow(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer exec-token" {
			t.Errorf("missing auth: %s", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/api/local-executor/v1/jobs/claim":
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"job":{"id":"job-1"},"leaseToken":"lease-1","leaseGeneration":1}`)
		default:
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"ok":true}`)
		}
	}))
	defer server.Close()
	client, err := New(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	claim, err := client.Claim(ctx, "exec-token")
	if err != nil {
		t.Fatal(err)
	}
	if claim.Job.ID != "job-1" {
		t.Fatalf("unexpected job: %+v", claim)
	}
	lease := Lease{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
	if err := client.Progress(ctx, "exec-token", claim.Job.ID, lease, "preparing"); err != nil {
		t.Fatal(err)
	}
	if err := client.Result(ctx, "exec-token", claim.Job.ID, lease, "artifact-1"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, "|") != "POST /api/local-executor/v1/jobs/claim|POST /api/local-executor/v1/jobs/job-1/progress|POST /api/local-executor/v1/jobs/job-1/result" {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestClientNoJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	client, _ := New(server.URL, nil)
	_, err := client.Claim(context.Background(), "token")
	if err != ErrNoJob {
		t.Fatalf("expected ErrNoJob, got %v", err)
	}
}
