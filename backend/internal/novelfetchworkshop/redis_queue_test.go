package novelfetchworkshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/mergeworker"
)

func TestRedisQueueFailRetryUsesOneAtomicCommand(t *testing.T) {
	ctx := context.Background()
	job := ProcessJob{
		ID: "retry", LeaseToken: "lease-retry", Owner: "alice", BatchID: "batch-1", PlatformID: "2", Attempts: 2,
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
			if len(args) != 3+len(keys)+3 || args[2] != strconv.Itoa(len(keys)) ||
				!reflect.DeepEqual(args[3:3+len(keys)], keys) || args[3+len(keys)] != job.ID || args[4+len(keys)] != job.LeaseToken {
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
	job := ProcessJob{ID: "missing", LeaseToken: "lease", Owner: "alice", BatchID: "batch-1", BookIDs: []string{"book-1"}}
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
			if args[2] == "3" {
				return "OK", nil
			}
			if len(args) != 8 || args[2] != "4" {
				t.Fatalf("reaper must use one atomic command per id: %v", args)
			}
			id := args[7]
			checked = append(checked, id)
			if args[3] != "test:job:"+id || args[4] != "test:lease:"+id || args[5] != "test:processing" || args[6] != "test:pending" {
				t.Fatalf("unexpected recovery key contract: %v", args[3:7])
			}
			for _, command := range []string{"'GET', KEYS[2]", "'GET', KEYS[1]", "'LREM', KEYS[3]", "'LPUSH', KEYS[4]"} {
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

func TestRedisClaimUsesIndependentLeaseTTL(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{"job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1"}`, "OK"}}
	queue := NewRedisQueueWithDo("test", fake.Do)
	job, err := queue.Claim(context.Background(), 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if job.LeaseToken == "" || fake.commands[2][2] != job.LeaseToken {
		t.Fatal("claim did not return the acquired token")
	}
	fake.responses = []any{"job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1"}`, "OK"}
	next, err := queue.Claim(context.Background(), 30*time.Second)
	if err != nil || next.LeaseToken == job.LeaseToken {
		t.Fatalf("second claim reused token or failed: %v", err)
	}
	payload, err := json.Marshal(job)
	if err != nil || strings.Contains(string(payload), job.LeaseToken) {
		t.Fatal("execution token leaked into durable payload")
	}
	set := fake.commands[2]
	if fake.commands[0][3] != "30" || set[0] != "SET" || set[len(set)-2] != "EX" || set[len(set)-1] != "600" {
		t.Fatalf("claim wait=%s lease command=%v", fake.commands[0][3], set)
	}
}

func TestRedisClaimRejectsLostLease(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{"job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1"}`, nil}}
	queue := NewRedisQueueWithDo("test", fake.Do)
	if job, err := queue.Claim(context.Background(), time.Second); err == nil || job.ID != "" {
		t.Fatalf("conflicting claim returned job=%+v err=%v", job, err)
	}
}

func TestRedisClaimRejectsReapedLeaseWindow(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{"job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1"}`, "OK", nil}}
	queue := NewRedisQueueWithDo("test", fake.Do)
	if _, err := queue.Claim(context.Background(), time.Second); !errors.Is(err, ErrProcessLeaseLost) {
		t.Fatalf("reaped between pop and SET: %v", err)
	}
}

func TestRedisRenewLeaseRequiresMatchingToken(t *testing.T) {
	for _, reply := range []int64{0, 1} {
		fake := &fakeRedisQueueDo{responses: []any{reply}}
		queue := NewRedisQueueWithDo("test", fake.Do)
		err := queue.RenewLease(context.Background(), ProcessJob{ID: "job", LeaseToken: "worker"}, 42*time.Second)
		if (reply == 0 && !errors.Is(err, ErrProcessLeaseLost)) || (reply == 1 && err != nil) {
			t.Fatalf("reply=%d err=%v", reply, err)
		}
		args := fake.commands[0]
		if args[0] != "EVAL" || !slices.Equal(args[2:], []string{"1", "test:lease:job", "worker", "42"}) || !strings.Contains(args[1], "'EXPIRE'") || !strings.Contains(args[1], "~= ARGV[1]") {
			t.Fatalf("renew lacks atomic token comparison: %v", args)
		}
	}
}

func TestRedisAckRejectsLostLease(t *testing.T) {
	ctx := context.Background()
	job := ProcessJob{ID: "job-1", Owner: "alice", BatchID: "batch", LeaseToken: "old-token"}
	for _, op := range []string{"ack", "fail", "retry"} {
		fake := &fakeRedisQueueDo{responses: []any{int64(0)}}
		queue := NewRedisQueueWithDo("test", fake.Do)
		var err error
		if op == "ack" {
			err = queue.Ack(ctx, job)
		} else {
			err = queue.Fail(ctx, job, op == "retry")
		}
		if !errors.Is(err, ErrProcessLeaseLost) || len(fake.commands) != 1 {
			t.Fatalf("%s err=%v commands=%v", op, err, fake.commands)
		}
		args := fake.commands[0]
		if args[0] != "EVAL" || !strings.Contains(args[1], "redis.call('GET', KEYS[2]) ~= ARGV[2]") || !slices.Contains(args, job.LeaseToken) {
			t.Fatalf("%s did not guard the transition: %v", op, args)
		}
	}
}

func TestRedisReaperKeepsMatchingLiveLease(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{[]string{"job"}, int64(0)}}
	q := NewRedisQueueWithDo("test", fake.Do)
	count, err := q.ReapStaleProcessing(context.Background())
	if err != nil || count != 0 || len(fake.commands) != 2 || !strings.Contains(fake.commands[1][1], "if redis.call('GET', KEYS[2]) then return 0 end") {
		t.Fatalf("live lease reaped: count=%d err=%v commands=%v", count, err, fake.commands)
	}
}

