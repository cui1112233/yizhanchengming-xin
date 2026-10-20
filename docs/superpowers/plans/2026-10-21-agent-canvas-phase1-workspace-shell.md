# Agent Canvas Phase 1: Workspace Shell and Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the existing Agent canvas into a stable workflow workspace with shared project navigation, a resizable node/assets drawer, searchable navigation, and complete canvas controls.

**Architecture:** Keep `AgentPageV2` as the project owner and `AgentCanvas` as the React Flow host. Add pure state helpers for view/panel persistence and navigation filtering, then compose focused header, drawer, and control components around the existing canvas without replacing its save queue or generation behavior.

**Tech Stack:** React 18, Ant Design 5, `@xyflow/react`, Lucide React, Node test runner, Go embedded frontend.

## Global Constraints

- Work directly on branch `v88`; do not create a branch and do not reset unrelated dirty files.
- The canonical source is React under `frontend/src`; compiled output is never copied back over source.
- The application runtime remains the Go embedded binary; no Node release path is added.
- Keep the existing `taskId`; view changes use `view=workflow|storyboard` and never create a second project.
- Middle-button drag and Space plus left-button drag pan the canvas.
- Shortcut help lists only commands that are implemented and tested.
- Do not deploy or push publicly during this local implementation plan.

---

## File Map

- `frontend/src/user/pages/agent-workspace-view.js`: normalize and serialize workspace view/panel state.
- `frontend/src/user/pages/agent-workspace-navigation.js`: pure node/asset search, filters, and group-tree projection.
- `frontend/src/user/components/AgentWorkspaceHeader.jsx`: project name, view switch, save state, counts, and panel toggles.
- `frontend/src/user/components/AgentWorkspaceDrawer.jsx`: resizable node/assets navigation drawer.
- `frontend/src/user/components/AgentCanvasControls.jsx`: minimap/display/zoom/layout controls and shortcut help.
- `frontend/src/user/pages/AgentPageV2.jsx`: owns shared project layout and URL view state.
- `frontend/src/user/components/AgentCanvas.jsx`: consumes workspace display settings and exposes focus/layout callbacks.
- `frontend/src/shared/styles/agent-workspace.css`: full-bleed layout and stable responsive dimensions.

### Task 1: Durable workspace view state

**Files:**
- Create: `frontend/src/user/pages/agent-workspace-view.js`
- Create: `frontend/src/user/pages/agent-workspace-view.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-state.js`
- Modify: `frontend/src/user/pages/agent-canvas-state.test.js`

**Interfaces:**
- Consumes: persisted canvas object from `getAgentCanvas(taskId)`.
- Produces: `normalizeWorkspaceView(value) -> WorkspaceViewState`, `workspaceViewPatch(current, patch) -> WorkspaceViewState`, and `workspaceViewQuery(value) -> 'workflow'|'storyboard'`.
- `WorkspaceViewState` shape: `{ activeView, drawerOpen, drawerTab, drawerWidth, conversationOpen, conversationWidth, snapToGrid, showMinimap, showConnections }`.

- [ ] **Step 1: Write failing normalization and save-queue tests**

```js
test('normalizes old canvases into the workflow workspace defaults', () => {
  assert.deepEqual(normalizeWorkspaceView(null), {
    activeView: 'workflow', drawerOpen: true, drawerTab: 'nodes', drawerWidth: 288,
    conversationOpen: true, conversationWidth: 360,
    snapToGrid: false, showMinimap: true, showConnections: true,
  });
});

test('merging a pending save retains local view state', () => {
  const merged = mergePendingCanvas(
    { revision: 4, viewState: { activeView: 'workflow' } },
    { revision: 3, viewState: { activeView: 'storyboard', drawerWidth: 340 } },
  );
  assert.equal(merged.viewState.activeView, 'storyboard');
  assert.equal(merged.viewState.drawerWidth, 340);
});
```

- [ ] **Step 2: Run the tests and verify the missing helper fails**

Run: `node --test frontend/src/user/pages/agent-workspace-view.test.js frontend/src/user/pages/agent-canvas-state.test.js`

Expected: FAIL because `normalizeWorkspaceView` is absent and the save merge does not preserve `viewState`.

- [ ] **Step 3: Implement normalization and merge behavior**

