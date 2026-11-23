package batchfactorytext

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Each changed range is between characters that every shortest alignment
// matches to the same partner. Offsets on both sides are normalized runes.
type editRange struct {
	start, end, replacementStart, replacementEnd int
	ambiguous                                    bool // At least one interior character could also be matched.
}
type runeAnchor struct {
	at       int
	possible bool
}

func mergeConflict() error {
	return fmt.Errorf("%w：这处修改在作者稿中有多个可能位置，请在编辑内容中核对后修改", ErrConflict)
}

// Banded LCS distances use insertion/deletion costs. A discovered distance <=
// band proves that no shortest alignment can leave the band. Limits bound both
// memory and total work across retries, independent of the number of repeats.
func editRanges(base, next []rune, remaining *int) ([]editRange, error) {
	n, m := len(base), len(next)
	if n == 0 || m == 0 {
		return []editRange{{end: n, replacementEnd: m}}, nil
	}

	maxSize := max(n, m)
	band := min(maxSize, max(2, n-m, m-n))
	for {
		width := 2*band + 3
		if band > 4096 || n+1 > *remaining/width {
			return nil, fmt.Errorf("%w：两份稿件差异过多，无法可靠定位，请在编辑内容中修改", ErrConflict)
		}
		size := (n + 1) * width
		*remaining -= size
		f := make([]uint16, size)
		const inf uint16 = 60000
		for i := range f {
			f[i] = inf
		}
		for i := 0; i <= n; i++ {
			for j := max(0, i-band); j <= min(m, i+band); j++ {
				at := i*width + j - i + band + 1
				if i == 0 && j == 0 {
					f[at] = 0
					continue
				}
				d := inf
				if i > 0 {
					d = min(d, f[at-width+1]+1)
				}
				if j > 0 {
					d = min(d, f[at-1]+1)
				}
				if i > 0 && j > 0 && base[i-1] == next[j-1] {
					d = min(d, f[at-width])
				}
				f[at] = d
			}
		}
		distance := f[n*width+m-n+band+1]
		if int(distance) <= band || band == maxSize {
			return rangesFromOptimalPaths(base, next, f, width, band, distance), nil
		}
		band = min(maxSize, band*2)
	}
}

func rangesFromOptimalPaths(base, next []rune, forward []uint16, width, band int, distance uint16) []editRange {
	n, m := len(base), len(next)
	anchors := make([]runeAnchor, n)
	below, row := make([]uint16, width), make([]uint16, width)
	const inf uint16 = 60000
	for k := range below {
		below[k] = inf
	}
	for i := n; i >= 0; i-- {
		for k := range row {
			row[k] = inf
		}
		match, canDelete := -1, false
		for j := min(m, i+band); j >= max(0, i-band); j-- {
			k := j - i + band + 1
			if i == n && j == m {
				row[k] = 0
				continue
			}
			cost := int(forward[i*width+k])
			if j < m {
				row[k] = min(row[k], row[k+1]+1)
			}
			if i < n {
				row[k] = min(row[k], below[k-1]+1)
				if cost+1+int(below[k-1]) == int(distance) {
					canDelete = true
				}
				if j < m && base[i] == next[j] {
					row[k] = min(row[k], below[k])
					if cost+int(below[k]) == int(distance) {
						if match == -1 {
							match = j
						} else {
							match = -2
						}
					}
				}
			}
		}
		if i < n {
			anchors[i] = runeAnchor{at: -1, possible: match != -1}
			if !canDelete && match >= 0 {
				anchors[i].at = match
			}
		}
		below, row = row, below
	}
	var out []editRange
	b, r := 0, 0
	uncertain := false
	for i, a := range anchors {
		if a.at < 0 {
			uncertain = uncertain || a.possible
			continue
		}
		if b != i || r != a.at {
			out = append(out, editRange{b, i, r, a.at, uncertain})
		}
		b, r, uncertain = i+1, a.at+1, false
	}
	if b != n || r != m {
		out = append(out, editRange{b, n, r, m, uncertain})
	}
	return out
}

// A trace supplies trustworthy anchors only if replaying its ordered changes
// exactly recreates the displayed version. System records for overlapping
// replacements may not replay; fall back to the bounded whole-text alignment.
func traceRanges(author, shown string, raw json.RawMessage) ([]editRange, bool) {
	if len(raw) == 0 || len(raw) > 4<<20 {
		return nil, false
	}
	var changes []struct{ Before, After string }
	if err := json.Unmarshal(raw, &changes); err != nil || len(changes) == 0 || len(changes) > 4096 {
		return nil, false
	}
	var spans []editRange
	a, s := 0, 0
	for _, c := range changes {
		before, after := NormalizeLineEndings(c.Before), NormalizeLineEndings(c.After)
		if before == "" {
			return nil, false
		}
		at := strings.Index(author, before)
		if at < 0 || !strings.HasPrefix(shown, author[:at]) {
			return nil, false
		}
		unchanged := utf8.RuneCountInString(author[:at])
		a += unchanged
		s += unchanged
		shown = shown[at:]
		if !strings.HasPrefix(shown, after) {
			return nil, false
		}
		ae, se := a+utf8.RuneCountInString(before), s+utf8.RuneCountInString(after)
		if before != after {
			spans = append(spans, editRange{start: s, end: se, replacementStart: a, replacementEnd: ae})
		}
		author, shown = author[at+len(before):], shown[len(after):]
		s, a = se, ae
	}
	return spans, author == shown
}

