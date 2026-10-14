package mergeworker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeWithAudioMapsTheMeasuredNarrationInsteadOfSourceAudio(t *testing.T) {
	dir := t.TempDir()
	video := filepath.Join(dir, "video.mp4")
	audio := filepath.Join(dir, "tts.mp3")
	if err := os.WriteFile(video, []byte("video"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio, []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got []string
	runner := &FFmpegRunner{Binary: "ffmpeg", Exec: func(_ context.Context, _ string, args ...string) error {
		got = append([]string(nil), args...)
		return nil
	}}
	if err := runner.MergeWithAudio(context.Background(), []string{video}, audio, filepath.Join(dir, "out.mp4"), 1, "9:16"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "-map [tts]") || strings.Contains(joined, "[0:a:0]") {
		t.Fatalf("ffmpeg args must replace source audio with TTS: %s", joined)
	}
}
