package pipeline

import (
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type Stage string

const (
	StageFetchBook       Stage = "fetch_book"
	StageResolveMetadata Stage = "resolve_metadata"
	StageAIClassify      Stage = "ai_classify"
	StageCreateBatch     Stage = "create_batch"
)

type PlanInput struct {
	Gender   novel.GenderResult
	HasStyle bool
	RunAt    time.Time
}

type Plan struct {
	RunAt  time.Time
	Stages []Stage
}

func BuildPlan(input PlanInput) Plan {
	stages := []Stage{StageFetchBook, StageResolveMetadata}
	if input.Gender.Gender == novel.GenderUnknown || !input.HasStyle {
		stages = append(stages, StageAIClassify)
	}
	stages = append(stages, StageCreateBatch)
	return Plan{RunAt: input.RunAt, Stages: stages}
}
