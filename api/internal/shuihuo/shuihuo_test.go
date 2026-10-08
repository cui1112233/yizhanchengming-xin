package shuihuo

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/objectkey"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type testStore struct {
	input            CreateMediaTaskInput
	asset            CreateAssetInput
	scopeErr         error
	createAssetErr   error
	createAssetHook  func()
	scopeCalls       int
	createAssetCalls int
}

func (*testStore) CreateSegment(context.Context, CreateSegmentInput) (Segment, error) {
	return Segment{}, nil
}
func (*testStore) UpdateSegment(context.Context, int64, int64, int64, UpdateSegmentInput) (Segment, error) {
	return Segment{}, nil
}
func (*testStore) ListSegments(context.Context, int64, int64) ([]Segment, error) { return nil, nil }
func (t *testStore) ValidateAssetScope(context.Context, int64, int64, int64) error {
	t.scopeCalls++
	return t.scopeErr
}
func (t *testStore) CreateAsset(_ context.Context, input CreateAssetInput) (Asset, error) {
	t.createAssetCalls++
	t.asset = input
	if t.createAssetHook != nil {
		t.createAssetHook()
	}
	if t.createAssetErr != nil {
		return Asset{}, t.createAssetErr
	}
	return Asset{ID: 7, ObjectKey: input.ObjectKey, Bucket: input.Bucket, Status: input.Status}, nil
}
func (*testStore) ListAssets(context.Context, int64, int64, int64) ([]Asset, error) { return nil, nil }
func (t *testStore) CreateMediaTask(_ context.Context, i CreateMediaTaskInput) (MediaTask, error) {
	t.input = i
	return MediaTask{Status: MediaPendingExecutor}, nil
}

type memoryObjects struct {
	putBucket, putKey string
	getKey            string
	deletedKey        string
	putCalls          int
	deleteCalls       int
	contents          []byte
	deleteErr         error
	deleteContextErr  error
	deleteHasDeadline bool
}

func (m *memoryObjects) PutObjectFromFile(_ context.Context, bucket, key, filename string) error {
	m.putCalls++
	data, err := os.ReadFile(filename)
	if err == nil {
		m.putBucket, m.putKey, m.contents = bucket, key, data
	}
	return err
}
func (m *memoryObjects) DeleteObject(ctx context.Context, _ string, key string) error {
	m.deleteCalls++
	m.deletedKey = key
	m.deleteContextErr = ctx.Err()
	_, m.deleteHasDeadline = ctx.Deadline()
	return m.deleteErr
}
func (m *memoryObjects) GetObject(_ context.Context, _ string, key string) (io.ReadCloser, error) {
	m.getKey = key
	return io.NopCloser(bytes.NewReader(m.contents)), nil
}

type countingReader struct {
	reads int
	data  []byte
}

func (r *countingReader) Read(p []byte) (int, error) {
	r.reads++
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func TestUploadAssetRejectsInvalidScopeBeforeReadingOrUploading(t *testing.T) {
	store := &testStore{scopeErr: ErrNotFound}
	objects := &memoryObjects{}
	service := NewService(store, objects)
	service.SetBucket("private-assets")
	body := &countingReader{data: []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}}

	_, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 99, SegmentID: 8, Type: AssetImage, Filename: "output.png", ContentType: "image/png", Body: body})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound", err)
	}
	if store.scopeCalls != 1 || store.createAssetCalls != 0 || body.reads != 0 || objects.putCalls != 0 || objects.deleteCalls != 0 {
		t.Fatalf("scope=%d create=%d reads=%d put=%d delete=%d", store.scopeCalls, store.createAssetCalls, body.reads, objects.putCalls, objects.deleteCalls)
	}
}

func TestUploadAssetDeletesNewObjectWhenDatabaseInsertFails(t *testing.T) {
	insertErr := errors.New("database insert failed")
	store := &testStore{createAssetErr: assetNotPersistedError(insertErr)}
	objects := &memoryObjects{}
	service := NewService(store, objects)
	service.SetBucket("private-assets")
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}

	_, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 9, SegmentID: 8, Type: AssetImage, Filename: "output.png", ContentType: "image/png", Body: bytes.NewReader(png)})
	if !errors.Is(err, insertErr) {
		t.Fatalf("err=%v, want insert failure", err)
	}
	if store.scopeCalls != 1 || store.createAssetCalls != 1 || objects.putCalls != 1 || objects.deleteCalls != 1 || objects.deletedKey == "" || objects.deletedKey != objects.putKey {
		t.Fatalf("scope=%d create=%d object=%+v", store.scopeCalls, store.createAssetCalls, objects)
	}
}

