package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"qiantie/backend/internal/prompt/luoshui"
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
// 正文注册在 luoshui 语料目录（v88_synopsis），管理后台发布覆盖版本后运行时自动生效。
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
	raw, err := callModel(ctx, luoshui.V88RuntimeBody(luoshui.V88SynopsisID), "请分析以下小说文本，生成故事摘要并严格按 JSON 格式返回：\n\n"+source)
	if err != nil {
		return SynopsisResult{}, err
	}
	return NormalizeSynopsisResult(raw, now)
}
