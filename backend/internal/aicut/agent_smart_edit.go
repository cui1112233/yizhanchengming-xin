package aicut

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
)

func (e *AgentToolExecutor) smartEdit(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	timeline, err := e.timelines.Timeline(ctx, run.Owner, run.ProjectID)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if call.Name == "get_transcript" {
		var input struct{}
		if decodeToolArguments(call.Arguments, &input) != nil {
			result.Message = "读取字幕参数无效"
			return result
		}
		captions := []map[string]any{}
		for _, track := range timeline.Tracks {
			if track.Kind == TrackSubtitle {
				for _, clip := range track.Clips {
					captions = append(captions, map[string]any{"id": clip.ID, "trackId": track.ID, "text": clip.Text, "startMs": clip.StartMS, "endMs": clip.StartMS + clip.DurationMS, "locked": track.Locked})
				}
			}
		}
		result.OK = true
		result.Revision = timeline.Revision
		result.Message = fmt.Sprintf("已读取 %d 条字幕", len(captions))
		result.Data = toolResultData(map[string]any{"captions": captions, "script": timeline.Script})
		return result
	}
	if timeline.Revision != run.ExpectedRevision {
		result.Message = "时间线已更新，请重新读取项目"
		return result
	}
	next := cloneTimeline(timeline)
	switch call.Name {
	case "select_caption_ranges", "compact_caption_gaps":
		var input struct {
			ClipIDs    []string `json:"clipIds"`
			Mode       string   `json:"mode"`
			PaddingMS  int64    `json:"paddingMs"`
			MaxPauseMS int64    `json:"maxPauseMs"`
		}
		if decodeToolArguments(call.Arguments, &input) != nil || input.PaddingMS < 0 || input.PaddingMS > 2000 || len(input.ClipIDs) > 500 {
			result.Message = "字幕选片参数无效"
			return result
		}
		duration := int64(0)
		captions := map[string]Clip{}
		var speech []trimInterval
		for _, track := range timeline.Tracks {
			for _, clip := range track.Clips {
				duration = max64(duration, clip.StartMS+clip.DurationMS)
				if track.Kind == TrackSubtitle && clip.Kind == ClipSubtitle {
					captions[clip.ID] = clip
					speech = append(speech, trimInterval{clip.StartMS, clip.StartMS + clip.DurationMS})
				}
			}
		}
		var cuts []trimInterval
		if call.Name == "compact_caption_gaps" {
			if input.MaxPauseMS == 0 {
				input.MaxPauseMS = 1000
			}
			if input.MaxPauseMS < 100 || input.MaxPauseMS > 10000 {
				result.Message = "保留停顿范围为 100–10000 毫秒"
				return result
			}
			if len(speech) < 2 {
				result.Message = "请先识别或导入至少两条字幕"
				return result
			}
			speech = mergeTrimIntervals(speech)
			for i := 1; i < len(speech); i++ {
				gap := speech[i].start - speech[i-1].end
				if gap > input.MaxPauseMS {
					left := speech[i-1].end + input.MaxPauseMS/2
					cuts = append(cuts, trimInterval{left, speech[i].start - (input.MaxPauseMS - input.MaxPauseMS/2)})
				}
			}
		} else {
			if len(input.ClipIDs) == 0 || (input.Mode != "keep" && input.Mode != "delete") {
				result.Message = "请选择有效字幕，并指定 keep 或 delete"
				return result
			}
			chosen := []trimInterval{}
			seen := map[string]bool{}
			for _, id := range input.ClipIDs {
				clip, ok := captions[id]
				if !ok || seen[id] {
					result.Message = "字幕编号不存在或重复：" + id
					return result
				}
				seen[id] = true
				chosen = append(chosen, trimInterval{max64(0, clip.StartMS-input.PaddingMS), min64(duration, clip.StartMS+clip.DurationMS+input.PaddingMS)})
			}
			chosen = mergeTrimIntervals(chosen)
			if input.Mode == "delete" {
				cuts = chosen
			} else {
				cursor := int64(0)
				for _, keep := range chosen {
					if cursor < keep.start {
						cuts = append(cuts, trimInterval{cursor, keep.start})
					}
					cursor = keep.end
				}
				if cursor < duration {
					cuts = append(cuts, trimInterval{cursor, duration})
				}
			}
		}
		if len(cuts) == 0 {
			result.OK = true
			result.Message = "没有需要删除的区间"
			result.Revision = timeline.Revision
			return result
		}
		next, err = applyAgentRippleCuts(timeline, cuts)
		if err != nil {
			result.Message = err.Error()
			return result
		}
		removed := int64(0)
		for _, cut := range cuts {
			removed += cut.end - cut.start
		}
		result = e.saveTimeline(ctx, run, call, next, fmt.Sprintf("已剪辑 %d 个区间，缩短 %.2f 秒", len(cuts), float64(removed)/1000))
		result.Data = toolResultData(map[string]any{"removedMs": removed, "rangeCount": len(cuts)})
		return result
	case "set_clip_animation", "set_clip_color":
		var input struct {
			TrackID   string           `json:"trackId"`
			ClipID    string           `json:"clipId"`
			Animation *ClipAnimation   `json:"animation"`
			Color     *ColorAdjustment `json:"color"`
		}
		if decodeToolArguments(call.Arguments, &input) != nil {
			result.Message = "片段参数无效"
			return result
		}
		_, _, clip, clipErr := editPlanWritableClip(&next, EditOperation{ID: call.ID, TrackID: input.TrackID, ClipID: input.ClipID})
		if clipErr != nil {
			result.Message = clipErr.Error()
			return result
		}
		if call.Name == "set_clip_animation" {
			if input.Animation == nil || !validClipAnimation(input.Animation) {
				result.Message = "动画参数无效"
				return result
			}
			clip.Animation = input.Animation
		} else {
			if input.Color == nil || (clip.Kind != ClipVideo && clip.Kind != ClipImage) {
				result.Message = "请选择画面片段并提供调色参数"
				return result
			}
			if clip.VisualProcessing == nil {
				clip.VisualProcessing = &VisualProcessing{}
			}
			clip.VisualProcessing.Color = *input.Color
			if !validVisualProcessing(clip.VisualProcessing) {
				result.Message = "调色参数超出范围"
				return result
			}
		}
	case "apply_edit_operations":
		var input struct {
			Operations []EditOperation `json:"operations"`
		}
		if decodeToolArguments(call.Arguments, &input) != nil || len(input.Operations) == 0 || len(input.Operations) > 50 {
			result.Message = "操作数量应为 1–50 项"
			return result
		}
		assets, assetErr := e.timelines.ListAssets(ctx, run.Owner, run.ProjectID)
		if assetErr != nil {
			result.Message = assetErr.Error()
			return result
		}
		ids := []string{}
		for _, asset := range assets {
			if len(ids) < 64 {
				ids = append(ids, asset.ID)
			}
		}
		// Request validation requires a selection even for operations using only existing timeline clips.
		if len(ids) == 0 {
			ids = []string{"timeline-only"}
		}
		for i := range input.Operations {
			op := &input.Operations[i]
			if op.Type == EditSetProjectFormat {
				result.Message = "画面比例请在自动成片中设置"
				return result
			}
			if op.Type == EditAddBGM {
				found := false
				for _, asset := range assets {
					if asset.ID == op.AssetID && asset.Kind == "audio" {
						found = true
					}
				}
				if !found {
					result.Message = "背景音乐必须来自本工程音频素材"
					return result
				}
			}
			digest := sha256.Sum256([]byte(run.ID + "\x00" + call.ID))
			op.ID = fmt.Sprintf("agent-%x-%d", digest[:12], i)
			op.Enabled = true
			if strings.TrimSpace(op.Reason) == "" {
				op.Reason = "按用户要求调整"
			}
		}
		modelID := run.ModelID
		if modelID == "" {
			modelID = "configured-model"
		}
		selectedAssets := []EditPlanAsset{}
		for _, asset := range assets {
			for _, id := range ids {
				if asset.ID == id {
					selectedAssets = append(selectedAssets, EditPlanAsset{ID: asset.ID, Name: asset.Name, Kind: asset.Kind, DurationMS: asset.DurationMS})
					break
				}
			}
		}
		plan := EditPlan{ID: "agent-plan", ClientRequestID: run.ID, Owner: run.Owner, ProjectID: run.ProjectID, BaseRevision: run.ExpectedRevision, ModelID: modelID, State: EditPlanReady, Request: EditPlanRequest{AssetIDs: ids, SelectedAssets: selectedAssets, Goal: run.Instruction, AspectRatio: "16:9", IncludeCaptions: true, IncludeTransitions: true, IncludeFilters: true, IncludeBGM: true}, Operations: input.Operations}
		next, _, err = ApplyEditPlan(next, plan)
		if err != nil {
			result.Message = err.Error()
			return result
		}
	}
	return e.saveTimeline(ctx, run, call, next, "剪辑设置已保存，可撤销恢复")
}

