package aicut

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHSLSharedReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("../../../frontend/src/user/pages/aicut/hslColor.reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Pixel    [4]uint8                     `json:"pixel"`
		Settings map[string]HSLBandAdjustment `json:"settings"`
		Want     [4]uint8                     `json:"want"`
	}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, vector := range vectors {
		if got := ApplyHSLToRGBA(vector.Pixel, vector.Settings); got != vector.Want {
			t.Errorf("pixel %v settings %v: got %v want %v", vector.Pixel, vector.Settings, got, vector.Want)
		}
	}
}

func TestHSLValidationAndJSONPersistence(t *testing.T) {
	processing := &VisualProcessing{Color: ColorAdjustment{HSL: map[string]HSLBandAdjustment{"orange": {Hue: 25, Saturation: -30, Luminance: 12.5}}}}
	if !validVisualProcessing(processing) {
		t.Fatal("valid HSL rejected")
	}
	data, err := json.Marshal(processing)
	if err != nil {
		t.Fatal(err)
	}
	var decoded VisualProcessing
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Color.HSL, processing.Color.HSL) {
		t.Fatal("HSL lost in stored JSON")
	}
	for _, settings := range []map[string]HSLBandAdjustment{{"red": {Hue: 101}}, {"red": {Saturation: math.NaN()}}, {"unknown": {Hue: 1}}} {
		processing.Color.HSL = settings
		if validVisualProcessing(processing) {
			t.Errorf("invalid HSL accepted: %v", settings)
		}
	}
}

func TestHSLTimelineStoreRoundTripAndIsolation(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	now := time.Now()
	if err := store.CreateProject(ctx, "alice", Project{ID: "hsl-round-trip", Title: "HSL", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	settings := map[string]HSLBandAdjustment{"orange": {Hue: 25, Saturation: -30, Luminance: 12.5}}
	timeline := Timeline{Tracks: []Track{{ID: "video", Kind: TrackVideo, Clips: []Clip{{ID: "clip", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &VisualProcessing{Color: ColorAdjustment{HSL: settings}}}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", "hsl-round-trip", timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	settings["orange"] = HSLBandAdjustment{Hue: 99}
	loaded, err := store.Timeline(ctx, "alice", "hsl-round-trip")
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Tracks[0].Clips[0].VisualProcessing.Color.HSL["orange"]; got.Hue != 25 || got.Saturation != -30 || got.Luminance != 12.5 {
		t.Fatalf("HSL lost or aliased in timeline: %v", got)
	}
}

func TestHSLCubeOrderingAndIdentity(t *testing.T) {
	if HasHSLAdjustments(nil) {
		t.Fatal("identity marked active")
	}
	cube := BuildHSLCube(map[string]HSLBandAdjustment{"red": {Hue: 100}})
	if cube == "" {
		t.Fatal("missing cube")
	}
	// Red varies fastest; first sample black, sample 32 is pure red transformed to orange.
	lines := strings.Split(strings.TrimSpace(cube), "\n")
	if lines[0] != "LUT_3D_SIZE 33" || lines[1] != "DOMAIN_MIN 0 0 0" || lines[2] != "DOMAIN_MAX 1 1 1" {
		t.Fatalf("invalid cube header: %v", lines[:3])
	}
	if lines[3] != "0.000000000 0.000000000 0.000000000" || lines[35] != "1.000000000 0.501960784 0.000000000" {
		t.Fatal("cube ordering or channel transform wrong")
	}
	if len(lines) != 33*33*33+3 {
		t.Fatalf("wrong lattice length %d", len(lines))
	}
}
