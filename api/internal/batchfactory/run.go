package batchfactory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type RunIntakeCreator interface {
	Create(context.Context, CreateIntakeInput) (CreateIntakeResult, error)
}

type RunStarter interface {
	Start(context.Context, StartInput) (pipeline.Job, error)
}

type CreateRunInput struct {
	Intake CreateIntakeInput
	RunAt  time.Time
}

type RunResult struct {
	IntakeID  string    `json:"intake_id"`
	JobID     string    `json:"job_id,omitempty"`
	GroupCount int      `json:"group_count"`
	BookCount int       `json:"book_count"`
	RunAt     time.Time `json:"run_at,omitempty"`
	Status    string    `json:"status,omitempty"`
}

type RunService struct {
	Intakes RunIntakeCreator
	Starter RunStarter
}

func (s RunService) CreateAndStart(ctx context.Context, input CreateRunInput) (RunResult, error) {
	if s.Intakes == nil || s.Starter == nil {
		return RunResult{}, errors.New("intake creator and starter are required")
	}
	intake, err := s.Intakes.Create(ctx, input.Intake)
	if err != nil {
		return RunResult{}, err
	}
	result := RunResult{
		IntakeID: intake.IntakeID,
		GroupCount: intake.GroupCount,
		BookCount: intake.BookCount,
		RunAt: input.RunAt,
	}
	job, err := s.Starter.Start(ctx, StartInput{IntakeID: intake.IntakeID, RunAt: input.RunAt})
	if err != nil {
		return result, fmt.Errorf("start pipeline for intake %s: %w", intake.IntakeID, err)
	}
	result.JobID = job.ID
	result.RunAt = job.RunAt
	result.Status = job.Status
	return result, nil
}
