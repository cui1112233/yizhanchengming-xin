package httpapi

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"strings"
	"testing"
)

func TestAICutColorGradeFFmpegRGBWheelsAndAlpha(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	for _, scenario := range []struct {
		name        string
		pixel, want color.NRGBA
		grade       aicut.ColorAdjustment
	}{
		{name: "rgb-red", pixel: color.NRGBA{R: 64, G: 100, B: 200, A: 77}, want: color.NRGBA{R: 128, G: 100, B: 200, A: 77}, grade: aicut.ColorAdjustment{RGBCurves: map[string][]aicut.ColorCurvePoint{"red": {{X: 0, Y: 0}, {X: .5, Y: 1}, {X: 1, Y: 1}}}}},
		{name: "shadow-wheel", pixel: color.NRGBA{R: 128, G: 128, B: 128, A: 77}, want: color.NRGBA{R: 140, G: 125, B: 125, A: 77}, grade: aicut.ColorAdjustment{Wheels: map[string]aicut.ColorWheelAdjustment{"shadows": {Saturation: 100}}}},
		{name: "rgb-before-hsl", pixel: color.NRGBA{R: 255, A: 255}, want: color.NRGBA{R: 255, G: 255, B: 255, A: 255}, grade: aicut.ColorAdjustment{RGBCurves: map[string][]aicut.ColorCurvePoint{"red": {{X: 0, Y: 0}, {X: 1, Y: .5}}}, HSL: map[string]aicut.HSLBandAdjustment{"red": {Luminance: 100}}}},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dir := t.TempDir()
			source := image.NewNRGBA(image.Rect(0, 0, 2, 1))
			source.SetNRGBA(0, 0, scenario.pixel)
			source.SetNRGBA(1, 0, color.NRGBA{R: 64, G: 100, B: 200, A: 0})
			input, output := filepath.Join(dir, "input.png"), filepath.Join(dir, "output.png")
			f, err := os.Create(input)
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, source); err != nil {
				t.Fatal(err)
			}
			f.Close()
			processing := &aicut.VisualProcessing{Color: scenario.grade}
			clips := []aicutRenderClip{{clip: aicut.Clip{VisualProcessing: processing}}}
			if err := prepareAICutHSLLUTs(dir, clips); err != nil {
				t.Fatal(err)
			}
			filters := buildAICutVisualProcessingFilters(processing, true, clips[0].hslLUTPath)
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
			got := color.NRGBAModel.Convert(decoded.At(0, 0)).(color.NRGBA)
			if got.A != scenario.want.A || absByte(got.R, scenario.want.R) > 2 || absByte(got.G, scenario.want.G) > 2 || absByte(got.B, scenario.want.B) > 2 {
				t.Errorf("got %v want %v", got, scenario.want)
			}
			if alpha := color.NRGBAModel.Convert(decoded.At(1, 0)).(color.NRGBA).A; alpha != 0 {
				t.Errorf("zero alpha became %d", alpha)
			}
		})
	}
}
