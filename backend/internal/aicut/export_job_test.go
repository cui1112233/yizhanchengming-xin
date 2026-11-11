package aicut

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStoreExportQueueClaimsOldestAndKeepsSnapshot(t *testing.T) {
	store := NewMemoryStore()
	now := time.Date(2026, 11, 11, 11, 0, 0, 0, time.UTC)
	project := Project{ID: "cut-1", Title: "队列工程", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	first := ExportJob{ID: "export-1", Owner: "alice", ProjectID: project.ID, Format: "mp4", ExpectedRevision: 3, State: ExportQueued, Options: ExportOptions{Resolution: "1080p", FrameRate: 30, Quality: "standard", Codec: "h264", BitDepth: 8}, ProjectSnapshot: project, TimelineSnapshot: Timeline{Revision: 3, Tracks: []Track{}}, CreatedAt: now, UpdatedAt: now}
	second := first
	second.ID, second.Format, second.CreatedAt, second.UpdatedAt = "export-2", "mp3", now.Add(time.Second), now.Add(time.Second)
	second.Options = ExportOptions{}
	for _, job := range []ExportJob{first, second} {
		if err := store.CreateExportJob(context.Background(), "alice", job); err != nil {
			t.Fatal(err)
		}
	}
	claimed, err := store.ClaimNextExportJob(context.Background(), now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if claimed.ID != first.ID || claimed.State != ExportRunning || claimed.TimelineSnapshot.Revision != 3 || claimed.ProjectSnapshot.AspectRatio != "16:9" {
		t.Fatalf("claimed=%#v", claimed)
	}
	items, err := store.ListExportJobs(context.Background(), "alice", project.ID)
	if err != nil || len(items) != 2 || items[0].ID != second.ID || items[1].ID != first.ID {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}

func TestMemoryStoreExportCancelIsIdempotentForQueuedRunningAndFinished(t *testing.T) {
	store := NewMemoryStore()
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	project := Project{ID: "cut-1", Title: "队列工程", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	job := ExportJob{ID: "export-1", Owner: "alice", ProjectID: project.ID, Format: "mp4", State: ExportQueued, ProjectSnapshot: project, TimelineSnapshot: Timeline{Tracks: []Track{}}, CreatedAt: now, UpdatedAt: now}
	if err := store.CreateExportJob(context.Background(), "alice", job); err != nil {
		t.Fatal(err)
	}
	canceled, err := store.RequestExportCancel(context.Background(), "alice", project.ID, job.ID, now.Add(time.Second))
	if err != nil || canceled.State != ExportCanceled || canceled.Progress != 100 {
		t.Fatalf("canceled=%#v err=%v", canceled, err)
	}
	again, err := store.RequestExportCancel(context.Background(), "alice", project.ID, job.ID, now.Add(2*time.Second))
	if err != nil || again.State != ExportCanceled {
		t.Fatalf("again=%#v err=%v", again, err)
	}
	if _, err := store.ClaimNextExportJob(context.Background(), now.Add(3*time.Second)); !errors.Is(err, ErrExportQueueEmpty) {
		t.Fatalf("claim err=%v", err)
	}
}

func TestExportJobValidationRejectsInvalidStateAndFormat(t *testing.T) {
	now := time.Now().UTC()
	base := ExportJob{ID: "export-1", Owner: "alice", ProjectID: "cut-1", Format: "mp4", State: ExportQueued, ProjectSnapshot: Project{ID: "cut-1", Title: "x", AspectRatio: "16:9"}, TimelineSnapshot: Timeline{Tracks: []Track{}}, CreatedAt: now, UpdatedAt: now}
	for _, mutate := range []func(*ExportJob){
		func(job *ExportJob) { job.Format = "avi" },
		func(job *ExportJob) { job.State = "mystery" },
		func(job *ExportJob) { job.Progress = 101 },
		func(job *ExportJob) { job.TimelineSnapshot = Timeline{}; job.ExpectedRevision = 2 },
	} {
		job := base
		mutate(&job)
		if err := validateExportJob(job); err == nil {
			t.Fatalf("accepted invalid job: %#v", job)
		}
	}
}
