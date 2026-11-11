package aicut

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrInvalidFillerTrim = errors.New("invalid ai-cut filler trim")

type FillerCandidate struct {
	ClipID string `json:"clipId"`
	Reason string `json:"reason,omitempty"`
}

type FillerTrimSummary struct {
	FillerCount int   `json:"fillerCount"`
	PauseCount  int   `json:"pauseCount"`
	RemovedMS   int64 `json:"removedMs"`
}

type trimInterval struct{ start, end int64 }

func ParseFillerAnalysis(raw string) ([]FillerCandidate, error) {
	cleaned := strings.TrimSpace(raw)
	if strings.HasPrefix(cleaned, "```") {
		if newline := strings.IndexByte(cleaned, '\n'); newline >= 0 {
			cleaned = cleaned[newline+1:]
		}
		cleaned = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(cleaned), "```"))
	}
	var envelope struct {
		Fillers []FillerCandidate `json:"fillers"`
	}
	if cleaned == "" || json.Unmarshal([]byte(cleaned), &envelope) != nil || envelope.Fillers == nil {
		return nil, fmt.Errorf("%w: 模型没有返回有效的水词分析", ErrInvalidFillerTrim)
	}
	seen := make(map[string]struct{}, len(envelope.Fillers))
	for index := range envelope.Fillers {
		envelope.Fillers[index].ClipID = strings.TrimSpace(envelope.Fillers[index].ClipID)
		envelope.Fillers[index].Reason = strings.TrimSpace(envelope.Fillers[index].Reason)
		if envelope.Fillers[index].ClipID == "" {
			return nil, fmt.Errorf("%w: 第 %d 条结果缺少字幕编号", ErrInvalidFillerTrim, index+1)
		}
		if _, ok := seen[envelope.Fillers[index].ClipID]; ok {
			return nil, fmt.Errorf("%w: 字幕 %s 被重复标记", ErrInvalidFillerTrim, envelope.Fillers[index].ClipID)
		}
		seen[envelope.Fillers[index].ClipID] = struct{}{}
	}
	return envelope.Fillers, nil
}

func ApplyFillerTrim(timeline Timeline, candidates []FillerCandidate) (Timeline, FillerTrimSummary, error) {
	known := make(map[string]Clip)
	locked := make(map[string]struct{})
	for _, track := range timeline.Tracks {
		if track.Kind != TrackSubtitle {
			continue
		}
		for _, clip := range track.Clips {
			if clip.Kind != ClipSubtitle {
				continue
			}
			known[clip.ID] = clip
			if track.Locked {
				locked[clip.ID] = struct{}{}
			}
		}
	}
	remove := make(map[string]struct{}, len(candidates))
	removed := int64(0)
	for _, candidate := range candidates {
		id := strings.TrimSpace(candidate.ClipID)
		if _, duplicate := remove[id]; duplicate {
			return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 字幕 %s 被重复标记", ErrInvalidFillerTrim, id)
		}
		clip, ok := known[id]
		if !ok {
			return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 字幕 %s 不存在", ErrInvalidFillerTrim, id)
		}
		if _, ok := locked[id]; ok {
			return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 字幕 %s 所在轨道已锁定", ErrInvalidFillerTrim, id)
		}
		remove[id] = struct{}{}
		removed += clip.DurationMS
	}
	if len(remove) == 0 {
		return timeline, FillerTrimSummary{}, nil
	}
	tracks := make([]Track, len(timeline.Tracks))
	for trackIndex, track := range timeline.Tracks {
		tracks[trackIndex] = track
		if track.Kind != TrackSubtitle {
			tracks[trackIndex].Clips = append([]Clip(nil), track.Clips...)
			continue
		}
		tracks[trackIndex].Clips = make([]Clip, 0, len(track.Clips))
		for _, clip := range track.Clips {
			if _, deleted := remove[clip.ID]; !deleted {
				tracks[trackIndex].Clips = append(tracks[trackIndex].Clips, clip)
			}
		}
	}
	result := Timeline{Revision: timeline.Revision, Tracks: tracks}
	if err := validateTimeline(result); err != nil {
		return Timeline{}, FillerTrimSummary{}, err
	}
	return result, FillerTrimSummary{FillerCount: len(candidates), RemovedMS: removed}, nil
}

