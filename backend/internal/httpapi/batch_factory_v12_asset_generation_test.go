package httpapi

import (
	"errors"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestAssetImageGenerationTasksExposeTerminalFailure(t *testing.T) {
	tasks := newAssetImageGenerationTasks()
	tasks.put(assetImageGenerationTask{ID: "task-1", Owner: "alice", BatchID: "batch-1", BookID: "book-1", AssetIDs: []string{"asset-1"}, Status: "processing"})
	tasks.finish("task-1", nil, errors.New("provider returned HTTP 503"))

	task, found := tasks.get("task-1")
	if !found || task.Status != "failed" || task.Error != "provider returned HTTP 503" {
		t.Fatalf("task = %#v, found=%v", task, found)
	}
	if task.ImageCount != 0 || len(task.AssetIDs) != 1 || task.AssetIDs[0] != "asset-1" {
		t.Fatalf("task result = %#v", task)
	}
}

func TestAssetImageGenerationTaskWithoutImagesCannotSucceed(t *testing.T) {
	tasks := newAssetImageGenerationTasks()
	tasks.put(assetImageGenerationTask{ID: "task-empty", Status: "processing"})
	tasks.finish("task-empty", nil, nil)

	task, found := tasks.get("task-empty")
	if !found || task.Status != "failed" || task.Error == "" {
		t.Fatalf("empty result must be a visible terminal failure: task=%#v found=%v", task, found)
	}
}

func TestAssetImageGenerationTasksKeepPartialSuccessVisible(t *testing.T) {
	tasks := newAssetImageGenerationTasks()
	tasks.put(assetImageGenerationTask{ID: "task-2", Status: "processing"})
	tasks.finish("task-2", []batchfactoryv11.BookAssetImage{{ID: "image-1"}}, errors.New("second asset failed"))

	task, found := tasks.get("task-2")
	if !found || task.Status != "succeeded" || task.ImageCount != 1 || task.Error != "second asset failed" {
		t.Fatalf("task = %#v, found=%v", task, found)
	}
}
