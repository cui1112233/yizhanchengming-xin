# Batch Factory Shared Compact Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Batch Factory workbench and single-book settings compact, aligned, and visually shared across comic, video, and text modes while improving the text-page preview.

**Architecture:** Keep `batch-factory-engine-*` and `shuihuo-workbench-*` as the shared visual primitives. Text mode supplies only its fields and preview content; it must not create a parallel card or field system. Browser tests assert real computed geometry at the 1021×901 acceptance viewport.

**Tech Stack:** React, Ant Design, CSS, Node test runner, esbuild, headless Chrome, Vite, Go embed.

## Global Constraints

- Work directly on branch `v88`; do not create a branch.
- Preserve unrelated dirty files and commit only the listed task files.
- Do not change backend APIs, data structures, production flow, or upload flow.
- Do not push or deploy; update only local port 4001.
- Comic, video, and text modes must share workbench header and form primitives.
- Text page previews must preserve 3:4, use contain behavior, stay centered, and must not force the row taller.

---

### Task 1: Shared single-book form layout

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/TextProductionSettingsForm.jsx:47-134`
- Modify: `frontend/src/user/pages/shuihuo-production.css:748-756`
- Test: `frontend/src/user/pages/shuihuo/TextProductionSettingsForm.browser.test.js`

**Interfaces:**
- Consumes: existing `TextProductionSettingsForm({ value, explicitKeys, onChange, onRestore, scope, open, activeSection })`.
- Produces: the same component interface, rendered through shared `.batch-factory-engine-grid`, `.batch-factory-engine-field`, and new shared grouping modifiers that media-mode forms may also reuse.

- [ ] **Step 1: Extend the real-browser test with layout and copy assertions**

Add assertions after rendering the layout and publish tabs:

```js
assert.equal(document.body.textContent.includes('复用已有标红词，本地排版。'), false);
const fields = [...document.querySelectorAll('.batch-factory-engine-field')];
assert.ok(fields.length > 4);
assert.ok(fields.every(field => getComputedStyle(field).display === 'grid'));
const firstRows = fields.slice(0, 4).map(field => field.getBoundingClientRect().height);
assert.ok(Math.max(...firstRows) - Math.min(...firstRows) <= 2);
```

For the publish tab, assert that uploader, organization, website profile, and category controls occupy the same two-column grid and that the sync action remains adjacent to the website-profile field.

- [ ] **Step 2: Run the focused browser test and verify RED**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/TextProductionSettingsForm.browser.test.js
```

Expected: FAIL because the redundant highlight note exists and current fields have uneven row composition.

- [ ] **Step 3: Implement the shared A-layout**

In `TextProductionSettingsForm.jsx`:

- Remove only the inline `<small>复用已有标红词，本地排版。</small>`.
- Add semantic grouping wrappers for “正文排版” and “背景与标题”, but keep every actual field on `.batch-factory-engine-field`.
- Keep “同步网站配置档” inside the website-profile field action area.
- Render auto-collection and auto-upload with the same shared switch-row modifier.

In `shuihuo-production.css`:

```css
.batch-factory-engine-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }
.batch-factory-engine-group { grid-column:1/-1; display:grid; grid-template-columns:subgrid; gap:12px 20px; }
.batch-factory-engine-group-title { grid-column:1/-1; }
.batch-factory-engine-field.is-switch-row { min-height:36px; }
```

Use existing theme variables and existing Ant Design controls. At the current `780px` breakpoint, all groups and fields collapse to one column.

- [ ] **Step 4: Run the focused browser test and verify GREEN**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/TextProductionSettingsForm.browser.test.js
```

Expected: PASS, with no redundant note and aligned field geometry.

- [ ] **Step 5: Commit Task 1**

```powershell
git add frontend/src/user/pages/shuihuo/TextProductionSettingsForm.jsx frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/TextProductionSettingsForm.browser.test.js
git commit -m "fix(batch-factory): align shared single-book settings"
```

### Task 2: Shared compact workbench header

**Files:**
- Modify: `frontend/src/user/pages/shuihuo-production.css:199-201,1089-1110,4244-4260`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryWorkbenchLayout.source.test.js`
- Test: `frontend/src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js`

**Interfaces:**
- Consumes: `.shuihuo-workbench-head` and `.batch-factory-workbench-table` used by comic, video, and text rows.
- Produces: one shared compact header rule; no `.is-text-mode` or text-only header override.

