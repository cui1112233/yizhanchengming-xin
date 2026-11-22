package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

type leaseProbeQueue struct {
	novelfetchworkshop.ProcessQueue
	readOwner                    func() bool
	renewed                      chan time.Duration
	renewErr                     error
	renewAfter                   <-chan struct{}
	owns                         bool
	acks, fails, ownershipChecks int
}

func (q *leaseProbeQueue) RenewLease(ctx context.Context, job novelfetchworkshop.ProcessJob, ttl time.Duration) error {
	select {
	case <-q.renewAfter:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case q.renewed <- ttl:
		return q.renewErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *leaseProbeQueue) OwnsLease(context.Context, novelfetchworkshop.ProcessJob) (bool, error) {
	q.ownershipChecks++
	if q.readOwner != nil {
		return q.readOwner(), nil
	}
	return q.owns, nil
}
func (q *leaseProbeQueue) Ack(context.Context, novelfetchworkshop.ProcessJob) error {
	q.acks++
	return nil
}
func (q *leaseProbeQueue) Fail(context.Context, novelfetchworkshop.ProcessJob, bool) error {
	q.fails++
	return nil
}

func TestWorkshopWorkerRenewsLeaseUntilCompletion(t *testing.T) { testWorkshopWorkerLease(t, false) }

func TestWorkshopWorkerCancelsWorkAfterLeaseLoss(t *testing.T) { testWorkshopWorkerLease(t, true) }

func testWorkshopWorkerLease(t *testing.T, lose bool) {
	t.Helper()
	parent, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	store := novelfetchworkshop.NewMemoryStore()
	const id = "33333333333"
	doc := novelfetchworkshop.Document{BookID: id, Gender: "男频", Original: "已经抓取的正文", ProcessedText: "已经抓取的正文", Meta: map[string]any{"originalStatus": "done", "style": "历史"}}
	if err := store.PutDocument(parent, "tester", doc); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	q := &leaseProbeQueue{ProcessQueue: novelfetchworkshop.NewMemoryQueue(), renewed: make(chan time.Duration, 4), owns: true, renewAfter: entered}
	if lose {
		q.renewErr = novelfetchworkshop.ErrProcessLeaseLost
	}
	uploads := 0
	worker := workshopProcessWorker{store: store, queue: q, registry: newWorkshopProcessRegistry(), leaseHeartbeat: time.Millisecond,
		autoSubmit: func(context.Context, novelfetchworkshop.ProcessJob) error { uploads++; return nil },
		resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
			return rewriteProviderFunc(func(ctx context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
				close(entered)
				select {
				case <-release:
				case <-ctx.Done():
				}
				// A provider can still deliver a successful response after cancellation.
				return `{"id":"33333333333","changed_lines":[{"line_no":1,"text":"新生成的第二版"}]}`, nil
			}), nil
		},
	}
	done := make(chan error, 1)
	go func() {
		done <- worker.process(parent, novelfetchworkshop.ProcessJob{ID: "job", LeaseToken: "owned", BatchID: "batch", Owner: "tester", BookIDs: []string{id}, SelectedVersions: []string{"ai1"}, AutoSubmit: true})
	}()
	for i := 0; i < 2; i++ {
		select {
		case ttl := <-q.renewed:
			if ttl != 10*time.Minute {
				t.Fatalf("renewed TTL %v", ttl)
			}
		case <-parent.Done():
			t.Fatal("worker did not renew while model was blocked")
		}
		if lose {
			break
		}
	}
	if !lose {
		close(release)
	}
	var processErr error
	select {
	case processErr = <-done:
	case <-parent.Done():
		t.Fatal("worker did not finish")
	}
	body, err := store.GetBody(context.Background(), "tester", id, "ai1")
	persisted, _ := store.GetDocument(context.Background(), "tester", id)
	if lose {
		if !errors.Is(processErr, novelfetchworkshop.ErrProcessLeaseLost) || err == nil || persisted.Versions["ai1"] != nil || uploads != 0 || q.acks != 0 || q.fails != 0 {
			t.Fatalf("lease loss err=%v body=%q versions=%v uploads=%d ack=%d fail=%d", processErr, body.Content, persisted.Versions, uploads, q.acks, q.fails)
		}
	} else if processErr != nil || err != nil || body.Content == "" || uploads != 1 || q.acks != 1 || q.fails != 0 || q.ownershipChecks == 0 {
		t.Fatalf("completion err=%v body=%q uploads=%d ack=%d fail=%d ownership=%d", processErr, body.Content, uploads, q.acks, q.fails, q.ownershipChecks)
	}
}

func TestWorkshopWorkerChecksLeaseBeforeUpload(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	q := &leaseProbeQueue{ProcessQueue: novelfetchworkshop.NewMemoryQueue(), owns: false}
	const book = "book"
	if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: book}); err != nil {
		t.Fatal(err)
	}
	worker := workshopProcessWorker{store: store, queue: q, registry: newWorkshopProcessRegistry(), autoSubmit: func(context.Context, novelfetchworkshop.ProcessJob) error {
		t.Fatal("submitted after ownership changed")
		return nil
	}}
	err := worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "job", Owner: "tester", BatchID: "batch", BookIDs: []string{book}, LeaseToken: "old", AutoSubmit: true})
	if !errors.Is(err, novelfetchworkshop.ErrProcessLeaseLost) || q.ownershipChecks == 0 || q.acks != 0 || q.fails != 0 {
		t.Fatalf("err=%v checks=%d ack=%d fail=%d", err, q.ownershipChecks, q.acks, q.fails)
	}
	saved, _ := store.GetDocument(ctx, "tester", book)
	if saved.Meta["autoSubmitReady"] == true {
		t.Fatal("worker marked an upload ready after its lease moved")
	}
}

func TestWorkshopWorkerRechecksLeaseAfterPreparingUpload(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	const book = "book"
	if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: book}); err != nil {
		t.Fatal(err)
	}
	q := &leaseProbeQueue{ProcessQueue: novelfetchworkshop.NewMemoryQueue(), readOwner: func() bool {
		doc, _ := store.GetDocument(ctx, "tester", book)
		return doc.Meta["autoSubmitReady"] != true
	}}
	uploads := 0
	worker := workshopProcessWorker{store: store, queue: q, registry: newWorkshopProcessRegistry(), autoSubmit: func(context.Context, novelfetchworkshop.ProcessJob) error { uploads++; return nil }}
	err := worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "job", Owner: "tester", BatchID: "batch", BookIDs: []string{book}, LeaseToken: "old", AutoSubmit: true})
	if !errors.Is(err, novelfetchworkshop.ErrProcessLeaseLost) || uploads != 0 || q.acks != 0 || q.fails != 0 {
		t.Fatalf("err=%v uploads=%d acks=%d fails=%d", err, uploads, q.acks, q.fails)
	}
}
