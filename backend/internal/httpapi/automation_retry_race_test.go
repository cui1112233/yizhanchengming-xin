package httpapi

import (
	"context"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// 回归：Retry 不能把“状态仓库里正在被后台 worker 修改的 map”直接交给
// publicAutomationJob 读取。真实故障会触发 concurrent map read and map write，
// 让整个 Go 服务进程退出。
func TestRetryBuildsPublicResponseBeforeWorkerMutatesJob(t *testing.T) {
	ctx := context.Background()
	for attempt := 0; attempt < 20; attempt++ {
		store := batchfactoryv11.NewMemoryStore()
		batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
			Title: "batch",
			Books: []batchfactoryv11.CreateBookInput{{Title: "book", SourceText: "正文"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		bookID := batch.Books[0].ID
		jobID := "job-retry-snapshot"
		state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
			jobID: {
				"id": jobID, "owner": "alice", "batchId": batch.ID, "state": "failed",
				"books": map[string]any{
					bookID: map[string]any{
						"bookId": bookID, "title": "book", "status": "failed", "stage": "comic_images",
						"nodes": map[string]any{
							"comic_images": map[string]any{
								"pageLedger": map[string]any{"page-1": map[string]any{"needsManual": false}},
							},
						},
					},
				},
			},
		}}
		coordinator := &BatchAutomationCoordinator{
			store:           store,
			automationStore: &memoryAutomationStateStore{state: state, found: true},
			now:             func() time.Time { return time.Date(2026, 11, 8, 1, 30, 0, 0, time.UTC) },
		}

		result, err := coordinator.Retry(ctx, "alice", batch.ID, []string{bookID})
		if err != nil {
			t.Fatalf("attempt %d: Retry() error = %v", attempt, err)
		}
		if result["id"] != jobID {
			t.Fatalf("attempt %d: response id = %v, want %s", attempt, result["id"], jobID)
		}
		coordinator.cancelRun(jobID)
	}
}

func newRetryRaceCoordinator(state automationStateFile) *BatchAutomationCoordinator {
	store := &memoryAutomationStateStore{state: state, found: true}
	return &BatchAutomationCoordinator{automationStore: store, now: func() time.Time {
		return time.Date(2026, 11, 5, 17, 0, 0, 0, time.UTC)
	}}
}

func retryRaceJob(entry map[string]any) automationStateFile {
	return automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {
			"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "running",
			"books": map[string]any{"book-1": entry},
		},
	}}
}

// retryAt 在内存里是 time.Time，落库后才是字符串；到点判定必须两者都认。
func TestRetryDueAcceptsInMemoryTimeValue(t *testing.T) {
	now := time.Date(2026, 11, 5, 17, 0, 0, 0, time.UTC)
	entry := map[string]any{"retryAt": now.Add(-time.Minute)}
	if !automationRetryDue(entry, now) {
		t.Fatal("内存 time.Time 形式的 retryAt 必须判为已到点")
	}
	entry["retryAt"] = now.Add(time.Minute)
	if automationRetryDue(entry, now) {
		t.Fatal("未到点的 retryAt 不得判为到点")
	}
	entry["retryAt"] = now.Add(-time.Minute).Format(time.RFC3339Nano)
	if !automationRetryDue(entry, now) {
		t.Fatal("落库字符串形式的 retryAt 必须判为已到点")
	}
}

// 调度器到点续跑与旧 worker 写回 failed 同时发生：续跑必须最终生效，
// 旧 worker 的迟到写入必须被 runVersion 拒绝。
func TestDispatchReopenDefeatsStaleWorkerWrite(t *testing.T) {
	now := time.Date(2026, 11, 5, 17, 0, 0, 0, time.UTC)
	ledgerRecord := map[string]any{"attempts": 2, "state": "skipped", "lastErrorAt": now.Add(-2 * time.Minute)}
	entry := map[string]any{
		"bookId": "book-1", "status": "failed", "stage": "comic_images",
		"retryRequested": true, "retryAt": now.Add(-time.Minute),
		"attempts":   1,
		"pageLedger": map[string]any{"video-2": ledgerRecord},
	}
	staleClone := cloneAutomationEntry(entry)

	if !reopenAutomationRetryBook(entry, now) {
		t.Fatal("到点续跑必须打开书")
	}
	if entry["status"] != "running" || entry["retryRequested"] != nil || entry["retryAt"] != nil {
		t.Fatalf("reopen 后状态不对: %#v", entry)
	}
	if automationNumberValue(entry[automationBookRunVersionField], 0) != 1 {
		t.Fatalf("runVersion=%v, want 1", entry[automationBookRunVersionField])
	}

	coordinator := newRetryRaceCoordinator(retryRaceJob(entry))
	written, err := coordinator.persistAutomationBookEntry("alice", "batch-1", "job-1", "book-1", staleClone, staleClone, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if written {
		t.Fatal("旧 worker 的失败快照不得覆盖调度器写入的 running")
	}
	state, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	latest, _ := state.Jobs["job-1"]["books"].(map[string]any)["book-1"].(map[string]any)
	if latest["status"] != "running" || latest["retryRequested"] != nil {
		t.Fatalf("续跑被旧写入覆盖: %#v", latest)
	}
	if _, ok := latest["pageLedger"].(map[string]any)["video-2"]; !ok {
		t.Fatal("并发写不得丢失 pageLedger")
	}
	if automationNumberValue(latest["attempts"], 0) != 1 {
		t.Fatalf("attempts=%v, want 1（旧写入不得改动累计次数）", latest["attempts"])
	}

	// 新一轮 worker（版本 1）可以正常写。
	fresh := cloneAutomationEntry(entry)
	fresh["message"] = "第二轮执行中"
	written, err = coordinator.persistAutomationBookEntry("alice", "batch-1", "job-1", "book-1", entry, fresh, 1, nil)
	if err != nil || !written {
		t.Fatalf("新一轮 worker 写入失败: written=%v err=%v", written, err)
	}
	reloaded, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	after, _ := reloaded.Jobs["job-1"]["books"].(map[string]any)["book-1"].(map[string]any)
	if automationStringValue(after["status"], "") != "running" || automationNumberValue(after["attempts"], 0) != 1 {
		t.Fatalf("续跑终态=%#v attempts=%v, want running/1", after["status"], after["attempts"])
	}

	// 续跑只做缺页：已完成页不进入工作集。
	videos := []batchfactoryv11.Video{{ID: "video-1"}, {ID: "video-2"}}
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "video-1", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}},
		{VideoID: "video-2"},
	}
	continuation := buildComicPageContinuation(videos, pages)
	if len(continuation.Pending) != 1 || continuation.Pending[0].ID != "video-2" {
		t.Fatalf("续跑工作集=%+v, want only video-2", continuation.Pending)
	}
}

