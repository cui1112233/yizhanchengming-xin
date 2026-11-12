package httpapi

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

func TestAICutHSLLUTPreparationAndFilter(t *testing.T) {
	clips := []aicutRenderClip{{clip: aicut.Clip{VisualProcessing: &aicut.VisualProcessing{Color: aicut.ColorAdjustment{HSL: map[string]aicut.HSLBandAdjustment{"red": {Hue: 100}}}}}}}
	dir := t.TempDir()
	if err := prepareAICutHSLLUTs(dir, clips); err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(clips[0].hslLUTPath) != dir {
		t.Fatal("LUT escaped render directory")
	}
	data, err := os.ReadFile(clips[0].hslLUTPath)
	if err != nil || !strings.Contains(string(data), "LUT_3D_SIZE 33") {
		t.Fatal("LUT missing", err)
	}
	filters := buildAICutVisualProcessingFilters(clips[0].clip.VisualProcessing, true, clips[0].hslLUTPath)
	if !strings.Contains(strings.Join(filters, ","), "lut3d=file='") || !strings.Contains(strings.Join(filters, ","), ":interp=trilinear") {
		t.Fatal("missing trilinear LUT", filters)
	}
	neutral := []aicutRenderClip{{clip: aicut.Clip{VisualProcessing: &aicut.VisualProcessing{}}}}
	if err := prepareAICutHSLLUTs(dir, neutral); err != nil || neutral[0].hslLUTPath != "" {
		t.Fatal("identity must not create LUT")
	}
}

func TestAICutHSLExportRefusesMissingLUT(t *testing.T) {
	clips := []aicutRenderClip{{path: "image.png", asset: aicut.Asset{Kind: "image"}, clip: aicut.Clip{Kind: aicut.ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &aicut.VisualProcessing{Color: aicut.ColorAdjustment{HSL: map[string]aicut.HSLBandAdjustment{"red": {Hue: 100}}}}}}}
	if _, err := buildAICutFFmpegArgs("16:9", clips, "mp4", "output.mp4"); err == nil || !strings.Contains(err.Error(), "HSL") {
		t.Fatalf("active HSL silently dropped: %v", err)
	}
}

func TestAICutHSLFFmpegChangesRedPreservesGreenAndPNGAlpha(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	source := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	source.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	source.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	source.SetNRGBA(2, 0, color.NRGBA{R: 255, A: 77})
	source.SetNRGBA(3, 0, color.NRGBA{R: 83, G: 83, B: 83, A: 0})
	input := filepath.Join(dir, "bars.png")
	output := filepath.Join(dir, "adjusted.png")
	f, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, source); err != nil {
		t.Fatal(err)
	}
	f.Close()
	processing := &aicut.VisualProcessing{Color: aicut.ColorAdjustment{HSL: map[string]aicut.HSLBandAdjustment{"red": {Hue: 100}}}}
	clips := []aicutRenderClip{{clip: aicut.Clip{VisualProcessing: processing}}}
	if err := prepareAICutHSLLUTs(dir, clips); err != nil {
		t.Fatal(err)
	}
	filters := buildAICutVisualProcessingFilters(processing, true, clips[0].hslLUTPath)
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", input, "-vf", strings.Join(filters, ","), "-frames:v", "1", output); err != nil {
		t.Fatal(err)
	}
	resultFile, err := os.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer resultFile.Close()
	decoded, err := png.Decode(resultFile)
	if err != nil {
		t.Fatal(err)
	}
	for x, want := range []color.NRGBA{{R: 255, G: 128, A: 255}, {G: 255, A: 255}, {R: 255, G: 128, A: 77}, {R: 83, G: 83, B: 83, A: 0}} {
		got := color.NRGBAModel.Convert(decoded.At(x, 0)).(color.NRGBA)
		if got.A != want.A {
			t.Errorf("pixel %d alpha got %d want %d", x, got.A, want.A)
		}
		if want.A != 0 && (absByte(got.R, want.R) > 1 || absByte(got.G, want.G) > 1 || absByte(got.B, want.B) > 1) {
			t.Errorf("pixel %d got %v want %v", x, got, want)
		}
	}
}

func absByte(a, b uint8) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}

func TestAICutHSLFFmpegKeysOriginalSourceBeforeColorAdjustment(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	for _, scenario := range []struct {
		name       string
		key        color.NRGBA
		processing aicut.VisualProcessing
	}{
		{name: "green-hue", key: color.NRGBA{G: 255, A: 255}, processing: aicut.VisualProcessing{ChromaKey: aicut.ChromaKey{Enabled: true, Color: "#00ff00", Similarity: .1, Blend: .02}, Color: aicut.ColorAdjustment{HSL: map[string]aicut.HSLBandAdjustment{"green": {Hue: 100}}}}},
		{name: "green-hue-exposure", key: color.NRGBA{G: 255, A: 255}, processing: aicut.VisualProcessing{ChromaKey: aicut.ChromaKey{Enabled: true, Color: "#00ff00", Similarity: .1, Blend: .02}, Color: aicut.ColorAdjustment{Exposure: 1, Gamma: 1.2, HSL: map[string]aicut.HSLBandAdjustment{"green": {Hue: 100}}}}},
		{name: "black-lift", key: color.NRGBA{A: 255}, processing: aicut.VisualProcessing{ChromaKey: aicut.ChromaKey{RemoveBlack: true, Color: "#000000", Similarity: .1, Blend: .02}, Color: aicut.ColorAdjustment{Curve: []float64{.4, .5, .6, .8, 1}, HSL: map[string]aicut.HSLBandAdjustment{"red": {Hue: 100}}}}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dir := t.TempDir()
			source := image.NewNRGBA(image.Rect(0, 0, 80, 16))
			for y := 0; y < 16; y++ {
				for x := 0; x < 80; x++ {
					pixel := scenario.key
					if x >= 16 {
						pixel = color.NRGBA{R: 255, A: 255}
					}
					if x >= 32 {
						pixel.A = 77
					}
					if x >= 48 {
						pixel.A = 0
					}
					if x >= 64 {
						pixel = scenario.key
						pixel.R = 28
						pixel.A = 77
					}
					source.SetNRGBA(x, y, pixel)
				}
			}
			input, output := filepath.Join(dir, "screen.png"), filepath.Join(dir, "keyed.png")
			f, err := os.Create(input)
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, source); err != nil {
				t.Fatal(err)
			}
			f.Close()
			clips := []aicutRenderClip{{clip: aicut.Clip{VisualProcessing: &scenario.processing}}}
			if err := prepareAICutHSLLUTs(dir, clips); err != nil {
				t.Fatal(err)
			}
			filters := buildAICutVisualProcessingFilters(&scenario.processing, true, clips[0].hslLUTPath)
			if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", input, "-vf", strings.Join(filters, ","), "-frames:v", "1", output); err != nil {
				t.Fatal(err)
			}
			f, err = os.Open(output)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			decoded, err := png.Decode(f)
			if err != nil {
				t.Fatal(err)
			}
			for _, sample := range []struct {
				x     int
				alpha uint8
			}{{8, 0}, {24, 255}, {40, 77}, {56, 0}, {72, 37}} {
				got := color.NRGBAModel.Convert(decoded.At(sample.x, 8)).(color.NRGBA)
				if absByte(got.A, sample.alpha) > 1 {
					t.Errorf("source-key alpha at %d got %d want %d; filters=%v", sample.x, got.A, sample.alpha, filters)
				}
			}
		})
	}
}
