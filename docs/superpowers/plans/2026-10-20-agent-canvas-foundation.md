# Agent Canvas Interaction Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the local V88 Agent canvas operate as a clean, keyboard-accessible production workspace instead of a form-filled node board.

**Architecture:** Keep `@xyflow/react` as the rendering engine and keep the existing Go canvas revision API unchanged. Put deterministic keyboard command parsing in `agent-canvas-commands.js`; let `AgentCanvas.jsx` apply those commands to its existing node, edge, history, clipboard, and viewport state. Node cards become compact previews and retain the existing modal as their inspector until the dedicated inspector phase.

**Tech Stack:** React 18, `@xyflow/react` 12, Ant Design, Lucide, Node test runner, Go embedded frontend.

## Global Constraints

- Local V88 only; do not publish publicly.
- Preserve unrelated dirty-worktree changes. Do not reset, clean, or overwrite compiled output as source.
- Use test-first changes and retain the existing MySQL/Redis/TOS contracts.
- Do not claim image generation is available: its current submit control remains disabled until its Go/executor lifecycle exists.
- Build React first, then rebuild the Go embedded binary and verify authenticated browser interaction on local port 4001.
- Do not create a branch, commit, or push during this task.

---

### Task 1: Define complete keyboard command parsing

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-commands.js`
- Test: `frontend/src/user/pages/agent-canvas-commands.test.js`

**Interfaces:**
- Consumes: browser key event fields plus `targetTagName` and `isContentEditable`.
- Produces: `canvasKeyboardCommand(event)` returning one of `select_all`, `copy`, `paste`, `duplicate`, `delete`, `undo`, `redo`, `group`, `ungroup`, `fit_canvas`, `focus_selection`, `zoom_in`, `zoom_out`, or `null`.

- [ ] **Step 1: Write failing command tests**

```js
assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'a', targetTagName: 'div' }), 'select_all');
assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'd', targetTagName: 'div' }), 'duplicate');
assert.equal(canvasKeyboardCommand({ shiftKey: true, key: '1', targetTagName: 'div' }), 'fit_canvas');
assert.equal(canvasKeyboardCommand({ key: 'Delete', targetTagName: 'div' }), 'delete');
assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'a', targetTagName: 'textarea' }), null);
```

- [ ] **Step 2: Run the focused test and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-commands.test.js`

Expected: failure because the new commands currently return `null`.

- [ ] **Step 3: Implement the minimal command map**

```js
if (key === 'a' && modifier) return 'select_all';
if (key === 'd' && modifier) return 'duplicate';
if (key === 'g' && modifier && shift) return 'ungroup';
if (key === 'g' && modifier) return 'group';
if (shift && key === '1') return 'fit_canvas';
if (shift && key === '2') return 'focus_selection';
```

Map copy, paste, delete, plus/minus, undo, and redo in the same pure function. Return `null` before all mappings when the target is editable.

- [ ] **Step 4: Re-run the focused test**

Run: `node --test frontend/src/user/pages/agent-canvas-commands.test.js`

Expected: all command tests pass.

### Task 2: Apply keyboard commands and permanent pan behavior

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Consumes: `canvasKeyboardCommand`, existing history, clipboard, node/edge actions, and `flowInstance`.
- Produces: working canvas shortcuts and middle-button panning without disabling normal left-button selection.

- [ ] **Step 1: Write failing source/behavior tests**

```js
assert.match(canvasSource, /panOnDrag=\{\[[^\]]*1[^\]]*\]\}/);
assert.match(canvasSource, /case 'select_all'/);
assert.match(canvasSource, /case 'duplicate'/);
assert.match(canvasSource, /case 'fit_canvas'/);
assert.match(canvasSource, /case 'focus_selection'/);
```

- [ ] **Step 2: Run the focused source test and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: failure because only undo/redo are currently wired and `panOnDrag` depends entirely on the mode switch.

- [ ] **Step 3: Implement command application**

Add one `applyKeyboardCommand(command)` switch in `AgentCanvas`.

```js
case 'select_all':
  setSelectedIds(nodesRef.current.map(node => node.id));
  return;
case 'fit_canvas':
  flowInstance?.fitView({ padding: 0.18 });
  return;
case 'focus_selection':
  flowInstance?.fitView({ nodes: selectedNodes, padding: 0.3 });
  return;
```

