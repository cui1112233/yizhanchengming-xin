# AI-CUT Resizable Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add real panel resizing, independent collapse, material drag-and-drop, project creation access, and working keyboard shortcuts to the V88 AI-CUT page.

**Architecture:** Keep project data in the existing AI-CUT APIs and timeline helpers. Put pure layout, drop-position, and keyboard decisions in a small `aicutWorkbench.js` module, while `AICutPage.jsx` owns React state and pointer/keyboard event wiring. CSS grid variables render the saved sizes and collapsed states without hard-coding theme colors.

**Tech Stack:** React, Ant Design, Lucide React, browser Pointer Events, HTML Drag and Drop, Node test runner, Vite, Go embed runtime.

## Global Constraints

- Work only on branch `v88`; do not create a new branch.
- Preserve the dirty worktree and unrelated changes.
- Do not push, deploy publicly, call paid models, or use real credentials.
- System prompts stay in the Go backend; this feature adds no frontend prompt text.
- Colors must continue to use existing theme variables.

---

### Task 1: Pure workbench interaction rules

**Files:**
- Create: `frontend/src/user/pages/aicutWorkbench.js`
- Create: `frontend/src/user/pages/aicutWorkbench.test.js`

**Interfaces:**
- Produces: `clampAICutLayout(layout, viewport)`, `timelineDropTime(clientX, rectLeft, scrollLeft, zoom)`, and `resolveAICutShortcut(eventLike, context)`.

- [ ] Write tests for minimum/maximum panel sizes, drop-time conversion, ignored typing, and supported shortcut actions.
- [ ] Run `node --test src/user/pages/aicutWorkbench.test.js` and confirm the missing-module failure.
- [ ] Implement the three pure helpers with no DOM dependency.
- [ ] Run the same test and confirm PASS.

### Task 2: React wiring and CSS layout

**Files:**
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-page.css`
- Modify: `frontend/src/user/pages/aicut-workbench.layout.test.js`

**Interfaces:**
- Consumes: Task 1 helpers and existing `createAICutProject`, timeline mutation, save, undo, redo, export, and preview functions.
- Produces: two vertical resize handles, one horizontal resize handle, material drop targets, material/timeline collapse controls, a topbar create button, and shortcut modal/handlers.

- [ ] Add failing layout-source assertions for handles, collapse controls, new-project access, and shortcut dialog.
- [ ] Run focused tests and confirm the new assertions fail.
- [ ] Wire layout state, pointer handlers, drag/drop handlers, keyboard actions, and modal controls in `AICutPage.jsx`.
- [ ] Add CSS grid variables, resize-handle hit areas, collapsed states, and responsive fallbacks in `aicut-page.css`.
- [ ] Run focused tests and confirm PASS.

### Task 3: Build and local runtime acceptance

**Files:**
- Modify only if evidence is recorded: `docs/acceptance/aicut-resizable-workbench.md`

**Interfaces:**
- Consumes: completed frontend source and existing Go embed build path.
- Produces: local browser evidence for the requested interactions.

- [ ] Run `node --test src/user/pages/aicutWorkbench.test.js src/user/pages/aicut-workbench.layout.test.js src/user/pages/aicutTimeline.test.js` from `frontend`.
- [ ] Run `npm run build` from `frontend`.
- [ ] Rebuild/restart only the confirmed local `qiantie-local.exe` through the repository script.
- [ ] Verify `/health`, `/api/runtime-build-info`, and `http://127.0.0.1:4001/ai-cut`.
- [ ] Browser-test resize, collapse/restore, material drop, new-project cancel, shortcut panel, and theme switching; record honest pass/partial results.

## Self-review

- Coverage: every requested interaction maps to Task 1 or Task 2; Task 3 proves the served page rather than only the source.
- No placeholders: each step names files, commands, outputs, and failure boundaries.
- Type consistency: layout helpers return plain numeric/state values consumed directly by the React page.

