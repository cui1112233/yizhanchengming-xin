package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

func TestAICutExportManagerTracksProgressAndCancellation(t *testing.T) {
	manager := newAICutExportManager()
	job, ctx := manager.create("alice", "cut-1", "mp4", time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC))
	if job.State != "queued" || job.Progress != 0 {
		t.Fatalf("job=%#v", job)
	}
	manager.update(job.ID, "running", 35, "", "", nil)
	if current, ok := manager.get("alice", "cut-1", job.ID); !ok || current.Progress != 35 {
		t.Fatalf("current=%#v ok=%v", current, ok)
	}
	canceled, ok := manager.cancel("alice", "cut-1", job.ID)
	if !ok || canceled.State != "cancel_requested" {
		t.Fatalf("canceled=%#v ok=%v", canceled, ok)
	}
	select {
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			t.Fatalf("ctx err=%v", ctx.Err())
		}
	default:
		t.Fatal("cancel did not stop render context")
	}
}

func TestBuildAICutFFmpegArgsRendersVideoClipsInTimelineOrder(t *testing.T) {
	clips := []aicutRenderClip{
		{clip: aicut.Clip{StartMS: 0, DurationMS: 2000, SourceStartMS: 500, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "first.mp4"},
		{clip: aicut.Clip{StartMS: 2000, DurationMS: 1000, PlaybackRate: 2}, asset: aicut.Asset{Kind: "image"}, path: "cover.png"},
	}
	args, err := buildAICutFFmpegArgs("9:16", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"-i first.mp4", "-loop 1 -t 1.000 -i cover.png", "trim=start=0.500:duration=2.000", "scale=1080:1920", "concat=n=2:v=1:a=0", "libx264", "output.mp4"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("args missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsAppliesVisualEffect(t *testing.T) {
	effect := &aicut.VisualEffect{Preset: "custom", Brightness: .2, Contrast: 1.1, Saturation: .8, Blur: 2, Grayscale: true}
	clips := []aicutRenderClip{{clip: aicut.Clip{DurationMS: 1000, PlaybackRate: 1, VisualEffect: effect}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"}}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"eq=brightness=0.200:contrast=1.100:saturation=0.800", "gblur=sigma=2.000", "hue=s=0"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("visual effect missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsMixesBGMIntoVideoExport(t *testing.T) {
	clips := []aicutRenderClip{
		{clip: aicut.Clip{StartMS: 0, DurationMS: 2000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"},
		{clip: aicut.Clip{StartMS: 0, DurationMS: 2000, Volume: 0.25, FadeInMS: 500, FadeOutMS: 500}, asset: aicut.Asset{Kind: "audio"}, path: "bgm.mp3"},
	}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"concat=n=1:v=1:a=0[outv]", "volume=0.250", "amix=inputs=1:duration=longest:normalize=0[outa]", "-map [outv] -map [outa]", "-c:a aac", "-shortest"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("args missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsRendersFlashTransition(t *testing.T) {
	clips := []aicutRenderClip{
		{clip: aicut.Clip{StartMS: 0, DurationMS: 2000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "first.mp4"},
		{clip: aicut.Clip{StartMS: 2000, DurationMS: 1500, PlaybackRate: 1, Transition: "flash-black", TransitionMS: 500}, asset: aicut.Asset{Kind: "video"}, path: "second.mp4"},
	}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "[v0][v1]xfade=transition=fadeblack:duration=0.500:offset=1.500[outv]") {
		t.Fatalf("flash transition was not rendered: %s", joined)
	}
}

func TestBuildAICutFFmpegArgsRendersFirstBatchTransitions(t *testing.T) {
	cases := map[string]string{
		"cross-dissolve": "dissolve", "fade": "fade", "flash-black": "fadeblack", "flash-white": "fadewhite",
		"slide-left": "slideleft", "slide-right": "slideright", "slide-up": "slideup", "slide-down": "slidedown", "zoom": "zoomin", "blur": "hblur",
		"transition-cross-dissolve": "dissolve", "transition-fade": "fade", "transition-flash-black": "fadeblack", "transition-flash-white": "fadewhite",
		"transition-push-left": "slideleft", "transition-push-right": "slideright", "transition-push-up": "slideup", "transition-push-down": "slidedown",
		"transition-wipe-left": "wipeleft", "transition-wipe-right": "wiperight", "transition-wipe-up": "wipeup", "transition-wipe-down": "wipedown",
		"transition-zoom": "zoomin", "transition-blur": "hblur", "transition-glitch": "pixelize", "transition-light-leak": "fadewhite",
	}
	for id, ffmpegName := range cases {
		clips := []aicutRenderClip{
			{clip: aicut.Clip{StartMS: 0, DurationMS: 2000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "first.mp4"},
			{clip: aicut.Clip{StartMS: 2000, DurationMS: 1500, PlaybackRate: 1, Transition: id, TransitionMS: 500}, asset: aicut.Asset{Kind: "video"}, path: "second.mp4"},
		}
		args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
		if err != nil {
			t.Fatalf("transition %s: %v", id, err)
		}
		if joined := strings.Join(args, " "); !strings.Contains(joined, "xfade=transition="+ffmpegName) {
			t.Fatalf("transition %s missing from %s", id, joined)
		}
	}
}

func TestBuildAICutCatalogFiltersMapsEveryStableID(t *testing.T) {
	ids := []string{
		"filter-original", "filter-bright", "filter-clear", "filter-natural", "filter-warm-sun", "filter-cool-blue", "filter-cinema", "filter-retro",
		"filter-monochrome", "filter-high-contrast", "filter-low-saturation", "filter-teal-orange", "filter-night", "filter-portrait", "filter-vivid", "filter-soft",
	}
	for _, id := range ids {
		filters, ok := buildAICutCatalogResourceFilters(id, map[string]float64{"intensity": .5})
		if !ok {
			t.Fatalf("catalog filter %s is not mapped", id)
		}
		if id != "filter-original" && len(filters) == 0 {
			t.Fatalf("catalog filter %s has no renderer", id)
		}
	}
}

func TestBuildAICutCatalogEffectsMapsEveryStableID(t *testing.T) {
	ids := []string{
		"effect-blur", "effect-vignette", "effect-glow", "effect-grain", "effect-shake", "effect-breathe-zoom",
		"effect-flash", "effect-glitch", "effect-chromatic", "effect-mosaic", "effect-fisheye", "effect-sharpen",
	}
	for _, id := range ids {
		filters, ok := buildAICutCatalogResourceFilters(id, map[string]float64{"intensity": .5, "speed": 1})
		if !ok || len(filters) == 0 {
			t.Fatalf("catalog effect %s is not renderable: %q", id, filters)
		}
	}
}

func TestBuildAICutFFmpegArgsAppliesCatalogFilterAndEffectTrack(t *testing.T) {
	clips := []aicutRenderClip{{
		clip:  aicut.Clip{DurationMS: 1000, PlaybackRate: 1, CatalogResourceID: "filter-warm-sun"},
		asset: aicut.Asset{Kind: "video"}, path: "scene.mp4",
		layerEffects: []aicut.Clip{{ResourceID: "effect-sharpen", Parameters: map[string]float64{"intensity": .5}}},
	}}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"colorbalance=", "unsharp="} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("catalog render chain missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutVisualEffectFiltersRendersAdvancedPresets(t *testing.T) {
	cases := map[string]string{"vignette": "vignette=", "glow": "gblur=", "shake": "crop=", "breathe": "zoompan=", "flash": "eq=brightness=", "mosaic": "pixelize="}
	for preset, expected := range cases {
		filters := strings.Join(buildAICutVisualEffectFilters(&aicut.VisualEffect{Preset: preset, Contrast: 1, Saturation: 1, Intensity: .5, Speed: 1}), ",")
		if !strings.Contains(filters, expected) {
			t.Fatalf("preset %s filters %q missing %q", preset, filters, expected)
		}
	}
}

func TestBuildAICutFFmpegArgsPreservesTimelineGapAndOverlappingLayers(t *testing.T) {
	clips := []aicutRenderClip{
		{clip: aicut.Clip{StartMS: 1000, DurationMS: 2000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "base.mp4"},
		{clip: aicut.Clip{StartMS: 1500, DurationMS: 1000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "image"}, path: "overlay.png"},
	}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{
		"color=c=black:s=1920x1080:d=3.000",
		"setpts=(PTS-STARTPTS)/1.000+1.000/TB",
		"overlay=eof_action=pass:enable='between(t,1.000,3.000)'",
		"overlay=eof_action=pass:enable='between(t,1.500,2.500)'",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("timeline composition missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutASSUsesTrackStyleTimingAndAnimation(t *testing.T) {
	timeline := aicut.Timeline{Tracks: []aicut.Track{{
		ID: "sub-1", Kind: aicut.TrackSubtitle,
		Settings: aicut.TrackSettings{SubtitleStyleID: "黄字大标题", SubtitleAnimation: "淡入上浮"},
		Clips:    []aicut.Clip{{ID: "s1", Kind: aicut.ClipSubtitle, StartMS: 1000, DurationMS: 2500, Text: "第一行\n第二行"}},
	}}}
	content := buildAICutASS(timeline, 1920, 1080)
	for _, expected := range []string{
		"PlayResX: 1920", "Style: Track0,Microsoft YaHei,72,&H0000FFFF", "Dialogue: 0,0:00:01.00,0:00:03.50,Track0", "{\\fad(200,200)}第一行\\N第二行",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("ASS missing %q:\n%s", expected, content)
		}
	}
}

func TestBuildAICutASSMapsChineseEditorTextResources(t *testing.T) {
	ids := []string{"text-basic-caption", "text-bold-title", "text-outline-caption", "text-shadow-caption", "text-panel-caption", "text-word-highlight", "text-news-bar", "text-lower-third", "text-comic", "text-handwriting", "text-vertical", "text-gradient-title"}
	for _, id := range ids {
		timeline := aicut.Timeline{Tracks: []aicut.Track{{
			ID: "sub-1", Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{TextResourceID: id},
			Clips: []aicut.Clip{{ID: "s1", Kind: aicut.ClipSubtitle, DurationMS: 1000, Text: "测试", TextResourceID: id}},
		}}}
		content := buildAICutASS(timeline, 1920, 1080)
		if !strings.Contains(content, "Dialogue:") || !strings.Contains(content, "测试") {
			t.Fatalf("text resource %s was not rendered: %s", id, content)
		}
	}
}

func TestBuildAICutFFmpegArgsAppliesCanvasTransform(t *testing.T) {
	clips := []aicutRenderClip{{clip: aicut.Clip{
		DurationMS: 1000, PlaybackRate: 1,
		CanvasTransform: &aicut.CanvasTransform{X: .1, Y: -.1, ScaleX: .8, ScaleY: .8, Rotation: 12, Opacity: .7},
	}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"}}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"rotate=", "colorchannelmixer=aa=0.700", "pad=1920:1080"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("canvas transform missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsMixesAudioAtClipStart(t *testing.T) {
	clips := []aicutRenderClip{
		{clip: aicut.Clip{StartMS: 0, DurationMS: 2000}, asset: aicut.Asset{Kind: "audio"}, path: "voice.wav"},
		{clip: aicut.Clip{StartMS: 2500, DurationMS: 1000, Volume: 0.25, FadeInMS: 200, FadeOutMS: 300}, asset: aicut.Asset{Kind: "audio"}, path: "music.mp3"},
	}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp3", "output.mp3")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"adelay=2500|2500", "volume=0.250", "afade=t=in:st=0:d=0.200", "afade=t=out:st=0.700:d=0.300", "amix=inputs=2:duration=longest", "libmp3lame", "output.mp3"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("args missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsRejectsEmptyTimeline(t *testing.T) {
	if _, err := buildAICutFFmpegArgs("9:16", nil, "mp4", "output.mp4"); err == nil || !strings.Contains(err.Error(), "没有可导出") {
		t.Fatalf("expected clear empty error, got %v", err)
	}
}

func TestAICutFFmpegArgsProducePlayableMP4(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.mp4"), filepath.Join(dir, "output.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-y", "-f", "lavfi", "-i", "color=c=red:s=64x64:d=0.4", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 400, SourceDurationMS: 400, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: input}}, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(output)
	if err != nil || info.Size() == 0 {
		t.Fatalf("rendered file invalid: info=%v err=%v", info, err)
	}
}

func TestAICutCatalogResourceFiltersExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	ids := []string{
		"filter-bright", "filter-clear", "filter-natural", "filter-warm-sun", "filter-cool-blue", "filter-cinema", "filter-retro",
		"filter-monochrome", "filter-high-contrast", "filter-low-saturation", "filter-teal-orange", "filter-night", "filter-portrait", "filter-vivid", "filter-soft",
		"effect-blur", "effect-vignette", "effect-glow", "effect-grain", "effect-shake", "effect-breathe-zoom", "effect-flash",
		"effect-glitch", "effect-chromatic", "effect-mosaic", "effect-fisheye", "effect-sharpen",
	}
	for _, id := range ids {
		filters, ok := buildAICutCatalogResourceFilters(id, map[string]float64{"intensity": .5, "speed": 1})
		if !ok {
			t.Fatalf("%s is not mapped", id)
		}
		err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=size=64x64:duration=0.15", "-vf", strings.Join(filters, ","), "-frames:v", "2", "-f", "null", "-")
		if err != nil {
			t.Fatalf("%s filter chain failed: %v (%s)", id, err, strings.Join(filters, ","))
		}
	}
}

func TestAICutCanvasTransformFiltersExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	filters := buildAICutCanvasTransformFilters(64, 64, &aicut.CanvasTransform{X: .1, Y: -.1, ScaleX: .8, ScaleY: .8, Rotation: 12, Opacity: .7})
	err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=size=64x64:duration=0.15", "-vf", strings.Join(filters, ","), "-frames:v", "2", "-f", "null", "-")
	if err != nil {
		t.Fatalf("canvas transform chain failed: %v (%s)", err, strings.Join(filters, ","))
	}
}
