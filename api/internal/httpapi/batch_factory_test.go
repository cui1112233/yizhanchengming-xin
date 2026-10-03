package httpapi

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeStarter struct{ input batchfactory.StartInput }
func (f *fakeStarter) Start(_ context.Context, input batchfactory.StartInput) (pipeline.Job, error) {
    f.input = input
    return pipeline.Job{ID: "job-1", BatchID: input.BatchID, RunAt: input.RunAt, Status: "queued"}, nil
}

func TestBatchFactoryCreateJobEndpoint(t *testing.T) {
    starter := &fakeStarter{}
    handler := NewBatchFactoryHandler(starter)
    body := `{"batch_id":"batch-123","needs_ai":false,"run_at":"2026-10-04T09:00:00Z"}`
    req := httptest.NewRequest(http.MethodPost, "/api/batch-factory/jobs", strings.NewReader(body))
    res := httptest.NewRecorder()

    handler.ServeHTTP(res, req)
    if res.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", res.Code, res.Body.String()) }
    if starter.input.BatchID != "batch-123" || starter.input.NeedsAI { t.Fatalf("bad input: %#v", starter.input) }
    wantRunAt := time.Date(2026,10,4,9,0,0,0,time.UTC)
    if !starter.input.RunAt.Equal(wantRunAt) { t.Fatalf("run_at=%v", starter.input.RunAt) }
    var payload map[string]any
    if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil { t.Fatal(err) }
    if payload["job_id"] != "job-1" || payload["status"] != "queued" { t.Fatalf("payload=%v", payload) }
}

func TestBatchFactoryCreateJobRejectsWrongMethod(t *testing.T) {
    handler := NewBatchFactoryHandler(&fakeStarter{})
    req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/jobs", nil)
    res := httptest.NewRecorder()
    handler.ServeHTTP(res, req)
    if res.Code != http.StatusMethodNotAllowed { t.Fatalf("status=%d", res.Code) }
}
