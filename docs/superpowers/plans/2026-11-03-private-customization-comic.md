# 私人定制漫画工作区 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (or superpowers:subagent-driven-development) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `/script` 入口改名为“私人定制”，保留现有剧本流程，并新增一个可直接粘贴正文、选择画风库配置、启动单篇漫画生成的工作区。

**Architecture:** 继续使用 `/script` 作为兼容路由，在 `ScriptPage` 内增加“剧本创作 / 漫画生成”工作区切换。漫画工作区复用 Batch Factory 的漫画配置目录、画风选择器、任务状态和媒体接口；由于现有手工批量接口要求书籍 ID，后端补一个最小的单正文创建适配接口，不复制漫画生产流水线。

**Tech Stack:** React、Ant Design、现有 Batch Factory V12 API、Go HTTP API、MySQL/Redis 任务存储、Go embed 本地构建。

**Spec:** `docs/superpowers/specs/2026-11-03-private-customization-comic-design.md`

## Global Constraints

- 左侧导航、首页入口和页面标题统一显示“私人定制”。
- 保留 `/script` 地址，旧链接和历史记录继续可用。
- 漫画工作区保留现有“画风库”选择、画风预览、全局参考图和图片描述词前后缀配置。
- 私人定制默认只生成漫画内容和合集，不自动向外部漫画后台发布。
- 预设提示词仍由后端提供，前端不新增硬编码系统提示词。
- 不重写现有批量工厂漫画流水线，不迁移或删除历史数据。
- 保留工作区现有脏文件，不回滚、清理、提交或远程发布无关改动。

## Review Focus

- 空标题：使用正文第一条非空短句作为可编辑默认标题；测试归入漫画表单任务。
- 空正文或正文过短：不发起任务并在当前页面提示；测试归入漫画表单任务。
- 画风库没有已发布项或预览图读取失败：仍可显示明确状态，不让页面卡死；测试归入画风配置任务。
- 任务启动成功但状态接口暂时失败：保留任务 ID并允许再次刷新；测试归入任务状态任务。
- 任务处于失败/部分完成：展示具体阶段和错误，并提供进入批量工厂详情/重试入口；测试归入任务状态任务。

---

### Task 1: 提取私人定制漫画的纯状态与请求组装逻辑

**Files:**
- Create: `frontend/src/user/pages/privateCustomizationComic.js`
- Test: `frontend/src/user/pages/privateCustomizationComic.test.js`

**Interfaces:**
- Produces `privateComicTitle(sourceText, title)`, `privateComicDraftKey`, `privateComicTaskSnapshot`, and `buildPrivateComicSettings` helpers for the page.
- `buildPrivateComicSettings({ comic, textModelId, imageModelId, imageProvider })` returns a Batch Factory-compatible `initialBatchSettings` object with `productionMode: 'comic'` and the selected `aiPromptConfig.comic` snapshot.

- [ ] **Step 1: Write failing tests** for title fallback, empty input validation, draft snapshot round-trip, and settings preserving style/reference/prefix/suffix/constraints without embedding system prompt text.
- [ ] **Step 2: Run the focused frontend test** and verify the new helper tests fail before implementation.
- [ ] **Step 3: Implement the pure helpers** with trimmed strings, a short first non-empty title line fallback, and JSON-safe snapshots.
- [ ] **Step 4: Run the focused test** and verify it passes.

### Task 2: Add a minimal owner-scoped API for one pasted comic source

**Files:**
- Create: `backend/internal/httpapi/private_customization_comic.go`
- Test: `backend/internal/httpapi/private_customization_comic_test.go`
- Modify: `backend/internal/httpapi/batch_factory_v12_aliases.go` or the V12 route registration file that owns the compatibility surface
- Modify: `frontend/src/shared/api/batchFactoryV12.js`

**Interfaces:**
- Add `POST /api/batch-factory/v12/private-comic/tasks`.
- Request: `{ title, sourceText, initialBatchSettings }`.
- Response: `{ batch, book }` with the owner-scoped IDs needed by the existing automation and runtime-summary APIs.
- Add frontend `createPrivateComicTask(payload)` wrapper.

- [ ] **Step 1: Write failing Go tests** proving a non-empty pasted source creates exactly one book with the supplied title/source text, stores comic settings, and rejects empty source text.
- [ ] **Step 2: Run the focused Go test** and verify failure on the missing route/handler.
- [ ] **Step 3: Implement the handler** by creating one intake/book through the existing store contracts, setting private-customization source metadata, and creating a single batch; do not introduce a second comic generation engine or public-publish behavior.
- [ ] **Step 4: Register the V12 route and add the API wrapper** while preserving existing V11/V12 aliases.
- [ ] **Step 5: Run `go test ./internal/httpapi -run 'TestPrivateCustomizationComic' -count=1`** and verify all focused tests pass.

### Task 3: Add the “私人定制” labels and two-workspace shell

**Files:**
- Modify: `frontend/src/shared/layouts/UserLayout.jsx`
- Modify: `frontend/src/user/pages/HomePage.jsx`
- Modify: `frontend/src/user/pages/ScriptPage.jsx`
- Modify: related script-page stylesheet if the existing layout needs tab styles
- Test: existing source tests for `UserLayout`, `HomePage`, and `ScriptPage`, plus a focused new source test if no suitable test exists

