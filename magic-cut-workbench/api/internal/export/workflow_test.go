package export

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"yizhanchengming/magic-cut-workbench/internal/material"
	worktask "yizhanchengming/magic-cut-workbench/internal/task"
	"yizhanchengming/magic-cut-workbench/internal/timeline"
)

type timelineFixture struct{ value timeline.Timeline }

func (f timelineFixture) Get(context.Context, string, string) (timeline.Timeline, error) {
	return f.value, nil
}

type materialFixture struct{ items []material.Material }

func (f materialFixture) List(context.Context, string, string) ([]material.Material, error) {
	return f.items, nil
}

type repositoryFixture struct {
	created Export
	task    worktask.Task
	started bool
	result  Result
	meta    material.Metadata
	failed  error
}

func (r *repositoryFixture) Create(_ context.Context, item Export, task worktask.Task) error {
	r.created, r.task = item, task
	return nil
}
func (r *repositoryFixture) Start(_ context.Context, _ string, _ time.Time) (int, error) {
	r.started = true
	r.task.Status = worktask.StatusRunning
	return 1, nil
}
func (r *repositoryFixture) Progress(_ context.Context, _ string, progress float64, _ time.Time) error {
	r.task.Progress = progress
	return nil
}
func (r *repositoryFixture) Succeed(_ context.Context, _ string, _ string, result Result, metadata material.Metadata, _ time.Time) error {
	r.result, r.meta = result, metadata
	r.task.Status, r.task.Progress = worktask.StatusSucceeded, 1
	return nil
}
func (r *repositoryFixture) Fail(_ context.Context, _ string, _ worktask.Status, _ string, message string, _ time.Time) error {
	r.failed = &ServiceError{Code: "FAILED", Err: context.Canceled}
	r.task.ErrorMessage = message
	return nil
}
func (r *repositoryFixture) TaskStatus(context.Context, string, string) (worktask.Status, error) {
	return r.task.Status, nil
}
func (r *repositoryFixture) Get(context.Context, string, string) (Export, error) {
	return r.created, nil
}

type rendererFixture struct{ plan RenderPlan }

func (r *rendererFixture) Run(_ context.Context, plan RenderPlan, progress func(float64)) (Result, error) {
	r.plan = plan
	progress(0.5)
	progress(1)
	return Result{OutputPath: plan.OutputPath, SizeBytes: 42, SHA256: "abc123"}, nil
}

type probeFixture struct{}

func (probeFixture) Probe(context.Context, string) (material.Metadata, error) {
	return material.Metadata{DurationMs: 1000, Width: 640, Height: 360, VideoCodec: "h264", AudioCodec: "aac"}, nil
}

func TestServiceCreatePersistsRunsAndFinishesExport(t *testing.T) {
	root := t.TempDir()
	repository := &repositoryFixture{}
	renderer := &rendererFixture{}
	now := time.Date(2026, 11, 20, 8, 0, 0, 0, time.UTC)
	service := NewService(ServiceOptions{
		StorageRoot: root,
		Timelines:   timelineFixture{value: timeline.Timeline{DurationMs: 1000, Tracks: []timeline.Track{{ID: "video-1", Kind: "video", Clips: []timeline.Clip{{ID: "clip_1", MaterialID: "mat_1", DurationMs: 1000, Speed: 1}}}}}},
		Materials:   materialFixture{items: []material.Material{{ID: "mat_1", StorageKey: filepath.Join(root, "source.mp4"), MediaType: "video"}}},
		Repository:  repository,
		Renderer:    renderer,
		Probe:       probeFixture{},
		NewExportID: func() string { return "mce_1" },
		NewTaskID:   func() string { return "mct_1" },
		Now:         func() time.Time { return now },
		Start:       func(work func()) { work() },
	})
	created, task, err := service.Create(context.Background(), "alice", CreateInput{ProjectID: "mcp_1", OutputName: "成片", Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "mce_1" || task.ID != "mct_1" || repository.created.TaskID != task.ID {
		t.Fatalf("created=%+v task=%+v", created, task)
	}
	if !repository.started || repository.task.Status != worktask.StatusSucceeded || repository.result.SHA256 != "abc123" {
		t.Fatalf("repository=%+v", repository)
	}
	if repository.meta.Width != 640 || filepath.Ext(renderer.plan.OutputPath) != ".mp4" {
		t.Fatalf("metadata=%+v plan=%+v", repository.meta, renderer.plan)
	}
}

func TestServiceCreateRejectsUnsafeOutputName(t *testing.T) {
	service := NewService(ServiceOptions{})
	_, _, err := service.Create(context.Background(), "alice", CreateInput{ProjectID: "mcp_1", OutputName: "../outside", Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "mp4"})
	if ErrorCode(err) != "EXPORT_NAME_INVALID" {
		t.Fatalf("error=%v code=%q", err, ErrorCode(err))
	}
}

func TestServiceRetriesOnlyOwnedFailedExportTask(t *testing.T) {
	root := t.TempDir()
	repository := &repositoryFixture{task: worktask.Task{Status: worktask.StatusFailed}}
	renderer := &rendererFixture{}
	service := NewService(ServiceOptions{
		StorageRoot: root,
		Timelines:   timelineFixture{value: timeline.Timeline{DurationMs: 1000, Tracks: []timeline.Track{{ID: "video-1", Kind: "video", Clips: []timeline.Clip{{ID: "clip_1", MaterialID: "mat_1", DurationMs: 1000, Speed: 1}}}}}},
		Materials:   materialFixture{items: []material.Material{{ID: "mat_1", StorageKey: filepath.Join(root, "source.mp4"), MediaType: "video"}}},
		Repository:  repository, Renderer: renderer, Probe: probeFixture{},
		NewExportID: func() string { return "mce_1" }, NewTaskID: func() string { return "mct_1" },
		Start: func(work func()) { work() },
	})
	_, _, err := service.Create(context.Background(), "alice", CreateInput{ProjectID: "mcp_1", OutputName: "成片", Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "mp4"})
	if err != nil {
		t.Fatal(err)
	}
	repository.task.Status = worktask.StatusFailed
	handled, err := service.RetryTask(context.Background(), "alice", "mct_1")
	if err != nil || !handled || repository.task.Status != worktask.StatusSucceeded {
		t.Fatalf("handled=%v err=%v task=%+v", handled, err, repository.task)
	}
	handled, err = service.RetryTask(context.Background(), "mallory", "mct_1")
	if err != nil || handled {
		t.Fatalf("cross-owner retry handled=%v err=%v", handled, err)
	}
}