- [ ] **Step 1: Add a geometry assertion for the shared table header**

In the browser fixture, render a real `.shuihuo-workbench-head` above the text row and assert:

```js
const head = document.querySelector('.shuihuo-workbench-head').getBoundingClientRect();
assert.ok(head.height >= 30 && head.height <= 34);
const labels = [...document.querySelectorAll('.shuihuo-workbench-head > div')];
assert.ok(labels.every(label => Math.abs(label.getBoundingClientRect().height - head.height) <= 1));
```

Add a source-level guard that the compact rule targets `.batch-factory-workbench-table .shuihuo-workbench-head` without a production-mode suffix.

- [ ] **Step 2: Run the focused tests and verify RED**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/batchFactoryWorkbenchLayout.source.test.js src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js
```

Expected: FAIL because the current header is 40px plus 10px vertical cell padding.

- [ ] **Step 3: Apply one shared compact header rule**

Add the final shared rule after existing overrides so cascade order is deterministic:

```css
.batch-factory-workbench-table .shuihuo-workbench-head { min-height:32px; }
.batch-factory-workbench-table .shuihuo-workbench-head > div { display:flex; align-items:center; min-height:32px; padding:5px 10px; }
```

Keep the first checkbox/sequence column centered through its existing selector. Do not reduce text below the current readable shared size.

- [ ] **Step 4: Run the focused tests and verify GREEN**

Run the same two-test command. Expected: PASS.

- [ ] **Step 5: Commit Task 2**

```powershell
git add frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/batchFactoryWorkbenchLayout.source.test.js frontend/src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js
git commit -m "fix(batch-factory): compact shared workbench header"
```

### Task 3: Complete centered text-page preview

**Files:**
- Modify: `frontend/src/user/pages/shuihuo-production.css:3343-3360`
- Test: `frontend/src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js`

**Interfaces:**
- Consumes: existing `.batch-factory-text-page-image` button and its real `<img>` rendered by `InlineTextBookPages`.
- Produces: a complete, centered, contained 3:4 preview within the shared cell skeleton; no React API changes.

- [ ] **Step 1: Add real-browser preview geometry assertions**

After loading a page with a real data-URL image, assert:

```js
const preview = document.querySelector('.batch-factory-text-page-image').getBoundingClientRect();
const image = document.querySelector('.batch-factory-text-page-image img').getBoundingClientRect();
assert.ok(Math.abs((image.left + image.right) / 2 - (preview.left + preview.right) / 2) <= 1);
assert.ok(image.width > 120);
assert.ok(image.right <= preview.right && image.bottom <= preview.bottom);
assert.ok(Math.abs(image.width / image.height - 0.75) < 0.02);
```

Retain the existing pager-bottom equality assertion.

- [ ] **Step 2: Run the browser test and verify RED**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js
```

Expected: FAIL because the fixed `height:150px;width:auto` preview remains narrow in a wide cell.

- [ ] **Step 3: Make the preview contained, centered, and responsive**

Use the existing preview button as the sizing box:

```css
.batch-factory-text-page-image { width:100%; min-height:0; height:100%; place-content:center; }
.batch-factory-text-page-image img { width:min(100%,180px); height:auto; max-height:100%; object-fit:contain; aspect-ratio:3/4; }
```

Keep the page header and pager full width. Do not set a new minimum row height.

- [ ] **Step 4: Run the browser test and verify GREEN**

Run the focused browser test. Expected: PASS, including pager delta 0 and complete centered preview.

- [ ] **Step 5: Run the full scoped verification**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/TextProductionSettingsForm.browser.test.js src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js src/user/pages/shuihuo/batchFactoryWorkbenchLayout.source.test.js src/user/pages/shuihuo/BatchFactoryCompactRows.source.test.js
npm run build
```

Expected: all tests PASS and Vite exits 0.

- [ ] **Step 6: Rebuild and restart only the confirmed local Go runtime**

Verify the port 4001 listener path is `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`, then run:

```powershell
.\scripts\start-local-go-review.ps1 -Detached
```

Check `/health`, `/api/runtime-build-info`, and the actual browser page at 1021×901.

- [ ] **Step 7: Commit Task 3**

```powershell
git add frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/InlineTextBookPages.layout.browser.test.js
git commit -m "fix(batch-factory): center responsive text previews"
```
