package generation

import (
	"context"
	"fmt"
	"strings"
)

const StoryModeMetaPrompt = `你是通用小说剧情视觉化编剧与分镜规划器。输入可以来自任意题材。保持稳定输出结构，并严格遵守以下规则：
1. 不得改变主剧情、人物核心关系、事件因果与关键事实；允许在不改变事实的前提下增强失落、伤心、愤怒、紧张、冲突、压迫、惊讶等情绪表达。
2. 首镜头优先使用强动作、强冲突、强情绪或明确事件，不要总让人物站着说话。
3. 每段必须写清空间：人物在哪里，例如卧室、客厅、巷子、医院、办公室、饭店、山路等。
4. 情绪必须转换成可见动作，例如摔下杯子、猛然转身、攥紧手机、推开门、后退一步；不要只写“她很生气”。
5. 根据剧情合理使用电影语言：特写、中景、全景、推镜、拉镜、摇镜、跟拍、环绕、景深、环境光、阴影与空间调度；不得为了炫技乱加镜头。
6. 输出结构保持稳定：场景与空间 -> 人物与状态 -> 首镜头钩子 -> 关键动作与冲突 -> 镜头与光影 -> 连续性与安全说明。
7. 安全要求：不得强化性暗示，不得色情化或性化未成年人，不做不必要的身体特写，不为了冲突强行增加违规行为。能用摔东西、离开、争吵、拒绝、对峙表达时优先使用这些动作。`

const MatchAudioDirectorRules = `matchAudio=false 时保持原 Director 行为，不强制使用音频时长。
matchAudio=true 时，后端提供的 audioDurationSec 是硬性总时长：第一镜 start 必须为 0.00；后一镜 start 必须等于前一镜 end；不得有 gap 或 overlap；最后一镜 end 必须严格等于 audioDurationSec。时间单位为秒并保留两位小数。shotDurationLimitSec 是单镜头硬上限，选择 15 秒表示每镜头 <=15 秒，不是固定 15 秒切片。必须按对白、旁白、停顿、动作复杂度、情绪节奏和镜头信息量分配现有镜头时间。禁止为了凑时长新增人物、剧情、对白、旁白或无关镜头。`

const DirectorPrompt = `你是短视频导演分镜规划器。将剧本与 Hook 转成可继续用于视频提示词编译的导演 JSON 输出。保持剧情连续、空间明确、动作可视化。
` + MatchAudioDirectorRules

const H3DirectorPrompt = `你是 H3 结构化导演。必须只输出可解析 JSON，schema_version 固定为 "h3-director/v1"，并提供结构化 director_cards。保持剧情连续、空间明确、动作可视化。
` + MatchAudioDirectorRules

type PromptReader interface {
	ResolvePrompt(context.Context, string) (Prompt, error)
}

type PromptResolver struct{ store PromptReader }

func NewPromptResolver(store PromptReader) *PromptResolver { return &PromptResolver{store: store} }

func (r *PromptResolver) Resolve(ctx context.Context, key string) (Prompt, error) {
	if r == nil || r.store == nil || strings.TrimSpace(key) == "" {
		return Prompt{}, ErrInvalid
	}
	prompt, err := r.store.ResolvePrompt(ctx, key)
	if err != nil {
		return Prompt{}, err
	}
	if !prompt.Enabled || prompt.Version <= 0 || strings.TrimSpace(prompt.Content) == "" {
		return Prompt{}, fmt.Errorf("%w: prompt %s unavailable", ErrUnavailable, key)
	}
	return prompt, nil
}

func DefaultPrompts() []Prompt {
	return []Prompt{
		{Key: PromptScript, Version: 1, Enabled: true, Content: "根据小说原文生成结构清晰、可继续进入 Hook/Director 的剧本。保留主要事实、人物和因果，不虚构关键剧情。"},
		{Key: PromptScriptPlotMode, Version: 1, Enabled: true, Content: StoryModeMetaPrompt},
		{Key: PromptHook, Version: 1, Enabled: true, Content: "基于已生成剧本提炼短视频开场 Hook。强化冲突与信息密度，但不得改变主剧情。"},
		{Key: PromptDirector, Version: 2, Enabled: true, Content: DirectorPrompt},
		{Key: PromptDirectorH3, Version: 2, Enabled: true, Content: H3DirectorPrompt},
		{Key: PromptFinal, Version: 1, Enabled: true, Content: "按后端确定性顺序编译最终提示词，不在前端拼装。"},
	}
}
