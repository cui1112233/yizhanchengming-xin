# Storyboard Resizable Material Grid Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users resize the two columns and two rows of the storyboard project-material grid by dragging its central splitters.

**Architecture:** Keep the split ratios as local `AgentStoryboardWorkspace` state so they apply immediately without changing project content or saved material order. A small pure helper normalizes pointer positions into bounded percentages; JSX renders two pointer-captured splitter controls, and CSS consumes the ratios through custom properties.

**Tech Stack:** React, Pointer Events, CSS Grid custom properties, Node built-in test runner.

## Global Constraints

- Only the storyboard project's four material sections are resizable.
- Preserve all material filtering, batch actions, sorting, uploads and inspector behavior.
- Default ratios remain 39% left / 61% right and 50% top / 50% bottom.
- Each row and column keeps a minimum usable size.
- At narrow mobile widths, retain the existing vertical material list and hide splitter controls.
- Do not create a branch, reset the dirty worktree, or publish this local change.

---

### Task 1: Add pure splitter ratio calculations

**Files:**
- Create: `frontend/src/user/pages/agent-storyboard-split.js`
- Test: `frontend/src/user/pages/agent-storyboard-split.test.js`

**Interfaces:**
- Produces `STORYBOARD_SPLIT_DEFAULTS`, `storyboardSplitRatio(value, minimum)`, and `storyboardSplitRatioFromPointer(pointer, bounds, axis)`.
- `AgentStoryboardWorkspace` consumes the helpers to convert client pointer positions into CSS percentage values.

- [ ] **Step 1: Write the failing test**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import { STORYBOARD_SPLIT_DEFAULTS, storyboardSplitRatio, storyboardSplitRatioFromPointer } from './agent-storyboard-split.js';

test('storyboard split ratios keep defaults and usable bounds', () => {
  assert.deepEqual(STORYBOARD_SPLIT_DEFAULTS, { column: 39, row: 50 });
  assert.equal(storyboardSplitRatio(-4, 24), 24);
  assert.equal(storyboardSplitRatio(98, 24), 76);
  assert.equal(storyboardSplitRatioFromPointer(200, { left: 0, width: 400 }, 'x'), 50);
  assert.equal(storyboardSplitRatioFromPointer(120, { top: 0, height: 400 }, 'y'), 30);
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-storyboard-split.test.js`

Expected: FAIL because the splitter module does not exist.

- [ ] **Step 3: Write the minimal implementation**

```js
export const STORYBOARD_SPLIT_DEFAULTS = Object.freeze({ column: 39, row: 50 });

export function storyboardSplitRatio(value, minimum = 24) {
  const safeMinimum = Math.min(45, Math.max(10, Number(minimum) || 24));
  const numeric = Number(value);
  const candidate = Number.isFinite(numeric) ? numeric : 50;
  return Math.round(Math.max(safeMinimum, Math.min(100 - safeMinimum, candidate)));
}

export function storyboardSplitRatioFromPointer(pointer, bounds, axis) {
  const start = axis === 'x' ? Number(bounds?.left) : Number(bounds?.top);
  const size = axis === 'x' ? Number(bounds?.width) : Number(bounds?.height);
  if (!Number.isFinite(start) || !Number.isFinite(size) || size <= 0) return 50;
  return storyboardSplitRatio(((Number(pointer) - start) / size) * 100);
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `node --test frontend/src/user/pages/agent-storyboard-split.test.js`

Expected: PASS.

### Task 2: Connect splitters to the storyboard grid

**Files:**
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-storyboard-workspace.source.test.js`

**Interfaces:**
- Consumes `storyboardSplitRatioFromPointer(pointer, bounds, axis)`.
- Produces two accessible splitter controls inside `.agent-storyboard-library__body`.

- [ ] **Step 1: Write the failing source contract test**

```js
assert.match(source, /agent-storyboard-grid-splitter--column/);
assert.match(source, /agent-storyboard-grid-splitter--row/);
assert.match(source, /onPointerDown=\{startGridResize\}/);
assert.match(source, /storyboardSplitRatioFromPointer/);
```

- [ ] **Step 2: Run the source contract test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-storyboard-workspace.source.test.js`

Expected: FAIL because no grid splitter controls exist.

- [ ] **Step 3: Add local pointer-resize state and markup**

```jsx
const [gridSplit, setGridSplit] = useState(STORYBOARD_SPLIT_DEFAULTS);
const gridResizeRef = useRef(null);

function startGridResize(event) {
  const axis = event.currentTarget.dataset.axis;
  const bounds = event.currentTarget.parentElement?.getBoundingClientRect();
  if (event.button !== 0 || !axis || !bounds) return;
  gridResizeRef.current = { axis, pointerId: event.pointerId, bounds };
  event.currentTarget.setPointerCapture?.(event.pointerId);
}
```

Render the column and row splitter buttons after the material sections. On pointer move, update only `gridSplit.column` or `gridSplit.row`; on pointer up/cancel, release capture and clear `gridResizeRef`.

- [ ] **Step 4: Drive the CSS grid using custom properties**

```css
.agent-page .agent-storyboard-library__body {
  grid-template-columns: minmax(170px, var(--storyboard-column-split, 39%)) minmax(0, 1fr);
  grid-template-rows: minmax(0, var(--storyboard-row-split, 50%)) minmax(0, 1fr);
}
.agent-page .agent-storyboard-grid-splitter--column { cursor: col-resize; }
.agent-page .agent-storyboard-grid-splitter--row { cursor: row-resize; }
```

Use a transparent 10px hit target aligned to the grid boundary, show the accent while resizing, and hide splitters under the existing `680px` media query.

- [ ] **Step 5: Run focused tests**

Run: `node --test frontend/src/user/pages/agent-storyboard-split.test.js frontend/src/user/pages/agent-storyboard-state.test.js frontend/src/user/pages/agent-storyboard-workspace.source.test.js`

Expected: PASS.

### Task 3: Build and verify the local UI

**Files:**
- Modify: generated local runtime only through the approved build script; do not edit generated assets as source.

**Interfaces:**
- Consumes the built React frontend through the local Go embedded runtime.
- Produces a rendered storyboard page where both splitters resize the intended axes.

- [ ] **Step 1: Build and start the local Go review runtime**

Run: `scripts/build-go-local.ps1` then `scripts/start-local-go-review.ps1 -UseYcmMysql`.

Expected: local service responds at `http://localhost:4001`.

- [ ] **Step 2: Verify the desktop interaction**

Open: `/agent/canvas?taskId=agent-381900adb37b8663b49729b2&view=storyboard`.

Drag the vertical splitter right and the horizontal splitter down. Confirm the image/video areas shrink or grow with the respective axis, material item selection still works, and the details panel still opens.

- [ ] **Step 3: Verify narrow layout**

At a width below 680px, confirm the material sections render as the existing vertical list and neither desktop splitter is present.
