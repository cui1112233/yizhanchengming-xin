package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

// 参考实现用 mode=selected|failed|all 解析批量范围；这里锁定 same 语义，
// 防止后续又把「全部」误当成只有显式 ids。
func TestResolveWorkshopTaskIDsSupportsReferenceModes(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	ctx := context.Background()
	documents := []novelfetchworkshop.Document{
		{BookID: "done-1", Meta: map[string]any{"status": "completed"}},
		{BookID: "failed-1", Meta: map[string]any{"status": "failed"}},
		{BookID: "failed-2", Meta: map[string]any{"status": "completed", "siteSubmitStatus": "failed"}},
	}
	for _, document := range documents {
		if err := store.PutDocument(ctx, "tester", document); err != nil {
			t.Fatal(err)
		}
	}

	all, err := resolveWorkshopTaskIDs(ctx, store, "tester", "all", nil, workshopTaskFailed)
	if err != nil || len(all) != len(documents) {
		t.Fatalf("mode=all => %v, %v", all, err)
	}
	failed, err := resolveWorkshopTaskIDs(ctx, store, "tester", "failed", nil, workshopTaskFailed)
	if err != nil || len(failed) != 1 || failed[0] != "failed-1" {
		t.Fatalf("mode=failed => %v, %v", failed, err)
	}
	submitFailed, err := resolveWorkshopTaskIDs(ctx, store, "tester", "failed", nil, workshopSubmitFailed)
	if err != nil || len(submitFailed) != 1 || submitFailed[0] != "failed-2" {
		t.Fatalf("submit mode=failed => %v, %v", submitFailed, err)
	}
	selected, err := resolveWorkshopTaskIDs(ctx, store, "tester", "selected", []string{"done-1", "done-1"}, workshopTaskFailed)
	if err != nil || len(selected) != 1 || selected[0] != "done-1" {
		t.Fatalf("mode=selected => %v, %v", selected, err)
	}
}