```js
export const DEFAULT_WORKSPACE_VIEW = Object.freeze({
  activeView: 'workflow', drawerOpen: true, drawerTab: 'nodes', drawerWidth: 288,
  conversationOpen: true, conversationWidth: 360,
  snapToGrid: false, showMinimap: true, showConnections: true,
});

export function normalizeWorkspaceView(value = {}) {
  const source = value && typeof value === 'object' ? value : {};
  return {
    ...DEFAULT_WORKSPACE_VIEW,
    ...source,
    activeView: source.activeView === 'storyboard' ? 'storyboard' : 'workflow',
    drawerWidth: Math.min(460, Math.max(240, Number(source.drawerWidth) || 288)),
    conversationWidth: Math.min(520, Math.max(320, Number(source.conversationWidth) || 360)),
  };
}
```

Export `mergePendingCanvas` for its test and merge `viewState` from the newest local snapshot while accepting the server revision.

- [ ] **Step 4: Run focused tests**

Run: `node --test frontend/src/user/pages/agent-workspace-view.test.js frontend/src/user/pages/agent-canvas-state.test.js`

Expected: PASS.

- [ ] **Step 5: Commit the state contract**

```powershell
git add frontend/src/user/pages/agent-workspace-view.js frontend/src/user/pages/agent-workspace-view.test.js frontend/src/user/pages/agent-canvas-state.js frontend/src/user/pages/agent-canvas-state.test.js
git commit -m "feat(agent): persist workspace view state"
```

### Task 2: Shared header and workflow/storyboard URL contract

**Files:**
- Create: `frontend/src/user/components/AgentWorkspaceHeader.jsx`
- Create: `frontend/src/user/pages/agent-workspace-header.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `WorkspaceViewState`, active task, save status, node count, and material count.
- Produces: `onViewChange(view)`, `onToggleDrawer()`, `onToggleConversation()`, and existing rename/save actions.

- [ ] **Step 1: Write the failing source contract test**

```js
test('header exposes both views and keeps the task id in the URL', () => {
  const header = read('frontend/src/user/components/AgentWorkspaceHeader.jsx');
  const page = read('frontend/src/user/pages/AgentPageV2.jsx');
  assert.match(header, /工作流/);
  assert.match(header, /故事板/);
  assert.match(page, /searchParams\.set\('view', view\)/);
  assert.match(page, /taskId/);
});
```

- [ ] **Step 2: Run the source test and verify failure**

Run: `node --test frontend/src/user/pages/agent-workspace-header.source.test.js`

Expected: FAIL because the shared header and stable `view` query update do not exist.

- [ ] **Step 3: Implement the header and URL synchronization**

```jsx
<Segmented
  value={viewState.activeView}
  options={[{ label: '工作流', value: 'workflow' }, { label: '故事板', value: 'storyboard' }]}
  onChange={onViewChange}
/>
```

In `AgentPageV2`, update only the `view` query key with `history.replaceState`; preserve `taskId` and all unrelated query keys. Send the view patch through the existing canvas save queue.

- [ ] **Step 4: Run tests and build**

Run: `node --test frontend/src/user/pages/agent-workspace-header.source.test.js frontend/src/user/pages/agent-workspace-view.test.js; npm --prefix frontend run build`

Expected: tests PASS and Vite build succeeds.

- [ ] **Step 5: Commit the shared header**

```powershell
git add frontend/src/user/components/AgentWorkspaceHeader.jsx frontend/src/user/pages/agent-workspace-header.source.test.js frontend/src/user/pages/AgentPageV2.jsx frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): add shared workspace header"
```

### Task 3: Searchable node and asset drawer

**Files:**
- Create: `frontend/src/user/pages/agent-workspace-navigation.js`
- Create: `frontend/src/user/pages/agent-workspace-navigation.test.js`
- Create: `frontend/src/user/components/AgentWorkspaceDrawer.jsx`
- Create: `frontend/src/user/pages/agent-workspace-drawer.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: canvas nodes/groups, `listAgentAssets(taskId, kind)`, search text, type/status/rating filters.
- Produces: `filterWorkspaceNodes(nodes, filters)`, `buildWorkspaceGroupTree(nodes, groups)`, `onLocateNode(id)`, and `onCreateFromAsset(asset)`.

- [ ] **Step 1: Write failing filter and group-tree tests**

