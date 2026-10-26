# Novel Fetch REA Control Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every visible Novel Fetch control real, remove or disable fake controls, and expose the same actionable states and configuration behavior verified in the REA legacy workflow.

**Architecture:** Keep the existing React Novel Fetch workbench and Go API boundaries. Add narrow source-contract tests for the React controls, Go tests for persisted configuration/status projection, then make the smallest changes in the existing page/API files. No legacy UI copy or new runtime layer.

**Tech Stack:** React, Ant Design, Node test runner, Go, MySQL-backed Novel Fetch store, Go embedded frontend.

## Global Constraints

- Maintain `v88` as the only maintained source branch.
- Preserve unrelated dirty worktree changes; stage only files belonging to this plan.
- Keep all system prompts and execution decisions in Go; React edits configuration and displays results.
- User-visible `121` text remains `视频管理系统`.
- Do not execute real novel fetching, paid AI generation, or external video-management uploads during tests.

---

### Task 1: Build the control-to-route contract audit

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Create: `frontend/src/user/pages/novel-fetch-control-contract.test.js`
- Inspect: `frontend/src/shared/api/novelFetchWorkshop.js`, `backend/internal/httpapi/novel_fetch_workshop.go`

**Interfaces:**
- Consumes the existing handler names imported by `NovelFetchPage.jsx`.
- Produces a source-contract test that fails when a visible action has no handler/API call or when a known fake action remains.

- [ ] **Step 1: Write the failing contract tests.**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const page = fs.readFileSync(new URL('./NovelFetchPage.jsx', import.meta.url), 'utf8');
const api = fs.readFileSync(new URL('../../shared/api/novelFetchWorkshop.js', import.meta.url), 'utf8');

test('visible workflow actions have Go-backed handlers', () => {
  for (const name of ['fetchWorkshopOriginal', 'startWorkshopProcess', 'previewWorkshopProcess', 'retryWorkshopTasks', 'stopWorkshopTasks', 'generateWorkshopAi', 'applyWorkshopRules', 'previewWebSubmit', 'startWebSubmit']) {
    assert.match(page, new RegExp(name));
    assert.match(api, new RegExp(name));
  }
});

