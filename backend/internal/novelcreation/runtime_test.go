package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRuntimeLockRejectsSecondWriterAndCancellationPersists(t *testing.T) {
	runtime := NewMemoryRuntime()
	release, err := runtime.AcquireProject(context.Background(), "project-1", "run-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	if _, err := runtime.AcquireProject(context.Background(), "project-1", "run-2", time.Minute); !IsCode(err, ErrProjectLocked) {
		t.Fatalf("err=%v", err)
	}
	if err := runtime.Cancel(context.Background(), "run-1"); err != nil {
		t.Fatal(err)
	}
	if cancelled, err := runtime.IsCancelled(context.Background(), "run-1"); err != nil || !cancelled {
		t.Fatalf("cancelled=%t err=%v", cancelled, err)
	}
}

func TestMemoryRuntimeQueuesRunIDs(t *testing.T) {
	runtime := NewMemoryRuntime()
	if err := runtime.Enqueue(context.Background(), "run-1"); err != nil {
		t.Fatal(err)
	}
	runID, err := runtime.Dequeue(context.Background())
	if err != nil || runID != "run-1" {
		t.Fatalf("runID=%q err=%v", runID, err)
	}
}

func TestMemoryRuntimeRetainsPublishedEvent(t *testing.T) {
	runtime := NewMemoryRuntime()
	event := RunEvent{RunID: "run-1", Type: "status", Status: RunPlanning}
	if err := runtime.Publish(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if len(runtime.events) != 1 || !reflect.DeepEqual(runtime.events[0], event) {
		t.Fatalf("events=%+v", runtime.events)
	}
}

func TestRedisRuntimeLockReleasesOnlyMatchingToken(t *testing.T) {
	redis := newRuntimeRedisFake()
	runtime := NewRedisRuntime("example", redis.Do)

	releaseFirst, err := runtime.AcquireProject(context.Background(), "project-1", "run-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.AcquireProject(context.Background(), "project-1", "run-2", time.Minute); !IsCode(err, ErrProjectLocked) {
		t.Fatalf("second acquire err=%v", err)
	}
	releaseFirst()

	releaseSecond, err := runtime.AcquireProject(context.Background(), "project-1", "run-2", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseSecond()
	releaseFirst()
	if _, err := runtime.AcquireProject(context.Background(), "project-1", "run-3", time.Minute); !IsCode(err, ErrProjectLocked) {
		t.Fatalf("stale release removed another writer lock: %v", err)
	}
}

func TestRedisRuntimeCancellationUses24HourTTL(t *testing.T) {
	redis := newRuntimeRedisFake()
	runtime := NewRedisRuntime("example", redis.Do)
	if err := runtime.Cancel(context.Background(), "run-1"); err != nil {
		t.Fatal(err)
	}
	if got := redis.ttl[runtime.cancelKey("run-1")]; got != 24*60*60 {
		t.Fatalf("cancel ttl=%d want=%d", got, 24*60*60)
	}
	if cancelled, err := runtime.IsCancelled(context.Background(), "run-1"); err != nil || !cancelled {
		t.Fatalf("cancelled=%t err=%v", cancelled, err)
	}
}

func TestRedisRuntimeQueuesRunIDs(t *testing.T) {
	redis := newRuntimeRedisFake()
	runtime := NewRedisRuntime("example", redis.Do)
	var _ Runtime = runtime

	if err := runtime.Enqueue(context.Background(), "run-1"); err != nil {
		t.Fatal(err)
	}
	runID, err := runtime.Dequeue(context.Background())
	if err != nil || runID != "run-1" {
		t.Fatalf("runID=%q err=%v", runID, err)
	}
}

func TestRedisRuntimePublishesSerializedWakeEvent(t *testing.T) {
	redis := newRuntimeRedisFake()
	runtime := NewRedisRuntime("example", redis.Do)
	event := RunEvent{EventID: 9, RunID: "run-1", SessionID: "session-1", ProjectID: "project-1", Type: "status", Step: "planning", Status: RunPlanning, Message: "正在规划", Progress: 20, Payload: json.RawMessage(`{"phase":"start"}`)}
	if err := runtime.Publish(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	payload := redis.published[runtime.eventChannel("run-1")]
	var published RunEvent
	if err := json.Unmarshal([]byte(payload), &published); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(published, event) {
		t.Fatalf("published=%+v want=%+v", published, event)
	}
}

func TestRedisRuntimeUsesFixedPrefixAndFullCommands(t *testing.T) {
	redis := newRuntimeRedisFake()
	runtime := NewRedisRuntime("caller-prefix-must-not-change-production-keys", redis.Do)
	ctx := context.Background()
	const prefix = "qiantie:novel-creation"

	if err := runtime.Enqueue(ctx, "run-1"); err != nil {
		t.Fatal(err)
	}
	if runID, err := runtime.Dequeue(ctx); err != nil || runID != "run-1" {
		t.Fatalf("runID=%q err=%v", runID, err)
	}
	release, err := runtime.AcquireProject(ctx, "project-1", "run-1", 1500*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Cancel(ctx, "run-1"); err != nil {
		t.Fatal(err)
	}
	if cancelled, err := runtime.IsCancelled(ctx, "run-1"); err != nil || !cancelled {
		t.Fatalf("cancelled=%t err=%v", cancelled, err)
	}
	event := RunEvent{EventID: 7, RunID: "run-1", SessionID: "session-1", ProjectID: "project-1", Type: "status", Status: RunPlanning, Payload: json.RawMessage(`{"step":"planning"}`)}
	if err := runtime.Publish(ctx, event); err != nil {
		t.Fatal(err)
	}
	release()

	if got, want := len(redis.calls), 7; got != want {
		t.Fatalf("redis calls=%d want=%d: %#v", got, want, redis.calls)
	}
	assertRuntimeRedisCall(t, redis.calls[0], "RPUSH", prefix+":queue", "run-1")
	assertRuntimeRedisCall(t, redis.calls[1], "LPOP", prefix+":queue")

	acquire := redis.calls[2]
	if len(acquire) != 6 || acquire[0] != "EVAL" || acquire[1] != "\nif redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2]) then return 1 end\nreturn 0\n" || acquire[2] != "1" || acquire[3] != prefix+":project-lock:project-1" || acquire[4] == "" || acquire[5] != "1500" {
		t.Fatalf("acquire command=%#v", acquire)
	}
	assertRuntimeRedisCall(t, redis.calls[3], "SET", prefix+":cancel:run-1", "1", "EX", "86400")
	assertRuntimeRedisCall(t, redis.calls[4], "EXISTS", prefix+":cancel:run-1")

	publish := redis.calls[5]
	if len(publish) != 3 || publish[0] != "PUBLISH" || publish[1] != prefix+":events:run-1" {
		t.Fatalf("publish command=%#v", publish)
	}
	var published RunEvent
	if err := json.Unmarshal([]byte(publish[2]), &published); err != nil || !reflect.DeepEqual(published, event) {
		t.Fatalf("published=%+v err=%v", published, err)
	}

	releaseCall := redis.calls[6]
	if len(releaseCall) != 5 || releaseCall[0] != "EVAL" || releaseCall[1] != "\nif redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) end\nreturn 0\n" || releaseCall[2] != "1" || releaseCall[3] != prefix+":project-lock:project-1" || releaseCall[4] != acquire[4] {
		t.Fatalf("release command=%#v", releaseCall)
	}
}

func TestRedisRuntimeReleaseFailureLeavesNewerLockUntouched(t *testing.T) {
	redis := newRuntimeRedisFake()
	runtime := NewRedisRuntime("ignored", redis.Do)
	release, err := runtime.AcquireProject(context.Background(), "project-1", "run-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	key := "qiantie:novel-creation:project-lock:project-1"
	if redis.locks[key] == "" {
		t.Fatalf("missing lock at fixed key %q", key)
	}
	redis.nextErr = errors.New("redis connection lost during release")
	release()
	if redis.locks[key] == "" {
		t.Fatal("failed release unexpectedly removed its lock")
	}
	redis.locks[key] = "newer-writer-token"
	release()
	if got := redis.locks[key]; got != "newer-writer-token" {
		t.Fatalf("stale release removed newer lock: %q", got)
	}
}

func TestRedisRuntimePropagatesTransportErrors(t *testing.T) {
	cause := errors.New("redis unavailable")
	runtime := NewRedisRuntime("ignored", func(context.Context, ...string) (any, error) { return nil, cause })
	assertRedisRuntimeOperationErrors(t, runtime, cause)
}

func TestRedisRuntimeWithoutDoDoesNotFallBackToMemory(t *testing.T) {
	runtime := NewRedisRuntime("ignored", nil)
	assertRedisRuntimeOperationErrors(t, runtime, nil)
}

func assertRedisRuntimeOperationErrors(t *testing.T, runtime *RedisRuntime, cause error) {
	t.Helper()
	event := RunEvent{RunID: "run-1", Type: "status", Status: RunPlanning}
	operations := []struct {
		name string
		run  func() error
	}{
		{"Enqueue", func() error { return runtime.Enqueue(context.Background(), "run-1") }},
		{"Dequeue", func() error { _, err := runtime.Dequeue(context.Background()); return err }},
		{"AcquireProject", func() error {
			_, err := runtime.AcquireProject(context.Background(), "project-1", "run-1", time.Minute)
			return err
		}},
		{"Cancel", func() error { return runtime.Cancel(context.Background(), "run-1") }},
		{"IsCancelled", func() error { _, err := runtime.IsCancelled(context.Background(), "run-1"); return err }},
		{"Publish", func() error { return runtime.Publish(context.Background(), event) }},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			err := operation.run()
			if err == nil {
				t.Fatal("unexpected success; Redis must not fall back to memory")
			}
			if cause != nil && !errors.Is(err, cause) {
				t.Fatalf("err=%v does not wrap cause=%v", err, cause)
			}
		})
	}
}

func assertRuntimeRedisCall(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("redis command=%#v want=%#v", got, want)
	}
}

type runtimeRedisFake struct {
	queue     []string
	locks     map[string]string
	values    map[string]string
	ttl       map[string]int
	published map[string]string
	calls     [][]string
	nextErr   error
}

func newRuntimeRedisFake() *runtimeRedisFake {
	return &runtimeRedisFake{locks: map[string]string{}, values: map[string]string{}, ttl: map[string]int{}, published: map[string]string{}}
}

func (f *runtimeRedisFake) Do(_ context.Context, args ...string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("missing redis command")
	}
	f.calls = append(f.calls, append([]string(nil), args...))
	if f.nextErr != nil {
		err := f.nextErr
		f.nextErr = nil
		return nil, err
	}
	switch args[0] {
	case "RPUSH":
		f.queue = append(f.queue, args[2])
		return int64(len(f.queue)), nil
	case "LPOP":
		if len(f.queue) == 0 {
			return nil, nil
		}
		value := f.queue[0]
		f.queue = f.queue[1:]
		return value, nil
	case "EVAL":
		script, key, token := args[1], args[3], args[4]
		if strings.Contains(script, "GET") {
			if f.locks[key] == token {
				delete(f.locks, key)
				return int64(1), nil
			}
			return int64(0), nil
		}
		if _, locked := f.locks[key]; locked {
			return int64(0), nil
		}
		f.locks[key] = token
		return int64(1), nil
	case "SET":
		f.values[args[1]] = args[2]
		f.ttl[args[1]] = 0
		for index := 3; index+1 < len(args); index += 2 {
			if args[index] == "EX" {
				var seconds int
				if _, err := fmt.Sscan(args[index+1], &seconds); err != nil {
					return nil, err
				}
				f.ttl[args[1]] = seconds
			}
		}
		return "OK", nil
	case "EXISTS":
		if _, exists := f.values[args[1]]; exists {
			return int64(1), nil
		}
		return int64(0), nil
	case "PUBLISH":
		f.published[args[1]] = args[2]
		return int64(1), nil
	default:
		return nil, fmt.Errorf("unsupported redis command %q", args[0])
	}
}
