package aicut

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AgentToolExecutor struct {
	collection *TimelineCollectionStore
	timelines  Store
	jobStore   Store
	agents     AgentStore
	now        func() time.Time
	composite  AgentCompositeTools
}

type AgentCompositeTools struct {
	ListASRModels         func(context.Context, string) ([]AgentASRModel, error)
	SmartSplitForTimeline func(context.Context, string, string, string, string, float64, int64, int64) (Timeline, int, string, error)
	OralCleanForTimeline  func(context.Context, string, string, string, string, int64, int64) (Timeline, FillerTrimSummary, string, error)
	SubmitASR             func(context.Context, string, ASRSubmit) (ASRJob, error)
	SmartSplit            func(context.Context, string, string, string, float64, int64, int64) (Timeline, int, string, error)
	ExtractAudio          func(context.Context, string, string, string) (Asset, error)
	OralClean             func(context.Context, string, string, string, int64, int64) (Timeline, FillerTrimSummary, string, error)
}

func NewAgentToolExecutor(timelines Store, agents AgentStore, now func() time.Time, collections ...*TimelineCollectionStore) *AgentToolExecutor {
	if now == nil {
		now = time.Now
	}
	var collection *TimelineCollectionStore
	if len(collections) > 0 {
		collection = collections[0]
	}
	return &AgentToolExecutor{jobStore: timelines, timelines: timelines, agents: agents, now: now, collection: collection}
}

func (e *AgentToolExecutor) ConfigureCompositeTools(tools AgentCompositeTools) {
	if e != nil {
		e.composite = tools
	}
}

func decodeToolArguments(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("unexpected trailing data")
	}
	return nil
}

func (e *AgentToolExecutor) Execute(ctx context.Context, run AgentRun, call ToolCall) (ToolResult, error) {
	persisted, err := e.agents.AgentRun(ctx, run.Owner, run.ProjectID, run.ID)
	if err != nil {
		return ToolResult{}, err
	}
	if canonicalASRTimeline(persisted.TimelineID) != canonicalASRTimeline(run.TimelineID) {
		return ToolResult{}, ErrAgentRunNotFound
	}
	if canonicalASRTimeline(run.TimelineID) != DefaultTimelineID {
		if e.collection == nil {
			return ToolResult{}, ErrTimelineJobUnsupported
		}
		scoped, err := e.collection.ScopedStore(ctx, run.Owner, run.ProjectID, run.TimelineID)
		if err != nil {
			return ToolResult{}, err
		}
		local := *e
		local.timelines = scoped
		if named := e.composite.SmartSplitForTimeline; named != nil {
			local.composite.SmartSplit = func(ctx context.Context, o, p, asset string, threshold float64, interval, revision int64) (Timeline, int, string, error) {
				return named(ctx, o, p, run.TimelineID, asset, threshold, interval, revision)
			}
		} else {
			local.composite.SmartSplit = nil
		}
		if named := e.composite.OralCleanForTimeline; named != nil {
			local.composite.OralClean = func(ctx context.Context, o, p, model string, pause, revision int64) (Timeline, FillerTrimSummary, string, error) {
				return named(ctx, o, p, run.TimelineID, model, pause, revision)
			}
		} else {
			local.composite.OralClean = nil
		}
		e = &local
	}
	call.ID, call.Name = strings.TrimSpace(call.ID), strings.TrimSpace(call.Name)
	if call.ID == "" || call.Name == "" {
		return ToolResult{}, ErrInvalidAgentRun
	}
	if saved, ok, err := e.agents.AgentToolResult(ctx, run.Owner, run.ProjectID, run.ID, call.ID); err != nil {
		return ToolResult{}, err
	} else if ok {
		return saved, nil
	}
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var executeErr error
	switch call.Name {
	case "get_transcript", "select_caption_ranges", "compact_caption_gaps", "set_clip_animation", "set_clip_color", "apply_edit_operations":
		result = e.smartEdit(ctx, run, call)
	case "move_clip":
		result = e.moveClip(ctx, run, call)
	case "delete_clip":
		result = e.deleteClip(ctx, run, call)
	case "split_clip":
		result = e.splitClip(ctx, run, call)
	case "set_playback_rate":
		result = e.setPlaybackRate(ctx, run, call)
	case "replace_subtitle_text":
		result = e.replaceSubtitleText(ctx, run, call)
	case "place_asset", "add_bgm":
		result = e.placeAsset(ctx, run, call)
	case "undo_timeline":
		result = e.undoRedo(ctx, run, call, false)
	case "redo_timeline":
		result = e.undoRedo(ctx, run, call, true)
	case "list_asr_models", "get_asr_jobs":
		result = e.asrReadTool(ctx, run, call)
	case "get_project_snapshot":
		result, executeErr = e.projectSnapshot(ctx, run, call)
	case "transcribe_asset":
		result = e.transcribeAsset(ctx, run, call)
	case "smart_split_asset":
		result = e.smartSplitAsset(ctx, run, call)
	case "extract_audio":
		result = e.extractAudio(ctx, run, call)
	case "clean_oral_pauses":
		result = e.cleanOralPauses(ctx, run, call)
	default:
		result.Message = "不允许调用工具：" + call.Name
	}
	if executeErr != nil {
		return ToolResult{}, executeErr
	}
	if err := e.agents.SaveAgentToolResult(ctx, run.Owner, run.ProjectID, run.ID, result); err != nil {
		return ToolResult{}, err
	}
	return result, nil
}

