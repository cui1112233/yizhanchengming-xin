package batchfactoryv11

import (
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"
)

const ComicSegmentReasonOverlongUnit = "overlong_indivisible_unit"

const (
	ComicMaxPageCount      = 35
	comicPageTextMin       = 90
	comicPageTextTargetMin = 90
	comicPageTextTargetMax = 120
	comicPageTextMax       = 150
	comicPageRepairMax     = 150
)

type ComicSourceBoundary struct {
	Offset  int    `json:"offset"`
	Type    string `json:"type"`
	Preview string `json:"preview,omitempty"`
}

type ComicSourceSegmentDraft struct {
	StartOffset    int    `json:"startOffset"`
	EndOffset      int    `json:"endOffset"`
	SourceText     string `json:"sourceText"`
	CharacterCount int    `json:"characterCount"`
	BoundaryType   string `json:"boundaryType"`
	Reason         string `json:"reason"`
}

func EnumerateComicSourceBoundaries(source string) []ComicSourceBoundary {
	if source == "" {
		return nil
	}
	boundaries := []ComicSourceBoundary{}
	quoteDepth := 0
	previous := rune(0)
	for offset, current := range source {
		next := offset + utf8.RuneLen(current)
		switch current {
		case '"':
			if quoteDepth == 0 {
				quoteDepth++
			} else {
				quoteDepth--
				if quoteDepth == 0 && isComicSentenceEnd(previous) && !nextNonSpaceIsOpeningQuote(source, next) {
					boundaries = appendComicBoundary(boundaries, next, "dialogue")
				}
			}
		case '“', '「', '『', '《':
			quoteDepth++
		case '”', '」', '』', '》':
			if quoteDepth > 0 {
				quoteDepth--
			}
			if quoteDepth == 0 && isComicSentenceEnd(previous) && !nextNonSpaceIsOpeningQuote(source, next) {
				boundaries = appendComicBoundary(boundaries, next, "dialogue")
			}
		default:
			if quoteDepth == 0 && isComicSentenceEnd(current) {
				boundaries = appendComicBoundary(boundaries, next, "sentence")
			}
			if quoteDepth == 0 && (current == '\n' || current == '\r') && !isContinuousDialogueBreak(source, offset, next) {
				boundaries = appendComicBoundary(boundaries, next, "paragraph")
			}
		}
		previous = current
	}
	boundaries = appendComicBoundary(boundaries, len(source), "end")
	sort.SliceStable(boundaries, func(i, j int) bool { return boundaries[i].Offset < boundaries[j].Offset })
	return boundaries
}

func FallbackComicSourceSegments(source string, boundaries []ComicSourceBoundary) ([]ComicSourceSegmentDraft, error) {
	if source == "" {
		return nil, fmt.Errorf("%w: 漫画正文为空", ErrInvalid)
	}
	boundaryByOffset := make(map[int]ComicSourceBoundary, len(boundaries))
	ordered := make([]ComicSourceBoundary, 0, len(boundaries))
	for _, boundary := range boundaries {
		if boundary.Offset <= 0 || boundary.Offset > len(source) {
			continue
		}
		if _, exists := boundaryByOffset[boundary.Offset]; exists {
			continue
		}
		boundaryByOffset[boundary.Offset] = boundary
		ordered = append(ordered, boundary)
	}
	if _, ok := boundaryByOffset[len(source)]; !ok {
		ordered = append(ordered, ComicSourceBoundary{Offset: len(source), Type: "end"})
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Offset < ordered[j].Offset })
	segments := []ComicSourceSegmentDraft{}
	for start := 0; start < len(source); {
		chosen := ComicSourceBoundary{}
		lastWithinMax := ComicSourceBoundary{}
		firstOverMax := ComicSourceBoundary{}
		for _, boundary := range ordered {
			if boundary.Offset <= start {
				continue
			}
			count := comicSourceCharacterCount(source[start:boundary.Offset])
			if count <= comicPageTextMax {
				lastWithinMax = boundary
				if count >= comicPageTextTargetMin && count <= comicPageTextTargetMax {
					chosen = boundary
				}
				continue
			}
			firstOverMax = boundary
			break
		}
		if chosen.Offset == 0 {
			if lastWithinMax.Offset > 0 {
				chosen = lastWithinMax
			} else {
				chosen = firstOverMax
			}
		}
		if chosen.Offset <= start {
			return nil, fmt.Errorf("%w: 正文缺少可用安全边界", ErrInvalid)
		}
		text := source[start:chosen.Offset]
		reason := "safe_boundary"
		if comicSourceCharacterCount(text) > comicPageTextMax {
			reason = ComicSegmentReasonOverlongUnit
		}
		segments = append(segments, ComicSourceSegmentDraft{StartOffset: start, EndOffset: chosen.Offset, SourceText: text, CharacterCount: comicSourceCharacterCount(text), BoundaryType: chosen.Type, Reason: reason})
		start = chosen.Offset
	}
	if len(segments) > 1 {
		last := segments[len(segments)-1]
		previous := &segments[len(segments)-2]
		if last.CharacterCount < comicPageTextMin && previous.CharacterCount+last.CharacterCount <= comicPageTextMax {
			previous.EndOffset = last.EndOffset
			previous.SourceText += last.SourceText
			previous.CharacterCount = comicSourceCharacterCount(previous.SourceText)
			previous.BoundaryType = last.BoundaryType
			segments = segments[:len(segments)-1]
		}
	}
	if err := ValidateComicSourceSegments(source, segments); err != nil {
		return nil, err
	}
	return segments, nil
}