// Trace spans outside the user's edit can remain opaque. Align the inside of a
// change only if an edit endpoint is strictly inside that span.
func refineTouchedRanges(shown, author []rune, spans []editRange, start int, user []editRange, remaining *int) ([]editRange, error) {
	var out []editRange
	u := 0
	for _, span := range spans {
		for u < len(user) && start+user[u].end <= span.start {
			u++
		}
		inside := false
		for i := u; i < len(user) && start+user[i].start < span.end; i++ {
			left, right := start+user[i].start, start+user[i].end
			if left > span.start || (right > span.start && right < span.end) {
				inside = true
				break
			}
		}
		if !inside {
			out = append(out, span)
			continue
		}
		local, err := editRanges(shown[span.start:span.end], author[span.replacementStart:span.replacementEnd], remaining)
		if err != nil {
			return nil, err
		}
		for _, part := range local {
			part.start += span.start
			part.end += span.start
			part.replacementStart += span.replacementStart
			part.replacementEnd += span.replacementStart
			out = append(out, part)
		}
	}
	return out, nil
}

type authorBoundary struct {
	left, right int
	known       bool
}

func authorBoundaries(shown int, changes []editRange) []authorBoundary {
	edges := make([]authorBoundary, shown+1)
	put := func(offset, pos int) {
		e := &edges[offset]
		if !e.known {
			*e = authorBoundary{pos, pos, true}
			return
		}
		e.left = min(e.left, pos)
		e.right = max(e.right, pos)
	}
	s, a := 0, 0
	for _, h := range changes {
		for ; s <= h.start; s++ {
			put(s, a)
			a++
		}
		s, a = h.end, h.replacementEnd
		put(s, a)
	}
	for ; s <= shown; s++ {
		put(s, a)
		a++
	}
	return edges
}

// MergeAuthorPageEdit projects the actual changes on a displayed page onto the
// author draft. Unedited system substitutions remain exclusive to production.
func MergeAuthorPageEdit(author string, v Version, start, end int, edited string) (string, error) {
	shown := v.ProductionText
	source := []rune(NormalizeLineEndings(author))
	base := []rune(NormalizeLineEndings(shown))
	page := []rune(NormalizeLineEndings(edited))
	if start < 0 || end > len(base) || start > end {
		return "", ErrConflict
	}
	sameSource := string(source) == string(base)
	hasCR := strings.ContainsRune(author, '\r')
	if sameSource && !hasCR {
		return string(source[:start]) + string(page) + string(source[end:]), nil
	}
	remaining := 8_000_000
	user, err := editRanges(base[start:end], page, &remaining)
	if err != nil {
		return "", err
	}
	var machine []editRange
	if !sameSource {
		var verified bool
		machine, verified = traceRanges(string(source), string(base), v.ChangeTrace)
		if verified {
			machine, err = refineTouchedRanges(base, source, machine, start, user, &remaining)
		} else {
			machine, err = editRanges(base, source, &remaining)
		}
		if err != nil {
			return "", err
		}
	}
	// The renderer normalizes line endings. Preserve untouched bytes from
	// imported author drafts when translating rune offsets back to storage.
	var rawOffsets []int
	if hasCR {
		rawOffsets = make([]int, 0, len(source)+1)
		for pos := 0; pos < len(author); {
			rawOffsets = append(rawOffsets, pos)
			r, n := utf8.DecodeRuneInString(author[pos:])
			pos += n
			if r == '\r' && pos < len(author) && author[pos] == '\n' {
				pos++
			}
		}
		rawOffsets = append(rawOffsets, len(author))
	}
	sourceSlice := func(lo, hi int) string {
		if hasCR {
			return author[rawOffsets[lo]:rawOffsets[hi]]
		}
		return string(source[lo:hi])
	}
	edges := authorBoundaries(len(base), machine)
	var out strings.Builder
	cursor := 0
	for _, h := range user {
		lo, hi := start+h.start, start+h.end
		l, r := edges[lo], edges[hi]
		if !l.known || !r.known || (lo == hi && l.left != l.right) {
			return "", mergeConflict()
		}
		from, to := l.right, r.left
		if from < cursor || to < from || (h.ambiguous && string(source[from:to]) != string(base[lo:hi])) {
			return "", mergeConflict()
		}
		out.WriteString(sourceSlice(cursor, from))
		out.WriteString(string(page[h.replacementStart:h.replacementEnd]))
		cursor = to
	}
	out.WriteString(sourceSlice(cursor, len(source)))
	return out.String(), nil
}
