package mergeworker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestEncodeRESPCommand(t *testing.T) {
	got := string(encodeRESPCommand("SET", "job:key", "value"))
	want := "*3\r\n$3\r\nSET\r\n$7\r\njob:key\r\n$5\r\nvalue\r\n"
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestReadRESPArray(t *testing.T) {
	reader := bufio.NewReader(bytes.NewBufferString("*2\r\n$5\r\nqueue\r\n$8\r\ntask-123\r\n"))
	value, err := readRESP(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value, []any{"queue", "task-123"}) {
		t.Fatalf("value=%#v", value)
	}
}

func TestRedisStoreCommandContract(t *testing.T) {
	fake := &fakeRedisDo{}
	store := &RedisStore{Prefix: "qiantie:test:merge", Do: fake.Do}
	job := Job{ID: "task-1", BatchID: "batch-1", Status: StateQueued, Sources: []Source{{ProductionJobID: "p1", VideoID: "v1", MediaURL: "https://media.example/1.mp4", Order: 0}}}
	created, err := store.Create(context.Background(), job)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != job.ID {
		t.Fatalf("created=%+v", created)
	}
	if err := store.Enqueue(context.Background(), job.ID); err != nil {
		t.Fatal(err)
	}
	fake.next = []any{"qiantie:test:merge:queue", "task-1"}
	id, err := store.Dequeue(context.Background(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if id != "task-1" {
		t.Fatalf("id=%q", id)
	}
	if len(fake.commands) < 3 || fake.commands[0][0] != "SET" || fake.commands[1][0] != "RPUSH" || fake.commands[2][0] != "BLPOP" {
		t.Fatalf("commands=%v", fake.commands)
	}
	var persisted Job
	if err := json.Unmarshal([]byte(fake.commands[0][2]), &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.Status != StateQueued || len(persisted.Sources) != 1 {
		t.Fatalf("persisted=%+v", persisted)
	}
}

type fakeRedisDo struct {
	commands [][]string
	next     any
}

func (f *fakeRedisDo) Do(_ context.Context, args ...string) (any, error) {
	f.commands = append(f.commands, append([]string(nil), args...))
	if f.next != nil {
		value := f.next
		f.next = nil
		return value, nil
	}
	return "OK", nil
}

// Exercise the transport deadline with a Redis reply arriving after the
// default ten-second I/O timeout but inside the requested blocking wait.
func TestRedisBlockingClaimWaitsForLateReply(t *testing.T) {
	if testing.Short() {
		t.Skip("exercises the real ten-second socket deadline")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 16*time.Second)
	defer cancel()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		if _, err := readRESP(bufio.NewReader(conn)); err != nil {
			return
		}
		select {
		case <-time.After(11 * time.Second):
			_, _ = io.WriteString(conn, "$3\r\njob\r\n")
		case <-ctx.Done():
		}
	}()
	store, err := NewRedisStore("redis://"+ln.Addr().String(), "test")
	if err != nil {
		t.Fatal(err)
	}
	value, err := store.Do(ctx, "BRPOPLPUSH", "pending", "processing", "30")
	if err != nil || value != "job" {
		t.Fatalf("blocking claim was cut short: value=%v err=%v", value, err)
	}
}
