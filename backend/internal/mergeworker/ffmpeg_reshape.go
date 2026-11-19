package mergeworker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"qiantie/backend/internal/ffmpegbin"
)

// 片段重拍需要的两件底层能力：从成片里抽帧做时间轴预览，以及把新片段拼回原视频。
// 两者都走项目统一的 ffmpeg 解析（ffmpegbin），与合并成片共用同一个可执行文件。

// ExtractFrames 均匀抽取 count 张预览帧，写到 outputDir，返回按顺序排列的文件路径。
func (r *FFmpegRunner) ExtractFrames(ctx context.Context, inputPath string, duration float64, count, width int, outputDir string) ([]string, error) {
	if strings.TrimSpace(inputPath) == "" {
		return nil, fmt.Errorf("ffmpeg input is required")
	}
	if count <= 0 {
		return nil, fmt.Errorf("frame count must be positive")
	}
	if duration <= 0 {
		return nil, fmt.Errorf("video duration is required")
	}
	if width <= 0 {
		width = 480
	}
	if err := os.MkdirAll(outputDir, 0o700); err != nil {
		return nil, err
	}
	fps := float64(count) / duration
	pattern := filepath.Join(outputDir, "frame-%03d.jpg")
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-i", inputPath,
		"-vf", fmt.Sprintf("fps=%s,scale=%d:-2", strconv.FormatFloat(fps, 'f', 6, 64), width),
		"-frames:v", strconv.Itoa(count),
		"-q:v", "3",
		pattern,
	}
	if err := r.run(ctx, args...); err != nil {
		return nil, err
	}
	frames := make([]string, 0, count)
	for index := 1; index <= count; index++ {
		path := filepath.Join(outputDir, fmt.Sprintf("frame-%03d.jpg", index))
		if _, err := os.Stat(path); err != nil {
			break
		}
		frames = append(frames, path)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("no preview frames were produced")
	}
	return frames, nil
}

// ExtractFrame 在指定时间点抓一张图，用于首帧/尾帧参考。
func (r *FFmpegRunner) ExtractFrame(ctx context.Context, inputPath string, at float64, outputPath string) error {
	if strings.TrimSpace(inputPath) == "" || strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("ffmpeg input and output are required")
	}
	if at < 0 {
		at = 0
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return err
	}
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-ss", strconv.FormatFloat(at, 'f', 3, 64),
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		outputPath,
	}
	return r.run(ctx, args...)
}

type mediaInfo struct {
	Duration float64
	HasAudio bool
}

// ProbeMedia 读取时长与是否含音轨；没有 ffprobe 时退回 ffmpeg 的解析结果。
func ProbeMedia(ctx context.Context, path string) (mediaInfo, error) {
	binary, isFFmpeg, err := ffmpegbin.ResolveProbe("")
	if err != nil {
		return mediaInfo{}, err
	}
	if isFFmpeg {
		return probeWithFFmpeg(ctx, binary, path)
	}
	args := []string{"-v", "error", "-show_entries", "stream=codec_type:format=duration", "-of", "json", path}
	output, err := runCapture(ctx, binary, args...)
	if err != nil {
		return mediaInfo{}, err
	}
	var parsed struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(output, &parsed); err != nil {
		return mediaInfo{}, err
	}
	info := mediaInfo{}
	for _, stream := range parsed.Streams {
		if stream.CodecType == "audio" {
			info.HasAudio = true
			break
		}
	}
	info.Duration, _ = strconv.ParseFloat(strings.TrimSpace(parsed.Format.Duration), 64)
	if info.Duration <= 0 {
		return mediaInfo{}, fmt.Errorf("media duration is unknown")
	}
	return info, nil
}

// probeWithFFmpeg 在没有 ffprobe 时用 ffmpeg -i 的输出估算时长，音轨信息保持保守（当作有）。
func probeWithFFmpeg(ctx context.Context, binary, path string) (mediaInfo, error) {
	output, _ := runCapture(ctx, binary, "-hide_banner", "-i", path)
	text := string(output)
	info := mediaInfo{HasAudio: strings.Contains(text, "Audio:")}
	const marker = "Duration:"
	at := strings.Index(text, marker)
	if at < 0 {
		return mediaInfo{}, fmt.Errorf("media duration is unknown")
	}
	line := strings.TrimSpace(strings.SplitN(text[at+len(marker):], ",", 2)[0])
	parts := strings.Split(line, ":")
	if len(parts) != 3 {
		return mediaInfo{}, fmt.Errorf("media duration is unknown")
	}
	hours, _ := strconv.ParseFloat(parts[0], 64)
	minutes, _ := strconv.ParseFloat(parts[1], 64)
	seconds, _ := strconv.ParseFloat(parts[2], 64)
	info.Duration = hours*3600 + minutes*60 + seconds
	if info.Duration <= 0 {
		return mediaInfo{}, fmt.Errorf("media duration is unknown")
	}
	return info, nil
}

