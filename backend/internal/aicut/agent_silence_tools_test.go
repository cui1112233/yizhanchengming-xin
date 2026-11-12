package aicut

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAgentRemoveAudioSilenceMapsRateRepeatsAndReplays(t *testing.T) {
	ctx := context.Background()
	store, _, run, tools := newAgentToolFixture(t, false)
	tl, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	tl.Script = "原剧本"
	tl.Markers = []TimelineMarker{{ID: "mark", TimeMS: 2500, Label: "点"}}
	tl.Tracks[0].Clips = []Clip{{ID: "v1", AssetID: "source", Kind: ClipVideo, DurationMS: 2000, SourceStartMS: 1000, SourceDurationMS: 6000, PlaybackRate: 2}, {ID: "v2", AssetID: "source", Kind: ClipVideo, StartMS: 2000, DurationMS: 1000, SourceStartMS: 2000, SourceDurationMS: 6000}}
	tl.Tracks[1].Clips = []Clip{{ID: "sub", Kind: ClipSubtitle, Text: "字幕不决定静音", StartMS: 2500, DurationMS: 500}}
	saved, err := store.SaveTimeline(ctx, run.Owner, run.ProjectID, tl, run.ExpectedRevision, tools.now())
	if err != nil {
		t.Fatal(err)
	}
	run.ExpectedRevision = saved.Revision
	calls := 0
	tools.composite.AnalyzeSilence = func(context.Context, string, string, string, SilenceOptions) (SilenceAnalysis, error) {
		calls++
		return SilenceAnalysis{AssetID: "source", DurationMS: 6000, Intervals: []SilenceInterval{{2000, 2400}}}, nil
	}
	call := ToolCall{ID: "silence", Name: "remove_audio_silence", Arguments: json.RawMessage(`{"assetId":"source"}`)}
	res, err := tools.Execute(ctx, run, call)
	if err != nil || !res.OK {
		t.Fatalf("%+v %v", res, err)
	}
	got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	if got.Script != "原剧本" || got.Markers[0].TimeMS != 1900 || got.Tracks[1].Clips[0].StartMS != 1900 || len(got.Tracks[0].Clips) != 3 || got.Tracks[0].Clips[1].SourceStartMS != 2400 {
		t.Fatalf("%+v", got)
	}
	replay, err := tools.Execute(ctx, run, call)
	if err != nil || !replay.OK || calls != 1 || replay.Revision != res.Revision {
		t.Fatalf("replay %+v %v calls=%d", replay, err, calls)
	}
}
func TestAgentSilenceRejectsUnsafeWithoutMutation(t *testing.T) {
	for _, kind := range []string{"locked", "curve", "all", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			store, _, run, tools := newAgentToolFixture(t, false)
			tl, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
			tl.Tracks[0].Clips = []Clip{{ID: "v", AssetID: "source", Kind: ClipVideo, DurationMS: 3000, SourceDurationMS: 3000}}
			if kind == "locked" {
				tl.Tracks[0].Locked = true
			}
			if kind == "curve" {
				tl.Tracks[0].Clips[0].Reversed = true
			}
			saved, err := store.SaveTimeline(ctx, run.Owner, run.ProjectID, tl, run.ExpectedRevision, tools.now())
			if err != nil {
				t.Fatal(err)
			}
			run.ExpectedRevision = saved.Revision
			tools.composite.AnalyzeSilence = func(context.Context, string, string, string, SilenceOptions) (SilenceAnalysis, error) {
				return SilenceAnalysis{AssetID: "source", DurationMS: 3000, AllSilent: kind == "all", Intervals: []SilenceInterval{{1000, 2000}}}, nil
			}
			if kind == "cancel" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			res, _ := tools.Execute(ctx, run, ToolCall{ID: "x", Name: "remove_audio_silence", Arguments: json.RawMessage(`{"assetId":"source"}`)})
			if res.OK {
				t.Fatal("unsafe accepted")
			}
			got, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
			if got.Revision != saved.Revision {
				t.Fatal("mutated")
			}
		})
	}
}
func TestAgentAnalyzeAudioSilenceIsReadOnly(t *testing.T) {
	ctx := context.Background()
	store, _, run, tools := newAgentToolFixture(t, false)
	tools.composite.AnalyzeSilence = func(context.Context, string, string, string, SilenceOptions) (SilenceAnalysis, error) {
		return SilenceAnalysis{AssetID: "source", DurationMS: 3000, Intervals: []SilenceInterval{{1000, 2000}}}, nil
	}
	res, err := tools.Execute(ctx, run, ToolCall{ID: "analyze", Name: "analyze_audio_silence", Arguments: json.RawMessage(`{"assetId":"source"}`)})
	if err != nil || !res.OK {
		t.Fatalf("%+v %v", res, err)
	}
	got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	if got.Revision != run.ExpectedRevision || got.Tracks[0].Clips[0].DurationMS != 2000 {
		t.Fatal("analysis edited timeline")
	}
}
func TestAgentSilenceRejectsCutFadesAndAnimationOnOtherTracks(t *testing.T) {
	for _, kind := range []string{"fade", "animation"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			store, _, run, tools := newAgentToolFixture(t, false)
			tl, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
			tl.Tracks[0].Clips = []Clip{{ID: "v", AssetID: "source", Kind: ClipVideo, DurationMS: 3000, SourceDurationMS: 3000}}
			other := Clip{ID: "bgm", Kind: ClipAudio, DurationMS: 3000, SourceDurationMS: 3000}
			if kind == "fade" {
				other.FadeInMS = 1000
			} else {
				other.Kind = ClipVideo
				other.Animation = &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "fade", DurationMS: 1000}}
			}
			trackKind := TrackAudio
			if kind == "animation" {
				trackKind = TrackVideo
			}
			tl.Tracks[1] = Track{ID: "music", Kind: trackKind, Clips: []Clip{other}}
			saved, err := store.SaveTimeline(ctx, run.Owner, run.ProjectID, tl, run.ExpectedRevision, tools.now())
			if err != nil {
				t.Fatal(err)
			}
			run.ExpectedRevision = saved.Revision
			tools.composite.AnalyzeSilence = func(context.Context, string, string, string, SilenceOptions) (SilenceAnalysis, error) {
				return SilenceAnalysis{AssetID: "source", DurationMS: 3000, Intervals: []SilenceInterval{{1000, 2000}}}, nil
			}
			res, err := tools.Execute(ctx, run, ToolCall{ID: "x", Name: "remove_audio_silence", Arguments: json.RawMessage(`{"assetId":"source"}`)})
			if err != nil || res.OK {
				t.Fatalf("unsafe animation/fade cut accepted: %+v %v", res, err)
			}
			got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
			if got.Revision != saved.Revision || len(got.Tracks[1].Clips) != 1 {
				t.Fatal("failed preflight changed tracks")
			}
		})
	}
}
func TestSilencePreflightAllowsAnimatedWholeClipShift(t *testing.T) {
	tl := Timeline{Tracks: []Track{{Clips: []Clip{{ID: "later", StartMS: 2000, DurationMS: 1000, FadeInMS: 300, Animation: &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "fade", DurationMS: 300}}}}}}}
	if err := preflightSilenceCuts(tl, []trimInterval{{500, 1000}}); err != nil {
		t.Fatalf("whole clip movement rejected: %v", err)
	}
}

func TestSilencePreflightAllowsClearedAnimationSettings(t *testing.T) {
	tl := Timeline{Tracks: []Track{{Clips: []Clip{{ID: "plain", StartMS: 0, DurationMS: 3000, Animation: &ClipAnimation{Entrance: ClipAnimationPhase{Preset: "none"}, Exit: ClipAnimationPhase{Preset: "none"}, Loop: ClipAnimationPhase{Preset: "none"}}}}}}}
	if err := preflightSilenceCuts(tl, []trimInterval{{1000, 2000}}); err != nil {
		t.Fatalf("cleared animation is not an active effect: %v", err)
	}
}
