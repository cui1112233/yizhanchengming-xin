package novelcreation

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Deterministic rules independently implemented from the retained native
// contract, docs/reverse-engineering/2026-11-17-luoshui-postwrite-validation.md.
//
//go:embed rules/postwrite.json
var postWriteFiles embed.FS

type LocalIssue struct {
	RuleID   string `json:"rule_id"`
	Severity string `json:"severity"`
	Category string `json:"category"`
	Message  string `json:"message"`
}
type LocalStats struct {
	TotalChars     int `json:"total_chars"`
	ParagraphCount int `json:"paragraph_count"`
	ErrorCount     int `json:"error_count"`
	WarningCount   int `json:"warning_count"`
}
type PostWriteResult struct {
	Passed bool         `json:"passed"`
	Issues []LocalIssue `json:"issues"`
	Stats  LocalStats   `json:"stats"`
}
type PostWriteInput struct {
	Title                                          string
	ExistingTitles, PreviousChapters, FatigueWords []string
	FirstPerson                                    bool
	Protagonist                                    string
}
type postWriteRule struct{ ID, Severity, Category, Message string }
type postWriteCorpus struct {
	Constants map[string]json.RawMessage
	Rules     []postWriteRule
}

var localCorpus = func() postWriteCorpus {
	var c postWriteCorpus
	b, e := postWriteFiles.ReadFile("rules/postwrite.json")
	if e != nil {
		panic(e)
	}
	if e = json.Unmarshal(b, &c); e != nil {
		panic(e)
	}
	return c
}()

func ruleWords(key string) []string {
	var out []string
	if e := json.Unmarshal(localCorpus.Constants[key], &out); e != nil {
		panic(e)
	}
	return out
}
func rulePattern(key string) *regexp.Regexp {
	var p string
	raw := localCorpus.Constants[key]
	if e := json.Unmarshal(raw, &p); e != nil {
		var v struct{ Pattern string }
		if e = json.Unmarshal(raw, &v); e != nil {
			panic(e)
		}
		p = v.Pattern
	}
	if key == "CHAPTER_NUMBER_PATTERN" {
		p = "(?i)" + strings.ReplaceAll(strings.ReplaceAll(p, `\s`, `[\p{Z}\t\n\r\f\v\x{0085}\x1c-\x1f]`), `\d`, `\p{Nd}`)
	}
	return regexp.MustCompile(p)
}
func stripLocalTitle(s string) string {
	if strings.HasPrefix(s, "#") {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			return s[i+1:]
		}
		return ""
	}
	return s
}
func localCharCount(s string) int {
	return len([]rune(strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(s)))
}
func pythonSpace(c rune) bool { return unicode.IsSpace(c) || (c >= 0x1c && c <= 0x1f) }
func localParagraphs(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, "\n") {
		if p = strings.TrimFunc(p, pythonSpace); p != "" {
			out = append(out, p)
		}
	}
	return out
}
func needsLengthNormalization(s string, target int) bool {
	if target <= 0 {
		return false
	}
	count := localCharCount(stripLocalTitle(s))
	delta := count - target
	if delta < 0 {
		delta = -delta
	}
	return int64(delta)*100 > int64(target)*45
}
func takeWords(v []string, n int) string {
	if len(v) > n {
		v = v[:n]
	}
	return strings.Join(v, "、")
}
func matchedWords(body string, words []string) []string {
	out := []string{}
	for _, w := range words {
		if strings.Contains(body, w) {
			out = append(out, w)
		}
	}
	return out
}

