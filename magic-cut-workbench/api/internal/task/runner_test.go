package task

import (
	"context"
	"reflect"
	"testing"
)

type recordingStore struct {
	events  *[]string
	pending []Task
}

func (s *recordingStore) Save(_ context.Context, task Task) error {
	*s.events = append(*s.events, "save:"+task.ID+":"+string(task.Status))
	return nil
}

func (s *recordingStore) Recoverable(context.Context) ([]Task, error) { return s.pending, nil }

type recordingQueue struct{ events *[]string }

func (q recordingQueue) Push(_ context.Context, key, id string) error {
	*q.events = append(*q.events, "queue:"+key+":"+id)
	return nil
}

func TestRunnerPersistsBeforeQueueing(t *testing.T) {
	events := []string{}
	runner := NewRunner(&recordingStore{events: &events}, recordingQueue{events: &events}, "magic_cut:dev:")
	err := runner.Enqueue(context.Background(), Task{ID: "mct_1", Status: StatusPending})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"save:mct_1:pending", "queue:magic_cut:dev:tasks:pending:mct_1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events=%v want=%v", events, want)
	}
}

func TestRunnerRecoveryRequeuesPersistedTasks(t *testing.T) {
	events := []string{}
	store := &recordingStore{events: &events, pending: []Task{{ID: "mct_a", Status: StatusPending}, {ID: "mct_b", Status: StatusRunning}}}
	runner := NewRunner(store, recordingQueue{events: &events}, "magic_cut:dev:")
	if err := runner.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"queue:magic_cut:dev:tasks:pending:mct_a",
		"queue:magic_cut:dev:tasks:pending:mct_b",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events=%v want=%v", events, want)
	}
}

func TestQueueKeyCannotEscapeConfiguredNamespace(t *testing.T) {
	if got := PendingQueueKey("magic_cut:dev:"); got != "magic_cut:dev:tasks:pending" {
		t.Fatalf("key=%q", got)
	}
}
