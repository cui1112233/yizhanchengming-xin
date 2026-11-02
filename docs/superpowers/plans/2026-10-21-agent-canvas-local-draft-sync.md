# Agent Canvas Local Draft Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the local Agent canvas recover unsynced work instantly from browser storage while batching quiet background syncs to the existing Go canvas API.

**Architecture:** Keep the React canvas as the immediate editing surface. Add a pure browser-storage module keyed by account and Agent task ID, then let `AgentPageV2` persist every durable canvas change locally before passing the newest snapshot to a trailing 1.2-second remote save queue. The server revision remains authoritative for cross-device work and conflict recovery; browser drafts only buffer local work until that remote revision accepts it.

**Tech Stack:** React, React Flow, browser `localStorage`, existing Go embedded frontend and `/api/agent/tasks/{id}/canvas` API, Node test runner.

## Global Constraints

- Scope is local V88 only; do not publish to the public site.
- Preserve the existing Go API, MySQL revision contract, Redis task behavior, and persisted project records.
- Browser draft keys must be account- and Agent-project-scoped: `qiantie:agent-canvas-draft:<account>:<taskId>`.
- Do not persist selection, open panels, menus, pointer state, or other transient React Flow UI state.
- Durable canvas edits render immediately, update local draft immediately, and batch remote sync after 1200ms of inactivity.
- Route changes and `pagehide` must preserve the local draft and attempt a final non-blocking remote flush.
- A remote save failure must keep the canvas visible and retain the draft; it must not silently drop user content.
- Do not commit, push, create a branch, or release as part of this task.

---

### Task 1: Isolated Canvas Draft Storage

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-draft-storage.js`
- Create: `frontend/src/user/pages/agent-canvas-draft-storage.test.js`

**Interfaces:**
- Produces: `canvasDraftKey(username, taskId): string`
- Produces: `loadCanvasDraft(storage, username, taskId): { canvas: Canvas, baseRevision: number, savedAt: string } | null`
- Produces: `saveCanvasDraft(storage, username, taskId, canvas, baseRevision): boolean`
- Produces: `clearCanvasDraft(storage, username, taskId): boolean`
- Produces: `isDraftAtLeastAsFresh(draft, serverCanvas): boolean`

- [ ] **Step 1: Write the failing storage tests**

```js
test('canvas drafts are isolated by account and project', () => {
  const storage = new MapStorage();
  saveCanvasDraft(storage, 'alice', 'task-a', canvas('local-a'), 3);
  saveCanvasDraft(storage, 'alice', 'task-b', canvas('local-b'), 4);
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-a').canvas.nodes[0].id, 'local-a');
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-b').canvas.nodes[0].id, 'local-b');
  assert.equal(loadCanvasDraft(storage, 'bob', 'task-a'), null);
});

test('invalid draft data is ignored without throwing', () => {
  const storage = new MapStorage();
  storage.setItem(canvasDraftKey('alice', 'task-a'), '{broken');
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-a'), null);
});

test('only drafts based on the current or newer revision can restore', () => {
  const draft = { canvas: canvas('local'), baseRevision: 4, savedAt: '2026-10-21T00:00:00.000Z' };
  assert.equal(isDraftAtLeastAsFresh(draft, { revision: 4 }), true);
  assert.equal(isDraftAtLeastAsFresh(draft, { revision: 5 }), false);
});
```

- [ ] **Step 2: Run the storage test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-canvas-draft-storage.test.js`  
Expected: FAIL because the draft-storage module does not exist.

- [ ] **Step 3: Implement the bounded, versioned storage contract**

```js
const storagePrefix = 'qiantie:agent-canvas-draft:';
const draftVersion = 1;

export function canvasDraftKey(username, taskId) {
  return `${storagePrefix}${encodeURIComponent(String(username || 'guest'))}:${encodeURIComponent(String(taskId || ''))}`;
}

export function saveCanvasDraft(storage, username, taskId, canvas, baseRevision) {
  if (!taskId || !canvas || typeof canvas !== 'object') return false;
  try {
    storage?.setItem(canvasDraftKey(username, taskId), JSON.stringify({
      version: draftVersion,
      canvas: stripTransientCanvasState(canvas),
      baseRevision: Number(baseRevision) || 0,
      savedAt: new Date().toISOString()
    }));
    return true;
  } catch {
    return false;
  }
}
```

