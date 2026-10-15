# React 页面迁移 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将小说获取、小说面板从旧 `public` iframe 迁移为 React 页面，并让页面只调用 Go 正式接口。

**Architecture:** 先切换现有 React 页面中的小说获取实现，使用已经存在的 Go workshop API；再梳理小说面板 iframe 的桥接接口，逐步把工作台功能拆到 React。迁移期间旧静态资源只保留作回退，不再作为默认页面入口。

**Tech Stack:** React、Ant Design、Go `net/http`、Go `embed`、Node test runner、Go test。

**Spec:** `docs/superpowers/specs/2026-10-15-react-page-migration-design.md`

## Global Constraints

- 维护分支使用 `v88`，不创建新分支，不推送远端。
- Go 是唯一业务接口实现；React 不硬编码提示词。
- 不删除用户现有未提交改动。
- 旧 `public` 工作台在通过验收前不得删除。
- 不把公网 `/api/tasks` 直接当成本地正式接口；以本地 Go `/api/novel-fetch-workshop/*`、`/api/novel-panel/*`、`/api/batch-factory/v12/*` 为准。

## Review Focus

- 小说获取切换页面后配置、知识库和处理规则是否仍保留。
- 原文抓取失败时是否显示具体错误而不是空白页面。
- 进入批量工厂时是否只传当前选中的版本和内容范围。
- 小说面板 `@` 资产选择是否保留缩略图、资产类型和完整删除行为。
- Go embed 构建后是否还偷偷依赖 `public/batch-rewrite` 或 `public/novel-panel`。

---

### Task 1: 小说获取 React 正式入口（已完成）

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Modify: `frontend/src/user/App.performance.test.js`
- Modify: `test/novel-fetch-performance.test.js`
- Test: existing frontend and novel-fetch source tests

**Interfaces:**
- Consumes: existing `LegacyNovelFetchPage` and `novelFetchWorkshop` API helpers.
- Produces: default `NovelFetchPage` renders the React implementation; no iframe is created.

- [x] **Step 1: Write the failing source test**

  Assert that the exported default page uses the React implementation and that the canonical page source no longer contains `src={\`/batch-rewrite/v2` or `novel-fetch-original-workbench`.

- [x] **Step 2: Run the focused test and verify it fails**

  Run: `node --test frontend/src/user/App.performance.test.js test/novel-fetch-performance.test.js`

- [x] **Step 3: Switch the default export**

  Make `NovelFetchPage` render the existing React workbench implementation, keep `LegacyNovelFetchPage` as the named compatibility export only while tests and UI parity are completed, and remove iframe-only status handling from the default path.

- [x] **Step 4: Update layout/status tests and styles**

  Replace iframe visibility assertions with React page assertions; keep the shared page spacing and loading states.

- [x] **Step 5: Run focused tests**

  Expected: all focused tests pass and no canonical novel-fetch page test references the old iframe.

- [x] **Step 6: Commit**

  `git add frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch.css frontend/src/user/App.performance.test.js test/novel-fetch-performance.test.js && git commit -m "refactor: make novel fetch react entry canonical"`

**实际结果：** `/novel-fetch` 和 `/novel-fetch-workshop` 现在共用完整的
`NovelFetchWorkshopPage` React 工作台，小说获取、任务、配置、知识库、处理规则、网站提交和日志都走 Go 的
`/api/novel-fetch-workshop/*` 接口；`NovelFetchPage.jsx` 只保留稳定的模块转出，不再包含旧 iframe。
本次提交：`92d2548f`。验证：`node --test frontend/src/user/App.performance.test.js` 3/3，
`npm --prefix frontend run build` 通过。

### Task 2: 小说获取链路验收与 Go 接口边界

**Files:**
- Modify: `frontend/src/shared/api/novelFetchWorkshop.js` only if an endpoint is missing
- Create/Modify: focused source tests under `test/` or `frontend/src/user/pages/`

**Interfaces:**
- Consumes: Task 1 React entry.
- Produces: a verified mapping for fetch, metadata classification, rules, knowledge, and batch-factory handoff.

- [ ] **Step 1: Add failing coverage for selected-version persistence and handoff payload.**
- [ ] **Step 2: Run the focused tests and observe the failure.**
- [ ] **Step 3: Implement persistence and Go-only handoff.**
- [ ] **Step 4: Run focused frontend tests and `go test ./internal/httpapi ./internal/novelfetchworkshop/...`.**
- [ ] **Step 5: Commit.**

### Task 3: 小说面板 React 迁移边界

**Files:**
- Modify: `frontend/src/user/pages/NovelPanelPage.jsx`
- Modify: `frontend/src/user/pages/NovelPanelPage.test.*` or add focused source test
- Preserve during migration: `public/novel-panel/workbench/**`

**Interfaces:**
- Consumes: existing Go `/api/novel-panel/*` bridge contracts.
- Produces: React-owned asset/project state and a documented list of remaining workbench-only behaviors.

- [ ] **Step 1: Add failing test that canonical page no longer depends on workbench iframe for migrated asset operations.**
- [ ] **Step 2: Implement one vertical slice: asset list, thumbnails, `@` insertion, save and reload.**
- [ ] **Step 3: Run focused React and Go asset tests.**
- [ ] **Step 4: Commit.**

### Task 4: 删除旧静态入口前的最终验收

**Files:**
- Modify: `scripts/build-go-local.ps1`
- Modify: `backend/internal/web/static.go`
- Delete only after acceptance: `public/batch-rewrite/**`, `public/novel-panel/**` and duplicate embed copies

- [ ] **Step 1: Add a build test proving React dist is sufficient without old workbench directories.**
- [ ] **Step 2: Run Go tests and build a single binary.**
- [ ] **Step 3: Start the binary and verify `/novel-fetch`, `/novel-panel`, `/batch-factory` plus key APIs.**
- [ ] **Step 4: Remove old assets only if all acceptance checks pass.**
- [ ] **Step 5: Commit the cleanup locally; do not push.**
