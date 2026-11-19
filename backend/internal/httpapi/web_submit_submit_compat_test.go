package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

type configuredSubmitConnector struct {
	submitCalls   int
	readbackCalls int
	lastInput     videomanagementconnector.SubmissionInput
}

type uncertainSubmitConnector struct {
	submitCalls   int
	readbackCalls int
}

type recordingSubmitConnector struct {
	submitCalls int
	lastInput   videomanagementconnector.SubmissionInput
}

func (c *recordingSubmitConnector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, nil
}
func (c *recordingSubmitConnector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}
func (c *recordingSubmitConnector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}
func (c *recordingSubmitConnector) Submit(_ context.Context, _ videomanagementconnector.Session, input videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submitCalls++
	c.lastInput = input
	return videomanagementconnector.Submission{ID: "remote-1", Status: "submitted"}, nil
}
func (c *recordingSubmitConnector) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{ID: "remote-1", Status: "submitted", Message: "已确认", ReadbackRaw: map[string]any{"success": true}}, nil
}

func (c *uncertainSubmitConnector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, nil
}
func (c *uncertainSubmitConnector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}
func (c *uncertainSubmitConnector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}
func (c *uncertainSubmitConnector) Submit(context.Context, videomanagementconnector.Session, videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submitCalls++
	return videomanagementconnector.Submission{ID: "remote-uncertain", Status: "submitted"}, nil
}
func (c *uncertainSubmitConnector) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	c.readbackCalls++
	if c.readbackCalls == 1 {
		return videomanagementconnector.Submission{}, context.DeadlineExceeded
	}
	return videomanagementconnector.Submission{ID: "remote-uncertain", Status: "submitted", Message: "已确认"}, nil
}

func (c *configuredSubmitConnector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, nil
}
func (c *configuredSubmitConnector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}
func (c *configuredSubmitConnector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}
func (c *configuredSubmitConnector) Submit(_ context.Context, _ videomanagementconnector.Session, input videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submitCalls++
	c.lastInput = input
	if c.submitCalls == 1 {
		return videomanagementconnector.Submission{}, context.DeadlineExceeded
	}
	return videomanagementconnector.Submission{ID: "remote-1", Status: "submitted"}, nil
}
func (c *configuredSubmitConnector) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	c.readbackCalls++
	return videomanagementconnector.Submission{ID: "remote-1", Status: "submitted", Message: "已确认", ReadbackRaw: map[string]any{"success": true}}, nil
}

func TestResolveSubmissionGenderUsesConfiguredJudgeForMissingGender(t *testing.T) {
	doc := novelfetchworkshop.Document{
		ProcessedText: "用于分类的测试文本",
		Meta:          map[string]any{},
	}

	resolveSubmissionGender(context.Background(), &doc, func(context.Context, novelfetchworkshop.BookInfo) (string, error) {
		return `{"gender":"女频","confidence":0.9,"reason":"正文情感关系线突出"}`, nil
	})

	if doc.Gender != "女频" || doc.GenderSource != novelfetchworkshop.GenderAI {
		t.Fatalf("gender resolution was not persisted on the document: %#v", doc)
	}
	if doc.Meta["gender"] != "女频" || doc.Meta["genderSource"] != "ai" {
		t.Fatalf("gender resolution was not mirrored to metadata: %#v", doc.Meta)
	}
}

func TestMarkSubmissionFailurePersistsVersionAndReason(t *testing.T) {
	meta := map[string]any{"siteSubmitDoneVersions": []any{"original"}}
	markSubmissionFailure(meta, "ai1", "请为该本选择性别和风格")

	if meta["siteSubmitStatus"] != "pending_upload" {
		t.Fatalf("status = %#v, want pending_upload", meta["siteSubmitStatus"])
	}
	if got := anyStrings(meta["siteSubmitFailedVersions"]); len(got) != 1 || got[0] != "ai1" {
		t.Fatalf("failed versions = %#v, want [ai1]", got)
	}
	errors := mapOrEmpty(meta["siteSubmitFailedErrors"])
	if errors["ai1"] != "请为该本选择性别和风格" {
		t.Fatalf("failed error = %#v", errors)
	}
}

