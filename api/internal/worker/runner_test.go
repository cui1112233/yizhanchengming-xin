package worker

import (
    "context"
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

type fakeExecutor struct{ called pipeline.Stage }
func (f *fakeExecutor) Execute(_ context.Context, _ pipeline.Job, stage pipeline.Stage) error { f.called = stage; return nil }

func TestRunnerLoadsDurableJobBeforeExecuting(t *testing.T) {
    source := &fakeSource{id:"job-1"}
    loader := &fakeLoader{job:pipeline.Job{ID:"job-1", BatchID:"batch-1", Status:"queued", Stages:[]pipeline.Stage{pipeline.StageFetchBook, pipeline.StageResolveMetadata}}}
    exec := &fakeExecutor{}
    runner := Runner{Source:source, Jobs:loader, Executor:exec}

    if err := runner.RunOnce(context.Background()); err != nil { t.Fatal(err) }
    if exec.called != pipeline.StageFetchBook { t.Fatalf("executed=%q", exec.called) }
}