// LimitComicSourceSegments merges already safe, contiguous segments without
// cutting a sentence or paragraph. The comic pipeline uses this as a hard
// page ceiling: later prompt extraction, image generation and collection
// upload all see the same capped list.
func LimitComicSourceSegments(source string, segments []ComicSourceSegmentDraft, maxPages int) ([]ComicSourceSegmentDraft, error) {
	if len(segments) == 0 || maxPages <= 0 {
		return nil, fmt.Errorf("%w: 漫画分页为空", ErrInvalid)
	}
	if len(segments) <= maxPages {
		return segments, nil
	}
	totalCharacters := comicSourceCharacterCount(source)
	target := totalCharacters / maxPages
	if target < 1 {
		target = 1
	}
	result := make([]ComicSourceSegmentDraft, 0, maxPages)
	startIndex := 0
	for page := 0; page < maxPages; page++ {
		remainingPages := maxPages - page - 1
		lastIndex := len(segments) - remainingPages
		endIndex := startIndex + 1
		if remainingPages == 0 {
			endIndex = lastIndex
		} else {
			bestDistance := int(^uint(0) >> 1)
			for candidate := startIndex + 1; candidate <= lastIndex; candidate++ {
				count := comicSourceCharacterCount(source[segments[startIndex].StartOffset:segments[candidate-1].EndOffset])
				distance := count - target
				if distance < 0 {
					distance = -distance
				}
				if distance < bestDistance {
					bestDistance = distance
					endIndex = candidate
				}
				if count >= target {
					break
				}
			}
		}
		startOffset := segments[startIndex].StartOffset
		endOffset := segments[endIndex-1].EndOffset
		merged := ComicSourceSegmentDraft{
			StartOffset:    startOffset,
			EndOffset:      endOffset,
			SourceText:     source[startOffset:endOffset],
			CharacterCount: comicSourceCharacterCount(source[startOffset:endOffset]),
			BoundaryType:   segments[endIndex-1].BoundaryType,
			Reason:         "max_page_count_35",
		}
		result = append(result, merged)
		startIndex = endIndex
	}
	if err := ValidateComicSourceSegments(source, result); err != nil {
		return nil, err
	}
	return result, nil
}

