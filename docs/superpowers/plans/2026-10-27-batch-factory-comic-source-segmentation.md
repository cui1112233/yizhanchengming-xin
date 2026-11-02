# Batch Factory Comic Source Segmentation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every comic page own one lossless, continuous novel-source segment, normally 90-120 Chinese characters and never split inside a complete sentence or dialogue, with read-only display and safe left/right manual splitting.

**Architecture:** Add a focused segmentation domain module and a durable per-video source-segment table. Comic planning creates active video-backed page rows from validated source ranges atomically; prompt extraction then consumes those rows. The existing comic page/image tables remain intact, while split operations preserve the current page ID, create one adjacent page, mark only affected prompts stale, and revalidate full-book continuity before commit.

**Tech Stack:** Go 1.x, MySQL 8, React, Ant Design, existing V12 HTTP client, Go embed frontend runtime.

## Global Constraints

- Normal page target is 90-120 source characters; normal upper limit is 150.
- Sentence, quotation, continuous dialogue, paragraph, scene, time, and POV completeness outrank the 150-character limit.
- Concatenating saved segments must reproduce the source text byte-for-byte.
- System prompts live only in the backend prompt/service layer; React must not hardcode them.
- Re-segmentation and manual splitting must not delete historical candidate images or primary images.
- Manual splitting must not call a model or create image tasks.
- Add a new migration; never alter an applied migration.
- Do not call real generation providers during tests or browser verification.

---

### Task 1: Safe-boundary segmentation engine

**Files:**
- Create: `backend/internal/batchfactoryv11/comic_source_segmenter.go`
- Create: `backend/internal/batchfactoryv11/comic_source_segmenter_test.go`

**Interfaces:**
- Produces: `ComicSourceBoundary`, `ComicSourceSegmentDraft`, `EnumerateComicSourceBoundaries(source string)`, `FallbackComicSourceSegments(source string, boundaries []ComicSourceBoundary)`, `ValidateComicSourceSegments(source string, segments []ComicSourceSegmentDraft)`.

- [ ] **Step 1: Write failing table tests** for Chinese sentence endings, nested quotation marks, continuous dialogue, paragraph boundaries, a single unit over 150 characters, short-tail merging, and exact source reconstruction.
- [ ] **Step 2: Run** `go test ./internal/batchfactoryv11 -run 'Test(Enumerate|Fallback|Validate)ComicSource' -count=1` and confirm failure because the new interfaces do not exist.
- [ ] **Step 3: Implement rune-aware boundary enumeration** while retaining byte offsets for exact slicing. Treat punctuation inside an open quote as unsafe until the closing quote, and coalesce adjacent dialogue lines into one indivisible unit.
- [ ] **Step 4: Implement deterministic packing** that chooses the best safe boundary in 90-120, may extend to 150, keeps overlong indivisible units whole, and merges a final short page only when the merged page stays at or below 150.
- [ ] **Step 5: Implement strict validation** for start/end coverage, adjacency, exact text slices, unique ranges, legal boundaries, and exact reconstruction.
- [ ] **Step 6: Rerun the focused tests** and require PASS.

### Task 2: Durable segment model and append-only schema

**Files:**
- Modify: `backend/internal/batchfactoryv11/types.go`
- Modify: `backend/internal/storage/batch_factory_v11_schema.go`
- Modify: `backend/internal/storage/batch_factory_v11_comic_schema_test.go`

**Interfaces:**
- Produces: `ComicSourceSegment`, `ComicSourceSegmentation`, `ComicSourceSplitInput`, `ComicSourceSegmentStore`.
- Produces migration `1100032` and table `batch_factory_v11_comic_source_segments`, keyed by `video_id`, containing byte offsets, exact text, character count, boundary type, reason, source digest, segmentation version, prompt stale state, media stale state, and revision.

