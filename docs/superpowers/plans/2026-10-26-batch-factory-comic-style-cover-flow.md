# Batch Factory Comic Style Cover Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the multi-row comic style grid with an accessible single-row Cover Flow whose centered card is the selected draft style.

**Architecture:** Keep the existing `presetId` configuration contract and reference-image separation. Add pure cyclic-index helpers in `batchFactoryComicStyle.js`, use them from the existing picker component for pointer, wheel, button, card, and keyboard navigation, and implement the 3D stage entirely with local CSS.

**Tech Stack:** React 18, Ant Design, lucide-react, CSS 3D transforms, Node test runner, Vite, Go embed runtime.

## Global Constraints

- The centered card is the selected style and updates only the current modal draft.
- Persist only through the existing `保存当前书覆盖` or `保存统一配置` action.
- Keep style selection separate from `设为漫画全局参考图`.
- Keep every preview at 3:4 and do not add a carousel dependency.
- Support reduced motion and mobile swipe behavior.
- Modify and commit only files belonging to this feature.

---

### Task 1: Cyclic Cover Flow Model

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.js`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

**Interfaces:**
- Produces: `wrapComicStyleIndex(index, count): number`
- Produces: `comicStyleRelativeOffset(index, activeIndex, count): number`
- Produces: `comicStyleIndexFromDrag(activeIndex, deltaX, count, threshold = 36): number`

- [ ] **Step 1: Write failing tests for wrapping, shortest relative offsets, and drag direction**

```js
assert.equal(wrapComicStyleIndex(-1, 5), 4);
assert.equal(wrapComicStyleIndex(5, 5), 0);
assert.equal(comicStyleRelativeOffset(4, 0, 5), -1);
assert.equal(comicStyleRelativeOffset(0, 4, 5), 1);
assert.equal(comicStyleIndexFromDrag(2, -50, 5), 3);
assert.equal(comicStyleIndexFromDrag(2, 50, 5), 1);
assert.equal(comicStyleIndexFromDrag(2, 20, 5), 2);
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

Expected: FAIL because the three Cover Flow helpers are not exported.

- [ ] **Step 3: Implement the pure helpers**

```js
export function wrapComicStyleIndex(index, count) {
  if (!Number.isFinite(count) || count <= 0) return 0;
  return ((Number(index) % count) + count) % count;
}

export function comicStyleRelativeOffset(index, activeIndex, count) {
  if (!Number.isFinite(count) || count <= 0) return 0;
  let offset = wrapComicStyleIndex(index, count) - wrapComicStyleIndex(activeIndex, count);
  if (offset > count / 2) offset -= count;
  if (offset < -count / 2) offset += count;
  return offset;
}

export function comicStyleIndexFromDrag(activeIndex, deltaX, count, threshold = 36) {
  if (Math.abs(deltaX) < threshold) return wrapComicStyleIndex(activeIndex, count);
  return wrapComicStyleIndex(activeIndex + (deltaX < 0 ? 1 : -1), count);
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

Expected: all tests pass.

### Task 2: Accessible Cover Flow Picker

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryComicStylePicker.jsx`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

**Interfaces:**
- Consumes: `wrapComicStyleIndex`, `comicStyleRelativeOffset`, `comicStyleIndexFromDrag`
- Preserves: `onChange({ ...comic, style: comicStyleSelection(preset), enabled: true })`

- [ ] **Step 1: Add failing source assertions for the required controls and selection text**

```js
const component = fs.readFileSync(new URL('./BatchFactoryComicStylePicker.jsx', import.meta.url), 'utf8');
assert.match(component, /aria-label="上一画风"/);
assert.match(component, /aria-label="下一画风"/);
assert.match(component, /当前画风：/);
assert.match(component, /已选择/);
assert.match(component, /onPointerDown/);
assert.match(component, /onWheel/);
assert.match(component, /onKeyDown/);
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

Expected: FAIL because the current picker is a static multi-row grid.

- [ ] **Step 3: Replace the grid markup with the Cover Flow stage**

Implementation requirements:

```jsx
<div className="batch-factory-comic-coverflow" role="radiogroup" aria-label="漫画画风库" tabIndex={0} onKeyDown={handleKeyDown} onWheel={handleWheel}>
  <button type="button" aria-label="上一画风" onClick={() => move(-1)}><ChevronLeft /></button>
  <div className="batch-factory-comic-coverflow-stage" onPointerDown={startDrag} onPointerUp={finishDrag}>
    {styles.map((preset, index) => <article className={`batch-factory-comic-style-card${selected ? ' is-selected' : ''}`} style={coverFlowStyle(offset)}>
      <button type="button" role="radio" aria-checked={selected} onClick={() => selectStyle(preset)}>...</button>
      {selected ? <span className="batch-factory-comic-style-selected"><Check />已选择</span> : null}
    </article>)}
  </div>
  <button type="button" aria-label="下一画风" onClick={() => move(1)}><ChevronRight /></button>
</div>
```

Use the selected preset for the background preview, heading text, description, and existing `设为漫画全局参考图` action. Throttle wheel navigation so one gesture does not skip many styles.

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

Expected: all tests pass.

### Task 3: 3D Layout, Build, and Runtime Acceptance

**Files:**
- Modify: `frontend/src/user/pages/shuihuo-production.css`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

**Interfaces:**
- Consumes component classes from Task 2.
- Produces a bounded 3D stage that does not change the modal data contract.

- [ ] **Step 1: Add failing CSS assertions**

```js
assert.match(stylesheet, /\.batch-factory-comic-coverflow-stage\s*\{[^}]*perspective:/s);
assert.match(stylesheet, /\.batch-factory-comic-style-card\s*\{[^}]*position:\s*absolute/s);
assert.match(stylesheet, /@media\s*\(prefers-reduced-motion:\s*reduce\)/s);
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js`

Expected: FAIL because Cover Flow CSS is absent.

- [ ] **Step 3: Implement bounded Cover Flow styles**

Use `perspective`, `transform-style: preserve-3d`, absolute 3:4 cards, a clipped stage, blue selected outline, selected badge, responsive card sizing, and reduced-motion overrides. Keep card radius at 8px or less and preserve readable dark text on the white metadata area.

- [ ] **Step 4: Run tests and production build**

Run:

```powershell
node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js
npm run build
```

Expected: focused tests pass and Vite exits 0.

- [ ] **Step 5: Commit only feature files**

```powershell
git add -- frontend/src/user/pages/shuihuo/batchFactoryComicStyle.js frontend/src/user/pages/shuihuo/batchFactoryComicStyle.test.js frontend/src/user/pages/shuihuo/BatchFactoryComicStylePicker.jsx frontend/src/user/pages/shuihuo-production.css
git commit -m "feat(batch-factory): add comic style cover flow"
```

- [ ] **Step 6: Rebuild and restart the local Go runtime**

Run:

```powershell
scripts/build-go-local.ps1
scripts/start-local-go-review.ps1 -UseLocalMysql -ListenAddr :4001
```

Expected: `/health` returns `ok: true` and `/api/runtime-build-info` reports the new commit SHA.

- [ ] **Step 7: Browser acceptance**

Verify both unified configuration and single-book constraints: selected card is centered, blue, checked, and titled; card click, drag, wheel, arrows, and keyboard change the draft; background and title transition; reference-image action remains separate; modal width stays bounded at desktop and mobile sizes.
