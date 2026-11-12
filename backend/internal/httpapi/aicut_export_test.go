package httpapi

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

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

func TestNormalizeAICutExportOptionsValidatesProfessionalSettings(t *testing.T) {
	defaults, err := normalizeAICutExportOptions("mp4", aicutExportOptions{})
	if err != nil || defaults.Resolution != "1080p" || defaults.FrameRate != 30 || defaults.Quality != "standard" || defaults.Codec != "h264" || defaults.BitDepth != 8 || defaults.BitrateMbps != 0 || defaults.HardwareAcceleration != "auto" {
		t.Fatalf("defaults=%#v err=%v", defaults, err)
	}
	valid, err := normalizeAICutExportOptions("mp4", aicutExportOptions{Resolution: "4k", FrameRate: 60, Quality: "high", Codec: "hevc", BitDepth: 10, BitrateMbps: 35, HardwareAcceleration: "software"})
	if err != nil || valid.Resolution != "4k" || valid.FrameRate != 60 || valid.Quality != "high" || valid.Codec != "hevc" || valid.BitDepth != 10 || valid.BitrateMbps != 35 || valid.HardwareAcceleration != "software" {
		t.Fatalf("valid=%#v err=%v", valid, err)
	}
	for _, invalid := range []aicutExportOptions{
		{Resolution: "8k", FrameRate: 30, Quality: "standard"},
		{Resolution: "1080p", FrameRate: 29, Quality: "standard"},
		{Resolution: "1080p", FrameRate: 30, Quality: "lossless"},
		{Resolution: "1080p", FrameRate: 30, Quality: "high", Codec: "vp9"},
		{Resolution: "1080p", FrameRate: 30, Quality: "high", Codec: "h264", BitDepth: 10},
		{Resolution: "1080p", FrameRate: 30, Quality: "high", Codec: "hevc", BitDepth: 12},
		{Resolution: "1080p", FrameRate: 30, Quality: "high", Codec: "hevc", BitDepth: 10, BitrateMbps: 101},
		{Resolution: "1080p", FrameRate: 30, Quality: "high", Codec: "h264", HardwareAcceleration: "gpu-only"},
	} {
		if _, err := normalizeAICutExportOptions("mp4", invalid); err == nil {
			t.Fatalf("accepted invalid options: %#v", invalid)
		}
	}
	webm, err := normalizeAICutExportOptions("webm", aicutExportOptions{Resolution: "4k", FrameRate: 60, Quality: "high", Codec: "av1", BitDepth: 10, BitrateMbps: 24})
	if err != nil || webm.Codec != "av1" || webm.BitDepth != 10 {
		t.Fatalf("webm=%#v err=%v", webm, err)
	}
	mp3, err := normalizeAICutExportOptions("mp3", aicutExportOptions{Resolution: "4k", FrameRate: 29, Quality: "lossless"})
	if err != nil || mp3 != (aicutExportOptions{}) {
		t.Fatalf("mp3 options=%#v err=%v", mp3, err)
	}
}

