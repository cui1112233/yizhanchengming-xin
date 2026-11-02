package httpapi

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"qiantie/backend/internal/agentworkspace"
)

const (
	agentAudioWaveformSampleRate = 8000
	agentAudioWaveformPeakCount  = 96
	agentAudioWaveformMaxPCM     = 10 << 20
)

// AgentAudioWaveformSummary is deliberately bounded so the UI can draw a
// preview without receiving the protected audio payload.
type AgentAudioWaveformSummary struct {
	SampleRate int       `json:"sampleRate"`
	DurationMs int64     `json:"durationMs"`
	Peaks      []float64 `json:"peaks"`
}

type agentAudioWaveform interface {
	Summary(context.Context, agentworkspace.Asset, []byte) (AgentAudioWaveformSummary, error)
}

type agentAudioWaveformFunc func(context.Context, agentworkspace.Asset, []byte) (AgentAudioWaveformSummary, error)

func (fn agentAudioWaveformFunc) Summary(ctx context.Context, asset agentworkspace.Asset, contents []byte) (AgentAudioWaveformSummary, error) {
	return fn(ctx, asset, contents)
}

// FFmpegAgentAudioWaveform decodes a short, mono PCM representation before
// downsampling it into fixed-width display peaks.
type FFmpegAgentAudioWaveform struct {
	Binary string
}

func (p *FFmpegAgentAudioWaveform) Summary(ctx context.Context, asset agentworkspace.Asset, contents []byte) (AgentAudioWaveformSummary, error) {
	if asset.Kind != "audio" || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(asset.ContentType)), "audio/") || len(contents) == 0 {
		return AgentAudioWaveformSummary{}, agentworkspace.ErrInvalidAsset
	}
	workDir, err := os.MkdirTemp("", "agent-audio-waveform-*")
	if err != nil {
		return AgentAudioWaveformSummary{}, err
	}
	defer os.RemoveAll(workDir)
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(asset.Name)), ".")
	if extension == "" || len(extension) > 8 {
		extension = "audio"
	}
	source := filepath.Join(workDir, "source."+extension)
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		return AgentAudioWaveformSummary{}, err
	}
	binaryName := strings.TrimSpace(p.Binary)
	if binaryName == "" {
		binaryName = "ffmpeg"
	}
	command := exec.CommandContext(ctx, binaryName, "-nostdin", "-hide_banner", "-loglevel", "error", "-i", source, "-vn", "-ac", "1", "-ar", "8000", "-f", "f32le", "pipe:1")
	stdout, err := command.StdoutPipe()
	if err != nil {
		return AgentAudioWaveformSummary{}, err
	}
	if err := command.Start(); err != nil {
		return AgentAudioWaveformSummary{}, err
	}
	pcm, readErr := io.ReadAll(io.LimitReader(stdout, agentAudioWaveformMaxPCM+1))
	waitErr := command.Wait()
	if readErr != nil {
		return AgentAudioWaveformSummary{}, readErr
	}
	if len(pcm) > agentAudioWaveformMaxPCM {
		return AgentAudioWaveformSummary{}, fmt.Errorf("音频过长，无法生成波形预览")
	}
	if waitErr != nil {
		return AgentAudioWaveformSummary{}, fmt.Errorf("音频波形解析失败")
	}
	return summarizeAgentAudioPCM(pcm), nil
}

func summarizeAgentAudioPCM(pcm []byte) AgentAudioWaveformSummary {
	sampleCount := len(pcm) / 4
	peaks := make([]float64, agentAudioWaveformPeakCount)
	if sampleCount == 0 {
		return AgentAudioWaveformSummary{SampleRate: agentAudioWaveformSampleRate, Peaks: peaks}
	}
	for column := range peaks {
		start := column * sampleCount / len(peaks)
		end := (column + 1) * sampleCount / len(peaks)
		if end <= start {
			end = start + 1
		}
		if end > sampleCount {
			end = sampleCount
		}
		for index := start; index < end; index++ {
			value := float64(math.Float32frombits(binary.LittleEndian.Uint32(pcm[index*4 : index*4+4])))
			if !math.IsNaN(value) && !math.IsInf(value, 0) && math.Abs(value) > peaks[column] {
				peaks[column] = math.Min(1, math.Abs(value))
			}
		}
		peaks[column] = math.Round(peaks[column]*1000) / 1000
	}
	return AgentAudioWaveformSummary{SampleRate: agentAudioWaveformSampleRate, DurationMs: int64(sampleCount) * 1000 / agentAudioWaveformSampleRate, Peaks: peaks}
}
