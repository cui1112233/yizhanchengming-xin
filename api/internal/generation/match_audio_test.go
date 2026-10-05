package generation

import (
	"errors"
	"strings"
	"testing"
)

func TestMatchAudioTimelineRepairsSmallClosingDriftToExactlyTwentyEightSeconds(t *testing.T) {
	input := `{"cards":[{"shot":"开门","start":0.00,"end":7.25},{"shot":"对白","start":7.25,"end":18.40},{"shot":"停顿","start":18.40,"end":27.96}]}`
	got, result, err := validateAndRepairDirectorOutput(input, 28000, 15000)
	if err != nil {
		t.Fatalf("validateAndRepairDirectorOutput: %v", err)
	}
	if !result.Repaired || result.DurationMS != 28000 {
		t.Fatalf("result = %#v", result)
	}
	for _, want := range []string{`"start":0.00`, `"end":28.00`} {
		if !strings.Contains(got, want) {
			t.Fatalf("repaired output missing %s: %s", want, got)
		}
	}
}

func TestMatchAudioRepairOnlyChangesTimingAndNeverAddsStoryContent(t *testing.T) {
	input := `{"cards":[{"shot":"她推开门","dialogue":"你来了","start":0.00,"end":9.00},{"shot":"他放下信封","narration":"房间安静下来","start":9.00,"end":18.00},{"shot":"她看向窗外","start":18.00,"end":27.96}]}`
	got, result, err := validateAndRepairDirectorOutput(input, 28000, 15000)
	if err != nil {
		t.Fatalf("validateAndRepairDirectorOutput: %v", err)
	}
	if !result.Repaired {
		t.Fatal("expected closing repair")
	}
	if strings.Count(got, `"shot"`) != 3 {
		t.Fatalf("repair changed shot count: %s", got)
	}
	for _, original := range []string{"她推开门", "你来了", "他放下信封", "房间安静下来", "她看向窗外"} {
		if !strings.Contains(got, original) {
			t.Fatalf("repair lost original story content %q: %s", original, got)
		}
	}
}

func TestMatchAudioTimelineRejectsSevereGapOverlapAndLargeDurationDrift(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"gap", `{"cards":[{"shot":"a","start":0.00,"end":5.00},{"shot":"b","start":8.00,"end":28.00}]}`},
		{"overlap", `{"cards":[{"shot":"a","start":0.00,"end":12.00},{"shot":"b","start":5.00,"end":28.00}]}`},
		{"large drift", `{"cards":[{"shot":"a","start":0.00,"end":10.00},{"shot":"b","start":10.00,"end":20.00}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateAndRepairDirectorOutput(tt.body, 28000, 15000)
			if !errors.Is(err, ErrTimelineValidation) {
				t.Fatalf("err = %v, want ErrTimelineValidation", err)
			}
		})
	}
}

func TestMatchAudioTimelineRejectsInvalidStructuresAndHardShotLimits(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		limit int64
	}{
		{"empty", `{"cards":[]}`, 15000},
		{"missing start", `{"cards":[{"shot":"a","end":10.00}]}`, 15000},
		{"missing end", `{"cards":[{"shot":"a","start":0.00}]}`, 15000},
		{"reversed", `{"cards":[{"shot":"a","start":8.00,"end":5.00}]}`, 15000},
		{"over fifteen", `{"cards":[{"shot":"a","start":0.00,"end":16.00},{"shot":"b","start":16.00,"end":28.00}]}`, 15000},
		{"over ten", `{"cards":[{"shot":"a","start":0.00,"end":11.00},{"shot":"b","start":11.00,"end":20.00}]}`, 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateAndRepairDirectorOutput(tt.body, 28000, tt.limit)
			if !errors.Is(err, ErrTimelineValidation) {
				t.Fatalf("err = %v, want ErrTimelineValidation", err)
			}
		})
	}
}

func TestFifteenSecondModeAllowsShotsShorterThanFifteenSeconds(t *testing.T) {
	input := `{"cards":[{"shot":"a","start":0.00,"end":7.00},{"shot":"b","start":7.00,"end":18.00},{"shot":"c","start":18.00,"end":28.00}]}`
	_, result, err := validateAndRepairDirectorOutput(input, 28000, 15000)
	if err != nil {
		t.Fatalf("15 second mode must be a maximum, not fixed slices: %v", err)
	}
	if result.DurationMS != 28000 {
		t.Fatalf("duration = %d", result.DurationMS)
	}
}

func TestMatchAudioTimelineUsesIntegerMillisecondsWithoutFloatDrift(t *testing.T) {
	input := `{"cards":[{"shot":"a","start":0.00,"end":9.33},{"shot":"b","start":9.33,"end":18.66},{"shot":"c","start":18.66,"end":28.00}]}`
	got, result, err := validateAndRepairDirectorOutput(input, 28000, 15000)
	if err != nil {
		t.Fatal(err)
	}
	if result.DurationMS != 28000 || strings.Contains(got, "27.999") {
		t.Fatalf("integer timeline drifted: result=%#v output=%s", result, got)
	}
}

func TestTimelineRepairToleranceIsCentralizedAndSmall(t *testing.T) {
	if TimelineRepairToleranceMS <= 0 || TimelineRepairToleranceMS > 500 {
		t.Fatalf("TimelineRepairToleranceMS = %d; want a small explicit tolerance", TimelineRepairToleranceMS)
	}
}

func TestH3DirectorMatchAudioTimelineIsValidatedToo(t *testing.T) {
	input := `{"schema_version":"h3-director/v1","director_cards":[{"shot":"a","start":0.00,"end":14.00},{"shot":"b","start":14.00,"end":28.00}]}`
	got, result, err := validateAndRepairDirectorOutput(input, 28000, 15000)
	if err != nil {
		t.Fatalf("H3 matchAudio validation: %v", err)
	}
	if result.DurationMS != 28000 || !strings.Contains(got, `"end":28.00`) {
		t.Fatalf("H3 output not preserved/validated: %s %#v", got, result)
	}
}
