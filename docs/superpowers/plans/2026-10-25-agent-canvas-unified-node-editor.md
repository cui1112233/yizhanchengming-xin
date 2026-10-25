# Agent Canvas Unified Node Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make text, image, video, and audio nodes use one Liblib-inspired selected-node toolbar and one unified bottom editor while retaining only existing V88 functions.

**Architecture:** Keep selection, positioning, and callback ownership in `AgentCanvas.jsx`; keep generation configuration and persisted task state in `AgentNodeComposer.jsx`. Extract the selected-node toolbar presentation into the existing top-action component, reuse the current image/video/audio operation callbacks, and express all light/dark styling through scoped Agent workspace variables.

**Tech Stack:** React 18, Ant Design 5, Lucide React, XYFlow React 12.11.6, CSS, Node built-in test runner, Vite, Go embed.

## Global Constraints

- Work directly on the canonical `v88` branch; do not create a branch or alter unrelated dirty files.
- Do not add backend APIs, database changes, generation protocols, models, credits, role-library, effects-library, or camera-library features.
- All visible actions must call an existing V88 callback or open an existing V88 panel.
- Text, image, video, and audio nodes must use the same selected-node shell and single-click opening behavior.
- Node bodies continue to follow canvas zoom; selected-node toolbar and editor stay stable in screen pixels through inverse scaling.
- Node controls remain interactive, while every non-control node area remains draggable.
- Light and dark themes must retain readable backgrounds, text, borders, hover, disabled, success, warning, and failure states.
- Do not click a final generate action during browser acceptance.
- Rebuild React and the Go-embedded binary; do not start the historical Node runtime.

---

### Task 1: Lock the unified interaction contract with focused tests

**Files:**
- Create: `frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js`
- Modify: `frontend/src/user/components/AgentCanvasVideoEditor.source.test.js`
- Test: `frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js`

**Interfaces:**
- Consumes: `AgentCanvas.jsx` selected-node state and operation callbacks; `AgentNodeComposer.jsx` generation controls; `agent-workspace.css` canvas variables.
- Produces: source-level acceptance for `text`, `image`, `video`, and `audio`, shared classes `agent-node-floating-toolbar` and `agent-node-floating-composer`, and semantic editor color variables.

- [ ] **Step 1: Write the failing source contract test**

```js
test('all editable node kinds open on one click and share inverse-scaled surfaces', () => {
  for (const kind of ['text', 'image', 'video', 'audio']) {
    assert.match(canvasSource, new RegExp(`'${kind}'`));
  }
  assert.match(canvasSource, /EDITABLE_NODE_TYPES/);
  assert.match(canvasSource, /agent-node-floating-toolbar/);
  assert.match(canvasSource, /agent-node-floating-composer/);
  assert.match(styles, /agent-node-floating-toolbar[^}]*scale\(var\(--agent-canvas-inverse-zoom\)\)/);
  assert.match(styles, /agent-node-floating-composer[^}]*scale\(var\(--agent-canvas-inverse-zoom\)\)/);
});

test('editor surfaces use semantic theme variables instead of fixed dark fills', () => {
  assert.match(styles, /--agent-editor-surface:/);
  assert.match(styles, /--agent-editor-action:/);
  assert.doesNotMatch(styles, /agent-(?:node-composer|video-quick-tools|audio-quick-tools)[^}]*background:\s*#202020/);
  assert.doesNotMatch(styles, /agent-audio-waveform__bars[^}]*background:\s*#242424/);
});

test('interactive controls opt out of dragging but node surfaces remain draggable', () => {
  assert.match(canvasSource, /agent-canvas-node-drag-surface/);
  assert.match(canvasSource, /nodrag nopan/);
  assert.doesNotMatch(canvasSource, /dragHandle:\s*['"]\.agent-canvas-node-drag-handle/);
});
```

- [ ] **Step 2: Run the focused tests and confirm the new contract fails**

Run: `node --test frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js frontend/src/user/components/AgentCanvasInteraction.source.test.js`

Expected: FAIL because text/image/audio do not yet share the selected-node floating toolbar/composer classes and fixed dark fills still exist.

- [ ] **Step 3: Keep the existing video assertions compatible with the generalized shell**

