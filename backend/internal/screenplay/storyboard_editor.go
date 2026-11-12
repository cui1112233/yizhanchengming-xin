package screenplay

import (
	"context"
	"fmt"
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
	if err := validateStoryboardShotInput(changes); err != nil {
		return Storyboard{}, err
	}
	clean := cleanStoryboardShot(changes)
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
	if shot.Duration <= 0 {
		return fmt.Errorf("%w: 镜头时长必须大于 0 秒", ErrInvalidArgument)
	}
	if strings.TrimSpace(shot.VideoDesc) == "" && strings.TrimSpace(shot.Characters) == "" && strings.TrimSpace(shot.Scene) == "" && strings.TrimSpace(shot.Props) == "" {
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
