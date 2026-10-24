# Agent Workflow Liblib UI Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 Agent Canvas 工作流视图整理为接近 Liblib 的可收起分组侧栏、简洁节点卡片和底部浮动工具条，同时保持现有节点连接、缩放和生成逻辑不变。

**Architecture:** 复用现有 `AgentCanvas` 的 React Flow 节点数据与命令状态，只调整工作流视图的侧栏渲染、节点分组展示和 CSS 视觉层。分组状态仅存在于前端视图状态，不写入画布数据；节点定位继续调用现有 focus/fitView 能力。

**Tech Stack:** React, React Flow, Ant Design, existing V88 workspace CSS, Node test runner, Vite, Go embed runtime.

## Global Constraints

- 所有改动先写入 Git `v88`，不得只改运行中的文件。
- 不回退或覆盖工作区其他脏改动。
- 不触发真实模型生成，不消耗额度。
- 前端改动后运行 `npm.cmd --prefix frontend run build` 和 `scripts/build-go-local.ps1`。
- 只重启 `backend/qiantie-local.exe` 并检查 `/health` 与 `/api/runtime-build-info`。

### Task 1: Map existing workflow layout

**Files:**
- Inspect: `frontend/src/user/components/AgentCanvas.jsx`
- Inspect: `frontend/src/user/components/AgentCanvasControls.jsx`
- Inspect: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

- [ ] Confirm the existing sidebar, canvas controls, node group bounds, and React Flow node class hooks.
- [ ] Add source assertions describing the desired collapsible group and Liblib-style dock structure.
- [ ] Run the focused source test and verify the new assertions fail before implementation.

### Task 2: Add collapsible workflow group sidebar

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

**Interfaces:**
- Consumes existing `nodes`, `selectedIds`, `flowInstance`, `groups`, and node presentation helpers.
- Produces a sidebar with searchable visual groups, per-group collapse state, node focus buttons, and existing group/node actions.

- [ ] Add local collapsed-group state keyed by group id.
- [ ] Render each group as a folder row with count, expand/collapse control, and child node rows.
- [ ] Keep existing node focus and more-action callbacks unchanged.
- [ ] Add compact dark styling, separators, scroll behavior, and selected-row states.
- [ ] Run focused source tests and existing canvas command tests.

### Task 3: Align bottom canvas dock and node card presentation

**Files:**
- Modify: `frontend/src/user/components/AgentCanvasControls.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Test: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

- [ ] Group the existing display/zoom/conversation controls into a single floating dock.
- [ ] Preserve all existing accessible labels and command handlers.
- [ ] Apply Liblib-like compact icon buttons, pill background, border, and bottom-center placement.
- [ ] Reduce node card chrome while retaining handles, selected state, media content, and generation composer behavior.
- [ ] Run focused source tests and frontend build.

### Task 4: Runtime verification

**Files:**
- Generated: `frontend/dist/*`, `backend/web/dist/*` through existing build scripts only.
- Runtime: `backend/qiantie-local.exe`

- [ ] Run `npm.cmd --prefix frontend run build`.
- [ ] Run `scripts/build-go-local.ps1`.
- [ ] Commit only the task files and generated embed output required by the build path.
- [ ] Restart local Go runtime and verify `/health` is 200.
- [ ] Verify `/api/runtime-build-info` reports the new Git SHA and `runtime=go`.
- [ ] Open the local Agent Canvas route and visually check the workflow view without generating media.

## Test Commands

```powershell
node --test frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js
node --test frontend/src/user/pages/agent-canvas-commands.test.js
npm.cmd --prefix frontend run build
.\scripts\build-go-local.ps1
```
