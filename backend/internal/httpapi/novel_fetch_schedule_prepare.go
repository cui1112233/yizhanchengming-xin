package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func scheduleInputMetadata(item map[string]any) workshopInputMetadata {
	order := scheduleStringSlice(item["column_order"])
	if len(order) == 0 {
		if text, ok := item["column_order"].(string); ok {
			order = workshopColumnOrderFromText(text)
		}
	}
	return workshopInputMetadata{
		Format:       firstText(item["parse_mode"], item["input_format"], item["inputFormat"]),
		ColumnPreset: firstText(item["column_preset_id"], item["column_preset"], item["columnPreset"]),
		ColumnOrder:  order,
		Normalized:   item["input_normalized"] == true,
	}
}

// The UI sends canonical rows; direct clients can supply explicit column order
// or a header. Preserve empty cells rather than shifting subsequent columns.
func parseScheduleDocuments(item map[string]any) ([]novelfetchworkshop.Document, int, int, int) {
	input := firstText(item["input_text"], item["inputText"])
	maxTxt := int(targetNumber(firstConfiguredValue(item["max_txt"], item["maxTxt"])))
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	metadata := scheduleInputMetadata(item)
	columns := metadata.ColumnOrder
	if item["input_normalized"] == true {
		columns = []string{"bookId", "bookName", "gender", "reason"}
	}
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	start := 0
	if item["input_normalized"] != true {
		for index, line := range lines {
			if index >= 8 {
				break
			}
			cells := splitScheduleCells(line)
			header := make([]string, len(cells))
			recognized, hasID := 0, false
			for i, cell := range cells {
				header[i] = scheduleColumnKey(cell)
				if header[i] != "" {
					recognized++
				}
				if header[i] == "bookId" {
					hasID = true
				}
			}
			if hasID && recognized >= 2 {
				columns, start = header, index+1
				break
			}
		}
	}
	if len(columns) == 0 {
		docs, parsed, empty, duplicate := parseWorkshopDocuments(input, maxTxt, metadata)
		for i := range docs {
			metadata.apply(docs[i].Meta)
		}
		return docs, parsed, empty, duplicate
	}
	canonical := []string{}
	rowInfo := map[string]map[string]string{}
	parsed, empty := 0, 0
	for _, line := range lines[start:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parsed++
		cells := splitScheduleCells(line)
		row := map[string]string{}
		// Older smart-form schedules saved display column metadata beside a
		// plain "ID title" line. Preserve that supported input, too.
		if len(cells) == 1 && item["input_normalized"] != true && (metadata.Format == "smart" || metadata.Format == "") {
			fallback, _, _, _ := parseWorkshopDocuments(line, maxTxt, metadata)
			if len(fallback) == 1 {
				row["bookId"], row["bookName"], row["gender"] = fallback[0].BookID, firstText(fallback[0].Meta["bookName"]), fallback[0].Gender
			}
		} else {
			for i, column := range columns {
				if i < len(cells) {
					row[scheduleColumnKey(column)] = strings.TrimSpace(cells[i])
				}
			}
		}
		id := novelFetchSafeBookID(row["bookId"])
		if id == "" {
			empty++
			continue
		}
		canonical = append(canonical, strings.Join([]string{id, row["bookName"], row["gender"], row["reason"]}, "\t"))
		if _, exists := rowInfo[id]; !exists {
			rowInfo[id] = row
		}
	}
	docs, _, _, duplicate := parseWorkshopDocuments(strings.Join(canonical, "\n"), maxTxt)
	for i := range docs {
		row := rowInfo[docs[i].BookID]
		docs[i].Meta["bookName"], docs[i].Meta["reason"] = row["bookName"], row["reason"]
		// A gender word in the recommendation is not a manual gender field.
		docs[i].Gender, docs[i].GenderSource, docs[i].GenderReason = "", "", ""
		delete(docs[i].Meta, "gender")
		delete(docs[i].Meta, "genderSource")
		gender := strings.TrimSpace(row["gender"])
		if gender == "男" {
			gender = "男频"
		}
		if gender == "女" {
			gender = "女频"
		}
		if gender == "男频" || gender == "女频" {
			docs[i].Gender, docs[i].GenderSource, docs[i].GenderReason = gender, novelfetchworkshop.GenderManual, "采用输入清单中的男女频"
			docs[i].Meta["gender"], docs[i].Meta["genderSource"] = gender, "input"
		}
		metadata.apply(docs[i].Meta)
	}
	return docs, parsed, empty, duplicate
}

