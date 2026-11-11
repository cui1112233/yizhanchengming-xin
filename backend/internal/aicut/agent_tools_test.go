package aicut

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func newAgentToolFixture(t *testing.T, locked bool) (*MemoryStore, *MemoryAgentStore, AgentRun, *AgentToolExecutor) {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2026, 11, 10, 6, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	project := Project{ID: "cut-1", Title: "测试", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	timeline, err := store.SaveTimeline(ctx, "alice", project.ID, Timeline{Tracks: []Track{{ID: "video-1", Kind: TrackVideo, Name: "视频1", Locked: locked, Clips: []Clip{{ID: "clip-a", Kind: ClipVideo, Name: "A", StartMS: 0, DurationMS: 2000, SourceDurationMS: 5000}, {ID: "clip-b", Kind: ClipVideo, Name: "B", StartMS: 5000, DurationMS: 2000, SourceDurationMS: 5000}}}, {ID: "subtitle-1", Kind: TrackSubtitle, Name: "字幕1", Clips: []Clip{{ID: "subtitle-2", Kind: ClipSubtitle, Name: "字幕", Text: "旧字幕", StartMS: 0, DurationMS: 1000}}}}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	agents := NewMemoryAgentStore()
	run := AgentRun{ID: "run-1", ClientRequestID: "request-1", Owner: "alice", ProjectID: project.ID, Instruction: "编辑", Mode: AgentModeAgent, ExpectedRevision: timeline.Revision, State: AgentQueued, CreatedAt: now, UpdatedAt: now}
	if _, err := agents.CreateAgentRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	return store, agents, run, NewAgentToolExecutor(store, agents, func() time.Time { return now.Add(time.Minute) })
}

func toolArgs(value any) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

func TestAgentToolMoveClipChangesTimelineRevision(t *testing.T) {
	store, _, run, executor := newAgentToolFixture(t, false)
	result, err := executor.Execute(context.Background(), run, ToolCall{ID: "call-1", Name: "move_clip", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-b", "startMs": 3000})})
	if err != nil || !result.OK || result.Revision != run.ExpectedRevision+1 {
		t.Fatalf("Execute(move_clip) = %#v, %v", result, err)
	}
	timeline, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
	if timeline.Tracks[0].Clips[1].StartMS != 3000 {
		t.Fatalf("clip start = %d, want 3000", timeline.Tracks[0].Clips[1].StartMS)
	}
}

func TestAgentToolDeleteClipRejectsLockedTrackAndPreservesTimeline(t *testing.T) {
	store, _, run, executor := newAgentToolFixture(t, true)
	result, err := executor.Execute(context.Background(), run, ToolCall{ID: "call-1", Name: "delete_clip", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-b"})})
	if err != nil || result.OK || result.Message != "轨道已锁定，不能删除片段" {
		t.Fatalf("Execute(delete locked) = %#v, %v", result, err)
	}
	timeline, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
	if timeline.Revision != run.ExpectedRevision || len(timeline.Tracks[0].Clips) != 2 {
		t.Fatalf("locked timeline changed: %#v", timeline)
	}
}

func TestAgentToolMoveClipRejectsStaleRevision(t *testing.T) {
	store, _, run, executor := newAgentToolFixture(t, false)
	current, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
	if _, err := store.SaveTimeline(context.Background(), run.Owner, run.ProjectID, current, current.Revision, time.Now()); err != nil {
		t.Fatal(err)
	}
	result, err := executor.Execute(context.Background(), run, ToolCall{ID: "call-1", Name: "move_clip", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-b", "startMs": 3000})})
	if err != nil || result.OK || result.Message != "时间线已更新，请刷新后重试" {
		t.Fatalf("Execute(stale move) = %#v, %v", result, err)
	}
}

func TestAgentToolCallRejectsUnknownTool(t *testing.T) {
	_, _, run, executor := newAgentToolFixture(t, false)
	result, err := executor.Execute(context.Background(), run, ToolCall{ID: "call-1", Name: "run_sql", Arguments: toolArgs(map[string]any{})})
	if err != nil || result.OK || result.Message != "不允许调用工具：run_sql" {
		t.Fatalf("Execute(unknown) = %#v, %v", result, err)
	}
}

func TestAgentToolCallReplayDoesNotApplyTwice(t *testing.T) {
	store, _, run, executor := newAgentToolFixture(t, false)
	call := ToolCall{ID: "call-1", Name: "delete_clip", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-b"})}
	first, err := executor.Execute(context.Background(), run, call)
	if err != nil || !first.OK {
		t.Fatalf("first Execute() = %#v, %v", first, err)
	}
	second, err := executor.Execute(context.Background(), run, call)
	if err != nil || !second.OK || second.Revision != first.Revision {
		t.Fatalf("replayed Execute() = %#v, %v", second, err)
	}
	timeline, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
	if len(timeline.Tracks[0].Clips) != 1 || timeline.Revision != first.Revision {
		t.Fatalf("duplicate tool applied twice: %#v", timeline)
	}
}

func TestAgentTimelineEditingToolsMatchWorkbenchRules(t *testing.T) {
	ctx := context.Background()
	store, _, run, executor := newAgentToolFixture(t, false)
	split, err := executor.Execute(ctx, run, ToolCall{ID: "split-1", Name: "split_clip", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-a", "offsetMs": 800})})
	if err != nil || !split.OK {
		t.Fatalf("split_clip = %#v, %v", split, err)
	}
	run.ExpectedRevision = split.Revision
	speed, err := executor.Execute(ctx, run, ToolCall{ID: "speed-1", Name: "set_playback_rate", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-b", "playbackRate": 2})})
	if err != nil || !speed.OK {
		t.Fatalf("set_playback_rate = %#v, %v", speed, err)
	}
	run.ExpectedRevision = speed.Revision
	replace, err := executor.Execute(ctx, run, ToolCall{ID: "replace-1", Name: "replace_subtitle_text", Arguments: toolArgs(map[string]any{"from": "旧", "to": "新"})})
	if err != nil || !replace.OK {
		t.Fatalf("replace_subtitle_text = %#v, %v", replace, err)
	}
	timeline, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	if len(timeline.Tracks[0].Clips) != 3 || timeline.Tracks[0].Clips[0].DurationMS != 800 || timeline.Tracks[0].Clips[2].DurationMS != 1000 || timeline.Tracks[1].Clips[0].Text != "新字幕" {
		t.Fatalf("edited timeline = %#v", timeline)
	}
}

func TestAgentPlaceAssetAndUndoRedo(t *testing.T) {
	ctx := context.Background()
	store, _, run, executor := newAgentToolFixture(t, false)
	asset := Asset{ID: "asset-audio", ProjectID: run.ProjectID, Kind: "audio", Name: "BGM", ContentType: "audio/mpeg", ObjectKey: "a.mp3", ByteSize: 12, DurationMS: 3000, CreatedAt: time.Now()}
	if err := store.CreateAsset(ctx, run.Owner, asset); err != nil {
		t.Fatal(err)
	}
	placed, err := executor.Execute(ctx, run, ToolCall{ID: "place-1", Name: "place_asset", Arguments: toolArgs(map[string]any{"assetId": asset.ID, "trackId": "audio-1", "startMs": 1000})})
	if err != nil || !placed.OK {
		t.Fatalf("place_asset = %#v, %v", placed, err)
	}
	run.ExpectedRevision = placed.Revision
	undone, err := executor.Execute(ctx, run, ToolCall{ID: "undo-1", Name: "undo_timeline", Arguments: toolArgs(map[string]any{})})
	if err != nil || !undone.OK {
		t.Fatalf("undo_timeline = %#v, %v", undone, err)
	}
	run.ExpectedRevision = undone.Revision
	redone, err := executor.Execute(ctx, run, ToolCall{ID: "redo-1", Name: "redo_timeline", Arguments: toolArgs(map[string]any{})})
	if err != nil || !redone.OK {
		t.Fatalf("redo_timeline = %#v, %v", redone, err)
	}
	timeline, _ := store.Timeline(ctx, run.Owner, run.ProjectID)
	track, found := findTimelineTrack(&timeline, "audio-1")
	if !found || len(track.Clips) != 1 || track.Clips[0].AssetID != asset.ID {
		t.Fatalf("placed asset missing after redo: %#v", timeline)
	}
}

func TestAgentSplitRejectsBoundaryWithoutMutation(t *testing.T) {
	store, _, run, executor := newAgentToolFixture(t, false)
	result, err := executor.Execute(context.Background(), run, ToolCall{ID: "split-bad", Name: "split_clip", Arguments: toolArgs(map[string]any{"trackId": "video-1", "clipId": "clip-a", "offsetMs": 2000})})
	if err != nil || result.OK || result.Message != "切分位置必须在片段内部" {
		t.Fatalf("split boundary = %#v, %v", result, err)
	}
	timeline, _ := store.Timeline(context.Background(), run.Owner, run.ProjectID)
	if timeline.Revision != run.ExpectedRevision || len(timeline.Tracks[0].Clips) != 2 {
		t.Fatalf("invalid split changed timeline: %#v", timeline)
	}
}

func TestAgentCompositeToolsCallConfiguredServices(t *testing.T) {
	_, _, run, executor := newAgentToolFixture(t, false)
	var asrCalls, splitCalls, extractCalls, cleanCalls int
	executor.ConfigureCompositeTools(AgentCompositeTools{
		SubmitASR: func(_ context.Context, owner string, input ASRSubmit) (ASRJob, error) {
			asrCalls++
			if owner != run.Owner || input.ProjectID != run.ProjectID || input.AssetID != "video-a" || input.ModelID != "whisper" || input.Language != "zh" || input.ExpectedRevision != run.ExpectedRevision {
				t.Fatalf("unexpected ASR input: owner=%q input=%#v", owner, input)
			}
			return ASRJob{ID: "asr-1", ProjectID: run.ProjectID, AssetID: input.AssetID, State: ASRQueued}, nil
		},
		SmartSplit: func(_ context.Context, owner, projectID, assetID string, threshold float64, minIntervalMS, expectedRevision int64) (Timeline, int, string, error) {
			splitCalls++
			if owner != run.Owner || projectID != run.ProjectID || assetID != "video-a" || threshold != 0.4 || minIntervalMS != 2000 || expectedRevision != run.ExpectedRevision {
				t.Fatalf("unexpected smart split input: %q %q %q %v %d %d", owner, projectID, assetID, threshold, minIntervalMS, expectedRevision)
			}
			return Timeline{Revision: run.ExpectedRevision + 1}, 3, "已智能切出 3 个新片段", nil
		},
		ExtractAudio: func(_ context.Context, owner, projectID, assetID string) (Asset, error) {
			extractCalls++
			if owner != run.Owner || projectID != run.ProjectID || assetID != "video-a" {
				t.Fatalf("unexpected extract input: %q %q %q", owner, projectID, assetID)
			}
			return Asset{ID: "audio-1", ProjectID: projectID, Kind: "audio", Name: "video-a.wav"}, nil
		},
		OralClean: func(_ context.Context, owner, projectID, modelID string, maxPauseMS, expectedRevision int64) (Timeline, FillerTrimSummary, string, error) {
			cleanCalls++
			if owner != run.Owner || projectID != run.ProjectID || modelID != "text-1" || maxPauseMS != 1000 || expectedRevision != run.ExpectedRevision {
				t.Fatalf("unexpected oral clean input: %q %q %q %d %d", owner, projectID, modelID, maxPauseMS, expectedRevision)
			}
			return Timeline{Revision: run.ExpectedRevision + 1}, FillerTrimSummary{FillerCount: 1}, "去口误停顿完成", nil
		},
	})

	checks := []struct {
		call         ToolCall
		wantRevision int64
	}{
		{ToolCall{ID: "asr-call", Name: "transcribe_asset", Arguments: toolArgs(map[string]any{"assetId": "video-a", "modelId": "whisper", "language": "zh"})}, 0},
		{ToolCall{ID: "split-call", Name: "smart_split_asset", Arguments: toolArgs(map[string]any{"assetId": "video-a", "threshold": 0.4, "minIntervalMs": 2000})}, run.ExpectedRevision + 1},
		{ToolCall{ID: "extract-call", Name: "extract_audio", Arguments: toolArgs(map[string]any{"assetId": "video-a"})}, 0},
		{ToolCall{ID: "clean-call", Name: "clean_oral_pauses", Arguments: toolArgs(map[string]any{"modelId": "text-1", "maxPauseMs": 1000})}, run.ExpectedRevision + 1},
	}
	for _, check := range checks {
		result, err := executor.Execute(context.Background(), run, check.call)
		if err != nil || !result.OK || result.Revision != check.wantRevision || len(result.Data) == 0 {
			t.Fatalf("Execute(%s) = %#v, %v", check.call.Name, result, err)
		}
	}
	if asrCalls != 1 || splitCalls != 1 || extractCalls != 1 || cleanCalls != 1 {
		t.Fatalf("composite calls = asr:%d split:%d extract:%d clean:%d", asrCalls, splitCalls, extractCalls, cleanCalls)
	}
}

func TestAgentCompositeToolsRejectInvalidArgumentsWithoutCallingService(t *testing.T) {
	_, _, run, executor := newAgentToolFixture(t, false)
	called := 0
	executor.ConfigureCompositeTools(AgentCompositeTools{
		SmartSplit: func(context.Context, string, string, string, float64, int64, int64) (Timeline, int, string, error) {
			called++
			return Timeline{}, 0, "", nil
		},
	})
	result, err := executor.Execute(context.Background(), run, ToolCall{ID: "bad-split", Name: "smart_split_asset", Arguments: toolArgs(map[string]any{"assetId": "video-a", "threshold": 2.0, "minIntervalMs": 10, "unknown": true})})
	if err != nil || result.OK || result.Message != "智能切镜参数无效" || called != 0 {
		t.Fatalf("invalid smart split = %#v, %v, called=%d", result, err, called)
	}
}

func TestAgentCompositeToolReplayDoesNotRunServiceTwice(t *testing.T) {
	_, _, run, executor := newAgentToolFixture(t, false)
	calls := 0
	executor.ConfigureCompositeTools(AgentCompositeTools{
		ExtractAudio: func(context.Context, string, string, string) (Asset, error) {
			calls++
			return Asset{ID: "audio-1", ProjectID: run.ProjectID, Kind: "audio", Name: "audio.wav"}, nil
		},
	})
	call := ToolCall{ID: "extract-once", Name: "extract_audio", Arguments: toolArgs(map[string]any{"assetId": "video-a"})}
	first, err := executor.Execute(context.Background(), run, call)
	if err != nil || !first.OK {
		t.Fatalf("first extract = %#v, %v", first, err)
	}
	second, err := executor.Execute(context.Background(), run, call)
	if err != nil || !second.OK || calls != 1 {
		t.Fatalf("replayed extract = %#v, %v, calls=%d", second, err, calls)
	}
}
