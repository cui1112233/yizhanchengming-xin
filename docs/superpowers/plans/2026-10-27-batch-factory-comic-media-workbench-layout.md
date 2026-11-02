# Batch Factory Comic Media Workbench Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Batch Factory comic media modal so the left side only displays the active primary comic and every page, reference, collection, and regeneration control lives in a fixed right-side workbench.

**Architecture:** Keep `ComicMediaPanel` as the owner of existing page, reference, collection, and publish state. Recompose its JSX into a left preview plus a right four-row grid (header, tabs, scrollable body, footer), and pass the existing `generateBookComicPage` function into the panel for current-page regeneration. No backend contract changes are required.

**Tech Stack:** React 18, Ant Design 5, existing Batch Factory V11 API helpers, CSS, Node test runner, Vite, Go embed runtime.

## Global Constraints

- The left column contains only the active primary comic or its empty/error state.
- The right column contains the work title, page status, comic-page and comic-collection tabs, candidates, page references, source content, collection generation, page navigation, and current-page regeneration.
- `上传参考图`, `漫画合集`, and `生成漫画合集` must be visible without opening a secondary menu.
- The right header, tabs, and footer stay fixed; only the right middle body scrolls.
- Existing page, reference, collection, and publish APIs remain unchanged.
- Current-page regeneration reuses `generateBookComicPage`; do not add a backend route.
- Do not trigger real image generation, collection generation, or publishing during browser verification.
- Preserve unrelated dirty worktree changes; do not commit, branch, push, reset, or clean.

---

### Task 1: Lock the approved workbench contract with a failing source test

**Files:**
- Create: `frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`
- Read: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Read: `frontend/src/user/pages/shuihuo-production.css`

**Interfaces:**
- Consumes: `ComicMediaPanel`, existing `makeCollection`, `uploadPageReference`, `ComicSourceContent`, and `generateBookComicPage`.
- Produces: source-level regression coverage for the approved DOM ownership and fixed-region CSS contract.

- [ ] **Step 1: Write the failing source test**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const jsx = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
const css = readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
const panel = jsx.slice(jsx.indexOf('function ComicMediaPanel'), jsx.indexOf('function MediaVersionPanel'));

test('comic media workbench keeps the primary comic alone on the left', () => {
  assert.match(panel, /batch-factory-comic-media-workbench/);
  assert.match(panel, /batch-factory-comic-media-preview[\s\S]*AuthenticatedComicPageImage/);
  assert.doesNotMatch(panel.match(/batch-factory-comic-media-preview[\s\S]*?batch-factory-comic-media-tools/)?.[0] || '', /上传参考图|漫画合集|ComicSourceContent/);
});

test('comic media controls live in the fixed right workbench', () => {
  assert.match(panel, /batch-factory-comic-media-tools[\s\S]*生成漫画合集[\s\S]*漫画页[\s\S]*漫画合集/);
  assert.match(panel, /batch-factory-comic-media-scroll[\s\S]*候选图[\s\S]*上传参考图[\s\S]*ComicSourceContent/);
  assert.match(panel, /batch-factory-comic-media-footer[\s\S]*重新生成本页/);
  assert.match(css, /\.batch-factory-comic-media-tools\{[^}]*grid-template-rows:[^}]*minmax\(0,1fr\)[^}]*\}/);
  assert.match(css, /\.batch-factory-comic-media-scroll\{[^}]*overflow-y:auto/);
});