```js
test('filters titles and summaries without losing group identity', () => {
  const result = filterWorkspaceNodes(nodes, { query: '雨夜', kinds: ['image'], statuses: [], ratings: [] });
  assert.deepEqual(result.map(item => item.id), ['image-2']);
  assert.equal(buildWorkspaceGroupTree(result, groups)[0].id, 'group-a');
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/pages/agent-workspace-drawer.source.test.js`

Expected: FAIL because navigation helpers and drawer do not exist.

- [ ] **Step 3: Implement the drawer**

```jsx
<Tabs activeKey={viewState.drawerTab} items={[
  { key: 'nodes', label: '画布节点', children: <NodeTree /> },
  { key: 'assets', label: '项目素材', children: <AssetTree /> },
]} />
```

Use pointer capture for the resize handle, clamp width to `240..460`, and persist the final width. Locating a node calls a single canvas callback; dragging an asset passes its durable asset ID and kind to node creation.

- [ ] **Step 4: Run focused tests and build**

Run: `node --test frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/pages/agent-workspace-drawer.source.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit the navigation drawer**

```powershell
git add frontend/src/user/pages/agent-workspace-navigation.js frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/components/AgentWorkspaceDrawer.jsx frontend/src/user/pages/agent-workspace-drawer.source.test.js frontend/src/user/pages/AgentPageV2.jsx frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): add canvas navigation drawer"
```

### Task 4: Canvas controls, minimap, snap, and automatic layout

**Files:**
- Create: `frontend/src/user/components/AgentCanvasControls.jsx`
- Create: `frontend/src/user/pages/agent-canvas-layout.js`
- Create: `frontend/src/user/pages/agent-canvas-layout.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-commands.js`
- Modify: `frontend/src/user/pages/agent-canvas-commands.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: normalized nodes/edges and workspace display flags.
- Produces: `layoutCanvas(nodes, edges, direction) -> nodes`, focus/fit callbacks, `snapToGrid`, `showMinimap`, and `showConnections` behavior.

- [ ] **Step 1: Write failing deterministic layout and shortcut tests**

```js
test('lays a connected chain left to right without moving locked nodes', () => {
  const result = layoutCanvas(nodes, edges, 'LR');
  assert.equal(result.find(node => node.id === 'locked').position.x, 20);
  assert.ok(result.find(node => node.id === 'b').position.x > result.find(node => node.id === 'a').position.x);
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-layout.test.js frontend/src/user/pages/agent-canvas-commands.test.js`

Expected: FAIL because `layoutCanvas` and the new tested control commands are absent.

- [ ] **Step 3: Implement controls and deterministic layout**

```jsx
<ReactFlow
  panOnDrag={[1, 2]}
  panOnScroll={false}
  snapToGrid={viewState.snapToGrid}
  snapGrid={[16, 16]}
>
  {viewState.showMinimap ? <MiniMap pannable zoomable /> : null}
</ReactFlow>
```

Implement a stable layer pass using typed edges, preserve locked-node positions, and record one undo snapshot for an automatic layout. Shortcut help is generated from the same command registry used by `canvasKeyboardCommand`.

- [ ] **Step 4: Run phase tests and build**

Run: `node --test frontend/src/user/pages/agent-workspace-view.test.js frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/pages/agent-canvas-layout.test.js frontend/src/user/pages/agent-canvas-commands.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit canvas controls**

```powershell
git add frontend/src/user/components/AgentCanvasControls.jsx frontend/src/user/pages/agent-canvas-layout.js frontend/src/user/pages/agent-canvas-layout.test.js frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-commands.js frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): complete canvas navigation controls"
```

### Task 5: Anchored connection handles and typed reference synchronization

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-connections.js`
- Create: `frontend/src/user/pages/agent-canvas-connections.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: source/target node kinds, existing edges, manual references, edge-derived references, and the blank-canvas release coordinate.
- Produces: `createTypedCanvasEdge(connection, nodes)`, `connectionDropChoices(sourceNode)`, and `syncEffectiveReferences(node, edges, nodes)`.
- Initial relation values: `reference`, `first_frame`, `last_frame`, `audio_source`, `script_source`, and `sequence`.

- [ ] **Step 1: Write failing handle, drop-menu, and reference-ownership tests**

```js
test('blank drop offers compatible nodes and connected assets become edge references', () => {
  assert.deepEqual(connectionDropChoices(imageNode).map(item => item.kind), ['image', 'video']);
  const edge = createTypedCanvasEdge({ source: 'image-1', target: 'video-1' }, [imageNode, videoNode]);
  assert.equal(edge.data.relation, 'first_frame');
  const target = syncEffectiveReferences(videoNode, [edge], [imageNode, videoNode]);
  assert.deepEqual(target.data.config.edgeReferences.map(item => item.assetId), ['asset-image-1']);
});

