# Batch Factory Text Page Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Batch Factory text rows show synchronized assigned copy and rendered page images, restore the text book engine/layout/publish configuration structure, and keep the work cover out of text carousel uploads.

**Architecture:** Add one focused inline text-page component that owns the shared page index for the two adjacent table cells and reuses the existing authenticated image renderer. Keep detailed editing in `TextMediaDialog`. Preserve cover artifacts as separate document metadata, but build collection publication images only from collection page items.

**Tech Stack:** React 18, Ant Design, Node test runner, Go, MySQL-backed Batch Factory repositories, Go `embed` local runtime.

## Global Constraints

- Work only in the existing `v88` branch; do not create another branch.
- Preserve unrelated dirty files and stage only files from this plan.
- Update and verify local port 4001 only; do not push or deploy until the user asks again.
- `页面内容` shows assigned source text; `文字页` shows the rendered active image for the same page.
- The cover remains separately editable and previewable, but never appears in the text collection image list or N8 carousel image array.
- Text book settings expose exactly `引擎配置 / 排版配置 / 发布配置` and retain field-level inheritance.

---

### Task 1: Synchronized inline text copy and rendered page image

**Files:**
- Create: `frontend/src/user/pages/shuihuo/InlineTextBookPages.jsx`
- Create: `frontend/src/user/pages/shuihuo/InlineTextBookPages.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo-production.css`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryTextContentEditor.source.test.js`

**Interfaces:**
- Consumes: `listTextPages(batchId, bookId, options)`, page fields `id`, `ordinal`, `text`, `activeArtifactId`, and artifact fields `id`, `contentUrl`.
- Produces: `InlineTextBookPages({batchId, book, pageTarget, automationActive, onOpen})`, where `onOpen(pageId)` opens the existing text media dialog at the same page.

- [ ] **Step 1: Write the failing component source test**

Assert that the new component reads pages, finds the active artifact, renders `AuthenticatedTextPageImage`, renders both `页面内容` and `文字页`, uses one `selectedIndex`, exposes previous/next controls, and calls `onOpen(page.id)`.

- [ ] **Step 2: Run the test and verify it fails**

Run: `node --test src/user/pages/shuihuo/InlineTextBookPages.source.test.js`

Expected: FAIL because `InlineTextBookPages.jsx` does not exist.

- [ ] **Step 3: Implement the shared inline component**

Create a component that:

```jsx
const page = pages[selectedIndex];
const activeArtifact = artifacts.find(item => item.id === page?.activeArtifactId && item.kind === 'page');
return <>
  <AssignedTextCell page={page} selectedIndex={selectedIndex} total={pages.length} onMove={setSelectedIndex} />
  <RenderedTextPageCell page={page} artifact={activeArtifact} selectedIndex={selectedIndex} total={pages.length} onMove={setSelectedIndex} onOpen={() => onOpen?.(page?.id)} />
