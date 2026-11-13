package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBatchCompletionNoticeHTTPIsAuthenticatedAndOwnerScoped(t *testing.T) {
	c, _, _ := priorityFixture(t)
	mux := http.NewServeMux()
	registerBatchAutomationRoutes(mux, BridgeAuth{Secret: "completion-test", Now: c.now}, c, "v12")
	for _, owner := range []string{"", "alice", "bob"} {
		req := httptest.NewRequest("GET", "/api/batch-factory/v12/automation/notifications", nil)
		if owner != "" {
			SignBridgeRequest(req, owner, false, c.now(), "completion-test")
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		want := 200
		if owner == "" {
			want = 401
		}
		if rec.Code != want {
			t.Fatalf("%s: %d %s", owner, rec.Code, rec.Body.String())
		}
		if owner != "" && rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("polling response cached")
		}
		if owner == "alice" && !strings.Contains(rec.Body.String(), "job-active") {
			t.Fatal("missing owned task")
		}
		if owner == "bob" && strings.Contains(rec.Body.String(), "job-active") {
			t.Fatal("foreign task leaked")
		}
	}
}

func TestBatchCompletionNoticeRequiresEveryBookSuccessfullyFinished(t *testing.T) {
	c, b, s := priorityFixture(t)
	job := s.state.Jobs["job-active"]
	job["state"] = "completed"
	job["completedAt"] = c.now()
	books := job["books"].(map[string]any)
	for _, raw := range books {
		raw.(map[string]any)["status"] = "completed"
	}
	if !batchAutomationCompletionSummary(job)["success"].(bool) {
		t.Fatal("all completed was not successful")
	}
	books[b.Books[2].ID].(map[string]any)["status"] = "queued"
	if batchAutomationCompletionSummary(job)["success"].(bool) {
		t.Fatal("queued priority book incorrectly reported done")
	}
	for _, status := range []string{"failed", "blocked", "paused", "running", "pending_confirmation"} {
		books[b.Books[2].ID].(map[string]any)["status"] = status
		if batchAutomationCompletionSummary(job)["success"].(bool) {
			t.Fatal(status)
		}
	}
	books[b.Books[2].ID].(map[string]any)["status"] = "completed"
	job["state"] = "failed"
	if batchAutomationCompletionSummary(job)["success"].(bool) {
		t.Fatal("failed job announced success")
	}
	job["state"] = "completed"
	job["books"] = map[string]any{}
	if batchAutomationCompletionSummary(job)["success"].(bool) {
		t.Fatal("empty job announced success")
	}
}

func TestBatchCompletionNoticesAreOwnerScopedAndSmall(t *testing.T) {
	c, _, s := priorityFixture(t)
	s.state.Jobs["foreign"] = map[string]any{"id": "foreign", "owner": "bob", "state": "running"}
	jobs, err := c.CompletionNotices("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0]["id"] != "job-active" {
		t.Fatal(jobs)
	}
	if _, ok := jobs[0]["books"]; ok {
		t.Fatal("full book payload leaked into polling summary")
	}
}
