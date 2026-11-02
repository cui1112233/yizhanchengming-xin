package batchfactoryv11

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type DirectorSettings struct {
	MaxVideoDuration    int      `json:"maxVideoDuration"`
	AudioTargetSeconds  int      `json:"audioTargetSeconds,omitempty"`
	FixedSingleVideo    bool     `json:"fixedSingleVideo"`
	AspectRatio         string   `json:"aspectRatio"`
	AllowedPrefixKeys   []string `json:"allowedPrefixKeys"`
	RequireVisualPrompt bool     `json:"requireVisualPrompt"`
	RequireH3Metadata   bool     `json:"requireH3Metadata"`
}

type NamedPrompt struct {
	Name          string `json:"name"`
	FormalName    string `json:"formal_name,omitempty"`
	Facts         string `json:"facts,omitempty"`
	Appearance    string `json:"appearance,omitempty"`
	Relationships string `json:"relationships,omitempty"`
	Prompt        string `json:"prompt"`
}

type DirectorShot struct {
	StartSec      int    `json:"start_sec"`
	EndSec        int    `json:"end_sec"`
	ShotType      string `json:"shot_type"`
	Camera        string `json:"camera"`
	Rhythm        string `json:"rhythm,omitempty"`
	Audio         string `json:"audio,omitempty"`
	VisualContext string `json:"visual_context,omitempty"`
	Lighting      string `json:"lighting,omitempty"`
	Description   string `json:"description"`
}

type DirectorVideo struct {
	ID           any            `json:"id"`
	SceneID      any            `json:"scene_id"`
	DurationSec  int            `json:"duration_sec"`
	Characters   []string       `json:"characters"`
	Props        []string       `json:"props"`
	Scene        string         `json:"scene"`
	SceneMemory  string         `json:"scene_memory,omitempty"`
	PrefixKey    string         `json:"prefix_key"`
	Shots        []DirectorShot `json:"shots"`
	VideoDesc    string         `json:"video_desc"`
	VisualPrompt string         `json:"visual_prompt,omitempty"`
	// FinalPrompt is the verbatim plain-text card body returned by the SD video
	// preset. When present it replaces the compact shot-timeline rendering.
	FinalPrompt string `json:"final_prompt,omitempty"`
}

type SourceCoverage struct {
	SourceComplete     bool   `json:"source_complete"`
	SourceEndMarker    string `json:"source_end_marker"`
	HasRemainingSource bool   `json:"has_remaining_source"`
}

type DirectorResult struct {
	Characters           []NamedPrompt         `json:"characters"`
	Scenes               []NamedPrompt         `json:"scenes"`
	Props                []NamedPrompt         `json:"props"`
	Storyboard           []DirectorVideo       `json:"storyboard"`
	SourceCoverage       SourceCoverage        `json:"source_coverage"`
	SmartUnifiedStyle    string                `json:"smart_unified_style,omitempty"`
	SmartUnifiedAnalysis *SmartUnifiedAnalysis `json:"smart_unified_analysis,omitempty"`
	H3Director           *H3DirectorDocument   `json:"h3_director,omitempty"`
}

// storyboardVideoPrompt is the durable card body for one VIDEO.  The provider
// returns video_desc as a compact internal summary, but public screenplay
// generation exposes the verified shot timeline as the actual video prompt.
// Keep the summary for search/audit only; never let it replace a card's
// camera, timing and visible-action instructions.
func storyboardVideoPrompt(video DirectorVideo) string {
	if strings.TrimSpace(video.FinalPrompt) != "" {
		return strings.TrimSpace(video.FinalPrompt)
	}
	if len(video.Shots) == 0 {
		return strings.TrimSpace(video.VideoDesc)
	}
	lines := []string{"镜头画面："}
	for _, shot := range video.Shots {
		lines = append(lines, fmt.Sprintf(
			"%02d:%02d-%02d:%02d | %s｜%s | %s",
			shot.StartSec/60, shot.StartSec%60,
			shot.EndSec/60, shot.EndSec%60,
			displayShotType(shot.ShotType),
			displayCamera(shot.Camera),
			strings.TrimSpace(shot.Description),
		))
	}
	return strings.Join(lines, "\n")
}

