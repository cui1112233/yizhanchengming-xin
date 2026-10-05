package video

import (
	"context"
	"errors"
	"testing"
)

type mergeSourceStore struct {
	*memoryMergeStore
	inputs []MergeInputAsset
	err    error
}

func (s *mergeSourceStore) ResolveSucceededMergeInputs(context.Context, int64, int64, []int64) ([]MergeInputAsset, error) {
	if s.err != nil {
		return nil, s.err
	}
	return copyMergeInputs(s.inputs), nil
}

func TestMergeStartFromProductionTasksUsesOnlyDurableSucceededAssets(t *testing.T) {
	store := &mergeSourceStore{
		memoryMergeStore: newMemoryMergeStore(),
		inputs: []MergeInputAsset{
			{ProductionTaskID: 11, URL: "https://tos.example/11.mp4", Order: 1},
			{ProductionTaskID: 12, URL: "https://tos.example/12.mp4", Order: 2},
		},
	}
	service := NewMergeService(store, &recordingMergeExecutor{})
	result, err := service.StartFromProductionTasks(context.Background(), MergeProductionStartRequest{
		BatchProjectID: 7,
		BookID: 9,
		ProductionTaskIDs: []int64{11, 12},
		AspectRatio: "9:16",
		Speed: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Attempt.Status != MergeQueued || len(result.Attempt.Inputs) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if result.Attempt.Inputs[0].URL != "https://tos.example/11.mp4" {
		t.Fatalf("inputs = %+v", result.Attempt.Inputs)
	}
}

func TestMergeStartFromProductionTasksRejectsVideoThatIsNotReady(t *testing.T) {
	store := &mergeSourceStore{memoryMergeStore: newMemoryMergeStore(), err: ErrMergeInputNotReady}
	service := NewMergeService(store, &recordingMergeExecutor{})
	_, err := service.StartFromProductionTasks(context.Background(), MergeProductionStartRequest{
		BatchProjectID: 7,
		BookID: 9,
		ProductionTaskIDs: []int64{11},
		AspectRatio: "9:16",
		Speed: 1,
	})
	if !errors.Is(err, ErrMergeInputNotReady) {
		t.Fatalf("err = %v", err)
	}
}
