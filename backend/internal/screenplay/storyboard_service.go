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

	"qiantie/backend/internal/prompt"
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
	IdempotencyKey             string `json:"idempotency_key"`
	RetryOf                    string `json:"retry_of"`
	ConfirmedRetry             bool   `json:"confirmed_retry"`
	generationSession          *storyboardGenerationSession
	ChapterID                  string           `json:"chapter_id"`
	PlanningChapterID          string           `json:"planning_chapter_id"`
	UsePlanning                *bool            `json:"use_planning"`
	UseOriginalContentDirectly bool             `json:"use_original_content_directly"`
	Owner                      string           `json:"owner"`
	ProjectID                  string           `json:"project_id"`
	PromptID                   string           `json:"prompt_id"`
	Engine                     StoryboardEngine `json:"engine"`
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
	GenerationAttempt    *StoryboardGenerationSummary `json:"generation_attempt,omitempty"`
	AddedAssetCount      int                          `json:"added_asset_count,omitempty"`
	Storyboards          []Storyboard                 `json:"storyboards"`
	EpisodeCount         int                          `json:"episode_count"`
	GroupCount           int                          `json:"group_count"`
	Warnings             []string                     `json:"warnings,omitempty"`
	PartialFailure       bool                         `json:"partial_failure"`
	FailedEpisodes       []int                        `json:"failed_episodes,omitempty"`
	RequiresConfirmation bool                         `json:"requires_confirmation"`
	PendingActions       []PendingAction              `json:"pending_actions,omitempty"`
}

// StoryboardService 实现能力 6（小洛 v7-15s）与能力 7（小云雀 v3）。
type StoryboardService struct {
	Authoring *AuthoringService
	Store     Store
	Generator TextGenerator
	Prompts   *luoshui.Registry
}

type episodeInput struct {
	No              int
	Content         string
	OriginalContent string `json:"original_content,omitempty"`
}

