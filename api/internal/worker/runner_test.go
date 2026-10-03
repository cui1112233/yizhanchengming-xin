package worker

import (
    "context"
    "errors"
    "testing"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeSource struct{ id string }
func (f *fakeSource) Next(context.Context) (string, error) { return f.id, nil }

type fakeLoader struct{ job pipeline.Job }
func (f *fakeLoader) LoadJob(_ context.Context, id string) (pipeline.Job, error) {
    if id != f.job.ID { return pipeline.Job{}, ErrJobNotFound }
    return f.job, nil
}

type fakeExecutor struct{ called pipeline.Stage; err error }
func (f *fakeExecutor) Execute(_ context.Context, _ pipeline.Job, stage pipeline.Stage) error { f.called = stage; return f.err }

type fakeProgress struct {
    stage pipeline.Stage
    jobDone bool
}
func (f *fakeProgress) CompleteStage(_ context.Context, _ string, stage pipeline.Stage) error { f.stage = stage; return nil }
func (f *fakeProgress) CompleteJob(_ context.Context, _ string) error { f.jobDone = true; return nil }

type fakeRequeue struct{ job pipeline.Job }
func (f *fakeRequeue) Enqueue(_ context.Context, job pipeline.Job) error { f.job = job; return nil }

func TestRunnerCompletesStageAndRequeuesWhenMoreWorkRemains(t *testing.T) {
    source := &fakeSource{id:"job-1"}
    loader := &fakeLoader{job:pipeline.Job{ID:"job-1", IntakeID:"intake-1", Status:"queued", Stages:[]pipeline.Stage{pipeline.StageFetchBook, pipeline.StageResolveMetadata}}}
    exec := &fakeExecutor{}
    progress := &fakeProgress{}
    requeue := &fakeRequeue{}
    runner := Runner{Source:source, Jobs:loader, Executor:exec, Progress:progress, Queue:requeue}

    if err := runner.RunOnce(context.Background()); err != nil { t.Fatal(err) }
    if exec.called != pipeline.StageFetchBook { t.Fatalf("executed=%q", exec.called) }
    if progress.stage != pipeline.StageFetchBook { t.Fatalf("completed=%q", progress.stage) }
    if requeue.job.ID != "job-1" { t.Fatalf("job not requeued: %#v", requeue.job) }
    if progress.jobDone { t.Fatal("job should not be complete") }
}

func TestRunnerCompletesJobAfterLastStage(t *testing.T) {
    source := &fakeSource{id:"job-1"}
    loader := &fakeLoader{job:pipeline.Job{ID:"job-1", IntakeID:"intake-1", Status:"queued", Stages:[]pipeline.Stage{pipeline.StageCreateBatch}}}
    exec := &fakeExecutor{}
    progress := &fakeProgress{}
    requeue := &fakeRequeue{}
    runner := Runner{Source:source, Jobs:loader, Executor:exec, Progress:progress, Queue:requeue}

    if err := runner.RunOnce(context.Background()); err != nil { t.Fatal(err) }
    if !progress.jobDone { t.Fatal("job should be complete") }
    if requeue.job.ID != "" { t.Fatalf("finished job must not be requeued: %#v", requeue.job) }
}

func TestRunnerRequeuesPendingJobWhenStageExecutionFails(t *testing.T) {
    source := &fakeSource{id:"job-1"}
    loader := &fakeLoader{job:pipeline.Job{ID:"job-1", IntakeID:"intake-1", Status:"queued", Stages:[]pipeline.Stage{pipeline.StageFetchBook}}}
    exec := &fakeExecutor{err: errors.New("upstream failed")}
    progress := &fakeProgress{}
    requeue := &fakeRequeue{}
    runner := Runner{Source:source, Jobs:loader, Executor:exec, Progress:progress, Queue:requeue}

    err := runner.RunOnce(context.Background())
    if err == nil { t.Fatal("expected execution error") }
    if requeue.job.ID != "job-1" { t.Fatalf("failed job was lost instead of requeued: %#v", requeue.job) }
    if progress.stage != "" || progress.jobDone { t.Fatalf("failed stage must remain pending: %#v", progress) }
}
