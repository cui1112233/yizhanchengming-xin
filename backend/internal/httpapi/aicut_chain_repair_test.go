package httpapi

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"strings"
	"testing"
)

func TestRepairCurvedTypewriterRespectsFinalOpacity(t *testing.T) {
	for _, opacity := range []float64{0, .5} {
		timeline := aicut.Timeline{Tracks: []aicut.Track{{Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{TextCurve: 40, SubtitleAnimation: "typewriter", TextOpacity: &opacity}, Clips: []aicut.Clip{{DurationMS: 1000, Text: "测试"}}}}}
		ass := buildAICutASS(timeline, 720, 720)
		want := fmt.Sprintf("\\alpha&H%02X&)}", int(math.Round((1-opacity)*255)))
		if !strings.Contains(ass, want) || strings.Contains(ass, "\\alpha&H00&)}") {
			t.Fatal(ass)
		}
	}
}

func TestRepairBreatheLayerKeepsAnimatedFramesAfterNonzeroStart(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	input, out := filepath.Join(dir, "motion.mp4"), filepath.Join(dir, "out.mp4")
	run := func(args ...string) {
		t.Helper()
		if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
			t.Fatal(err)
		}
	}
	run("-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=64x64:r=24:d=1", "-c:v", "libx264", input)
	clips := []aicutRenderClip{{clip: aicut.Clip{StartMS: 2000, DurationMS: 1000}, asset: aicut.Asset{Kind: "video"}, path: input, layerEffects: []aicut.Clip{{StartMS: 2000, DurationMS: 1000, ResourceID: "effect-breathe-zoom"}}}}
	args, err := buildAICutFFmpegArgsWithOptions("1:1", clips, "mp4", out, aicutExportOptions{FrameRate: 24, Resolution: "720p", Quality: "draft"}, "")
	if err != nil {
		t.Fatal(err)
	}
	run(args...)
	samples := [][]byte{}
	for _, stamp := range []string{"2.1", "2.7"} {
		raw := filepath.Join(dir, stamp+".raw")
		run("-hide_banner", "-loglevel", "error", "-ss", stamp, "-i", out, "-frames:v", "1", "-vf", "scale=32:32,format=rgb24", "-f", "rawvideo", raw)
		data, err := os.ReadFile(raw)
		if err != nil || len(data) == 0 {
			t.Fatal("frame missing", err)
		}
		samples = append(samples, data)
	}
	if bytes.Equal(samples[0], samples[1]) {
		t.Fatal("breathe layer froze animated source after timeline offset")
	}
	difference := 0.0
	for i, v := range samples[0] {
		difference += math.Abs(float64(v) - float64(samples[1][i]))
	}
	difference /= float64(len(samples[0]))
	if difference < 2 {
		t.Fatalf("source animation froze; averagepixel difference%f", difference)
	}
	if strings.Contains(strings.Join(args, " "), "fps=30") {
		t.Fatal("breathe layer ignored selected24fps")
	}
}

func TestRepairFinalVolumeFollowsLoudnessProcessing(t *testing.T) {
	args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 300, Volume: 0, AudioProcessing: &aicut.AudioProcessing{Denoise: .5, LoudnessEnabled: true, TargetLUFS: -16}}, asset: aicut.Asset{Kind: "audio"}, path: "tone.wav"}}, "mp3", "silent.mp3")
	if err != nil {
		t.Fatal(err)
	}
	s := strings.Join(args, " ")
	if strings.Index(s, "volume=0.000") < strings.Index(s, "loudnorm=") {
		t.Fatalf("normalization can override requested volume: %s", s)
	}
}

