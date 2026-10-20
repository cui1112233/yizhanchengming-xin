# Agent Canvas Collapsible Node Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Agent media nodes display as compact previews by default, open one inline editor at a time, anchor connection controls to the visible preview, and let users hide or restore saved edges without deleting them.

**Architecture:** Keep persisted canvas data and all Go APIs unchanged. Add one small pure presentation module for canonical node types, keep temporary editor state inside `AgentCanvas`, render handles inside a preview shell, and express edge visibility through the existing `appearance` object plus React Flow's `hidden` edge property. Move existing video-job controls into the inline editor so removing the competing media modal does not remove any working video action.

**Tech Stack:** React 18, Ant Design 5, `@xyflow/react` 12, Lucide icons, Node's built-in test runner, Vite, Go `embed`, PowerShell local-review scripts.

## Global Constraints

- Work only on branch `v88`; do not create another branch.
- Preserve all unrelated dirty files and stage only the exact files listed by each task.
- `AgentCanvas.jsx` and `agent-workspace.css` are already dirty before this plan. Before each commit, inspect `git diff` and `git diff --cached` for the listed paths; do not stage a hunk unless it belongs to the approved Agent canvas chain or is a required prerequisite already present in those same files.
- Do not modify MySQL migration history or add a migration; this change uses the existing canvas JSON contract.
- Do not add or restore a Node runtime or release path; the final local application is the Go-embedded React build.
- Do not copy Liblib or Shenlang source, branding, private endpoints, account data, or project data.
- Do not submit image, audio, video, or other paid/provider generation jobs during browser acceptance.
- Do not deploy publicly. Rebuild and restart only the verified local Go review runtime on port `4001`.
- Image and audio generation remain unavailable. Existing Agent video generation, cancellation, retry, and confirmation behavior must remain available.
- The current canvas colors remain the product palette; this work changes structure, spacing, and interaction, not branding.

---

## File Map

- Create `frontend/src/user/pages/agent-canvas-node-presentation.js`: canonical node-type and media-type helpers only.
- Create `frontend/src/user/pages/agent-canvas-node-presentation.test.js`: pure tests for legacy and unknown node types.
- Create `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`: focused source/CSS contract for editor, handles, asset actions, video actions, and edge visibility.
- Modify `frontend/src/user/pages/agent-canvas-commands.js`: new nodes use canonical `text`, not legacy `copy`.
- Modify `frontend/src/user/pages/agent-canvas-commands.test.js`: update the new-node contract.
- Modify `frontend/src/user/pages/agent-canvas-appearance.js`: add normalized `edgesVisible` with a backward-compatible `true` default.
- Modify `frontend/src/user/pages/agent-canvas-appearance.test.js`: cover default and explicit hidden state.
- Modify `frontend/src/user/components/AgentCanvas.jsx`: editor state, preview shell, inline video actions, handle placement, normalized serialization, and edge visibility tool.
- Modify `frontend/src/user/pages/agent-canvas-context-menu.test.js`: replace old always-visible/media-modal assertions with the approved behavior.
- Modify `frontend/src/shared/styles/agent-workspace.css`: compact media geometry, collapsible editor, attached conditional handles, and responsive behavior.

---

### Task 1: Canonical Node Types and Backward-Compatible Appearance

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-node-presentation.js`
- Create: `frontend/src/user/pages/agent-canvas-node-presentation.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-commands.js`
- Modify: `frontend/src/user/pages/agent-canvas-commands.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-appearance.js`
- Modify: `frontend/src/user/pages/agent-canvas-appearance.test.js`

**Interfaces:**
- Produces: `normalizeCanvasNodeType(value): 'text' | 'image' | 'video' | 'audio'`.
- Produces: `isMediaCanvasNodeType(value): boolean`.
- Produces: `normalizeCanvasAppearance(value).edgesVisible: boolean`, default `true`.
- Consumed later by: `toFlowNode`, `toCanvas`, media editor routing, and `toFlowEdge` in `AgentCanvas.jsx`.

- [ ] **Step 1: Write failing node-presentation and appearance tests**

```js
// agent-canvas-node-presentation.test.js
import test from 'node:test';
import assert from 'node:assert/strict';
import { isMediaCanvasNodeType, normalizeCanvasNodeType } from './agent-canvas-node-presentation.js';

test('legacy, blank, and unknown node types become text', () => {
  for (const value of ['copy', '', null, 'unexpected']) {
    assert.equal(normalizeCanvasNodeType(value), 'text');
  }
});

