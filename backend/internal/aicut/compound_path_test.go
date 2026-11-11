package aicut

import (
	"encoding/json"
	"testing"
)

func TestCompoundPathAndInternalTrackSettingsValidateAndRoundTrip(t *testing.T) {
	clip := Clip{
		ID: "clip-1", Kind: ClipVideo, Name: "镜头", DurationMS: 1000, SourceDurationMS: 1000,
		CompoundID: "parent", CompoundName: "上层",
		CompoundPath: []CompoundRef{{ID: "child", Name: "子层"}, {ID: "parent", Name: "上层"}},
	}
	timeline := Timeline{Tracks: []Track{
		{ID: "video", Kind: TrackVideo, Name: "画面", Clips: []Clip{clip}},
		{ID: "inside-audio", Kind: TrackAudio, Name: "组内声音", Settings: TrackSettings{CompoundID: "parent", CompoundName: "上层"}, Clips: []Clip{}},
	}}
	if err := validateTimeline(timeline); err != nil {
		t.Fatalf("valid nested compound timeline rejected: %v", err)
	}
	payload, err := encodeClipEditorJSON(clip)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Clip
	if err := decodeClipEditorJSON(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.CompoundPath) != 2 || decoded.CompoundPath[0].ID != "child" || decoded.CompoundPath[1].ID != "parent" {
		t.Fatalf("compound path lost: %s", payload)
	}
	settings, _ := json.Marshal(timeline.Tracks[1].Settings)
	if string(settings) != `{"compoundId":"parent","compoundName":"上层"}` {
		t.Fatalf("internal track settings lost: %s", settings)
	}
}

func TestCompoundPathValidationRejectsInconsistentOrRepeatedPath(t *testing.T) {
	base := Clip{ID: "clip-1", Kind: ClipVideo, Name: "镜头", DurationMS: 1000, SourceDurationMS: 1000}
	for _, clip := range []Clip{
		{ID: base.ID, Kind: base.Kind, Name: base.Name, DurationMS: base.DurationMS, SourceDurationMS: base.SourceDurationMS, CompoundID: "wrong", CompoundPath: []CompoundRef{{ID: "child", Name: "子层"}}},
		{ID: base.ID, Kind: base.Kind, Name: base.Name, DurationMS: base.DurationMS, SourceDurationMS: base.SourceDurationMS, CompoundID: "child", CompoundPath: []CompoundRef{{ID: "child"}, {ID: "child"}}},
	} {
		if err := validateTimeline(Timeline{Tracks: []Track{{ID: "video", Kind: TrackVideo, Clips: []Clip{clip}}}}); err != ErrInvalidTimeline {
			t.Fatalf("invalid compound path accepted: %#v, err=%v", clip.CompoundPath, err)
		}
	}
}