func splitScheduleCells(line string) []string {
	for _, delimiter := range []string{"\t", "|", ","} {
		if strings.Contains(line, delimiter) {
			return strings.Split(line, delimiter)
		}
	}
	return []string{line}
}

func scheduleColumnKey(value string) string {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "_", "")) {
	case "bookid", "id", "书籍id", "小说id", "书号", "书籍编号":
		return "bookId"
	case "bookname", "title", "书名", "作品名称", "小说名称":
		return "bookName"
	case "gender", "男女频", "性别", "频道":
		return "gender"
	case "reason", "理由", "推荐理由", "推荐原因", "简介":
		return "reason"
	}
	return ""
}

func scheduleBookSkipReason(ctx context.Context, store novelfetchworkshop.Store, owner string, doc novelfetchworkshop.Document, item map[string]any) string {
	meta := doc.Meta
	status := strings.ToLower(firstText(meta["status"], meta["queueState"]))
	if status == "running" || status == "queued" || status == "processing" {
		return "已有任务正在执行，不重复启动"
	}
	if novelFetchStopped(status) || meta["manualOnly"] == true || meta["manual_only"] == true || meta["rewriteValidationHold"] == true || novelFetchRetryScheduleHasManualHold(meta) {
		return "任务已停止或需要人工处理，请先明确重试"
	}
	if meta["scheduledHold"] == true && firstText(meta["scheduleId"], meta["schedule_id"]) != firstText(item["id"]) {
		return "已由其他定时任务安排，不重复占用"
	}
	versions := scheduleStringSlice(item["selected_versions"])
	if len(versions) == 0 {
		versions = []string{"original"}
	}
	if item["auto_submit"] != false {
		done := scheduleStringSlice(firstConfiguredValue(meta["siteSubmitDoneVersions"], meta["site_submit_done_versions"]))
		for _, version := range versions {
			if !seenString(done, version) {
				return ""
			}
		}
		return "所选版本已经提交成功"
	}
	for _, version := range versions {
		if version == "original" {
			if strings.TrimSpace(firstText(doc.ProcessedText, doc.Original)) != "" {
				continue
			}
			version = "processed_text"
		} else if strings.TrimSpace(toString(doc.Versions[version])) != "" {
			continue
		}
		body, err := store.GetBody(ctx, owner, doc.BookID, version)
		if err != nil || strings.TrimSpace(body.Content) == "" {
			return ""
		}
	}
	return "所选内容已经生成完成"
}

// An interrupted preparation is repeatable: it retains the schedule ID and
// never overwrites prose, AI versions or submission receipts on an existing book.
func prepareNovelFetchScheduleTasks(ctx context.Context, store novelfetchworkshop.Store, owner string, item map[string]any, now time.Time) error {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return err
	}
	if _, exists := item["selected_versions"]; !exists {
		item["selected_versions"] = novelFetchAutomationVersions(settings)
	}
	if versions := scheduleStringSlice(item["selected_versions"]); len(versions) > 0 {
		normalizedVersions, normalizeErr := normalizeNovelFetchAIVersions(versions)
		if normalizeErr != nil {
			return normalizeErr
		}
		item["selected_versions"] = normalizedVersions
	}
	if maxTxt := int(targetNumber(firstConfiguredValue(item["max_txt"], item["maxTxt"]))); novelFetchMaxTxtInvalid(maxTxt) {
		return fmt.Errorf("max_txt 不能超过 %d", novelFetchMaxTxtLimit)
	}
	runAt, err := parseNovelFetchScheduleTime(firstText(item["run_at"], item["runAt"]))
	if err != nil {
		return err
	}
	_, batchID := novelFetchScheduleDispatchIDs(owner, firstText(item["id"]), runAt)
	batchID = firstText(item["batch_id"], batchID)
	documents, parsed, empty, duplicates := parseScheduleDocuments(item)
	ids, created, requested, skippedIDs := []string{}, scheduleStringSlice(item["created_ids"]), []string{}, []string{}
	skipped := []any{}
	allComplete := true
	for _, incoming := range documents {
		requested = append(requested, incoming.BookID)
		doc, lookupErr := store.GetDocument(ctx, owner, incoming.BookID)
		isNew := errors.Is(lookupErr, novelfetchworkshop.ErrNotFound)
		if lookupErr != nil && !isNew {
			return lookupErr
		}
		if isNew {
			doc = incoming
		} else {
			if reason := scheduleBookSkipReason(ctx, store, owner, doc, item); reason != "" {
				skippedIDs = append(skippedIDs, doc.BookID)
				skipped = append(skipped, map[string]any{"book_id": doc.BookID, "reason": reason})
				if reason != "所选版本已经提交成功" && reason != "所选内容已经生成完成" {
					allComplete = false
				}
				continue
			}
		}
		if doc.Meta == nil {
			doc.Meta = map[string]any{}
		}
		doc.Meta["scheduleId"], doc.Meta["schedule_id"] = item["id"], item["id"]
		doc.Meta["scheduledHold"], doc.Meta["scheduledRunAt"] = true, runAt.UTC().Format(time.RFC3339Nano)
		doc.Meta["status"], doc.Meta["queueState"] = "scheduled_waiting", "scheduled_waiting"
		doc.Meta["currentBatchId"] = batchID
		if isNew {
			doc.Meta["platformId"] = firstText(item["platform_id"], "2")
			doc.Meta["processingFingerprint"] = novelFetchProcessingFingerprint(settings)
			doc.Meta["scheduleCreatedBy"] = item["id"]
		}
		delete(doc.Meta, "currentProcessJobId")
		delete(doc.Meta, "current_process_job_id")
		if err := store.PutDocument(ctx, owner, doc); err != nil {
			return err
		}
		ids = append(ids, doc.BookID)
		if firstText(doc.Meta["scheduleCreatedBy"]) == firstText(item["id"]) && !seenString(created, doc.BookID) {
			created = append(created, doc.BookID)
		}
	}
	item["book_ids"], item["created_ids"], item["requested_book_ids"] = ids, created, requested
	item["skipped_existing_book_ids"], item["skipped_books"] = skippedIDs, skipped
	item["parsed_count"], item["empty_id_count"], item["duplicate_count"] = parsed, empty, duplicates
	item["batch_id"], item["tasks_prepared"], item["status"] = batchID, true, "scheduled"
	if len(ids) == 0 {
		item["status"] = "skipped_existing"
		item["error"] = "没有新派发任务，请查看每本书的跳过原因"
		if len(skippedIDs) > 0 && allComplete {
			item["status"] = "completed"
			delete(item, "error")
		}
	}
	item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
	if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
		return fmt.Errorf("保存待执行书籍失败: %w", err)
	}
	return nil
}

