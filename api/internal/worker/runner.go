package worker

import (
	"context"
	"errors"
	"fmt"

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

type ProgressStore interface {
	CompleteStage(context.Context, string, pipeline.Stage) error
	CompleteJob(context.Context, string) error
}

type JobQueue interface {
	Enqueue(context.Context, pipeline.Job) error
}

type Runner struct {
	Source   JobSource
	Jobs     JobLoader
	Executor StageExecutor
	Progress ProgressStore
	Queue    JobQueue
}

func (r Runner) RunOnce(ctx context.Context) error {
	if r.Source == nil || r.Jobs == nil || r.Executor == nil || r.Progress == nil || r.Queue == nil {
		return errors.New("worker source, job loader, executor, progress store and queue are required")
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
		return errors.New("pipeline job has no pending stages")
	}
	stage := job.Stages[0]
	if err := r.Executor.Execute(ctx, job, stage); err != nil {
		if queueErr := r.Queue.Enqueue(ctx, job); queueErr != nil {
			return fmt.Errorf("execute stage %s: %v; requeue failed: %w", stage, err, queueErr)
		}
		return err
	}
	if err := r.Progress.CompleteStage(ctx, job.ID, stage); err != nil {
		return err
	}
	if len(job.Stages) == 1 {
		return r.Progress.CompleteJob(ctx, job.ID)
	}
	return r.Queue.Enqueue(ctx, job)
}