test('supported media types remain media types', () => {
  for (const value of ['image', 'video', 'audio']) {
    assert.equal(normalizeCanvasNodeType(value), value);
    assert.equal(isMediaCanvasNodeType(value), true);
  }
  assert.equal(isMediaCanvasNodeType('copy'), false);
});
```

Extend `agent-canvas-appearance.test.js` so the default expected object includes `edgesVisible: true`, and the explicit-false case passes `edgesVisible: false` and expects it to remain false. Change the `createCanvasNode` expectation in `agent-canvas-commands.test.js` from `data.type: 'copy'` to `data.type: 'text'`.

- [ ] **Step 2: Run the focused tests and verify they fail for the missing module/property and old `copy` default**

Run from `frontend`:

```powershell
node --test src/user/pages/agent-canvas-node-presentation.test.js src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-commands.test.js
```

Expected: FAIL because `agent-canvas-node-presentation.js` does not exist, `edgesVisible` is absent, and new nodes still use `copy`.

- [ ] **Step 3: Add the minimal canonical helpers and defaults**

```js
// agent-canvas-node-presentation.js
const supportedNodeTypes = new Set(['text', 'image', 'video', 'audio']);
const mediaNodeTypes = new Set(['image', 'video', 'audio']);

export function normalizeCanvasNodeType(value) {
  const type = String(value || '').trim().toLowerCase();
  return supportedNodeTypes.has(type) ? type : 'text';
}

export function isMediaCanvasNodeType(value) {
  return mediaNodeTypes.has(normalizeCanvasNodeType(value));
}
```

Add `edgesVisible: true` to `defaultCanvasAppearance` and normalize it with the same explicit-boolean pattern already used by the other appearance settings. Change `createCanvasNode` to emit `data.type: 'text'`.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run the Step 2 command again.

Expected: all node-presentation, appearance, and command tests PASS.

- [ ] **Step 5: Commit only Task 1 files**

```powershell
git add -- frontend/src/user/pages/agent-canvas-node-presentation.js frontend/src/user/pages/agent-canvas-node-presentation.test.js frontend/src/user/pages/agent-canvas-commands.js frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/pages/agent-canvas-appearance.js frontend/src/user/pages/agent-canvas-appearance.test.js
git commit -m "fix: normalize agent canvas presentation state"
```

---

### Task 2: One Inline Media Editor and Explicit Asset Actions

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Consumes: `normalizeCanvasNodeType` and `isMediaCanvasNodeType` from Task 1.
- Produces: temporary `activeEditorNodeID: string` state in `AgentCanvas`.
- Produces: `CanvasNode` props `isEditorOpen`, `onOpenEditor`, and explicit `onPickImageAsset` behavior.
- Preserves: text-node double-click manual editing; media-node double-click does not open the manual modal.

- [ ] **Step 1: Write the failing source contract**

Create `agent-canvas-collapsible-editor.source.test.js` that reads `AgentCanvas.jsx` and `agent-workspace.css` and asserts these exact contracts:

```js
test('media nodes render one temporary inline editor controlled by preview activation', () => {
  assert.match(canvasSource, /const \[activeEditorNodeID, setActiveEditorNodeID\] = useState\(''\)/);
  assert.match(canvasSource, /isEditorOpen=\{activeEditorNodeID === props\.id\}/);
  assert.match(canvasSource, /onOpenEditor=\{setActiveEditorNodeID\}/);
  assert.match(canvasSource, /const generationPanel = isEditorOpen/);
  assert.match(canvasSource, /\{generationPanel\}/);
  assert.match(canvasSource, /setActiveEditorNodeID\(''\)/);
});

test('image preview opens editing while a named editor action opens the asset picker', () => {
  assert.match(canvasSource, /aria-label="展开图片编辑"/);
  assert.match(canvasSource, /aria-label=\{data\.assetRef\?\.url \? '更换图片素材' : '添加图片素材'\}/);
  assert.doesNotMatch(canvasSource, /aria-label="选择图片素材"[\s\S]*?onClick=\{event => \{ stopCanvasEvent\(event\); onPickImageAsset/);
});

test('only text nodes use the manual double-click editor', () => {
  assert.match(canvasSource, /if \(normalizeCanvasNodeType\(node\.data\?\.type\) !== 'text'\) return;/);
});

test('loaded and serialized nodes both use the canonical type', () => {
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.type\)/);
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.data\.type\)/);
});
```

Update the old `agent-canvas-context-menu.test.js` assertions that require permanently visible media controls and a media manual fallback. Keep tests for prompt/model/reference controls, but require them inside the conditional editor.

- [ ] **Step 2: Run the source contract and verify it fails**

```powershell
node --test src/user/pages/agent-canvas-collapsible-editor.source.test.js src/user/pages/agent-canvas-context-menu.test.js
```

Expected: FAIL because no active editor state exists, the generation panel is unconditional, and image preview still opens the asset picker.

- [ ] **Step 3: Implement temporary editor state and event boundaries**

In `AgentCanvas`:

```jsx
const [activeEditorNodeID, setActiveEditorNodeID] = useState('');