func comicSourceSegmentsForPageCount(source string, boundaries []ComicSourceBoundary, pageCount int) ([]ComicSourceSegmentDraft, error) {
	if source == "" || pageCount <= 0 {
		return nil, fmt.Errorf("%w: 漫画原文分段为空", ErrInvalid)
	}
	ordered := make([]ComicSourceBoundary, 0, len(boundaries))
	seen := map[int]bool{}
	for _, boundary := range boundaries {
		if boundary.Offset <= 0 || boundary.Offset > len(source) || seen[boundary.Offset] {
			continue
		}
		seen[boundary.Offset] = true
		ordered = append(ordered, boundary)
	}
	if !seen[len(source)] {
		ordered = append(ordered, ComicSourceBoundary{Offset: len(source), Type: "end"})
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Offset < ordered[j].Offset })
	if len(ordered) < pageCount {
		return nil, fmt.Errorf("%w: 安全边界不足以生成 %d 页", ErrInvalid, pageCount)
	}

	totalCharacters := comicSourceCharacterCount(source)
	target := totalCharacters / pageCount
	if target < comicPageTextTargetMin {
		target = comicPageTextTargetMin
	} else if target > comicPageTextTargetMax {
		target = comicPageTextTargetMax
	}
	type state struct {
		score int
		prev  int
		ok    bool
	}
	build := func(minimum int) ([]ComicSourceSegmentDraft, bool) {
		positions := make([]int, len(ordered)+1)
		for index, boundary := range ordered {
			positions[index+1] = boundary.Offset
		}
		dp := make([][]state, pageCount+1)
		for page := range dp {
			dp[page] = make([]state, len(positions))
		}
		dp[0][0] = state{ok: true}
		for page := 1; page <= pageCount; page++ {
			for end := 1; end < len(positions); end++ {
				for start := 0; start < end; start++ {
					if !dp[page-1][start].ok {
						continue
					}
					count := comicSourceCharacterCount(source[positions[start]:positions[end]])
					overlongIndivisible := count > comicPageRepairMax && end == start+1
					if count < minimum || (count > comicPageRepairMax && !overlongIndivisible) {
						continue
					}
					penalty := count - target
					if penalty < 0 {
						penalty = -penalty
					}
					if count < comicPageTextMin {
						penalty += 1000
					}
					if overlongIndivisible {
						penalty += 10000
					}
					score := dp[page-1][start].score + penalty
					if !dp[page][end].ok || score < dp[page][end].score {
						dp[page][end] = state{score: score, prev: start, ok: true}
					}
				}
			}
		}
		end := len(positions) - 1
		if !dp[pageCount][end].ok {
			return nil, false
		}
		indexes := make([]int, pageCount+1)
		indexes[pageCount] = end
		for page := pageCount; page > 0; page-- {
			indexes[page-1] = dp[page][indexes[page]].prev
		}
		segments := make([]ComicSourceSegmentDraft, pageCount)
		for page := range segments {
			start, end := positions[indexes[page]], positions[indexes[page+1]]
			text := source[start:end]
			characterCount := comicSourceCharacterCount(text)
			reason := "safe_boundary_page_count"
			if characterCount > comicPageRepairMax {
				reason = ComicSegmentReasonOverlongUnit
			}
			segments[page] = ComicSourceSegmentDraft{StartOffset: start, EndOffset: end, SourceText: text, CharacterCount: characterCount, BoundaryType: ordered[indexes[page+1]-1].Type, Reason: reason}
		}
		return segments, true
	}
	segments, ok := build(90)
	if !ok {
		segments, ok = build(1)
	}
	if !ok {
		return nil, fmt.Errorf("%w: 无法按安全边界生成 %d 页", ErrInvalid, pageCount)
	}
	if err := ValidateComicSourceSegments(source, segments); err != nil {
		return nil, err
	}
	return segments, nil
}

func ValidateComicSourceSegments(source string, segments []ComicSourceSegmentDraft) error {
	if source == "" || len(segments) == 0 {
		return fmt.Errorf("%w: 漫画原文分段为空", ErrInvalid)
	}
	next := 0
	for index, segment := range segments {
		if segment.StartOffset != next || segment.EndOffset <= segment.StartOffset || segment.EndOffset > len(source) {
			return fmt.Errorf("%w: 第 %d 段原文范围不连续", ErrInvalid, index+1)
		}
		if source[segment.StartOffset:segment.EndOffset] != segment.SourceText {
			return fmt.Errorf("%w: 第 %d 段原文与范围不一致", ErrInvalid, index+1)
		}
		next = segment.EndOffset
	}
	if next != len(source) {
		return fmt.Errorf("%w: 漫画原文没有完整覆盖", ErrInvalid)
	}
	return nil
}

func comicSourceCharacterCount(value string) int {
	count := 0
	for _, current := range value {
		if !unicode.IsSpace(current) {
			count++
		}
	}
	return count
}

func isComicSentenceEnd(value rune) bool {
	switch value {
	case '。', '！', '？', '!', '?', '；', ';':
		return true
	default:
		return false
	}
}

func isComicOpeningQuote(value rune) bool {
	switch value {
	case '"', '“', '「', '『':
		return true
	default:
		return false
	}
}

func nextNonSpaceIsOpeningQuote(source string, offset int) bool {
	for _, current := range source[offset:] {
		if unicode.IsSpace(current) {
			continue
		}
		return isComicOpeningQuote(current)
	}
	return false
}

func isContinuousDialogueBreak(source string, currentOffset, nextOffset int) bool {
	before := rune(0)
	for _, current := range source[:currentOffset] {
		if !unicode.IsSpace(current) {
			before = current
		}
	}
	if before != '"' && before != '”' && before != '」' && before != '』' {
		return false
	}
	return nextNonSpaceIsOpeningQuote(source, nextOffset)
}

func appendComicBoundary(boundaries []ComicSourceBoundary, offset int, kind string) []ComicSourceBoundary {
	if len(boundaries) > 0 && boundaries[len(boundaries)-1].Offset == offset {
		if boundaries[len(boundaries)-1].Type != "paragraph" {
			boundaries[len(boundaries)-1].Type = kind
		}
		return boundaries
	}
	return append(boundaries, ComicSourceBoundary{Offset: offset, Type: kind})
}
