package httpapi

import (
	"context"
	"errors"
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

type expiredAccountConnector struct {
	configuredSubmitConnector
	expired       bool
	sessionChecks int
}

func (c *expiredAccountConnector) CheckSession(context.Context, videomanagementconnector.Session) error {
	c.sessionChecks++
	if c.expired {
		return videomanagementconnector.ErrSessionExpired
	}
	return nil
}

func (c *expiredAccountConnector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{Cookie: "sid=new", TargetUsername: "fixture"}, nil
}

func (c *expiredAccountConnector) Submit(_ context.Context, _ videomanagementconnector.Session, _ videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submitCalls++
	if c.expired {
		return videomanagementconnector.Submission{}, videomanagementconnector.ErrSessionExpired
	}
	return videomanagementconnector.Submission{ID: "remote-1", Status: "submitted"}, nil
}

func TestExpiredUploadAccountPausesWithoutUnknownReceiptAndResumesAfterLogin(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	usersDir := t.TempDir()
	_ = store.PutConfig(ctx, "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true}}})
	for _, id := range []string{"12345678", "22345678"} {
		doc := readyIndependentUploadDocument(id)
		doc.Meta["platformId"] = "2"
		doc.Versions["ai1"] = "她走进书店，与多年未见的朋友重逢。"
		_ = store.PutDocument(ctx, "tester", doc)
		_, _ = store.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{BookID: id, BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1", State: "ready"}, Content: doc.Versions["ai1"].(string)})
	}
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := saveUploadTargetSession(usersDir, "tester", map[string]any{"cookie": "sid=old"}); err != nil {
		t.Fatal(err)
	}
	connector := &expiredAccountConnector{expired: true}
	_, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"12345678"}, Versions: []string{"ai1"}}, nil, connector)
	var blocked *webSubmitRequestError
	if !errors.As(err, &blocked) || connector.submitCalls != 1 || connector.readbackCalls != 0 {
		t.Fatalf("expiry not surfaced: %v calls=%d", err, connector.submitCalls)
	}
	doc, _ := store.GetDocument(ctx, "tester", "12345678")
	if doc.Meta["siteSubmitStatus"] != "waiting_login" || len(mapOrEmpty(doc.Meta["siteSubmitPendingConfirmation"])) != 0 {
		t.Fatalf("expiry incorrectly became uncertain upload: %#v", doc.Meta)
	}
	_, err = executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"22345678"}, Versions: []string{"ai1"}}, nil, connector)
	if err == nil || connector.submitCalls != 1 {
		t.Fatal("expired account continued uploading sibling book")
	}
	// Exercise the authenticated HTTP login route, including saved-session recovery.
	connector.expired = false
	mux := http.NewServeMux()
	middleware := BridgeAuth{Secret: "test", AuthStore: &authTestStore{account: auth.Account{Username: "tester", Active: true}, session: auth.Session{Username: "tester", Token: "test-session", ExpiresAt: time.Now().Add(time.Hour)}}}
	registerNovelFetchUploadSessionWithStore(mux, middleware, usersDir, store, connector)
	req := httptest.NewRequest(http.MethodPost, "/api/novel-fetch-workshop/upload/login", strings.NewReader(`{"username":"fixture","password":"fixture-only"}`))
	req.Header.Set("Authorization", "Bearer test-session")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("login route failed: %d %s", response.Code, response.Body)
	}
	doc, _ = store.GetDocument(ctx, "tester", "12345678")
	if !novelFetchAutomationEligible(doc, time.Now()) {
		t.Fatalf("login did not release saved upload: %#v", doc.Meta)
	}
	result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{"12345678"}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || result.Success != 1 || connector.submitCalls != 2 || connector.readbackCalls != 1 {
		t.Fatalf("resume failed: %v %#v", err, result)
	}
	doc, _ = store.GetDocument(ctx, "tester", "12345678")
	if doc.Versions["ai1"] != "她走进书店，与多年未见的朋友重逢。" {
		t.Fatal("login recovery changed completed prose")
	}
}

