package queue

import (
    "context"
    "encoding/json"
    "testing"
    "time"

    "github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakePusher struct {
    key string
    payload []byte
}

func (f *fakePusher) Push(_ context.Context, key string, payload []byte) error {
    f.key = key
    f.payload = append([]byte(nil), payload...)
    return nil
}

func TestRedisQueueEnqueuesOnlyJobReferencePayload(t *testing.T) {
    pusher := &fakePusher{}
    q := NewRedisQueueWithPusher("qiantie:pipeline:ready", pusher)
    job := pipeline.Job{ID:"job-1", BatchID:"batch-1", RunAt:time.Date(2026,10,4,9,0,0,0,time.UTC), Status:"queued"}

    if err := q.Enqueue(context.Background(), job); err != nil { t.Fatal(err) }
    if pusher.key != "qiantie:pipeline:ready" { t.Fatalf("key=%q", pusher.key) }
    var payload map[string]any
    if err := json.Unmarshal(pusher.payload, &payload); err != nil { t.Fatal(err) }
    if payload["job_id"] != "job-1" { t.Fatalf("payload=%v", payload) }
    if _, ok := payload["stages"]; ok { t.Fatalf("redis payload must not duplicate durable stage state: %v", payload) }
}
