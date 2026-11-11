package aicut

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

func TestMemoryStorePersistsOwnerScopedProjectTimeline(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 9, 14, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	project := Project{ID: "cut-1", Title: "第一条成片", AspectRatio: "9:16", CreatedAt: now, UpdatedAt: now}

	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Project(ctx, "bob", project.ID); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("other owner read err=%v, want ErrProjectNotFound", err)
	}

	timeline := Timeline{
		Revision: 0,
		Tracks: []Track{{
			ID: "video-1", Kind: TrackVideo, Name: "主画面", Order: 0,
			Clips: []Clip{{ID: "clip-1", AssetID: "asset-1", Kind: ClipVideo, Name: "开场", StartMS: 0, DurationMS: 5000, SourceDurationMS: 9000}},
		}},
	}
	saved, err := store.SaveTimeline(ctx, "alice", project.ID, timeline, 0, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != 1 || len(saved.Tracks) != 1 || len(saved.Tracks[0].Clips) != 1 {
		t.Fatalf("saved=%+v", saved)
	}
	if _, err := store.SaveTimeline(ctx, "alice", project.ID, timeline, 0, now.Add(2*time.Minute)); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale save err=%v, want ErrRevisionConflict", err)
	}
}

func TestMemoryStorePersistsProjectCanvasAppearance(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 11, 16, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-alpha", Title: "透明工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateProjectAppearance(ctx, "alice", "cut-alpha", "transparent", "#123456", now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if updated.BackgroundMode != "transparent" || updated.BackgroundColor != "#123456" {
		t.Fatalf("updated=%+v", updated)
	}
	loaded, err := store.Project(ctx, "alice", "cut-alpha")
	if err != nil || loaded.BackgroundMode != "transparent" || loaded.BackgroundColor != "#123456" {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	if _, err := store.UpdateProjectAppearance(ctx, "alice", "cut-alpha", "gradient", "#123456", now); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("invalid mode err=%v", err)
	}
	if _, err := store.UpdateProjectAppearance(ctx, "alice", "cut-alpha", "color", "red", now); !errors.Is(err, ErrInvalidProject) {
		t.Fatalf("invalid color err=%v", err)
	}
}

func TestMemoryStoreTimelineHistorySurvivesSaveUndoRedoAndBranch(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 10, 3, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-history", Title: "历史", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	first, err := store.SaveTimeline(ctx, "alice", "cut-history", Timeline{Script: "第一版", Tracks: []Track{}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.SaveTimeline(ctx, "alice", "cut-history", Timeline{Script: "第二版", Tracks: []Track{}}, first.Revision, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.TimelineHistory(ctx, "alice", "cut-history")
	if err != nil || !state.CanUndo || state.CanRedo {
		t.Fatalf("state=%+v err=%v", state, err)
	}
	undone, state, err := store.UndoTimeline(ctx, "alice", "cut-history", second.Revision, now.Add(2*time.Second))
	if err != nil || undone.Script != "第一版" || undone.Revision != 3 || !state.CanUndo || !state.CanRedo {
		t.Fatalf("undo=%+v state=%+v err=%v", undone, state, err)
	}
	redone, state, err := store.RedoTimeline(ctx, "alice", "cut-history", undone.Revision, now.Add(3*time.Second))
	if err != nil || redone.Script != "第二版" || redone.Revision != 4 || !state.CanUndo || state.CanRedo {
		t.Fatalf("redo=%+v state=%+v err=%v", redone, state, err)
	}
	undone, _, err = store.UndoTimeline(ctx, "alice", "cut-history", redone.Revision, now.Add(4*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	branched, err := store.SaveTimeline(ctx, "alice", "cut-history", Timeline{Script: "分支版", Tracks: []Track{}}, undone.Revision, now.Add(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	state, err = store.TimelineHistory(ctx, "alice", "cut-history")
	if err != nil || state.CanRedo || branched.Script != "分支版" {
		t.Fatalf("branch=%+v state=%+v err=%v", branched, state, err)
	}
	if _, _, err := store.RedoTimeline(ctx, "alice", "cut-history", branched.Revision, now); !errors.Is(err, ErrTimelineHistoryBoundary) {
		t.Fatalf("redo boundary=%v", err)
	}
}

func TestMemoryStoreRejectsInvalidClipBounds(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 9, 14, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-1", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	_, err := store.SaveTimeline(ctx, "alice", "cut-1", Timeline{Tracks: []Track{{
		ID: "video-1", Kind: TrackVideo, Clips: []Clip{{ID: "clip-1", Kind: ClipVideo, StartMS: -1, DurationMS: 1000}},
	}}}, 0, now)
	if !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("err=%v, want ErrInvalidTimeline", err)
	}
}

