package novelcreation

import (
	"fmt"
	promptcatalog "qiantie/backend/internal/prompt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"unicode/utf8"
)

func validateAgentRequirements(a AgentSettings) error {
	total := 0
	for _, group := range []struct {
		name    string
		items   []AgentPrompt
		writing bool
	}{{"自定义审稿", a.ReviewPrompts, false}, {"自定义写作", a.CustomWritePrompts, true}} {
		fail := func(reason string) error { return &DomainError{Code: ErrInputRequired, Message: group.name + reason} }
		if len(group.items) > 50 {
			return fail("最多保存50条")
		}
		ids := map[string]bool{}
		for i, p := range group.items {
			prefix := fmt.Sprintf("第%d条：", i+1)
			if strings.TrimSpace(p.ID) == "" || len(p.ID) > 128 || ids[p.ID] {
				return fail(prefix + "条目标识为空、重复或过长")
			}
			ids[p.ID] = true
			if strings.TrimSpace(p.Title) == "" || utf8.RuneCountInString(p.Title) > 100 {
				return fail(prefix + "标题不能为空且最多100字")
			}
			if p.Enabled && strings.TrimSpace(p.Content) == "" {
				return fail(prefix + "启用前请填写要求内容")
			}
			size := utf8.RuneCountInString(p.Content)
			total += size
			if size > 8000 {
				return fail(prefix + "内容最多8000字")
			}
			if group.writing {
				if p.Category != "long_fiction" && p.Category != "short_fiction" && p.Category != "imitation" {
					return fail(prefix + "请选择长篇正文、短篇小说或仿写创作")
				}
			} else if p.Kind != "custom" {
				return fail(prefix + "当前只接受自定义审稿项")
			}
		}
	}
	if total > 24000 {
		return &DomainError{Code: ErrInputRequired, Message: "审稿与写作要求内容合计最多24000字"}
	}
	return nil
}

// Category follows the current six-mode contract. The original short/fanfic
// pipeline is not yet equivalent; this does not claim native algorithm parity.
func agentRequirementsSystemPrompt(input RunContext, definition promptcatalog.Definition) string {
	if input.Agent == nil {
		return outputContractSystemPrompt(definition)
	}
	var review, writing []novelprompt.UserRequirement
	category := "long_fiction"
	switch input.Session.Mode {
	case "short_original":
		category = "short_fiction"
	case "long_imitation", "short_imitation":
		category = "imitation"
	}
	for _, p := range input.Agent.ReviewPrompts {
		if p.Enabled && strings.TrimSpace(p.Content) != "" {
			review = append(review, novelprompt.UserRequirement{Title: p.Title, Content: p.Content})
		}
	}
	for _, p := range input.Agent.CustomWritePrompts {
		if p.Enabled && p.Category == category && strings.TrimSpace(p.Content) != "" {
			writing = append(writing, novelprompt.UserRequirement{Title: p.Title, Content: p.Content})
		}
	}
	return novelprompt.WithUserRequirements(definition, review, writing)
}
