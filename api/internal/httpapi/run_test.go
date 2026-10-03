package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
)

type fakeRunCreator struct {
	input  batchfactory.CreateRunInput
	result batchfactory.RunResult
	err    error
}

func (f *fakeRunCreator) CreateAndStart(_ context.Context, input batchfactory.CreateRunInput) (batchfactory.RunResult, error) {
	f.input = input
	return f.result, f.err
}

func TestRunHandlerCreatesImmediateRunFromGroupedBooks(t *testing.T) {
	creator := &fakeRunCreator{result: batchfactory.RunResult{IntakeID:"intake-1", JobID:"job-1", GroupCount:2, BookCount:3, Status:"queued"}}
	handler := NewRunHandler(creator, func(*http.Request) (string,error) { return "user-1", nil })
	body := `{"title":"今晚批量","groups":[{"platform_id":"zhihu","platform_name":"知乎","max_txt":2000,"books":[{"book_id":"z1"},{"book_id":"z2"}]},{"platform_id":"dianzhong","platform_name":"点众","books":[{"book_id":"d1"}]}]}`
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/batch-factory/runs", strings.NewReader(body)))
	if w.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
	if creator.input.Intake.Owner != "user-1" || len(creator.input.Intake.Groups) != 2 || !creator.input.RunAt.IsZero() {
		t.Fatalf("input=%+v", creator.input)
	}
}

func TestRunHandlerUsesSameEndpointForScheduledAutomation(t *testing.T) {
	creator := &fakeRunCreator{result: batchfactory.RunResult{IntakeID:"intake-1", JobID:"job-1", Status:"queued"}}
	handler := NewRunHandler(creator, func(*http.Request) (string,error) { return "user-1", nil })
	body := `{"run_at":"2026-10-04T09:30:00Z","groups":[{"platform_id":"2","platform_name":"番茄","books":[{"book_id":"b1"}]}]}`
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/batch-factory/runs", strings.NewReader(body)))
	if w.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
	want := time.Date(2026,10,4,9,30,0,0,time.UTC)
	if !creator.input.RunAt.Equal(want) { t.Fatalf("run_at=%v", creator.input.RunAt) }
}

func TestRunHandlerRejectsClientOwner(t *testing.T) {
	creator := &fakeRunCreator{}
	handler := NewRunHandler(creator, func(*http.Request) (string,error) { return "trusted", nil })
	body := `{"owner":"spoofed","groups":[{"platform_id":"2","platform_name":"番茄","books":[{"book_id":"b1"}]}]}`
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/batch-factory/runs", strings.NewReader(body)))
	if w.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
}
