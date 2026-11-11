package batchfactoryv11

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

const maxVideoDurationProbeBytes = 512 << 20
const videoDurationProbeAttempts = 3

// HTTPVideoDurationProbe downloads a completed HTTPS video to a bounded
// stream, then measures the real container duration with the built-in
// MP4/WAV/MP3/FLAC/ADTS parser or an external ffprobe when available.
type HTTPVideoDurationProbe struct {
	Client     *http.Client
	RetryDelay func(attempt int) time.Duration
}

func (p HTTPVideoDurationProbe) DurationSeconds(ctx context.Context, rawURL string) (float64, error) {
	var lastErr error
	for attempt := 0; attempt < videoDurationProbeAttempts; attempt++ {
		seconds, err := p.durationSecondsOnce(ctx, rawURL)
		if err == nil {
			return seconds, nil
		}
		lastErr = err
		if attempt+1 == videoDurationProbeAttempts || !retryableVideoDurationProbeError(ctx, err) {
			break
		}
		delay := 250 * time.Millisecond * time.Duration(attempt+1)
		if p.RetryDelay != nil {
			delay = p.RetryDelay(attempt + 1)
		}
		if delay <= 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(delay):
		}
	}
	return 0, fmt.Errorf("media duration probe failed after %d attempts: %w", videoDurationProbeAttempts, lastErr)
}

func (p HTTPVideoDurationProbe) durationSecondsOnce(ctx context.Context, rawURL string) (float64, error) {
	endpoint, err := validateProductionURL(strings.TrimSpace(rawURL))
	if err != nil {
		return 0, fmt.Errorf("invalid completed media URL: %w", err)
	}
	client := p.Client
	if client == nil {
		client = safehttp.NewClientWithTimeout(3 * time.Minute)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("media download returned HTTP %d", response.StatusCode)
	}
	video, err := io.ReadAll(io.LimitReader(response.Body, maxVideoDurationProbeBytes+1))
	if err != nil {
		return 0, err
	}
	if len(video) == 0 || len(video) > maxVideoDurationProbeBytes {
		return 0, fmt.Errorf("media duration probe requires 1-%d bytes", maxVideoDurationProbeBytes)
	}
	durationMS, err := (FFprobeH3AudioDuration{}).DurationMSExact(ctx, video)
	if err != nil {
		return 0, err
	}
	seconds := float64(durationMS) / 1000
	if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 0, fmt.Errorf("media duration probe returned invalid duration")
	}
	return seconds, nil
}

func retryableVideoDurationProbeError(ctx context.Context, err error) bool {
	if err == nil || ctx.Err() != nil {
		return false
	}
	message := strings.ToLower(err.Error())
	// Bad source URLs and a permanent client-side rejection cannot recover by
	// reading the exact same object again. Transport failures, stream resets,
	// and 5xx/429 CDN responses can.
	if strings.Contains(message, "invalid completed media url") || strings.Contains(message, "media duration probe requires") || strings.Contains(message, "returned invalid duration") || strings.Contains(message, "media download returned http 4") {
		return false
	}
	return true
}
