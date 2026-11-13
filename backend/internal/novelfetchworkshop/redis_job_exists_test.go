package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestProcessJobExistsUsesExplicitRedisExistence(t *testing.T) {
	for _, count := range []int64{0, 1} {
		q := NewRedisQueueWithDo("test", func(_ context.Context, args ...string) (any, error) {
			if len(args) != 2 || args[0] != "EXISTS" || args[1] != "test:job:job" {
				t.Fatalf("unexpected command %v", args)
			}
			return count, nil
		})
		exists, err := q.ProcessJobExists(context.Background(), "job")
		if err != nil || exists != (count == 1) {
			t.Fatalf("count=%d exists=%v err=%v", count, exists, err)
		}
	}
}
