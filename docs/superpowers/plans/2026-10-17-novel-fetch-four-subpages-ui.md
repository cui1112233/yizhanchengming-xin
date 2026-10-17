# Novel Fetch Four Subpages UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild task, knowledge, configuration, and submission/history as four clear independent React + Antd subpages while preserving every existing action and backend contract.

**Architecture:** `NovelFetchPage.jsx` remains the owner of data and mutations. Focused presentation components provide the common subpage shell and controlled cylinder list; pure cylinder math and keyboard-selection behavior live in a separately tested JavaScript module. CSS stays scoped under `.novel-fetch-workbench`.

**Tech Stack:** React 18, Ant Design 5, Lucide React, CSS, Node test runner, Vite, Go embed local review service.

## Global Constraints

- Work directly on branch `v88`; do not create another branch or worktree.
- Modify React source first; never copy compiled output back into source.
- User-facing `121` naming remains `视频管理系统`.
- Preserve full `originalRaw`, process only to `maxTxt`, and show processed/raw counts.
- Preserve `input_ready` as `分类信息已就绪` and `running/processing` as `正在执行中…`.
- Do not change Go APIs, MySQL, Redis, queue protocols, or TOS.
- Do not trigger real submission, delete real data, or deploy publicly.
- Formal pages must not contain the mockup review navigation.

---

### Task 1: Implement and test cylinder-list behavior

**Files:**
- Create: `frontend/src/user/pages/novel-fetch-cylinder.js`
- Create: `frontend/src/user/pages/novel-fetch-cylinder.test.js`

**Interfaces:**
- Produces: `cylinderVisualState(distance: number): { rotateX: number, translateZ: number, scale: number, opacity: number, brightness: number }`.
- Produces: `nextCylinderIndex(length: number, currentIndex: number, direction: -1 | 1): number`.

- [ ] Write failing tests with hand-derived boundary values: distance `0` returns the centered state; large distances clamp rotation to 48 degrees, scale to 0.78, opacity to 0.18; keyboard movement clamps at list boundaries.
- [ ] Run `node --test frontend/src/user/pages/novel-fetch-cylinder.test.js` and confirm failure because the module is missing.
- [ ] Implement the two pure functions without DOM access.
- [ ] Re-run the test and confirm PASS.
- [ ] Commit both files with `git commit -m "web: 增加知识库圆柱列表计算"`.

### Task 2: Add focused presentation components

**Files:**
- Create: `frontend/src/user/pages/novel-fetch-workbench-ui.jsx`
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`

**Interfaces:**
- Produces: `NovelFetchSubpageShell({ title, description, onBack, actions, children })`.
- Produces: `NovelFetchMetricStrip({ items })`.
- Produces: `NovelFetchCylinderList({ items, selectedId, onSelect, getTitle, getMeta, ariaLabel })`.
- Consumes: `cylinderVisualState` and `nextCylinderIndex` from Task 1.

- [ ] Implement the shell with an icon-only back button whose tooltip/accessible name is “返回处理页”.
- [ ] Implement the metric strip as a stable responsive grid.
- [ ] Implement the controlled cylinder list using `requestAnimationFrame`, center snapping, click selection, and ArrowUp/ArrowDown selection.
- [ ] Import the components in `NovelFetchPage.jsx` without changing data behavior.
- [ ] Run `node --test frontend/src/user/pages/novel-fetch-cylinder.test.js` and `pnpm --dir frontend build`; both must pass.
- [ ] Commit with `git commit -m "web: 增加小说获取子页面展示组件"`.

### Task 3: Rebuild the task subpage

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`

**Interfaces:**
- Consumes all existing task callbacks and `TaskDetailDrawer`.
- Produces a task page with current/history tabs, date toolbar, conditional selection toolbar, task table, batch summary, and collapsed current-batch menu.

- [ ] Wrap the page in `NovelFetchSubpageShell` and retain current/history switching.
- [ ] Separate date controls from selected-task controls.
- [ ] Show selected-task actions only when one or more tasks are selected; retain retry, stop, delete, AI generation, sensitive repair, AI count, rule scope, rule processing, and selected submission.
- [ ] Move whole-batch retry, factory transfer, rule processing, submit, resubmit, delete-failed, and delete-all into an Antd `Dropdown` labelled “当前批次操作”.
- [ ] Keep task detail hidden until the row “详情” button opens `TaskDetailDrawer`; preserve refetch and restore-original actions.
- [ ] Run `node --test frontend/src/user/pages/novel-fetch-status.test.js frontend/src/user/pages/novel-fetch-cylinder.test.js` and `pnpm --dir frontend build`.
- [ ] Commit with `git commit -m "web: 重整小说获取任务页面"`.

