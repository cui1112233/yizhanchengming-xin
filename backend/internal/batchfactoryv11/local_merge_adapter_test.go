package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/mergeworker"
)

type localMergeDownloadFunc func(context.Context, []MergeMedia, string, func(int, int)) ([]string, error)

func (fn localMergeDownloadFunc) Download(ctx context.Context, sources []MergeMedia, dir string, onProgress func(int, int)) ([]string, error) {
	return fn(ctx, sources, dir, onProgress)
}

type localMergeRunFunc func(context.Context, []string, string, float64) error

func (fn localMergeRunFunc) Merge(ctx context.Context, inputs []string, output string, speed float64) error {
	return fn(ctx, inputs, output, speed)
}

func TestLocalMergeAdapterCreatesProtectedArtifactForCompletedMerge(t *testing.T) {
	artifactRoot := t.TempDir()
	artifacts := localartifact.NewStore(artifactRoot, 1<<20)
	adapter := NewLocalMergeAdapter(artifacts)
	if adapter.WorkRoot != artifactRoot {
		t.Fatalf("work root=%q want artifact root=%q", adapter.WorkRoot, artifactRoot)
	}
	adapter.Downloader = localMergeDownloadFunc(func(_ context.Context, _ []MergeMedia, dir string, onProgress func(int, int)) ([]string, error) {
		input := filepath.Join(dir, "input.mp4")
		err := os.WriteFile(input, []byte("input"), 0o600)
		if err == nil && onProgress != nil {
			onProgress(1, 1)
		}
		return []string{input}, err
	})
	adapter.Merger = localMergeRunFunc(func(_ context.Context, _ []string, output string, _ float64) error {
		return os.WriteFile(output, []byte("0000ftypisom-local-merged-video"), 0o600)
	})

	queued, err := adapter.Submit(context.Background(), "batch-1", []MergeMedia{{ProductionJobID: "job-1", VideoID: "video-1", MediaURL: "https://media.example/one.mp4", Order: 0}}, MergeOptions{Speed: 1})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if queued.Status != MergeQueued || queued.ProviderTaskID == "" || queued.ProgressPhase != "queued" || queued.ProgressTotal != 1 {
		t.Fatalf("queued=%+v", queued)
	}

	deadline := time.Now().Add(time.Second)
	for {
		result, pollErr := adapter.Poll(context.Background(), "batch-1", queued)
		if pollErr != nil {
			t.Fatalf("poll: %v", pollErr)
		}
		if result.Status == MergeSucceeded {
			if result.ProgressPhase != "completed" || result.ProgressCurrent != 1 || result.ProgressTotal != 1 {
				t.Fatalf("completed progress=%+v", result)
			}
			if !strings.HasPrefix(result.OutputURL, "/api/batch-factory/v12/batches/batch-1/merge-media/") {
				t.Fatalf("output=%q", result.OutputURL)
			}
			artifactID := filepath.Base(result.OutputURL)
			file, openErr := artifacts.Open(artifactID + ".mp4")
			if openErr != nil {
				t.Fatalf("open merged artifact: %v", openErr)
			}
			_ = file.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("merge did not finish: %+v", result)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLocalMergeAdapterReclaimsOnlyStaleInterruptedMergeWorkspaces(t *testing.T) {
	root := t.TempDir()
	stale := filepath.Join(root, "qiantie-local-merge-stale")
	recent := filepath.Join(root, "qiantie-local-merge-recent")
	unrelated := filepath.Join(root, "merge_retained_output")
	for _, dir := range []string{stale, recent, unrelated} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-2 * localMergeWorkspaceRecoveryAge)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}

	_ = NewLocalMergeAdapter(localartifact.NewStore(root, 1<<20))

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale interrupted workspace should be removed, stat err=%v", err)
	}
	for _, dir := range []string{recent, unrelated} {
		if _, err := os.Stat(dir); err != nil {
			t.Fatalf("workspace %q should be retained: %v", dir, err)
		}
	}
}