const canvasNodeTypes = useMemo(() => ({
  agent: props => <CanvasNode
    {...props}
    isEditorOpen={activeEditorNodeID === props.id}
    onOpenEditor={setActiveEditorNodeID}
    onManualEdit={handleNodeContentDoubleClick}
    onPickImageAsset={openImageAssetPicker}
    onPickReferenceAsset={openReferenceAssetPanel}
    onRemoveReference={removeReference}
    onNodeConfigChange={handleNodeConfigChange}
    onSubmitVideo={submitVideoForNode}
    imageModelOptions={configuredImageModelOptions}
    videoModelOptions={configuredVideoModelOptions}
    modelLoadError={modelLoadError}
  />
}), [activeEditorNodeID, configuredImageModelOptions, configuredVideoModelOptions, handleNodeContentDoubleClick, modelLoadError]);
```

Apply these event rules:

- Media preview activation calls `onOpenEditor(id)` and does not commit canvas data.
- Image source asset selection moves to an explicit editor button labelled `添加图片素材` or `更换图片素材`.
- `onPaneClick` closes the active editor as well as context/drop menus.
- `Escape` closes context menus and the active editor in the same key action.
- Clicking a text node clears the active media editor.
- Media-node double-click returns without opening `manualNode`; text double-click continues to call `openManualNode`.
- `toFlowNode` normalizes `node.type` before rendering, and `toCanvas` normalizes `node.data.type` before serialization, so the next content-changing save writes legacy `copy`, blank, and unknown types as `text`.
- When a genuinely different external canvas is loaded, clear `activeEditorNodeID`; ordinary echoed `onChange` data continues to be ignored by the existing signature check.
- Editor controls keep `nodrag nopan` and stop pointer/double-click propagation.

Use `useUpdateNodeInternals` inside `CanvasNode`, where React Flow context exists:

```jsx
const updateNodeInternals = useUpdateNodeInternals();
useLayoutEffect(() => {
  updateNodeInternals(id);
}, [id, isEditorOpen, updateNodeInternals]);
```

The manual modal remains for text title/content/status editing only; remove unreachable image/video/audio branches after Task 3 moves video job actions inline.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run the Step 2 command again.

Expected: both files PASS, with no assertion requiring permanent media panels or media double-click modals.

- [ ] **Step 5: Commit only the editor-state files**

```powershell
git add -- frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "feat: add collapsible agent media editors"
```

---

### Task 3: Preserve the Complete Video Job Workflow Inline

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Test: `frontend/src/user/pages/agent-canvas-video-jobs.test.js`

**Interfaces:**
- Consumes: existing `videoNodeViewState(job)` and Agent video API functions.
- Produces: node-ID-based handlers `cancelVideo(nodeID)`, `retryVideo(nodeID)`, and `confirmVideo(nodeID)`.
- Produces: inline video props `videoJob`, `videoState`, `videoBusy`, `videoActionError`, `onCancelVideo`, `onRetryVideo`, `onConfirmVideo`.

- [ ] **Step 1: Extend the source contract with the four working video actions**

```js
test('expanded video editor retains generate cancel retry and confirm actions', () => {
  for (const label of ['生成视频', '取消生成', '重试生成', '确认回填']) {
    assert.match(canvasSource, new RegExp(`>${label}<`));
  }
  assert.match(canvasSource, /videoNodeViewState\(videoJob\)/);
  assert.match(canvasSource, /onCancelVideo\?\.\(id\)/);
  assert.match(canvasSource, /onRetryVideo\?\.\(id\)/);
  assert.match(canvasSource, /onConfirmVideo\?\.\(id\)/);
});
```

- [ ] **Step 2: Run video tests and verify the new source contract fails**

```powershell
node --test src/user/pages/agent-canvas-collapsible-editor.source.test.js src/user/pages/agent-canvas-video-jobs.test.js
```

Expected: the existing pure video-job tests PASS; the new inline-action assertion FAILS because cancel/retry/confirm still depend on `manualNode`.

- [ ] **Step 3: Refactor video handlers to use the node ID and render status inline**

Resolve the job with `videoJobs[nodeID]`, pass it to `CanvasNode`, and render the existing result preview, error text, status, and allowed actions beneath the video prompt toolbar. Refactor handlers from `manualNode.id` to their explicit argument:

```jsx
async function cancelVideo(nodeID) {
  const job = videoJobs[nodeID];
  if (!taskId || !job) return;
  setVideoBusy(true);
  setVideoActionError('');
  try {
    const result = await cancelAgentVideoJob(taskId, job.id);
    setVideoJobs(current => ({ ...current, [nodeID]: result.job }));
  } catch (error) {
    setVideoActionError(error?.message || '取消视频任务失败');
  } finally {
    setVideoBusy(false);
  }
}

