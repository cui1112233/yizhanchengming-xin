# Agent Canvas Usability Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Repair the local Agent canvas connection workflow and remove generation controls that pretend to work without a real Agent task API.

**Architecture:** Keep React Flow as the interaction engine and preserve the existing Go canvas persistence contract. Add one pure decision helper for connection completion versus empty-canvas drop, then let `AgentCanvas.jsx` manage pointer feedback and render truthful capability states. Only the existing Agent video job API remains executable in this phase.

**Tech Stack:** React 18, `@xyflow/react` 12, Ant Design, Node test runner, Go embedded frontend.

## Global Constraints

- Work only in local V88. Do not publish publicly.
- Do not create a branch, commit, push, reset, clean, or overwrite unrelated dirty-worktree changes.
- Preserve the current Go canvas revision and MySQL persistence contracts.
- The right-edge `+` is the connection source; the left-edge port is the connection target; the reference-panel `+` opens project materials.
- Empty-canvas connection drops offer only `文本` and `图片`.
- Image and audio submit controls must remain visibly unavailable until Agent-specific task APIs exist.
- Rebuild React, embed it in the Go binary, restart local port 4001, and verify through the real browser.

---

### Task 1: Make connection completion deterministic

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.js`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Produce `shouldOpenConnectionDrop({ sourceID, connectionCompleted, droppedOnHandle })` returning a boolean.
- A completed connection or a drop on a real handle never opens the empty-canvas creation menu.

- [ ] Write failing unit tests for missing source, completed connection, handle drop, and true empty-canvas drop.
- [ ] Run the focused test and confirm the new tests fail because the helper does not exist.
- [ ] Implement the pure helper without React or DOM dependencies.
- [ ] Re-run the focused test and confirm all decision cases pass.

### Task 2: Attach ports and expose valid connection feedback

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- `connectionCompletedRef` records whether React Flow accepted a connection during the active drag.
- `connectionSourceID` drives the `is-connecting` canvas state.
- `isValidConnection` rejects self-links and duplicate links before drop.

- [ ] Add failing source/style assertions for the named target handle, overlapping source port, connection radius, valid-connection callback, and completion guard.
- [ ] Run the focused test and confirm those assertions fail.
- [ ] Add the named target handle and keep the source `+` centered directly over the node border with no horizontal translate.
- [ ] Track connection start, accepted connection, and end so a successful connection cannot also open the empty-drop menu.
- [ ] Add a larger connection radius and visible target feedback while dragging.
- [ ] Re-run the focused canvas tests.

### Task 3: Make generation controls truthful

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Video nodes keep the existing `onSubmitVideo` path.
- Image and audio nodes show `暂未接入生成执行器` and keep their submit buttons disabled.

- [ ] Add failing assertions that video alone calls `onSubmitVideo` and image/audio expose a disabled capability message.
- [ ] Run the focused test and confirm failure.
- [ ] Render a capability message and disable non-video submission without removing prompt/model editing.
- [ ] Re-run the focused test.

### Task 4: Rebuild and perform browser acceptance

**Files:**
- No further source changes expected.

- [ ] Run all Agent canvas unit and source tests.
- [ ] Build the React frontend.
- [ ] Stop only the verified `qiantie-local.exe` listener on port 4001, rebuild the Go embedded binary, restart with the review MySQL config, and run local health checks.
- [ ] In the authenticated browser, verify the source `+` is attached, drag-to-empty opens exactly `文本` and `图片`, a valid existing-node connection does not also open the menu, and image/audio submit controls cannot pretend to run.
- [ ] Run `git diff --check` on the changed plan, source, style, and test files. Do not commit or push.
