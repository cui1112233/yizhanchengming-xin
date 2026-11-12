package aicut

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const MaxSilenceDurationMS int64 = 2 * 60 * 60 * 1000

type SilenceOptions struct {
	MinSilenceMS int64   `json:"minSilenceMs"`
	ThresholdDB  float64 `json:"thresholdDb"`
	PaddingMS    int64   `json:"paddingMs"`
}
type SilenceInterval struct {
	StartMS int64 `json:"startMs"`
	EndMS   int64 `json:"endMs"`
}
type SilenceAnalysis struct {
	AssetID    string            `json:"assetId"`
	DurationMS int64             `json:"durationMs"`
	SampleRate int               `json:"sampleRate"`
	Channels   int               `json:"channels"`
	WindowMS   int               `json:"windowMs"`
	Options    SilenceOptions    `json:"options"`
	Intervals  []SilenceInterval `json:"intervals"`
	AllSilent  bool              `json:"allSilent"`
}

func NormalizeSilenceOptions(o SilenceOptions) (SilenceOptions, error) {
	if o.MinSilenceMS == 0 {
		o.MinSilenceMS = 500
	}
	if o.ThresholdDB == 0 {
		o.ThresholdDB = -40
	}
	if o.MinSilenceMS < 100 || o.MinSilenceMS > 60000 || math.IsNaN(o.ThresholdDB) || math.IsInf(o.ThresholdDB, 0) || o.ThresholdDB < -90 || o.ThresholdDB > -10 || o.PaddingMS < 0 || o.PaddingMS > 2000 {
		return o, errors.New("静音参数无效：最短静音 100–60000 毫秒，阈值 -90 至 -10 dB，保留边缘 0–2000 毫秒")
	}
	return o, nil
}

// AnalyzeSilencePCM reads interleaved signed 16-bit little-endian PCM. Each 20ms
// window uses the loudest channel RMS so opposite-phase stereo cannot cancel.
func AnalyzeSilencePCM(ctx context.Context, r io.Reader, sampleRate, channels int, opts SilenceOptions) (SilenceAnalysis, error) {
	opts, err := NormalizeSilenceOptions(opts)
	if err != nil {
		return SilenceAnalysis{}, err
	}
	if r == nil || sampleRate < 1000 || sampleRate > 192000 || channels < 1 || channels > 8 {
		return SilenceAnalysis{}, errors.New("音频采样参数无效")
	}
	a := SilenceAnalysis{SampleRate: sampleRate, Channels: channels, WindowMS: 20, Options: opts, Intervals: []SilenceInterval{}, AllSilent: true}
	framesPerWindow := sampleRate / 50
	buf := make([]byte, framesPerWindow*channels*2)
	var frames int64
	silentStart := int64(-1)
	threshold := math.Pow(10, opts.ThresholdDB/20)
	flush := func(end int64) {
		if silentStart >= 0 && end-silentStart >= opts.MinSilenceMS {
			start, stop := silentStart+opts.PaddingMS, end-opts.PaddingMS
			if stop > start {
				a.Intervals = append(a.Intervals, SilenceInterval{start, stop})
			}
		}
		silentStart = -1
	}
	for {
		if err := ctx.Err(); err != nil {
			return SilenceAnalysis{}, err
		}
		n, readErr := io.ReadFull(r, buf)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return SilenceAnalysis{}, readErr
		}
		if n == 0 {
			break
		}
		if n%(channels*2) != 0 {
			return SilenceAnalysis{}, errors.New("音频采样不完整")
		}
		count := n / (channels * 2)
		loudest := 0.0
		for ch := 0; ch < channels; ch++ {
			sum := 0.0
			for i := 0; i < count; i++ {
				v := float64(int16(binary.LittleEndian.Uint16(buf[(i*channels+ch)*2:]))) / 32768
				sum += v * v
			}
			loudest = math.Max(loudest, math.Sqrt(sum/float64(count)))
		}
		start := frames * 1000 / int64(sampleRate)
		frames += int64(count)
		end := frames * 1000 / int64(sampleRate)
		if end > MaxSilenceDurationMS {
			return SilenceAnalysis{}, errors.New("单次静音分析最长支持 2 小时")
		}
		if loudest <= threshold {
			if silentStart < 0 {
				silentStart = start
			}
		} else {
			a.AllSilent = false
			flush(start)
		}
		if readErr != nil {
			break
		}
	}
	if frames == 0 {
		return SilenceAnalysis{}, errors.New("素材没有可分析的音频采样")
	}
	a.DurationMS = frames * 1000 / int64(sampleRate)
	flush(a.DurationMS)
	return a, nil
}
