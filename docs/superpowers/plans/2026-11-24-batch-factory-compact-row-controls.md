# Batch Factory Compact Row Controls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Compact the per-book controls for comic, video, and text production while aligning the table header, book title row, and paired text-page pagers.

**Architecture:** Keep the change inside the existing Batch Factory workbench CSS scope. Add one final compact-density block that overrides legacy rules without changing React behavior, then protect the scope and exact layout contract with a focused CSS source test and real browser geometry checks.

**Tech Stack:** React, Ant Design, CSS, Node.js test runner, Vite, Go embed, Scrapling browser acceptance.

## Global Constraints

- Only controls inside `.batch-factory-workbench` may shrink; do not alter global Ant Design sizing.
- Row buttons target 26px height and 12px-or-smaller readable text.
- The table header targets about 40px; only the `序号` header uses 8px.
- Page text, comic/video/text media previews, top toolbar, dialogs, and production behavior remain unchanged.
- At narrow widths, retain horizontal scrolling instead of shrinking page content.
- Work only on local 4001; do not push or deploy.

---

### Task 1: Add and protect the compact row layout contract

**Files:**
- Modify: `frontend/src/user/pages/shuihuo-production.css`
- Create: `frontend/src/user/pages/shuihuo/BatchFactoryCompactRows.source.test.js`

**Interfaces:**
- Consumes: Existing classes `.batch-factory-book-content-action`, `.batch-factory-book-config-region`, `.batch-factory-action-group`, `.batch-factory-book-title-row`, `.batch-factory-text-page-cell`, `.batch-factory-text-rendered-cell`, and `.batch-factory-cell-pager`.
- Produces: A workbench-scoped compact CSS contract shared by comic, video, and text rows.

- [ ] **Step 1: Write the failing contract test**

```js
import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const css = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');

test('batch factory row controls share one compact workbench-scoped contract', () => {
  assert.match(css, /BF_COMPACT_ROW_CONTROLS_START[\s\S]*BF_COMPACT_ROW_CONTROLS_END/);
  const block = css.match(/BF_COMPACT_ROW_CONTROLS_START([\s\S]*?)BF_COMPACT_ROW_CONTROLS_END/)?.[1] || '';
  assert.match(block, /\.batch-factory-workbench-table \.shuihuo-workbench-head\s*\{[^}]*height:\s*40px/);
  assert.match(block, /\.shuihuo-workbench-head\s*>\s*div:first-child\s*\{[^}]*font-size:\s*8px/);
  assert.match(block, /\.batch-factory-book-content-action\.ant-btn\s*\{[^}]*height:\s*26px/);
  assert.match(block, /\.batch-factory-book-config-region\s*\{[^}]*padding:\s*4px 6px/);
  assert.match(block, /\.batch-factory-text-page-cell\s*>\s*\.batch-factory-cell-shell/);
  assert.doesNotMatch(block, /\.shuihuo-workbench-toolbar|\.batch-factory-text-page-content\s+p|\.batch-factory-text-page-image\s+img/);
});
```

- [ ] **Step 2: Run the test and verify the compact block is missing**

Run: `node --test src/user/pages/shuihuo/BatchFactoryCompactRows.source.test.js`

Expected: FAIL because `BF_COMPACT_ROW_CONTROLS_START` is absent.

- [ ] **Step 3: Add the final scoped CSS block**

Append a final block after the existing Batch Factory density rules:

