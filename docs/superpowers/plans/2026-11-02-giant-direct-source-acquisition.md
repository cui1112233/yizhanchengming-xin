# Giant Direct Source Acquisition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a Giant-material book that fails under `direct_first` stop safely, with no hidden 121 retry, OCR job, or production run until a user explicitly chooses a new action.

**Architecture:** Add one backend source-state gate which recognizes the persisted `direct_first` failure state before every automatic production fetch. Keep legacy empty-text recovery unchanged for all other books. Make the creation flow and list summary read the same persisted stage fields so a failed direct read is shown as a failure, not as a background read.

**Tech Stack:** Go, Go `net/http` tests, React, Node built-in test runner, MySQL-backed batch store, embedded Vite frontend.

## Global Constraints

- Work in `F:\脚本测试\一战晟铭-v88`; do not create a branch, reset, clean, revert unrelated changes, commit, push, or deploy.
- The backend remains Go in `backend`; the frontend remains React + Ant Design in `frontend`.
- Do not make paid model, OCR, video, or publishing calls while verifying this change.
- Do not put upstream tokens, cookies, URLs, or authorization headers into diagnostics or UI state.
- `direct_first` failure must be terminal for automatic paths; OCR may only be created by an explicit user action while an executor is online.

---

### Task 1: Add a backend direct-first terminal-state gate

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_automation.go:1222-1312`
- Modify: `backend/internal/httpapi/batch_factory_automation_test.go`

**Interfaces:**
- Produces: `isDirectFirstSourceBlocked(book batchfactoryv11.Book) bool`.
- Consumes: `book.SourceMetadata["originalReadStrategy"]`, `book.SourceMetadata["originalReadStage"]`, and `book.SourceMetadata["contentPending"]`.
- Guarantees: `ensureAutomationSource` returns a readable error before it invokes `fetchBatchSourceWithRetry` for a blocked book.

- [ ] **Step 1: Write the failing Go test**

Add a test that constructs a `BatchAutomationCoordinator` with an HTTP test upstream which increments a request counter. Give it a book with empty source text and metadata:

```go
map[string]any{
    "sourceReadStrategy": "direct_first",
    "originalReadStage": "failed",
    "originalReadError": "121 upstream business error 400: 参数不完整",
    "contentPending": false,
}
```

Call `ensureAutomationSource` and assert the error contains `书城原文获取失败，等待人工处理`, and assert the upstream request counter remains `0`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/httpapi -run TestEnsureAutomationSourceDirectFirstFailureDoesNotRefetch -count=1`

Expected: FAIL because the current implementation calls the upstream fetch fallback.

- [ ] **Step 3: Add the minimal gate**

Place `isDirectFirstSourceBlocked` above `ensureAutomationSource`. It returns true only when strategy is `direct_first` and stage is `failed` or `source_blocked`. At the beginning of `ensureAutomationSource`, after the existing stored-raw reprocess case and before the empty-text fetch fallback, return:

```go
return book, fmt.Errorf("书城原文获取失败，等待人工处理：%s", reasonOrDefault(strings.TrimSpace(stringValue(book.SourceMetadata["originalReadError"])), "请手动重试直取或确认后使用 OCR"))
```

- [ ] **Step 4: Run the focused test to verify it passes**

Run: `go test ./internal/httpapi -run TestEnsureAutomationSourceDirectFirstFailureDoesNotRefetch -count=1`

Expected: PASS and no upstream request.

- [ ] **Step 5: Add a regression test for legacy recovery**

Add a second test for an empty legacy book with no `direct_first` strategy. Its upstream request counter must be `1`, proving the gate does not break legacy automatic recovery.

- [ ] **Step 6: Run both source-gate tests**

Run: `go test ./internal/httpapi -run 'TestEnsureAutomationSource(DirectFirstFailureDoesNotRefetch|LegacyEmptyBookStillFetches)' -count=1`

Expected: PASS.

### Task 2: Refuse automatic production for blocked books

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_automation.go:710-775`
- Modify: `backend/internal/httpapi/batch_factory_automation_test.go`

**Interfaces:**
- Consumes: `isDirectFirstSourceBlocked(book)` from Task 1.
- Produces: a preflight error containing `书城原文获取失败，等待人工处理` before a production job is persisted or launched.
- Guarantees: scheduled, retry, and refresh routes which enter `Start` cannot turn `direct_first` failure into production work.

- [ ] **Step 1: Write the failing Go test**

Create a batch with configured model settings and one book whose metadata matches Task 1. Call the automation-start HTTP route with a valid preset. Assert a client error response, assert its body includes `书城原文获取失败，等待人工处理`, and assert the automation status remains idle.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/httpapi -run TestBatchAutomationStartRejectsDirectFirstBlockedBook -count=1`

Expected: FAIL because preflight currently accepts the book and creates an automation job.

