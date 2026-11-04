package batchfactoryv11

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newResubmitTestService(t *testing.T) (*MemoryStore, Batch, Book, Video, *ProductionService) {
	t.Helper()
	store, batch, book, video := seedCompiledVideo(t)
	service := &ProductionService{
		Store: store, Compiler: &PromptCompilerService{Store: store}, Enabled: true,
		Adapter: &recordingProductionAdapter{ref: ProviderTaskRef{State: ProductionQueued}},
		Model:   FrozenVideoModel{ID: "yd2.0-mini", MaxDuration: 15},
	}
	return store, batch, book, video, service
}

func queueTaskWithoutReceipt(t *testing.T, store *MemoryStore, batch Batch, book Book, video Video, requestID string, queuedFor time.Duration) ProductionJob {
	t.Helper()
	job, err := store.CreateProductionJob(context.Background(), ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: requestID,
		DirectorRevisionID: book.DirectorRevision.ID,
		Tasks: []ProductionTask{{
			VideoID: video.ID, Status: ProductionQueued, Attempt: 1,
			FinalPromptHash: "queued-without-receipt", CompiledPrompt: "queued without receipt",
			UpdatedAt: time.Now().UTC().Add(-queuedFor),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return job
}

func TestResubmitQueuedWithoutReceiptOnceOnly(t *testing.T) {
	ctx := context.Background()

	// 1) Stale queued task without a provider id is re-submitted and the old
	// task is cancelled with the durable marker.
	store, batch, book, video, service := newResubmitTestService(t)
	queueTaskWithoutReceipt(t, store, batch, book, video, "queued-stale", 3*time.Minute)
	job, err := service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, "resubmit-1", VideoProviderPersonalAPI, ProductionOptions{})
	if err != nil {
		t.Fatalf("first resubmit should be allowed: %v", err)
	}
	if len(job.Tasks) != 1 {
		t.Fatalf("resubmitted job tasks = %+v", job.Tasks)
	}
	updated, err := store.FindProductionJob(ctx, "alice", batch.ID, book.ID, "queued-stale")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Tasks[0].Status != ProductionCancelled || updated.Tasks[0].ErrorMessage != SupersededWithoutReceiptMarker {
		t.Fatalf("superseded task = %+v", updated.Tasks[0])
	}

	// 2) The replacement task is again queued without a receipt and now stale:
	// the durable marker blocks a second re-submission for the same video.
	replacement := job.Tasks[0]
	replacement.UpdatedAt = time.Now().UTC().Add(-3 * time.Minute)
	if _, err := store.UpdateProductionTask(ctx, "alice", job.ID, replacement.ID, replacement); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, "resubmit-2", VideoProviderPersonalAPI, ProductionOptions{}); !errors.Is(err, ErrConflict) {
		t.Fatalf("second resubmit must be blocked, err=%v", err)
	}

	// 3) A queued task that already has a provider id is never re-submitted.
	store2, batch2, book2, video2, service2 := newResubmitTestService(t)
	withReceipt, err := store2.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch2.ID, BookID: book2.ID, RequestID: "queued-with-receipt",
		DirectorRevisionID: book2.DirectorRevision.ID,
		Tasks: []ProductionTask{{
			VideoID: video2.ID, Status: ProductionQueued, Attempt: 1, ProviderTaskID: "provider-receipt",
			FinalPromptHash: "receipt", CompiledPrompt: "receipt", UpdatedAt: time.Now().UTC().Add(-10 * time.Minute),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service2.SubmitBookProductionWithOptions(ctx, "alice", batch2.ID, book2.ID, "resubmit-receipt", VideoProviderPersonalAPI, ProductionOptions{}); !errors.Is(err, ErrConflict) {
		t.Fatalf("queued task with receipt must block resubmission, err=%v", err)
	}
	if withReceipt.ID == "" {
		t.Fatal("fixture job missing id")
	}

	// 4) A queued task inside the 2 minute grace period is not re-submitted.
	store3, batch3, book3, video3, service3 := newResubmitTestService(t)
	queueTaskWithoutReceipt(t, store3, batch3, book3, video3, "queued-fresh", 30*time.Second)
	if _, err := service3.SubmitBookProductionWithOptions(ctx, "alice", batch3.ID, book3.ID, "resubmit-early", VideoProviderPersonalAPI, ProductionOptions{}); !errors.Is(err, ErrConflict) {
		t.Fatalf("fresh queued task must block resubmission, err=%v", err)
	}
}
