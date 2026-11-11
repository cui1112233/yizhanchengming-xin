package aicut

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRedisAgentQueueNotifyAndWaitPreserveRunID(t *testing.T) {
	commands := make([][]string, 0)
	queue := NewRedisAgentQueue("qiantie:aicut:agent", func(_ context.Context, command ...string) (any, error) {
		commands = append(commands, append([]string(nil), command...))
		if command[0] == "BLPOP" {
			return []any{"qiantie:aicut:agent:wake", "run-1"}, nil
		}
		return int64(1), nil
	})
	if err := queue.Notify(context.Background(), "run-1"); err != nil {
		t.Fatal(err)
	}
	runID, err := queue.Wait(context.Background(), time.Second)
	if err != nil || runID != "run-1" {
		t.Fatalf("Wait() = %q, %v", runID, err)
	}
	want := [][]string{{"RPUSH", "qiantie:aicut:agent:wake", "run-1"}, {"BLPOP", "qiantie:aicut:agent:wake", "1"}}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands=%v want=%v", commands, want)
	}
}