test('current-page regeneration reuses the existing comic page generator', () => {
  assert.match(panel, /onGeneratePage/);
  assert.match(panel, /onGeneratePage\?\.\(selectedVideoId/);
  assert.match(jsx, /<ComicMediaPanel[\s\S]*onGeneratePage=\{[\s\S]*generateBookComicPage/);
});
```

- [ ] **Step 2: Run the test and verify the intended RED failure**

Run: `node --test frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

Expected: FAIL because the approved workbench class names and `onGeneratePage` callback do not exist yet.

---

### Task 2: Recompose `ComicMediaPanel` into the approved two-column workbench

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx:2198-2349`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx:4718`
- Test: `frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

**Interfaces:**
- Consumes: `onGeneratePage(videoId: string, prompt: string) => Promise<void>` supplied by the parent.
- Produces: `ComicMediaPanel({ batchId, book, comicPublishSettings, initialVideoId, onSaved, onGeneratePage, generatingPage })`.
- Preserves: `choosePrimary`, `uploadPageReference`, `applyPageReferenceToAll`, `savePageReferences`, `makeCollection`, `openComicPublish`, and `ComicSourceContent` behavior.

- [ ] **Step 1: Add the current-page generation callback to the panel contract**

Change the component signature to accept `onGeneratePage` and `generatingPage`. Add a handler that exits without a selected video, then calls:

```js
await onGeneratePage?.(
  selectedVideoId,
  selectedVideo?.comicImagePrompt || selectedVideo?.comicPrompt || ''
);
```

The handler must not call `load()` optimistically; the parent generation path and existing batch refresh remain the source of truth.

- [ ] **Step 2: Replace the outer tab-owned layout with one workbench shell**

Render this ownership order:

```jsx
<div className="batch-factory-comic-media-workbench">
  <section className="batch-factory-comic-media-preview">...</section>
  <aside className="batch-factory-comic-media-tools">
    <header className="batch-factory-comic-media-header">...</header>
    <Tabs className="batch-factory-comic-media-tabs" ... />
    <div className="batch-factory-comic-media-scroll">...</div>
    <footer className="batch-factory-comic-media-footer">...</footer>
  </aside>
</div>
```

The preview section contains only the primary image or the existing empty state. Move the work name, page label, status tag, and primary `生成漫画合集` button into the right header.

- [ ] **Step 3: Build the comic-page scroll body**

Inside the active page tab body, render in this exact order:

1. A four-column candidate grid with selected-primary state.
2. The existing page-reference block with visible `上传参考图`, `应用到全部`, and conditional `清空` actions.
3. `ComicSourceContent` in embedded mode.

Keep candidate selection, reference loading states, and source splitting wired to the existing handlers.

- [ ] **Step 4: Keep collection content in the same right scroll body**

Use controlled tab state inside `ComicMediaPanel`. The `漫画合集` tab shows the existing collection list and publish controls. Keep `生成漫画合集` in the fixed right header so it remains visible on either tab; remove the duplicate button from the collection content.

- [ ] **Step 5: Add the fixed footer**

The footer contains previous page, `NN / total`, next page, and `重新生成本页`. Disable page navigation at the ends. Disable regeneration when no current video, another panel operation is busy, or `generatingPage` is true.

- [ ] **Step 6: Pass the existing generator from the parent**

At the `ComicMediaPanel` call site, pass:

```jsx
onGeneratePage={(videoId, prompt) => generateBookComicPage(mediaBook, videoId, prompt)}
generatingPage={String(actionBusy).startsWith(`comic-image-${mediaBook.id}-`)}
```

- [ ] **Step 7: Run the focused test and verify GREEN**

Run: `node --test frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js frontend/src/user/pages/shuihuo/comicSourceContent.test.js`

Expected: all tests PASS.

---

### Task 3: Implement the fixed two-column visual system

**Files:**
- Modify: `frontend/src/user/pages/shuihuo-production.css:4130-4140`
- Test: `frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

**Interfaces:**
- Consumes: the `batch-factory-comic-media-*` class structure from Task 2.
- Produces: desktop 59/41 preview-to-tools ratio, fixed right regions, compact four-column candidates, and responsive stacking below 900px.

- [ ] **Step 1: Replace the old viewer/sidebar rules**

Define the workbench as a bounded grid with `height:min(760px, calc(100vh - 150px))`, `min-height:620px`, and `grid-template-columns:minmax(0,59fr) minmax(430px,41fr)`. Give the preview a neutral dark canvas and center the image with `object-fit:contain`.

- [ ] **Step 2: Lock the right panel rows**

Set `.batch-factory-comic-media-tools` to `grid-template-rows:auto auto minmax(0,1fr) auto`, `min-height:0`, and `overflow:hidden`. Set `.batch-factory-comic-media-scroll` to `min-height:0; overflow-y:auto`.

- [ ] **Step 3: Style the compact controls**

Use a restrained neutral palette, one blue selection/primary accent, green only for ready/success status, 8px maximum radius, thin separators, stable candidate aspect ratios, and no nested floating cards. Candidate thumbnails stay in one four-column row on desktop and remain visibly smaller than the primary comic.

- [ ] **Step 4: Add responsive behavior**

Below 900px, stack preview above tools, keep the preview at a stable minimum height, and let the right panel own its internal scroll. Text buttons must wrap instead of overflowing.

- [ ] **Step 5: Rerun the focused source test**

Run: `node --test frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js frontend/src/user/pages/shuihuo/comicSourceContent.test.js`

Expected: all tests PASS.

---

### Task 4: Build and verify the embedded local runtime

**Files:**
- Modify only files required by failures caused by Tasks 1-3.

**Interfaces:**
- Produces: the approved layout in the Go-embedded local runtime at `http://localhost:4001/batch-factory`.

- [ ] **Step 1: Run focused frontend tests**

Run: `node --test frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js frontend/src/user/pages/shuihuo/comicSourceContent.test.js`

- [ ] **Step 2: Check only this feature's diff**

Run: `git diff --check -- frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js docs/superpowers/specs/2026-10-27-batch-factory-comic-media-workbench-layout-design.md docs/superpowers/plans/2026-10-27-batch-factory-comic-media-workbench-layout.md`

- [ ] **Step 3: Build the frontend**

Run: `npm --prefix frontend run build`

Expected: exit code 0.

- [ ] **Step 4: Rebuild the Go-embedded binary**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Expected: `backend/qiantie-local.exe` is rebuilt with the new `backend/web/dist` assets.

- [ ] **Step 5: Restart only the local review process**

Resolve the PID listening on port 4001 and verify its executable path is `backend/qiantie-local.exe`. Stop only that PID, then start the configured local review service with the existing project script.

- [ ] **Step 6: Verify health and build identity**

Run: `powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1` and inspect `/api/runtime-build-info`.

- [ ] **Step 7: Perform authenticated browser acceptance without generation**

Open the existing Batch Factory comic media modal and verify:

- Left side contains only the full primary comic.
- Right header shows work name, current page, status, and `生成漫画合集`.
- `漫画页` and `漫画合集` tabs are visible.
- Four candidates stay smaller than the primary comic.
- `上传参考图` is visible under candidates.
- Right middle scrolls while header, tabs, and footer remain fixed.
- Footer shows page navigation and `重新生成本页`.
- No controls overlap at the current desktop viewport.
- Do not click upload, regeneration, collection generation, or publishing.
