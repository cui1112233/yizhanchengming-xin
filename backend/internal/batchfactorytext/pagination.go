package batchfactorytext

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultTargetCharacters = 150
const DefaultMaxPages = 35

type PageLimitError struct{ Pages, Limit int }

func (e *PageLimitError) Unwrap() error { return ErrInvalid }

func (e *PageLimitError) Error() string {
	return fmt.Sprintf("当前排版需要 %d 页，超过设置的 %d 页上限。正文未截断，请减小字号、行距或段距，或提高最多页数后重新分页", e.Pages, e.Limit)
}

// SentenceBreaks reports legal manual split points in the same coordinate space as Paginate.
func SentenceBreaks(text string) []int {
	runes := []rune(NormalizeLineEndings(text))
	var qs straightQuotes
	var breaks []int
	for start := 0; start < len(runes); {
		end := paragraphEnd(runes, start)
		for start < end {
			next := qs.sentenceEnd(runes, start, end)
			if next < len(runes) {
				// The paginator can carry part of a trailing separator to the next page.
				first := separatorStart(runes, start, next)
				if first == start {
					first = next
				}
				for offset := first; offset <= next; offset++ {
					breaks = append(breaks, offset)
				}
			}
			start = next
		}
	}
	return breaks
}

type PaginationInput struct {
	Text   string
	Layout LayoutConfig
}

// MeasureFunc uses the renderer's actual layout geometry, including its height limit.
type MeasureFunc func(text string, layout LayoutConfig) (fits bool, usedHeight float64, err error)

// SentenceOverflowError identifies the unbreakable range in normalized rune coordinates.
type SentenceOverflowError struct {
	Page        int
	StartOffset int
	EndOffset   int
	UsedHeight  float64
}

func (e *SentenceOverflowError) Error() string {
	return fmt.Sprintf("第 %d 页的完整句子超出可用空间（第 %d 至 %d 字），请调小字号或增大文字区域后重试", e.Page, e.StartOffset+1, e.EndOffset)
}

type pagePacker struct {
	ctx     context.Context
	runes   []rune
	layout  LayoutConfig
	measure MeasureFunc
	pages   []PageDraft
	start   int // Current page begins here.
	end     int // End of the last accepted unit.
}

// Paginate packs complete paragraphs first. Only a paragraph too tall on an
// empty page is divided into sentences. Every page is a source slice.
func Paginate(ctx context.Context, input PaginationInput, measure MeasureFunc) ([]PageDraft, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if measure == nil || input.Layout.TargetCharacters < 0 || !utf8.ValidString(input.Text) || strings.TrimSpace(input.Text) == "" {
		return nil, ErrInvalid
	}
	if mode := input.Layout.PaginationMode; mode != "" && mode != "auto" && mode != "target" {
		return nil, ErrInvalid
	}
	if input.Layout.TargetCharacters == 0 {
		input.Layout.TargetCharacters = DefaultTargetCharacters
	}
	limit := input.Layout.MaxPages
	if limit == 0 {
		limit = DefaultMaxPages
	}
	if limit < 1 || limit > 1000 {
		return nil, ErrInvalid
	}
	p := pagePacker{ctx: ctx, runes: []rune(NormalizeLineEndings(input.Text)), layout: input.Layout, measure: measure}
	reading := map[int]bool{}
	if resolved, size, err := resolveLayout(input.Layout); err == nil {
		reading = readingBreaks(string(p.runes), resolved, size.X-2*resolved.Padding)
	}
	var quotes straightQuotes
	for start := 0; start < len(p.runes); {
		end := paragraphEnd(p.runes, start)
		for offset := start + 1; offset < end; offset++ {
			if reading[offset] {
				end = offset
				break
			}
		}
		fits, _, err := p.accept(end)
		if err != nil {
			return nil, err
		}
		if fits {
			for i := start; i < end; i++ {
				quotes.observe(p.runes, i)
			}
		} else {
			for i := start; i < end; {
				next := quotes.sentenceEnd(p.runes, i, end)
				fits, height, err := p.accept(next)
				if err != nil {
					return nil, err
				}
				if !fits {
					return nil, &SentenceOverflowError{Page: len(p.pages) + 1, StartOffset: p.start, EndOffset: next, UsedHeight: height}
				}
				i = next
			}
		}
		start = end
	}
	p.flush()
	if len(p.pages) > limit {
		return nil, &PageLimitError{Pages: len(p.pages), Limit: limit}
	}
	if err := validateDrafts(input.Text, p.pages); err != nil {
		return nil, err
	}
	return p.pages, nil
}