test('legacy-only placeholder actions are not presented as active controls', () => {
  assert.doesNotMatch(page, /onClick=\{\(\) => \{\s*message\.(info|success)\([^}]+\)\s*\}\}/);
});
```

- [ ] **Step 2: Run the contract test and record the first failing control.**

Run: `node --test src/user/pages/novel-fetch-control-contract.test.js`

Expected: FAIL if any visible action is only a message or missing from the API module.

- [ ] **Step 3: Inspect each failure and either wire it to the existing Go API or remove/disable it.**

For a real action, use the existing pattern `await handleWebSubmit(existingApiCall, successMessage)` or `await existingApiCall(...); await loadTasks()`. For an unsupported action, replace the active button with a disabled control carrying a reason in its tooltip, or remove it from the page.

- [ ] **Step 4: Run the contract test and the existing Novel Fetch source tests.**

Run: `node --test src/user/pages/novel-fetch-control-contract.test.js src/user/pages/novel-fetch-status.test.js src/user/pages/novel-fetch-submit.test.js`

Expected: PASS.

- [ ] **Step 5: Commit the contract cleanup.**

```bash
git add frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-control-contract.test.js
git commit -m "fix: remove novel fetch fake controls"
```

### Task 2: Expose and persist REA sensitive-AI controls

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Modify: `backend/internal/novelfetchworkshop/seed/app.json`
- Test: `backend/internal/httpapi/novel_fetch_workshop_config_test.go`
- Test: `backend/internal/httpapi/novel_fetch_rewrite_test.go`

**Interfaces:**
- `sensitive_ai` config uses `group_hits`, `concurrency`, `max_hits_per_task`, `context_chars`, `retries`, `temperature`, and `prompt`.
- Existing `applySensitiveAIRepair` remains the single execution entry point.

- [ ] **Step 1: Add failing config round-trip tests.**

Assert that saving a `sensitive_ai` patch preserves all seven REA fields and that a subsequent GET returns the same numeric/boolean values.

- [ ] **Step 2: Add failing UI source assertions for the controls.**

Assert the page contains labels and form bindings for send mode, concurrency, maximum hits, context length, retries, temperature, and prompt, and that the values are inside the `sensitive_ai` form namespace.

- [ ] **Step 3: Implement the controls using the existing Ant Design Form.**

Use `InputNumber` bounds `concurrency: 1..32`, `max_hits_per_task: 1..1000`, `context_chars: 0..200`, `retries: 0..3`, `temperature: 0..2`. Keep the mode select mapped to `group_hits` rather than introducing a second incompatible field.

- [ ] **Step 4: Verify the Go execution tests for group, limit, unsafe output, and retry behavior.**

Run: `go test ./internal/httpapi -run 'TestApplySensitiveAIRepair|TestNovelFetchWorkshopConfig' -count=1`

Expected: PASS with no partial result persisted.

- [ ] **Step 5: Commit the sensitive-AI parity changes.**

```bash
git add frontend/src/user/pages/NovelFetchPage.jsx backend/internal/httpapi/novel_fetch_workshop.go backend/internal/novelfetchworkshop/seed/app.json backend/internal/httpapi/novel_fetch_workshop_config_test.go backend/internal/httpapi/novel_fetch_rewrite_test.go
git commit -m "feat: expose rea sensitive ai controls"
```

### Task 3: Restore detailed task and submission states

**Files:**
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Modify: `backend/internal/novelfetchworkshop/status.go`
- Modify: `frontend/src/user/pages/novel-fetch-status.js`
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Test: `backend/internal/novelfetchworkshop/status_test.go`
- Test: `frontend/src/user/pages/novel-fetch-status.test.js`

**Interfaces:**
- Task projection exposes `site_submit_pending_versions`, `site_submit_waiting_generate_versions`, `site_submit_ready_versions`, `site_submit_expected_versions`, and existing failure/partial fields when present.
- `novelFetchTaskStatus` returns separate `waiting_retry`, `waiting_configuration`, `waiting_login`, `pending_submit`, and `failed` statuses.

- [ ] **Step 1: Add failing status and projection tests.**

Cover waiting fetch retry, waiting sensitive AI, waiting configuration/login, pending AI generation, pending submission, partial submission, failed version, and submitted version. Assert the UI summary counts waiting retry separately from active processing.

- [ ] **Step 2: Implement backend projection and status precedence.**

Copy existing metadata fields without renaming them, and ensure failure precedes waiting, waiting precedes processing, and pending submission precedes completed.

- [ ] **Step 3: Implement compact visible status details.**

Add a status detail column or expandable detail using the existing task table/detail components. Show the exact versions and blocking reason; do not replace the existing table with a new page.

- [ ] **Step 4: Run focused tests.**

Run: `go test ./internal/novelfetchworkshop ./internal/httpapi` and `node --test src/user/pages/novel-fetch-status.test.js`.

Expected: PASS.

- [ ] **Step 5: Commit the state projection changes.**

```bash
git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/novelfetchworkshop/status.go frontend/src/user/pages/novel-fetch-status.js frontend/src/user/pages/NovelFetchPage.jsx backend/internal/novelfetchworkshop/status_test.go frontend/src/user/pages/novel-fetch-status.test.js
git commit -m "fix: show novel fetch blocking states"
```

### Task 4: Verify configuration persistence and cross-page handoff

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Inspect: `frontend/src/user/pages/novel-fetch-batch-factory.js`, `frontend/src/shared/api/batchFactoryV11.js`, `frontend/src/shared/api/novelFetchWorkshop.js`; modify only the file identified by the failing handoff test.
- Test: `frontend/src/user/pages/novel-fetch-batch-factory.test.js`
- Test: `frontend/src/shared/api/novelFetchWorkshop.test.js`

**Interfaces:**
- Handoff payload carries `bookId`, selected source version, and selected AI versions.
- Configuration save/read uses the same field names and account scope.

- [ ] **Step 1: Add failing handoff and config-readback tests.**

Assert the payload uses edited/selected source text and selected version, and that saved input format, max text length, workflow switches, queue settings, upload profiles, and submit versions survive a fresh GET.

- [ ] **Step 2: Fix mismatched field names or stale local-only state.**

After save, update React state from the server response and reload the task/config projection. Do not use fake dates or browser-only defaults to make a task appear complete.

- [ ] **Step 3: Run focused frontend tests.**

Run: `node --test src/user/pages/novel-fetch-batch-factory.test.js src/shared/api/novelFetchWorkshop.test.js`.

Expected: PASS.

- [ ] **Step 4: Build and verify the runtime.**

Run: `npm run build`, `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`, then restart the local Go runtime and run `powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1`.

Expected: all health routes return 200 where applicable and `/api/runtime-build-info` reports the new Git SHA.

- [ ] **Step 5: Commit the final handoff/build verification.**

```bash
git add frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-batch-factory.js frontend/src/shared/api/batchFactoryV11.js frontend/src/shared/api/novelFetchWorkshop.js frontend/src/user/pages/novel-fetch-batch-factory.test.js frontend/src/shared/api/novelFetchWorkshop.test.js
git commit -m "fix: verify novel fetch handoff and config readback"
```