- [ ] **Step 1: Add failing schema assertions** for migration version `1100032`, table ownership/book indexes, video foreign key, source offsets/text/digest, stale flags, and revision.
- [ ] **Step 2: Run** `go test ./internal/storage -run Comic -count=1` and confirm the migration assertions fail.
- [ ] **Step 3: Add the public types and optional store interface** without expanding the legacy base `Store` contract.
- [ ] **Step 4: Add `V11ComicSourceSegmentStatements` and append migration 1100032** with a new checksum marker.
- [ ] **Step 5: Rerun storage tests** and require PASS.

### Task 3: Atomic memory and MySQL persistence

**Files:**
- Create: `backend/internal/batchfactoryv11/comic_source_segments_memory.go`
- Create: `backend/internal/batchfactoryv11/comic_source_segments_mysql.go`
- Create: `backend/internal/batchfactoryv11/comic_source_segments_store_test.go`
- Modify: `backend/internal/batchfactoryv11/memory_store.go`
- Modify: `backend/internal/batchfactoryv11/mysql_store.go`

**Interfaces:**
- Implements: `ReplaceComicSourceSegments(ctx, owner, batchID, bookID, sourceDigest string, drafts []ComicSourceSegmentDraft) (ComicSourceSegmentation, error)`.
- Implements: `GetComicSourceSegmentation(ctx, owner, batchID, bookID string) (ComicSourceSegmentation, error)`.
- Implements: `SplitComicSourceSegment(ctx, owner, batchID, bookID, videoID string, input ComicSourceSplitInput) (ComicSourceSegmentation, error)`.

- [ ] **Step 1: Write failing store contract tests** using `MemoryStore`: replacement creates ordered video-backed pages, retrieval marks changed-source data stale, left split preserves the original ID on the latter half, right split preserves it on the first half, unaffected IDs and prompts remain unchanged, and invalid/version-conflicted splits make no change.
- [ ] **Step 2: Run** `go test ./internal/batchfactoryv11 -run ComicSourceSegmentStore -count=1` and confirm failure.
- [ ] **Step 3: Implement memory persistence** with lock-scoped clone-then-validate-then-swap behavior.
- [ ] **Step 4: Implement MySQL replacement in one transaction**: lock book, validate drafts, orphan old active video rows without deleting their comic media, insert new video/video-record/segment rows, and commit only after continuity validation.
- [ ] **Step 5: Implement transactional left/right split**: lock all active rows, validate the chosen safe boundary and expected revision, shift ordinals, insert one adjacent video, update the stable page, clear affected prompts, mark the retained comic page `prompt_stale`, and validate the reconstructed source before commit.
- [ ] **Step 6: Project segment fields in batch reads** by left joining the new table so `book.videos` exposes source text, character count, boundaries, stale flags, and revision.
- [ ] **Step 7: Run focused store tests and existing comic page tests** and require PASS.

### Task 4: Segmentation service and comic-stage ordering

**Files:**
- Create: `backend/internal/batchfactoryv11/comic_source_service.go`
- Create: `backend/internal/batchfactoryv11/comic_source_service_test.go`
- Modify: `backend/internal/batchfactoryv11/stage_service.go`
- Modify: `backend/internal/httpapi/batch_factory_automation.go`

**Interfaces:**
- Produces: `DirectorService.RunComicSourceSegmentation(ctx, owner, batchID, bookID) (ComicSourceSegmentation, error)`.
- The model receives only numbered safe boundaries and returns boundary IDs; invalid, missing, reordered, provider-failed, or non-covering output falls back to deterministic segments.

- [ ] **Step 1: Write failing provider tests** for valid model-selected boundaries and fallback on malformed JSON, unsafe IDs, missing coverage, and provider failure.
- [ ] **Step 2: Run** `go test ./internal/batchfactoryv11 -run ComicSourceSegmentation -count=1` and confirm failure.
- [ ] **Step 3: Implement the backend-only selection contract** with compact boundary IDs and exact snippets; cap request batching without changing page identity.
- [ ] **Step 4: Validate model choices and persist only complete results**; on any provider/output failure persist the deterministic fallback instead.
- [ ] **Step 5: In comic mode, make director planning call source segmentation** instead of the video storyboard planner. Before comic prompt extraction, ensure a current segmentation exists; segment first if missing or stale.
- [ ] **Step 6: Replace untouched-comic bootstrap automation** with explicit source segmentation, assets, and prompt stages so assets and prompts consume page segments rather than full-book storyboard output.
- [ ] **Step 7: Run stage, automation, segmentation, and existing comic prompt tests** and require PASS.

