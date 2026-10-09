package video

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPerformanceControlledPersonalAPIConcurrency(t *testing.T) {
	var submitCount atomic.Int64
	var pollCount atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/create":
			id := submitCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"task_id":"job-%d"}`, id)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tasks/"):
			pollCount.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"status":"succeeded","output_url":"https://example.invalid/video.mp4"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider, err := NewPersonalAPIProvider(ProviderConfig{
		ProviderKey: ProviderPersonalAPI,
		Model:       ModelYD20Mini,
		CreateURL:   server.URL + "/create",
		TasksURL:    server.URL + "/tasks",
	}, "test-secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}

	for _, concurrency := range []int{10, 50, 100} {
		concurrency := concurrency
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			beforeSubmit := submitCount.Load()
			beforePoll := pollCount.Load()
			latencies := make([]time.Duration, concurrency)
			errs := make(chan error, concurrency)
			start := make(chan struct{})
			var wg sync.WaitGroup
			wg.Add(concurrency)
			started := time.Now()

			for i := 0; i < concurrency; i++ {
				i := i
				go func() {
					defer wg.Done()
					<-start
					one := time.Now()
					submitted, err := provider.Submit(context.Background(), SubmitRequest{Model: ModelYD20Mini, Prompt: "controlled load"})
					if err == nil {
						var polled PollResult
						polled, err = provider.Poll(context.Background(), submitted.ProviderJobID)
						if err == nil && polled.Status != TaskSucceeded {
							err = fmt.Errorf("poll status=%s, want %s", polled.Status, TaskSucceeded)
						}
					}
					latencies[i] = time.Since(one)
					errs <- err
				}()
			}
			close(start)
			wg.Wait()
			duration := time.Since(started)
			close(errs)

			failed := 0
			for err := range errs {
				if err != nil {
					failed++
					t.Logf("controlled provider error: %v", err)
				}
			}
			sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
			p50 := perfVideoPercentile(latencies, 0.50)
			p95 := perfVideoPercentile(latencies, 0.95)
			p99 := perfVideoPercentile(latencies, 0.99)
			requests := (submitCount.Load() - beforeSubmit) + (pollCount.Load() - beforePoll)
			rps := float64(requests) / duration.Seconds()
			t.Logf("PERF_VIDEO controlled=personal_api concurrency=%d total_tasks=%d success=%d failed=%d provider_requests=%d p50=%s p95=%s p99=%s duration=%s request_rps=%.2f",
				concurrency, concurrency, concurrency-failed, failed, requests, p50, p95, p99, duration, rps)
			if failed != 0 {
				t.Fatalf("controlled provider failed=%d", failed)
			}
			if requests != int64(concurrency*2) {
				t.Fatalf("provider requests=%d, want %d (one submit + one poll per task)", requests, concurrency*2)
			}
		})
	}
}

func TestPerformanceFFmpegExecutorConcurrencyBoundary(t *testing.T) {
	for _, concurrency := range []int{1, 2, 4} {
		concurrency := concurrency
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			runner := &perfCommandRunner{delay: 30 * time.Millisecond}
			root := t.TempDir()
			executor := NewFFmpegExecutor(FFmpegExecutorConfig{
				Binary:     "ffmpeg-test",
				TempRoot:   root,
				Runner:     runner,
				Downloader: perfDownloader{},
				Artifacts:  perfFileStore{},
			})

			start := make(chan struct{})
			errs := make(chan error, concurrency)
			var wg sync.WaitGroup
			wg.Add(concurrency)
			started := time.Now()
			for i := 0; i < concurrency; i++ {
				i := i
				go func() {
					defer wg.Done()
					<-start
					_, err := executor.Execute(context.Background(), MergeExecutionRequest{
						JobID:       int64(i + 1),
						AttemptID:   int64(i + 1),
						AspectRatio: "9:16",
						Speed:       1,
						Inputs: []MergeInputAsset{{ProductionTaskID: int64(i + 1), Order: 1, URL: "https://example.invalid/input.mp4"}},
					})
					errs <- err
				}()
			}
			close(start)
			wg.Wait()
			duration := time.Since(started)
			close(errs)
			failed := 0
			for err := range errs {
				if err != nil {
					failed++
					t.Logf("ffmpeg probe error: %v", err)
				}
			}
			maxActive := runner.maxActive.Load()
			t.Logf("PERF_FFMPEG concurrency=%d success=%d failed=%d max_active=%d duration=%s temp_root=%s",
				concurrency, concurrency-failed, failed, maxActive, duration, root)
			if failed != 0 {
				t.Fatalf("ffmpeg probe failed=%d", failed)
			}
			if maxActive < 1 || maxActive > int64(concurrency) {
				t.Fatalf("invalid max_active=%d for concurrency=%d", maxActive, concurrency)
			}
		})
	}
}

func perfVideoPercentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values))*p + 0.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

type perfCommandRunner struct {
	delay     time.Duration
	active    atomic.Int64
	maxActive atomic.Int64
}

func (r *perfCommandRunner) Run(ctx context.Context, _ string, args ...string) ([]byte, error) {
	active := r.active.Add(1)
	defer r.active.Add(-1)
	for {
		current := r.maxActive.Load()
		if active <= current || r.maxActive.CompareAndSwap(current, active) {
			break
		}
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(r.delay):
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("missing output path")
	}
	if err := os.WriteFile(args[len(args)-1], []byte("fake-video"), 0o600); err != nil {
		return nil, err
	}
	return nil, nil
}

type perfDownloader struct{}

func (perfDownloader) Download(_ context.Context, _ string, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	return os.WriteFile(destination, []byte("fake-input"), 0o600)
}

type perfFileStore struct{}

func (perfFileStore) PersistFile(_ context.Context, sourcePath, objectKey string) (Artifact, error) {
	if _, err := os.Stat(sourcePath); err != nil {
		return Artifact{}, err
	}
	return Artifact{Bucket: "test", ObjectKey: objectKey, URL: "https://example.invalid/" + objectKey}, nil
}