// ApplyOralClean is the timeline-changing Agent tool. Unlike the direct
// subtitle analyzer above, it removes the marked spoken intervals and shortens
// silent gaps across every unlocked track while preserving clip order.
func ApplyOralClean(timeline Timeline, candidates []FillerCandidate, maxPauseMS int64) (Timeline, FillerTrimSummary, error) {
	if maxPauseMS < 100 || maxPauseMS > 10000 {
		return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 停顿上限必须在 0.1 到 10 秒之间", ErrInvalidFillerTrim)
	}
	known := make(map[string]Clip)
	locked := make(map[string]struct{})
	captions := make([]Clip, 0)
	for _, track := range timeline.Tracks {
		if track.Kind != TrackSubtitle {
			continue
		}
		for _, clip := range track.Clips {
			if clip.Kind != ClipSubtitle {
				continue
			}
			known[clip.ID] = clip
			if track.Locked {
				locked[clip.ID] = struct{}{}
			} else {
				captions = append(captions, clip)
			}
		}
	}
	intervals := make([]trimInterval, 0, len(candidates)+len(captions))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		id := strings.TrimSpace(candidate.ClipID)
		if _, duplicate := seen[id]; duplicate {
			return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 字幕 %s 被重复标记", ErrInvalidFillerTrim, id)
		}
		seen[id] = struct{}{}
		clip, ok := known[id]
		if !ok {
			return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 字幕 %s 不存在", ErrInvalidFillerTrim, id)
		}
		if _, ok := locked[id]; ok {
			return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 字幕 %s 所在轨道已锁定", ErrInvalidFillerTrim, id)
		}
		intervals = append(intervals, trimInterval{start: clip.StartMS, end: clip.StartMS + clip.DurationMS})
	}
	sort.Slice(captions, func(i, j int) bool { return captions[i].StartMS < captions[j].StartMS })
	pauseCount := 0
	for index := 1; index < len(captions); index++ {
		previousEnd := captions[index-1].StartMS + captions[index-1].DurationMS
		if gap := captions[index].StartMS - previousEnd; gap > maxPauseMS {
			intervals = append(intervals, trimInterval{start: captions[index].StartMS - (gap - maxPauseMS), end: captions[index].StartMS})
			pauseCount++
		}
	}
	intervals = mergeTrimIntervals(intervals)
	if len(intervals) == 0 {
		return timeline, FillerTrimSummary{}, nil
	}
	for _, track := range timeline.Tracks {
		if !track.Locked {
			continue
		}
		for _, clip := range track.Clips {
			for _, cut := range intervals {
				if clip.StartMS < cut.end && clip.StartMS+clip.DurationMS > cut.start {
					return Timeline{}, FillerTrimSummary{}, fmt.Errorf("%w: 轨道 %s 已锁定，不能裁剪", ErrInvalidFillerTrim, track.Name)
				}
			}
		}
	}
	removed := int64(0)
	for _, cut := range intervals {
		removed += cut.end - cut.start
	}
	usedIDs := make(map[string]struct{})
	for _, track := range timeline.Tracks {
		for _, clip := range track.Clips {
			usedIDs[clip.ID] = struct{}{}
		}
	}
	tracks := make([]Track, len(timeline.Tracks))
	for trackIndex, track := range timeline.Tracks {
		tracks[trackIndex] = track
		tracks[trackIndex].Clips = nil
		for _, clip := range track.Clips {
			segments := subtractTrimIntervals(clip.StartMS, clip.StartMS+clip.DurationMS, intervals)
			for segmentIndex, segment := range segments {
				part := clip
				part.StartMS = mapTrimmedTime(segment.start, intervals)
				part.DurationMS = segment.end - segment.start
				part.SourceStartMS = clip.SourceStartMS + (segment.start - clip.StartMS)
				if segmentIndex > 0 {
					part.ID = uniqueTrimClipID(fmt.Sprintf("%s-trim-%d", clip.ID, segmentIndex), usedIDs)
					usedIDs[part.ID] = struct{}{}
				}
				tracks[trackIndex].Clips = append(tracks[trackIndex].Clips, part)
			}
		}
	}
	result := Timeline{Revision: timeline.Revision, Tracks: tracks}
	if err := validateTimeline(result); err != nil {
		return Timeline{}, FillerTrimSummary{}, err
	}
	return result, FillerTrimSummary{FillerCount: len(candidates), PauseCount: pauseCount, RemovedMS: removed}, nil
}

func mergeTrimIntervals(input []trimInterval) []trimInterval {
	filtered := input[:0]
	for _, item := range input {
		if item.end > item.start {
			filtered = append(filtered, item)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].start == filtered[j].start {
			return filtered[i].end < filtered[j].end
		}
		return filtered[i].start < filtered[j].start
	})
	result := make([]trimInterval, 0, len(filtered))
	for _, item := range filtered {
		if len(result) == 0 || item.start > result[len(result)-1].end {
			result = append(result, item)
			continue
		}
		if item.end > result[len(result)-1].end {
			result[len(result)-1].end = item.end
		}
	}
	return result
}

func subtractTrimIntervals(start, end int64, cuts []trimInterval) []trimInterval {
	cursor := start
	result := make([]trimInterval, 0, len(cuts)+1)
	for _, cut := range cuts {
		if cut.end <= cursor || cut.start >= end {
			continue
		}
		if cut.start > cursor {
			result = append(result, trimInterval{cursor, min64(cut.start, end)})
		}
		if cut.end > cursor {
			cursor = cut.end
		}
		if cursor >= end {
			break
		}
	}
	if cursor < end {
		result = append(result, trimInterval{cursor, end})
	}
	return result
}

func mapTrimmedTime(value int64, cuts []trimInterval) int64 {
	removed := int64(0)
	for _, cut := range cuts {
		if value >= cut.end {
			removed += cut.end - cut.start
			continue
		}
		if value > cut.start {
			removed += value - cut.start
		}
		break
	}
	return value - removed
}

func uniqueTrimClipID(base string, used map[string]struct{}) string {
	if _, ok := used[base]; !ok {
		return base
	}
	for index := 2; ; index++ {
		candidate := fmt.Sprintf("%s-%d", base, index)
		if _, ok := used[candidate]; !ok {
			return candidate
		}
	}
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
