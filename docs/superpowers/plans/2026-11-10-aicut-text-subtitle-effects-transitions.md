# AI-CUT Text, Subtitle, Effects, and Transitions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the first usable batch of AI-CUT text, subtitle, visual-effect, transition, and speech-recognition controls with matching preview, persistence, validation, and export behavior.

**Architecture:** Extend the existing timeline contract instead of introducing a second editor state. React renders previews from normalized timeline values; Go validates the same stable IDs and FFmpeg export maps those IDs to supported filters. Existing MySQL/Redis ASR jobs remain the only speech-recognition path.

**Tech Stack:** React, Ant Design, Vitest/Node tests, Go, MySQL, Redis, FFmpeg, TOS-compatible object storage.

## Global Constraints

- Work only in `F:\脚本测试\一战晟铭-v88` and preserve unrelated dirty changes.
- Do not create a branch, push, deploy, or make paid model calls.
- Frontend preview and backend export must consume the same stable preset IDs.
- API credentials remain in the Go backend configuration.
- Do not commit while unrelated worktree changes cannot be isolated safely.

---

### Task 1: Stable preset catalog and timeline validation

**Files:**
- Modify: `frontend/src/user/pages/aicutEffects.js`
- Modify: `frontend/src/user/pages/aicutEffects.test.js`
- Modify: `frontend/src/user/pages/aicutTimeline.js`
- Modify: `frontend/src/user/pages/aicutTimeline.test.js`
- Modify: `backend/internal/aicut/types.go`
- Modify: `backend/internal/aicut/types_test.go`

**Interfaces:**
- Produces: `visualEffectCatalog`, `transitionCatalog`, `textStyleCatalog`, `subtitleStyleCatalog`, `subtitleAnimationCatalog` and normalizers returning serializable timeline values.
- Consumes: existing `Timeline`, `Track`, `Clip`, `VisualEffect` JSON contracts.

- [ ] Add failing frontend tests proving all agreed preset IDs normalize to bounded values and unknown IDs fall back safely.
- [ ] Run `npm test -- --run src/user/pages/aicutEffects.test.js src/user/pages/aicutTimeline.test.js` from `frontend`; expect the new cases to fail.
- [ ] Implement catalogs and normalizers without adding a second state store.
- [ ] Add failing Go tests for all accepted transition/effect IDs and invalid numeric ranges.
- [ ] Run `go test ./internal/aicut`; expect the new validation cases to fail.
- [ ] Extend Go timeline validation and rerun both focused suites until they pass.

### Task 2: Real editor controls and theme-correct overlays

**Files:**
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-page.css`
- Modify: `frontend/src/user/pages/aicut-workbench.layout.test.js`
- Modify: `frontend/src/user/pages/aicut-controls.contract.test.js`

**Interfaces:**
- Consumes: Task 1 catalogs and normalizers.
- Produces: usable cards, property controls, disabled reasons, and editor-scoped Ant Design overlay theme.

- [ ] Add failing source/layout tests for usable add/drag actions, effect parameters, transition placement validation, and dark/light modal classes.
- [ ] Run the two focused frontend test files and verify failure.
- [ ] Replace static feature cards with catalog-driven controls that update the selected unlocked clip or subtitle track.
- [ ] Mount popup containers inside the AI-CUT theme root and style modals, selects, dropdowns, tooltips, and color pickers from theme variables.
- [ ] Rerun focused tests and verify pass.

### Task 3: Text and subtitle timeline operations

**Files:**
- Modify: `frontend/src/user/pages/aicutTimeline.js`
- Modify: `frontend/src/user/pages/aicutTimeline.test.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicutPreview.js`
- Modify: `frontend/src/user/pages/aicutPreview.test.js`

**Interfaces:**
- Produces: create, split, merge, duplicate, replace, style, animation, and transform operations for text/subtitle clips.
- Consumes: existing `commitTimeline` undo/redo path and Task 1 preset IDs.

- [ ] Add failing tests for empty text rejection, split/merge timing, duplicate IDs, locked-track rejection, style persistence, and active preview animation metadata.
- [ ] Run focused timeline/preview tests and verify failure.
- [ ] Implement pure timeline operations first, then wire buttons and direct player editing through `commitTimeline`.
- [ ] Ensure every action is undoable and does not mutate the previous timeline object.
- [ ] Rerun focused tests and verify pass.

### Task 4: ASR interaction completion

**Files:**
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-page.css`
- Modify: `frontend/src/user/pages/aicut-workbench.layout.test.js`
- Modify: `backend/internal/aicut/asr_service.go`
- Modify: `backend/internal/aicut/asr_service_test.go`

**Interfaces:**
- Consumes: existing `createAICutASRJob`, `listAICutASRJobs`, `cancelAICutASRJob` APIs.
- Produces: visible source selection, duplicate-job protection, append-new-track behavior, conflict errors, progress, cancel, and retry controls.

- [ ] Add failing Go test that a second active job for the same project and asset is rejected without creating another task.
- [ ] Add failing frontend contract tests for media selection, no-model navigation, retry, cancellation, and non-destructive completion copy.
- [ ] Implement backend duplicate protection and frontend task controls.
- [ ] Run `go test ./internal/aicut` and focused frontend ASR/layout tests; verify pass.

### Task 5: Export parity for supported presets

**Files:**
- Modify: `backend/internal/aicut/export.go`
- Modify: `backend/internal/aicut/export_test.go`
- Modify: `backend/internal/aicut/ffmpeg.go`
- Modify: `backend/internal/aicut/ffmpeg_test.go`

**Interfaces:**
- Consumes: validated timeline preset IDs from Task 1.
- Produces: deterministic FFmpeg arguments for the supported transition, effect, text, and subtitle subset; unsupported IDs return an explicit error.

- [ ] Add table-driven failing tests for every supported preset and for unsupported-effect errors.
- [ ] Run `go test ./internal/aicut`; verify failure.
- [ ] Map supported IDs to bounded FFmpeg filter expressions and escape user text/font paths safely.
- [ ] Rerun focused Go tests and verify pass.

### Task 6: Cross-feature acceptance and local runtime proof

**Files:**
- Modify: `docs/acceptance/aicut-first-effects-subtitles.md`
- Modify only if required by failures: files owned by Tasks 1-5.

**Interfaces:**
- Consumes: completed Tasks 1-5.
- Produces: repeatable evidence for normal, empty, duplicate, invalid, and canceled operations.

- [ ] Run focused frontend tests and `npm run build`.
- [ ] Run `go test ./internal/aicut ./internal/httpapi`.
- [ ] Rebuild embedded frontend assets using the repository's existing local build path.
- [ ] Restart only the confirmed local `qiantie-local.exe` process and verify `/health` plus `/api/runtime-build-info`.
- [ ] In the logged-in browser, test one video with the same inputs for add/edit/undo/refresh/export, ASR success or configured-model absence, duplicate action, invalid placement, and cancel.
- [ ] Record screenshots, runtime identity, output path, and any remaining unavailable capability in the acceptance document.

