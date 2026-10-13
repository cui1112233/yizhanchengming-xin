package batchfactoryv11

import (
	"encoding/json"
	"fmt"
	"strings"
)

var DirectorPrefixKeys = []string{
	"general_anime",
	"modern_conflict",
	"ancient_drama",
	"xuanhuan_action",
	"suspense",
	"era_drama",
}

type TextCompletionRequest struct {
	SystemPrompt        string
	UserPrompt          string
	Temperature         float64
	MaxTokens           int
	DisableJSONResponse bool
}

type PromptContract struct {
	SystemPrompt  string
	UserPrompt    string
	Temperature   float64
	MaxTokens     int
	Normalization DirectorSettings
	// SDTextProtocol marks the plain-text output protocol used by the SD video
	// preset: the model returns one ===VIDEO NN=== section per VIDEO and each
	// section body is stored verbatim as the card prompt.
	SDTextProtocol bool
}

// AIReasoningPromptConfig is saved with the batch. It is intentionally kept
// in the V11 settings patch so batch, book and VIDEO overrides still resolve
// through the same settings chain. Prompt text is generation guidance only;
// it never bypasses the director output validation.
// PresetSnapshot is resolved by the trusted Node bridge. The browser may only
// choose its ID; the current published body is frozen here with its identity so
// a director revision can always be traced to the actual rule it used.
type PresetSnapshot struct {
	ID                 string `json:"presetId"`
	Name               string `json:"presetName"`
	Slot               string `json:"presetSlot"`
	Version            int    `json:"presetVersion"`
	Key                string `json:"presetKey"`
	Body               string `json:"body"`
	ConstraintCategory string `json:"constraintCategory"`
}

type AIReasoningScope struct {
	Enabled bool     `json:"enabled"`
	Scope   string   `json:"scope"`
	BookIDs []string `json:"bookIds"`
}

type AIReasoningAssetModule struct {
	AIReasoningScope
	Extraction PresetSnapshot `json:"extraction"`
	Character  PresetSnapshot `json:"character"`
	Scene      PresetSnapshot `json:"scene"`
	Prop       PresetSnapshot `json:"prop"`
}

type AIReasoningConstraintModule struct {
	AIReasoningScope
	BaseSetup  AIReasoningBaseSetup `json:"baseSetup"`
	Selections []PresetSnapshot     `json:"selections"`
}

type AIReasoningBaseSetup struct {
	Enabled bool `json:"enabled"`
}

type AIReasoningPromptModule struct {
	AIReasoningScope
	PresetSnapshot
	// Prompt only reads legacy settings saved before typed selection support.
	// The Node bridge no longer accepts or creates it.
	Prompt string `json:"prompt"`
}

// ScriptStoryboardComposition is the trusted snapshot of the normal public
// Script workbench contract. Node resolves these published preset bodies before
// Go receives settings, so Batch Factory never has to call the public chat API.
type ScriptStoryboardComposition struct {
	Segmented         PresetSnapshot `json:"segmented"`
	Shotlist          PresetSnapshot `json:"shotlist"`
	General           PresetSnapshot `json:"general"`
	CharacterFocus    PresetSnapshot `json:"characterFocus"`
	AudioMatch        PresetSnapshot `json:"audioMatch"`
	CardProtocol      PresetSnapshot `json:"cardProtocol"`
	ConstraintWrapper PresetSnapshot `json:"constraintWrapper"`
}

type AIReasoningPromptConfig struct {
	Assets            AIReasoningAssetModule      `json:"assets"`
	Constraints       AIReasoningConstraintModule `json:"constraints"`
	Hook              AIReasoningPromptModule     `json:"hook"`
	OriginalDirector  AIReasoningPromptModule     `json:"originalDirector"`
	ViralDirector     AIReasoningPromptModule     `json:"viralDirector"`
	Video             AIReasoningPromptModule     `json:"video"`
	Visual            AIReasoningPromptModule     `json:"visual"`
	Prefix            AIReasoningPromptModule     `json:"prefix"`
	ScriptComposition ScriptStoryboardComposition `json:"scriptComposition"`
}

const (
	fixedSingleVideoDirectorMaxTokens = 14000
	standardDirectorMaxTokens         = 16000
	audioPlanningDirectorMaxTokens    = 18000
)

