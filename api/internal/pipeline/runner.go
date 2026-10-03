package pipeline

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Stage string

const (
	StageFetchBook       Stage = "fetch_book"
	StageResolveMetadata Stage = "resolve_metadata"
	StageAIClassify      Stage = "ai_classify"
	StageCreateBatch     Stage = "create_batch"
)

type PlanInput struct {
	RunAt time.Time
}

type Plan struct {
	RunAt  time.Time
	Stages []Stage
}

type Job struct {
	ID        string    `json:"id"`
	IntakeID  string    `json:"intake_id"`
	BatchID   string    `json:"batch_id,omitempty"`
	RunAt     time.Time `json:"run_at"`
	Stages    []Stage   `json:"stages"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func BuildPlan(input PlanInput) Plan {
	return Plan{
		RunAt: input.RunAt,
		Stages: []Stage{StageFetchBook, StageResolveMetadata, StageAIClassify, StageCreateBatch},
	}
}

func NewIntakeJob(intakeID string, plan Plan, now time.Time) Job {
	return Job{
		ID:        newID(),
		IntakeID:  intakeID,
		RunAt:     plan.RunAt,
		Stages:    append([]Stage(nil), plan.Stages...),
		Status:    "queued",
		CreatedAt: now,
	}
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}
	return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
}