**Interfaces:**
- Keep the route `/script` and internal script/history identifiers unchanged.
- Add page-local workspace state with keys `script` and `comic`, defaulting to `script`.
- Render the existing script UI unchanged under `script`; render the new comic workspace under `comic`.

- [ ] **Step 1: Add failing source assertions** for navigation/home/page copy (`私人定制`), workspace labels (`剧本创作`, `漫画生成`), and the absence of any route migration.
- [ ] **Step 2: Run the focused frontend source tests** and verify the old labels still fail the new assertions.
- [ ] **Step 3: Change only user-facing labels and add the workspace switcher**, keeping existing script controls, history, model config, and drafts intact.
- [ ] **Step 4: Run the focused source tests** and verify the original script layout assertions still pass.

### Task 4: Build the comic workspace with existing style/config components

**Files:**
- Create: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Create: `frontend/src/user/pages/privateCustomizationComic.css` only if existing shared styles cannot express the layout
- Modify: `frontend/src/user/pages/ScriptPage.jsx`
- Reuse: `frontend/src/user/pages/shuihuo/BatchFactoryComicStylePicker.jsx`, `BatchFactoryUnifiedSettingsModal.jsx`/comic prompt controls, and the existing preset/catalog API
- Test: `frontend/src/user/pages/PrivateComicWorkspace.test.jsx` or the project’s established source-test equivalent

**Interfaces:**
- Props: `{ catalog, modelOptions, draft, onDraftChange, onSubmit, task, onRefresh, onOpenBatchFactory }`.
- Draft fields: `{ title, sourceText, comic, textModelId, imageModelId, imageProvider }`.
- The workspace exposes the exact user actions `开始生成漫画`, `刷新状态`, `查看漫画详情` and `重试/继续` when applicable.

- [ ] **Step 1: Write failing UI/source tests** for title/source inputs, the reused `BatchFactoryComicStylePicker`, the comic prompt/constraint selections, explicit no-publish copy, and disabled submit on empty source.
- [ ] **Step 2: Run the focused frontend test** and verify failure before the workspace exists.
- [ ] **Step 3: Implement the comic form** with pasted text, editable title fallback, a collapsible “漫画设置” region, existing画风库 cover-flow, reference-image switches, prefix/suffix, comic prompt/image-derivation and comic constraints.
- [ ] **Step 4: Ensure catalog bodies and system prompt text remain backend-sourced**; the browser stores only preset IDs/metadata and user-entered overrides.
- [ ] **Step 5: Run the focused UI/source tests** and verify they pass.

### Task 5: Wire creation, automatic execution, status polling, and result handoff

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Modify: `frontend/src/user/pages/ScriptPage.jsx`
- Modify: `frontend/src/shared/api/batchFactoryV12.js` only for wrappers needed by the workspace
- Test: `frontend/src/user/pages/PrivateComicWorkspace.test.jsx` or a focused helper/source test

**Interfaces:**
- Submit flow calls `createPrivateComicTask`, then existing `startBatchAutomation(batchId, { presetId, runMode: 'video_no_submit', autoPublish: false })` using the available saved automation configuration.
- Poll flow calls existing `getBatchRuntimeSummary(batchId)` and `getBatch(batchId)` without exposing batch terminology in the main status copy.
- Persist the latest draft/task snapshot in page-safe local storage; do not store credentials or prompt bodies.

- [ ] **Step 1: Add failing tests** for successful create/start, saved task restoration, polling refresh, failure copy with retry/detail action, and explicit no-auto-publish payload.
- [ ] **Step 2: Run the focused tests** and verify they fail before wiring.
- [ ] **Step 3: Implement submit/poll/retry behavior**, preventing duplicate submits while a request is active and retaining the task ID if a refresh request fails.
- [ ] **Step 4: Add a completed-state result card** with page/collection progress and a link that opens the existing Batch Factory detail/media view for full preview and publishing actions.
- [ ] **Step 5: Run the focused tests** and verify they pass.

### Task 6: Verify source, build, embedded runtime, and visible behavior

**Files:**
- No new product files; inspect the modified paths above.

- [ ] **Step 1: Run focused frontend tests** for the new workspace, style picker reuse, labels, and script regression coverage.
- [ ] **Step 2: Run focused backend tests** for the new one-source endpoint and existing Batch Factory comic route tests.
- [ ] **Step 3: Run the frontend build and Go tests** from the real checkout, then rebuild embedded assets using the existing local build script.
- [ ] **Step 4: Restart only the confirmed local Go runtime**, check `/health` and `/api/runtime-build-info`, and open `/script` in the authenticated local browser.
- [ ] **Step 5: Visually verify** the nav says“私人定制”, script tab remains usable, comic tab accepts pasted content,画风库 is visible, submit is blocked for empty text, and a real task reaches a persisted status/result without automatic publishing. Do not claim completion from build output alone.

## Self-Review

- Spec coverage: navigation, two workspaces, direct paste, title fallback, style/reference/constraint reuse, single-task creation, no-publish mode, status/retry, history/draft preservation, and batch-factory handoff are covered by Tasks 1–6.
- Type consistency: the task wrapper returns `{ batch, book }`; the submit flow uses `batch.id`; status polling uses that same ID; the comic draft carries `aiPromptConfig.comic` into `initialBatchSettings`.
- Review focus: the five failure modes above each have a named test owner.
- Proportion: the plan keeps the existing Batch Factory production engine intact and limits the backend addition to one owner-scoped source adapter.
