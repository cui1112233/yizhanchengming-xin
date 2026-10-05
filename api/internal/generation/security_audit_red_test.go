package generation

import (
	"context"
	"errors"
	"strings"
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

func TestSecurityAuditProviderFailureDoesNotPersistOrReturnRawUpstreamDetail(t *testing.T) {
	service, store, provider := serviceFixture()
	const upstreamDetail = "provider payload failed at /srv/private/provider-request.json"
	provider.errors[0] = errors.New(upstreamDetail)

	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3,
		BookID:         11,
		DirectorMode:   DirectorNormal,
		RequestID:      "security-redaction",
	})
	if err == nil {
		t.Fatal("expected provider failure")
	}
	if strings.Contains(result.Error, upstreamDetail) || strings.Contains(result.Run.ErrorMessage, upstreamDetail) {
		t.Fatalf("raw provider failure leaked in result: %#v", result)
	}
	if len(store.stageRuns) == 0 {
		t.Fatal("expected failed stage audit row")
	}
	if strings.Contains(store.stageRuns[0].ErrorMessage, upstreamDetail) {
		t.Fatalf("raw provider failure persisted in browser-facing stage error: %#v", store.stageRuns[0])
	}
}
