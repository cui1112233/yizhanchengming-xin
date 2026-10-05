package video

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type recordingCapacityDownloader struct {
	mu   sync.Mutex
	urls []string
}

func (d *recordingCapacityDownloader) Download(_ context.Context, rawURL, destination string) error {
	d.mu.Lock()
	d.urls = append(d.urls, rawURL)
	d.mu.Unlock()
	return os.WriteFile(destination, []byte(strings.Repeat("x", 1024)), 0o600)
}

func (d *recordingCapacityDownloader) snapshot() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.urls...)
}

func TestPerformanceMergeFragmentCapacity10_20_50(t *testing.T) {
	for _, fragments := range []int{10, 20, 50} {
		fragments := fragments
		t.Run(fmt.Sprintf("fragments_%d", fragments), func(t *testing.T) {
			root := t.TempDir()
			downloader := &recordingCapacityDownloader{}
			runner := &perfCommandRunner{}
			store := &perfFileStore{}
			executor := NewFFmpegExecutor(FFmpegExecutorConfig{
				Binary: "ffmpeg-test", TempRoot: root, Runner: runner, Downloader: downloader, Artifacts: store,
			})
			inputs := make([]MergeInputAsset, 0, fragments)
			for i := fragments; i >= 1; i-- {
				inputs = append(inputs, MergeInputAsset{ProductionTaskID: int64(i), Order: i, URL: fmt.Sprintf("https://example.invalid/fragment-%03d.mp4", i)})
			}
			var before runtime.MemStats
			runtime.ReadMemStats(&before)
			started := time.Now()
			artifact, err := executor.Execute(context.Background(), MergeExecutionRequest{JobID: int64(fragments), AttemptID: 1, Inputs: inputs, AspectRatio: "9:16", Speed: 1})
			duration := time.Since(started)
			var after runtime.MemStats
			runtime.ReadMemStats(&after)
			if err != nil {
				t.Fatal(err)
			}
			gotURLs := downloader.snapshot()
			ordered := sort.SliceIsSorted(gotURLs, func(i, j int) bool { return gotURLs[i] < gotURLs[j] })
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("PERF_MERGE fragments=%d success=1 failed=0 duration=%s input_bytes=%d output_url=%s output_key=%s heap_delta=%d temp_entries_after=%d ordered=%t",
				fragments, duration, fragments*1024, artifact.URL, artifact.ObjectKey, int64(after.HeapAlloc)-int64(before.HeapAlloc), len(entries), ordered)
			if len(gotURLs) != fragments {
				t.Fatalf("downloaded=%d want=%d", len(gotURLs), fragments)
			}
			if !ordered {
				t.Fatalf("merge inputs were not normalized by order: first=%v", gotURLs[:minInt(5, len(gotURLs))])
			}
			if len(entries) != 0 {
				t.Fatalf("temp root retained %d work directories", len(entries))
			}
		})
	}
}

func TestPerformanceMergeRetryDoesNotRegenerateVideo(t *testing.T) {
	store := newMemoryMergeStore()
	executor := &recordingMergeExecutor{err: fmt.Errorf("controlled merge failure")}
	service := NewMergeService(store, executor)
	started, err := service.Start(context.Background(), MergeStartRequest{
		BatchProjectID: 1,
		BookID:         1,
		Inputs:         []MergeInputAsset{{ProductionTaskID: 11, Order: 1, URL: "https://example.invalid/1.mp4"}},
		AspectRatio:    "9:16",
		Speed:          1,
	})
	if err != nil {
		t.Fatal(err)
	}
	failed, err := service.ExecuteAttempt(context.Background(), started.Attempt.ID)
	if err == nil || failed.Status != MergeFailed {
		t.Fatalf("first merge status=%s err=%v, want failed", failed.Status, err)
	}
	retry, err := service.RetryAttempt(context.Background(), failed.ID)
	if err != nil {
		t.Fatal(err)
	}
	executor.err = nil
	executor.artifact = Artifact{Bucket: "test", ObjectKey: "merge/retry.mp4", URL: "https://example.invalid/retry.mp4"}
	completed, err := service.ExecuteAttempt(context.Background(), retry.Attempt.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("PERF_MERGE_RETRY attempts=2 first_status=%s retry_status=%s video_submit_calls=%d output=%s", failed.Status, completed.Status, executor.videoSubmitCalls, completed.OutputURL)
	if completed.Status != MergeSucceeded {
		t.Fatalf("retry status=%s want succeeded", completed.Status)
	}
	if executor.videoSubmitCalls != 0 {
		t.Fatalf("merge retry triggered video submit calls=%d", executor.videoSubmitCalls)
	}
}

type capacityUploader struct {
	delay     time.Duration
	failKey   string
	calls     atomic.Int64
	active    atomic.Int64
	maxActive atomic.Int64
}

func (u *capacityUploader) PutObjectFromFile(ctx context.Context, _ string, key, filename string) error {
	u.calls.Add(1)
	active := u.active.Add(1)
	defer u.active.Add(-1)
	for {
		current := u.maxActive.Load()
		if active <= current || u.maxActive.CompareAndSwap(current, active) {
			break
		}
	}
	if _, err := os.Stat(filename); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(u.delay):
	}
	if u.failKey != "" && strings.Contains(key, u.failKey) {
		return fmt.Errorf("controlled upload failure")
	}
	return nil
}

