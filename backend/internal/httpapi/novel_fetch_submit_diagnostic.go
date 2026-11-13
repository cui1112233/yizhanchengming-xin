package httpapi

import (
	"context"
	"fmt"
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
