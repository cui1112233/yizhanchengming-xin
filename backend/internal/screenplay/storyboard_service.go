package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/prompt/luoshui"
)

// 分镜词两套并存（规格 §6 / §7）：v7 与小云雀 v3 输出结构同为
// storyboard[] 六字段，但语义与校验不同，按「剧集 × 引擎」独立落表。
const (
	StoryboardMaxGroupSecondsV7 = 15.0
	StoryboardMinGroupSecondsV3 = 10.0
	StoryboardMaxGroupSecondsV3 = 15.0
)

type StoryboardRequest struct {
	Owner     string           `json:"owner"`
	ProjectID string           `json:"project_id"`
	Engine    StoryboardEngine `json:"engine"`
	// EpisodeNo 为 0 时按全部剧集逐集生成；>0 只生成该集。
	EpisodeNo int            `json:"episode_no"`
	Content   string         `json:"content"`
	Episodes  map[int]string `json:"episodes"`
	Style     string         `json:"style"`
	Synopsis  string         `json:"synopsis"`
	Model     string         `json:"model"`
	Confirmed bool           `json:"confirmed"`
}

type StoryboardResult struct {
	Storyboards          []Storyboard    `json:"storyboards"`
	EpisodeCount         int             `json:"episode_count"`
	GroupCount           int             `json:"group_count"`
	Warnings             []string        `json:"warnings,omitempty"`
	PartialFailure       bool            `json:"partial_failure"`
	FailedEpisodes       []int           `json:"failed_episodes,omitempty"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	PendingActions       []PendingAction `json:"pending_actions,omitempty"`
}

// StoryboardService 实现能力 6（小洛 v7-15s）与能力 7（小云雀 v3）。
type StoryboardService struct {
	Store     Store
	Generator TextGenerator
	Prompts   *luoshui.Registry
}

type episodeInput struct {
	No      int
	Content string
}

// Generate 逐集调用分镜词提示词。每集一次调用，集内分段由模型内部完成
// （规格 §6「一次一集；集内按场景片段分段」；agent.pyd step5 以「第{N}集」拼提示词）。
func (s *StoryboardService) Generate(ctx context.Context, req StoryboardRequest) (StoryboardResult, error) {
	engine, promptID, err := resolveStoryboardEngine(req.Engine)
	if err != nil {
		return StoryboardResult{}, err
	}
	project, err := s.Store.GetProject(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return StoryboardResult{}, err
	}
	if project.PermissionMode == PermissionAuto && !req.Confirmed {
		return StoryboardResult{
			RequiresConfirmation: true,
			PendingActions: []PendingAction{newPendingAction("skill", "5_storyboard",
				fmt.Sprintf("运行分镜词（%s）", engine),
				"调用分镜词提示词并写入分镜表；当前为 auto 权限，需人工确认后执行")},
		}, nil
	}

	episodes, err := s.resolveEpisodes(ctx, req)
	if err != nil {
		return StoryboardResult{}, err
	}
	if len(episodes) == 0 {
		return StoryboardResult{}, ErrEmptyContent
	}

	systemPrompt, ok := s.prompts().RuntimeBody(promptID)
	if !ok {
		return StoryboardResult{}, fmt.Errorf("%w: 缺少提示词 %s", ErrInvalidArgument, promptID)
	}
	roles, scenes, props, err := s.libraryPlaceholders(ctx, req)
	if err != nil {
		return StoryboardResult{}, err
	}
	style := firstNonEmpty(project.StyleDesc, req.Style, project.ArtStyle)
	synopsis := firstNonEmpty(req.Synopsis, project.Synopsis)

	result := StoryboardResult{EpisodeCount: len(episodes)}
	known := libraryNameSets(roles, scenes, props)
	var firstErr error
	for index, episode := range episodes {
		content := strings.TrimSpace(episode.Content)
		if content == "" {
			// 规格 §6 边界标注原版「空 CURRENT_SCENE_CONTENT 无校验」。
			// TODO(对照确认): 为避免把空输入送进付费接口，这里改为直接报错；
			// 逐项对照时需确认原版是否真的会产生一次空调用。
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: 第 %d 集", ErrEmptyContent, episode.No)
			}
			result.FailedEpisodes = append(result.FailedEpisodes, episode.No)
			result.Warnings = append(result.Warnings, fmt.Sprintf("第 %d 集内容为空，已跳过", episode.No))
			continue
		}
		vars := map[string]string{
			"CURRENT_SCENE_CONTENT":   content,
			"CURRENT_SCENE_POSITION":  strconv.Itoa(index + 1),
			"TOTAL_SCENE_COUNT":       strconv.Itoa(len(episodes)),
			"ROLES_PLACEHOLDER":       roles,
			"SCENES_PLACEHOLDER":      scenes,
			"PROPS_PLACEHOLDER":       props,
			"style":                   style,
			"synopsis":                synopsis,
			"PREVIOUS_SCENE_POSITION": "",
			"PREVIOUS_SCENE_CONTENT":  "",
			"NEXT_SCENE_POSITION":     "",
			"NEXT_SCENE_CONTENT":      "",
		}
		if index > 0 {
			vars["PREVIOUS_SCENE_POSITION"] = strconv.Itoa(index)
			vars["PREVIOUS_SCENE_CONTENT"] = strings.TrimSpace(episodes[index-1].Content)
		}
		if index+1 < len(episodes) {
			vars["NEXT_SCENE_POSITION"] = strconv.Itoa(index + 2)
			vars["NEXT_SCENE_CONTENT"] = strings.TrimSpace(episodes[index+1].Content)
		}

		response, callErr := s.Generator.Generate(ctx, TextRequest{
			Model:        firstNonEmpty(req.Model, "gpt-4o-mini"),
			SystemPrompt: luoshui.Render(systemPrompt, vars),
			UserPrompt:   content,
			JSONResponse: true,
		})
		if callErr != nil {
			wrapped := wrapAIError(callErr)
			if ctx.Err() != nil {
				return result, wrapped
			}
			if firstErr == nil {
				firstErr = wrapped
			}
			result.FailedEpisodes = append(result.FailedEpisodes, episode.No)
			result.Warnings = append(result.Warnings, fmt.Sprintf("第 %d 集生成失败：%v", episode.No, wrapped))
			continue
		}

		shots, parseErr := parseStoryboardJSON(strings.TrimSpace(response.Content))
		if parseErr != nil {
			if firstErr == nil {
				firstErr = parseErr
			}
			result.FailedEpisodes = append(result.FailedEpisodes, episode.No)
			result.Warnings = append(result.Warnings, fmt.Sprintf("第 %d 集分镜解析失败：%v", episode.No, parseErr))
			continue
		}
		for i := range shots {
			shots[i].ID = newScreenplayID("sb_shot")
		}
		storyboard := Storyboard{
			ID:            newScreenplayID("sb"),
			ProjectID:     req.ProjectID,
			OwnerUsername: req.Owner,
			EpisodeNo:     episode.No,
			Engine:        engine,
			Status:        "done",
			SourceSHA256:  contentDigest(content),
			ShotCount:     len(shots),
			Shots:         shots,
		}
		saved, saveErr := s.Store.SaveStoryboard(ctx, storyboard)
		if saveErr != nil {
			return result, saveErr
		}
		result.Storyboards = append(result.Storyboards, saved)
		result.GroupCount += len(saved.Shots)
		result.Warnings = append(result.Warnings, validateStoryboardGroups(engine, episode.No, saved.Shots, known)...)
	}
	if len(result.Storyboards) == 0 && firstErr != nil {
		return result, firstErr
	}
	if len(result.FailedEpisodes) > 0 {
		// 部分剧集成功落库时不能整体报成功：把失败集号交给上层，
		// 由上层明确提示「第 N 集失败」并决定是否重跑。
		result.PartialFailure = true
		numbers := make([]string, 0, len(result.FailedEpisodes))
		for _, no := range result.FailedEpisodes {
			numbers = append(numbers, strconv.Itoa(no))
		}
		return result, fmt.Errorf("%w：第 %s 集分镜生成失败，已保存 %d 集", ErrPartialFailure, strings.Join(numbers, "、"), len(result.Storyboards))
	}
	return result, nil
}

func (s *StoryboardService) prompts() *luoshui.Registry {
	if s.Prompts != nil {
		return s.Prompts
	}
	return luoshui.MustDefault()
}

// resolveEpisodes 的优先级：Episodes 覆盖 > Content 覆盖 > 激活剧本按集切分。
func (s *StoryboardService) resolveEpisodes(ctx context.Context, req StoryboardRequest) ([]episodeInput, error) {
	if len(req.Episodes) > 0 {
		numbers := make([]int, 0, len(req.Episodes))
		for number := range req.Episodes {
			if req.EpisodeNo > 0 && number != req.EpisodeNo {
				continue
			}
			numbers = append(numbers, number)
		}
		sort.Ints(numbers)
		episodes := make([]episodeInput, 0, len(numbers))
		for _, number := range numbers {
			episodes = append(episodes, episodeInput{No: number, Content: req.Episodes[number]})
		}
		return episodes, nil
	}
	if strings.TrimSpace(req.Content) != "" {
		number := req.EpisodeNo
		if number <= 0 {
			number = 1
		}
		return []episodeInput{{No: number, Content: req.Content}}, nil
	}
	document, err := s.Store.ActiveDocument(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return nil, err
	}
	episodes := splitEpisodes(document.Content)
	if len(episodes) == 0 {
		episodes = []episodeInput{{No: 1, Content: document.Content}}
	}
	if req.EpisodeNo > 0 {
		filtered := make([]episodeInput, 0, 1)
		for _, episode := range episodes {
			if episode.No == req.EpisodeNo {
				filtered = append(filtered, episode)
			}
		}
		return filtered, nil
	}
	return episodes, nil
}

var storyboardEpisodeHeadingRe = regexp.MustCompile(`(?m)^#[ \t]*第[ \t]*([0-9]+)[ \t]*集[^\n]*$`)

// splitEpisodes 按集标题原文切分，保留正文逐字不变。
func splitEpisodes(content string) []episodeInput {
	positions := storyboardEpisodeHeadingRe.FindAllStringSubmatchIndex(content, -1)
	if len(positions) == 0 {
		return nil
	}
	episodes := make([]episodeInput, 0, len(positions))
	for i, position := range positions {
		number, err := strconv.Atoi(content[position[2]:position[3]])
		if err != nil || number <= 0 {
			continue
		}
		end := len(content)
		if i+1 < len(positions) {
			end = positions[i+1][0]
		}
		episodes = append(episodes, episodeInput{No: number, Content: strings.TrimSpace(content[position[0]:end])})
	}
	return episodes
}

// libraryPlaceholders 把资产库压成提示词声明的三个占位符。
//
// TODO(推断): 规格未给出 {ROLES_PLACEHOLDER} 等占位符的逐字序列化格式，
// 这里按「登记名：描述」逐行输出，只使用 name 作为可绑定素材名。
func (s *StoryboardService) libraryPlaceholders(ctx context.Context, req StoryboardRequest) (string, string, string, error) {
	characters, err := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, AssetCharacter)
	if err != nil {
		return "", "", "", err
	}
	scenes, err := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, AssetScene)
	if err != nil {
		return "", "", "", err
	}
	props, err := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, AssetProp)
	if err != nil {
		return "", "", "", err
	}
	return formatAssetLibrary(characters), formatAssetLibrary(scenes), formatAssetLibrary(props), nil
}

