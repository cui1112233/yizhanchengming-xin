# Batch Factory Text Workbench UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the text mode to batch creation, full unified/book settings, book rows, and a comic-parity text media library with server previews and page editing.

**Architecture:** Keep the current Shuihuo Batch Factory shell and extract new text components into focused files. Reuse base controls and modal geometry, but keep `TextMediaPanel` independent from `ComicMediaPanel`; both consume their own APIs and state. All previews come from the Go renderer.

**Tech Stack:** React 18, Ant Design 5, existing Batch Factory CSS, Node test runner

## Global Constraints

- The production switch reads `视频生成 | 漫画生成 | 文本生成`.
- Text mode unified tabs are `模型配置 | 排版配置 | 发布配置`.
- Unified configuration contains all whole-book fields; book configuration mirrors them and adds only inheritance/override behavior.
- Text media layout keeps the comic modal size, left preview, right tabs, candidate versions, page arrows, and bottom controls.
- Labels use `文字页`, `本页背景`, and `页面内容`.
- Default page target is 150, content range is 4000, and aspect is 3:4.
- No system prompt text is hard-coded in frontend files.
- Preserve unrelated dirty worktree changes and commit only files named by each task.

---

### Task 1: Add text mode to batch creation and mode presentation

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryAutomationSchedule.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryNovelListHelpers.js`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.source.test.js`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryProductionMode.test.js`

**Interfaces:**
- Consumes: the three-mode helper from plan 01.
- Produces: text new-batch defaults `{ productionMode:'text', textPageTarget:150, contentRange:4000, imageAspectRatio:'3:4' }`.

- [ ] **Step 1: Add failing UI source assertions**

Assert `文本生成` appears in both switches, text mode shows “单页字数”, and defaults are 150/4000/3:4.

- [ ] **Step 2: Run and verify failure**

Run: `node --test src/user/pages/shuihuo/BatchFactoryCreateModal.source.test.js src/user/pages/shuihuo/batchFactoryProductionMode.test.js` from `frontend`
Expected: FAIL because text is missing.

- [ ] **Step 3: Add the third mode and text defaults**

Keep video/comic defaults untouched. When mode changes to text, seed only missing text fields; do not overwrite values already edited in the draft.

- [ ] **Step 4: Add text labels to automation presets and source summary**

`automationPresetModeLabel('text')` returns `文本`. The source summary reads `已带入 · N 字 · 单页约 150 字`.

- [ ] **Step 5: Run and commit**

Run the focused tests above; expect PASS.

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.jsx frontend/src/user/pages/shuihuo/batchFactoryAutomationSchedule.js frontend/src/user/pages/shuihuo/batchFactoryNovelListHelpers.js frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.source.test.js frontend/src/user/pages/shuihuo/batchFactoryProductionMode.test.js
git commit -m "feat: add text mode batch entry"
```

### Task 2: Build one complete text configuration schema for unified and book forms

**Files:**
- Create: `frontend/src/user/pages/shuihuo/textProductionConfig.js`
- Create: `frontend/src/user/pages/shuihuo/TextProductionSettingsForm.jsx`
- Create: `frontend/src/user/pages/shuihuo/textProductionConfig.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js`

**Interfaces:**
- Produces: `TEXT_PRODUCTION_CONFIG_KEYS`, `defaultTextProductionConfig()`, `normalizeTextProductionConfig(value)`, `textProductionBookOverride(batchPatch, bookPatch)`.
- Produces: `TextProductionSettingsForm({ value, inherited, explicitKeys, onChange, onRestore, scope })`.

- [ ] **Step 1: Add failing key-parity tests**

The expected key list includes model/sensitive fields, `highlightReuseEnabled`, content range, target characters, aspect, font, size, weight, spacing, padding, alignment, colors, max occupancy, background, cover, publish mapping, auto collection, auto upload, and retry policy. Read legacy `publishSettings.biaohongReuse` only as compatibility when the new top-level field is absent.

Assert unified and book forms import the same `TEXT_PRODUCTION_CONFIG_KEYS`; book form may not define a second list.

- [ ] **Step 2: Run and verify failure**

Run: `node --test src/user/pages/shuihuo/textProductionConfig.test.js src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js` from `frontend`
Expected: FAIL because the schema does not exist.

- [ ] **Step 3: Implement defaults and sparse override helpers**