func toolResultData(value any) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

func (e *AgentToolExecutor) transcribeAsset(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		AssetID  string `json:"assetId"`
		ModelID  string `json:"modelId"`
		Language string `json:"language"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.AssetID) == "" || strings.TrimSpace(input.ModelID) == "" {
		result.Message = "字幕识别参数无效"
		return result
	}
	if e.composite.SubmitASR == nil {
		result.Message = "ASR 服务尚未配置"
		return result
	}
	job, err := e.composite.SubmitASR(ctx, run.Owner, ASRSubmit{ProjectID: run.ProjectID, TimelineID: run.TimelineID, AssetID: strings.TrimSpace(input.AssetID), ModelID: strings.TrimSpace(input.ModelID), Language: strings.TrimSpace(input.Language), ExpectedRevision: run.ExpectedRevision})
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Data = true, "字幕识别任务已提交", toolResultData(map[string]any{"job": job})
	return result
}

func (e *AgentToolExecutor) smartSplitAsset(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		AssetID       string  `json:"assetId"`
		Threshold     float64 `json:"threshold"`
		MinIntervalMS int64   `json:"minIntervalMs"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.AssetID) == "" {
		result.Message = "智能切镜参数无效"
		return result
	}
	if input.Threshold == 0 {
		input.Threshold = .4
	}
	if input.MinIntervalMS == 0 {
		input.MinIntervalMS = 2000
	}
	if input.Threshold <= 0 || input.Threshold > 1 || input.MinIntervalMS < 100 || input.MinIntervalMS > 60000 {
		result.Message = "智能切镜参数无效"
		return result
	}
	if e.composite.SmartSplit == nil {
		result.Message = "智能切镜服务尚未配置"
		return result
	}
	timeline, count, message, err := e.composite.SmartSplit(ctx, run.Owner, run.ProjectID, strings.TrimSpace(input.AssetID), input.Threshold, input.MinIntervalMS, run.ExpectedRevision)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Revision = true, message, timeline.Revision
	result.Data = toolResultData(map[string]any{"timeline": timeline, "splitCount": count})
	return result
}

