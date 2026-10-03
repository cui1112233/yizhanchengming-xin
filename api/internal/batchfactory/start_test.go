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

func TestStartUsesSamePipelineForImmediateAndScheduled(t *testing.T) {
    store := &fakeJobStore{}
    queue := &fakeQueue{}
    svc := StartService{Jobs: store, Queue: queue}

    runAt := time.Date(2026, 10, 4, 9, 0, 0, 0, time.UTC)
    job, err := svc.Start(context.Background(), StartInput{IntakeID: "intake-1", RunAt: runAt, NeedsAI: true})
    if err != nil { t.Fatal(err) }
    if job.IntakeID != "intake-1" { t.Fatalf("intake id = %q", job.IntakeID) }
    if !job.RunAt.Equal(runAt) { t.Fatalf("run_at = %v", job.RunAt) }
    if len(job.Stages) != 4 || job.Stages[0] != pipeline.StageFetchBook || job.Stages[1] != pipeline.StageResolveMetadata || job.Stages[2] != pipeline.StageAIClassify || job.Stages[3] != pipeline.StageCreateBatch {
        t.Fatalf("unexpected stages: %#v", job.Stages)
    }
    if store.saved.ID == "" || queue.enqueued.ID != store.saved.ID { t.Fatalf("job was not durably saved then enqueued: saved=%#v queued=%#v", store.saved, queue.enqueued) }
}

func TestStartSkipsAIWhenMetadataComplete(t *testing.T) {
    store := &fakeJobStore{}
    queue := &fakeQueue{}
    svc := StartService{Jobs: store, Queue: queue}

    job, err := svc.Start(context.Background(), StartInput{IntakeID: "intake-2", NeedsAI: false})
    if err != nil { t.Fatal(err) }
    for _, stage := range job.Stages {
        if stage == pipeline.StageAIClassify { t.Fatalf("AI stage should be skipped: %#v", job.Stages) }
    }
}
