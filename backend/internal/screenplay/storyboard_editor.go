package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
)

const minimumStoryboardSplitDuration = 0.4

func CreateStoryboardShot(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, afterShotID string, shot StoryboardShot) (Storyboard, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, err
	}
	if err := validateStoryboardShotInput(shot); err != nil {
		return Storyboard{}, err
	}
	shot = cleanStoryboardShot(shot)
	shot.ID = newScreenplayID("sb_shot")
	insertAt := len(storyboard.Shots)
	if strings.TrimSpace(afterShotID) != "" {
		index := storyboardShotIndex(storyboard.Shots, afterShotID)
		if index < 0 {
			return Storyboard{}, fmt.Errorf("%w: 分镜镜头 %s", ErrNotFound, afterShotID)
		}
		insertAt = index + 1
	}
	storyboard.Shots = append(storyboard.Shots, StoryboardShot{})
	copy(storyboard.Shots[insertAt+1:], storyboard.Shots[insertAt:])
	storyboard.Shots[insertAt] = shot
	return saveEditedStoryboard(ctx, store, storyboard)
}

func UpdateStoryboardShot(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, shotID string, changes StoryboardShot) (Storyboard, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, err
	}
	index := storyboardShotIndex(storyboard.Shots, shotID)
	if index < 0 {
		return Storyboard{}, fmt.Errorf("%w: 分镜镜头 %s", ErrNotFound, shotID)
	}
	clean := cleanStoryboardShot(changes)
	preserveStoryboardProtocol(&clean, storyboard.Shots[index])
	if err := validateStoryboardShotInput(clean); err != nil {
		return Storyboard{}, err
	}
	clean.ID = storyboard.Shots[index].ID
	clean.StoryboardID = storyboard.Shots[index].StoryboardID
	storyboard.Shots[index] = clean
	return saveEditedStoryboard(ctx, store, storyboard)
}

func DeleteStoryboardShot(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, shotID string) (Storyboard, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, err
	}
	index := storyboardShotIndex(storyboard.Shots, shotID)
	if index < 0 {
		return Storyboard{}, fmt.Errorf("%w: 分镜镜头 %s", ErrNotFound, shotID)
	}
	storyboard.Shots = append(storyboard.Shots[:index], storyboard.Shots[index+1:]...)
	return saveEditedStoryboard(ctx, store, storyboard)
}

func SplitStoryboardShot(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, shotID string) (Storyboard, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, err
	}
	index := storyboardShotIndex(storyboard.Shots, shotID)
	if index < 0 {
		return Storyboard{}, fmt.Errorf("%w: 分镜镜头 %s", ErrNotFound, shotID)
	}
	if storyboard.Shots[index].Duration <= minimumStoryboardSplitDuration {
		return Storyboard{}, fmt.Errorf("%w: 镜头时长必须大于 %.1f 秒才能拆分", ErrInvalidState, minimumStoryboardSplitDuration)
	}
	first := storyboard.Shots[index]
	second := first
	first.Duration = first.Duration / 2
	second.Duration = first.Duration
	second.ID = newScreenplayID("sb_shot")
	storyboard.Shots[index] = first
	storyboard.Shots = append(storyboard.Shots, StoryboardShot{})
	copy(storyboard.Shots[index+2:], storyboard.Shots[index+1:])
	storyboard.Shots[index+1] = second
	return saveEditedStoryboard(ctx, store, storyboard)
}

func MergeStoryboardShotWithNext(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, shotID string) (Storyboard, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, err
	}
	index := storyboardShotIndex(storyboard.Shots, shotID)
	if index < 0 {
		return Storyboard{}, fmt.Errorf("%w: 分镜镜头 %s", ErrNotFound, shotID)
	}
	if index+1 >= len(storyboard.Shots) {
		return Storyboard{}, fmt.Errorf("%w: 最后一条分镜没有下一条可以合并", ErrInvalidState)
	}
	next := storyboard.Shots[index+1]
	first := &storyboard.Shots[index]
	mergedFields := map[string]any{}
	if len(first.ProtocolFields) > 0 {
		if err := json.Unmarshal(first.ProtocolFields, &mergedFields); err != nil {
			return Storyboard{}, err
		}
	}
	if mergedFields == nil {
		mergedFields = map[string]any{}
	}
	mergedFields["merged_shots"] = []StoryboardShot{*first, next}
	merged, err := json.Marshal(mergedFields)
	if err != nil {
		return Storyboard{}, err
	}
	first.ProtocolFields = merged
	first.ImagePrompt = mergeStoryboardText(first.ImagePrompt, next.ImagePrompt, "\n")
	first.Line = mergeStoryboardText(first.Line, next.Line, "\n")
	first.VoiceActor = mergeStoryboardText(first.VoiceActor, next.VoiceActor, "；")
	first.Emotion = mergeStoryboardText(first.Emotion, next.Emotion, "；")
	if len(next.SourceElements) > 0 {
		if len(first.SourceElements) == 0 {
			first.SourceElements = next.SourceElements
		} else if string(first.SourceElements) != string(next.SourceElements) {
			first.SourceElements, _ = json.Marshal([]json.RawMessage{first.SourceElements, next.SourceElements})
		}
	}
	storyboard.Shots[index].Duration += next.Duration
	storyboard.Shots[index].Characters = mergeStoryboardText(storyboard.Shots[index].Characters, next.Characters, "；")
	storyboard.Shots[index].Scene = mergeStoryboardText(storyboard.Shots[index].Scene, next.Scene, "；")
	storyboard.Shots[index].Props = mergeStoryboardText(storyboard.Shots[index].Props, next.Props, "；")
	storyboard.Shots[index].VideoDesc = mergeStoryboardText(storyboard.Shots[index].VideoDesc, next.VideoDesc, "\n")
	storyboard.Shots = append(storyboard.Shots[:index+1], storyboard.Shots[index+2:]...)
	return saveEditedStoryboard(ctx, store, storyboard)
}