func formatAssetLibrary(assets []Asset) string {
	lines := make([]string, 0, len(assets))
	for _, asset := range assets {
		description := firstNonEmpty(asset.Description, asset.Identity)
		if description == "" {
			lines = append(lines, asset.Name)
			continue
		}
		lines = append(lines, asset.Name+"："+description)
	}
	return strings.Join(lines, "\n")
}

type knownNames struct {
	characters map[string]bool
	scenes     map[string]bool
	props      map[string]bool
}

func libraryNameSets(roles, scenes, props string) knownNames {
	return knownNames{
		characters: nameSet(roles),
		scenes:     nameSet(scenes),
		props:      nameSet(props),
	}
}

func nameSet(library string) map[string]bool {
	set := map[string]bool{}
	for _, line := range strings.Split(library, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, _, _ := strings.Cut(line, "：")
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			set[trimmed] = true
		}
	}
	return set
}

func resolveStoryboardEngine(engine StoryboardEngine) (StoryboardEngine, string, error) {
	switch StoryboardEngine(strings.TrimSpace(string(engine))) {
	case "", EngineXiaoyunqueV3:
		// 批次默认走小云雀 v3（规格 §7）。
		return EngineXiaoyunqueV3, "storyboard_xiaoyunque_v3", nil
	case EngineXiaoluoV7:
		return EngineXiaoluoV7, "storyboard_xiaoluo_v7_15s", nil
	}
	return "", "", fmt.Errorf("%w: 未知分镜引擎 %q", ErrInvalidArgument, engine)
}