async function retryVideo(nodeID) {
  const job = videoJobs[nodeID];
  if (!taskId || !job) return;
  setVideoBusy(true);
  setVideoActionError('');
  try {
    const result = await retryAgentVideoJob(taskId, job.id, makeVideoJobIdempotencyKey());
    setVideoJobs(current => ({ ...current, [nodeID]: result.job }));
  } catch (error) {
    setVideoActionError(error?.message || '重试视频任务失败');
  } finally {
    setVideoBusy(false);
  }
}

async function confirmVideo(nodeID) {
  const job = videoJobs[nodeID];
  if (!taskId || !job) return;
  setVideoBusy(true);
  setVideoActionError('');
  try {
    const result = await confirmAgentVideoJob(taskId, job.id);
    setVideoJobs(current => ({ ...current, [nodeID]: result.job }));
    await onServerCanvasChange?.();
  } catch (error) {
    setVideoActionError(error?.message || '确认回填失败');
  } finally {
    setVideoBusy(false);
  }
}
```

Do not submit any job from tests. Remove media-specific fields and video job controls from the manual text modal once the inline equivalents exist.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run the Step 2 command again.

Expected: source contract and video-job state tests PASS.

- [ ] **Step 5: Commit only the video parity change**

```powershell
git add -- frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js
git commit -m "fix: keep agent video actions in inline editor"
```

---

### Task 4: Attach Conditional Connection Handles to the Preview Shell

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Consumes: React Flow `Handle`, `Position`, `selected`, and root `.is-connecting` state.
- Produces: `.agent-canvas-node-preview-shell` as the single geometry owner for preview content and both handles.
- Preserves: `connectionRadius={24}`, duplicate/self-edge rejection, empty-drop menu, and reference synchronization.

- [ ] **Step 1: Add failing structure and CSS assertions**

```js
test('connection handles belong to the visible preview shell and stay hidden until relevant', () => {
  assert.match(canvasSource, /className="agent-canvas-node-preview-shell"[\s\S]*?<Handle type="target"[\s\S]*?<Handle type="source"/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-node-preview-shell\s*\{[^}]*position:\s*relative/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-node-quick-add\s*\{[^}]*opacity:\s*0[^}]*pointer-events:\s*none/);
  assert.match(canvasStyles, /\.react-flow__node\.selected[\s\S]*?\.agent-canvas-node-quick-add/);
  assert.match(canvasStyles, /\.agent-canvas\.is-connecting[\s\S]*?\.agent-canvas-node-target/);
  assert.match(canvasStyles, /\.agent-canvas-node-target\.connecting\.valid/);
});
```

Update context-menu tests so they require real source/target handles inside the shell rather than always visible ports.

- [ ] **Step 2: Run the handle tests and verify they fail**

```powershell
node --test src/user/pages/agent-canvas-collapsible-editor.source.test.js src/user/pages/agent-canvas-context-menu.test.js src/user/pages/agent-canvas-commands.test.js
```

Expected: FAIL because handles are siblings of `.agent-canvas-node-card`, media wrappers are wider than previews, and source handles are always interactive.

- [ ] **Step 3: Move handles and replace oversized transparent geometry**

Render both handles inside `.agent-canvas-node-preview-shell`. The image preview activates the editor; its separate asset button remains in the editor from Task 2. Preserve the current video, audio, and text preview branches inside `.agent-canvas-node-card`:

```jsx
<article className={`agent-canvas-node agent-canvas-node--${nodeType} agent-canvas-node--${data.status}`}>
  <div className="agent-canvas-node-preview-shell">
    <Handle type="target" position={Position.Left} className="agent-canvas-node-target" aria-label="连接到此节点" />
    <div className="agent-canvas-node-card">
      <header><span>{nodeType}</span><strong>{data.title}</strong></header>
      {nodeType === 'image' ? <button type="button" className="agent-canvas-node-image-select" aria-label="展开图片编辑" onClick={() => onOpenEditor?.(id)}>
        {data.assetRef?.url ? <img className="agent-canvas-node-media" src={data.assetRef.url} alt="" /> : <span className="agent-canvas-node-image-placeholder" aria-label="待生成图片"><ImagePlus size={27} /></span>}
      </button> : null}
      {nodeType === 'video' && data.assetRef?.url ? <video className="agent-canvas-node-media" src={data.assetRef.url} preload="metadata" muted controls /> : null}
      {nodeType === 'video' && !data.assetRef?.url ? <button type="button" className="agent-canvas-node-video-placeholder" onClick={() => onOpenEditor?.(id)}>等待生成视频</button> : null}
      {nodeType === 'audio' ? <button type="button" className="agent-canvas-node-audio-placeholder" onClick={() => onOpenEditor?.(id)}>音频设置</button> : null}
      {nodeType === 'text' ? <p>{data.summary || '等待 Agent 写入结果'}</p> : null}
      <footer><span>v{data.version}</span><span>{data.status}</span></footer>
    </div>
    <Handle type="source" position={Position.Right} className="agent-canvas-node-quick-add" aria-label="拖出连接" />
  </div>
  {generationPanel}