func (p *pagePacker) flush() {
	if p.end == p.start {
		return
	}
	p.pages = append(p.pages, PageDraft{Text: string(p.runes[p.start:p.end]), StartOffset: p.start, EndOffset: p.end})
	p.start = p.end
}

// At a legal break near the soft target, prefer the closer character count.
// On an empty page geometry alone decides whether the next unit fits.
func (p *pagePacker) accept(end int) (bool, float64, error) {
	target := p.layout.TargetCharacters
	before, after := p.end-p.start, end-p.start
	if p.layout.PaginationMode != "auto" && before > 0 && after > target && (before >= target || after-target > target-before) {
		p.flush()
	}
	for {
		fits, height, err := p.measureTo(end)
		if err != nil {
			return false, 0, err
		}
		if fits {
			p.end = end
			return true, height, nil
		}
		if p.start == p.end {
			if end < len(p.runes) && unicode.IsSpace(p.runes[end-1]) {
				ok, err := p.carrySeparator(end)
				return ok, height, err
			}
			return false, height, nil
		}
		p.flush()
	}
}

func (p *pagePacker) measureTo(end int) (bool, float64, error) {
	if err := p.ctx.Err(); err != nil {
		return false, 0, err
	}
	fits, height, err := p.measure(string(p.runes[p.start:end]), p.layout)
	if err != nil {
		return false, 0, err
	}
	return fits, height, p.ctx.Err()
}

func separatorStart(runes []rune, start, end int) int {
	for end > start && unicode.IsSpace(runes[end-1]) {
		end--
	}
	return end
}

// When the words fit exactly, a separator after a whole paragraph or sentence
// may straddle the page break. Keep as much on this page as its geometry permits.
func (p *pagePacker) carrySeparator(end int) (bool, error) {
	contentEnd := separatorStart(p.runes, p.start, end)
	if contentEnd == p.start {
		return false, nil
	}
	fits, _, err := p.measureTo(contentEnd)
	if !fits || err != nil {
		return false, err
	}
	lo, hi := contentEnd, end
	for hi-lo > 1 {
		mid := lo + (hi-lo)/2
		fits, _, err := p.measureTo(mid)
		if err != nil {
			return false, err
		}
		if fits {
			lo = mid
		} else {
			hi = mid
		}
	}
	p.end = lo
	p.flush()
	return true, nil
}

// Attach blank separator lines to the preceding paragraph and preserve the
// indentation of the following paragraph. Leading blank lines stay with page 1.
func paragraphEnd(text []rune, start int) int {
	i := start
	for i < len(text) && unicode.IsSpace(text[i]) {
		i++
	}
	for i < len(text) && text[i] != '\n' {
		i++
	}
	if i == len(text) {
		return i
	}
	i++
	end := i
	for i < len(text) && unicode.IsSpace(text[i]) {
		if text[i] == '\n' {
			end = i + 1
		}
		i++
	}
	if i == len(text) {
		return i
	}
	return end
}

// Straight quotes use the same character on both sides. Follow their state
// across sentences and paragraphs, ignoring apostrophes inside English words.
type straightQuotes struct {
	single, double bool
	nextSingle     int // cached lookahead shared by consecutive possessives
}

