# Batch Factory Text Automation Publish And Acceptance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Run text books through source, sensitive processing, pagination, rendering, validation, collection, upload, recovery, and real local acceptance without mixing video or comic state.

**Architecture:** Add a text workflow adapter to the existing automation coordinator and stage contract. The coordinator asks the text service to reach a target state; text-specific checkpoints live in MySQL, while the existing publisher service receives a typed image collection with a strict `text` kind. UI status is projected from backend facts.

**Tech Stack:** Go 1.26, MySQL, Redis queue/locks, TOS, React 18, Node tests, Go embed

## Global Constraints

- The visible text chain is `等待处理 → 文本处理 → 正在分页 → 正在渲染 → 正在校验 → 合集就绪 → 正在上传 → 已上传`.
- One failed book does not stop other books in the batch.
- Retry resumes from the failed stage and never overwrites a locked uploaded page.
- Text and comic may use the same publisher service but cannot be mixed in one request.
- Already published text uses `再次上传` and requires confirmation.
- “已上传” requires the existing success receipt/readback rule, not only request acceptance.
- Local implementation and acceptance come before any remote push or online deployment.
- Preserve unrelated dirty worktree changes and commit only files named by each task.

---

### Task 1: Extend the shared automation stage contract

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/automationStages.json`
- Modify: `backend/internal/httpapi/batch_factory_automation_stages.go`
- Modify: `backend/internal/httpapi/batch_automation_design_test.go`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryBookState.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryBookState.test.js`

**Interfaces:**
- Produces text runtime stages: `text_process`, `text_paginate`, `text_render`, `text_validate`, `text_collection`, `text_publish`.
- Produces text book timeline and terminal state from backend stage summaries.

- [ ] **Step 1: Add failing contract and timeline tests**

Assert the JSON and Go contracts contain identical text stage IDs/labels, and the frontend timeline maps each running/failed/completed stage to the approved Chinese copy.

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/httpapi -run AutomationStagesContract -count=1` from `backend`
Run: `node --test src/user/pages/shuihuo/batchFactoryBookState.test.js` from `frontend`
Expected: FAIL because text stages are absent.

- [ ] **Step 3: Add stage definitions and state projection**

Keep video and comic definitions unchanged. Text completion requires a verified publish record when auto upload is enabled; collection-ready is terminal only for a non-submit run mode.

- [ ] **Step 4: Run and commit**

Run both focused tests; expect PASS.

```bash
git add frontend/src/user/pages/shuihuo/automationStages.json backend/internal/httpapi/batch_factory_automation_stages.go backend/internal/httpapi/batch_automation_design_test.go frontend/src/user/pages/shuihuo/batchFactoryBookState.js frontend/src/user/pages/shuihuo/batchFactoryBookState.test.js
git commit -m "feat: add text automation stages"
```

### Task 2: Add the text workflow service and idempotent checkpoints

**Files:**
- Create: `backend/internal/batchfactorytext/workflow.go`
- Create: `backend/internal/batchfactorytext/workflow_test.go`
- Modify: `backend/internal/httpapi/batch_factory_automation.go`
- Modify: `backend/internal/httpapi/batch_factory_automation_preflight.go`
- Modify: `backend/internal/httpapi/batch_factory_automation_test.go`

**Interfaces:**
- Produces: `Workflow.Ensure(ctx context.Context, req EnsureRequest) (EnsureResult, error)`.
- `EnsureRequest` includes owner, batch/book IDs, frozen settings, target `collection|published`, and force flags.
- Consumes: sensitive processor, repository, paginator, renderer, artifact store, and collection service from plans 01/02.

- [ ] **Step 1: Write failing workflow tests**

Use fakes to prove the order `process → paginate → render pages and cover → validate → collection`, successful checkpoints are skipped on retry, render retries only failed pages, locked pages/cover are skipped, and duplicate `Ensure` calls reuse the same content/config fingerprint.

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/batchfactorytext -run Workflow -count=1` from `backend`
Expected: FAIL because `Workflow` does not exist.