func filterDueScheduleBooks(ctx context.Context, store novelfetchworkshop.Store, owner string, item map[string]any) ([]string, bool, error) {
	eligible := []string{}
	allComplete := true
	skipped := arrayOrEmpty(item["skipped_books"])
	for _, id := range scheduleStringSlice(item["book_ids"]) {
		doc, err := store.GetDocument(ctx, owner, id)
		if err != nil {
			return nil, false, err
		}
		reason := scheduleBookSkipReason(ctx, store, owner, doc, item)
		// A task that became busy after preparation still belongs to this
		// schedule. Queue deduplication will defer this book independently.
		if reason == "已有任务正在执行，不重复启动" && firstText(doc.Meta["scheduleId"], doc.Meta["schedule_id"]) == firstText(item["id"]) {
			reason = ""
		}
		if reason == "" && firstText(doc.Meta["scheduleId"], doc.Meta["schedule_id"]) != firstText(item["id"]) {
			reason = "书籍已由其他任务接管，本定时任务不再派发"
		}
		if reason == "" {
			eligible = append(eligible, id)
			continue
		}
		if reason != "所选版本已经提交成功" && reason != "所选内容已经生成完成" {
			allComplete = false
		}
		skipped = append(skipped, map[string]any{"book_id": id, "reason": reason})
	}
	item["skipped_books"] = skipped
	return eligible, allComplete, nil
}

// Recreate only missing IDs from the same interrupted dispatch's saved input.
func ensureScheduleRecoveryDocuments(ctx context.Context, store novelfetchworkshop.Store, owner string, item map[string]any) error {
	documents, _, _, _ := parseScheduleDocuments(item)
	byID := map[string]novelfetchworkshop.Document{}
	for _, doc := range documents {
		byID[doc.BookID] = doc
	}
	for _, id := range scheduleStringSlice(item["book_ids"]) {
		_, err := store.GetDocument(ctx, owner, id)
		if err == nil {
			continue
		}
		if !errors.Is(err, novelfetchworkshop.ErrNotFound) {
			return err
		}
		doc, found := byID[id]
		if !found {
			return fmt.Errorf("定时派发书籍 %s 的恢复输入缺失，未重新生成", id)
		}
		doc.Meta["platformId"], doc.Meta["currentBatchId"] = firstText(item["platform_id"], "2"), item["batch_id"]
		doc.Meta["scheduleId"], doc.Meta["scheduledHold"] = item["id"], true
		if err := store.PutDocument(ctx, owner, doc); err != nil {
			return err
		}
	}
	return nil
}
