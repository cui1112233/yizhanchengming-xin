package generation

import (
	"context"
	"strings"
	"testing"
)

func TestMatchAudioRetryReexecutesDirectorOnlyWithAuthoritativeMeasurement(t *testing.T) {
	service, store, provider := serviceFixture()
	service.audioProber = &fakeAudioProber{durationMS: 28000}
	if _, err := service.MeasureAudio(context.Background(), AudioMeasurementRequest{
		BatchProjectID: 3,
		BookID:         11,
		AudioAsset:     "/audio/book-11.mp3",
	}); err != nil {
		t.Fatal(err)
	}

	provider.responses = []string{
		"SCRIPT",
		"HOOK",
		`{"cards":[{"shot":"bad","start":0.00,"end":20.00}]}`,
	}
	failed, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID:       3,
		BookID:               11,
		HookEnabled:          true,
		DirectorMode:         DirectorNormal,
		MatchAudio:           true,
		AudioDurationSec:     999.99,
		ShotDurationLimitSec: 15,
		RequestID:            "match-audio-fail",
	})
	if err == nil {
		t.Fatal("expected Director validation failure")
	}
	if len(provider.calls) != 3 {
		t.Fatalf("initial provider calls = %d", len(provider.calls))
	}
	badDirector, err := store.LatestStageRun(context.Background(), failed.Run.ID, StageDirector)
	if err != nil {
		t.Fatal(err)
	}
	if badDirector.Status != StatusFailed || !strings.Contains(badDirector.ValidationResult, `"valid":false`) {
		t.Fatalf("failed Director validation not recorded: %#v", badDirector)
	}

	provider.responses = append(provider.responses,
		`{"cards":[{"shot":"a","start":0.00,"end":14.00},{"shot":"b","start":14.00,"end":28.00}]}`,
	)
	callsBefore := len(provider.calls)
	result, err := service.RetryStage(context.Background(), RetryStageRequest{
		BatchProjectID: 3,
		BookID:         11,
		Stage:          StageDirector,
		RequestID:      "match-audio-retry",
	})
	if err != nil {
		t.Fatalf("RetryStage: %v", err)
	}
	if len(provider.calls) != callsBefore+1 {
		t.Fatalf("retry must call Director only once: %d -> %d", callsBefore, len(provider.calls))
	}
	if provider.calls[len(provider.calls)-1].AudioDurationSec != 28.00 {
		t.Fatalf("retry lost authoritative duration: %.2f", provider.calls[len(provider.calls)-1].AudioDurationSec)
	}

	scriptRuns, hookRuns, directorRuns := 0, 0, 0
	for _, stage := range result.Stages {
		switch stage.Stage {
		case StageScript:
			scriptRuns++
		case StageHook:
			hookRuns++
		case StageDirector:
			directorRuns++
		}
	}
	if scriptRuns != 1 || hookRuns != 1 || directorRuns != 2 {
		t.Fatalf("retry reran prerequisites: script=%d hook=%d director=%d", scriptRuns, hookRuns, directorRuns)
	}
	latest, err := store.LatestStageRun(context.Background(), failed.Run.ID, StageDirector)
	if err != nil {
		t.Fatal(err)
	}
	if latest.Status != StatusCompleted || !strings.Contains(latest.ValidationResult, `"valid":true`) {
		t.Fatalf("retried Director not completed/validated: %#v", latest)
	}
}
