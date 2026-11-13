package batchfactoryv11

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ManualIntakeInput uses Novel Fetch's book-list protocol. When the caller has
// already fetched an original, it is persisted with the matching Book ID.
type ManualIntakeInput struct {
	Title      string `json:"title"`
	PlatformID string `json:"platformId"`
	// PlatformName is the novel-fetch book-city label shown to people. PlatformID
	// stays internal so raw-content requests can continue using the configured ID.
	PlatformName             string            `json:"platformName"`
	ParseMode                string            `json:"parseMode"`
	ColumnPresetID           string            `json:"columnPresetId"`
	ColumnOrder              string            `json:"columnOrder"`
	InputText                string            `json:"inputText"`
	ContentRangeLines        int               `json:"contentRangeLines,omitempty"`
	ComicAnalysisCharacters  int               `json:"comicAnalysisCharacters,omitempty"`
	ContentCaptureCharacters int               `json:"contentCaptureCharacters,omitempty"`
	ScheduledAt              string            `json:"scheduledAt,omitempty"`
	SourceTextByBookID       map[string]string `json:"sourceTextByBookId,omitempty"`
	// SourceMetadataByBookID carries authoritative fetch facts (for example 121
	// category/genre). Manual columns keep precedence when the book is created.
	SourceMetadataByBookID map[string]map[string]any `json:"sourceMetadataByBookId,omitempty"`
	// InitialBatchSettings lets a first-time user configure the batch before
	// there is a workbench to open. The route persists it atomically with the
	// intake-derived batch instead of creating an unconfigured intermediary.
	InitialBatchSettings SettingsPatch `json:"initialBatchSettings,omitempty"`
	// Groups 可选：多书城分组录入。非空时每组按各自平台解析，忽略顶层 PlatformID/InputText。
	Groups []ManualIntakeGroup `json:"groups,omitempty"`
}

// ManualIntakeGroup 是一个书城下的一批粘贴文本。
type ManualIntakeGroup struct {
	PlatformID   string `json:"platformId"`
	PlatformName string `json:"platformName"`
	InputText    string `json:"inputText"`
	// SourceTextByBookID 可选：该组已抓到的正文，键为书 ID。非空时分组解析
	// 用它替代顶层同名 map，使跨书城撞 ID 的两本书各取各的正文。
	SourceTextByBookID     map[string]string         `json:"sourceTextByBookId,omitempty"`
	SourceMetadataByBookID map[string]map[string]any `json:"sourceMetadataByBookId,omitempty"`
}
type manualRow struct {
	bookID, paidID, freeID, title, gender, style, tags, reason, rating, sourceLine, mode string
	columns                                                                              []string
}

