package aicut

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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
		case EditPlaceAsset:
			applyErr = editPlanPlaceAsset(&next, normalized, operation, false)
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
			original := timeline.Script
			for _, track := range timeline.Tracks {
				for _, clip := range track.Clips {
					if clip.Kind == ClipSubtitle {
						original += "\n" + clip.Text
					}
				}
			}
			if !strings.Contains(original, operation.Text) {
				applyErr = editPlanApplyError(operation, "字幕缺少已有剧本或字幕原文依据，不能编造对白")
			} else {
				applyErr = editPlanAddSubtitle(&next, operation, false)
			}
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
				clip.CatalogParameters = make(map[string]float64, len(operation.Parameters))
				for key, value := range operation.Parameters {
					clip.CatalogParameters[key] = value
				}
			}
		case EditAddEffectLayer:
			track, ok := editPlanTrack(&next, operation.TrackID)
			if !ok || track.Locked || track.Kind != TrackEffect {
				applyErr = editPlanApplyError(operation, "特效轨道不可写")
			} else {
				track.Clips = append(track.Clips, Clip{ID: "plan-" + operation.ID, Kind: ClipEffect, Name: operation.ResourceID, StartMS: operation.StartMS, DurationMS: operation.DurationMS, SourceDurationMS: operation.DurationMS, ResourceID: operation.ResourceID, Parameters: operation.Parameters, PlaybackRate: 1, Volume: 1})
			}
		case EditAddBGM:
			applyErr = editPlanPlaceAsset(&next, normalized, operation, true)
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
	summary.EstimatedDurationMS = EditPlanTimelineDuration(next)
	return next, summary, nil
}

// EditPlanTimelineDuration measures the result, never the model's estimate.
func EditPlanTimelineDuration(t Timeline) int64 {
	var end int64
	for _, track := range t.Tracks {
		for _, clip := range track.Clips {
			if e := clip.StartMS + clip.DurationMS; e > end {
				end = e
			}
		}
	}
	return end
}
func editPlanPlaceAsset(t *Timeline, plan EditPlan, op EditOperation, audio bool) error {
	if !editPlanAssetSelected(plan, op.AssetID) {
		return editPlanApplyError(op, "素材不在已选列表中")
	}
	var asset *EditPlanAsset
	for i := range plan.Request.SelectedAssets {
		if plan.Request.SelectedAssets[i].ID == op.AssetID {
			asset = &plan.Request.SelectedAssets[i]
			break
		}
	}
	if asset == nil {
		return editPlanApplyError(op, "缺少经过核验的素材信息")
	}
	kind := ClipVideo
	trackKind := TrackVideo
	if audio {
		if asset.Kind != "audio" {
			return editPlanApplyError(op, "背景音乐必须是音频素材")
		}
		kind = ClipAudio
		trackKind = TrackAudio
	} else if asset.Kind == "image" {
		kind = ClipImage
	} else if asset.Kind != "video" {
		return editPlanApplyError(op, "画面只接受视频或图片")
	}
	if asset.Kind == "image" {
		if op.SourceStartMS != 0 {
			return editPlanApplyError(op, "图片没有源时间偏移")
		}
	} else if asset.DurationMS <= 0 || op.SourceStartMS > asset.DurationMS || op.DurationMS > asset.DurationMS-op.SourceStartMS {
		return editPlanApplyError(op, "放置范围超过实际素材时长")
	}
	track, ok := editPlanTrack(t, op.TrackID)
	if !ok {
		t.Tracks = append(t.Tracks, Track{ID: op.TrackID, Kind: trackKind, Name: "自动剪辑", Order: len(t.Tracks), Settings: TrackSettings{PreviewVisible: true}, Clips: []Clip{}})
		track = &t.Tracks[len(t.Tracks)-1]
	}
	if track.Locked || track.Kind != trackKind {
		return editPlanApplyError(op, "目标轨道类型不符或已锁定")
	}
	id := "plan-" + op.ID
	for _, c := range track.Clips {
		if c.ID == id || editPlanIntervalsOverlap(op.StartMS, op.DurationMS, c.StartMS, c.DurationMS) {
			return editPlanApplyError(op, "放置后与现有片段重叠或编号重复")
		}
	}
	duration := asset.DurationMS
	if kind == ClipImage {
		duration = op.DurationMS
	}
	volume := 1.0
	if audio {
		volume = op.Volume
	}
	track.Clips = append(track.Clips, Clip{ID: id, AssetID: asset.ID, Kind: kind, Name: asset.Name, StartMS: op.StartMS, DurationMS: op.DurationMS, SourceStartMS: op.SourceStartMS, SourceDurationMS: duration, PlaybackRate: 1, Volume: volume, FadeInMS: op.FadeInMS, FadeOutMS: op.FadeOutMS})
	return nil
}
