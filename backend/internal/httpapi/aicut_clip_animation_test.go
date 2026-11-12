package httpapi

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

func TestAICutClipAnimationFFmpegFrames(t *testing.T) {
	ffmpeg, err := ffmpegbin.ResolveFFmpeg()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, phase, preset string
		time                float64
		wantAlpha           int
		wantX               int
		wantY               int
	}{
		{"fade midpoint", "entrance", "fade", .25, 127, 32, 24},
		{"slide blank right", "entrance", "slide", .25, 0, 40, 24},
		{"slide visible left", "entrance", "slide", .25, 255, 4, 24},
		{"zoom corner transparent", "entrance", "zoom", 0, 0, 0, 24},
		{"zoom center opaque", "entrance", "zoom", .25, 255, 32, 24},
		{"exit fade midpoint", "exit", "fade", .75, 127, 32, 24},
		{"exit slide left transparent", "exit", "slide", .75, 0, 4, 24},
		{"exit zoom corner transparent", "exit", "zoom", .75, 0, 0, 24},
		{"loop float top transparent", "loop", "float", .25, 0, 32, 0},
		{"loop pulse center opaque", "loop", "pulse", .5, 255, 32, 24},
	} {
		t.Run(tc.name, func(t *testing.T) {
			animation := &aicut.ClipAnimation{}
			phase := aicut.ClipAnimationPhase{Preset: tc.preset, DurationMS: 500}
			switch tc.phase {
			case "entrance":
				animation.Entrance = phase
			case "exit":
				animation.Exit = phase
			case "loop":
				phase.DurationMS = 1000
				animation.Loop = phase
			}
			clip := aicut.Clip{StartMS: 1000, DurationMS: 1000, Animation: animation}
			filters := append([]string{"format=rgba", "setpts=PTS+1/TB"}, buildAICutClipAnimationFilters(clip, true)...)
			// Select the same output-time sample used by browser sampling, with a nonzero timeline start.
			filters = append(filters, "select='eq(n,"+map[float64]string{0: "0", .25: "1", .5: "2", .75: "3"}[tc.time]+")'")
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, ffmpeg, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=red:s=64x48:r=4:d=1", "-vf", strings.Join(filters, ","), "-frames:v", "1", "-f", "rawvideo", "-pix_fmt", "rgba", "-")
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			raw, err := cmd.Output()
			if err != nil {
				t.Fatalf("%v: %s", err, stderr.String())
			}
			if len(raw) != 64*48*4 {
				t.Fatalf("bad frame size %d", len(raw))
			}
			got := int(raw[(tc.wantY*64+tc.wantX)*4+3])
			if got < tc.wantAlpha-2 || got > tc.wantAlpha+2 {
				t.Fatalf("alpha %d want %d", got, tc.wantAlpha)
			}
		})
	}
}

func TestAICutClipAnimationMP4Export(t *testing.T) {
	ffmpeg, err := ffmpegbin.ResolveFFmpeg()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "red.png")
	output := filepath.Join(dir, "fade.mp4")
	source := image.NewNRGBA(image.Rect(0, 0, 64, 36))
	for y := 0; y < 36; y++ {
		for x := 0; x < 64; x++ {
			source.SetNRGBA(x, y, color.NRGBA{R: 255, A: 255})
		}
	}
	f, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, source); err != nil {
		t.Fatal(err)
	}
	f.Close()
	clip := aicutRenderClip{clip: aicut.Clip{DurationMS: 500, Animation: &aicut.ClipAnimation{Entrance: aicut.ClipAnimationPhase{Preset: "fade", DurationMS: 500}}}, asset: aicut.Asset{Kind: "image"}, path: input}
	args, err := buildAICutFFmpegArgsWithOptions("16:9", []aicutRenderClip{clip}, "mp4", output, aicutExportOptions{Resolution: "720p", FrameRate: 30, Quality: "standard", Codec: "h264", BitDepth: 8, HardwareAcceleration: "software"}, "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	if raw, err := exec.CommandContext(ctx, ffmpeg, args...).CombinedOutput(); err != nil {
		t.Fatalf("export: %v %s", err, raw)
	}
	// At 0.2 output seconds a 0.5-second fade is 40% red over black.
	raw, err := exec.CommandContext(ctx, ffmpeg, "-hide_banner", "-loglevel", "error", "-ss", "0.2", "-i", output, "-vf", "scale=2:2", "-frames:v", "1", "-f", "rawvideo", "-pix_fmt", "rgb24", "-").Output()
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 12 || raw[0] < 90 || raw[0] > 115 || raw[1] > 8 || raw[2] > 8 {
		t.Fatalf("export midpoint mismatch: %v", raw)
	}
}
