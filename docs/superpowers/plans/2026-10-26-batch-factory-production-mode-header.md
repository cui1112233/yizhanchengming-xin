# Batch Factory Production Mode Header Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the mutually exclusive video/comic production-mode control into the unified-config title and make AI reasoning show only the selected mode's settings.

**Architecture:** Keep `productionMode` in the existing unified draft patch and reuse `batchFactoryProductionModeVisibility`. Remove the separate media tab, place one Ant Design `Segmented` control in the modal title, and render comic prompt/image-derivation fields inside the existing AI reasoning form without clearing inactive mode values.

**Tech Stack:** React 18, Ant Design 5, Node test runner, Vite, Go embed runtime.

## Global Constraints

- Work directly on Git branch `v88`; do not create a branch or worktree.
- System prompt bodies remain server-side; the frontend only selects published preset IDs.
- Remove the separate `画面与视频` tab and never restore the legacy `画面提示词` selector.
- Switching modes updates only the modal draft; saving persists it and cancelling discards it.
- Preserve saved settings for the inactive mode.

---

### Task 1: Lock The New Unified-Config Layout With Tests

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

**Interfaces:**
- Consumes: `BatchFactoryUnifiedSettingsModal.jsx` source text.
- Produces: regression assertions for the title switch, removed media tab, and comic-only fields.

- [ ] **Step 1: Write failing source assertions**

Add assertions equivalent to:

```js
assert.match(unifiedSource, /<Segmented[^>]+options=\{\[\{ value: 'video', label: '视频生成' \}, \{ value: 'comic', label: '漫画生成' \}\]\}/);
assert.doesNotMatch(unifiedSource, /key: 'media', label: '画面与视频'/);
assert.match(unifiedSource, /visibility\.comic \? <section[^>]*><header><b>漫画生成<\/b>/);
assert.match(unifiedSource, /<b>漫画提示词<\/b>/);
assert.match(unifiedSource, /<b>图片描述词推导<\/b>/);
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```powershell
cd frontend
node --test src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
```

Expected: the new layout assertions fail because the media tab and its mode cards still exist.

---

### Task 2: Move Production Mode And Comic Settings

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

**Interfaces:**
- Consumes: `batchFactoryProductionModeVisibility(value)` and the existing `aiPromptConfig.comic` shape.
- Produces: a title-level `Segmented` control that writes `draftPatch.productionMode`, plus mode-specific AI settings.

- [ ] **Step 1: Add the title production-mode control**

Use the normalized mode and update only the draft:

```jsx
const productionMode = batchFactoryProductionModeVisibility(draftPatch.productionMode).mode;
<Segmented
  value={productionMode}
  options={[{ value: 'video', label: '视频生成' }, { value: 'comic', label: '漫画生成' }]}
  onChange={productionMode => setDraftPatch(current => ({ ...current, productionMode }))}
/>
```

- [ ] **Step 2: Remove the separate media tab**

Delete this tab item:

```jsx
{ key: 'media', label: '画面与视频', children: <BatchFactoryMediaSettingsForm ... /> }
```

Keep `模型配置`, `AI 推理`, and `发布统一`.

- [ ] **Step 3: Add comic-only generation settings to AI reasoning**

Inside `BatchFactoryAiReasoningForm`, derive the catalog options and update the existing comic object:

```jsx
const comicPageRules = catalog.filter(item => item.slot === 'batch.comic-page');
const imageDerivationRules = catalog.filter(item => item.slot === 'batch.comic-image-derivation');
const updateComic = next => {
  const nextComic = { ...comic, ...next };
  onChange({ ...config, comic: { ...nextComic, enabled: comicEnabled(nextComic) } });
};
```

Render only when `visibility.comic`:

```jsx
<section className="batch-factory-engine-card">
  <header><b>漫画生成</b></header>
  <div className="batch-factory-engine-card-body batch-factory-engine-grid">
    <Select value={comic.presetId || undefined} options={comicPageRules.map(selectOption)} />
    <Select value={comic.imageDerivation?.presetId || undefined} options={imageDerivationRules.map(selectOption)} />
  </div>
</section>
```

Selection handlers must store full preset metadata through `presetValue` and must not modify `config.video`.

- [ ] **Step 4: Run focused tests and verify GREEN**

```powershell
cd frontend
node --test src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
```

Expected: all focused tests pass.

- [ ] **Step 5: Commit the source and test**

```powershell
git add frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
git commit -m "feat(batch-factory): move production mode into unified header"
```

---

### Task 3: Build And Live Acceptance

**Files:**
- Verify generated frontend output without committing it.
- Verify local Go embedded runtime.

**Interfaces:**
- Consumes: committed React source.
- Produces: a live local page on `http://localhost:4001/batch-factory` serving the committed SHA.

- [ ] **Step 1: Build the production frontend**

```powershell
cd frontend
npm run build
```

Expected: Vite exits with code 0.

- [ ] **Step 2: Restart the local Go review runtime**

Stop only the verified `backend/qiantie-local.exe` listener on port 4001, then run:

```powershell
powershell -File scripts/start-local-go-review.ps1 -UseYcmMysql
```

- [ ] **Step 3: Verify runtime identity and browser behavior**

Confirm `/health` is OK and `/api/runtime-build-info` reports the committed SHA. In the browser:

1. Open unified configuration.
2. Confirm the title shows `视频生成 | 漫画生成`.
3. Confirm the `画面与视频` tab is absent.
4. Switch the draft to comic and confirm only comic AI settings appear.
5. Switch the draft to video and confirm only video AI settings appear.
6. Confirm `画面提示词` never appears.
7. Cancel without saving or triggering generation.

- [ ] **Step 4: Record verification evidence**

Report the focused test count, Vite build result, runtime SHA, and browser-visible mode-specific fields.
