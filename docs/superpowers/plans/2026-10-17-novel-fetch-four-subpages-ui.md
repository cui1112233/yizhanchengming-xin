# Novel Fetch Four Subpages UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the existing task, knowledge, configuration, and submission/history views as four clear independent React + Antd subpages while preserving every existing action and backend contract.

**Architecture:** `NovelFetchPage.jsx` remains the owner of loading, mutation, and routing state. New focused presentation components provide the shared subpage header, compact status strips, and controlled cylinder list; the existing API calls remain in the owner and are passed into the new layouts as callbacks. CSS is scoped under `.novel-fetch-workbench` and responsive behavior degrades the cylinder list to a flat list.

**Tech Stack:** React 18, Ant Design 5, Lucide React, CSS, Node test runner, Vite, Go embed local review service.

## Global Constraints

- Work directly on branch `v88`; do not create a feature branch.
- Modify source first; never copy `frontend/dist` or other compiled output back into source.
- User-facing `121` naming remains `视频管理系统`.
- Preserve full `originalRaw`, process only up to `maxTxt`, and keep processed/raw counts visible.
- Preserve `input_ready` as `分类信息已就绪` and `running/processing` as `正在执行中…`.
- Do not change Go APIs, MySQL, Redis, queue protocols, or TOS for this UI task.
- Do not trigger real external submission, delete real tasks, overwrite knowledge data, or deploy publicly.
- Formal pages must not contain the four-page review navigation shown in the mockup.

---

### Task 1: Lock the page structure with source contract tests

**Files:**
- Create: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`
- Read: `frontend/src/user/pages/NovelFetchPage.jsx`
- Read: `frontend/src/user/pages/novel-fetch.css`

**Interfaces:**
- Consumes: the current `NovelFetchPage.jsx` source and CSS source as UTF-8 strings.
- Produces: regression assertions for independent subpage shells, task batch menu, detail drawer, knowledge/rules tabs, configuration groups, submit/history/log tabs, and responsive cylinder fallback.

- [ ] **Step 1: Write the failing source contract test**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const page = fs.readFileSync(new URL('./NovelFetchPage.jsx', import.meta.url), 'utf8');
const css = fs.readFileSync(new URL('./novel-fetch.css', import.meta.url), 'utf8');

test('novel fetch exposes four independent subpage shells without review navigation', () => {
  assert.match(page, /NovelFetchSubpageShell/);
  assert.match(page, /title="任务"/);
  assert.match(page, /title="知识库"/);
  assert.match(page, /title="配置"/);
  assert.match(page, /title="提交与记录"/);
  assert.doesNotMatch(page, /设计稿预览工具/);
});

test('task detail is a drawer and current batch actions are collapsed', () => {
  assert.match(page, /TaskDetailDrawer/);
  assert.match(page, /当前批次操作/);
  assert.match(page, /重试当前批次失败/);
  assert.match(page, /删除全部/);
});

test('knowledge includes controlled cylinder list and complete rule workspace', () => {
  assert.match(page, /NovelFetchCylinderList/);
  for (const label of ['资料库', '爆款开头拆解', '处理规则', '规则 AI 助手', '高级 JSON']) assert.match(page, new RegExp(label));
  assert.match(css, /prefers-reduced-motion/);
});
```

- [ ] **Step 2: Run the new test and verify it fails**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL because the shared shell, cylinder component, and final layout markers do not exist yet.

- [ ] **Step 3: Commit the failing contract test**

```powershell
git add -- frontend/src/user/pages/novel-fetch-subpages.layout.test.js
git commit -m "test: 固定小说获取四页交互契约"
```

### Task 2: Add focused presentation components

**Files:**
- Create: `frontend/src/user/pages/novel-fetch-workbench-ui.jsx`
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: `activeKey`, `items`, `onChange`, `title`, `description`, `onBack`, and existing Antd children.
- Produces: `NovelFetchSubpageShell`, `NovelFetchMetricStrip`, and `NovelFetchCylinderList` exports.
- `NovelFetchCylinderList` signature: `{ items, selectedId, onSelect, getTitle, getMeta, ariaLabel }`.

- [ ] **Step 1: Add component export assertions**

```js
const ui = fs.readFileSync(new URL('./novel-fetch-workbench-ui.jsx', import.meta.url), 'utf8');
for (const name of ['NovelFetchSubpageShell', 'NovelFetchMetricStrip', 'NovelFetchCylinderList']) {
  assert.match(ui, new RegExp(`export function ${name}`));
}
```

- [ ] **Step 2: Run the test and verify missing component failure**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL because `novel-fetch-workbench-ui.jsx` does not exist.

- [ ] **Step 3: Implement the shared shell and controlled list**

Use Antd `Button`, `Tabs`, and Lucide `ArrowLeft`. The cylinder list must calculate visual distance in `requestAnimationFrame`, set `aria-selected`, scroll the selected item to center, accept keyboard ArrowUp/ArrowDown, and call `onSelect` without mutating item data.

