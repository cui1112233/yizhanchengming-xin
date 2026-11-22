package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
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

func TestWebSubmitConfirmedVersionWithUnchangedFingerprintIsNotUploaded(t *testing.T) {
	store, usersDir, doc := newSubmissionEvidenceTask(t)
	connector := &recordingSubmitConnector{}
	result, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{
		TaskIDs: []string{doc.BookID}, Versions: []string{"ai1", "ai2"},
	}, nil, connector)
	if err != nil || result.Failed != 0 || len(result.Results) != 2 || connector.submitCalls != 1 || string(connector.lastInput.Content) != "第二版完整正文" {
		t.Fatalf("only ai2 should upload: result=%+v calls=%d input=%+v error=%v", result, connector.submitCalls, connector.lastInput, err)
	}
	if row := mapOrEmpty(result.Results[0]); row["version"] != "ai1" || row["status"] != "skipped" || row["reason"] != "confirmed_result" {
		t.Fatalf("confirmed ai1 was not skipped: %+v", row)
	}
}

func TestWebSubmitUnknownSubmissionOnlyReadsBackEvenWhenForced(t *testing.T) {
	for _, status := range []string{"pending_verification", "pending_confirmation"} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			store, usersDir, doc := newSubmissionEvidenceTask(t)
			doc.Meta["siteSubmitStatus"] = status
			doc.Meta["siteSubmitFailedVersions"] = []string{"ai2"}
			if err := store.PutDocument(ctx, "tester", doc); err != nil {
				t.Fatal(err)
			}
			connector := &configuredSubmitConnector{}
			result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{
				TaskIDs: []string{doc.BookID}, Versions: []string{"ai2"}, Force: true,
			}, nil, connector)
			if err != nil || result.Success != 0 || result.Failed != 1 || len(result.Results) != 1 || mapOrEmpty(result.Results[0])["status"] != "pending_confirmation" || connector.submitCalls != 0 || connector.readbackCalls != 1 {
				t.Fatalf("unattributed receipt must stay pending: result=%+v uploads=%d readbacks=%d error=%v", result, connector.submitCalls, connector.readbackCalls, err)
			}
			saved, err := store.GetDocument(ctx, "tester", doc.BookID)
			if err != nil || containsVersion(saved.Meta["siteSubmitDoneVersions"], "ai2") || saved.Meta["siteSubmitStatus"] != "pending_confirmation" {
				t.Fatalf("unattributed ai2 receipt changed persisted evidence: %+v %v", saved, err)
			}
		})
	}
}

func TestSubmitReadbackOperationMustIdentifyOneVersion(t *testing.T) {
	doc := novelfetchworkshop.Document{Meta: map[string]any{"siteSubmitAttempts": map[string]any{
		"ai1": map[string]any{"operation_id": "shared-attempt"},
		"ai2": map[string]any{"operation_id": "shared-attempt"},
	}}}
	for _, scenario := range []struct {
		name, operation, version string
		want                     bool
	}{
		{name: "shared operation alone", operation: "shared-attempt"},
		{name: "different version", operation: "shared-attempt", version: "ai1"},
		{name: "different operation", operation: "another-attempt", version: "ai2"},
		{name: "operation and version match", operation: "shared-attempt", version: "ai2", want: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			receipt := videomanagementconnector.Submission{ID: "candidate-receipt", ReadbackRaw: map[string]any{
				"operation_id": scenario.operation, "version_id": scenario.version,
			}}
			if got := workshopPendingReadbackMatches(receipt, doc, "12345678901", "ai2"); got != scenario.want {
				t.Fatalf("ai2 receipt attribution = %v, want %v", got, scenario.want)
			}
		})
	}
}

