package novelfetchworkshop

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestMemoryPartialRetryReleasesOnlyCompletedBookLocks(t *testing.T) {
	ctx := context.Background()
	q := NewMemoryQueue()
	job := ProcessJob{ID: "partial", Owner: "tester", BatchID: "batch", PlatformID: "2", BookIDs: []string{"success", "failed"}}
	if ok, _, err := q.EnqueueUniqueBooks(ctx, job, time.Hour); err != nil || !ok {
		t.Fatalf("enqueue: %v %v", ok, err)
	}
	claimed, _ := q.Claim(ctx, time.Second)
	claimed.BookIDs = []string{"failed"}
	if err := q.Fail(ctx, claimed, true); err != nil {
		t.Fatal(err)
	}
	if _, exists := q.dedupe["book:tester:2:success"]; exists {
		t.Fatal("completed book still locked after partial retry")
	}
	if q.dedupe["book:tester:2:failed"] != job.ID {
		t.Fatal("failed book lost its retry lock")
	}
	duplicate := job
	duplicate.ID = "overlap"
	duplicate.BookIDs = []string{"failed"}
	if queued, existing, err := q.EnqueueUniqueBooks(ctx, duplicate, time.Hour); err != nil || queued || existing != job.ID {
		t.Fatalf("retry lock failed: %v %s %v", queued, existing, err)
	}
	retried, _ := q.Claim(ctx, time.Second)
	if len(retried.DedupeBookIDs) != 1 || retried.DedupeBookIDs[0] != "failed" {
		t.Fatalf("persisted retry lock scope not reduced: %#v", retried)
	}
	if err := q.Ack(ctx, retried); err != nil {
		t.Fatal(err)
	}
	if len(q.dedupe) != 0 {
		t.Fatalf("terminal retry left lock keys: %#v", q.dedupe)
	}
}

// A small stateful Redis command double exercises lock ownership and persisted
// job scopes instead of only asserting the sequence of command names.
type partialRetryRedis struct {
	values map[string]string
	lists  map[string][]string
}

func (r *partialRetryRedis) do(_ context.Context, args ...string) (any, error) {
	switch args[0] {
	case "SET":
		r.values[args[1]] = args[2]
		return "OK", nil
	case "GET":
		if value, ok := r.values[args[1]]; ok {
			return value, nil
		}
		return nil, nil
	case "DEL":
		for _, key := range args[1:] {
			delete(r.values, key)
		}
		return int64(1), nil
	case "RPUSH":
		r.lists[args[1]] = append(r.lists[args[1]], args[2])
		return int64(1), nil
	case "BRPOPLPUSH":
		list := r.lists[args[1]]
		if len(list) == 0 {
			return nil, nil
		}
		id := list[len(list)-1]
		r.lists[args[1]] = list[:len(list)-1]
		r.lists[args[2]] = append(r.lists[args[2]], id)
		return id, nil
	case "LREM":
		var kept []string
		for _, id := range r.lists[args[1]] {
			if id != args[3] {
				kept = append(kept, id)
			}
		}
		r.lists[args[1]] = kept
		return int64(1), nil
	case "EVAL":
		if args[1] != releaseUniqueBooksScript {
			return nil, fmt.Errorf("unexpected script")
		}
		count, _ := strconv.Atoi(args[2])
		owner := args[3+count]
		for _, key := range args[3 : 3+count] {
			if r.values[key] == owner {
				delete(r.values, key)
			}
		}
		return int64(1), nil
	}
	return nil, fmt.Errorf("unexpected command %s", args[0])
}

func TestRedisPartialRetryReleasesCompletedScopeAndPreservesOtherJobs(t *testing.T) {
	ctx := context.Background()
	state := &partialRetryRedis{values: map[string]string{}, lists: map[string][]string{}}
	q := NewRedisQueueWithDo("partial-test", state.do)
	// No DedupeBookIDs deliberately: Claim must upgrade an older persisted job.
	job := ProcessJob{ID: "partial", Owner: "tester", BatchID: "batch", PlatformID: "2", BookIDs: []string{"success", "failed", "new-owner"}, DedupeKey: "books:tester:2"}
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}
	successKey, failedKey, otherKey := q.bookDedupeKey("tester", "2", "success"), q.bookDedupeKey("tester", "2", "failed"), q.bookDedupeKey("tester", "2", "new-owner")
	state.values[successKey], state.values[failedKey], state.values[otherKey] = job.ID, job.ID, "other-job"
	claimed, err := q.Claim(ctx, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	claimed.BookIDs = []string{"failed"}
	if err := q.Fail(ctx, claimed, true); err != nil {
		t.Fatal(err)
	}
	if state.values[successKey] != "" || state.values[failedKey] != job.ID || state.values[otherKey] != "other-job" {
		t.Fatalf("partial retry lock ownership violated: %#v", state.values)
	}
	retried, err := q.Claim(ctx, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(retried.DedupeBookIDs) != 1 || retried.DedupeBookIDs[0] != "failed" {
		t.Fatalf("old scope survived requeue: %#v", retried)
	}
	if err := q.Ack(ctx, retried); err != nil {
		t.Fatal(err)
	}
	if state.values[failedKey] != "" || state.values[successKey] != "" || state.values[q.jobKey(job.ID)] != "" {
		t.Fatalf("terminal retry leaked job locks: %#v", state.values)
	}
	if state.values[otherKey] != "other-job" {
		t.Fatal("ack deleted another job's lock")
	}
}
