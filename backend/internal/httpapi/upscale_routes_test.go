package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/upscale"
)

type upscaleFakeAPI struct {
	configured bool
	taskID     string
	uploaded   string
	task       upscale.Task
	outcome    upscale.TerminateOutcome
}

func (f *upscaleFakeAPI) Configured() bool { return f.configured }

func (f *upscaleFakeAPI) RequestUploadURL(context.Context) (upscale.UploadTarget, error) {
	return upscale.UploadTarget{FileID: "FILE-1", UploadURL: "https://upload.example.com/x", Method: http.MethodPut}, nil
}

func (f *upscaleFakeAPI) UploadMedia(_ context.Context, _ upscale.UploadTarget, body io.Reader, _ int64, _ string) error {
	raw, _ := io.ReadAll(body)
	f.uploaded = string(raw)
	return nil
}

func (f *upscaleFakeAPI) SubmitEnhance(context.Context, string, string) (string, error) {
	return f.taskID, nil
}

func (f *upscaleFakeAPI) QueryTask(context.Context, string) (upscale.Task, error) {
	return f.task, nil
}

func (f *upscaleFakeAPI) Terminate(context.Context, []string) (upscale.TerminateOutcome, error) {
	return f.outcome, nil
}

func newUpscaleTestRouter(api *upscaleFakeAPI) (http.Handler, time.Time) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	service := upscale.NewService(api, func() time.Time { return now })
	service.NewID = func() string { return "upscale-route-test" }
	mux := http.NewServeMux()
	RegisterUpscaleRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, service)
	return mux, now
}

func upscaleMultipartBody(t *testing.T, filename string, content []byte, fields map[string]string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return body, writer.FormDataContentType()
}

func TestUpscaleRoutesRequireAuth(t *testing.T) {
	router, _ := newUpscaleTestRouter(&upscaleFakeAPI{configured: true, taskID: "t1"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/ai-upscale/config", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", recorder.Code)
	}
}

func TestUpscaleConfigReportsMissingCredentials(t *testing.T) {
	router, now := newUpscaleTestRouter(&upscaleFakeAPI{configured: false})
	recorder := agentRequest(t, router, now, http.MethodGet, "/api/ai-upscale/config", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Configured  bool                `json:"configured"`
		Resolutions []map[string]string `json:"resolutions"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Configured {
		t.Fatal("configured must be false without api key")
	}
	if len(payload.Resolutions) != 2 || payload.Resolutions[0]["value"] != "1080p" || payload.Resolutions[1]["value"] != "2k" {
		t.Fatalf("resolutions=%v", payload.Resolutions)
	}
}

func TestUpscaleSubmitWithoutCredentialsReturnsExplicitError(t *testing.T) {
	router, now := newUpscaleTestRouter(&upscaleFakeAPI{configured: false})
	body, contentType := upscaleMultipartBody(t, "demo.mp4", []byte("video"), map[string]string{"resolution": "1080p"})
	req := httptest.NewRequest(http.MethodPost, "/api/ai-upscale/jobs", body)
	req.Header.Set("Content-Type", contentType)
	SignBridgeRequest(req, "alice", true, now, "secret")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("未配置火山凭证")) {
		t.Fatalf("body=%s", recorder.Body.String())
	}
}

func TestUpscaleSubmitCreatesJobAndStatusPolling(t *testing.T) {
	api := &upscaleFakeAPI{configured: true, taskID: "amk-tool-enhance-video-1"}
	router, now := newUpscaleTestRouter(api)

	body, contentType := upscaleMultipartBody(t, "demo.mp4", []byte("video-bytes"), map[string]string{"resolution": "2k"})
	req := httptest.NewRequest(http.MethodPost, "/api/ai-upscale/jobs", body)
	req.Header.Set("Content-Type", contentType)
	SignBridgeRequest(req, "alice", true, now, "secret")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var created struct {
		Job upscale.Job `json:"job"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Job.Status != upscale.StatusRunning || created.Job.TaskID != api.taskID || created.Job.Resolution != "2k" {
		t.Fatalf("job=%+v", created.Job)
	}
	if api.uploaded != "video-bytes" {
		t.Fatalf("uploaded=%q", api.uploaded)
	}

	api.task = upscale.Task{TaskID: api.taskID, Status: "completed", VideoURL: "https://cdn.example.com/out.mp4"}
	statusRecorder := agentRequest(t, router, now, http.MethodGet, "/api/ai-upscale/jobs/"+created.Job.ID, "")
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", statusRecorder.Code, statusRecorder.Body.String())
	}
	var polled struct {
		Job upscale.Job `json:"job"`
	}
	if err := json.Unmarshal(statusRecorder.Body.Bytes(), &polled); err != nil {
		t.Fatal(err)
	}
	if polled.Job.Status != upscale.StatusCompleted || polled.Job.ResultURL != "https://cdn.example.com/out.mp4" {
		t.Fatalf("job=%+v", polled.Job)
	}
}

func TestUpscaleCancelMapsMissingJob(t *testing.T) {
	router, now := newUpscaleTestRouter(&upscaleFakeAPI{configured: true, taskID: "t1"})
	recorder := agentRequest(t, router, now, http.MethodPost, "/api/ai-upscale/jobs/missing/cancel", "")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestUpscaleSubmitRejectsNonVideoFile(t *testing.T) {
	router, now := newUpscaleTestRouter(&upscaleFakeAPI{configured: true, taskID: "t1"})
	body, contentType := upscaleMultipartBody(t, "note.txt", []byte("text"), map[string]string{"resolution": "1080p"})
	req := httptest.NewRequest(http.MethodPost, "/api/ai-upscale/jobs", body)
	req.Header.Set("Content-Type", contentType)
	SignBridgeRequest(req, "alice", true, now, "secret")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestUpscaleCancelSuccess(t *testing.T) {
	api := &upscaleFakeAPI{configured: true, taskID: "t1", outcome: upscale.TerminateOutcome{Terminated: []string{"t1"}, Failed: map[string]string{}}}
	router, now := newUpscaleTestRouter(api)

	body, contentType := upscaleMultipartBody(t, "demo.mp4", []byte("video-bytes"), map[string]string{"resolution": "1080p"})
	req := httptest.NewRequest(http.MethodPost, "/api/ai-upscale/jobs", body)
	req.Header.Set("Content-Type", contentType)
	SignBridgeRequest(req, "alice", true, now, "secret")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var created struct {
		Job upscale.Job `json:"job"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	cancel := agentRequest(t, router, now, http.MethodPost, "/api/ai-upscale/jobs/"+created.Job.ID+"/cancel", "")
	if cancel.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", cancel.Code, cancel.Body.String())
	}
	var cancelled struct {
		Job upscale.Job `json:"job"`
	}
	if err := json.Unmarshal(cancel.Body.Bytes(), &cancelled); err != nil {
		t.Fatal(err)
	}
	if cancelled.Job.Status != upscale.StatusCancelled {
		t.Fatalf("job=%+v", cancelled.Job)
	}
}
