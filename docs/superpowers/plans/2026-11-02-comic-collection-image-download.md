# Comic Collection Image Download Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a user download every selected primary image in one comic collection as a correctly named ZIP file.

**Architecture:** The backend exposes one authenticated collection-download route. It validates collection ownership and page completeness, builds a ZIP from the collection's ordered primary images, and returns it as a download. The existing collection card calls that route and triggers a browser Blob download.

**Tech Stack:** Go `net/http` and `archive/zip`; React with Ant Design; existing Batch Factory store and authenticated API client.

## Global Constraints

- ZIP name is the sanitized book title plus `.zip`.
- Internal image names are sanitized book title plus `-001`, `-002`, and so on, ordered by collection page order.
- Download only collection primary images and reject incomplete collections.
- Preserve existing dirty worktree changes and do not commit, push, or deploy.

---

### Task 1: Server ZIP download

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/httpapi/batch_factory_comic_collection_download_test.go`

**Interfaces:**
- Produces: `GET /api/batch-factory/v12/batches/{batchID}/books/{bookID}/comic-collections/{collectionID}/download`
- Produces: a ZIP response with `Content-Disposition: attachment`.

- [ ] **Step 1: Write the failing route test**

Create an in-memory batch, two comic pages with selected primary images, and a collection. Request the route and assert HTTP 200, ZIP response headers, and ordered entries `测试书-001.png`, `测试书-002.png`.

- [ ] **Step 2: Run the focused test and verify it fails because the route is absent.**

- [ ] **Step 3: Add the smallest server handler.**

Load the owner-scoped batch and collection, validate every item has a readable primary image, write the ordered bytes to a ZIP, then send the ZIP as an attachment.

- [ ] **Step 4: Add and pass the incomplete-collection test.**

Create a collection item without an available primary image and assert the handler returns an error instead of an incomplete archive.

- [ ] **Step 5: Run the focused Go tests.**

### Task 2: Collection-card download control

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: the existing Batch Factory API client module that contains `downloadComicPromptZip`
- Test: `frontend/src/user/pages/shuihuo/comicMediaWorkbench.source.test.js`

**Interfaces:**
- Consumes: the collection download route from Task 1.
- Produces: a `下载图片` action for each comic collection card.

- [ ] **Step 1: Write a failing source test**

Assert the collection card exposes `下载图片` and the implementation invokes a collection ZIP download function.

- [ ] **Step 2: Run the focused frontend test and verify it fails.**

- [ ] **Step 3: Add the API client function and button handler.**

Fetch the download route through the existing authenticated client, create a Blob URL, and use the sanitized book title for the browser download name. Disable only that collection's button while packing.

- [ ] **Step 4: Run the focused frontend test and verify it passes.**

### Task 3: Build and local acceptance

**Files:**
- Verify: `backend/qiantie-local.exe`
- Verify: `http://localhost:4001/batch-factory`

- [ ] **Step 1: Run affected Go and frontend tests.**
- [ ] **Step 2: Build embedded frontend assets and rebuild the local Go executable.**
- [ ] **Step 3: Restart only the confirmed local `qiantie-local.exe` listener on port 4001.**
- [ ] **Step 4: Verify health and the download route with a collection containing images.**