### Task 4: Rebuild knowledge and processing rules

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`

**Interfaces:**
- Consumes existing knowledge/rule state and callbacks.
- Produces internal workspaces `library`, `opening`, and `rules`; cylinder selection remains controlled by `knowledgeSelectedId`.

- [ ] Replace vertical `knowledge + rules` rendering with three internal tabs.
- [ ] In “资料库”, retain usage prompt, all ten statistics, six knowledge types, search, refresh, add, edit, duplicate, delete, AI optimize, full JSON, and use `NovelFetchCylinderList` for item selection.
- [ ] In “爆款开头拆解”, retain style, extra requirement, source, editable result, AI analysis, normalize, confirm-save, and clear.
- [ ] In “处理规则”, retain all six rule types, all three scopes, load current source, preview, AI suggestion, save, full editor, AI helper, advanced JSON, test text, processed result, trace, and suggestions.
- [ ] Preserve local drafts when selection changes; loading and empty states must never call save methods.
- [ ] Run `node --test frontend/src/user/pages/novel-fetch-cylinder.test.js test/novel-fetch-knowledge-recovery.test.js` and `pnpm --dir frontend build`.
- [ ] Commit with `git commit -m "web: 整理知识库与处理规则工作区"`.

### Task 5: Reorganize configuration

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`

**Interfaces:**
- Consumes `configForm`, existing config load/save/test callbacks, profile state, sync callbacks, and `saveVersionConfig`.
- Produces five mounted configuration sections and a read-only effective-value summary.

- [ ] Replace the all-open collapse with navigation for workflow/fetch, supplement, AI, platform/style catalogs, and version profiles.
- [ ] Keep one Antd Form mounted so switching sections cannot erase unsaved values.
- [ ] Use `Form.useWatch` plus `onValuesChange` to show selected values and unsaved state; clear dirty state only after a successful save.
- [ ] Retain every existing field, test-text-model action, all three sync actions, default/per-version profiles, three version prompts, and both save actions.
- [ ] Keep real system prompt contents backend-owned; add no new frontend prompt literal.
- [ ] Run `node --test frontend/src/shared/api/novelFetchWorkshop.test.js` and `pnpm --dir frontend build`.
- [ ] Commit with `git commit -m "web: 重组小说获取配置页面"`.

### Task 6: Rebuild submission and records

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`

**Interfaces:**
- Consumes existing web submission config, plan, result, history, logs, and callbacks.
- Produces internal stages `ready`, `result`, `history`, and `logs`; mounting the page performs no external submission.

- [ ] Add a non-interactive environment/config/preview/execute status strip and four internal tabs.
- [ ] Keep environment detection, three sync actions, preview, organization, minimum chars, retries, skip-submitted, save configuration, selection summary, and start confirmation in “提交准备”.
- [ ] Keep current results and exact single-version failure retry in “本次结果”.
- [ ] Keep refreshable history in “提交历史”.
- [ ] Keep task-specific and global logs in “操作日志”.
- [ ] Disable start until tasks are selected and preview data exists; do not call start on mount.
- [ ] Run `node --test test/novel-fetch-selected-submit-payload-guard.test.js test/novel-fetch-upload-policy.test.js` and `pnpm --dir frontend build`.
- [ ] Commit with `git commit -m "web: 整理小说获取提交与记录页面"`.

### Task 7: Responsive polish and real local acceptance

**Files:**
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Generated by build: `frontend/dist/**`

**Interfaces:**
- Consumes all completed subpages and existing Go local review scripts.
- Produces responsive desktop/mobile UI and verified local Go embed rendering.

- [ ] Add desktop and narrow layout rules for the subpage shell, task toolbar/menu, cylinder list, configuration navigation, submission stages, tables, and drawer.
- [ ] Under `prefers-reduced-motion: reduce` and the mobile breakpoint, remove cylinder transforms and smooth scrolling while retaining selection highlight.
- [ ] Run focused Node tests and `pnpm --dir frontend build`; all must pass.
- [ ] Run `scripts/build-go-local.ps1`, `scripts/start-local-go-review.ps1 -UseYcmMysql`, and `scripts/check-local-go.ps1`.
- [ ] In the local browser, visit `?tab=tasks`, `knowledge`, `config`, and `submit`; exercise only navigation, menu, drawer, tabs, scrolling, and responsive layouts. Do not click delete or external submission actions.
- [ ] Verify no page-level horizontal overflow, clipped button text, overlap, console error, or blank panel at desktop and narrow sizes.
- [ ] Commit source plus newly built embed assets with `git commit -m "web: 完成小说获取四个子页面重设计"`.
