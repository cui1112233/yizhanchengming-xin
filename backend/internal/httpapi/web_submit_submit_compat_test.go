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
)

func TestResolveSubmissionGenderUsesConfiguredJudgeForMissingGender(t *testing.T) {
	doc := novelfetchworkshop.Document{
		ProcessedText: "用于分类的测试文本",
		Meta:          map[string]any{},
	}

	resolveSubmissionGender(context.Background(), &doc, func(context.Context, novelfetchworkshop.BookInfo) (string, error) {
		return "女频", nil
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
	recordNovelFetchStyleDecision(meta, novelfetchworkshop.StyleDecision{Style: "现代悬疑", Source: "ai", Reason: "AI 判断"}, nil)
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
		BookID: "book-preview", Gender: "女频", Meta: map[string]any{"style": "待判定"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "book-preview", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "正文"}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}}}
	RegisterWebSubmitPreviewCompat(mux, middleware, store)
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/preview", stringsReader(`{"taskIds":["book-preview"],"versions":["ai1"]}`))
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Groups  []any            `json:"groups"`
		Skipped []map[string]any `json:"skipped"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Groups) != 0 || len(payload.Skipped) != 1 {
		t.Fatalf("payload=%s", res.Body.String())
	}
	if got := payload.Skipped[0]; got["status"] != "blocked" || got["error"] != webSubmitUnclassifiedStyleError {
		t.Fatalf("skipped=%#v", got)
	}
}

func TestWebSubmitBlocksUnclassifiedStyleAndPersistsFailureLog(t *testing.T) {
	now := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutConfig(context.Background(), "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true}}}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "tester", novelfetchworkshop.Document{
		BookID: "book-submit", Gender: "女频", Meta: map[string]any{"style": "未判定"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutBody(context.Background(), "tester", novelfetchworkshop.BodyRecord{BookID: "book-submit", BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "正文"}); err != nil {
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
	middleware := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }, AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}}}
	RegisterWebSubmitSubmitCompat(mux, middleware, usersDir, store, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/batch-rewrite/web-submit/submit", stringsReader(`{"taskIds":["book-submit"],"versions":["ai1"]}`))
	SignBridgeRequest(req, "tester", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var payload struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Results) != 1 || payload.Results[0]["error"] != webSubmitUnclassifiedStyleError {
		t.Fatalf("payload=%s", res.Body.String())
	}
	doc, err := store.GetDocument(context.Background(), "tester", "book-submit")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Logs) != 1 {
		t.Fatalf("logs=%#v", doc.Logs)
	}
	log := doc.Logs[0].(map[string]any)
	if log["event"] != "site_submit" || log["version"] != "ai1" || log["status"] != "failed" || log["error"] != webSubmitUnclassifiedStyleError || log["externalRequested"] != false || log["time"] == "" {
		t.Fatalf("log=%#v", log)
	}
}

func stringsReader(value string) *strings.Reader { return strings.NewReader(value) }
