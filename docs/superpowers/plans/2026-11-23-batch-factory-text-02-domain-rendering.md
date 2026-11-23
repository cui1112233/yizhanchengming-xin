# Batch Factory Text Domain Pagination And Rendering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist text-production versions and pages, paginate without cutting sentences, render canonical 3:4 images on the Go server, and build durable text collections.

**Architecture:** Add a focused `batchfactorytext` domain package with repository interfaces, a deterministic paginator, and a canonical renderer. MySQL stores document/page/version facts, local artifact storage and TOS store bytes, and HTTP handlers expose previews and edits without reusing comic business state.

**Tech Stack:** Go 1.26, MySQL, Redis-compatible job orchestration, TOS, freetype, golang.org/x/image

## Global Constraints

- Default page target is 150 characters and default content range is 4000 characters.
- Default canvas is 3:4 and default maximum text height is 75%.
- A sentence is never split in the middle; a long paragraph may continue only after a complete sentence.
- Preview and final images use the same Go renderer.
- Uploaded finished pages are locked and automation cannot overwrite them.
- Complete original text and editable production text are different durable versions.
- MySQL is authoritative; Redis is not used to persist page or collection truth.
- Preserve unrelated dirty worktree changes and commit only files named by each task.

---

### Task 1: Add V12 text-production tables

**Files:**
- Modify: `backend/internal/storage/batch_factory_v12_schema.go`
- Modify: `backend/internal/storage/batch_factory_v12_schema_test.go`

**Interfaces:**
- Produces migration `1200113` with text documents, versions, pages, page artifacts, backgrounds, collections, and collection items.

- [ ] **Step 1: Write the failing schema assertions**

Assert that `V12TextProductionStatements()` creates these tables:

```text
batch_factory_v12_text_documents
batch_factory_v12_text_versions
batch_factory_v12_text_pages
batch_factory_v12_text_page_artifacts
batch_factory_v12_text_backgrounds
batch_factory_v12_text_collections
batch_factory_v12_text_collection_items
```

Also assert migration number `1200113` and checksum `batch-factory-v12-text-production-v1`.

- [ ] **Step 2: Run the schema test and verify failure**

Run: `go test ./internal/storage -run V12TextProduction -count=1` from `backend`
Expected: FAIL because the migration is absent.

- [ ] **Step 3: Add append-only DDL**

The page table must include `ordinal_no`, `start_offset`, `end_offset`, `page_text`, `layout_json`, `background_json`, `source_kind`, `locked`, `status`, `last_error`, `revision`, and timestamps. Add unique `(owner_username, book_id, pagination_version, ordinal_no)` and book/status indexes.

The document table stores `original_text`, `active_version_id`, `pagination_version`, and `config_snapshot_json`; the versions table stores immutable `production_text`, sensitive mode, change trace, and content digest.

- [ ] **Step 4: Run storage tests**

Run: `go test ./internal/storage -run 'V12(TextProduction|Migrations)' -count=1` from `backend`
Expected: PASS.

- [ ] **Step 5: Commit the migration**

```bash
git add backend/internal/storage/batch_factory_v12_schema.go backend/internal/storage/batch_factory_v12_schema_test.go
git commit -m "feat: add text production schema"
```

### Task 2: Define the text domain and MySQL repository

**Files:**
- Create: `backend/internal/batchfactorytext/types.go`
- Create: `backend/internal/batchfactorytext/repository.go`
- Create: `backend/internal/batchfactorytext/mysql_repository.go`
- Create: `backend/internal/batchfactorytext/mysql_repository_test.go`

**Interfaces:**
- Produces: `LayoutConfig`, `Background`, `Document`, `Version`, `Page`, `PageArtifact`, `Collection`.
- Produces: `Repository` methods `SaveOriginal`, `CreateVersion`, `ReplaceSystemPages`, `PatchPage`, `LockPageArtifact`, `RestoreSystemPage`, `CreateCollection`, and list methods.

- [ ] **Step 1: Write repository contract tests with sqlmock**

```go
type LayoutConfig struct {
    AspectRatio string `json:"aspectRatio"`
    TargetCharacters int `json:"targetCharacters"`
    FontID string `json:"fontId"`
    FontSize float64 `json:"fontSize"`
    FontWeight int `json:"fontWeight"`
    LineHeight float64 `json:"lineHeight"`
    ParagraphSpacing float64 `json:"paragraphSpacing"`
    Padding int `json:"padding"`
    Align string `json:"align"`
    TextColor string `json:"textColor"`
    HighlightColor string `json:"highlightColor"`
    MaxHeightRatio float64 `json:"maxHeightRatio"`
}
```