func TestLocalMergeAdapterCreatesMissingWorkspaceBeforeMerge(t *testing.T) {
	// 容器重建后挂载的工作目录还没被创建时，合成必须自动建目录并成功，
	// 而不是 MkdirTemp 直接报 "no such file or directory" 让所有合成失败。
	missing := filepath.Join(t.TempDir(), "nested", "local-executor-artifacts")
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("workspace should not exist yet, stat err=%v", err)
	}
	adapter := NewLocalMergeAdapter(localartifact.NewStore(t.TempDir(), 1<<20))
	adapter.WorkRoot = missing
	adapter.Downloader = localMergeDownloadFunc(func(_ context.Context, _ []MergeMedia, dir string, _ func(int, int)) ([]string, error) {
		input := filepath.Join(dir, "input.mp4")
		return []string{input}, os.WriteFile(input, []byte("input"), 0o600)
	})
	adapter.Merger = localMergeRunFunc(func(_ context.Context, _ []string, output string, _ float64) error {
		return os.WriteFile(output, []byte("0000ftypisom-local-merged-video"), 0o600)
	})

	queued, err := adapter.Submit(context.Background(), "batch-1", []MergeMedia{{ProductionJobID: "job-1", VideoID: "video-1", MediaURL: "https://media.example/one.mp4", Order: 0}}, MergeOptions{Speed: 1})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		result, pollErr := adapter.Poll(context.Background(), "batch-1", queued)
		if pollErr != nil {
			t.Fatalf("poll: %v", pollErr)
		}
		if result.Status == MergeFailed {
			t.Fatalf("merge failed: %s", result.ErrorMessage)
		}
		if result.Status == MergeSucceeded {
			if info, statErr := os.Stat(missing); statErr != nil || !info.IsDir() {
				t.Fatalf("workspace dir was not created: %v", statErr)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("merge did not finish: %+v", result)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

type localMergeOutputFunc func(context.Context, string, string) (string, error)

func (fn localMergeOutputFunc) PutMerged(ctx context.Context, taskID, filePath string) (string, error) {
	return fn(ctx, taskID, filePath)
}

var _ mergeworker.ObjectStore = localMergeOutputFunc(nil)

func TestLocalMergeAdapterStoresCompletedOutputInTOS(t *testing.T) {
	artifactDir := t.TempDir()
	artifacts := localartifact.NewStore(artifactDir, 1<<20)
	adapter := NewLocalMergeAdapter(artifacts)
	adapter.Output = localMergeOutputFunc(func(_ context.Context, taskID, filePath string) (string, error) {
		if taskID == "" {
			t.Fatal("task id is required")
		}
		if _, err := os.Stat(filePath); err != nil {
			t.Fatalf("merged output must exist before TOS upload: %v", err)
		}
		return "https://media.example/batch-merged/" + taskID + ".mp4", nil
	})
	adapter.Downloader = localMergeDownloadFunc(func(_ context.Context, _ []MergeMedia, dir string, _ func(int, int)) ([]string, error) {
		input := filepath.Join(dir, "input.mp4")
		return []string{input}, os.WriteFile(input, []byte("input"), 0o600)
	})
	adapter.Merger = localMergeRunFunc(func(_ context.Context, _ []string, output string, _ float64) error {
		return os.WriteFile(output, []byte("0000ftypisom-local-merged-video"), 0o600)
	})

	queued, err := adapter.Submit(context.Background(), "batch-1", []MergeMedia{{ProductionJobID: "job-1", VideoID: "video-1", MediaURL: "https://media.example/one.mp4", Order: 0}}, MergeOptions{Speed: 1})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		result, pollErr := adapter.Poll(context.Background(), "batch-1", queued)
		if pollErr != nil {
			t.Fatalf("poll: %v", pollErr)
		}
		if result.Status == MergeSucceeded {
			if got, want := result.OutputURL, "https://media.example/batch-merged/"+queued.ProviderTaskID+".mp4"; got != want {
				t.Fatalf("output=%q want=%q", got, want)
			}
			entries, readErr := os.ReadDir(artifactDir)
			if readErr != nil {
				t.Fatalf("read artifact dir: %v", readErr)
			}
			if len(entries) != 0 {
				t.Fatalf("TOS merge must not persist a local finished artifact: %v", entries)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("merge did not finish: %+v", result)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLocalMergeAdapterReservesAnHourForSlowRemoteOutput(t *testing.T) {
	adapter := NewLocalMergeAdapter(localartifact.NewStore(t.TempDir(), 1<<20))
	adapter.Downloader = localMergeDownloadFunc(func(_ context.Context, _ []MergeMedia, dir string, _ func(int, int)) ([]string, error) {
		input := filepath.Join(dir, "input.mp4")
		return []string{input}, os.WriteFile(input, []byte("input"), 0o600)
	})
	adapter.Merger = localMergeRunFunc(func(_ context.Context, _ []string, output string, _ float64) error {
		return os.WriteFile(output, []byte("0000ftypisom-local-merged-video"), 0o600)
	})
	deadlineChecked := make(chan error, 1)
	adapter.Output = localMergeOutputFunc(func(ctx context.Context, taskID, _ string) (string, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			deadlineChecked <- errors.New("remote output needs a bounded context")
			return "", errors.New("missing deadline")
		}
		if remaining := time.Until(deadline); remaining < 59*time.Minute {
			deadlineChecked <- fmt.Errorf("remote output only has %s remaining", remaining.Round(time.Second))
			return "", errors.New("remote deadline is too short")
		}
		deadlineChecked <- nil
		return "https://media.example/batch-merged/" + taskID + ".mp4", nil
	})

	queued, err := adapter.Submit(context.Background(), "batch-1", []MergeMedia{{ProductionJobID: "job-1", VideoID: "video-1", MediaURL: "https://media.example/one.mp4", Order: 0}}, MergeOptions{Speed: 1})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		result, pollErr := adapter.Poll(context.Background(), "batch-1", queued)
		if pollErr != nil {
			t.Fatalf("poll: %v", pollErr)
		}
		if result.Status == MergeSucceeded || result.Status == MergeFailed {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("merge did not finish: %+v", result)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := <-deadlineChecked; err != nil {
		t.Fatal(err)
	}
}

func TestLocalMergeAdapterSerializesActiveMergeAndUploadWorkspaces(t *testing.T) {
	// The book scheduler's concurrency does not constrain this adapter: it can
	// submit many completed books together. A single host must not begin every
	// merge/upload at once, otherwise the shared TOS egress is saturated.
	adapter := NewLocalMergeAdapter(localartifact.NewStore(t.TempDir(), 1<<20))
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	adapter.Downloader = localMergeDownloadFunc(func(_ context.Context, _ []MergeMedia, dir string, _ func(int, int)) ([]string, error) {
		input := filepath.Join(dir, "input.mp4")
		return []string{input}, os.WriteFile(input, []byte("input"), 0o600)
	})
	adapter.Merger = localMergeRunFunc(func(_ context.Context, _ []string, output string, _ float64) error {
		started <- struct{}{}
		<-release
		return os.WriteFile(output, []byte("0000ftypisom-local-merged-video"), 0o600)
	})

	source := []MergeMedia{{ProductionJobID: "job-1", VideoID: "video-1", MediaURL: "https://media.example/one.mp4", Order: 0}}
	first, err := adapter.Submit(context.Background(), "batch-1", source, MergeOptions{Speed: 1})
	if err != nil {
		t.Fatalf("submit first: %v", err)
	}
	second, err := adapter.Submit(context.Background(), "batch-1", source, MergeOptions{Speed: 1})
	if err != nil {
		t.Fatalf("submit second: %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first merge did not start")
	}
	select {
	case <-started:
		t.Fatal("second merge started while the first still owned the merge/upload lane")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("queued merge did not start after the lane was released")
	}
	deadline := time.Now().Add(time.Second)
	for {
		one, oneErr := adapter.Poll(context.Background(), "batch-1", first)
		two, twoErr := adapter.Poll(context.Background(), "batch-1", second)
		if oneErr != nil || twoErr != nil {
			t.Fatalf("poll errors: first=%v second=%v", oneErr, twoErr)
		}
		if one.Status == MergeSucceeded && two.Status == MergeSucceeded {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("merges did not finish: first=%+v second=%+v", one, two)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLocalMergeAdapterResumesPersistedRunningJobAfterProcessRestart(t *testing.T) {
	// MergeJob and its source selection are persisted. Losing the adapter's
	// in-memory task map during a deploy must resume that job, not report a
	// fabricated permanent failure that requires another manual retry.
	adapter := NewLocalMergeAdapter(localartifact.NewStore(t.TempDir(), 1<<20))
	adapter.Downloader = localMergeDownloadFunc(func(_ context.Context, _ []MergeMedia, dir string, _ func(int, int)) ([]string, error) {
		input := filepath.Join(dir, "input.mp4")
		return []string{input}, os.WriteFile(input, []byte("input"), 0o600)
	})
	adapter.Merger = localMergeRunFunc(func(_ context.Context, _ []string, output string, _ float64) error {
		return os.WriteFile(output, []byte("0000ftypisom-local-merged-video"), 0o600)
	})
	persisted := MergeJob{
		ProviderTaskID: "local-merge-persisted",
		Status:         MergeRunning,
		Speed:          1,
		Sources:        []MergeMedia{{ProductionJobID: "job-1", VideoID: "video-1", MediaURL: "https://media.example/one.mp4", Order: 0}},
	}

	deadline := time.Now().Add(time.Second)
	for {
		result, err := adapter.Poll(context.Background(), "batch-1", persisted)
		if err != nil {
			t.Fatalf("poll: %v", err)
		}
		if result.Status == MergeFailed {
			t.Fatalf("persisted active merge must resume, got failure: %+v", result)
		}
		if result.Status == MergeSucceeded {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("persisted active merge did not resume: %+v", result)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
