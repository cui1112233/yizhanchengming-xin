package shuihuo

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
)

type testStore struct {
	input CreateMediaTaskInput
	asset CreateAssetInput
}

func (*testStore) CreateSegment(context.Context, CreateSegmentInput) (Segment, error) {
	return Segment{}, nil
}
func (*testStore) UpdateSegment(context.Context, int64, int64, int64, UpdateSegmentInput) (Segment, error) {
	return Segment{}, nil
}
func (*testStore) ListSegments(context.Context, int64, int64) ([]Segment, error) { return nil, nil }
func (t *testStore) CreateAsset(_ context.Context, input CreateAssetInput) (Asset, error) {
	t.asset = input
	return Asset{ID: 7, ObjectKey: input.ObjectKey, Bucket: input.Bucket, Status: input.Status}, nil
}
func (*testStore) ListAssets(context.Context, int64, int64, int64) ([]Asset, error) { return nil, nil }
func (t *testStore) CreateMediaTask(_ context.Context, i CreateMediaTaskInput) (MediaTask, error) {
	t.input = i
	return MediaTask{Status: MediaPendingExecutor}, nil
}

type memoryObjects struct {
	putBucket, putKey string
	contents          []byte
}

func (m *memoryObjects) PutObjectFromFile(_ context.Context, bucket, key, filename string) error {
	data, err := os.ReadFile(filename)
	if err == nil {
		m.putBucket, m.putKey, m.contents = bucket, key, data
	}
	return err
}
func (m *memoryObjects) GetObject(context.Context, string, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.contents)), nil
}

func TestUploadAssetGeneratesScopedTOSKeyAndDoesNotTrustBrowserKey(t *testing.T) {
	store, objects := &testStore{}, &memoryObjects{}
	service := NewService(store, objects)
	service.SetBucket("private-assets")
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}
	asset, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 9, Type: AssetImage, Filename: "../../mine.png", ContentType: "image/png", Body: bytes.NewReader(png)})
	if err != nil {
		t.Fatal(err)
	}
	if asset.Bucket != "private-assets" || objects.putBucket != "private-assets" || objects.putKey == "" || store.asset.ObjectKey != objects.putKey {
		t.Fatalf("asset=%+v object=%+v store=%+v", asset, objects, store.asset)
	}
	if bytes.Contains(store.asset.Metadata, []byte("../")) {
		t.Fatalf("unsafe filename metadata: %s", store.asset.Metadata)
	}
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
func (*testStore) MarkExecutorUnavailable(context.Context, int64, string, string) (MediaTask, error) {
	return MediaTask{Status: MediaFailed, ErrorCode: "executor_unavailable"}, nil
}
func (*testStore) GetAsset(context.Context, int64, int64, int64) (Asset, error) {
	return Asset{}, ErrNotFound
}
func TestVideoTaskIsRecordedAsPendingExecutorWithoutProviderCall(t *testing.T) {
	m := &testStore{}
	got, e := NewService(m).CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, Kind: MediaVideo, ProductionTaskID: 11})
	if e != nil || got.Status != MediaPendingExecutor {
		t.Fatalf("task=%+v err=%v", got, e)
	}
	if _, e = NewService(m).CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, Kind: MediaVideo}); e == nil {
		t.Fatal("video task lacking existing production task accepted")
	}
}

func TestImageAndAudioTasksTruthfullyReportMissingExecutor(t *testing.T) {
	m := &testStore{}
	for _, kind := range []MediaKind{MediaImage, MediaAudio} {
		got, err := NewService(m).CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, Kind: kind})
		if err != nil || got.Status != MediaFailed || got.ErrorCode != "executor_unavailable" {
			t.Fatalf("kind=%s task=%+v err=%v", kind, got, err)
		}
	}
}
