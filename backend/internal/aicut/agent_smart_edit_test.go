package aicut

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestAgentCaptionSelectionKeepsSpeechAndRateOffsets(t *testing.T) {
	ctx := context.Background()
	store, _, run, tools := newAgentToolFixture(t, false)
	timeline, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	timeline.Tracks[0].Clips = []Clip{{ID: "video", Kind: ClipVideo, Name: "source", DurationMS: 5000, SourceDurationMS: 10000, PlaybackRate: 2, Volume: 1}}
	timeline.Tracks[1].Clips = []Clip{{ID: "speech", Kind: ClipSubtitle, Name: "caption", Text: "保留这一句", StartMS: 1000, DurationMS: 1000}}
	saved, err := store.SaveTimeline(ctx, run.Owner, run.ProjectID, timeline, run.ExpectedRevision, tools.now())
	if err != nil {
		t.Fatal(err)
	}
	run.ExpectedRevision = saved.Revision
	result, err := tools.Execute(ctx, run, ToolCall{ID: "select", Name: "select_caption_ranges", Arguments: json.RawMessage(`{"clipIds":["speech"],"mode":"keep","paddingMs":200}`)})
	if err != nil || !result.OK {
		t.Fatalf("%#v %v", result, err)
	}
	got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	video := got.Tracks[0].Clips[0]
	if video.StartMS != 0 || video.DurationMS != 1400 || video.SourceStartMS != 1600 || got.Tracks[1].Clips[0].StartMS != 200 {
		t.Fatalf("wrong selection: %#v", got)
	}
	undone, _, err := store.UndoTimeline(ctx, run.Owner, run.ProjectID, got.Revision, tools.now())
	if err != nil || undone.Tracks[0].Clips[0].DurationMS != 5000 {
		t.Fatalf("undo: %#v %v", undone, err)
	}
}

func TestAgentSmartEditRejectsLockedAndUnknownCaptions(t *testing.T) {
	for _, args := range []string{`{"clipIds":["missing"],"mode":"keep"}`, `{"clipIds":["subtitle-2"],"mode":"keep"}`} {
		store, _, run, tools := newAgentToolFixture(t, true)
		result, err := tools.Execute(context.Background(), run, ToolCall{ID: "select", Name: "select_caption_ranges", Arguments: json.RawMessage(args)})
		if err != nil || result.OK {
			t.Fatalf("accepted invalid selection %#v %v", result, err)
		}
		got, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
		if got.Revision != run.ExpectedRevision {
			t.Fatal("invalid selection changed timeline")
		}
	}
}

func TestAgentAppliesAnimationAndRejectsInvalidPreset(t *testing.T) {
	store, _, run, tools := newAgentToolFixture(t, false)
	result, err := tools.Execute(context.Background(), run, ToolCall{ID: "animate", Name: "set_clip_animation", Arguments: json.RawMessage(`{"trackId":"video-1","clipId":"clip-a","animation":{"entrance":{"preset":"fade","durationMs":300}}}`)})
	if err != nil || !result.OK {
		t.Fatalf("%#v %v", result, err)
	}
	got, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
	if got.Tracks[0].Clips[0].Animation.Entrance.Preset != "fade" {
		t.Fatal("animation not persisted")
	}
	run.ExpectedRevision = got.Revision
	result, err = tools.Execute(context.Background(), run, ToolCall{ID: "bad", Name: "set_clip_animation", Arguments: json.RawMessage(`{"trackId":"video-1","clipId":"clip-a","animation":{"entrance":{"preset":"unknown","durationMs":300}}}`)})
	if err != nil || result.OK {
		t.Fatalf("invalid animation accepted %#v %v", result, err)
	}
}

