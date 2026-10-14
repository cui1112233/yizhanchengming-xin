package mergeworker

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type CommandExecutor func(context.Context, string, ...string) error

type FFmpegRunner struct {
	Binary string
	Exec   CommandExecutor
}

func BuildFFmpegArgs(manifestPath, outputPath string, speed float64, aspectRatio ...string) ([]string, error) {
	if strings.TrimSpace(manifestPath) == "" || strings.TrimSpace(outputPath) == "" {
		return nil, fmt.Errorf("ffmpeg manifest and output are required")
	}
	if speed == 0 {
		speed = 1
	}
	if speed < 0.5 || speed > 4 {
		return nil, fmt.Errorf("speed must be between 0.5 and 4")
	}
	inputs, err := concatInputs(manifestPath)
	if err != nil {
		return nil, err
	}
	width, height, err := mergeCanvas(aspectRatio...)
	if err != nil {
		return nil, err
	}
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
	}
	for _, input := range inputs {
		args = append(args, "-i", input)
	}
	filters := make([]string, 0, len(inputs)*2+1)
	concatInputs := make([]string, 0, len(inputs)*2)
	speedValue := strconv.FormatFloat(speed, 'f', -1, 64)
	for index := range inputs {
		video := fmt.Sprintf("[%d:v:0]settb=AVTB,setpts=PTS-STARTPTS,fps=30,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black,setsar=1", index, width, height, width, height)
		if speed != 1 {
			video += ",setpts=PTS/" + speedValue
		}
		filters = append(filters, video+fmt.Sprintf("[v%d]", index))
		audio := fmt.Sprintf("[%d:a:0]aresample=48000,aformat=channel_layouts=stereo,asetpts=PTS-STARTPTS", index)
		if speed != 1 {
			audio += ",atempo=" + atempoFilter(speed)
		}
		filters = append(filters, audio+fmt.Sprintf("[a%d]", index))
		concatInputs = append(concatInputs, fmt.Sprintf("[v%d][a%d]", index, index))
	}
	filters = append(filters, strings.Join(concatInputs, "")+fmt.Sprintf("concat=n=%d:v=1:a=1[v][a]", len(inputs)))
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[v]", "-map", "[a]")
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "192k",
		"-movflags", "+faststart", outputPath,
	)
	return args, nil
}

func concatInputs(manifestPath string) ([]string, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	paths := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "file '") || !strings.HasSuffix(line, "'") {
			return nil, fmt.Errorf("invalid ffmpeg concat manifest entry")
		}
		path := strings.TrimSuffix(strings.TrimPrefix(line, "file '"), "'")
		path = strings.ReplaceAll(path, "'\\\\''", "'")
		if strings.TrimSpace(path) == "" {
			return nil, fmt.Errorf("invalid ffmpeg concat manifest entry")
		}
		paths = append(paths, path)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("ffmpeg inputs are required")
	}
	return paths, nil
}

func mergeCanvas(aspectRatio ...string) (int, int, error) {
	aspect := "9:16"
	if len(aspectRatio) > 0 && strings.TrimSpace(aspectRatio[0]) != "" {
		aspect = strings.TrimSpace(aspectRatio[0])
	}
	switch aspect {
	case "9:16":
		return 720, 1280, nil
	case "16:9":
		return 1280, 720, nil
	case "1:1":
		return 1080, 1080, nil
	default:
		return 0, 0, fmt.Errorf("unsupported merge aspect ratio %q", aspect)
	}
}

func atempoFilter(speed float64) string {
	parts := []string{}
	for speed > 2 {
		parts = append(parts, "2.0")
		speed /= 2
	}
	for speed < 0.5 {
		parts = append(parts, "0.5")
		speed /= 0.5
	}
	parts = append(parts, strconv.FormatFloat(speed, 'f', -1, 64))
	return strings.Join(parts, ",atempo=")
}

func (r *FFmpegRunner) Merge(ctx context.Context, inputPaths []string, outputPath string, speed float64) error {
	return r.merge(ctx, inputPaths, outputPath, speed, "")
}

// MergeWithAspect normalizes every source to the selected production canvas
// before concatenation. Video providers may return different pixel dimensions
// even when a book requested the same aspect ratio.
func (r *FFmpegRunner) MergeWithAspect(ctx context.Context, inputPaths []string, outputPath string, speed float64, aspectRatio string) error {
	return r.merge(ctx, inputPaths, outputPath, speed, aspectRatio)
}

