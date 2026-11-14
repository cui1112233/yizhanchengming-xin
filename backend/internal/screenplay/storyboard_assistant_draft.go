package screenplay

import (
	"context"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strconv"
	"strings"
)

type StoryboardAssistantDraftRequest struct {
	ChapterID         string `json:"chapter_id"`
	Task              string `json:"task"`
	PlanMode          string `json:"plan_mode"`
	PlanningChapterID string `json:"planning_chapter_id"`
	UseOriginal       bool   `json:"use_original_content_directly"`
	PlanDensity       string `json:"plan_density"`
}

func (s *AuthoringService) StoryboardAssistantDraft(ctx context.Context, owner, id string, in StoryboardAssistantDraftRequest) (string, error) {
	p, e := s.Store.GetProject(ctx, owner, id)
	if e != nil {
		return "", e
	}
	state, e := s.Load(ctx, owner, id)
	if e != nil {
		return "", e
	}
	index := -1
	for i, c := range state.Chapters {
		if c.ID == in.ChapterID {
			index = i
			break
		}
	}
	if index < 0 {
		return "", ErrNotFound
	}
	if in.Task != "plan" && in.Task != "generate" {
		return "", ErrInvalidArgument
	}
	modes := map[string]string{"new": "先生成并保存新规划", "existing": "复用已有规划，不改写规划正文", "none": "不使用任何已有规划，直接基于完整输入生成"}
	mode, ok := modes[in.PlanMode]
	if !ok {
		return "", ErrInvalidArgument
	}
	if in.Task == "plan" {
		mode = "仅生成并保存新规划，不执行分镜生成"
	}
	source := "当前章节完整剧本"
	if in.UseOriginal {
		source = "当前章节完整原文"
	}
	planSource := ""
	if in.Task == "generate" && in.PlanMode == "existing" {
		found := false
		for _, c := range state.Chapters {
			if c.ID == in.PlanningChapterID && strings.TrimSpace(c.PlanningText) != "" {
				planSource = "复用规划来源：「" + c.Title + "」。"
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("%w: 未选择有效的已有规划", ErrInvalidArgument)
		}
	}
	density := in.PlanDensity
	if density == "" {
		density = EffectiveCreativeConfig(p).PlanDensity
	}
	labels := map[string]string{"compact": "更少小节", "standard": "标准", "more": "更多小节", "sd25": "SD2.5专项规划（30秒一组）"}
	task := map[string]string{"plan": "仅生成并保存规划，本次不生成分镜", "generate": "生成分镜脚本"}[in.Task]
	return strings.NewReplacer("{chapter}", state.Chapters[index].Title, "{episode}", strconv.Itoa(index+1), "{task}", task, "{mode}", mode, "{density}", labels[NormalizePlanDensity(density)], "{source}", source, "{planning_source}", planSource).Replace(prompt.ScreenplayStoryboardAssistantDraft), nil
}
