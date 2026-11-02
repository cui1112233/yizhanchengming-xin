package novelfetchworkshop

import "context"

// ClassificationResult is the shared result of the novel metadata stage.
// Both Novel Fetch and Batch Factory use the same upstream-first, AI-fallback
// ordering through this function.
type ClassificationResult struct {
	Gender GenderDecision `json:"gender"`
	Style  StyleDecision  `json:"style"`
}

func ClassifyBook(ctx context.Context, manualGender string, info BookInfo, judge AIGenderJudge, provider RewriteProvider) ClassificationResult {
	result := ClassificationResult{
		Gender: ResolveGender(ctx, manualGender, info, judge),
		Style:  ResolveStyle(ctx, info, provider),
	}
	return completeClassification(result)
}

// completeClassification never leaves the pair empty. A missing value is
// derived from the other one, and an unknown pair falls back to the most
// common female modern style. Production continues; the user can adjust the
// values afterwards instead of being blocked by a confirmation step.
func completeClassification(result ClassificationResult) ClassificationResult {
	if result.Style.Style == "" && result.Gender.Gender != "" && result.Gender.Gender != "待确认" {
		result.Style = StyleDecision{
			Style:      defaultStyleForGender(result.Gender.Gender),
			Source:     "fallback",
			Reason:     "平台与 AI 未给出明确风格，已按男女频采用默认风格",
			Confidence: 0.5,
		}
	}
	if result.Gender.Gender == "" || result.Gender.Gender == "待确认" {
		if result.Style.Style != "" {
			result.Gender = GenderDecision{
				Gender:     genderForStyle(result.Style.Style),
				Source:     GenderFallback,
				Reason:     "平台与 AI 未给出明确男女频，已按风格采用默认男女频",
				Confidence: 0.5,
			}
		} else {
			result.Gender = GenderDecision{Gender: "女频", Source: GenderFallback, Reason: "未能判断男女频，已采用默认女频", Confidence: 0.5}
			result.Style = StyleDecision{Style: "现代通用", Source: "fallback", Reason: "未能判断风格，已采用默认风格", Confidence: 0.5}
		}
	}
	return result
}