</article>
```

Define `generationPanel` immediately before the return by assigning the current complete `<section className="agent-canvas-generation-panel nodrag nopan">` block when `isEditorOpen && isMediaCanvasNodeType(nodeType)` is true, and `null` otherwise. Move that existing block without changing or duplicating its prompt, references, model settings, availability message, or submit action.

Set media node and editor width from the same `340px` outer width. Hide handles with `opacity: 0; pointer-events: none`; reveal the source on preview hover, node selection, or preview `:focus-within`; reveal targets only under `.agent-canvas.is-connecting`. Style React Flow's `.connecting.valid` target state with the accent glow, while invalid or duplicate targets remain dim and do not advertise acceptance. Keep handles mounted so React Flow geometry and keyboard accessibility stay stable. Use at most `8px` corner radii and retain the current V88 palette.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run the Step 2 command again.

Expected: all handle, connection, and source-contract tests PASS.

- [ ] **Step 5: Commit only the handle and styling files**

```powershell
git add -- frontend/src/user/components/AgentCanvas.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "fix: anchor agent canvas handles to previews"
```

---

### Task 5: Hide or Restore Existing Edges Without Data Loss

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Test: `frontend/src/user/pages/agent-canvas-appearance.test.js`

**Interfaces:**
- Consumes: `appearance.edgesVisible` from Task 1.
- Produces: `toFlowEdge(edge, appearance).hidden`.
- Produces: toolrail action labelled `隐藏连线` when visible and `显示连线` when hidden.
- Preserves: the complete `edgesRef.current` array and `toCanvas(...).edges` serialization.

- [ ] **Step 1: Add failing edge-visibility source assertions**

```js
test('edge visibility changes rendering without deleting serialized edges', () => {
  assert.match(canvasSource, /hidden:\s*!appearance\.edgesVisible/);
  assert.match(canvasSource, /appearance\.edgesVisible \? '隐藏连线' : '显示连线'/);
  assert.match(canvasSource, /updateAppearance\(\{ edgesVisible: !appearance\.edgesVisible \}\)/);
  assert.doesNotMatch(canvasSource, /appearance\.edgesVisible \? edges : \[\]/);
});
```

- [ ] **Step 2: Run appearance and source tests and verify they fail**

```powershell
node --test src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-collapsible-editor.source.test.js
```

Expected: appearance normalization PASS from Task 1; source contract FAILS because flow edges and the toolrail do not yet use the property.

- [ ] **Step 3: Apply visual-only edge hiding and add the icon tool**

Add the property in `toFlowEdge`:

```js
return {
  ...edge,
  type: 'canvas',
  hidden: !appearance.edgesVisible,
  markerEnd: { type: MarkerType.ArrowClosed, color: style.stroke },
  data: { relation: edge.relation || edge.data?.relation || 'reference', appearance }
};
```

Add a Lucide `Eye`/`EyeOff` tool beside the connection tool. It calls `updateAppearance({ edgesVisible: !appearance.edgesVisible })`; never call `setEdges([])` and never filter `edgesRef.current`. The temporary drag line remains React Flow connection state and therefore remains visible.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run the Step 2 command again.

Expected: all appearance and visibility tests PASS.

- [ ] **Step 5: Commit only the edge-visibility files**

```powershell
git add -- frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js
git commit -m "feat: toggle agent canvas edge visibility"
```

---

### Task 6: Full Regression, Embedded Go Rebuild, and Real Browser Acceptance

**Files:**
- Verify only: all files changed by Tasks 1-5.
- Generated and ignored build output may change locally, but must not be used as source implementation or included in a source commit.

**Interfaces:**
- Consumes: the complete implementation from Tasks 1-5.
- Produces: verified local Go-embedded runtime on `http://localhost:4001`.
- Produces: browser evidence for the exact task `agent-381900adb37b8663b49729b2` without provider submission.

