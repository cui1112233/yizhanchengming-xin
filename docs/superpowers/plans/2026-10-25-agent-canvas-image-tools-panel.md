# Agent Canvas Image Tools Panel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the debug-like image tool row with a compact, theme-safe image processing panel backed only by the existing Go image-tool operations.

**Architecture:** `AgentImageTools` remains the focused owner of image-tool presentation and local crop form state. `AgentNodeComposer` only controls whether the panel is open and labels the entry; the existing `onImageTool` callback remains the sole submission path. Scoped CSS under `.agent-page` supplies dark/light-safe visuals through existing Agent workspace variables.

**Tech Stack:** React, Ant Design InputNumber, Lucide icons, Node test runner, Vite, existing Go embed release scripts.

## Global Constraints

- Work directly on canonical branch `v88`; do not create another branch.
- Preserve unrelated dirty files and stage only files belonging to this change.
- Keep the existing operations `rotate`, `mirror`, `grid_split`, and `crop`; do not add APIs or database changes.
- Do not trigger a real image processing or generation task during acceptance.
- Build React first, then rebuild and restart only the verified Go runtime.

---

### Task 1: Lock The Panel Contract

**Files:**
- Modify: `frontend/src/user/pages/agent-image-tools.source.test.js`
- Modify: `frontend/src/user/pages/agent-node-composer.source.test.js`

**Interfaces:**
- Consumes: existing `AgentImageTools({ sourceAsset, busy, onApply })` and `AgentNodeComposer` source.
- Produces: source-level regression coverage for the entry label, action labels, crop disclosure, and theme-safe class names.

- [ ] **Step 1: Add failing source assertions**

Require `图片处理`, `旋转 90°`, `水平镜像`, `2 x 2 拆图`, `裁剪`, `生成裁剪结果`, `agent-image-tools__action-grid`, `agent-image-tools__crop-fields`, and a crop disclosure state.

- [ ] **Step 2: Run tests and verify the new assertions fail**

Run:

```powershell
node --test frontend/src/user/pages/agent-image-tools.source.test.js frontend/src/user/pages/agent-node-composer.source.test.js
```

Expected: new image-panel assertions fail against the current raw numeric row.

### Task 2: Implement The Compact Theme-Safe Panel

**Files:**
- Modify: `frontend/src/user/components/AgentImageTools.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-image-tools.source.test.js`
- Test: `frontend/src/user/pages/agent-node-composer.source.test.js`

**Interfaces:**
- Consumes: `sourceAsset`, `busy`, and `onApply({ operation, sourceAssetId, settings })`.
- Produces: the same callback payloads for `rotate`, `mirror`, `grid_split`, and `crop`; no backend contract changes.

- [ ] **Step 1: Replace the raw row with four named action buttons**

Keep exact payloads:

```js
apply('rotate', { degrees: 90 });
apply('mirror', { direction: 'horizontal' });
apply('grid_split', { columns: 2, rows: 2 });
apply('crop', crop);
```

- [ ] **Step 2: Make crop a disclosure instead of an immediate action**

Add local `cropOpen` state. The crop action toggles a labeled parameter area; it never calls `onApply` until “生成裁剪结果” is clicked.

- [ ] **Step 3: Add scoped theme styling**

Style the panel, action grid, selected crop action, labeled fields, Ant Design number inputs, handlers, and disabled states with existing variables such as `--agent-editor-panel`, `--agent-editor-border`, `--agent-text`, `--agent-muted`, and `--agent-editor-action-hover`. Use responsive grid tracks so narrow nodes wrap without overlap.

- [ ] **Step 4: Run focused tests**

Run:

```powershell
node --test frontend/src/user/pages/agent-image-tools.source.test.js frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-node-composer-actions.test.js
```

Expected: all focused tests pass, apart from any explicitly recorded pre-existing stale assertions in the broad composer source file.

- [ ] **Step 5: Commit only the panel implementation**

Stage the two components, scoped CSS, and the two focused tests. Do not stage unrelated Agent Canvas or Director Desk changes.

### Task 3: Build And Local Acceptance

**Files:**
- Generated: `frontend/dist/**`
- Generated: `backend/web/dist/**`
- Generated: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: the tested React source and repository Go embed scripts.
- Produces: a traceable local Go runtime whose Git and runtime manifest SHAs match the implementation commit.

- [ ] **Step 1: Build the frontend**

Run `npm run build` from `frontend` and require exit code `0`.

- [ ] **Step 2: Build and restart the Go runtime**

Run `scripts/build-go-local.ps1`, stop only the process whose executable path is exactly `backend/qiantie-local.exe`, then start `scripts/start-local-go-review.ps1 -UseYcmMysql` in the background.

- [ ] **Step 3: Verify runtime identity**

Require `/health` status `200`, `runtime=go`, and matching latest `git_sha` and `runtime_manifest_sha` from `/api/runtime-build-info`.

- [ ] **Step 4: Verify both themes in the current canvas**

Open the image node, click “图片处理”, confirm the four named actions, expand crop, and verify no white control blocks in dark and light themes. Do not click rotate, mirror, split, or “生成裁剪结果”.
