package prompt

import (
	"encoding/json"
	"errors"
	"strings"
)

type AICutCaption struct {
	ClipID  string `json:"clipId"`
	StartMS int64  `json:"startMs"`
	EndMS   int64  `json:"endMs"`
	Text    string `json:"text"`
}

type AICutCaptionContract struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float64
	MaxTokens    int
}

func AICutCaptionCorrectionContract(script string, captions []AICutCaption) (AICutCaptionContract, error) {
	script = strings.TrimSpace(script)
	if script == "" {
		return AICutCaptionContract{}, errors.New("校对剧本不能为空")
	}
	if len(captions) == 0 {
		return AICutCaptionContract{}, errors.New("时间线上没有可校对的字幕")
	}
	payload, _ := json.Marshal(map[string]any{"script": script, "captions": captions})
	return AICutCaptionContract{
		SystemPrompt: strings.TrimSpace(`你负责依据剧本校对中文视频字幕。只修正能够由剧本明确证明的错别字、同音字、人物名和标点，不得改写剧情，不得删减信息，不能修改时间，也不能创造输入中不存在的字幕编号。只返回一个 JSON 对象，格式为 {"corrections":[{"clipId":"原字幕编号","text":"修正后的完整字幕","reason":"简短依据"}]}。没有需要修改的字幕时返回 {"corrections":[]}。不要输出 Markdown 或解释。`),
		UserPrompt:   "请校对以下 JSON 数据中的字幕：\n" + string(payload),
		Temperature:  0.1,
		MaxTokens:    4096,
	}, nil
}
