package pipeline

import (
	"reflect"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

func TestBuildPlanSkipsAIWhenMetadataIsDeterministic(t *testing.T) {
	plan := BuildPlan(PlanInput{
		Gender:   novel.GenderResult{Gender: novel.GenderMale, Source: novel.GenderSource121Category},
		HasStyle: true,
	})
	want := []Stage{StageFetchBook, StageResolveMetadata, StageCreateBatch}
	if !reflect.DeepEqual(plan.Stages, want) {
		t.Fatalf("stages = %#v, want %#v", plan.Stages, want)
	}
}

func TestBuildPlanUsesAIOnlyForMissingMetadata(t *testing.T) {
	plan := BuildPlan(PlanInput{
		Gender:   novel.GenderResult{Gender: novel.GenderUnknown, Source: novel.GenderSourceUnresolved},
		HasStyle: false,
	})
	want := []Stage{StageFetchBook, StageResolveMetadata, StageAIClassify, StageCreateBatch}
	if !reflect.DeepEqual(plan.Stages, want) {
		t.Fatalf("stages = %#v, want %#v", plan.Stages, want)
	}
}

func TestImmediateAndScheduledPlansUseSameStages(t *testing.T) {
	future := time.Date(2026, 10, 4, 8, 0, 0, 0, time.UTC)
	immediate := BuildPlan(PlanInput{Gender: novel.GenderResult{Gender: novel.GenderMale}, HasStyle: true})
	scheduled := BuildPlan(PlanInput{Gender: novel.GenderResult{Gender: novel.GenderMale}, HasStyle: true, RunAt: future})
	if !reflect.DeepEqual(immediate.Stages, scheduled.Stages) {
		t.Fatalf("immediate stages %#v != scheduled stages %#v", immediate.Stages, scheduled.Stages)
	}
	if !scheduled.RunAt.Equal(future) {
		t.Fatalf("scheduled run_at = %v", scheduled.RunAt)
	}
}
