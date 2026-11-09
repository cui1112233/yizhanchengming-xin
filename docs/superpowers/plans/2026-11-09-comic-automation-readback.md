# Comic Automation Readback Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop invalid comic publish jobs before paid generation, continue in-flight pages without manual dead ends, and show generated comic images quickly through progressive thumbnails.

**Architecture:** Keep the three concerns isolated. The N8 publisher owns read-only per-book directory validation, the automation coordinator owns retry disposition, and the React image component owns progressive display. Existing page records and successful images remain authoritative and reusable.

**Tech Stack:** Go, React, Ant Design, Node test runner, Go `testing`.

## Global Constraints

- Work directly on branch `v88`; do not create another branch.
- Do not start any paid image generation or production retry during implementation.
- Do not deploy until local tests, frontend build, embedded Go build, and local runtime checks pass.
- Reuse successful comic pages and only continue missing pages.

---

### Task 1: Read-only per-book comic publish preflight

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_n8_comic_publish.go`
- Modify: `backend/internal/httpapi/batch_factory_automation.go`
- Modify: `backend/internal/httpapi/batch_factory_automation_preflight.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/httpapi/batch_factory_comic_audio_match_test.go`
- Test: `backend/internal/httpapi/batch_factory_automation_preflight_test.go`

**Interfaces:**
- Produces: `ValidateBookSettings(context.Context, owner, batchID, bookID string, settings batchfactoryv11.SettingsPatch) error`.
- Consumes: `comicAudioCategoryName`, `enabledComicAudioCategory`, `enabledComicAudio`, and the existing cached N8 directory readers.

- [ ] **Step 1: Write failing tests**

Add tests proving that automatic matching rejects a missing mapped category, manual matching rejects an unavailable audio, and validation does not persist `comicAudioSelection`.

```go
err := publisher.ValidateBookSettings(ctx, "alice", batch.ID, book.ID, settings)
if err == nil || !strings.Contains(err.Error(), "不存在、未启用或名称不唯一") {
    t.Fatalf("error=%v", err)
}
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `go test ./internal/httpapi -run 'TestComicAudioBookPreflight|TestAutomationPreflightBlocksInvalidComicAudio' -count=1`

Expected: FAIL because `ValidateBookSettings` does not exist and automation preflight does not call it.

- [ ] **Step 3: Implement the minimal read-only validation**

Parse settings, load the current book, resolve manual or automatic category, confirm exactly one enabled category, and confirm at least one enabled audio. Do not call `randomAudioIndex` and do not save settings.

- [ ] **Step 4: Wire validation into `full_submit` per-book preflight**

Change the coordinator callback to receive owner/batch/book identity and append a blocking `comic_publish_directory_invalid` issue only for comic `full_submit` jobs.

- [ ] **Step 5: Run focused and package tests**

Run: `go test ./internal/httpapi -run 'TestComicAudio|TestAutomationPreflight' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add backend/internal/httpapi
git commit -m "fix: validate comic publish audio before automation"
```

### Task 2: Keep in-flight pages in automatic continuation

**Files:**
- Modify: `backend/internal/httpapi/automation_failure.go`
- Test: `backend/internal/httpapi/automation_failure_test.go`
- Test: `backend/internal/httpapi/batch_factory_automation_test.go`

**Interfaces:**
- Consumes: `automationCodePageInFlight`, `recordAutomationPageLedger`, and existing batch retry scheduling.
- Produces: a first in-flight conflict with `NeedsManual == false` and a retryable continuation state.

- [ ] **Step 1: Write a failing page-ledger test**

```go
outcome := recordAutomationPageLedger(entry, partialInFlightError(), "comic_images", now)
if outcome.NeedsManual {
    t.Fatal("PAGE_IN_FLIGHT must continue automatically")
}
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `go test ./internal/httpapi -run 'TestRecordAutomationPageLedgerKeepsInFlightRetryable' -count=1`

Expected: FAIL because the first in-flight error is currently marked manual.

- [ ] **Step 3: Make in-flight disposition repairable without charging retry budget**

Set the in-flight disposition to remain skipped/retryable rather than `needs_manual`; preserve its stable `PAGE_IN_FLIGHT` code. Do not change deterministic 4xx handling.

- [ ] **Step 4: Verify continuation reuses completed pages**

Run: `go test ./internal/httpapi -run 'TestRecordAutomationPageLedgerKeepsInFlightRetryable|TestGenerateComicWithRetry|TestBatchRetry' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add backend/internal/httpapi/automation_failure.go backend/internal/httpapi/automation_failure_test.go backend/internal/httpapi/batch_factory_automation_test.go
git commit -m "fix: continue comic pages that are still in flight"
```

### Task 3: Progressive comic image display

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicStatus.js`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryComicStatus.test.js`
- Test: `frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

**Interfaces:**
- Consumes: `resolveComicImageSources({url, thumbnailUrl})` and `getComicPageImageBlob`.
- Produces: thumbnail-first display with original replacement for preview mode.

- [ ] **Step 1: Write failing tests**

Assert that the main preview and candidate grid pass `thumbnailUrl`, and that preview mode starts with the thumbnail rather than the original.

```js
assert.match(panel, /imageUrl=\{primary\.url\} thumbnailUrl=\{primary\.thumbnailUrl\}/);
assert.match(panel, /selectedPage\?\.images[\s\S]*thumbnailUrl=\{image\.thumbnailUrl\}/);
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStatus.test.js src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

Expected: FAIL because the main preview and candidates omit thumbnails and preview mode prefers the original immediately.

- [ ] **Step 3: Implement thumbnail-first progressive loading**

Fetch and show the thumbnail first. In preview mode, fetch the original in the background and replace the displayed object URL only after it completes. Retain the thumbnail if original loading fails.

- [ ] **Step 4: Pass thumbnails through every comic display call site**

Update main preview and candidate grid to pass `thumbnailUrl`. Keep list and collection lazy loading unchanged.

- [ ] **Step 5: Run focused tests and frontend build**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStatus.test.js src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

Run: `npm run build`

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add frontend/src/user/pages/shuihuo
git commit -m "fix: show comic thumbnails before original images"
```

### Task 4: Integrated local verification

**Files:**
- Verify: `backend/qiantie-local.exe`
- Verify: embedded frontend assets produced by the normal build scripts.

**Interfaces:**
- Consumes: the three preceding commits.
- Produces: local browser acceptance evidence without starting paid jobs.

- [ ] **Step 1: Run complete focused backend tests**

Run: `go test ./internal/httpapi ./internal/batchfactoryv11 -count=1`

Expected: PASS.

- [ ] **Step 2: Build the embedded local binary**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Expected: exit code 0 and a rebuilt `backend/qiantie-local.exe`.

- [ ] **Step 3: Restart only the confirmed local review runtime**

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`

Expected: `localhost:4001` responds and the process path is the rebuilt local binary.

- [ ] **Step 4: Verify local runtime identity and UI**

Check `/health`, `/api/runtime-build-info`, open the existing batch read-only, and confirm thumbnail-first display. Do not click generation or retry.

- [ ] **Step 5: Commit any verification-only source corrections**

Only if local verification exposes a source defect, add the smallest tested correction and commit it separately. Otherwise make no additional commit.