- [ ] **Step 3: Implement stage fingerprinting**

Use SHA-256 over production text digest plus the settings that affect each stage. Background/color/cover changes invalidate render and collection, layout/text changes invalidate pagination onward, and publish mapping changes invalidate only publish.

- [ ] **Step 4: Connect the coordinator through a narrow adapter**

Add a `textEnsure` function field to `BatchAutomationCoordinator`, matching the existing comic dependency-injection pattern. In preflight, text requires source and valid render settings; it does not require image or video models. AI sensitive mode additionally requires a text model.

- [ ] **Step 5: Verify batch isolation and restart recovery**

Persist each stage with `BookStageService`; create tests where one book fails rendering while the next completes, and where a new coordinator instance resumes from stored stages.

- [ ] **Step 6: Run and commit**

Run: `go test ./internal/batchfactorytext ./internal/httpapi -run 'TextWorkflow|Automation.*Text' -count=1` from `backend`
Expected: PASS.

```bash
git add backend/internal/batchfactorytext/workflow.go backend/internal/batchfactorytext/workflow_test.go backend/internal/httpapi/batch_factory_automation.go backend/internal/httpapi/batch_factory_automation_preflight.go backend/internal/httpapi/batch_factory_automation_test.go
git commit -m "feat: automate text production workflow"
```

### Task 3: Generalize image-collection publishing with strict type guards

**Files:**
- Create: `backend/internal/collectionpublish/types.go`
- Create: `backend/internal/collectionpublish/service.go`
- Create: `backend/internal/collectionpublish/service_test.go`
- Modify: `backend/internal/httpapi/batch_factory_n8_comic_publish.go`
- Create: `backend/internal/httpapi/batch_factory_text_publish.go`
- Create: `backend/internal/httpapi/batch_factory_text_publish_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `frontend/src/shared/api/batchFactoryV12.js`

**Interfaces:**
- Produces: `CollectionKindComic` and `CollectionKindText`.
- Produces: `PublishRequest{Kind, Owner, BatchID, BookID, CollectionID, Producer, OrganizationID, CategoryID, Force}`.
- Produces text publish and status endpoints with idempotent request keys.

- [ ] **Step 1: Add failing service tests**

Test comic-only, text-only, mixed-kind rejection, first publish idempotency, forced republish creates a new external intent, and accepted-but-unconfirmed responses remain pending.

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/collectionpublish ./internal/httpapi -run 'CollectionPublish|TextPublish' -count=1` from `backend`
Expected: FAIL because the generic service is absent.

- [ ] **Step 3: Extract only the reusable publisher layer**

Keep comic HTTP routes and comic domain types intact. Move account/organization/category validation, upload ordering, idempotency, and readback into `collectionpublish`; adapt comic and text collections into the shared request.

- [ ] **Step 4: Add explicit mixed-kind validation**

Both service and HTTP layer reject a request containing collection IDs from different kinds with `漫画与文本不能一起上传，请分开选择。`.

- [ ] **Step 5: Add text frontend API methods**

Expose `publishTextCollection` and `getTextPublishStatus`. A normal repeat returns the existing record; `force:true` creates a confirmed republish.

- [ ] **Step 6: Run and commit**

Run: `go test ./internal/collectionpublish ./internal/httpapi -run 'CollectionPublish|TextPublish|N8ComicPublish' -count=1` from `backend`
Expected: PASS with existing comic tests unchanged.

```bash
git add backend/internal/collectionpublish backend/internal/httpapi/batch_factory_n8_comic_publish.go backend/internal/httpapi/batch_factory_text_publish.go backend/internal/httpapi/batch_factory_text_publish_test.go backend/internal/httpapi/router.go frontend/src/shared/api/batchFactoryV12.js
git commit -m "feat: publish text image collections"
```

