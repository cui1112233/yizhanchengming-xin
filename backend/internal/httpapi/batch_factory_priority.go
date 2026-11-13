package httpapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func (c *BatchAutomationCoordinator) claimAutomationPriorityBookWithRetry(ctx context.Context, owner, batchID, jobID string, epoch int64, fallback batchfactoryv11.Book) (batchfactoryv11.Book, map[string]any, error) {
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		book, entry, err := c.claimAutomationPriorityBook(ctx, owner, batchID, jobID, epoch, fallback)
		if err == nil {
			return book, entry, nil
		}
		last = err
		if ctx.Err() != nil || errors.Is(err, batchfactoryv11.ErrConflict) || errors.Is(err, batchfactoryv11.ErrNotFound) {
			break
		}
		if attempt < 2 {
			select {
			case <-ctx.Done():
				return fallback, nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 150 * time.Millisecond):
			}
		}
	}
	return fallback, nil, last
}

func (c *BatchAutomationCoordinator) markAutomationPriorityAttention(owner, batchID, jobID string, epoch int64, cause error) error {
	unlock := c.lockBatchRun(owner, batchID)
	defer unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isCurrentRun(jobID, epoch) {
		return nil
	}
	state, err := c.loadLocked()
	if err != nil {
		return err
	}
	job := state.Jobs[jobID]
	if job == nil || job["owner"] != owner || job["batchId"] != batchID || job["state"] != "running" {
		return nil
	}
	state, job = writablePriorityJob(state, jobID)
	job["state"], job["lastError"], job["updatedAt"] = "needs_attention", "领取下一本失败，队列已保留，请稍后点击继续："+cause.Error(), c.now()
	return c.saveLocked(state)
}

