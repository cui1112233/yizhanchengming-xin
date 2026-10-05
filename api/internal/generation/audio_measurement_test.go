package generation

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeAudioProber struct {
	durationMS int64
	err        error
	calls      int
}

func (p *fakeAudioProber) ProbeDurationMS(_ context.Context, _ string) (int64, error) {
	p.calls++
	return p.durationMS, p.err
}

func TestServiceConstructionDoesNotRequireFFprobe(t *testing.T) {
	service, _, _ := serviceFixture()
	service.audioProber = FFprobeAudioProber{Binary: "/definitely/missing/ffprobe"}
	if err := service.validate(); err != nil {
		t.Fatalf("service startup must not require ffprobe: %v", err)
	}
}

func TestFFprobeUnavailableReturnsExplicitErrorOnlyWhenMeasuring(t *testing.T) {
	prober := FFprobeAudioProber{Binary: "/definitely/missing/ffprobe"}
	_, err := prober.ProbeDurationMS(context.Background(), "/tmp/audio.mp3")
	if !errors.Is(err, ErrAudioProbeUnavailable) {
		t.Fatalf("err = %v, want ErrAudioProbeUnavailable", err)
	}
	if !strings.Contains(err.Error(), "audio_probe_unavailable") {
		t.Fatalf("missing stable error code: %v", err)
	}
}

func TestMeasureAudioPersistsTraceableMeasurementAndReusesSameAsset(t *testing.T) {
	service, _, _ := serviceFixture()
	prober := &fakeAudioProber{durationMS: 28000}
	service.audioProber = prober

	first, err := service.MeasureAudio(context.Background(), AudioMeasurementRequest{
		BatchProjectID: 3,
		BookID:         11,
		AudioAsset:     "/srv/audio/book-11-v1.mp3",
	})
	if err != nil {
		t.Fatalf("MeasureAudio first: %v", err)
	}
	second, err := service.MeasureAudio(context.Background(), AudioMeasurementRequest{
		BatchProjectID: 3,
		BookID:         11,
		AudioAsset:     "/srv/audio/book-11-v1.mp3",
	})
	if err != nil {
		t.Fatalf("MeasureAudio second: %v", err)
	}
	if prober.calls != 1 {
		t.Fatalf("ffprobe calls = %d, want 1 for cached measurement", prober.calls)
	}
	if first.ID != second.ID || first.DurationMS != 28000 || first.MeasuredAt.IsZero() {
		t.Fatalf("measurement not persisted/reused: first=%#v second=%#v", first, second)
	}
	if first.BookID != 11 || first.BatchProjectID != 3 || first.AudioAsset == "" {
		t.Fatalf("measurement source is not traceable: %#v", first)
	}
}

func TestMatchAudioRequiresPersistedMeasurementAndDoesNotTrustRequestDuration(t *testing.T) {
	service, store, provider := serviceFixture()
	service.audioProber = &fakeAudioProber{durationMS: 28000}
	if _, err := service.MeasureAudio(context.Background(), AudioMeasurementRequest{BatchProjectID: 3, BookID: 11, AudioAsset: "/srv/audio/11.mp3"}); err != nil {
		t.Fatal(err)
	}
	provider.responses = []string{"SCRIPT", "HOOK", `{"cards":[{"shot":"a","start":0.00,"end":14.00},{"shot":"b","start":14.00,"end":28.00}]}`}

	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3,
		BookID: 11,
		HookEnabled: true,
		DirectorMode: DirectorNormal,
		MatchAudio: true,
		AudioDurationSec: 999.99,
		ShotDurationLimitSec: 15,
		RequestID: "authoritative-audio",
	})
	if err != nil {
		t.Fatalf("RunBook: %v", err)
	}
	if len(provider.calls) != 3 {
		t.Fatalf("provider calls = %d", len(provider.calls))
	}
	directorCall := provider.calls[2]
	if directorCall.AudioDurationSec != 28.00 {
		t.Fatalf("Director trusted request duration: got %.2f want 28.00", directorCall.AudioDurationSec)
	}
	director, err := store.LatestStageRun(context.Background(), result.Run.ID, StageDirector)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(director.InputSnapshot, `"audioDurationSec":28`) || !strings.Contains(director.ValidationResult, `"valid":true`) {
		t.Fatalf("authoritative duration/validation not recorded: %#v", director)
	}
}

func TestMatchAudioWithoutMeasurementFailsBeforeProviderExecution(t *testing.T) {
	service, _, provider := serviceFixture()
	_, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3,
		BookID: 11,
		HookEnabled: true,
		DirectorMode: DirectorNormal,
		MatchAudio: true,
		AudioDurationSec: 28,
		ShotDurationLimitSec: 15,
		RequestID: "no-audio",
	})
	if !errors.Is(err, ErrAudioMeasurementRequired) {
		t.Fatalf("err = %v, want ErrAudioMeasurementRequired", err)
	}
	if len(provider.calls) != 0 {
		t.Fatalf("provider must not run without authoritative audio measurement, calls=%d", len(provider.calls))
	}
}
