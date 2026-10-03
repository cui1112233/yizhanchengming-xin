package worker

import (
	"context"
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

var ErrJobNotFound = errors.New("pipeline job not found")

type JobSource interface {
	Next(context.Context) (string, error)
}

type JobLoader interface {
	LoadJob(context.Context, string) (pipeline.Job, error)
}

type StageExecutor interface {
	Execute(context.Context, pipeline.Job, pipeline.Stage) error
}

type Runner struct {
	Source   JobSource
	Jobs     JobLoader
	Executor StageExecutor
}

func (r Runner) RunOnce(ctx context.Context) error {
	if r.Source == nil || r.Jobs == nil || r.Executor == nil {
		return errors.New("worker source, job loader and executor are required")
	}
	jobID, err := r.Source.Next(ctx)
	if err != nil {
		return err
	}
	job, err := r.Jobs.LoadJob(ctx, jobID)
	if err != nil {
		return err
	}
	if len(job.Stages) == 0 {
		return errors.New("pipeline job has no stages")
	}
	return r.Executor.Execute(ctx, job, job.Stages[0])
}
