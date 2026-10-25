# Batch Factory Comic Style And Reference Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the 25 recovered comic art-style presets and previews to Batch Factory, compile style text only on the Go backend, and pass explicitly selected reference images to comic asset and comic-page image generation.

**Architecture:** Extend the existing prompt catalog with a dedicated `batch.comic-style` slot and backend-served embedded previews. Add one Batch Factory comic-image compiler that resolves batch/book inheritance, builds the final prompt, and orders reference images; both asset generation and comic-page generation call it. The existing candidate, primary-image, collection, provider, local executor, and settings-patch mechanisms remain authoritative.

**Tech Stack:** Go, embedded JSON/JPEG resources, React, Ant Design, existing Batch Factory settings JSON, local artifact storage/TOS-compatible media URLs, Node test runner, Go tests.

## Global Constraints

- Work directly on the current `v88` branch; do not create another branch.
- Preserve unrelated dirty files and never reset or clean the worktree.
- System prompt bodies live only in `backend/internal/prompt`; the frontend stores preset metadata only.
- The three comic-page prompt presets remain separate from the 25 art-style presets.
- Selecting an art-style card does not send its preview image to AI unless the user explicitly chooses `设为漫画全局参考图`.
- Comic mode and video mode remain mutually exclusive; comic fields must not affect video generation.
- Frontend changes must be rebuilt and served through the Go embedded-binary path.

---

### Task 1: Publish 25 Art-Style Presets And Preview Resources

**Files:**
- Modify: `backend/internal/prompt/defaults/comic_presets.json`
- Modify: `backend/internal/prompt/comic_presets.go`
- Modify: `backend/internal/prompt/catalog.go`
- Modify: `backend/internal/httpapi/presets.go`
- Create: `backend/internal/prompt/defaults/comic_style_previews/01.jpg` through `25.jpg`
- Modify: `backend/internal/prompt/comic_presets_test.go`
- Modify: `backend/internal/httpapi/presets_test.go`

**Interfaces:**
- Produces catalog items with `slot="batch.comic-style"` and `previewUrl="/api/presets/comic-style/{presetId}/preview"`.
- Produces `prompt.OpenComicStylePreview(presetID string) ([]byte, string, bool)` for the authenticated preview route.

- [ ] Write tests that require exactly 25 published `batch.comic-style` entries, unique IDs, non-empty backend bodies, and valid JPEG previews.
- [ ] Run the focused prompt and preset-route tests and confirm they fail because the style slot and preview endpoint do not exist.
- [ ] Copy the 25 recovered JPEGs into the embedded backend resource directory and add 25 style records from `DECRYPTED_ART_STYLES.json`.
- [ ] Extend public preset metadata with `previewUrl` while continuing to hide `body`.
- [ ] Add the authenticated preview route with `image/jpeg`, `nosniff`, and immutable cache headers.
- [ ] Run the focused tests and confirm all 25 preset-to-preview mappings pass.
- [ ] Commit only Task 1 files with `feat(batch-factory): add comic style catalog`.

### Task 2: Compile Comic Style And References On The Backend

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_prompt_config.go`
- Modify: `backend/internal/batchfactoryv11/director_contract.go`
- Modify: `backend/internal/batchfactoryv11/comic_prompt_service.go`
- Create: `backend/internal/batchfactoryv11/comic_image_contract.go`
- Create: `backend/internal/batchfactoryv11/comic_image_contract_test.go`
- Modify: `backend/internal/batchfactoryv11/comic_prompt_service_test.go`

**Interfaces:**
- Consumes `aiPromptConfig.comic.style`, `imagePromptPrefix`, `imagePromptSuffix`, and `globalReferenceImage` from effective settings.
- Produces `BuildComicImageContract(settings SettingsPatch, pagePrompt string, assetPrompts []string, assetImageURLs []string, pageReferenceURLs []string) (ComicImageContract, error)`.
- `ComicImageContract` exposes `Prompt`, `ReferenceImageURLs`, `StylePresetID`, and `SnapshotHash`.

- [ ] Write failing tests for batch-to-book inheritance, style snapshot hydration, final prompt order, preview exclusion, explicit global-reference inclusion, and video-mode isolation.
- [ ] Run the focused tests and confirm failures identify the missing style slot/compiler.
- [ ] Hydrate `comic.style` from `batch.comic-style` exactly like the existing comic page/assets selections.
- [ ] Split the fixed Q-version `comicReferenceStyleLock` into style-neutral layout/text safety rules and remove unconditional Q-style wording from compilation.
- [ ] Implement deterministic prompt order: style body, user prefix, structure rules, asset/page content, user suffix, negative constraints.
- [ ] Implement deterministic reference order with de-duplication and explicit inclusion only; preview URLs are never inferred from the style selection.
- [ ] Run focused compiler and comic-prompt tests.
- [ ] Commit only Task 2 files with `feat(batch-factory): compile comic image contracts`.

### Task 3: Send References Through Asset And Comic-Page Generation

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Modify: `backend/internal/httpapi/batch_factory_v12_asset_generation.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Modify: `backend/internal/httpapi/agent_image_generator.go`
- Modify: `backend/internal/httpapi/batch_factory_v11_asset_images_test.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_pages_test.go`