Replace video-only class expectations with shared class expectations while retaining assertions for single-click opening, stable node size, video operations, and actual callbacks. Do not weaken assertions for `trim`, `crop`, `separate_audio`, `extract_frames`, download, cancel, retry, or confirm.

- [ ] **Step 4: Re-run the focused tests to preserve the red baseline**

Run: `node --test frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js frontend/src/user/components/AgentCanvasInteraction.source.test.js`

Expected: FAIL only on not-yet-implemented unified behavior; existing drag and video callback tests remain green.

- [ ] **Step 5: Commit the test contract**

```powershell
git add -- frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js
git commit -m "test(agent-canvas): define unified node editor contract"
```

### Task 2: Generalize selected-node opening and floating surfaces

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposerTopActions.jsx`
- Test: `frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js`

**Interfaces:**
- Consumes: `openNodeComposer(nodeID)`, `activeEditorNodeID`, existing `AgentImageTools`, `AgentVideoTools`, audio operation callbacks, and `--agent-canvas-inverse-zoom`.
- Produces: `EDITABLE_NODE_TYPES = new Set(['text', 'image', 'video', 'audio'])`, a shared selected-node quick-action array, and the classes `agent-node-floating-toolbar` and `agent-node-floating-composer`.

- [ ] **Step 1: Add the editable-kind gate and single-click behavior**

```jsx
const EDITABLE_NODE_TYPES = new Set(['text', 'image', 'video', 'audio']);

function canOpenNodeEditor(node) {
  return EDITABLE_NODE_TYPES.has(normalizeCanvasNodeType(node?.data?.type));
}
```

Use the gate in the existing node-click handler so a normal click opens all four editors. Preserve multi-select behavior and stop opening when the event originates from `button`, `a`, `input`, `textarea`, `select`, Ant popups, media controls, or `.nodrag` content.

- [ ] **Step 2: Replace video/audio-only toolbar placement with one selected-node toolbar mount**

Build action descriptors from existing callbacks:

```jsx
const quickActions = nodeType === 'image'
  ? imageQuickActions
  : nodeType === 'video'
    ? videoQuickActions
    : nodeType === 'audio'
      ? audioQuickActions
      : textQuickActions;
```

Render the descriptors through `AgentNodeComposerTopActions` inside `agent-node-floating-toolbar nodrag nopan`. Image actions call the existing image apply callback; video actions call the existing video apply/open-tools callbacks; audio actions call `applyAudioOperation`; text actions update the existing text generation type/config or open existing parameter controls.

- [ ] **Step 3: Generalize composer positioning without changing node geometry**

```jsx
{isEditorOpen && EDITABLE_NODE_TYPES.has(nodeType)
  ? <div className="agent-node-floating-composer nodrag nopan">{composer}</div>
  : composer}
```

Keep `--agent-canvas-inverse-zoom: 1 / Math.max(canvasZoom, 0.25)` on the selected node and leave pinned Batch Factory composers on their current inline path.

- [ ] **Step 4: Preserve drag behavior and interactive exclusions**

Keep the node preview/body classes as `agent-canvas-node-drag-surface`; keep editor, toolbar, media controls, inputs, and popovers under `nodrag nopan`. Do not configure XYFlow with a six-dot-only `dragHandle` selector.

- [ ] **Step 5: Run the focused interaction tests**

Run: `node --test frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js frontend/src/user/components/AgentCanvasInteraction.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

Expected: PASS for shared opening, inverse scaling, video operations, general node dragging, protected controls, and collapse behavior.

- [ ] **Step 6: Commit the interaction implementation**

```powershell
git add -- frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/AgentNodeComposerTopActions.jsx frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js
git commit -m "feat(agent-canvas): unify selected node tools"
```

### Task 3: Reshape the bottom editor around real node functions