```css
/* BF_COMPACT_ROW_CONTROLS_START */
.batch-factory-workbench-table .shuihuo-workbench-head {
  height: 40px;
  min-height: 40px;
}
.batch-factory-workbench-table .shuihuo-workbench-head > div:first-child { font-size: 8px; }
.batch-factory-workbench .batch-factory-book-title-row { align-items: center; min-height: 20px; }
.batch-factory-workbench .batch-factory-book-title-row .batch-factory-production-source { align-items: center; }
.batch-factory-workbench .batch-factory-book-content-actions { gap: 5px !important; margin-top: 6px; }
.batch-factory-workbench .batch-factory-book-content-action.ant-btn {
  height: 26px;
  min-height: 26px;
  padding: 0 10px;
  font-size: 11px;
  line-height: 24px;
}
.batch-factory-workbench .batch-factory-book-config-regions {
  grid-template-rows: repeat(3, minmax(40px, 1fr)) !important;
  gap: 4px !important;
}
.batch-factory-workbench .batch-factory-book-config-region { padding: 4px 6px !important; }
.batch-factory-workbench .batch-factory-book-config-region b { font-size: 11px; }
.batch-factory-workbench .batch-factory-book-config-region small { font-size: 10px; }
.batch-factory-workbench .batch-factory-action-group.is-utility,
.batch-factory-workbench .batch-factory-action-group.is-production,
.batch-factory-workbench .batch-factory-action-group.is-recovery {
  min-height: 26px;
  padding-block: 2px;
}
.batch-factory-workbench .batch-factory-action-button-grid .ant-btn,
.batch-factory-workbench .batch-factory-action-group > .ant-btn {
  height: 24px;
  min-height: 24px;
  padding-inline: 6px;
  font-size: 10px;
  line-height: 22px;
}
.batch-factory-workbench .batch-factory-text-page-cell,
.batch-factory-workbench .batch-factory-text-rendered-cell { display: grid; align-items: stretch; }
.batch-factory-workbench .batch-factory-text-page-cell > .batch-factory-cell-shell,
.batch-factory-workbench .batch-factory-text-rendered-cell > .batch-factory-cell-shell {
  align-self: stretch;
  height: 100% !important;
  grid-template-rows: 36px minmax(0, 1fr) 32px !important;
}
/* BF_COMPACT_ROW_CONTROLS_END */
```

- [ ] **Step 4: Run the focused test**

Run: `node --test src/user/pages/shuihuo/BatchFactoryCompactRows.source.test.js`

Expected: PASS, 1 test, 0 failures.

- [ ] **Step 5: Run nearby Batch Factory source tests**

Run: `node --test src/user/pages/shuihuo/InlineTextBookPages.source.test.js src/user/pages/shuihuo/batchFactoryWorkbenchLayout.source.test.js src/user/pages/shuihuo/BatchFactoryCompactRows.source.test.js`

Expected: PASS, 0 failures.

- [ ] **Step 6: Commit the scoped CSS and test**

```powershell
git add -- frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/BatchFactoryCompactRows.source.test.js
git commit -m "fix(batch-factory): compact row controls and align text pages"
```

### Task 2: Build, restart, and verify local 4001 geometry

**Files:**
- Verify only: `frontend/dist/`, `backend/web/dist/`, local embedded binary.

**Interfaces:**
- Consumes: Task 1 CSS contract.
- Produces: A local Go-embedded runtime serving the compact layout.

- [ ] **Step 1: Build the frontend**

Run: `npm run build` from `frontend`.

Expected: Vite exits 0; chunk-size warnings are allowed.

- [ ] **Step 2: Rebuild and restart only the confirmed local runtime**

Confirm the listener executable is `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`, stop that exact PID, then run:

```powershell
.\scripts\start-local-go-review.ps1 -Detached
```

Expected: a new PID listens on 4001.

- [ ] **Step 3: Verify runtime identity**

Run requests to `/health` and `/api/runtime-build-info`.

Expected: health is `true`, runtime is `go`, and `runtime_manifest_sha` contains the implementation commit.

- [ ] **Step 4: Verify the visible geometry in a real authenticated browser**

On the existing local acceptance batch, check computed geometry:

```js
const header = document.querySelector('.batch-factory-workbench-table .shuihuo-workbench-head');
const orderHeader = header.firstElementChild;
const action = document.querySelector('.batch-factory-book-content-action');
const copyPager = document.querySelector('.batch-factory-text-page-cell .batch-factory-cell-pager');
const imagePager = document.querySelector('.batch-factory-text-rendered-cell .batch-factory-cell-pager');
({
  headerHeight: header.getBoundingClientRect().height,
  orderFont: getComputedStyle(orderHeader).fontSize,
  actionHeight: action.getBoundingClientRect().height,
  pagerDelta: Math.abs(copyPager.getBoundingClientRect().bottom - imagePager.getBoundingClientRect().bottom)
});
```

Expected: `headerHeight === 40`, `orderFont === '8px'`, `actionHeight === 26`, and `pagerDelta <= 1`.

- [ ] **Step 5: Check all three production modes and themes without side effects**

Inspect one comic row, one video row, and the text acceptance row in dark and light themes. Confirm row buttons and cards are compact, titles/tags share a centerline, and no production, retry, upload, or paid model action is triggered.

- [ ] **Step 6: Remove the temporary browser session and authentication row**

Close the browser session and delete only the temporary `auth_sessions` record created for acceptance.
