package mergeworker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qiantie/backend/internal/ffmpegbin"
)

func TestBuildSpliceFilterSegments(t *testing.T) {
	filters, err := buildSpliceFilter(12, 3, 5.5)
	if err != nil {
		t.Fatalf("中段重拍应可拼接：%v", err)
	}
	joined := strings.Join(filters, ";")
	for _, want := range []string{"trim=start=0:end=3.000", "trim=start=5.500:end=12.000", "concat=n=3:v=1:a=1[v][a]"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("缺少 %q：%s", want, joined)
		}
	}
	filters, err = buildSpliceFilter(12, 0, 4)
	if err != nil || !strings.Contains(strings.Join(filters, ";"), "concat=n=2") {
		t.Fatalf("起点为 0 时应只拼两段：%v %v", filters, err)
	}
	filters, err = buildSpliceFilter(12, 8, 12)
	if err != nil || !strings.Contains(strings.Join(filters, ";"), "concat=n=2") {
		t.Fatalf("到结尾时应只拼两段：%v %v", filters, err)
	}
	if _, err = buildSpliceFilter(12, 4, 4); err == nil {
		t.Fatal("空片段应拒绝")
	}
}

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skipf("本机没有 ffmpeg：%v", err)
	}
}

// 用两段不同颜色的测试视频验证抽帧与拼接，避免依赖任何模型接口。
func TestReshootExtractFramesAndSplice(t *testing.T) {
	requireFFmpeg(t)
	ctx := context.Background()
	directory := t.TempDir()
	source := filepath.Join(directory, "source.mp4")
	insert := filepath.Join(directory, "insert.mp4")
	runner := &FFmpegRunner{}
	if err := runner.run(ctx, "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=blue:s=320x568:d=6:r=30",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6",
		"-shortest", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", source); err != nil {
		t.Fatalf("生成测试视频失败：%v", err)
	}
	// 新片段刻意不带音轨，验证补静音后仍能拼接。
	if err := runner.run(ctx, "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=red:s=320x568:d=2:r=25",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", insert); err != nil {
		t.Fatalf("生成新片段失败：%v", err)
	}
	frames, err := runner.ExtractFrames(ctx, source, 6, 20, 240, filepath.Join(directory, "frames"))
	if err != nil {
		t.Fatalf("抽帧失败：%v", err)
	}
	if len(frames) < 18 || len(frames) > 21 {
		t.Fatalf("20 帧时间轴数量异常：%d", len(frames))
	}
	output := filepath.Join(directory, "out.mp4")
	if err := runner.Splice(ctx, source, 2, 4, insert, output, "9:16"); err != nil {
		t.Fatalf("拼接失败：%v", err)
	}
	info, err := ProbeMedia(ctx, output)
	if err != nil {
		t.Fatalf("读取成片失败：%v", err)
	}
	// 原 6 秒去掉 2-4 秒，加 2 秒新片段，应约等于 6 秒。
	if info.Duration < 5.2 || info.Duration > 6.8 {
		t.Fatalf("成片时长异常：%.2f（预期约 6 秒）", info.Duration)
	}
	stat, err := os.Stat(output)
	if err != nil || stat.Size() == 0 {
		t.Fatalf("成片文件为空：%v", err)
	}
}

func TestExtractFrameAtPosition(t *testing.T) {
	requireFFmpeg(t)
	ctx := context.Background()
	directory := t.TempDir()
	source := filepath.Join(directory, "source.mp4")
	runner := &FFmpegRunner{}
	if err := runner.run(ctx, "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=green:s=160x284:d=3:r=24",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", source); err != nil {
		t.Fatalf("生成测试视频失败：%v", err)
	}
	target := filepath.Join(directory, "first.jpg")
	if err := runner.ExtractFrame(ctx, source, 1.5, target); err != nil {
		t.Fatalf("抽单帧失败：%v", err)
	}
	stat, err := os.Stat(target)
	if err != nil || stat.Size() < 100 {
		t.Fatalf("首帧图片异常：%v", err)
	}
}
