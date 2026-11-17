package httpapi

import (
	"context"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
	"time"
)

type orphanCheckQueue struct {
	*novelfetchworkshop.MemoryQueue
	exists bool
}

func (q *orphanCheckQueue) ProcessJobExists(context.Context, string) (bool, error) {
	return q.exists, nil
}
func TestQueuedRecoveryRequiresMissingRedisJob(t *testing.T) {
	for _, exists := range []bool{true, false} {
		s := novelfetchworkshop.NewMemoryStore()
		now := time.Now().UTC()
		doc := novelfetchworkshop.Document{BookID: "orphan", Meta: map[string]any{"status": "queued", "currentProcessJobId": "old"}}
		_ = s.PutDocument(context.Background(), "owner", doc)
		doc.UpdatedAt = now.Add(-time.Hour).Format(time.RFC3339Nano)
		q := &orphanCheckQueue{MemoryQueue: novelfetchworkshop.NewMemoryQueue(), exists: exists}
		recoverStaleNovelFetchTasks(context.Background(), s, "owner", map[string]any{}, []novelfetchworkshop.Document{doc}, now, q)
		got, _ := s.GetDocument(context.Background(), "owner", doc.BookID)
		want := "waiting_retry"
		if exists {
			want = "queued"
		}
		if got.Meta["status"] != want {
			t.Fatalf("exists=%v status=%v", exists, got.Meta["status"])
		}
	}
}

// 没有 currentProcessJobId 的 queued 文档不可能被任何 Redis 任务持有，
// 必须恢复成 waiting_retry，否则会永远停留在“排队中”。
func TestQueuedRecoveryWithoutJobIDIsTreatedAsOrphan(t *testing.T) {
	s := novelfetchworkshop.NewMemoryStore()
	now := time.Now().UTC()
	doc := novelfetchworkshop.Document{BookID: "orphan-no-job", Meta: map[string]any{"status": "queued"}}
	_ = s.PutDocument(context.Background(), "owner", doc)
	doc.UpdatedAt = now.Add(-time.Hour).Format(time.RFC3339Nano)
	q := &orphanCheckQueue{MemoryQueue: novelfetchworkshop.NewMemoryQueue(), exists: true}
	recoverStaleNovelFetchTasks(context.Background(), s, "owner", map[string]any{}, []novelfetchworkshop.Document{doc}, now, q)
	got, _ := s.GetDocument(context.Background(), "owner", doc.BookID)
	if got.Meta["status"] != "waiting_retry" {
		t.Fatalf("status=%v, want waiting_retry", got.Meta["status"])
	}
}
