package httpapi

import (
	"context"
	"errors"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
	"time"
)

func TestScheduleDueWhilePausedResumesExactlyOnce(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	q := novelfetchworkshop.NewMemoryQueue()
	now := time.Now().UTC()
	_ = s.PutConfig(ctx, "owner", map[string]any{"novel_fetch_automation": map[string]any{"enabled": false}, "novel_fetch_schedules": []any{map[string]any{"id": "pause-due", "status": "scheduled", "run_at": now.Add(-time.Minute).Format(time.RFC3339), "input_text": "12345678901\t定时验证", "platform_id": "2", "selected_versions": []any{"ai1"}}}})
	if err := dispatchDueNovelFetchSchedules(ctx, s, q, "owner", now); err != nil {
		t.Fatal(err)
	}
	if _, err := q.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatalf("paused schedule dispatched: %v", err)
	}
	cfg, _ := s.GetConfig(ctx, "owner")
	if cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)["status"] != "paused" {
		t.Fatal("pause state missing")
	}
	cfg["novel_fetch_automation"] = map[string]any{"enabled": true}
	_ = s.PutConfig(ctx, "owner", cfg)
	if err := dispatchDueNovelFetchSchedules(ctx, s, q, "owner", now); err != nil {
		t.Fatal(err)
	}
	job, err := q.Claim(ctx, time.Millisecond)
	if err != nil || !job.AutoSubmit || job.ScheduleID != "pause-due" {
		t.Fatalf("resumed job=%+v err=%v", job, err)
	}
	_ = q.Ack(ctx, job)
	_ = dispatchDueNovelFetchSchedules(ctx, s, q, "owner", now.Add(time.Minute))
	if _, err := q.Claim(ctx, time.Millisecond); !errors.Is(err, novelfetchworkshop.ErrQueueEmpty) {
		t.Fatal("schedule dispatched twice")
	}
}
