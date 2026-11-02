# Batch Factory Comic Reference Chips Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show comic reference images as named, previewable, horizontally scrolling material chips with source-aware deletion rules.

**Architecture:** The Go preview handler describes each effective reference URL as a structured item while preserving the existing URL array for compatibility. A focused React component loads protected thumbnails, renders Ant Design previews, and only calls the existing page-reference save path for manually uploaded items.

**Tech Stack:** Go, React, Ant Design, Node test runner, Go test.

## Global Constraints

- Work directly in the current development branch; do not create a branch or commit unless explicitly requested.
- Preserve unrelated dirty files.
- System prompt text remains in the Go prompt module; the frontend must not hard-code prompt contents.
- Built-in references are view-only here; only manually uploaded page references are removable.

---

### Task 1: Structured reference metadata

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Test: `backend/internal/httpapi/batch_factory_comic_pages_test.go`

**Interfaces:**
- Consumes: effective reference URLs from `BuildComicImageContract`.
- Produces: `comicReferenceImageResponse` and `referenceImages` in the prompt-preview response.

- [x] **Step 1: Write a failing Go test** that checks an urban scene reference is named `都市`, marked `scene-library`, and not removable, while a page upload is named `手动参考图 1`, marked `manual-upload`, and removable.
- [x] **Step 2: Run the focused Go test** and confirm it fails because `referenceImages` is absent.
- [x] **Step 3: Implement URL classification and response mapping** without changing reference ordering or the existing `referenceImageUrls` field.
- [x] **Step 4: Run the focused Go test** and confirm it passes.

### Task 2: Reference material chip component

**Files:**
- Create: `frontend/src/user/pages/shuihuo/ComicReferenceChips.jsx`
- Create: `frontend/src/user/pages/shuihuo/ComicReferenceChips.test.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo-production.css`

**Interfaces:**
- Consumes: `{ id, name, url, source, removable }[]`, authenticated blob loader, and `onRemove(url)`.
- Produces: a one-row `.batch-factory-comic-reference-strip` with image preview and source-aware remove controls.

- [x] **Step 1: Write failing component tests** for visible names/count, thumbnail rendering, locked system items, and removable manual items.
- [x] **Step 2: Run the focused frontend test** and confirm it fails because the component does not exist.
- [x] **Step 3: Implement the component and styles**, using the authenticated blob loader and Ant Design image preview.
- [x] **Step 4: Wire it into the comic prompt panel** and use the existing page-reference save call to remove one manual item after confirmation.
- [x] **Step 5: Run the focused frontend tests** and confirm they pass.

### Task 3: Runtime verification

**Files:**
- Verify only; no planned source changes.

**Interfaces:**
- Consumes: built frontend and Go-embedded runtime.
- Produces: verified local UI at `http://localhost:4001`.

- [x] **Step 1: Run relevant frontend and Go test suites.**
- [x] **Step 2: Build the frontend and Go local binary.**
- [x] **Step 3: Restart the local `:4001` service and run the local health check.**
- [x] **Step 4: Open the target batch, inspect the comic prompt dialog, preview a thumbnail, and verify delete visibility.**
