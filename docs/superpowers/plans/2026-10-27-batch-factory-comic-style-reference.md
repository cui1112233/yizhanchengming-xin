# Batch Factory Comic Style Reference Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the selected comic style image the default global reference for character assets and comic pages while excluding scene and prop assets and retiring the scene reference library from the user workflow.

**Architecture:** The frontend derives a typed global-reference snapshot from the selected style and preserves a manual upload as an override. The backend receives the asset kind when compiling an image contract and independently enforces that only character assets can receive the global reference. Legacy scene-library references are ignored by the generation contract.

**Tech Stack:** React, Ant Design, Node test runner, Go, Go `testing`, Go embedded frontend.

## Global Constraints

- Do not create a branch, commit, reset, clean, or alter unrelated dirty files.
- System prompt bodies remain in the backend prompt module; the frontend stores preset metadata only.
- The scene reference library remains stored for historical compatibility but is absent from the user workflow and new generation requests.
- `applyToAssets` means character assets only.

---

### Task 1: Frontend style-reference state

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.test.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.js`

**Interfaces:**
- Produces: `comicStylePreviewReference(preset)` returning a `style-library` reference with enabled asset/page scopes.
- Produces: `comicReferenceForSelectedStyle(currentReference, preset)` preserving manual uploads and replacing legacy style/scene defaults.

- [ ] Write failing tests proving the style image becomes the default reference, a manual upload survives a style change, and a legacy scene reference is replaced.
- [ ] Run `node --test src/user/pages/shuihuo/batchFactoryComicStyle.test.js` and verify the new assertions fail because selection currently stores style metadata only.
- [ ] Implement the two helpers with explicit `source` values and default switches.
- [ ] Re-run the focused test and verify it passes.

### Task 2: Frontend picker and manual override behavior

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicStyle.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryComicStylePicker.jsx`

**Interfaces:**
- Consumes: `comicReferenceForSelectedStyle(currentReference, preset)`.
- Keeps: `uploadComicGlobalReference(batchId, dataUrl)` for manual replacement.

- [ ] Add source assertions that the picker no longer renders `场景参考图库`, says the style image is sent to AI, and labels the asset switch `用于资产图（仅人物）`.
- [ ] Run the focused frontend test and verify those source assertions fail.
- [ ] Remove scene catalog filtering and selection UI, update `selectStyle` to save both style and derived reference, mark uploaded references as `manual-upload`, and make delete restore the selected style reference.
- [ ] Re-run the focused frontend test and verify it passes.

### Task 3: Backend generation enforcement

**Files:**
- Modify: `backend/internal/batchfactoryv11/comic_image_contract_test.go`
- Modify: `backend/internal/batchfactoryv11/comic_image_contract.go`
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_reference.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_reference_test.go`

**Interfaces:**
- Changes: `BuildComicImageContract(settings, purpose, assetKind, content, assetImageURLs, pageReferenceURLs)`.
- Uses asset kinds `character`, `scene`, and `prop`; page calls pass an empty asset kind.

- [ ] Replace the former “style previews never become references” test with tests proving style/manual references reach characters and pages, never scenes/props, and legacy `scene:` references never reach prompts or reference URLs.
- [ ] Add an upload-route assertion that uploaded references return `source: manual-upload`.
- [ ] Run `go test ./internal/batchfactoryv11 ./internal/httpapi -run 'ComicImageContract|BatchComicReference' -count=1` and verify failures reflect the old behavior.
- [ ] Add `assetKind` to the image contract, permit `style-library` and `manual-upload` references, reject legacy scene references, and apply the asset reference only when `assetKind == "character"`.
- [ ] Pass `asset.Kind` from asset generation and an empty kind from page generation; mark uploaded references as `manual-upload`.
- [ ] Re-run the focused Go tests and verify they pass.

### Task 4: Regression, build, and live acceptance

**Files:**
- Generated: `frontend/dist/**`
- Generated: `backend/web/dist/**`
- Generated: local Go executable via `scripts/build-go-local.ps1`

**Interfaces:**
- Uses the existing local `:4001` runtime and `/api/runtime-build-info` check.

- [ ] Run the full comic-style frontend tests and focused asset/page Go tests.
- [ ] Run `npm --prefix frontend run build`.
- [ ] Run `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`.
- [ ] Restart the local service using the existing project script without changing MySQL, Redis, or credentials.
- [ ] Open the Batch Factory comic settings and verify the scene library is absent, selecting a style fills the global reference, all three switches default on, and the asset label says only characters.
- [ ] Verify the running version using `/api/runtime-build-info` and `scripts/check-local-go.ps1`.
- [ ] Do not submit a paid image-generation request during UI acceptance; generation is resumed only after configuration acceptance.
