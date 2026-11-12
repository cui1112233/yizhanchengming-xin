package aicut

import (
	"errors"
	"testing"
)

func editPlanBaseTimeline() Timeline {
	return Timeline{Revision: 7, Script: "开场对白\n自动字幕", Tracks: []Track{
		{ID: "video-1", Kind: TrackVideo, Name: "主画面", Order: 0, Clips: []Clip{
			{ID: "clip-1", AssetID: "asset-video", Kind: ClipVideo, Name: "镜头一", StartMS: 0, DurationMS: 2000, SourceStartMS: 0, SourceDurationMS: 5000, PlaybackRate: 1, Volume: 1},
			{ID: "clip-2", AssetID: "asset-video", Kind: ClipVideo, Name: "镜头二", StartMS: 2000, DurationMS: 2000, SourceStartMS: 2000, SourceDurationMS: 5000, PlaybackRate: 1, Volume: 1},
		}},
		{ID: "subtitle-1", Kind: TrackSubtitle, Name: "字幕", Order: 1, Settings: TrackSettings{PreviewVisible: true}, Clips: []Clip{}},
		{ID: "effect-1", Kind: TrackEffect, Name: "特效", Order: 2, Clips: []Clip{}},
		{ID: "audio-1", Kind: TrackAudio, Name: "音频", Order: 3, Clips: []Clip{}},
	}}
}

func editPlanForApply(operations ...EditOperation) EditPlan {
	plan := validEditPlanForTest()
	plan.BaseRevision = 7
	plan.Request.IncludeTransitions = true
	plan.Request.IncludeFilters = true
	plan.Request.IncludeBGM = true
	plan.Request.SelectedAssets = []EditPlanAsset{{ID: "asset-bgm", Kind: "audio", Name: "背景音乐", DurationMS: 10000}}
	plan.Request.AssetIDs = []string{"asset-video", "asset-bgm"}
	plan.Operations = operations
	return plan
}

func findClipForEditPlanTest(t *testing.T, timeline Timeline, trackID, clipID string) Clip {
	t.Helper()
	for _, track := range timeline.Tracks {
		if track.ID != trackID {
			continue
		}
		for _, clip := range track.Clips {
			if clip.ID == clipID {
				return clip
			}
		}
	}
	t.Fatalf("missing clip %s/%s", trackID, clipID)
	return Clip{}
}