```jsx
export function NovelFetchSubpageShell({ title, description, onBack, actions, children }) {
  return <section className="novel-fetch-subpage">
    <header className="novel-fetch-subpage__header">
      <Button aria-label="返回处理页" icon={<ArrowLeft size={16} />} onClick={onBack} />
      <div className="novel-fetch-subpage__title"><Typography.Title level={3}>{title}</Typography.Title><Typography.Text type="secondary">{description}</Typography.Text></div>
      <Space wrap>{actions}</Space>
    </header>
    {children}
  </section>;
}
```

- [ ] **Step 4: Import the components into `NovelFetchPage.jsx`**

```jsx
import { NovelFetchCylinderList, NovelFetchMetricStrip, NovelFetchSubpageShell } from './novel-fetch-workbench-ui.jsx';
```

- [ ] **Step 5: Run the source contract test**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: shared component assertions PASS; page layout assertions may still fail until later tasks.

- [ ] **Step 6: Commit the focused UI primitives**

```powershell
git add -- frontend/src/user/pages/novel-fetch-workbench-ui.jsx frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-subpages.layout.test.js
git commit -m "web: 增加小说获取子页面展示组件"
```

### Task 3: Rebuild the task subpage without persistent detail panels

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: existing task state and callbacks including `loadTasks`, `loadCurrentBatch`, `loadHistoryBatches`, `handleRetryTasks`, `handleDeleteTasks`, `handleGenerateAi`, `handleSensitiveRepair`, `applySelectedRules`, `applyAllRules`, `submitVisibleTasks`, `transferToBatchFactory`, and `TaskDetailDrawer`.
- Produces: `tasksPanel` rendered inside `NovelFetchSubpageShell`; no API signature changes.

- [ ] **Step 1: Extend the test for collapsed whole-batch actions and row details**

Assert the page contains `Dropdown`, `当前批次操作`, `TaskDetailDrawer`, `详情`, `重新抓取`, and `恢复原文`, while the CSS contains `.novel-fetch-task-selection-bar` and `.novel-fetch-task-summary`.

- [ ] **Step 2: Run the test and verify the new assertions fail**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL on task layout class names and collapsed action marker.

- [ ] **Step 3: Compose the task page**

Use Antd `Tabs` for current/history views, a separate date toolbar, a conditional selected-task action bar, the existing table, and an Antd `Dropdown` for whole-batch operations. Put delete actions in a danger group. Keep `TaskDetailDrawer` closed until the row `详情` button sets `taskDetail`.

- [ ] **Step 4: Preserve task behavior and status semantics**

Verify every existing callback remains reachable and keep `TaskStatusTag`, processed/raw counts, real date filtering, pagination, and historical rerun actions unchanged.

- [ ] **Step 5: Run task and status tests**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js frontend/src/user/pages/novel-fetch-status.test.js`

Expected: PASS.

- [ ] **Step 6: Commit the task page**

```powershell
git add -- frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch.css frontend/src/user/pages/novel-fetch-subpages.layout.test.js
git commit -m "web: 重整小说获取任务页面"
```

### Task 4: Separate knowledge, opening analysis, and processing rules

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: existing knowledge and rule state, `loadKnowledge`, `saveKnowledgeUsagePrompt`, `saveKnowledge`, `handleOpeningAnalyze`, `handleOpeningSaveAnalyzed`, `handleRulePreview`, `handleRuleSuggest`, `saveRuleConfiguration`, and `RuleEditor`.
- Produces: one knowledge subpage with internal keys `library`, `opening`, and `rules`; selection stays controlled by `knowledgeSelectedId`.

- [ ] **Step 1: Add exhaustive labels to the layout test**

Assert all six rule types, three rule scopes, knowledge actions, opening analysis actions, `处理后文本`, `阶段追踪`, and `规则建议` remain in the source.

- [ ] **Step 2: Run the test and observe missing grouped layout markers**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL until the internal workspace tabs and cylinder component are used.

- [ ] **Step 3: Build the three internal workspaces**

Replace the current `panels.knowledge + panels.rules` vertical concatenation with Antd tabs. The library tab contains usage prompt, all ten statistics, type/search controls, `NovelFetchCylinderList`, and the complete entry editor. The opening tab contains style, extra requirement, source, editable result, analyze, save, normalize, and clear. The rules tab retains the complete `RuleEditor`, AI helper, advanced JSON, test text, result, trace, and suggestions.

- [ ] **Step 4: Protect drafts and empty knowledge responses**

Keep selection controlled, preserve the current draft before changing the centered cylinder item, and do not call a save method from load or empty-state handling.

- [ ] **Step 5: Run knowledge and layout tests**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js test/novel-fetch-knowledge-recovery.test.js`

Expected: PASS.

- [ ] **Step 6: Commit the knowledge and rule page**

```powershell
git add -- frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch.css frontend/src/user/pages/novel-fetch-subpages.layout.test.js
git commit -m "web: 整理知识库与处理规则工作区"
```

