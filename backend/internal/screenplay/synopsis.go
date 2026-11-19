package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// SynopsisResult 是「故事摘要 Agent」的结构化产物，对应洛水 synopsis_result.json：
// {title, synopsis, audience, genre, tags}，前端剧本页顶部据此渲染标题/题材标签/故事概要。
type SynopsisResult struct {
	Title       string    `json:"title"`
	Synopsis    string    `json:"synopsis"`
	Audience    string    `json:"audience"`
	Genre       string    `json:"genre"`
	Tags        []string  `json:"tags"`
	GeneratedAt time.Time `json:"generated_at,omitempty"`
}

// 洛水 screenplay.pyd 的 SYNOPSIS_AGENT_PROMPT（提取原文规格，中文原样保留要点）：
// 摘要 300-500 字，涵盖起因→发展→高潮→结局，保留核心人物姓名与关键事件；
// 标题 2-10 字；audience 男频/女频二选一；genre 枚举；tags 3-6 个、每个 2-4 字。
const synopsisSystemPrompt = `你是一个专业的故事摘要与项目信息提取助手。你的任务是分析小说文本，提炼故事核心脉络，生成精炼的故事摘要，并识别作品的受众类型、题材分类和关键标签，以 JSON 格式返回。

【核心原则】
摘要必须做到"精炼传神、脉络清晰"——用最少的文字传递最完整的故事信息，让读者一读便知故事全貌。

【摘要要求】
- 字数控制在 300-500字，不得少于300字，不得超过500字
- 必须涵盖故事的完整脉络：起因→发展→高潮→结局（或当前进展）
- 必须保留核心人物的姓名和关键事件的具体描述
- 必须体现故事的核心冲突和情感主线
- 语言流畅自然，具有文学性，能让读者快速了解故事全貌
- 不要逐句复述原文，而是提炼概括，抓住关键转折点
- 保持客观叙述视角，不加入个人评价或推荐语
- 如果原文有明确的世界观设定（仙侠/玄幻/都市/历史等），摘要开头应简要交代背景

【标题要求】
- 根据故事内容提炼一个简洁有力的标题
- 标题应能概括故事主题或核心冲突
- 控制在 2-10个字 以内
- 标题风格应贴合故事类型（仙侠→古韵诗意；悬疑→紧张悬念；言情→情感共鸣）

【受众与分类标签要求】
你必须从文本内容中精准识别以下信息：
1. audience（受众频道）— 必须二选一："男频" 或 "女频"
2. genre（题材分类）— 选择最匹配的一个：现代言情、古代言情、仙侠玄幻、都市异能、悬疑推理、科幻未来、历史军事、游戏竞技、恐怖灵异、二次元、古代宫廷、末日生存、武侠江湖、其他
3. tags（关键词标签）— 3-6个，每个2-4字，从题材设定/情感基调/核心元素/风格氛围维度提取，避免过于笼统

【输出格式要求】
只返回合法 JSON，不得输出任何解释性文字。
{"title":"故事标题（2-10字）","synopsis":"300-500字的故事摘要，涵盖起因、发展、高潮、结局","audience":"男频 或 女频","genre":"题材分类","tags":["标签1","标签2","标签3"]}`

var synopsisGenres = map[string]bool{
	"现代言情": true, "古代言情": true, "仙侠玄幻": true, "都市异能": true, "悬疑推理": true,
	"科幻未来": true, "历史军事": true, "游戏竞技": true, "恐怖灵异": true, "二次元": true,
	"古代宫廷": true, "末日生存": true, "武侠江湖": true, "其他": true,
}

type synopsisDraft struct {
	Title    string   `json:"title"`
	Synopsis string   `json:"synopsis"`
	Audience string   `json:"audience"`
	Genre    string   `json:"genre"`
	Tags     []string `json:"tags"`
}

// NormalizeSynopsisResult 解析模型输出并按洛水规则归一化：
// audience 只保留 男频/女频（非法回退女频）；genre 白名单（非法回退其他）；
// tags 保留 3-6 个、每个 2-4 字、去重；synopsis 为空则报错（长度仅提示不强制，容错模型）。
func NormalizeSynopsisResult(raw string, now time.Time) (SynopsisResult, error) {
	obj, ok := extractJSONObject(raw)
	if !ok {
		return SynopsisResult{}, fmt.Errorf("%w: 模型输出没有可解析的故事摘要 JSON", ErrInvalidState)
	}
	var draft synopsisDraft
	if err := json.Unmarshal([]byte(obj), &draft); err != nil {
		return SynopsisResult{}, fmt.Errorf("%w: 故事摘要 JSON 解析失败：%v", ErrInvalidState, err)
	}
	synopsis := strings.TrimSpace(draft.Synopsis)
	if synopsis == "" {
		return SynopsisResult{}, fmt.Errorf("%w: 模型未返回故事摘要内容", ErrInvalidState)
	}
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		title = "未命名故事"
	}
	audience := strings.TrimSpace(draft.Audience)
	if audience != "男频" && audience != "女频" {
		audience = "女频"
	}
	genre := strings.TrimSpace(draft.Genre)
	if !synopsisGenres[genre] {
		genre = "其他"
	}
	tags := make([]string, 0, 6)
	seen := map[string]bool{}
	for _, item := range draft.Tags {
		tag := strings.TrimSpace(item)
		n := utf8.RuneCountInString(tag)
		if tag == "" || n < 2 || n > 4 || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
		if len(tags) == 6 {
			break
		}
	}
	return SynopsisResult{
		Title: title, Synopsis: synopsis, Audience: audience, Genre: genre, Tags: tags,
		GeneratedAt: now.UTC(),
	}, nil
}

// GenerateSynopsis 从原文生成故事摘要。callModel 为外部注入的文本模型调用（走 V88 自己的通道）。
func GenerateSynopsis(ctx context.Context, sourceText string, callModel func(ctx context.Context, systemPrompt, userPrompt string) (string, error), now time.Time) (SynopsisResult, error) {
	source := strings.TrimSpace(sourceText)
	if source == "" {
		return SynopsisResult{}, fmt.Errorf("%w: 还没有原文内容，无法生成故事摘要", ErrInvalidArgument)
	}
	if callModel == nil {
		return SynopsisResult{}, fmt.Errorf("%w: 未配置文本模型", ErrInvalidArgument)
	}
	// 洛水：AI 推理前先取原文；过长时截断到可读范围，避免超出模型上下文。
	const maxRunes = 60000
	if utf8.RuneCountInString(source) > maxRunes {
		source = string([]rune(source)[:maxRunes])
	}
	raw, err := callModel(ctx, synopsisSystemPrompt, "请分析以下小说文本，生成故事摘要并严格按 JSON 格式返回：\n\n"+source)
	if err != nil {
		return SynopsisResult{}, err
	}
	return NormalizeSynopsisResult(raw, now)
}