func TestUploadAssetCleanupUsesDetachedBoundedContextAfterRequestCancellation(t *testing.T) {
	insertErr := errors.New("database insert failed")
	requestContext, cancel := context.WithCancel(context.Background())
	store := &testStore{createAssetErr: assetNotPersistedError(insertErr), createAssetHook: cancel}
	objects := &memoryObjects{}
	service := NewService(store, objects)
	service.SetBucket("private-assets")
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}

	_, err := service.UploadAsset(requestContext, UploadAssetInput{BatchProjectID: 4, BookID: 9, SegmentID: 8, Type: AssetImage, Filename: "output.png", ContentType: "image/png", Body: bytes.NewReader(png)})
	if !errors.Is(err, insertErr) {
		t.Fatalf("err=%v, want insert failure", err)
	}
	if objects.deleteCalls != 1 || objects.deletedKey != objects.putKey {
		t.Fatalf("deleteCalls=%d deleted=%q uploaded=%q", objects.deleteCalls, objects.deletedKey, objects.putKey)
	}
	if objects.deleteContextErr != nil {
		t.Fatalf("cleanup context err=%v, want a context detached from request cancellation", objects.deleteContextErr)
	}
	if !objects.deleteHasDeadline {
		t.Fatal("cleanup context has no deadline; want bounded compensation")
	}
}

func TestUploadAssetCleanupFailurePreservesOriginalPersistenceError(t *testing.T) {
	insertErr := errors.New("database insert failed")
	cleanupErr := errors.New("tos delete unavailable")
	store := &testStore{createAssetErr: assetNotPersistedError(insertErr)}
	objects := &memoryObjects{deleteErr: cleanupErr}
	service := NewService(store, objects)
	service.SetBucket("private-assets")
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}

	_, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 9, SegmentID: 8, Type: AssetImage, Filename: "output.png", ContentType: "image/png", Body: bytes.NewReader(png)})
	if !errors.Is(err, insertErr) {
		t.Fatalf("err=%v, want original persistence error", err)
	}
	if !strings.Contains(err.Error(), "cleanup") || !strings.Contains(err.Error(), cleanupErr.Error()) {
		t.Fatalf("err=%v, want truthful cleanup failure boundary", err)
	}
	if objects.deleteCalls != 1 || objects.deletedKey != objects.putKey {
		t.Fatalf("deleteCalls=%d deleted=%q uploaded=%q", objects.deleteCalls, objects.deletedKey, objects.putKey)
	}
}

func TestUploadAssetDoesNotDeleteObjectWhenDatabaseCommitOutcomeIsUnknown(t *testing.T) {
	commitErr := errors.New("commit outcome unknown")
	store := &testStore{createAssetErr: assetCommitOutcomeUnknownError(commitErr)}
	objects := &memoryObjects{}
	service := NewService(store, objects)
	service.SetBucket("private-assets")
	png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}

	_, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 9, SegmentID: 8, Type: AssetImage, Filename: "output.png", ContentType: "image/png", Body: bytes.NewReader(png)})
	if !errors.Is(err, commitErr) {
		t.Fatalf("err=%v, want commit failure", err)
	}
	if objects.putCalls != 1 || objects.deleteCalls != 0 {
		t.Fatalf("put=%d delete=%d; uncertain commit must retain the referenced object", objects.putCalls, objects.deleteCalls)
	}
}

func TestUploadAssetWithoutConfiguredObjectStoreStillFailsTruthfully(t *testing.T) {
	store := &testStore{}
	service := NewService(store)
	service.SetBucket("private-assets")

	_, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 9, Type: AssetImage, Body: bytes.NewReader([]byte("png"))})
	if !errors.Is(err, ErrStorageUnavailable) {
		t.Fatalf("err=%v, want ErrStorageUnavailable", err)
	}
	if store.scopeCalls != 0 || store.createAssetCalls != 0 {
		t.Fatalf("scope=%d create=%d", store.scopeCalls, store.createAssetCalls)
	}
}

