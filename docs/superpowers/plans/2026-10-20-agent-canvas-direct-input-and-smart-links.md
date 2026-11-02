# Agent Canvas Direct Input and Smart Links Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make media generation nodes editable directly on the canvas and make connections create usable references instead of decorative lines.

**Architecture:** Keep React Flow responsible for pointer and edge events. Put pure node/reference derivation in `agent-canvas-context-menu.js`, then let `AgentCanvas.jsx` use it to update the existing canvas revision through `commit`. The Go API, MySQL schema, and canvas serialization format remain unchanged because node `config` already persists reference assets.

**Tech Stack:** React 18, `@xyflow/react`, Ant Design, Lucide, Node test runner, Go embedded frontend.

## Global Constraints

- Local V88 only. Do not publish, commit, push, reset, or clean the dirty workspace.
- Preserve current Go canvas save and revision conflict behavior.
- Right-edge `+` means draw a connection; panel `+` means add project material.
- A connection dropped onto empty canvas offers only text and image creation.
- Build React before rebuilding the embedded local Go binary, then verify local port 4001.

---

### Task 1: Derive automatic reference assets from canvas edges

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.js`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Produces `syncNodeReferencesFromEdges({ nodes, edges })`, returning new nodes only when a connected media asset changes an eligible target reference list.
- Video targets accept image and video assets; image targets accept image assets only; asset IDs deduplicate and preserve edge order.

- [ ] **Step 1: Write failing tests** for an image edge feeding a video target, a video edge feeding a video target, rejection of a video reference by an image target, and duplicate edges.
- [ ] **Step 2: Run** `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js` and confirm the missing helper causes the new tests to fail.
- [ ] **Step 3: Implement** the pure edge-to-reference derivation without mutating source nodes, edge input, or existing manually selected references.
- [ ] **Step 4: Re-run** the focused test and confirm all reference cases pass.

### Task 2: Turn a dropped connection into a linked text or image node

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Consumes React Flow `onConnectStart`, `onConnectEnd`, and `onConnect` events.
- Produces a `添加参考节点` menu at an empty-drop canvas coordinate with buttons `文本` and `图片`.

- [ ] **Step 1: Write failing source tests** for the connection-drop state, the two menu buttons, and creation through `createNodeAt` plus `connectCanvasNodes`.
- [ ] **Step 2: Run** the focused canvas test and confirm the new assertions fail.
- [ ] **Step 3: Implement** empty-drop detection, screen-to-flow coordinate conversion, menu placement, and the two linked creation actions; retain normal existing-node connections.
- [ ] **Step 4: Add scoped CSS** for the compact drop menu without changing the V88 color system.
- [ ] **Step 5: Re-run** the focused test and confirm it passes.

### Task 3: Render direct node input panels

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Media node data uses existing `data.config` fields: `prompt`, `model`, `ratio`, `resolution`, `durationSeconds`, `voice`, `speed`, and `referenceAssets`.
- The direct panel `+` opens the existing project asset picker in a reference mode anchored below the node.

- [ ] **Step 1: Write failing source tests** for prompt inputs, direct model controls, the panel reference `+`, and distinct right-side link handle semantics.
- [ ] **Step 2: Run** the focused test and confirm it fails before the UI implementation.
- [ ] **Step 3: Implement** an inline panel for image, video, and audio nodes. Debounce only through the existing save path; each ordinary input change updates canvas node config via `commit`.
- [ ] **Step 4: Render** selected references as removable thumbnails plus a count. Keep text nodes compact.
- [ ] **Step 5: Re-run** focused tests.

### Task 4: Connect automatic references to rendering and persistence

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- `handleConnect(connection)` first creates the edge, then calls `syncNodeReferencesFromEdges` before `commit`.
- The same synchronization runs when a node is created from the connection-drop menu.

- [ ] **Step 1: Write a failing source-level integration test** that verifies the connection handler invokes edge/reference synchronization before saving.
- [ ] **Step 2: Run** the focused test and confirm failure.
- [ ] **Step 3: Implement** the integration with the existing history and save path; preserve manually added references when synchronizing.
- [ ] **Step 4: Re-run** the full focused canvas suite.

### Task 5: Build and browser acceptance

**Files:**
- No further source changes expected.

- [ ] **Step 1: Run** `node --test frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-history.test.js frontend/src/user/pages/agent-canvas-video-jobs.test.js`.
- [ ] **Step 2: Run** `npm --prefix frontend run build`, then `scripts/build-go-local.ps1`, restart the verified local `qiantie-local.exe`, and run `scripts/check-local-go.ps1`.
- [ ] **Step 3: Verify in browser:** direct prompt/settings panels, material `+`, empty-drop text/image menu, existing-media connection auto reference, reload persistence, and the right-side drag handle.
- [ ] **Step 4: Run** `git diff --check` for the five changed source/test/document files. Do not commit or push.