func TestSubmitReadbackRejectsExplicitOperationMismatchDespiteMatchingEvidence(t *testing.T) {
	for _, pendingID := range []string{"12345678901", "candidate-receipt"} {
		t.Run(pendingID, func(t *testing.T) {
			fingerprint := "current-ai2-fingerprint"
			doc := novelfetchworkshop.Document{Meta: map[string]any{
				"siteSubmitPendingConfirmation": map[string]any{"ai2": pendingID},
				"siteSubmitAttempts":            map[string]any{"ai2": map[string]any{"operation_id": "operation-b", "content_sha256": fingerprint}},
			}}
			receipt := videomanagementconnector.Submission{ID: "candidate-receipt", ReadbackRaw: map[string]any{
				"bookid": "12345678901", "version_id": "ai2", "operation_id": "operation-a", "content_sha256": fingerprint,
			}}
			if workshopPendingReadbackMatches(receipt, doc, "12345678901", "ai2") {
				t.Fatal("a conflicting operation identifier was overridden by weaker evidence")
			}
		})
	}
}

func TestWebSubmitForceAuditsOperationAndPriorReceipt(t *testing.T) {
	ctx := context.Background()
	store, usersDir, doc := newSubmissionEvidenceTask(t)
	connector := &cloneSubmissionEvidenceSpy{}
	result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{
		TaskIDs: []string{doc.BookID}, Versions: []string{"ai1"}, OperationID: "force-operation", Force: true,
	}, nil, connector)
	if err != nil || result.Success != 1 || connector.submitCalls != 0 || connector.cloneSource != "first-ai1-receipt" || connector.content != "第一版完整正文" {
		t.Fatalf("force did not create a second receipt: result=%+v connector=%+v error=%v", result, connector, err)
	}
	saved, err := store.GetDocument(ctx, "tester", doc.BookID)
	if err != nil {
		t.Fatal(err)
	}
	if priorWorkshopSubmitReceipt(saved, "ai1") != "cloned-ai1-receipt" {
		t.Fatalf("second receipt not stored: %+v", saved.Logs)
	}
	diagnostic := workshopSubmitDiagnostic(saved)
	for _, value := range arrayOrEmpty(diagnostic["logs"]) {
		entry := mapOrEmpty(value)
		if entry["operation_id"] == "force-operation" && entry["version"] == "ai1" && entry["force"] == true && entry["prior_receipt_id"] == "first-ai1-receipt" {
			return
		}
	}
	t.Fatalf("force audit lacks operation and prior receipt: %+v", diagnostic["logs"])
}

func TestWebSubmitDisabledSkipUsesExistingSubmitAndAuditsPriorReceipt(t *testing.T) {
	ctx := context.Background()
	store, usersDir, doc := newSubmissionEvidenceTask(t)
	cfg, err := store.GetConfig(ctx, "tester")
	if err != nil {
		t.Fatal(err)
	}
	mapOrEmpty(mapOrEmpty(cfg["appConfig"])["web_submit"])["skip_submitted"] = false
	if err := store.PutConfig(ctx, "tester", cfg); err != nil {
		t.Fatal(err)
	}
	connector := &cloneSubmissionEvidenceSpy{}
	result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{
		TaskIDs: []string{doc.BookID}, Versions: []string{"ai1"}, OperationID: "submit-again-operation",
	}, nil, connector)
	if err != nil || result.Success != 1 || connector.submitCalls != 1 || connector.cloneSource != "" {
		t.Fatalf("skip setting was not respected: result=%+v connector=%+v error=%v", result, connector, err)
	}
	saved, err := store.GetDocument(ctx, "tester", doc.BookID)
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range saved.Logs {
		entry := mapOrEmpty(raw)
		if entry["operation_id"] == "submit-again-operation" && entry["prior_receipt_id"] == "first-ai1-receipt" && entry["skip_submitted"] == false {
			return
		}
	}
	t.Fatalf("disabled skip audit missing: %+v", saved.Logs)
}

func TestWebSubmitForceFailsWhenCloneIsUnavailable(t *testing.T) {
	ctx := context.Background()
	store, usersDir, doc := newSubmissionEvidenceTask(t)
	connector := &recordingSubmitConnector{}
	result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{
		TaskIDs: []string{doc.BookID}, Versions: []string{"ai1"}, Force: true,
	}, nil, connector)
	if err != nil || result.Failed != 1 || result.Success != 0 || len(result.Results) != 1 || firstText(mapOrEmpty(result.Results[0])["error"]) == "" || connector.submitCalls != 0 {
		t.Fatalf("unsupported forced resend was not blocked: result=%+v uploads=%d error=%v", result, connector.submitCalls, err)
	}
	saved, err := store.GetDocument(ctx, "tester", doc.BookID)
	if err != nil || priorWorkshopSubmitReceipt(saved, "ai1") != "first-ai1-receipt" || saved.Meta["siteSubmitStatus"] != "submitted" {
		t.Fatalf("blocked forced resend lost old confirmation: %+v %v", saved, err)
	}
}

