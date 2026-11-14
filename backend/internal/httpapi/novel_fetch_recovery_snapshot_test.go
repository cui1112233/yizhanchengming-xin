package httpapi

import (
	"context"
	"qiantie/backend/internal/novelfetchworkshop"
	"reflect"
	"testing"
	"time"
)

func TestNovelFetchCancelledBooksDoNotSpinAutomation(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	q := novelfetchworkshop.NewMemoryQueue()
	s.PutConfig(ctx, "alice", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": true}})
	s.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "waiting_retry", "siteSubmitStatus": "failed"}})
	q.CancelBooks(ctx, "alice", []string{"12345678901"})
	for i := 0; i < 3; i++ {
		job, err := dispatchNovelFetchAutomationOwner(ctx, s, q, "alice", time.Now())
		if err != nil || job != nil {
			t.Fatalf("cancelled task dispatched: job=%v err=%v", job, err)
		}
	}
	d, _ := s.GetDocument(ctx, "alice", "12345678901")
	if firstText(d.Meta["status"]) != "stopped" {
		t.Fatal(d.Meta)
	}
	if novelFetchAutomationEligible(d, time.Now()) {
		t.Fatal("stopped upload failure must remain stopped")
	}
	result, err := enqueueWorkshopRetries(ctx, s, q, "alice", []string{d.BookID}, false)
	if err != nil || result.Queued != 1 {
		t.Fatalf("explicit retry failed: %+v %v", result, err)
	}
	cancelled, _ := q.IsBookCancelled(ctx, "alice", d.BookID)
	if cancelled {
		t.Fatal("explicit retry must clear cancellation")
	}
}

func TestNovelFetchRecoveryPreservesNonconsecutiveVersionsAndGroups(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	q := novelfetchworkshop.NewMemoryQueue()
	s.PutConfig(ctx, "alice", map[string]any{novelFetchAutomationConfigKey: map[string]any{"enabled": true}, "text_model_id": "changed-model"})
	for i, id := range []string{"12345678901", "12345678902"} {
		s.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: id, MaxTxt: 4000, Meta: map[string]any{"status": "waiting_retry"}})
		job := novelfetchworkshop.ProcessJob{Owner: "alice", BookIDs: []string{id}, BatchID: []string{"batch1", "batch2"}[i], ScheduleID: []string{"schedule1", "schedule2"}[i], PlatformID: []string{"2", "3"}[i], MaxTxt: []int{1234, 2345}[i], SelectedVersions: []string{"ai3"}, TextModelID: "original-model", RewriteMethods: map[string]string{"ai3": "high_imitation"}, SensitiveAIEnabled: true, AutoSubmit: true}
		if err := saveNovelFetchJobSnapshot(ctx, s, job); err != nil {
			t.Fatal(err)
		}
	}
	job, err := dispatchNovelFetchAutomationOwner(ctx, s, q, "alice", time.Now())
	if err != nil || job == nil {
		t.Fatalf("%v %v", job, err)
	}
	if len(job.BookIDs) != 1 || !reflect.DeepEqual(job.SelectedVersions, []string{"ai3"}) || job.TextModelID != "original-model" || job.ScheduleID == "" || job.RewriteMethods["ai3"] != "high_imitation" {
		t.Fatalf("snapshot lost or incompatible books grouped: %+v", job)
	}
	doc, _ := s.GetDocument(ctx, "alice", job.BookIDs[0])
	want := novelFetchRecoveryJob(doc, map[string]any{}, "alice", true)
	if job.MaxTxt != want.MaxTxt || job.PlatformID != want.PlatformID || job.BatchID != want.BatchID {
		t.Fatalf("parameters drifted: %+v vs %+v", job, want)
	}
}

func TestNovelFetchManualRetryKeepsOriginalVersionSelection(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	q := novelfetchworkshop.NewMemoryQueue()
	s.PutConfig(ctx, "alice", map[string]any{"text_model_id": "new-model"})
	s.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "12345678901", Meta: map[string]any{"status": "failed", "aiCount": 5}})
	saveNovelFetchJobSnapshot(ctx, s, novelfetchworkshop.ProcessJob{Owner: "alice", BookIDs: []string{"12345678901"}, BatchID: "old-batch", PlatformID: "2", MaxTxt: 3000, SelectedVersions: []string{"ai3"}, TextModelID: "old-model"})
	result, err := enqueueWorkshopRetries(ctx, s, q, "alice", []string{"12345678901"}, false)
	if err != nil || result.Queued != 1 {
		t.Fatalf("%+v %v", result, err)
	}
	job, err := q.Claim(ctx, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(job.SelectedVersions, []string{"ai3"}) || job.TextModelID != "old-model" || job.MaxTxt != 3000 {
		t.Fatalf("retry drift: %+v", job)
	}
}

func TestNovelFetchWorkerPassesSelectedModelToResolver(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	q := novelfetchworkshop.NewMemoryQueue()
	s.PutDocument(ctx, "alice", novelfetchworkshop.Document{BookID: "12345678901", OriginalRaw: "saved", Original: "saved", ProcessedText: "saved", Meta: map[string]any{"originalStatus": "done"}})
	seen := ""
	w := workshopProcessWorker{store: s, queue: q, resolveRewriteProvider: func(ctx context.Context, _ string) (novelfetchworkshop.RewriteProvider, error) {
		seen = novelfetchworkshop.ProcessTextModel(ctx)
		return nil, nil
	}}
	// Resolve during sensitive processing before classification; no external call.
	_ = w.execute(ctx, novelfetchworkshop.ProcessJob{Owner: "alice", BookIDs: []string{"12345678901"}, TextModelID: "chosen-model", SensitiveAIEnabled: true})
	if seen != "chosen-model" {
		t.Fatalf("model=%q", seen)
	}
}
