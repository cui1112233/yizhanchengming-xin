package novelcreation

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

// 洛水前端文风统计指纹 Tln 的等价实现。
// 证据与公式见 docs/reverse-engineering/2026-11-17-luoshui-novel-style-statistics.md。
// 关键约束：长度一律按 JavaScript UTF16 码元计算，不能用 Go 字节数或 rune 数替代。
type ParagraphRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type StyleFingerprint struct {
	AvgSentenceLength    float64        `json:"avgSentenceLength"`
	SentenceLengthStdDev float64        `json:"sentenceLengthStdDev"`
	AvgParagraphLength   float64        `json:"avgParagraphLength"`
	ParagraphLengthRange ParagraphRange `json:"paragraphLengthRange"`
	VocabularyDiversity  float64        `json:"vocabularyDiversity"`
	TopPatterns          []string       `json:"topPatterns"`
	RhetoricalFeatures   []string       `json:"rhetoricalFeatures"`
	SourceName           string         `json:"sourceName,omitempty"`
}

func roundTo(value float64, scale float64) float64 {
	return math.Round(value*scale) / scale
}

var whitespaceOnly = regexp.MustCompile(`\s+`)

// styleSentences 按原版切分：中文句末符和换行，不含英文标点、省略号和分号。
func styleSentences(text string) []string {
	out := []string{}
	for _, part := range regexp.MustCompile(`[。！？\n]`).Split(text, -1) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// styleParagraphs 按空行切段。
func styleParagraphs(text string) []string {
	out := []string{}
	for _, part := range regexp.MustCompile(`\n\s*\n`).Split(text, -1) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

var (
	rhetoricSimile          = regexp.MustCompile(`[像如仿佛似](?:是|同|一般|一样)`)
	rhetoricRhetoric        = regexp.MustCompile(`难道|怎么可能|岂不是|何尝不`)
	rhetoricExaggeration    = regexp.MustCompile(`天崩地裂|惊天动地|翻天覆地|震耳欲聋`)
	rhetoricPersonification = regexp.MustCompile(`[风雨雪月花树草石](?:在|像|仿佛).*?(?:笑|哭|叹|呻|吟|怒|舞)`)
	rhetoricShortRhythm     = regexp.MustCompile(`[。！？][^。！？]{1,8}[。！？]`)
)

func isStylePunctuation(r rune) bool {
	return r == '，' || r == '。' || r == '；'
}

// countParallelism 等价于 /[，。；]([^，。；]{2,6})[，。；]\1/g。
// Go 的 RE2 不支持反向引用，因此按原正则的贪婪长度 6→2 与全局非重叠规则手写。
func countParallelism(text string) int {
	runes := []rune(text)
	count := 0
	for i := 0; i < len(runes); i++ {
		if !isStylePunctuation(runes[i]) {
			continue
		}
		for length := 6; length >= 2; length-- {
			groupEnd := i + 1 + length
			repeatEnd := groupEnd + 1 + length
			if repeatEnd > len(runes) {
				continue
			}
			if !isStylePunctuation(runes[groupEnd]) {
				continue
			}
			group := runes[i+1 : groupEnd]
			bad := false
			for _, r := range group {
				if isStylePunctuation(r) {
					bad = true
					break
				}
			}
			if bad {
				continue
			}
			if string(runes[groupEnd+1:repeatEnd]) != string(group) {
				continue
			}
			count++
			i = repeatEnd - 1
			break
		}
	}
	return count
}

func rhetoricalFeatures(text string) []string {
	entries := []struct {
		name  string
		count int
	}{
		{"比喻(像/如/仿佛)", len(rhetoricSimile.FindAllStringIndex(text, -1))},
		{"排比", countParallelism(text)},
		{"反问", len(rhetoricRhetoric.FindAllStringIndex(text, -1))},
		{"夸张", len(rhetoricExaggeration.FindAllStringIndex(text, -1))},
		{"拟人", len(rhetoricPersonification.FindAllStringIndex(text, -1))},
		{"短句节奏", len(rhetoricShortRhythm.FindAllStringIndex(text, -1))},
	}
	out := []string{}
	for _, entry := range entries {
		if entry.count >= 2 {
			out = append(out, entry.name+"("+itoa(entry.count)+"处)")
		}
	}
	return out
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// jsPrefix 按 JavaScript slice(0, n) 的 UTF16 语义取前缀。
func jsPrefix(value string, units int) string {
	used := 0
	out := []rune{}
	for _, r := range value {
		width := 1
		if r > 0xFFFF {
			width = 2
		}
		if used+width > units {
			break
		}
		used += width
		out = append(out, r)
	}
	return string(out)
}

func topSentencePatterns(sentences []string) []string {
	counts := map[string]int{}
	order := []string{}
	for _, sentence := range sentences {
		if jsLength(sentence) < 2 {
			continue
		}
		prefix := jsPrefix(sentence, 2)
		if _, ok := counts[prefix]; !ok {
			order = append(order, prefix)
		}
		counts[prefix]++
	}
	// 稳定排序：次数降序，次数相同保持首次出现顺序。
	sort.SliceStable(order, func(i, j int) bool { return counts[order[i]] > counts[order[j]] })
	out := []string{}
	for _, prefix := range order {
		if len(out) >= 5 {
			break
		}
		if counts[prefix] >= 3 {
			out = append(out, prefix+"...("+itoa(counts[prefix])+"次)")
		}
	}
	return out
}

var vocabularyStrip = regexp.MustCompile(`[\s\n\r，。！？、：；""''（）【】《》\d]`)

func vocabularyDiversity(text string) float64 {
	stripped := vocabularyStrip.ReplaceAllString(text, "")
	if stripped == "" {
		return 0
	}
	codepoints := map[rune]bool{}
	for _, r := range stripped {
		codepoints[r] = true
	}
	return roundTo(float64(len(codepoints))/float64(jsLength(stripped)), 1000)
}

// buildStyleFingerprint 按原版 Tln 输出全部统计字段。
func buildStyleFingerprint(text, sourceName string) StyleFingerprint {
	f := StyleFingerprint{
		ParagraphLengthRange: ParagraphRange{},
		TopPatterns:          []string{},
		RhetoricalFeatures:   rhetoricalFeatures(text),
		SourceName:           strings.TrimSpace(sourceName),
	}
	sentences := styleSentences(text)
	paragraphs := styleParagraphs(text)
	lengths := make([]int, 0, len(sentences))
	total := 0
	for _, sentence := range sentences {
		value := jsLength(whitespaceOnly.ReplaceAllString(sentence, ""))
		lengths = append(lengths, value)
		total += value
	}
	if len(lengths) > 0 {
		avg := float64(total) / float64(len(lengths))
		f.AvgSentenceLength = roundTo(avg, 10)
		if len(lengths) > 1 {
			sum := 0.0
			for _, value := range lengths {
				diff := float64(value) - avg
				sum += diff * diff
			}
			f.SentenceLengthStdDev = roundTo(math.Sqrt(sum/float64(len(lengths))), 10)
		}
	}
	if len(paragraphs) > 0 {
		paragraphLengths := make([]int, 0, len(paragraphs))
		sum := 0
		for _, paragraph := range paragraphs {
			value := jsLength(whitespaceOnly.ReplaceAllString(paragraph, ""))
			paragraphLengths = append(paragraphLengths, value)
			sum += value
		}
		f.AvgParagraphLength = math.Round(float64(sum) / float64(len(paragraphLengths)))
		minValue, maxValue := paragraphLengths[0], paragraphLengths[0]
		for _, value := range paragraphLengths {
			if value < minValue {
				minValue = value
			}
			if value > maxValue {
				maxValue = value
			}
		}
		f.ParagraphLengthRange = ParagraphRange{Min: minValue, Max: maxValue}
	}
	f.VocabularyDiversity = vocabularyDiversity(text)
	f.TopPatterns = topSentencePatterns(sentences)
	return f
}