func TestRepairColorUsesPreviewRGBMath(t *testing.T) {
	got := aicutBasicColor([3]float64{.2, .3, .4}, aicut.ColorAdjustment{Exposure: 1, Temperature: .5, Tint: -.5, Gamma: 1})
	for i, want := range []float64{.45, .56, .75} {
		if math.Abs(got[i]-want) > 1e-9 {
			t.Fatalf("channel%d=%f want%f", i, got[i], want)
		}
	}
	clips := []aicutRenderClip{{clip: aicut.Clip{VisualProcessing: &aicut.VisualProcessing{Color: aicut.ColorAdjustment{Exposure: 1, Gamma: 1}}}}}
	if err := prepareAICutColorLUTs(t.TempDir(), clips); err != nil {
		t.Fatal(err)
	}
	filters := buildAICutVisualProcessingFilters(clips[0].clip.VisualProcessing, false, clips[0].hslLUTPath)
	if len(filters) != 1 || !strings.HasPrefix(filters[0], "lut3d=") {
		t.Fatalf("basic color escaped common LUT: %v", filters)
	}
}
func TestRepairGradientAndTextOpacityExport(t *testing.T) {
	zero := 0.0
	timeline := aicut.Timeline{Tracks: []aicut.Track{{Kind: aicut.TrackSubtitle, Settings: aicut.TrackSettings{TextResourceID: "text-gradient-title", TextOpacity: &zero}, Clips: []aicut.Clip{{DurationMS: 1000, Text: "甲乙丙"}}}}}
	ass := buildAICutASS(timeline, 720, 720)
	for _, want := range []string{"\\1c&H6BD6FF&", "\\1c&HFF8BB7&", "\\1c&HFFDC63&", "\\alpha&HFF&"} {
		if !strings.Contains(ass, want) {
			t.Fatalf("missing%s: %s", want, ass)
		}
	}
}
func TestRepairCatalogReplacesLegacyFilterOnce(t *testing.T) {
	args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 1000, CatalogResourceID: "filter-bright", CatalogParameters: map[string]float64{"intensity": 1}, VisualEffect: &aicut.VisualEffect{Brightness: .7, Contrast: 1, Saturation: 1}}, asset: aicut.Asset{Kind: "video"}, path: "a.mp4"}}, "mp4", "o.mp4")
	if err != nil {
		t.Fatal(err)
	}
	s := strings.Join(args, " ")
	if !strings.Contains(s, "brightness=0.120") || strings.Contains(s, "brightness=0.700") {
		t.Fatal(s)
	}
}

func TestRepairActualFFmpegWindowsRampAndSilence(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	archive := "../../../_local-archive-20261028/aicut-chain-repair"
	if err := os.MkdirAll(archive, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(archive, "ffmpeg-")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dir)
	video := filepath.Join(dir, "black.mp4")
	out := filepath.Join(dir, "window.mp4")
	run := func(args ...string) {
		t.Helper()
		if err := ffmpegbin.Run(t.Context(), "ffmpeg", args...); err != nil {
			t.Fatal(err)
		}
	}
	run("-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=black:s=64x64:r=30:d=2", "-c:v", "libx264", video)
	clips := []aicutRenderClip{{clip: aicut.Clip{DurationMS: 1000, SpeedCurve: []aicut.SpeedCurvePoint{{Position: 0, Rate: .5}, {Position: 1, Rate: 2}}}, asset: aicut.Asset{Kind: "video"}, path: video, layerEffects: []aicut.Clip{{StartMS: 400, DurationMS: 200, ResourceID: "effect-flash", Parameters: map[string]float64{"intensity": 1}}}}}
	args, err := buildAICutFFmpegArgsWithOptions("1:1", clips, "mp4", out, aicutExportOptions{Resolution: "720p", Quality: "draft"}, "")
	if err != nil {
		t.Fatal(err)
	}
	run(args...)
	for _, sample := range []struct {
		time   string
		bright bool
	}{{"0.1", false}, {"0.5", true}, {"0.8", false}} {
		raw := filepath.Join(dir, "sample"+sample.time+".raw")
		run("-hide_banner", "-loglevel", "error", "-ss", sample.time, "-i", out, "-frames:v", "1", "-vf", "crop=2:2:0:0,scale=1:1,format=gray", "-f", "rawvideo", raw)
		data, err := os.ReadFile(raw)
		if err != nil || len(data) != 1 {
			t.Fatalf("sample missing: %v %v", data, err)
		}
		if (data[0] > 10) != sample.bright {
			t.Fatalf("window at %s: gray=%d", sample.time, data[0])
		}
	}
	// Source frame brightness identifies the consumed source time. At output .5s,
	// integral(.5+1.5*t, 0.. .5) is .4375s, rather than flattened .625s.
	rampSource, rampOut, rampPixel := filepath.Join(dir, "ramp-source.mp4"), filepath.Join(dir, "ramp-out.mp4"), filepath.Join(dir, "ramp.raw")
	run("-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "nullsrc=s=64x64:r=30:d=2,geq=lum='N*4':cb=128:cr=128", "-c:v", "libx264", "-crf", "0", rampSource)
	clips[0].path = rampSource
	clips[0].layerEffects = nil
	args, err = buildAICutFFmpegArgsWithOptions("1:1", clips, "mp4", rampOut, aicutExportOptions{Resolution: "720p", Quality: "draft"}, "")
	if err != nil {
		t.Fatal(err)
	}
	run(args...)
	run("-hide_banner", "-loglevel", "error", "-ss", "0.5", "-i", rampOut, "-frames:v", "1", "-vf", "crop=2:2:0:0,scale=1:1,format=gray", "-f", "rawvideo", rampPixel)
	rampData, err := os.ReadFile(rampPixel)
	ref := filepath.Join(dir, "ref.raw")
	run("-hide_banner", "-loglevel", "error", "-ss", "0.4375", "-i", rampSource, "-frames:v", "1", "-vf", "crop=2:2:0:0,scale=1:1,format=gray", "-f", "rawvideo", ref)
	refData, _ := os.ReadFile(ref)
	t.Logf("ramp sample %v source at integrated timestamp %v", rampData, refData)
	if err != nil || len(rampData) != 1 || len(refData) != 1 || math.Abs(float64(rampData[0])-float64(refData[0])) > 8 {
		t.Fatalf("ramp source at .5 sec must correspond .4375 sec: %v %v", rampData, err)
	}
	audio := filepath.Join(dir, "tone.wav")
	silent := filepath.Join(dir, "silent.mp3")
	run("-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=0.3", audio)
	args, err = buildAICutFFmpegArgs("1:1", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 300, Volume: 0, AudioProcessing: &aicut.AudioProcessing{Denoise: .5, LoudnessEnabled: true, TargetLUFS: -16}}, asset: aicut.Asset{Kind: "audio"}, path: audio}}, "mp3", silent)
	if err != nil {
		t.Fatal(err)
	}
	run(args...)
	pcm := filepath.Join(dir, "silent.raw")
	run("-hide_banner", "-loglevel", "error", "-i", silent, "-f", "s16le", pcm)
	data, err := os.ReadFile(pcm)
	if err != nil || len(data) == 0 {
		t.Fatal("silent decode missing")
	}
	for _, b := range data {
		if b != 0 {
			t.Fatal("zero volume generated audible samples")
		}
	}
}