Reuse existing copy/paste/delete/group/ungroup/history helpers. Generate duplicate nodes using the existing copy/paste helper with an offset derived from the selected node bounds. Set React Flow `panOnDrag` to include mouse button `1` and preserve the current left-button pan mode by selecting `[0, 1]` only in pan mode and `[1]` otherwise.

- [ ] **Step 4: Re-run the focused test**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-commands.test.js`

Expected: all assertions pass.

### Task 3: Make every canvas node a compact preview

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Consumes: node `data`, existing manual editor modal, and media asset references.
- Produces: compact cards that open their existing editor by double click and no longer contain permanent media parameter fields.

- [ ] **Step 1: Write failing node-layout tests**

```js
assert.doesNotMatch(canvasSource, /agent-canvas-node-inline-settings/);
assert.doesNotMatch(canvasSource, /agent-canvas-image-editor/);
assert.match(canvasSource, /onDoubleClick=\{handleNodeDoubleClick\}/);
assert.match(canvasStyles, /\.agent-page \.agent-canvas-node--video \{ width: 220px;/);
```

- [ ] **Step 2: Run the focused test and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: failure because video and image editors are currently embedded in cards.

- [ ] **Step 3: Remove embedded editors and retain the inspector path**

Remove video inline controls and the image expanded editor from `CanvasNode`. Keep preview, title, status, version, left/right handles, and quick-add. Route double-click for every editable node through `onManualEdit` so the current modal remains the single editor surface.

- [ ] **Step 4: Tighten card styles**

Use one stable media-card width, fixed preview ratio, quiet status footer, and a compact quick-add control. Delete only styles that belong exclusively to the removed embedded editors; retain the existing V88 colors and edge motion settings.

- [ ] **Step 5: Re-run the focused test**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: compact-node assertions pass and existing context menu coverage remains green.

### Task 4: Add visible shortcut help

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Consumes: modal state local to `AgentCanvas` and the same command vocabulary from Task 1.
- Produces: a user-visible `快捷键` control and modal listing only shipped commands.

- [ ] **Step 1: Write failing shortcut-dialog tests**

```js
assert.match(canvasSource, /aria-label="快捷键"/);
assert.match(canvasSource, /title="快捷键操作"/);
assert.match(canvasSource, /中键拖动/);
assert.match(canvasSource, /Ctrl\/Cmd \+ A/);
assert.match(canvasSource, /Shift \+ 1/);
```

- [ ] **Step 2: Run the focused test and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: failure because the canvas has no shortcut button or modal.

- [ ] **Step 3: Implement modal and control**

Add an icon button next to the zoom controls and an Ant Design `Modal`. Its rows explain panning, zooming, selection, editing, and the commands implemented by Tasks 1 and 2. Do not list image/audio generation or group actions unless their handlers are enabled.

- [ ] **Step 4: Re-run the focused test**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-commands.test.js`

Expected: all shortcut and command tests pass.

### Task 5: Build and verify the local Go workspace

**Files:**
- No source changes expected.

- [ ] **Step 1: Run all focused canvas tests**

Run: `node --test frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-history.test.js frontend/src/user/pages/agent-canvas-video-jobs.test.js`

Expected: pass.

- [ ] **Step 2: Build the React frontend**

Run: `npm --prefix frontend run build`

Expected: successful Vite build.

- [ ] **Step 3: Rebuild and restart the local Go-embedded runtime**

Run: `scripts/build-go-local.ps1`, then `scripts/start-local-go-review.ps1 -UseYcmMysql`, then `scripts/check-local-go.ps1`.

Expected: local 4001 serves the rebuilt embedded frontend.

- [ ] **Step 4: Browser acceptance**

Open the authenticated local `/agent/canvas` page and verify middle-button panning, shortcut modal, compact nodes, double-click editor, copy/paste, delete, undo/redo, fit, and focus selection.

- [ ] **Step 5: Review only changed files**

Run: `git diff --check -- frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-commands.js frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/shared/styles/agent-workspace.css docs/superpowers`

Expected: no whitespace errors. Do not commit or push.
