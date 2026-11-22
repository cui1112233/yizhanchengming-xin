package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRedisQueueFailRetryUsesOneAtomicCommand(t *testing.T) {
	ctx := context.Background()
	job := ProcessJob{
		ID: "retry", Owner: "alice", BatchID: "batch-1", PlatformID: "2", Attempts: 2,
		BookIDs: []string{"retry-book"}, DedupeBookIDs: []string{"completed-book", "retry-book"},
		DedupeKey: "books:alice:2", LastError: "persistence_checkpoint:{\"type\":\"original_fetch\",\"version\":1}",
	}
	connectionErr := errors.New("simulated Redis connection failure")
	for _, replyErr := range []error{nil, connectionErr} {
		t.Run(fmt.Sprint(replyErr), func(t *testing.T) {
			var calls [][]string
			queue := NewRedisQueueWithDo("test", func(_ context.Context, args ...string) (any, error) {
				calls = append(calls, slices.Clone(args))
				if args[0] == "RPUSH" || (args[0] == "EVAL" && len(args) >= 9) {
					return int64(1), replyErr
				}
				return int64(1), nil
			})
			err := queue.Fail(ctx, job, true)
			if !errors.Is(err, replyErr) {
				t.Fatalf("Fail error=%v, want %v", err, replyErr)
			}
			if len(calls) != 1 || calls[0][0] != "EVAL" {
				t.Fatalf("Fail retry issued independent writes: %v", calls)
			}
			args := calls[0]
			keys := []string{queue.jobKey(job.ID), queue.leaseKey(job.ID), queue.processingKey(), queue.pendingKey(), queue.bookDedupeKey(job.Owner, job.PlatformID, "completed-book")}
			if len(args) != 3+len(keys)+2 || args[2] != strconv.Itoa(len(keys)) ||
				!reflect.DeepEqual(args[3:3+len(keys)], keys) || args[3+len(keys)] != job.ID {
				t.Fatalf("unexpected retry key contract: %v", args[2:])
			}
			var retryJob ProcessJob
			if err := json.Unmarshal([]byte(args[len(args)-1]), &retryJob); err != nil {
				t.Fatal(err)
			}
			if retryJob.LastError != job.LastError || !reflect.DeepEqual(retryJob.DedupeBookIDs, job.BookIDs) || retryJob.Attempts != job.Attempts {
				t.Fatalf("retry payload lost checkpoint or retry scope: %#v", retryJob)
			}
			for _, command := range []string{"'GET', KEYS[1]", "'TYPE'", "'LREM', KEYS[3]", "'SET', KEYS[1]", "'LPUSH', KEYS[4]", "'DEL', KEYS[2]"} {
				if !strings.Contains(args[1], command) {
					t.Errorf("retry script missing %s", command)
				}
			}
		})
	}
}

func TestRedisQueueFailRetryRejectsMissingJob(t *testing.T) {
	queue := NewRedisQueueWithDo("test", func(_ context.Context, args ...string) (any, error) {
		if args[0] != "EVAL" {
			t.Fatalf("missing job mutated Redis outside EVAL: %v", args)
		}
		return int64(0), nil
	})
	job := ProcessJob{ID: "missing", Owner: "alice", BatchID: "batch-1", BookIDs: []string{"book-1"}}
	if err := queue.Fail(context.Background(), job, true); err == nil {
		t.Fatal("missing job should not be reported as requeued")
	}
}

func TestRedisQueueReapPreservesCheckpointAndCleansMissingPayload(t *testing.T) {
	ctx := context.Background()
	const checkpoint = "persistence_checkpoint:{\"type\":\"original_fetch\",\"version\":1}"
	leased := map[string]bool{"alive": true}
	payloads := map[string]string{
		"dead": `{"id":"dead","owner":"alice","batchId":"batch-1","platformId":"2","bookIds":["book-1"],"lastError":"persistence_checkpoint:{\"type\":\"original_fetch\",\"version\":1}"}`,
	}
	processing := []string{"alive", "dead", "orphan"}
	var pending []string
	var checked []string
	queue := NewRedisQueueWithDo("test", func(_ context.Context, args ...string) (any, error) {
		switch args[0] {
		case "LRANGE":
			return slices.Clone(processing), nil
		case "EVAL":
			if len(args) != 8 || args[2] != "4" {
				t.Fatalf("reaper must use one atomic command per id: %v", args)
			}
			id := args[7]
			checked = append(checked, id)
			if args[3] != "test:job:"+id || args[4] != "test:lease:"+id || args[5] != "test:processing" || args[6] != "test:pending" {
				t.Fatalf("unexpected recovery key contract: %v", args[3:7])
			}
			for _, command := range []string{"'EXISTS', KEYS[2]", "'GET', KEYS[1]", "'LREM', KEYS[3]", "'LPUSH', KEYS[4]"} {
				if !strings.Contains(args[1], command) {
					t.Errorf("recovery script missing %s", command)
				}
			}
			if strings.Contains(args[1], "'DEL', KEYS[1]") {
				t.Fatal("recovery script deletes the job payload")
			}
			if leased[id] {
				return int64(0), nil
			}
			processing = slices.DeleteFunc(processing, func(item string) bool { return item == id })
			if payloads[id] != "" {
				pending = append(pending, id)
			}
			return int64(1), nil
		case "BRPOPLPUSH":
			if len(pending) == 0 {
				t.Fatal("recovered job was not in pending")
			}
			id := pending[0]
			pending = pending[1:]
			return id, nil
		case "GET":
			return payloads[strings.TrimPrefix(args[1], "test:job:")], nil
		case "SET":
			return "OK", nil
		default:
			t.Fatalf("unexpected non-atomic recovery command: %v", args)
			return nil, nil
		}
	})
	count, err := queue.ReapStaleProcessing(ctx)
	if err != nil || count != 2 || !reflect.DeepEqual(checked, []string{"alive", "dead", "orphan"}) {
		t.Fatalf("reaped=%d checked=%v err=%v", count, checked, err)
	}
	if !reflect.DeepEqual(processing, []string{"alive"}) || !reflect.DeepEqual(pending, []string{"dead"}) {
		t.Fatalf("processing=%v pending=%v", processing, pending)
	}
	job, err := queue.Claim(ctx, time.Second)
	if err != nil || job.ID != "dead" || job.LastError != checkpoint {
		t.Fatalf("claimed recovered job lost checkpoint: job=%#v err=%v", job, err)
	}
}
