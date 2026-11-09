# Private Comic Single Model Config Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the conflicting private-comic model controls so the current comic has one effective configuration source.

**Architecture:** Keep account-level script defaults in `ScriptPage`, but treat them only as initial values for empty private-comic drafts. In comic mode, reuse the existing modal as a current-comic editor and render the toolbar model values as read-only summaries. Script mode keeps its existing default-saving behavior.

**Tech Stack:** React, Ant Design, Node test runner, Vite, Go embed runtime.

## Global Constraints

- Work directly on the maintained `v88` branch; do not create another branch.
- Do not change or deploy production.
- Use test-first development and verify the embedded local runtime on port 4001.
- Preserve unrelated worktree files.

---

### Task 1: Lock the single-entry behavior with source tests

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.source.test.js`

**Interfaces:**
- Consumes: `PrivateComicWorkspace.jsx` and `ScriptPage.jsx` source text.
- Produces: regression assertions for read-only toolbar summaries and comic-specific modal save behavior.

- [ ] **Step 1: Write failing tests**

Add assertions that the workspace has no model `Select` change handlers, displays resolved model labels, and that `ScriptPage` initializes/saves the modal from/to `privateComicDraft` in comic mode.

- [ ] **Step 2: Run the focused test and confirm RED**

Run: `node --test src/user/pages/PrivateComicWorkspace.source.test.js`

Expected: FAIL because editable toolbar selects and global-only save behavior still exist.

- [ ] **Step 3: Commit only after the implementation passes**

The test and implementation belong in the same behavior commit.

### Task 2: Make the comic modal the only editor

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Modify: `frontend/src/user/pages/ScriptPage.jsx`

**Interfaces:**
- Consumes: `privateComicDraft`, `scriptConfigDraft`, `customizationWorkspace`, available model catalogs.
- Produces: comic-specific modal initialization and save path; read-only model summaries.

- [ ] **Step 1: Implement minimal comic-specific modal behavior**

When `customizationWorkspace === 'comic'`, open from `privateComicDraft` with account defaults only as fallbacks; save `textModelId`, `imageProvider`, and `imageModelId` back to `privateComicDraft`; use comic-specific title, button copy, and explanatory copy; hide video-only fields.

- [ ] **Step 2: Seed only empty private-comic model fields**

After model/default loading, fill empty `privateComicDraft` fields from `scriptModelSelection` and `scriptImageProvider` without overwriting restored current-comic values.

- [ ] **Step 3: Replace editable toolbar selects**

Resolve the selected IDs against all available text/image models and render plain read-only values. Keep the existing “配置” button as the only edit action.

- [ ] **Step 4: Run focused tests and confirm GREEN**

Run: `node --test src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/privateCustomizationComic.test.js`

Expected: all tests pass.

### Task 3: Build and verify the real local runtime

**Files:**
- Generated: `frontend/dist/`
- Generated: `backend/web/dist/`
- Generated: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: verified React source.
- Produces: current Go embedded local runtime on port 4001.

- [ ] **Step 1: Run the full frontend test suite**

Run: `npm test -- --runInBand`

Expected: zero failures.

- [ ] **Step 2: Build and restart the approved local Go runtime**

Run the project-provided local Go review/build script used by the current V88 runtime; restart only the confirmed `qiantie-local.exe` listener.

- [ ] **Step 3: Verify runtime identity and health**

Check `/health` and `/api/runtime-build-info`; both must respond successfully and identify the new Git commit.

- [ ] **Step 4: Verify the visible page**

Open `http://localhost:4001/script`, switch to manga generation, and confirm the toolbar is read-only while the modal edits the current manga values.

- [ ] **Step 5: Commit**

Commit source, tests, and required embedded assets with a traceable message such as `fix: unify private comic model configuration`.