// 导演模型返回的景别/机位是英文枚举，卡片上直接展示会中英混杂，
// 这里只做展示层的中文映射，未知值原样保留。
var shotTypeLabels = map[string]string{
	"extreme_wide_shot": "大远景",
	"wide_shot":         "远景",
	"long_shot":         "远景",
	"full_shot":         "全景",
	"medium_long_shot":  "中远景",
	"medium_shot":       "中景",
	"medium_close_up":   "中近景",
	"close_up":          "特写",
	"extreme_close_up":  "大特写",
	"extreme_closeup":   "大特写",
	"closeup":           "特写",
	"insert_shot":       "插入镜头",
	"over_the_shoulder": "过肩镜头",
}

var cameraLabels = map[string]string{
	"eye_level":         "平视",
	"high_angle":        "俯拍",
	"low_angle":         "仰拍",
	"dutch_angle":       "斜角",
	"bird_eye_view":     "鸟瞰",
	"top_down":          "俯拍",
	"over_the_shoulder": "过肩",
	"static":            "固定",
	"fixed":             "固定",
	"push_in":           "推镜",
	"pull_out":          "拉镜",
	"pan_left":          "左摇",
	"pan_right":         "右摇",
	"tilt_up":           "上摇",
	"tilt_down":         "下摇",
	"tracking":          "跟拍",
	"handheld":          "手持",
	"zoom_in":           "推近",
	"zoom_out":          "拉远",
}

func displayShotType(value string) string {
	return displayEnumLabel(value, shotTypeLabels)
}

func displayCamera(value string) string {
	return displayEnumLabel(value, cameraLabels)
}

func displayEnumLabel(value string, labels map[string]string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(trimmed, "-", "_"), " ", "_"))
	if label, ok := labels[key]; ok {
		return label
	}
	return trimmed
}

var (
	fencedDirectorJSON      = regexp.MustCompile("(?is)```(?:json)?\\s*([\\s\\S]*?)```")
	sdVideoSectionPattern   = regexp.MustCompile(`(?m)^===VIDEO\s*(\d+)\s*===\s*$`)
	sdDurationPattern       = regexp.MustCompile(`^时长[：:]\s*(\d+)\s*秒?$`)
	sdAssetReferencePattern = regexp.MustCompile(`^资产引用[：:]\s*人物\s*[=：:]\s*(.*?)\s*[；;]\s*场景\s*[=：:]\s*(.*?)\s*[；;]\s*道具\s*[=：:]\s*(.*?)\s*$`)
)

func parseSDAssetNames(value string) []string {
	parts := strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool {
		return r == '、' || r == '，' || r == ',' || r == '；' || r == ';'
	})
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" || name == "无" || name == "暂无" || name == "—" || name == "-" {
			continue
		}
		values = append(values, name)
	}
	return values
}

func parseSDAssetReference(value string) (characters []string, scene string, props []string, matched bool) {
	match := sdAssetReferencePattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 4 {
		return nil, "", nil, false
	}
	characters = parseSDAssetNames(match[1])
	sceneValues := parseSDAssetNames(match[2])
	if len(sceneValues) > 0 {
		scene = sceneValues[0]
	}
	props = parseSDAssetNames(match[3])
	return characters, scene, props, true
}

// parseSDDirectorText splits the SD preset's plain-text director output into
// one DirectorVideo per ===VIDEO NN=== section. Each section keeps its full
// body verbatim in FinalPrompt; only the 时长：X秒 line is consumed as metadata.
func parseSDDirectorText(raw string, maxVideoDuration int) (DirectorResult, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return DirectorResult{}, fmt.Errorf("导演模型返回为空")
	}
	markers := sdVideoSectionPattern.FindAllStringSubmatchIndex(text, -1)
	if len(markers) == 0 {
		return DirectorResult{}, fmt.Errorf("未找到 ===VIDEO NN=== 分段标记")
	}
	result := DirectorResult{SourceCoverage: SourceCoverage{SourceComplete: true}}
	for i, marker := range markers {
		end := len(text)
		if i+1 < len(markers) {
			end = markers[i+1][0]
		}
		body := strings.TrimSpace(text[marker[1]:end])
		lines := strings.Split(body, "\n")
		durationMatch := sdDurationPattern.FindStringSubmatch(strings.TrimSpace(lines[0]))
		if durationMatch == nil {
			return DirectorResult{}, fmt.Errorf("第 %d 个 VIDEO 段缺少“时长：X秒”行", i+1)
		}
		duration, err := strconv.Atoi(durationMatch[1])
		if err != nil || duration < 1 || duration > maxVideoDuration {
			return DirectorResult{}, fmt.Errorf("第 %d 个 VIDEO 时长必须在 1-%d 秒之间", i+1, maxVideoDuration)
		}
		characters, scene, props, hasAssetReferences := []string(nil), "", []string(nil), false
		promptLines := lines[1:]
		if len(promptLines) > 0 {
			characters, scene, props, hasAssetReferences = parseSDAssetReference(promptLines[0])
			if hasAssetReferences {
				promptLines = promptLines[1:]
			}
		}
		prompt := strings.TrimSpace(strings.Join(promptLines, "\n"))
		if prompt == "" {
			return DirectorResult{}, fmt.Errorf("第 %d 个 VIDEO 段缺少提示词正文", i+1)
		}
		result.Storyboard = append(result.Storyboard, DirectorVideo{DurationSec: duration, Characters: characters, Scene: scene, Props: props, FinalPrompt: prompt})
	}
	return result, nil
}

