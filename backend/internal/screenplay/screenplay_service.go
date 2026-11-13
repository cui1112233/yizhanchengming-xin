package screenplay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/prompt/luoshui"
)

// MaxSingleCallRunes 是单次提交的硬上限：超过 60 万字必须走 segments（规格 §1 边界）。
const MaxSingleCallRunes = 600000

type A1Request struct {
	Owner     string            `json:"owner"`
	ProjectID string            `json:"project_id"`
	Content   string            `json:"content"`
	Segments  []string          `json:"segments"`
	Model     string            `json:"model"`
	PromptID  string            `json:"prompt_id"`
	ExtraVars map[string]string `json:"extra_vars"`
	Confirmed bool              `json:"confirmed"`
}

type BRequest struct {
	Owner         string            `json:"owner"`
	ProjectID     string            `json:"project_id"`
	Novel         string            `json:"novel"`
	Model         string            `json:"model"`
	PromptID      string            `json:"prompt_id"`
	ChapterRange  string            `json:"chapter_range"`
	EpisodeTarget string            `json:"episode_target"`
	ExtraNote     string            `json:"extra_note"`
	ExtraVars     map[string]string `json:"extra_vars"`
	Confirmed     bool              `json:"confirmed"`
}

// ScreenplayResult 同时承载「已落库结果」与「auto 权限下的待确认计划」。
type ScreenplayResult struct {
	Document             Document        `json:"document,omitempty"`
	Shots                []Shot          `json:"shots,omitempty"`
	Characters           []Character     `json:"characters,omitempty"`
	Scenes               []Scene         `json:"scenes,omitempty"`
	EpisodeCount         int             `json:"episode_count"`
	ShotCount            int             `json:"shot_count"`
	SegmentCount         int             `json:"segment_count"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	PendingActions       []PendingAction `json:"pending_actions,omitempty"`
}

// ScreenplayService 实现能力 1（A1）与能力 2（B 档）。
type ScreenplayService struct {
	Store     Store
	Generator TextGenerator
	Prompts   *luoshui.Registry
}

func (s *ScreenplayService) GenerateA1(ctx context.Context, req A1Request) (ScreenplayResult, error) {
	content, segments, err := normalizeSource(req.Content, req.Segments)
	if err != nil {
		return ScreenplayResult{}, err
	}
	project, err := s.Store.GetProject(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return ScreenplayResult{}, err
	}
	if plan, blocked := pendingPlan(project, req.Confirmed, "1_screenplay", "运行 A1 小说转剧本", "调用 A1 提示词并写入 screenplay.md"); blocked {
		return plan, nil
	}
	promptID := defaultString(req.PromptID, selectedScriptPrompt(project), "a1_novel_to_screenplay")
	selected, resolveErr := ResolveProjectPrompt(project, promptID, "script")
	if resolveErr != nil {
		return ScreenplayResult{}, fmt.Errorf("%w: %v", ErrInvalidArgument, resolveErr)
	}
	systemPrompt, renderErr := renderScriptLibrary(ctx, s.Store, req.Owner, req.ProjectID, selected.Body)
	if renderErr != nil {
		return ScreenplayResult{}, renderErr
	}

	outputs := make([]string, 0, len(segments))
	for _, segment := range segments {
		response, err := s.Generator.Generate(ctx, TextRequest{
			Model:        defaultString(req.Model, project.ModelID),
			SystemPrompt: systemPrompt,
			JSONResponse: promptID == prompt.CreativeScript144,
			UserPrompt:   styledAssetContent(project, luoshui.Render(segment, req.ExtraVars)),
		})
		if err != nil {
			return ScreenplayResult{}, wrapAIError(err)
		}
		if ctx.Err() != nil {
			return ScreenplayResult{}, ctx.Err()
		}
		output := strings.TrimSpace(response.Content)
		if output == "" {
			return ScreenplayResult{}, ErrAIEmptyContent
		}
		outputs = append(outputs, response.Content)
	}
	content = strings.Join(outputs, "\n\n")
	if strings.TrimSpace(content) == "" {
		return ScreenplayResult{}, ErrAIEmptyBody
	}
	return s.persist(ctx, req.Owner, req.ProjectID, "a1", content, len(segments))
}

func (s *ScreenplayService) GenerateB(ctx context.Context, req BRequest) (ScreenplayResult, error) {
	novel := strings.TrimSpace(req.Novel)
	if novel == "" {
		return ScreenplayResult{}, ErrEmptyNovel
	}
	if len([]rune(novel)) > MaxSingleCallRunes {
		return ScreenplayResult{}, fmt.Errorf("%w: 超过 60 万字，请按章拆分后分批提交", ErrInvalidArgument)
	}
	project, err := s.Store.GetProject(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return ScreenplayResult{}, err
	}
	if plan, blocked := pendingPlan(project, req.Confirmed, "1_screenplay", "运行 B 档导演级剧本", "调用 B 档提示词并写入 screenplay.md"); blocked {
		return plan, nil
	}
	promptID := defaultString(req.PromptID, selectedScriptPrompt(project), "b_director_screenplay")
	selected, resolveErr := ResolveProjectPrompt(project, promptID, "script")
	if resolveErr != nil {
		return ScreenplayResult{}, fmt.Errorf("%w: %v", ErrInvalidArgument, resolveErr)
	}
	systemPrompt, renderErr := renderScriptLibrary(ctx, s.Store, req.Owner, req.ProjectID, selected.Body)
	if renderErr != nil {
		return ScreenplayResult{}, renderErr
	}
	userPrompt := buildBNovelPrompt(req, novel)
	response, err := s.Generator.Generate(ctx, TextRequest{
		Model:        defaultString(req.Model, project.ModelID),
		SystemPrompt: systemPrompt,
		JSONResponse: promptID == prompt.CreativeScript144,
		UserPrompt:   styledAssetContent(project, luoshui.Render(userPrompt, req.ExtraVars)),
	})
	if err != nil {
		return ScreenplayResult{}, wrapAIError(err)
	}
	if ctx.Err() != nil {
		return ScreenplayResult{}, ctx.Err()
	}
	output := strings.TrimSpace(response.Content)
	if output == "" {
		return ScreenplayResult{}, ErrAIEmptyContent
	}
	return s.persist(ctx, req.Owner, req.ProjectID, "b", response.Content, 1)
}

func (s *ScreenplayService) persist(ctx context.Context, owner, projectID, source, content string, segmentCount int) (ScreenplayResult, error) {
	editable, err := editableScreenplayOutput(content)
	if err != nil {
		return ScreenplayResult{}, err
	}
	parsed := ParseScreenplayMarkdown(editable)
	if len(parsed.Shots) == 0 || len(parsed.Episodes) == 0 {
		return ScreenplayResult{}, fmt.Errorf("%w screenplay.md 未产出可解析的集/场", ErrArtifactInvalid)
	}
	documentID := newScreenplayID("doc")
	for i := range parsed.Shots {
		parsed.Shots[i].ID = newScreenplayID("shot")
		parsed.Shots[i].DocumentID = documentID
		parsed.Shots[i].ProjectID = projectID
		parsed.Shots[i].OwnerUsername = owner
	}
	for i := range parsed.Characters {
		parsed.Characters[i].ID = newScreenplayID("char")
		parsed.Characters[i].DocumentID = documentID
		parsed.Characters[i].ProjectID = projectID
		parsed.Characters[i].OwnerUsername = owner
	}
	for i := range parsed.Scenes {
		parsed.Scenes[i].ID = newScreenplayID("scene")
		parsed.Scenes[i].DocumentID = documentID
		parsed.Scenes[i].ProjectID = projectID
		parsed.Scenes[i].OwnerUsername = owner
	}
	doc := Document{
		ID:            documentID,
		ProjectID:     projectID,
		OwnerUsername: owner,
		Source:        source,
		Content:       editable,
		RawResult:     content,
		ChapterCount:  0,
		EpisodeCount:  len(parsed.Episodes),
		SegmentCount:  segmentCount,
	}
	saved, err := s.Store.SaveScreenplay(ctx, doc, parsed.Shots, parsed.Characters, parsed.Scenes)
	if err != nil {
		return ScreenplayResult{}, err
	}
	return ScreenplayResult{
		Document:     saved,
		Shots:        parsed.Shots,
		Characters:   parsed.Characters,
		Scenes:       parsed.Scenes,
		EpisodeCount: len(parsed.Episodes),
		ShotCount:    len(parsed.Shots),
		SegmentCount: segmentCount,
	}, nil
}

func (s *ScreenplayService) prompts() *luoshui.Registry {
	if s.Prompts != nil {
		return s.Prompts
	}
	return luoshui.MustDefault()
}

// normalizeSource 校验空输入与分段，返回待调用片段列表。
func normalizeSource(content string, segments []string) (string, []string, error) {
	if len(segments) > 0 {
		cleaned := make([]string, 0, len(segments))
		for _, segment := range segments {
			trimmed := strings.TrimSpace(segment)
			if trimmed == "" {
				return "", nil, ErrEmptyContent
			}
			cleaned = append(cleaned, trimmed)
		}
		return strings.Join(cleaned, "\n\n"), cleaned, nil
	}
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "", nil, ErrEmptyContent
	}
	if len([]rune(trimmed)) > MaxSingleCallRunes {
		return "", nil, fmt.Errorf("%w: 超过 60 万字，必须使用 segments 分段", ErrInvalidArgument)
	}
	return trimmed, []string{trimmed}, nil
}

func buildBNovelPrompt(req BRequest, novel string) string {
	var builder strings.Builder
	if value := strings.TrimSpace(req.ChapterRange); value != "" {
		builder.WriteString("本次改编章节：")
		builder.WriteString(value)
		builder.WriteString("\n")
	}
	if value := strings.TrimSpace(req.EpisodeTarget); value != "" {
		builder.WriteString("单集目标时长：")
		builder.WriteString(value)
		builder.WriteString("\n")
	}
	if value := strings.TrimSpace(req.ExtraNote); value != "" {
		builder.WriteString("特殊要求：")
		builder.WriteString(value)
		builder.WriteString("\n")
	}
	if builder.Len() > 0 {
		builder.WriteString("\n")
	}
	builder.WriteString("# 待改编的小说原文\n\n")
	builder.WriteString(novel)
	return builder.String()
}

// pendingPlan 在 auto 权限下只返回计划，不调用模型、不写库。
func pendingPlan(project Project, confirmed bool, stepKey, title, detail string) (ScreenplayResult, bool) {
	if project.PermissionMode != PermissionAuto || confirmed {
		return ScreenplayResult{}, false
	}
	return ScreenplayResult{
		RequiresConfirmation: true,
		PendingActions: []PendingAction{{
			ID:              newScreenplayID("action"),
			Kind:            "skill",
			StepKey:         stepKey,
			Title:           title,
			Detail:          detail + "；当前为 auto 权限，需人工确认后执行",
			RequiresConfirm: true,
			CreatedAt:       time.Now().UTC(),
		}},
	}, true
}

func defaultString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func newScreenplayID(prefix string) string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buffer)
}
