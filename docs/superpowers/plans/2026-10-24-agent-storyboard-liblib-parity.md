# Agent Storyboard Liblib Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the V88 Storyboard material workspace match Liblib's measured desktop proportions, independent draggable regions, and media density without removing existing V88 functions.

**Architecture:** Keep `AgentPageV2` as the owner of the outer director/material split and `AgentStoryboardWorkspace` as the owner of the material grid. Persist normalized layout values through the existing Storyboard view state and keep the current inspector, material APIs, and generation paths unchanged.

**Tech Stack:** React 18, Ant Design, Lucide, CSS Grid/Flexbox, Node test runner, Go embed runtime.

## Global Constraints

- Work only on Git branch `v88` and preserve all unrelated dirty changes.
- Do not copy Liblib source, assets, branding, or compiled output.
- Do not change backend APIs or trigger paid generation.
- Do not publish the public site.
- Do not commit the dirty checkout without separate explicit approval.
- Frontend output must be rebuilt and embedded through the existing Go build scripts.

---

### Task 1: Normalize Storyboard Layout State

**Files:**
- Modify: `frontend/src/user/pages/agent-storyboard-split.js`
- Modify: `frontend/src/user/pages/agent-storyboard-split.test.js`
- Modify: `frontend/src/user/pages/agent-storyboard-state.js`
- Modify: `frontend/src/user/pages/agent-storyboard-state.test.js`

**Interfaces:**
- Consumes: saved `viewState.storyboard` data.
- Produces: normalized `directorSplit`, `columnSplit`, `leftRowSplit`, `rightRowSplit`, and `sectionOrder` values.

- [ ] Add failing tests for defaults 31/25/66/55, clamping, per-split resets, and valid unique section slots.
- [ ] Run the focused split/state tests and confirm the old `{ column: 39, row: 50 }` contract fails.
- [ ] Implement normalization helpers without changing category, item order, or selection behavior.
- [ ] Re-run the focused tests and confirm they pass.

### Task 2: Build Three Independent Material Resize Controls

**Files:**
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Modify: `frontend/src/user/pages/agent-storyboard-workspace.source.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: normalized material ratios and `onStoryboardStateChange(patch)`.
- Produces: column, left-row, and right-row splitters with pointer drag and double-click reset.

- [ ] Add failing source tests requiring three named splitters, distinct CSS variables, independent columns, and reset handlers.
- [ ] Run the source test and confirm the current shared row splitter fails.
- [ ] Replace the shared four-cell grid with two independently split material columns.
- [ ] Keep loading, error, empty state, section filters, cards, and inspector behavior intact.
- [ ] Re-run split, state, and source tests.

### Task 3: Add The Outer Director Resize Control

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `viewState.storyboard.directorSplit`.
- Produces: a desktop-only director/material separator that persists its ratio through `handleStoryboardStateChange`.

- [ ] Add failing layout tests for the 31/69 default, accessible separator, drag cursor, and double-click reset.
- [ ] Run the focused layout test and confirm failure.
- [ ] Add the outer splitter and clamp it against usable minimum widths.
- [ ] Preserve the existing responsive overlay behavior at 1000px and below.
- [ ] Re-run the focused layout tests.

### Task 4: Match Material Density And Remove Empty Space

**Files:**
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Modify: `frontend/src/user/pages/agent-storyboard-workspace.source.test.js`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: existing material sections and shot loading/error data.
- Produces: compact toolbar, 32px text/audio rows, dense image/video grids, conditional shot strip, and Storyboard-only pet hiding.

- [ ] Add failing tests for conditional shot rendering, compact media sizes, and the Storyboard pet rule.
- [ ] Run the tests and confirm the current unconditional shot strip and card sizes fail.
- [ ] Implement density rules and render the shot strip only for loading, error, or real shots.
- [ ] Add section-slot drag affordance while preserving click-to-collapse and card reorder.
- [ ] Run all Storyboard-focused tests.

### Task 5: Build And Live Browser Acceptance

**Files:**
- Generated: `frontend/dist/**`
- Generated: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: tested React source.
- Produces: the local Go-embedded runtime at `http://localhost:4001`.

- [ ] Run `npm.cmd --prefix frontend run build`.
- [ ] Run `scripts/build-go-local.ps1`.
- [ ] Restart with `scripts/start-local-go-review.ps1 -UseYcmMysql` only after verifying the current `:4001` executable path belongs to this workspace.
- [ ] Verify `/api/runtime-build-info` and the active listener.
- [ ] At 1890 x 1070, measure the rendered 31/69 and 25/75 geometry.
- [ ] Drag all four separators, double-click reset each, and verify independent movement.
- [ ] Verify empty-shot removal, hidden pet, scroll isolation, material selection, filters, batch mode, and add-to-conversation without triggering generation.

