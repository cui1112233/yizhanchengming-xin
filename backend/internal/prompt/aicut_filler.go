package prompt

import (
	"encoding/json"
	"errors"
	"strings"
)

// AICutFillerContract is a clean-room contract derived from observable input
// and output behavior. It intentionally does not contain third-party prompt text.
func AICutFillerContract(captions []AICutCaption) (AICutCaptionContract, error) {
	if len(captions) == 0 {
		return AICutCaptionContract{}, errors.New("时间线上没有可处理的字幕")
	}
	payload, _ := json.Marshal(map[string]any{"captions": captions})
	return AICutCaptionContract{
		SystemPrompt: strings.TrimSpace(`你负责识别口播字幕中的纯水词行。只有当整条字幕完全由语气词、犹豫声、无信息口头禅或它们的重复组成时，才能标记删除。只要字幕含有人名、动作、数字、剧情、问答或完整含义，就必须保留；拿不准时也保留。不得创造输入中不存在的字幕编号。只返回 JSON：{"fillers":[{"clipId":"原字幕编号","reason":"简短原因"}]}。没有候选时返回 {"fillers":[]}，不要输出 Markdown 或解释。`),
		UserPrompt:   "请分析以下字幕 JSON，只返回整行纯水词候选：\n" + string(payload),
		Temperature:  0.0,
		MaxTokens:    2048,
	}, nil
}
