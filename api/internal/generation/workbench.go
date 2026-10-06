package generation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const maxWorkbenchSourceRunes = 180000

const workbenchExtractionPrompt = "你是小说人物与场景结构化提取器。只提取原文明确存在的人物与场景，不补写身份、关系、外貌或地点。必须只返回合法 JSON 对象，结构为 characters 和 scenes 两个数组。characters 每项包含 id、name、description、protagonist；scenes 每项包含 id、name、description。protagonist 仅在原文主视角或核心主角有明确依据时为 true。不要输出 Markdown、解释或代码围栏。"

func normalizeWorkbenchRequest(req *RunBookRequest) error {
	if req == nil {
		return ErrInvalid
	}
	if req.Action == "" {
		req.Action = WorkbenchActionGenerate
	}
	if !req.Workbench && req.Action == WorkbenchActionGenerate {
		return nil
	}
	req.Workbench = true
	if req.Action != WorkbenchActionGenerate && req.Action != WorkbenchActionExtract {
		return fmt.Errorf("%w: unsupported workbench action", ErrInvalid)
	}
	if utf8.RuneCountInString(req.SourceText) > maxWorkbenchSourceRunes {
		return fmt.Errorf("%w: sourceText exceeds %d characters", ErrInvalid, maxWorkbenchSourceRunes)
	}
	if req.OpeningMode == "" {
		req.OpeningMode = OpeningContinuous
	}
	switch req.OpeningMode {
	case OpeningContinuous, OpeningSegmented:
		req.HookEnabled = false
	case OpeningHook:
		req.HookEnabled = true
	default:
		return fmt.Errorf("%w: unsupported openingMode", ErrInvalid)
	}
	if req.OutputMode == "" {
		req.OutputMode = OutputCanvas
	}
	switch req.OutputMode {
	case OutputCanvas, OutputScript, OutputStory, OutputShotlist, OutputQ:
	default:
		return fmt.Errorf("%w: unsupported outputMode", ErrInvalid)
	}
	if req.OutputMode == OutputStory {
		req.PlotMode = true
	}
	var err error
	req.Characters, err = normalizeScriptEntities(req.Characters, "character")
	if err != nil {
		return err
	}
	req.Scenes, err = normalizeScriptEntities(req.Scenes, "scene")
	return err
}

func normalizeScriptEntities(values []ScriptEntity, kind string) ([]ScriptEntity, error) {
	if len(values) > 40 {
		return nil, fmt.Errorf("%w: too many %s entities", ErrInvalid, kind)
	}
	out := make([]ScriptEntity, 0, len(values))
	for index, value := range values {
		value.ID = strings.TrimSpace(value.ID)
		value.Kind = kind
		value.Name = strings.TrimSpace(value.Name)
		value.Description = strings.TrimSpace(value.Description)
		if value.Name == "" {
			continue
		}
		if value.ID == "" {
			value.ID = fmt.Sprintf("%s-%d", kind, index+1)
		}
		if len(value.ReferenceImages) > 8 {
			value.ReferenceImages = value.ReferenceImages[:8]
		}
		refs := make([]string, 0, len(value.ReferenceImages))
		for _, raw := range value.ReferenceImages {
			url := strings.TrimSpace(raw)
			if url == "" {
				continue
			}
			if len(url) > 2048 {
				return nil, fmt.Errorf("%w: reference image URL too long", ErrInvalid)
			}
			refs = append(refs, url)
		}
		value.ReferenceImages = refs
		out = append(out, value)
	}
	return out, nil
}

