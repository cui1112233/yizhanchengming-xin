# Batch Factory Production Mode Settings Visibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make unified and single-book Batch Factory configuration render only the controls used by the active video or comic production mode while preserving inactive settings.

**Architecture:** Add one small pure visibility model shared by both configuration surfaces. Use its `video` and `comic` flags for conditional React rendering; do not delete inactive values or change the saved schema. Keep backend execution isolation on the existing `productionMode` contract and verify it with focused existing Go tests.

**Tech Stack:** React 18, Ant Design, Node test runner, Vite, Go, Go embed runtime.

## Global Constraints

- Work directly on the canonical `v88` branch; do not create a feature branch.
- Preserve inactive video and comic configuration values when switching modes.
- Do not modify prompt bodies, model catalogs, provider contracts, or publishing behavior.
- Do not overwrite unrelated dirty files, especially the existing single-book and unified source test files.
- Build React source first, then rebuild the Go embedded runtime.

---

### Task 1: Shared Production Mode Visibility Model

**Files:**
- Create: `frontend/src/user/pages/shuihuo/batchFactoryProductionMode.js`
- Create: `frontend/src/user/pages/shuihuo/batchFactoryProductionMode.test.js`

**Interfaces:**
- Produces: `batchFactoryProductionMode(value): 'video' | 'comic'`
- Produces: `batchFactoryProductionModeVisibility(value): { mode: 'video' | 'comic', video: boolean, comic: boolean }`

- [ ] **Step 1: Write the failing pure behavior tests**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import {
  batchFactoryProductionMode,
  batchFactoryProductionModeVisibility
} from './batchFactoryProductionMode.js';

test('defaults historical settings to video mode', () => {
  assert.equal(batchFactoryProductionMode(undefined), 'video');
  assert.deepEqual(batchFactoryProductionModeVisibility('unknown'), {
    mode: 'video', video: true, comic: false
  });
});

test('comic mode exposes comic settings and hides video settings', () => {
  assert.deepEqual(batchFactoryProductionModeVisibility('comic'), {
    mode: 'comic', video: false, comic: true
  });
});
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js`

Expected: FAIL because `batchFactoryProductionMode.js` does not exist.

- [ ] **Step 3: Implement the visibility model**

```js
export function batchFactoryProductionMode(value) {
  return value === 'comic' ? 'comic' : 'video';
}