// directorOutputTokenBudget keeps the H3 director response proportional to
// the work it has been asked to plan. A single VIDEO still needs enough room
// for the shared asset library and structured shot timeline, while a real
// audio-following run may legitimately need the full long-form budget.
func directorOutputTokenBudget(snapshot DirectorSnapshot) int {
	if snapshot.AudioTargetSeconds > 0 {
		return audioPlanningDirectorMaxTokens
	}
	if snapshot.FixedSingleVideo {
		return fixedSingleVideoDirectorMaxTokens
	}
	return standardDirectorMaxTokens
}

func aiReasoningPromptConfig(patch SettingsPatch) AIReasoningPromptConfig {
	raw, ok := patch["aiPromptConfig"]
	if !ok {
		return AIReasoningPromptConfig{}
	}
	var config AIReasoningPromptConfig
	if json.Unmarshal(raw, &config) != nil {
		return AIReasoningPromptConfig{}
	}
	// Character and scene output are separately selectable V12 renderers. Props
	// remain on the combined extraction contract.
	config.Assets.Prop = PresetSnapshot{}
	return config
}

func (m AIReasoningScope) appliesTo(book Book) bool {
	if !m.Enabled {
		return false
	}
	if m.Scope != "custom" {
		return true
	}
	for _, id := range m.BookIDs {
		if id == book.ID {
			return true
		}
	}
	return false
}

func (m AIReasoningPromptModule) body() string {
	if body := strings.TrimSpace(m.Body); body != "" {
		return body
	}
	return strings.TrimSpace(m.Prompt)
}

func (m AIReasoningPromptModule) appliesTo(book Book) bool {
	return m.AIReasoningScope.appliesTo(book) && m.body() != ""
}

// videoAppliesTo deliberately ignores Enabled. Video prompting is the required
// director input; Enabled was an old UI-only switch and persisted false values
// must not make a selected video preset disappear from a later run.
func (m AIReasoningPromptModule) videoAppliesTo(book Book) bool {
	if m.body() == "" {
		return false
	}
	if m.Scope != "custom" {
		return true
	}
	for _, id := range m.BookIDs {
		if id == book.ID {
			return true
		}
	}
	return false
}

func (m AIReasoningAssetModule) appliesTo(book Book, selection PresetSnapshot) bool {
	return m.AIReasoningScope.appliesTo(book) && strings.TrimSpace(selection.Body) != ""
}

func storyboardDurationLabel(seconds int) string {
	if seconds == 15 {
		return "15s"
	}
	return "10s"
}

func interpolateStoryboardPreset(body string, duration int) string {
	seconds := storyboardDurationLabel(duration)
	limit := "10"
	end := "00:10"
	if duration == 15 {
		limit, end = "15", "00:15"
	}
	replacer := strings.NewReplacer(
		"{10s或15s}", seconds,
		"{X}", limit,
		"{2X}", fmt.Sprintf("%d", duration*2),
		"{duration}", seconds,
		"{结束时间}", end,
	)
	return replacer.Replace(strings.TrimSpace(body))
}