</>;
```

Poll every two seconds only while `automationActive`. Abort reads on unmount, clamp the selected index after page-count changes, and show explicit loading, read-failure, and ungenerated-page states.

- [ ] **Step 4: Replace the two text-only row branches**

In `BatchFactoryNovelList.jsx`, remove `DeferredInlineTextPages` and the text-only `文本媒体库` button branch. Render one `InlineTextBookPages` in the text branch so it supplies both adjacent cells, and pass:

```jsx
onOpen={pageId => openTextBook(book, 'pages', pageId)}
```

Extend `openTextBook` to accept and store the initial text page ID without changing comic/video behavior.

- [ ] **Step 5: Add compact image-cell styling**

Add styles for a 3:4-aware image stage, empty state, page label, shared pager geometry, and a visible `查看 / 调整` action. Do not embed the full media-library tool panel in the row.

- [ ] **Step 6: Run focused and surrounding frontend tests**

Run:

```powershell
node --test src/user/pages/shuihuo/InlineTextBookPages.source.test.js src/user/pages/shuihuo/BatchFactoryTextContentEditor.source.test.js src/user/pages/shuihuo/TextMediaPanel.source.test.js src/user/pages/shuihuo/textMediaState.test.js
```

Expected: all tests PASS.

- [ ] **Step 7: Commit the inline workbench change**

```powershell
git add frontend/src/user/pages/shuihuo/InlineTextBookPages.jsx frontend/src/user/pages/shuihuo/InlineTextBookPages.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/shuihuo/BatchFactoryTextContentEditor.source.test.js
git commit -m "feat(batch-factory): align inline text pages"
```

### Task 2: Text single-book configuration structure

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/textProductionConfig.js`
- Modify: `frontend/src/user/pages/shuihuo/TextProductionSettingsForm.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Create: `frontend/src/user/pages/shuihuo/TextProductionSettingsForm.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/textProductionConfig.test.js`

**Interfaces:**
- Consumes: `TEXT_PRODUCTION_SECTIONS`, `textProductionBookOverride`, and `updateTextProductionSettings`.
- Produces: the same dotted settings keys and save payloads, with user-facing tabs labelled `引擎配置`, `排版配置`, and `发布配置`.

- [ ] **Step 1: Write failing structure and inheritance tests**

Add assertions that text single-book mode renders all three required labels, keeps the inherited/override marker, and serializes changes from each section without dropping sibling nested settings.

- [ ] **Step 2: Run the focused tests and verify the label test fails**

Run:

```powershell
node --test src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js src/user/pages/shuihuo/textProductionConfig.test.js
```

Expected: the `引擎配置` assertion FAILS because the first text section is still labelled `模型配置`.

- [ ] **Step 3: Align the text section labels and modal copy**

Change the first text section label to `引擎配置`. Keep its existing text model, sensitive handling, AI repair, failure strategy, and keyword highlight fields. Keep the other two labels as `排版配置` and `发布配置`. Update the text modal title/footer copy so it describes the same three-section single-book configuration rather than a separate simplified form.

- [ ] **Step 4: Verify all text configuration tests**

Run the two tests from Step 2 plus:

```powershell
node --test src/user/pages/shuihuo/TextProductionSettingsForm.source.test.js
```

Expected: all tests PASS and dotted nested settings remain intact.

- [ ] **Step 5: Commit the configuration alignment**

```powershell
git add frontend/src/user/pages/shuihuo/textProductionConfig.js frontend/src/user/pages/shuihuo/TextProductionSettingsForm.jsx frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js frontend/src/user/pages/shuihuo/textProductionConfig.test.js
git commit -m "fix(batch-factory): restore text book settings sections"
```

### Task 3: Exclude work covers from text collections and publication images

**Files:**
- Modify: `backend/internal/batchfactorytext/workflow_artifacts.go`
- Modify: `backend/internal/batchfactorytext/workflow.go`
- Modify: `backend/internal/batchfactorytext/workflow_test.go`
- Modify: `backend/internal/httpapi/batch_factory_text_publish.go`
- Modify: `backend/internal/httpapi/batch_factory_text_publish_test.go`
- Modify: `frontend/src/user/pages/shuihuo/TextMediaPanel.jsx`
- Modify: `frontend/src/user/pages/shuihuo/TextMediaPanel.source.test.js`

**Interfaces:**
- Consumes: `Collection.Items` as the ordered frozen text-page manifest and `Collection.CoverArtifactID` as optional external-cover provenance.
- Produces: N8 `PublishRequest.Images` containing only ordered collection page images.

- [ ] **Step 1: Write failing backend publication tests**

Change the existing publication fixture to include three page artifacts plus one cover. Assert:

```go
if len(client.images) != 3 {
    t.Fatalf("published images=%d, want the three text pages only", len(client.images))
}
```

Also assert the first uploaded bytes are page 1, not the cover, and that a ready page-only collection can be composed without treating the cover as an item.

- [ ] **Step 2: Run the backend tests and verify they fail**

Run:

```powershell
go test ./internal/httpapi -run TestTextPublishSignedReadbackAndForcedPending -count=1
go test ./internal/batchfactorytext -run "Test.*Collection|Test.*Cover" -count=1
```

Expected: the HTTP publication test FAILS because the current code prepends the cover to `out.Pages`.

- [ ] **Step 3: Build publication pages only from collection items**

In `batch_factory_text_publish.go`, stop calling `publishPage` for `c.CoverArtifactID`. Iterate only over `c.Items`, preserving their order and ordinal values. Keep the cover reference available as metadata/provenance, but do not append it to `cp.Collection.Pages`.

- [ ] **Step 4: Remove cover readiness from collection-only composition**

Update collection matching and manifest verification so page readiness controls collection creation. Cover validation remains in the separate cover workflow and cover editor, but a missing cover does not turn a complete page manifest into an unready text collection. Keep `CoverArtifactID` optional in stored records for backward compatibility.

- [ ] **Step 5: Remove the cover choice from collection preview**

In `TextMediaPanel.jsx`, the collections tab must list only `文字页 N` buttons and preview the first collection item by default. The separate `封面` tab remains unchanged.

- [ ] **Step 6: Run backend and frontend collection tests**

Run:

```powershell
go test ./internal/batchfactorytext ./internal/httpapi -run "Text|Collection|Cover" -count=1
node --test src/user/pages/shuihuo/TextMediaPanel.source.test.js src/user/pages/shuihuo/textMediaState.test.js
```

Expected: all tests PASS; a three-page collection publishes three images.

- [ ] **Step 7: Commit the cover boundary fix**

```powershell
git add backend/internal/batchfactorytext/workflow_artifacts.go backend/internal/batchfactorytext/workflow.go backend/internal/batchfactorytext/workflow_test.go backend/internal/httpapi/batch_factory_text_publish.go backend/internal/httpapi/batch_factory_text_publish_test.go frontend/src/user/pages/shuihuo/TextMediaPanel.jsx frontend/src/user/pages/shuihuo/TextMediaPanel.source.test.js
git commit -m "fix(batch-factory): keep text covers outside collections"
```

### Task 4: Build, restart local 4001, and perform real-book acceptance

**Files:**
- Verify only; no planned source changes.

**Interfaces:**
- Consumes: local `qiantie-local.exe`, `/health`, `/api/runtime-build-info`, and the existing text acceptance batch.
- Produces: browser evidence for the approved UI and local runtime SHA.

- [ ] **Step 1: Run the full focused suites**

```powershell
go test ./internal/batchfactorytext ./internal/httpapi -count=1
node --test src/user/pages/shuihuo/InlineTextBookPages.source.test.js src/user/pages/shuihuo/BatchFactoryTextContentEditor.source.test.js src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js src/user/pages/shuihuo/textProductionConfig.test.js src/user/pages/shuihuo/TextMediaPanel.source.test.js src/user/pages/shuihuo/textMediaState.test.js
npm run build
```

Expected: all tests and the production frontend build PASS.

- [ ] **Step 2: Restart only the confirmed local port 4001 runtime**

Resolve the listener, verify its executable path is `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`, stop that PID, and run:

```powershell
.\scripts\start-local-go-review.ps1 -Detached
```

- [ ] **Step 3: Verify runtime identity**

Check `/health` and `/api/runtime-build-info`; the runtime and manifest SHA must equal the final local commit.

- [ ] **Step 4: Verify the existing three-page text acceptance book**

Use a temporary local acceptance session, then confirm:

1. page 1 assigned text and page 1 rendered image appear side by side;
2. one next-page action changes both cells to page 2;
3. clicking the image opens the media library on page 2;
4. single-book settings show the three approved labels;
5. the collection preview lists only three text pages and no cover entry;
6. saving text production mode returns HTTP 200 without `生产模式无效`.

- [ ] **Step 5: Clean acceptance state**

Close the browser session and delete only the temporary local authentication session created for acceptance. Do not remove user batches, artifacts, or unrelated worktree files.

- [ ] **Step 6: Report local-only completion**

Report local commits, test counts, runtime SHA, real-book evidence, and explicitly state that GitHub and production were not updated.
