package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"qiantie/backend/internal/batchfactoryv11"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
)

// Native evidence and CM compatibility boundaries: PATCH-SPEC.md.
type nativePatch struct{ Target, Replacement string }
type nativePatchResult struct {
	Route        *nativeRevisionRoute `json:"route,omitempty"`
	Content      string               `json:"content"`
	Applied      int                  `json:"applied_patch_count"`
	Skipped      int                  `json:"skipped_patch_count"`
	Touched      int                  `json:"touched_chars"`
	Mode         string               `json:"mode"`
	UpdatedState string               `json:"updated_state"`
}

var nativePatchPattern = regexp.MustCompile(`(?s)--- PATCH(?:\s+\d+)? ---\s*TARGET_TEXT:\s*(.*?)\s*REPLACEMENT_TEXT:\s*(.*?)\s*--- END PATCH ---`)
var nativeSectionPattern = regexp.MustCompile(`(?m)^=== ([A-Z_]+) ===[\t\r ]*$`)

// Python str.isspace includes these four controls in addition to Unicode White_Space.
func nativeSpace(r rune) bool { return unicode.IsSpace(r) || r >= 0x1c && r <= 0x1f }

// Keep original byte boundaries for each normalized rune: replacement must not
// collapse whitespace in the untouched parts of a chapter.
func normalizedNativeText(s string) (string, []int, []int) {
	var b strings.Builder
	var starts, ends []int
	pendingStart, pendingEnd := -1, -1
	for i, r := range s {
		end := i + utf8.RuneLen(r)
		if nativeSpace(r) {
			if pendingStart < 0 {
				pendingStart = i
			}
			pendingEnd = end
			continue
		}
		if pendingStart >= 0 && len(starts) > 0 {
			b.WriteByte(' ')
			starts = append(starts, pendingStart)
			ends = append(ends, pendingEnd)
		}
		pendingStart = -1
		b.WriteRune(r)
		starts = append(starts, i)
		ends = append(ends, end)
	}
	return b.String(), starts, ends
}

func uniqueNativeMatch(s, target string) int {
	if target == "" {
		return -1
	}
	i := strings.Index(s, target)
	if i < 0 || strings.Contains(s[i+len(target):], target) {
		return -1
	}
	return i
}

func applyNativePatches(content string, patches []nativePatch) nativePatchResult {
	result := nativePatchResult{Content: content, Mode: "patches"}
	for _, p := range patches {
		start, end := -1, -1
		if i := uniqueNativeMatch(result.Content, p.Target); i >= 0 {
			start, end = i, i+len(p.Target)
		} else {
			normalizedTarget, _, _ := normalizedNativeText(p.Target)
			if utf8.RuneCountInString(normalizedTarget) >= 10 {
				normalized, starts, ends := normalizedNativeText(result.Content)
				if i := uniqueNativeMatch(normalized, normalizedTarget); i >= 0 {
					first := utf8.RuneCountInString(normalized[:i])
					last := first + utf8.RuneCountInString(normalizedTarget) - 1
					start, end = starts[first], ends[last]
				}
			}
		}
		if start < 0 {
			result.Skipped++
			continue
		}
		result.Content = result.Content[:start] + p.Replacement + result.Content[end:]
		result.Applied++
		result.Touched += utf8.RuneCountInString(p.Target)
	}
	return result
}

