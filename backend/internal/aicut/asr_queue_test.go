package aicut

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRedisASRQueueNotifiesAndWaits(t *testing.T) {
	commands := [][]string{}
	queue := NewRedisASRQueue("qiantie:aicut:asr", func(_ context.Context, args ...string) (any, error) {
		commands = append(commands, append([]string(nil), args...))
		if args[0] == "BLPOP" {
			return []any{"qiantie:aicut:asr:wake", "asr-1"}, nil
		}
		return int64(1), nil
	})
	if err := queue.Notify(context.Background(), "asr-1"); err != nil {
		t.Fatal(err)
	}
	if err := queue.Wait(context.Background(), time.Second); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"RPUSH", "qiantie:aicut:asr:wake", "asr-1"}, {"BLPOP", "qiantie:aicut:asr:wake", "1"}}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("commands=%#v", commands)
	}
}