var manualDigits = regexp.MustCompile(`^\d{6,25}$`)
var manualDigitsInLine = regexp.MustCompile(`\d{6,25}`)
var manualBookIDAndTitle = regexp.MustCompile(`^(\d{6,25})\s+(.+)$`)
var manualPastedBookIDAndTitle = regexp.MustCompile(`^(\d{6,25})(\D.+)$`)
var manualSpaces = regexp.MustCompile(`\s{2,}`)
var manualMonth = regexp.MustCompile(`\d{4}[-/年]\d{1,2}`)
var manualAliases = map[string]string{
	"书籍id": "book_id", "书id": "book_id", "id": "book_id", "bookid": "book_id",
	"付费书籍id": "paid_book_id", "付费id": "paid_book_id", "paidbookid": "paid_book_id", "paidid": "paid_book_id",
	"免费书籍id": "free_book_id", "免费id": "free_book_id", "freebookid": "free_book_id", "freeid": "free_book_id",
	"书籍名称": "book_name", "书名": "book_name", "名称": "book_name", "标题": "book_name", "小说名": "book_name", "作品名": "book_name", "bookname": "book_name", "name": "book_name", "title": "book_name",
	"男女频": "gender", "性别": "gender", "频道": "gender", "男频女频": "gender", "频类": "gender", "gender": "gender",
	"风格": "style", "风格类型": "style", "类型": "style", "分类": "style", "ai识别类型": "style", "识别类型": "style", "style": "style", "type": "style",
	"标签": "tags", "题材": "tags", "关键词": "tags", "卖点": "tags", "tag": "tags", "tags": "tags",
	"推荐理由": "reason", "理由": "reason", "推荐语": "reason", "简介": "reason", "文案": "reason", "卖点文案": "reason", "reason": "reason", "desc": "reason",
	"内容评级": "rating", "评级": "rating", "等级": "rating", "rating": "rating", "rank": "rating",
}
var manualPresets = map[string][]string{
	"paid_name_reason":             {"paid_book_id", "book_name", "reason"},
	"paid_name_gender_reason":      {"paid_book_id", "book_name", "gender", "reason"},
	"free_paid_name_gender_reason": {"free_book_id", "paid_book_id", "book_name", "gender", "reason"},
	"sample_input":                 {"book_id", "book_name", "reason", "gender", "tags", "rating"},
	"full_metadata":                {"book_id", "book_name", "gender", "style", "tags", "reason", "rating"},
	"full_11":                      {"ignore", "free_book_id", "paid_book_id", "book_name", "rating", "ignore", "ignore", "ignore", "ignore", "gender", "reason"},
}
var manualModes = map[string][]string{
	"fixed_full_11":    manualPresets["full_11"],
	"fixed_from_b":     {"free_book_id", "paid_book_id", "book_name", "rating", "ignore", "ignore", "ignore", "ignore", "gender", "reason"},
	"fixed_paid_basic": {"paid_book_id", "book_name", "gender", "reason"},
}

