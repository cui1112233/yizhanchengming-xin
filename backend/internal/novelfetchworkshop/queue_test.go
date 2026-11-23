package novelfetchworkshop

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/mergeworker"
)

func TestMemoryQueueClaimAckAndRetry(t *testing.T) {
	queue := NewMemoryQueue()
	job := ProcessJob{ID: "job-1", Owner: "tester", BatchID: "batch-1", BookIDs: []string{"12345678901"}}
	if err := queue.Enqueue(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	claimed, err := queue.Claim(context.Background(), time.Second)
	if err != nil || claimed.ID != job.ID {
		t.Fatalf("claimed=%+v err=%v", claimed, err)
	}
	claimed.Attempts++
	if err := queue.Fail(context.Background(), claimed, true); err != nil {
		t.Fatal(err)
	}
	retried, err := queue.Claim(context.Background(), time.Second)
	if err != nil || retried.Attempts != 1 {
		t.Fatalf("retried=%+v err=%v", retried, err)
	}
	if err := queue.Ack(context.Background(), retried); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryQueueEnqueueUniqueBooksProtectsOverlappingBatches(t *testing.T) {
	ctx := context.Background()
	queue := NewMemoryQueue()
	first := ProcessJob{ID: "job-first", Owner: "tester", BatchID: "batch-first", PlatformID: "2", BookIDs: []string{"book-a", "book-b"}}
	queued, existingID, err := queue.EnqueueUniqueBooks(ctx, first, time.Hour)
	if err != nil || !queued || existingID != first.ID {
		t.Fatalf("first enqueue queued=%v existingID=%q err=%v", queued, existingID, err)
	}

	claimed, err := queue.Claim(ctx, time.Second)
	if err != nil || claimed.ID != first.ID {
		t.Fatalf("claim first job=%#v err=%v", claimed, err)
	}
	if err := queue.Fail(ctx, claimed, true); err != nil {
		t.Fatal(err)
	}
	duplicate := ProcessJob{ID: "job-overlap", Owner: "tester", BatchID: "batch-overlap", PlatformID: "2", BookIDs: []string{"book-b", "book-c"}}
	queued, existingID, err = queue.EnqueueUniqueBooks(ctx, duplicate, time.Hour)
	if err != nil || queued || existingID != first.ID {
		t.Fatalf("retry released its active book lock: queued=%v existingID=%q err=%v", queued, existingID, err)
	}
	claimed, err = queue.Claim(ctx, time.Second)
	if err != nil || claimed.ID != first.ID {
		t.Fatalf("claim retried job=%#v err=%v", claimed, err)
	}
	if err := queue.Ack(ctx, claimed); err != nil {
		t.Fatal(err)
	}
	third := ProcessJob{ID: "job-third", Owner: "tester", BatchID: "batch-third", PlatformID: "2", BookIDs: []string{"book-c"}}
	queued, _, err = queue.EnqueueUniqueBooks(ctx, third, time.Hour)
	if err != nil || !queued {
		t.Fatalf("failed overlapping enqueue left a partial book lock: queued=%v err=%v", queued, err)
	}
	retry := ProcessJob{ID: "job-rerun", Owner: "tester", BatchID: "batch-rerun", PlatformID: "2", BookIDs: []string{"book-b"}}
	queued, _, err = queue.EnqueueUniqueBooks(ctx, retry, time.Hour)
	if err != nil || !queued {
		t.Fatalf("terminal ack did not release the source book lock: queued=%v err=%v", queued, err)
	}
}

func TestMemoryQueuePreservesScheduledAutoSubmit(t *testing.T) {
	queue := NewMemoryQueue()
	want := ProcessJob{ID: "job-auto-submit", Owner: "tester", BatchID: "batch-auto-submit", BookIDs: []string{"12345678901"}, AutoSubmit: true}
	if err := queue.Enqueue(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	got, err := queue.Claim(context.Background(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !got.AutoSubmit {
		t.Fatalf("automatic upload intent was lost: %#v", got)
	}
}

func TestMemoryQueueCancelledBookRejectsWholeEnqueue(t *testing.T) {
	ctx := context.Background()
	q := NewMemoryQueue()
	job := ProcessJob{ID: "cancelled-batch", Owner: "owner", BatchID: "batch", PlatformID: "2", BookIDs: []string{"free", "stopped"}}
	if err := q.CancelBooks(ctx, job.Owner, []string{"stopped"}); err != nil {
		t.Fatal(err)
	}
	queued, existing, err := q.EnqueueUniqueBooks(ctx, job, time.Hour)
	if queued || existing != "" || !errors.Is(err, ErrBookCancelled) {
		t.Errorf("cancelled batch queued=%v existing=%q err=%v", queued, existing, err)
	}
	if len(q.pending) != 0 || len(q.dedupe) != 0 {
		t.Fatalf("rejected batch left pending=%d dedupe=%d", len(q.pending), len(q.dedupe))
	}
	if err := q.ClearCancelledBooks(ctx, job.Owner, []string{"stopped"}); err != nil {
		t.Fatal(err)
	}
	queued, _, err = q.EnqueueUniqueBooks(ctx, job, time.Hour)
	if !queued || err != nil || len(q.pending) != 1 {
		t.Fatalf("explicit retry queued=%v pending=%d err=%v", queued, len(q.pending), err)
	}
}

func TestRedisQueueCancelledBookRejectsWholeEnqueue(t *testing.T) {
	url := os.Getenv("V88_NOVEL_FETCH_TEST_REDIS_URL")
	if url == "" {
		t.Skip("explicit isolated Redis endpoint required")
	}
	prefix := fmt.Sprintf("cancel-enqueue-test:%d", time.Now().UnixNano())
	transport, err := mergeworker.NewRedisStore(url, prefix)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	q := NewRedisQueueWithDo(prefix, RedisDoFunc(transport.Do))
	job := ProcessJob{ID: "mixed", Owner: "owner", BatchID: "batch", PlatformID: "2", BookIDs: []string{"free", "stopped", "free"}}
	if err = q.CancelBooks(ctx, job.Owner, []string{"stopped"}); err != nil {
		t.Fatal(err)
	}
	defer q.ClearCancelledBooks(ctx, job.Owner, []string{"stopped"})
	queued, existing, err := q.EnqueueUniqueBooks(ctx, job, time.Hour)
	if queued || existing != "" || !errors.Is(err, ErrBookCancelled) {
		t.Fatalf("queued=%v existing=%q err=%v", queued, existing, err)
	}
	count, err := transport.Do(ctx, "EXISTS", q.jobKey(job.ID), q.pendingKey(), q.bookDedupeKey(job.Owner, "2", "free"), q.bookDedupeKey(job.Owner, "2", "stopped"))
	if err != nil || count != int64(0) {
		t.Fatalf("rejected batch left keys: %v %v", count, err)
	}
	marked, err := q.IsBookCancelled(ctx, job.Owner, "stopped")
	if err != nil || !marked {
		t.Fatal(marked, err)
	}
	if err := q.ClearCancelledBooks(ctx, job.Owner, []string{"stopped"}); err != nil {
		t.Fatal(err)
	}
	queued, _, err = q.EnqueueUniqueBooks(ctx, job, time.Hour)
	if err != nil || !queued {
		t.Fatal(queued, err)
	}
	got, err := q.Claim(ctx, time.Second)
	if err != nil || got.ID != job.ID {
		t.Fatal(got, err)
	}
	if err = q.Ack(ctx, got); err != nil {
		t.Fatal(err)
	}
}

func TestRedisQueueUsesPendingProcessingAndLeaseKeys(t *testing.T) {
	fake := &fakeRedisQueueDo{}
	queue := NewRedisQueueWithDo("qiantie:test:novel", fake.Do)
	job := ProcessJob{ID: "job-1", Owner: "tester", BatchID: "batch-1"}
	if err := queue.Enqueue(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	fake.responses = []any{"job-1", `{"id":"job-1","owner":"tester","batchId":"batch-1"}`, "OK"}
	claimed, err := queue.Claim(context.Background(), 30*time.Second)
	if err != nil || claimed.ID != job.ID {
		t.Fatalf("claimed=%+v err=%v commands=%v", claimed, err, fake.commands)
	}
	fake.responses = []any{int64(1)}
	if err := queue.Ack(context.Background(), claimed); err != nil {
		t.Fatal(err)
	}
	if !fake.hasCommand("BRPOPLPUSH") || !fake.hasCommand("SET") || !fake.hasCommand("EVAL") || fake.hasCommand("LREM") || fake.hasCommand("DEL") {
		t.Fatalf("redis queue contract missing command: %v", fake.commands)
	}
}

func TestRedisQueueEnqueueUniqueUsesDistributedDedupeKey(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{"OK", "OK", "OK"}}
	queue := NewRedisQueueWithDo("qiantie:test:novel", fake.Do)
	job := ProcessJob{ID: "job-unique", Owner: "tester", BatchID: "batch-unique"}
	queued, err := queue.EnqueueUnique(context.Background(), "tester:book-1", job, time.Hour)
	if err != nil || !queued {
		t.Fatalf("queued=%v err=%v commands=%v", queued, err, fake.commands)
	}
	if len(fake.commands) == 0 || len(fake.commands[0]) < 6 || fake.commands[0][0] != "SET" || fake.commands[0][3] != "NX" || fake.commands[0][4] != "EX" {
		t.Fatalf("missing Redis NX/EX dedupe lock: %v", fake.commands)
	}
	fake.responses = []any{nil}
	queued, err = queue.EnqueueUnique(context.Background(), "tester:book-1", ProcessJob{ID: "job-duplicate", Owner: "tester", BatchID: "batch-unique"}, time.Hour)
	if err != nil || queued {
		t.Fatalf("duplicate queued=%v err=%v", queued, err)
	}
}

func TestRedisQueueEnqueueUniqueBooksUsesAtomicPerBookKeys(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{[]any{int64(1), "job-books"}, []any{int64(0), "job-books"}}}
	queue := NewRedisQueueWithDo("qiantie:test:novel", fake.Do)
	job := ProcessJob{ID: "job-books", Owner: "tester", BatchID: "batch-books", PlatformID: "2", BookIDs: []string{"book-a", "book-b"}}
	queued, existingID, err := queue.EnqueueUniqueBooks(context.Background(), job, time.Hour)
	if err != nil || !queued || existingID != job.ID {
		t.Fatalf("first book enqueue queued=%v existingID=%q err=%v", queued, existingID, err)
	}
	if len(fake.commands) != 1 || fake.commands[0][0] != "EVAL" || fake.commands[0][2] != "6" ||
		!slices.Equal(fake.commands[0][7:9], []string{queue.cancelKey(job.Owner, "book-a"), queue.cancelKey(job.Owner, "book-b")}) {
		t.Fatalf("expected one atomic eval with both book locks and cancellation keys: %v", fake.commands)
	}
	job.BookIDs = []string{"book-b", "book-c"}
	queued, existingID, err = queue.EnqueueUniqueBooks(context.Background(), job, time.Hour)
	if err != nil || queued || existingID != "job-books" {
		t.Fatalf("overlap result queued=%v existingID=%q err=%v", queued, existingID, err)
	}
}

func TestRedisQueueAckReleasesOnlyBookLocksOwnedByCompletedJob(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{[]any{int64(1), "job-books"}, "1", "OK", "1"}}
	queue := NewRedisQueueWithDo("qiantie:test:novel", fake.Do)
	job := ProcessJob{ID: "job-books", Owner: "tester", BatchID: "batch-books", PlatformID: "2", BookIDs: []string{"book-a", "book-b"}}
	if queued, _, err := queue.EnqueueUniqueBooks(context.Background(), job, time.Hour); err != nil || !queued {
		t.Fatalf("enqueue books queued=%v err=%v", queued, err)
	}
	job.LeaseToken = "token"
	job.DedupeKey = "books:tester:2"
	if err := queue.Ack(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if len(fake.commands) != 2 || fake.commands[1][0] != "EVAL" || !strings.Contains(fake.commands[1][1], "== ARGV[1]") {
		t.Fatalf("ack did not compare-and-delete the completed job's book locks: %v", fake.commands)
	}
}

func TestRedisQueueRecoversStaleJobWithoutReleasingOwnedBookLocks(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{"1"}}
	queue := NewRedisQueueWithDo("qiantie:test:novel", fake.Do)
	const jobID = "job-stale"
	recovered, err := queue.RecoverStaleJob(context.Background(), jobID)
	if err != nil || !recovered {
		t.Fatalf("recovered=%v err=%v", recovered, err)
	}
	if len(fake.commands) != 1 || fake.commands[0][0] != "EVAL" || fake.commands[0][2] != "4" {
		t.Fatalf("stale job recovery must use one EVAL with job, lease and list keys: %v", fake.commands)
	}
	args := fake.commands[0]
	want := []string{queue.jobKey(jobID), queue.leaseKey(jobID), queue.processingKey(), queue.pendingKey(), jobID}
	if !slices.Equal(args[3:], want) || strings.Contains(args[1], "'DEL', KEYS[1]") ||
		!strings.Contains(args[1], "'LPUSH', KEYS[4]") {
		t.Fatalf("stale job lost its payload or book ownership: %v", args)
	}
}

type fakeRedisQueueDo struct {
	commands  [][]string
	responses []any
}

func (f *fakeRedisQueueDo) Do(_ context.Context, args ...string) (any, error) {
	f.commands = append(f.commands, append([]string(nil), args...))
	if len(f.responses) == 0 {
		return "OK", nil
	}
	value := f.responses[0]
	f.responses = f.responses[1:]
	return value, nil
}

func (f *fakeRedisQueueDo) hasCommand(name string) bool {
	for _, command := range f.commands {
		if len(command) > 0 && command[0] == name {
			return true
		}
	}
	return false
}