// ParseSDDirectorText validates the SD plain-text director protocol for
// consumers that reuse the Batch Factory director contract without its store.
func ParseSDDirectorText(raw string, maxVideoDuration int) (DirectorResult, error) {
	return parseSDDirectorText(raw, maxVideoDuration)
}

func ParseDirectorJSON(value string) (json.RawMessage, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return nil, fmt.Errorf("导演模型返回为空")
	}
	candidate := text
	if match := fencedDirectorJSON.FindStringSubmatch(text); len(match) > 1 {
		candidate = strings.TrimSpace(match[1])
	}
	if json.Valid([]byte(candidate)) {
		return json.RawMessage(candidate), nil
	}
	start := strings.Index(candidate, "{")
	end := strings.LastIndex(candidate, "}")
	if start >= 0 && end > start {
		fragment := candidate[start : end+1]
		if json.Valid([]byte(fragment)) {
			return json.RawMessage(fragment), nil
		}
	}
	return nil, fmt.Errorf("导演模型没有返回合法 JSON")
}

func cleanDirectorText(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func decodeDirectorObject(raw json.RawMessage) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("导演模型没有返回合法 JSON")
	}
	if value == nil {
		return nil, fmt.Errorf("导演模型没有返回合法 JSON")
	}
	return value, nil
}

func directorInteger(value any, field string) (int, error) {
	var number float64
	switch typed := value.(type) {
	case json.Number:
		parsed, err := strconv.ParseFloat(string(typed), 64)
		if err != nil {
			return 0, fmt.Errorf("%s 必须是整数", field)
		}
		number = parsed
	case float64:
		number = typed
	case float32:
		number = float64(typed)
	case int:
		return typed, nil
	case int64:
		return int(typed), nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, fmt.Errorf("%s 必须是整数", field)
		}
		number = parsed
	default:
		return 0, fmt.Errorf("%s 必须是整数", field)
	}
	if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
		return 0, fmt.Errorf("%s 必须是整数", field)
	}
	return int(number), nil
}

func directorRoundedSecond(value any, field string) (int, error) {
	var number float64
	switch typed := value.(type) {
	case json.Number:
		parsed, err := strconv.ParseFloat(string(typed), 64)
		if err != nil {
			return 0, fmt.Errorf("%s 必须是数字", field)
		}
		number = parsed
	case float64:
		number = typed
	case float32:
		number = float64(typed)
	case int:
		return typed, nil
	case int64:
		return int(typed), nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, fmt.Errorf("%s 必须是数字", field)
		}
		number = parsed
	default:
		return 0, fmt.Errorf("%s 必须是数字", field)
	}
	if math.IsNaN(number) || math.IsInf(number, 0) || number < 0 {
		return 0, fmt.Errorf("%s 必须是非负数字", field)
	}
	return int(math.Round(number)), nil
}

func directorArray(value any) []any {
	if array, ok := value.([]any); ok {
		return array
	}
	return []any{}
}

func directorMap(value any) (map[string]any, bool) {
	entry, ok := value.(map[string]any)
	return entry, ok
}

func firstDirectorText(entry map[string]any, fields ...string) string {
	for _, field := range fields {
		if text := cleanDirectorText(entry[field]); text != "" {
			return text
		}
	}
	return ""
}