func TestAgentCompactGapsKeepsNaturalPauseAndSync(t *testing.T) {
	ctx := context.Background()
	store, _, run, tools := newAgentToolFixture(t, false)
	timeline, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	timeline.Tracks[0].Clips = []Clip{{ID: "video", Kind: ClipVideo, DurationMS: 7000, SourceDurationMS: 7000}}
	timeline.Tracks[1].Clips = []Clip{{ID: "first", Kind: ClipSubtitle, Text: "第一句", DurationMS: 1000}, {ID: "second", Kind: ClipSubtitle, Text: "第二句", StartMS: 5000, DurationMS: 1000}}
	saved, err := store.SaveTimeline(ctx, run.Owner, run.ProjectID, timeline, run.ExpectedRevision, tools.now())
	if err != nil {
		t.Fatal(err)
	}
	run.ExpectedRevision = saved.Revision
	result, err := tools.Execute(ctx, run, ToolCall{ID: "compact", Name: "compact_caption_gaps", Arguments: json.RawMessage(`{"maxPauseMs":1000}`)})
	if err != nil || !result.OK {
		t.Fatalf("%#v %v", result, err)
	}
	got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	if got.Tracks[1].Clips[1].StartMS != 2000 {
		t.Fatalf("wrong pause: %#v", got)
	}
}

func TestAgentAtomicOperationsAndColor(t *testing.T) {
	ctx := context.Background()
	store, _, run, tools := newAgentToolFixture(t, false)
	result, err := tools.Execute(ctx, run, ToolCall{ID: "color", Name: "set_clip_color", Arguments: json.RawMessage(`{"trackId":"video-1","clipId":"clip-a","color":{"exposure":0.2,"temperature":0.1,"gamma":1}}`)})
	if err != nil || !result.OK {
		t.Fatalf("color %#v %v", result, err)
	}
	got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	if got.Tracks[0].Clips[0].VisualProcessing.Color.Exposure != 0.2 {
		t.Fatal("color not saved")
	}
	run.ExpectedRevision = got.Revision
	result, err = tools.Execute(ctx, run, ToolCall{ID: "batch", Name: "apply_edit_operations", Arguments: json.RawMessage(`{"operations":[{"type":"set_filter","trackId":"video-1","clipId":"clip-a","resourceId":"filter-cinema"},{"type":"move_clip","trackId":"video-1","clipId":"missing","startMs":2000}]}`)})
	if err != nil || result.OK {
		t.Fatalf("invalid atomic edit accepted %#v %v", result, err)
	}
	got, _ = store.Timeline(ctx, run.Owner, run.ProjectID)
	if got.Revision != run.ExpectedRevision || got.Tracks[0].Clips[0].CatalogResourceID != "" {
		t.Fatal("partial edit saved")
	}
	result, err = tools.Execute(ctx, run, ToolCall{ID: "filter", Name: "apply_edit_operations", Arguments: json.RawMessage(`{"operations":[{"type":"set_filter","trackId":"video-1","clipId":"clip-a","resourceId":"filter-cinema"}]}`)})
	if err != nil || !result.OK {
		t.Fatalf("filter %#v %v", result, err)
	}
	got, _ = store.Timeline(ctx, run.Owner, run.ProjectID)
	if got.Tracks[0].Clips[0].CatalogResourceID != "filter-cinema" {
		t.Fatal("filter not persisted")
	}
}

func TestAgentConsecutiveCaptionOperationsHaveUniqueIDs(t *testing.T) {
	ctx := context.Background()
	store, _, run, tools := newAgentToolFixture(t, false)
	for i, id := range []string{"first-call", "second-call"} {
		args := fmt.Sprintf(`{"operations":[{"type":"add_caption","trackId":"subtitle-1","startMs":%d,"durationMs":500,"text":"旧字幕","resourceId":"text-basic-caption"}]}`, 2000+i*1000)
		result, err := tools.Execute(ctx, run, ToolCall{ID: id, Name: "apply_edit_operations", Arguments: json.RawMessage(args)})
		if err != nil || !result.OK {
			t.Fatalf("call %d: %#v %v", i, result, err)
		}
		run.ExpectedRevision = result.Revision
	}
	got, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	if len(got.Tracks[1].Clips) != 3 {
		t.Fatal("captions not added")
	}
}
