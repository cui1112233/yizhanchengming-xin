# Agent Canvas Asset Mentions And Image Paste Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real project-asset mentions to node prompts and persist pasted screenshots as image nodes.

**Architecture:** A pure helper module owns mention parsing, asset filtering, insertion, and clipboard-image selection. `AgentNodeComposer` owns the local popup state; `AgentCanvas` owns authenticated asset listing/upload and durable node commits.

**Tech Stack:** React, Ant Design, React Flow, Node test runner, existing Go asset API and Go embed release path.

## Global Constraints

- Work directly on canonical branch `v88`; do not create another branch.
- Preserve unrelated dirty files and stage only files belonging to this feature.
- Do not add APIs, database changes, or trigger real generation.
- Build React first, then the Go embedded binary, and run only the Go application runtime.

---

### Task 1: Asset input rules

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-asset-input.js`
- Create: `frontend/src/user/pages/agent-canvas-asset-input.test.js`

**Interfaces:**
- Produces: `assetMentionTrigger`, `filterMentionAssets`, `insertAssetMention`, `referenceKindsForPromptNode`, and `firstClipboardImage`.

- [x] Write tests for valid `@` ranges, filtering, insertion, node type rules, and image clipboard selection.
- [x] Run the test and confirm it fails because the module is absent.
- [x] Implement the smallest pure helpers that satisfy the tested behavior.
- [x] Run the test and confirm it passes.

### Task 2: Prompt mention picker

**Files:**
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-asset-input.test.js`

**Interfaces:**
- Consumes: Task 1 helpers and existing `listAgentAssets`, `appendReferenceAsset`, `patchNodeConfig`, and canvas commit functions.
- Produces: `onListMentionAssets` and `onMentionAsset` callbacks from the canvas to each composer.

- [x] Extend failing tests to require text-node references and duplicate-safe asset binding.
- [x] Run focused tests and confirm the new expectations fail.
- [x] Add the popup state and callbacks, inserting readable prompt text while persisting actual asset references.
- [x] Add theme-safe compact popup styling.
- [x] Run focused tests and confirm they pass.

### Task 3: External screenshot paste

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-asset-input.test.js`

**Interfaces:**
- Consumes: `firstClipboardImage`, existing `uploadAgentAsset`, `addAssetNode`, and React Flow coordinate conversion.
- Produces: a canvas-level paste handler that uploads and commits one image node.

- [x] Add failing tests for clipboard image selection and editable-target exclusion.
- [x] Run focused tests and confirm the new expectations fail.
- [x] Add the canvas paste handler, upload status, coordinate placement, and error feedback.
- [x] Run focused tests and confirm they pass.

### Task 4: Build and local acceptance

**Files:**
- Generated only: frontend distribution and Go embedded binary.

- [x] Run all focused Agent Canvas tests.
- [x] Build the frontend.
- [ ] Commit only feature source, tests, and docs to `v88`.
- [ ] Build the Go embedded binary and restart the verified local Go runtime.
- [ ] Check `/health` and `/api/runtime-build-info` against the new commit.
- [ ] Verify `@` selection and screenshot paste in the current canvas without triggering generation.