func (e *AgentToolExecutor) extractAudio(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		AssetID string `json:"assetId"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.AssetID) == "" {
		result.Message = "音频分离参数无效"
		return result
	}
	if e.composite.ExtractAudio == nil {
		result.Message = "音频分离服务尚未配置"
		return result
	}
	asset, err := e.composite.ExtractAudio(ctx, run.Owner, run.ProjectID, strings.TrimSpace(input.AssetID))
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Data = true, "音频已分离并加入素材库", toolResultData(map[string]any{"asset": asset})
	return result
}

func (e *AgentToolExecutor) cleanOralPauses(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		ModelID    string `json:"modelId"`
		MaxPauseMS int64  `json:"maxPauseMs"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil {
		result.Message = "去口误停顿参数无效"
		return result
	}
	if strings.TrimSpace(input.ModelID) == "" {
		input.ModelID = run.ModelID
	}
	if input.MaxPauseMS == 0 {
		input.MaxPauseMS = 1000
	}
	if input.MaxPauseMS < 100 || input.MaxPauseMS > 10000 {
		result.Message = "去口误停顿参数无效"
		return result
	}
	if e.composite.OralClean == nil {
		result.Message = "去口误停顿服务尚未配置"
		return result
	}
	timeline, summary, message, err := e.composite.OralClean(ctx, run.Owner, run.ProjectID, strings.TrimSpace(input.ModelID), input.MaxPauseMS, run.ExpectedRevision)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Revision = true, message, timeline.Revision
	result.Data = toolResultData(map[string]any{"timeline": timeline, "summary": summary})
	return result
}

func (e *AgentToolExecutor) saveTimeline(ctx context.Context, run AgentRun, call ToolCall, timeline Timeline, success string) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	saved, err := e.timelines.SaveTimeline(ctx, run.Owner, run.ProjectID, timeline, run.ExpectedRevision, e.now().UTC())
	if errors.Is(err, ErrRevisionConflict) {
		result.Message = "时间线已更新，请刷新后重试"
		return result
	}
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Revision = true, success, saved.Revision
	return result
}

func (e *AgentToolExecutor) splitClip(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		TrackID  string `json:"trackId"`
		ClipID   string `json:"clipId"`
		OffsetMS int64  `json:"offsetMs"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.TrackID) == "" || strings.TrimSpace(input.ClipID) == "" {
		result.Message = "切分片段参数无效"
		return result
	}
	timeline, failed, ok, err := e.loadWritableTimeline(ctx, run)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if !ok {
		failed.CallID, failed.Name = call.ID, call.Name
		return failed
	}
	track, found := findTimelineTrack(&timeline, input.TrackID)
	if !found {
		result.Message = "轨道不存在"
		return result
	}
	if track.Locked {
		result.Message = "轨道已锁定，不能切分片段"
		return result
	}
	index, found := findTrackClip(track, input.ClipID)
	if !found {
		result.Message = "片段不存在"
		return result
	}
	original := track.Clips[index]
	if input.OffsetMS <= 0 || input.OffsetMS >= original.DurationMS {
		result.Message = "切分位置必须在片段内部"
		return result
	}
	rate := original.PlaybackRate
	if rate <= 0 {
		rate = 1
	}
	right := original
	right.ID = original.ID + "-split-" + call.ID
	right.StartMS = original.StartMS + input.OffsetMS
	right.DurationMS = original.DurationMS - input.OffsetMS
	right.SourceStartMS = original.SourceStartMS + int64(float64(input.OffsetMS)*rate)
	original.DurationMS = input.OffsetMS
	track.Clips[index] = original
	track.Clips = append(track.Clips[:index+1], append([]Clip{right}, track.Clips[index+1:]...)...)
	return e.saveTimeline(ctx, run, call, timeline, "片段已切分")
}

func (e *AgentToolExecutor) setPlaybackRate(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		TrackID      string  `json:"trackId"`
		ClipID       string  `json:"clipId"`
		PlaybackRate float64 `json:"playbackRate"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || input.PlaybackRate < .25 || input.PlaybackRate > 4 || strings.TrimSpace(input.TrackID) == "" || strings.TrimSpace(input.ClipID) == "" {
		result.Message = "变速参数无效，范围为 0.25x–4x"
		return result
	}
	timeline, failed, ok, err := e.loadWritableTimeline(ctx, run)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if !ok {
		failed.CallID, failed.Name = call.ID, call.Name
		return failed
	}
	track, found := findTimelineTrack(&timeline, input.TrackID)
	if !found {
		result.Message = "轨道不存在"
		return result
	}
	if track.Locked {
		result.Message = "轨道已锁定，不能修改变速"
		return result
	}
	index, found := findTrackClip(track, input.ClipID)
	if !found {
		result.Message = "片段不存在"
		return result
	}
	previous := track.Clips[index].PlaybackRate
	if previous <= 0 {
		previous = 1
	}
	track.Clips[index].PlaybackRate = input.PlaybackRate
	track.Clips[index].DurationMS = max64(1, int64(float64(track.Clips[index].DurationMS)*previous/input.PlaybackRate+.5))
	return e.saveTimeline(ctx, run, call, timeline, "片段变速已更新")
}