func TestUploadAssetKeyPrefixAndSavedReadback(t *testing.T) {
	for _, raw := range []string{"", "staging/"} {
		t.Run(raw, func(t *testing.T) {
			store, objects := &testStore{}, &memoryObjects{}
			service := NewService(store, objects)
			service.SetBucket("private-assets")
			prefix, err := objectkey.ParsePrefix(raw)
			if err != nil {
				t.Fatal(err)
			}
			service.SetKeyPrefix(prefix)
			png := []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}
			asset, err := service.UploadAsset(context.Background(), UploadAssetInput{BatchProjectID: 4, BookID: 9, Type: AssetImage, Filename: "output.png", ContentType: "image/png", Body: bytes.NewReader(png)})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(asset.ObjectKey, raw+"shuihuo/project-4/book-9/") || strings.Count(asset.ObjectKey, "staging/") != strings.Count(raw, "staging/") || asset.ObjectKey != objects.putKey || asset.ObjectKey != store.asset.ObjectKey {
				t.Fatalf("asset=%+v upload=%q saved=%q", asset, objects.putKey, store.asset.ObjectKey)
			}
			// A changed runtime namespace must not change an existing saved reference.
			next, err := objectkey.ParsePrefix("next-run/")
			if err != nil {
				t.Fatal(err)
			}
			service.SetKeyPrefix(next)
			_, body, err := service.OpenAsset(context.Background(), 4, 9, asset.ID)
			if err != nil {
				t.Fatal(err)
			}
			defer body.Close()
			data, err := io.ReadAll(body)
			if err != nil || !bytes.Equal(data, png) || objects.getKey != asset.ObjectKey {
				t.Fatalf("read=%q saved=%q err=%v", objects.getKey, asset.ObjectKey, err)
			}
		})
	}
}

func TestUploadAssetPrefixDoesNotRewriteImportedReference(t *testing.T) {
	store := &testStore{}
	service := NewService(store)
	prefix, err := objectkey.ParsePrefix("staging/")
	if err != nil {
		t.Fatal(err)
	}
	service.SetKeyPrefix(prefix)
	asset, err := service.CreateAsset(context.Background(), CreateAssetInput{BatchProjectID: 4, BookID: 9, Type: AssetAudio, Bucket: "legacy", ObjectKey: "existing/audio.mp3"})
	if err != nil || asset.ObjectKey != "existing/audio.mp3" {
		t.Fatalf("asset=%+v err=%v", asset, err)
	}
}

type generatedRoundTripper func(*http.Request) (*http.Response, error)

