package queue

import (
    "context"
    "testing"
)

type fakePopper struct{ payload []byte }
func (f *fakePopper) Pop(context.Context, string) ([]byte, error) { return append([]byte(nil), f.payload...), nil }

func TestRedisSourceReturnsJobIDOnly(t *testing.T) {
    source := NewRedisSourceWithPopper("qiantie:pipeline:ready", &fakePopper{payload:[]byte(`{"job_id":"job-42"}`)})
    id, err := source.Next(context.Background())
    if err != nil { t.Fatal(err) }
    if id != "job-42" { t.Fatalf("id=%q", id) }
}