Test that `ReplaceSystemPages` updates system pages in one transaction, preserves `locked=TRUE` rows, and increments `pagination_version` only after all inserts succeed.

- [ ] **Step 2: Run the repository test and verify failure**

Run: `go test ./internal/batchfactorytext -run Repository -count=1` from `backend`
Expected: FAIL because the package is absent.

- [ ] **Step 3: Implement typed repository methods**

Use optimistic revisions for page edits. Return a typed conflict error when `expectedRevision` is stale. Never update `original_text` from page edits.

- [ ] **Step 4: Verify transaction and lock behavior**

Run: `go test ./internal/batchfactorytext -run 'Repository|Locked' -count=1` from `backend`
Expected: PASS.

- [ ] **Step 5: Commit the domain store**

```bash
git add backend/internal/batchfactorytext/types.go backend/internal/batchfactorytext/repository.go backend/internal/batchfactorytext/mysql_repository.go backend/internal/batchfactorytext/mysql_repository_test.go
git commit -m "feat: persist text production pages"
```

### Task 3: Implement complete-sentence pagination

**Files:**
- Create: `backend/internal/batchfactorytext/pagination.go`
- Create: `backend/internal/batchfactorytext/pagination_test.go`

**Interfaces:**
- Produces: `Paginate(ctx context.Context, input PaginationInput, measure MeasureFunc) ([]PageDraft, error)`.
- `MeasureFunc(text string, layout LayoutConfig) (fits bool, usedHeight float64, err error)` is supplied by the renderer.

- [ ] **Step 1: Add failing table tests**

Cover short paragraphs, dialogue punctuation, Chinese/English terminators, a paragraph longer than one page, blank lines, and emoji. Every case must assert:

```go
joined := strings.Join(pageTexts(pages), "")
if joined != normalizedInput { t.Fatalf("content mismatch") }
if strings.Contains(joined, "长段落续页") { t.Fatal("internal marker leaked") }
```

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/batchfactorytext -run Paginate -count=1` from `backend`
Expected: FAIL because `Paginate` is undefined.

- [ ] **Step 3: Implement paragraph-first, sentence-second packing**

Split paragraphs without dropping separators. Try the whole next paragraph; if it does not fit an empty page, split it using sentence boundaries `。！？!?；;…` plus closing quotes. Use rune offsets, not byte offsets.

- [ ] **Step 4: Add property-style invariants**

Generate deterministic combinations from a fixed seed and assert content equality, monotonic offsets, non-empty pages, and no page exceeding `MeasureFunc`.

- [ ] **Step 5: Run paginator tests**

Run: `go test ./internal/batchfactorytext -run Paginate -count=1` from `backend`
Expected: PASS.

- [ ] **Step 6: Commit pagination**

```bash
git add backend/internal/batchfactorytext/pagination.go backend/internal/batchfactorytext/pagination_test.go
git commit -m "feat: paginate text on sentence boundaries"
```

### Task 4: Build the canonical server renderer

**Files:**
- Create: `backend/internal/renderfont/catalog.go`
- Move: `backend/internal/watermark/NotoSansSC.ttf` to `backend/internal/renderfont/fonts/NotoSansSC.ttf`
- Copy: `frontend/public/assets/fonts/home-xianghai.ttf` to `backend/internal/renderfont/fonts/XiangHai.ttf`
- Modify: `backend/internal/watermark/watermark.go`
- Create: `backend/internal/batchfactorytext/render.go`
- Create: `backend/internal/batchfactorytext/render_test.go`
- Modify: `backend/internal/watermark/watermark_test.go`

**Interfaces:**
- Produces: `renderfont.Open(id string, size float64) (font.Face, error)` and `renderfont.Catalog() []FontInfo`.
- Produces: `Renderer.Measure(text string, layout LayoutConfig) (MeasureResult, error)`.
- Produces: `Renderer.Render(ctx context.Context, input RenderInput) ([]byte, string, error)` returning PNG bytes and `image/png`.
- Produces: `HighlightResolver.Ranges(ctx context.Context, owner, bookID, versionID, text string) ([]HighlightRange, error)`; no reusable data returns an empty slice, not an error.

- [ ] **Step 1: Add failing font and render tests**

Assert both bundled fonts load, unknown fonts return a visible error, 3:4 renders at 1080×1440, solid background pixels match the configured color, and the measured content never exceeds 75% height by default.

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/renderfont ./internal/batchfactorytext ./internal/watermark -run 'Font|Render|Watermark' -count=1` from `backend`
Expected: FAIL before the shared font catalog exists.

