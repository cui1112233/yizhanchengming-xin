package novelcreation

import (
	"encoding/json"
	"regexp"
	"strings"

	novelprompt "qiantie/backend/internal/prompt/novelcreation"
)

// Static derivation, including scope precedence: ROUTING-SPEC.md.
type nativeRevisionRoute struct {
	Mode   string `json:"mode"`
	Reason string `json:"reason"`
}

func compileNativeRoutePatterns(values []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(values))
	for _, v := range values {
		out = append(out, regexp.MustCompile("(?i)"+v))
	}
	return out
}

var nativeStructuralPatterns = compileNativeRoutePatterns(novelprompt.NativeRevisionStructuralPatterns)
var nativeLocalPatterns = compileNativeRoutePatterns(novelprompt.NativeRevisionLocalPatterns)

func nativeScopeField(raw json.RawMessage) (string, bool) {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	switch v := value.(type) {
	case string:
		return v, v != ""
	case bool:
		return "", v
	case float64:
		return "", v != 0
	case []any:
		return "", len(v) > 0
	case map[string]any:
		return "", len(v) > 0
	}
	return "", false
}

func nativeIssueScope(i auditIssue) string {
	if scope, ok := nativeScopeField(i.RepairScope); ok {
		return scope
	}
	scope, _ := nativeScopeField(i.RepairScopeSnake)
	return scope
}

func matchesNativeRoutePatterns(i auditIssue, patterns []*regexp.Regexp) bool {
	var category string
	_ = json.Unmarshal(i.Category, &category)
	text := category + " " + i.Description
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func resolveNativeRevisionRoute(issues []auditIssue) nativeRevisionRoute {
	if len(issues) == 0 {
		return nativeRevisionRoute{"allow-full", "审稿问题为空"}
	}
	var scoped, blocking []auditIssue
	for _, i := range issues {
		nonInfo := strings.ToLower(i.Severity) != "info"
		_, camel := nativeScopeField(i.RepairScope)
		_, snake := nativeScopeField(i.RepairScopeSnake)
		// Preserve original and/or precedence; snake scope can include an info item.
		if nonInfo && camel || snake {
			scoped = append(scoped, i)
		}
		if nonInfo {
			blocking = append(blocking, i)
		}
	}
	if len(scoped) > 0 {
		allLocal := true
		for _, i := range scoped {
			scope := nativeIssueScope(i)
			if scope == "structural" {
				return nativeRevisionRoute{"rewrite-only", "审稿明确标注结构问题"}
			}
			if scope != "local" {
				allLocal = false
			}
		}
		if len(scoped) == len(blocking) && allLocal {
			return nativeRevisionRoute{"patch-only", "审稿问题全部明确标注局部范围"}
		}
	}
	if len(blocking) == 0 {
		return nativeRevisionRoute{"patch-only", "仅包含提示信息"}
	}
	allLocal := true
	for _, i := range blocking {
		if matchesNativeRoutePatterns(i, nativeStructuralPatterns) {
			return nativeRevisionRoute{"rewrite-only", "命中原版结构问题规则"}
		}
		if !matchesNativeRoutePatterns(i, nativeLocalPatterns) {
			allLocal = false
		}
	}
	if allLocal {
		return nativeRevisionRoute{"patch-only", "全部命中原版局部问题规则"}
	}
	return nativeRevisionRoute{"allow-full", "存在未归类问题，保留两种修订选择"}
}
