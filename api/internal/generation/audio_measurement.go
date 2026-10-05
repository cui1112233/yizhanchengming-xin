package generation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const AudioDurationQuantumMS int64 = 10

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

func canonicalAudioDurationMS(durationMS int64) int64 {
	if durationMS <= 0 {
		return durationMS
	}
	return ((durationMS + AudioDurationQuantumMS/2) / AudioDurationQuantumMS) * AudioDurationQuantumMS
}

func audioAssetHash(asset string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(asset)))
	return hex.EncodeToString(sum[:])
}

func safeLocalAudioAsset(asset string) bool {
	asset = strings.TrimSpace(asset)
	if asset == "" || strings.HasPrefix(asset, "-") || strings.ContainsRune(asset, '\x00') {
		return false
	}
	parsed, err := url.Parse(asset)
	if err != nil {
		return false
	}
	// Service-internal callers may pass an absolute server-materialized temp
	// file. Network protocols are never accepted by ffprobe.
	return parsed.Scheme == "" && parsed.Host == ""
}

// BrowserAudioAssetAllowed is the HTTP trust boundary for audio measurement.
// A browser may identify only a server-managed relative asset; it may not make
// ffprobe open an arbitrary absolute/traversal path on the server.
func BrowserAudioAssetAllowed(asset string) bool {
	asset = strings.TrimSpace(asset)
	if !safeLocalAudioAsset(asset) {
		return false
	}
	if filepath.IsAbs(asset) || strings.HasPrefix(asset, "/") || strings.Contains(asset, "\\") {
		return false
	}
	for _, segment := range strings.Split(filepath.ToSlash(asset), "/") {
		if segment == ".." {
			return false
		}
	}
	cleaned := filepath.ToSlash(filepath.Clean(asset))
	return cleaned != "." && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}

func (s *Service) MeasureAudio(ctx context.Context, req AudioMeasurementRequest) (AudioMeasurement, error) {
	if err := s.validate(); err != nil {
		return AudioMeasurement{}, err
	}
	asset := strings.TrimSpace(req.AudioAsset)
	if req.BatchProjectID <= 0 || req.BookID <= 0 || !safeLocalAudioAsset(asset) {
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
	durationMS = canonicalAudioDurationMS(durationMS)
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
