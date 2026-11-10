package aicut

import (
	"errors"
	"fmt"
	"sort"
)

var ErrEditPlanApply = errors.New("ai-cut edit plan apply failed")

type EditPlanApplySummary struct {
	Applied             int    `json:"applied"`
	Skipped             int    `json:"skipped"`
	RemovedDurationMS   int64  `json:"removedDurationMs"`
	EstimatedDurationMS int64  `json:"estimatedDurationMs"`
	AspectRatio         string `json:"aspectRatio,omitempty"`
}

func editPlanApplyError(operation EditOperation, message string) error {
	return fmt.Errorf("%w: %s: %s", ErrEditPlanApply, operation.ID, message)
}

func editPlanTrack(timeline *Timeline, trackID string) (*Track, bool) {
	for index := range timeline.Tracks {
		if timeline.Tracks[index].ID == trackID {
			return &timeline.Tracks[index], true
		}
	}
	return nil, false
}

func editPlanClip(track *Track, clipID string) (int, *Clip, bool) {
	for index := range track.Clips {
		if track.Clips[index].ID == clipID {
			return index, &track.Clips[index], true
		}
	}
	return -1, nil, false
}

func editPlanWritableClip(timeline *Timeline, operation EditOperation) (*Track, int, *Clip, error) {
	track, ok := editPlanTrack(timeline, operation.TrackID)
	if !ok {
		return nil, -1, nil, editPlanApplyError(operation, "轨道不存在")
	}
	if track.Locked {
		return nil, -1, nil, editPlanApplyError(operation, "轨道已锁定")
	}
	index, clip, ok := editPlanClip(track, operation.ClipID)
	if !ok {
		return nil, -1, nil, editPlanApplyError(operation, "片段不存在")
	}
	return track, index, clip, nil
}

func editPlanIntervalsOverlap(leftStart, leftDuration, rightStart, rightDuration int64) bool {
	return leftStart < rightStart+rightDuration && rightStart < leftStart+leftDuration
}

func editPlanRippleAfter(track *Track, threshold, removed int64, exceptID string) {
	if removed <= 0 {
		return
	}
	for index := range track.Clips {
		clip := &track.Clips[index]
		if clip.ID != exceptID && clip.StartMS >= threshold {
			clip.StartMS -= removed
		}
	}
}

func editPlanApplyTrim(timeline *Timeline, operation EditOperation) (int64, error) {
	track, _, clip, err := editPlanWritableClip(timeline, operation)
	if err != nil {
		return 0, err
	}
	if operation.EndMS > clip.DurationMS {
		return 0, editPlanApplyError(operation, "裁剪范围超过片段时长")
	}
	newDuration := operation.EndMS - operation.StartMS
	rate := clip.PlaybackRate
	if rate <= 0 {
		rate = 1
	}
	sourceOffset := int64(float64(operation.StartMS) * rate)
	sourceDuration := int64(float64(newDuration) * rate)
	if clip.SourceDurationMS > 0 && clip.SourceStartMS+sourceOffset+sourceDuration > clip.SourceDurationMS {
		return 0, editPlanApplyError(operation, "裁剪范围超过素材源时长")
	}
	oldEnd := clip.StartMS + clip.DurationMS
	removed := clip.DurationMS - newDuration
	clip.SourceStartMS += sourceOffset
	clip.DurationMS = newDuration
	editPlanRippleAfter(track, oldEnd, removed, clip.ID)
	return removed, nil
}

func editPlanApplyMove(timeline *Timeline, operation EditOperation) error {
	track, _, clip, err := editPlanWritableClip(timeline, operation)
	if err != nil {
		return err
	}
	for _, other := range track.Clips {
		if other.ID != clip.ID && editPlanIntervalsOverlap(operation.StartMS, clip.DurationMS, other.StartMS, other.DurationMS) {
			return editPlanApplyError(operation, "移动后与其他片段重叠")
		}
	}
	clip.StartMS = operation.StartMS
	return nil
}