**Files:**
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/pages/agent-node-composer-actions.js`
- Modify: `frontend/src/user/components/AgentNodeParameterPopover.jsx`
- Modify: `frontend/src/user/components/AgentNodeSettingsPanel.jsx`
- Test: `frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js`

**Interfaces:**
- Consumes: `compactComposerActionPlan`, `AgentModelPicker`, `AgentNodeParameterPopover`, `AgentNodeSettingsPanel`, persisted `generationState`, video job state, candidate arrays, and current `onSubmit`/retry/cancel/confirm callbacks.
- Produces: one continuous `agent-node-composer` surface with top function chips, prompt body, and `agent-node-composer__bottom-bar`.

- [ ] **Step 1: Add exact per-kind real-action assertions**

```js
for (const label of ['生成类型', '文字长度', '创作风格']) assert.match(composerSource, new RegExp(label));
for (const label of ['参考素材', '更换当前图片', '编辑图片', '查看实际输入']) assert.match(composerSource, new RegExp(label));
for (const label of ['处理视频', '截取', '裁剪', '提取音频', '抽取画面帧']) assert.match(composerSource, new RegExp(label));
for (const label of ['处理方式', '自定义切分', '智能切分', '变速']) assert.match(composerSource, new RegExp(label));
for (const fake of ['积分', '角色库', '特效库', '运镜库']) assert.doesNotMatch(composerSource, new RegExp(fake));
```

- [ ] **Step 2: Make the action plan explicit for all four kinds**

Extend the current descriptor builder so it returns only supported action IDs. Keep conditional actions conditional: `confirm-image` requires image candidates, `confirm-audio` requires audio candidates, `retry` requires retry state, and `upload-video-management` requires a publishable merge result.

- [ ] **Step 3: Render one continuous editor hierarchy**

Use this fixed order:

```jsx
<section className={`agent-node-composer agent-node-composer--${kind} nodrag nopan`}>
  <AgentNodeComposerTopActions primaryActions={topActions.direct} moreActions={topActions.more} />
  <div className="agent-node-composer__body">...</div>
  <div className="agent-node-composer__bottom-bar agent-node-composer__toolbar">...</div>
</section>
```

Keep references, input preview, errors, candidates, and tool-result previews inside the body. Keep prompt state local until blur and submit `draftPrompt`, not stale node data.

- [ ] **Step 4: Keep the bottom row connected to the existing configuration**

The bottom row must retain `AgentModelPicker`, `AgentNodeParameterPopover`, `AgentNodeSettingsPanel`, task state, and the existing `submitNodeGeneration` call. Style the submit control as a circular icon button but retain its accessible label and disabled reason.

- [ ] **Step 5: Keep persisted review actions reachable**

Retain visible or More-menu entries for image/audio candidate confirmation and video cancel/retry/confirm. Preserve current callbacks `onConfirmImage`, `onConfirmAudio`, `onCancelVideo`, `onRetryVideo`, and `onConfirmVideo` exactly.

- [ ] **Step 6: Run composer tests**

Run: `node --test frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js frontend/src/user/pages/agent-node-composer-actions.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

Expected: PASS with no loss of model selection, parameter/settings controls, candidate confirmation, retry/cancel, or persisted status.

- [ ] **Step 7: Commit the editor hierarchy**

```powershell
git add -- frontend/src/user/components/AgentNodeComposer.jsx frontend/src/user/pages/agent-node-composer-actions.js frontend/src/user/components/AgentNodeParameterPopover.jsx frontend/src/user/components/AgentNodeSettingsPanel.jsx frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js
git commit -m "feat(agent-canvas): unify node generation editor"
```

### Task 4: Apply the shared light/dark visual system

**Files:**
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js`

**Interfaces:**
- Consumes: existing `.agent-page` theme variables and shared toolbar/composer classes from Tasks 2-3.
- Produces: `--agent-editor-surface`, `--agent-editor-toolbar`, `--agent-editor-action`, `--agent-editor-action-hover`, `--agent-editor-border`, `--agent-editor-text`, `--agent-editor-muted`, and semantic status colors in both themes.

- [ ] **Step 1: Define semantic variables in both theme scopes**

```css
.agent-page {
  --agent-editor-surface: rgba(28, 29, 32, .98);
  --agent-editor-toolbar: rgba(24, 25, 28, .98);
  --agent-editor-action: rgba(255, 255, 255, .06);
  --agent-editor-action-hover: rgba(255, 255, 255, .11);
  --agent-editor-border: rgba(255, 255, 255, .12);
  --agent-editor-text: #f4f4f5;
  --agent-editor-muted: #a1a1aa;
}

