package httpapi

import (
	"context"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
	"testing"
	"time"
)

func TestSubmitReadbackCannotConfirmAnOlderUpload(t *testing.T) {
	started := time.Date(2026, 11, 13, 4, 44, 11, 0, time.UTC).Format(time.RFC3339Nano)
	old := videomanagementconnector.Submission{ID: "old", ReadbackRaw: map[string]any{"bookid": "book", "addtime": "2026-11-10 23:54:40"}}
	if submissionMatchesAttempt(old, "book", started) {
		t.Fatal("old upload confirmed a new attempt")
	}
	fresh := videomanagementconnector.Submission{ID: "new", ReadbackRaw: map[string]any{"bookid": "book", "addtime": "2026-11-13 12:45:35"}}
	if !submissionMatchesAttempt(fresh, "book", started) {
		t.Fatal("fresh upload was rejected")
	}
	if submissionMatchesAttempt(fresh, "other-receipt", started) {
		t.Fatal("different remote receipt accepted")
	}
}

func TestHistoricalReadbackExtractsFreshMatchingReceiptWithoutChangingStoredLog(t *testing.T) {
	old := map[string]any{"id": 235423, "bookid": "book", "addtime": "2026-10-08 09:22:39"}
	fresh := map[string]any{"id": 288072, "bookid": "book", "addtime": "2026-11-13 19:26:09"}
	receipt := map[string]any{"verified": true, "submission_id": "", "remote_record": map[string]any{"data": []any{old, fresh}}}
	entry := map[string]any{"time": "2026-11-13T11:24:43Z", "remote_receipt": receipt}
	got := repairHistoricalSubmitReceipt(entry, "book")
	if firstText(mapOrEmpty(got["remote_receipt"])["submission_id"]) != "288072" {
		t.Fatalf("receipt=%v", got)
	}
	if receipt["submission_id"] != "" {
		t.Fatal("read mutated persisted evidence")
	}
	if repairHistoricalSubmitReceipt(entry, "other")["remote_receipt"].(map[string]any)["submission_id"] != "" {
		t.Fatal("other book receipt leaked")
	}
}

func TestWebSubmitBlockedBeforeUploadStillHasTaskLog(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutDocument(context.Background(), "owner", novelfetchworkshop.Document{BookID: "blocked", Meta: map[string]any{}})
	_, err := executeWebSubmit(context.Background(), "owner", t.TempDir(), store, webSubmitRequest{TaskIDs: []string{"blocked"}}, nil)
	if err == nil {
		t.Fatal("missing configuration should block")
	}
	doc, _ := store.GetDocument(context.Background(), "owner", "blocked")
	if len(workshopSubmitDiagnostic(doc)["logs"].([]any)) != 1 {
		t.Fatal("preflight failure has no visible task log")
	}
}

func TestWorkshopSubmitDiagnosticReadsDurableWorkerLogs(t *testing.T) {
	doc := novelfetchworkshop.Document{Meta: map[string]any{"siteSubmitStatus": "submitted", "siteSubmitDoneVersions": []any{"ai1"}}, Logs: []any{
		map[string]any{"event": "rewrite_failed", "error": "历史模型错误"},
		map[string]any{"event": "auto_submit_started", "time": "2026-11-13T04:44:10Z"},
		map[string]any{"event": "site_submit", "status": "submitted", "version": "ai1", "time": "2026-11-13T04:44:11Z", "remote_receipt": map[string]any{"verified": true, "submission_id": "287175"}},
	}}
	got := workshopSubmitDiagnostic(doc)
	logs := got["logs"].([]any)
	if len(logs) != 2 {
		t.Fatalf("submit log lost or unrelated error included: %#v", logs)
	}
	result := got["result"].(map[string]any)
	if result["status"] != "submitted" || mapOrEmpty(result["remote_receipt"])["verified"] != true {
		t.Fatalf("receipt lost: %#v", result)
	}
}

func TestWorkshopSubmitDiagnosticExplainsNoAttemptAndLegacyHistory(t *testing.T) {
	empty := workshopSubmitDiagnostic(novelfetchworkshop.Document{Meta: map[string]any{"aiStatus": "failed", "error": "文本模型不可用"}})
	if empty["result"] != nil || empty["reason"] != "文本模型不可用" {
		t.Fatalf("empty attempt appears successful or unexplained: %#v", empty)
	}
	legacy := workshopSubmitDiagnostic(novelfetchworkshop.Document{Meta: map[string]any{"websiteSubmitHistory": []any{map[string]any{"status": "failed", "error": "未登录"}}}})
	if len(legacy["logs"].([]any)) != 1 {
		t.Fatal("legacy history lost")
	}
}
