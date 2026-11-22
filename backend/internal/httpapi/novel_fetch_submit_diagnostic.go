package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

func recordWorkshopSubmitBlocked(ctx context.Context, store novelfetchworkshop.Store, owner string, ids []string, operationID, status, message string) {
	for _, id := range cleanStrings(ids) {
		document, err := store.GetDocument(ctx, owner, id)
		if err != nil {
			continue
		}
		document.Logs = append(document.Logs, map[string]any{"event": "site_submit_blocked", "status": status, "operation_id": operationID, "error": message, "externalRequested": false, "time": time.Now().UTC().Format(time.RFC3339Nano)})
		_ = store.PutDocument(ctx, owner, document)
	}
}

func novelFetchSubmissionNeedsReadback(meta map[string]any, version string) bool {
	pending := mapOrEmpty(meta["siteSubmitPendingConfirmation"])
	if _, found := pending[version]; found || containsVersion(meta["siteSubmitExecutingVersions"], version) {
		return true
	}
	if len(pending) != 0 || len(anyStrings(meta["siteSubmitExecutingVersions"])) != 0 {
		return false
	}
	switch strings.ToLower(firstText(meta["siteSubmitStatus"], meta["site_submit_status"])) {
	case "pending_verification", "pending_confirmation", "unknown":
		failed := anyStrings(firstConfiguredValue(meta["siteSubmitFailedVersions"], meta["site_submit_failed_versions"]))
		if len(failed) != 0 {
			return containsVersion(failed, version)
		}
		done := anyStrings(firstConfiguredValue(meta["siteSubmitDoneVersions"], meta["site_submit_done_versions"]))
		return !containsVersion(done, version)
	}
	return false
}

func priorWorkshopSubmitReceipt(document novelfetchworkshop.Document, version string) string {
	for i := len(document.Logs) - 1; i >= 0; i-- {
		entry := mapOrEmpty(document.Logs[i])
		if id := workshopVerifiedSubmitReceipt(entry); entry["version"] == version && id != "" {
			return id
		}
	}
	return ""
}

func workshopVerifiedSubmitReceipt(entry map[string]any) string {
	if entry["event"] != "site_submit" || entry["status"] != "submitted" {
		return ""
	}
	// Historical listings may help display old logs but do not identify the
	// version that created a remote record.
	receipt := mapOrEmpty(entry["remote_receipt"])
	if receipt["verified"] != true {
		return ""
	}
	return firstText(receipt["submission_id"])
}

func workshopHasConfirmedReceipt(doc novelfetchworkshop.Document, id string) bool {
	for _, raw := range doc.Logs {
		if workshopVerifiedSubmitReceipt(mapOrEmpty(raw)) == id {
			return true
		}
	}
	for _, raw := range arrayOrEmpty(doc.Meta["websiteSubmitHistory"]) {
		entry := mapOrEmpty(raw)
		receipt := mapOrEmpty(entry["remote_receipt"])
		if entry["status"] == "submitted" && receipt["verified"] == true && firstText(receipt["submission_id"]) == id {
			return true
		}
	}
	return false
}

func workshopRemoteRecordID(value any) string {
	switch value.(type) {
	case string, json.Number:
		return strings.TrimSpace(fmt.Sprint(value))
	default:
		return ""
	}
}

func workshopPrecloneReceiptIDs(ctx context.Context, target videomanagementconnector.Service, session videomanagementconnector.Session, bookID, sourceID string) ([]string, error) {
	reader, ok := target.(videomanagementconnector.BookReadbackService)
	if !ok {
		return nil, fmt.Errorf("连接器不支持按页读取历史回执")
	}
	const pageSize, maxPages = 100, 50
	ids := []string{}
	seen := map[string]bool{}
	total := 0
	for page := 1; page <= maxPages; page++ {
		raw, err := reader.GetBook(ctx, session, bookID, page, pageSize)
		if err != nil {
			return nil, err
		}
		payload := mapOrEmpty(raw)
		rows, _ := payload["data"].([]any)
		count, countErr := strconv.Atoi(strings.TrimSpace(fmt.Sprint(payload["total"])))
		if payload["success"] != true || countErr != nil || count < 1 || count > pageSize*maxPages || len(rows) == 0 || len(rows) > pageSize {
			return nil, fmt.Errorf("视频管理系统未返回完整历史回执清单")
		}
		if page == 1 {
			total = count
		}
		if count != total || len(ids)+len(rows) > total {
			return nil, fmt.Errorf("读取历史回执期间清单发生变化")
		}
		for _, row := range rows {
			record := mapOrEmpty(row)
			id := workshopRemoteRecordID(record["id"])
			if id == "" || workshopRemoteRecordID(record["bookid"]) != bookID || seen[id] {
				return nil, fmt.Errorf("历史回执中包含重复或无法确认归属的记录")
			}
			seen[id] = true
			ids = append(ids, id)
		}
		if len(ids) == total {
			if !seen[sourceID] {
				return nil, fmt.Errorf("历史回执清单中没有本版本的原回执")
			}
			return ids, nil
		}
	}
	return nil, fmt.Errorf("历史回执页数过多，未能完整核对")
}

