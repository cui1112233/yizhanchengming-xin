package httpapi

import (
	"encoding/binary"
	"math"
	"os/exec"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"strings"
	"testing"
)

func TestAICutStereoBalanceFilters(t *testing.T) {
	if filters := buildAICutBalanceFilters(0); len(filters) != 0 {
		t.Fatalf("center must be identity: %v", filters)
	}
	for _, tc := range []struct {
		pan  float64
		want string
	}{{1, "c0=0.000000000*c0|c1=1.000000000*c1"}, {-1, "c0=1.000000000*c0|c1=0.000000000*c1"}, {.25, "c0=0.750000000*c0|c1=1.000000000*c1"}} {
		if got := strings.Join(buildAICutBalanceFilters(tc.pan), ","); !strings.Contains(got, tc.want) {
			t.Fatalf("pan %v filters: %s", tc.pan, got)
		}
	}
	args, err := buildAICutFFmpegArgs("16:9", []aicutRenderClip{{clip: aicut.Clip{Kind: aicut.ClipAudio, DurationMS: 1000, AudioProcessing: &aicut.AudioProcessing{Pan: .25}}, asset: aicut.Asset{Kind: "audio"}, path: "input.wav"}}, "mp3", "output.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(args, " "), "pan=stereo|c0=0.750000000*c0|c1=1.000000000*c1") {
		t.Fatal("balance missing from actual export graph")
	}
}

func TestAICutStereoBalanceRealSignal(t *testing.T) {
	ffmpeg, err := ffmpegbin.ResolveFFmpeg()
	if err != nil {
		t.Skip(err)
	}
	render := func(source, filter string) []float64 {
		t.Helper()
		args := []string{"-v", "error", "-f", "lavfi", "-i", source}
		if filter != "" {
			args = append(args, "-af", filter)
		}
		args = append(args, "-t", "0.1", "-acodec", "pcm_f64le", "-f", "f64le", "-")
		output, err := exec.Command(ffmpeg, args...).Output()
		if err != nil {
			t.Fatalf("signal render: %v", err)
		}
		result := make([]float64, len(output)/8)
		for i := range result {
			result[i] = math.Float64frombits(binary.LittleEndian.Uint64(output[i*8 : i*8+8]))
		}
		return result
	}
	// Independent frequencies ensure that accidentally crossmixing the channels fails.
	source := "aevalsrc=0.2*sin(2*PI*440*t)|0.3*sin(2*PI*880*t):s=48000:d=0.1:c=stereo"
	before := render(source, "")
	for _, tc := range []struct{ pan, left, right float64 }{{0, 1, 1}, {.375, .625, 1}, {-.625, 1, .375}, {1, 0, 1}, {-1, 1, 0}} {
		after := render(source, strings.Join(buildAICutBalanceFilters(tc.pan), ","))
		if len(after) != len(before) {
			t.Fatal("sample duration changed")
		}
		for i, value := range after {
			gain := tc.left
			if i%2 == 1 {
				gain = tc.right
			}
			if math.Abs(value-before[i]*gain) > 1e-7 {
				t.Fatalf("pan %v sample %d expected %v got %v", tc.pan, i, before[i]*gain, value)
			}
		}
	}
	mono := render("aevalsrc=0.2*sin(2*PI*440*t):s=48000:d=0.1:c=mono", strings.Join(buildAICutBalanceFilters(.5), ","))
	for i := 0; i < len(mono); i += 2 {
		expected := .2 * math.Sin(2*math.Pi*440*float64(i/2)/48000)
		if math.Abs(mono[i]-expected*.5) > 1e-7 || math.Abs(mono[i+1]-expected) > 1e-7 {
			t.Fatalf("mono must duplicate at unity before balance: %v %v", mono[i], mono[i+1])
		}
	}
}