func normalizeDirectorNamedPrompts(value any, label string) ([]NamedPrompt, error) {
	items := directorArray(value)
	seen := map[string]struct{}{}
	out := make([]NamedPrompt, 0, len(items))
	for index, item := range items {
		entry, ok := directorMap(item)
		if !ok {
			return nil, fmt.Errorf("%s[%d] 格式无效", label, index)
		}
		name := firstDirectorText(entry, "name", "名称", "角色名称", "场景名称")
		formalName := firstDirectorText(entry, "formal_name", "formalName", "正式人物名单", "正式名称", "canonical_name")
		facts := firstDirectorText(entry, "facts", "人物事实", "事实", "character_facts")
		appearance := firstDirectorText(entry, "appearance", "人物外形", "外形", "character_appearance")
		relationships := firstDirectorText(entry, "relationships", "人物关系", "关系", "character_relationships")
		prompt := firstDirectorText(entry, "prompt", "提示词", "description", "描述")
		if prompt == "" {
			prompt = appearance
		}
		if name == "" {
			return nil, fmt.Errorf("%s[%d] 缺少名称", label, index)
		}
		if prompt == "" {
			return nil, fmt.Errorf("%s[%d] 缺少提示词", label, index)
		}
		// 人物关系是结构化分析资料，不是人物的视觉锚点。旧版模型会把
		// relationships 也抄进 prompt，导致基础设定把“某人的妻子”当成
		// 外形提示词传给视频/图片模型；在统一入口把这类关系文字剥掉。
		if label == "characters" {
			prompt = mergeVisualAssetPrompt(facts, appearance, prompt, relationships)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("%s存在重复名称：%s", label, name)
		}
		seen[name] = struct{}{}
		out = append(out, NamedPrompt{Name: name, FormalName: formalName, Facts: facts, Appearance: appearance, Relationships: relationships, Prompt: prompt})
	}
	return out, nil
}

func mergeVisualAssetPrompt(facts, appearance, prompt, relationships string) string {
	parts := make([]string, 0, 2)
	if value := visualFactsWithoutRelationships(facts, relationships); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(appearance); value != "" {
		parts = append(parts, value)
	}
	if len(parts) > 0 {
		return strings.Trim(strings.Join(parts, "，"), " ，，")
	}
	return visualPromptWithoutRelationships(prompt, relationships)
}

// compiledAssetPrompt 在编译分镜时取资产文本：资产设置里保存的“合并资产提示词”
// 才是权威文本，必须优先使用；只有它为空时才回退到“事实 + 外形”拼装。
func compiledAssetPrompt(asset BookAsset) string {
	if value := visualPromptWithoutRelationships(asset.Prompt, asset.Relationships); value != "" {
		return value
	}
	return mergeVisualAssetPrompt(asset.Facts, asset.Appearance, asset.Prompt, asset.Relationships)
}

func visualPromptWithoutRelationships(prompt, relationships string) string {
	return visualFactsWithoutRelationships(prompt, relationships)
}

// visualFactsWithoutRelationships keeps facts that help render a character
// (age, gender, condition and role) but removes clauses that describe how the
// character is related to somebody else. Relationship data stays in the
// structured Relationships field and is consumed by story logic, not by the
// visual model.
func visualFactsWithoutRelationships(value, relationships string) string {
	clauses := relationshipClauses(value)
	if len(clauses) == 0 {
		return ""
	}
	kept := make([]string, 0, len(clauses))
	for _, clause := range clauses {
		if isRelationshipClause(clause, relationships) {
			continue
		}
		kept = append(kept, clause)
	}
	return strings.Trim(strings.Join(kept, "，"), " ，，；。")
}

