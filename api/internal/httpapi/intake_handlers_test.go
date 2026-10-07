package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeIntakeAPI struct {
	createInput   intake.CreateIntakeInput
	createResult  intake.Intake
	createBooks   []intake.Book
	createErr     error
	executeID     int64
	executeMax    int
	executeResult intake.ExecuteResult
	executeErr    error
}

func (f *fakeIntakeAPI) CreateIntake(_ context.Context, input intake.CreateIntakeInput) (intake.Intake, []intake.Book, error) {
	f.createInput = input
	return f.createResult, f.createBooks, f.createErr
}

func (f *fakeIntakeAPI) ExecuteIntake(_ context.Context, id int64, maxText int) (intake.ExecuteResult, error) {
	f.executeID = id
	f.executeMax = maxText
	return f.executeResult, f.executeErr
}

func (f *fakeIntakeAPI) RestoreBook(_ context.Context, intakeID, bookID int64, _ int) (intake.Book, error) {
	return intake.Book{ID: bookID, IntakeID: intakeID}, f.executeErr
}

type fakeReader struct {
	intakes []intake.Intake
	books   map[int64][]intake.Book
	err     error
}

func (f *fakeReader) ListIntakes(context.Context) ([]intake.Intake, error) {
	return f.intakes, f.err
}

func (f *fakeReader) ListBooks(_ context.Context, id int64) ([]intake.Book, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.books[id], nil
}

type fakePipelineAPI struct {
	request pipeline.CreateRequest
	result  pipeline.CreateResult
	err     error
}

func (f *fakePipelineAPI) Create(_ context.Context, request pipeline.CreateRequest) (pipeline.CreateResult, error) {
	f.request = request
	return f.result, f.err
}

func TestCreateIntakeEndpoint(t *testing.T) {
	api := &fakeIntakeAPI{
		createResult: intake.Intake{ID: 11, Name: "知乎+黑岩", Status: intake.StatusPending},
		createBooks:  []intake.Book{{ID: 21, IntakeID: 11, Source: "知乎付费", PlatformID: "15", ExternalBookID: "1001", Title: "测试书", Status: intake.BookStatusPending}},
	}
	handler := NewHandler(Dependencies{Intakes: api})
	body := []byte(`{"name":"知乎+黑岩","groups":[{"source":"知乎付费","platformId":"15","books":[{"bookId":"1001","title":"测试书","gender":"女频","style":"情感"}]}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(api.createInput.Groups) != 1 || api.createInput.Groups[0].PlatformID != "15" || api.createInput.Groups[0].Books[0].BookID != "1001" {
		t.Fatalf("create input = %+v", api.createInput)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["intake"]; !ok {
		t.Fatalf("response = %v", payload)
	}
	if _, ok := payload["books"]; !ok {
		t.Fatalf("response = %v", payload)
	}
}

func TestCreateIntakeRejectsMalformedJSON(t *testing.T) {
	api := &fakeIntakeAPI{}
	handler := NewHandler(Dependencies{Intakes: api})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", bytes.NewBufferString(`{"name":`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if len(api.createInput.Groups) != 0 {
		t.Fatalf("service called for malformed JSON")
	}
}

func TestListIntakesAndBooksEndpoints(t *testing.T) {
	reader := &fakeReader{
		intakes: []intake.Intake{{ID: 11, Name: "知乎+黑岩", Status: intake.StatusCompleted}},
		books:   map[int64][]intake.Book{11: {{ID: 21, IntakeID: 11, Source: "知乎付费", PlatformID: "15", ExternalBookID: "1001", Title: "测试书", OriginalText: "不应出现在列表接口", Gender: "女频", Style: "情感", Status: intake.BookStatusFetched}}},
	}
	handler := NewHandler(Dependencies{Reader: reader})

	for _, path := range []string{"/api/v1/intakes", "/api/v1/intakes/11/books"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", path, rec.Code, rec.Body.String())
		}
		if bytes.Contains(rec.Body.Bytes(), []byte("不应出现在列表接口")) {
			t.Fatalf("%s leaked originalText in list response: %s", path, rec.Body.String())
		}
	}
}

func TestExecuteIntakeEndpointMapsPartialFailureAsValidResult(t *testing.T) {
	api := &fakeIntakeAPI{executeResult: intake.ExecuteResult{IntakeID: 11, Fetched: 2, Failed: 1, Status: intake.StatusPartial}}
	handler := NewHandler(Dependencies{Intakes: api})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes/11/execute", bytes.NewBufferString(`{"maxText":4000}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if api.executeID != 11 || api.executeMax != 4000 {
		t.Fatalf("execute = id:%d max:%d", api.executeID, api.executeMax)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"partial_failed"`)) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestExecuteIntakeValidatesPathAndMaxText(t *testing.T) {
	api := &fakeIntakeAPI{}
	handler := NewHandler(Dependencies{Intakes: api})
	cases := []struct {
		path string
		body string
	}{
		{path: "/api/v1/intakes/not-a-number/execute", body: `{"maxText":4000}`},
		{path: "/api/v1/intakes/11/execute", body: `{"maxText":10}`},
		{path: "/api/v1/intakes/11/execute", body: `{"maxText":100001}`},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewBufferString(tc.body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d body=%s", tc.path, rec.Code, rec.Body.String())
		}
	}
}

func TestCreateBatchProjectEndpointSupportsImmediateAndScheduled(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	pipelineAPI := &fakePipelineAPI{result: pipeline.CreateResult{
		Project: intake.BatchProject{ID: 51, IntakeID: 11, Name: "项目"},
		Run:     intake.Run{ID: 71, BatchProjectID: 51, RunAt: now, Status: intake.RunStatusPending},
	}}
	handler := NewHandler(Dependencies{Pipeline: pipelineAPI})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes/11/batch-projects", bytes.NewBufferString(`{"name":"项目"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("immediate status = %d body=%s", rec.Code, rec.Body.String())
	}
	if pipelineAPI.request.IntakeID != 11 || !pipelineAPI.request.RunAt.IsZero() {
		t.Fatalf("immediate request = %+v", pipelineAPI.request)
	}

	scheduledAt := "2026-10-06T08:30:00Z"
	req = httptest.NewRequest(http.MethodPost, "/api/v1/intakes/11/batch-projects", bytes.NewBufferString(`{"name":"自动化","runAt":"`+scheduledAt+`"}`))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("scheduled status = %d body=%s", rec.Code, rec.Body.String())
	}
	if pipelineAPI.request.RunAt.Format(time.RFC3339) != scheduledAt {
		t.Fatalf("scheduled runAt = %v", pipelineAPI.request.RunAt)
	}
}

func TestServiceErrorIsNotReportedAsSuccess(t *testing.T) {
	api := &fakeIntakeAPI{createErr: errors.New("database unavailable")}
	handler := NewHandler(Dependencies{Intakes: api})
	body := []byte(`{"name":"测试","groups":[{"source":"番茄付费","platformId":"2","books":[{"bookId":"1"}]}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code < 400 {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
