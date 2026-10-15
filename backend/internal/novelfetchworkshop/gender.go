package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"strings"
)

type AIGenderJudge func(ctx context.Context, info BookInfo) (string, error)

type GenderSource string

const (
	GenderManual   GenderSource = "manual"
	GenderBookInfo GenderSource = "bookinfo"
	GenderAI       GenderSource = "ai"
	GenderUnknown  GenderSource = "unknown"
)

type BookInfo struct {
	Category string `json:"category"`
	Genre    string `json:"genre"`
	WorkTitle string `json:"workTitle,omitempty"`
	Themes    string `json:"themes,omitempty"`
	Labels    string `json:"labels,omitempty"`
	// Text is supplied only to the AI fallback. It is deliberately omitted
	// from the persisted book-info JSON because it is already stored as the
	// document body.
	Text string `json:"-"`
}

// UnmarshalJSON accepts the upstream's inconsistent genre representation.
// Some 121 responses use a numeric category code while others use its text.
// Keep the public BookInfo contract textual so callers do not need to know
// which representation the upstream used.
func (info *BookInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		Category  string          `json:"category"`
		Genre     json.RawMessage `json:"genre"`
		WorkTitle string          `json:"work_title"`
		Theme     []struct{ Name string `json:"name"` } `json:"theme"`
		Labels    []struct{ Name string `json:"name"` } `json:"labels"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	genre, err := decodeUpstreamTextOrNumber(raw.Genre)
	if err != nil {
		return err
	}
	themes := make([]string, 0, len(raw.Theme))
	for _, item := range raw.Theme {
		if text := strings.TrimSpace(item.Name); text != "" {
			themes = append(themes, text)
		}
	}
	labels := make([]string, 0, len(raw.Labels))
	for _, item := range raw.Labels {
		if text := strings.TrimSpace(item.Name); text != "" {
			labels = append(labels, text)
		}
	}
	*info = BookInfo{Category: raw.Category, Genre: genre, WorkTitle: raw.WorkTitle, Themes: strings.Join(themes, "、"), Labels: strings.Join(labels, "、")}
	return nil
}

func decodeUpstreamTextOrNumber(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err != nil {
		return "", err
	}
	return number.String(), nil
}

type GenderDecision struct {
	Gender string       `json:"gender"`
	Source GenderSource `json:"genderSource"`
	Reason string       `json:"genderReason"`
}

func ResolveGender(ctx context.Context, manual string, info BookInfo, judge AIGenderJudge) GenderDecision {
	if strings.TrimSpace(manual) != "" {
		return DecideGender(manual, BookInfo{}, "")
	}
	category := normalizeGender(info.Category)
	genre := normalizeGender(info.Genre)
	if (category != "" && genre != "" && category != genre) || category == "" && genre == "" {
		if judge != nil {
			aiGender, err := judge(ctx, info)
			if err == nil {
				return DecideGender("", info, aiGender)
			}
		}
		return GenderDecision{Gender: "待确认", Source: GenderUnknown, Reason: "分类信息无法确定且 AI 未返回有效判断"}
	}
	return DecideGender("", info, "")
}

func DecideGender(manual string, info BookInfo, aiGender string) GenderDecision {
	if strings.TrimSpace(manual) != "" {
		if gender := normalizeGender(manual); gender != "" {
			return GenderDecision{Gender: gender, Source: GenderManual, Reason: "采用手工填写的男女频"}
		}
		return GenderDecision{Gender: "待确认", Source: GenderUnknown, Reason: "手工填写的男女频不明确"}
	}

	category := normalizeGender(info.Category)
	genre := normalizeGender(info.Genre)
	if category != "" && genre != "" && category != genre {
		if gender := normalizeGender(aiGender); gender != "" {
			return GenderDecision{Gender: gender, Source: GenderAI, Reason: "分类信息冲突，采用 AI 判断结果"}
		}
		return GenderDecision{Gender: "待确认", Source: GenderUnknown, Reason: "分类信息冲突且没有有效 AI 判断结果"}
	}
	if category == "" && genre == "" && (strings.TrimSpace(info.Category) != "" || strings.TrimSpace(info.Genre) != "") {
		if gender := normalizeGender(aiGender); gender != "" {
			return GenderDecision{Gender: gender, Source: GenderAI, Reason: "分类信息不明确，采用 AI 判断结果"}
		}
		return GenderDecision{Gender: "待确认", Source: GenderUnknown, Reason: "分类信息不明确且没有有效 AI 判断结果"}
	}
	if category != "" || genre != "" {
		return GenderDecision{Gender: firstGender(category, genre), Source: GenderBookInfo, Reason: "采用书籍分类信息中的明确男女频"}
	}
	if gender := normalizeGender(aiGender); gender != "" {
		return GenderDecision{Gender: gender, Source: GenderAI, Reason: "书籍分类信息不明确，采用 AI 判断结果"}
	}
	return GenderDecision{Gender: "待确认", Source: GenderUnknown, Reason: "没有有效的手工、书籍分类或 AI 判断结果"}
}

func normalizeGender(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	male := strings.Contains(value, "男")
	female := strings.Contains(value, "女")
	if male == female {
		return ""
	}
	if male {
		return "男频"
	}
	return "女频"
}

func firstGender(category, genre string) string {
	if category != "" {
		return category
	}
	return genre
}
