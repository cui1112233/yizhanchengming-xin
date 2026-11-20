package novelcreation

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Source: recovered novels.pipeline.short_fiction_parsers. The intentionally
// loose Markdown fallback is preserved; default generation asks for CHAPTER
// blocks. Evidence is in docs/verification/novel-full-draft-20261120.
type shortFullDraft struct {
	Title, OpeningHook, RawContent string
	Chapters                       []ShortChapter
}

const shortFullSpace = `[\t\n\r\f\v\x{001c}-\x{001f}\x{0085}\p{Z}]`

var shortFullMarker = regexp.MustCompile(`(?im)^` + shortFullSpace + `*===` + shortFullSpace + `*([A-Zİı0-9_ ]+)` + shortFullSpace + `*===` + shortFullSpace + `*$`)
var shortFullSanitizeMarker = regexp.MustCompile(`(?im)^` + shortFullSpace + `*===` + shortFullSpace + `*[A-Zİı0-9_ ]+` + shortFullSpace + `*===` + shortFullSpace + `*$`)
var shortFullFenceStart = regexp.MustCompile(`(?i)^` + "```" + `(?:md|markdown)?` + shortFullSpace + `*`)
var shortFullFenceEnd = regexp.MustCompile("```" + shortFullSpace + `*$`)
var shortFullTitlePrefix = regexp.MustCompile(`^#+` + shortFullSpace + `*`)
var shortFullBookQuotes = regexp.MustCompile(`^《(.+)》$`)

func shortFullIsSpace(r rune) bool  { return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f) }
func shortFullTrim(s string) string { return strings.TrimFunc(s, shortFullIsSpace) }

func shortFullCharacterCount(s string) int {
	count := 0
	for _, r := range s {
		if !shortFullIsSpace(r) {
			count++
		}
	}
	return count
}

func shortFullNormalizeTitle(s string) string {
	// Python splitlines includes Unicode separators; treating them as line
	// boundaries preserves which nonempty line supplies the title.
	lines := strings.FieldsFunc(s, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\v' || r == '\f' || r == 0x85 || r == 0x2028 || r == 0x2029 || (r >= 0x1c && r <= 0x1e)
	})
	for _, line := range lines {
		line = shortFullTrim(shortFullTitlePrefix.ReplaceAllString(line, ""))
		if line != "" {
			return shortFullTrim(shortFullBookQuotes.ReplaceAllString(line, "$1"))
		}
	}
	return ""
}

func shortFullNormalizeChapterTitle(s string, n int) string {
	title := shortFullNormalizeTitle(s)
	re := regexp.MustCompile(`^第` + shortFullSpace + `*` + strconv.Itoa(n) + shortFullSpace + `*章` + shortFullSpace + `*`)
	title = shortFullTrim(re.ReplaceAllString(title, ""))
	if title == "" {
		return fmt.Sprintf("第%d章", n)
	}
	return title
}

func shortFullBlocks(raw, tag string) []string {
	markers := shortFullMarker.FindAllStringSubmatchIndex(raw, -1)
	result := []string{}
	for i, m := range markers {
		// Python re.IGNORECASE also folds dotted/dotless I into ASCII I.
		name := strings.NewReplacer("İ", "I", "ı", "i").Replace(shortFullTrim(raw[m[2]:m[3]]))
		if !strings.EqualFold(name, tag) {
			continue
		}
		end := len(raw)
		if i+1 < len(markers) {
			end = markers[i+1][0]
		}
		result = append(result, shortFullTrim(raw[m[1]:end]))
	}
	return result
}

func shortFullFirstBlock(raw, tag string) string {
	blocks := shortFullBlocks(raw, tag)
	if len(blocks) > 0 {
		return blocks[0]
	}
	return ""
}

func shortFullLastBlock(raw, tag string) string {
	result := ""
	for _, block := range shortFullBlocks(raw, tag) {
		if block != "" {
			result = block
		}
	}
	return result
}

func shortFullDuplicateTitleBody(raw string, n int) string {
	pattern := regexp.MustCompile(`(?im)^` + shortFullSpace + `*===` + shortFullSpace + `*CHAPTER ` + strconv.Itoa(n) + ` TITLE` + shortFullSpace + `*===` + shortFullSpace + `*$`)
	matches := pattern.FindAllStringIndex(raw, -1)
	if len(matches) < 2 {
		return ""
	}
	text := strings.TrimLeft(raw[matches[1][1]:], "\n")
	boundary := regexp.MustCompile(`(?im)^` + shortFullSpace + `*===` + shortFullSpace + `*(?:CHAPTER` + shortFullSpace + `+\p{Nd}+` + shortFullSpace + `+(?:TITLE|CONTENT)|SHORT_FICTION_[A-Zİı0-9_ ]+)` + shortFullSpace + `*===` + shortFullSpace + `*$`)
	if m := boundary.FindStringIndex(text); m != nil {
		text = text[:m[0]]
	}
	return shortFullTrim(text)
}

func shortFullMarkdownTitle(raw string, n int) string {
	re := regexp.MustCompile(`(?m)^##` + shortFullSpace + `*(?:第` + shortFullSpace + `*` + strconv.Itoa(n) + shortFullSpace + `*章` + shortFullSpace + `*)?(.+)$`)
	if m := re.FindStringSubmatch(raw); m != nil {
		return shortFullTrim(m[1])
	}
	return ""
}