- [ ] **Step 3: Add the preflight rejection**

At the start of each `for _, book := range books` iteration in `validateAutomationPreflight`, append one reason when `isDirectFirstSourceBlocked(book)` is true:

```go
reasons = append(reasons, fmt.Sprintf("《%s》书城原文获取失败，等待人工处理", strings.TrimSpace(book.Title)))
```

Use `未命名作品` when the title is empty. Preserve the existing de-duplication of reasons.

- [ ] **Step 4: Run the focused preflight test to verify it passes**

Run: `go test ./internal/httpapi -run TestBatchAutomationStartRejectsDirectFirstBlockedBook -count=1`

Expected: PASS with no persisted running/scheduled automation.

- [ ] **Step 5: Run the affected backend package**

Run: `go test ./internal/httpapi -count=1`

Expected: PASS.

### Task 3: Do not auto-start Giant automation after direct failures

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.jsx:405-514`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialFlow.source.test.js`

**Interfaces:**
- Consumes: the direct-read result bookkeeping (`originalReadStage`, `originalReadError`, `contentPending`).
- Produces: a `hasDirectReadFailure` accumulator scoped to the Giant direct-read loop.
- Guarantees: `startBatchAutomation(batchId, giantAutomationPlan)` is called only when every planned direct read succeeded; failed books retain `contentPending: false`.

- [ ] **Step 1: Write the failing frontend source test**

Add assertions that the component declares `hasDirectReadFailure`, sets it in the direct-read catch block, and guards automatic start with `giantAutomationPlan?.presetId && !hasDirectReadFailure`. Assert the failed metadata assigns `contentPending: false`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialFlow.source.test.js`

Expected: FAIL because the existing component starts automation whenever a preset exists and leaves the failure pending.

- [ ] **Step 3: Implement the smallest frontend guard**

Initialize `let hasDirectReadFailure = false` before direct reads. Set it to `true` inside the failure branch, write `contentPending: false`, and add `&& !hasDirectReadFailure` to the start condition. When there are failures, show one warning that direct reads failed and production was not started.

- [ ] **Step 4: Run the focused test to verify it passes**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialFlow.source.test.js`

Expected: PASS.

### Task 4: Make the list summary count only active reads

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx:716-720,3622`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryDirectFirstStatus.source.test.js`

**Interfaces:**
- Produces: `isBatchFactoryContentReadRunning(book)` or an equivalent local predicate.
- Consumes: `sourceMetadata.originalReadStage` and `sourceMetadata.contentPending`.
- Guarantees: only `direct`, `ocr`, or an explicit running OCR state appears in the “正在后台读取” summary; `failed` and `source_blocked` never do.

- [ ] **Step 1: Write the failing frontend source test**

Add assertions that the summary predicate contains the active direct/OCR stages, excludes `failed`, and that the summary copy keeps `正在后台读取` only for this active predicate.

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryDirectFirstStatus.source.test.js`

Expected: FAIL because the summary currently filters only `contentPending`.

- [ ] **Step 3: Implement the state-specific summary**

Add a small pure predicate near the existing source status helpers. Use it for both `pendingContentBooks` and `giantWaitingBooks`. Continue to show the existing per-book `书城读取失败` state, without treating it as active work.

- [ ] **Step 4: Run the focused test to verify it passes**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryDirectFirstStatus.source.test.js`

Expected: PASS.

### Task 5: Verify the integrated local runtime without external work

**Files:**
- No production source changes expected.

**Interfaces:**
- Consumes: Tasks 1-4.
- Produces: a rebuilt embedded frontend and a restarted verified local `qiantie-local.exe` at port `4001`.

- [ ] **Step 1: Run the affected frontend source tests**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryDirectFirstStatus.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialFlow.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialPendingProgress.source.test.js`

Expected: PASS.

- [ ] **Step 2: Build the frontend and Go embedded binary**

Run the repository's established frontend build and local Go embedded-build commands, preserving unrelated dirty files.

Expected: both builds succeed.

- [ ] **Step 3: Restart only the confirmed local review runtime**

Use the established `scripts/start-local-go-review.ps1 -UseYcmMysql` workflow after confirming the running process is the local `qiantie-local.exe`.

Expected: `http://localhost:4001/health` and `/api/runtime-build-info` return successfully.

- [ ] **Step 4: Inspect the existing Giant batch without invoking production or OCR**

Read `batch_d7bffda3c94ec2813ef7cbb556d8c032` through the authenticated/local API or current browser. Confirm failed direct books remain failed, are not in an OCR queue, and do not appear as “正在后台读取”.

- [ ] **Step 5: Record verification results**

Report changed paths, focused test/build output, runtime identity, and the visible batch state. Do not claim a live upstream success because this plan intentionally performs no paid/production/OCR action.