test('disconnecting removes only edge references and retains manually selected assets', () => {
  const target = syncEffectiveReferences(videoNodeWithManualReference, [], nodes);
  assert.deepEqual(target.data.config.manualReferences.map(item => item.assetId), ['manual-1']);
  assert.deepEqual(target.data.config.edgeReferences, []);
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-connections.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: FAIL because typed connection creation and separate manual/edge reference ownership do not exist.

- [ ] **Step 3: Implement anchored handles and blank-drop creation**

```jsx
<Handle type="source" position={Position.Right} className="agent-node__output-handle" aria-label="拉出连接" />
```

Keep the `+` visual inside the right-edge handle box in every collapsed/expanded state. Start a live edge preview on drag; create a typed edge when dropped on a compatible handle; when dropped on blank canvas, open image/video quick choices at the release coordinate plus `更多节点`. Creating the chosen node immediately connects it and persists both changes in one canvas revision.

- [ ] **Step 4: Run connection/editor tests and build**

Run: `node --test frontend/src/user/pages/agent-canvas-connections.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-node-composer.source.test.js; npm --prefix frontend run build`

Expected: PASS, build succeeds, and the `+` handle has no absolute offset outside the node boundary.

- [ ] **Step 5: Commit connection behavior**

```powershell
git add frontend/src/user/pages/agent-canvas-connections.js frontend/src/user/pages/agent-canvas-connections.test.js frontend/src/user/pages/agent-canvas-context-menu.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/AgentNodeComposer.jsx frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): connect nodes with typed references"
```

### Task 6: Phase 1 Go-embedded browser acceptance

**Files:**
- Create: `test/agent-workspace-shell.source.test.js`
- Modify: `docs/superpowers/specs/2026-10-21-agent-canvas-liblib-workspace-design.md`

**Interfaces:**
- Consumes: the complete Phase 1 workspace.
- Produces: verified local behavior at `/agent/canvas?taskId=<id>&view=workflow` in the rebuilt Go binary.

- [ ] **Step 1: Add the release-source guard**

```js
test('workspace shell is sourced from React and Go embed remains the release path', () => {
  assert.match(read('frontend/src/user/pages/AgentPageV2.jsx'), /AgentWorkspaceHeader/);
  assert.match(read('scripts/deploy-go-embedded.ps1'), /go/i);
});
```

- [ ] **Step 2: Run all Phase 1 checks**

Run: `$workspaceTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-workspace-*.test.js').FullName; node --test test/agent-workspace-shell.source.test.js $workspaceTests frontend/src/user/pages/agent-canvas-layout.test.js frontend/src/user/pages/agent-canvas-connections.test.js frontend/src/user/pages/agent-canvas-commands.test.js; npm --prefix frontend run build; Set-Location backend; go test ./internal/httpapi ./internal/agentworkspace; go build ./cmd/api`

Expected: all commands succeed.

- [ ] **Step 3: Rebuild and start the local Go review runtime**

```powershell
Set-Location F:\脚本测试\一战晟铭-v88
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql
.\scripts\check-local-go.ps1
```

Expected: the review service listens on `http://localhost:4001` and reports the newly built local SHA.

- [ ] **Step 4: Perform authenticated browser acceptance**

Verify the real UI can switch views without losing `taskId`, resize/reopen both panels, search and locate a node, drag an asset onto the canvas, toggle minimap/grid/connections, run automatic layout, pan with the middle button, save, reload, and restore all settings. Capture desktop and narrow-width screenshots and confirm no toolbar, drawer, node, or conversation overlap.

- [ ] **Step 5: Commit acceptance evidence**

```powershell
git add test/agent-workspace-shell.source.test.js docs/superpowers/specs/2026-10-21-agent-canvas-liblib-workspace-design.md
git commit -m "test(agent): verify workspace shell acceptance"
```
