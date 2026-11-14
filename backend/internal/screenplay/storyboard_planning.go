package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
)

func NormalizePlanDensity(value any) string {
	switch v := value.(type) {
	case float64:
		if v == 0 {
			return "compact"
		}
		if v == 2 {
			return "more"
		}
		if v == 3 {
			return "sd25"
		}
	case int:
		return NormalizePlanDensity(float64(v))
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		switch s {
		case "0", "compact":
			return "compact"
		case "2", "more":
			return "more"
		case "3", "sd25", "sd2.5", "sd2_5", "sd2.5专项", "sd2.5专项规划", "sd25专项", "sd25专项规划", "seedance25", "seedance2.5":
			return "sd25"
		}
	}
	return "standard"
}

type StoryboardPlanRequest struct {
	IdempotencyKey   string `json:"idempotencyKey"`
	ChapterID        string `json:"chapter_id"`
	Model            string `json:"model"`
	ExpectedRevision *int64 `json:"expected_revision"`
	UseOriginal      bool   `json:"use_original_content_directly"`
}
type StoryboardPlanResult struct {
	Chapter  Chapter `json:"chapter"`
	Revision int64   `json:"revision"`
}

func (s *AuthoringService) GenerateStoryboardPlan(ctx context.Context, owner, project string, in StoryboardPlanRequest, generator TextGenerator) (StoryboardPlanResult, error) {
	var out StoryboardPlanResult
	if in.ExpectedRevision == nil {
		return out, ErrRevisionConflict
	}
	if generator == nil {
		return out, ErrAPIKeyMissing
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" || len(in.IdempotencyKey) > 256 {
		return out, fmt.Errorf("%w: 请提供有效规划请求标识", ErrInvalidArgument)
	}
	err := s.WithLock(ctx, owner, project, func() error {
		p, e := s.Store.GetProject(ctx, owner, project)
		if e != nil {
			return e
		}
		state, e := s.Load(ctx, owner, project)
		if e != nil {
			return e
		}
		index := -1
		for i, c := range state.Chapters {
			if c.ID == in.ChapterID {
				index = i
				break
			}
		}
		if index < 0 {
			return ErrNotFound
		}
		chapter := state.Chapters[index]
		content := chapter.ScriptText
		if in.UseOriginal {
			content = chapter.SourceText
		}
		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("%w: 目标章节的%s为空", ErrEmptyContent, map[bool]string{true: "原文", false: "完整剧本"}[in.UseOriginal])
		}
		model := firstNonEmpty(in.Model, p.ModelID)
		if model == "" {
			return fmt.Errorf("%w: 请选择文字模型", ErrInvalidArgument)
		}
		selected, configDigest, style, e := storyboardPlanningContext(p)
		if e != nil {
			return e
		}
		density := NormalizePlanDensity(EffectiveCreativeConfig(p).PlanDensity)
		digest := contentDigest(fmt.Sprintf("%s\n%s\n%s\n%s\n%t\n%s", model, selected.ID, contentDigest(selected.Body), configDigest, in.UseOriginal, content))
		if chapter.PlanningRequestKey == in.IdempotencyKey {
			if chapter.PlanningRequestDigest != digest {
				return fmt.Errorf("%w: 同一规划请求标识的输入已变化，请核对已有结果", ErrInvalidArgument)
			}
			if chapter.PlanningStatus == "succeeded" {
				out = StoryboardPlanResult{Chapter: chapter, Revision: state.Revision}
				return nil
			}
			return fmt.Errorf("%w: 本次规划状态为%s，请核对已有结果后再明确重试", ErrInvalidState, chapter.PlanningStatus)
		}
		if chapter.PlanningStatus == "running" {
			return fmt.Errorf("%w: 本章节规划正在生成，请等待后刷新", ErrInvalidState)
		}
		if state.Revision != *in.ExpectedRevision {
			return ErrRevisionConflict
		}
		chapter.PlanningRequestKey = in.IdempotencyKey
		chapter.PlanningRequestDigest = digest
		chapter.PlanningStatus = "running"
		chapter.PlanningModel = model
		state.Chapters[index] = chapter
		reserved, e := s.Save(ctx, owner, project, state, state.Revision)
		if e != nil {
			return e
		}
		state = reserved
		response, runErr := generator.Generate(ctx, TextRequest{Model: model, SystemPrompt: prompt.ScreenplayPlanningReference(selected.ID, selected.Name, selected.Body, density, style), UserPrompt: content, MaxOutputTokens: 16000, Temperature: .4})
		if ctx.Err() != nil {
			runErr = ctx.Err()
		}
		text := strings.TrimSpace(response.Content)
		if runErr == nil && (text == "" || len([]rune(text)) > 100000) {
			runErr = fmt.Errorf("%w: 规划正文为空或过长", ErrArtifactInvalid)
		}
		if runErr == nil {
			chapter.PlanningText = text
			chapter.PlanningPromptID = selected.ID
			chapter.PlanningPromptDigest = contentDigest(selected.Body)
			chapter.PlanningConfigDigest = configDigest
			chapter.PlanDensity = density
			chapter.PlanningSourceSHA256 = contentDigest(content)
			chapter.PlanningUseOriginal = in.UseOriginal
			chapter.PlanningStatus = "succeeded"
		} else {
			chapter.PlanningStatus = "failed"
			if ctx.Err() != nil {
				chapter.PlanningStatus = "cancelled"
			}
		}
		state.Chapters[index] = chapter
		saved, e := s.Save(context.WithoutCancel(ctx), owner, project, state, state.Revision)
		if e != nil {
			return e
		}
		out = StoryboardPlanResult{Chapter: chapter, Revision: saved.Revision}
		return runErr
	})
	return out, err
}

// Planning context fingerprints every saved setting used by the planner. The
// complete selected body is hashed directly, even when a stored digest is stale.
func storyboardPlanningContext(p Project) (CreativePromptSnapshot, string, string, error) {
	cfg := EffectiveCreativeConfig(p)
	selected, err := ResolveProjectPrompt(p, cfg.StoryboardPromptID, "storyboard")
	if err != nil {
		return selected, "", "", fmt.Errorf("%w: %v", ErrInvalidArgument, err)
	}
	if strings.TrimSpace(selected.Body) == "" {
		return selected, "", "", fmt.Errorf("%w: 所选分镜提示词正文为空", ErrInvalidArgument)
	}
	style := firstNonEmpty(p.StyleDesc, p.ArtStyle)
	if cfg.Style != nil {
		style = cfg.Style.Prompt
	}
	raw, _ := json.Marshal(struct {
		Density   string
		Style     *CreativeStyleSnapshot
		StyleText string
	}{NormalizePlanDensity(cfg.PlanDensity), cfg.Style, style})
	return selected, contentDigest(string(raw)), style, nil
}