func isRelationshipClause(clause, relationships string) bool {
	clause = strings.TrimSpace(clause)
	if clause == "" {
		return false
	}
	for _, relationship := range relationshipClauses(relationships) {
		relationship = strings.TrimSpace(relationship)
		if len([]rune(relationship)) >= 2 && (strings.Contains(clause, relationship) || strings.Contains(relationship, clause)) {
			return true
		}
	}
	// Facts from older extraction prompts often shorten the relationship text,
	// e.g. "陆淮安的妻子" versus the structured "陆淮安的合法妻子".
	for _, marker := range []string{
		"妻子", "丈夫", "老公", "老婆", "配偶", "夫妻", "妻婿", "孙媳", "孙子", "孙女",
		"祖父", "祖母", "爷爷", "奶奶", "父亲", "母亲", "爸爸", "妈妈", "儿子", "女儿",
		"兄弟", "姐妹", "长辈", "晚辈", "夫家", "主仆", "旧爱", "旧相识", "恋人", "爱人",
		"情感牵扯", "女主角的", "男主角的",
	} {
		if strings.Contains(clause, marker) {
			return true
		}
	}
	return false
}

func relationshipClauses(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	value = strings.NewReplacer("；", "\n", ";", "\n", "。", "\n", ".", "\n", "，", "\n", ",", "\n").Replace(value)
	parts := strings.Split(value, "\n")
	clauses := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			clauses = append(clauses, part)
		}
	}
	return clauses
}

func normalizeDirectorRefs(value any) []string {
	values := []any{}
	switch typed := value.(type) {
	case []any:
		values = typed
	case string:
		normalized := strings.ReplaceAll(typed, "，", ",")
		for _, part := range strings.Split(normalized, ",") {
			values = append(values, part)
		}
	default:
		if value != nil {
			values = append(values, value)
		}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, item := range values {
		name := cleanDirectorText(item)
		if entry, ok := directorMap(item); ok {
			name = cleanDirectorText(entry["name"])
		}
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func directorNameSet(values []NamedPrompt) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value.Name] = struct{}{}
	}
	return out
}

func normalizeDirectorShots(value any, durationSec, videoIndex int, requireH3Metadata bool) ([]DirectorShot, error) {
	return normalizeDirectorShotsWithOffset(value, durationSec, videoIndex, requireH3Metadata, 0)
}

// normalizeDirectorShotsWithOffset accepts a provider timeline that is
// otherwise valid but uses the whole-book clock for a later VIDEO. Stored
// VIDEO cards are segment-local by contract, so only an explicitly detected
// cumulative offset is removed. Arbitrary gaps are still rejected below.
func normalizeDirectorShotsWithOffset(value any, durationSec, videoIndex int, requireH3Metadata bool, timelineOffset int) ([]DirectorShot, error) {
	items := directorArray(value)
	if len(items) == 0 {
		return nil, fmt.Errorf("storyboard[%d] 至少需要一个镜头", videoIndex)
	}
	cursor := 0
	out := make([]DirectorShot, 0, len(items))
	for shotIndex, item := range items {
		entry, ok := directorMap(item)
		if !ok {
			return nil, fmt.Errorf("storyboard[%d].shots[%d] 格式无效", videoIndex, shotIndex)
		}
		startValue := entry["start_sec"]
		if startValue == nil {
			startValue = entry["startSec"]
		}
		endValue := entry["end_sec"]
		if endValue == nil {
			endValue = entry["endSec"]
		}
		startSec, err := directorRoundedSecond(startValue, fmt.Sprintf("storyboard[%d].shots[%d].start_sec", videoIndex, shotIndex))
		if err != nil {
			return nil, err
		}
		endSec, err := directorRoundedSecond(endValue, fmt.Sprintf("storyboard[%d].shots[%d].end_sec", videoIndex, shotIndex))
		if err != nil {
			return nil, err
		}
		if timelineOffset > 0 {
			startSec -= timelineOffset
			endSec -= timelineOffset
		}
		if startSec != cursor {
			return nil, fmt.Errorf("storyboard[%d] 镜头时间轴必须连续，期望从 %d 秒开始", videoIndex, cursor)
		}
		if endSec <= startSec {
			return nil, fmt.Errorf("storyboard[%d].shots[%d] 结束时间必须大于开始时间", videoIndex, shotIndex)
		}
		// Providers commonly return an otherwise valid shot with a Chinese field
		// name (or `shot_description`) for its visible action. Keep the persisted
		// contract canonical as `description`, while accepting these equivalent
		// representations at the boundary. We still reject an actually empty shot
		// because a VIDEO cannot be planned from it.
		description := firstDirectorText(entry, "description", "shot_description", "shotDescription", "visual_description", "visualDescription", "desc", "画面描述", "画面内容", "画面", "prompt")
		if description == "" {
			return nil, fmt.Errorf("storyboard[%d].shots[%d] 缺少画面描述", videoIndex, shotIndex)
		}
		rhythm := firstDirectorText(entry, "rhythm", "pace", "节奏")
		audio := firstDirectorText(entry, "audio", "audio_raw", "audioRaw", "sound", "dialogue", "声音", "音频", "对白")
		visualContext := firstDirectorText(entry, "visual_context", "visualContext", "scene_context", "sceneContext", "场景上下文", "场景")
		lighting := firstDirectorText(entry, "lighting", "light", "light_description", "lightDescription", "光影", "光线")
		if requireH3Metadata && rhythm == "" {
			return nil, fmt.Errorf("H3 storyboard[%d].shots[%d] 缺少 rhythm", videoIndex, shotIndex)
		}
		if requireH3Metadata && audio == "" {
			return nil, fmt.Errorf("H3 storyboard[%d].shots[%d] 缺少 audio（无声音时填写“无”）", videoIndex, shotIndex)
		}
		if requireH3Metadata && visualContext == "" {
			return nil, fmt.Errorf("H3 storyboard[%d].shots[%d] 缺少 visual_context", videoIndex, shotIndex)
		}
		if requireH3Metadata && lighting == "" {
			return nil, fmt.Errorf("H3 storyboard[%d].shots[%d] 缺少 lighting", videoIndex, shotIndex)
		}
		cursor = endSec
		out = append(out, DirectorShot{
			StartSec:      startSec,
			EndSec:        endSec,
			ShotType:      firstDirectorText(entry, "shot_type", "shotType", "景别"),
			Camera:        firstDirectorText(entry, "camera", "运镜"),
			Rhythm:        rhythm,
			Audio:         audio,
			VisualContext: visualContext,
			Lighting:      lighting,
			Description:   description,
		})
	}
	if out[len(out)-1].EndSec != durationSec {
		return nil, fmt.Errorf("storyboard[%d] 最后一个镜头必须结束在 %d 秒", videoIndex, durationSec)
	}
	return out, nil
}

