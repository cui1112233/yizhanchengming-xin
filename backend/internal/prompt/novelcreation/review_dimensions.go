package novelcreation

import (
	_ "embed"
	"encoding/json"
	"fmt"
	promptcatalog "qiantie/backend/internal/prompt"
	"strings"
)

// Source and exact file offsets: docs/reverse-engineering/novel-review-dimensions-20261119.
//
//go:embed review_dimensions.json
var reviewDimensionsJSON []byte

type ReviewDimension struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	DefaultEnabled bool   `json:"defaultEnabled"`
}

// Return a new catalog so callers cannot mutate the system preset.
func ReviewDimensionCatalog() []ReviewDimension {
	var items []ReviewDimension
	if err := json.Unmarshal(reviewDimensionsJSON, &items); err != nil {
		panic(err)
	}
	return items
}

func DefaultReviewDimensionIDs() []int {
	ids := []int{}
	for _, d := range ReviewDimensionCatalog() {
		if d.DefaultEnabled {
			ids = append(ids, d.ID)
		}
	}
	return ids
}

// WithReviewDimensions keeps content-only selection separate from the output contract.
// Nil is a legacy frozen run; an explicit empty list disables all optional dimensions.
func WithReviewDimensions(def promptcatalog.Definition, system string, ids *[]int) string {
	if ids == nil {
		return system
	}
	switch def.ID {
	case ChapterAudit.ID, AgentChapterAudit.ID, ShortReviewPrompt.ID, ChapterRevision.ID, ShortBodyRevision.ID, ShortOutlineRevision.ID:
	default:
		return system
	}
	selected := map[int]bool{}
	for _, id := range *ids {
		selected[id] = true
	}
	var lines []string
	for _, d := range ReviewDimensionCatalog() {
		if selected[d.ID] {
			lines = append(lines, fmt.Sprintf("%d. %s：%s", d.ID, d.Title, d.Content))
		}
	}
	block := "内置审稿维度（仅以下勾选项参与额外维度检查；未勾选项不作为额外扣分依据。缺少正典、前文、设定或备忘资料时注明无法核验，不得编造资料或声称已通过；仍遵守基础输出合同）：\n"
	if len(lines) == 0 {
		block += "未选择额外内置审稿维度。"
	} else {
		block += strings.Join(lines, "\n")
	}
	// Insert before the custom-review block, preserving original built-in -> custom order.
	marker := "\n\n自定义补充审稿要求"
	if i := strings.Index(system, marker); i >= 0 {
		return system[:i] + "\n\n" + block + system[i:]
	}
	return system + "\n\n" + block
}