func cleanManual(v string) string { return strings.TrimSpace(strings.TrimPrefix(v, "\ufeff")) }
func manualHeader(v string) string {
	return strings.NewReplacer(" ", "", "_", "", "-", "").Replace(strings.ToLower(cleanManual(v)))
}
func manualField(v string) string { return manualAliases[manualHeader(v)] }
func manualGender(v string) string {
	v = cleanManual(v)
	switch v {
	case "男", "男频", "男性", "男生", "男向", "男频文":
		return "男频"
	case "女", "女频", "女性", "女生", "女向", "女频文":
		return "女频"
	}
	if strings.Contains(v, "男频") && !strings.Contains(v, "女频") {
		return "男频"
	}
	if strings.Contains(v, "女频") && !strings.Contains(v, "男频") {
		return "女频"
	}
	return ""
}
func manualRating(v string) bool {
	switch strings.ToUpper(v) {
	case "S", "S+", "A", "A+", "B", "B+", "C", "C+":
		return true
	}
	return false
}
func splitManual(line string) []string {
	if strings.Contains(line, "\t") {
		return cleanManualList(strings.Split(line, "\t"), false)
	}
	if strings.Contains(line, "|") {
		return cleanManualList(strings.Split(line, "|"), false)
	}
	if manualSpaces.MatchString(line) {
		return cleanManualList(manualSpaces.Split(line, -1), true)
	}
	if strings.Contains(line, ",") {
		return splitManualCSV(line)
	}
	if parts := manualBookIDAndTitle.FindStringSubmatch(cleanManual(line)); len(parts) == 3 {
		return []string{parts[1], cleanManual(parts[2])}
	}
	return []string{cleanManual(line)}
}
func cleanManualList(in []string, drop bool) []string {
	out := []string{}
	for _, v := range in {
		v = cleanManual(v)
		if !drop || v != "" {
			out = append(out, v)
		}
	}
	return out
}
func splitManualCSV(line string) []string {
	out := []string{}
	var b strings.Builder
	quoted := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if quoted {
			if c == '"' {
				if i+1 < len(line) && line[i+1] == '"' {
					b.WriteByte(c)
					i++
				} else {
					quoted = false
				}
			} else {
				b.WriteByte(c)
			}
			continue
		}
		if c == '"' {
			quoted = true
		} else if c == ',' {
			out = append(out, cleanManual(b.String()))
			b.Reset()
		} else {
			b.WriteByte(c)
		}
	}
	return append(out, cleanManual(b.String()))
}
func headerScore(cells []string) int {
	seen := map[string]bool{}
	for _, c := range cells {
		if f := manualField(c); f != "" {
			seen[f] = true
		}
	}
	return len(seen)
}
func hasHeader(cells []string) bool { return headerScore(cells) >= 2 }
func columnsFor(input ManualIntakeInput) []string {
	mode := strings.TrimSpace(input.ParseMode)
	if v, ok := manualModes[mode]; ok {
		return append([]string(nil), v...)
	}
	if v, ok := manualPresets[strings.TrimSpace(input.ColumnPresetID)]; ok {
		return append([]string(nil), v...)
	}
	if mode != "custom" {
		return nil
	}
	out := []string{}
	for _, part := range strings.FieldsFunc(input.ColumnOrder, func(r rune) bool { return unicode.IsSpace(r) || strings.ContainsRune(",，、/|", r) }) {
		if f := manualField(part); f != "" {
			out = append(out, f)
		} else if n := manualHeader(part); n == "ignore" || n == "忽略" || n == "跳过" || n == "空" {
			out = append(out, "ignore")
		}
	}
	return out
}
func applyManual(row *manualRow, field, value string) {
	if value == "" || field == "" || field == "ignore" {
		return
	}
	switch field {
	case "book_id":
		row.bookID = value
	case "paid_book_id":
		row.paidID = value
	case "free_book_id":
		row.freeID = value
	case "book_name":
		row.title = value
	case "gender":
		row.gender = manualGender(value)
	case "style":
		row.style = value
	case "tags":
		row.tags = value
	case "reason":
		row.reason = value
	case "rating":
		row.rating = strings.ToUpper(value)
	}
}
func nonEmpty(in []string) []string {
	out := []string{}
	for _, v := range in {
		if v = cleanManual(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
func finalize(row manualRow, mode string) manualRow {
	if row.bookID == "" {
		if row.paidID != "" {
			row.bookID = row.paidID
		} else if row.freeID != "" {
			row.bookID = row.freeID
		} else {
			row.bookID = manualDigitsInLine.FindString(row.sourceLine)
		}
	}
	if parts := manualPastedBookIDAndTitle.FindStringSubmatch(row.bookID); len(parts) == 3 {
		row.bookID = parts[1]
		if row.title == "" {
			row.title = cleanManual(parts[2])
		}
	}
	row.gender = manualGender(row.gender)
	row.mode = mode
	return row
}
func mappedManual(cells, columns []string, mode string) manualRow {
	row := manualRow{sourceLine: strings.Join(cells, "\t"), columns: append([]string(nil), columns...)}
	for i, f := range columns {
		if i >= len(cells) {
			break
		}
		v := cells[i]
		if (f == "reason" || f == "tags") && i == len(columns)-1 {
			v = strings.Join(nonEmpty(cells[i:]), " ")
		}
		applyManual(&row, f, v)
		if (f == "reason" || f == "tags") && i == len(columns)-1 {
			break
		}
	}
	return finalize(row, mode)
}
func inferredManual(cells []string, mode string) manualRow {
	clean := nonEmpty(cells)
	row := manualRow{sourceLine: strings.Join(clean, "\t")}
	ids := []int{}
	for i, v := range clean {
		if manualDigits.MatchString(v) {
			ids = append(ids, i)
		}
	}
	idIndex := -1
	if len(ids) >= 2 && manualMonth.MatchString(clean[0]) {
		row.freeID, row.paidID, row.bookID, idIndex = clean[ids[0]], clean[ids[1]], clean[ids[1]], ids[1]
	} else if len(ids) > 0 {
		row.bookID, idIndex = clean[ids[0]], ids[0]
	}
	used := map[int]bool{}
	for _, i := range ids {
		used[i] = true
	}
	for i, v := range clean {
		if g := manualGender(v); g != "" && row.gender == "" {
			row.gender = g
			used[i] = true
		}
		if manualRating(v) && row.rating == "" {
			row.rating = strings.ToUpper(v)
			used[i] = true
		}
	}
	for i := idIndex + 1; i < len(clean); i++ {
		if !used[i] {
			row.title = clean[i]
			used[i] = true
			break
		}
	}
	rest := []string{}
	for i, v := range clean {
		if !used[i] {
			rest = append(rest, v)
		}
	}
	if len(rest) > 0 {
		row.reason, row.tags = strings.Join(rest, " "), strings.Join(rest, " ")
	}
	return finalize(row, mode)
}
func parseManualRows(input ManualIntakeInput) []manualRow {
	mode := strings.TrimSpace(input.ParseMode)
	switch mode {
	case "smart", "header", "multi_header", "fixed_full_11", "fixed_from_b", "fixed_paid_basic", "custom":
	default:
		mode = "smart"
	}
	raw := [][]string{}
	for _, line := range strings.Split(input.InputText, "\n") {
		if cleanManual(line) != "" {
			raw = append(raw, splitManual(line))
		}
	}
	if len(raw) == 0 {
		return nil
	}
	columns := columnsFor(input)
	data := [][]string{}
	for _, cells := range raw {
		if !hasHeader(cells) {
			data = append(data, cells)
		}
	}
	if mode == "smart" || mode == "header" || mode == "multi_header" {
		limit := 5
		if mode == "multi_header" {
			limit = 8
		}
		if limit > len(raw) {
			limit = len(raw)
		}
		first := -1
		for i := 0; i < limit; i++ {
			if hasHeader(raw[i]) {
				first = i
				break
			}
		}
		if first >= 0 {
			last := first
			for last+1 < limit && hasHeader(raw[last+1]) {
				last++
			}
			width := 0
			for _, cells := range raw[:last+1] {
				if len(cells) > width {
					width = len(cells)
				}
			}
			headers := make([]string, width)
			for _, cells := range raw[:last+1] {
				for i, c := range cells {
					if f := manualField(c); f != "" {
						headers[i] = f
					}
				}
			}
			out := []manualRow{}
			for _, cells := range raw[last+1:] {
				if !hasHeader(cells) {
					out = append(out, mappedManual(cells, headers, mode))
				}
			}
			return out
		}
	}
	out := []manualRow{}
	for _, cells := range data {
		if len(columns) > 0 {
			out = append(out, mappedManual(cells, columns, mode))
		} else {
			out = append(out, inferredManual(cells, mode))
		}
	}
	return out
}

// ParseManualBookList mirrors Novel Fetch parsing, deduplicating each final
// book ID while retaining the first raw input line for later audit.
func ParseManualBookList(input ManualIntakeInput) ([]CreateBookInput, error) {
	input.PlatformID = strings.TrimSpace(input.PlatformID)
	input.PlatformName = strings.TrimSpace(input.PlatformName)
	if input.PlatformID == "" || strings.TrimSpace(input.InputText) == "" {
		return nil, ErrInvalid
	}
	merged := map[string]manualRow{}
	order := []string{}
	for _, row := range parseManualRows(input) {
		if row.bookID == "" {
			continue
		}
		if old, ok := merged[row.bookID]; ok {
			if row.title != "" {
				old.title = row.title
			}
			if row.gender != "" {
				old.gender = row.gender
			}
			if row.style != "" {
				old.style = row.style
			}
			if row.tags != "" {
				old.tags = row.tags
			}
			if row.reason != "" {
				old.reason = row.reason
			}
			if row.rating != "" {
				old.rating = row.rating
			}
			if row.paidID != "" {
				old.paidID = row.paidID
			}
			if row.freeID != "" {
				old.freeID = row.freeID
			}
			merged[row.bookID] = old
		} else {
			merged[row.bookID] = row
			order = append(order, row.bookID)
		}
	}
	if len(order) == 0 {
		return nil, ErrInvalid
	}
	out := make([]CreateBookInput, 0, len(order))
	for _, id := range order {
		row := merged[id]
		fetchedMeta := input.SourceMetadataByBookID[id]
		title := row.title
		if title == "" {
			title = "小说 " + id
		}
		if title == "小说 "+id {
			if fetchedTitle, ok := fetchedMeta["bookName"].(string); ok && strings.TrimSpace(fetchedTitle) != "" {
				title = strings.TrimSpace(fetchedTitle)
			}
		}
		meta := map[string]any{"platformName": input.PlatformName, "gender": row.gender, "style": row.style, "tags": row.tags, "reason": row.reason, "rating": row.rating, "paidBookId": row.paidID, "freeBookId": row.freeID, "sourceLine": row.sourceLine, "parseMode": row.mode, "parseColumns": row.columns, "sourceMode": "manual_original"}
		if fetchedMeta["sourceMode"] == "custom_copy" {
			for _, key := range []string{"sourceMode", "customCopyRecordId", "customCopyCreatedAt", "customCopyPlatform"} {
				if value, ok := fetchedMeta[key].(string); ok {
					meta[key] = value
				}
			}
		}
		if category, ok := fetchedMeta["category"].(string); ok && strings.TrimSpace(category) != "" {
			meta["category"] = strings.TrimSpace(category)
		}
		if genre, ok := fetchedMeta["genre"]; ok && genre != nil && strings.TrimSpace(fmt.Sprint(genre)) != "" {
			meta["genre"] = genre
		}
		if row.gender != "" {
			meta["genderSource"] = "manual"
		} else if fetchedGender, ok := fetchedMeta["gender"].(string); ok && manualGender(fetchedGender) != "" {
			meta["gender"] = manualGender(fetchedGender)
			if source, ok := fetchedMeta["genderSource"].(string); ok && strings.TrimSpace(source) != "" {
				meta["genderSource"] = strings.TrimSpace(source)
			}
		}
		sourceText := strings.TrimSpace(input.SourceTextByBookID[id])
		out = append(out, CreateBookInput{ID: id, BookID: id, Title: title, Platform: input.PlatformID, SourceText: sourceText, TxtText: sourceText, SourceMetadata: meta})
	}
	return out, nil
}

// ParseGroupedManualBookLists 按书城分组解析粘贴文本。每组复用
// ParseManualBookList 的全部解析/组内去重规则，仅平台信息按组取值；
// 跨组重复 ID 保留（不同书城可能撞 ID，视为两本书）。
func ParseGroupedManualBookLists(base ManualIntakeInput) ([]CreateBookInput, error) {
	if len(base.Groups) == 0 {
		return nil, ErrInvalid
	}
	out := make([]CreateBookInput, 0)
	for _, group := range base.Groups {
		groupInput := base
		groupInput.PlatformID = strings.TrimSpace(group.PlatformID)
		groupInput.PlatformName = strings.TrimSpace(group.PlatformName)
		groupInput.InputText = group.InputText
		if len(group.SourceTextByBookID) > 0 {
			groupInput.SourceTextByBookID = group.SourceTextByBookID
		}
		if len(group.SourceMetadataByBookID) > 0 {
			groupInput.SourceMetadataByBookID = group.SourceMetadataByBookID
		}
		books, err := ParseManualBookList(groupInput)
		if err != nil {
			return nil, err
		}
		out = append(out, books...)
	}
	if len(out) == 0 {
		return nil, ErrInvalid
	}
	return out, nil
}