func TestAICutAutomaticHardwareEncoderSelectionAndSoftwareFallback(t *testing.T) {
	available := map[string]bool{"h264_nvenc": true, "h264_qsv": true, "hevc_qsv": true}
	if got := selectAICutHardwareEncoder("h264", 8, available); got != "h264_nvenc" {
		t.Fatalf("expected nvenc first, got %q", got)
	}
	if got := selectAICutHardwareEncoder("hevc", 10, available); got != "hevc_qsv" {
		t.Fatalf("expected qsv fallback, got %q", got)
	}
	if got := selectAICutHardwareEncoder("vp9", 8, available); got != "" {
		t.Fatalf("vp9 must keep software encoder, got %q", got)
	}

	clips := []aicutRenderClip{{clip: aicut.Clip{DurationMS: 1000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"}}
	hardwareArgs, err := buildAICutFFmpegArgsWithOptions("16:9", clips, "mp4", "output.mp4", aicutExportOptions{Codec: "h264", HardwareAcceleration: "auto", ResolvedVideoEncoder: "h264_nvenc"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if joined := strings.Join(hardwareArgs, " "); !strings.Contains(joined, "-c:v h264_nvenc") || strings.Contains(joined, "-c:v libx264") {
		t.Fatalf("hardware args did not select nvenc: %s", joined)
	}
	softwareArgs, err := buildAICutFFmpegArgsWithOptions("16:9", clips, "mp4", "output.mp4", aicutExportOptions{Codec: "h264", HardwareAcceleration: "software"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if joined := strings.Join(softwareArgs, " "); !strings.Contains(joined, "-c:v libx264") {
		t.Fatalf("software fallback missing libx264: %s", joined)
	}
}

func TestAICutExportOptionsSupportTransparentVP9AndProRes4444(t *testing.T) {
	webm, err := normalizeAICutExportOptions("webm", aicutExportOptions{Codec: "vp9", Alpha: true})
	if err != nil || !webm.Alpha {
		t.Fatalf("webm=%+v err=%v", webm, err)
	}
	mov, err := normalizeAICutExportOptions("mov", aicutExportOptions{Codec: "prores4444", BitDepth: 12, Alpha: true})
	if err != nil || mov.Codec != "prores4444" || !mov.Alpha {
		t.Fatalf("mov=%+v err=%v", mov, err)
	}
	if _, err := normalizeAICutExportOptions("mp4", aicutExportOptions{Codec: "h264", Alpha: true}); err == nil || !strings.Contains(err.Error(), "透明") {
		t.Fatalf("mp4 alpha err=%v", err)
	}
	if _, err := normalizeAICutExportOptions("mov", aicutExportOptions{Codec: "h264", Alpha: true}); err == nil || !strings.Contains(err.Error(), "ProRes 4444") {
		t.Fatalf("mov codec err=%v", err)
	}
}

func TestAICutTransparentExportBuildsAlphaPreservingFilterAndCodec(t *testing.T) {
	clips := []aicutRenderClip{{
		clip:  aicut.Clip{ID: "image-1", Kind: aicut.ClipImage, Name: "透明人物", DurationMS: 1000, SourceDurationMS: 1000},
		asset: aicut.Asset{ID: "asset-1", Kind: "image", Name: "person.png"},
		path:  "person.png",
	}}
	webm, err := buildAICutFFmpegArgsWithOptions("16:9", clips, "webm", "output.webm", aicutExportOptions{Resolution: "720p", FrameRate: 30, Quality: "standard", Codec: "vp9", BitDepth: 8, Alpha: true}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(webm, " ")
	for _, want := range []string{"libvpx-vp9", "yuva420p", "-auto-alt-ref 0", "color=c=black@0.0"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("webm args missing %q: %s", want, joined)
		}
	}
	mov, err := buildAICutFFmpegArgsWithOptions("16:9", clips, "mov", "output.mov", aicutExportOptions{Resolution: "720p", FrameRate: 30, Quality: "standard", Codec: "prores4444", BitDepth: 12, Alpha: true}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined = strings.Join(mov, " ")
	for _, want := range []string{"prores_ks", "-profile:v 4", "yuva444p12le"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("mov args missing %q: %s", want, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsUsesProfessionalExportSettings(t *testing.T) {
	clips := []aicutRenderClip{{clip: aicut.Clip{DurationMS: 1000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"}}
	args, err := buildAICutFFmpegArgsWithOptions("16:9", clips, "mp4", "output.mp4", aicutExportOptions{Resolution: "720p", FrameRate: 25, Quality: "high"}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"scale=1280:720", "fps=25", "-crf 18"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("professional export missing %q: %s", expected, joined)
		}
	}
	webm, err := buildAICutFFmpegArgsWithOptions("9:16", clips, "webm", "output.webm", aicutExportOptions{Resolution: "720p", FrameRate: 60, Quality: "draft"}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined = strings.Join(webm, " ")
	for _, expected := range []string{"scale=720:1280", "fps=60", "-crf 38"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("professional webm export missing %q: %s", expected, joined)
		}
	}
	hevc, err := buildAICutFFmpegArgsWithOptions("16:9", clips, "mp4", "output.mp4", aicutExportOptions{Resolution: "4k", FrameRate: 60, Quality: "high", Codec: "hevc", BitDepth: 10, BitrateMbps: 35}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined = strings.Join(hevc, " ")
	for _, expected := range []string{"scale=3840:2160", "fps=60", "-c:v libx265", "-pix_fmt yuv420p10le", "-b:v 35M", "-maxrate 35M", "-bufsize 70M"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("professional HEVC export missing %q: %s", expected, joined)
		}
	}
	av1, err := buildAICutFFmpegArgsWithOptions("9:16", clips, "webm", "output.webm", aicutExportOptions{Resolution: "1080p", FrameRate: 30, Quality: "standard", Codec: "av1", BitDepth: 10}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined = strings.Join(av1, " ")
	for _, expected := range []string{"-c:v libaom-av1", "-pix_fmt yuv420p10le", "-cpu-used 6"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("professional AV1 export missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsRendersSpeedCurveAsRealSegments(t *testing.T) {
	clips := []aicutRenderClip{{clip: aicut.Clip{
		DurationMS: 1600, SourceStartMS: 500, PlaybackRate: 1.375,
		SpeedCurve: []aicut.SpeedCurvePoint{{Position: 0, Rate: .5}, {Position: .5, Rate: 1.5}, {Position: 1, Rate: 2}},
	}, asset: aicut.Asset{Kind: "video"}, path: "curve.mp4"}}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{
		"split=2", "trim=start=0.500000:duration=0.800000", "sqrt(0.250000000",
		"trim=start=1.300000:duration=1.400000", "sqrt(2.250000000", "concat=n=2:v=1:a=0",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("speed curve export missing %q: %s", expected, joined)
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

func TestBuildAICutFFmpegArgsAppliesGeneratedPersonMatte(t *testing.T) {
	clip := aicut.Clip{ID: "person", AssetID: "video-1", Kind: aicut.ClipVideo, DurationMS: 1200, SourceStartMS: 200, SourceDurationMS: 5000, PlaybackRate: 1}
	clip.VisualProcessing = &aicut.VisualProcessing{PersonSegmentation: aicut.PersonSegmentation{
		Enabled: true, MatteAssetID: "mask-1", Threshold: .55, Feather: .08, Signature: aicutPersonSegmentationSignature(clip),
	}}
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: clip, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4", mattePath: "person-mask.webm"}}, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"-i scene.mp4 -i person-mask.webm", "[1:v]trim=start=0:duration=1.200", "format=gray", "boxblur", "lut", "alphamerge"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("person matte export missing %q: %s", expected, joined)
		}
	}
}

func TestBuildAICutFFmpegArgsRejectsStalePersonMatte(t *testing.T) {
	clip := aicut.Clip{ID: "person", AssetID: "video-1", Kind: aicut.ClipVideo, DurationMS: 1200, SourceDurationMS: 5000, PlaybackRate: 1,
		VisualProcessing: &aicut.VisualProcessing{PersonSegmentation: aicut.PersonSegmentation{Enabled: true, MatteAssetID: "mask-1", Threshold: .5, Feather: .1, Signature: "old"}},
	}
	_, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: clip, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4", mattePath: "person-mask.webm"}}, "mp4", "output.mp4")
	if err == nil || !strings.Contains(err.Error(), "人物蒙版已过期") {
		t.Fatalf("stale matte err=%v", err)
	}
}

func TestBuildAICutFFmpegArgsComposesBlendModesWithAlphaMask(t *testing.T) {
	clips := []aicutRenderClip{
		{clip: aicut.Clip{ID: "base", Kind: aicut.ClipVideo, StartMS: 0, DurationMS: 1000, SourceDurationMS: 1000}, asset: aicut.Asset{Kind: "video"}, path: "base.mp4"},
		{clip: aicut.Clip{ID: "top", Kind: aicut.ClipImage, StartMS: 0, DurationMS: 1000, SourceDurationMS: 1000, BlendMode: "screen"}, asset: aicut.Asset{Kind: "image"}, path: "top.png"},
	}
	args, err := buildAICutFFmpegArgs("1:1", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"split=2", "alphaextract", "blend=all_mode=screen", "alphamerge", "overlay=eof_action=pass"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("blend export missing %q: %s", expected, joined)
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
		"filter-golden-hour", "filter-noir", "filter-vintage", "filter-cyberpunk", "filter-vaporwave", "filter-mint-haze", "filter-bleach-bypass", "filter-winter-cool", "filter-rosy", "filter-film", "filter-super-8", "filter-instant-print",
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
		"color=c=black:s=1920x1080:r=30:d=3.000",
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

func TestBuildAICutASSUsesRichTextStyleAndRealKaraokeTiming(t *testing.T) {
	timeline := aicut.Timeline{Tracks: []aicut.Track{{ID: "text", Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{
		FontFamily: "KaiTi", FontSize: 72, TextColor: "#ffcc00", StrokeColor: "#112233", StrokeWidth: 3,
		BackgroundColor: "#00000080", ShadowColor: "#223344cc", ShadowOffsetX: 4, ShadowOffsetY: 6, ShadowBlur: 8,
		LetterSpacing: 4, LineHeight: 1.4, TextAlign: "left", SubtitleAnimation: "word-highlight",
	}, Clips: []aicut.Clip{{ID: "c", Kind: aicut.ClipSubtitle, DurationMS: 1200, Text: "剪辑开始"}}}}}
	content := buildAICutASS(timeline, 1920, 1080)
	for _, expected := range []string{"Style: Track0,KaiTi,72", "&H0000CCFF", "&H00332211", ",3,", "\\4c&H33443322&", "\\xshad4.00", "\\yshad6.00", "\\blur8.00", "\\fsp4", "{\\k30}剪", "{\\k30}辑"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("ASS missing %q:\n%s", expected, content)
		}
	}
}

func TestBuildAICutASSRendersCurvedTextAsPositionedRotatedGlyphs(t *testing.T) {
	timeline := aicut.Timeline{Tracks: []aicut.Track{{ID: "curve", Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{
		FontFamily: "KaiTi", FontSize: 60, TextCurve: 120, LetterSpacing: 2,
	}, Clips: []aicut.Clip{{ID: "c", Kind: aicut.ClipSubtitle, DurationMS: 1200, Text: "曲线文字"}}}}}
	content := buildAICutASS(timeline, 1920, 1080)
	if strings.Count(content, "Dialogue:") != 4 || !strings.Contains(content, "\\an5\\pos(") || !strings.Contains(content, "\\frz-") || !strings.Contains(content, "\\frz") {
		t.Fatalf("curved glyph events missing:\n%s", content)
	}
}

func TestBuildAICutASSRendersMixedTextRuns(t *testing.T) {
	timeline := aicut.Timeline{Tracks: []aicut.Track{{ID: "mixed", Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{
		FontSize: 60, TextColor: "#ffffff", TextRuns: []aicut.TextStyleRun{{Start: 1, End: 3, TextColor: "#ff0000", FontSize: 80, Bold: true, Italic: true}},
	}, Clips: []aicut.Clip{{ID: "c", Kind: aicut.ClipSubtitle, DurationMS: 1200, Text: "混排文字"}}}}}
	content := buildAICutASS(timeline, 1920, 1080)
	for _, expected := range []string{"{\\c&H000000FF&\\fs80\\b1\\i1}排", "{\\c&H000000FF&\\fs80\\b1\\i1}文", "{\\c&H00FFFFFF&\\fs60\\b0\\i0}混"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("mixed run missing %q:\n%s", expected, content)
		}
	}
}

func TestBuildAICutFFmpegArgsLoadsProjectFontDirectory(t *testing.T) {
	dir := t.TempDir()
	fontsDir := filepath.Join(dir, "fonts")
	if err := os.MkdirAll(fontsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	subtitlePath := filepath.Join(dir, "subtitles.ass")
	if err := os.WriteFile(subtitlePath, []byte("[Script Info]"), 0o600); err != nil {
		t.Fatal(err)
	}
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 1000, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"}}, "mp4", "output.mp4", subtitlePath)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "fontsdir='") || !strings.Contains(joined, escapeAICutFilterPath(fontsDir)) {
		t.Fatalf("custom font directory missing: %s", joined)
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
		expectedText := "测试"
		if id == "text-gradient-title" {
			expectedText = aicutGradientASSText("测试")
		}
		if id == "text-vertical" {
			expectedText = "测\\N试"
		}
		if !strings.Contains(content, "Dialogue:") || !strings.Contains(content, expectedText) {
			t.Fatalf("text resource %s was not rendered: %s", id, content)
		}
	}
}

func TestBuildAICutASSRendersVerticalAndNewsTextTemplates(t *testing.T) {
	timeline := aicut.Timeline{Tracks: []aicut.Track{
		{ID: "vertical", Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{TextResourceID: "text-vertical"}, Clips: []aicut.Clip{{ID: "v", Kind: aicut.ClipSubtitle, DurationMS: 1000, Text: "竖排"}}},
		{ID: "news", Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{TextResourceID: "text-news-bar"}, Clips: []aicut.Clip{{ID: "n", Kind: aicut.ClipSubtitle, DurationMS: 1000, Text: "新闻"}}},
	}}
	content := buildAICutASS(timeline, 1920, 1080)
	for _, expected := range []string{"竖\\N排", "Style: Track1", "&H22EF5E15", ",3,0,0,1,40,40,70,1"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("template output missing %q:\n%s", expected, content)
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

func TestAICutCanvasTransformCanRenderFullyTransparent(t *testing.T) {
	filters := strings.Join(buildAICutCanvasTransformFilters(64, 64, &aicut.CanvasTransform{ScaleX: 1, ScaleY: 1, Opacity: 0}), ",")
	if !strings.Contains(filters, "colorchannelmixer=aa=0.000") {
		t.Fatalf("zero opacity must not fall back to opaque: %s", filters)
	}
}

func TestBuildAICutFFmpegArgsRendersVisualAndAudioKeyframes(t *testing.T) {
	x0, x1 := 0.0, 1.0
	s0, s1 := 1.0, 1.5
	r0, r1 := 0.0, 30.0
	o0, o1 := 1.0, 0.4
	v0, v1 := 0.2, 1.2
	clips := []aicutRenderClip{
		{clip: aicut.Clip{DurationMS: 2000, PlaybackRate: 1, Keyframes: []aicut.ClipKeyframe{
			{ID: "v0", OffsetMS: 0, Values: aicut.KeyframeValues{X: &x0, ScaleX: &s0, ScaleY: &s0, Rotation: &r0, Opacity: &o0}, Easing: &aicut.KeyframeEasing{Type: "ease-in", X1: .42, X2: 1, Y2: 1}},
			{ID: "v1", OffsetMS: 1000, Values: aicut.KeyframeValues{X: &x1, ScaleX: &s1, ScaleY: &s1, Rotation: &r1, Opacity: &o1}},
		}}, asset: aicut.Asset{Kind: "video"}, path: "scene.mp4"},
		{clip: aicut.Clip{DurationMS: 2000, Volume: 1, Keyframes: []aicut.ClipKeyframe{
			{ID: "a0", OffsetMS: 0, Values: aicut.KeyframeValues{Volume: &v0}},
			{ID: "a1", OffsetMS: 1000, Values: aicut.KeyframeValues{Volume: &v1}},
		}}, asset: aicut.Asset{Kind: "audio"}, path: "voice.wav"},
	}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, expected := range []string{"scale=w=", "eval=frame", "rotate=", "geq=", "volume='if(", "0.200000", "1.200000", "0.163555"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("keyframe export missing %q: %s", expected, joined)
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

func TestBuildAICutFFmpegArgsAudioProcessing(t *testing.T) {
	output := filepath.Join(t.TempDir(), "processed.mp3")
	clips := []aicutRenderClip{{path: "voice.wav", asset: aicut.Asset{Kind: "audio"}, clip: aicut.Clip{
		ID: "audio-1", Kind: aicut.ClipAudio, DurationMS: 2000, SourceDurationMS: 2000,
		AudioProcessing: &aicut.AudioProcessing{Denoise: .5, EQLowDB: 3, EQMidDB: -2, EQHighDB: 4, ChannelMode: "right", LoudnessEnabled: true, TargetLUFS: -14},
	}}}
	args, err := buildAICutFFmpegArgs("16:9", clips, "mp3", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"afftdn=", "equalizer=f=120", "equalizer=f=1000", "equalizer=f=8000", "pan=stereo|c0=c1|c1=c1", "loudnorm=I=-14.0:TP=-1.5:LRA=11"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q: %s", want, joined)
		}
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

func TestAICutPersonMatteProducesPlayableMP4(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, matte, output := filepath.Join(dir, "input.mp4"), filepath.Join(dir, "matte.mp4"), filepath.Join(dir, "output.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=red:s=64x64:d=0.4", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=black:s=64x64:d=0.4,drawbox=x=0:y=0:w=32:h=64:color=white:t=fill", "-c:v", "libx264", "-pix_fmt", "yuv420p", matte); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "person", AssetID: "video-1", Kind: aicut.ClipVideo, DurationMS: 400, SourceDurationMS: 400, PlaybackRate: 1}
	clip.VisualProcessing = &aicut.VisualProcessing{PersonSegmentation: aicut.PersonSegmentation{Enabled: true, MatteAssetID: "mask-1", Threshold: .5, Feather: .08, Signature: aicutPersonSegmentationSignature(clip)}}
	args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{clip: clip, asset: aicut.Asset{Kind: "video"}, path: input, mattePath: matte}}, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(output)
	if err != nil || info.Size() == 0 {
		t.Fatalf("person matte render invalid: info=%v err=%v", info, err)
	}
}

func TestAICutProfessionalCodecsProduceTenBitVideo(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	if _, err := ffmpegbin.ResolveFFprobe(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "input.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=blue:s=64x64:d=0.2", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, format, codec, extension, expectedCodec string
	}{
		{name: "HEVC", format: "mp4", codec: "hevc", extension: "mp4", expectedCodec: "hevc,yuv420p10le"},
		{name: "AV1", format: "webm", codec: "av1", extension: "webm", expectedCodec: "av1,yuv420p10le"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(dir, "output-"+test.codec+"."+test.extension)
			args, err := buildAICutFFmpegArgsWithOptions("1:1", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 200, SourceDurationMS: 200, PlaybackRate: 1}, asset: aicut.Asset{Kind: "video"}, path: input}}, test.format, output, aicutExportOptions{Resolution: "720p", FrameRate: 24, Quality: "draft", Codec: test.codec, BitDepth: 10}, "")
			if err != nil {
				t.Fatal(err)
			}
			if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
				t.Fatal(err)
			}
			command, _, err := ffmpegbin.Command(t.Context(), "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name,pix_fmt", "-of", "csv=p=0", output)
			if err != nil {
				t.Fatal(err)
			}
			result, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("ffprobe: %v: %s", err, result)
			}
			if strings.TrimSpace(string(result)) != test.expectedCodec {
				t.Fatalf("codec probe=%q want=%q", strings.TrimSpace(string(result)), test.expectedCodec)
			}
		})
	}
}

func TestAICutProRes4444ExportProducesAlphaVideo(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	if _, err := ffmpegbin.ResolveFFprobe(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "transparent.png"), filepath.Join(dir, "alpha.mov")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=red@0.45:s=64x64:d=0.1,format=rgba", "-frames:v", "1", input); err != nil {
		t.Fatal(err)
	}
	args, err := buildAICutFFmpegArgsWithOptions("1:1", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 200, SourceDurationMS: 200, PlaybackRate: 1}, asset: aicut.Asset{Kind: "image"}, path: input}}, "mov", output, aicutExportOptions{Resolution: "720p", FrameRate: 24, Quality: "draft", Codec: "prores4444", BitDepth: 12, Alpha: true}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	command, _, err := ffmpegbin.Command(t.Context(), "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name,pix_fmt", "-of", "csv=p=0", output)
	if err != nil {
		t.Fatal(err)
	}
	result, err := command.Output()
	if err != nil || !strings.Contains(string(result), "prores,yuva444p12le") {
		t.Fatalf("ffprobe=%q err=%v", result, err)
	}
}

func TestAICutCatalogResourceFiltersExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	ids := []string{
		"filter-bright", "filter-clear", "filter-natural", "filter-warm-sun", "filter-cool-blue", "filter-cinema", "filter-retro",
		"filter-monochrome", "filter-high-contrast", "filter-low-saturation", "filter-teal-orange", "filter-night", "filter-portrait", "filter-vivid", "filter-soft",
		"filter-golden-hour", "filter-noir", "filter-vintage", "filter-cyberpunk", "filter-vaporwave", "filter-mint-haze", "filter-bleach-bypass", "filter-winter-cool", "filter-rosy", "filter-film", "filter-super-8", "filter-instant-print",
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

func TestAICutAnimatedCanvasTransformFiltersExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	x0, x1 := 0.0, 0.25
	s0, s1 := 1.0, 1.2
	r0, r1 := 0.0, 15.0
	o0, o1 := 1.0, 0.5
	clip := aicut.Clip{DurationMS: 200, Keyframes: []aicut.ClipKeyframe{
		{ID: "k0", OffsetMS: 0, Values: aicut.KeyframeValues{X: &x0, ScaleX: &s0, ScaleY: &s0, Rotation: &r0, Opacity: &o0}, Easing: &aicut.KeyframeEasing{Type: "custom", X1: .15, Y1: .1, X2: .75, Y2: .95}},
		{ID: "k1", OffsetMS: 100, Values: aicut.KeyframeValues{X: &x1, ScaleX: &s1, ScaleY: &s1, Rotation: &r1, Opacity: &o1}},
	}}
	filters := buildAICutAnimatedCanvasTransformFilters(64, 64, clip, false)
	dir := t.TempDir()
	pattern := filepath.Join(dir, "frame-%02d.png")
	err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=red:size=64x64:duration=0.2", "-vf", strings.Join(filters, ","), "-frames:v", "4", pattern)
	if err != nil {
		t.Fatalf("animated canvas transform failed: %v (%s)", err, strings.Join(filters, ","))
	}
	first, firstErr := os.ReadFile(filepath.Join(dir, "frame-01.png"))
	last, lastErr := os.ReadFile(filepath.Join(dir, "frame-04.png"))
	if firstErr != nil || lastErr != nil || bytes.Equal(first, last) {
		t.Fatalf("keyframes did not change rendered frames: first=%v last=%v equal=%v", firstErr, lastErr, bytes.Equal(first, last))
	}
}

func TestAICutAudioProcessingFiltersExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "stereo.wav"), filepath.Join(dir, "processed.mp3")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "aevalsrc=sin(440*2*PI*t)|sin(660*2*PI*t):s=44100:d=0.3", input); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "audio", Kind: aicut.ClipAudio, DurationMS: 300, SourceDurationMS: 300, FadeInMS: 50, FadeOutMS: 60, AudioProcessing: &aicut.AudioProcessing{Denoise: .5, EQLowDB: 3, EQMidDB: -2, EQHighDB: 4, ChannelMode: "swap", LoudnessEnabled: true, TargetLUFS: -16, VoiceEffect: "robot"}}
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "audio"}, clip: clip}}, "mp3", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"afade=t=in:st=0:d=0.050", "afade=t=out:st=0.240:d=0.060", "chorus=", "aecho="} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q: %s", want, joined)
		}
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(output)
	if err != nil || info.Size() == 0 {
		t.Fatalf("processed export invalid: info=%v err=%v", info, err)
	}
}

func TestAICutVoiceEffectsBuildDistinctExecutableFFmpegChains(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "voice.wav")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=0.5", input); err != nil {
		t.Fatal(err)
	}
	seenChains := map[string]string{}
	for _, effect := range []string{"natural", "enhance", "bright", "bass", "treble", "telephone", "radio", "megaphone", "robot", "chipmunk", "deep", "sweet", "whisper", "distorted", "lofi", "echo", "room", "hall", "plate", "cathedral", "stadium", "underwater", "vinyl"} {
		filters := buildAICutVoiceEffectFilters(effect)
		if effect != "natural" && len(filters) == 0 {
			t.Fatalf("%s has no server render chain", effect)
		}
		chain := strings.Join(filters, ",")
		if previous, duplicated := seenChains[chain]; effect != "natural" && duplicated {
			t.Fatalf("%s duplicates %s render chain: %s", effect, previous, chain)
		}
		seenChains[chain] = effect
		output := filepath.Join(dir, effect+".mp3")
		clip := aicut.Clip{ID: effect, Kind: aicut.ClipAudio, DurationMS: 500, SourceDurationMS: 500, Volume: 1, AudioProcessing: &aicut.AudioProcessing{VoiceEffect: effect}}
		args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "audio"}, clip: clip}}, "mp3", output)
		if err != nil {
			t.Fatalf("%s args: %v", effect, err)
		}
		if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
			t.Fatalf("%s render: %v (%s)", effect, err, strings.Join(args, " "))
		}
		if info, err := os.Stat(output); err != nil || info.Size() == 0 {
			t.Fatalf("%s output invalid: info=%v err=%v", effect, info, err)
		}
	}
}

