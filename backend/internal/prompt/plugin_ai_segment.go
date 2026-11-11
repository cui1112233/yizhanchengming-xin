package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AISegmentClip is the minimum timeline information exposed to the text model.
// File paths, model credentials and draft internals never enter the prompt.
type AISegmentClip struct {
	Index        int     `json:"index"`
	StartSeconds float64 `json:"startSeconds"`
	EndSeconds   float64 `json:"endSeconds"`
	Subtitle     string  `json:"subtitle,omitempty"`
}

// AISegmentSystemPrompt is independently written for 一战晟铭. The model may
// suggest boundaries, but the HTTP service remains responsible for enforcing
// continuity and snapping every boundary to an actual clip edge.
func AISegmentSystemPrompt() string {
	return strings.TrimSpace(`你负责为连续短剧视频建议分集边界。只能选择输入中已经存在的片段编号，不能创造时间点，也不能拆开片段。结合剧本、字幕和累计时长，在剧情阶段结束、冲突转折或钩子出现的位置结束当前集。输出必须是单个 JSON 对象，格式为 {"episodes":[{"end_clip":数字,"reason":"简短原因"}]}。不要输出 Markdown 或解释。`)
}

func AISegmentUserPrompt(script string, clips []AISegmentClip, frontCount int, frontMin, frontMax, backMin, backTarget, backMax float64) string {
	payload, _ := json.Marshal(map[string]any{
		"script": script,
		"clips":  clips,
		"durationRules": map[string]any{
			"frontCount": frontCount, "frontMinSeconds": frontMin, "frontMaxSeconds": frontMax,
			"laterMinSeconds": backMin, "laterTargetSeconds": backTarget, "laterMaxSeconds": backMax,
		},
		"coverageRules": []string{"从片段1开始", "最后一集覆盖最后一个片段", "各集连续且不重叠", "短尾优先并入上一集"},
	})
	return fmt.Sprintf("请根据以下数据建议每集的最后一个片段编号：\n%s", payload)
}