func ValidatePostWrite(content string, in PostWriteInput) PostWriteResult {
	body := stripLocalTitle(content)
	paras := localParagraphs(body)
	r := PostWriteResult{Passed: true, Issues: []LocalIssue{}, Stats: LocalStats{TotalChars: localCharCount(body), ParagraphCount: len(paras)}}
	add := func(index int, args ...any) {
		d := localCorpus.Rules[index]
		message := d.Message
		if len(args) > 0 {
			for i, value := range args {
				args[i] = fmt.Sprint(value)
			}
			message = fmt.Sprintf(message, args...)
		}
		r.Issues = append(r.Issues, LocalIssue{d.ID, d.Severity, d.Category, message})
		if d.Severity == "error" {
			r.Passed = false
			r.Stats.ErrorCount++
		}
		if d.Severity == "warning" {
			r.Stats.WarningCount++
		}
	}
	if n := len(rulePattern("NOT_BUT_PATTERN").FindAllString(body, -1)); n > 0 {
		add(0, n)
	}
	if n := strings.Count(body, "——"); n > 0 {
		add(1, n)
	}
	if found := matchedWords(body, ruleWords("ANALYSIS_TERMS")); len(found) > 0 {
		add(2, takeWords(found, 5))
	}
	if n := len(rulePattern("CHAPTER_NUMBER_PATTERN").FindAllString(body, -1)); n > 0 {
		add(3, n)
	}
	normalize := func(s string) string {
		return strings.Map(func(c rune) rune {
			if pythonSpace(c) || strings.ContainsRune("，。！？、：；''（）", c) {
				return -1
			}
			return c
		}, s)
	}
	if in.Title != "" {
		for _, old := range in.ExistingTitles {
			if normalize(in.Title) == normalize(old) {
				add(4, in.Title)
				break
			}
		}
	}
	if r.Stats.TotalChars >= 500 {
		n := 0
		for _, w := range ruleWords("SURPRISE_MARKERS_ZH") {
			n += strings.Count(body, w)
		}
		allowed := r.Stats.TotalChars / 3000
		if allowed < 1 {
			allowed = 1
		}
		if n > allowed {
			add(5, n, 3000, allowed)
		}
	}
	fatigue := []string{}
	for _, w := range in.FatigueWords {
		if w != "" {
			if n := strings.Count(body, w); n > 1 {
				fatigue = append(fatigue, fmt.Sprintf("「%s」×%d", w, n))
			}
		}
	}
	if len(fatigue) > 0 {
		add(6, takeWords(fatigue, 5))
	}
	for _, p := range ruleWords("META_NARRATION_PATTERNS") {
		if regexp.MustCompile(p).MatchString(body) {
			add(7)
			break
		}
	}
	if found := matchedWords(body, ruleWords("PREACHING_WORDS")); len(found) >= 2 {
		add(8, strings.Join(found, "、"))
	}
	if n := len(rulePattern("COLLECTIVE_SHOCK").FindAllString(body, -1)); n >= 2 {
		add(9, n)
	}
	sentences := []string{}
	for _, s := range rulePattern("LE_SENTENCE_SPLIT").Split(body, -1) {
		if s = strings.TrimFunc(s, pythonSpace); s != "" {
			sentences = append(sentences, s)
		}
	}
	consecutive, longest := 0, 0
	for _, s := range sentences {
		if strings.HasSuffix(s, "了") || strings.Contains(s, "了，") || strings.Contains(s, "了。") {
			consecutive++
		} else {
			consecutive = 0
		}
		if consecutive > longest {
			longest = consecutive
		}
	}
	if len(sentences) >= 6 && longest >= 6 {
		add(10, longest)
	}
	overflow := 0
	for _, p := range paras {
		if localCharCount(p) > 300 {
			overflow++
		}
	}
	if overflow > 0 {
		add(11, overflow, 300)
	}
	clean := func(s string) []rune { return []rune(strings.NewReplacer("\n", "", " ", "").Replace(s)) }
	current := clean(body)
	if len(in.PreviousChapters) > 0 && len(current) >= 12 {
		previous := in.PreviousChapters
		if len(previous) > 3 {
			previous = previous[len(previous)-3:]
		}
		joined := []rune{}
		for _, s := range previous {
			joined = append(joined, clean(s)...)
		}
		old := map[string]bool{}
		for i := 0; i+6 <= len(joined); i++ {
			old[string(joined[i:i+6])] = true
		}
		counts := map[string]int{}
		order := []string{}
		for i := 0; i+6 <= len(current); i++ {
			v := string(current[i : i+6])
			if counts[v] == 0 {
				order = append(order, v)
			}
			counts[v]++
		}
		repeated := []string{}
		for _, v := range order {
			first := []rune(v)[0]
			if counts[v] >= 2 && old[v] && first >= '一' && first <= '鿿' {
				repeated = append(repeated, v)
			}
		}
		if len(repeated) >= 3 {
			add(12, len(repeated), 6, takeWords(repeated, 5))
		}
	}
	if len(paras) > 15 {
		dialogue := 0
		for _, p := range paras {
			v := []rune(p)
			end := len(v)
			if end > 3 {
				end = 3
			}
			if strings.HasPrefix(p, "“") || strings.HasPrefix(p, `"`) || strings.Contains(string(v[:end]), "「") {
				dialogue++
			}
		}
		ratio := float64(dialogue) / float64(len(paras))
		if ratio < .05 {
			add(13, fmt.Sprintf("%.0f%%", ratio*100))
		}
	}
	if in.FirstPerson && r.Stats.TotalChars >= 800 && in.Protagonist != "" {
		wo := strings.Count(body, "我")
		third := len(rulePattern("THIRD_PERSON_INNER").FindAllString(body, -1))
		names := strings.Count(body, in.Protagonist)
		if wo < 12 && third >= 3 && names >= 6 && names > wo {
			add(14, third, in.Protagonist, names, wo)
		}
	}
	return r
}

// These are transparent CM measurements, not a recovered original AI-tell
// classifier or an assertion that a paragraph was AI-written.
func localExpressionMetrics(content string) map[string]any {
	paras := localParagraphs(stripLocalTitle(content))
	lengths := map[int]int{}
	for _, p := range paras {
		lengths[localCharCount(p)]++
	}
	counts := map[string]int{}
	for _, w := range ruleWords("PREACHING_WORDS") {
		if n := strings.Count(content, w); n > 0 {
			counts[w] = n
		}
	}
	return map[string]any{"implementation": "cm_local_metrics_v1", "original_algorithm_verified": false, "paragraph_count": len(paras), "paragraph_length_histogram": lengths, "formula_phrase_counts": counts, "note": "本地表达指标供人工判断，不改写正文；原版AI味判定算法尚未对齐"}
}
