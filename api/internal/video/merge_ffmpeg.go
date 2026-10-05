package video

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type MediaDownloader interface {
	Download(context.Context, string, string) error
}

type FileArtifactStore interface {
	PersistFile(context.Context, string, string) (Artifact, error)
}

type FFmpegExecutorConfig struct {
	Binary     string
	Timeout    time.Duration
	TempRoot   string
	Runner     CommandRunner
	Downloader MediaDownloader
	Artifacts  FileArtifactStore
}

type FFmpegExecutor struct {
	binary     string
	timeout    time.Duration
	tempRoot   string
	runner     CommandRunner
	downloader MediaDownloader
	artifacts  FileArtifactStore
}

func NewFFmpegExecutor(config FFmpegExecutorConfig) *FFmpegExecutor {
	binary := strings.TrimSpace(config.Binary)
	if binary == "" {
		binary = "ffmpeg"
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}
	runner := config.Runner
	if runner == nil {
		runner = execCommandRunner{}
	}
	downloader := config.Downloader
	if downloader == nil {
		downloader = &HTTPMediaDownloader{Client: &http.Client{Timeout: 10 * time.Minute}}
	}
	return &FFmpegExecutor{
		binary: binary, timeout: timeout, tempRoot: strings.TrimSpace(config.TempRoot),
		runner: runner, downloader: downloader, artifacts: config.Artifacts,
	}
}

func (e *FFmpegExecutor) Execute(ctx context.Context, req MergeExecutionRequest) (Artifact, error) {
	if e == nil || e.runner == nil || e.downloader == nil || e.artifacts == nil {
		return Artifact{}, providerError(ErrorProviderUnavailable, "merge executor dependencies unavailable", nil)
	}
	inputs, aspect, speed, err := normalizeMergeRequest(req.Inputs, req.AspectRatio, req.Speed)
	if err != nil {
		return Artifact{}, err
	}
	root := e.tempRoot
	if root == "" {
		root = os.TempDir()
	}
	workDir, err := os.MkdirTemp(root, "task14-merge-*")
	if err != nil {
		return Artifact{}, fmt.Errorf("video: create merge temp directory: %w", err)
	}
	defer os.RemoveAll(workDir)

	inputPaths := make([]string, 0, len(inputs))
	for index, input := range inputs {
		path := filepath.Join(workDir, fmt.Sprintf("input-%03d.mp4", index+1))
		if err := e.downloader.Download(ctx, input.URL, path); err != nil {
			return Artifact{}, fmt.Errorf("video: download merge input %d: %w", index+1, err)
		}
		inputPaths = append(inputPaths, path)
	}
	outputPath := filepath.Join(workDir, "output.mp4")
	args, err := BuildFFmpegArgs(inputPaths, outputPath, speed, aspect)
	if err != nil {
		return Artifact{}, err
	}

	runCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	stderr, runErr := e.runner.Run(runCtx, e.binary, args...)
	if runErr != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) || errors.Is(runErr, context.DeadlineExceeded) {
			return Artifact{}, fmt.Errorf("%w: %s", ErrMergeTimeout, boundedMergeStderr(stderr))
		}
		var execErr *exec.Error
		if errors.As(runErr, &execErr) || errors.Is(runErr, exec.ErrNotFound) || errors.Is(runErr, os.ErrNotExist) {
			return Artifact{}, fmt.Errorf("%w: %s", ErrFFmpegUnavailable, e.binary)
		}
		message := boundedMergeStderr(stderr)
		if message == "" {
			message = runErr.Error()
		}
		return Artifact{}, fmt.Errorf("video: ffmpeg merge failed: %s", message)
	}
	if _, err := os.Stat(outputPath); err != nil {
		return Artifact{}, fmt.Errorf("video: ffmpeg did not create output: %w", err)
	}
	return e.artifacts.PersistFile(ctx, outputPath, fmt.Sprintf("merge/%d/%d.mp4", req.JobID, req.AttemptID))
}

func BuildFFmpegArgs(inputPaths []string, outputPath string, speed float64, aspectRatio string) ([]string, error) {
	if len(inputPaths) == 0 || strings.TrimSpace(outputPath) == "" {
		return nil, fmt.Errorf("video: ffmpeg inputs and output are required")
	}
	if speed == 0 {
		speed = 1
	}
	if speed < 0.5 || speed > 4 {
		return nil, fmt.Errorf("video: ffmpeg speed must be between 0.5 and 4")
	}
	width, height, err := mergeCanvas(aspectRatio)
	if err != nil {
		return nil, err
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	for _, input := range inputPaths {
		input = strings.TrimSpace(input)
		if input == "" {
			return nil, fmt.Errorf("video: ffmpeg input path is empty")
		}
		args = append(args, "-i", input)
	}
	filters := make([]string, 0, len(inputPaths)*2+1)
	concatInputs := make([]string, 0, len(inputPaths)*2)
	speedValue := strconv.FormatFloat(speed, 'f', -1, 64)
	for index := range inputPaths {
		videoFilter := fmt.Sprintf("[%d:v:0]settb=AVTB,setpts=PTS-STARTPTS,fps=30,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black,setsar=1", index, width, height, width, height)
		if speed != 1 {
			videoFilter += ",setpts=PTS/" + speedValue
		}
		filters = append(filters, videoFilter+fmt.Sprintf("[v%d]", index))
		audioFilter := fmt.Sprintf("[%d:a:0]aresample=48000,aformat=channel_layouts=stereo,asetpts=PTS-STARTPTS", index)
		if speed != 1 {
			audioFilter += ",atempo=" + atempoFilter(speed)
		}
		filters = append(filters, audioFilter+fmt.Sprintf("[a%d]", index))
		concatInputs = append(concatInputs, fmt.Sprintf("[v%d][a%d]", index, index))
	}
	filters = append(filters, strings.Join(concatInputs, "")+fmt.Sprintf("concat=n=%d:v=1:a=1[v][a]", len(inputPaths)))
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[v]", "-map", "[a]")
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart", outputPath,
	)
	return args, nil
}

func mergeCanvas(aspect string) (int, int, error) {
	switch strings.TrimSpace(aspect) {
	case "", "9:16":
		return 720, 1280, nil
	case "16:9":
		return 1280, 720, nil
	case "1:1":
		return 1080, 1080, nil
	default:
		return 0, 0, fmt.Errorf("video: unsupported merge aspect ratio %q", aspect)
	}
}

func atempoFilter(speed float64) string {
	parts := make([]string, 0, 3)
	for speed > 2 {
		parts = append(parts, "2.0")
		speed /= 2
	}
	for speed < 0.5 {
		parts = append(parts, "0.5")
		speed /= 0.5
	}
	parts = append(parts, strconv.FormatFloat(speed, 'f', -1, 64))
	return strings.Join(parts, ",atempo=")
}

func boundedMergeStderr(raw []byte) string {
	message := strings.TrimSpace(string(raw))
	if len(message) > 1024 {
		message = message[:1024]
	}
	return message
}

type execCommandRunner struct{}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

type HTTPMediaDownloader struct {
	Client *http.Client
}

func (d *HTTPMediaDownloader) Download(ctx context.Context, rawURL, destination string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("video: merge input URL must use https")
	}
	client := d.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("video: merge input returned HTTP %d", resp.StatusCode)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	const maxMergeInputBytes int64 = 2 << 30
	written, copyErr := io.Copy(file, io.LimitReader(resp.Body, maxMergeInputBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written > maxMergeInputBytes {
		return fmt.Errorf("video: merge input exceeds 2 GiB limit")
	}
	return nil
}
