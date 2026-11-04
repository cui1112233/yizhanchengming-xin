package batchfactoryv11

import (
	"encoding/json"
	"testing"
)

func TestEffectiveVideoSettingsKeepImageRatioIndependent(t *testing.T) {
	legacySquareImage := SettingsPatch{"aspectRatio": json.RawMessage(`"1:1"`)}
	if got := EffectiveVideoAspectRatio(legacySquareImage); got != "9:16" {
		t.Fatalf("legacy square image ratio must fall back to vertical video, got %q", got)
	}
	if got := EffectiveImageAspectRatio(legacySquareImage); got != "1:1" {
		t.Fatalf("legacy square image ratio = %q, want 1:1", got)
	}
	if got := EffectiveVideoResolution(legacySquareImage); got != "720p" {
		t.Fatalf("default video resolution = %q, want 720p", got)
	}

	separate := SettingsPatch{
		"imageAspectRatio": json.RawMessage(`"1:1"`),
		"videoAspectRatio": json.RawMessage(`"16:9"`),
		"videoResolution":  json.RawMessage(`"1080p"`),
	}
	if got := EffectiveVideoAspectRatio(separate); got != "16:9" {
		t.Fatalf("video ratio = %q, want 16:9", got)
	}
	if got := EffectiveImageAspectRatio(separate); got != "1:1" {
		t.Fatalf("image ratio = %q, want 1:1", got)
	}
	if got := EffectiveVideoResolution(separate); got != "1080p" {
		t.Fatalf("video resolution = %q, want 1080p", got)
	}
}

func TestEffectiveImageAspectRatioSupportsFourNewRatios(t *testing.T) {
	for _, ratio := range []string{"2:3", "3:4", "4:3", "3:2"} {
		values := SettingsPatch{"imageAspectRatio": json.RawMessage(`"` + ratio + `"`)}
		if got := EffectiveImageAspectRatio(values); got != ratio {
			t.Fatalf("image ratio = %q, want %q", got, ratio)
		}
	}
}

func TestEffectiveImageAspectRatioKeepsLegacyValuesAndUnknownFallback(t *testing.T) {
	for _, ratio := range []string{"16:9", "9:16", "1:1"} {
		values := SettingsPatch{"imageAspectRatio": json.RawMessage(`"` + ratio + `"`)}
		if got := EffectiveImageAspectRatio(values); got != ratio {
			t.Fatalf("legacy image ratio = %q, want %q", got, ratio)
		}
	}
	unknown := SettingsPatch{"imageAspectRatio": json.RawMessage(`"7:5"`)}
	if got := EffectiveImageAspectRatio(unknown); got != "9:16" {
		t.Fatalf("unknown image ratio = %q, want 9:16 fallback", got)
	}
}

func TestImageAspectRatioOrientationDistinguishesPortraitAndLandscape(t *testing.T) {
	for _, ratio := range []string{"2:3", "3:4", "9:16"} {
		if !IsPortraitImageAspectRatio(ratio) || ImageAspectRatioOrientation(ratio) != "竖幅构图" {
			t.Fatalf("%s must be portrait", ratio)
		}
	}
	for _, ratio := range []string{"4:3", "3:2", "16:9"} {
		if IsPortraitImageAspectRatio(ratio) || ImageAspectRatioOrientation(ratio) != "横幅构图" {
			t.Fatalf("%s must be landscape", ratio)
		}
	}
	if IsPortraitImageAspectRatio("1:1") || ImageAspectRatioOrientation("1:1") != "方形构图" {
		t.Fatal("1:1 must be square")
	}
}
