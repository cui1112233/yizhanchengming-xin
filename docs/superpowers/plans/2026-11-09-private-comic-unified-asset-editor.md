# Private Comic Unified Asset Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every asset in the single-book comic workspace open the same shared asset editor shell already used by Script and Batch Factory, including editable fields and existing image versions.

**Architecture:** Keep `PrivateStudioAssetEditor` as the only asset-editor shell. `PrivateComicWorkspace` owns the editor presentation state, while `ScriptPage` owns API mutations and reloads the canonical Batch Factory asset data after every successful action. No backend endpoint or duplicate modal/CSS is introduced.

**Tech Stack:** React 18, Ant Design 5, Node test runner, existing Batch Factory V12 APIs.

## Global Constraints

- Reuse `PrivateStudioAssetEditor`; do not create a second asset modal shell.
- Reuse `comicAssets` and `comicAssetImages` for display.
- Reuse existing `updateBookAsset`, `uploadBookAssetImage`, `setPrimaryBookAssetImage`, and asset-image generation APIs.
- Merely opening or closing the editor must not save data or call a model.
- Keep Script asset fields separate from Comic asset fields.
- Verify locally at `http://localhost:4001/script`; do not deploy online.

---

### Task 1: Lock the unified-editor contract with source tests

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.source.test.js`
- Modify: `frontend/src/user/pages/private-studio/PrivateStudioAssetEditor.source.test.js`

**Interfaces:**
- Consumes: `PrivateStudioAssetEditor` and `PrivateComicWorkspace` source.
- Produces: regression assertions that reject the old passive preview modal and require the shared editor plus callback wiring.

- [ ] **Step 1: Write the failing tests**

Add assertions equivalent to:

```js
assert.match(source, /import \{ PrivateStudioAssetEditor \}/);
assert.match(source, /<button[^>]+private-comic-asset-name/);
assert.match(source, /<PrivateStudioAssetEditor/);
assert.match(source, /onSaveAsset/);
assert.match(source, /onUploadAssetImage/);
assert.match(source, /onSetPrimaryAssetImage/);
assert.doesNotMatch(source, /title=\{assetPreview \? `\$\{assetPreview\.title\}/);
```

Extend the shared-shell test so all three consumers are required:

```js
assert.match(comicPage, /import \{ PrivateStudioAssetEditor \}/);
assert.match(comicPage, /<PrivateStudioAssetEditor/);
```

- [ ] **Step 2: Run tests and verify RED**

Run:

```powershell
node --test src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/private-studio/PrivateStudioAssetEditor.source.test.js
```

Expected: failure because `PrivateComicWorkspace` still renders plain spans and the old list-only modal.

- [ ] **Step 3: Commit the failing contract test**

```powershell
git add frontend/src/user/pages/PrivateComicWorkspace.source.test.js frontend/src/user/pages/private-studio/PrivateStudioAssetEditor.source.test.js
git commit -m "test(script): require unified comic asset editor"
```

### Task 2: Add canonical asset mutation callbacks

**Files:**
- Modify: `frontend/src/user/pages/ScriptPage.jsx`

**Interfaces:**
- Produces:
  - `savePrivateComicAsset(asset, fields): Promise<boolean>`
  - `uploadPrivateComicAssetImage(asset, file): Promise<boolean>`
  - `setPrivateComicPrimaryAssetImage(asset, image): Promise<boolean>`
- Each successful mutation calls `loadPrivateComicProductionData(batchId, bookId)` so the single-book page and Batch Factory read the same persisted records.

- [ ] **Step 1: Implement the minimal callbacks**

Import the existing API functions and wire handlers into `PrivateComicWorkspace`:

```jsx
onSaveAsset={savePrivateComicAsset}
onUploadAssetImage={uploadPrivateComicAssetImage}
onSetPrimaryAssetImage={setPrivateComicPrimaryAssetImage}
```

Use `expectedRevision` from the selected asset when saving, validate PNG/JPG/WEBP before upload, and report API errors through the existing `privateComicError` state.

- [ ] **Step 2: Keep image generation on the existing path**

Continue passing `onGenerateAssetImage`; do not add a second generation function or polling loop.

### Task 3: Replace the passive asset preview with the shared editor

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Modify: existing private-studio or Batch Factory CSS only if the shared classes need a small compatibility rule; do not create a new modal stylesheet.

**Interfaces:**
- Consumes callbacks from Task 2.
- Produces clickable asset buttons and a `PrivateStudioAssetEditor` instance for the selected asset.

- [ ] **Step 1: Make compact names and full-list rows clickable**

Use real buttons:

```jsx
<button type="button" className="private-comic-asset-name" onClick={() => setEditingAsset({ ...asset })}>
  {asset.name || '未命名资产'}
</button>
```

The “查看全部” modal remains only a chooser; clicking a row closes it and opens the shared editor.

- [ ] **Step 2: Render the shared editor**

Map existing Comic fields into the same sections used by Batch Factory:

```jsx
<PrivateStudioAssetEditor
  open={Boolean(editingAsset)}
  title={`编辑资产 · ${editingAsset?.name || '未命名资产'}`}
  basic={/* category, image model, aspect ratio, name */}
  settings={/* character facts, relationships, appearance; empty note for scene/prop */}
  prompts={/* full prompt textarea */}
  aside={/* primary image, upload, AI generation, all versions */}
  footer={/* cancel and save */}
/>
```

- [ ] **Step 3: Preserve explicit-action semantics**

Opening/closing only changes local `editingAsset`. Save, upload, set-primary, and AI-generate call their corresponding callback only from their explicit controls.

- [ ] **Step 4: Run focused tests and verify GREEN**

```powershell
node --test src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/private-studio/PrivateStudioAssetEditor.source.test.js
```

Expected: both files pass.

### Task 4: Build and visible local acceptance

**Files:**
- Modify generated embedded assets only through the repository's established frontend build/embed workflow.

**Interfaces:**
- Consumes completed source changes.
- Produces a locally running `localhost:4001` UI showing the shared editor.

- [ ] **Step 1: Run the full frontend test suite**

```powershell
npm test
```

Expected: exit code 0 and no new failures.

- [ ] **Step 2: Build the frontend**

```powershell
npm run build
```

Expected: exit code 0.

- [ ] **Step 3: Rebuild/restart the confirmed local Go runtime**

Use the repository's existing local review script, then verify `/health` and `/api/runtime-build-info` identify the new local build.

- [ ] **Step 4: Perform visible browser acceptance**

At `http://localhost:4001/script`:

1. Open 漫画生成.
2. Click an asset name and confirm the shared editor sections appear.
3. Confirm its prompt and existing image versions are visible.
4. Close without saving and confirm no mutation/model request occurs.
5. Open one asset from “查看全部” and confirm identical behavior.

- [ ] **Step 5: Commit implementation**

```powershell
git add frontend/src/user/pages/ScriptPage.jsx frontend/src/user/pages/PrivateComicWorkspace.jsx frontend/src/user/pages/PrivateComicWorkspace.source.test.js frontend/src/user/pages/private-studio/PrivateStudioAssetEditor.source.test.js
git commit -m "feat(script): open comic assets in unified editor"
```