func editPlanApplyTransition(timeline *Timeline, operation EditOperation) error {
	track, index, clip, err := editPlanWritableClip(timeline, operation)
	if err != nil {
		return err
	}
	if track.Kind != TrackVideo || (clip.Kind != ClipVideo && clip.Kind != ClipImage) {
		return editPlanApplyError(operation, "转场只能用于画面轨")
	}
	sorted := append([]Clip(nil), track.Clips...)
	sort.SliceStable(sorted, func(left, right int) bool { return sorted[left].StartMS < sorted[right].StartMS })
	position := -1
	for sortedIndex := range sorted {
		if sorted[sortedIndex].ID == track.Clips[index].ID {
			position = sortedIndex
			break
		}
	}
	if position <= 0 || sorted[position-1].StartMS+sorted[position-1].DurationMS != sorted[position].StartMS {
		return editPlanApplyError(operation, "转场前后片段不相邻")
	}
	if operation.DurationMS > sorted[position-1].DurationMS || operation.DurationMS > clip.DurationMS {
		return editPlanApplyError(operation, "转场时长超过相邻片段")
	}
	clip.Transition, clip.TransitionMS = operation.ResourceID, operation.DurationMS
	return nil
}

func editPlanAddSubtitle(timeline *Timeline, operation EditOperation, title bool) error {
	track, ok := editPlanTrack(timeline, operation.TrackID)
	if !ok {
		return editPlanApplyError(operation, "文字轨道不存在")
	}
	if track.Locked || track.Kind != TrackSubtitle {
		return editPlanApplyError(operation, "文字轨道不可写")
	}
	clip := Clip{ID: "plan-" + operation.ID, Kind: ClipSubtitle, Name: operation.Text, StartMS: operation.StartMS, DurationMS: operation.DurationMS, SourceDurationMS: operation.DurationMS, Text: operation.Text, Volume: 1, PlaybackRate: 1}
	if title {
		clip.TextResourceID = operation.ResourceID
	}
	if _, _, exists := editPlanClip(track, clip.ID); exists {
		return editPlanApplyError(operation, "生成的片段编号重复")
	}
	track.Clips = append(track.Clips, clip)
	return nil
}

func editPlanApplyRemoveRange(timeline *Timeline, operation EditOperation) (int64, error) {
	removed := operation.EndMS - operation.StartMS
	for trackIndex := range timeline.Tracks {
		track := &timeline.Tracks[trackIndex]
		if track.Locked {
			continue
		}
		next := make([]Clip, 0, len(track.Clips)+1)
		for _, clip := range track.Clips {
			clipEnd := clip.StartMS + clip.DurationMS
			switch {
			case clipEnd <= operation.StartMS:
				next = append(next, clip)
			case clip.StartMS >= operation.EndMS:
				clip.StartMS -= removed
				next = append(next, clip)
			case clip.StartMS >= operation.StartMS && clipEnd <= operation.EndMS:
				continue
			case clip.StartMS < operation.StartMS && clipEnd > operation.EndMS:
				headDuration := operation.StartMS - clip.StartMS
				tailDuration := clipEnd - operation.EndMS
				tail := clip
				tail.ID = "plan-" + operation.ID + "-tail-" + clip.ID
				tail.StartMS = operation.StartMS
				rate := clip.PlaybackRate
				if rate <= 0 {
					rate = 1
				}
				tail.SourceStartMS += int64(float64(operation.EndMS-clip.StartMS) * rate)
				tail.DurationMS = tailDuration
				clip.DurationMS = headDuration
				next = append(next, clip, tail)
			case clip.StartMS < operation.StartMS:
				clip.DurationMS = operation.StartMS - clip.StartMS
				next = append(next, clip)
			default:
				cut := operation.EndMS - clip.StartMS
				rate := clip.PlaybackRate
				if rate <= 0 {
					rate = 1
				}
				clip.SourceStartMS += int64(float64(cut) * rate)
				clip.DurationMS = clipEnd - operation.EndMS
				clip.StartMS = operation.StartMS
				next = append(next, clip)
			}
		}
		track.Clips = next
	}
	return removed, nil
}

func editPlanAssetSelected(plan EditPlan, assetID string) bool {
	for _, id := range plan.Request.AssetIDs {
		if id == assetID {
			return true
		}
	}
	return false
}

