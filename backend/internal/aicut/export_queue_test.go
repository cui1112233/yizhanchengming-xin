package aicut

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRedisExportQueueNotifiesAndWaits(t *testing.T) {
	commands := make([][]string, 0)
	queue := NewRedisExportQueue("qiantie:aicut:export", func(_ context.Context, args ...string) (any, error) {
		commands = append(commands, append([]string(nil), args...))
		if args[0] == "BLPOP" {
			return []any{"qiantie:aicut:export:wake", "export-1"}, nil
		}
		return int64(1), nil
	})
	if err := queue.Notify(context.Background(), "export-1"); err != nil {
		t.Fatal(err)
	}
	if err := queue.Wait(context.Background(), 2*time.Second); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"RPUSH", "qiantie:aicut:export:wake", "export-1"}, {"BLPOP", "qiantie:aicut:export:wake", "2"}}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands=%#v", commands)
	}
}
