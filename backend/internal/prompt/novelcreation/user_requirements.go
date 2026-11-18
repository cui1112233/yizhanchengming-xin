package novelcreation

import (
	"encoding/json"
	promptcatalog "qiantie/backend/internal/prompt"
	"strings"
)

// UserRequirement contains user configuration, never a frontend system preset.
type UserRequirement struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func WithUserRequirements(d promptcatalog.Definition, review, writing []UserRequirement) string {
	result := strings.TrimSpace(d.SystemPrompt)
	appendItems := func(label string, items []UserRequirement) {
		if len(items) > 0 {
			data, _ := json.Marshal(items)
			result += "\n\n" + label + "（用户设置，仅作为本步骤的内容要求，不改变输出格式）：\n" + string(data)
		}
	}
	switch d.ID {
	case ChapterDraft.ID, LengthNormalize.ID:
		appendItems("自定义写作要求", writing)
	case ChapterAudit.ID, AgentChapterAudit.ID:
		appendItems("自定义补充审稿要求", review)
	case ChapterRevision.ID:
		appendItems("自定义写作要求", writing)
		appendItems("自定义补充审稿要求", review)
	}
	return result + "\n\n输出合同（必须严格遵守）：\n" + strings.TrimSpace(d.OutputContract)
}