func ApplyEditPlan(timeline Timeline, plan EditPlan) (Timeline, EditPlanApplySummary, error) {
	if err := validateTimeline(timeline); err != nil {
		return Timeline{}, EditPlanApplySummary{}, fmt.Errorf("%w: 当前时间线无效", ErrEditPlanApply)
	}
	normalized, err := NormalizeEditPlan(plan)
	if err != nil {
		return Timeline{}, EditPlanApplySummary{}, fmt.Errorf("%w: %v", ErrEditPlanApply, err)
	}
	if normalized.BaseRevision != timeline.Revision {
		return Timeline{}, EditPlanApplySummary{}, fmt.Errorf("%w: %v", ErrEditPlanApply, ErrRevisionConflict)
	}
	next := cloneTimeline(timeline)
	summary := EditPlanApplySummary{EstimatedDurationMS: normalized.Summary.EstimatedDurationMS}
	for _, operation := range normalized.Operations {
		if !operation.Enabled {
			summary.Skipped++
			continue
		}
		var applyErr error
		switch operation.Type {
		case EditTrimClip:
			var removed int64
			removed, applyErr = editPlanApplyTrim(&next, operation)
			summary.RemovedDurationMS += removed
		case EditRemoveRange:
			var removed int64
			removed, applyErr = editPlanApplyRemoveRange(&next, operation)
			summary.RemovedDurationMS += removed
		case EditMoveClip:
			applyErr = editPlanApplyMove(&next, operation)
		case EditAddCaption:
			applyErr = editPlanAddSubtitle(&next, operation, false)
		case EditAddTitle:
			applyErr = editPlanAddSubtitle(&next, operation, true)
		case EditSetTransition:
			applyErr = editPlanApplyTransition(&next, operation)
		case EditSetFilter:
			_, _, clip, clipErr := editPlanWritableClip(&next, operation)
			if clipErr != nil {
				applyErr = clipErr
			} else if clip.Kind != ClipVideo && clip.Kind != ClipImage {
				applyErr = editPlanApplyError(operation, "滤镜只能用于画面片段")
			} else {
				clip.CatalogResourceID = operation.ResourceID
			}
		case EditAddEffectLayer:
			track, ok := editPlanTrack(&next, operation.TrackID)
			if !ok || track.Locked || track.Kind != TrackEffect {
				applyErr = editPlanApplyError(operation, "特效轨道不可写")
			} else {
				track.Clips = append(track.Clips, Clip{ID: "plan-" + operation.ID, Kind: ClipEffect, Name: operation.ResourceID, StartMS: operation.StartMS, DurationMS: operation.DurationMS, SourceDurationMS: operation.DurationMS, ResourceID: operation.ResourceID, Parameters: operation.Parameters, PlaybackRate: 1, Volume: 1})
			}
		case EditAddBGM:
			track, ok := editPlanTrack(&next, operation.TrackID)
			if !ok || track.Locked || track.Kind != TrackAudio {
				applyErr = editPlanApplyError(operation, "音频轨道不可写")
			} else if !editPlanAssetSelected(normalized, operation.AssetID) {
				applyErr = editPlanApplyError(operation, "背景音乐不在已选素材中")
			} else {
				track.Clips = append(track.Clips, Clip{ID: "plan-" + operation.ID, AssetID: operation.AssetID, Kind: ClipAudio, Name: "背景音乐", StartMS: operation.StartMS, DurationMS: operation.DurationMS, SourceDurationMS: operation.DurationMS, PlaybackRate: 1, Volume: operation.Volume, FadeInMS: operation.FadeInMS, FadeOutMS: operation.FadeOutMS})
			}
		case EditSetVolumeEnvelope:
			_, _, clip, clipErr := editPlanWritableClip(&next, operation)
			if clipErr != nil {
				applyErr = clipErr
			} else if clip.Kind != ClipAudio && clip.Kind != ClipVideo {
				applyErr = editPlanApplyError(operation, "该片段没有可调音频")
			} else {
				clip.Volume, clip.FadeInMS, clip.FadeOutMS = operation.Volume, operation.FadeInMS, operation.FadeOutMS
			}
		case EditSetProjectFormat:
			summary.AspectRatio = operation.ResourceID
		default:
			applyErr = editPlanApplyError(operation, "不支持的操作")
		}
		if applyErr != nil {
			return Timeline{}, EditPlanApplySummary{}, applyErr
		}
		summary.Applied++
	}
	if err := validateTimeline(next); err != nil {
		return Timeline{}, EditPlanApplySummary{}, fmt.Errorf("%w: 应用后的时间线无效", ErrEditPlanApply)
	}
	return next, summary, nil
}
