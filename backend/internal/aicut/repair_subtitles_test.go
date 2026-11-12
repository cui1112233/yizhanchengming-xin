package aicut

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestRepairSRTKeepsScriptAndMarkers(t *testing.T) {
	before := Timeline{Revision: 3, Script: "保留剧本", Markers: []TimelineMarker{{ID: "m", TimeMS: 500, Label: "提示"}}}
	got, _, err := ApplySRTCues(before, "x.srt", []SRTCue{{0, 1000, "字幕"}}, newASRID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Script != before.Script || !reflect.DeepEqual(got.Markers, before.Markers) {
		t.Fatalf("metadata lost: %#v", got)
	}
}
func TestRepairASRMapsRepeatedTrimmedRateInstances(t *testing.T) {
	store, now := createASRFixture(t)
	ctx := context.Background()
	before := Timeline{Script: "保留剧本", Markers: []TimelineMarker{{ID: "m", TimeMS: 100, Label: "保留"}}, Tracks: []Track{{ID: "audio", Kind: TrackAudio, Clips: []Clip{{ID: "one", AssetID: "asset-1", Kind: ClipAudio, StartMS: 5000, DurationMS: 1000, SourceStartMS: 1000, SourceDurationMS: 10000, PlaybackRate: 2}, {ID: "two", AssetID: "asset-1", Kind: ClipAudio, StartMS: 9000, DurationMS: 2000, SourceStartMS: 2000, SourceDurationMS: 10000, PlaybackRate: 1}}}}}
	saved, err := store.SaveTimeline(ctx, "alice", "cut-1", before, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	service := NewASRService(ASRServiceOptions{Store: store, OpenObject: func(context.Context, Asset) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBufferString("x")), nil
	}, ResolveProvider: func(context.Context, string, string) (ASRProvider, error) {
		return &fixedASRProvider{cues: []SRTCue{{0, 2000, "first"}, {2000, 3500, "second"}}}, nil
	}, Now: func() time.Time { return now }})
	job, err := service.Submit(ctx, "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "asr", ExpectedRevision: saved.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.ProcessNext(ctx); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Timeline(ctx, "alice", "cut-1")
	clips := got.Tracks[len(got.Tracks)-1].Clips
	want := []SRTCue{{5000, 5500, "first"}, {5500, 6000, "second"}, {9000, 10500, "second"}}
	if len(clips) != len(want) {
		t.Fatalf("mapped clips=%#v", clips)
	}
	for i, c := range clips {
		if c.StartMS != want[i].StartMS || c.StartMS+c.DurationMS != want[i].EndMS || c.Text != want[i].Text {
			t.Fatalf("clip %d=%#v want %#v", i, c, want[i])
		}
	}
	if got.Script != before.Script || !reflect.DeepEqual(got.Markers, before.Markers) || !reflect.DeepEqual(got.Tracks[0], saved.Tracks[0]) {
		t.Fatalf("unrelated content changed: %#v", got)
	}
	completed, _ := store.ASRJob(ctx, "alice", "cut-1", job.ID)
	if completed.Imported != 3 {
		t.Fatalf("imported=%d", completed.Imported)
	}
}
func TestRepairASRRejectsComplexSourceBeforeSave(t *testing.T) {
	for _, kind := range []string{"curve", "reverse", "freeze", "compound"} {
		t.Run(kind, func(t *testing.T) {
			store, now := createASRFixture(t)
			ctx := context.Background()
			clip := Clip{ID: "c", AssetID: "asset-1", Kind: ClipVideo, StartMS: 0, DurationMS: 1000, SourceDurationMS: 10000}
			switch kind {
			case "curve":
				clip.SpeedCurve = []SpeedCurvePoint{{Position: 0, Rate: 1}, {Position: 1, Rate: 2}}
			case "reverse":
				clip.Reversed = true
			case "freeze":
				v := int64(1)
				clip.FreezeSourceMS = &v
			case "compound":
				clip.CompoundID = "compound"
				clip.CompoundName = "compound"
			}
			original, err := store.SaveTimeline(ctx, "alice", "cut-1", Timeline{Script: "kept", Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{clip}}}}, 0, now)
			if err != nil {
				t.Fatal(err)
			}
			service := NewASRService(ASRServiceOptions{Store: store, OpenObject: func(context.Context, Asset) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewBufferString("x")), nil
			}, ResolveProvider: func(context.Context, string, string) (ASRProvider, error) {
				return &fixedASRProvider{cues: []SRTCue{{0, 1000, "text"}}}, nil
			}})
			_, err = service.Submit(ctx, "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "m", ExpectedRevision: original.Revision})
			if err != nil {
				t.Fatal(err)
			}
			if err = service.ProcessNext(ctx); err == nil {
				t.Fatal("complex source accepted")
			}
			got, _ := store.Timeline(ctx, "alice", "cut-1")
			if !reflect.DeepEqual(got, original) {
				t.Fatal("failed ASR changed timeline")
			}
		})
	}
}
func TestRepairASRNamedTimelineMappingPreservesOtherTimelines(t *testing.T) {
	store, now := createASRFixture(t)
	ctx := context.Background()
	collections := NewMemoryTimelineCollectionStore(store)
	a, err := collections.Create(ctx, "alice", "cut-1", "A")
	if err != nil {
		t.Fatal(err)
	}
	b, err := collections.Create(ctx, "alice", "cut-1", "B")
	if err != nil {
		t.Fatal(err)
	}
	before := Timeline{Script: "named script", Markers: []TimelineMarker{{ID: "m", TimeMS: 200, Label: "keep"}}, Tracks: []Track{{ID: "audio", Kind: TrackAudio, Clips: []Clip{{ID: "one", AssetID: "asset-1", Kind: ClipAudio, StartMS: 3000, DurationMS: 1000, SourceStartMS: 2000, SourceDurationMS: 5000}}}}}
	saved, err := collections.Save(ctx, "alice", "cut-1", a.ID, before, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	service := NewASRService(ASRServiceOptions{Store: store, Timelines: collections, OpenObject: func(context.Context, Asset) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBufferString("x")), nil
	}, ResolveProvider: func(context.Context, string, string) (ASRProvider, error) {
		return &fixedASRProvider{cues: []SRTCue{{2000, 2500, "named"}}}, nil
	}})
	_, err = service.Submit(ctx, "alice", ASRSubmit{ProjectID: "cut-1", TimelineID: a.ID, AssetID: "asset-1", ModelID: "m", ExpectedRevision: saved.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.ProcessNext(ctx); err != nil {
		t.Fatal(err)
	}
	got, _ := collections.Timeline(ctx, "alice", "cut-1", a.ID)
	if got.Script != before.Script || !reflect.DeepEqual(got.Markers, before.Markers) || got.Tracks[1].Clips[0].StartMS != 3000 {
		t.Fatalf("bad named import %#v", got)
	}
	for _, id := range []string{b.ID, DefaultTimelineID} {
		untouched, err := collections.Timeline(ctx, "alice", "cut-1", id)
		if err != nil || untouched.Revision != 0 || len(untouched.Tracks) > 0 {
			t.Fatalf("other timeline changed: %#v %v", untouched, err)
		}
	}
}