- [ ] **Step 3: Centralize the embedded font bytes**

Embed each font once in `renderfont`. Update watermark to request `noto-sans-sc` rather than embedding its own copy. Keep existing watermark output tests passing.

- [ ] **Step 4: Implement measurement and rendering with one layout engine**

Decode solid color or background image, calculate the content rectangle, wrap by rune using the selected face, apply line/paragraph spacing and alignment, and draw highlight runs using the existing highlight ranges. `Measure` and `Render` must share the same line-layout function.

When `highlightReuseEnabled` is false, skip highlight lookup. When it is true, map the existing illustration highlight records to rune ranges; missing records render normal text and do not fail the page.

- [ ] **Step 5: Add golden invariants without committing fragile screenshots**

Decode the PNG and assert dimensions, background sample pixels, non-background text pixels, and stable SHA-256 for one fixed Noto test card. Keep the fixture text and layout in the test.

- [ ] **Step 6: Run rendering tests**

Run: `go test ./internal/renderfont ./internal/batchfactorytext ./internal/watermark -count=1` from `backend`
Expected: PASS.

- [ ] **Step 7: Commit the renderer**

```bash
git add backend/internal/renderfont backend/internal/batchfactorytext/render.go backend/internal/batchfactorytext/render_test.go backend/internal/watermark frontend/public/assets/fonts/home-xianghai.ttf
git commit -m "feat: render canonical text pages on server"
```

### Task 5: Add text page, background, preview, upload, and collection APIs

**Files:**
- Create: `backend/internal/httpapi/batch_factory_text_pages.go`
- Create: `backend/internal/httpapi/batch_factory_text_pages_test.go`
- Create: `backend/internal/httpapi/batch_factory_text_media.go`
- Create: `backend/internal/httpapi/batch_factory_text_media_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `frontend/src/shared/api/batchFactoryV12.js`
- Create: `frontend/src/shared/api/batchFactoryText.test.js`

**Interfaces:**
- Produces endpoints under `/api/batch-factory/v12/batches/{batchId}/books/{bookId}/text-*`.
- Produces frontend functions `prepareTextBook`, `listTextPages`, `patchTextPage`, `splitTextPage`, `mergeTextPage`, `previewTextPage`, `renderTextPage`, `uploadTextPageArtifact`, `restoreTextPage`, `previewTextCover`, `renderTextCover`, `uploadTextCover`, `restoreTextCover`, `listTextBackgrounds`, `uploadTextBackground`, `createTextCollection`, `listTextCollections`.

- [ ] **Step 1: Write failing signed HTTP tests**

Test owner isolation, text-mode guard, optimistic revision conflict, complete-sentence split/merge, locked-page preservation, valid image upload, invalid image rejection, pure-color preview, background-library upload, generated title cover, uploaded cover lock/restore, and collection refusal when the cover or any selected page lacks a ready artifact.

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/httpapi -run BatchFactoryText -count=1` from `backend`
Expected: FAIL because routes are not registered.

- [ ] **Step 3: Register the typed handlers**

Handlers validate `EffectiveProductionMode(...) == "text"`, call the text service, store generated bytes through the existing local artifact/TOS handoff, and return authenticated media URLs rather than file-system paths.

- [ ] **Step 4: Add frontend API wrappers and request-shape tests**

Run: `node --test src/shared/api/batchFactoryText.test.js` from `frontend`
Expected: PASS with encoded batch/book/page identifiers and correct methods.

- [ ] **Step 5: Run backend and frontend API tests**

Run: `go test ./internal/httpapi -run BatchFactoryText -count=1` from `backend`
Run: `node --test src/shared/api/batchFactoryText.test.js` from `frontend`
Expected: PASS.

- [ ] **Step 6: Commit the API slice**

```bash
git add backend/internal/httpapi/batch_factory_text_pages.go backend/internal/httpapi/batch_factory_text_pages_test.go backend/internal/httpapi/batch_factory_text_media.go backend/internal/httpapi/batch_factory_text_media_test.go backend/internal/httpapi/router.go frontend/src/shared/api/batchFactoryV12.js frontend/src/shared/api/batchFactoryText.test.js
git commit -m "feat: expose text page production APIs"
```
