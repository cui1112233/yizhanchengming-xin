package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type Store interface {
	GetIntake(ctx context.Context, id int64) (intake.Intake, error)
	CreateBatchProject(ctx context.Context, project intake.BatchProject) (intake.BatchProject, error)
	CreateRun(ctx context.Context, run intake.Run) (intake.Run, error)
}

type idempotentRunStore interface {
	CreateRunIdempotent(ctx context.Context, run intake.Run, idempotencyKey string) (intake.Run, error)
}

type Clock func() time.Time

type Service struct {
	store Store
	now   Clock
}

type CreateRequest struct {
	IntakeID       int64
	Name           string
	RunAt          time.Time
	IdempotencyKey string
}

type CreateResult struct {
	Project intake.BatchProject `json:"project"`
	Run     intake.Run          `json:"run"`
}

func NewService(store Store, now Clock) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, now: now}
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (CreateResult, error) {
	if s.store == nil {
		return CreateResult{}, fmt.Errorf("pipeline store is required")
	}
	if request.IntakeID <= 0 {
		return CreateResult{}, fmt.Errorf("intakeId 必须大于 0")
	}

	intakeValue, err := s.store.GetIntake(ctx, request.IntakeID)
	if err != nil {
		return CreateResult{}, fmt.Errorf("读取 intake: %w", err)
	}
	if intakeValue.Status != intake.StatusCompleted {
		return CreateResult{}, fmt.Errorf("intake 状态为 %s，只有 completed 才能创建批量项目", intakeValue.Status)
	}

	now := s.now().UTC()
	runAt := request.RunAt
	immediate := runAt.IsZero()
	if immediate {
		runAt = now
	} else {
		runAt = runAt.UTC()
		if !runAt.After(now) {
			return CreateResult{}, fmt.Errorf("run_at 必须晚于当前时间")
		}
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = strings.TrimSpace(intakeValue.Name)
	}
	if name == "" {
		name = fmt.Sprintf("批量项目-%d", request.IntakeID)
	}

	project, err := s.store.CreateBatchProject(ctx, intake.BatchProject{
		IntakeID: request.IntakeID,
		Name:     name,
	})
	if err != nil {
		return CreateResult{}, fmt.Errorf("创建批量项目: %w", err)
	}

	runInput := intake.Run{BatchProjectID: project.ID, RunAt: runAt, Status: intake.RunStatusPending}
	if idempotent, ok := s.store.(idempotentRunStore); ok {
		key := strings.TrimSpace(request.IdempotencyKey)
		if key == "" {
			if immediate {
				key = fmt.Sprintf("pipeline:intake:%d:immediate", request.IntakeID)
			} else {
				key = fmt.Sprintf("pipeline:intake:%d:scheduled:%s", request.IntakeID, runAt.Format(time.RFC3339Nano))
			}
		}
		run, err := idempotent.CreateRunIdempotent(ctx, runInput, key)
		if err != nil {
			return CreateResult{}, fmt.Errorf("创建执行记录: %w", err)
		}
		return CreateResult{Project: project, Run: run}, nil
	}

	run, err := s.store.CreateRun(ctx, runInput)
	if err != nil {
		return CreateResult{}, fmt.Errorf("创建执行记录: %w", err)
	}
	return CreateResult{Project: project, Run: run}, nil
}
