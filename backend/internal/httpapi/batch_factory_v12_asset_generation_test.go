package httpapi

import (
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func newTestAssetImageTasks(t *testing.T) *assetImageGenerationTasks {
	t.Helper()
	t.Setenv("QIANTIE_ASSET_IMAGE_TASKS_PATH", filepath.Join(t.TempDir(), "asset-image-tasks.json"))
	return newAssetImageGenerationTasks()
}

// 服务重启后，仍在生成中的任务不应变成 404，而应恢复为明确的失败状态。
func TestAssetImageGenerationTasksMarkInterruptedTasksFailed(t *testing.T) {
	first := newTestAssetImageTasks(t)
	first.put(assetImageGenerationTask{
		ID: "task-1", Owner: "owner", BatchID: "batch", BookID: "book",
		AssetIDs: []string{"asset-1"}, Status: "processing", UpdatedAt: time.Now().UTC(),
	})

	second := newAssetImageGenerationTasks()
	task, ok := second.get("task-1")
	if !ok {
		t.Fatalf("重启后应能读到任务记录")
	}
	if task.Status != "failed" {
		t.Fatalf("中断任务应标记为 failed，实际 %q", task.Status)
	}
	if task.Error == "" {
		t.Fatalf("中断任务应带有原因说明")
	}
}

func TestAssetImageGenerationTasksPersistFinishedStatus(t *testing.T) {
	first := newTestAssetImageTasks(t)
	first.put(assetImageGenerationTask{
		ID: "task-2", Owner: "owner", BatchID: "batch", BookID: "book",
		Status: "processing", UpdatedAt: time.Now().UTC(),
	})
	first.finish("task-2", []batchfactoryv11.BookAssetImage{{AssetID: "asset-1"}}, nil)

	second := newAssetImageGenerationTasks()
	task, ok := second.get("task-2")
	if !ok {
		t.Fatalf("重启后应能读到已完成任务")
	}
	if task.Status != "succeeded" || task.ImageCount != 1 {
		t.Fatalf("已完成任务状态应保留，实际 status=%q count=%d", task.Status, task.ImageCount)
	}
}
