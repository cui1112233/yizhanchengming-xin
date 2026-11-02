package novelfetchworkshop

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrQueueEmpty = errors.New("novel fetch queue is empty")

type ProcessJob struct {
	ID                 string            `json:"id"`
	Owner              string            `json:"owner"`
	BatchID            string            `json:"batchId"`
	ScheduleID         string            `json:"scheduleId,omitempty"`
	BookIDs            []string          `json:"bookIds"`
	PlatformID         string            `json:"platformId"`
	MaxTxt             int               `json:"maxTxt"`
	SensitiveAIEnabled bool              `json:"sensitiveAiEnabled"`
	AutoSubmit         bool              `json:"autoSubmit,omitempty"`
	SelectedVersions   []string          `json:"selectedVersions,omitempty"`
	RewriteMethods     map[string]string `json:"rewriteMethods,omitempty"`
	TextModelID        string            `json:"textModelId"`
	Attempts           int               `json:"attempts"`
	LastError          string            `json:"lastError,omitempty"`
	DedupeKey          string            `json:"dedupeKey,omitempty"`
}

type ProcessQueue interface {
	Enqueue(context.Context, ProcessJob) error
	Claim(context.Context, time.Duration) (ProcessJob, error)
	Ack(context.Context, ProcessJob) error
	Fail(context.Context, ProcessJob, bool) error
}

type ProcessCancellationQueue interface {
	CancelBooks(context.Context, string, []string) error
	ClearCancelledBooks(context.Context, string, []string) error
	IsBookCancelled(context.Context, string, string) (bool, error)
}

type ProcessDedupQueue interface {
	EnqueueUnique(context.Context, string, ProcessJob, time.Duration) (bool, error)
}

type ProcessBookDedupQueue interface {
	EnqueueUniqueBooks(context.Context, ProcessJob, time.Duration) (bool, string, error)
}

type ProcessStaleJobRecoveryQueue interface {
	RecoverStaleJob(context.Context, string) (bool, error)
}

type MemoryQueue struct {
	mu        sync.Mutex
	pending   chan ProcessJob
	inflight  map[string]ProcessJob
	cancelled map[string]bool
	dedupe    map[string]string
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{pending: make(chan ProcessJob, 64), inflight: map[string]ProcessJob{}, cancelled: map[string]bool{}, dedupe: map[string]string{}}
}

func (q *MemoryQueue) EnqueueUnique(ctx context.Context, key string, job ProcessJob, _ time.Duration) (bool, error) {
	q.mu.Lock()
	if existingID := q.dedupe[key]; existingID != "" {
		q.mu.Unlock()
		return false, nil
	}
	q.dedupe[key] = job.ID
	q.mu.Unlock()
	job.DedupeKey = key
	if err := q.Enqueue(ctx, job); err != nil {
		q.mu.Lock()
		if q.dedupe[key] == job.ID {
			delete(q.dedupe, key)
		}
		q.mu.Unlock()
		return false, err
	}
	return true, nil
}

func (q *MemoryQueue) EnqueueUniqueBooks(ctx context.Context, job ProcessJob, _ time.Duration) (bool, string, error) {
	if err := validateProcessJob(job); err != nil {
		return false, "", err
	}
	if len(job.BookIDs) == 0 || strings.TrimSpace(job.PlatformID) == "" {
		return false, "", fmt.Errorf("platform and source book ids are required for unique enqueue")
	}
	keys := processBookDedupeKeys(job)
	q.mu.Lock()
	for _, key := range keys {
		if existingID := q.dedupe[key]; existingID != "" {
			q.mu.Unlock()
			return false, existingID, nil
		}
	}
	for _, key := range keys {
		q.dedupe[key] = job.ID
	}
	job.DedupeKey = processBookDedupeMarker(job)
	q.mu.Unlock()
	if err := q.Enqueue(ctx, job); err != nil {
		q.mu.Lock()
		for _, key := range keys {
			if q.dedupe[key] == job.ID {
				delete(q.dedupe, key)
			}
		}
		q.mu.Unlock()
		return false, "", err
	}
	return true, job.ID, nil
}

func processBookDedupeMarker(job ProcessJob) string {
	return "books:" + strings.TrimSpace(job.Owner) + ":" + strings.TrimSpace(job.PlatformID)
}

func processBookDedupeKeys(job ProcessJob) []string {
	seen := map[string]bool{}
	keys := make([]string, 0, len(job.BookIDs))
	for _, bookID := range job.BookIDs {
		bookID = strings.TrimSpace(bookID)
		if bookID == "" || seen[bookID] {
			continue
		}
		seen[bookID] = true
		keys = append(keys, "book:"+strings.TrimSpace(job.Owner)+":"+strings.TrimSpace(job.PlatformID)+":"+bookID)
	}
	return keys
}

func cancellationKey(owner, bookID string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(bookID)
}

func (q *MemoryQueue) CancelBooks(_ context.Context, owner string, bookIDs []string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, bookID := range bookIDs {
		q.cancelled[cancellationKey(owner, bookID)] = true
	}
	return nil
}

func (q *MemoryQueue) ClearCancelledBooks(_ context.Context, owner string, bookIDs []string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, bookID := range bookIDs {
		delete(q.cancelled, cancellationKey(owner, bookID))
	}
	return nil
}

func (q *MemoryQueue) IsBookCancelled(_ context.Context, owner, bookID string) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.cancelled[cancellationKey(owner, bookID)], nil
}

func (q *MemoryQueue) Enqueue(ctx context.Context, job ProcessJob) error {
	if err := validateProcessJob(job); err != nil {
		return err
	}
	select {
	case q.pending <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *MemoryQueue) Claim(ctx context.Context, wait time.Duration) (ProcessJob, error) {
	if wait <= 0 {
		wait = time.Second
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case job := <-q.pending:
		q.mu.Lock()
		q.inflight[job.ID] = job
		q.mu.Unlock()
		return job, nil
	case <-timer.C:
		return ProcessJob{}, ErrQueueEmpty
	case <-ctx.Done():
		return ProcessJob{}, ctx.Err()
	}
}

func (q *MemoryQueue) Ack(_ context.Context, job ProcessJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.inflight, job.ID)
	q.releaseDedupe(job)
	return nil
}

func (q *MemoryQueue) Fail(ctx context.Context, job ProcessJob, retry bool) error {
	q.mu.Lock()
	delete(q.inflight, job.ID)
	q.mu.Unlock()
	if retry {
		if err := q.Enqueue(ctx, job); err != nil {
			q.mu.Lock()
			q.releaseDedupe(job)
			q.mu.Unlock()
			return err
		}
		return nil
	}
	q.mu.Lock()
	q.releaseDedupe(job)
	q.mu.Unlock()
	return nil
}

func (q *MemoryQueue) releaseDedupe(job ProcessJob) {
	if strings.HasPrefix(job.DedupeKey, "books:") {
		for _, key := range processBookDedupeKeys(job) {
			if q.dedupe[key] == job.ID {
				delete(q.dedupe, key)
			}
		}
		return
	}
	if q.dedupe[job.DedupeKey] == job.ID {
		delete(q.dedupe, job.DedupeKey)
	}
}

func validateProcessJob(job ProcessJob) error {
	if strings.TrimSpace(job.ID) == "" || strings.TrimSpace(job.Owner) == "" || strings.TrimSpace(job.BatchID) == "" {
		return fmt.Errorf("process job id, owner and batch id are required")
	}
	return nil
}