func TestPerformanceArtifactPersistenceConcurrency10_50(t *testing.T) {
	for _, concurrency := range []int{10, 50} {
		concurrency := concurrency
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			uploader := &capacityUploader{delay: 2 * time.Millisecond}
			store, err := NewArtifactStore(ArtifactStoreConfig{Bucket: "test", PublicBaseURL: "https://example.invalid", Uploader: uploader})
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			paths := make([]string, concurrency)
			for i := range paths {
				paths[i] = filepath.Join(root, fmt.Sprintf("artifact-%03d.bin", i))
				if err := os.WriteFile(paths[i], []byte(strings.Repeat("a", 4096)), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var before runtime.MemStats
			runtime.ReadMemStats(&before)
			started := time.Now()
			var wg sync.WaitGroup
			var failed atomic.Int64
			wg.Add(concurrency)
			for i := 0; i < concurrency; i++ {
				i := i
				go func() {
					defer wg.Done()
					if _, err := store.PersistFile(context.Background(), paths[i], fmt.Sprintf("capacity/%03d.bin", i)); err != nil {
						failed.Add(1)
					}
				}()
			}
			wg.Wait()
			duration := time.Since(started)
			var after runtime.MemStats
			runtime.ReadMemStats(&after)
			t.Logf("PERF_ARTIFACT concurrency=%d total=%d success=%d failed=%d uploader_calls=%d retries=0 max_active=%d duration=%s heap_delta=%d source_files_remaining=%d",
				concurrency, concurrency, concurrency-int(failed.Load()), failed.Load(), uploader.calls.Load(), uploader.maxActive.Load(), duration, int64(after.HeapAlloc)-int64(before.HeapAlloc), countRegularFiles(root))
			if failed.Load() != 0 || uploader.calls.Load() != int64(concurrency) {
				t.Fatalf("artifact persistence failed=%d calls=%d", failed.Load(), uploader.calls.Load())
			}
		})
	}
}

func TestPerformanceArtifactUploadFailureHasNoHiddenRetry(t *testing.T) {
	uploader := &capacityUploader{failKey: "fail"}
	store, err := NewArtifactStore(ArtifactStoreConfig{Bucket: "test", PublicBaseURL: "https://example.invalid", Uploader: uploader})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "artifact.bin")
	if err := os.WriteFile(path, []byte("controlled"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, persistErr := store.PersistFile(context.Background(), path, "capacity/fail.bin")
	t.Logf("PERF_ARTIFACT_FAILURE calls=%d retries=%d failed=%t", uploader.calls.Load(), maxInt64(0, uploader.calls.Load()-1), persistErr != nil)
	if persistErr == nil || uploader.calls.Load() != 1 {
		t.Fatalf("persist err=%v calls=%d want controlled failure with one attempt", persistErr, uploader.calls.Load())
	}
}

func countRegularFiles(root string) int {
	entries, err := os.ReadDir(root)
	if err != nil {
		return -1
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}

func minInt(a, b int) int {
	if a < b { return a }
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b { return a }
	return b
}