func applyAgentRippleCuts(timeline Timeline, cuts []trimInterval) (Timeline, error) {
	cuts = mergeTrimIntervals(cuts)
	next := cloneTimeline(timeline)
	used := map[string]struct{}{}
	for _, track := range timeline.Tracks {
		for _, clip := range track.Clips {
			used[clip.ID] = struct{}{}
		}
	}
	for ti, track := range timeline.Tracks {
		next.Tracks[ti].Clips = nil
		for _, clip := range track.Clips {
			segments := subtractTrimIntervals(clip.StartMS, clip.StartMS+clip.DurationMS, cuts)
			unchanged := len(segments) == 1 && segments[0].start == clip.StartMS && segments[0].end == clip.StartMS+clip.DurationMS
			moved := mapTrimmedTime(clip.StartMS, cuts) != clip.StartMS
			if track.Locked && (!unchanged || moved) {
				return Timeline{}, fmt.Errorf("轨道 %s 已锁定，请先解锁再同步剪辑", track.Name)
			}
			if !unchanged && (len(clip.Keyframes) > 0 || len(clip.SpeedCurve) > 0 || clip.Reversed || clip.FreezeSourceMS != nil || clip.CompoundID != "" || clip.Kind == ClipTransition || clip.Transition != "" || clip.TransitionTrackID != "") {
				return Timeline{}, fmt.Errorf("片段 %s 含关键帧、曲线变速、复合片段或转场，请先处理后再自动选片", clip.Name)
			}
			for si, segment := range segments {
				part := clip
				part.StartMS = mapTrimmedTime(segment.start, cuts)
				part.DurationMS = segment.end - segment.start
				rate := clip.PlaybackRate
				if rate <= 0 {
					rate = 1
				}
				part.SourceStartMS = clip.SourceStartMS + int64(math.Round(float64(segment.start-clip.StartMS)*rate))
				if si > 0 {
					part.ID = uniqueTrimClipID(clip.ID+"-agent", used)
					used[part.ID] = struct{}{}
				}
				part.FadeInMS = min64(part.FadeInMS, part.DurationMS)
				part.FadeOutMS = min64(part.FadeOutMS, part.DurationMS)
				next.Tracks[ti].Clips = append(next.Tracks[ti].Clips, part)
			}
		}
	}
	for i := range next.Markers {
		next.Markers[i].TimeMS = mapTrimmedTime(next.Markers[i].TimeMS, cuts)
	}
	if err := validateTimeline(next); err != nil {
		return Timeline{}, err
	}
	return next, nil
}
