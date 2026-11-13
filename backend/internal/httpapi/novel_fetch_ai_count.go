package httpapi

import (
	"fmt"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// Counts are the target for the next manual generation, never a mutation of
// an already queued ProcessJob's selected version snapshot.
func workshopAIRequestedCount(document novelfetchworkshop.Document, config map[string]any, explicit int) int {
	rewrite, _ := config["rewrite"].(map[string]any)
	if app, ok := config["appConfig"].(map[string]any); ok {
		if configured, ok := app["rewrite"].(map[string]any); ok {
			rewrite = configured
		}
	}
	maximum := intValue(rewrite["max_ai_count"])
	if maximum < 1 {
		maximum = 5
	}
	if maximum > 20 {
		maximum = 20
	}
	count := explicit
	if count < 1 {
		count = intValue(document.Meta["aiCount"])
	}
	if count < 1 {
		count = intValue(document.Meta["ai_count"])
	}
	if count < 1 {
		count = intValue(rewrite["default_ai_count"])
	}
	if count < 1 {
		count = 1
	}
	if count > maximum {
		count = maximum
	}
	return count
}

func workshopMissingAIVersions(document novelfetchworkshop.Document, count int) []string {
	missing := []string{}
	for index := 1; index <= count; index++ {
		id := fmt.Sprintf("ai%d", index)
		if ref, ok := document.BodyRefs[id]; ok && ref.State == "ready" {
			continue
		}
		if state, ok := document.Versions[id].(map[string]any); ok {
			if strings.EqualFold(firstText(state["state"]), "ready") {
				continue
			}
		} else if strings.TrimSpace(firstText(document.Versions[id])) != "" {
			continue
		}
		missing = append(missing, id)
	}
	return missing
}