func effectiveWorkbenchSource(req RunBookRequest, original string) (string, error) {
	source := strings.TrimSpace(original)
	if req.Workbench && strings.TrimSpace(req.SourceText) != "" {
		source = strings.TrimSpace(req.SourceText)
	}
	if source == "" {
		return "", fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	if utf8.RuneCountInString(source) > maxWorkbenchSourceRunes {
		return "", fmt.Errorf("%w: source text exceeds %d characters", ErrInvalid, maxWorkbenchSourceRunes)
	}
	return source, nil
}

func openingModeRule(mode OpeningMode) string {
	switch mode {
	case OpeningHook:
		return "爆款开头：从原文已有事实中选择最强冲突、反差或悬念作为开场，随后自然衔接原文；禁止虚构关键事件、身份、结果或关系。"
	case OpeningSegmented:
		return "分段开头：按原文自然剧情段落拆分，每段先给出本段最有信息量的可视化入口，再按原文因果推进；不得跨段偷跑后文关键结果。"
	default:
		return "连续开头：严格按原文时间线连续展开，从原文起点开始可视化；不闪回、不前置结局、不创造原文不存在的冲突。"
	}
}

func outputModeRule(mode OutputMode) string {
	switch mode {
	case OutputScript:
		return "剧本模式：按专业短剧场次、动作和对白组织；空间或时间真正变化时才换场。"
	case OutputStory:
		return "剧情模式：每个分镜卡内部使用连续时间轴，写清可见动作、人物状态、道具与情绪变化。"
	case OutputShotlist:
		return "分镜模式：输出 ### 分镜N 卡片；卡内时间从 00:00 开始，镜头首尾连续，写清景别、机位、运镜、动作与镜头落点。"
	case OutputQ:
		return "Q版模式：保持原文人物关系、事件和动作事实不变，以可爱夸张但不篡改剧情的视觉方式组织分镜。"
	default:
		return "画布模式：输出 ### 分镜N 卡片；每张卡片可独立编辑和用于视频生成，卡内时间轴从 00:00 开始并保持动作连续。"
	}
}

func workbenchConstraintText(value ScriptConstraints) string {
	parts := make([]string, 0, 4)
	if text := strings.TrimSpace(value.VisualPrefix); text != "" {
		parts = append(parts, "画面前缀："+text)
	}
	if text := strings.TrimSpace(value.Quality); text != "" {
		parts = append(parts, "画质约束："+text)
	}
	if text := strings.TrimSpace(value.PictureLimit); text != "" {
		parts = append(parts, "画面限制："+text)
	}
	if text := strings.TrimSpace(value.NegativePrompt); text != "" {
		parts = append(parts, "负面提示词："+text)
	}
	return strings.Join(parts, "\n")
}

func mergeWorkbenchRules(base, workbench string) string {
	base = strings.TrimSpace(base)
	workbench = strings.TrimSpace(workbench)
	if base == "" {
		return workbench
	}
	if workbench == "" {
		return base
	}
	return base + "\n" + workbench
}

func buildScriptWorkbenchUserPrompt(req RunBookRequest, source string) string {
	characters, _ := json.Marshal(req.Characters)
	scenes, _ := json.Marshal(req.Scenes)
	sections := []string{
		"【权威原文】\n" + source,
		"【开头模式】" + openingModeLabel(req.OpeningMode) + "\n" + openingModeRule(req.OpeningMode),
		"【输出模式】" + outputModeLabel(req.OutputMode) + "\n" + outputModeRule(req.OutputMode),
	}
	if len(req.Characters) > 0 {
		sections = append(sections, "【已确认人物】\n"+string(characters))
	}
	if len(req.Scenes) > 0 {
		sections = append(sections, "【已确认场景】\n"+string(scenes))
	}
	if constraints := workbenchConstraintText(req.Constraints); constraints != "" {
		sections = append(sections, "【本次约束】\n"+constraints)
	}
	sections = append(sections, "只输出最终可编辑成品，不复述分析过程。人物、场景、参考图仅作为一致性事实；主角标记只影响镜头关注优先级，不允许把未出场主角强行写入剧情。")
	return strings.Join(sections, "\n\n")
}

func openingModeLabel(mode OpeningMode) string {
	switch mode {
	case OpeningHook:
		return "爆款开头"
	case OpeningSegmented:
		return "分段开头"
	default:
		return "连续开头"
	}
}

func outputModeLabel(mode OutputMode) string {
	switch mode {
	case OutputScript:
		return "剧本模式"
	case OutputStory:
		return "剧情模式"
	case OutputShotlist:
		return "分镜模式"
	case OutputQ:
		return "Q版模式"
	default:
		return "画布模式"
	}
}

func cleanJSONBlock(value string) string {
	text := strings.TrimSpace(value)
	if strings.HasPrefix(text, "```") {
		if index := strings.Index(text, "\n"); index >= 0 {
			text = text[index+1:]
		}
		text = strings.TrimSpace(strings.TrimSuffix(text, "```"))
	}
	return text
}

func parseEntityExtraction(value string) (EntityExtractionResult, error) {
	var decoded EntityExtractionResult
	if err := json.Unmarshal([]byte(cleanJSONBlock(value)), &decoded); err != nil {
		return EntityExtractionResult{}, fmt.Errorf("%w: extraction output is not valid JSON", ErrUnavailable)
	}
	characters, err := normalizeScriptEntities(decoded.Characters, "character")
	if err != nil {
		return EntityExtractionResult{}, err
	}
	scenes, err := normalizeScriptEntities(decoded.Scenes, "scene")
	if err != nil {
		return EntityExtractionResult{}, err
	}
	if len(characters) == 0 && len(scenes) == 0 {
		return EntityExtractionResult{}, fmt.Errorf("%w: extraction returned no characters or scenes", ErrUnavailable)
	}
	decoded.Characters, decoded.Scenes = characters, scenes
	return decoded, nil
}

func (s *Service) extractWorkbenchEntities(ctx context.Context, bookID int64, source string) (EntityExtractionResult, error) {
	if s == nil || s.provider == nil {
		return EntityExtractionResult{}, ErrUnavailable
	}
	output, err := s.provider.Complete(ctx, TextRequest{
		BookID:       bookID,
		Stage:        StageExtract,
		SystemPrompt: workbenchExtractionPrompt,
		UserPrompt:   source,
	})
	if err != nil {
		return EntityExtractionResult{}, err
	}
	return parseEntityExtraction(output)
}