func TestWorkerLoginFailureDoesNotImmediatelyRequeueOrRegenerate(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	doc := readyIndependentUploadDocument("12345678")
	doc.Versions["ai1"] = "已经完成的正文。"
	_ = store.PutDocument(ctx, "tester", doc)
	calls := 0
	worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), autoSubmit: func(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
		calls++
		err := &webSubmitRequestError{Status: http.StatusConflict, Message: workshopLoginRequiredMessage, Code: "UPLOAD_LOGIN_REQUIRED"}
		markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, job.BookIDs, err)
		return err
	}}
	worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "login-wait", Owner: "tester", BookIDs: []string{doc.BookID}, SelectedVersions: []string{"ai1"}, AutoSubmit: true})
	if _, err := queue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatal("login failure was immediately retried")
	}
	stored, _ := store.GetDocument(ctx, "tester", doc.BookID)
	if calls != 1 || stored.Meta["status"] != "waiting_login" || stored.Meta["siteSubmitStatus"] != "waiting_login" || stored.Versions["ai1"] != doc.Versions["ai1"] {
		t.Fatalf("login wait lost state: calls=%d meta=%#v", calls, stored.Meta)
	}
}

func TestLoginRecoveryPreservesOtherOwnersAndStoppedBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	for _, owner := range []string{"tester", "other"} {
		for _, id := range []string{"waiting", "stopped"} {
			doc := readyIndependentUploadDocument(id)
			doc.Meta["siteSubmitStatus"] = "waiting_login"
			if id == "stopped" {
				doc.Meta["status"] = "stopped"
			}
			_ = store.PutDocument(ctx, owner, doc)
		}
	}
	if err := resumeWorkshopUploadsAfterLogin(ctx, store, "tester"); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"tester", "stopped"}, {"other", "waiting"}, {"other", "stopped"}} {
		doc, _ := store.GetDocument(ctx, pair[0], pair[1])
		if doc.Meta["siteSubmitStatus"] != "waiting_login" {
			t.Fatalf("login resumed unrelated work: %v", pair)
		}
	}
}

func TestPausedAccountProbeWaitsThenRecoversWithoutUpload(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	dir := t.TempDir()
	session := map[string]any{"cookie": "fixture-cookie"}
	_ = saveUploadTargetSession(dir, "tester", session)
	doc := readyIndependentUploadDocument("12345678")
	doc.Meta["siteSubmitStatus"] = "waiting_login"
	_ = store.PutDocument(ctx, "tester", doc)
	_ = pauseWorkshopSubmitAccount(ctx, store, "tester", session)
	connector := &expiredAccountConnector{expired: true}
	recoverWorkshopSubmitAccount(ctx, store, queue, "tester", dir, connector)
	if connector.sessionChecks != 0 {
		t.Fatal("paused account probed before cooldown")
	}
	makeDue := func() {
		cfg, _ := store.GetConfig(ctx, "tester")
		mapOrEmpty(cfg[workshopSubmitPauseKey])["next_probe_at"] = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano)
		_ = store.PutConfig(ctx, "tester", cfg)
	}
	makeDue()
	recoverWorkshopSubmitAccount(ctx, store, queue, "tester", dir, connector)
	if connector.sessionChecks != 1 || connector.submitCalls != 0 {
		t.Fatal("probe sent upload or skipped verification")
	}
	connector.expired = false
	makeDue()
	recoverWorkshopSubmitAccount(ctx, store, queue, "tester", dir, connector)
	stored, _ := store.GetDocument(ctx, "tester", doc.BookID)
	if connector.sessionChecks != 2 || connector.submitCalls != 0 || stored.Meta["siteSubmitStatus"] != "pending_upload" {
		t.Fatalf("verified account did not recover: %#v", stored.Meta)
	}
}