func (e *AgentToolExecutor) replaceSubtitleText(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || input.From == "" {
		result.Message = "字幕替换参数无效"
		return result
	}
	timeline, failed, ok, err := e.loadWritableTimeline(ctx, run)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if !ok {
		failed.CallID, failed.Name = call.ID, call.Name
		return failed
	}
	count := 0
	for trackIndex := range timeline.Tracks {
		track := &timeline.Tracks[trackIndex]
		if track.Kind != TrackSubtitle || track.Locked {
			continue
		}
		for clipIndex := range track.Clips {
			occurrences := strings.Count(track.Clips[clipIndex].Text, input.From)
			if occurrences > 0 {
				track.Clips[clipIndex].Text = strings.ReplaceAll(track.Clips[clipIndex].Text, input.From, input.To)
				count += occurrences
			}
		}
	}
	if count == 0 {
		result.Message = "没有找到可替换的字幕"
		return result
	}
	return e.saveTimeline(ctx, run, call, timeline, fmt.Sprintf("已替换 %d 处字幕", count))
}

func (e *AgentToolExecutor) placeAsset(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		AssetID string `json:"assetId"`
		TrackID string `json:"trackId"`
		StartMS int64  `json:"startMs"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.AssetID) == "" || strings.TrimSpace(input.TrackID) == "" || input.StartMS < 0 {
		result.Message = "放置素材参数无效"
		return result
	}
	asset, err := e.timelines.Asset(ctx, run.Owner, run.ProjectID, input.AssetID)
	if err != nil {
		result.Message = "素材不存在"
		return result
	}
	timeline, failed, ok, err := e.loadWritableTimeline(ctx, run)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if !ok {
		failed.CallID, failed.Name = call.ID, call.Name
		return failed
	}
	trackKind := TrackVideo
	clipKind := ClipKind(asset.Kind)
	if asset.Kind == "audio" {
		trackKind = TrackAudio
	}
	if asset.Kind == "subtitle" {
		trackKind = TrackSubtitle
	}
	track, found := findTimelineTrack(&timeline, input.TrackID)
	if !found {
		timeline.Tracks = append(timeline.Tracks, Track{ID: input.TrackID, Kind: trackKind, Name: input.TrackID, Order: len(timeline.Tracks), Clips: []Clip{}})
		track = &timeline.Tracks[len(timeline.Tracks)-1]
	}
	if track.Locked {
		result.Message = "轨道已锁定，不能放置素材"
		return result
	}
	if track.Kind != trackKind {
		result.Message = "素材类型与轨道不匹配"
		return result
	}
	duration := asset.DurationMS
	if duration <= 0 {
		if asset.Kind == "image" {
			duration = 5000
		} else {
			result.Message = "素材时长无效"
			return result
		}
	}
	track.Clips = append(track.Clips, Clip{ID: "agent-clip-" + call.ID, AssetID: asset.ID, Kind: clipKind, Name: asset.Name, StartMS: input.StartMS, DurationMS: duration, SourceDurationMS: duration, PlaybackRate: 1, Volume: 1})
	return e.saveTimeline(ctx, run, call, timeline, "素材已放入时间线")
}

func (e *AgentToolExecutor) undoRedo(ctx context.Context, run AgentRun, call ToolCall, redo bool) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct{}
	if err := decodeToolArguments(call.Arguments, &input); err != nil {
		result.Message = "撤销重做参数无效"
		return result
	}
	var timeline Timeline
	var err error
	if redo {
		timeline, _, err = e.timelines.RedoTimeline(ctx, run.Owner, run.ProjectID, run.ExpectedRevision, e.now().UTC())
	} else {
		timeline, _, err = e.timelines.UndoTimeline(ctx, run.Owner, run.ProjectID, run.ExpectedRevision, e.now().UTC())
	}
	if errors.Is(err, ErrRevisionConflict) {
		result.Message = "时间线已更新，请刷新后重试"
		return result
	}
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Revision = true, timeline.Revision
	if redo {
		result.Message = "时间线已重做"
	} else {
		result.Message = "时间线已撤销"
	}
	return result
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (e *AgentToolExecutor) loadWritableTimeline(ctx context.Context, run AgentRun) (Timeline, ToolResult, bool, error) {
	timeline, err := e.timelines.Timeline(ctx, run.Owner, run.ProjectID)
	if err != nil {
		return Timeline{}, ToolResult{}, false, err
	}
	if timeline.Revision != run.ExpectedRevision {
		return Timeline{}, ToolResult{OK: false, Message: "时间线已更新，请刷新后重试", Revision: timeline.Revision}, false, nil
	}
	return timeline, ToolResult{}, true, nil
}

func findTimelineTrack(timeline *Timeline, trackID string) (*Track, bool) {
	for index := range timeline.Tracks {
		if timeline.Tracks[index].ID == trackID {
			return &timeline.Tracks[index], true
		}
	}
	return nil, false
}

func findTrackClip(track *Track, clipID string) (int, bool) {
	for index := range track.Clips {
		if track.Clips[index].ID == clipID {
			return index, true
		}
	}
	return -1, false
}

func (e *AgentToolExecutor) moveClip(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		TrackID string `json:"trackId"`
		ClipID  string `json:"clipId"`
		StartMS int64  `json:"startMs"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.TrackID) == "" || strings.TrimSpace(input.ClipID) == "" || input.StartMS < 0 {
		result.Message = "移动片段参数无效"
		return result
	}
	timeline, failed, ok, err := e.loadWritableTimeline(ctx, run)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if !ok {
		failed.CallID, failed.Name = call.ID, call.Name
		return failed
	}
	track, found := findTimelineTrack(&timeline, input.TrackID)
	if !found {
		result.Message = "轨道不存在"
		return result
	}
	if track.Locked {
		result.Message = "轨道已锁定，不能移动片段"
		return result
	}
	index, found := findTrackClip(track, input.ClipID)
	if !found {
		result.Message = "片段不存在"
		return result
	}
	track.Clips[index].StartMS = input.StartMS
	saved, err := e.timelines.SaveTimeline(ctx, run.Owner, run.ProjectID, timeline, run.ExpectedRevision, e.now().UTC())
	if errors.Is(err, ErrRevisionConflict) {
		result.Message = "时间线已更新，请刷新后重试"
		return result
	}
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Revision = true, "片段已移动", saved.Revision
	return result
}