func TestAICutPitchSemitonesBuildsExecutableDurationPreservingChain(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "voice.wav"), filepath.Join(dir, "pitched.mp3")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=0.5", input); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "pitch", Kind: aicut.ClipAudio, DurationMS: 500, SourceDurationMS: 500, Volume: 1, AudioProcessing: &aicut.AudioProcessing{PitchSemitones: 7}}
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "audio"}, clip: clip}}, "mp3", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"asetrate=44100*1.498307", "aresample=44100", "atempo=0.667420"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("pitch chain missing %q: %s", want, joined)
		}
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("pitched export invalid: info=%v err=%v", info, err)
	}
}

func TestAICutVisualProcessingFiltersExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "green.mp4"), filepath.Join(dir, "processed.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=green:s=64x64:d=0.3", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "v", Kind: aicut.ClipVideo, DurationMS: 300, SourceDurationMS: 300, VisualProcessing: &aicut.VisualProcessing{
		Color:     aicut.ColorAdjustment{Exposure: .4, Temperature: .2, Gamma: 1.1, CurvePoints: []aicut.ColorCurvePoint{{X: 0, Y: 0}, {X: .37, Y: .82}, {X: 1, Y: 1}}},
		ChromaKey: aicut.ChromaKey{Enabled: true, Color: "#00ff00", Similarity: .2, Blend: .05},
		Mask: aicut.VisualMask{Type: "bezier", BezierPoints: []aicut.VisualBezierPoint{
			{X: .1, Y: .1, InX: .05, InY: .2, OutX: .3, OutY: .05},
			{X: .9, Y: .15, InX: .7, InY: .05, OutX: .95, OutY: .35},
			{X: .75, Y: .9, InX: .9, InY: .7, OutX: .55, OutY: .95},
			{X: .2, Y: .8, InX: .4, InY: .95, OutX: .05, OutY: .6},
		}, Feather: .04},
		Geometry: aicut.VisualGeometry{CropLeft: .1, CropTop: .05, CropRight: .1, CropBottom: .05, FlipHorizontal: true, FlipVertical: true},
		Details:  aicut.VisualDetails{Sharpen: .4, Vignette: .5, Grain: .25},
	}}
	clips := []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "video"}, clip: clip}}
	if err := prepareAICutColorLUTs(dir, clips); err != nil {
		t.Fatal(err)
	}
	args, err := buildAICutFFmpegArgs("1:1", clips, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"crop=", "hflip", "vflip", "lut3d=", "unsharp=", "vignette=", "noise=", "pow(g(X,Y)-255,2)", "alpha(X,Y)*ld(0)*ld(0)*(3-2*ld(0))"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q: %s", want, joined)
		}
	}
	if strings.Index(joined, "st(0,clip") > strings.Index(joined, "lut3d=") {
		t.Fatal("source-key alpha must be computed before color grading")
	}

	if !strings.Contains(joined, "0.100000") || !strings.Contains(joined, "0.214815") {
		t.Fatalf("bezier mask sample points missing from export: %s", joined)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("visual export invalid: info=%v err=%v", info, err)
	}
}

