package aicut

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClipAnimationValidationAndShortClipDurations(t *testing.T) {
	if !validClipAnimation(nil) {
		t.Fatal("old clip rejected")
	}
	value := &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "fade", DurationMS: 800}, Exit: ClipAnimationPhase{Preset: "slide", DurationMS: 800}}
	if !validClipAnimation(value) {
		t.Fatal("valid phases rejected")
	}
	in, out := ClipAnimationDurations(value, 1000)
	if in != 500 || out != 500 {
		t.Fatalf("overlap: %v %v", in, out)
	}
	value.Entrance.DurationMS = -1
	if validClipAnimation(value) {
		t.Fatal("negative duration accepted")
	}
	value.Entrance.DurationMS = 100
	value.Loop.Preset = "fade"
	if validClipAnimation(value) {
		t.Fatal("invalid loop preset accepted")
	}
}

func TestClipAnimationJSONPersistenceCompatibility(t *testing.T) {
	old := Clip{ID: "old"}
	raw, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"animation"`) {
		t.Fatal("old clip acquired animation")
	}
	old.Animation = &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "fade", DurationMS: 500}, Exit: ClipAnimationPhase{Preset: "zoom", DurationMS: 700}, Loop: ClipAnimationPhase{Preset: "float", DurationMS: 2000}}
	raw, err = json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	var restored Clip
	if err = json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Animation == nil || *restored.Animation != *old.Animation {
		t.Fatalf("animation lost on roundtrip %s", raw)
	}
}
