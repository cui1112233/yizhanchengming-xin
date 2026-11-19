package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// 分镜合并按洛水 2026-10-29 版规则实现：按分镜顺序把相邻分镜打包成组，
// 每组推荐时长之和不超过所选最长时长，每组只保留第一个分镜，其余并入后删除。
// 这里是纯本地整理，不调用任何模型接口。
const (
	StoryboardMergeDefaultMaxSeconds = 30.0
	StoryboardMergeOptionMinSeconds  = 10.0
	StoryboardMergeOptionMidSeconds  = 15.0
	StoryboardMergeOptionMaxSeconds  = 30.0
)

// StoryboardMergeGroup 是一组待合并的相邻分镜。KeepShotID 是该组保留的第一个分镜。
type StoryboardMergeGroup struct {
	KeepShotID string   `json:"keep_shot_id"`
	ShotIDs    []string `json:"shot_ids"`
	Seconds    float64  `json:"seconds"`
}

// StoryboardMergePlan 是合并预览：合并前后分镜数、待合并组数与分组明细。
type StoryboardMergePlan struct {
	MaxSeconds    float64                `json:"max_seconds"`
	BeforeCount   int                    `json:"before_count"`
	AfterCount    int                    `json:"after_count"`
	RemovedCount  int                    `json:"removed_count"`
	PendingGroups int                    `json:"pending_groups"`
	PendingShots  int                    `json:"pending_shots"`
	UntimedCount  int                    `json:"untimed_count"`
	Groups        []StoryboardMergeGroup `json:"groups"`
}

// PlanStoryboardMerge 只计算分组，不写库；弹窗预览与确认执行共用同一份算法。
func PlanStoryboardMerge(shots []StoryboardShot, maxSeconds float64) (StoryboardMergePlan, error) {
	plan := StoryboardMergePlan{MaxSeconds: maxSeconds, BeforeCount: len(shots), AfterCount: len(shots), Groups: []StoryboardMergeGroup{}}
	if math.IsNaN(maxSeconds) || math.IsInf(maxSeconds, 0) || maxSeconds <= 0 {
		return plan, fmt.Errorf("%w: 最长合并时长必须大于 0 秒", ErrInvalidArgument)
	}
	timed := 0
	for _, shot := range shots {
		if shot.Duration > 0 {
			timed++
		}
	}
	if len(shots) == 0 {
		return plan, fmt.Errorf("%w: 当前没有分镜可合并", ErrInvalidState)
	}
	plan.UntimedCount = len(shots) - timed
	if timed == 0 {
		return plan, fmt.Errorf("%w: 当前分镜没有推荐时长，无法按时长合并", ErrInvalidState)
	}

	group := make([]int, 0, len(shots))
	var sum float64
	flush := func() {
		defer func() { group = group[:0]; sum = 0 }()
		if len(group) < 2 {
			return
		}
		ids := make([]string, 0, len(group))
		for _, index := range group {
			ids = append(ids, shots[index].ID)
		}
		plan.Groups = append(plan.Groups, StoryboardMergeGroup{KeepShotID: shots[group[0]].ID, ShotIDs: ids, Seconds: roundStoryboardSeconds(sum)})
		plan.PendingShots += len(group)
		plan.AfterCount -= len(group) - 1
	}
	for index, shot := range shots {
		if shot.Duration <= 0 {
			// 没有推荐时长的分镜不能参与打包，同时作为分组断点，避免跨过它合并。
			flush()
			continue
		}
		if len(group) > 0 && sum+shot.Duration > maxSeconds {
			flush()
		}
		group = append(group, index)
		sum += shot.Duration
	}
	flush()
	plan.PendingGroups = len(plan.Groups)
	plan.RemovedCount = plan.BeforeCount - plan.AfterCount
	return plan, nil
}

// MergeStoryboardShotsByDuration 按最长时长把当前集分镜批量合并，返回合并后的分镜与预览明细。
func MergeStoryboardShotsByDuration(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, maxSeconds float64) (Storyboard, StoryboardMergePlan, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, StoryboardMergePlan{}, err
	}
	plan, err := PlanStoryboardMerge(storyboard.Shots, maxSeconds)
	if err != nil {
		return storyboard, plan, err
	}
	if len(plan.Groups) == 0 {
		return storyboard, plan, fmt.Errorf("%w: 按最长 %s 秒没有可合并的分镜", ErrInvalidState, trimStoryboardSeconds(maxSeconds))
	}

	groupByKeep := make(map[string]StoryboardMergeGroup, len(plan.Groups))
	dropped := make(map[string]string, plan.PendingShots-len(plan.Groups))
	for _, group := range plan.Groups {
		groupByKeep[group.KeepShotID] = group
		for _, id := range group.ShotIDs[1:] {
			dropped[id] = group.KeepShotID
		}
	}

	merged := make([]StoryboardShot, 0, plan.AfterCount)
	for _, shot := range storyboard.Shots {
		if _, remove := dropped[shot.ID]; remove {
			continue
		}
		if group, ok := groupByKeep[shot.ID]; ok {
			merged = append(merged, mergeStoryboardShotGroup(group, storyboard.Shots))
			continue
		}
		merged = append(merged, shot)
	}
	storyboard.Shots = merged
	saved, err := saveEditedStoryboard(ctx, store, storyboard)
	if err != nil {
		return saved, plan, err
	}
	if len(dropped) > 0 {
		reattachMergedStoryboardMedia(ctx, store, owner, projectID, dropped)
	}
	return saved, plan, nil
}

