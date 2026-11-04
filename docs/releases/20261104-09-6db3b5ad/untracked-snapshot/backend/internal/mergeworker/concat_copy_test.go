package mergeworker

import "testing"

func validClip() probeStream {
	return probeStream{
		CodecName: "h264", Width: 720, Height: 1280, PixFmt: "yuv420p",
		SampleAspectRatio: "1:1", RFrameRate: "30/1", TimeBase: "1/15360",
		SampleRate: "48000", Channels: 2,
	}
}

func TestConcatCopyCompatibleAcceptsIdenticalClips(t *testing.T) {
	clips := []probeStream{validClip(), validClip()}
	if !concatCopyCompatible(clips, 720, 1280) {
		t.Fatalf("identical clips should be compatible")
	}
}

func TestConcatCopyCompatibleRejectsParameterDrift(t *testing.T) {
	cases := map[string]func(*probeStream){
		"frame rate": func(clip *probeStream) { clip.RFrameRate = "24/1" },
		"resolution": func(clip *probeStream) { clip.Width = 1280 },
		"pixel fmt":  func(clip *probeStream) { clip.PixFmt = "yuv422p" },
		"time base":  func(clip *probeStream) { clip.TimeBase = "1/12288" },
		"audio rate": func(clip *probeStream) { clip.SampleRate = "44100" },
		"audio ch":   func(clip *probeStream) { clip.Channels = 1 },
	}
	for name, mutate := range cases {
		second := validClip()
		mutate(&second)
		if concatCopyCompatible([]probeStream{validClip(), second}, 720, 1280) {
			t.Fatalf("%s drift must fall back to re-encode", name)
		}
	}
}

func TestConcatCopySwitchDefaultsOff(t *testing.T) {
	t.Setenv(EnvConcatCopy, "")
	if concatCopyEnabled() {
		t.Fatalf("container concat must be off by default")
	}
	t.Setenv(EnvConcatCopy, "1")
	if !concatCopyEnabled() {
		t.Fatalf("container concat should follow the explicit switch")
	}
}
