package upscale

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

type fakeAPI struct {
	configured       bool
	target           UploadTarget
	uploaded         string
	requestErr       error
	uploadErr        error
	submitErr        error
	taskID           string
	submitVideoURL   string
	submitResolution string
	task             Task
	taskErr          error
	terminateOutcome TerminateOutcome
	terminateErr     error
	terminateCalls   [][]string
}

func (f *fakeAPI) Configured() bool { return f.configured }

func (f *fakeAPI) RequestUploadURL(context.Context) (UploadTarget, error) {
	if f.requestErr != nil {
		return UploadTarget{}, f.requestErr
	}
	return f.target, nil
}

func (f *fakeAPI) UploadMedia(_ context.Context, _ UploadTarget, body io.Reader, _ int64, _ string) error {
	if f.uploadErr != nil {
		return f.uploadErr
	}
	raw, _ := io.ReadAll(body)
	f.uploaded = string(raw)
	return nil
}

func (f *fakeAPI) SubmitEnhance(_ context.Context, videoURL, resolution string) (string, error) {
	if f.submitErr != nil {
		return "", f.submitErr
	}
	f.submitVideoURL = videoURL
	f.submitResolution = resolution
	return f.taskID, nil
}

func (f *fakeAPI) QueryTask(context.Context, string) (Task, error) {
	if f.taskErr != nil {
		return Task{}, f.taskErr
	}
	return f.task, nil
}

func (f *fakeAPI) Terminate(_ context.Context, taskIDs []string) (TerminateOutcome, error) {
	f.terminateCalls = append(f.terminateCalls, taskIDs)
	if f.terminateErr != nil {
		return TerminateOutcome{}, f.terminateErr
	}
	return f.terminateOutcome, nil
}

func newTestService(api API) (*Service, time.Time) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	service := NewService(api, func() time.Time { return now })
	service.NewID = func() string { return "upscale-test" }
	return service, now
}

func configuredFake() *fakeAPI {
	return &fakeAPI{
		configured: true,
		target:     UploadTarget{FileID: "FILE-1", UploadURL: "https://upload.example.com/x", Method: "PUT"},
		taskID:     "amk-tool-enhance-video-1",
	}
}

func TestServiceSubmitUploadsThenSubmits(t *testing.T) {
	api := configuredFake()
	service, _ := newTestService(api)

	job, err := service.Submit(context.Background(), "alice", strings.NewReader("video-bytes"), 11, "video/mp4", "demo.mp4", "1080p")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if api.uploaded != "video-bytes" {
		t.Fatalf("uploaded=%q", api.uploaded)
	}
	if api.submitVideoURL != "mediakit://FILE-1" || api.submitResolution != "1080p" {
		t.Fatalf("videoURL=%q resolution=%q", api.submitVideoURL, api.submitResolution)
	}
	if job.ID != "upscale-test" || job.Status != StatusRunning || job.TaskID != api.taskID {
		t.Fatalf("job=%+v", job)
	}
	if job.Progress != nil {
		t.Fatalf("running progress must be unknown, got %v", *job.Progress)
	}
	if job.FileName != "demo.mp4" || job.Resolution != "1080p" {
		t.Fatalf("job=%+v", job)
	}
}

func TestServiceSubmitRejectsMissingConfiguration(t *testing.T) {
	api := configuredFake()
	api.configured = false
	service, _ := newTestService(api)

	_, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p")
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err=%v", err)
	}
	if api.uploaded != "" {
		t.Fatal("must not upload without configuration")
	}
}

func TestServiceSubmitRejectsUnsupportedResolution(t *testing.T) {
	service, _ := newTestService(configuredFake())

	_, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "4k")
	if !errors.Is(err, ErrInvalidResolution) {
		t.Fatalf("err=%v", err)
	}
}

func TestServiceSubmitMarksJobFailedWhenUploadFails(t *testing.T) {
	api := configuredFake()
	api.uploadErr = &APIError{Status: 403, Message: "signature expired"}
	service, _ := newTestService(api)

	job, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "2k")
	if err == nil {
		t.Fatal("expected error")
	}
	if job.Status != StatusFailed || !strings.Contains(job.Error, "signature expired") {
		t.Fatalf("job=%+v", job)
	}
}