// parseStoryboardJSON 解析 storyboard[] 六字段。两套引擎输出同构，
// 因此共用一份解析；字段差异由 validateStoryboardGroups 校验。
func parseStoryboardJSON(raw string) ([]StoryboardShot, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, ErrAIEmptyContent
	}
	object, ok := extractJSONObject(raw)
	if !ok {
		return nil, ErrAIInvalidJSON
	}
	var envelope map[string]any
	if err := json.Unmarshal([]byte(object), &envelope); err != nil {
		return nil, fmt.Errorf("%w %v", ErrAIJSONParseRaw, err)
	}
	groupsRaw, ok := envelope["storyboard"]
	if !ok {
		return nil, ErrAIInvalidStoryboard
	}
	groups, ok := groupsRaw.([]any)
	if !ok || len(groups) == 0 {
		return nil, ErrAIInvalidStoryboard
	}
	shots := make([]StoryboardShot, 0, len(groups))
	for index, groupRaw := range groups {
		group, ok := groupRaw.(map[string]any)
		if !ok {
			continue
		}
		videoDesc := anyString(group["video_desc"])
		if strings.TrimSpace(videoDesc) == "" {
			continue
		}
		seq := int(anyNumber(group["id"]))
		if seq <= 0 {
			seq = index + 1
		}
		shots = append(shots, StoryboardShot{
			Seq:        seq,
			SceneID:    int(anyNumber(group["scene_id"])),
			Duration:   anyNumber(group["duration_sec"]),
			Characters: anyString(group["characters"]),
			Scene:      anyString(group["scene"]),
			Props:      anyString(group["props"]),
			VideoDesc:  videoDesc,
		})
	}
	if len(shots) == 0 {
		return nil, ErrAIInvalidStoryboard
	}
	return shots, nil
}