func TestRepairZeroVolumeAndEffectWindow(t *testing.T) {
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 2000, Volume: 0}, asset: aicut.Asset{Kind: "audio"}, path: "a.wav"}}, "mp3", "a.mp3")
	if err != nil || !strings.Contains(strings.Join(args, " "), "volume=0.000") {
		t.Fatalf("zero changed: %v %v", args, err)
	}
	args, err = buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 10000}, asset: aicut.Asset{Kind: "video"}, path: "v.mp4", layerEffects: []aicut.Clip{{StartMS: 4000, DurationMS: 1000, ResourceID: "effect-blur"}}}}, "mp4", "v.mp4")
	if err != nil || !strings.Contains(strings.Join(args, " "), "gte(t,4.000)*lt(t,5.000)") {
		t.Fatalf("missing effect window: %v %v", args, err)
	}
}

func TestRepairContinuousRamp(t *testing.T) {
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: aicut.Clip{DurationMS: 4000, SpeedCurve: []aicut.SpeedCurvePoint{{Position: 0, Rate: .5}, {Position: 1, Rate: 2}}}, asset: aicut.Asset{Kind: "video"}, path: "v.mp4"}}, "mp4", "v.mp4")
	if err != nil || !strings.Contains(strings.Join(args, " "), "sqrt(") {
		t.Fatalf("linear ramp flattened: %v %v", args, err)
	}
}

func TestRepairZeroCatalogIntensity(t *testing.T) {
	filters, _ := buildAICutCatalogResourceFilters("filter-noir", map[string]float64{"intensity": 0})
	if len(filters) != 0 {
		t.Fatalf("zero intensity applies filters: %v", filters)
	}
}

func TestRepairCompositionUsesClipLocalEffectPhase(t *testing.T) {
	args, err := buildAICutFFmpegArgs("1:1", []aicutRenderClip{{clip: aicut.Clip{StartMS: 2000, DurationMS: 1000}, asset: aicut.Asset{Kind: "video"}, path: "v.mp4", layerEffects: []aicut.Clip{{StartMS: 2400, DurationMS: 200, ResourceID: "effect-flash"}}}}, "mp4", "v.mp4")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "sin((t-2.000000)*") || !strings.Contains(joined, "gte(t,2.400)*lt(t,2.600)") {
		t.Fatal(joined)
	}
}