export function batchFactoryProductionModeVisibility(value) {
  const mode = batchFactoryProductionMode(value);
  return { mode, video: mode === 'video', comic: mode === 'comic' };
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js`

Expected: 2 tests pass.

- [ ] **Step 5: Commit the helper and test**

```powershell
git add -- frontend/src/user/pages/shuihuo/batchFactoryProductionMode.js frontend/src/user/pages/shuihuo/batchFactoryProductionMode.test.js
git commit -m "test(batch-factory): define production mode visibility"
```

### Task 2: Unified Configuration Visibility

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Create: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

**Interfaces:**
- Consumes: `batchFactoryProductionModeVisibility(value.productionMode)`
- Preserves: the complete `draftPatch`, including inactive-mode fields

- [ ] **Step 1: Write failing source-contract tests for unified configuration**

The dedicated source test must assert that:

```js
assert.match(unifiedSource, /batchFactoryProductionModeVisibility/);
assert.match(unifiedSource, /visibility\.video\s*\?/);
assert.match(unifiedSource, /visibility\.comic\s*\?/);
assert.match(unifiedSource, /visibility\.video\s*&&\s*sections\.includes\('audio'\)/);
assert.match(unifiedSource, /visibility\.video\s*\?\s*<section[^>]*>.*?<b>视频设置<\/b>/s);
assert.match(unifiedSource, /visibility\.comic\s*\?\s*<BatchFactoryComicStylePicker/s);
```

It must also assert that `setProductionMode` spreads the existing `value` and `aiPromptConfig`, proving inactive values are retained.

- [ ] **Step 2: Run the dedicated source test and verify RED**

Run: `node --test src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

Expected: FAIL because the unified modal does not use the shared visibility model and still renders video/audio details in comic mode.

- [ ] **Step 3: Gate unified model and audio controls**

In `BatchFactoryEngineSettingsForm`:

- derive `visibility` from `value.productionMode`;
- always render text model, image provider/model, and image aspect ratio;
- render video provider/model, video aspect ratio, resolution, and storyboard duration only when `visibility.video`;
- render audio source, voice, style, speed, pitch, and the follow-audio section only when `visibility.video`.

Use conditional React fragments rather than CSS hiding so inactive controls are absent from the accessible tree.

- [ ] **Step 4: Gate unified media details**

Keep both mode headers and switches visible. Change their bodies to:

```jsx
{visibility.video ? <div className="batch-factory-ai-module-body">...</div> : null}
{visibility.comic ? <div className="batch-factory-ai-module-body">...</div> : null}
```

Do not remove inactive values inside `setProductionMode`.

- [ ] **Step 5: Gate unified AI reasoning sections**

- Use the video asset prompt and standard constraint categories in video mode.
- Use the comic asset prompt and comic constraint categories in comic mode.
- Render `BatchFactoryComicStylePicker` only for `visibility.comic`.
- Render the complete video settings section only for `visibility.video`.

- [ ] **Step 6: Run focused tests and verify GREEN**

Run:

```powershell
node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
```

Expected: all tests pass.

- [ ] **Step 7: Commit unified visibility**

```powershell
git add -- frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
git commit -m "feat(batch-factory): separate unified settings by production mode"
```

### Task 3: Single-Book Configuration Visibility

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

**Interfaces:**
- Consumes: `batchFactoryProductionModeVisibility(form.productionMode)`
- Preserves: `buildBookRegionUpdate` inheritance and override behavior

- [ ] **Step 1: Extend the source-contract test and verify RED**

Add assertions proving that the single-book source:

```js
assert.match(bookSource, /batchFactoryProductionModeVisibility/);
assert.match(bookSource, /visibility\.video\s*\?/);
assert.match(bookSource, /visibility\.comic\s*\?/);
assert.match(bookSource, /visibility\.comic\s*\?\s*<BatchFactoryComicStylePicker/s);
```

Add negative assertions that comic rendering branches do not contain `音频来源`, `配音音色`, `视频模型`, or `视频提示词`.

Run: `node --test src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

Expected: FAIL because the single-book engine still renders video/audio controls for comic mode.

- [ ] **Step 2: Gate the single-book engine section**

- Keep the production mode control visible.
- Keep text/image model and image aspect controls shared.
- Render video provider/model, video format, duration, and audio cards only for `visibility.video`.
- Do not remove hidden values from `form`, `bookPatch`, or `buildBookRegionUpdate`.

- [ ] **Step 3: Gate single-book prompt sections**

- Render only the active mode's generation prompt controls.
- Keep the asset selector mode-aware, using the existing comic or script catalog.
- Render comic image constraints and style picker only for `visibility.comic`.
- Render standalone video prompt settings only for `visibility.video`.

- [ ] **Step 4: Run focused frontend tests**

Run:

```powershell
node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js src/user/pages/shuihuo/batchFactoryComicStyle.test.js
```

Expected: all focused tests pass.

- [ ] **Step 5: Run backend mode-isolation tests**

Run:

```powershell
go test ./backend/internal/batchfactoryv11 -run "ProductionMode|ComicPrompt|ComicImage" -count=1
```

Expected: existing video/comic routing and comic prompt contract tests pass.

- [ ] **Step 6: Commit single-book visibility**

```powershell
git add -- frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
git commit -m "feat(batch-factory): separate book settings by production mode"
```

### Task 4: Build, Go Embed, and Browser Acceptance

**Files:**
- Verify only; no new source file is required.

**Interfaces:**
- Consumes: completed unified and single-book visibility behavior
- Produces: a locally running Go binary whose runtime SHA contains all feature commits

- [ ] **Step 1: Run the complete focused verification set**

```powershell
node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js src/user/pages/shuihuo/batchFactoryComicStyle.test.js
npm run build
```

Expected: focused tests and Vite production build pass.

- [ ] **Step 2: Rebuild and restart the local Go embedded runtime**

Run the repository-approved local Go build/start scripts. Do not copy `frontend/dist` back into source directories and do not use the retired Node or Docker release path.

- [ ] **Step 3: Verify runtime identity**

Check `/health` and `/api/runtime-build-info`. Confirm `ok: true`, `runtime: go`, and `git_sha`/`runtime_manifest_sha` include the feature commits.

- [ ] **Step 4: Verify unified configuration in the browser**

Without saving test edits:

- video mode shows video prompt, video model, audio source, TTS, and follow-audio controls but no comic style library;
- comic mode shows comic prompt, image derivation, comic constraints, style library, and comic reference controls but no video/audio controls;
- switching back restores the prior video values.

Click `取消` after verification.

- [ ] **Step 5: Verify single-book configuration in the browser**

Repeat the video/comic visibility checks for `小妾造反后`. Confirm asset and constraint settings follow the selected mode. Click `取消` after verification.

- [ ] **Step 6: Final repository check**

Run `git diff --check` and confirm only this feature's intended files are committed. Report unrelated pre-existing test failures separately without changing those files.