### Task 5: Read/split HTTP endpoints

**Files:**
- Create: `backend/internal/httpapi/batch_factory_comic_source_segments.go`
- Create: `backend/internal/httpapi/batch_factory_comic_source_segments_test.go`
- Modify: `backend/internal/httpapi/router.go`

**Interfaces:**
- `GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-source-segments`
- `POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-source-segments/segment`
- `POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-source-segments/{videoId}/split`

- [ ] **Step 1: Write failing authenticated route tests** for list, deterministic segmentation through a stubbed service, left/right split, bad boundaries, stale revisions, ownership, and error status mapping.
- [ ] **Step 2: Run** `go test ./internal/httpapi -run ComicSourceSegments -count=1` and confirm failure.
- [ ] **Step 3: Implement thin handlers** that use the authenticated owner, decode `{direction,boundaryOffset,expectedRevision}`, and return the full current segmentation after every mutation.
- [ ] **Step 4: Register routes only when the store implements `ComicSourceSegmentStore`** and keep model work out of manual split handlers.
- [ ] **Step 5: Rerun focused route tests** and require PASS.

### Task 6: Comic page source-content UI

**Files:**
- Create: `frontend/src/user/pages/shuihuo/ComicSourceContent.jsx`
- Create: `frontend/src/user/pages/shuihuo/comicSourceContent.js`
- Create: `frontend/src/user/pages/shuihuo/comicSourceContent.test.js`
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/BatchFactoryWorkbenchPage.css`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

**Interfaces:**
- API helpers: `getComicSourceSegments`, `segmentComicSource`, `splitComicSourceSegment`.
- UI props: `ComicSourceContent({batchId, book, video, ordinal, onChanged})`.

- [ ] **Step 1: Write failing pure tests** for safe-boundary option labels, character counts, stale labels, and left/right split payloads; add source-level assertions that the comic library renders `画面内容` below the image area.
- [ ] **Step 2: Run the focused Node tests** and confirm failure.
- [ ] **Step 3: Add API helpers** using the existing authenticated client.
- [ ] **Step 4: Build the read-only source panel** with `第 N 段 · X 字`, collapsed/full display, safe-boundary selector, icon buttons with tooltips for `向左拆分` and `向右拆分`, conflict/error messages, and a visible stale-media warning.
- [ ] **Step 5: Mount it under the active comic image** and refresh the batch after splitting so page order, prompts, and media state stay synchronized.
- [ ] **Step 6: Add restrained responsive styles** that fit the existing dense workbench and do not resize the comic image frame.
- [ ] **Step 7: Run focused frontend tests** and require PASS.

### Task 7: End-to-end regression and embedded runtime acceptance

**Files:**
- Modify only files required by failures caused by Tasks 1-6.

- [ ] **Step 1: Run focused Go suites:** `go test ./internal/batchfactoryv11 ./internal/httpapi ./internal/storage -count=1`.
- [ ] **Step 2: Run focused frontend tests** for comic source content, API paths, and Batch Factory source rendering.
- [ ] **Step 3: Run** `git diff --check` and inspect only this feature's diff in the dirty worktree.
- [ ] **Step 4: Build frontend:** `npm --prefix frontend run build`.
- [ ] **Step 5: Rebuild embedded Go runtime:** `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`.
- [ ] **Step 6: Confirm the process listening on 4001 is `backend/qiantie-local.exe`, stop only that process, then restart with** `scripts/start-local-go-review.ps1`.
- [ ] **Step 7: Verify runtime health and identity with** `scripts/check-local-go.ps1` and `/api/runtime-build-info`.
- [ ] **Step 8: Browser-check authenticated `http://localhost:4001/batch-factory`** in comic mode: source segments are page-sized, exact text is shown below the image, navigation follows segments, both split directions work on a test-safe boundary, stale warnings appear, and no generation request is submitted.
