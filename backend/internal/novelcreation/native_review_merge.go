package novelcreation

import (
	"encoding/json"
	"fmt"
)

// Source: docs/reverse-engineering/luoshui-novel-native-20261119/SPEC.md.
// Native assess calls optional post-write checks; _merge_deterministic_issues
// maps error -> critical and forces passed=false without changing model score.
// The server freezes this behavior in new runs; old journals stay unchanged.
func mergePostWriteIntoAudit(input RunContext, plan chapterPlanOutput, content string, audit chapterAuditOutput, raw json.RawMessage) (chapterAuditOutput, json.RawMessage, error) {
	if input.ReviewCycleVersion < 1 || input.Agent == nil || !input.Agent.EnablePostWrite {
		return audit, raw, nil
	}
	// 使用冻结的跨章资料执行全部确定性检查；缺资料的项必须在结果里标为未检查。
	checks := ValidatePostWrite(content, input.PostWrite.postWriteInput(plan.Title))
	seen := map[string]bool{}
	for _, i := range audit.Issues {
		seen[i.ID] = true
	}
	blocked := false
	for _, issue := range checks.Issues {
		severity := issue.Severity
		if severity == "error" {
			severity = "critical"
			blocked = true
		} else if severity == "warning" {
			// CM's existing audit schema uses medium for non-blocking warnings.
			severity = "medium"
		}
		id := "post_write/" + issue.RuleID
		for n := 2; seen[id]; n++ {
			id = fmt.Sprintf("post_write/%s/%d", issue.RuleID, n)
		}
		seen[id] = true
		category, _ := json.Marshal(issue.Category)
		audit.Issues = append(audit.Issues, auditIssue{ID: id, Severity: severity, Description: issue.Message, Category: category, Source: "post_write"})
	}
	if blocked {
		passed := false
		audit.Passed = &passed
	}
	var payload map[string]json.RawMessage
	if e := json.Unmarshal(raw, &payload); e != nil {
		return audit, nil, e
	}
	payload["issues"], _ = json.Marshal(audit.Issues)
	payload["passed"], _ = json.Marshal(audit.Passed)
	payload["post_write"], _ = json.Marshal(checks)
	payload["post_write_status"], _ = json.Marshal(input.PostWrite.postWriteStatus())
	merged, e := json.Marshal(payload)
	return audit, merged, e
}
