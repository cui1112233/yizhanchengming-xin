package httpapi

import (
	"fmt"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// recordNovelFetchStyleDecision keeps the style decision and its failure state
// in the same legacy metadata object that the old workbench displays.
func recordNovelFetchStyleDecision(meta map[string]any, decision novelfetchworkshop.StyleDecision, resolveErr error) {
	if meta == nil {
		return
	}
	style := strings.TrimSpace(decision.Style)
	if style != "" {
		if decision.Confidence <= 0 && decision.Source != "ai" {
			decision.Confidence = 1
		}
		meta["style"] = style
		meta["styleSource"] = decision.Source
		meta["styleReason"] = decision.Reason
		meta["styleConfidence"] = decision.Confidence
		meta["styleStatus"] = "classified"
		delete(meta, "styleError")
		return
	}
	reason := strings.TrimSpace(decision.Reason)
	if resolveErr != nil {
		reason = strings.TrimSpace(resolveErr.Error())
	}
	if reason == "" {
		reason = "未返回有效风格"
	}
	meta["styleStatus"] = "failed"
	meta["styleError"] = reason
	meta["styleReason"] = reason
	meta["styleConfidence"] = 0.0
}

func inferNovelFetchStyleFailure(meta map[string]any) {
	if meta == nil {
		return
	}
	if status, ok := meta["styleStatus"]; ok && strings.TrimSpace(fmt.Sprint(status)) != "" {
		return
	}
	style := ""
	if value, ok := meta["style"]; ok {
		style = strings.TrimSpace(fmt.Sprint(value))
	}
	if style != "" && style != "<nil>" && style != "待判定" && style != "未判定" {
		meta["styleStatus"] = "classified"
		return
	}
	failed, ok := meta["siteSubmitFailedErrors"].(map[string]any)
	if !ok {
		return
	}
	for _, raw := range failed {
		message := strings.TrimSpace(fmt.Sprint(raw))
		if strings.Contains(message, "风格") {
			meta["styleStatus"] = "failed"
			meta["styleError"] = message
			return
		}
	}
}
