package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMergeRetryCreatesNewAttemptAndReusesSuccessfulVideoAssets(t *testing.T) {
	store := newMemoryMergeStore()
	executor := &recordingMergeExecutor{err: errors.New("ffmpeg failed")}
	service := NewMergeService(store, executor)
	inputs := []MergeInputAsset{
		{ProductionTaskID: 101, URL: "https://cdn.example/video-1.mp4", Order: 1},
		{ProductionTaskID: 102, URL: "https://cdn.example/video-2.mp4", Order: 2},
	}

	started, err := service.Start(context.Background(), MergeStartRequest{BatchProjectID: 7, BookID: 9, Inputs: inputs, AspectRatio: "9:16", Speed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if started.Attempt.Attempt != 1 || started.Attempt.Status != MergeQueued {
		t.Fatalf("start = %+v", started)
	}
	_, err = service.ExecuteAttempt(context.Background(), started.Attempt.ID)
	if err == nil {
		t.Fatal("expected first merge attempt to fail")
	}
	failed, err := store.GetMergeAttempt(context.Background(), started.Attempt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if failed.Status != MergeFailed || failed.ErrorMessage == "" {
		t.Fatalf("failed attempt = %+v", failed)
	}

	executor.err = nil
	executor.artifact = Artifact{Bucket: "videos", ObjectKey: "merge/7/9/2.mp4", URL: "https://tos.example/merge.mp4"}
	retried, err := service.RetryAttempt(context.Background(), failed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if retried.Attempt.Attempt != 2 || retried.Attempt.ID == failed.ID {
		t.Fatalf("retry did not create a new attempt: %+v", retried)
	}
	if got := retried.Attempt.Inputs; len(got) != 2 || got[0].ProductionTaskID != 101 || got[1].ProductionTaskID != 102 {
		t.Fatalf("retry did not reuse video assets: %+v", got)
	}
	if executor.videoSubmitCalls != 0 {
		t.Fatalf("merge retry called VIDEO provider %d times", executor.videoSubmitCalls)
	}
	completed, err := service.ExecuteAttempt(context.Background(), retried.Attempt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != MergeSucceeded || completed.OutputURL != "https://tos.example/merge.mp4" {
		t.Fatalf("completed retry = %+v", completed)
	}
	attempts, err := store.ListMergeAttempts(context.Background(), started.Job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(attempts) != 2 || attempts[0].Status != MergeFailed || attempts[1].Status != MergeSucceeded {
		t.Fatalf("merge history = %+v", attempts)
	}
}

func TestBuildFFmpegArgsUsesStructuredWhitelist(t *testing.T) {
	args, err := BuildFFmpegArgs([]string{"/tmp/a;touch-pwned.mp4", "/tmp/b.mp4"}, "/tmp/out.mp4", 1, "9:16")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-filter_complex") || !strings.Contains(joined, "720:1280") {
		t.Fatalf("args missing normalized filter: %v", args)
	}
	if strings.Contains(joined, "sh -c") || strings.Contains(joined, "bash -c") {
		t.Fatalf("args invoke shell: %v", args)
	}
	foundLiteral := false
	for i, arg := range args {
		if arg == "/tmp/a;touch-pwned.mp4" && i > 0 && args[i-1] == "-i" {
			foundLiteral = true
		}
	}
	if !foundLiteral {
		t.Fatalf("input path was not passed as a literal argv element: %v", args)
	}
	if _, err := BuildFFmpegArgs([]string{"/tmp/a.mp4"}, "/tmp/out.mp4", 9, "9:16"); err == nil {
		t.Fatal("expected speed whitelist rejection")
	}
	if _, err := BuildFFmpegArgs([]string{"/tmp/a.mp4"}, "/tmp/out.mp4", 1, "4:3"); err == nil {
		t.Fatal("expected aspect-ratio whitelist rejection")
	}
}

func TestFFmpegExecutorTimeoutCleanupAndMissingBinaryAreExplicit(t *testing.T) {
	tempRoot := t.TempDir()
	downloader := &fakeMediaDownloader{}
	uploader := &fakeFileArtifactStore{artifact: Artifact{Bucket: "videos", ObjectKey: "merge/out.mp4", URL: "https://tos.example/out.mp4"}}

	blocking := &fakeCommandRunner{run: func(ctx context.Context, _ string, _ ...string) ([]byte, error) {
		<-ctx.Done()
		return []byte("timed out stderr"), ctx.Err()
	}}
	executor := NewFFmpegExecutor(FFmpegExecutorConfig{Binary: "ffmpeg", Timeout: 5 * time.Millisecond, TempRoot: tempRoot, Runner: blocking, Downloader: downloader, Artifacts: uploader})
	_, err := executor.Execute(context.Background(), MergeExecutionRequest{JobID: 1, AttemptID: 1, Inputs: []MergeInputAsset{{ProductionTaskID: 1, URL: "https://cdn.example/a.mp4", Order: 1}}, AspectRatio: "9:16", Speed: 1})
	if !errors.Is(err, ErrMergeTimeout) {
		t.Fatalf("timeout err = %v", err)
	}
	entries, readErr := os.ReadDir(tempRoot)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("temp files were not cleaned: %v", entries)
	}

	missing := NewFFmpegExecutor(FFmpegExecutorConfig{Binary: filepath.Join(tempRoot, "missing-ffmpeg"), Timeout: time.Second, TempRoot: tempRoot, Downloader: downloader, Artifacts: uploader})
	_, err = missing.Execute(context.Background(), MergeExecutionRequest{JobID: 2, AttemptID: 1, Inputs: []MergeInputAsset{{ProductionTaskID: 1, URL: "https://cdn.example/a.mp4", Order: 1}}, AspectRatio: "9:16", Speed: 1})
	if !errors.Is(err, ErrFFmpegUnavailable) {
		t.Fatalf("missing ffmpeg err = %v", err)
	}
}

func TestFFmpegExecutorPersistsSuccessfulResultToTOSBoundary(t *testing.T) {
	tempRoot := t.TempDir()
	runner := &fakeCommandRunner{run: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		output := args[len(args)-1]
		if err := os.WriteFile(output, []byte("merged-video"), 0o600); err != nil {
			return nil, err
		}
		return nil, nil
	}}
	uploader := &fakeFileArtifactStore{artifact: Artifact{Bucket: "videos", ObjectKey: "merge/4/5.mp4", URL: "https://tos.example/final.mp4"}}
	executor := NewFFmpegExecutor(FFmpegExecutorConfig{Binary: "ffmpeg", Timeout: time.Second, TempRoot: tempRoot, Runner: runner, Downloader: &fakeMediaDownloader{}, Artifacts: uploader})
	artifact, err := executor.Execute(context.Background(), MergeExecutionRequest{JobID: 4, AttemptID: 5, Inputs: []MergeInputAsset{{ProductionTaskID: 10, URL: "https://cdn.example/10.mp4", Order: 1}}, AspectRatio: "16:9", Speed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.URL != "https://tos.example/final.mp4" || uploader.calls != 1 {
		t.Fatalf("artifact=%+v uploader calls=%d", artifact, uploader.calls)
	}
}
