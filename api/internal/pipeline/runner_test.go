package pipeline

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildPlanAlwaysIncludesConditionalAIStage(t *testing.T) {
	plan := BuildPlan(PlanInput{})
	want := []Stage{StageFetchBook, StageResolveMetadata, StageAIClassify, StageCreateBatch}
	if !reflect.DeepEqual(plan.Stages, want) {
		t.Fatalf("stages = %#v, want %#v", plan.Stages, want)
	}
}

func TestImmediateAndScheduledPlansUseSameStages(t *testing.T) {
	future := time.Date(2026, 10, 4, 8, 0, 0, 0, time.UTC)
	immediate := BuildPlan(PlanInput{})
	scheduled := BuildPlan(PlanInput{RunAt: future})
	if !reflect.DeepEqual(immediate.Stages, scheduled.Stages) {
		t.Fatalf("immediate stages %#v != scheduled stages %#v", immediate.Stages, scheduled.Stages)
	}
	if !scheduled.RunAt.Equal(future) {
		t.Fatalf("scheduled run_at = %v", scheduled.RunAt)
	}
}
