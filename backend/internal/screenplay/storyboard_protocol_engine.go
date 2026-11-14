package screenplay

import (
	"fmt"
	"qiantie/backend/internal/prompt"
)

func ValidStoryboardEngine(engine StoryboardEngine) bool {
	_, _, err := resolveStoryboardEngine(engine)
	return err == nil
}

func StoryboardEngineForPrompt(id string) (StoryboardEngine, error) {
	p, err := prompt.ResolveCreativePrompt(id, "storyboard")
	if err != nil || !p.Supported {
		return "", fmt.Errorf("%w: 分镜模板不可用", ErrInvalidArgument)
	}
	switch p.ID {
	case prompt.CreativeStoryboard307:
		return EngineXiaoluoV9, nil
	case "storyboard_xiaoluo_v7_15s":
		return EngineXiaoluoV7, nil
	case "storyboard_xiaoyunque_v3":
		return EngineXiaoyunqueV3, nil
	}
	if p.SourceID > 0 {
		return StoryboardEngine(fmt.Sprintf("prompt_%d", p.SourceID)), nil
	}
	return "", ErrInvalidArgument
}
