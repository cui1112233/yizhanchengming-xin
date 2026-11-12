package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
)

type exportTestNotifier struct{ ids []string }

func (n *exportTestNotifier) Notify(_ context.Context, id string) error {
	n.ids = append(n.ids, id)
	return nil
}

type failingExportNotifier struct{}

func (failingExportNotifier) Notify(context.Context, string) error {
	return errors.New("redis unavailable")
}

type exportTestObjects struct{}

func (exportTestObjects) Put(context.Context, agentworkspace.Asset, io.Reader) error { return nil }
func (exportTestObjects) Open(context.Context, agentworkspace.Asset) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), errors.New("not used")
}
func (exportTestObjects) Delete(context.Context, agentworkspace.Asset) error { return nil }

func TestAICutExportServiceSubmitsMultipleFormatsAndRejectsExactDuplicate(t *testing.T) {
	store := aicut.NewMemoryStore()
	now := time.Date(2026, 11, 11, 15, 0, 0, 0, time.UTC)
	project := aicut.Project{ID: "cut-1", Title: "批量导出", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	notifier := &exportTestNotifier{}
	service := NewAICutExportService(store, exportTestObjects{}, notifier, func() time.Time { return now })
	timeline := aicut.Timeline{Revision: 0, Tracks: []aicut.Track{}}
	settings := aicut.ExportOptions{Resolution: "1080p", FrameRate: 30, Quality: "standard", Codec: "h264", BitDepth: 8}
	first, err := service.Submit(context.Background(), "alice", project, timeline, "mp4", settings)
	if err != nil || first.State != aicut.ExportQueued {
		t.Fatalf("first=%#v err=%v", first, err)
	}
	if _, err := service.Submit(context.Background(), "alice", project, timeline, "mp4", settings); err == nil {
		t.Fatal("duplicate export was accepted")
	}
	second, err := service.Submit(context.Background(), "alice", project, timeline, "mp3", aicut.ExportOptions{})
	if err != nil || second.ID == first.ID {
		t.Fatalf("second=%#v err=%v", second, err)
	}
	jobs, err := service.List(context.Background(), "alice", project.ID)
	if err != nil || len(jobs) != 2 || len(notifier.ids) != 2 {
		t.Fatalf("jobs=%#v notified=%#v err=%v", jobs, notifier.ids, err)
	}
}

func TestAICutExportServiceCancelsQueuedJobIdempotently(t *testing.T) {
	store := aicut.NewMemoryStore()
	now := time.Date(2026, 11, 11, 16, 0, 0, 0, time.UTC)
	project := aicut.Project{ID: "cut-1", Title: "批量导出", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	service := NewAICutExportService(store, exportTestObjects{}, &exportTestNotifier{}, func() time.Time { return now })
	job, err := service.Submit(context.Background(), "alice", project, aicut.Timeline{Tracks: []aicut.Track{}}, "mp3", aicut.ExportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	canceled, err := service.Cancel(context.Background(), "alice", project.ID, job.ID)
	if err != nil || canceled.State != aicut.ExportCanceled || canceled.Progress != 100 {
		t.Fatalf("canceled=%#v err=%v", canceled, err)
	}
	again, err := service.Cancel(context.Background(), "alice", project.ID, job.ID)
	if err != nil || again.State != aicut.ExportCanceled {
		t.Fatalf("again=%#v err=%v", again, err)
	}
}

func TestAICutExportServicePersistsNotifierFailureForVisibleRetry(t *testing.T) {
	store := aicut.NewMemoryStore()
	now := time.Date(2026, 11, 11, 17, 0, 0, 0, time.UTC)
	project := aicut.Project{ID: "cut-redis", Title: "队列失败", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	service := NewAICutExportService(store, exportTestObjects{}, failingExportNotifier{}, func() time.Time { return now })
	job, err := service.Submit(context.Background(), "alice", project, aicut.Timeline{Tracks: []aicut.Track{}}, "mp3", aicut.ExportOptions{})
	if err == nil || job.State != aicut.ExportFailed || job.Progress != 100 || job.Error != "Redis 导出队列不可用，请稍后重试" {
		t.Fatalf("job=%#v err=%v", job, err)
	}
	jobs, listErr := service.List(context.Background(), "alice", project.ID)
	if listErr != nil || len(jobs) != 1 || jobs[0].State != aicut.ExportFailed {
		t.Fatalf("jobs=%#v err=%v", jobs, listErr)
	}
}

func TestAICutExportServiceNormalizesSnapshotOnlyBackgroundForDuplicateDetection(t *testing.T) {
	store := aicut.NewMemoryStore()
	now := time.Date(2026, 11, 11, 18, 0, 0, 0, time.UTC)
	project := aicut.Project{ID: "cut-bg", Title: "背景快照", AspectRatio: "16:9", BackgroundColor: "#123456", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	service := NewAICutExportService(store, exportTestObjects{}, &exportTestNotifier{}, func() time.Time { return now })
	timeline := aicut.Timeline{Revision: 3, Tracks: []aicut.Track{}}
	options := aicut.ExportOptions{Resolution: "1080p", FrameRate: 30, Quality: "standard", Codec: "h264", BitDepth: 8, BackgroundColor: "#123456"}
	job, err := service.Submit(context.Background(), "alice", project, timeline, "mp4", options)
	if err != nil || job.Options.BackgroundColor != "" || job.ProjectSnapshot.BackgroundColor != "#123456" {
		t.Fatalf("job=%#v err=%v", job, err)
	}
	if _, err := service.Submit(context.Background(), "alice", project, timeline, "mp4", options); err == nil {
		t.Fatal("same snapshot was accepted twice after option normalization")
	}
}

func TestAICutExportServiceNamedTimelineSnapshotAndJobIsolation(t *testing.T) {
	ctx := context.Background()
	base := aicut.NewMemoryStore()
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	project := aicut.Project{ID: "p", Title: "P", CreatedAt: now, UpdatedAt: now}
	_ = base.CreateProject(ctx, "alice", project)
	collection := aicut.NewMemoryTimelineCollectionStore(base)
	a, _ := collection.Create(ctx, "alice", "p", "A")
	b, _ := collection.Create(ctx, "alice", "p", "B")
	named, _ := collection.Save(ctx, "alice", "p", a.ID, aicut.Timeline{Script: "named snapshot"}, 0, now)
	svc := NewAICutExportService(base, exportTestObjects{}, &exportTestNotifier{}, func() time.Time { return now }, collection)
	job, e := svc.Submit(ctx, "alice", project, named, "mp3", aicut.ExportOptions{TimelineID: a.ID})
	if e != nil {
		t.Fatal(e)
	}
	if job.Options.TimelineID != a.ID || job.TimelineSnapshot.Script != "named snapshot" {
		t.Fatalf("job=%+v", job)
	}
	_, _ = collection.Save(ctx, "alice", "p", a.ID, aicut.Timeline{Script: "later"}, 1, now)
	persisted, _ := base.ExportJob(ctx, "alice", "p", job.ID)
	if persisted.TimelineSnapshot.Script != "named snapshot" {
		t.Fatal("snapshot mutated")
	}
	for _, id := range []string{aicut.DefaultTimelineID, b.ID} {
		scope := context.WithValue(ctx, aicutTimelineScopeKey{}, id)
		items, e := svc.List(scope, "alice", "p")
		if e != nil || len(items) != 0 {
			t.Fatalf("scope=%s list=%+v %v", id, items, e)
		}
		if _, e = svc.Job(scope, "alice", "p", job.ID); !errors.Is(e, aicut.ErrExportJobNotFound) {
			t.Fatalf("scope=%s job=%v", id, e)
		}
		if _, e = svc.Cancel(scope, "alice", "p", job.ID); !errors.Is(e, aicut.ErrExportJobNotFound) {
			t.Fatalf("scope=%s cancel=%v", id, e)
		}
	}
	scope := context.WithValue(ctx, aicutTimelineScopeKey{}, a.ID)
	if _, e = svc.Cancel(scope, "alice", "p", job.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Submit(ctx, "alice", project, named, "mp3", aicut.ExportOptions{TimelineID: a.ID}); !errors.Is(e, aicut.ErrRevisionConflict) {
		t.Fatalf("stale named=%v", e)
	}
}
