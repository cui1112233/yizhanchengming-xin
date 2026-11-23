package batchfactorytext

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultTargetCharacters = 150

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
	if input.Layout.TargetCharacters == 0 {
		input.Layout.TargetCharacters = DefaultTargetCharacters
	}
	p := pagePacker{ctx: ctx, runes: []rune(NormalizeLineEndings(input.Text)), layout: input.Layout, measure: measure}
	for start := 0; start < len(p.runes); {
		end := paragraphEnd(p.runes, start)
		fits, _, err := p.accept(end)
		if err != nil {
			return nil, err
		}
		if !fits {
			for i := start; i < end; {
				next := sentenceEnd(p.runes[:end], i)
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
	if before > 0 && after > target && (before >= target || after-target > target-before) {
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

// When the words fit exactly, a separator after a whole paragraph or sentence
// may straddle the page break. Keep as much on this page as its geometry permits.
func (p *pagePacker) carrySeparator(end int) (bool, error) {
	contentEnd := end
	for contentEnd > p.start && unicode.IsSpace(p.runes[contentEnd-1]) {
		contentEnd--
	}
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

func sentenceEnd(text []rune, start int) int {
	for i := start; i < len(text); i++ {
		if !strings.ContainsRune("。！？!?；;…", text[i]) && !(text[i] == '.' && isFinalPeriod(text, i)) {
			continue
		}
		i++
		// Repeated punctuation plus any closing quotes/brackets belong to
		// the sentence that just ended, as does the space between sentences.
		for i < len(text) && strings.ContainsRune("。.!?！？；;…’”」』）)]】》〉\"'", text[i]) {
			i++
		}
		for i < len(text) && unicode.IsSpace(text[i]) {
			i++
		}
		return i
	}
	return len(text)
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
