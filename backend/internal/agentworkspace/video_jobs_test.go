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
	if _, err := store.UpdateVideoJob(ctx, "alice", VideoJob{ID: "job-1", ProjectID: "project-1", State: VideoJobAwaitingConfirmation, ResultAssetID: "asset-1"}, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConfirmVideoJob(ctx, "alice", "project-1", "job-1", now); !errors.Is(err, ErrVideoJobNodeChanged) {
		t.Fatalf("confirm err=%v", err)
	}
}
