# Agent Workflow Liblib Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a Liblib-aligned V88 workflow canvas with consistent nodes, subtle handles and edges, and automatic recovery from an off-screen saved viewport.

**Architecture:** Keep the existing React Flow graph and persistence contracts. Add one pure viewport-decision helper, update canvas presentation defaults, and apply scoped JSX/CSS changes to the existing node and edge renderers.

**Tech Stack:** React, `@xyflow/react`, Lucide React, CSS, Node test runner, Vite, Go embed.

## Global Constraints

- Work on the existing `v88` checkout and preserve unrelated dirty files.
- Do not change generation business logic, node schema, API contracts, or Storyboard state.
- Do not start paid generation.
- Do not publish or push remotely.
- Build the frontend and `backend/qiantie-local.exe`, then verify the Go runtime on `http://localhost:4001`.

---

### Task 1: Recover Off-Screen Workflow Viewports

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-viewport.js`
- Create: `frontend/src/user/pages/agent-canvas-viewport.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`

**Interfaces:**
- Produces: `shouldRecoverCanvasViewport(nodes, viewport, canvasSize, nodeSizes)` returning a boolean.
- Consumes: saved React Flow node positions, viewport transform, and measured canvas/node dimensions.

- [ ] Write literal rectangle tests for visible, partially visible, completely off-screen, empty, and invalid viewport cases.
- [ ] Run the new test and confirm it fails because the helper is missing.
- [ ] Implement the pure helper and add a one-time post-initialization `fitView` call only when it returns true.
- [ ] Confirm recovery never commits node positions or writes viewport state directly.
- [ ] Run viewport and canvas-state tests.

### Task 2: Establish Neutral Edge Presentation

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-appearance.js`
- Modify: `frontend/src/user/pages/agent-canvas-appearance.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Produces: neutral default edge style, arrowless static `CanvasEdge`, selected/hover CSS, optional activity overlay.
- Consumes: existing normalized appearance and edge visibility settings.

- [ ] Change tests first to require a neutral thin non-animated default while preserving explicit legacy presets.
- [ ] Run tests and confirm the old coral animated default fails.
- [ ] Add the neutral option and remove `MarkerType.ArrowClosed` from normal edges.
- [ ] Add selected, hover, connection-preview, and reduced-motion styles.
- [ ] Run appearance, connection, and canvas source tests.

### Task 3: Unify Node Shells, Icons, Handles, And Media Ratios

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Interfaces:**
- Produces: common node header structure, stable type-specific dimensions, 8px visible handle dots with 24px hit targets.
- Consumes: existing `canvasNodeTypeLabel`, Lucide icons, node status presentation, and node editor callbacks.

- [ ] Add failing layout/source assertions for type icons, header structure, stable 16:9 media previews, and handle hit/visual sizing.
- [ ] Run the focused tests and confirm the old markup/styles fail.
- [ ] Add a node-kind icon mapper and compact shared header markup without changing editor behavior.
- [ ] Apply scoped node, media, hover, selection, and handle styles.
- [ ] Run all focused Agent Canvas presentation and interaction tests.

### Task 4: Embedded Runtime And Browser Acceptance

**Files:**
- Generated: `frontend/dist/**`
- Generated: `backend/web/dist/**`
- Generated: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: Tasks 1-3.
- Produces: authenticated local acceptance evidence from the Go-embedded frontend.

- [ ] Run focused workflow tests and the existing Storyboard layout tests.
- [ ] Run `npm.cmd --prefix frontend run build`.
- [ ] Run `.\scripts\build-go-local.ps1`.
- [ ] Verify port 4001 belongs to this workspace binary before restarting it with `.\scripts\start-local-go-review.ps1 -UseYcmMysql`.
- [ ] Browser-check node visibility, drag, connection drag, edge selection, zoom, Storyboard round trip, refresh, console errors, and runtime build info without triggering generation.
- [ ] Run `git diff --check` on targeted source files and report local-only completion.

