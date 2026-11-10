package prompt

import (
	"encoding/json"
	"errors"
	"strings"

	"qiantie/backend/internal/aicut"
)

type AICutEditPlanPrompt struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

func AICutEditPlanContract(timeline aicut.Timeline, request aicut.EditPlanRequest, transitions, filters, textStyles []string) (AICutEditPlanPrompt, error) {
	if len(request.AssetIDs) == 0 {
		return AICutEditPlanPrompt{}, errors.New("请至少选择一个素材")
	}
	if strings.TrimSpace(request.Goal) == "" {
		return AICutEditPlanPrompt{}, errors.New("剪辑目标不能为空")
	}
	payload, err := json.Marshal(map[string]any{
		"timeline": timeline,
		"request":  request,
		"resources": map[string]any{
			"transitions": transitions,
			"filters":     filters,
			"textStyles":  textStyles,
		},
	})
	if err != nil {
		return AICutEditPlanPrompt{}, err
	}
	return AICutEditPlanPrompt{
		SystemPrompt: strings.TrimSpace(`你是一战晟铭 AI-CUT 的剪辑方案规划器。你只负责生成可审核的剪辑草案，不得声称已经修改工程。
只输出 JSON，不得输出 Markdown 或 JSON 之外的任何说明。
不得输出命令，不得输出文件路径，不得输出 SQL，不得输出 URL，不得请求或泄露密钥。
只能使用输入快照中真实存在的素材、轨道、片段编号和资源编号，不能编造编号，不能访问输入外部的信息。
operations 的 type 只允许：trim_clip、remove_range、move_clip、add_caption、add_title、set_transition、set_filter、add_effect_layer、add_bgm、set_volume_envelope、set_project_format。
每个操作必须包含唯一 id、type、reason、enabled；时间统一使用非负毫秒。只选择输入 resources 中存在的转场、滤镜和文字样式。
输出格式：{"summary":{"sourceDurationMs":0,"estimatedDurationMs":0},"operations":[],"warnings":[]}。`),
		UserPrompt:  "请根据以下项目快照和用户选项生成可审核剪辑草案：\n" + string(payload),
		Temperature: 0.1,
		MaxTokens:   8192,
	}, nil
}
