package novelcreation

import (
	"sort"
	"strings"
	"unicode"
)

// 洛水 novels.services.memory_store.retrieve_relevant_chapters 的服务端等价实现。
// 原版使用 SQLite FTS5 + 关键词重叠 + 近因 + 邻近度打分；CM 的章节正文已经在
// MySQL，因此直接在内存中做同样的排序，不引入第二套存储。
const (
	retrievalLimit        = 3
	retrievalTokenLimit   = 24
	retrievalExcerptRunes = 500
	retrievalMaxTokens    = 10
)

type RetrievedChapter struct {
	Number  int     `json:"number"`
	Title   string  `json:"title"`
	Excerpt string  `json:"excerpt"`
	Score   float64 `json:"score"`
}

// retrievalTokens 提取用于重叠打分的关键词。
// 中文按 2 字滑窗，英文按 3+ 字母单词，数字按整数；小写、去重保序。
// 原版把整段中文当作一个 token，实际几乎无法命中其他章节；这里保留同样
// 的打分公式，但用 2 字窗口让中文召回真正生效。
func retrievalTokens(text string) []string {
	tokens := []string{}
	seen := map[string]bool{}
	add := func(value string) {
		if len(tokens) >= retrievalTokenLimit || value == "" || seen[value] {
			return
		}
		seen[value] = true
		tokens = append(tokens, value)
	}
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if !unicode.Is(unicode.Han, runes[i]) {
			continue
		}
		if i+1 < len(runes) && unicode.Is(unicode.Han, runes[i+1]) {
			add(string(runes[i : i+2]))
		}
	}
	var word []rune
	flush := func() {
		if len(word) >= 3 {
			add(strings.ToLower(string(word)))
		}
		word = word[:0]
	}
	for _, r := range runes {
		if unicode.IsLetter(r) && !unicode.Is(unicode.Han, r) {
			word = append(word, r)
			continue
		}
		flush()
	}
	flush()
	return tokens
}

func chapterExcerpt(content string) string {
	body := stripLocalTitle(content)
	body = strings.TrimSpace(body)
	runes := []rune(body)
	if len(runes) > retrievalExcerptRunes {
		runes = runes[:retrievalExcerptRunes]
	}
	return string(runes)
}

// retrieveRelevantChapters 按洛水公式打分：重叠词×2 + 近因×1 + 邻近度×1.2。
// 只考虑目标章之前已保存的章节，返回按章号升序的命中片段。
func retrieveRelevantChapters(chapters []Chapter, query string, currentChapter, limit int) []RetrievedChapter {
	if limit <= 0 {
		limit = retrievalLimit
	}
	ordered := make([]Chapter, 0, len(chapters))
	for _, chapter := range chapters {
		if currentChapter > 0 && chapter.Number >= currentChapter {
			continue
		}
		ordered = append(ordered, chapter)
	}
	if len(ordered) == 0 {
		return nil
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })
	tokens := retrievalTokens(query)
	if len(tokens) > retrievalMaxTokens {
		tokens = tokens[:retrievalMaxTokens]
	}
	maxNumber := ordered[len(ordered)-1].Number
	type scored struct {
		chapter RetrievedChapter
	}
	hits := []scored{}
	for _, chapter := range ordered {
		blob := strings.ToLower(chapter.Title + chapter.Content)
		overlap := 0
		for _, token := range tokens {
			if strings.Contains(blob, token) {
				overlap++
			}
		}
		recency := 0.0
		if maxNumber > 0 {
			recency = float64(chapter.Number) / float64(maxNumber)
		}
		distance := maxNumber - chapter.Number
		proximity := 1 / (1 + float64(distance))
		score := float64(overlap)*2 + recency + proximity*1.2
		if overlap == 0 && len(tokens) > 0 {
			// 没有关键词命中时只保留最近章节，避免把无关旧章塞进上下文。
			if distance > retrievalLimit {
				continue
			}
		}
		hits = append(hits, scored{RetrievedChapter{Number: chapter.Number, Title: chapter.Title, Excerpt: chapterExcerpt(chapter.Content), Score: score}})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].chapter.Score == hits[j].chapter.Score {
			return hits[i].chapter.Number > hits[j].chapter.Number
		}
		return hits[i].chapter.Score > hits[j].chapter.Score
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	out := make([]RetrievedChapter, 0, len(hits))
	for _, hit := range hits {
		out = append(out, hit.chapter)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out
}

// retrievalQuery 构造检索词：用户指令优先，没有指令时退到上一章标题与摘要。
func retrievalQuery(userMessage string, chapters []Chapter) string {
	if text := strings.TrimSpace(userMessage); text != "" {
		return text
	}
	if len(chapters) == 0 {
		return ""
	}
	latest := chapters[len(chapters)-1]
	return latest.Title + " " + chapterExcerpt(latest.Content)
}