// mergeStoryboardShotGroup 把一组分镜并入保留的第一条：
// 时长求和；视频提示词、画面提示词、台词按顺序拼接；角色、场景、道具取并集；
// 原始协议里记录成员，便于回查被并入的内容。
func mergeStoryboardShotGroup(group StoryboardMergeGroup, shots []StoryboardShot) StoryboardShot {
	members := make([]StoryboardShot, 0, len(group.ShotIDs))
	for _, id := range group.ShotIDs {
		if index := storyboardShotIndex(shots, id); index >= 0 {
			members = append(members, shots[index])
		}
	}
	if len(members) == 0 {
		return StoryboardShot{}
	}
	keep := members[0]
	var total float64
	var videoDesc, imagePrompt, line, voiceActor, emotion []string
	var characters, scene, props []string
	var sourceElements []json.RawMessage
	for _, member := range members {
		total += member.Duration
		videoDesc = append(videoDesc, member.VideoDesc)
		imagePrompt = append(imagePrompt, member.ImagePrompt)
		line = append(line, member.Line)
		voiceActor = append(voiceActor, member.VoiceActor)
		emotion = append(emotion, member.Emotion)
		characters = append(characters, member.Characters)
		scene = append(scene, member.Scene)
		props = append(props, member.Props)
		if json.Valid(member.SourceElements) && len(member.SourceElements) > 0 {
			sourceElements = append(sourceElements, json.RawMessage(member.SourceElements))
		}
	}
	keep.Duration = roundStoryboardSeconds(total)
	keep.VideoDesc = joinStoryboardLines(videoDesc, "\n")
	keep.ImagePrompt = joinStoryboardLines(imagePrompt, "\n")
	keep.Line = joinStoryboardLines(line, "\n")
	keep.VoiceActor = joinStoryboardNames(voiceActor)
	keep.Emotion = joinStoryboardNames(emotion)
	keep.Characters = joinStoryboardNames(characters)
	keep.Scene = joinStoryboardNames(scene)
	keep.Props = joinStoryboardNames(props)
	if len(sourceElements) > 0 {
		if len(sourceElements) == 1 {
			keep.SourceElements = sourceElements[0]
		} else if raw, err := json.Marshal(sourceElements); err == nil {
			keep.SourceElements = raw
		}
	}
	fields := map[string]any{}
	if len(keep.ProtocolFields) > 0 {
		_ = json.Unmarshal(keep.ProtocolFields, &fields)
	}
	if fields == nil {
		fields = map[string]any{}
	}
	fields["merged_shots"] = members
	fields["merged_count"] = len(members)
	if raw, err := json.Marshal(fields); err == nil {
		keep.ProtocolFields = raw
	}
	return keep
}

// reattachMergedStoryboardMedia 把被并入分镜的照片、视频挂到保留分镜上，避免素材变孤儿。
func reattachMergedStoryboardMedia(ctx context.Context, store Store, owner, projectID string, dropped map[string]string) {
	mediaStore, ok := store.(MediaStore)
	if !ok || mediaStore == nil {
		return
	}
	_, _ = mediaStore.MutateMedia(ctx, owner, projectID, func(state *MediaState) error {
		if state == nil {
			return nil
		}
		for index := range state.Items {
			if keep, hit := dropped[state.Items[index].ShotID]; hit {
				state.Items[index].ShotID = keep
			}
		}
		if state.Selected != nil {
			for shotID, mediaID := range state.Selected {
				keep, hit := dropped[shotID]
				if !hit {
					continue
				}
				if _, exists := state.Selected[keep]; !exists {
					state.Selected[keep] = mediaID
				}
				delete(state.Selected, shotID)
			}
		}
		return nil
	})
}

// joinStoryboardLines 按顺序拼接多段文本，空段自动跳过，相同内容不重复。
func joinStoryboardLines(values []string, separator string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if len(parts) > 0 && parts[len(parts)-1] == value {
			continue
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, separator)
}

// joinStoryboardNames 把逗号、分号、顿号分隔的名单求并集，保持出现顺序。
func joinStoryboardNames(values []string) string {
	seen := map[string]bool{}
	names := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.FieldsFunc(value, storyboardNameSeparator) {
			part = strings.TrimSpace(part)
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			names = append(names, part)
		}
	}
	return strings.Join(names, "；")
}

func storyboardNameSeparator(r rune) bool {
	switch r {
	case ',', '，', '、', ';', '；', '/', '|':
		return true
	}
	return false
}

func roundStoryboardSeconds(value float64) float64 {
	return math.Round(value*10) / 10
}

func trimStoryboardSeconds(value float64) string {
	return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%.1f", value), ".0"), ".")
}