func TestSubmissionGenderStyleFallsBackToDocumentWithoutInventingStyle(t *testing.T) {
	doc := novelfetchworkshop.Document{
		Gender:   "女频",
		BookInfo: novelfetchworkshop.BookInfo{Genre: "现代"},
		Meta:     map[string]any{},
	}

	gender, style := submissionGenderStyle(doc)
	if gender != "女" {
		t.Fatalf("gender = %q, want 女", gender)
	}
	if style != "" {
		t.Fatalf("style = %q, want empty until classification exists", style)
	}
}

func TestSubmissionGenderStylePrefersSavedMetadata(t *testing.T) {
	doc := novelfetchworkshop.Document{
		Gender: "男频",
		Meta: map[string]any{
			"gender": "女频",
			"style":  "现代悬疑",
		},
	}

	gender, style := submissionGenderStyle(doc)
	if gender != "女" || style != "现代悬疑" {
		t.Fatalf("metadata was not preferred: gender=%q style=%q", gender, style)
	}
}

func TestSyncNovelFetchClassificationMetadataPersistsResolvedGender(t *testing.T) {
	doc := novelfetchworkshop.Document{Gender: "女频", GenderSource: novelfetchworkshop.GenderAI, GenderReason: "AI 判断", Meta: map[string]any{}}
	syncNovelFetchClassificationMetadata(&doc)
	if got := doc.Meta["gender"]; got != "女频" {
		t.Fatalf("meta gender = %#v, want 女频", got)
	}
	if got := doc.Meta["genderSource"]; got != "ai" {
		t.Fatalf("meta genderSource = %#v, want ai", got)
	}
}

func TestStyleNeedsResolutionTreatsPendingDisplayValueAsUnknown(t *testing.T) {
	if !styleNeedsResolution(map[string]any{"style": "待判定"}) {
		t.Fatal("待判定 must trigger style resolution")
	}
	if styleNeedsResolution(map[string]any{"style": "现代悬疑"}) {
		t.Fatal("a supported style must not trigger a second AI call")
	}
}

func TestRecordNovelFetchStyleDecisionPersistsFailure(t *testing.T) {
	meta := map[string]any{"style": "待判定"}
	recordNovelFetchStyleDecision(meta, novelfetchworkshop.StyleDecision{Reason: "AI 未返回有效风格"}, nil)
	if meta["styleStatus"] != "failed" {
		t.Fatalf("styleStatus = %#v, want failed", meta["styleStatus"])
	}
	if meta["styleError"] != "AI 未返回有效风格" {
		t.Fatalf("styleError = %#v", meta["styleError"])
	}
}

func TestRecordNovelFetchStyleDecisionPersistsSuccessAndClearsFailure(t *testing.T) {
	meta := map[string]any{"styleError": "旧错误", "styleStatus": "failed"}
	recordNovelFetchStyleDecision(meta, novelfetchworkshop.StyleDecision{Style: "现代悬疑", Source: "ai", Reason: "AI 判断", Confidence: 0.9}, nil)
	if meta["style"] != "现代悬疑" || meta["styleStatus"] != "classified" {
		t.Fatalf("style result = %#v", meta)
	}
	if _, ok := meta["styleError"]; ok {
		t.Fatalf("styleError should be cleared: %#v", meta)
	}
}

func TestInferNovelFetchStyleFailureFromExistingSubmitError(t *testing.T) {
	meta := map[string]any{
		"siteSubmitFailedErrors": map[string]any{"ai1": "该任务尚未识别或选择上传风格，无法提交到视频管理系统。"},
	}
	inferNovelFetchStyleFailure(meta)
	if meta["styleStatus"] != "failed" || meta["styleError"] == nil {
		t.Fatalf("style failure was not inferred: %#v", meta)
	}
}