func (e *AgentToolExecutor) deleteClip(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	var input struct {
		TrackID string `json:"trackId"`
		ClipID  string `json:"clipId"`
	}
	if err := decodeToolArguments(call.Arguments, &input); err != nil || strings.TrimSpace(input.TrackID) == "" || strings.TrimSpace(input.ClipID) == "" {
		result.Message = "删除片段参数无效"
		return result
	}
	timeline, failed, ok, err := e.loadWritableTimeline(ctx, run)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if !ok {
		failed.CallID, failed.Name = call.ID, call.Name
		return failed
	}
	track, found := findTimelineTrack(&timeline, input.TrackID)
	if !found {
		result.Message = "轨道不存在"
		return result
	}
	if track.Locked {
		result.Message = "轨道已锁定，不能删除片段"
		return result
	}
	index, found := findTrackClip(track, input.ClipID)
	if !found {
		result.Message = "片段不存在"
		return result
	}
	track.Clips = append(track.Clips[:index], track.Clips[index+1:]...)
	saved, err := e.timelines.SaveTimeline(ctx, run.Owner, run.ProjectID, timeline, run.ExpectedRevision, e.now().UTC())
	if errors.Is(err, ErrRevisionConflict) {
		result.Message = "时间线已更新，请刷新后重试"
		return result
	}
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.OK, result.Message, result.Revision = true, "片段已删除", saved.Revision
	return result
}

