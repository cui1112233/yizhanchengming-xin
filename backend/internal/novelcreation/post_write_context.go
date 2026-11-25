package novelcreation

import (
	"sort"
	"strings"
)

// 本文件冻结本地后写校验所需的跨章资料。参考洛水 novels.post_write_validator
// 的输入合同：历史标题、前 3 章正文、疲劳词、叙事人称与主角名。
//
// 该结构只服务零模型成本的确定性检查，不进入任何 LLM 请求，因此向新任务的
// PromptSnapshot 增加字段不会改变已有任务的请求指纹。
const (
	postWriteChapterChars = 12000
	postWriteTotalChars   = 36000
	postWriteChapterLimit = 3
)

// PostWriteContext 是一次创作运行可用的真实后写检查资料。
// 缺失项必须由调用方显式标注为未检查，不能用截断结尾冒充全文。
type PostWriteContext struct {
	ExistingTitles    []string `json:"existing_titles"`
	PreviousFullText  []string `json:"previous_full_text"`
	FatigueWords      []string `json:"fatigue_words"`
	FirstPerson       bool     `json:"first_person"`
	FirstPersonSource string   `json:"first_person_source"`
	Protagonist       string   `json:"protagonist"`
	Genre             string   `json:"genre"`
	SourceChapters    []int    `json:"source_chapters"`
}

// 洛水 GenreProfile._COMMON_FATIGUE。
var commonFatigueWords = []string{"忍不住", "不由得", "情不自禁", "下意识", "缓缓", "微微", "淡淡"}

// 洛水 GenreProfile 内置题材疲劳词（50 册 §4.2 节选）。
var genreFatigueWords = map[string][]string{
	"xuanhuan": {"气势", "威压", "灵气", "丹田", "经脉", "吞噬", "领悟"},
	"玄幻":       {"气势", "威压", "灵气", "丹田", "经脉", "吞噬", "领悟"},
	"xianxia":  {"道韵", "天劫", "灵根", "仙气", "剑意", "心境"},
	"仙侠":       {"道韵", "天劫", "灵根", "仙气", "剑意", "心境"},
	"urban":    {"眼眸", "薄唇", "磁性", "霸道", "冷冽"},
	"都市":       {"眼眸", "薄唇", "磁性", "霸道", "冷冽"},
	"romance":  {"心跳加速", "脸颊绯红", "小鹿乱撞", "温柔", "宠溺"},
	"言情":       {"心跳加速", "脸颊绯红", "小鹿乱撞", "温柔", "宠溺"},
	"suspense": {"蛛丝马迹", "真相大白", "恍然大悟", "细思极恐"},
	"mystery":  {"蛛丝马迹", "真相大白", "恍然大悟", "细思极恐"},
	"悬疑":       {"蛛丝马迹", "真相大白", "恍然大悟", "细思极恐"},
	"history":  {"拱手", "微臣", "圣上", "社稷"},
	"历史":       {"拱手", "微臣", "圣上", "社稷"},
	"scifi":    {"量子", "维度", "矩阵", "纳米", "全息"},
	"科幻":       {"量子", "维度", "矩阵", "纳米", "全息"},
	"game":     {"经验值", "暴击", "闪避", "系统提示"},
	"游戏":       {"经验值", "暴击", "闪避", "系统提示"},
	"horror":   {"毛骨悚然", "后背发凉", "头皮发麻", "不寒而栗"},
	"恐怖":       {"毛骨悚然", "后背发凉", "头皮发麻", "不寒而栗"},
}

