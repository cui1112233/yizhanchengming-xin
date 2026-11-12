package aicut

import (
	"strings"
	"testing"
)

func TestClipAnimationMySQLEditorMetadataRoundTrip(t *testing.T) {
	original := Clip{ID: "clip", Animation: &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "fade", DurationMS: 600}, Exit: ClipAnimationPhase{Preset: "slide", DurationMS: 400}, Loop: ClipAnimationPhase{Preset: "pulse", DurationMS: 2000}}, BlendMode: "screen"}
	encoded, err := encodeClipEditorJSON(original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"animation"`) {
		t.Fatalf("SQL editor metadata dropped animation: %s", encoded)
	}
	var first, second Clip
	if err = decodeClipEditorJSON(encoded, &first); err != nil {
		t.Fatal(err)
	}
	if err = decodeClipEditorJSON(encoded, &second); err != nil {
		t.Fatal(err)
	}
	if first.Animation == nil || *first.Animation != *original.Animation || first.BlendMode != "screen" {
		t.Fatalf("animation metadata mismatch %+v", first)
	}
	original.Animation.Entrance.DurationMS = 900
	first.Animation.Exit.Preset = "zoom"
	if second.Animation.Entrance.DurationMS != 600 || second.Animation.Exit.Preset != "slide" {
		t.Fatal("metadata decoder aliased animation between clips")
	}
	if err = decodeClipEditorJSON([]byte(`{}`), &first); err != nil {
		t.Fatal(err)
	}
	if first.Animation != nil {
		t.Fatal("legacy metadata must clear animation to neutral")
	}
	legacy, err := encodeClipEditorJSON(Clip{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(legacy), `"animation"`) {
		t.Fatal("old SQL clip acquired animation")
	}
}

func TestClipAnimationCloneTimelineDoesNotAlias(t *testing.T) {
	animation := &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "fade", DurationMS: 500}}
	original := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{ID: "c", Kind: ClipVideo, DurationMS: 1000, Animation: animation}}}}}
	copied := cloneTimeline(original)
	copied.Tracks[0].Clips[0].Animation.Entrance.DurationMS = 200
	if animation.Entrance.DurationMS != 500 {
		t.Fatal("memory clone aliased animation")
	}
	job := cloneExportJob(ExportJob{TimelineSnapshot: original})
	job.TimelineSnapshot.Tracks[0].Clips[0].Animation.Entrance.Preset = "zoom"
	if animation.Entrance.Preset != "fade" {
		t.Fatal("export snapshot aliased animation")
	}
}
