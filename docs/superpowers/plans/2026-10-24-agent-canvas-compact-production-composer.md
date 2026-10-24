# Agent Canvas Compact Production Composer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make all five Batch Factory mapped canvas cards use the approved compact creative composer without replacing their existing execution or synchronization paths.

**Architecture:** Batch Factory remains the only owner of book data, assets, production jobs, audio plans, merges, and upload status. The frontend adds one pure action-mapping helper and compact action controls; existing `AgentCanvas` callbacks continue to execute the existing mapped operations. Each card has three areas: real top actions, central content/status, and a one-line bottom configuration bar with the primary action.

**Tech Stack:** React 18, Ant Design 5, lucide-react, @xyflow/react, Node built-in test runner, Vite, Go embedded frontend.

## Global Constraints

- Work only from `v88`; preserve unrelated dirty files.
- Edit React source only, never `frontend/dist` or other generated output.
- Do not add a generation, prompt, asset, merge, upload, or data-copy API.
- Keep mapped-node `batchId`, `bookId`, `resourceType`, `resourceId`, and existing synchronization unchanged.
- Do not trigger image, video, audio, merge, or upload execution while checking the UI.
- Show only actions backed by existing callbacks. Do not display a fixed version, cost, or invented mode.
- Show `重试` only when the live state allows retry.
- Finish with frontend build, Go embedded local build, and a browser review at `http://localhost:4001`.

---

## Files

- Create `frontend/src/user/pages/agent-node-composer-actions.js`: maps node state to permitted direct and more-menu action IDs.
- Create `frontend/src/user/pages/agent-node-composer-actions.test.js`: tests the map for script, image, video, audio, and merge cards.
- Create `frontend/src/user/components/AgentNodeComposerTopActions.jsx`: compact buttons and a no-arrow `更多` menu.
- Create `frontend/src/user/components/AgentNodeComposerReferenceAssets.jsx`: compact reference count, picker entry, and removal list.
- Modify `frontend/src/user/components/AgentBatchFactoryPromptSelector.jsx`: add a popover trigger display mode while preserving its current inline default.
- Modify `frontend/src/user/components/AgentNodeComposer.jsx`: use the compact layout for image, video, audio, and merge cards.
- Modify `frontend/src/user/components/AgentScriptNodeComposer.jsx`: use the compact layout for mapped script cards only.
- Modify `frontend/src/shared/styles/agent-workspace.css`: compact frame and overflow/focus rules.
- Modify `frontend/src/user/pages/agent-node-composer.source.test.js`, `frontend/src/user/pages/agent-canvas-batch-factory.test.js`, `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`, and `frontend/src/user/components/AgentBatchFactoryPublish.source.test.js`: preserve actual actions and mapped behavior.

## Interfaces

```js
export function compactComposerActionPlan({
  kind, source, resourceType, canRetry = false,
  hasImageCandidates = false, hasAudioCandidates = false,
  canPublish = false, hasViralCandidate = false
} = {}) {
  // { primary: string[], more: string[] }
}
```

```jsx
export function AgentNodeComposerTopActions({ primaryActions = [], moreActions = [] }) {
  // action: { id, label, icon, onClick, disabled, visible }
}

export function AgentNodeComposerReferenceAssets({ references = [], disabled = false, onAdd, onRemove }) {}
```

### Task 1: Define the Allowed Action Map

**Files:**
- Create `frontend/src/user/pages/agent-node-composer-actions.js`
- Create `frontend/src/user/pages/agent-node-composer-actions.test.js`

**Produces:** `compactComposerActionPlan`, used by both node composer components.

- [ ] **Step 1: Write failing tests for the action map**

```js
test('video keeps references and templates on top', () => {
  assert.deepEqual(compactComposerActionPlan({ kind: 'video', source: 'batch_factory', resourceType: 'video' }), {
    primary: ['reference-assets', 'prompt-templates'], more: ['preview-inputs']
  });
});
test('failed mapped video adds retry only in more', () => {
  assert.deepEqual(compactComposerActionPlan({ kind: 'video', source: 'batch_factory', resourceType: 'video', canRetry: true }).more, ['preview-inputs', 'retry']);
});
test('mapped script keeps real production actions', () => {
  assert.deepEqual(compactComposerActionPlan({ kind: 'script', source: 'batch_factory', resourceType: 'book_script' }).primary, ['prompt-templates', 'extract-assets', 'generate-storyboard', 'generate-viral']);
});
test('merge exposes organization and publish only when it can publish', () => {
  assert.deepEqual(compactComposerActionPlan({ kind: 'video', source: 'batch_factory', resourceType: 'merge', canPublish: true }), { primary: ['publish-organization'], more: ['upload-video-management'] });
});
```

- [ ] **Step 2: Verify the tests fail**

Run: `node --test frontend/src/user/pages/agent-node-composer-actions.test.js`

Expected: import failure because the helper does not exist.

- [ ] **Step 3: Implement the pure map**

Implement fixed base arrays for `script`, `image`, `video`, `audio`, and `merge`. Select merge before generic video. Add `retry` only for `canRetry`; add image/audio review actions only when candidates exist; add upload only when `canPublish`; omit `prompt-templates` when `source !== 'batch_factory'`. The helper contains no React and no callbacks.