// fatigueWordsForGenre 合并通用疲劳词与题材疲劳词，保持原顺序去重。
func fatigueWordsForGenre(genre string) []string {
	out := append([]string(nil), commonFatigueWords...)
	seen := map[string]bool{}
	for _, w := range out {
		seen[w] = true
	}
	key := strings.ToLower(strings.TrimSpace(genre))
	for _, w := range genreFatigueWords[key] {
		if w != "" && !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	return out
}

// firstPersonFromRules 识别用户在作品规则里声明的叙事人称。
// 原版把叙事人称作为结构化书规则；CM 当前以自由文本规则承载，
// 因此这里只认显式声明，并把来源写回检查结果。
func firstPersonFromRules(rules []string) (bool, string) {
	for _, rule := range rules {
		text := strings.TrimSpace(rule)
		if text == "" {
			continue
		}
		if strings.Contains(text, "第三人称") {
			return false, "rules"
		}
		if strings.Contains(text, "第一人称") {
			return true, "rules"
		}
	}
	return false, "absent"
}

// buildPostWriteContext 按洛水口径收集 targetChapter 之前的章节资料。
// 只取最近 postWriteChapterLimit 章正文，并设置单章与总量上限；
// SourceChapters 记录真正参与跨章检查的章节号，便于前端如实展示覆盖范围。
func buildPostWriteContext(chapters []Chapter, targetChapter int, settings ProjectSettings, project Project) PostWriteContext {
	out := PostWriteContext{
		ExistingTitles:   []string{},
		PreviousFullText: []string{},
		FatigueWords:     fatigueWordsForGenre(project.Genre),
		Protagonist:      strings.TrimSpace(project.Protagonist),
		Genre:            strings.TrimSpace(project.Genre),
		SourceChapters:   []int{},
	}
	out.FirstPerson, out.FirstPersonSource = firstPersonFromRules(settings.Rules)

	ordered := make([]Chapter, 0, len(chapters))
	for _, chapter := range chapters {
		if targetChapter > 0 && chapter.Number >= targetChapter {
			continue
		}
		ordered = append(ordered, chapter)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Number < ordered[j].Number })
	for _, chapter := range ordered {
		if title := strings.TrimSpace(chapter.Title); title != "" {
			out.ExistingTitles = append(out.ExistingTitles, title)
		}
	}
	start := len(ordered) - postWriteChapterLimit
	if start < 0 {
		start = 0
	}
	total := 0
	for _, chapter := range ordered[start:] {
		if total >= postWriteTotalChars {
			break
		}
		body := []rune(chapter.Content)
		if len(body) > postWriteChapterChars {
			body = body[:postWriteChapterChars]
		}
		if total+len(body) > postWriteTotalChars {
			body = body[:postWriteTotalChars-total]
		}
		if len(body) == 0 {
			continue
		}
		total += len(body)
		out.PreviousFullText = append(out.PreviousFullText, string(body))
		out.SourceChapters = append(out.SourceChapters, chapter.Number)
	}
	return out
}

// postWriteContextOf 兼容没有该字段的历史任务：返回空资料，检查项显示未检查。
func postWriteContextOf(value *PostWriteContext) PostWriteContext {
	if value == nil {
		return PostWriteContext{ExistingTitles: []string{}, PreviousFullText: []string{}, FatigueWords: []string{}, SourceChapters: []int{}, FirstPersonSource: "absent"}
	}
	return *value
}

// postWriteStatus 如实报告每一种检查是否具备执行所需的真实资料。
// 没有资料的项不得显示为已通过。
func (p PostWriteContext) postWriteStatus() map[string]string {
	status := map[string]string{}
	if len(p.PreviousFullText) > 0 {
		status["cross_chapter"] = "checked_full_text"
	} else {
		status["cross_chapter"] = "not_checked_missing_full_text"
	}
	if len(p.ExistingTitles) > 0 {
		status["duplicate_title"] = "checked_history_titles"
	} else {
		status["duplicate_title"] = "not_checked_missing_history_titles"
	}
	if len(p.FatigueWords) > 0 {
		status["fatigue"] = "checked_genre_fatigue_words"
	} else {
		status["fatigue"] = "not_checked_missing_fatigue_words"
	}
	if p.FirstPersonSource == "rules" {
		status["person"] = "checked_rules_declaration"
	} else {
		status["person"] = "not_checked_missing_person_declaration"
	}
	if p.Protagonist != "" {
		status["protagonist"] = "available"
	} else {
		status["protagonist"] = "not_checked_missing_protagonist"
	}
	return status
}

// postWriteInput 把冻结资料转换为确定性检查入参。
func (p PostWriteContext) postWriteInput(title string) PostWriteInput {
	return PostWriteInput{
		Title:            title,
		ExistingTitles:   p.ExistingTitles,
		PreviousChapters: p.PreviousFullText,
		FatigueWords:     p.FatigueWords,
		FirstPerson:      p.FirstPerson,
		Protagonist:      p.Protagonist,
	}
}