func (e *AgentToolExecutor) projectSnapshot(ctx context.Context, run AgentRun, call ToolCall) (ToolResult, error) {
	if len(bytes.TrimSpace(call.Arguments)) > 0 && string(bytes.TrimSpace(call.Arguments)) != "{}" && string(bytes.TrimSpace(call.Arguments)) != "null" {
		return ToolResult{CallID: call.ID, Name: call.Name, Message: "读取项目参数无效"}, nil
	}
	timeline, err := e.timelines.Timeline(ctx, run.Owner, run.ProjectID)
	if err != nil {
		return ToolResult{}, err
	}
	assets, err := e.timelines.ListAssets(ctx, run.Owner, run.ProjectID)
	if err != nil {
		return ToolResult{}, err
	}
	data, err := json.Marshal(map[string]any{"timeline": timeline, "assets": assets, "currentTextModelId": run.ModelID, "oralCleanDefault": "current_text_model"})
	if err != nil {
		return ToolResult{}, fmt.Errorf("encode project snapshot: %w", err)
	}
	return ToolResult{CallID: call.ID, Name: call.Name, OK: true, Message: "项目快照已读取", Revision: timeline.Revision, Data: data}, nil
}

// AgentASRModel deliberately exposes only selection metadata, never provider secrets.
type AgentASRModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func AgentASRModelsFromCatalog(settings map[string]any) []AgentASRModel {
	out := []AgentASRModel{}
	var items []map[string]any
	switch raw := settings["modelCatalog"].(type) {
	case []map[string]any:
		items = raw
	case []any:
		for _, v := range raw {
			if m, ok := v.(map[string]any); ok {
				items = append(items, m)
			}
		}
	}
	text := func(v any) string {
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(v))
	}
	for _, m := range items {
		id := text(m["id"])
		// The same resolver validates enabled status and complete provider configuration;
		// constructing a provider performs no network requests.
		if id == "" || text(m["kind"]) != "asr" || text(m["credential"]) == "" || text(m["modelId"]) == "" {
			continue
		}
		if _, err := ASRProviderFromCatalog(settings, id, nil); err != nil {
			continue
		}
		name := text(m["name"])
		if name == "" {
			name = text(m["modelId"])
		}
		out = append(out, AgentASRModel{ID: id, Name: name})
	}
	return out
}
func (e *AgentToolExecutor) asrReadTool(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	r := ToolResult{CallID: call.ID, Name: call.Name}
	var args struct{}
	if err := decodeToolArguments(call.Arguments, &args); err != nil {
		r.Message = "查询参数无效"
		return r
	}
	if call.Name == "list_asr_models" {
		if e.composite.ListASRModels == nil {
			r.Message = "ASR 模型目录尚未配置"
			return r
		}
		models, err := e.composite.ListASRModels(ctx, run.Owner)
		if err != nil {
			r.Message = err.Error()
			return r
		}
		r.OK = true
		r.Message = "已读取可用 ASR 模型"
		r.Data = toolResultData(map[string]any{"models": models})
		return r
	}
	jobs, err := e.jobStore.ListASRJobs(ctx, run.Owner, run.ProjectID)
	if err != nil {
		r.Message = err.Error()
		return r
	}
	scoped := []ASRJob{}
	for _, job := range jobs {
		if canonicalASRTimeline(job.TimelineID) == canonicalASRTimeline(run.TimelineID) {
			scoped = append(scoped, job)
		}
	}
	r.OK = true
	r.Message = "已读取当前时间线字幕识别任务"
	r.Data = toolResultData(map[string]any{"jobs": scoped})
	return r
}
