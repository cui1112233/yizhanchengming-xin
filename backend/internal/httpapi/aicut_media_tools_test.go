package httpapi

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	"qiantie/backend/internal/aicut"
)

func TestAnalyzeAICutPCMBuildsBoundedWaveformAndFindsBeats(t *testing.T) {
	const sampleRate = 2000
	const durationMS = 2000
	samples := make([]int16, sampleRate*durationMS/1000)
	for _, centerMS := range []int{500, 1000, 1500} {
		center := centerMS * sampleRate / 1000
		for offset := -20; offset <= 20; offset++ {
			index := center + offset
			if index >= 0 && index < len(samples) {
				samples[index] = int16(28000 * math.Cos(float64(offset)*math.Pi/40))
			}
		}
	}
	var pcm bytes.Buffer
	if err := binary.Write(&pcm, binary.LittleEndian, samples); err != nil {
		t.Fatal(err)
	}

	got, err := analyzeAICutPCM(&pcm, sampleRate, durationMS, 40)
	if err != nil {
		t.Fatal(err)
	}
	if got.DurationMS != durationMS || got.SampleRate != sampleRate || len(got.Peaks) != 40 {
		t.Fatalf("unexpected analysis shape: %+v", got)
	}
	if len(got.BeatsMS) != 3 {
		t.Fatalf("beats=%v want three pulses", got.BeatsMS)
	}
	for index, want := range []int64{500, 1000, 1500} {
		if delta := got.BeatsMS[index] - want; delta < -60 || delta > 60 {
			t.Fatalf("beat[%d]=%d want near %d", index, got.BeatsMS[index], want)
		}
	}
}

func TestAnalyzeAICutPCMRejectsInvalidBoundsAndEmptyAudio(t *testing.T) {
	if _, err := analyzeAICutPCM(bytes.NewReader(nil), 0, 1000, 40); err == nil {
		t.Fatal("expected invalid sample rate error")
	}
	if _, err := analyzeAICutPCM(bytes.NewReader(nil), 2000, 1000, 3); err == nil {
		t.Fatal("expected invalid point count error")
	}
	if _, err := analyzeAICutPCM(bytes.NewReader(nil), 2000, 1000, 40); err == nil {
		t.Fatal("expected empty audio error")
	}
}

func TestParseAICutSceneTimesDeduplicatesCloseCuts(t *testing.T) {
	output := []byte("showinfo pts_time:0.2\nshowinfo pts_time:0.4\nshowinfo pts_time:1.1\nshowinfo pts_time:2.0\n")
	if got, want := parseAICutSceneTimes(output, 500), []int64{200, 1100, 2000}; !reflect.DeepEqual(got, want) {
		t.Fatalf("times=%v want=%v", got, want)
	}
}

func TestApplyAICutSceneSplitsKeepsSourceAndTimelineContinuous(t *testing.T) {
	timeline := aicut.Timeline{Revision: 2, Tracks: []aicut.Track{{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "clip-1", AssetID: "asset-1", Kind: aicut.ClipVideo, Name: "原片", StartMS: 1000, SourceStartMS: 500, DurationMS: 3000, SourceDurationMS: 5000}}}}}
	got, count := applyAICutSceneSplits(timeline, "asset-1", []int64{1000, 2200, 4500})
	if count != 2 || len(got.Tracks[0].Clips) != 3 {
		t.Fatalf("count=%d clips=%+v", count, got.Tracks[0].Clips)
	}
	clips := got.Tracks[0].Clips
	if clips[0].StartMS != 1000 || clips[0].SourceStartMS != 500 || clips[0].DurationMS != 500 || clips[1].StartMS != 1500 || clips[1].SourceStartMS != 1000 || clips[2].StartMS != 2700 || clips[2].DurationMS != 1300 {
		t.Fatalf("unexpected split geometry: %+v", clips)
	}
}

func TestApplyAICutSceneSplitsDoesNotChangeLockedTrack(t *testing.T) {
	timeline := aicut.Timeline{Tracks: []aicut.Track{{ID: "video-1", Kind: aicut.TrackVideo, Locked: true, Clips: []aicut.Clip{{ID: "clip-1", AssetID: "asset-1", Kind: aicut.ClipVideo, DurationMS: 2000}}}}}
	got, count := applyAICutSceneSplits(timeline, "asset-1", []int64{1000})
	if count != 0 || len(got.Tracks[0].Clips) != 1 {
		t.Fatalf("locked track changed: %+v", got)
	}
}

func TestAICutVisualFrameTimesSamplesInsideVideoBounds(t *testing.T) {
	if got, want := aicutVisualFrameTimes(10000, 4), []int64{1000, 3667, 6333, 9000}; !reflect.DeepEqual(got, want) {
		t.Fatalf("times=%v want=%v", got, want)
	}
	if got, want := aicutVisualFrameTimes(600, 6), []int64{60, 156, 252, 348, 444, 540}; !reflect.DeepEqual(got, want) {
		t.Fatalf("short times=%v want=%v", got, want)
	}
	if got := aicutVisualFrameTimes(0, 4); len(got) != 0 {
		t.Fatalf("zero duration times=%v", got)
	}
}

func TestAICutTrackingFrameTimesCoverRequestedSourceRange(t *testing.T) {
	got, err := aicutTrackingFrameTimes(10000, 1200, 3000, 4)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int64{1200, 2200, 3200, 4200}; !reflect.DeepEqual(got, want) {
		t.Fatalf("times=%v want=%v", got, want)
	}
	if _, err := aicutTrackingFrameTimes(10000, -1, 3000, 4); err == nil {
		t.Fatal("expected negative start error")
	}
	if _, err := aicutTrackingFrameTimes(10000, 9000, 2000, 4); err == nil {
		t.Fatal("expected out of range error")
	}
	if _, err := aicutTrackingFrameTimes(10000, 0, 3000, 1); err == nil {
		t.Fatal("expected count error")
	}
}