func (f generatedRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestUploadAssetGeneratedMediaKeyPrefix(t *testing.T) {
	for _, raw := range []string{"", "staging/"} {
		for _, tc := range []struct {
			kind        MediaKind
			contentType string
			body        []byte
		}{
			{MediaImage, "image/png", []byte{'\x89', 'P', 'N', 'G', '\r', '\n', '\x1a', '\n', 0, 0, 0, 0}},
			{MediaAudio, "audio/wav", []byte("RIFF\x00\x00\x00\x00WAVEfmt ")},
		} {
			t.Run(raw+string(tc.kind), func(t *testing.T) {
				store, objects := &testStore{}, &memoryObjects{}
				service := NewService(store, objects)
				service.SetBucket("private")
				prefix, err := objectkey.ParsePrefix(raw)
				if err != nil {
					t.Fatal(err)
				}
				service.SetKeyPrefix(prefix)
				requests := 0
				worker := &Worker{service: service, client: &http.Client{Transport: generatedRoundTripper(func(*http.Request) (*http.Response, error) {
					requests++
					return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(tc.body))}, nil
				})}}
				asset, err := worker.persist(context.Background(), MediaTask{BatchProjectID: 4, BookID: 9, SegmentID: 2, Kind: tc.kind}, GeneratedMedia{URL: "https://provider.example/output", ContentType: tc.contentType})
				if err != nil {
					t.Fatal(err)
				}
				if requests != 1 || !strings.HasPrefix(asset.ObjectKey, raw+"shuihuo/project-4/book-9/") || asset.ObjectKey != objects.putKey || store.asset.ObjectKey != asset.ObjectKey || store.asset.SegmentID != 2 || !bytes.Equal(objects.contents, tc.body) {
					t.Fatalf("asset=%+v upload=%q saved=%+v requests=%d", asset, objects.putKey, store.asset, requests)
				}
			})
		}
	}
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
func (t *testStore) GetAsset(context.Context, int64, int64, int64) (Asset, error) {
	if t.asset.ObjectKey != "" {
		return Asset{ID: 7, ObjectKey: t.asset.ObjectKey, Bucket: t.asset.Bucket, Status: t.asset.Status}, nil
	}
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

type configuredProvider struct{}

func (configuredProvider) Available(kind MediaKind) bool { return kind == MediaImage }
func (configuredProvider) Generate(context.Context, MediaTask, string) (GeneratedMedia, error) {
	return GeneratedMedia{}, nil
}

type queueFake struct {
	message taskruntime.Message
	calls   int
}

func (q *queueFake) Enqueue(_ context.Context, m taskruntime.Message) error {
	q.calls++
	q.message = m
	return nil
}
func (*queueFake) Claim(context.Context, string, time.Duration) (taskruntime.Delivery, error) {
	return taskruntime.Delivery{}, taskruntime.ErrQueueEmpty
}
func (*queueFake) Ack(context.Context, taskruntime.Delivery) error                 { return nil }
func (*queueFake) Nack(context.Context, taskruntime.Delivery, time.Duration) error { return nil }

type configuredStore struct {
	testStore
	queued      int64
	createCalls int
	taskErr     error
}

func (s *configuredStore) CreateMediaTask(context.Context, CreateMediaTaskInput) (MediaTask, error) {
	s.createCalls++
	if s.taskErr != nil {
		return MediaTask{}, s.taskErr
	}
	return MediaTask{ID: 42, Kind: MediaImage, Status: MediaPendingExecutor, BatchProjectID: 1, BookID: 2}, nil
}
func (s *configuredStore) QueueMediaTask(_ context.Context, id int64) (MediaTask, error) {
	s.queued = id
	return MediaTask{ID: id, Kind: MediaImage, Status: MediaQueued}, nil
}
func (*configuredStore) MediaTaskForExecution(context.Context, int64) (MediaTask, Segment, error) {
	return MediaTask{}, Segment{}, nil
}
func (*configuredStore) StartMediaTask(context.Context, int64) (bool, error) { return false, nil }
func (*configuredStore) CompleteMediaTask(context.Context, int64, int64) (bool, error) {
	return false, nil
}
func (*configuredStore) FailMediaTask(context.Context, int64, string, string, bool) (bool, error) {
	return false, nil
}

func TestConfiguredImageProviderQueuesSharedRuntimeTask(t *testing.T) {
	store, queue := &configuredStore{}, &queueFake{}
	service := NewService(store)
	service.SetMediaExecutor(configuredProvider{}, queue)
	task, err := service.CreateMediaTask(context.Background(), CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, Kind: MediaImage})
	if err != nil || task.Status != MediaQueued || store.queued != 42 || queue.message.TaskKey != "shuihuo-media:42" {
		t.Fatalf("task=%+v queued=%d message=%+v err=%v", task, store.queued, queue.message, err)
	}
}

func TestRejectedMediaTaskScopeNeverQueues(t *testing.T) {
	for _, kind := range []MediaKind{MediaImage, MediaAudio, MediaVideo} {
		t.Run(string(kind), func(t *testing.T) {
			store, queue := &configuredStore{taskErr: ErrNotFound}, &queueFake{}
			service := NewService(store)
			service.SetMediaExecutor(configuredProvider{}, queue)
			input := CreateMediaTaskInput{BatchProjectID: 1, BookID: 2, SegmentID: 3, SourceAssetID: 4, Kind: kind}
			if kind == MediaVideo {
				input.ProductionTaskID = 5
			}

			_, err := service.CreateMediaTask(context.Background(), input)
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("err=%v, want ErrNotFound", err)
			}
			if store.createCalls != 1 || store.queued != 0 || queue.calls != 0 {
				t.Fatalf("create=%d queued=%d enqueue=%d", store.createCalls, store.queued, queue.calls)
			}
		})
	}
}