```js
export const TEXT_PRODUCTION_DEFAULTS = Object.freeze({
  textPageTarget: 150,
  contentRange: 4000,
  imageAspectRatio: '3:4',
  textFontId: 'noto-sans-sc',
  textFontSize: 44,
  textFontWeight: 400,
  textLineHeight: 1.65,
  textParagraphSpacing: 24,
  textPadding: 96,
  textAlign: 'left',
  textColor: '#FFFFFF',
  textHighlightColor: '#E53935',
  textMaxHeightRatio: 0.75,
  textBackground: { type: 'color', color: '#111318' }
});
```

- [ ] **Step 4: Render the same fields in unified and book scopes**

Unified scope edits direct values. Book scope shows inherited values, writes only changed fields, and exposes “恢复继承” for each explicit field. Place AI/sensitive controls under model configuration, visual fields under layout, and upload fields under publish.

- [ ] **Step 5: Verify unified completeness**

Add a test that every `TEXT_PRODUCTION_CONFIG_KEYS` entry appears in the unified form binding and can be represented by the book override helper.

- [ ] **Step 6: Run and commit**

Run: `node --test src/user/pages/shuihuo/textProductionConfig.test.js src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js` from `frontend`
Expected: PASS.

```bash
git add frontend/src/user/pages/shuihuo/textProductionConfig.js frontend/src/user/pages/shuihuo/TextProductionSettingsForm.jsx frontend/src/user/pages/shuihuo/textProductionConfig.test.js frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js
git commit -m "feat: add complete text production settings"
```

### Task 3: Add text page state helpers and authenticated previews

**Files:**
- Create: `frontend/src/user/pages/shuihuo/textMediaState.js`
- Create: `frontend/src/user/pages/shuihuo/textMediaState.test.js`
- Create: `frontend/src/user/pages/shuihuo/AuthenticatedTextPageImage.jsx`
- Create: `frontend/src/user/pages/shuihuo/AuthenticatedTextPageImage.source.test.js`

**Interfaces:**
- Produces: `textPagePresentation(page)`, `textBookCompletion(pages)`, `canCreateTextCollection(pages, ids)`, `textPageDirtyImpact(before, after)`.
- Produces: authenticated image component that uses the existing token-aware media loading pattern.

- [ ] **Step 1: Add failing state tests**

Cover empty, rendering, ready, failed, locked upload, stale layout, and mixed ready/not-ready collection selection. Assert locked upload is always presented as the current artifact.

- [ ] **Step 2: Run and verify failure**

Run: `node --test src/user/pages/shuihuo/textMediaState.test.js src/user/pages/shuihuo/AuthenticatedTextPageImage.source.test.js` from `frontend`
Expected: FAIL because helpers are absent.

- [ ] **Step 3: Implement pure state helpers**

Keep network calls out of this module. `textPageDirtyImpact` returns one of `none|render|reflow`; background/color changes return `render`, while text/font/size/spacing/aspect changes return `reflow`.

- [ ] **Step 4: Implement authenticated preview loading**

Follow `AuthenticatedComicPageImage` behavior: request with the user token, create an object URL, revoke it on cleanup, and never expose a local storage path.

- [ ] **Step 5: Run and commit**

Run the focused tests; expect PASS.

```bash
git add frontend/src/user/pages/shuihuo/textMediaState.js frontend/src/user/pages/shuihuo/textMediaState.test.js frontend/src/user/pages/shuihuo/AuthenticatedTextPageImage.jsx frontend/src/user/pages/shuihuo/AuthenticatedTextPageImage.source.test.js
git commit -m "feat: add text media presentation state"
```

### Task 4: Build the comic-parity TextMediaPanel