func TestWebSubmitPreviewBlocksUnclassifiedStyle(t *testing.T) {
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "700001", Gender: "女频", Meta: map[string]any{"style": "待判定", "platformId": "1"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "700001", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "正文"}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}}
	RegisterWebSubmitPreviewCompat(mux, middleware, store)
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/preview", stringsReader(`{"taskIds":["700001"],"versions":["ai1"]}`))
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Groups  []any            `json:"groups"`
		Skipped []map[string]any `json:"skipped"`
		Plan    map[string]any   `json:"plan"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Groups) != 0 || len(payload.Skipped) != 1 {
		t.Fatalf("payload=%s", res.Body.String())
	}
	if payload.Plan == nil {
		t.Fatalf("preview must include the plan consumed by the React page: %s", res.Body.String())
	}
	if got := payload.Skipped[0]; got["status"] != "waiting_classify" || got["error"] != "无法映射网站字段：风格类型" {
		t.Fatalf("skipped=%#v", got)
	}
}

func TestWebSubmitAppliesSavedProfileAndDoesNotResendAfterUncertainSubmit(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "retry_times": 1, "min_file_size": 0, "skip_submitted": false,
		"organization": "org-9", "selected_profile": "profile-x", "profile_bindings": map[string]any{"ai1": "profile-y"},
		"upload_profiles": []any{map[string]any{"id": "profile-y", "platform_id": "6", "organization": "org-profile", "advanced": map[string]any{"ziti": 4}}},
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "123456", Gender: "女频", Meta: map[string]any{"style": "现代悬疑", "platformId": "1", "siteSubmitDoneVersions": []any{"ai1"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "123456", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &configuredSubmitConnector{}
	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}}
	RegisterWebSubmitSubmitCompat(mux, middleware, usersDir, store, nil, nil, connector)
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/submit", stringsReader(`{"taskIds":["123456"],"versions":["ai1"]}`))
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK || connector.submitCalls != 1 || connector.readbackCalls != 0 {
		t.Fatalf("status=%d submits=%d readbacks=%d body=%s", res.Code, connector.submitCalls, connector.readbackCalls, res.Body.String())
	}
	if fields := connector.lastInput.FormFields; fields["organization"] != "org-9" || fields["platform_id"] != "6" || fields["ziti"] != "4" {
		t.Fatalf("saved submit settings were not applied: %#v", fields)
	}
	var payload struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil || len(payload.Results) != 1 || payload.Results[0]["status"] != "pending_confirmation" {
		t.Fatalf("payload=%s err=%v", res.Body.String(), err)
	}
	if payload.Results[0]["operation_id"] == "" || payload.Results[0]["profile"] != "profile-y" {
		t.Fatalf("submission audit is incomplete: %#v", payload.Results[0])
	}
	document, err := store.GetDocument(context.Background(), "tester", "123456")
	if err != nil || mapOrEmpty(document.Meta["siteSubmitAttempts"])["ai1"] == nil {
		t.Fatalf("pre-send attempt was not persisted: document=%#v err=%v", document, err)
	}
	second, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"123456"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || second.Success != 1 || connector.submitCalls != 1 || connector.readbackCalls != 1 {
		t.Fatalf("uncertain retry must only read back: result=%#v err=%v submits=%d readbacks=%d", second, err, connector.submitCalls, connector.readbackCalls)
	}
}

func TestNovelFetchClassificationGateKeepsUnresolvedTasksOutOfAutomaticRewrite(t *testing.T) {
	for _, document := range []novelfetchworkshop.Document{
		{Gender: "待判定", Meta: map[string]any{"style": "现代悬疑"}},
		{Gender: "女频", Meta: map[string]any{"style": "待判定"}},
	} {
		if novelFetchClassificationResolved(document) {
			t.Fatalf("unresolved classification must pause automatic rewrite: %#v", document)
		}
	}
	if !novelFetchClassificationResolved(novelfetchworkshop.Document{Gender: "女频", Meta: map[string]any{"style": "现代悬疑"}}) {
		t.Fatal("confirmed gender and style should allow automatic rewrite")
	}
	if !novelFetchClassificationResolved(novelfetchworkshop.Document{Gender: "女频", GenderConfidence: 0.3, Meta: map[string]any{"style": "现代悬疑"}}) {
		t.Fatal("a valid gender and style should not be blocked by a confidence score")
	}
}

func TestNovelFetchProcessingFingerprintUsesNormalizedRules(t *testing.T) {
	canonical := novelFetchProcessingFingerprint(map[string]any{"sensitive": map[string]any{"enabled": true}, "knowledgeLibraries": map[string]any{"tone": []any{"清晰"}}})
	legacy := novelFetchProcessingFingerprint(map[string]any{"appConfig": map[string]any{
		"sensitive": map[string]any{"enabled": true}, "knowledge": map[string]any{"tone": []any{"清晰"}},
	}})
	changed := novelFetchProcessingFingerprint(map[string]any{"sensitive": map[string]any{"enabled": true}, "knowledgeLibraries": map[string]any{"tone": []any{"克制"}}})
	if canonical != legacy {
		t.Fatalf("equivalent normalized settings differ: %s != %s", canonical, legacy)
	}
	if canonical == changed {
		t.Fatal("meaningful processing rule change did not change fingerprint")
	}
}

func TestWebSubmitUncertainOutcomeOnlyReadsBackAndNeverUploadsTwice(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "retry_times": 0, "skip_submitted": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "654321", Gender: "女频", Meta: map[string]any{"style": "现代悬疑", "platformId": "1"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "654321", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &uncertainSubmitConnector{}
	first, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"654321"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || first.Success != 1 || first.Failed != 0 || connector.submitCalls != 1 {
		t.Fatalf("first result=%#v err=%v submits=%d", first, err, connector.submitCalls)
	}
	document, _ := store.GetDocument(context.Background(), "tester", "654321")
	if document.Meta["siteSubmitStatus"] != "submitted" {
		t.Fatalf("submitted upload status=%#v", document.Meta)
	}
	// 已确认的版本再次提交时必须被跳过，绝不能重复上传。
	second, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"654321"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || connector.submitCalls != 1 || second.Success != 0 {
		t.Fatalf("second result=%#v err=%v submits=%d", second, err, connector.submitCalls)
	}
}

func TestWebSubmitBlocksUnclassifiedStyleAndPersistsFailureLog(t *testing.T) {
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true, "min_file_size": 0}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "700002", Gender: "女频", Meta: map[string]any{"style": "未判定", "platformId": "1"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "700002", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "正文"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}}
	RegisterWebSubmitSubmitCompat(mux, middleware, usersDir, store, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/submit", stringsReader(`{"taskIds":["700002"],"versions":["ai1"]}`))
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		OK            bool             `json:"ok"`
		Result        string           `json:"result"`
		SuccessGroups int              `json:"success_groups"`
		FailedGroups  int              `json:"failed_groups"`
		Results       []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.OK || payload.Result != "all_ok" || payload.SuccessGroups != 0 || payload.FailedGroups != 0 {
		t.Fatalf("deferred classification must not be reported as a hard failure: %s", res.Body.String())
	}
	if len(payload.Results) != 1 || payload.Results[0]["error"] != "无法映射网站字段：风格类型" {
		t.Fatalf("payload=%s", res.Body.String())
	}
	if payload.Results[0]["status"] != "waiting_classify" || payload.Results[0]["deferred"] != true {
		t.Fatalf("blocked item must be deferred, not failed: %s", res.Body.String())
	}
	doc, err := store.GetDocument(context.Background(), "tester", "700002")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Logs) != 1 {
		t.Fatalf("logs=%#v", doc.Logs)
	}
	log := doc.Logs[0].(map[string]any)
	if log["event"] != "site_submit" || log["version"] != "ai1" || log["status"] != "waiting_classify" || log["error"] != "无法映射网站字段：风格类型" || log["externalRequested"] != false || log["time"] == "" {
		t.Fatalf("log=%#v", log)
	}
	if doc.Meta["siteSubmitStatus"] != "waiting_classify" {
		t.Fatalf("site status must wait for classification: %#v", doc.Meta)
	}
}

func TestWebSubmitOverallResultLabels(t *testing.T) {
	cases := []struct {
		success int
		failed  int
		want    string
	}{
		{success: 2, failed: 0, want: "all_ok"},
		{success: 1, failed: 1, want: "partial"},
		{success: 0, failed: 2, want: "all_failed"},
	}
	for _, tc := range cases {
		if got := webSubmitOverallResult(tc.success, tc.failed); got != tc.want {
			t.Fatalf("success=%d failed=%d got=%s want=%s", tc.success, tc.failed, got, tc.want)
		}
	}
}

func TestDecorateWebSubmitResultsAddsFailureCodes(t *testing.T) {
	items := []any{
		map[string]any{"bookId": "ok", "status": "submitted"},
		map[string]any{"bookId": "bad", "status": "error", "error": "正文不足"},
		map[string]any{"bookId": "pending", "status": "pending_confirmation"},
	}
	decorated := decorateWebSubmitResults(items)
	if len(decorated) != 3 {
		t.Fatalf("decorated=%#v", decorated)
	}
	bad := decorated[1].(map[string]any)
	if bad["code"] != "WEB_SUBMIT_FAILED" || bad["message"] != "正文不足" {
		t.Fatalf("bad=%#v", bad)
	}
	pending := decorated[2].(map[string]any)
	if pending["code"] != "WEB_SUBMIT_PENDING_CONFIRMATION" || strings.TrimSpace(toString(pending["message"])) == "" {
		t.Fatalf("pending=%#v", pending)
	}
	// 同一套中文状态词汇：接口直接给出 status_label，页面不再显示英文原始状态。
	if got := decorated[0].(map[string]any)["status_label"]; got != "已提交" {
		t.Fatalf("submitted label=%#v", got)
	}
	if got := bad["status_label"]; got != "失败" {
		t.Fatalf("failed label=%#v", got)
	}
	if got := pending["status_label"]; got != "待远端核验" {
		t.Fatalf("pending label=%#v", got)
	}
	if got := webSubmitStatusText("waiting_configuration"); got != "等待配置" {
		t.Fatalf("waiting label=%#v", got)
	}
}

func TestWebSubmitDryRunPersistsPlanManifest(t *testing.T) {
	now := time.Date(2026, 10, 18, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "123456", Gender: "女频", Meta: map[string]any{"style": "现代悬疑", "platformId": "1"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "123456", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: now.Add(time.Hour)}}}
	RegisterWebSubmitSubmitCompat(mux, middleware, usersDir, store, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/submit", stringsReader(`{"taskIds":["123456"],"versions":["ai1"],"dry_run":true}`))
	req.Header.Set("Authorization", "Bearer test-session")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		DryRun          bool             `json:"dry_run"`
		Requested       int              `json:"requested"`
		GroupCount      int              `json:"group_count"`
		SubmissionRunID string           `json:"submission_run_id"`
		PlannedItems    []map[string]any `json:"planned_items"`
		DeferredItems   []map[string]any `json:"deferred_items"`
		MissingItems    []map[string]any `json:"missing_items"`
		Versions        []string         `json:"versions"`
		PlanComplete    bool             `json:"plan_complete"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.DryRun || payload.Requested != 1 || payload.GroupCount != 1 || payload.SubmissionRunID == "" {
		t.Fatalf("dry_run payload=%s", res.Body.String())
	}
	if len(payload.PlannedItems) != 1 || len(payload.DeferredItems) != 0 || len(payload.MissingItems) != 0 || !payload.PlanComplete {
		t.Fatalf("plan accounting is incomplete: %s", res.Body.String())
	}
	if len(payload.Versions) != 1 || payload.Versions[0] != "ai1" {
		t.Fatalf("versions=%#v", payload.Versions)
	}
	raw, err := os.ReadFile(filepath.Join(usersDir, "tester", "web-submit", "manifests", payload.SubmissionRunID+".json"))
	if err != nil {
		t.Fatalf("dry_run manifest was not persisted: %v", err)
	}
	var manifest struct {
		Status         string           `json:"status"`
		Reason         string           `json:"reason"`
		RequestedCount int              `json:"requested_count"`
		Items          []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Status != "dry_run" || manifest.Reason != "dry_run" || manifest.RequestedCount != 1 {
		t.Fatalf("manifest=%s", string(raw))
	}
	if len(manifest.Items) != 1 || manifest.Items[0]["state"] != "ready" {
		t.Fatalf("manifest items=%#v", manifest.Items)
	}
}

