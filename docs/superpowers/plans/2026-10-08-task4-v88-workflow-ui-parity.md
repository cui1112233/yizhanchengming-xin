# Task 4 V88 Workflow UI Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the four V88 workflow pages to the new React UI while keeping Go/MySQL/TOS as the only business fact source.

**Architecture:** Keep `UserShell`, `RouterApp.jsx`, and shared `api.js` untouched. Page-local components consume existing Go API adapters, keep unsaved form text only in component state, and reload persisted facts after every mutation. Existing Go prompt, stage, intake, storyboard and novel-panel services remain the sole execution and persistence boundary.

**Tech Stack:** Go `net/http`, MySQL/Goose, React, Ant Design, Vitest, Testing Library, Playwright.

## Global Constraints

- Do not add Node API calls, iframe, postMessage business bridge, Bearer token, `/api/chat`, localStorage, or sessionStorage.
- Do not alter `前台/src/UserShell.jsx`, `前台/src/RouterApp.jsx`, or `前台/src/api.js`.
- Use existing prompt catalog metadata; no large system prompt text in React.
- A missing Provider/TOS/product contract is documented in `docs/parity/task4-missing-go-contracts.md`; it never produces a simulated success.
- All write paths must use existing same-origin protected Go endpoints and preserve 401/403/409/422 semantics.

---

### Task 1: Establish task-four parity ledger and automated prohibition checks

**Files:**
- Create: `docs/parity/task4-missing-go-contracts.md`
- Modify: `前台/src/ScriptWorkspace.test.jsx`
- Modify: `前台/src/NovelFetchPage.test.jsx`
- Modify: `前台/src/NovelFetchWorkshop.test.jsx`
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.test.jsx`

**Interfaces:**
- Consumes: existing page exports and Go API adapters from `前台/src/api.js`.
- Produces: regression tests proving the pages do not rely on browser business storage or embedding.

- [ ] **Step 1: Write failing prohibition tests**

```js
expect(document.querySelector('iframe')).toBeNull()
expect(window.sessionStorage.getItem('workshopUploadItems')).toBeNull()
```

- [ ] **Step 2: Run the four page tests to verify they fail only where a prohibited dependency exists**

Run: `CI=1 npm test -- --run src/ScriptWorkspace.test.jsx src/NovelFetchPage.test.jsx src/NovelFetchWorkshop.test.jsx src/novel-panel/NovelPanelWorkbench.test.jsx`

- [ ] **Step 3: Remove the dependency or record a server contract gap**

Create a section per page with endpoint, persisted model, permission behavior, and any unsatisfied real contract. No test fixture is a production fallback.

- [ ] **Step 4: Re-run the four page tests**

Run: `CI=1 npm test -- --run src/ScriptWorkspace.test.jsx src/NovelFetchPage.test.jsx src/NovelFetchWorkshop.test.jsx src/novel-panel/NovelPanelWorkbench.test.jsx`

### Task 2: Recreate Script workspace layout on persisted Go facts

**Files:**
- Modify: `前台/src/ScriptWorkspace.jsx`
- Modify: `前台/src/script-workspace.css`
- Modify: `前台/src/ScriptWorkspace.test.jsx`
- Modify: `api/internal/httpapi/script_storyboard_handlers_test.go`

**Interfaces:**
- Consumes: `getScriptWorkspace`, `saveScriptOriginalText`, `saveProductionSettings`, stage APIs, and storyboard APIs.
- Produces: persisted original text, characters, scenes, constraints, duration, stage history, cards, and downstream recompile behavior.

- [ ] **Step 1: Write failing UI tests for server reload, 403, conflict preservation, and retry**

```js
expect(api.getScriptWorkspace).toHaveBeenCalledWith(3)
expect(screen.getByRole('textbox', { name: '小说原文' }).value).toBe('第一章 原文')
```

- [ ] **Step 2: Verify failure before the layout change**

Run: `CI=1 npm test -- --run src/ScriptWorkspace.test.jsx`

- [ ] **Step 3: Port visual sections from legacy ScriptPage without legacy request/storage code**

Keep old-card visual hierarchy, drawers, labels and responsive behavior; wire actions to persisted Go functions only.

- [ ] **Step 4: Add/adjust handler tests for failed and recompiled storyboard requests**

```go
if got := response.Code; got != http.StatusForbidden { t.Fatalf("got %d", got) }
```

- [ ] **Step 5: Run frontend and Go focused tests**

Run: `CI=1 npm test -- --run src/ScriptWorkspace.test.jsx && go test ./internal/httpapi ./internal/generation`

### Task 3: Recreate Novel Fetch and Workshop without browser handoff

**Files:**
- Modify: `前台/src/NovelFetchPage.jsx`
- Modify: `前台/src/NovelFetchWorkshop.jsx`
- Modify: `前台/src/NovelFetchPage.test.jsx`
- Modify: `前台/src/NovelFetchWorkshop.test.jsx`
- Modify: `api/internal/httpapi/workshop_handlers_test.go`

**Interfaces:**
- Consumes: intakes/books/execute/workshop/restore endpoints and prompt catalog metadata.
- Produces: source/platform/batch IDs, progress, safe errors, rule/prompt state, original text recovery, and server-only task status.

- [ ] **Step 1: Write failing tests for a persisted workshop reopen and 403/failed retry**

```js
expect(await screen.findByText('服务端原文')).toBeTruthy()
expect(await screen.findByText('无权限查看小说获取任务。')).toBeTruthy()
```

- [ ] **Step 2: Verify failure on old layout behavior**

Run: `CI=1 npm test -- --run src/NovelFetchPage.test.jsx src/NovelFetchWorkshop.test.jsx`

- [ ] **Step 3: Port legacy visual layout and remove every upload handoff storage path**

Use server task status, route query parameters, and persisted workshop settings. A missing upload/TOS submission contract goes in the parity ledger and has no success control.

- [ ] **Step 4: Verify Go same-origin and persistence behavior**

Run: `go test ./internal/httpapi ./internal/workshop ./internal/intake`

- [ ] **Step 5: Re-run focused UI tests**

Run: `CI=1 npm test -- --run src/NovelFetchPage.test.jsx src/NovelFetchWorkshop.test.jsx`

### Task 4: Complete React Novel Panel parity and server-history recovery

**Files:**
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.jsx`
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.css`
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.test.jsx`
- Modify: `api/internal/novelpanel/service_test.go`
- Modify: `api/internal/httpapi/novel_panel_handlers_test.go`

