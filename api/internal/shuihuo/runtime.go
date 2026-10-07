package shuihuo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type mediaRuntimeStore interface {
	QueueMediaTask(context.Context, int64) (MediaTask, error)
	MediaTaskForExecution(context.Context, int64) (MediaTask, Segment, error)
	StartMediaTask(context.Context, int64) (bool, error)
	CompleteMediaTask(context.Context, int64, int64) (bool, error)
	FailMediaTask(context.Context, int64, string, string, bool) (bool, error)
}

func mediaTaskKey(id int64) string { return fmt.Sprintf("shuihuo-media:%d", id) }
func decodeMediaTaskKey(key string) (int64, error) {
	var id int64
	if _, err := fmt.Sscanf(key, "shuihuo-media:%d", &id); err != nil || id < 1 {
		return 0, fmt.Errorf("invalid media task key")
	}
	return id, nil
}

func (s *Service) configureMediaExecutor(provider MediaProvider, queue taskruntime.Queue) {
	s.provider, s.queue = provider, queue
}
func (s *Service) enqueueMediaTask(ctx context.Context, task MediaTask) (MediaTask, error) {
	if s.provider == nil || !s.provider.Available(task.Kind) {
		return s.store.MarkExecutorUnavailable(ctx, task.ID, "executor_unavailable", "configure a Go image or TTS provider before running this media task")
	}
	runtime, ok := s.store.(mediaRuntimeStore)
	if !ok {
		return MediaTask{}, fmt.Errorf("shuihuo: media runtime store unavailable")
	}
	if s.queue == nil {
		_, _ = runtime.FailMediaTask(ctx, task.ID, "queue_unavailable", "Redis media queue is not configured", true)
		task.Status = MediaRetryableFailed
		task.ErrorCode = "queue_unavailable"
		task.ErrorMessage = "Redis media queue is not configured"
		return task, nil
	}
	queued, err := runtime.QueueMediaTask(ctx, task.ID)
	if err != nil {
		return MediaTask{}, err
	}
	if err := s.queue.Enqueue(ctx, taskruntime.Message{TaskKey: mediaTaskKey(task.ID)}); err != nil {
		_, _ = runtime.FailMediaTask(ctx, task.ID, "queue_unavailable", "Redis media queue is unavailable", true)
		return MediaTask{}, err
	}
	return queued, nil
}

// Worker consumes the shared taskruntime queue. MySQL remains the state source;
// a duplicate Redis delivery is harmless because StartMediaTask is conditional.
type Worker struct {
	service *Service
	queue   taskruntime.Queue
	owner   string
	client  *http.Client
}

func NewWorker(service *Service, queue taskruntime.Queue, owner string) *Worker {
	return &Worker{service: service, queue: queue, owner: owner, client: &http.Client{Timeout: 90 * time.Second}}
}
func (w *Worker) Run(ctx context.Context) error {
	for {
		if err := w.RunOnce(ctx, time.Second); err != nil && err != taskruntime.ErrQueueEmpty {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}
func (w *Worker) RunOnce(ctx context.Context, poll time.Duration) error {
	delivery, err := w.queue.Claim(ctx, w.owner, poll)
	if err != nil {
		return err
	}
	id, err := decodeMediaTaskKey(delivery.Message.TaskKey)
	if err != nil {
		_ = w.queue.Ack(ctx, delivery)
		return err
	}
	err = w.execute(ctx, id)
	if err != nil {
		_ = w.queue.Nack(ctx, delivery, 5*time.Second)
		return err
	}
	return w.queue.Ack(ctx, delivery)
}
func (w *Worker) execute(ctx context.Context, id int64) error {
	runtime, ok := w.service.store.(mediaRuntimeStore)
	if !ok {
		return fmt.Errorf("shuihuo: media runtime store unavailable")
	}
	started, err := runtime.StartMediaTask(ctx, id)
	if err != nil || !started {
		return err
	}
	task, segment, err := runtime.MediaTaskForExecution(ctx, id)
	if err != nil {
		return w.fail(ctx, runtime, id, "task_load_failed", err, true)
	}
	output, err := w.service.provider.Generate(ctx, task, segment.Text)
	if err != nil {
		return w.fail(ctx, runtime, id, "provider_failed", err, true)
	}
	asset, err := w.persist(ctx, task, output)
	if err != nil {
		return w.fail(ctx, runtime, id, "artifact_persist_failed", err, true)
	}
	_, err = runtime.CompleteMediaTask(ctx, id, asset.ID)
	return err
}
func (w *Worker) fail(ctx context.Context, store mediaRuntimeStore, id int64, code string, cause error, retryable bool) error {
	_, updateErr := store.FailMediaTask(ctx, id, code, safeMediaError(cause), retryable)
	if updateErr != nil {
		return updateErr
	}
	return nil
}
func safeMediaError(err error) string {
	if err == nil {
		return ""
	}
	value := err.Error()
	for _, marker := range []string{"bearer ", "api key", "secret", "token"} {
		if strings.Contains(strings.ToLower(value), marker) {
			return "provider request failed"
		}
	}
	if len(value) > 1024 {
		value = value[:1024]
	}
	return value
}
func (w *Worker) persist(ctx context.Context, task MediaTask, output GeneratedMedia) (Asset, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, output.URL, nil)
	if err != nil {
		return Asset{}, err
	}
	response, err := w.client.Do(request)
	if err != nil {
		return Asset{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Asset{}, fmt.Errorf("provider artifact returned status %d", response.StatusCode)
	}
	contentType := output.ContentType
	if contentType == "" {
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	}
	typeValue := AssetImage
	if task.Kind == MediaAudio {
		typeValue = AssetAudio
	}
	filename := "output" + extensionFor(contentType, task.Kind)
	return w.service.UploadAsset(ctx, UploadAssetInput{BatchProjectID: task.BatchProjectID, BookID: task.BookID, SegmentID: task.SegmentID, Type: typeValue, Filename: filename, ContentType: contentType, Body: io.LimitReader(response.Body, maxGeneratedBytes(task.Kind))})
}
func extensionFor(contentType string, kind MediaKind) string {
	parts := strings.Split(contentType, "/")
	ext := ""
	if len(parts) == 2 {
		ext = path.Ext(parts[1])
	}
	if ext != "" {
		return ext
	}
	if kind == MediaAudio {
		return ".mp3"
	}
	return ".png"
}
func maxGeneratedBytes(kind MediaKind) int64 {
	if kind == MediaAudio {
		return 200 << 20
	}
	return 20 << 20
}
