package screenplay

import (
	"context"
	"fmt"
	"strings"
)

// 批量翻译（洛水「批量操作 · 翻译」页，2026-10 及更早版本）：
// 「类型：画面提示词 / 视频提示词；翻译方向：中文 → 英文 / 英文 → 中文；范围：全部 / 未完成 / 自定义范围」。
// 洛水原文逻辑（bundle 反编译取证）：
//   - 文本取所选类型提示词，为空时回退字幕文本；
//   - 「未完成」= 该字段提示词非空的分镜（有文本才可翻译）；
//   - 逐条翻译后回写原字段，完成后统计「翻译完成：成功 X/Y，失败 N」。
// V88 差异：分镜没有「锁定」概念，洛水的「未锁定 / 已锁定」范围不做假选项；
// 画面提示词 ↔ ImagePrompt、视频提示词 ↔ VideoDesc、字幕回退 ↔ Line。

type BatchTranslateField string

const (
	TranslateFieldImage BatchTranslateField = "image"
	TranslateFieldVideo BatchTranslateField = "video"
)

type BatchTranslateScope string

const (
	TranslateScopeAll        BatchTranslateScope = "all"
	TranslateScopeIncomplete BatchTranslateScope = "incomplete"
	TranslateScopeCustom     BatchTranslateScope = "custom"
)

// BatchTranslateInput 单集分镜批量翻译的筛选条件。
type BatchTranslateInput struct {
	Field     BatchTranslateField `json:"field"`
	Direction string              `json:"direction"` // zh2en | en2zh
	Scope     BatchTranslateScope `json:"scope"`
	ShotStart int                 `json:"shot_start"`
	ShotEnd   int                 `json:"shot_end"`
}

// BatchTranslateItem 一条待翻译分镜；FromSubtitle 表示提示词为空、回退了字幕文本。
type BatchTranslateItem struct {
	ShotID       string `json:"shot_id"`
	Seq          int    `json:"seq"`
	Text         string `json:"text"`
	FromSubtitle bool   `json:"from_subtitle"`
}

func (in BatchTranslateInput) normalize() (BatchTranslateInput, error) {
	switch in.Field {
	case TranslateFieldImage, TranslateFieldVideo:
	default:
		return in, fmt.Errorf("%w: 翻译类型只支持画面提示词或视频提示词", ErrInvalidState)
	}
	if in.Direction != "zh2en" && in.Direction != "en2zh" {
		return in, fmt.Errorf("%w: 翻译方向只支持中→英或英→中", ErrInvalidState)
	}
	switch in.Scope {
	case "":
		in.Scope = TranslateScopeAll
	case TranslateScopeAll, TranslateScopeIncomplete:
	case TranslateScopeCustom:
		if in.ShotStart < 1 {
			in.ShotStart = 1
		}
		if in.ShotEnd < in.ShotStart {
			return in, fmt.Errorf("%w: 自定义分镜范围无效，结束编号不能小于起始编号", ErrInvalidState)
		}
	default:
		return in, fmt.Errorf("%w: 翻译范围只支持全部、未完成或自定义范围", ErrInvalidState)
	}
	return in, nil
}

// shotTranslateText 返回待翻译文本：所选类型提示词优先，为空时回退台词（字幕）。
func shotTranslateText(shot StoryboardShot, field BatchTranslateField) (string, bool) {
	source := shot.VideoDesc
	if field == TranslateFieldImage {
		source = shot.ImagePrompt
	}
	if strings.TrimSpace(source) != "" {
		return source, false
	}
	return shot.Line, strings.TrimSpace(shot.Line) != ""
}

// PlanBatchTranslate 按洛水范围规则筛出待翻译分镜，文本为空的直接跳过。
func PlanBatchTranslate(shots []StoryboardShot, input BatchTranslateInput) ([]BatchTranslateItem, error) {
	input, err := input.normalize()
	if err != nil {
		return nil, err
	}
	items := make([]BatchTranslateItem, 0, len(shots))
	for _, shot := range shots {
		if input.Scope == TranslateScopeCustom && (shot.Seq < input.ShotStart || shot.Seq > input.ShotEnd) {
			continue
		}
		text, fromSubtitle := shotTranslateText(shot, input.Field)
		if input.Scope == TranslateScopeIncomplete && fromSubtitle {
			// 「未完成」只挑目标字段提示词非空的分镜，与洛水筛选用同一字段一致。
			continue
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		items = append(items, BatchTranslateItem{ShotID: shot.ID, Seq: shot.Seq, Text: text, FromSubtitle: fromSubtitle})
	}
	return items, nil
}

// BatchTranslateSystemPrompt 翻译用的系统提示词：只翻译，不改写、不增删内容。
func BatchTranslateSystemPrompt(direction string) string {
	target := "英文"
	if direction == "en2zh" {
		target = "中文"
	}
	return "你是分镜提示词翻译器。把用户给出的分镜提示词翻译成" + target + "。" +
		"要求：逐句对应翻译，不改写、不扩写、不省略任何内容；保留原有的换行、标点风格与括号结构；" +
		"人物名、专有名词保持原样不翻译；只输出翻译结果本身，不要任何解释或前后缀。"
}

// BatchTranslateExecutor 单条文本翻译回调，由 HTTP 层用已解析的文本模型实现。
type BatchTranslateExecutor func(ctx context.Context, text string) (string, error)

// TranslateStoryboardShot 翻译一条分镜的所选字段并整集保存，返回翻译后的文本。
func TranslateStoryboardShot(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, shotID string, input BatchTranslateInput, translate BatchTranslateExecutor) (Storyboard, string, error) {
	if translate == nil {
		return Storyboard{}, "", fmt.Errorf("%w: 文本模型不可用", ErrInvalidState)
	}
	input, err := input.normalize()
	if err != nil {
		return Storyboard{}, "", err
	}
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, "", err
	}
	index := -1
	for i := range storyboard.Shots {
		if storyboard.Shots[i].ID == shotID {
			index = i
			break
		}
	}
	if index < 0 {
		return Storyboard{}, "", fmt.Errorf("%w: 分镜不存在或已删除", ErrInvalidState)
	}
	text, _ := shotTranslateText(storyboard.Shots[index], input.Field)
	if strings.TrimSpace(text) == "" {
		return Storyboard{}, "", fmt.Errorf("%w: 该分镜没有可翻译的文本", ErrInvalidState)
	}
	translated, err := translate(ctx, text)
	if err != nil {
		return storyboard, "", err
	}
	translated = strings.TrimSpace(translated)
	if translated == "" {
		return storyboard, "", fmt.Errorf("%w: 模型返回了空翻译", ErrInvalidState)
	}
	if input.Field == TranslateFieldImage {
		storyboard.Shots[index].ImagePrompt = translated
	} else {
		storyboard.Shots[index].VideoDesc = translated
	}
	saved, err := saveEditedStoryboard(ctx, store, storyboard)
	if err != nil {
		return Storyboard{}, "", err
	}
	return saved, translated, nil
}