### Task 4: Wire text logs, retry, automatic collection, and upload controls

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/TextMediaPanel.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryNovelListHelpers.js`
- Create: `frontend/src/user/pages/shuihuo/batchFactoryTextAutomation.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryImmediatePreflight.source.test.js`

**Interfaces:**
- Consumes: text stage summaries and publish status.
- Produces: text preflight messages, text logs, page/collection/upload retry actions, and correct `上传网络/再次上传` labels.

- [ ] **Step 1: Add failing UI behavior tests**

Assert text automation does not require image/video models, requires a text model only for AI optimization, shows the six text stages, disables upload before collection, and confirms republish.

- [ ] **Step 2: Run and verify failure**

Run: `node --test src/user/pages/shuihuo/batchFactoryTextAutomation.test.js src/user/pages/shuihuo/BatchFactoryImmediatePreflight.source.test.js` from `frontend`
Expected: FAIL before text behavior exists.

- [ ] **Step 3: Add text preflight and retry routing**

Retry a page render through the page endpoint, collection through collection creation, and upload through publish. Do not send a text retry to comic image generation or video production.

- [ ] **Step 4: Add visible task/log copy**

Display the failed book, stage, page number when present, and backend action message. Keep earlier successful stage results visible.

- [ ] **Step 5: Run and commit**

Run the focused tests and `npm run build` from `frontend`; expect PASS.

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/TextMediaPanel.jsx frontend/src/user/pages/shuihuo/batchFactoryNovelListHelpers.js frontend/src/user/pages/shuihuo/batchFactoryTextAutomation.test.js frontend/src/user/pages/shuihuo/BatchFactoryImmediatePreflight.source.test.js
git commit -m "feat: expose text automation and retry controls"
```

### Task 5: Run full local verification and embedded-runtime acceptance

**Files:**
- Create: `docs/acceptance/2026-11-23-batch-factory-text-generation.md`
- Modify only if failures require fixes: files already named in plans 01-04

**Interfaces:**
- Verifies source, build, embedded binary, local HTTP runtime, visible UI, generated artifacts, collections, and upload state separately.

- [ ] **Step 1: Run backend package tests**

Run from `backend`:

```powershell
go test ./internal/batchfactorysensitive ./internal/batchfactorytext ./internal/collectionpublish ./internal/storage ./internal/httpapi -count=1
```

Expected: PASS.

- [ ] **Step 2: Run focused and full frontend tests**

Run from `frontend`:

```powershell
node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js src/user/pages/shuihuo/batchFactorySensitiveMode.test.js src/user/pages/shuihuo/textProductionConfig.test.js src/user/pages/shuihuo/textMediaState.test.js src/user/pages/shuihuo/TextMediaPanel.source.test.js src/user/pages/shuihuo/batchFactoryTextAutomation.test.js
npm test
npm run build
```

Expected: PASS.

- [ ] **Step 3: Rebuild the embedded Go binary**

Use the repository's confirmed local build/start script, then check:

```text
GET http://localhost:4001/health
GET http://localhost:4001/api/runtime-build-info
```

Expected: health is successful and runtime build identity matches the new local commit.

- [ ] **Step 4: Perform visible local acceptance with one real book**

Create a text batch and verify default 150/4000/3:4, system optimization, whole-book solid background, sentence-safe pagination, page edit/reflow, page-only background rerender, uploaded-page lock, collection, and upload mapping. Capture exact book ID, page count, collection ID, publish record, and screenshots in the acceptance document.

- [ ] **Step 5: Exercise failure recovery**

Force one render failure, restart the local runtime, retry, and verify completed pages and locked uploads remain unchanged. Verify another book in the same batch continues.

- [ ] **Step 6: Verify mixed upload rejection**

Select one comic collection and one text collection; confirm frontend blocks it and a direct signed backend request returns the same Chinese error.

- [ ] **Step 7: Commit acceptance evidence**

```bash
git add docs/acceptance/2026-11-23-batch-factory-text-generation.md
git commit -m "docs: record text generation acceptance"
```

- [ ] **Step 8: Stop before remote actions**

Report local commit IDs, changed paths, tests, build identity, book/page/collection/upload evidence, and remaining risks. Do not push GitHub or deploy online until the user explicitly authorizes those actions after reviewing local acceptance.
