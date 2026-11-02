# 小说获取与批量工厂实用性修复 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复小说获取与批量工厂中已确认的并发、后台返回、分类门槛、流程入口和批次刷新问题。

**Architecture:** 保留现有 Go 队列工作器和 MySQL/内存 Store，不改变上游接口。通过测试先锁定工作器并发与异步返回契约，再统一分类状态文案和前端刷新行为；所有阶段仍独立保存并可从失败阶段继续。

**Tech Stack:** Go、React、Ant Design、Node test runner、现有 Novel Fetch 与 Batch Factory API。

## Global Constraints

- 保留原文、处理正文、分类、改写、提交的独立阶段状态。
- 不改变上游原文接口协议。
- 不在测试或本地验证中触发真实抓取、付费 AI 调用或外部提交。
- 不清理工作区中与本任务无关的已有修改。
- 继续在当前 `v88` 主分支工作，不新建分支。

---

### Task 1: 修复 Novel Fetch 并发与后台启动契约

**Files:**
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Test: `backend/internal/httpapi/novel_fetch_workshop_test.go`

**Interfaces:**
- Consumes: `novelFetchConfiguredConcurrency`, `workshopProcessWorker.process`, `process/start` route and `novelfetchworkshop.ProcessQueue`.
- Produces: configured `fetch.concurrency` limits concurrent per-book execution; `POST /api/novel-fetch-workshop/process/start` returns `202` before any upstream response is needed.

- [ ] **Step 1: Lock the concurrency contract with a focused test**

  Keep the existing `TestWorkshopProcessWorkerHonorsConfiguredFetchConcurrency` and add an assertion that `novelFetchConfiguredConcurrency` returns the stored value for an integer config. Run:

  ```powershell
  go test ./internal/httpapi -run 'TestWorkshopProcessWorkerHonorsConfiguredFetchConcurrency|TestNovelFetchWorkshopProcessStartReturnsBeforeBackgroundFetchCompletes' -count=1
  ```

  Expected before the fix: the concurrency test observes only one active request and the start test times out waiting for the response.

- [ ] **Step 2: Trace and fix the worker boundary**

  Ensure the worker starts one goroutine per active book, acquires the configured semaphore before `worker.execute`, and never performs `worker.execute` from the HTTP handler. Keep queue insertion and document state writes in the handler, and keep fetch/classification/rewrite in the worker. Do not add a second queue or change `ProcessQueue` signatures.

- [ ] **Step 3: Run the focused tests**

  ```powershell
  go test ./internal/httpapi -run 'TestWorkshopProcessWorkerHonorsConfiguredFetchConcurrency|TestNovelFetchWorkshopProcessStartReturnsBeforeBackgroundFetchCompletes|TestWorkshopProcessWorkerIsolatesFailedBookAndDefersOnlyThatBook' -count=1
  ```

  Expected: all pass; the blocking upstream test returns `202` while the fake upstream remains blocked.

- [ ] **Step 4: Commit the independently testable backend fix**

  ```powershell
  git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/httpapi/novel_fetch_workshop_test.go
  git commit -m "fix: keep novel fetch queued and honor fetch concurrency"
  ```

### Task 2: Make classification state truthful across Novel Fetch and Batch Factory

