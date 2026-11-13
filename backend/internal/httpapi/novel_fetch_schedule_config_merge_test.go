package httpapi

import (
	"context"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
)

func TestWorkshopScheduleRuntimeMergePreservesUserConfigAndCancellation(t *testing.T) {
	ctx := context.Background()
	s := novelfetchworkshop.NewMemoryStore()
	_ = s.PutConfig(ctx, "alice", map[string]any{"label": "new user config", "novel_fetch_schedules": []any{map[string]any{"id": "s1", "status": "dispatching", "run_at": "new user time"}}})
	if err := saveNovelFetchScheduleItem(ctx, s, "alice", map[string]any{"id": "s1", "status": "queued", "run_at": "old stale time"}); err != nil {
		t.Fatal(err)
	}
	current, _ := s.GetConfig(ctx, "alice")
	row := mapOrEmpty(arrayOrEmpty(current["novel_fetch_schedules"])[0])
	if current["label"] != "new user config" || row["run_at"] != "new user time" || row["status"] != "queued" {
		t.Fatal("runtime write replaced user fields")
	}
	row["status"] = "cancelled"
	if err := s.PutConfig(ctx, "alice", current); err != nil {
		t.Fatal(err)
	}
	if err := saveNovelFetchScheduleItem(ctx, s, "alice", map[string]any{"id": "s1", "status": "queued"}); err != nil {
		t.Fatal(err)
	}
	current, _ = s.GetConfig(ctx, "alice")
	if mapOrEmpty(arrayOrEmpty(current["novel_fetch_schedules"])[0])["status"] != "cancelled" {
		t.Fatal("cancelled schedule resurrected")
	}
}
