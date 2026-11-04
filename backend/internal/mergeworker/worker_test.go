package mergeworker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkerProcessOneCompletesJob(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	queue := NewMemoryQueue()
	job, err := store.Create(ctx, Job{
		ID: "merge-task-1", BatchID: "batch-1", Status: StateQueued,
		Sources: []Source{{ProductionJobID: "p1", VideoID: "v1", MediaURL: "https://media.example/1.mp4", Order: 0}},
		Speed:   1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(ctx, job.ID); err != nil {
		t.Fatal(err)
	}

	download := &fakeDownloader{}
	merge := &fakeMerger{}
	output := &fakeObjectStore{url: "https://cdn.example.com/merged.mp4"}
	worker := &Worker{Store: store, Queue: queue, Downloader: download, Merger: merge, Output: output, WorkRoot: t.TempDir()}
	if err := worker.ProcessOne(ctx, time.Second); err != nil {
		t.Fatal(err)
	}
	updated, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != StateSucceeded || updated.OutputURL != output.url || updated.ErrorMessage != "" {
		t.Fatalf("updated=%+v", updated)
	}
	if !download.called || !merge.called || !output.called {
		t.Fatalf("download=%v merge=%v output=%v", download.called, merge.called, output.called)
	}
}

func TestWorkerProcessOneMarksFailure(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	queue := NewMemoryQueue()
	job, _ := store.Create(ctx, Job{ID: "merge-task-2", BatchID: "batch-1", Status: StateQueued, Sources: []Source{{ProductionJobID: "p1", VideoID: "v1", MediaURL: "https://media.example/1.mp4", Order: 0}}})
	_ = queue.Enqueue(ctx, job.ID)
	worker := &Worker{Store: store, Queue: queue, Downloader: &fakeDownloader{err: errors.New("network detail that must not leak")}, Merger: &fakeMerger{}, Output: &fakeObjectStore{}, WorkRoot: t.TempDir()}
	if err := worker.ProcessOne(ctx, time.Second); err == nil {
		t.Fatal("expected processing error")
	}
	updated, _ := store.Get(ctx, job.ID)
	if updated.Status != StateFailed || updated.ErrorMessage != "merge execution failed" {
		t.Fatalf("updated=%+v", updated)
	}
}

func TestWorkerSkipsAlreadyTerminalJob(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	queue := NewMemoryQueue()
	job, _ := store.Create(ctx, Job{ID: "merge-task-3", BatchID: "batch-1", Status: StateSucceeded, OutputURL: "https://cdn.example.com/already.mp4"})
	_ = queue.Enqueue(ctx, job.ID)
	download := &fakeDownloader{}
	worker := &Worker{Store: store, Queue: queue, Downloader: download, Merger: &fakeMerger{}, Output: &fakeObjectStore{}, WorkRoot: t.TempDir()}
	if err := worker.ProcessOne(ctx, time.Second); err != nil {
		t.Fatal(err)
	}
	if download.called {
		t.Fatal("terminal job should not execute again")
	}
}

type fakeDownloader struct {
	called bool
	err    error
}

func (f *fakeDownloader) Download(_ context.Context, _ []Source, dir string) ([]string, error) {
	f.called = true
	if f.err != nil {
		return nil, f.err
	}
	p := filepath.Join(dir, "0000.mp4")
	if err := os.WriteFile(p, []byte("video"), 0o600); err != nil {
		return nil, err
	}
	return []string{p}, nil
}

type fakeMerger struct {
	called bool
	err    error
}

func (f *fakeMerger) Merge(_ context.Context, _ []string, output string, _ float64) error {
	f.called = true
	if f.err != nil {
		return f.err
	}
	return os.WriteFile(output, []byte("merged"), 0o600)
}

type fakeObjectStore struct {
	called bool
	url    string
	err    error
}

func (f *fakeObjectStore) PutMerged(_ context.Context, _ string, _ string) (string, error) {
	f.called = true
	if f.err != nil {
		return "", f.err
	}
	return f.url, nil
}