func usesCumulativeStoryboardClock(value any, timelineOffset, durationSec int) bool {
	if timelineOffset <= 0 {
		return false
	}
	items := directorArray(value)
	if len(items) == 0 {
		return false
	}
	first, ok := directorMap(items[0])
	if !ok {
		return false
	}
	last, ok := directorMap(items[len(items)-1])
	if !ok {
		return false
	}
	firstStart, err := directorRoundedSecond(first["start_sec"], "storyboard.start_sec")
	if err != nil && first["start_sec"] == nil {
		firstStart, err = directorRoundedSecond(first["startSec"], "storyboard.startSec")
	}
	if err != nil {
		return false
	}
	lastEnd, err := directorRoundedSecond(last["end_sec"], "storyboard.end_sec")
	if err != nil && last["end_sec"] == nil {
		lastEnd, err = directorRoundedSecond(last["endSec"], "storyboard.endSec")
	}
	if err != nil {
		return false
	}
	return firstStart == timelineOffset && lastEnd == timelineOffset+durationSec
}

func NormalizeDirectorOutput(raw json.RawMessage, settings DirectorSettings) (DirectorResult, error) {
	if settings.MaxVideoDuration < 1 || settings.MaxVideoDuration > 60 {
		return DirectorResult{}, fmt.Errorf("maxVideoDuration 超出允许范围")
	}
	if settings.AudioTargetSeconds < 0 {
		return DirectorResult{}, fmt.Errorf("audioTargetSeconds 不能小于 0")
	}
	if settings.FixedSingleVideo && settings.AudioTargetSeconds > 0 {
		return DirectorResult{}, fmt.Errorf("固定开头只生产 VIDEO01，不能同时执行配音分镜规划")
	}
	aspectRatio := strings.TrimSpace(settings.AspectRatio)
	if aspectRatio == "" {
		aspectRatio = "9:16"
	}
	if aspectRatio != "9:16" && aspectRatio != "16:9" {
		return DirectorResult{}, fmt.Errorf("aspectRatio 仅支持 9:16 或 16:9")
	}

	root, err := decodeDirectorObject(raw)
	if err != nil {
		return DirectorResult{}, err
	}
	characters, err := normalizeDirectorNamedPrompts(root["characters"], "characters")
	if err != nil {
		return DirectorResult{}, err
	}
	scenes, err := normalizeDirectorNamedPrompts(root["scenes"], "scenes")
	if err != nil {
		return DirectorResult{}, err
	}
	props, err := normalizeDirectorNamedPrompts(root["props"], "props")
	if err != nil {
		return DirectorResult{}, err
	}
	characterNames := directorNameSet(characters)
	sceneNames := directorNameSet(scenes)
	propNames := directorNameSet(props)
	allowedPrefixes := map[string]struct{}{}
	for _, key := range settings.AllowedPrefixKeys {
		allowedPrefixes[key] = struct{}{}
	}

	sourceVideos := directorArray(root["storyboard"])
	if len(sourceVideos) == 0 {
		return DirectorResult{}, fmt.Errorf("storyboard 不能为空")
	}
	if settings.AudioTargetSeconds > 0 {
		minimumVideos := audioMinimumVideoCount(settings.AudioTargetSeconds, settings.MaxVideoDuration)
		if len(sourceVideos) < minimumVideos {
			return DirectorResult{}, fmt.Errorf("配音规划需要至少 %d 个 VIDEO 才能覆盖 %d 秒（单段上限 %d 秒），当前只有 %d 个", minimumVideos, settings.AudioTargetSeconds, settings.MaxVideoDuration, len(sourceVideos))
		}
	}
	storyboard := make([]DirectorVideo, 0, len(sourceVideos))
	bookTimelineOffset := 0
	for videoIndex, item := range sourceVideos {
		video, ok := directorMap(item)
		if !ok {
			return DirectorResult{}, fmt.Errorf("storyboard[%d] 格式无效", videoIndex)
		}
		durationValue := video["duration_sec"]
		if durationValue == nil {
			durationValue = video["durationSec"]
		}
		durationSec, err := directorInteger(durationValue, fmt.Sprintf("storyboard[%d].duration_sec", videoIndex))
		if err != nil {
			return DirectorResult{}, err
		}
		if durationSec < 1 || durationSec > settings.MaxVideoDuration {
			return DirectorResult{}, fmt.Errorf("storyboard[%d] 时长必须在 1-%d 秒之间", videoIndex, settings.MaxVideoDuration)
		}
		characterRefs := normalizeDirectorRefs(video["characters"])
		propRefs := normalizeDirectorRefs(video["props"])
		scene := cleanDirectorText(video["scene"])
		if sceneEntry, ok := directorMap(video["scene"]); ok {
			scene = cleanDirectorText(sceneEntry["name"])
		}
		for _, name := range characterRefs {
			if _, exists := characterNames[name]; !exists {
				return DirectorResult{}, fmt.Errorf("storyboard[%d] 引用了未知人物：%s", videoIndex, name)
			}
		}
		for _, name := range propRefs {
			if _, exists := propNames[name]; !exists {
				return DirectorResult{}, fmt.Errorf("storyboard[%d] 引用了未知道具：%s", videoIndex, name)
			}
		}
		if scene != "" {
			if _, exists := sceneNames[scene]; !exists {
				return DirectorResult{}, fmt.Errorf("storyboard[%d] 引用了未知场景：%s", videoIndex, scene)
			}
		}

		prefixKey := firstDirectorText(video, "prefix_key", "prefixKey")
		if len(allowedPrefixes) > 0 && prefixKey != "" {
			if _, exists := allowedPrefixes[prefixKey]; !exists {
				return DirectorResult{}, fmt.Errorf("storyboard[%d] 使用了未知前缀类型：%s", videoIndex, prefixKey)
			}
		}
		videoDesc := firstDirectorText(video, "video_desc", "videoDesc")
		if videoDesc == "" {
			return DirectorResult{}, fmt.Errorf("storyboard[%d] 缺少 video_desc", videoIndex)
		}
		visualPrompt := firstDirectorText(video, "visual_prompt", "visualPrompt", "image_prompt", "imagePrompt")
		if settings.RequireVisualPrompt && visualPrompt == "" {
			return DirectorResult{}, fmt.Errorf("storyboard[%d] 缺少 visual_prompt", videoIndex)
		}
		if !settings.RequireVisualPrompt {
			visualPrompt = ""
		}
		shots, err := normalizeDirectorShots(video["shots"], durationSec, videoIndex, settings.RequireH3Metadata)
		if err != nil && strings.Contains(err.Error(), "镜头时间轴必须连续") && usesCumulativeStoryboardClock(video["shots"], bookTimelineOffset, durationSec) {
			shots, err = normalizeDirectorShotsWithOffset(video["shots"], durationSec, videoIndex, settings.RequireH3Metadata, bookTimelineOffset)
		}
		if err != nil {
			return DirectorResult{}, err
		}
		sceneMemory := firstDirectorText(video, "scene_memory", "sceneMemory", "continuity_state", "continuityState")
		if settings.RequireH3Metadata && sceneMemory == "" {
			return DirectorResult{}, fmt.Errorf("H3 storyboard[%d] 缺少 scene_memory", videoIndex)
		}

		id := video["id"]
		if id == nil {
			id = videoIndex + 1
		}
		sceneID := video["scene_id"]
		if sceneID == nil {
			sceneID = video["sceneId"]
		}
		if sceneID == nil {
			sceneID = videoIndex + 1
		}
		storyboard = append(storyboard, DirectorVideo{
			ID:           id,
			SceneID:      sceneID,
			DurationSec:  durationSec,
			Characters:   characterRefs,
			Props:        propRefs,
			Scene:        scene,
			SceneMemory:  sceneMemory,
			VisualPrompt: visualPrompt,
			PrefixKey:    prefixKey,
			Shots:        shots,
			VideoDesc:    videoDesc,
		})
		bookTimelineOffset += durationSec
	}
	if settings.AudioTargetSeconds > 0 {
		total := 0
		for _, video := range storyboard {
			total += video.DurationSec
		}
		if total != settings.AudioTargetSeconds {
			return DirectorResult{}, fmt.Errorf("所有分镜总时长必须严格等于配音规划时长 %d 秒，当前为 %d 秒", settings.AudioTargetSeconds, total)
		}
	}

	coverage := SourceCoverage{SourceComplete: true, SourceEndMarker: "", HasRemainingSource: false}
	if rawCoverage, ok := directorMap(root["source_coverage"]); ok {
		coverage = SourceCoverage{
			SourceComplete:     rawCoverage["source_complete"] == true,
			SourceEndMarker:    cleanDirectorText(rawCoverage["source_end_marker"]),
			HasRemainingSource: rawCoverage["has_remaining_source"] == true,
		}
	}
	return DirectorResult{
		Characters:     characters,
		Scenes:         scenes,
		Props:          props,
		Storyboard:     storyboard,
		SourceCoverage: coverage,
	}, nil
}