// Generate 按章节内场次调用并分别保存原回包；失败续跑复用已完成场次。
func (s *StoryboardService) Generate(ctx context.Context, req StoryboardRequest) (out StoryboardResult, runErr error) {
	if !storyboardGenerationKeyRe.MatchString(req.IdempotencyKey) {
		return out, fmt.Errorf("%w: 请提供稳定的分镜生成idempotencyKey", ErrInvalidArgument)
	}
	if s.Generator == nil {
		return out, ErrAPIKeyMissing
	}
	var session *storyboardGenerationSession
	defer func() {
		if session != nil {
			session.finish(ctx, &out, &runErr)
		}
	}()
	engine, promptID, err := resolveStoryboardEngine(req.Engine)
	if err != nil {
		return StoryboardResult{}, err
	}
	project, err := s.Store.GetProject(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return StoryboardResult{}, err
	}
	if req.PromptID != "" {
		promptID = req.PromptID
	} else if project.CreativeConfig != nil && project.CreativeConfig.StoryboardPromptID != "" {
		promptID = project.CreativeConfig.StoryboardPromptID
	}
	selected, resolveErr := ResolveProjectPrompt(project, promptID, "storyboard")
	if resolveErr != nil {
		return StoryboardResult{}, fmt.Errorf("%w: %v", ErrInvalidArgument, resolveErr)
	}
	if !selected.Supported || selected.OutputProtocol == "" {
		return out, fmt.Errorf("%w: 所选分镜模板未完成协议核验：%s", ErrInvalidArgument, selected.UnsupportedReason)
	}
	// The selected prompt defines the output protocol, including validation and
	// persistence identity. A stale client engine must never widen its limits.
	switch selected.ID {
	case "luoshui_storyboard_v9_30s_307":
		engine = EngineXiaoluoV9
	case "storyboard_xiaoluo_v7_15s":
		engine = EngineXiaoluoV7
	case "storyboard_xiaoyunque_v3":
		engine = EngineXiaoyunqueV3
	default:
		if selected.SourceID == 0 || !selected.Supported || !strings.HasPrefix(selected.OutputProtocol, "storyboard_") {
			return StoryboardResult{}, fmt.Errorf("%w: 分镜提示词 %s 不可执行：%s", ErrInvalidArgument, selected.Name, selected.UnsupportedReason)
		}
		engine = StoryboardEngine(fmt.Sprintf("prompt_%d", selected.SourceID))
	}
	if project.PermissionMode == PermissionAuto && !req.Confirmed {
		return StoryboardResult{
			RequiresConfirmation: true,
			PendingActions: []PendingAction{newPendingAction("skill", "5_storyboard",
				fmt.Sprintf("运行分镜词（%s）", engine),
				"调用分镜词提示词并写入分镜表；当前为 auto 权限，需人工确认后执行")},
		}, nil
	}

	authoring, err := s.storyboardAuthoringState(ctx, project)
	if err != nil {
		return StoryboardResult{}, err
	}
	req, err = s.storyboardChapterInput(req, authoring)
	if err != nil {
		return StoryboardResult{}, err
	}
	prefix := ""
	if ps, ok := s.Store.(StoryboardPreferencesStore); ok {
		preferences, e := ps.LoadStoryboardPreferences(ctx, req.Owner)
		if e != nil {
			return StoryboardResult{}, e
		}
		prefix = preferences.SelectedPrefix()
		if len([]rune(prefix)) > 5000 {
			return StoryboardResult{}, fmt.Errorf("%w: 分镜前缀超过5000字", ErrInvalidArgument)
		}
	}
	episodes, err := s.resolveEpisodes(ctx, req)
	if err != nil {
		return StoryboardResult{}, err
	}
	if len(episodes) == 0 {
		return StoryboardResult{}, ErrEmptyContent
	}
	if selected.SourceID > 0 {
		for i := range episodes {
			episodes[i].OriginalContent = storyboardReferenceOriginal(req, authoring, episodes[i])
		}
	}
	// Check the whole selection before reserving or generating any episode.
	// A heading and the document preamble are not episode body text.
	for _, episode := range episodes {
		body := episode.Content
		for _, position := range storyboardEpisodeHeadingRe.FindAllStringSubmatchIndex(body, -1) {
			if _, ok := storyboardEpisodeNumber(body[position[2]:position[3]]); ok {
				body = body[position[1]:]
				break
			}
		}
		if strings.TrimSpace(body) == "" {
			return StoryboardResult{}, fmt.Errorf("%w: 第%d集正文为空，请补充后再生成", ErrEmptyContent, episode.No)
		}
	}
	// Validate every selected source unit before reserving or calling a provider.
	sceneInputs := map[int][]storyboardSceneInput{}
	sceneCounts := map[int]int{}
	for _, ep := range episodes {
		units, e := storyboardScenesForEpisode(req, authoring, ep)
		if e != nil {
			return out, fmt.Errorf("第%d集：%w", ep.No, e)
		}
		sceneInputs[ep.No] = units
		sceneCounts[ep.No] = len(units)
	}
	planningByEpisode := map[int]string{}
	for _, episode := range episodes {
		planning, e := storyboardPlanningText(req, authoring, episode.No, project)
		if e != nil {
			return StoryboardResult{}, e
		}
		planningByEpisode[episode.No] = planning
	}

	roles, scenes, props, err := s.libraryPlaceholders(ctx, req)
	if err != nil {
		return StoryboardResult{}, err
	}
	style := firstNonEmpty(project.StyleDesc, req.Style, project.ArtStyle)
	synopsis := firstNonEmpty(req.Synopsis, project.Synopsis)
	store, ok := s.Store.(StoryboardGenerationStore)
	if !ok {
		return out, fmt.Errorf("%w: 分镜生成持久状态未配置", ErrInvalidState)
	}
	assets, e := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, "")
	if e != nil {
		return out, e
	}
	numbers := []int{}
	for _, ep := range episodes {
		numbers = append(numbers, ep.No)
	}
	external := map[int][]Storyboard{}
	currentBoards, e := s.Store.ListStoryboards(ctx, req.Owner, req.ProjectID, engine)
	if e != nil {
		return out, e
	}
	if EffectiveCreativeConfig(project).StoryboardUseContextHistory {
		external = generationExternalHistory(currentBoards, numbers)
	}
	base := map[string]any{"owner": req.Owner, "project": req.ProjectID, "engine": engine, "model": firstNonEmpty(req.Model, project.ModelID, "gpt-4o-mini"), "prompt_id": selected.ID, "prompt_digest": contentDigest(selected.Body), "config": EffectiveCreativeConfig(project), "prefix": prefix, "episodes": episodes, "planning": planningByEpisode, "style": style, "synopsis": synopsis, "chapter_id": req.ChapterID, "planning_chapter_id": req.PlanningChapterID, "use_original": req.UseOriginalContentDirectly, "use_planning": req.UsePlanning, "external_history": external, "scene_contract": "chapter-scenes-v1"}
	if selected.SourceID > 0 {
		base["scene_contract"] = "zip-chapter-scene-message-v1"
	}
	var replay *StoryboardGenerationAttempt
	base["scene_counts"] = sceneCounts
	base["scene_inputs"] = sceneInputs
	base["output_protocol"] = selected.OutputProtocol
	base["parser_version"] = selected.ParserVersion
	session, replay, e = beginStoryboardGeneration(ctx, store, req, engine, numbers, base, assets, external, currentBoards)
	if e != nil {
		if replay != nil {
			result, _ := replay.replay()
			return result, e
		}
		return out, e
	}
	if replay != nil {
		return replay.replay()
	}
	req.generationSession = session
	roles, scenes, props = session.attempt.Roles, session.attempt.Scenes, session.attempt.Props

	result := StoryboardResult{EpisodeCount: len(episodes)}
	known := libraryNameSets(roles, scenes, props)
	var firstErr error
	for _, episode := range episodes {
		if completed := session.attempt.Episodes[episode.No]; completed.Status == "succeeded" && completed.Storyboard != nil {
			result.Storyboards = append(result.Storyboards, *completed.Storyboard)
			result.GroupCount += len(completed.Storyboard.Shots)
			continue
		}
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
		shots, sceneErr := s.generateSceneShots(ctx, req, episode, sceneInputs[episode.No], selected, roles, scenes, props, style, synopsis, prefix, planningByEpisode[episode.No], firstNonEmpty(req.Model, project.ModelID, "gpt-4o-mini"), EffectiveCreativeConfig(project).StoryboardUseContextHistory, EffectiveCreativeConfig(project).AspectRatio)
		if sceneErr != nil {
			if ctx.Err() != nil {
				return result, wrapAIError(ctx.Err())
			}
			if firstErr == nil {
				firstErr = sceneErr
			}
			result.FailedEpisodes = append(result.FailedEpisodes, episode.No)
			result.Warnings = append(result.Warnings, sceneErr.Error())
			if session.attempt.Episodes[episode.No].Status == "unknown" {
				return result, sceneErr
			}
			continue
		}
		if EffectiveCreativeConfig(project).StoryboardAutoSyncAssets {
			syncReq := req
			syncReq.EpisodeNo = episode.No
			added, e := s.syncStoryboardMissingAssets(ctx, project, syncReq, shots)
			if checkpointErr := session.recordAssets(ctx, episode.No, added); checkpointErr != nil {
				return result, checkpointErr
			}
			result.AddedAssetCount += len(added)
			if e != nil {
				if checkpointErr := session.episodeError(episode.No, e); checkpointErr != nil {
					return result, checkpointErr
				}
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				if firstErr == nil {
					firstErr = e
				}
				result.FailedEpisodes = append(result.FailedEpisodes, episode.No)
				result.Warnings = append(result.Warnings, fmt.Sprintf("第%d集预设补充失败：%v", episode.No, e))
				continue
			}
			for _, a := range added {
				switch a.Type {
				case AssetCharacter:
					known.characters[a.Name] = true
				case AssetScene:
					known.scenes[a.Name] = true
				case AssetProp:
					known.props[a.Name] = true
				}
			}
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
		warnings := validateStoryboardGroups(engine, episode.No, storyboard.Shots, known)
		committed, saveErr := store.CommitStoryboardGeneration(ctx, req.Owner, req.ProjectID, req.IdempotencyKey, episode.No, storyboard, warnings)
		if saveErr != nil {
			return result, saveErr
		}
		session.attempt = committed
		saved := *committed.Episodes[episode.No].Storyboard
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
		if err := validateUniqueStoryboardEpisodes(splitEpisodes(req.Content)); err != nil {
			return nil, err
		}
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
	if err := validateUniqueStoryboardEpisodes(episodes); err != nil {
		return nil, err
	}
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

// A title needs a line of its own and a separator after 集; chapter headings
// and narrative mentions such as 第2集会再见 remain body text.
var storyboardEpisodeHeadingRe = regexp.MustCompile(`(?m)^[ \t]*(?:#{1,6}[ \t]*)?第[ \t]*([0-9〇零一二三四五六七八九十百千万两]+)[ \t]*集([ \t：:·—\-][^\r\n]*)?\r?$`)

func storyboardEpisodeNumber(text string) (int, bool) {
	if number, err := strconv.ParseInt(text, 10, 64); err == nil {
		return int(number), number > 0 && number <= 9007199254740991
	}
	digits := map[rune]int64{'〇': 0, '零': 0, '一': 1, '二': 2, '两': 2, '三': 3, '四': 4, '五': 5, '六': 6, '七': 7, '八': 8, '九': 9}
	units := map[rune]int64{'十': 10, '百': 100, '千': 1000, '万': 10000}
	var total, section, digit, positional int64
	hasUnit := false
	for _, char := range text {
		if value, ok := digits[char]; ok {
			digit = value
			positional = positional*10 + value
			if positional > 9007199254740991 {
				return 0, false
			}
			continue
		}
		unit, ok := units[char]
		if !ok {
			return 0, false
		}
		hasUnit = true
		if unit == 10000 {
			total += (section + digit) * unit
			section, digit = 0, 0
		} else {
			if digit == 0 {
				digit = 1
			}
			section += digit * unit
			digit = 0
		}
	}
	if !hasUnit {
		total = positional
	} else {
		total += section + digit
	}
	return int(total), total > 0 && total <= 9007199254740991
}

func validateUniqueStoryboardEpisodes(episodes []episodeInput) error {
	seen := make(map[int]bool, len(episodes))
	for _, episode := range episodes {
		if seen[episode.No] {
			return fmt.Errorf("%w: 第%d集编号重复，请修改后再生成", ErrInvalidArgument, episode.No)
		}
		seen[episode.No] = true
	}
	return nil
}

// splitEpisodes 按集标题原文切分，保留正文逐字不变。
func splitEpisodes(content string) []episodeInput {
	positions := storyboardEpisodeHeadingRe.FindAllStringSubmatchIndex(content, -1)
	valid := make([][]int, 0, len(positions))
	for _, position := range positions {
		if _, ok := storyboardEpisodeNumber(content[position[2]:position[3]]); ok {
			valid = append(valid, position)
		}
	}
	episodes := make([]episodeInput, 0, len(valid))
	for i, position := range valid {
		number, _ := storyboardEpisodeNumber(content[position[2]:position[3]])
		start, end := position[0], len(content)
		if i == 0 {
			start = 0
		}
		if i+1 < len(valid) {
			end = valid[i+1][0]
		}
		episodes = append(episodes, episodeInput{No: number, Content: strings.TrimSpace(content[start:end])})
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
	if strings.HasPrefix(string(engine), "prompt_") {
		p, e := prompt.ResolveCreativePrompt(strings.TrimPrefix(string(engine), "prompt_"), "storyboard")
		if e == nil && p.Supported {
			return engine, p.ID, nil
		}
		return "", "", fmt.Errorf("%w: 不可用分镜模板%s", ErrInvalidArgument, engine)
	}
	switch StoryboardEngine(strings.TrimSpace(string(engine))) {
	case "", EngineXiaoyunqueV3:
		// 批次默认走小云雀 v3（规格 §7）。
		return EngineXiaoyunqueV3, "storyboard_xiaoyunque_v3", nil
	case EngineXiaoluoV9:
		return EngineXiaoluoV9, "luoshui_storyboard_v9_30s_307", nil
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
			return nil, fmt.Errorf("%w: 第%d组必须是对象", ErrAIInvalidStoryboard, index+1)
		}
		videoDesc := anyString(group["video_desc"])
		imagePrompt := anyString(group["image_prompt"])
		if imagePrompt == "" {
			imagePrompt = anyString(group["image_desc"])
		}
		if strings.TrimSpace(videoDesc) == "" && strings.TrimSpace(imagePrompt) == "" {
			return nil, fmt.Errorf("%w: 第%d组视频词和图片词均为空", ErrAIInvalidStoryboard, index+1)
		}
		protocolFields, marshalErr := json.Marshal(group)
		if marshalErr != nil {
			return nil, marshalErr
		}
		var sourceElements json.RawMessage
		if v, ok := group["source_elements"]; ok {
			sourceElements, _ = json.Marshal(v)
		}
		seq := int(anyNumber(group["id"]))
		if seq <= 0 {
			seq = index + 1
		}
		shots = append(shots, StoryboardShot{
			Seq:            seq,
			SceneID:        int(anyNumber(group["scene_id"])),
			Duration:       anyNumber(group["duration_sec"]),
			Characters:     anyString(group["characters"]),
			Scene:          anyString(group["scene"]),
			Props:          anyString(group["props"]),
			VideoDesc:      videoDesc,
			ImagePrompt:    imagePrompt,
			VoiceActor:     anyString(group["voice_actor"]),
			Emotion:        anyString(group["emotion"]),
			Line:           anyString(group["line"]),
			SourceElements: sourceElements,
			ProtocolFields: protocolFields,
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
		case EngineXiaoluoV9:
			if shot.Duration <= 0 || shot.Duration > 30 {
				warnings = append(warnings, fmt.Sprintf("%s duration_sec=%.1f 不在 0–30s 区间", prefix, shot.Duration))
			}
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