**Files:**
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Modify: `backend/internal/httpapi/batch_factory_v12_single_fetch.go`
- Modify: `frontend/src/user/pages/BatchFactoryWorkbenchPage.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `backend/internal/httpapi/novel_fetch_workshop_test.go`
- Test: existing Batch Factory classification tests and focused frontend source tests

**Interfaces:**
- Consumes: `novelFetchClassificationResolved`, `finalizeFetchedBatchClassification`, batch book `sourceMetadata`.
- Produces: a shared user-facing state meaning “原文已抓取，等待分类确认”; automatic rewrite remains blocked until classification is confirmed; retry resumes from classification without fetching again.

- [ ] **Step 1: Add a regression test for the unresolved classification state**

  Create a document with `originalStatus=done`, `classifyStatus=pending_confirmation`, and saved `OriginalRaw`; assert that the worker marks it `waiting_classification`, does not call the upstream fetcher, and returns the classification-pending error.

- [ ] **Step 2: Normalize Batch Factory metadata and messages**

  Keep `classifyStatus` values (`classified`, `failed`, `pending_confirmation`) as the API contract. Change only the displayed copy and action availability so a failed or low-confidence classification reads as waiting for manual confirmation rather than “production is unaffected”. Preserve raw and processed text and never reset `originalStatus` during classification retry.

- [ ] **Step 3: Add focused UI source assertions**

  Assert the Batch Factory source contains the waiting-classification wording and does not claim that classification failure is harmless when automatic rewrite is blocked. Assert the single-book retry action points to classification retry when raw source metadata is already complete.

- [ ] **Step 4: Run classification tests**

  ```powershell
  go test ./internal/httpapi -run 'Classification|classif|NovelFetch' -count=1
  node --test frontend/src/user/pages/novel-fetch-account-scope.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.classification.source.test.js
  ```

- [ ] **Step 5: Commit the state-contract fix**

  ```powershell
  git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/httpapi/batch_factory_v12_single_fetch.go frontend/src/user/pages/BatchFactoryWorkbenchPage.jsx frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx backend/internal/httpapi/novel_fetch_workshop_test.go
  git commit -m "fix: align novel classification gate and batch status"
  ```

### Task 3: Remove the duplicate process preview entry and preserve existing task visibility

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`
- Test: `frontend/src/user/pages/novel-fetch-input-format.test.js`

**Interfaces:**
- Consumes: `normalizeNovelFetchInput`, existing process form submit and preview data state.
- Produces: one clear input-to-queue path; format validation remains available without a second “解析预览” action.

- [ ] **Step 1: Update the failing layout assertion**

  Keep input normalization and error-row handling tests. Change the layout test to assert that the process page has no duplicate “解析预览” action while the input format/recognition panel remains present.

- [ ] **Step 2: Remove only the duplicate action block**

  Delete the button and status block around the `解析预览` heading at `NovelFetchPage.jsx:1772`; retain the input format selector, normalized-row summary, and submit validation. Do not remove the backend preview endpoint because other callers still use it.

- [ ] **Step 3: Run focused frontend tests**

  ```powershell
  node --test src/user/pages/novel-fetch-input-format.test.js src/user/pages/novel-fetch-subpages.layout.test.js src/user/pages/novel-fetch-progress.test.js
  ```

- [ ] **Step 4: Commit the flow simplification**

  ```powershell
  git add frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-subpages.layout.test.js frontend/src/user/pages/novel-fetch-input-format.test.js
  git commit -m "fix: simplify novel fetch input flow"
  ```

### Task 4: Add low-frequency Batch Factory home refresh and final verification

**Files:**
- Modify: `frontend/src/user/pages/BatchFactoryWorkbenchPage.jsx`
- Modify: `frontend/src/user/pages/BatchFactoryWorkbenchPage.css` only if the refresh state needs a shared visual treatment
- Test: `frontend/src/user/pages/batch-factory-workbench-home.source.test.js`

**Interfaces:**
- Consumes: `listBatchSummaries`, `refreshBatches`, `document.hidden`.
- Produces: visible home page refreshes batch summaries every 15 seconds; hidden page pauses; errors keep the last successful list and expose retry feedback.

- [ ] **Step 1: Add the refresh behavior test**

  Assert the page registers a visibility-aware 15-second refresh and that `refreshBatches` does not replace the current list with an empty array on a failed request.

- [ ] **Step 2: Implement guarded polling**

  Add one `useEffect` with a timer and `visibilitychange` listener. Reuse `refreshBatches({ quiet: true })`; on failure set the error but do not call `setBatches([])`. Clear the timer and listener on unmount. Keep the existing manual refresh button.

- [ ] **Step 3: Run frontend tests and build**

  ```powershell
  node --test src/user/pages/batch-factory-workbench-home.source.test.js src/user/pages/novel-fetch-progress.test.js
  npm.cmd run build
  ```

- [ ] **Step 4: Run focused Go tests and diff checks**

  ```powershell
  go test ./internal/httpapi ./internal/novelfetchworkshop ./internal/batchfactoryv11 -count=1
  git diff --check
  ```

  Record unrelated pre-existing failures separately; do not alter unrelated files to make the full suite green.

- [ ] **Step 5: Commit the refresh and verification changes**

  ```powershell
  git add frontend/src/user/pages/BatchFactoryWorkbenchPage.jsx frontend/src/user/pages/batch-factory-workbench-home.source.test.js
  git commit -m "fix: refresh batch factory home state"
  ```
