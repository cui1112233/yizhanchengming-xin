package batchfactory

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type JobStore interface {
	CreateJob(context.Context, pipeline.Job) error
}

type JobQueue interface {
	Enqueue(context.Context, pipeline.Job) error
}

type StartInput struct {
	IntakeID string
	RunAt    time.Time
}

type StartService struct {
	Jobs  JobStore
	Queue JobQueue
	Now   func() time.Time
}

func (s StartService) Start(ctx context.Context, input StartInput) (pipeline.Job, error) {
	if s.Jobs == nil || s.Queue == nil {
		return pipeline.Job{}, errors.New("job store and queue are required")
	}
	intakeID := strings.TrimSpace(input.IntakeID)
	if intakeID == "" {
		return pipeline.Job{}, errors.New("intake id is required")
	}
	plan := pipeline.BuildPlan(pipeline.PlanInput{RunAt: input.RunAt})
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	job := pipeline.NewIntakeJob(intakeID, plan, now)
	if err := s.Jobs.CreateJob(ctx, job); err != nil {
		return pipeline.Job{}, err
	}
	if err := s.Queue.Enqueue(ctx, job); err != nil {
		return pipeline.Job{}, err
	}
	return job, nil
}