- [ ] **Step 1: Run all focused Agent canvas tests**

```powershell
node --test src/user/pages/agent-canvas-node-presentation.test.js src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-canvas-context-menu.test.js src/user/pages/agent-canvas-collapsible-editor.source.test.js src/user/pages/agent-canvas-video-jobs.test.js src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-workspace.layout.test.js
```

Expected: all focused tests PASS with zero failures.

- [ ] **Step 2: Build the React source**

```powershell
npm run build
```

Run from `frontend`.

Expected: Vite exits `0`; no source or CSS compilation error.

- [ ] **Step 3: Rebuild and restart only the local Go review runtime**

From the repository root:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1
```

Before restart, identify the listener owning port `4001`; do not stop unrelated processes. Expected: the rebuilt `backend/qiantie-local.exe` owns `4001`, `/health` passes, and `/api/runtime-build-info` reports the current local Git SHA and service `go-api`.

- [ ] **Step 4: Perform authenticated browser acceptance on the real canvas**

Open:

```text
http://localhost:4001/agent/canvas?taskId=agent-381900adb37b8663b49729b2
```

Verify all of the following in the rendered page:

1. Initial media nodes show previews only; no detached floating `+` controls appear.
2. Hovering or selecting a preview reveals one right-side `拖出连接` handle on the preview border.
3. Starting a drag reveals left-side targets on other preview borders.
4. Dropping on a valid node creates one edge and does not open the empty-drop menu; undo returns to the initial edge count.
5. Dropping on blank canvas still offers exactly `文本` and `图片`; dismiss without creating if not needed.
6. Clicking an image/video/audio preview opens its inline editor; another preview switches the single open editor.
7. Blank-canvas click and `Escape` close the editor; clicking prompt/model/native media controls does not close it.
8. The explicit image asset action opens the project material panel; close it without uploading or choosing an asset.
9. A video node still shows the correct existing job state and only the actions allowed by `videoNodeViewState`; do not click `生成视频`, `重试生成`, or any provider action.
10. `隐藏连线` makes saved edges disappear and changes to `显示连线`; restoring it returns the same edge count.
11. Repeat handle/editor checks at zoom `100%` and one non-default zoom level.
12. Browser console and failed-network list contain no new errors caused by this change.
13. Record the canvas revision/save indicator before and after opening and closing an editor without changing a field; it must not create a save or history entry.

If a disposable node or edge is created for testing, use undo/delete and confirm the final saved node/edge counts match the initial counts before finishing.

- [ ] **Step 5: Inspect scope and commit any final test-only correction**

```powershell
git diff --check
git status --short
git diff --name-only HEAD~5..HEAD
```

Expected: no whitespace errors; no migration, backend, deployment, unrelated dirty, or generated `frontend/dist` file was added to these commits. If browser acceptance exposes a defect, return to the task that owns that behavior, add a failing regression assertion to its named test file, apply the correction to its named source file, rerun Steps 1-4, and use that task's explicit `git add` command. Do not create an empty or catch-all commit when no correction was required.

---

## Completion Evidence

Report all of these separately rather than treating one as proof of another:

- Commit SHAs for each implementation task.
- Focused test count and zero-failure result.
- React build result.
- Go binary rebuild and local runtime SHA.
- Authenticated browser interaction result for editor switching, attached handles, connection drag/undo, asset-panel opening, edge hide/restore, zoom, and browser errors.
- Explicit statement that no provider generation was submitted and no public deployment occurred.
