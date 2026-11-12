package aicut

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestColorGradePersistsThroughTimelineStoreWithoutAliasing(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	now := time.Now()
	if err := store.CreateProject(ctx, "alice", Project{ID: "color-grade", Title: "Color", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	grade := ColorAdjustment{RGBCurves: map[string][]ColorCurvePoint{"red": {{X: 0, Y: 0}, {X: .5, Y: 1}, {X: 1, Y: 1}}}, Wheels: map[string]ColorWheelAdjustment{"shadows": {Hue: 120, Saturation: 25, Luminance: 10}}}
	timeline := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{ID: "c", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &VisualProcessing{Color: grade}}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", "color-grade", timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	grade.RGBCurves["red"][1].Y = 0
	grade.Wheels["shadows"] = ColorWheelAdjustment{}
	loaded, err := store.Timeline(ctx, "alice", "color-grade")
	if err != nil {
		t.Fatal(err)
	}
	got := loaded.Tracks[0].Clips[0].VisualProcessing.Color
	if got.RGBCurves["red"][1].Y != 1 || got.Wheels["shadows"].Saturation != 25 {
		t.Fatalf("grade lost or aliased %v", got)
	}
}

func TestColorGradeSharedReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("../../../frontend/src/user/pages/aicut/colorGrade.reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Pixel [4]uint8        `json:"pixel"`
		Color ColorAdjustment `json:"color"`
		Want  [4]uint8        `json:"want"`
	}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, v := range vectors {
		if got := ApplyColorGradeToRGBA(v.Pixel, v.Color); got != v.Want {
			t.Errorf("got %v want %v color=%v", got, v.Want, v.Color)
		}
	}
}

func TestColorGradeCurvesWheelsAndAlpha(t *testing.T) {
	for _, vector := range []struct {
		pixel, want [4]uint8
		color       ColorAdjustment
	}{
		{pixel: [4]uint8{64, 100, 200, 77}, want: [4]uint8{128, 100, 200, 77}, color: ColorAdjustment{RGBCurves: map[string][]ColorCurvePoint{"red": {{X: 0, Y: 0}, {X: .5, Y: 1}, {X: 1, Y: 1}}}}},
		{pixel: [4]uint8{128, 128, 128, 77}, want: [4]uint8{140, 125, 125, 77}, color: ColorAdjustment{Wheels: map[string]ColorWheelAdjustment{"shadows": {Saturation: 100}}}},
		{pixel: [4]uint8{0, 0, 0, 255}, want: [4]uint8{255, 255, 255, 255}, color: ColorAdjustment{Wheels: map[string]ColorWheelAdjustment{"shadows": {Luminance: 100}}}},
		{pixel: [4]uint8{255, 255, 255, 255}, want: [4]uint8{0, 0, 0, 255}, color: ColorAdjustment{Wheels: map[string]ColorWheelAdjustment{"highlights": {Luminance: -100}}}},
	} {
		if got := ApplyColorGradeToRGBA(vector.pixel, vector.color); got != vector.want {
			t.Errorf("got %v want %v", got, vector.want)
		}
		vector.pixel[3] = 0
		if got := ApplyColorGradeToRGBA(vector.pixel, vector.color); got != vector.pixel {
			t.Fatal("transparent source changed")
		}
	}
	if HasColorGradeAdjustments(ColorAdjustment{Wheels: map[string]ColorWheelAdjustment{"shadows": {Hue: 120}}}) {
		t.Fatal("unused hue is not a grade")
	}
}

func TestColorGradeValidation(t *testing.T) {
	for _, color := range []ColorAdjustment{
		{RGBCurves: map[string][]ColorCurvePoint{"unknown": {{X: 0, Y: 0}, {X: 1, Y: 1}}}},
		{RGBCurves: map[string][]ColorCurvePoint{"red": {{X: 0, Y: 0}, {X: 0, Y: 1}}}},
		{Wheels: map[string]ColorWheelAdjustment{"unknown": {Saturation: 10}}},
		{Wheels: map[string]ColorWheelAdjustment{"shadows": {Saturation: -1}}},
		{Wheels: map[string]ColorWheelAdjustment{"midtones": {Hue: 181}}},
	} {
		if validVisualProcessing(&VisualProcessing{Color: color}) {
			t.Errorf("invalid color grade accepted: %v", color)
		}
	}
}