func TestAICutRemoveBlackUsesSourceRGBBlackKey(t *testing.T) {
	processing := &aicut.VisualProcessing{ChromaKey: aicut.ChromaKey{RemoveBlack: true, Color: "#00ff00", Similarity: .22, Blend: .04}}
	filters := buildAICutVisualProcessingFilters(processing, true)
	joined := strings.Join(filters, ",")
	if !strings.Contains(joined, "pow(r(X,Y)-0,2)+pow(g(X,Y)-0,2)+pow(b(X,Y)-0,2)") || !strings.Contains(joined, "-0.220000000)/0.040000000") || !strings.Contains(joined, "alpha(X,Y)*ld(0)*ld(0)*(3-2*ld(0))") {
		t.Fatalf("remove black filter=%q", joined)
	}
}

func TestAICutObjectRemovalBuildsExecutableDelogoAfterCanvasScale(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "marked.mp4"), filepath.Join(dir, "cleaned.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=160x90:d=0.3", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "remove", Kind: aicut.ClipVideo, DurationMS: 300, SourceDurationMS: 300, VisualProcessing: &aicut.VisualProcessing{
		ObjectRemoval: aicut.ObjectRemoval{Enabled: true, X: .5, Y: .5, Width: .25, Height: .25},
	}}
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "video"}, clip: clip}}, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if scaleIndex, delogoIndex := strings.Index(joined, "scale=1920:1080"), strings.Index(joined, "delogo=x=720:y=405:w=480:h=270:show=0"); scaleIndex < 0 || delogoIndex < 0 || delogoIndex < scaleIndex {
		t.Fatalf("object removal filter must use final canvas coordinates after scale: %s", joined)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("object removal export invalid: info=%v err=%v", info, err)
	}
}

func TestAICutFrameInterpolationExecutesWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "motion.mp4")
	output := filepath.Join(dir, "interpolated.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=64x64:r=12:d=1", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "flow", Kind: aicut.ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &aicut.VisualProcessing{FrameInterpolation: "optical-flow"}}
	args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "video"}, clip: clip}}, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "minterpolate=fps=30:mi_mode=mci:mc_mode=aobmc:me_mode=bidir:vsbmc=1") {
		t.Fatalf("optical-flow filter missing: %s", joined)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("interpolated export invalid: info=%v err=%v", info, err)
	}
	blend := strings.Join(buildAICutFrameInterpolationFilters(&aicut.VisualProcessing{FrameInterpolation: "blend"}, 60), ",")
	if blend != "minterpolate=fps=60:mi_mode=blend" {
		t.Fatalf("blend filter=%q", blend)
	}
	if filters := buildAICutFrameInterpolationFilters(&aicut.VisualProcessing{FrameInterpolation: "off"}, 60); len(filters) != 0 {
		t.Fatalf("off should not add filters: %#v", filters)
	}
}

func TestAICutStabilizationUsesTwoPassFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "shaky.mp4")
	output := filepath.Join(dir, "stable.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=96x96:r=15:d=1,crop=64:64:x=16+8*sin(n/2):y=16+5*cos(n/3)", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	clip := aicut.Clip{ID: "stable", Kind: aicut.ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &aicut.VisualProcessing{Stabilization: aicut.VideoStabilization{Enabled: true, Smoothness: 30, Zoom: 8, Tripod: true}}}
	prepared := []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "video"}, clip: clip}}
	if err := prepareAICutStabilizationTransforms(t.Context(), dir, prepared); err != nil {
		t.Fatal(err)
	}
	if prepared[0].stabilizationPath == "" {
		t.Fatal("stabilization transform path missing")
	}
	if info, err := os.Stat(prepared[0].stabilizationPath); err != nil || info.Size() == 0 {
		t.Fatalf("stabilization analysis missing: info=%v err=%v", info, err)
	}
	args, err := buildAICutFFmpegArgs("1:1", prepared, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"vidstabtransform=input=", "smoothing=30", "zoom=8.000", "tripod=1"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("stabilization args missing %q: %s", want, joined)
		}
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("stabilized export invalid: info=%v err=%v", info, err)
	}
}