**Files:**
- Create: `frontend/src/user/pages/shuihuo/TextMediaPanel.jsx`
- Create: `frontend/src/user/pages/shuihuo/TextMediaPanel.source.test.js`
- Create: `frontend/src/user/pages/shuihuo/text-media-panel.css`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batch-factory-automation-theme.css`

**Interfaces:**
- Consumes text page APIs from plan 02.
- Produces `TextMediaPanel({ batchId, book, settings, initialPageId, onSaved, onClose })`.

- [ ] **Step 1: Write failing structural tests**

Assert the component contains left preview, `封面`, `文字页` and `文本合集` views, candidate artifacts, `本页背景`, `页面内容`, previous/next arrows, and bottom actions `展开`, `合成当前书`, `上传网络`.

- [ ] **Step 2: Run and verify failure**

Run: `node --test src/user/pages/shuihuo/TextMediaPanel.source.test.js` from `frontend`
Expected: FAIL because the component does not exist.

- [ ] **Step 3: Implement loading and selection**

Load pages and collections when opened, preserve the selected page across refresh, poll only while a page is rendering, and stop polling on close. Use the same modal dimensions and grid proportions as the comic media workbench.

- [ ] **Step 4: Wire the mode-specific media modal**

In `BatchFactoryNovelList.jsx`, render `TextMediaPanel` when effective mode is `text`; keep `ComicMediaPanel` for comic and `MediaVersionPanel` for video. Do not add text branches inside the comic component.

- [ ] **Step 5: Add text badge and mode labels to book rows**

Every text-mode row, thumbnail, filter, and media header displays `文本`, never `漫画`.

- [ ] **Step 6: Run source tests and build**

Run: `node --test src/user/pages/shuihuo/TextMediaPanel.source.test.js src/user/pages/shuihuo/batchFactoryProductionMode.test.js` from `frontend`
Run: `npm run build` from `frontend`
Expected: PASS.

- [ ] **Step 7: Commit the media shell**

```bash
git add frontend/src/user/pages/shuihuo/TextMediaPanel.jsx frontend/src/user/pages/shuihuo/TextMediaPanel.source.test.js frontend/src/user/pages/shuihuo/text-media-panel.css frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/batch-factory-automation-theme.css
git commit -m "feat: add text media workbench"
```

### Task 5: Add page editing, backgrounds, preview, locking, and collection controls

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/TextMediaPanel.jsx`
- Modify: `frontend/src/user/pages/shuihuo/TextMediaPanel.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/text-media-panel.css`
- Create: `frontend/src/user/pages/shuihuo/textMediaActions.test.js`

**Interfaces:**
- Consumes: page patch/preview/render/upload/restore/background/collection APIs.
- Preserves: uploaded artifact lock and optimistic page revision.

- [ ] **Step 1: Add failing action tests**

Test that content/font/aspect edits request reflow, color/background edits request render only, split/merge preserve complete sentences, page upload sets a locked artifact, cover upload sets a locked cover, both restore actions require confirmation, collection stays disabled until cover and all selected pages are ready, and an already uploaded collection displays `再次上传`.

- [ ] **Step 2: Run and verify failure**

Run: `node --test src/user/pages/shuihuo/textMediaActions.test.js src/user/pages/shuihuo/TextMediaPanel.source.test.js` from `frontend`
Expected: FAIL before actions are wired.

- [ ] **Step 3: Add debounced server preview**

After 300 ms without edits, send the page text plus effective layout/background to `previewTextPage`. Cancel the previous request when inputs change. Display server errors beside the preview without discarding the last successful image.

Add explicit “在完整句子后拆页” and “与下一页合并” actions. The client sends the requested boundary; the server validates that no sentence is cut and returns the authoritative pages.

- [ ] **Step 4: Add background controls**

Provide solid color picker/input, background-library selection, upload, “应用到整本书”, and current-page override. Show a low-contrast warning returned by the server; never auto-change colors.

- [ ] **Step 5: Add page upload and restore**

Uploading a finished page marks it locked and displays `上传成品页`. `恢复系统生成` requires confirmation, removes the user lock through the API, and rerenders that page.

Add the same candidate/lock/restore behavior for the title cover. The system cover preview uses book title, selected background, and cover font settings from the server renderer.

- [ ] **Step 6: Add collection and publish affordances**

Keep `合成当前书` disabled until the selected page range is ready. Show `重新合成` after a collection exists and `再次上传` after a successful publish record.

- [ ] **Step 7: Run tests and build**

Run: `node --test src/user/pages/shuihuo/textMediaActions.test.js src/user/pages/shuihuo/TextMediaPanel.source.test.js` from `frontend`
Run: `npm run build` from `frontend`
Expected: PASS.

- [ ] **Step 8: Commit workbench interactions**

```bash
git add frontend/src/user/pages/shuihuo/TextMediaPanel.jsx frontend/src/user/pages/shuihuo/TextMediaPanel.source.test.js frontend/src/user/pages/shuihuo/text-media-panel.css frontend/src/user/pages/shuihuo/textMediaActions.test.js
git commit -m "feat: edit and preview text pages"
```
