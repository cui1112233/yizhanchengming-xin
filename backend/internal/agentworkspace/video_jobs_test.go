package agentworkspace

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryVideoJobsAreOwnerScopedAndIdempotent(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC)
	for _, item := range []struct{ owner, id string }{{"alice", "project-1"}, {"alice", "project-2"}, {"bob", "project-1"}} {
		if err := store.CreateProject(ctx, item.owner, Project{ID: item.id, Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}

	created, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{
		ID: "job-1", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1,
		ModelKey: "yd2-mini-video", Settings: map[string]any{"prompt": "雨夜追逐"},
		IdempotencyKey: "idem-1", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.State != VideoJobPreparing || created.OwnerUsername != "alice" {
		t.Fatalf("created=%+v", created)
	}
	if _, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{
		ID: "job-2", ProjectID: "project-2", NodeID: "video-1", NodeVersion: 1,
		ModelKey: "yd2-mini-video", IdempotencyKey: "idem-1", Now: now,
	}); !errors.Is(err, ErrVideoJobExists) {
		t.Fatalf("duplicate idempotency err=%v", err)
	}
	if _, err := store.VideoJob(ctx, "bob", "project-1", "job-1"); !errors.Is(err, ErrVideoJobNotFound) {
		t.Fatalf("cross-owner read err=%v", err)
	}
}

func TestMemoryListVideoJobsStaysInsideOwnerProjectAndReturnsNewestFirst(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	for _, item := range []struct{ owner, id string }{{"alice", "project-1"}, {"alice", "project-2"}, {"bob", "project-1"}} {
		if err := store.CreateProject(ctx, item.owner, Project{ID: item.id, Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range []struct {
		owner, id, projectID, nodeID, key string
		at                                time.Time
	}{
		{"alice", "old", "project-1", "video-1", "alice-old", now},
		{"alice", "new", "project-1", "video-1", "alice-new", now.Add(time.Minute)},
		{"alice", "other-project", "project-2", "video-2", "alice-other", now.Add(2 * time.Minute)},
		{"bob", "other-owner", "project-1", "video-3", "bob-job", now.Add(3 * time.Minute)},
	} {
		if _, err := store.CreateVideoJob(ctx, input.owner, CreateVideoJobInput{
			ID: input.id, ProjectID: input.projectID, NodeID: input.nodeID, NodeVersion: 1,
			ModelKey: "minimax-h3-video", IdempotencyKey: input.key, Now: input.at,
		}); err != nil {
			t.Fatal(err)
		}
	}

	jobs, err := store.ListVideoJobs(ctx, "alice", "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].ID != "new" || jobs[1].ID != "old" {
		t.Fatalf("jobs=%+v", jobs)
	}
}

func TestMemoryVideoJobConfirmRejectsNodeVersionConflict(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC)
	project := Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: Canvas{Version: 1, Nodes: []Node{{
		ID: "video-1", Type: "video", Title: "视频", Version: 2, Status: "draft",
	}}, Edges: []Edge{}}}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	_, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{
		ID: "job-1", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1,
		ModelKey: "yd2-mini-video", IdempotencyKey: "idem-1", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := store.VideoJob(ctx, "alice", "project-1", "job-1")
	if err != nil {
		t.Fatal(err)
	}
	for index, state := range []VideoJobState{VideoJobQueued, VideoJobRunning, VideoJobResultReady, VideoJobIngesting, VideoJobAwaitingConfirmation} {
		job.State = state
		if state == VideoJobAwaitingConfirmation {
			job.ResultAssetID = "asset-1"
		}
		job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(time.Duration(index+1)*time.Second))
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.ConfirmVideoJob(ctx, "alice", "project-1", "job-1", now); !errors.Is(err, ErrVideoJobNodeChanged) {
		t.Fatalf("confirm err=%v", err)
	}
}

func TestMemoryVideoJobRejectsStaleAndBackwardTransitions(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 20, 4, 0, 0, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{ID: "job-1", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State = VideoJobQueued
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	job.State = VideoJobRunning
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	staleRunning := cloneVideoJob(job)
	job.State = VideoJobResultReady
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	job.State = VideoJobIngesting
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(4*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	job.State, job.ResultAssetID = VideoJobAwaitingConfirmation, "asset-1"
	if _, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}

	staleRunning.State = VideoJobRunning
	if _, err = store.UpdateVideoJob(ctx, "alice", staleRunning, now.Add(6*time.Second)); !errors.Is(err, ErrVideoJobState) {
		t.Fatalf("stale running update err=%v, want ErrVideoJobState", err)
	}
	stored, err := store.VideoJob(ctx, "alice", "project-1", "job-1")
	if err != nil || stored.State != VideoJobAwaitingConfirmation {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
}

func TestMemoryVideoJobFailedStateCannotReturnToRunning(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 20, 4, 10, 0, 0, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{ID: "job-1", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State = VideoJobFailed
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	job.State = VideoJobRunning
	if _, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(2*time.Second)); !errors.Is(err, ErrVideoJobState) {
		t.Fatalf("failed-to-running err=%v, want ErrVideoJobState", err)
	}
}

func TestVideoJobTimestampsUseMySQLMicrosecondPrecision(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 20, 6, 0, 0, 123456789, time.UTC)
	if err := store.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	job, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{ID: "job-1", ProjectID: "project-1", NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if job.CreatedAt.Nanosecond()%1000 != 0 || job.UpdatedAt.Nanosecond()%1000 != 0 {
		t.Fatalf("created=%s updated=%s, want MySQL DATETIME(6) precision", job.CreatedAt, job.UpdatedAt)
	}
	job.State = VideoJobQueued
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(987654321*time.Nanosecond))
	if err != nil {
		t.Fatal(err)
	}
	if job.UpdatedAt.Nanosecond()%1000 != 0 {
		t.Fatalf("updated=%s, want MySQL DATETIME(6) precision", job.UpdatedAt)
	}
}

func TestMemoryRetryVideoJobRejectsChangedNodeVersion(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 20, 6, 10, 0, 0, time.UTC)
	project := Project{ID: "project-1", Title: "项目", CreatedAt: now, UpdatedAt: now, Canvas: Canvas{Version: 1, Nodes: []Node{{ID: "video-1", Type: "video", Version: 2}}}}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	job, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{ID: "job-1", ProjectID: project.ID, NodeID: "video-1", NodeVersion: 1, ModelKey: "minimax-h3-video", IdempotencyKey: "request-1", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	job.State = VideoJobFailed
	job, err = store.UpdateVideoJob(ctx, "alice", job, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.RetryVideoJob(ctx, "alice", project.ID, job.ID, RetryVideoJobInput{ID: "job-2", IdempotencyKey: "request-2", Now: now.Add(2 * time.Second)})
	if !errors.Is(err, ErrVideoJobNodeChanged) {
		t.Fatalf("retry err=%v, want ErrVideoJobNodeChanged", err)
	}
}
