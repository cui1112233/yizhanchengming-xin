package novelcreation

import (
	_ "embed"
	"fmt"
	catalog "qiantie/backend/internal/prompt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Templates recovered from novels.pyd b5b22086, inkos.short_fiction_prompts.
// Only dynamic placeholder syntax is changed to a non-executing Go renderer.
// Evidence: docs/verification/novel-full-draft-20261120/FINDINGS.md.
//
//go:embed short_full/writer-system.txt
var shortFullWriterSystem string

//go:embed short_full/writer-user.txt
var ShortFullWriterUserTemplate string

//go:embed short_full/continuation-user.txt
var ShortFullContinuationUserTemplate string

//go:embed short_full/review-system.txt
var shortFullReviewSystem string

//go:embed short_full/review-user.txt
var ShortFullReviewUserTemplate string

//go:embed short_full/revision-user.txt
var ShortFullRevisionUserTemplate string

//go:embed short_full/package-system.txt
var shortFullPackageSystem string

//go:embed short_full/package-user.txt
var ShortFullPackageUserTemplate string

//go:embed short_full/chapter-block.txt
var ShortFullChapterBlockTemplate string

const (
	ShortFullWriterUserID       = "novel-creation.short-full-writer-user"
	ShortFullContinuationUserID = "novel-creation.short-full-continuation-user"
	ShortFullReviewUserID       = "novel-creation.short-full-review-user"
	ShortFullRevisionUserID     = "novel-creation.short-full-revision-user"
	ShortFullPackageUserID      = "novel-creation.short-full-package-user"
	ShortFullChapterBlockID     = "novel-creation.short-full-chapter-block"
)

var ShortFullWriter = catalog.Definition{ID: "novel-creation.short-full-draft", Version: 1, SystemPrompt: shortFullWriterSystem}
var ShortFullContinuation = catalog.Definition{ID: "novel-creation.short-full-continuation", Version: 1, SystemPrompt: shortFullWriterSystem}
var ShortFullRevision = catalog.Definition{ID: "novel-creation.short-full-revision", Version: 1, SystemPrompt: shortFullWriterSystem}
var ShortFullReview = catalog.Definition{ID: "novel-creation.short-full-review", Version: 1, SystemPrompt: shortFullReviewSystem}
var ShortFullPackage = catalog.Definition{ID: "novel-creation.short-full-package", Version: 1, SystemPrompt: shortFullPackageSystem}

type ShortFullPromptInput struct {
	Direction, OutlineMarkdown, DraftMarkdown, DraftTitle, Review string
	ChapterCount, CharsPerChapter                                 int
	MissingChapters                                               []int
}

var fullPlaceholder = regexp.MustCompile(`\{\{[^{}]+\}\}`)

func fullPromptTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f) })
}

// Validate the stored template before substituting; braces in user material
// remain plain data, never recursively interpreted as template instructions.
func RenderShortFullPrompt(template, chapterBlock string, in ShortFullPromptInput) (string, error) {
	if in.ChapterCount < 1 || in.ChapterCount > 50 || in.CharsPerChapter < 1 {
		return "", fmt.Errorf("全篇短篇模板参数无效")
	}
	values := map[string]string{
		"{{direction}}": in.Direction, "{{outline_markdown}}": in.OutlineMarkdown,
		"{{outline_markdown_trimmed}}": fullPromptTrim(in.OutlineMarkdown),
		"{{draft_markdown}}":           in.DraftMarkdown, "{{draft_markdown_trimmed}}": fullPromptTrim(in.DraftMarkdown),
		"{{existing_draft_markdown}}": in.DraftMarkdown,
		"{{draft_title}}":             in.DraftTitle, "{{review}}": in.Review, "{{review_trimmed}}": fullPromptTrim(in.Review),
		"{{chapter_count}}": strconv.Itoa(in.ChapterCount), "{{chars_per_chapter}}": strconv.Itoa(in.CharsPerChapter),
		"{{chapter_blocks}}": "", "{{missing_chapters}}": "",
	}
	for _, token := range fullPlaceholder.FindAllString(template, -1) {
		if _, ok := values[token]; !ok {
			return "", fmt.Errorf("全篇模板包含未知占位符：%s", token)
		}
	}
	for _, token := range fullPlaceholder.FindAllString(chapterBlock, -1) {
		if token != "{{n}}" {
			return "", fmt.Errorf("章节模板包含未知占位符：%s", token)
		}
	}
	numbers := in.MissingChapters
	if len(numbers) == 0 {
		numbers = make([]int, in.ChapterCount)
		for i := range numbers {
			numbers[i] = i + 1
		}
	}
	blocks := make([]string, 0, len(numbers))
	missing := make([]string, 0, len(numbers))
	seen := map[int]bool{}
	for _, n := range numbers {
		if n < 1 || n > in.ChapterCount || seen[n] {
			return "", fmt.Errorf("缺章编号无效或重复")
		}
		seen[n] = true
		missing = append(missing, strconv.Itoa(n))
		blocks = append(blocks, strings.TrimRight(strings.ReplaceAll(chapterBlock, "{{n}}", strconv.Itoa(n)), "\r\n"))
	}
	values["{{chapter_blocks}}"] = strings.Join(blocks, "\n")
	values["{{missing_chapters}}"] = strings.Join(missing, ", ")
	pairs := make([]string, 0, len(values)*2)
	for k, v := range values {
		pairs = append(pairs, k, v)
	}
	return strings.NewReplacer(pairs...).Replace(template), nil
}
