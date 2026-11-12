package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"strconv"
	"strings"
	"time"
)

// RunAICutAnalyzeSilence reads an owned source asset, decodes its actual audio,
// and reports source-time silence intervals. It never reads subtitles or saves edits.
func RunAICutAnalyzeSilence(ctx context.Context, store aicut.Store, objects agentworkspace.AssetObjectStore, owner, projectID, assetID string, options aicut.SilenceOptions) (aicut.SilenceAnalysis, error) {
	options, err := aicut.NormalizeSilenceOptions(options)
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	if store == nil || objects == nil {
		return aicut.SilenceAnalysis{}, errors.New("素材存储尚未配置")
	}
	if err := ctx.Err(); err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	asset, err := store.Asset(ctx, owner, projectID, assetID)
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	if asset.Kind != "audio" && asset.Kind != "video" {
		return aicut.SilenceAnalysis{}, errors.New("静音分析只支持音频或视频素材")
	}
	if asset.DurationMS > aicut.MaxSilenceDurationMS {
		return aicut.SilenceAnalysis{}, errors.New("单次静音分析最长支持 2 小时")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	dir, err := os.MkdirTemp("", "qiantie-aicut-silence-")
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "source"+path.Ext(asset.Name))
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	if err := ctx.Err(); err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	probe, _, err := ffmpegbin.Command(ctx, "ffprobe", "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=channels,start_time:format=duration,start_time", "-of", "json", input)
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	output, err := probe.Output()
	if err != nil {
		if ctx.Err() != nil {
			return aicut.SilenceAnalysis{}, ctx.Err()
		}
		return aicut.SilenceAnalysis{}, errors.New("无法读取素材音频信息")
	}
	var info struct {
		Streams []struct {
			Channels  int    `json:"channels"`
			StartTime string `json:"start_time"`
		} `json:"streams"`
		Format struct {
			Duration  string `json:"duration"`
			StartTime string `json:"start_time"`
		} `json:"format"`
	}
	if json.Unmarshal(output, &info) != nil || len(info.Streams) == 0 {
		return aicut.SilenceAnalysis{}, errors.New("素材没有音轨，无法进行静音检测")
	}
	// Decoded raw PCM starts at the first audio sample, while source cuts use
	// container time. Reject shifted streams until that mapping is implemented.
	startTime := func(value string) (float64, error) {
		if value == "" || value == "N/A" {
			return 0, nil
		}
		n, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, errors.New("素材音轨起始时间无效")
		}
		return n, nil
	}
	audioStart, err := startTime(info.Streams[0].StartTime)
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	formatStart, err := startTime(info.Format.StartTime)
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	if math.Abs(audioStart-formatStart) > 0.000001 {
		return aicut.SilenceAnalysis{}, errors.New("素材音轨与视频起始时间不一致，暂不支持自动静音剪辑，请先对齐音轨")
	}
	channels := info.Streams[0].Channels
	if channels < 1 || channels > 2 {
		return aicut.SilenceAnalysis{}, errors.New("静音分析暂支持单声道或双声道音频")
	}
	seconds, err := strconv.ParseFloat(info.Format.Duration, 64)
	if err != nil || seconds <= 0 || seconds > float64(aicut.MaxSilenceDurationMS)/1000 {
		return aicut.SilenceAnalysis{}, errors.New("音频时长无效或超过 2 小时")
	}
	const rate = 16000
	command, _, err := ffmpegbin.Command(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", input, "-map", "0:a:0", "-vn", "-t", "7200.020", "-ac", strconv.Itoa(channels), "-ar", strconv.Itoa(rate), "-f", "s16le", "-acodec", "pcm_s16le", "pipe:1")
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	defer stdout.Close()
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	analysis, analyzeErr := aicut.AnalyzeSilencePCM(ctx, stdout, rate, channels, options)
	// Stop the producer before Wait if the consumer exits early; otherwise a full
	// pipe can strand FFmpeg and its stdout reader indefinitely.
	if analyzeErr != nil {
		_ = stdout.Close()
		_ = command.Process.Kill()
	}
	waitErr := command.Wait()
	if err := ctx.Err(); err != nil {
		return aicut.SilenceAnalysis{}, err
	}
	if analyzeErr != nil {
		return aicut.SilenceAnalysis{}, analyzeErr
	}
	if waitErr != nil {
		return aicut.SilenceAnalysis{}, fmt.Errorf("静音音频解码失败：%s", strings.TrimSpace(compactAICutToolOutput(stderr.Bytes())))
	}
	analysis.AssetID = asset.ID
	return analysis, nil
}