func TestAICutReverseAndFreezeFrameExecuteWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "motion.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=64x64:r=30:d=1", "-c:v", "libx264", "-pix_fmt", "yuv420p", input); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		clip aicut.Clip
		want string
	}{
		{name: "reverse", clip: aicut.Clip{ID: "reverse", Kind: aicut.ClipVideo, DurationMS: 600, SourceDurationMS: 1000, Reversed: true}, want: "reverse"},
		{name: "freeze", clip: func() aicut.Clip {
			value := int64(300)
			return aicut.Clip{ID: "freeze", Kind: aicut.ClipVideo, DurationMS: 600, SourceDurationMS: 1000, FreezeSourceMS: &value}
		}(), want: "tpad=stop_mode=clone"},
	} {
		t.Run(test.name, func(t *testing.T) {
			output := filepath.Join(dir, test.name+".mp4")
			args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{path: input, asset: aicut.Asset{Kind: "video"}, clip: test.clip}}, "mp4", output)
			if err != nil {
				t.Fatal(err)
			}
			if joined := strings.Join(args, " "); !strings.Contains(joined, test.want) {
				t.Fatalf("args missing %q: %s", test.want, joined)
			}
			if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
				t.Fatal(err)
			}
			if info, err := os.Stat(output); err != nil || info.Size() == 0 {
				t.Fatalf("export invalid: info=%v err=%v", info, err)
			}
		})
	}
}