func insideEnglishWord(text []rune, i int) bool {
	return i > 0 && i+1 < len(text) && unicode.In(text[i-1], unicode.Latin) && unicode.In(text[i+1], unicode.Latin)
}

// An s-apostrophe followed by a separated word can be a possessive, even
// inside quoted speech: 'James' cat.' A quote around a word ending in s
// is also possible: 'James' left. Inspect the next standalone quote.
func beforePossessedWord(text []rune, i int) bool {
	if i == 0 || (text[i-1] != 's' && text[i-1] != 'S') {
		return false
	}
	j := i + 1
	for j < len(text) && text[j] != '\n' && unicode.IsSpace(text[j]) {
		j++
	}
	return j > i+1 && j < len(text) && (unicode.In(text[j], unicode.Latin) || unicode.IsDigit(text[j]))
}

func (q *straightQuotes) hasSingleCloserAhead(text []rune, i int) bool {
	if q.nextSingle <= i {
		q.nextSingle = len(text)
		for j := i + 1; j < len(text); j++ {
			if text[j] == '\'' && !insideEnglishWord(text, j) && !beforePossessedWord(text, j) {
				q.nextSingle = j
				break
			}
		}
	}
	j := q.nextSingle
	return j < len(text) && !unicode.IsSpace(text[j-1]) && !strings.ContainsRune("“‘「『（([", text[j-1])
}

func (q *straightQuotes) state(text []rune, i int) *bool {
	switch text[i] {
	case '"':
		return &q.double
	case '\'':
		if insideEnglishWord(text, i) {
			return nil // don't, O'Neil
		}
		if !q.single && i > 0 && (text[i-1] == 's' || text[i-1] == 'S') {
			return nil // predicative possessive, e.g. "it's the dogs'."
		}
		if beforePossessedWord(text, i) && q.hasSingleCloserAhead(text, i) {
			return nil
		}
		return &q.single
	default:
		return nil
	}
}

func (q *straightQuotes) observe(text []rune, i int) {
	if open := q.state(text, i); open != nil {
		*open = !*open
	}
}

func (quotes *straightQuotes) sentenceEnd(text []rune, start, end int) int {
	for i := start; i < end; i++ {
		quotes.observe(text, i)
		if !strings.ContainsRune("。！？!?；;…", text[i]) && !(text[i] == '.' && isFinalPeriod(text, i)) {
			continue
		}
		i++
		// Repeated punctuation plus any closing quotes/brackets belong to
		// the sentence that just ended, as does the space between sentences.
		for i < end {
			if strings.ContainsRune("。.!?！？；;…’”」』）)]】》〉", text[i]) {
				i++
				continue
			}
			open := quotes.state(text, i)
			if open == nil || (!*open && hasFollowingContent(text[i+1:])) {
				break // A new opening quote belongs to the next sentence.
			}
			*open = false
			i++
		}
		for i < end && unicode.IsSpace(text[i]) {
			i++
		}
		return i
	}
	return end
}

func hasFollowingContent(text []rune) bool {
	for _, ch := range text {
		if !unicode.IsSpace(ch) {
			return true
		}
	}
	return false
}

func isFinalPeriod(text []rune, i int) bool {
	if i+1 < len(text) && text[i+1] == '.' {
		return true
	}
	// A dot inside a number, hostname or an English word is not a stop.
	if i+1 < len(text) && (unicode.IsDigit(text[i+1]) || unicode.In(text[i+1], unicode.Latin)) {
		return false
	}
	j := i
	for j > 0 && unicode.In(text[j-1], unicode.Latin) {
		j--
	}
	// Protect initials and dotted abbreviations such as J. Doe or e.g.
	if i-j == 1 && (unicode.IsUpper(text[j]) || (j >= 2 && text[j-1] == '.' && unicode.In(text[j-2], unicode.Latin))) {
		return false
	}
	switch strings.ToLower(string(text[j:i])) {
	case "mr", "mrs", "ms", "dr", "prof", "sr", "jr", "st":
		return false
	}
	return true
}
