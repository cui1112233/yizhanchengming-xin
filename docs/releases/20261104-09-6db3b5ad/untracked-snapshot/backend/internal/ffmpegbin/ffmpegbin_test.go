package ffmpegbin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFakeBinary(t *testing.T, dir, base string) string {
	t.Helper()
	name := exeName(base)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("fake"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	return path
}

func inIsolatedDir(t *testing.T) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	isolated := t.TempDir()
	if err := os.Chdir(isolated); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(original)
	})
}

func TestResolvePrefersBundledToolsDirectory(t *testing.T) {
	root := t.TempDir()
	ffmpegDir := filepath.Join(root, "ffmpeg")
	if err := os.MkdirAll(ffmpegDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	want := writeFakeBinary(t, ffmpegDir, "ffmpeg")

	t.Setenv(EnvToolsDir, root)
	t.Setenv(EnvFFmpegPath, "")

	got, err := ResolveFFmpeg()
	if err != nil {
		t.Fatalf("ResolveFFmpeg: %v", err)
	}
	if !filepath.IsAbs(got) || !strings.EqualFold(got, want) {
		t.Fatalf("expected bundled ffmpeg %q, got %q", want, got)
	}
}

func TestResolveExplicitEnvWinsOverBundled(t *testing.T) {
	root := t.TempDir()
	bundledDir := filepath.Join(root, "ffmpeg")
	if err := os.MkdirAll(bundledDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFakeBinary(t, bundledDir, "ffmpeg")

	explicitDir := t.TempDir()
	explicit := writeFakeBinary(t, explicitDir, "ffmpeg")

	t.Setenv(EnvToolsDir, root)
	t.Setenv(EnvFFmpegPath, explicit)

	got, err := ResolveFFmpeg()
	if err != nil {
		t.Fatalf("ResolveFFmpeg: %v", err)
	}
	if got != explicit {
		t.Fatalf("expected explicit %q, got %q", explicit, got)
	}
}

func TestResolveReportsMissingWithExpectedPaths(t *testing.T) {
	inIsolatedDir(t)
	root := t.TempDir()
	t.Setenv(EnvToolsDir, root)
	t.Setenv(EnvFFmpegPath, "")
	t.Setenv("PATH", "")

	_, err := ResolveFFmpeg()
	if err == nil {
		t.Fatalf("expected missing error")
	}
	message := err.Error()
	if !strings.Contains(message, "缺少视频合成引擎（ffmpeg）") {
		t.Fatalf("missing message is not user facing: %s", message)
	}
	if !strings.Contains(message, "期望路径") {
		t.Fatalf("missing message lacks expected path: %s", message)
	}
	if !IsMissing(err) {
		t.Fatalf("IsMissing should report true")
	}
}

func TestResolveProbeFallsBackToBundledFFmpeg(t *testing.T) {
	inIsolatedDir(t)
	root := t.TempDir()
	ffmpegDir := filepath.Join(root, "ffmpeg")
	if err := os.MkdirAll(ffmpegDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ffmpegPath := writeFakeBinary(t, ffmpegDir, "ffmpeg")
	t.Setenv(EnvToolsDir, root)
	t.Setenv(EnvFFprobePath, "")
	t.Setenv(EnvFFmpegPath, "")
	t.Setenv("PATH", "")

	got, isFFmpeg, err := ResolveProbe("")
	if err != nil {
		t.Fatalf("ResolveProbe: %v", err)
	}
	if !isFFmpeg || !strings.EqualFold(got, ffmpegPath) {
		t.Fatalf("expected ffmpeg fallback %q, got %q (isFFmpeg=%v)", ffmpegPath, got, isFFmpeg)
	}
}

func TestCommandRejectsMissingExplicitPath(t *testing.T) {
	inIsolatedDir(t)
	missing := filepath.Join(t.TempDir(), "no-such-ffmpeg")
	t.Setenv(EnvFFmpegPath, missing)
	t.Setenv(EnvToolsDir, t.TempDir())
	t.Setenv("PATH", "")

	_, _, err := Command(context.Background(), missing)
	if err == nil || !IsMissing(err) {
		t.Fatalf("expected missing error, got %v", err)
	}
}

func TestRunReturnsMissengineMessageInsteadOfBareNotFound(t *testing.T) {
	inIsolatedDir(t)
	t.Setenv(EnvToolsDir, t.TempDir())
	t.Setenv(EnvFFmpegPath, "")
	t.Setenv("PATH", "")

	err := Run(context.Background(), "ffmpeg", "-version")
	if err == nil || !IsMissing(err) {
		t.Fatalf("expected missing error, got %v", err)
	}
	if !strings.Contains(err.Error(), "缺少视频合成引擎") {
		t.Fatalf("runtime error is not user facing: %s", err.Error())
	}
}