- [ ] **Step 4: Run the focused test**

Run: `node --test frontend/src/user/pages/agent-node-composer-actions.test.js`

Expected: PASS, including image candidates, audio candidates, and a non-Batch-Factory video.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/user/pages/agent-node-composer-actions.js frontend/src/user/pages/agent-node-composer-actions.test.js
git commit -m "feat(agent): map compact composer actions"
```

### Task 2: Add the Compact Top Controls

**Files:**
- Create `frontend/src/user/components/AgentNodeComposerTopActions.jsx`
- Create `frontend/src/user/components/AgentNodeComposerReferenceAssets.jsx`
- Modify `frontend/src/user/components/AgentBatchFactoryPromptSelector.jsx`
- Modify `frontend/src/user/pages/agent-node-composer.source.test.js`

**Consumes:** Action IDs from Task 1 and existing reference/prompt callbacks.

- [ ] **Step 1: Extend source tests first**

```js
assert.match(composerSource, /AgentNodeComposerTopActions/);
assert.match(composerSource, /AgentNodeComposerReferenceAssets/);
assert.match(composerSource, /aria-label="更多功能"/);
assert.match(composerSource, /onPickReferenceAsset\?\.\(node\.id, event\)/);
assert.match(composerSource, /onRemoveReference\?\.\(node\.id, asset\.id\)/);
assert.match(composerSource, /display="popover"/);
```

- [ ] **Step 2: Verify the expanded test fails**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js`

Expected: compact controls are absent.

- [ ] **Step 3: Implement the controls**

`AgentNodeComposerTopActions` renders primary descriptor buttons and a single Ant Design `Dropdown` using the lucide `Ellipsis` icon, `aria-label="更多功能"`, and no down-arrow glyph. It filters invisible descriptors and stops pointer events so React Flow does not close the editor.

`AgentNodeComposerReferenceAssets` shows one `添加参考素材` action with a count. Its popover lists the current authenticated image/video/audio references with the existing remove callback and has one add action that invokes the existing picker callback.

Add `display = 'inline'` to `AgentBatchFactoryPromptSelector`. For `display="popover"`, keep its existing load/filter/save/error state and `onSaved` behavior, but put the same selects inside a popover opened by `选择提示词`.

- [ ] **Step 4: Run the source test**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js`

Expected: PASS. Existing inline selector callers are unchanged.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/user/components/AgentNodeComposerTopActions.jsx frontend/src/user/components/AgentNodeComposerReferenceAssets.jsx frontend/src/user/components/AgentBatchFactoryPromptSelector.jsx frontend/src/user/pages/agent-node-composer.source.test.js
git commit -m "feat(agent): add compact composer controls"
```

### Task 3: Compact Image, Video, Audio, and Final Merge Cards

**Files:**
- Modify `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify `frontend/src/shared/styles/agent-workspace.css`
- Modify `frontend/src/user/pages/agent-node-composer.source.test.js`
- Modify `frontend/src/user/components/AgentBatchFactoryPublish.source.test.js`
- Modify `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

**Consumes:** Tasks 1-2 and existing `onPreviewInputs`, retry, candidate confirmation, submit, organization, and publish callbacks.

- [ ] **Step 1: Add failing compact-layout assertions**

```js
assert.match(composerSource, /agent-node-composer__top-actions/);
assert.match(composerSource, /agent-node-composer__body/);
assert.match(composerSource, /agent-node-composer__bottom-bar/);
assert.match(composerSource, /compactComposerActionPlan/);
assert.match(composerSource, /onPreviewInputs\?\.\(node\.id, selectedModel\)/);
assert.match(composerSource, /onRetryVideo\?\.\(node\.id\)/);
assert.match(composerSource, /onConfirmImage\?\.\(node\.id, selectedImageID\)/);
assert.match(composerSource, /onConfirmAudio\?\.\(node\.id, selectedAudioID\)/);
```

Also require `onPublishOrganizationChange`, `onPublishBatchFactory`, and `合成最终视频` in the publish source test.

- [ ] **Step 2: Verify the tests fail**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/components/AgentBatchFactoryPublish.source.test.js`

Expected: compact zone selectors are absent.

- [ ] **Step 3: Refactor the render tree, preserving every callback**

Compute the Task 1 action plan from `kind`, mapped source/resource type, retry state, candidate state, and publication state. Bind descriptors locally: reference assets use the Task 2 component; templates use the popover selector; preview invokes `onPreviewInputs`; retry uses the existing generic/video retry; image/audio confirmation uses its existing callback; upload invokes `onPublishBatchFactory` only after the existing organization condition is satisfied.

Put the action strip at top. Keep prompt local state and `onBlur={persistPrompt}` in the central body. Keep generated candidates/results in that body. For `resourceType === 'merge'`, show merge readiness, organization selection, and publication state in the body instead of a prompt field. Keep `AgentModelPicker`, `AgentNodeParameterPopover`, `AgentNodeSettingsPanel`, timing warning, state, and the existing primary submit callback in the bottom bar. Do not change request payloads, `batchFactoryMappedAction`, or any API function.

Update CSS to use one dark compact frame with a top action row, central prompt, and non-wrapping bottom bar. At narrow widths, the bottom bar scrolls horizontally rather than overlap or hide controls. Preserve keyboard focus and existing React Flow z-index behavior.

- [ ] **Step 4: Run focused tests and build**

Run: `node --test frontend/src/user/pages/agent-node-composer-actions.test.js frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/components/AgentBatchFactoryPublish.source.test.js; npm --prefix frontend run build`

Expected: PASS and a successful Vite build.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/user/components/AgentNodeComposer.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/components/AgentBatchFactoryPublish.source.test.js
git commit -m "feat(agent): compact mapped media composers"
```

