package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"qiantie/backend/internal/ffmpegbin"
)

const maxH3AudioBytes = 64 << 20

type H3AudioDurationProbe interface {
	DurationMS(context.Context, []byte) (int64, error)
}

// FFprobeH3AudioDuration 测量媒体时长，但不再强制依赖系统 PATH 里的 ffprobe：
//  1. 优先用纯 Go 内置解析（WAV/MP3/M4A(MP4)/FLAC/AAC-ADTS），跨平台一致；
//  2. 内置解析不支持时，尝试显式配置或 PATH 中的 ffprobe/ffmpeg；
//  3. 仍然拿不到时返回按字节数/码率估算的时长并标记 Estimated，
//     让配音与逐行时长阶段继续执行而不是整条任务失败。
type FFprobeH3AudioDuration struct {
	// Binary 为显式指定的 ffprobe 路径；为空时依次读取
	// QIANTIE_FFPROBE_PATH、QIANTIE_FFMPEG_PATH 与 PATH。
	Binary string
}

var ffmpegDurationPattern = regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+(?:\.\d+)?)`)

func (p FFprobeH3AudioDuration) Measure(ctx context.Context, audio []byte) (AudioDuration, error) {
	if len(audio) == 0 {
		return AudioDuration{}, fmt.Errorf("%w: audio bytes are required", ErrInvalid)
	}
	if duration, ok := parseBuiltinAudioDuration(audio); ok && duration.MS > 0 {
		return duration, nil
	}
	if duration, err := p.measureExternal(ctx, audio); err == nil && duration.MS > 0 {
		return duration, nil
	}
	return estimateAudioDurationByBytes(audio), nil
}

func (p FFprobeH3AudioDuration) DurationMS(ctx context.Context, audio []byte) (int64, error) {
	duration, err := p.Measure(ctx, audio)
	if err != nil {
		return 0, err
	}
	return duration.MS, nil
}

// DurationMSExact 只接受精确时长（内置解析或外部探针），不做估算。
// 视频时长探针使用它，避免用字节估算值驱动最终的音画合并。
func (p FFprobeH3AudioDuration) DurationMSExact(ctx context.Context, audio []byte) (int64, error) {
	if len(audio) == 0 {
		return 0, fmt.Errorf("%w: media bytes are required", ErrInvalid)
	}
	if duration, ok := parseBuiltinAudioDuration(audio); ok && duration.MS > 0 {
		return duration.MS, nil
	}
	duration, err := p.measureExternal(ctx, audio)
	if err != nil {
		return 0, err
	}
	return duration.MS, nil
}

func (p FFprobeH3AudioDuration) measureExternal(ctx context.Context, audio []byte) (AudioDuration, error) {
	binaryPath, isFFmpeg, resolveErr := ffmpegbin.ResolveProbe(p.Binary)
	if resolveErr != nil {
		return AudioDuration{}, fmt.Errorf("%w: 无法测量媒体时长：内置解析不支持该格式（已支持 WAV/MP3/M4A/MP4/FLAC/AAC-ADTS），且未找到 ffprobe/ffmpeg。%v", ErrInvalid, resolveErr)
	}
	file, err := os.CreateTemp("", "qiantie-media-probe-*")
	if err != nil {
		return AudioDuration{}, err
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(audio); err != nil {
		file.Close()
		return AudioDuration{}, err
	}
	if err := file.Close(); err != nil {
		return AudioDuration{}, err
	}
	if isFFmpeg {
		// ffmpeg 在没有指定输出时会以非零码退出，这里只解析稳定的 Duration 行。
		command, _, cmdErr := ffmpegbin.Command(ctx, binaryPath, "-hide_banner", "-i", path)
		if cmdErr != nil {
			return AudioDuration{}, fmt.Errorf("%w: %v", ErrInvalid, cmdErr)
		}
		out, _ := command.CombinedOutput()
		match := ffmpegDurationPattern.FindSubmatch(out)
		if len(match) != 4 {
			return AudioDuration{}, fmt.Errorf("%w: ffmpeg 未能解析媒体时长（%s）", ErrInvalid, binaryPath)
		}
		hours, hourErr := strconv.ParseFloat(string(match[1]), 64)
		minutes, minuteErr := strconv.ParseFloat(string(match[2]), 64)
		seconds, secondErr := strconv.ParseFloat(string(match[3]), 64)
		if hourErr != nil || minuteErr != nil || secondErr != nil {
			return AudioDuration{}, fmt.Errorf("%w: audio probe returned invalid duration", ErrInvalid)
		}
		durationMS := int64(math.Round((hours*3600 + minutes*60 + seconds) * 1000))
		if durationMS <= 0 {
			return AudioDuration{}, fmt.Errorf("%w: audio probe returned zero duration", ErrInvalid)
		}
		return AudioDuration{MS: durationMS, Source: "external:ffmpeg"}, nil
	}
	command, _, cmdErr := ffmpegbin.Command(ctx, binaryPath, "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	if cmdErr != nil {
		return AudioDuration{}, fmt.Errorf("%w: ffprobe 执行失败：%v", ErrInvalid, cmdErr)
	}
	out, err := command.Output()
	if err != nil {
		return AudioDuration{}, fmt.Errorf("%w: ffprobe 执行失败（%s）：%v", ErrInvalid, binaryPath, err)
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return AudioDuration{}, fmt.Errorf("%w: audio probe returned invalid duration", ErrInvalid)
	}
	durationMS := int64(math.Round(seconds * 1000))
	if durationMS <= 0 {
		return AudioDuration{}, fmt.Errorf("%w: audio probe returned zero duration", ErrInvalid)
	}
	return AudioDuration{MS: durationMS, Source: "external:ffprobe"}, nil
}

// locateProbeBinary 返回可用的外部探针路径，以及该路径是否为 ffmpeg。
func locateProbeBinary(explicit string) (string, bool) {
	path, isFFmpeg, err := ffmpegbin.ResolveProbe(explicit)
	if err != nil {
		return "", false
	}
	return path, isFFmpeg
}

func (s *H3KernelService) MeasureAudio(ctx context.Context, owner, batchID, bookID string, audio []byte) (H3AudioMeasurementRevision, error) {
	if s == nil || s.Store == nil {
		return H3AudioMeasurementRevision{}, ErrUnavailable
	}
	if len(audio) == 0 || len(audio) > maxH3AudioBytes {
		return H3AudioMeasurementRevision{}, fmt.Errorf("%w: audio bytes must be between 1 and %d bytes", ErrInvalid, maxH3AudioBytes)
	}
	repository, ok := s.Store.(H3Repository)
	if !ok {
		return H3AudioMeasurementRevision{}, ErrUnavailable
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return H3AudioMeasurementRevision{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return H3AudioMeasurementRevision{}, err
	}
	if book.DirectorRevision == nil || book.DirectorRevision.Output.H3Director == nil {
		return H3AudioMeasurementRevision{}, fmt.Errorf("%w: active H3 director revision is required", ErrConflict)
	}
	probe := s.AudioProbe
	if probe == nil {
		probe = FFprobeH3AudioDuration{}
	}
	durationMS, err := probe.DurationMS(ctx, audio)
	if err != nil {
		return H3AudioMeasurementRevision{}, err
	}
	if durationMS <= 0 {
		return H3AudioMeasurementRevision{}, fmt.Errorf("%w: probed audio duration must be greater than zero", ErrInvalid)
	}
	digest := sha256.Sum256(audio)
	contentHash := hex.EncodeToString(digest[:])
	document := book.DirectorRevision.Output.H3Director
	return repository.PersistH3AudioMeasurement(ctx, owner, batchID, bookID, H3AudioMeasurement{
		AssetID:             "h3-audio-" + contentHash[:24],
		ContentHash:         contentHash,
		DurationMS:          durationMS,
		VideoSourceRevision: document.VideoSourceRevision,
		VideoSourceHash:     document.VideoSourceHash,
	})
}