`stripTransientCanvasState` must copy only supported canvas fields (`version`, `nodes`, `edges`, `viewport`, `appearance`, `viewState`), remove node `selected`, and reject malformed arrays instead of allowing broken JSON to reach React Flow.

- [ ] **Step 4: Run the storage test to verify it passes**

Run: `node --test frontend/src/user/pages/agent-canvas-draft-storage.test.js`  
Expected: PASS.

- [ ] **Step 5: Commit**

Do not commit in this working tree; leave the tested changes uncommitted for the user.

### Task 2: Draft-Aware Trailing Sync Queue

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-state.js`
- Modify: `frontend/src/user/pages/agent-canvas-state.test.js`

**Interfaces:**
- Consumes: a durable canvas snapshot supplied by `AgentPageV2` after it has been locally stored.
- Produces: `createCanvasSaveQueue({ revision, delay: 1200, save, reload, onCanvas, onStatus, onSynced }): CanvasSaveQueue`
- Produces: `queue.enqueue(canvas)`, `queue.flush()`, `queue.retry()`, `queue.acceptServerCanvas(canvas)`, `queue.dispose()`.

- [ ] **Step 1: Write failing batching and completion tests**

```js
test('ten edits during the quiet period make one remote save with the newest canvas', async () => {
  const calls = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 20,
    save: async (next, revision) => { calls.push({ next, revision }); return { canvas: { ...next, revision: 3 } }; },
    reload: async () => ({ canvas: emptyCanvas(2) })
  });
  for (let number = 0; number < 10; number += 1) queue.enqueue(canvas(`node-${number}`));
  await wait(40);
  assert.equal(calls.length, 1);
  assert.equal(calls[0].next.nodes[0].id, 'node-9');
});

test('a successful remote save reports the revision that may clear its matching local draft', async () => {
  const synced = [];
  const queue = createCanvasSaveQueue({
    revision: 2, delay: 0,
    save: async next => ({ canvas: { ...next, revision: 3 } }),
    reload: async () => ({ canvas: emptyCanvas(2) }),
    onSynced: canvas => synced.push(canvas)
  });
  queue.enqueue(canvas('latest'));
  await queue.flush();
  assert.equal(synced.at(-1).revision, 3);
});
```

- [ ] **Step 2: Run queue tests to verify they fail**

Run: `node --test frontend/src/user/pages/agent-canvas-state.test.js`  
Expected: FAIL because the new `onSynced` behavior is absent or the trailing behavior differs.

- [ ] **Step 3: Extend the queue without changing its server revision protections**

```js
export function createCanvasSaveQueue({ revision = 0, delay = 1200, save, reload, onCanvas, onStatus, onSynced }) {
  // retain newest pending canvas; each enqueue clears and restarts the timer
  // after a successful save call onSynced(savedCanvas) only if no newer pending canvas exists
}
```

Retain the existing conflict reload/merge path. Do not clear a draft after an older save response if a newer canvas is pending. Keep `flush()` for page exit and retry, but do not expose it as a normal visible “保存” control.

- [ ] **Step 4: Run queue regression tests**

Run: `node --test frontend/src/user/pages/agent-canvas-state.test.js`  
Expected: PASS, including existing conflict, task-switch, and server-confirmation tests.

- [ ] **Step 5: Commit**

Do not commit in this working tree; leave the tested changes uncommitted for the user.

### Task 3: Connect the Canvas Page to Local Drafts

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-state.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-draft-storage.test.js`

**Interfaces:**
- Consumes: `loadCanvasDraft`, `saveCanvasDraft`, `clearCanvasDraft`, `isDraftAtLeastAsFresh` from Task 1.
- Consumes: `onSynced(canvas)` from Task 2.
- Produces: one `applyCanvasEdit(nextCanvas)` callback that updates React state, writes the scoped local draft, and enqueues remote sync.

- [ ] **Step 1: Add failing page-source contract checks**

```js
test('AgentPageV2 writes every durable canvas edit to scoped local draft before enqueueing', () => {
  const source = readFileSync(new URL('./AgentPageV2.jsx', import.meta.url), 'utf8');
  assert.match(source, /function applyCanvasEdit\(nextCanvas\)/);
  assert.match(source, /saveCanvasDraft\(window\.localStorage/);
  assert.match(source, /canvasSaveQueueRef\.current\?\.enqueue\(nextCanvas\)/);
});

test('page exit retains local draft and flushes pending remote work', () => {
  const source = readFileSync(new URL('./AgentPageV2.jsx', import.meta.url), 'utf8');
  assert.match(source, /addEventListener\('pagehide'/);
  assert.match(source, /saveCanvasDraft\(window\.localStorage/);
  assert.match(source, /queue\?\.flush\(\)/);
});
```