// validateStoryboardGroups 把规格里的硬约束变成可核对的问题清单，不修改产物。
func validateStoryboardGroups(engine StoryboardEngine, episodeNo int, shots []StoryboardShot, known knownNames) []string {
	warnings := make([]string, 0)
	for _, shot := range shots {
		prefix := fmt.Sprintf("第 %d 集第 %d 组", episodeNo, shot.Seq)
		switch engine {
		case EngineXiaoluoV7:
			if shot.Duration <= 0 {
				warnings = append(warnings, prefix+" duration_sec 缺失或为 0")
			} else if shot.Duration > StoryboardMaxGroupSecondsV7 {
				warnings = append(warnings, fmt.Sprintf("%s duration_sec=%.1f 超过 15s 硬上限", prefix, shot.Duration))
			}
		case EngineXiaoyunqueV3:
			if shot.Duration <= 0 {
				warnings = append(warnings, prefix+" duration_sec 缺失或为 0")
			} else if shot.Duration < StoryboardMinGroupSecondsV3 || shot.Duration > StoryboardMaxGroupSecondsV3 {
				warnings = append(warnings, fmt.Sprintf("%s duration_sec=%.1f 不在 10–15s 区间", prefix, shot.Duration))
			}
			if firstLine, _, _ := strings.Cut(shot.VideoDesc, "\n"); !strings.HasPrefix(strings.TrimSpace(firstLine), "画风:") {
				warnings = append(warnings, prefix+" video_desc 首行不是「画风:」风格行")
			}
		}
		if strings.TrimSpace(shot.Scene) == "" {
			warnings = append(warnings, prefix+" scene 为空")
		}
		warnings = append(warnings, unknownNames(prefix, "characters", shot.Characters, known.characters)...)
		warnings = append(warnings, unknownNames(prefix, "scene", shot.Scene, known.scenes)...)
		warnings = append(warnings, unknownNames(prefix, "props", shot.Props, known.props)...)
	}
	return warnings
}

func unknownNames(prefix, field, value string, known map[string]bool) []string {
	if len(known) == 0 || strings.TrimSpace(value) == "" {
		return nil
	}
	warnings := make([]string, 0)
	for _, name := range splitAssetNames(value) {
		if known[name] {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s %s 中的「%s」不在信息库登记名中", prefix, field, name))
	}
	return warnings
}

func splitAssetNames(value string) []string {
	names := make([]string, 0, 4)
	for _, field := range strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case '，', ',', '、', '；', ';', '/', '|', '\n', '@':
			return true
		}
		return false
	}) {
		name := strings.TrimSpace(field)
		if name != "" && name != "无" {
			names = append(names, name)
		}
	}
	return names
}

func anyString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := anyString(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "，")
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func anyNumber(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case json.Number:
		number, err := typed.Float64()
		if err != nil {
			return 0
		}
		return number
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(typed, "s")), 64)
		if err != nil {
			return 0
		}
		return number
	case nil:
		return 0
	default:
		return 0
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func newPendingAction(kind, stepKey, title, detail string) PendingAction {
	return PendingAction{
		ID:              newScreenplayID("action"),
		Kind:            kind,
		StepKey:         stepKey,
		Title:           title,
		Detail:          detail,
		RequiresConfirm: true,
		CreatedAt:       time.Now().UTC(),
	}
}