// needs_manual 的书不得被自动续跑重新激活。
func TestRetryDoesNotReactivateNeedsManual(t *testing.T) {
	now := time.Date(2026, 11, 5, 17, 0, 0, 0, time.UTC)
	entry := map[string]any{
		"bookId": "book-1", "status": "failed", "needsManual": true,
		"retryRequested": true, "retryAt": now.Add(-time.Minute),
	}
	if reopenAutomationRetryBook(entry, now) {
		t.Fatal("needs_manual 的书不得被自动打开")
	}
	if entry["status"] != "failed" || !automationBoolValue(entry["needsManual"]) {
		t.Fatalf("needs_manual 状态被破坏: %#v", entry)
	}
	if entry["retryRequested"] != nil || entry["retryAt"] != nil {
		t.Fatal("needs_manual 书的重试标记必须清理")
	}
	if automationNumberValue(entry[automationBookRunVersionField], 0) != 0 {
		t.Fatal("未重新打开的书不得递增 runVersion")
	}
}

// 多本书同时到点、并发写书，账本不得丢记录。
func TestConcurrentRetryDispatchKeepsLedger(t *testing.T) {
	now := time.Date(2026, 11, 5, 17, 0, 0, 0, time.UTC)
	books := map[string]any{}
	for _, bookID := range []string{"book-1", "book-2"} {
		entry := map[string]any{
			"bookId": bookID, "status": "failed", "stage": "comic_images",
			"retryRequested": true, "retryAt": now.Add(-time.Minute),
			"pageLedger": map[string]any{bookID + "-page": map[string]any{"attempts": 1, "state": "skipped"}},
		}
		if !reopenAutomationRetryBook(entry, now) {
			t.Fatalf("%s 未打开", bookID)
		}
		books[bookID] = entry
	}
	state := automationStateFile{Version: 1, Jobs: map[string]map[string]any{
		"job-1": {"id": "job-1", "owner": "alice", "batchId": "batch-1", "state": "running", "books": books},
	}}
	coordinator := newRetryRaceCoordinator(state)
	clones := map[string]map[string]any{}
	targets := map[string]map[string]any{}
	for _, bookID := range []string{"book-1", "book-2"} {
		targets[bookID] = books[bookID].(map[string]any)
		entry := cloneAutomationEntry(targets[bookID])
		entry["message"] = "并发写入"
		clones[bookID] = entry
	}

	var wait sync.WaitGroup
	for index := 0; index < 20; index++ {
		for _, bookID := range []string{"book-1", "book-2"} {
			wait.Add(1)
			go func(bookID string) {
				defer wait.Done()
				if _, err := coordinator.persistAutomationBookEntry("alice", "batch-1", "job-1", bookID, targets[bookID], clones[bookID], 1, nil); err != nil {
					t.Errorf("persist %s: %v", bookID, err)
				}
			}(bookID)
		}
	}
	wait.Wait()

	latest, err := coordinator.loadLocked()
	if err != nil {
		t.Fatal(err)
	}
	got, _ := latest.Jobs["job-1"]["books"].(map[string]any)
	if len(got) != 2 {
		t.Fatalf("books=%d, want 2", len(got))
	}
	for _, bookID := range []string{"book-1", "book-2"} {
		entry, _ := got[bookID].(map[string]any)
		ledger, _ := entry["pageLedger"].(map[string]any)
		if _, ok := ledger[bookID+"-page"]; !ok {
			t.Fatalf("%s 的 pageLedger 丢失: %#v", bookID, entry)
		}
	}
}

// run 收尾判断「还有没有已排期续跑」时，retryAt 可能是内存 time.Time。
// 漏判会让 job 提前写成 failed 并清掉书上的 retryAt，续跑就永远不会发生。
func TestJobFinalizeKeepsPendingRetryJobRunning(t *testing.T) {
	now := time.Date(2026, 11, 5, 17, 0, 0, 0, time.UTC)
	books := map[string]any{
		"book-1": map[string]any{
			"bookId": "book-1", "status": "completed",
		},
		"book-2": map[string]any{
			"bookId": "book-2", "status": "failed", "stage": "comic_images",
			"retryRequested": true, "retryAt": now.Add(30 * time.Second),
		},
	}
	if !automationPendingBookRetry(books) {
		t.Fatal("内存 time.Time 形式的 retryAt 必须判为存在待续跑计划")
	}
	delete(books["book-2"].(map[string]any), "retryAt")
	if automationPendingBookRetry(books) {
		t.Fatal("没有 retryAt 的书不得判为待续跑")
	}
}