// NormalizeAssetExtractionOutput accepts the compact asset-only result used by
// the per-book asset modal. It deliberately does not inspect storyboard data.
func NormalizeAssetExtractionOutput(raw json.RawMessage) (DirectorAssets, error) {
	root, err := decodeDirectorObject(raw)
	if err != nil {
		return DirectorAssets{}, err
	}
	characters, err := normalizeDirectorNamedPrompts(root["characters"], "characters")
	if err != nil {
		return DirectorAssets{}, err
	}
	scenes, err := normalizeDirectorNamedPrompts(root["scenes"], "scenes")
	if err != nil {
		return DirectorAssets{}, err
	}
	props, err := normalizeDirectorNamedPrompts(root["props"], "props")
	if err != nil {
		return DirectorAssets{}, err
	}
	if len(characters)+len(scenes)+len(props) == 0 {
		return DirectorAssets{}, fmt.Errorf("人物、场景、道具至少应提取一项")
	}
	return DirectorAssets{Characters: characters, Scenes: scenes, Props: props}, nil
}

// SmartUnifiedAnalysisFromAssetExtractionOutput reads the optional visual
// baseline emitted beside the normal asset arrays. Invalid or missing analysis
// is intentionally handled by the caller as an enhancement failure: assets
// remain valid and must never be discarded because this optional field failed.
func SmartUnifiedAnalysisFromAssetExtractionOutput(raw json.RawMessage) (*SmartUnifiedAnalysis, error) {
	root, err := decodeDirectorObject(raw)
	if err != nil {
		return nil, err
	}
	value, ok := root["smart_unified_analysis"]
	if !ok || value == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if string(encoded) == "null" {
		return nil, nil
	}
	return parseSmartUnifiedAnalysis(string(encoded))
}
