package novelcreation

import (
	"strings"
)

// 写前检查与写手标题的协议约束。
// 参考实现把 PRE_WRITE_CHECK 与 CHAPTER_TITLE / CHAPTER_CONTENT 分成独立区块；
// CM 保留 JSON 外壳，但要求检查行只出现在 preWriteCheck 字段，不能混入 content。

var chapterTitlePrefixes = []string{"第", "Chapter "}

// validatePreWriteCheck 只校验结构化字段，不要求必须输出。
// 空数组表示模型省略了写前检查，不影响正文落盘。
func validatePreWriteCheck(rows []preWriteCheckRow) error {
	seen := map[string]bool{}
	for index := range rows {
		item := strings.TrimSpace(rows[index].Item)
		record := strings.TrimSpace(rows[index].Record)
		if item == "" || record == "" {
			return &DomainError{Code: ErrInvalidOutput, Message: "写前检查缺少检查项或本章记录"}
		}
		if seen[item] {
			return &DomainError{Code: ErrInvalidOutput, Message: "写前检查项重复：" + item}
		}
		seen[item] = true
		rows[index].Item = item
		rows[index].Record = record
		rows[index].Note = strings.TrimSpace(rows[index].Note)
	}
	return nil
}

// normalizeDraftTitle 去掉「第X章」前缀与书名号，只保留章节名。
func normalizeDraftTitle(value string) string {
	title := strings.TrimSpace(value)
	title = strings.Trim(title, "《》\"“”'")
	title = strings.TrimSpace(title)
	for {
		trimmed := stripChapterPrefix(title)
		if trimmed == title {
			break
		}
		title = trimmed
	}
	return title
}

func stripChapterPrefix(title string) string {
	if strings.HasPrefix(title, "第") {
		rest := []rune(title[3:])
		index := 0
		for index < len(rest) && (rest[index] >= '0' && rest[index] <= '9' || rest[index] == '零' || rest[index] == '一' || rest[index] == '二' || rest[index] == '三' || rest[index] == '四' || rest[index] == '五' || rest[index] == '六' || rest[index] == '七' || rest[index] == '八' || rest[index] == '九' || rest[index] == '十' || rest[index] == '百' || rest[index] == '千') {
			index++
		}
		if index == 0 || index >= len(rest) || rest[index] != '章' {
			return title
		}
		stripped := strings.TrimSpace(strings.TrimLeft(string(rest[index+1:]), "：:、.． -—"))
		if stripped == "" {
			return title
		}
		return stripped
	}
	if strings.HasPrefix(title, "Chapter ") {
		rest := []rune(title[len("Chapter "):])
		index := 0
		for index < len(rest) && rest[index] >= '0' && rest[index] <= '9' {
			index++
		}
		if index == 0 {
			return title
		}
		return strings.TrimSpace(strings.TrimLeft(string(rest[index:]), ":：.、 "))
	}
	return title
}

// adoptDraftTitle 决定是否采用写手标题。
// 返回空串表示保留章节计划标题。
func adoptDraftTitle(draftTitle, plannedTitle string, existingTitles []string) string {
	title := strings.TrimSpace(draftTitle)
	if title == "" || title == strings.TrimSpace(plannedTitle) {
		return ""
	}
	if len([]rune(title)) > 60 {
		return ""
	}
	normalize := func(value string) string {
		return strings.Map(func(r rune) rune {
			if strings.ContainsRune("，。！？、：；''（）《》\"“” \t\n\r", r) {
				return -1
			}
			return r
		}, value)
	}
	target := normalize(title)
	for _, existing := range existingTitles {
		if normalize(existing) == target {
			return ""
		}
	}
	return title
}