func TestWebSubmitForceDoesNotGuessPriorVersionFromHistoricalListing(t *testing.T) {
	ctx := context.Background()
	store, usersDir, doc := newSubmissionEvidenceTask(t)
	doc.Logs = []any{map[string]any{
		"event": "site_submit", "status": "submitted", "version": "ai1", "time": "2026-11-13T11:24:43Z",
		"remote_receipt": map[string]any{"verified": true, "remote_record": map[string]any{"data": []any{
			map[string]any{"id": "ai2-receipt", "bookid": doc.BookID, "addtime": "2026-11-13 19:26:09"},
		}}},
	}}
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	connector := &cloneSubmissionEvidenceSpy{}
	result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{
		TaskIDs: []string{doc.BookID}, Versions: []string{"ai1"}, OperationID: "force-ai1", Force: true,
	}, nil, connector)
	if err != nil || result.Success != 0 || result.Failed != 1 || len(result.Results) != 1 || mapOrEmpty(result.Results[0])["code"] != "WEB_SUBMIT_RESUBMISSION_UNAVAILABLE" || connector.cloneCalls != 0 || connector.submitCalls != 0 {
		t.Fatalf("force used an unattributed historical receipt: result=%+v connector=%+v error=%v", result, connector, err)
	}
}

type cloneSubmissionEvidenceSpy struct {
	recordingSubmitConnector
	cloneSource string
	content     string
	cloneCalls  int
	receipt     *videomanagementconnector.Submission
}

func (c *cloneSubmissionEvidenceSpy) CloneSubmission(_ context.Context, _ videomanagementconnector.Session, source string, input videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.cloneSource, c.content = source, string(input.Content)
	c.cloneCalls++
	if c.receipt != nil {
		return *c.receipt, nil
	}
	return videomanagementconnector.Submission{ID: "cloned-ai1-receipt", Status: "submitted"}, nil
}

func newSubmissionEvidenceTask(t *testing.T) (*novelfetchworkshop.MemoryStore, string, novelfetchworkshop.Document) {
	t.Helper()
	const owner = "tester"
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(ctx, owner, map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	doc := novelfetchworkshop.Document{BookID: "12345678901", Gender: "女频", Meta: map[string]any{
		"status": "completed", "style": "现代悬疑", "platformId": "1", "siteSubmitStatus": "submitted",
		"siteSubmitDoneVersions": []string{"ai1"}, "autoSubmitReady": true,
	}, Logs: []any{
		map[string]any{"event": "site_submit", "status": "submitted", "version": "ai1", "operation_id": "first-operation",
			"remote_receipt": map[string]any{"verified": true, "submission_id": "first-ai1-receipt"}},
	}}
	if err := store.PutDocument(ctx, owner, doc); err != nil {
		t.Fatal(err)
	}
	for version, content := range map[string]string{"ai1": "第一版完整正文", "ai2": "第二版完整正文"} {
		if _, err := store.PutBody(ctx, owner, novelfetchworkshop.BodyRecord{BookID: doc.BookID, BodyRef: novelfetchworkshop.BodyRef{VersionID: version}, Content: content}); err != nil {
			t.Fatal(err)
		}
	}
	var err error
	doc, err = store.GetDocument(ctx, owner, doc.BookID)
	if err != nil {
		t.Fatal(err)
	}
	doc.Meta["autoSubmitReadyFingerprint"] = novelFetchUploadReadyFingerprint(ctx, store, owner, doc, []string{"ai1", "ai2"})
	if err := store.PutDocument(ctx, owner, doc); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, owner), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, owner, "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	return store, usersDir, doc
}