func TestWebSubmitExecutingVersionOnlyReadsBack(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "654321", Gender: "女频", Meta: map[string]any{
		"style": "现代悬疑", "platformId": "1", "siteSubmitExecutingVersions": []any{"ai1"},
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "654321", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &configuredSubmitConnector{}
	result, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"654321"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || connector.submitCalls != 0 || connector.readbackCalls != 1 || result.Success != 1 {
		t.Fatalf("executing version must only read back: result=%#v err=%v submits=%d readbacks=%d", result, err, connector.submitCalls, connector.readbackCalls)
	}
	document, _ := store.GetDocument(context.Background(), "tester", "654321")
	if containsVersion(document.Meta["siteSubmitExecutingVersions"], "ai1") || !containsVersion(document.Meta["siteSubmitDoneVersions"], "ai1") {
		t.Fatalf("executing state was not resolved: %#v", document.Meta)
	}
}

func TestWebSubmitForceResubmitsConfirmedVersion(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "skip_submitted": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "777777", Gender: "女频", Meta: map[string]any{
		"style": "现代悬疑", "platformId": "1", "siteSubmitDoneVersions": []any{"ai1"},
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "777777", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &uncertainSubmitConnector{}
	request := webSubmitRequest{TaskIDs: []string{"777777"}, Versions: []string{"ai1"}, Force: true}
	result, err := executeWebSubmit(context.Background(), "tester", usersDir, store, request, nil, connector)
	if err != nil || connector.submitCalls != 1 || result.Success != 1 {
		t.Fatalf("force must resubmit confirmed version: result=%#v err=%v submits=%d", result, err, connector.submitCalls)
	}
	request.Force = false
	result, err = executeWebSubmit(context.Background(), "tester", usersDir, store, request, nil, connector)
	if err != nil || connector.submitCalls != 1 || result.Success != 0 || len(result.Results) != 1 {
		t.Fatalf("without force confirmed version must be skipped: result=%#v err=%v submits=%d", result, err, connector.submitCalls)
	}
	if row, _ := result.Results[0].(map[string]any); row["status"] != "skipped" {
		t.Fatalf("result=%#v", result.Results[0])
	}
}

func TestWebSubmitUsesSourceBookIDForUpload(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "owner-777", Gender: "女频", Meta: map[string]any{
		"style": "现代悬疑", "platformId": "1", "source_book_id": "888888",
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "owner-777", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文内容，继续描述情节发展。"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &recordingSubmitConnector{}
	result, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"owner-777"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || connector.submitCalls != 1 || result.Success != 1 {
		t.Fatalf("result=%#v err=%v submits=%d", result, err, connector.submitCalls)
	}
	if connector.lastInput.BookID != "888888" || connector.lastInput.Filename != "888888.txt" {
		t.Fatalf("upload must use source book id: %#v", connector.lastInput)
	}
}

func TestWebSubmitRejectsInvalidSiteBookID(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "owner-777", Gender: "女频", Meta: map[string]any{
		"style": "现代悬疑", "platformId": "1",
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "owner-777", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "满足字数的正文内容，继续描述情节发展。"}); err != nil {
		t.Fatal(err)
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &recordingSubmitConnector{}
	result, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"owner-777"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || connector.submitCalls != 0 || len(result.Results) != 1 {
		t.Fatalf("result=%#v err=%v submits=%d", result, err, connector.submitCalls)
	}
	row, _ := result.Results[0].(map[string]any)
	if row["status"] != "waiting_config" || row["reason"] != "invalid_site_book_id" {
		t.Fatalf("row=%#v", row)
	}
}

func TestWebSubmitSkipsDuplicateContent(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{
		"enabled": true, "min_file_size": 0, "organization": "org-1",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{BookID: "777777", Gender: "女频", Meta: map[string]any{
		"style": "现代悬疑", "platformId": "1", "siteSubmitDoneVersions": []any{"ai1"},
	}}); err != nil {
		t.Fatal(err)
	}
	for version, content := range map[string]string{
		"ai1": "满足字数的正文内容\r\n继续描述情节发展。",
		"ai2": "满足字数的正文内容\n继续描述情节发展。",
	} {
		if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "777777", BodyRef: novelfetchworkshop.BodyRef{VersionID: version}, Content: content}); err != nil {
			t.Fatal(err)
		}
	}
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	connector := &recordingSubmitConnector{}
	result, err := executeWebSubmit(context.Background(), "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"777777"}, Versions: []string{"ai1", "ai2"}}, nil, connector)
	if err != nil || connector.submitCalls != 0 {
		t.Fatalf("duplicate content must not upload: result=%#v err=%v submits=%d", result, err, connector.submitCalls)
	}
	found := false
	for _, item := range result.Results {
		if row, ok := item.(map[string]any); ok && row["reason"] == "duplicate_content" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing duplicate_content result: %#v", result.Results)
	}
}

func stringsReader(value string) *strings.Reader { return strings.NewReader(value) }
