package aicut

import "testing"

func TestParseFillerAnalysisRejectsDuplicatesAndAcceptsEmpty(t *testing.T) {
	items, err := ParseFillerAnalysis(`{"fillers":[]}`)
	if err != nil || len(items) != 0 {
		t.Fatalf("empty analysis=%v err=%v", items, err)
	}
	for _, raw := range []string{
		`not-json`,
		`{"fillers":null}`,
		`{"fillers":[{"clipId":""}]}`,
		`{"fillers":[{"clipId":"sub-1"},{"clipId":"sub-1"}]}`,
	} {
		if _, err := ParseFillerAnalysis(raw); err == nil {
			t.Fatalf("ParseFillerAnalysis(%q) unexpectedly succeeded", raw)
		}
	}
}

func TestApplyFillerTrimOnlyDeletesMatchedSubtitleRows(t *testing.T) {
	timeline := Timeline{Revision: 4, Tracks: []Track{
		{ID: "subtitle-1", Kind: TrackSubtitle, Clips: []Clip{
			{ID: "sub-1", Kind: ClipSubtitle, StartMS: 0, DurationMS: 1000, SourceDurationMS: 1000, Text: "开场"},
			{ID: "sub-2", Kind: ClipSubtitle, StartMS: 1000, DurationMS: 500, SourceDurationMS: 500, Text: "嗯"},
			{ID: "sub-3", Kind: ClipSubtitle, StartMS: 4000, DurationMS: 1000, SourceDurationMS: 1000, Text: "继续"},
		}},
		{ID: "video-1", Kind: TrackVideo, Clips: []Clip{
			{ID: "video-clip", Kind: ClipVideo, AssetID: "asset-1", StartMS: 0, DurationMS: 6000, SourceDurationMS: 6000},
		}},
	}}

	result, summary, err := ApplyFillerTrim(timeline, []FillerCandidate{{ClipID: "sub-2", Reason: "整行只有语气词"}})
	if err != nil {
		t.Fatal(err)
	}
	if summary.FillerCount != 1 || summary.PauseCount != 0 || summary.RemovedMS != 500 {
		t.Fatalf("summary=%+v", summary)
	}
	if got := result.Tracks[0].Clips; len(got) != 2 || got[1].StartMS != 4000 || got[1].Text != "继续" {
		t.Fatalf("subtitle clips=%+v", got)
	}
	video := result.Tracks[1].Clips
	if len(video) != 1 || video[0].StartMS != 0 || video[0].DurationMS != 6000 || video[0].SourceStartMS != 0 {
		t.Fatalf("video clips=%+v", video)
	}
}

func TestApplyFillerTrimIsRepeatSafeAndRejectsUnknownOrLockedSubtitle(t *testing.T) {
	timeline := Timeline{Revision: 1, Tracks: []Track{{ID: "subtitle-1", Kind: TrackSubtitle, Clips: []Clip{{ID: "sub-1", Kind: ClipSubtitle, DurationMS: 500, SourceDurationMS: 500, Text: "嗯"}}}}}
	if _, _, err := ApplyFillerTrim(timeline, []FillerCandidate{{ClipID: "missing"}}); err == nil {
		t.Fatal("unknown filler id should fail")
	}
	timeline.Tracks[0].Locked = true
	if _, _, err := ApplyFillerTrim(timeline, []FillerCandidate{{ClipID: "sub-1"}}); err == nil {
		t.Fatal("locked subtitle track should fail")
	}
	if result, summary, err := ApplyFillerTrim(timeline, nil); err != nil || summary.RemovedMS != 0 || result.Revision != timeline.Revision {
		t.Fatalf("empty repeat result=%+v summary=%+v err=%v", result, summary, err)
	}
}

func TestApplyOralCleanCutsMarkedSpeechAndLongPausesAcrossTracks(t *testing.T) {
	timeline := Timeline{Revision: 4, Tracks: []Track{
		{ID: "subtitle-1", Kind: TrackSubtitle, Clips: []Clip{
			{ID: "sub-1", Kind: ClipSubtitle, StartMS: 0, DurationMS: 1000, SourceDurationMS: 1000, Text: "开场"},
			{ID: "sub-2", Kind: ClipSubtitle, StartMS: 1000, DurationMS: 500, SourceDurationMS: 500, Text: "嗯"},
			{ID: "sub-3", Kind: ClipSubtitle, StartMS: 4000, DurationMS: 1000, SourceDurationMS: 1000, Text: "继续"},
		}},
		{ID: "video-1", Kind: TrackVideo, Clips: []Clip{{ID: "video-clip", Kind: ClipVideo, AssetID: "asset-1", StartMS: 0, DurationMS: 6000, SourceDurationMS: 6000}}},
	}}

	result, summary, err := ApplyOralClean(timeline, []FillerCandidate{{ClipID: "sub-2"}}, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if summary.FillerCount != 1 || summary.PauseCount != 1 || summary.RemovedMS != 2000 {
		t.Fatalf("summary=%+v", summary)
	}
	if got := result.Tracks[0].Clips; len(got) != 2 || got[1].StartMS != 2000 || got[1].Text != "继续" {
		t.Fatalf("subtitle clips=%+v", got)
	}
	video := result.Tracks[1].Clips
	if len(video) != 3 || video[0].DurationMS != 1000 || video[1].StartMS != 1000 || video[1].SourceStartMS != 1500 || video[2].StartMS != 2000 || video[2].SourceStartMS != 4000 {
		t.Fatalf("video clips=%+v", video)
	}
}