func workshopPendingReadbackMatches(receipt videomanagementconnector.Submission, doc novelfetchworkshop.Document, bookID, version string) bool {
	record := mapOrEmpty(receipt.ReadbackRaw)
	recordBook := workshopRemoteRecordID(record["bookid"])
	id := strings.TrimSpace(receipt.ID)
	if id == "" || (recordBook != "" && recordBook != bookID) || workshopHasConfirmedReceipt(doc, id) {
		return false
	}
	attempts := mapOrEmpty(doc.Meta["siteSubmitAttempts"])
	attempt := mapOrEmpty(attempts[version])
	if containsVersion(attempt["preclone_receipt_ids"], id) {
		return false
	}
	remoteOperation, savedOperation := firstText(record["operation_id"]), firstText(attempt["operation_id"])
	hasOperationEvidence := remoteOperation != "" && savedOperation != ""
	if hasOperationEvidence && remoteOperation != savedOperation {
		return false
	}
	pending := mapOrEmpty(doc.Meta["siteSubmitPendingConfirmation"])
	if expected := firstText(pending[version]); expected != "" && expected != bookID {
		for other, candidate := range pending {
			if other != version && firstText(candidate) == id {
				return false
			}
		}
		return id == expected
	}
	recordVersion := firstText(record["version_id"], record["version"])
	if recordVersion != "" && recordVersion != version {
		return false
	}
	matchesEvidence := func(key string, compare func(string, string) bool) bool {
		value := firstText(record[key])
		if value == "" || !compare(value, firstText(attempt[key])) {
			return false
		}
		if recordVersion == "" {
			for other, candidate := range attempts {
				if other != version && compare(value, firstText(mapOrEmpty(candidate)[key])) {
					return false
				}
			}
		}
		return true
	}
	if hasOperationEvidence && matchesEvidence("operation_id", func(received, saved string) bool { return received == saved }) {
		return true
	}
	if !matchesEvidence("content_sha256", strings.EqualFold) {
		return false
	}
	// If the connector supplies a creation time, stale content cannot prove
	// that a new attempt using the same body was received.
	return firstText(record["addtime"]) == "" || submissionMatchesAttempt(receipt, bookID, firstText(attempt["started_at"]))
}

// Read the same durable events that submission workers write. Metadata-only
// history belongs to imported tasks and is retained as a fallback.
func workshopSubmitDiagnostic(document novelfetchworkshop.Document) map[string]any {
	logs := []any{}
	for _, raw := range document.Logs {
		entry := mapOrEmpty(raw)
		event := firstText(entry["event"])
		if strings.HasPrefix(event, "site_submit") || strings.HasPrefix(event, "auto_submit") {
			if event == "site_submit" {
				entry = repairHistoricalSubmitReceipt(entry, document.BookID)
			}
			logs = append(logs, entry)
		}
	}
	if len(logs) == 0 {
		if history, ok := document.Meta["websiteSubmitHistory"].([]any); ok {
			logs = append(logs, history...)
		}
	}
	if len(logs) > 200 {
		logs = logs[len(logs)-200:]
	}
	var result any
	for i := len(logs) - 1; i >= 0; i-- {
		entry := mapOrEmpty(logs[i])
		if firstText(entry["event"]) == "site_submit" || firstText(entry["event"]) == "" {
			result = entry
			break
		}
	}
	reason := ""
	if len(logs) == 0 {
		reason = firstText(document.Meta["automationError"], document.Meta["error"])
		if reason == "" {
			reason = "尚未发起提交，请先完成正文与 AI 文案处理"
		}
	}
	return map[string]any{"meta": legacyBatchRewriteTask(document), "result": result, "logs": logs, "reason": reason}
}

// Old receipts saved the whole readback list without extracting its record ID.
// Enrich the response only when book and attempt time identify a fresh record.
func repairHistoricalSubmitReceipt(entry map[string]any, bookID string) map[string]any {
	receipt := mapOrEmpty(entry["remote_receipt"])
	if receipt["verified"] != true || firstText(receipt["submission_id"]) != "" {
		return entry
	}
	raw := mapOrEmpty(receipt["remote_record"])
	rows, _ := raw["data"].([]any)
	for _, value := range rows {
		record := mapOrEmpty(value)
		if firstText(record["bookid"]) != bookID {
			continue
		}
		id := strings.TrimSpace(fmt.Sprint(record["id"]))
		if id == "" || id == "<nil>" {
			continue
		}
		confirmed := videomanagementconnector.Submission{ID: id, ReadbackRaw: record}
		if !submissionMatchesAttempt(confirmed, bookID, firstText(entry["time"])) {
			continue
		}
		updated := map[string]any{}
		for k, v := range entry {
			updated[k] = v
		}
		updatedReceipt := map[string]any{}
		for k, v := range receipt {
			updatedReceipt[k] = v
		}
		updatedReceipt["submission_id"] = id
		updated["remote_receipt"] = updatedReceipt
		return updated
	}
	return entry
}
