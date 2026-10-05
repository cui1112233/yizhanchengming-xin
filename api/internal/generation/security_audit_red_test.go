package generation

import (
	"context"
	"errors"
	"testing"
)

func TestSecurityAuditMeasureAudioRejectsRemoteURLBeforeFFprobe(t *testing.T) {
	service, _, _ := serviceFixture()
	prober := &fakeAudioProber{durationMS: 1000}
	service.audioProber = prober

	_, err := service.MeasureAudio(context.Background(), AudioMeasurementRequest{
		BatchProjectID: 3,
		BookID:         11,
		AudioAsset:     "http://169.254.169.254/latest/meta-data/",
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("err=%v, want ErrInvalid for browser-controlled remote audio URL", err)
	}
	if prober.calls != 0 {
		t.Fatalf("ffprobe calls=%d, want 0 before SSRF-capable URL reaches ffprobe", prober.calls)
	}
}
