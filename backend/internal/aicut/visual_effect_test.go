package aicut

import (
	"errors"
	"math"
	"testing"
)

func TestValidateTimelineAcceptsVisualEffect(t *testing.T) {
	timeline := Timeline{Tracks: []Track{{ID: "v1", Kind: TrackVideo, Clips: []Clip{{ID: "c1", Kind: ClipVideo, DurationMS: 1000, VisualEffect: &VisualEffect{Preset: "warm", Brightness: 0.1, Contrast: 1.1, Saturation: 1.2, Blur: 2}}}}}}
	if err := validateTimeline(timeline); err != nil {
		t.Fatalf("validateTimeline() error = %v", err)
	}
}

func TestValidateTimelineAcceptsFirstBatchEffectsAndTransitions(t *testing.T) {
	effects := []string{"original", "bright", "cool-cinema", "warm", "grayscale", "soft-focus", "vignette", "glow", "shake", "breathe", "flash", "mosaic", "custom"}
	transitions := []string{"cross-dissolve", "fade", "flash-black", "flash-white", "slide-left", "slide-right", "slide-up", "slide-down", "zoom", "blur"}
	for _, preset := range effects {
		timeline := Timeline{Tracks: []Track{{ID: "v1", Kind: TrackVideo, Clips: []Clip{{ID: "c1", Kind: ClipVideo, DurationMS: 1000, VisualEffect: &VisualEffect{Preset: preset, Contrast: 1, Saturation: 1, Intensity: .5, Speed: 1}}}}}}
		if err := validateTimeline(timeline); err != nil {
			t.Fatalf("preset %s rejected: %v", preset, err)
		}
	}
	for _, transition := range transitions {
		timeline := Timeline{Tracks: []Track{{ID: "v1", Kind: TrackVideo, Clips: []Clip{{ID: "c1", Kind: ClipVideo, DurationMS: 1000, Transition: transition, TransitionMS: 250}}}}}
		if err := validateTimeline(timeline); err != nil {
			t.Fatalf("transition %s rejected: %v", transition, err)
		}
	}
}

func TestValidateTimelineRejectsInvalidVisualEffect(t *testing.T) {
	cases := []VisualEffect{
		{Preset: "unknown", Contrast: 1, Saturation: 1},
		{Preset: "custom", Brightness: 1.1, Contrast: 1, Saturation: 1},
		{Preset: "custom", Contrast: .4, Saturation: 1},
		{Preset: "custom", Contrast: 1, Saturation: 3.1},
		{Preset: "custom", Contrast: 1, Saturation: 1, Blur: 21},
		{Preset: "custom", Brightness: math.NaN(), Contrast: 1, Saturation: 1},
		{Preset: "shake", Contrast: 1, Saturation: 1, Intensity: 1.1, Speed: 1},
		{Preset: "shake", Contrast: 1, Saturation: 1, Intensity: .5, Speed: .05},
	}
	for _, effect := range cases {
		timeline := Timeline{Tracks: []Track{{ID: "v1", Kind: TrackVideo, Clips: []Clip{{ID: "c1", Kind: ClipVideo, DurationMS: 1000, VisualEffect: &effect}}}}}
		if err := validateTimeline(timeline); !errors.Is(err, ErrInvalidTimeline) {
			t.Fatalf("effect=%+v error=%v, want ErrInvalidTimeline", effect, err)
		}
	}
}

func TestValidateTimelineAcceptsEffectTrackAndCanvasTransform(t *testing.T) {
	timeline := Timeline{Tracks: []Track{
		{ID: "fx1", Kind: TrackEffect, Clips: []Clip{{ID: "e1", Kind: ClipEffect, Name: "发光", StartMS: 100, DurationMS: 900, ResourceID: "effect-glow", Parameters: map[string]float64{"intensity": .5}}}},
		{ID: "v1", Kind: TrackVideo, Clips: []Clip{{ID: "c1", Kind: ClipVideo, DurationMS: 1000, CanvasTransform: &CanvasTransform{X: .2, Y: -.1, ScaleX: 1.2, ScaleY: .8, Rotation: 12, Opacity: .7}}}},
	}}
	if err := validateTimeline(timeline); err != nil {
		t.Fatalf("validateTimeline() error = %v", err)
	}
}

func TestValidateTimelineAcceptsChineseEditorTransitionIDs(t *testing.T) {
	for _, transition := range []string{"transition-cross-dissolve", "transition-push-left", "transition-wipe-right", "transition-glitch", "transition-light-leak"} {
		timeline := Timeline{Tracks: []Track{{ID: "v1", Kind: TrackVideo, Clips: []Clip{{ID: "c1", Kind: ClipVideo, DurationMS: 1000, Transition: transition, TransitionMS: 250}}}}}
		if err := validateTimeline(timeline); err != nil {
			t.Fatalf("transition %s rejected: %v", transition, err)
		}
	}
}

func TestValidateTimelinePreservesCatalogResourceID(t *testing.T) {
	timeline := Timeline{Tracks: []Track{{ID: "v1", Kind: TrackVideo, Clips: []Clip{{
		ID: "c1", Kind: ClipVideo, AssetID: "asset-1", DurationMS: 1000, CatalogResourceID: "filter-warm-sun",
	}}}}}
	if err := validateTimeline(timeline); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTimelinePreservesTextResourceID(t *testing.T) {
	timeline := Timeline{Tracks: []Track{{ID: "s1", Kind: TrackSubtitle, Settings: TrackSettings{TextResourceID: "text-comic"}, Clips: []Clip{{
		ID: "c1", Kind: ClipSubtitle, DurationMS: 1000, Text: "测试", TextResourceID: "text-comic",
	}}}}}
	if err := validateTimeline(timeline); err != nil {
		t.Fatal(err)
	}
}
