# Batch Factory Production Mode Colors Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the unified-configuration production-mode switch an unmistakable blue video state and coral-red comic state.

**Architecture:** Reuse the normalized mode already computed by `BatchFactoryProductionModeControl` and expose it as a scoped modifier class. Style only the selected Ant Design segmented item inside this control, leaving its value flow and every other segmented control unchanged.

**Tech Stack:** React 18, Ant Design 5, CSS, Node test runner, Vite, Go embed runtime.

## Global Constraints

- Work directly on `v88`; do not create a branch or worktree.
- Preserve all unrelated dirty-worktree changes.
- Video selected state is solid blue with white text.
- Comic selected state is solid coral red with white text.
- Keep the existing compact dimensions and title placement.
- Do not change production-mode persistence or prompt behavior.

---

### Task 1: Lock The State Colors With A Failing Test

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`
- Read: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Read: `frontend/src/user/pages/shuihuo-production.css`

**Interfaces:**
- Consumes: source text for the shared production-mode control and its scoped stylesheet.
- Produces: regression assertions for `is-video`, `is-comic`, `#1677ff`, and `#e85d75`.

- [ ] **Step 1: Add the failing source assertion**

Add a test that requires the component class to include the normalized mode and requires scoped video/comic selected-item rules:

```js
test('production mode switch uses distinct visible colors for video and comic', () => {
  assert.match(unifiedSource, /batch-factory-production-mode-switch is-\$\{productionMode\}/);
  assert.match(stylesheet, /\.batch-factory-production-mode-switch\.is-video[\s\S]*#1677ff/);
  assert.match(stylesheet, /\.batch-factory-production-mode-switch\.is-comic[\s\S]*#e85d75/);
  assert.match(stylesheet, /\.ant-segmented-item-selected[\s\S]*color:\s*#fff/);
});
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
```

Expected: the new test fails because the component has no mode modifier and the stylesheet has no state colors.

---

### Task 2: Implement The Scoped Video And Comic Selected States

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo-production.css`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

**Interfaces:**
- Consumes: `productionMode` normalized to `video` or `comic`.
- Produces: `batch-factory-production-mode-switch is-video|is-comic` and scoped selected-item styling.

- [ ] **Step 1: Add the mode modifier**

Change the shared component class to:

```jsx
className={`batch-factory-production-mode-switch is-${productionMode}`}
```

- [ ] **Step 2: Add minimal scoped CSS**

Add a neutral border and scoped state colors:

```css
.batch-factory-production-mode-switch{margin-left:8px;flex:none;border:1px solid var(--border-strong);box-shadow:0 2px 8px rgba(0,0,0,.12)}
.batch-factory-production-mode-switch .ant-segmented-item-selected{color:#fff;font-weight:600}
.batch-factory-production-mode-switch.is-video .ant-segmented-item-selected{background:#1677ff;box-shadow:0 2px 7px rgba(22,119,255,.32)}
.batch-factory-production-mode-switch.is-comic .ant-segmented-item-selected{background:#e85d75;box-shadow:0 2px 7px rgba(232,93,117,.32)}
```

- [ ] **Step 3: Run the focused test and verify GREEN**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
```

Expected: all focused tests pass.

- [ ] **Step 4: Build the frontend**

Run:

```powershell
cd frontend
npm run build
```

Expected: Vite exits with code 0.

- [ ] **Step 5: Commit the source change**

```powershell
git add frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
git commit -m "style(batch-factory): emphasize production mode switch"
```

---

### Task 3: Rebuild And Verify The Live Go Runtime

**Files:**
- Verify generated frontend output without committing it.
- Verify `backend/qiantie-local.exe` and the local page.

**Interfaces:**
- Consumes: the committed React and CSS change.
- Produces: a local Go runtime whose build info reports the new commit SHA.

- [ ] **Step 1: Restart only the verified port-4001 Go runtime**

Confirm the listener path is `backend/qiantie-local.exe`, stop that PID only, then run:

```powershell
powershell -File scripts/start-local-go-review.ps1 -UseYcmMysql
```

- [ ] **Step 2: Verify runtime identity**

Confirm `/health` returns `ok: true` and `/api/runtime-build-info` reports the new commit SHA for both `git_sha` and `runtime_manifest_sha`.

- [ ] **Step 3: Verify both colors in the browser**

Open the existing batch unified configuration and confirm:

1. `视频生成` selected background computes to `rgb(22, 119, 255)` and text to white.
2. Switching the draft to `漫画生成` changes the selected background to `rgb(232, 93, 117)` and keeps text white.
3. The control size and title placement remain stable.
4. Cancel without saving or generating.

Repeat the state-color check in `新建批量 · 统一配置`, then cancel both dialogs.
