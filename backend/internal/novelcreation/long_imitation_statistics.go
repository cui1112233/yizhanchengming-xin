package novelcreation

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

// These local formulas implement the recovered Tln/qln frontend contract.
// They are independent of legacy styleStats and are not native AI scores.
type ImitationStatistics struct {
	AvgSentenceLength    float64 `json:"avgSentenceLength"`
	SentenceLengthStdDev float64 `json:"sentenceLengthStdDev"`
	AvgParagraphLength   int     `json:"avgParagraphLength"`
	ParagraphLengthRange struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"paragraphLengthRange"`
	VocabularyDiversity float64           `json:"vocabularyDiversity"`
	TopPatterns         []json.RawMessage `json:"topPatterns"`
	RhetoricalFeatures  []string          `json:"rhetoricalFeatures"`
	SourceName          string            `json:"sourceName,omitempty"`
}
type ImitationDimension struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Desc  string `json:"desc"`
}

func ImitationStyleDimensions() []ImitationDimension {
	return []ImitationDimension{{"narrativeVoice", "叙事声音与语气", "冷峻/热烈/讽刺/温情/...，附1-2个原文例句"}, {"dialogueStyle", "对话风格", "角色说话的共性特征：句子长短、口头禅倾向、方言痕迹、对话节奏"}, {"sceneDescription", "场景描写特征", "五感偏好、意象选择、描写密度、环境与情绪的关联方式"}, {"transitions", "转折与衔接手法", "场景如何切换、时间跳跃的处理方式、段落间的过渡特征"}, {"pacing", "节奏特征", "长短句分布、段落长度偏好、高潮/舒缓的交替方式"}, {"diction", "词汇偏好", "高频特色用词、比喻/修辞倾向、口语化程度"}, {"emotionalExpression", "情绪表达方式", "直白抒情 vs 动作外化、内心独白的频率和风格"}, {"distinctiveHabits", "独特习惯", "任何值得模仿的个人写作习惯"}}
}
func jsWhitespace(r rune) bool {
	return (r >= 9 && r <= 13) || r == 32 || r == 0xA0 || r == 0x1680 || (r >= 0x2000 && r <= 0x200A) || r == 0x2028 || r == 0x2029 || r == 0x202F || r == 0x205F || r == 0x3000 || r == 0xFEFF
}
func jsTrim(s string) string { return strings.TrimFunc(s, jsWhitespace) }
func jsNonspaceLength(s string) int {
	n := 0
	for _, r := range s {
		if !jsWhitespace(r) {
			if r > 0xffff {
				n += 2
			} else {
				n++
			}
		}
	}
	return n
}
func jsLength(s string) int {
	n := 0
	for _, r := range s {
		n++
		if r > 0xffff {
			n++
		}
	}
	return n
}

// Counting uses exact UTF16 units. JSON storage cannot contain lone surrogates:
// top-pattern display replaces an isolated half while its count key stays exact.
func jsUnitsJSON(units []uint16) json.RawMessage {
	raw, _ := json.Marshal(string(utf16.Decode(units)))
	return raw
}
func imitationChunks(s string) []json.RawMessage {
	u := utf16.Encode([]rune(s))
	out := []json.RawMessage{}
	for i := 0; i < len(u); {
		end := i + 50000
		if end > len(u) {
			end = len(u)
		}
		if end < len(u) && u[end-1] >= 0xD800 && u[end-1] <= 0xDBFF && u[end] >= 0xDC00 && u[end] <= 0xDFFF {
			end--
		}
		out = append(out, jsUnitsJSON(u[i:end]))
		i = end
	}
	return out
}

var imitationSentenceSplit = regexp.MustCompile(`[。！？\n]`)
var imitationParagraphSplit = regexp.MustCompile("\n[\t\n\v\f\r \u00a0\u1680\u2000-\u200a\u2028\u2029\u202f\u205f\u3000\ufeff]*\n")

func ImitationStyleStatistics(text, source string) ImitationStatistics {
	v := ImitationStatistics{TopPatterns: []json.RawMessage{}, RhetoricalFeatures: []string{}, SourceName: source}
	sentences := []string{}
	lengths := []int{}
	sum := 0
	for _, s := range imitationSentenceSplit.Split(text, -1) {
		s = jsTrim(s)
		if s != "" {
			sentences = append(sentences, s)
			n := jsNonspaceLength(s)
			lengths = append(lengths, n)
			sum += n
		}
	}
	if len(lengths) > 0 {
		avg := float64(sum) / float64(len(lengths))
		v.AvgSentenceLength = math.Round(avg*10) / 10
		if len(lengths) > 1 {
			var variance float64
			for _, n := range lengths {
				variance += math.Pow(float64(n)-avg, 2)
			}
			v.SentenceLengthStdDev = math.Round(math.Sqrt(variance/float64(len(lengths)))*10) / 10
		}
	}
	sum = 0
	paragraphs := 0
	for _, p := range imitationParagraphSplit.Split(text, -1) {
		p = jsTrim(p)
		if p == "" {
			continue
		}
		n := jsNonspaceLength(p)
		sum += n
		if paragraphs == 0 || n < v.ParagraphLengthRange.Min {
			v.ParagraphLengthRange.Min = n
		}
		if n > v.ParagraphLengthRange.Max {
			v.ParagraphLengthRange.Max = n
		}
		paragraphs++
	}
	if paragraphs > 0 {
		v.AvgParagraphLength = int(math.Round(float64(sum) / float64(paragraphs)))
	}
	unique := map[rune]bool{}
	denom := 0
	for _, r := range text {
		if jsWhitespace(r) || strings.ContainsRune("，。！？、：；\"'（）【】《》", r) || (r >= '0' && r <= '9') {
			continue
		}
		unique[r] = true
		denom++
		if r > 0xffff {
			denom++
		}
	}
	if denom > 0 {
		v.VocabularyDiversity = math.Round(float64(len(unique))/float64(denom)*1000) / 1000
	}
	type prefix struct {
		units        []uint16
		count, order int
		numeric      bool
		index        uint64
	}
	counts := map[string]*prefix{}
	ordered := []*prefix{}
	for _, s := range sentences {
		u := utf16.Encode([]rune(s))
		if len(u) < 2 {
			continue
		}
		key := fmt.Sprintf("%04x%04x", u[0], u[1])
		p := counts[key]
		if p == nil {
			p = &prefix{units: append([]uint16(nil), u[:2]...), order: len(ordered)}
			k := string(utf16.Decode(u[:2]))
			if n, e := strconv.ParseUint(k, 10, 32); e == nil && n < 4294967295 && strconv.FormatUint(n, 10) == k {
				p.numeric = true
				p.index = n
			}
			counts[key] = p
			ordered = append(ordered, p)
		}
		p.count++
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.count != b.count {
			return a.count > b.count
		}
		if a.numeric != b.numeric {
			return a.numeric
		}
		if a.numeric {
			return a.index < b.index
		}
		return a.order < b.order
	})
	for i, p := range ordered {
		if i >= 5 {
			break
		}
		if p.count >= 3 {
			suffix := utf16.Encode([]rune(fmt.Sprintf("...(%d次)", p.count)))
			v.TopPatterns = append(v.TopPatterns, jsUnitsJSON(append(append([]uint16{}, p.units...), suffix...)))
		}
	}
	// Map surrogate units into a private supplementary range so Go regex's
	// quantifiers count JS UTF16 units rather than Unicode code points.
	units := utf16.Encode([]rune(text))
	var mapped strings.Builder
	for _, u := range units {
		r := rune(u)
		if u >= 0xD800 && u <= 0xDFFF {
			r += 0xF0000
		}
		mapped.WriteRune(r)
	}
	m := mapped.String()
	rhetorical := []struct{ name, pattern string }{{"比喻(像/如/仿佛)", `[像如仿佛似](?:是|同|一般|一样)`}, {"排比", ""}, {"反问", `难道|怎么可能|岂不是|何尝不`}, {"夸张", `天崩地裂|惊天动地|翻天覆地|震耳欲聋`}, {"拟人", `[风雨雪月花树草石](?:在|像|仿佛)[^\n\r\x{2028}\x{2029}]*?(?:笑|哭|叹|呻|吟|怒|舞)`}, {"短句节奏", `[。！？][^。！？]{1,8}[。！？]`}}
	for _, rule := range rhetorical {
		n := 0
		if rule.pattern == "" {
			for i := 0; i < len(units); {
				end := -1
				if strings.ContainsRune("，。；", rune(units[i])) {
					for size := 6; size >= 2; size-- {
						sep := i + 1 + size
						last := sep + 1 + size
						if last > len(units) || !strings.ContainsRune("，。；", rune(units[sep])) {
							continue
						}
						valid := true
						for j := 0; j < size; j++ {
							if strings.ContainsRune("，。；", rune(units[i+1+j])) || units[i+1+j] != units[sep+1+j] {
								valid = false
								break
							}
						}
						if valid {
							end = last
							break
						}
					}
				}
				if end >= 0 {
					n++
					i = end
				} else {
					i++
				}
			}
		} else {
			n = len(regexp.MustCompile(rule.pattern).FindAllStringIndex(m, -1))
		}
		if n >= 2 {
			v.RhetoricalFeatures = append(v.RhetoricalFeatures, fmt.Sprintf("%s(%d处)", rule.name, n))
		}
	}
	return v
}