func TestServiceJobMapsCompletedTask(t *testing.T) {
	api := configuredFake()
	api.task = Task{TaskID: api.taskID, Status: "completed", VideoURL: "https://cdn.example.com/out.mp4"}
	service, _ := newTestService(api)
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}

	job, err := service.Job(context.Background(), "alice", "upscale-test")
	if err != nil {
		t.Fatalf("job: %v", err)
	}
	if job.Status != StatusCompleted || job.Progress == nil || *job.Progress != 100 {
		t.Fatalf("job=%+v", job)
	}
	if job.ResultURL != "https://cdn.example.com/out.mp4" {
		t.Fatalf("result=%q", job.ResultURL)
	}
}

func TestServiceJobMapsFailedTask(t *testing.T) {
	api := configuredFake()
	api.task = Task{TaskID: api.taskID, Status: "failed", ErrorCode: "DownloadFailed", ErrorMessage: "cannot download source"}
	service, _ := newTestService(api)
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}

	job, err := service.Job(context.Background(), "alice", "upscale-test")
	if err != nil {
		t.Fatalf("job: %v", err)
	}
	if job.Status != StatusFailed || !strings.Contains(job.Error, "cannot download source") || !strings.Contains(job.Error, "DownloadFailed") {
		t.Fatalf("job=%+v", job)
	}
}

func TestServiceJobKeepsRunningWhenUpstreamUnknown(t *testing.T) {
	api := configuredFake()
	api.task = Task{TaskID: api.taskID, Status: "running"}
	service, _ := newTestService(api)
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}

	job, err := service.Job(context.Background(), "alice", "upscale-test")
	if err != nil {
		t.Fatalf("job: %v", err)
	}
	if job.Status != StatusRunning || job.Progress != nil {
		t.Fatalf("job=%+v", job)
	}
}

func TestServiceJobIsOwnerScoped(t *testing.T) {
	service, _ := newTestService(configuredFake())
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := service.Job(context.Background(), "bob", "upscale-test"); !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestServiceCancelTerminatesUpstreamTask(t *testing.T) {
	api := configuredFake()
	api.terminateOutcome = TerminateOutcome{Terminated: []string{api.taskID}, Failed: map[string]string{}}
	service, _ := newTestService(api)
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}

	job, err := service.Cancel(context.Background(), "alice", "upscale-test")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if job.Status != StatusCancelled {
		t.Fatalf("job=%+v", job)
	}
	if len(api.terminateCalls) != 1 || len(api.terminateCalls[0]) != 1 || api.terminateCalls[0][0] != api.taskID {
		t.Fatalf("terminate calls=%v", api.terminateCalls)
	}
}

func TestServiceCancelConflictKeepsRunning(t *testing.T) {
	api := configuredFake()
	api.terminateOutcome = TerminateOutcome{Failed: map[string]string{api.taskID: "task termination conflicts with its current state"}}
	service, _ := newTestService(api)
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}

	job, err := service.Cancel(context.Background(), "alice", "upscale-test")
	if err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("err=%v", err)
	}
	if job.Status != StatusRunning {
		t.Fatalf("job=%+v", job)
	}
}

func TestServiceCancelIsIdempotentForTerminalJobs(t *testing.T) {
	api := configuredFake()
	api.task = Task{TaskID: api.taskID, Status: "completed", VideoURL: "https://cdn.example.com/out.mp4"}
	service, _ := newTestService(api)
	if _, err := service.Submit(context.Background(), "alice", strings.NewReader("x"), 1, "video/mp4", "a.mp4", "1080p"); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := service.Job(context.Background(), "alice", "upscale-test"); err != nil {
		t.Fatalf("job: %v", err)
	}

	job, err := service.Cancel(context.Background(), "alice", "upscale-test")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if job.Status != StatusCompleted || len(api.terminateCalls) != 0 {
		t.Fatalf("job=%+v calls=%v", job, api.terminateCalls)
	}
}