[data-theme='light'] .agent-page {
  --agent-editor-surface: rgba(255, 255, 255, .98);
  --agent-editor-toolbar: rgba(250, 250, 250, .98);
  --agent-editor-action: #f4f4f5;
  --agent-editor-action-hover: #e9e9ec;
  --agent-editor-border: rgba(24, 24, 27, .14);
  --agent-editor-text: #18181b;
  --agent-editor-muted: #71717a;
}
```

Match the repository's actual theme selector if it differs; keep the semantic variable names unchanged.

- [ ] **Step 2: Style the toolbar as a compact selected-node surface**

Give `agent-node-floating-toolbar` an absolute anchor above the node, inverse scale, 38px target height, 7px radius, one-pixel semantic border, light shadow, horizontal overflow, and a high enough z-index to clear nodes and edges.

- [ ] **Step 3: Style the composer as one continuous surface**

Give `agent-node-floating-composer` the existing absolute node anchor and inverse scaling. Give `agent-node-composer` a 680px target width with viewport bounds, 15px radius, semantic background/border, one body column, and a continuous bottom row. Remove fixed `#202020` and `#242424` fills from toolbar/editor/waveform rules.

- [ ] **Step 4: Finish responsive and Ant Design states**

Keep action and bottom rows horizontally scrollable on narrow widths; keep text uncompressed; apply scoped foreground/background variables to Ant Select, Dropdown, Tooltip, and Popover content. Ensure disabled controls have visible border/icon contrast and pointer feedback.

- [ ] **Step 5: Run style and full focused tests**

Run: `node --test frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js frontend/src/user/components/AgentCanvasVideoEditor.source.test.js frontend/src/user/components/AgentCanvasInteraction.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-node-composer-actions.test.js`

Expected: PASS.

- [ ] **Step 6: Commit the visual system**

```powershell
git add -- frontend/src/shared/styles/agent-workspace.css frontend/src/user/components/AgentCanvasUnifiedEditor.source.test.js
git commit -m "style(agent-canvas): polish unified node editor"
```

### Task 5: Build, embed, restart, and verify the live UI

**Files:**
- Modify through build only: `frontend/dist/**`
- Modify through build only: Go embedded runtime artifacts recorded by `scripts/build-go-local.ps1`
- Verify: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: completed React source and the approved Go-only release scripts.
- Produces: a running Go-embedded V88 UI whose runtime SHA matches the final task commit.

- [ ] **Step 1: Run the focused and broader frontend tests separately**

Run focused command from Task 4 and record its exact pass/fail count.

Run: `npm test --prefix frontend`

Expected: report current results separately; do not revert current interaction to satisfy historical source assertions.

- [ ] **Step 2: Build React**

Run: `npm run build --prefix frontend`

Expected: Vite exits 0 and updates `frontend/dist` from current source.

- [ ] **Step 3: Commit only task-owned source and generated frontend assets**

Review with `git diff --name-only` and stage only the files created or changed by this plan plus the newly generated matching `frontend/dist` assets. Preserve all unrelated dirty files.

```powershell
git commit -m "feat(agent-canvas): deliver unified node editor"
```

- [ ] **Step 4: Build and restart the Go-embedded runtime**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`

Expected: `backend/qiantie-local.exe` is the application listener on port 4001; no legacy Node application service is started.

- [ ] **Step 5: Verify runtime identity**

Run: `Invoke-WebRequest http://localhost:4001/health -UseBasicParsing`

Expected: HTTP 200.

Run: `Invoke-RestMethod http://localhost:4001/api/runtime-build-info | ConvertTo-Json -Depth 6`

Expected: `runtime` is `go`, and both `git_sha` and `runtime_manifest_sha` equal `git rev-parse HEAD`.

- [ ] **Step 6: Perform non-generating browser acceptance**

Open `/agent/canvas?taskId=agent-bf-dbba32e978f13f1909e3b928&view=workflow`. In dark and light themes, click one text, image, video, and audio node; confirm the correct real toolbar and unified editor appear. At 100%, about 70%, and about 45% canvas zoom, confirm toolbar/editor pixel size remains stable. Drag each node from an empty area; then verify text input, video/audio playback, parameter popover, settings panel, group collapse, canvas tools, and edge controls remain usable. Do not click the final generate button.

- [ ] **Step 7: Record final evidence**

Report focused tests, broader historical failures if any, frontend build, Go build/restart, health, SHA identity, dark/light checks, zoom checks, and the fact that no generation request was submitted.