func automationPriorityIDs(job map[string]any) []string {
	var ids []string
	switch raw := job["priorityBookIds"].(type) {
	case []string:
		ids = append(ids, raw...)
	case []any:
		for _, v := range raw {
			if id, ok := v.(string); ok {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func writablePriorityJob(state automationStateFile, jobID string) (automationStateFile, map[string]any) {
	jobs := make(map[string]map[string]any, len(state.Jobs))
	for id, job := range state.Jobs {
		jobs[id] = job
	}
	state.Jobs = jobs
	job := cloneAutomationMap(jobs[jobID])
	jobs[jobID] = job
	return state, job
}

func (c *BatchAutomationCoordinator) PrioritizeBook(ctx context.Context, owner, batchID, bookID string) (map[string]any, error) {
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	var selected *batchfactoryv11.Book
	for i := range batch.Books {
		if batch.Books[i].ID == strings.TrimSpace(bookID) {
			selected = &batch.Books[i]
			break
		}
	}
	if selected == nil {
		return nil, batchfactoryv11.ErrNotFound
	}
	unlockBatch := c.lockBatchRun(owner, batchID)
	defer unlockBatch()
	c.mu.Lock()
	defer c.mu.Unlock()
	state, err := c.loadLocked()
	if err != nil {
		return nil, err
	}
	var job map[string]any
	var jobID string
	for id, candidate := range state.Jobs {
		if candidate["owner"] == owner && candidate["batchId"] == batchID && (candidate["state"] == "running" || candidate["state"] == "paused") {
			job = candidate
			jobID = id
			break
		}
	}
	if job == nil {
		return nil, fmt.Errorf("%w: 当前没有生产队列，请先启动自动生产", batchfactoryv11.ErrConflict)
	}
	state, job = writablePriorityJob(state, jobID)
	books, _ := job["books"].(map[string]any)
	if books == nil {
		books = map[string]any{}
		job["books"] = books
	}
	entry, _ := books[selected.ID].(map[string]any)
	if entry != nil && entry["status"] != "queued" && entry["status"] != "paused" {
		return nil, fmt.Errorf("%w: 只能插队等待中的书，正在处理或已结束的书不会重复执行", batchfactoryv11.ErrConflict)
	}
	ids := automationPriorityIDs(job)
	if len(ids) > 0 && ids[0] == selected.ID && entry != nil && automationBoolValue(entry["priorityRequested"]) {
		return publicAutomationJob(job), nil
	}
	next := []string{selected.ID}
	for _, id := range ids {
		if id != selected.ID {
			next = append(next, id)
		}
	}
	if entry == nil {
		entry = map[string]any{"bookId": selected.ID, "title": selected.Title, "status": "queued", "stage": automationFollowUpStage, "attempts": 0, "retryCount": 0}
		books[selected.ID] = entry
	}
	entry["priorityRequested"], entry["message"], entry["updatedAt"] = true, "已插队，等待当前书完成后优先处理", c.now()
	appendAutomationBookLog(entry, c.now(), "priority", "已插队，正在处理的书保持执行")
	job["priorityBookIds"], job["updatedAt"] = next, c.now()
	job["counts"] = automationCountsFromBooks(books)
	if err := c.saveLocked(state); err != nil {
		return nil, err
	}
	return publicAutomationJob(job), nil
}

// Called only after a slot is available. Choosing and marking running share the
// same state lock as PrioritizeBook, so a request cannot race an old selection.
func (c *BatchAutomationCoordinator) claimAutomationPriorityBook(ctx context.Context, owner, batchID, jobID string, epoch int64, fallback batchfactoryv11.Book) (batchfactoryv11.Book, map[string]any, error) {
	unlockBatch := c.lockBatchRun(owner, batchID)
	defer unlockBatch()
	c.mu.Lock()
	defer c.mu.Unlock()
	if ctx.Err() != nil || !c.isCurrentRun(jobID, epoch) {
		return fallback, nil, batchfactoryv11.ErrConflict
	}
	state, err := c.loadLocked()
	if err != nil {
		return fallback, nil, err
	}
	job := state.Jobs[jobID]
	if job == nil {
		return fallback, nil, nil
	}
	if job["owner"] != owner || job["batchId"] != batchID {
		return fallback, nil, batchfactoryv11.ErrNotFound
	}
	if job["state"] != "running" {
		return fallback, nil, batchfactoryv11.ErrConflict
	}
	state, job = writablePriorityJob(state, jobID)
	books, _ := job["books"].(map[string]any)
	chosen := fallback
	for _, id := range automationPriorityIDs(job) {
		entry, _ := books[id].(map[string]any)
		if entry == nil || (entry["status"] != "queued" && entry["status"] != "paused") {
			continue
		}
		if id != fallback.ID {
			batch, loadErr := c.store.GetBatch(ctx, owner, batchID)
			if loadErr != nil {
				return fallback, nil, loadErr
			}
			found := false
			for _, book := range batch.Books {
				if book.ID == id {
					chosen = book
					found = true
					break
				}
			}
			if !found {
				return fallback, nil, batchfactoryv11.ErrNotFound
			}
		}
		break
	}
	entry, _ := books[chosen.ID].(map[string]any)
	if entry == nil {
		return fallback, nil, nil
	}
	if automationEntryTerminal(entry) && !automationBoolValue(entry["batchRetryPending"]) {
		return chosen, cloneAutomationEntry(entry), nil
	}
	if entry["status"] == "running" {
		// 旧一轮 run（服务重启前的进程，或已被取消的上一轮）会把书留在
		// running。若继续按“正在执行”拒绝认领，新一轮整轮 run 会在第一本书
		// 就 revision conflict 退出，job 永远悬在 running、无法重试。
		if automationNumberValue(entry[automationBookClaimEpochField], 0) == int(epoch) {
			return fallback, nil, batchfactoryv11.ErrConflict
		}
		entry["message"] = "上一轮执行已中断，本轮重新认领"
		appendAutomationBookLog(entry, c.now(), automationStringValue(entry["stage"], ""), entry["message"].(string))
	}
	next := []string{}
	for _, id := range automationPriorityIDs(job) {
		if id != chosen.ID {
			next = append(next, id)
		}
	}
	job["priorityBookIds"] = next
	delete(entry, "priorityRequested")
	markAutomationBookQueued(entry, c.now())
	markAutomationBookStarted(entry, c.now())
	entry[automationBookClaimEpochField] = int(epoch)
	job["updatedAt"] = c.now()
	job["counts"] = automationCountsFromBooks(books)
	if err := c.saveLocked(state); err != nil {
		return fallback, nil, err
	}
	return chosen, cloneAutomationEntry(entry), nil
}

// Keep the displaced book first among ordinary waiting books; remove the chosen
// priority book from its old position. A newly appended book is inserted once.
func replaceAutomationQueuedBook(queue []batchfactoryv11.Book, qi int, displaced, chosen batchfactoryv11.Book) []batchfactoryv11.Book {
	if displaced.ID == chosen.ID {
		return queue
	}
	result := append([]batchfactoryv11.Book(nil), queue[:qi]...)
	result = append(result, displaced)
	for _, book := range queue[qi:] {
		if book.ID != chosen.ID {
			result = append(result, book)
		}
	}
	return result
}