### Task 4: Compact the Mapped Script Card

**Files:**
- Modify `frontend/src/user/components/AgentScriptNodeComposer.jsx`
- Modify `frontend/src/shared/styles/agent-workspace.css`
- Modify `frontend/src/user/pages/agent-node-composer.source.test.js`
- Modify `frontend/src/user/pages/agent-canvas-batch-factory.test.js`

**Consumes:** Tasks 1-2 and existing `runBatchFactoryStage`, `onCommitSource`, and text model callbacks.

- [ ] **Step 1: Add failing script assertions**

```js
assert.match(scriptComposerSource, /AgentNodeComposerTopActions/);
assert.match(scriptComposerSource, /contentRangeLines/);
assert.match(scriptComposerSource, /runBatchFactoryStage\('assets'\)/);
assert.match(scriptComposerSource, /runBatchFactoryStage\('director'\)/);
assert.match(scriptComposerSource, /runBatchFactoryStage\('viral'\)/);
assert.match(scriptComposerSource, /adoptViralCandidate/);
assert.match(scriptComposerSource, /AgentBatchFactoryPromptSelector[\s\S]*display="popover"/);
```

- [ ] **Step 2: Verify the tests fail**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-batch-factory.test.js`

Expected: the script composer still exposes numbered `agent-script-step` sections.

- [ ] **Step 3: Refactor only the mapped script path**

For `isBatchFactoryScript`, use the same top/middle/bottom skeleton. Top actions are template selection, `提取图片资产`, `生成视频分镜`, and `生成爆款候选`; `保存行数` and `采用候选` belong in the no-arrow more menu. The bottom bar contains text-model selection and the `contentRangeLines` input. Preserve `draftSource`, `onCommitSource`, stage busy/errors, viral candidate display, and candidate adoption. Do not modify the manual script splitting and shot-batch workflow.

- [ ] **Step 4: Run tests and build**

Run: `node --test frontend/src/user/pages/agent-node-composer-actions.test.js frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-batch-factory.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js; npm --prefix frontend run build`

Expected: PASS and a successful Vite build.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/user/components/AgentScriptNodeComposer.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-batch-factory.test.js
git commit -m "feat(agent): compact mapped script composer"
```

### Task 5: Verify Shared Production Behavior in the Embedded Runtime

**Files:**
- Modify only files from Tasks 1-4 when verification exposes a real regression.

- [ ] **Step 1: Run the focused frontend suite**

Run: `node --test frontend/src/user/pages/agent-node-composer-actions.test.js frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-batch-factory.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/components/AgentBatchFactoryPublish.source.test.js`

Expected: PASS. Restore an existing callback if this finds a lost action; do not create a replacement API.

- [ ] **Step 2: Run mapped projection regressions**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run 'Test.*BatchFactory|Test.*Projection|Test.*Agent' -count=1; Set-Location ..`

Expected: PASS. Projection still creates script, image, video, audio, and merge nodes with current IDs and edges.

- [ ] **Step 3: Build the local embedded review runtime**

Run: `npm --prefix frontend run build; .\scripts\build-go-local.ps1; .\scripts\start-local-go-review.ps1 -UseYcmMysql`

Expected: frontend build and Go embedded build succeed; `http://localhost:4001` uses the new source.

- [ ] **Step 4: Check the real mapped book in the browser without executing jobs**

Open `/agent/canvas?taskId=agent-bf-dbba32e978f13f1909e3b928&batchId=batch_2f462bb6608219f96838d30435620386&bookId=book_4090786e93fcf091439f12369c9dd02b&view=workflow`. Open script, image, video, audio, and merge cards. Confirm: each uses the compact shape; top actions match the type; `更多` has no arrow glyph; reference selection still opens the material picker; prompt selection opens saved templates; model/parameters remain reachable; merge still offers organization, compose, and upload status. Do not click generate, compose, or upload.

- [ ] **Step 5: Record only intentional changes**

Run: `git status --short; git log -1 --oneline`

Expected: no unrelated dirty file is staged, changed, reverted, or committed.

## Review

- Tasks 3-4 cover all five confirmed node types.
- Existing Batch Factory data ownership and mapped execution routes remain unchanged.
- Tasks 1-2 prevent visual controls from inventing actions.
- Task 5 verifies source tests, Go projection, embedded runtime, and browser behavior.
