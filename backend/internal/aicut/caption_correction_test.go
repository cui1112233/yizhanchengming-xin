package aicut

import (
	"strings"
	"testing"
)

func TestParseCaptionCorrectionsAndApplyWithoutChangingTiming(t *testing.T) {
	raw := "```json\n{\"corrections\":[{\"clipId\":\"sub-1\",\"text\":\"洛水之畔\",\"reason\":\"修正同音字\"}]}\n```"
	items, err := ParseCaptionCorrections(raw)
	if err != nil {
		t.Fatal(err)
	}
	timeline := Timeline{Revision: 2, Tracks: []Track{{ID: "subtitle-1", Kind: TrackSubtitle, Clips: []Clip{
		{ID: "sub-1", Kind: ClipSubtitle, Name: "字幕", StartMS: 1200, DurationMS: 1800, SourceDurationMS: 1800, Text: "落水之畔"},
	}}}}
	updated, applied, err := ApplyCaptionCorrections(timeline, items)
	if err != nil {
		t.Fatal(err)
	}
	clip := updated.Tracks[0].Clips[0]
	if applied != 1 || clip.Text != "洛水之畔" {
		t.Fatalf("applied=%d clip=%+v", applied, clip)
	}
	if clip.StartMS != 1200 || clip.DurationMS != 1800 || clip.SourceDurationMS != 1800 {
		t.Fatalf("caption correction changed timing: %+v", clip)
	}
}

func TestCaptionCorrectionsRejectInvalidOrUnknownResults(t *testing.T) {
	for _, raw := range []string{"", `{"wrong":[]}`, `{"corrections":[{"clipId":"","text":"x"}]}`} {
		if _, err := ParseCaptionCorrections(raw); err == nil {
			t.Fatalf("ParseCaptionCorrections(%q) unexpectedly succeeded", raw)
		}
	}
	timeline := Timeline{Tracks: []Track{{ID: "subtitle-1", Kind: TrackSubtitle, Clips: []Clip{{ID: "sub-1", Kind: ClipSubtitle, DurationMS: 1000, SourceDurationMS: 1000, Text: "原文"}}}}}
	_, _, err := ApplyCaptionCorrections(timeline, []CaptionCorrection{{ClipID: "missing", Text: "改字"}})
	if err == nil || !strings.Contains(err.Error(), "没有匹配到可修改的字幕") {
		t.Fatalf("unknown clip error=%v", err)
	}
}