// Splice 把原视频的保留前段、新片段、保留后段按顺序拼成一条新视频。
// 先把两段素材统一到同一画布与音轨规格，再按时间裁剪拼接，避免分辨率或音轨缺失导致失败。
func (r *FFmpegRunner) Splice(ctx context.Context, originalPath string, start, end float64, insertPath, outputPath, aspectRatio string) error {
	if strings.TrimSpace(originalPath) == "" || strings.TrimSpace(insertPath) == "" || strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("ffmpeg inputs and output are required")
	}
	width, height, err := mergeCanvas(aspectRatio)
	if err != nil {
		return err
	}
	original, err := ProbeMedia(ctx, originalPath)
	if err != nil {
		return fmt.Errorf("读取原视频信息失败：%w", err)
	}
	if end > original.Duration {
		end = original.Duration
	}
	if start < 0 {
		start = 0
	}
	if end-start < 0.05 {
		return fmt.Errorf("原片段范围无效")
	}
	insert, err := ProbeMedia(ctx, insertPath)
	if err != nil {
		return fmt.Errorf("读取新片段信息失败：%w", err)
	}
	workDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return err
	}
	normalizedOriginal := filepath.Join(workDir, "reshoot-original.mp4")
	normalizedInsert := filepath.Join(workDir, "reshoot-insert.mp4")
	if err := r.normalizeSegment(ctx, originalPath, normalizedOriginal, width, height, original.HasAudio); err != nil {
		return fmt.Errorf("原视频标准化失败：%w", err)
	}
	if err := r.normalizeSegment(ctx, insertPath, normalizedInsert, width, height, insert.HasAudio); err != nil {
		return fmt.Errorf("新片段标准化失败：%w", err)
	}
	filters, err := buildSpliceFilter(original.Duration, start, end)
	if err != nil {
		return err
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	args = append(args, "-i", normalizedOriginal, "-i", normalizedInsert)
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[v]", "-map", "[a]")
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "192k",
		"-movflags", "+faststart", outputPath,
	)
	return r.run(ctx, args...)
}

// buildSpliceFilter 生成「前段 + 新片段 + 后段」的 filter_complex，段数为 1-3。
func buildSpliceFilter(originalDuration, start, end float64) ([]string, error) {
	if originalDuration <= 0 {
		return nil, fmt.Errorf("原视频时长未知")
	}
	if end > originalDuration {
		end = originalDuration
	}
	if start < 0 {
		start = 0
	}
	if end-start < 0.05 {
		return nil, fmt.Errorf("原片段范围无效")
	}
	const tolerance = 0.05
	segments := make([]string, 0, 3)
	labels := make([]string, 0, 3)
	appendSegment := func(video, audio, label string) {
		segments = append(segments, video+";"+audio)
		labels = append(labels, label)
	}
	if start > tolerance {
		appendSegment(
			fmt.Sprintf("[0:v]trim=start=0:end=%s,setpts=PTS-STARTPTS[v0]", ffSeconds(start)),
			fmt.Sprintf("[0:a]atrim=start=0:end=%s,asetpts=PTS-STARTPTS[a0]", ffSeconds(start)),
			"[v0][a0]",
		)
	}
	appendSegment("[1:v]setpts=PTS-STARTPTS[v1]", "[1:a]asetpts=PTS-STARTPTS[a1]", "[v1][a1]")
	if originalDuration-end > tolerance {
		appendSegment(
			fmt.Sprintf("[0:v]trim=start=%s:end=%s,setpts=PTS-STARTPTS[v2]", ffSeconds(end), ffSeconds(originalDuration)),
			fmt.Sprintf("[0:a]atrim=start=%s:end=%s,asetpts=PTS-STARTPTS[a2]", ffSeconds(end), ffSeconds(originalDuration)),
			"[v2][a2]",
		)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("没有可拼接的片段")
	}
	filters := append([]string{}, segments...)
	filters = append(filters, strings.Join(labels, "")+fmt.Sprintf("concat=n=%d:v=1:a=1[v][a]", len(labels)))
	return filters, nil
}

// runCapture 执行命令并返回合并输出；探测类命令即使非零退出也需要输出内容。
func runCapture(ctx context.Context, binary string, args ...string) ([]byte, error) {
	command, path, err := ffmpegbin.Command(ctx, binary, args...)
	if err != nil {
		return nil, err
	}
	output, runErr := command.CombinedOutput()
	if runErr != nil {
		return output, fmt.Errorf("ffmpeg 执行失败（%s）：%w", path, runErr)
	}
	return output, nil
}

func ffSeconds(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}

// normalizeSegment 把任意来源统一为 30fps、目标画布、48k 立体声，缺音轨时补静音。
func (r *FFmpegRunner) normalizeSegment(ctx context.Context, inputPath, outputPath string, width, height int, hasAudio bool) error {
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", inputPath}
	filter := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black,setsar=1,fps=30", width, height, width, height)
	if hasAudio {
		args = append(args, "-map", "0:v:0", "-map", "0:a:0", "-vf", filter,
			"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
			"-c:a", "aac", "-b:a", "192k", "-ar", "48000", "-ac", "2", "-shortest", outputPath)
		return r.run(ctx, args...)
	}
	args = append(args, "-f", "lavfi", "-i", "anullsrc=r=48000:cl=stereo",
		"-map", "0:v:0", "-map", "1:a:0", "-vf", filter,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "192k", "-shortest", outputPath)
	return r.run(ctx, args...)
}