- [ ] **Step 2: Run page contract tests to verify they fail**

Run: `node --test frontend/src/user/pages/agent-canvas-state.test.js frontend/src/user/pages/agent-canvas-draft-storage.test.js`  
Expected: FAIL because `applyCanvasEdit` and page-exit draft behavior are absent.

- [ ] **Step 3: Load, apply, preserve, and clear drafts at defined points**

```jsx
const applyCanvasEdit = useCallback((nextCanvas) => {
  if (!activeTaskIdRef.current) return;
  setCanvas(nextCanvas);
  saveCanvasDraft(window.localStorage, getCurrentUsername(), activeTaskIdRef.current, nextCanvas, canvasRevisionRef.current);
  canvasSaveQueueRef.current?.enqueue(nextCanvas);
}, []);
```

When a project loads, create the queue from the server revision, then load the scoped draft. Restore a draft only when its `baseRevision >= loadedCanvas.revision`; enqueue that restored draft using the 1200ms delay. Clear the scoped draft from `onSynced` only when the saved snapshot is still the latest local draft. Register `pagehide`, `online`, and window `focus` handlers that preserve the latest local snapshot and safely call queue `flush` or `retry` as applicable. Replace direct `canvasSaveQueueRef.current?.enqueue(...)` calls in canvas, storyboard, and server-result update paths with this single callback where they represent a durable user canvas edit.

Keep existing server-confirmation paths using `acceptServerCanvas`; after a confirmed server result is accepted, write the merged current canvas back to the local draft only when unsynced local changes remain.

- [ ] **Step 4: Run targeted page and state tests**

Run: `node --test frontend/src/user/pages/agent-canvas-state.test.js frontend/src/user/pages/agent-canvas-draft-storage.test.js`  
Expected: PASS.

- [ ] **Step 5: Commit**

Do not commit in this working tree; leave the tested changes uncommitted for the user.

### Task 4: Remove Normal Save UI and Verify Actual Browser Behavior

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css` only if a removed save control leaves layout space.
- Test: existing focused Node tests plus an authenticated local browser session at `http://localhost:4001/agent/canvas?taskId=agent-381900adb37b8663b49729b2&view=workflow`.

**Interfaces:**
- Consumes: the durable edit callback and background statuses from Task 3.
- Produces: no normal save button; only a compact offline/recovery notice when a remote sync cannot proceed.

- [ ] **Step 1: Add failing source checks for invisible normal saving**

```js
test('normal canvas status does not render a manual save command', () => {
  const source = readFileSync(new URL('./AgentPageV2.jsx', import.meta.url), 'utf8');
  assert.doesNotMatch(source, /onClick=\{saveCanvasNow\}[^>]*>保存/);
  assert.match(source, /canvasSaveStatus === 'error'/);
});
```

- [ ] **Step 2: Run the focused test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-canvas-state.test.js`  
Expected: FAIL while normal save controls still render.

- [ ] **Step 3: Limit visible feedback to recovery states**

Remove the normal “已保存 / 正在保存 / 保存” action from the workspace header. Keep only one compact offline/error notice with an explicit retry action after a real failure. Do not surface every queue transition as a toast or status change.

- [ ] **Step 4: Build and restart the embedded local runtime**

Run: `npm --prefix frontend run build`  
Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`  
Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`  
Expected: all commands complete, local service responds on port 4001.

- [ ] **Step 5: Perform browser acceptance in the logged-in local page**

1. Move two nodes, alter a prompt, and create one connection within one second; confirm the changes remain visible without a save control.
2. Immediately switch to storyboard and back, then refresh; confirm the newest node positions, prompt, main image, and reference image remain.
3. Select an existing project image as a node main image, add a second image as a reference, refresh, and confirm both persist.
4. Temporarily simulate an unavailable save request; edit a node, refresh, and confirm the local draft restores. Restore connectivity and confirm it syncs without a manual save command.
5. Run one configured generation flow through result confirmation, refresh, and verify the returned asset remains attached to its node.

- [ ] **Step 6: Commit**

Do not commit in this working tree; leave the tested changes uncommitted for the user.
