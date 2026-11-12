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
		"timeline":       timeline,
		"request":        request,
		"selectedAssets": request.SelectedAssets,
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
只能使用输入快照中真实存在的素材、片段编号和资源编号，不能编造素材或源时长，不能访问输入外部的信息。已有轨道操作必须引用真实轨道；place_asset 和 add_bgm 可指定新轨道编号，由引擎创建。
用户关闭的选项必须遵守：includeCaptions=false 禁止 add_caption；includeTransitions=false 禁止 set_transition；includeFilters=false 禁止 set_filter 和 add_effect_layer；includeBgm=false 禁止 add_bgm。
目标时长是尽量接近的软目标，受可用素材限制，不得编造源时长。素材名不代表已理解画面；当前没有画面帧、声音识别结果，不得声称看懂画面或编造对白。add_caption 只能引用 timeline.script 或现有字幕中真实存在的原文，没有原文则不要添加字幕并在 warnings 说明。
可用 place_asset 在空工程组装视频/图片：assetId 必须来自 selectedAssets，trackId 可使用现有 video 轨编号或新建合法编号；startMs 是时间线位置，sourceStartMs 是源视频起点，durationMs 是实际放置时长。视频 sourceStartMs+durationMs 不得超过真实 durationMs；图片 sourceStartMs 必须为0，durationMs 按目标合理安排。引擎会创建新轨及片段，片段编号为 plan-加操作id；不能凭空延长视频。音频 add_bgm 同样只能放置选中真实音频的有效源区间，音频轨可新建。
operations 的 type 只允许：place_asset、trim_clip、remove_range、move_clip、add_caption、add_title、set_transition、set_filter、add_effect_layer、add_bgm、set_volume_envelope、set_project_format。
每个操作必须包含唯一 id、type、reason、enabled；时间统一使用非负毫秒。只选择输入 resources 中存在的转场、滤镜和文字样式。
输出格式：{"summary":{"sourceDurationMs":0,"estimatedDurationMs":0},"operations":[],"warnings":[]}。`),
		UserPrompt:  "请根据以下项目快照和用户选项生成可审核剪辑草案：\n" + string(payload),
		Temperature: 0.1,
		MaxTokens:   8192,
	}, nil
}
