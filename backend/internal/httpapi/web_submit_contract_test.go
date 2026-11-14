package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

type contractChangedConnector struct {
	configuredSubmitConnector
	healthy bool
	probes  int
}

func (c *contractChangedConnector) Submit(context.Context, videomanagementconnector.Session, videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	c.submitCalls++
	return videomanagementconnector.Submission{}, videomanagementconnector.ErrSiteContractChanged
}

func (c *contractChangedConnector) ProbeSubmissionContract(context.Context, videomanagementconnector.Session) error {
	c.probes++
	if !c.healthy {
		return videomanagementconnector.ErrSiteContractChanged
	}
	return nil
}

func TestContractPauseProbesReadOnlyAndRecoversReceiptWithoutResubmission(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	dir := t.TempDir()
	queue := novelfetchworkshop.NewMemoryQueue()
	if err := store.PutConfig(ctx, "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true}}}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"12345678", "22345678"} {
		doc := readyIndependentUploadDocument(id)
		doc.Meta["platformId"] = "2"
		if err := store.PutDocument(ctx, "tester", doc); err != nil {
			t.Fatal(err)
		}
		if _, err := store.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{BookID: id, BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1", State: "ready"}, Content: "她走进书店，与多年未见的朋友重逢。"}); err != nil {
			t.Fatal(err)
		}
	}
	session := map[string]any{"cookie": "fixture", "targetUsername": "external"}
	if err := saveUploadTargetSession(dir, "tester", session); err != nil {
		t.Fatal(err)
	}
	connector := &contractChangedConnector{}
	request := webSubmitRequest{TaskIDs: []string{"12345678", "22345678"}, Versions: []string{"ai1"}}
	_, err := executeWebSubmit(ctx, "tester", dir, store, request, nil, connector)
	if workshopSubmitBlockedStatus(err) != "waiting_configuration" || connector.submitCalls != 1 {
		t.Fatalf("circuit did not stop siblings: %v calls=%d", err, connector.submitCalls)
	}
	doc, _ := store.GetDocument(ctx, "tester", "12345678")
	if len(mapOrEmpty(doc.Meta["siteSubmitPendingConfirmation"])) != 1 || novelFetchAutomationEligible(doc, time.Now()) {
		t.Fatalf("lost pending receipt or remained eligible: %#v", doc.Meta)
	}
	// A fresh login must not clear a structural failure.
	if err := resumeWorkshopUploadsAfterLogin(ctx, store, "tester"); err != nil {
		t.Fatal(err)
	}
	_, err = executeWebSubmit(ctx, "tester", dir, store, request, nil, connector)
	if err == nil || connector.submitCalls != 1 {
		t.Fatal("login bypassed contract circuit")
	}
	recoverWorkshopSubmitAccount(ctx, store, queue, "tester", dir, connector)
	if connector.probes != 0 {
		t.Fatal("probed before cooldown")
	}
	makeDue := func() {
		cfg, _ := store.GetConfig(ctx, "tester")
		mapOrEmpty(cfg[workshopSubmitPauseKey])["next_probe_at"] = time.Now().Add(-time.Second).Format(time.RFC3339Nano)
		if err := store.PutConfig(ctx, "tester", cfg); err != nil {
			t.Fatal(err)
		}
	}
	makeDue()
	recoverWorkshopSubmitAccount(ctx, store, queue, "tester", dir, connector)
	if connector.probes != 1 || connector.submitCalls != 1 {
		t.Fatal("failed probe sent a write")
	}
	connector.healthy = true
	makeDue()
	recoverWorkshopSubmitAccount(ctx, store, queue, "tester", dir, connector)
	doc, _ = store.GetDocument(ctx, "tester", "12345678")
	if doc.Meta["siteSubmitStatus"] != "pending_confirmation" {
		t.Fatalf("did not retain receipt: %#v", doc.Meta)
	}
	request.TaskIDs = []string{"12345678"}
	result, err := executeWebSubmit(ctx, "tester", dir, store, request, nil, connector)
	if err != nil || result.Success != 1 || connector.submitCalls != 1 || connector.readbackCalls != 1 {
		t.Fatalf("receipt recovery resubmitted: %v %#v calls=%d reads=%d", err, result, connector.submitCalls, connector.readbackCalls)
	}
}

func TestExternalSubmitScopeSurvivesOwnersCookiesAndLegacySessions(t *testing.T) {
	a := workshopExternalSubmitScope(map[string]any{"baseUrl": "https://TWO.121w.com/tttadmin/", "targetUsername": "alice", "cookie": "old"})
	b := workshopExternalSubmitScope(map[string]any{"baseUrl": "http://two.121w.com", "targetUsername": "alice", "cookie": "new"})
	legacy := workshopExternalSubmitScope(map[string]any{})
	if a != b || a != legacy {
		t.Fatal("same target site can acquire independent write scopes")
	}
}

func TestWorkerContractFailureDoesNotImmediatelyRequeueOrRegenerate(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	doc := readyIndependentUploadDocument("12345678")
	doc.Versions["ai1"] = "已完成的正文。"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), autoSubmit: func(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
		err := workshopContractBlockedError()
		markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, job.BookIDs, err)
		return err
	}}
	worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "contract-wait", Owner: "tester", BookIDs: []string{doc.BookID}, SelectedVersions: []string{"ai1"}, AutoSubmit: true})
	if _, err := queue.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatal("contract failure immediately requeued")
	}
	stored, _ := store.GetDocument(ctx, "tester", doc.BookID)
	if stored.Meta["status"] != "waiting_configuration" || novelFetchAutomationEligible(stored, time.Now()) || stored.Versions["ai1"] != doc.Versions["ai1"] {
		t.Fatalf("contract wait lost state: %#v", stored.Meta)
	}
}
