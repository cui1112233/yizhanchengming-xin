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
