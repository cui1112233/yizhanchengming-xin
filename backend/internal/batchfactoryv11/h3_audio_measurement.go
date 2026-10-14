package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const maxH3AudioBytes = 64 << 20

type H3AudioDurationProbe interface {
	DurationMS(context.Context, []byte) (int64, error)
}

type FFprobeH3AudioDuration struct {
	Binary string
}

var ffmpegDurationPattern = regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+(?:\.\d+)?)`)

func (p FFprobeH3AudioDuration) DurationMS(ctx context.Context, audio []byte) (int64, error) {
	if len(audio) == 0 {
		return 0, fmt.Errorf("%w: audio bytes are required", ErrInvalid)
	}
	file, err := os.CreateTemp("", "qiantie-h3-audio-*")
	if err != nil {
		return 0, err
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(audio); err != nil {
		file.Close()
		return 0, err
	}
	if err := file.Close(); err != nil {
		return 0, err
	}
	binary := strings.TrimSpace(p.Binary)
	if binary == "" {
		binary = "ffprobe"
	}
	out, err := exec.CommandContext(ctx, binary, "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path).Output()
	if err != nil {
		// The local Windows bundle may ship ffmpeg without a separate ffprobe.
		// Use ffmpeg's stable Duration line as a compatible fallback instead of
		// making the whole automation stop before video generation.
		fallback, lookErr := exec.LookPath("ffmpeg")
		if lookErr != nil {
			return 0, fmt.Errorf("%w: audio probe failed: %v", ErrInvalid, err)
		}
		probeOut, probeErr := exec.CommandContext(ctx, fallback, "-hide_banner", "-i", path, "-f", "null", "-").CombinedOutput()
		if probeErr != nil && len(probeOut) == 0 {
			return 0, fmt.Errorf("%w: audio probe failed: %v", ErrInvalid, err)
		}
		match := ffmpegDurationPattern.FindSubmatch(probeOut)
		if len(match) != 4 {
			return 0, fmt.Errorf("%w: audio probe failed: %v", ErrInvalid, err)
		}
		hours, hourErr := strconv.ParseFloat(string(match[1]), 64)
		minutes, minuteErr := strconv.ParseFloat(string(match[2]), 64)
		seconds, secondErr := strconv.ParseFloat(string(match[3]), 64)
		if hourErr != nil || minuteErr != nil || secondErr != nil {
			return 0, fmt.Errorf("%w: audio probe returned invalid duration", ErrInvalid)
		}
		return int64(math.Round((hours*3600 + minutes*60 + seconds) * 1000)), nil
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 0, fmt.Errorf("%w: audio probe returned invalid duration", ErrInvalid)
	}
	durationMS := int64(math.Round(seconds * 1000))
	if durationMS <= 0 {
		return 0, fmt.Errorf("%w: audio probe returned zero duration", ErrInvalid)
	}
	return durationMS, nil
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