**Interfaces:**
- Extends `generateShuihuoImageArtifact` with `referenceImageURLs []string`.
- Local Doubao payload uses the existing `referenceImages`/`imageUrls` fields.
- Unsupported providers return a clear conflict error instead of silently dropping references.

- [ ] Write failing HTTP tests proving asset generation includes the global reference when `applyToAssets=true` and comic-page generation includes it when `applyToPages=true`.
- [ ] Add tests proving either scope switch removes the image and video mode sends no comic references.
- [ ] Extend the shared image-generation helper to pass references to the local executor and provider adapters.
- [ ] Make asset generation call the comic compiler only for comic-source assets in comic mode.
- [ ] Make comic-page generation compile the final prompt server-side and ignore any browser attempt to supply a style body.
- [ ] Return model-capability errors in user-facing Chinese when references are unsupported.
- [ ] Run the HTTP and local-executor image-generation tests.
- [ ] Commit only Task 3 files with `feat(batch-factory): send comic reference images`.

### Task 4: Add Unified And Per-Book Art-Style UI

**Files:**
- Create: `frontend/src/user/pages/shuihuo/BatchFactoryComicStylePicker.jsx`
- Create: `frontend/src/user/pages/shuihuo/BatchFactoryComicStylePicker.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactory.css`

**Interfaces:**
- `BatchFactoryComicStylePicker({catalog, value, onChange, onUsePreviewAsReference})` renders image cards and returns preset metadata only.
- Unified and book forms read/write the same `aiPromptConfig.comic` shape.

- [ ] Write source tests requiring the 25-card picker, selected state, inherited/custom state, global prefix/suffix fields, reference preview, and two scope switches.
- [ ] Run the focused frontend tests and confirm the new component is missing.
- [ ] Implement a compact Ant Design modal/grid with 3:4 cards, lazy images, names, categories, keyboard selection, preview, and explicit `设为漫画全局参考图` command.
- [ ] Mount the same control in unified constraints and per-book constraints; show inherited values without duplicating the fields elsewhere.
- [ ] Keep the three comic-page prompts in the existing comic prompt selector and label the new control `画风库`.
- [ ] Add stable responsive dimensions so card loading and long names do not shift or overlap the layout.
- [ ] Run focused frontend tests.
- [ ] Commit only Task 4 files with `feat(batch-factory): add comic style picker`.

### Task 5: Reference Upload, Stale-State Feedback, And End-To-End Verification

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactory.css`

**Interfaces:**
- Adds authenticated upload/content endpoints for the comic global reference using existing image validation and artifact storage.
- Generation responses expose the compiler `snapshotHash`; mismatches produce `配置已变化，建议重新生成` without deleting old candidates.

- [ ] Write failing route/UI tests for upload, removal, scope changes, snapshot mismatch, and preserved old candidates.
- [ ] Implement upload/content access through the existing owner-authenticated image store and save only stable resource metadata in settings.
- [ ] Add stale badges to affected asset images and comic pages; do not delete or replace candidates automatically.
- [ ] Run focused Go and frontend tests.
- [ ] Run `go test ./backend/internal/prompt ./backend/internal/batchfactoryv11 ./backend/internal/httpapi` from the repository module layout.
- [ ] Run the frontend Batch Factory tests and production build.
- [ ] Rebuild/restart through the Go embedded-binary path and verify `/health` plus `/api/runtime-build-info`.
- [ ] Browser-check desktop and mobile layouts, 25 preview images, unified/book inheritance, and video-mode isolation without spending image quota.
- [ ] With explicit generation authorization already provided, use the existing book and Doubao executor for one asset image and one comic page, then verify candidate, primary image, and comic collection persistence.
- [ ] Commit only Task 5 files with `feat(batch-factory): finish comic style workflow`.