**Interfaces:**
- Consumes: `getNovelPanel`, `saveNovelPanel`, `listNovelPanelHistory`, `restoreNovelPanelHistory` and the Go novel-panel MySQL store.
- Produces: a direct React workbench for text, characters, relationships, storyboard, style, and auditable history restoration.

- [ ] **Step 1: Write failing tests for direct React rendering, service reload, 403, and history restore**

```js
expect(document.querySelector('iframe')).toBeNull()
expect(await screen.findByText(/修订 1/)).toBeTruthy()
```

- [ ] **Step 2: Verify failure before changing components**

Run: `CI=1 npm test -- --run src/novel-panel/NovelPanelWorkbench.test.jsx`

- [ ] **Step 3: Apply legacy workbench visual structure and only direct Go-backed controls**

Keep old cards, drawers, responsive CSS and copy. Preserve current revision conflict behavior and never emit a provider-generated result locally.

- [ ] **Step 4: Run backend module and handler tests**

Run: `go test ./internal/novelpanel ./internal/httpapi`

- [ ] **Step 5: Re-run panel UI tests**

Run: `CI=1 npm test -- --run src/novel-panel/NovelPanelWorkbench.test.jsx`

### Task 5: Integrated verification and LAN acceptance evidence

**Files:**
- Create: `docs/parity/task4-acceptance-evidence.md`
- Modify: `前台/e2e/novel-fetch.spec.js`

**Interfaces:**
- Consumes: fully wired local Go server and `10.0.11.112:18080` deployment.
- Produces: reproducible test commands and image paths for reference/public and LAN screenshots.

- [ ] **Step 1: Add an E2E assertion for persisted intake reopening and retry**

```js
await expect(page.getByRole('button', { name: '重试失败项' })).toBeVisible()
```

- [ ] **Step 2: Run complete Go verification**

Run: `go test ./...`

- [ ] **Step 3: Run complete UI verification and build**

Run: `CI=1 npm test -- --run && npm run build`

- [ ] **Step 4: Capture authenticated LAN screenshots after create/save/reload/re-login/403/retry checks**

Run: Playwright against `http://10.0.11.112:18080`; retain one screenshot per page and label actual pass/fail evidence. Do not deploy ECS.

- [ ] **Step 5: Commit and publish main**

Run: `git add docs/parity 前台 api && git commit -m "feat(parity): migrate task4 V88 workflows [skip ci]" && git push origin main && git rev-parse HEAD && git rev-parse origin/main`
