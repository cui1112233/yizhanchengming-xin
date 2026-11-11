package aicut

import (
	"errors"
	"testing"
)

func TestValidateTimelineAcceptsMarkersAndKeyframes(t *testing.T) {
	volume := 0.75
	timeline := Timeline{
		Markers: []TimelineMarker{{ID: "marker-1", TimeMS: 500, Label: "节拍 1", Color: "#22c55e", Source: "audio-beat", AssetID: "audio-1"}},
		Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{
			ID: "c", Kind: ClipVideo, DurationMS: 2000, SourceDurationMS: 2000,
			Keyframes: []ClipKeyframe{{ID: "k1", OffsetMS: 500, Values: KeyframeValues{X: floatPtr(.2), Volume: &volume}}},
		}}}},
	}
	if err := validateTimeline(timeline); err != nil {
		t.Fatalf("validateTimeline() error = %v", err)
	}
}

func TestValidateTimelineAcceptsAndRejectsSpeedCurves(t *testing.T) {
	valid := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{
		ID: "c", Kind: ClipVideo, DurationMS: 1600, SourceDurationMS: 4000,
		SpeedCurve: []SpeedCurvePoint{{Position: 0, Rate: .5}, {Position: .5, Rate: 2}, {Position: 1, Rate: .5}},
	}}}}}
	if err := validateTimeline(valid); err != nil {
		t.Fatalf("valid speed curve: %v", err)
	}
	invalid := valid
	invalid.Tracks[0].Clips[0].SpeedCurve = []SpeedCurvePoint{{Position: 0, Rate: 1}, {Position: .7, Rate: 5}, {Position: 1, Rate: 1}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("expected invalid speed curve")
	}
}

func TestValidateTimelineChecksConsumedSourceSpanForStaticAndCurveSpeed(t *testing.T) {
	slowCurve := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{
		ID: "slow", Kind: ClipVideo, DurationMS: 2000, SourceStartMS: 1000, SourceDurationMS: 2500,
		PlaybackRate: .5, SpeedCurve: []SpeedCurvePoint{{Position: 0, Rate: .5}, {Position: 1, Rate: .5}},
	}}}}}
	if err := validateTimeline(slowCurve); err != nil {
		t.Fatalf("slow curve only consumes 1000ms of source: %v", err)
	}
	tooFast := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{
		ID: "fast", Kind: ClipVideo, DurationMS: 1000, SourceStartMS: 0, SourceDurationMS: 1500, PlaybackRate: 2,
	}}}}}
	if err := validateTimeline(tooFast); !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("2x clip consumes 2000ms source, error=%v", err)
	}
}

func TestValidateTimelineRejectsInvalidMarkersAndKeyframes(t *testing.T) {
	invalidMarker := Timeline{Markers: []TimelineMarker{{ID: "", TimeMS: 0}}, Tracks: []Track{}}
	if err := validateTimeline(invalidMarker); !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("marker error=%v", err)
	}
	invalidFrame := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{ID: "c", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, Keyframes: []ClipKeyframe{{ID: "k", OffsetMS: 1200}}}}}}}
	if err := validateTimeline(invalidFrame); !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("keyframe error=%v", err)
	}
}

func TestClipEditorJSONRoundTripPreservesAutomationAndCatalogFields(t *testing.T) {
	x := .3
	clip := Clip{CatalogResourceID: "filter-bright", ResourceID: "effect-glow", CanvasTransform: &CanvasTransform{ScaleX: 1, ScaleY: 1, Opacity: 1}, Keyframes: []ClipKeyframe{{ID: "k", OffsetMS: 200, Values: KeyframeValues{X: &x}}}, SpeedCurve: []SpeedCurvePoint{{Position: 0, Rate: 1}, {Position: 1, Rate: 2}}}
	encoded, err := encodeClipEditorJSON(clip)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Clip
	if err := decodeClipEditorJSON(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.CatalogResourceID != clip.CatalogResourceID || len(decoded.Keyframes) != 1 || decoded.Keyframes[0].Values.X == nil || *decoded.Keyframes[0].Values.X != x || len(decoded.SpeedCurve) != 2 || decoded.SpeedCurve[1].Rate != 2 {
		t.Fatalf("decoded=%+v", decoded)
	}
}

func floatPtr(value float64) *float64 { return &value }
