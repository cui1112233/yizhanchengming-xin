package batchfactory

import (
    "context"
    "testing"
    "time"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeJobStore struct{ saved pipeline.Job }
func (f *fakeJobStore) CreateJob(_ context.Context, job pipeline.Job) error { f.saved = job; return nil }

type fakeQueue struct{ enqueued pipeline.Job }
func (f *fakeQueue) Enqueue(_ context.Context, job pipeline.Job) error { f.enqueued = job; return nil }

func TestStartUsesOnePipelineForImmediateAndScheduled(t *testing.T) {
    store := &fakeJobStore{}
    queue := &fakeQueue{}
    svc := StartService{Jobs: store, Queue: queue}

    runAt := time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC)
    job, err := svc.Start(context.Background(), StartInput{IntakeID: "intake-1", RunAt: runAt})
    if err != nil { t.Fatal(err) }
    if job.IntakeID != "intake-1" { t.Fatalf("intake id = %q", job.IntakeID) }
    if !job.RunAt.Equal(runAt) { t.Fatalf("run_at = %v", job.RunAt) }
    want := []pipeline.Stage{pipeline.StageFetchBook, pipeline.StageResolveMetadata, pipeline.StageAIClassify, pipeline.StageCreateBatch}
    if len(job.Stages) != len(want) { t.Fatalf("unexpected stages: %#v", job.Stages) }
    for i := range want { if job.Stages[i] != want[i] { t.Fatalf("unexpected stages: %#v", job.Stages) } }
    if store.saved.ID == "" || queue.enqueued.ID != store.saved.ID { t.Fatalf("job was not durably saved then enqueued: saved=%#v queued=%#v", store.saved, queue.enqueued) }
}
