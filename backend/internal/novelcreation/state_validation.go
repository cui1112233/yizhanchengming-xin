package novelcreation

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 状态一致性的确定性校验：只检查结构化矛盾，不调用模型、不改正文。
// 需要语义判断的问题（例如事实是否被正文支持）不在这里冒充已检查。
type stateValidationOutput struct {
	Passed   bool     `json:"passed"`
	Warnings []string `json:"warnings"`
}

// validateStateConsistency 对照上一版结算与新结算，报告可结构化的矛盾。
func validateStateConsistency(previous json.RawMessage, settlement truthSettlementOutput, foundation foundationPlanOutput) stateValidationOutput {
	out := stateValidationOutput{Passed: true, Warnings: []string{}}
	var old truthSettlementOutput
	if len(previous) > 0 {
		if err := json.Unmarshal(previous, &old); err != nil {
			out.Passed = false
			out.Warnings = append(out.Warnings, "上一版结算资料无法解析，无法比对状态变更")
			return out
		}
	}
	if err := uniqueSettlementIdentifiers(settlement); err != nil {
		out.Passed = false
		out.Warnings = append(out.Warnings, err.Error())
	}
	oldFacts := map[string]string{}
	for _, fact := range old.Facts {
		if id := strings.TrimSpace(fact.ID); id != "" {
			oldFacts[id] = strings.TrimSpace(fact.Statement)
		}
	}
	for _, fact := range settlement.Facts {
		id := strings.TrimSpace(fact.ID)
		if id == "" {
			out.Passed = false
			out.Warnings = append(out.Warnings, "存在没有编号的事实条目")
			continue
		}
		if statement, ok := oldFacts[id]; ok && statement != "" && strings.TrimSpace(fact.Statement) != statement {
			out.Warnings = append(out.Warnings, fmt.Sprintf("事实「%s」的表述与已保存版本不同，请确认是否为有意修订", id))
		}
	}
	oldHooks := map[string]string{}
	for _, hook := range old.Foreshadowing {
		oldHooks[strings.TrimSpace(hook.Thread)] = strings.TrimSpace(hook.Status)
	}
	for _, hook := range settlement.Foreshadowing {
		thread := strings.TrimSpace(hook.Thread)
		if thread == "" {
			out.Passed = false
			out.Warnings = append(out.Warnings, "存在没有主题的伏笔条目")
			continue
		}
		if strings.TrimSpace(hook.Status) != "open" && strings.TrimSpace(hook.Status) != "resolved" {
			out.Passed = false
			out.Warnings = append(out.Warnings, fmt.Sprintf("伏笔「%s」状态无效：%s", thread, hook.Status))
		}
		if oldHooks[thread] == "resolved" && strings.TrimSpace(hook.Status) == "open" {
			out.Warnings = append(out.Warnings, fmt.Sprintf("已解决的伏笔「%s」再次以未解决状态出现", thread))
		}
	}
	if len(foundation.Characters) > 0 {
		known := map[string]bool{}
		for _, character := range foundation.Characters {
			known[strings.TrimSpace(character.Name)] = true
		}
		for _, change := range settlement.CharacterChanges {
			name := strings.TrimSpace(change.Character)
			if name != "" && !known[name] {
				out.Warnings = append(out.Warnings, fmt.Sprintf("人物变化引用了基础设定中不存在的角色「%s」", name))
			}
		}
	}
	return out
}