// MergeWithAudio concatenates the video pictures while replacing all source
// audio with the supplied narration track. The narration is deliberately not
// tempo-adjusted: audio-following production already plans the picture speed
// against its measured duration.
func (r *FFmpegRunner) MergeWithAudio(ctx context.Context, inputPaths []string, audioPath, outputPath string, speed float64, aspectRatio string) error {
	if len(inputPaths) == 0 || strings.TrimSpace(audioPath) == "" {
		return fmt.Errorf("ffmpeg video and audio inputs are required")
	}
	workDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return err
	}
	manifestPath := filepath.Join(workDir, "concat.txt")
	manifest, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(manifest)
	for _, inputPath := range inputPaths {
		absolute, err := filepath.Abs(inputPath)
		if err != nil {
			_ = manifest.Close()
			return err
		}
		if _, err := fmt.Fprintf(writer, "file '%s'\n", escapeConcatPath(absolute)); err != nil {
			_ = manifest.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = manifest.Close()
		return err
	}
	if err := manifest.Close(); err != nil {
		return err
	}
	width, height, err := mergeCanvas(aspectRatio)
	if err != nil {
		return err
	}
	if speed == 0 {
		speed = 1
	}
	if speed < 0.5 || speed > 4 {
		return fmt.Errorf("speed must be between 0.5 and 4")
	}
	inputs, err := concatInputs(manifestPath)
	if err != nil {
		return err
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	for _, input := range inputs {
		args = append(args, "-i", input)
	}
	args = append(args, "-i", audioPath)
	filters := make([]string, 0, len(inputs)+2)
	concatVideos := make([]string, 0, len(inputs))
	speedValue := strconv.FormatFloat(speed, 'f', -1, 64)
	for index := range inputs {
		video := fmt.Sprintf("[%d:v:0]settb=AVTB,setpts=PTS-STARTPTS,fps=30,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black,setsar=1", index, width, height, width, height)
		if speed != 1 {
			video += ",setpts=PTS/" + speedValue
		}
		filters = append(filters, video+fmt.Sprintf("[v%d]", index))
		concatVideos = append(concatVideos, fmt.Sprintf("[v%d]", index))
	}
	filters = append(filters, strings.Join(concatVideos, "")+fmt.Sprintf("concat=n=%d:v=1:a=0[v]", len(inputs)))
	filters = append(filters, fmt.Sprintf("[%d:a:0]aresample=48000,aformat=channel_layouts=stereo,asetpts=PTS-STARTPTS[tts]", len(inputs)))
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[v]", "-map", "[tts]", "-shortest")
	args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "192k", "-movflags", "+faststart", outputPath)
	return r.run(ctx, args...)
}

// ConcatAudioFiles creates one continuous narration track from the measured
// per-line TTS files. It intentionally copies the encoded audio so the
// narration timing is not changed between planning and final mix.
func (r *FFmpegRunner) ConcatAudioFiles(ctx context.Context, inputPaths []string, outputPath string) error {
	if len(inputPaths) == 0 || strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("ffmpeg audio inputs are required")
	}
	workDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return err
	}
	manifestPath := filepath.Join(workDir, "audio-concat.txt")
	manifest, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(manifest)
	for _, inputPath := range inputPaths {
		absolute, err := filepath.Abs(inputPath)
		if err != nil {
			_ = manifest.Close()
			return err
		}
		if _, err := fmt.Fprintf(writer, "file '%s'\n", escapeConcatPath(absolute)); err != nil {
			_ = manifest.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		_ = manifest.Close()
		return err
	}
	if err := manifest.Close(); err != nil {
		return err
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", manifestPath, "-c", "copy", outputPath}
	return r.run(ctx, args...)
}

func (r *FFmpegRunner) merge(ctx context.Context, inputPaths []string, outputPath string, speed float64, aspectRatio string) error {
	if len(inputPaths) == 0 {
		return fmt.Errorf("ffmpeg inputs are required")
	}
	workDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(workDir, 0o700); err != nil {
		return err
	}
	manifestPath := filepath.Join(workDir, "concat.txt")
	manifest, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(manifest)
	for _, inputPath := range inputPaths {
		absolute, err := filepath.Abs(inputPath)
		if err != nil {
			manifest.Close()
			return err
		}
		if _, err := fmt.Fprintf(writer, "file '%s'\n", escapeConcatPath(absolute)); err != nil {
			manifest.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		manifest.Close()
		return err
	}
	if err := manifest.Close(); err != nil {
		return err
	}
	args, err := BuildFFmpegArgs(manifestPath, outputPath, speed, aspectRatio)
	if err != nil {
		return err
	}
	binary := strings.TrimSpace(r.Binary)
	if binary == "" {
		binary = "ffmpeg"
	}
	run := r.Exec
	if run == nil {
		run = func(ctx context.Context, name string, args ...string) error {
			cmd := exec.CommandContext(ctx, name, args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				message := strings.TrimSpace(string(output))
				if len(message) > 512 {
					message = message[:512]
				}
				if message == "" {
					return fmt.Errorf("ffmpeg failed: %w", err)
				}
				return fmt.Errorf("ffmpeg failed: %s", message)
			}
			return nil
		}
	}
	return run(ctx, binary, args...)
}

func (r *FFmpegRunner) run(ctx context.Context, args ...string) error {
	binary := strings.TrimSpace(r.Binary)
	if binary == "" {
		binary = "ffmpeg"
	}
	run := r.Exec
	if run == nil {
		run = func(ctx context.Context, name string, args ...string) error {
			cmd := exec.CommandContext(ctx, name, args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				message := strings.TrimSpace(string(output))
				if len(message) > 512 {
					message = message[:512]
				}
				if message == "" {
					return fmt.Errorf("ffmpeg failed: %w", err)
				}
				return fmt.Errorf("ffmpeg failed: %s", message)
			}
			return nil
		}
	}
	return run(ctx, binary, args...)
}

func escapeConcatPath(value string) string {
	return strings.ReplaceAll(value, "'", "'\\''")
}
