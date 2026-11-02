# Agent Canvas Generation Review Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give every Agent canvas generation node one clear lifecycle, and let users review and choose among multiple generated images before any result replaces the node preview.

**Architecture:** Keep text generation synchronous and keep the existing durable video-job lifecycle. Add an `awaiting_confirmation` state to generic node executions so image results can be persisted as project assets without immediately changing the canvas. The frontend derives one normalized presentation state for text, image, and video nodes, restores pending image choices from the execution list, and applies only the asset explicitly selected by the user.

**Tech Stack:** Go, MySQL, React 18, Ant Design, `@xyflow/react`, Node test runner, Go embedded frontend.

## Global Constraints

- Work directly on the canonical `v88` branch; do not create a feature branch.
- Preserve all unrelated dirty worktree changes.
- Keep the application runtime Go-only and publish only through the Go embed path.
- Persist generated media as project assets before showing it as a selectable result.
- Removing or rejecting a result never deletes its project asset.
- Do not push or deploy publicly in this plan.

---

### Task 1: Persist image results as a reviewable execution

**Files:**
- Modify: `backend/internal/agentworkspace/store.go`
- Modify: `backend/internal/agentworkspace/memory_store.go`
- Modify: `backend/internal/agentworkspace/mysql_store.go`
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Test: `backend/internal/agentworkspace/store_test.go`
- Test: `backend/internal/agentworkspace/mysql_store_test.go`
- Test: `backend/internal/httpapi/agent_node_generation_test.go`

**Interfaces:**
- Produces: `SetNodeExecutionAwaitingConfirmation(ctx, owner, projectID, executionID, detail, now) (Execution, error)`.
- Produces: image generation response `{ execution, resultAssets }` without changing the canvas.

- [ ] **Step 1: Write failing store and route tests**

Assert that an image generation with two outputs stores both assets, changes the execution to `awaiting_confirmation`, records `nodeId`, `nodeVersion`, and `resultAssetIds`, and leaves the node version and asset reference unchanged.

- [ ] **Step 2: Run the focused tests and verify the expected failure**

Run: `go test ./internal/agentworkspace ./internal/httpapi -run 'NodeExecutionAwaitingConfirmation|ImageGenerationPersistsEveryRequestedResult'`

Expected: failure because the store transition does not exist and image generation currently applies the first result immediately.

- [ ] **Step 3: Implement the durable transition**

Allow only `running -> awaiting_confirmation`. Store the supplied review detail and timestamp. Change image generation to persist every returned asset, then transition the execution without updating the canvas.

- [ ] **Step 4: Re-run the focused tests**

Run: `go test ./internal/agentworkspace ./internal/httpapi -run 'NodeExecutionAwaitingConfirmation|ImageGenerationPersistsEveryRequestedResult'`

Expected: pass.

### Task 2: Confirm one selected image safely

**Files:**
- Modify: `backend/internal/agentworkspace/memory_store.go`
- Modify: `backend/internal/agentworkspace/mysql_store.go`
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Test: `backend/internal/agentworkspace/store_test.go`
- Test: `backend/internal/agentworkspace/mysql_store_test.go`
- Test: `backend/internal/httpapi/agent_node_generation_test.go`

**Interfaces:**
- Produces: `POST /api/agent/tasks/{id}/image-generations/{executionID}/confirm` with `{ assetId }`.
- Returns: `{ canvas, execution, resultAsset }`.

- [ ] **Step 1: Write failing confirmation tests**

Cover valid selection, an asset outside the execution candidates, cross-project access, repeated confirmation, and a node version changed after generation.

- [ ] **Step 2: Run the focused tests and verify the expected failure**

Run: `go test ./internal/agentworkspace ./internal/httpapi -run 'ConfirmImageGeneration|AwaitingConfirmation'`

Expected: failure because there is no confirmation route and `ApplyNodeExecution` accepts only `running`.

- [ ] **Step 3: Implement confirmation**

Let `ApplyNodeExecution` accept `awaiting_confirmation` for image executions. The route must locate the execution, verify owner/project, verify the selected asset is listed in `resultAssetIds`, resolve the asset from the project asset store, and then apply it using the original node version snapshot.

- [ ] **Step 4: Re-run the focused tests**

Run: `go test ./internal/agentworkspace ./internal/httpapi -run 'ConfirmImageGeneration|AwaitingConfirmation'`

Expected: pass.

### Task 3: Normalize generation state in the canvas UI

**Files:**
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/pages/agent-canvas-node-generation.js`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Test: `frontend/src/user/pages/agent-canvas-node-generation.test.js`
- Test: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

**Interfaces:**
- Produces: `nodeExecutionReview(executions, nodeID)` returning the newest pending image execution and its candidates.
- Produces: `confirmAgentImageGeneration(taskId, executionId, assetId)`.
- Produces: one UI state vocabulary: `准备生成`, `正在生成`, `等待选择`, `生成失败`, `已完成`.

- [ ] **Step 1: Write failing frontend tests**

Assert that the newest pending execution wins, completed/failed entries do not mask it, and the image composer renders candidate selection and explicit confirmation controls.

- [ ] **Step 2: Run tests and verify the expected failure**

Run: `node --test frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

Expected: failure because image generation immediately expects a canvas response and the composer has no candidate review surface.

- [ ] **Step 3: Implement the normalized state and review UI**

Load executions with the canvas, retain the selected candidate per node, render all candidate thumbnails in a compact horizontal strip, and provide `使用这张` plus `重新生成`. A newly generated response updates the execution state locally; reload restores it from `/executions`.

- [ ] **Step 4: Re-run frontend tests**

Run: `node --test frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-canvas-video-jobs.test.js`

Expected: pass.

### Task 4: Build and browser-verify the Go workspace

**Files:**
- No source changes expected.

**Interfaces:**
- Verifies: durable reload behavior, exact result selection, and unchanged video controls.

- [ ] **Step 1: Run the complete focused canvas and Go test sets**

Run the Agent canvas Node tests plus `go test ./internal/agentworkspace ./internal/httpapi` from `backend`.

- [ ] **Step 2: Build the frontend and Go binary**

Use an isolated frontend output, then run the approved local Go build and review scripts.

- [ ] **Step 3: Verify in the authenticated local browser**

Generate multiple images, confirm that none replaces the preview before selection, select a non-first result, confirm it, reload the page, and verify the chosen asset remains while other generated assets stay in the project asset collection.

- [ ] **Step 4: Review and commit only scoped files**

Check the exact diff and commit the implementation to `v88`. Do not push or deploy publicly.
