package generation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type AudioProber interface {
	ProbeDurationMS(context.Context, string) (int64, error)
}

type FFprobeAudioProber struct {
	Binary string
}

func (p FFprobeAudioProber) ProbeDurationMS(ctx context.Context, audioAsset string) (int64, error) {
	binary := strings.TrimSpace(p.Binary)
	if binary == "" {
		binary = "ffprobe"
	}
	path, err := exec.LookPath(binary)
	if err != nil {
		return 0, fmt.Errorf("%w: ffprobe executable not found", ErrAudioProbeUnavailable)
	}
	cmd := exec.CommandContext(ctx, path,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		audioAsset,
	)
	out, err := cmd.Output()
	if err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return 0, fmt.Errorf("%w: ffprobe executable unavailable", ErrAudioProbeUnavailable)
		}
		return 0, fmt.Errorf("%w: audio_probe_failed", ErrUnavailable)
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 0, fmt.Errorf("%w: invalid audio duration", ErrUnavailable)
	}
	return int64(math.Round(seconds * 1000)), nil
}

func audioAssetHash(asset string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(asset)))
	return hex.EncodeToString(sum[:])
}

func (s *Service) MeasureAudio(ctx context.Context, req AudioMeasurementRequest) (AudioMeasurement, error) {
	if err := s.validate(); err != nil {
		return AudioMeasurement{}, err
	}
	asset := strings.TrimSpace(req.AudioAsset)
	if req.BatchProjectID <= 0 || req.BookID <= 0 || asset == "" {
		return AudioMeasurement{}, ErrInvalid
	}
	if _, err := s.store.GetBookForProject(ctx, req.BatchProjectID, req.BookID); err != nil {
		return AudioMeasurement{}, err
	}
	hash := audioAssetHash(asset)
	if existing, err := s.store.AudioMeasurementByAsset(ctx, req.BatchProjectID, req.BookID, hash); err == nil {
		if existing.DurationMS > 0 {
			return existing, nil
		}
	} else if !errors.Is(err, ErrNotFound) {
		return AudioMeasurement{}, err
	}
	if s.audioProber == nil {
		return AudioMeasurement{}, fmt.Errorf("%w: no audio prober configured", ErrAudioProbeUnavailable)
	}
	durationMS, err := s.audioProber.ProbeDurationMS(ctx, asset)
	if err != nil {
		return AudioMeasurement{}, err
	}
	if durationMS <= 0 {
		return AudioMeasurement{}, fmt.Errorf("%w: invalid measured audio duration", ErrUnavailable)
	}
	measuredAt := s.now().UTC()
	return s.store.CreateAudioMeasurement(ctx, AudioMeasurement{
		BatchProjectID: req.BatchProjectID,
		BookID:         req.BookID,
		AudioAsset:     asset,
		AssetHash:      hash,
		DurationMS:     durationMS,
		MeasuredAt:     measuredAt,
	})
}

func (s *Service) AudioMeasurement(ctx context.Context, projectID, bookID int64) (AudioMeasurement, error) {
	if err := s.validate(); err != nil {
		return AudioMeasurement{}, err
	}
	if projectID <= 0 || bookID <= 0 {
		return AudioMeasurement{}, ErrInvalid
	}
	return s.store.LatestAudioMeasurement(ctx, projectID, bookID)
}