func TestAICutBlendModeExecutesWithFFmpeg(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	base := filepath.Join(dir, "base.mp4")
	top := filepath.Join(dir, "top.png")
	output := filepath.Join(dir, "blended.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=blue:s=64x64:d=0.4", "-c:v", "libx264", "-pix_fmt", "yuv420p", base); err != nil {
		t.Fatal(err)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=red@0.7:s=32x32:d=0.04", "-frames:v", "1", top); err != nil {
		t.Fatal(err)
	}
	clips := []aicutRenderClip{
		{path: base, asset: aicut.Asset{Kind: "video"}, clip: aicut.Clip{ID: "base", Kind: aicut.ClipVideo, DurationMS: 400, SourceDurationMS: 400}},
		{path: top, asset: aicut.Asset{Kind: "image"}, clip: aicut.Clip{ID: "top", Kind: aicut.ClipImage, DurationMS: 400, SourceDurationMS: 400, BlendMode: "multiply", CanvasTransform: &aicut.CanvasTransform{ScaleX: .5, ScaleY: .5, Opacity: .7}}},
	}
	args, err := buildAICutFFmpegArgs("1:1", clips, "mp4", output)
	if err != nil {
		t.Fatal(err)
	}
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(output); err != nil || info.Size() == 0 {
		t.Fatalf("blend export invalid: info=%v err=%v", info, err)
	}
	raw := filepath.Join(dir, "blended.rgb")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", output, "-frames:v", "1", "-vf", "scale=64:64,format=rgb24", "-f", "rawvideo", raw); err != nil {
		t.Fatal(err)
	}
	pixels, err := os.ReadFile(raw)
	if err != nil || len(pixels) < 64*64*3 {
		t.Fatalf("cannot inspect blended pixels: bytes=%d err=%v", len(pixels), err)
	}
	pixel := func(x, y int) []byte { offset := (y*64 + x) * 3; return pixels[offset : offset+3] }
	corner, center := pixel(2, 2), pixel(32, 32)
	if corner[2] < 120 || corner[2] <= corner[0] {
		t.Fatalf("transparent outside area did not preserve blue base: rgb=%v", corner)
	}
	if int(center[0])+int(center[1])+int(center[2]) >= int(corner[0])+int(corner[1])+int(corner[2]) || center[2] >= corner[2] {
		t.Fatalf("multiply mode did not darken overlapped center: corner=%v center=%v", corner, center)
	}
}