func TestApplyEditPlanAppliesEnabledOperationsOnly(t *testing.T) {
	plan := editPlanForApply(
		EditOperation{ID: "op-trim", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: 200, EndMS: 1800, Reason: "去掉首尾空白", Enabled: true},
		EditOperation{ID: "op-caption", Type: EditAddCaption, TrackID: "subtitle-1", StartMS: 250, DurationMS: 900, Text: "开场对白", Reason: "补充字幕", Enabled: true},
		EditOperation{ID: "op-transition", Type: EditSetTransition, TrackID: "video-1", ClipID: "clip-2", ResourceID: "transition-fade", DurationMS: 300, Reason: "柔和连接", Enabled: true},
		EditOperation{ID: "op-disabled", Type: EditRemoveRange, StartMS: 3000, EndMS: 3500, Reason: "测试关闭", Enabled: false},
	)
	result, summary, err := ApplyEditPlan(editPlanBaseTimeline(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Applied != 3 || summary.Skipped != 1 {
		t.Fatalf("summary=%#v", summary)
	}
	if result.Revision != 7 {
		t.Fatalf("pure application changed revision to %d", result.Revision)
	}
	trimmed := findClipForEditPlanTest(t, result, "video-1", "clip-1")
	if trimmed.StartMS != 0 || trimmed.SourceStartMS != 200 || trimmed.DurationMS != 1600 {
		t.Fatalf("trimmed=%#v", trimmed)
	}
	if shifted := findClipForEditPlanTest(t, result, "video-1", "clip-2"); shifted.StartMS != 1600 {
		t.Fatalf("ripple-shifted clip=%#v", shifted)
	}
	caption := findClipForEditPlanTest(t, result, "subtitle-1", "plan-op-caption")
	if caption.Text != "开场对白" || caption.StartMS != 250 || caption.DurationMS != 900 {
		t.Fatalf("caption=%#v", caption)
	}
	transition := findClipForEditPlanTest(t, result, "video-1", "clip-2")
	if transition.Transition != "transition-fade" || transition.TransitionMS != 300 {
		t.Fatalf("transition=%#v", transition)
	}
}

func TestApplyEditPlanRejectsMissingClip(t *testing.T) {
	plan := editPlanForApply(EditOperation{ID: "op-1", Type: EditMoveClip, TrackID: "video-1", ClipID: "missing", StartMS: 500, Reason: "移动", Enabled: true})
	if _, _, err := ApplyEditPlan(editPlanBaseTimeline(), plan); !errors.Is(err, ErrEditPlanApply) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyEditPlanRejectsLockedTrack(t *testing.T) {
	timeline := editPlanBaseTimeline()
	timeline.Tracks[0].Locked = true
	plan := editPlanForApply(EditOperation{ID: "op-1", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: 100, EndMS: 1500, Reason: "裁剪", Enabled: true})
	if _, _, err := ApplyEditPlan(timeline, plan); !errors.Is(err, ErrEditPlanApply) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyEditPlanRejectsTrimBeyondClipAndSource(t *testing.T) {
	plan := editPlanForApply(EditOperation{ID: "op-1", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: 100, EndMS: 2500, Reason: "越界裁剪", Enabled: true})
	if _, _, err := ApplyEditPlan(editPlanBaseTimeline(), plan); !errors.Is(err, ErrEditPlanApply) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyEditPlanRejectsOverlappingMove(t *testing.T) {
	plan := editPlanForApply(EditOperation{ID: "op-1", Type: EditMoveClip, TrackID: "video-1", ClipID: "clip-2", StartMS: 1000, Reason: "移动到重叠区", Enabled: true})
	if _, _, err := ApplyEditPlan(editPlanBaseTimeline(), plan); !errors.Is(err, ErrEditPlanApply) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyEditPlanRejectsTransitionWithoutAdjacentPreviousClip(t *testing.T) {
	timeline := editPlanBaseTimeline()
	timeline.Tracks[0].Clips[1].StartMS = 2500
	plan := editPlanForApply(EditOperation{ID: "op-1", Type: EditSetTransition, TrackID: "video-1", ClipID: "clip-2", ResourceID: "transition-fade", DurationMS: 300, Reason: "添加转场", Enabled: true})
	if _, _, err := ApplyEditPlan(timeline, plan); !errors.Is(err, ErrEditPlanApply) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyEditPlanRejectsBGMOutsideSelectedAssets(t *testing.T) {
	plan := editPlanForApply(EditOperation{ID: "op-1", Type: EditAddBGM, TrackID: "audio-1", AssetID: "asset-not-selected", StartMS: 0, DurationMS: 1000, Volume: .2, Reason: "添加音乐", Enabled: true})
	if _, _, err := ApplyEditPlan(editPlanBaseTimeline(), plan); !errors.Is(err, ErrEditPlanApply) {
		t.Fatalf("err=%v", err)
	}
}

func TestApplyEditPlanReturnsProjectFormatWithoutDuplicatingItInTimeline(t *testing.T) {
	plan := editPlanForApply(EditOperation{ID: "op-format", Type: EditSetProjectFormat, ResourceID: "16:9", Reason: "改为横屏", Enabled: true})
	result, summary, err := ApplyEditPlan(editPlanBaseTimeline(), plan)
	if err != nil {
		t.Fatal(err)
	}
	if summary.AspectRatio != "16:9" || result.Revision != 7 {
		t.Fatalf("summary=%#v timeline=%#v", summary, result)
	}
}

func TestEditPlanPlacesActualAssetsAndRejectsInventedRanges(t *testing.T) {
	p := editPlanForApply(EditOperation{ID: "v", Type: EditPlaceAsset, TrackID: "main", AssetID: "asset-video", StartMS: 0, SourceStartMS: 500, DurationMS: 1500, Reason: "按素材顺序拼接", Enabled: true}, EditOperation{ID: "i", Type: EditPlaceAsset, TrackID: "main", AssetID: "image", StartMS: 1500, DurationMS: 2500, Reason: "展示图片", Enabled: true})
	p.Request.AssetIDs = append(p.Request.AssetIDs, "image")
	p.Request.SelectedAssets = []EditPlanAsset{{ID: "asset-video", Name: "真实视频.mp4", Kind: "video", DurationMS: 3000}, {ID: "image", Name: "真实图片.png", Kind: "image"}}
	timeline := Timeline{Revision: 7, Tracks: []Track{}}
	got, sum, err := ApplyEditPlan(timeline, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Tracks) != 1 || len(got.Tracks[0].Clips) != 2 || got.Tracks[0].Clips[0].SourceStartMS != 500 || got.Tracks[0].Clips[0].SourceDurationMS != 3000 || got.Tracks[0].Clips[1].Kind != ClipImage || sum.EstimatedDurationMS != 4000 {
		t.Fatalf("got=%+v sum=%+v", got, sum)
	}
	p.Operations[0].SourceStartMS = 2000
	if _, _, err := ApplyEditPlan(timeline, p); err == nil {
		t.Fatal("source overrun accepted")
	}
	p.Operations[0].AssetID = "invented"
	if _, _, err := ApplyEditPlan(timeline, p); err == nil {
		t.Fatal("invented asset accepted")
	}
	if len(timeline.Tracks) != 0 {
		t.Fatal("input mutated")
	}
}

func TestEditPlanDoesNotInventDialogue(t *testing.T) {
	p := editPlanForApply(EditOperation{ID: "caption", Type: EditAddCaption, TrackID: "subtitle-1", Text: "凭空编造的对白", DurationMS: 1000, Reason: "自动字幕", Enabled: true})
	if _, _, err := ApplyEditPlan(editPlanBaseTimeline(), p); err == nil {
		t.Fatal("invented dialogue accepted without transcript")
	}
	timeline := editPlanBaseTimeline()
	timeline.Script = "凭空编造的对白"
	if _, _, err := ApplyEditPlan(timeline, p); err != nil {
		t.Fatal(err)
	}
}

func TestEditPlanFilterPreservesRequestedIntensity(t *testing.T) {
	for _, intensity := range []float64{0, .8} {
		p := editPlanForApply(EditOperation{ID: "filter", Type: EditSetFilter, TrackID: "video-1", ClipID: "clip-1", ResourceID: "filter-cinema", Parameters: map[string]float64{"intensity": intensity}, Reason: "r", Enabled: true})
		timeline, _, err := ApplyEditPlan(editPlanBaseTimeline(), p)
		if err != nil {
			t.Fatal(err)
		}
		clip := findClipForEditPlanTest(t, timeline, "video-1", "clip-1")
		value, exists := clip.CatalogParameters["intensity"]
		if !exists || value != intensity {
			t.Fatalf("intensity=%v exists=%v want=%v", value, exists, intensity)
		}
	}
}
func TestEditPlanFilterRejectsInvalidIntensity(t *testing.T) {
	for _, intensity := range []float64{-1, 1.1} {
		p := editPlanForApply(EditOperation{ID: "filter", Type: EditSetFilter, TrackID: "video-1", ClipID: "clip-1", ResourceID: "filter-cinema", Parameters: map[string]float64{"intensity": intensity}, Reason: "r", Enabled: true})
		if _, _, err := ApplyEditPlan(editPlanBaseTimeline(), p); err == nil {
			t.Fatalf("accepted intensity=%v", intensity)
		}
	}
}