### Task 5: Reorganize configuration with effective-value feedback

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: `configForm`, `loadConfig`, `saveWorkshopConfig`, `testWorkshopAi`, version profile state, sync callbacks, and `saveVersionConfig`.
- Produces: five controlled configuration sections and a read-only effective-value summary; persisted payload remains unchanged.

- [ ] **Step 1: Add configuration coverage assertions**

Assert the five group labels, `当前生效值`, `有未保存的更改`, `保存配置`, `测试文本模型`, all sync actions, and `保存版本配置` are present.

- [ ] **Step 2: Run the test and verify grouped configuration failure**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL on effective-value and dirty-state markers.

- [ ] **Step 3: Replace the all-open collapse with controlled section navigation**

Keep one Antd `Form` mounted while switching visible sections so values are never unmounted or cleared. Use `Form.useWatch` and `onValuesChange` to show selected values and an unsaved badge; reset the dirty flag only after a successful save.

- [ ] **Step 4: Integrate the existing version profile panel as the fifth section**

Keep all sync callbacks, default profile, per-version bindings, three prompts, and save action. Do not hardcode system prompt contents in the frontend.

- [ ] **Step 5: Run layout and API source tests**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js frontend/src/shared/api/novelFetchWorkshop.test.js`

Expected: PASS.

- [ ] **Step 6: Commit the configuration page**

```powershell
git add -- frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch.css frontend/src/user/pages/novel-fetch-subpages.layout.test.js
git commit -m "web: 重组小说获取配置页面"
```

### Task 6: Rebuild submission, result, history, and logs as stages

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: existing web submission config, plan, result, history, global logs, `handleWebSubmit`, and all existing environment/sync/preview/start/retry callbacks.
- Produces: one submission subpage with internal keys `ready`, `result`, `history`, and `logs`; no submission runs on mount.

- [ ] **Step 1: Add submission-stage coverage assertions**

Assert all four stage labels, all four sync/check actions, preview, start, settings, summary, single-version retry, refresh history, task log selection, and global logs remain discoverable.

- [ ] **Step 2: Run the test and verify stage layout failure**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL until submit and logs stop rendering as one vertical page.

- [ ] **Step 3: Compose the four submission stages**

Use a stage strip plus Antd tabs. Keep preparation controls in `ready`, current-run rows in `result`, persisted records in `history`, and task/global logs in `logs`. Preserve confirmation before `startWebSubmit` and disable start until tasks are selected and a preview exists.

- [ ] **Step 4: Preserve retry and duplicate-skip safety**

Single-version retry passes only `{ taskIds: [row.bookId], versions: [row.version] }`; `skip_submitted` remains enabled unless explicitly turned off.

- [ ] **Step 5: Run submission guard and layout tests**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js test/novel-fetch-selected-submit-payload-guard.test.js test/novel-fetch-upload-policy.test.js`

Expected: PASS.

- [ ] **Step 6: Commit the submission page**

```powershell
git add -- frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch.css frontend/src/user/pages/novel-fetch-subpages.layout.test.js
git commit -m "web: 整理小说获取提交与记录页面"
```

### Task 7: Responsive polish, build, and local Go acceptance

**Files:**
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Test: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`
- Generated by build only: `frontend/dist/**`

**Interfaces:**
- Consumes: all completed subpages and the approved Go embed review scripts.
- Produces: responsive desktop/mobile layouts, reduced-motion fallback, successful frontend build, and local route acceptance.

- [ ] **Step 1: Add responsive source assertions**

Assert CSS contains breakpoints for subpage grids, task actions, configuration navigation, submit stages, flat cylinder fallback, and `prefers-reduced-motion: reduce`.

- [ ] **Step 2: Run the layout test and verify any missing responsive markers**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL only for missing responsive rules.

- [ ] **Step 3: Complete responsive CSS**

Ensure page-level horizontal overflow is absent, tables scroll within their own containers, button text is not clipped, the drawer fits narrow screens, and the cylinder list becomes flat at the mobile breakpoint and under reduced motion.

- [ ] **Step 4: Run focused tests and the frontend build**

Run:

```powershell
node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js frontend/src/user/pages/novel-fetch-status.test.js frontend/src/shared/api/novelFetchWorkshop.test.js
pnpm --dir frontend build
```

Expected: all tests PASS and Vite build exits 0.

- [ ] **Step 5: Build and start the Go embedded local review service**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1
```

Expected: health and `/novel-fetch` checks pass on the local review service. This proves local serving, not real provider submission.

- [ ] **Step 6: Inspect desktop and narrow viewports**

Open `/novel-fetch?tab=tasks`, `knowledge`, `config`, and `submit`. Verify return actions, menu/drawer behavior, internal tabs, no overlap, and no page-level horizontal overflow. Do not click destructive or external submission actions.

- [ ] **Step 7: Commit the final responsive and verification changes**

```powershell
git add -- frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-workbench-ui.jsx frontend/src/user/pages/novel-fetch.css frontend/src/user/pages/novel-fetch-subpages.layout.test.js frontend/dist
git commit -m "web: 完成小说获取四个子页面重设计"
```