func decodeNativeRevision(reply, content string, issues []auditIssue) (chapterRevisionOutput, nativePatchResult, error) {
	var v chapterRevisionOutput
	report := nativePatchResult{}
	// Compatibility for existing CM providers returning the former JSON contract.
	// This is labelled explicitly; it is not original Luoshui output.
	if strings.HasPrefix(strings.TrimSpace(reply), "{") {
		if e := decodeAndValidate(reply, &v, func() error { return validateRevision(v, issues) }); e != nil {
			return v, report, e
		}
		report.Mode = "cm_legacy_json"
		report.Content = v.Content
		return v, report, nil
	}
	sections := map[string]string{}
	matches := nativeSectionPattern.FindAllStringSubmatchIndex(reply, -1)
	for i, m := range matches {
		name := reply[m[2]:m[3]]
		if _, duplicate := sections[name]; duplicate {
			return v, report, fmt.Errorf("修订结果重复出现 %s 段", name)
		}
		end := len(reply)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		sections[name] = strings.TrimSpace(reply[m[1]:end])
	}
	v.Strategy = sections["FIXED_ISSUES"]
	v.ResolvedIssueIDs = []string{}
	patches, whole := sections["PATCHES"], sections["REVISED_CONTENT"]
	if v.Strategy == "" || (patches == "") == (whole == "") {
		return v, report, errors.New("修订结果须包含修正说明，且局部补丁与整章正文只能提供一种")
	}
	if patches != "" {
		var parsed []nativePatch
		for _, m := range nativePatchPattern.FindAllStringSubmatch(patches, -1) {
			if m[1] != "" {
				parsed = append(parsed, nativePatch{m[1], m[2]})
			}
		}
		if len(parsed) == 0 {
			return v, report, errors.New("未返回有效局部补丁，原文未修改")
		}
		report = applyNativePatches(content, parsed)
		if report.Applied == 0 {
			return v, report, errors.New("局部补丁未能唯一匹配章节原文，原文未修改")
		}
		v.Content = report.Content
	} else {
		v.Content = whole
		report.Content = whole
		report.Mode = "revised_content"
	}
	report.UpdatedState = sections["UPDATED_STATE"]
	return v, report, validateRevision(v, issues)
}

func (p *Pipeline) callNativeRevision(ctx context.Context, input RunContext, sink EventSink, provider batchfactoryv11.DirectorProvider, spec modelCallSpec, target any) (json.RawMessage, error) {
	payload, ok := spec.userPayload.(map[string]any)
	if !ok {
		return nil, errors.New("修订输入无效")
	}
	content, _ := payload["chapter_content"].(string)
	issues, _ := payload["audit_issues"].([]auditIssue)
	spec.definition = novelprompt.NativeChapterRevision
	var route *nativeRevisionRoute
	if input.ReviewCycleVersion >= 3 {
		resolved := resolveNativeRevisionRoute(issues)
		route = &resolved
		spec.definition = novelprompt.RoutedNativeChapterRevision(resolved.Mode)
		copy := make(map[string]any, len(payload)+2)
		for k, v := range payload {
			copy[k] = v
		}
		copy["revision_output_mode"] = resolved.Mode
		copy["revision_route_reason"] = resolved.Reason
		spec.userPayload = copy
	}
	if input.LongImitation != nil {
		spec.definition.SystemPrompt += novelprompt.LongImitationSystemSupplement
	}
	user, e := marshalPromptPayload(withThinkingLevel(spec.userPayload, input.Run.ThinkingLevel))
	if e != nil {
		return nil, e
	}
	request := batchfactoryv11.TextCompletionRequest{SystemPrompt: agentRequirementsSystemPrompt(input, spec.definition), UserPrompt: user, Temperature: spec.temperature, MaxTokens: spec.maxTokens, DisableJSONResponse: true}
	reply, e := p.complete(ctx, input, sink, provider, spec.step, spec.definition.ID+spec.key, request, spec.definition.Version)
	if e != nil {
		return nil, e
	}
	if input.LongImitation != nil {
		if limitError := checkImitationRawReply(spec.definition.ID, reply); limitError != nil {
			return nil, limitError
		}
	}
	revision, report, e := decodeNativeRevision(reply, content, issues)
	if e != nil {
		return nil, &DomainError{Code: ErrInvalidOutput, Message: e.Error(), Cause: e}
	}
	if route != nil {
		report.Route = route
		if route.Mode == "patch-only" && report.Mode != "patches" {
			return nil, &DomainError{Code: ErrInvalidOutput, Message: "本轮要求局部修改，模型返回了整章正文，原文未修改"}
		}
		if route.Mode == "rewrite-only" && report.Mode == "patches" {
			return nil, &DomainError{Code: ErrInvalidOutput, Message: "本轮要求整章重写，模型只返回局部补丁，原文未修改"}
		}
	}
	output, ok := target.(*chapterRevisionOutput)
	if !ok {
		return nil, errors.New("修订输出类型无效")
	}
	*output = revision
	// Preserve a valid JSON snapshot for existing storage/readback, including raw
	// native sections. UPDATED_STATE remains evidence, not an unaudited truth write.
	return json.Marshal(struct {
		chapterRevisionOutput
		RawOutput string            `json:"raw_output"`
		Native    nativePatchResult `json:"native_revision"`
	}{revision, reply, report})
}
