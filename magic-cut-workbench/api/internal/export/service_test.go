package export

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"yizhanchengming/magic-cut-workbench/internal/material"
	"yizhanchengming/magic-cut-workbench/internal/timeline"
)

func TestBuildRenderPlanUsesTimelineOrderAndSafeArguments(t *testing.T) {
	items := []material.Material{
		{ID: "mat_b", StorageKey: filepath.Join(t.TempDir(), "b.mp4"), MediaType: "video"},
		{ID: "mat_a", StorageKey: filepath.Join(t.TempDir(), "a.mp4"), MediaType: "video"},
	}
	value := timeline.Timeline{DurationMs: 2000, Tracks: []timeline.Track{{ID: "video-1", Kind: "video", Clips: []timeline.Clip{
		{ID: "clip_b", MaterialID: "mat_b", StartMs: 1000, DurationMs: 1000, SourceStartMs: 0, Speed: 1},
		{ID: "clip_a", MaterialID: "mat_a", StartMs: 0, DurationMs: 1000, SourceStartMs: 250, Speed: 1},
	}}}}
	plan, err := BuildRenderPlan(value, items, CreateInput{Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "mp4"}, filepath.Join(t.TempDir(), "result.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	if plan.DurationMs != 2000 || len(plan.Inputs) != 2 {
		t.Fatalf("plan = %+v", plan)
	}
	if plan.Inputs[0].MaterialID != "mat_a" || plan.Inputs[1].MaterialID != "mat_b" {
		t.Fatalf("input order = %+v", plan.Inputs)
	}
	if plan.Command != "ffmpeg" || len(plan.Args) == 0 || plan.Args[len(plan.Args)-1] != plan.OutputPath {
		t.Fatalf("unsafe or incomplete argv plan: %+v", plan)
	}
}

func TestBuildRenderPlanRejectsMissingMaterialAndUnsupportedContainer(t *testing.T) {
	value := timeline.Timeline{Tracks: []timeline.Track{{ID: "video-1", Kind: "video", Clips: []timeline.Clip{{ID: "clip_1", MaterialID: "missing", DurationMs: 1000, Speed: 1}}}}}
	_, err := BuildRenderPlan(value, nil, CreateInput{Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "mp4"}, filepath.Join(t.TempDir(), "out.mp4"))
	if ErrorCode(err) != "EXPORT_MATERIAL_MISSING" {
		t.Fatalf("error = %v code=%q", err, ErrorCode(err))
	}
	_, err = BuildRenderPlan(value, []material.Material{{ID: "missing", StorageKey: "input.mp4", MediaType: "video"}}, CreateInput{Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "exe"}, filepath.Join(t.TempDir(), "out.exe"))
	if ErrorCode(err) != "EXPORT_FORMAT_NOT_SUPPORTED" {
		t.Fatalf("error = %v code=%q", err, ErrorCode(err))
	}
}

func TestFFmpegExportsPlayableOneSecondMP4(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe is not installed")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source.mp4")
	generate := exec.Command(ffmpegPath, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=blue:s=320x240:r=25:d=1", "-f", "lavfi", "-i", "anullsrc=channel_layout=stereo:sample_rate=48000", "-shortest", "-c:v", "libx264", "-c:a", "aac", "-pix_fmt", "yuv420p", "-y", source)
	if output, runErr := generate.CombinedOutput(); runErr != nil {
		t.Fatalf("generate fixture: %v: %s", runErr, output)
	}
	value := timeline.Timeline{DurationMs: 1000, Tracks: []timeline.Track{{ID: "video-1", Kind: "video", Clips: []timeline.Clip{{ID: "clip_1", MaterialID: "mat_1", DurationMs: 1000, Speed: 1}}}}}
	outputPath := filepath.Join(root, "export.mp4")
	plan, err := BuildRenderPlan(value, []material.Material{{ID: "mat_1", StorageKey: source, MediaType: "video"}}, CreateInput{Width: 640, Height: 360, FrameRate: 25, VideoCodec: "h264", AudioCodec: "aac", Container: "mp4"}, outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var latest float64
	result, err := (FFmpeg{}).Run(context.Background(), plan, func(progress float64) { latest = progress })
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := (material.FFprobe{}).Probe(context.Background(), outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if result.SizeBytes <= 0 || result.SHA256 == "" || latest != 1 {
		t.Fatalf("result=%+v progress=%v", result, latest)
	}
	if metadata.DurationMs < 900 || metadata.DurationMs > 1100 || metadata.Width != 640 || metadata.Height != 360 || metadata.VideoCodec != "h264" || metadata.AudioCodec != "aac" {
		t.Fatalf("metadata = %+v", metadata)
	}
}