func storyboardFocusNames(values SettingsPatch) []string {
	names := rawStrings(values, "starredCharacterNames")
	seen := map[string]bool{}
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func selectedH3VideoPreset(video AIReasoningPromptModule) bool {
	presetID := strings.TrimSpace(video.ID)
	if presetID != "" {
		return presetID == "batch-video-h3-director"
	}
	return strings.TrimSpace(video.Key) == "h3-video-normal"
}

func h3OnlyRuleBody(body string) bool {
	value := strings.ToLower(strings.TrimSpace(body))
	for _, marker := range []string{"h3", "<subject", "scene_memory", "visual_context", "h3_director"} {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

func selectedNonH3VideoPreset(video AIReasoningPromptModule) bool {
	return strings.TrimSpace(video.ID) != "" && !selectedH3VideoPreset(video)
}

func selectedSDVideoPreset(video AIReasoningPromptModule) bool {
	return strings.TrimSpace(video.ID) == "batch-video-sd"
}

func appendPublicStoryboardComposition(rules []string, config AIReasoningPromptConfig, snapshot DirectorSnapshot) []string {
	if selectedNonH3VideoPreset(config.Video) {
		return rules
	}
	composition := config.ScriptComposition
	for _, preset := range []PresetSnapshot{composition.Segmented, composition.General, composition.Shotlist} {
		if body := interpolateStoryboardPreset(preset.Body, snapshot.MaxVideoDuration); body != "" {
			rules = append(rules, body)
		}
	}
	if names := storyboardFocusNames(snapshot.Effective); len(names) > 0 {
		body := strings.NewReplacer(
			"{focusCharacters}", strings.Join(names, "、"),
			"{focusCount}", fmt.Sprintf("%d", len(names)),
		).Replace(strings.TrimSpace(composition.CharacterFocus.Body))
		if body != "" {
			rules = append(rules, body)
		}
	}
	if snapshot.AudioTargetSeconds > 0 {
		duration := snapshot.AudioDurationSeconds
		if duration <= 0 {
			duration = float64(snapshot.AudioTargetSeconds)
		}
		body := strings.NewReplacer(
			"{audioDurationSec}", fmt.Sprintf("%g", duration),
			"{unitMaxSec}", fmt.Sprintf("%d", snapshot.MaxVideoDuration),
			"{duration}", storyboardDurationLabel(snapshot.MaxVideoDuration),
		).Replace(strings.TrimSpace(composition.AudioMatch.Body))
		if body != "" {
			rules = append(rules, body)
		}
	}
	return rules
}

func (m AIReasoningConstraintModule) appliesTo(book Book) bool {
	return m.AIReasoningScope.appliesTo(book)
}

func BuildHookContract(book Book, snapshot DirectorSnapshot) PromptContract {
	system := strings.TrimSpace(`你是短剧爆款开头改写导演。生成一段可人工审核的 Hook。
必须在不改变人物身份、因果关系和安全边界的前提下，把原文中的核心冲突提前并放大。
情绪升级必须表现为可见冲突和行为升级（visible conflict / behavior escalation），不能只增加“非常生气”等形容词。
保留能衔接后续原文的信息。只输出 Hook 正文，不要解释。`)
	if config := aiReasoningPromptConfig(snapshot.Effective); config.Hook.appliesTo(book) {
		system += "\n\n当前爆款开头改编规则：\n" + config.Hook.body()
	}
	return PromptContract{
		SystemPrompt: system,
		UserPrompt:   "小说标题：" + book.Title + "\n\n原文：\n" + book.SourceText,
		Temperature:  0.75,
		MaxTokens:    5000,
	}
}

// BuildWorkingFrontRewriteContract creates a reviewable replacement candidate
// for the current book's working front. It deliberately returns prose rather
// than a Hook revision: accepting it is a separate, explicit user action.
func BuildWorkingFrontRewriteContract(book Book, opening PresetSnapshot) PromptContract {
	system := strings.TrimSpace(`你是短剧爆款开头改写编辑。把当前前贴文本改写为更有冲突、悬念和可视化行动的短剧开头。
不得改变人物身份、关键因果、时间关系或安全边界；不得编造后续不存在的重要设定。
保留与原文后续内容衔接所需的信息。只输出可直接替换的正文，不要标题、说明或 Markdown。`)
	if body := strings.TrimSpace(opening.Body); body != "" {
		system += "\n\n当前衍生开篇规则：\n" + body
	}
	return PromptContract{
		SystemPrompt: system,
		UserPrompt:   "小说标题：" + book.Title + "\n\n当前前贴文本：\n" + book.SourceText,
		Temperature:  0.75,
		MaxTokens:    5000,
	}
}

// BuildOpeningVariantsContract asks the meta-prompt model for N alternative
// VIDEO01 openings. Follow-up storyboard prompts are included so every variant
// can be checked against the shot it must hand over to.
func BuildOpeningVariantsContract(book Book, first DirectorVideo, followUps []DirectorVideo, meta PresetSnapshot, variantCount, maxVideoDuration int) PromptContract {
	system := strings.TrimSpace(meta.Body)
	if system == "" {
		system = strings.TrimSpace(`你是短剧开场变体导演。为分镜一生成多个画面感更强、且能自然衔接分镜二的开场变体提示词；人物、场景、剧情走向与时长协议必须与原分镜一致。每个变体用 ===VARIANT N=== 分段，首行为 时长：X秒。`)
	}
	system += fmt.Sprintf("\n\n【本次任务】需要变体数量：%d。单段时长上限：%d 秒。", variantCount, maxVideoDuration)
	system += `

【机器输出格式（必须严格遵守）】
每个变体必须独占一个分段；标记必须单独一行，下一行必须是时长，随后必须是非空正文。不要把“时长”写在标记同一行，不要输出额外说明。示例：
===VARIANT 1===
时长：10秒
可直接替换的分镜一提示词正文`
	var user strings.Builder
	user.WriteString("小说标题：" + book.Title + "\n\n")
	// 原分镜一与后续分镜统一走 storyboardVideoPrompt：SD 直出时它返回整段
	// FinalPrompt（行为不变），结构化分镜时把镜头/动作描述渲染成“镜头画面”文本。
	user.WriteString("【原分镜一（VIDEO01）提示词】\n" + strings.TrimSpace(storyboardVideoPrompt(first)) + "\n")
	for index, video := range followUps {
		label := fmt.Sprintf("分镜%d（VIDEO%02d）提示词", index+2, index+2)
		user.WriteString("\n【" + label + "】\n" + strings.TrimSpace(storyboardVideoPrompt(video)) + "\n")
	}
	user.WriteString(fmt.Sprintf("\n需要变体数量：%d", variantCount))
	return PromptContract{SystemPrompt: system, UserPrompt: user.String(), Temperature: 0.8, MaxTokens: 8000}
}

// BuildAssetExtractionContract intentionally asks for assets only. It is used
// by the asset modal's regeneration action and must never create or replace
// storyboard/VIDEO records.
func BuildAssetExtractionContract(book Book, snapshot DirectorSnapshot) (PromptContract, error) {
	if strings.TrimSpace(book.SourceText) == "" {
		return PromptContract{}, fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	system := strings.TrimSpace(`你是短剧制作资产提取器。只从小说原文提取实际出现的人物、场景与关键道具，并为每项写可直接用于图片生成的中文视觉提示词。
不要生成分镜、镜头、VIDEO、改写正文或解释。不要杜撰原文没有出现的重要资产。
只输出一个合法 JSON 对象，格式严格为：
{"characters":[{"name":"人物名","prompt":"外形、服装、年龄感、气质等可见特征"}],"scenes":[{"name":"场景名","prompt":"空间、时段、陈设、氛围等可见特征"}],"props":[{"name":"道具名","prompt":"材质、外观、状态等可见特征"}]}
每个数组可以为空；每个元素必须同时有非空 name 与 prompt。`)
	config := aiReasoningPromptConfig(snapshot.Effective)
	smartUnifiedSelected := smartUnifiedSelectedForRevision(book, config)
	if smartUnifiedSelected {
		system += `

【智能统一与资产同次返回】
除 characters、scenes、props 外，顶层还必须返回 smart_unified_analysis。它是本书后续 VIDEO 的视觉基线，格式严格为：
"smart_unified_analysis":{"schema_version":"h3-style-system/v1","prompt":"可直接注入 VIDEO 的中文视觉基线","fields":{"final_genre":"题材","trailer_style":"影像风格","story_era":"时代背景"},"preset":{"id":"script-constraint-prefix-smart-unified","name":"智能统一","version":1}}
不得单独解释或要求再次调用；四个字段都必须在这一次 JSON 返回中。`
		for _, selection := range config.Constraints.Selections {
			if selection.ID == smartUnifiedPrefixPresetID && constraintCategory(selection) == "prefix" && strings.TrimSpace(selection.Body) != "" {
				system += "\n\n当前智能统一规则：\n" + strings.TrimSpace(selection.Body)
				break
			}
		}
	}
	h3AssetSelected := config.Assets.appliesTo(book, config.Assets.Extraction) && usesCombinedAssetRenderer(config.Assets.Extraction)
	h3CharacterSelected := config.Assets.appliesTo(book, config.Assets.Character) && usesH3CharacterRenderer(config.Assets.Character)
	h3Selected := h3AssetSelected || h3CharacterSelected
	if h3Selected {
		h3Preset := config.Assets.Character
		if h3AssetSelected {
			h3Preset = config.Assets.Extraction
		}
		if h3AssetSelected {
			// A full H3 asset preset is one complete request: it first reasons over
			// roster/facts/relationships/scenes/props, then returns every detailed
			// character appearance in that same JSON response. Do not split it back
			// into an implementation-only second appearance call.
			system = strings.TrimSpace(h3Preset.Body)
		} else if factsRule, _ := h3CharacterPromptPhases(h3Preset.Body); factsRule != "" {
			system = factsRule
		}
	}
	if !h3Selected && config.Assets.appliesTo(book, config.Assets.Extraction) {
		system += "\n\n当前人物场景道具提取规则：\n" + strings.TrimSpace(config.Assets.Extraction.Body)
	}
	if !h3Selected && config.Assets.appliesTo(book, config.Assets.Character) && !usesH3CharacterRenderer(config.Assets.Character) {
		system += "\n\n当前人物提示词输出规则：\n" + strings.TrimSpace(config.Assets.Character.Body)
	}
	if !h3Selected && config.Assets.appliesTo(book, config.Assets.Scene) && !usesH3SceneRenderer(config.Assets.Scene) {
		system += "\n\n当前场景提示词输出规则：\n" + strings.TrimSpace(config.Assets.Scene.Body)
	}
	payloadValues := map[string]any{"book_id": book.ID, "title": book.Title, "source_text": book.SourceText}
	// style.system is produced before asset extraction and frozen on the book.
	// Every selected asset rule receives the result as production context; the
	// visual-prefix switch later controls only whether this result is shown or
	// injected into the final VIDEO prompt.
	if analysis, err := parseSmartUnifiedAnalysis(rawString(snapshot.Effective, "h3StyleAnalysis", "")); err == nil && analysis != nil {
		payloadValues["style_system_analysis"] = analysis.Prompt
	}
	payload, err := json.MarshalIndent(payloadValues, "", "  ")
	if err != nil {
		return PromptContract{}, err
	}
	if h3AssetSelected {
		system += `

【批量工厂统一返回契约】
在同一个合法 JSON 对象中返回 characters、scenes、props 三个数组。characters 的每项必须同时返回 name 与 prompt；prompt 必须是该人物可直接用于图片和视频一致性的完整详细外形，而不是简短标签。scenes 与 props 的每项也必须同时有非空 name 与 prompt。不得返回第二份外形结果、不得要求后续逐人调用。`
	}
	return PromptContract{SystemPrompt: system, UserPrompt: string(payload), Temperature: 0.2, MaxTokens: 4000}, nil
}

func BuildDirectorContract(book Book, hook HookRevision, snapshot DirectorSnapshot) (PromptContract, error) {
	mode := snapshot.Mode
	if mode != "original" && mode != "viral" {
		return PromptContract{}, fmt.Errorf("%w: unsupported director mode", ErrInvalid)
	}
	if mode == "viral" && (hook.ID == "" || hook.Status != "approved") {
		return PromptContract{}, fmt.Errorf("%w: viral mode requires approved Hook", ErrConflict)
	}
	durationRule := fmt.Sprintf("每个 VIDEO 的 duration_sec 必须为 1-%d 的整数。", snapshot.MaxVideoDuration)
	if snapshot.FixedSingleVideo {
		durationRule = fmt.Sprintf("固定开头已开启：只生产 VIDEO01；其 duration_sec 必须为 1-%d 的整数。其他 storyboard 仍需按原文完整分析并保留，但不会进入视频生产。", snapshot.MaxVideoDuration)
	}
	if rawBool(snapshot.Effective, "audioPlanningEnabled", false) {
		audioSeconds := snapshot.AudioTargetSeconds
		if audioSeconds > 0 {
			minimumVideos := audioMinimumVideoCount(audioSeconds, snapshot.MaxVideoDuration)
			durationRule += fmt.Sprintf(" 分镜规划跟随配音已开启：当前书真实配音约 %.2f 秒，规划整数目标为 %d 秒；至少规划 %d 个 VIDEO，全部 VIDEO 的 duration_sec 总和必须严格等于 %d 秒。按剧情节点、对白/旁白密度和情绪节奏分配各段时长，不要机械平均；不得用重复动作、空镜或静止画面凑时长。", snapshot.AudioDurationSeconds, audioSeconds, minimumVideos, audioSeconds)
		}
	}
	config := aiReasoningPromptConfig(snapshot.Effective)
	if selectedSDVideoPreset(config.Video) {
		return buildSDDirectorContract(book, hook, snapshot, config)
	}
	requireH3Metadata := usesH3VideoRenderer(config.Video)
	// Video style prefixes are retired. Keep the legacy snapshot field readable
	// for existing books, but never let it constrain a new director revision.
	allowedPrefixKeys := append([]string(nil), DirectorPrefixKeys...)
	system := `你是 Batch Factory V11 的导演。只输出合法 JSON，不要输出 Markdown。
JSON 顶层必须包含 characters、scenes、props、storyboard、source_coverage。
storyboard 每项必须包含 duration_sec、characters、props、scene、prefix_key、shots、video_desc。
shots 中的每一项必须包含 start_sec、end_sec、shot_type、camera、description；description 是该镜头可拍、可见的画面动作描述，不能为空。不要把它改写成 shot_description、画面描述或其它字段名。
shots 必须从 0 秒开始连续、无空白无重叠，最后一个 end_sec 必须等于 duration_sec。
分镜只“点名引用”，不许描写外形：storyboard.characters 只写人物名字数组、scene 只写场景名字、props 只写道具名字，名字必须来自顶层信息库。
人物/场景长什么样，一律以资产设置为准，你不负责重新生成人物外形。
用户消息中如附 assets 资产清单（本书已提取确认的人物/场景/道具），你的职责是“引用”，不是“重造”：
- 清单里已有的资产，顶层对应项的 name 与 prompt 都必须逐字照抄清单，严禁改名、删减、扩写或用自己的话重新描写外形；prompt 不能留空。
- 只允许补充原文中新出现、清单里确实没有的资产；新增项才需要写 prompt（人物写外形/服装/年龄感/气质，场景写空间/时段/陈设/氛围，道具写材质/外观/状态）。
没有附 assets 清单时，characters/scenes/props 每项必须包含 name 与非空 prompt，prompt 是可直接用于图片和视频一致性的完整可见特征。
prefix_key 只能从指定列表选择。
原文模式按原文顺序完整覆盖；爆款模式以已批准 Hook 开场，并继续覆盖原文。
不得把普通情绪只换成形容词；冲突升级要通过动作、表情、对白和可见行为呈现。
` + durationRule + "\n画幅：" + snapshot.AspectRatio + "\n允许的 prefix_key：" + strings.Join(allowedPrefixKeys, ", ")
	rules := appendPublicStoryboardComposition(nil, config, snapshot)
	if config.Assets.appliesTo(book, config.Assets.Extraction) {
		rules = append(rules, "人物/场景/道具统一提取方案：一次提取原文实际出现的人物、场景与关键道具。\n"+strings.TrimSpace(config.Assets.Extraction.Body))
	}
	if config.Constraints.appliesTo(book) && !selectedNonH3VideoPreset(config.Video) {
		constraintBodies := []string{}
		if config.Constraints.BaseSetup.Enabled {
			constraintBodies = append(constraintBodies, "基础设定：每个 storyboard 必须带入当前情节出现的人物与场景；人物与场景只能引用本次输出的顶层信息库，不能省略场景，也不能凭空新增资产。")
		}
		for _, selection := range config.Constraints.Selections {
			if body := strings.TrimSpace(selection.Body); body != "" {
				constraintBodies = append(constraintBodies, body)
			}
		}
		if len(constraintBodies) > 0 {
			wrapper := interpolateStoryboardPreset(config.ScriptComposition.ConstraintWrapper.Body, snapshot.MaxVideoDuration)
			if wrapper != "" {
				rules = append(rules, wrapper)
			}
			rules = append(rules, "约束设置：以下规则约束本次所有输出，且必须保持可拍、可见。\n"+strings.Join(constraintBodies, "\n"))
		}
	}
	if config.Video.videoAppliesTo(book) {
		rules = append(rules, "视频提示词：以下规则约束本次 VIDEO 分镜规划、镜头、动作、运镜与 storyboard.video_desc。\n"+config.Video.body())
	}
	if requireH3Metadata {
		rules = append(rules, `H3 canonical director metadata（最高优先级）：
每个 storyboard 必须额外输出非空 scene_memory，记录承接上一 VIDEO 的人物位置、服装/道具状态、空间轴线、动作结果和未完成动作；首个 VIDEO 也要写明初始状态。
		每个 shots 项必须额外输出非空 visual_context（该镜头可见的空间、时段、陈设、天气或环境状态）、lighting（具体光源、方向、色温、对比和材质/暗部表现）、rhythm（镜头节奏，例如“克制铺陈”“骤然加速”）和 audio（该镜头的对白、旁白、环境声或“无”）。不得省略字段；没有声音时 audio 必须明确写为“无”。
		scene_memory、visual_context、lighting、rhythm、audio 是后端 H3 VIDEO 编译器的结构化输入，不是展示性说明，必须与 description 的人物、动作、时间线保持一致。`)
	}
	if !selectedNonH3VideoPreset(config.Video) {
		if cardProtocol := interpolateStoryboardPreset(config.ScriptComposition.CardProtocol.Body, snapshot.MaxVideoDuration); cardProtocol != "" {
			rules = append(rules, cardProtocol)
		}
	}
	if len(rules) > 0 {
		system += "\n\n当前批量已启用的 AI 推理规则：\n" + strings.Join(rules, "\n\n")
	}
	// The public workbench persists Markdown cards. V11 preserves the same
	// boundaries and duration semantics, but its durable director output is JSON.
	// This final instruction resolves the transport difference without changing
	// the published public prompt bodies above.
	system += "\n\nV11 JSON 存储适配（最高优先级）：公网协议中的每一张 `### 分镜N` 对应 storyboard 数组中的一个对象；保留其独立卡片边界、剧情连续性与时长规则。不得输出任何 Markdown 标题、前言、分隔线或 JSON 以外的文字，只输出前述合法 JSON。"
	payload := map[string]any{
		"book_id":     book.ID,
		"title":       book.Title,
		"mode":        mode,
		"source_text": book.SourceText,
	}
	if assets := payloadAssets(book); len(assets) > 0 {
		payload["assets"] = assets
	}
	if snapshot.AudioTargetSeconds > 0 {
		payload["audio_planning"] = map[string]any{
			"enabled":             true,
			"measured_seconds":    snapshot.AudioDurationSeconds,
			"target_seconds":      snapshot.AudioTargetSeconds,
			"unit_max_seconds":    snapshot.MaxVideoDuration,
			"minimum_video_count": audioMinimumVideoCount(snapshot.AudioTargetSeconds, snapshot.MaxVideoDuration),
		}
	}
	if mode == "viral" {
		payload["approved_hook"] = hook.Text
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return PromptContract{}, err
	}
	temperature := 0.35
	if mode == "viral" {
		temperature = 0.65
	}
	return PromptContract{
		SystemPrompt: system,
		UserPrompt:   string(encoded),
		Temperature:  temperature,
		MaxTokens:    directorOutputTokenBudget(snapshot),
		Normalization: DirectorSettings{
			MaxVideoDuration:   snapshot.MaxVideoDuration,
			AudioTargetSeconds: snapshot.AudioTargetSeconds,
			FixedSingleVideo:   snapshot.FixedSingleVideo,
			AspectRatio:        snapshot.AspectRatio,
			AllowedPrefixKeys:  allowedPrefixKeys,
			// 画面提示词必须在已有分镜卡后单独提取，导演阶段只生成
			// 公网“生成剧本”对应的分镜视频提示词与镜头结构。
			RequireVisualPrompt: false,
			RequireH3Metadata:   requireH3Metadata,
		},
	}, nil
}

// buildSDDirectorContract lets the selected SD video preset drive the director
// call exactly like public script generation: the preset body is the system
// prompt and the model answers with plain-text VIDEO cards. No JSON schema is
// imposed; every ===VIDEO NN=== section is stored verbatim as the card prompt.
func buildSDDirectorContract(book Book, hook HookRevision, snapshot DirectorSnapshot, config AIReasoningPromptConfig) (PromptContract, error) {
	system := "你是 Batch Factory 导演。按以下视频提示词规则为整本书规划 VIDEO 分镜。\n\n" + config.Video.body()
	system += fmt.Sprintf(`

输出协议（最高优先级）：只输出纯文本，不要 JSON、不要 Markdown 代码块。每个 VIDEO 一段：段首一行 ===VIDEO 01===（编号从 01 递增），第二行写 时长：X秒（X 为 1-%d 的整数秒），第三行必须写资产引用，格式固定为“资产引用：人物=人物甲、人物乙；场景=场景名；道具=道具甲、道具乙”。只列本段实际出现的资产；没有某类资产时写“无”。人物、场景、道具名称必须逐字使用用户消息 assets 清单中的 name，不能改名或换称呼。第四行开始才是该 VIDEO 按预设模板渲染的完整提示词正文（含段内执行约束、[场景 N]与总时长、[镜头 N]逐镜描述）。不要输出统一风格、统一人物、最终导出画质约束、最终导出负面提示词——这些由系统约束设置开关统一注入，AI 只负责输出正文。按原文顺序完整覆盖整本书。

分段规则：先按预设时长规则逐场景规划时长（保留三位小数），再按原文顺序把场景往 VIDEO 里装；一个 VIDEO 内各场景时长之和一旦达到 %d 秒上限，就必须结束当前段、新开下一个 ===VIDEO NN=== 段继续装，直到全部内容覆盖完。默认在场景与场景的边界切段；仅当单个场景本身就超过上限时，才允许在该场景的镜头边界处切开。段首整数秒等于该段全部镜头时长之和（镜头小数可在边界处微调凑整）；最后一段可以短于上限。严禁把超过上限的内容塞进同一段，也严禁漏掉原文内容。全部内容总时长不超过上限时，只输出一个 VIDEO 段是允许的。
预设正文中的 ${...} 占位符是旧多阶段流程的注入点，本次为单次调用，忽略占位符语法，直接使用用户消息中的原文与资产资料。
人物沿用规则：用户消息附了 assets 资产清单时，正文里涉及任何人物/场景/道具的外形，都必须按名字逐字使用清单里对应的 prompt，严禁改名、换称呼或重新发明外形；只有清单里没有、原文中新出现的人物/场景/道具，才允许按原文补充。
画幅：%s`, snapshot.MaxVideoDuration, snapshot.MaxVideoDuration, snapshot.AspectRatio)
	if snapshot.FixedSingleVideo {
		system += "\n固定开头已开启：只输出 ===VIDEO 01=== 一段。"
	}
	if snapshot.AudioTargetSeconds > 0 {
		system += fmt.Sprintf("\n分镜规划跟随配音已开启：当前书真实配音约 %.2f 秒，规划整数目标为 %d 秒；至少输出 %d 个 VIDEO 段，全部段的时长之和必须严格等于 %d 秒。按剧情节点、对白/旁白密度和情绪节奏分配各段时长，不要机械平均；不得用重复动作、空镜或静止画面凑时长。", snapshot.AudioDurationSeconds, snapshot.AudioTargetSeconds, audioMinimumVideoCount(snapshot.AudioTargetSeconds, snapshot.MaxVideoDuration), snapshot.AudioTargetSeconds)
	}
	if len(snapshot.Effective["sdStoryScriptRows"]) > 0 {
		system += "\n本次已完成 SD_STORY_SCRIPT。用户消息中的 story_script_rows 是冻结的逐行故事剧本；导演只能按 source_index 顺序为这些行填写固定镜头画面，不得新增、删除、合并、拆分、改写原文行或改变时间契约。"
	}
	payload := map[string]any{
		"book_id":     book.ID,
		"title":       book.Title,
		"mode":        snapshot.Mode,
		"source_text": book.SourceText,
	}
	if assets := payloadAssets(book); len(assets) > 0 {
		payload["assets"] = assets
	}
	if analysis, err := parseSmartUnifiedAnalysis(rawString(snapshot.Effective, "h3StyleAnalysis", "")); err == nil && analysis != nil {
		payload["style_system_analysis"] = analysis.Prompt
	}
	if rows := snapshot.Effective["sdStoryScriptRows"]; len(rows) > 0 {
		var decoded any
		if json.Unmarshal(rows, &decoded) == nil {
			payload["story_script_rows"] = decoded
		}
	}
	if lineDurations := snapshot.Effective["audioLineDurations"]; len(lineDurations) > 0 {
		var decoded any
		if json.Unmarshal(lineDurations, &decoded) == nil {
			payload["audio_line_durations"] = decoded
		}
	}
	if snapshot.AudioTargetSeconds > 0 {
		payload["audio_planning"] = map[string]any{
			"enabled":             true,
			"measured_seconds":    snapshot.AudioDurationSeconds,
			"target_seconds":      snapshot.AudioTargetSeconds,
			"unit_max_seconds":    snapshot.MaxVideoDuration,
			"minimum_video_count": audioMinimumVideoCount(snapshot.AudioTargetSeconds, snapshot.MaxVideoDuration),
		}
	}
	if snapshot.Mode == "viral" {
		payload["approved_hook"] = hook.Text
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return PromptContract{}, err
	}
	temperature := 0.35
	if snapshot.Mode == "viral" {
		temperature = 0.65
	}
	return PromptContract{
		SystemPrompt:   system,
		UserPrompt:     string(encoded),
		Temperature:    temperature,
		MaxTokens:      audioPlanningDirectorMaxTokens,
		SDTextProtocol: true,
		Normalization: DirectorSettings{
			MaxVideoDuration:   snapshot.MaxVideoDuration,
			AudioTargetSeconds: snapshot.AudioTargetSeconds,
			FixedSingleVideo:   snapshot.FixedSingleVideo,
			AspectRatio:        snapshot.AspectRatio,
		},
	}, nil
}

// payloadAssets flattens the book's extracted asset records (人物/场景/道具)
// into name+prompt lists so every video preset can anchor its top-level asset
// library on the frozen assets instead of re-inventing appearances.
func payloadAssets(book Book) map[string]any {
	groups := map[string][]map[string]string{}
	for _, asset := range book.AssetRecords {
		kind := strings.TrimSpace(asset.Kind)
		name := strings.TrimSpace(asset.Name)
		if kind == "" || name == "" {
			continue
		}
		groups[kind] = append(groups[kind], map[string]string{"name": name, "prompt": strings.TrimSpace(asset.Prompt)})
	}
	if len(groups) == 0 {
		return nil
	}
	out := make(map[string]any, len(groups))
	for kind, items := range groups {
		out[kind] = items
	}
	return out
}