func shortFullMarkdownBody(raw string, n int) string {
	// Both native numbered groups are optional. It can match the first ##
	// section for more than one n; this recovered defect is documented/tested.
	re := regexp.MustCompile(`(?m)^##` + shortFullSpace + `*(?:第` + shortFullSpace + `*` + strconv.Itoa(n) + shortFullSpace + `*章` + shortFullSpace + `*)?.*$\n`)
	m := re.FindStringIndex(raw)
	if m == nil {
		return ""
	}
	text := raw[m[1]:]
	end := regexp.MustCompile(`(?m)^##` + shortFullSpace + `*(?:第` + shortFullSpace + `*` + strconv.Itoa(n+1) + shortFullSpace + `*章` + shortFullSpace + `*)?.*$`)
	if next := end.FindStringIndex(text); next != nil {
		text = text[:next[0]]
	}
	return shortFullTrim(text)
}

func shortFullSanitize(s string) string {
	s = shortFullFenceStart.ReplaceAllString(s, "")
	s = shortFullFenceEnd.ReplaceAllString(s, "")
	s = shortFullSanitizeMarker.ReplaceAllString(s, "")
	return shortFullTrim(s)
}

func parseShortFullDraft(raw string, expected int) (shortFullDraft, error) {
	if expected < 1 || expected > 50 || !utf8.ValidString(raw) || jsLength(raw) > LongImitationMaxUTF16 {
		return shortFullDraft{}, fmt.Errorf("全篇短篇章节数或文本范围无效")
	}
	title := shortFullFirstBlock(raw, "SHORT_FICTION_TITLE")
	if title == "" {
		if m := regexp.MustCompile(`(?m)^#` + shortFullSpace + `+(.+)$`).FindStringSubmatch(raw); m != nil {
			title = m[1]
		}
	}
	title = shortFullNormalizeTitle(title)
	if title == "" {
		title = "未命名短篇"
	}
	hook := shortFullFirstBlock(raw, "SHORT_FICTION_OPENING_HOOK")
	if hook == "" {
		hook = shortFullFirstBlock(raw, "OPENING_HOOK")
	}
	draft := shortFullDraft{Title: title, OpeningHook: hook, RawContent: raw, Chapters: make([]ShortChapter, 0, expected)}
	for n := 1; n <= expected; n++ {
		name := shortFullFirstBlock(raw, fmt.Sprintf("CHAPTER %d TITLE", n))
		if name == "" {
			name = shortFullMarkdownTitle(raw, n)
		}
		body := shortFullLastBlock(raw, fmt.Sprintf("CHAPTER %d CONTENT", n))
		if body == "" {
			body = shortFullDuplicateTitleBody(raw, n)
		}
		if body == "" {
			body = shortFullMarkdownBody(raw, n)
		}
		draft.Chapters = append(draft.Chapters, ShortChapter{Number: n, Title: shortFullNormalizeChapterTitle(name, n), Content: shortFullSanitize(body)})
	}
	return draft, nil
}

func missingShortFullChapters(d shortFullDraft) []int {
	result := []int{}
	for _, c := range d.Chapters {
		if shortFullTrim(c.Content) == "" {
			result = append(result, c.Number)
		}
	}
	return result
}
func validateShortFullDraft(d shortFullDraft, expected int) error {
	if len(d.Chapters) != expected {
		return fmt.Errorf("短篇草稿应有%d章，实际%d章", expected, len(d.Chapters))
	}
	if missing := missingShortFullChapters(d); len(missing) > 0 {
		return fmt.Errorf("短篇草稿不完整：以下章节为空：%v", missing)
	}
	return nil
}
func renderShortFullDraft(d shortFullDraft) string {
	parts := []string{"# " + d.Title}
	if d.OpeningHook != "" {
		parts = append(parts, "## 开篇钩子\n\n"+d.OpeningHook)
	}
	for _, c := range d.Chapters {
		title := shortFullTrim(c.Title)
		if title == "" {
			title = fmt.Sprintf("第%d章", c.Number)
		} else if !regexp.MustCompile(`^第` + shortFullSpace + `*` + strconv.Itoa(c.Number) + shortFullSpace + `*章`).MatchString(title) {
			title = fmt.Sprintf("第%d章 %s", c.Number, title)
		}
		parts = append(parts, "## "+title+"\n\n"+c.Content)
	}
	return strings.Join(parts, "\n\n")
}
func estimateShortFullTokens(chapters, chars int) int {
	value := int(math.Ceil(float64(chapters)*float64(chars)*2.2)) + 4096
	if value < 12288 {
		return 12288
	}
	return value
}

type shortFullPackage struct {
	Title, Intro, CoverPrompt string
	SellingPoints             []string
}

func parseShortFullPackage(raw string, fallbackTitle ...string) shortFullPackage {
	first := func(tags ...string) string {
		for _, tag := range tags {
			if value := shortFullFirstBlock(raw, tag); value != "" {
				return value
			}
		}
		return ""
	}
	title := first("SHORT_FICTION_PACKAGE_TITLE", "SHORT_FICTION_TITLE")
	if title == "" {
		title = "未命名短篇"
		if len(fallbackTitle) > 0 {
			title = fallbackTitle[0]
		}
	}
	result := shortFullPackage{Title: shortFullNormalizeTitle(title), Intro: first("SHORT_FICTION_INTRO", "INTRO"), CoverPrompt: first("SHORT_FICTION_COVER_PROMPT", "COVER_PROMPT"), SellingPoints: []string{}}
	bullet := regexp.MustCompile(`^` + shortFullSpace + `*[-*]` + shortFullSpace + `*`)
	for _, line := range strings.Split(first("SHORT_FICTION_SELLING_POINTS", "SELLING_POINTS"), "\n") {
		if text := shortFullTrim(bullet.ReplaceAllString(line, "")); text != "" {
			result.SellingPoints = append(result.SellingPoints, text)
		}
	}
	return result
}