func TestTimelinePersistsScriptSubtitleStyleAndReviewMarkers(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 10, 2, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "cut-1", Title: "测试", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	timeline := Timeline{Script: "第一场", Tracks: []Track{
		{ID: "subtitle-1", Kind: TrackSubtitle, Name: "字幕1", Settings: TrackSettings{SubtitleStyleID: "yellow-black", SubtitleRhythm: "word", SubtitleAnimation: "pop", PreviewVisible: true}},
		{ID: "video-1", Kind: TrackVideo, Name: "视频1", Clips: []Clip{{ID: "clip-1", Kind: ClipVideo, Name: "镜头", DurationMS: 1000, SourceDurationMS: 1000, Transition: "flash-black", TransitionMS: 500, ReviewStatus: "review", ReviewReason: "需要人工确认", LinkGroupID: "link-1"}}},
	}}
	saved, err := store.SaveTimeline(ctx, "alice", "cut-1", timeline, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Timeline(ctx, "alice", "cut-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Script != "第一场" || loaded.Tracks[0].Settings.SubtitleStyleID != "yellow-black" || loaded.Tracks[1].Clips[0].ReviewStatus != "review" || loaded.Tracks[1].Clips[0].TransitionMS != 500 || loaded.Tracks[1].Clips[0].LinkGroupID != "link-1" {
		t.Fatalf("loaded=%+v saved=%+v", loaded, saved)
	}
}

func TestTimelinePersistsRichTextTrackSettings(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 11, 5, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	if err := store.CreateProject(ctx, "alice", Project{ID: "text-rich", Title: "文字", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	timeline := Timeline{Tracks: []Track{{ID: "text-1", Kind: TrackSubtitle, Settings: TrackSettings{
		FontFamily: "一战测试字体", FontAssetID: "aicut-asset-font1", FontSize: 72, TextColor: "#ffcc00", StrokeColor: "#112233", StrokeWidth: 3,
		BackgroundColor: "#00000080", ShadowColor: "#223344cc", ShadowOffsetX: 4, ShadowOffsetY: 6, ShadowBlur: 8,
		LetterSpacing: 4, LineHeight: 1.4, TextAlign: "left", TextCurve: 120, SubtitleAnimation: "typewriter",
		TextRuns: []TextStyleRun{{Start: 1, End: 3, TextColor: "#ff0000", FontSize: 80, Bold: true, Italic: true}},
	}, Clips: []Clip{{ID: "c", Kind: ClipSubtitle, DurationMS: 1000, Text: "剪辑开始"}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", "text-rich", timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Timeline(ctx, "alice", "text-rich")
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Tracks[0].Settings; got.FontFamily != "一战测试字体" || got.FontAssetID != "aicut-asset-font1" || got.FontSize != 72 || got.TextColor != "#ffcc00" || got.LetterSpacing != 4 || got.TextAlign != "left" || got.TextCurve != 120 || got.ShadowColor != "#223344cc" || got.ShadowOffsetX != 4 || got.ShadowOffsetY != 6 || got.ShadowBlur != 8 || len(got.TextRuns) != 1 || !got.TextRuns[0].Bold {
		t.Fatalf("settings not persisted: %+v", got)
	}
}

func TestTimelineRejectsInvalidRichTextTrackSettings(t *testing.T) {
	for _, settings := range []TrackSettings{{FontSize: 500, TextColor: "red"}, {ShadowColor: "red"}, {ShadowOffsetX: 41}, {ShadowOffsetY: -41}, {ShadowBlur: 21}, {TextCurve: 181}, {TextRuns: []TextStyleRun{{Start: 3, End: 2}}}, {TextRuns: []TextStyleRun{{Start: 0, End: 2, TextColor: "red"}}}} {
		err := validateTimeline(Timeline{Tracks: []Track{{ID: "text-1", Kind: TrackSubtitle, Settings: settings, Clips: []Clip{{ID: "c", Kind: ClipSubtitle, DurationMS: 1000, Text: "测试"}}}}})
		if !errors.Is(err, ErrInvalidTimeline) {
			t.Fatalf("settings=%+v err=%v want invalid timeline", settings, err)
		}
	}
}

func TestMemoryStorePersistsAudioProcessing(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	now := time.Now()
	project := Project{ID: "audio-processing", Title: "音频处理", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	timeline := Timeline{Tracks: []Track{{ID: "audio-1", Kind: TrackAudio, Name: "音频1", Clips: []Clip{{
		ID: "clip-1", Kind: ClipAudio, Name: "对白", DurationMS: 1000, SourceDurationMS: 1000,
		AudioProcessing: &AudioProcessing{Denoise: .6, EQLowDB: 2, EQMidDB: -1, EQHighDB: 3, ChannelMode: "mono", LoudnessEnabled: true, TargetLUFS: -14, VoiceEffect: "robot"},
	}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", project.ID, timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Timeline(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	processing := loaded.Tracks[0].Clips[0].AudioProcessing
	if processing == nil || processing.VoiceEffect != "robot" {
		t.Fatalf("voice effect not persisted: %#v", processing)
	}
	if processing == nil || processing.ChannelMode != "mono" || processing.EQHighDB != 3 || !processing.LoudnessEnabled || processing.TargetLUFS != -14 {
		t.Fatalf("audio processing not persisted: %#v", processing)
	}
}

func TestTimelineRejectsInvalidAudioProcessing(t *testing.T) {
	base := Timeline{Tracks: []Track{{ID: "audio-1", Kind: TrackAudio, Clips: []Clip{{ID: "clip-1", Kind: ClipAudio, DurationMS: 1000, SourceDurationMS: 1000}}}}}
	for _, processing := range []*AudioProcessing{{Denoise: 1.1}, {EQLowDB: -13}, {EQMidDB: 13}, {ChannelMode: "surround"}, {VoiceEffect: "alien"}, {PitchSemitones: 12.1}, {PitchSemitones: -12.1}, {LoudnessEnabled: true, TargetLUFS: -7}, {LoudnessEnabled: true, TargetLUFS: -25}} {
		invalid := base
		invalid.Tracks[0].Clips[0].AudioProcessing = processing
		if err := validateTimeline(invalid); err == nil {
			t.Fatalf("accepted invalid processing: %#v", processing)
		}
	}
	video := Timeline{Tracks: []Track{{ID: "video-1", Kind: TrackVideo, Clips: []Clip{{ID: "clip-v", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, AudioProcessing: &AudioProcessing{Denoise: .5}}}}}}
	if err := validateTimeline(video); err == nil {
		t.Fatal("accepted audio processing on video")
	}
}

func TestTimelineAcceptsBoundedPitchSemitones(t *testing.T) {
	for _, pitch := range []float64{-12, -7.5, 0, 4, 12} {
		timeline := Timeline{Tracks: []Track{{ID: "audio-1", Kind: TrackAudio, Clips: []Clip{{ID: "clip-1", Kind: ClipAudio, DurationMS: 1000, SourceDurationMS: 1000, AudioProcessing: &AudioProcessing{PitchSemitones: pitch}}}}}}
		if err := validateTimeline(timeline); err != nil {
			t.Fatalf("pitch %.1f rejected: %v", pitch, err)
		}
	}
}

func TestTimelineAcceptsExtendedAudioEffectPresets(t *testing.T) {
	effects := []string{"bright", "bass", "treble", "radio", "megaphone", "deep", "sweet", "whisper", "distorted", "lofi", "echo", "room", "hall", "plate", "cathedral", "stadium", "underwater", "vinyl"}
	for _, effect := range effects {
		timeline := Timeline{Tracks: []Track{{ID: "audio-1", Kind: TrackAudio, Clips: []Clip{{ID: "clip-1", Kind: ClipAudio, DurationMS: 1000, SourceDurationMS: 1000, AudioProcessing: &AudioProcessing{VoiceEffect: effect}}}}}}
		if err := validateTimeline(timeline); err != nil {
			t.Fatalf("preset %s rejected: %v", effect, err)
		}
	}
}

func TestTimelinePersistsAndValidatesVisualProcessing(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	now := time.Now()
	project := Project{ID: "visual-processing", Title: "高级画面", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	processing := &VisualProcessing{
		FrameInterpolation: "optical-flow",
		Stabilization:      VideoStabilization{Enabled: true, Smoothness: 60, Zoom: 12, Tripod: true},
		Color:              ColorAdjustment{Exposure: .5, Temperature: .2, Tint: -.1, Highlights: .3, Shadows: -.2, Gamma: 1.1, Curve: []float64{0, .2, .55, .85, 1}, CurvePoints: []ColorCurvePoint{{X: 0, Y: 0}, {X: .18, Y: .3}, {X: .63, Y: .52}, {X: 1, Y: 1}}},
		ChromaKey:          ChromaKey{Enabled: true, RemoveBlack: true, Color: "#00ff00", Similarity: .25, Blend: .1},
		PersonSegmentation: PersonSegmentation{Enabled: true, MatteAssetID: "mask-1", Threshold: .55, Feather: .1, Signature: "video|0|1000|1|0|none"},
		ObjectRemoval:      ObjectRemoval{Enabled: true, X: .42, Y: .36, Width: .24, Height: .18},
		Mask:               VisualMask{Type: "ellipse", X: .5, Y: .5, Width: .8, Height: .7, Feather: .1},
		Geometry:           VisualGeometry{CropLeft: .1, CropTop: .05, CropRight: .1, CropBottom: .05, FlipHorizontal: true},
		Details:            VisualDetails{Sharpen: .4, Vignette: .5, Grain: .2},
	}
	timeline := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{ID: "c", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: processing}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", project.ID, timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Timeline(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Tracks[0].Clips[0].VisualProcessing; got == nil || got.FrameInterpolation != "optical-flow" || !got.Stabilization.Enabled || got.Stabilization.Smoothness != 60 || got.Stabilization.Zoom != 12 || !got.Stabilization.Tripod || got.Mask.Type != "ellipse" || got.ChromaKey.Color != "#00ff00" || !got.ChromaKey.RemoveBlack || !got.PersonSegmentation.Enabled || got.PersonSegmentation.MatteAssetID != "mask-1" || !got.ObjectRemoval.Enabled || got.ObjectRemoval.X != .42 || got.ObjectRemoval.Height != .18 || got.Color.Exposure != .5 || got.Geometry.CropLeft != .1 || got.Geometry.CropBottom != .05 || !got.Geometry.FlipHorizontal || got.Details.Sharpen != .4 || got.Details.Vignette != .5 || got.Details.Grain != .2 {
		t.Fatalf("visual processing not persisted: %#v", got)
	}
	invalid := timeline
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Color: ColorAdjustment{Exposure: 3}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid visual processing")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Color: ColorAdjustment{Curve: []float64{0, .5, 1}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid color curve")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Color: ColorAdjustment{CurvePoints: []ColorCurvePoint{{X: 0, Y: 0}, {X: .6, Y: .4}, {X: .5, Y: .7}, {X: 1, Y: 1}}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted unsorted color curve points")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Mask: VisualMask{Type: "ellipse", X: .5, Y: .5, Width: 1, Height: 1, Feather: .6}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid mask feather")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Mask: VisualMask{Type: "polygon", Points: []VisualMaskPoint{{X: .1, Y: .1}, {X: .9, Y: .1}}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted polygon mask with too few points")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Mask: VisualMask{Type: "polygon", Points: []VisualMaskPoint{{X: .1, Y: .1}, {X: .5, Y: .5}, {X: .9, Y: .9}}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted polygon mask without area")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Mask: VisualMask{Type: "bezier", BezierPoints: []VisualBezierPoint{{X: .1, Y: .1}}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted bezier mask with too few anchors")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Mask: VisualMask{Type: "bezier", BezierPoints: []VisualBezierPoint{{X: .1, Y: .1, InX: -1, InY: .1, OutX: .2, OutY: .1}, {X: .9, Y: .1, InX: .8, InY: .1, OutX: 1, OutY: .2}, {X: .5, Y: .9, InX: .6, InY: .8, OutX: .4, OutY: .8}}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted bezier mask with invalid handle")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Geometry: VisualGeometry{CropLeft: .6, CropRight: .5}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid horizontal crop")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Details: VisualDetails{Sharpen: 1.1}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid sharpen strength")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Details: VisualDetails{Vignette: -.1}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid vignette strength")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{Details: VisualDetails{Grain: math.NaN()}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid grain strength")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{ObjectRemoval: ObjectRemoval{Enabled: true, X: .5, Y: .5, Width: .01, Height: .2}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid object removal region")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{FrameInterpolation: "fake"}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid frame interpolation mode")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{PersonSegmentation: PersonSegmentation{Enabled: true, Threshold: .5, Feather: .1, Signature: "sig"}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted missing person matte")
	}
	invalid.Tracks[0].Clips[0].VisualProcessing = &VisualProcessing{PersonSegmentation: PersonSegmentation{Enabled: true, MatteAssetID: "mask", Threshold: 1.1, Feather: .1, Signature: "sig"}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid person threshold")
	}
	invalid.Tracks[0].Clips[0] = Clip{ID: "image", Kind: ClipImage, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &VisualProcessing{FrameInterpolation: "blend"}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted frame interpolation on image")
	}
	invalid.Tracks[0].Clips[0] = Clip{ID: "video", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &VisualProcessing{Stabilization: VideoStabilization{Enabled: true, Smoothness: 0}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted invalid stabilization smoothness")
	}
	invalid.Tracks[0].Clips[0] = Clip{ID: "image", Kind: ClipImage, DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &VisualProcessing{Stabilization: VideoStabilization{Enabled: true, Smoothness: 15}}}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted stabilization on image")
	}
}

func TestTimelinePersistsAndValidatesKeyframeEasing(t *testing.T) {
	custom := &KeyframeEasing{Type: "custom", X1: .2, Y1: .1, X2: .8, Y2: .9}
	timeline := Timeline{Tracks: []Track{{
		ID: "v", Kind: TrackVideo, Clips: []Clip{{
			ID: "c", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000,
			Keyframes: []ClipKeyframe{
				{ID: "k1", OffsetMS: 0, Values: KeyframeValues{}, Easing: custom},
				{ID: "k2", OffsetMS: 1000, Values: KeyframeValues{}},
			},
		}},
	}}}
	if err := validateTimeline(timeline); err != nil {
		t.Fatalf("valid custom easing rejected: %v", err)
	}
	invalid := timeline
	invalid.Tracks[0].Clips[0].Keyframes[0].Easing = &KeyframeEasing{Type: "custom", X1: -1, Y1: 0, X2: 1, Y2: 1}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted out-of-range easing handle")
	}
	invalid.Tracks[0].Clips[0].Keyframes[0].Easing = &KeyframeEasing{Type: "bounce", X1: 0, Y1: 0, X2: 1, Y2: 1}
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted unknown easing type")
	}
	duplicateOffset := Timeline{Tracks: []Track{{
		ID: "v", Kind: TrackVideo, Clips: []Clip{{
			ID: "c", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 1000,
			Keyframes: []ClipKeyframe{
				{ID: "k1", OffsetMS: 500, Values: KeyframeValues{}},
				{ID: "k2", OffsetMS: 500, Values: KeyframeValues{}},
			},
		}},
	}}}
	if err := validateTimeline(duplicateOffset); err == nil {
		t.Fatal("accepted two keyframes at the same offset")
	}
}

func TestTimelinePersistsAndValidatesReverseAndFreezeFrame(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	now := time.Now()
	project := Project{ID: "reverse-freeze", Title: "倒放定格", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	freezeSourceMS := int64(500)
	timeline := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{ID: "c", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 2000, FreezeSourceMS: &freezeSourceMS}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", project.ID, timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Timeline(ctx, "alice", project.ID)
	if err != nil || loaded.Tracks[0].Clips[0].FreezeSourceMS == nil || *loaded.Tracks[0].Clips[0].FreezeSourceMS != 500 {
		t.Fatalf("freeze frame not persisted: %#v err=%v", loaded, err)
	}
	invalid := timeline
	invalid.Tracks[0].Clips[0].FreezeSourceMS = nil
	invalid.Tracks[0].Clips[0].Reversed = true
	if err := validateTimeline(invalid); err != nil {
		t.Fatalf("valid reverse rejected: %v", err)
	}
	invalid.Tracks[0].Clips[0].FreezeSourceMS = &freezeSourceMS
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted simultaneous reverse and freeze")
	}
	invalid.Tracks[0].Clips[0].Reversed = false
	badFreeze := int64(2000)
	invalid.Tracks[0].Clips[0].FreezeSourceMS = &badFreeze
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted freeze frame outside source")
	}
}

func TestTimelinePersistsAndValidatesBlendMode(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	now := time.Now()
	project := Project{ID: "blend-mode", Title: "混合模式", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	timeline := Timeline{Tracks: []Track{{ID: "v", Kind: TrackVideo, Clips: []Clip{{ID: "c", Kind: ClipImage, DurationMS: 1000, SourceDurationMS: 1000, BlendMode: "multiply"}}}}}
	if _, err := store.SaveTimeline(ctx, "alice", project.ID, timeline, 0, now); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Timeline(ctx, "alice", project.ID)
	if err != nil || loaded.Tracks[0].Clips[0].BlendMode != "multiply" {
		t.Fatalf("blend mode not persisted: %#v err=%v", loaded, err)
	}
	for _, blendMode := range []string{"normal", "multiply", "screen", "overlay", "soft-light", "difference", "add"} {
		valid := timeline
		valid.Tracks[0].Clips[0].BlendMode = blendMode
		if err := validateTimeline(valid); err != nil {
			t.Fatalf("valid blend mode %q rejected: %v", blendMode, err)
		}
	}
	invalid := timeline
	invalid.Tracks[0].Clips[0].BlendMode = "unknown"
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted unknown blend mode")
	}
	invalid.Tracks[0].Clips[0].BlendMode = "screen"
	invalid.Tracks[0].Clips[0].Kind = ClipAudio
	if err := validateTimeline(invalid); err == nil {
		t.Fatal("accepted blend mode on audio")
	}
}

func TestTimelineRejectsUnknownReviewStatus(t *testing.T) {
	err := validateTimeline(Timeline{Tracks: []Track{{ID: "video-1", Kind: TrackVideo, Clips: []Clip{{ID: "clip-1", Kind: ClipVideo, DurationMS: 1000, ReviewStatus: "maybe"}}}}})
	if !errors.Is(err, ErrInvalidTimeline) {
		t.Fatalf("err=%v, want ErrInvalidTimeline", err)
	}
}
