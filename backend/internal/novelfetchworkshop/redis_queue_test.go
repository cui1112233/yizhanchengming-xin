package novelfetchworkshop

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// 消费者崩溃后 job 只留在 processing、租约已过期。回收必须只处理这种情况，
// 持有效租约的活跃任务不能被误删。
func TestRedisQueueReapsOnlyProcessingJobsWithoutLease(t *testing.T) {
	ctx := context.Background()
	recovered := ""
	do := func(_ context.Context, args ...string) (any, error) {
		switch args[0] {
		case "LRANGE":
			return []any{"alive", "dead"}, nil
		case "EXISTS":
			if strings.HasSuffix(args[1], ":lease:alive") {
				return int64(1), nil
			}
			return int64(0), nil
		case "GET":
			return `{"id":"dead","owner":"alice","batchId":"batch-1","platformId":"2","bookIds":["book-1"],"dedupeKey":"books:alice:2"}`, nil
		case "EVAL":
			recovered = fmt.Sprint(args[len(args)-1])
			return int64(1), nil
		default:
			return nil, nil
		}
	}
	queue := NewRedisQueueWithDo("qiantie:novel-fetch", do)
	count, err := queue.ReapStaleProcessing(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("reaped=%d, want 1", count)
	}
	if recovered != "dead" {
		t.Fatalf("recovered job=%q, want only the job without a lease", recovered)
	}
}

func TestRedisQueueReapIgnoresLiveLeases(t *testing.T) {
	ctx := context.Background()
	do := func(_ context.Context, args ...string) (any, error) {
		switch args[0] {
		case "LRANGE":
			return []any{"still-running"}, nil
		case "EXISTS":
			return int64(1), nil
		default:
			t.Fatalf("unexpected redis command %q for a live lease", args[0])
			return nil, nil
		}
	}
	queue := NewRedisQueueWithDo("qiantie:novel-fetch", do)
	count, err := queue.ReapStaleProcessing(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("reaped=%d, want 0 for a live lease", count)
	}
}
