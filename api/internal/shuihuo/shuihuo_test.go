package shuihuo

import (
	"context"
	"testing"
)

type testStore struct{ input CreateMediaTaskInput }

func (*testStore) CreateSegment(context.Context, CreateSegmentInput) (Segment, error) {
	return Segment{}, nil
}
func (*testStore) UpdateSegment(context.Context, int64, int64, int64, UpdateSegmentInput) (Segment, error) {
	return Segment{}, nil
}
func (*testStore) ListSegments(context.Context, int64, int64) ([]Segment, error)    { return nil, nil }
func (*testStore) CreateAsset(context.Context, CreateAssetInput) (Asset, error)     { return Asset{}, nil }
func (*testStore) ListAssets(context.Context, int64, int64, int64) ([]Asset, error) { return nil, nil }
func (t *testStore) CreateMediaTask(_ context.Context, i CreateMediaTaskInput) (MediaTask, error) {
	t.input = i
	return MediaTask{Status: MediaPendingExecutor}, nil
}
func (*testStore) ListMediaTasks(context.Context, int64, int64) ([]MediaTask, error) { return nil, nil }
func (*testStore) ReorderSegments(context.Context, int64, int64, []int64) ([]Segment, error) {
	return nil, nil
}
func (*testStore) ListCandidates(context.Context, int64, int64, int64) ([]Candidate, error) {
	return nil, nil
}
func (*testStore) SelectCandidate(context.Context, int64, int64, int64, int64) (Candidate, error) {
	return Candidate{}, nil
}
func (*testStore) RetryMediaTask(context.Context, int64, int64, int64) (MediaTask, error) {
	return MediaTask{Status: MediaPendingExecutor}, nil
}
func TestImageTaskIsRecordedAsPendingExecutorWithoutProviderCall(t *testing.T) {
	m := &testStore{}
	got, e := NewService(m).CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, Kind: MediaImage})
	if e != nil || got.Status != MediaPendingExecutor {
		t.Fatalf("task=%+v err=%v", got, e)
	}
	if _, e = NewService(m).CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, Kind: MediaVideo}); e == nil {
		t.Fatal("video task lacking existing production task accepted")
	}
}