func TestRedisOwnsLeaseChecksClaimToken(t *testing.T) {
	for _, tc := range []struct {
		value any
		want  bool
	}{{"new", false}, {nil, false}, {"old", true}} {
		fake := &fakeRedisQueueDo{responses: []any{tc.value}}
		q := NewRedisQueueWithDo("test", fake.Do)
		ok, err := q.OwnsLease(context.Background(), ProcessJob{ID: "job", LeaseToken: "old"})
		if ok != tc.want || err != nil || !slices.Equal(fake.commands[0], []string{"GET", "test:lease:job"}) {
			t.Fatalf("value=%v got=%v err=%v", tc.value, ok, err)
		}
	}
}

func TestRedisLeaseHandlesMissingValuesFromProductionTransport(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		const request = "*2\r\n$3\r\nGET\r\n$6\r\nabsent\r\n"
		if _, err := io.ReadFull(conn, make([]byte, len(request))); err == nil {
			_, _ = io.WriteString(conn, "$-1\r\n")
		}
	}()
	transport, err := mergeworker.NewRedisStore("redis://"+ln.Addr().String(), "test")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, missing := transport.Do(ctx, "GET", "absent")
	if !mergeworker.IsRedisNil(missing) {
		t.Fatalf("expected a real RESP nil reply, got %v", missing)
	}
	q := NewRedisQueueWithDo("test", func(context.Context, ...string) (any, error) { return nil, missing })
	owned, err := q.OwnsLease(ctx, ProcessJob{ID: "job", LeaseToken: "old"})
	if owned || err != nil {
		t.Errorf("expired lease: owned=%v err=%v", owned, err)
	}
	if _, err := q.Claim(ctx, time.Second); !errors.Is(err, ErrQueueEmpty) {
		t.Errorf("empty queue: %v", err)
	}
	q.Do = func(_ context.Context, args ...string) (any, error) {
		switch args[0] {
		case "BRPOPLPUSH":
			return "job", nil
		case "GET":
			return `{"id":"job","owner":"alice","batchId":"batch"}`, nil
		default:
			return nil, missing
		}
	}
	if _, err := q.Claim(ctx, time.Second); !errors.Is(err, ErrProcessLeaseLost) {
		t.Errorf("SET NX refused: %v", err)
	}
}
