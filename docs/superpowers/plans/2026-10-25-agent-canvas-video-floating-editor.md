# Agent Canvas Video Floating Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make video selection reveal fixed-screen-size tools above and an editor below without resizing the video node.

**Architecture:** Keep both controls anchored inside the selected React Flow node, but remove them from normal layout with absolute positioning. Apply the reciprocal of the current canvas zoom to the floating surfaces so React Flow movement is inherited while viewport scaling is cancelled.

**Tech Stack:** React, `@xyflow/react`, CSS custom properties, Node test runner, Vite, Go embed runtime.

## Global Constraints

- Work directly on the maintained `v88` branch and preserve unrelated dirty files.
- Modify React source first; compiled assets are build output only.
- Use the Go embedded runtime path and do not start a Node application server.
- Do not add APIs, database changes, fake video operations, or trigger generation.

---

### Task 1: Lock The Floating Layout Contract

**Files:**
- Modify: `frontend/src/user/components/AgentCanvasVideoEditor.source.test.js`

**Interfaces:**
- Consumes: `canvasZoom` exposed as `--agent-canvas-zoom` by `AgentNodeComposer`.
- Produces: source assertions for fixed node width, absolute overlays, and reciprocal zoom scaling.

- [ ] **Step 1: Replace the old adaptive-width assertion with assertions that reject an open-state node width and require floating classes.**
- [ ] **Step 2: Run `node --test src/user/components/AgentCanvasVideoEditor.source.test.js` from `frontend` and confirm the new assertions fail against the current implementation.**

### Task 2: Implement Fixed Node And Inverse-Scaled Floating Surfaces

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `canvasZoom: number`, `isEditorOpen: boolean`, and the existing `AgentVideoQuickTools` and `AgentNodeComposer` callbacks.
- Produces: `.agent-video-floating-toolbar` and `.agent-video-floating-composer` surfaces anchored to the selected video node.

- [ ] **Step 1: Wrap the video toolbar and composer with dedicated floating anchors without changing their real actions.**
- [ ] **Step 2: Remove the video open-state width override.**
- [ ] **Step 3: Apply absolute positioning and `scale(calc(1 / max(var(--agent-canvas-zoom), 0.25)))` with top/bottom transform origins.**
- [ ] **Step 4: Run the focused source test and confirm all assertions pass.**

### Task 3: Build, Commit, And Verify The Go Runtime

**Files:**
- Generated: `frontend/dist/**`
- Embedded binary: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: the finished React source and existing local MySQL review environment.
- Produces: a Git commit on `v88` and a Go runtime whose build SHA matches that commit.

- [ ] **Step 1: Run the focused Agent Canvas tests and the frontend production build.**
- [ ] **Step 2: Stage only the approved source, tests, design, plan, and required generated assets; preserve unrelated dirty changes.**
- [ ] **Step 3: Commit to `v88`, rebuild the embedded Go binary, and restart it with `scripts/start-local-go-review.ps1`.**
- [ ] **Step 4: Verify `/health`, `/api/runtime-build-info`, and the live browser at multiple zoom levels without clicking Generate.**