func loadEditableStoryboard(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine) (Storyboard, error) {
	if store == nil || strings.TrimSpace(owner) == "" || strings.TrimSpace(projectID) == "" || episodeNo <= 0 || strings.TrimSpace(string(engine)) == "" {
		return Storyboard{}, fmt.Errorf("%w: 项目、用户、集数和分镜引擎不能为空", ErrInvalidArgument)
	}
	return store.GetStoryboard(ctx, owner, projectID, episodeNo, engine)
}

func saveEditedStoryboard(ctx context.Context, store Store, storyboard Storyboard) (Storyboard, error) {
	for index := range storyboard.Shots {
		storyboard.Shots[index].Seq = index + 1
	}
	storyboard.ShotCount = len(storyboard.Shots)
	return store.SaveStoryboard(ctx, storyboard)
}

func validateStoryboardShotInput(shot StoryboardShot) error {
	imageOnly := strings.TrimSpace(shot.VideoDesc) == "" && strings.TrimSpace(shot.ImagePrompt) != ""
	if shot.Duration < 0 || (shot.Duration == 0 && !imageOnly && !storyboardProtocolOmitsDuration(shot.ProtocolFields)) {
		return fmt.Errorf("%w: 镜头时长必须大于 0 秒", ErrInvalidArgument)
	}
	if strings.TrimSpace(shot.VideoDesc) == "" && strings.TrimSpace(shot.Characters) == "" && strings.TrimSpace(shot.Scene) == "" && strings.TrimSpace(shot.Props) == "" && strings.TrimSpace(shot.ImagePrompt) == "" && strings.TrimSpace(shot.Line) == "" {
		return fmt.Errorf("%w: 分镜内容不能为空", ErrInvalidArgument)
	}
	return nil
}

func cleanStoryboardShot(shot StoryboardShot) StoryboardShot {
	shot.Characters = strings.TrimSpace(shot.Characters)
	shot.Scene = strings.TrimSpace(shot.Scene)
	shot.Props = strings.TrimSpace(shot.Props)
	shot.VideoDesc = strings.TrimSpace(shot.VideoDesc)
	shot.Seq = 0
	shot.ProjectID = ""
	shot.OwnerUsername = ""
	return shot
}

func storyboardShotIndex(shots []StoryboardShot, shotID string) int {
	shotID = strings.TrimSpace(shotID)
	for index := range shots {
		if shots[index].ID == shotID {
			return index
		}
	}
	return -1
}

func mergeStoryboardText(first, second, separator string) string {
	first, second = strings.TrimSpace(first), strings.TrimSpace(second)
	if first == "" {
		return second
	}
	if second == "" || first == second {
		return first
	}
	return first + separator + second
}

// Absent extension fields in older/manual clients retain the saved protocol.
// Explicitly supplied editable strings can be cleared by sending an empty string.
func preserveStoryboardProtocol(dst *StoryboardShot, src StoryboardShot) {
	edited := map[string]bool{}
	for _, key := range dst.EditedProtocolFields {
		edited[key] = true
	}
	if !edited["image_prompt"] && dst.ImagePrompt == "" {
		dst.ImagePrompt = src.ImagePrompt
	}
	if !edited["voice_actor"] && dst.VoiceActor == "" {
		dst.VoiceActor = src.VoiceActor
	}
	if !edited["emotion"] && dst.Emotion == "" {
		dst.Emotion = src.Emotion
	}
	if !edited["line"] && dst.Line == "" {
		dst.Line = src.Line
	}
	if len(dst.SourceElements) == 0 {
		dst.SourceElements = src.SourceElements
	}
	if len(dst.ProtocolFields) == 0 {
		dst.ProtocolFields = src.ProtocolFields
	}
	if dst.SourceSceneIndex == 0 {
		dst.SourceSceneIndex = src.SourceSceneIndex
	}
	if dst.SourceSceneLocalOrder == 0 {
		dst.SourceSceneLocalOrder = src.SourceSceneLocalOrder
	}
	dst.EditedProtocolFields = nil
}

// ValidateStoryboardShotInput shares editor content rules with file import.
func ValidateStoryboardShotInput(shot StoryboardShot) error { return validateStoryboardShotInput(shot) }

// An untimed video row is editable only when its saved protocol matches the
// current supported catalog and that contract does not supply or require time.
func storyboardProtocolOmitsDuration(raw json.RawMessage) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return false
	}
	if _, present := fields["duration_sec"]; present {
		return false
	}
	var contract struct {
		PromptID string `json:"prompt_id"`
		Digest   string `json:"prompt_digest"`
		Protocol string `json:"output_protocol"`
		Parser   string `json:"parser_version"`
	}
	if json.Unmarshal(fields["_cm_contract"], &contract) != nil {
		return false
	}
	selected, err := prompt.ResolveCreativePrompt(contract.PromptID, "storyboard")
	if err != nil || !selected.Supported || selected.SourceID == 0 || selected.Digest != contract.Digest || selected.OutputProtocol != contract.Protocol || selected.ParserVersion != contract.Parser || selected.FixedDurationSec != 0 {
		return false
	}
	for _, field := range selected.RequiredFields {
		if field == "duration_sec" {
			return false
		}
	}
	return true
}
