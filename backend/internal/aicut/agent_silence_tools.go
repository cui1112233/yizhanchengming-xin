package aicut

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
)

func (e *AgentToolExecutor) audioSilence(ctx context.Context, run AgentRun, call ToolCall) ToolResult {
	result := ToolResult{CallID: call.ID, Name: call.Name}
	fail := func(err error) ToolResult { result.Message = err.Error(); return result }
	var input struct {
		AssetID      string  `json:"assetId"`
		MinSilenceMS int64   `json:"minSilenceMs"`
		ThresholdDB  float64 `json:"thresholdDb"`
		PaddingMS    int64   `json:"paddingMs"`
	}
	if decodeToolArguments(call.Arguments, &input) != nil || strings.TrimSpace(input.AssetID) == "" {
		return fail(errors.New("静音分析参数无效"))
	}
	opts, err := NormalizeSilenceOptions(SilenceOptions{input.MinSilenceMS, input.ThresholdDB, input.PaddingMS})
	if err != nil {
		return fail(err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	timeline, err := e.timelines.Timeline(ctx, run.Owner, run.ProjectID)
	if err != nil {
		return fail(err)
	}
	removing := call.Name == "remove_audio_silence"
	if removing && timeline.Revision != run.ExpectedRevision {
		return fail(ErrRevisionConflict)
	}
	if e.composite.AnalyzeSilence == nil {
		return fail(errors.New("静音采样服务尚未配置"))
	}
	analysis, err := e.composite.AnalyzeSilence(ctx, run.Owner, run.ProjectID, strings.TrimSpace(input.AssetID), opts)
	if err != nil {
		return fail(err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if analysis.AssetID != strings.TrimSpace(input.AssetID) || analysis.DurationMS <= 0 || analysis.DurationMS > MaxSilenceDurationMS {
		return fail(errors.New("静音采样结果无效"))
	}
	for _, interval := range analysis.Intervals {
		if interval.StartMS < 0 || interval.EndMS <= interval.StartMS || interval.EndMS > analysis.DurationMS {
			return fail(errors.New("静音采样区间无效"))
		}
	}
	if !removing {
		result.OK = true
		result.Revision = timeline.Revision
		result.Message = fmt.Sprintf("已按真实音频采样检测出 %d 段静音", len(analysis.Intervals))
		result.Data = toolResultData(analysis)
		return result
	}
	if analysis.AllSilent {
		return fail(errors.New("整份素材均低于静音阈值，请调整阈值或手动确认保留区间"))
	}
	cuts, err := mapAudioSilenceCuts(timeline, analysis)
	if err != nil {
		return fail(err)
	}
	if len(cuts) == 0 {
		result.OK = true
		result.Revision = timeline.Revision
		result.Message = "没有需要删除的静音区间"
		result.Data = toolResultData(map[string]any{"analysis": analysis, "removedMs": 0, "rangeCount": 0})
		return result
	}
	if err := preflightSilenceCuts(timeline, cuts); err != nil {
		return fail(err)
	}
	next, err := applyAgentRippleCuts(timeline, cuts)
	if err != nil {
		return fail(err)
	}
	retained := false
	for _, track := range next.Tracks {
		if len(track.Clips) > 0 {
			retained = true
		}
	}
	if !retained {
		return fail(errors.New("静音剪辑将清空整个项目，已停止"))
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	removed := int64(0)
	for _, cut := range cuts {
		removed += cut.end - cut.start
	}
	result = e.saveTimeline(ctx, run, call, next, fmt.Sprintf("已按真实音频删除 %d 段静音，缩短 %.2f 秒", len(cuts), float64(removed)/1000))
	if result.OK {
		result.Data = toolResultData(map[string]any{"analysis": analysis, "removedMs": removed, "rangeCount": len(cuts)})
	}
	return result
}

func preflightSilenceCuts(timeline Timeline, cuts []trimInterval) error {
	for _, track := range timeline.Tracks {
		for _, clip := range track.Clips {
			segments := subtractTrimIntervals(clip.StartMS, clip.StartMS+clip.DurationMS, cuts)
			unchanged := len(segments) == 1 && segments[0].start == clip.StartMS && segments[0].end == clip.StartMS+clip.DurationMS
			if !unchanged && (clip.FadeInMS > 0 || clip.FadeOutMS > 0 || HasClipAnimation(clip.Animation)) {
				return fmt.Errorf("片段 %s 含淡入淡出或动画，请先处理后再自动静音剪辑", clip.Name)
			}
		}
	}
	return nil
}

func mapAudioSilenceCuts(timeline Timeline, analysis SilenceAnalysis) ([]trimInterval, error) {
	var cuts []trimInterval
	found := false
	for _, track := range timeline.Tracks {
		for _, clip := range track.Clips {
			if clip.AssetID != analysis.AssetID || (clip.Kind != ClipAudio && clip.Kind != ClipVideo) {
				continue
			}
			found = true
			if clip.Reversed || clip.FreezeSourceMS != nil || len(clip.SpeedCurve) > 0 || clip.CompoundID != "" || len(clip.CompoundPath) > 0 {
				return nil, fmt.Errorf("片段 %s 含倒放、定格、曲线变速或复合内容，无法映射静音区间", clip.Name)
			}
			rate := clip.PlaybackRate
			if rate == 0 {
				rate = 1
			}
			if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
				return nil, errors.New("片段速度无效")
			}
			sourceEnd := float64(clip.SourceStartMS) + float64(clip.DurationMS)*rate
			for _, interval := range analysis.Intervals {
				start := math.Max(float64(interval.StartMS), float64(clip.SourceStartMS))
				end := math.Min(float64(interval.EndMS), sourceEnd)
				if end <= start {
					continue
				}
				// Round inward: never remove audible source samples at either edge.
				left := clip.StartMS + int64(math.Ceil((start-float64(clip.SourceStartMS))/rate))
				right := clip.StartMS + int64(math.Floor((end-float64(clip.SourceStartMS))/rate))
				if right > left {
					cuts = append(cuts, trimInterval{left, right})
				}
			}
		}
	}
	if !found {
		return nil, errors.New("该素材尚未放入当前时间线")
	}
	return mergeTrimIntervals(cuts), nil
}
