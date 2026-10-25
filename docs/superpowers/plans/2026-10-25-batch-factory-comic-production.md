# Batch Factory Comic Production Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有批量工厂完成视频/漫画互斥生产、漫画资产、漫画页候选图/主图和漫画合集。

**Architecture:** 书级 `productionMode` 是唯一模式事实来源，前端视频/漫画开关只作互斥显示。漫画资产继续使用 `BookAsset` 并以 `source=comic` 隔离；漫画页和漫画合集使用独立 MySQL 记录，不复用视频任务或资产图片记录。

**Tech Stack:** Go、MySQL、Redis、React、Ant Design、Go embed、Node test、Go test。

## Global Constraints

- 只在 `v88` 开发，不覆盖已有无关改动。
- 迁移只追加在 `1100027` 之后，不能修改已执行迁移。
- 系统提示词正文只由后端 `prompt` 模块保存和冻结；浏览器不得硬编码。
- 漫画模式不创建视频任务、不调用视频合成、FFmpeg、配音或上传后台。
- 漫画合集只保存分镜主图的有序快照，不拼图、不上传。
- 每项先写失败测试，再做最小实现、验证通过并独立提交。

## File Structure

| 文件 | 职责 |
| --- | --- |
| `backend/internal/prompt/defaults/comic_presets.json` | 漫画资产、漫画页、图片描述词推导预设。 |
| `backend/internal/batchfactoryv11/comic_assets_service.go` | 提取并保存 `source=comic` 的资产。 |
| `backend/internal/batchfactoryv11/comic_page_service.go` | 漫画页候选图和主图服务。 |
| `backend/internal/batchfactoryv11/comic_collection_service.go` | 按分镜主图建立漫画合集。 |
| `backend/internal/batchfactoryv11/comic_pages_mysql.go` | 漫画页和漫画合集的 MySQL 读写。 |
| `backend/internal/storage/batch_factory_v11_schema.go` | 追加漫画页和漫画合集迁移。 |
| `backend/internal/httpapi/batch_factory_v12_comic_pages.go` | 漫画页和漫画合集 HTTP 接口。 |
| `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx` | 模式开关、漫画提示词、图片描述词推导。 |
| `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx` | 漫画动作、分镜卡、漫画页/漫画合集媒体库。 |

---

### Task 1: 预设目录与互斥生产模式

**Files:**
- Modify: `backend/internal/prompt/defaults/comic_presets.json`
- Modify: `backend/internal/prompt/comic_presets.go`
- Modify: `backend/internal/httpapi/batch_factory_prompt_config.go`
- Modify: `backend/internal/batchfactoryv11/director_contract.go`
- Test: `backend/internal/prompt/comic_presets_test.go`
- Test: `backend/internal/httpapi/batch_factory_comic_prompt_config_test.go`

**Interfaces:** `productionMode` 只能是 `video` 或 `comic`，空值兼容为 `video`。提示词槽位为 `batch.comic-assets`、`batch.comic-page`、`batch.comic-image-derivation`；`VIP 智能角色推导` 是普通漫画资产提示词选项。

- [ ] **Step 1: Write the failing test**

```go
func TestComicCatalogHasAssetAndImageDerivationChoices(t *testing.T) {
    requirePreset(t, catalog, "batch-factory", "batch.comic-assets", "vip-smart-character-derive")
    requirePreset(t, catalog, "batch-factory", "batch.comic-image-derivation", "comic-image-derive-default")
}
func TestNormalizeProductionModeDefaultsToVideo(t *testing.T) {
    if got := normalizeProductionMode(SettingsPatch{}); got != "video" { t.Fatal(got) }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/prompt ./internal/httpapi ./internal/batchfactoryv11 -run 'TestComicCatalogHasAssetAndImageDerivationChoices|TestNormalizeProductionModeDefaultsToVideo' -count=1`

Expected: FAIL because the slots and normalizer do not exist.

- [ ] **Step 3: Write minimal implementation**

Add backend-owned presets and make `hydrateBatchFactoryPromptConfig` freeze `comic.assets`, `comic`, and `comic.imageDerivation`. Implement `normalizeProductionMode`, rejecting invalid values and treating empty legacy values as `video`.

- [ ] **Step 4: Run test to verify it passes**

Run: same command as Step 2. Expected: PASS.

- [ ] **Step 5: Commit**

Run: `git add backend/internal/prompt backend/internal/httpapi/batch_factory_prompt_config.go backend/internal/batchfactoryv11/director_contract.go; git commit -m "feat: add comic production preset selections"`

### Task 2: 漫画资产独立提取

**Files:**
- Create: `backend/internal/batchfactoryv11/comic_assets_service.go`
- Modify: `backend/internal/batchfactoryv11/types.go`
- Modify: `backend/internal/batchfactoryv11/stage_runs.go`
- Modify: `backend/internal/batchfactoryv11/stage_service.go`
- Modify: `backend/internal/batchfactoryv11/book_assets_mysql.go`
- Test: `backend/internal/batchfactoryv11/comic_assets_service_test.go`

**Interfaces:** `RunComicAssetExtraction(ctx, owner, batchID, bookID) ([]BookAsset, error)`; new stage `BookStageComicAssets = "comic-assets"`; automatic rows only replace `source=comic`.

- [ ] **Step 1: Write the failing test**

```go
func TestRunComicAssetExtractionReplacesOnlyComicAssets(t *testing.T) {
    // Seed director/manual/comic assets; refresh replaces only comic rows.
}
func TestComicAssetStageRejectsVideoMode(t *testing.T) {
    // Assert ErrConflict before the text provider is called.
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/batchfactoryv11 -run 'TestRunComicAssetExtractionReplacesOnlyComicAssets|TestComicAssetStageRejectsVideoMode' -count=1`

Expected: FAIL because the comic asset stage does not exist.

- [ ] **Step 3: Write minimal implementation**

Build a text-model contract from frozen `comic.assets`, reuse the existing character/scene/prop JSON parser, and add a store method that deletes only `source='comic'` before inserting replacements. Keep `BookStageAssets` for video assets; use the explicit comic stage.

- [ ] **Step 4: Run test to verify it passes**

Run: same command as Step 2. Expected: PASS.

- [ ] **Step 5: Commit**

Run: `git add backend/internal/batchfactoryv11; git commit -m "feat: isolate comic asset extraction"`

### Task 3: 漫画页和漫画合集持久化

**Files:**
- Modify: `backend/internal/storage/batch_factory_v11_schema.go`
- Modify: `backend/internal/storage/batch_factory_v11_schema_test.go`
- Modify: `backend/internal/batchfactoryv11/types.go`
- Create: `backend/internal/batchfactoryv11/comic_pages_mysql.go`
- Modify: `backend/internal/batchfactoryv11/memory_store.go`
- Test: `backend/internal/batchfactoryv11/comic_pages_test.go`

**Interfaces:**

```go
type ComicPage struct { ID, BatchID, BookID, VideoID, ComicPrompt, ImagePrompt, ModelID, AspectRatio, Status string }
type ComicPageImage struct { ID, ComicPageID, StorageRef, MediaType, Source string; IsPrimary bool; Revision int64 }
type ComicCollection struct { ID, BatchID, BookID, Status string; Items []ComicCollectionItem }
```

Migration `1100028` creates comic page, page image, collection, and collection item tables. Image bytes remain in the existing local-artifact/TOS boundary.

- [ ] **Step 1: Write the failing test**

```go
func TestComicPageImagesKeepVersionsAndOnePrimary(t *testing.T) { /* switch primary and reload */ }
func TestComicCollectionRejectsMissingStoryboardPrimary(t *testing.T) { /* expect ErrConflict */ }
func TestV11ComicMigrationIsAdditive(t *testing.T) { /* assert version 1100028 */ }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/storage ./internal/batchfactoryv11 -run 'TestV11ComicMigrationIsAdditive|TestComicPageImagesKeepVersionsAndOnePrimary|TestComicCollectionRejectsMissingStoryboardPrimary' -count=1`

Expected: FAIL because the storage models do not exist.

- [ ] **Step 3: Write minimal implementation**

Add owner-scoped MySQL and MemoryStore CRUD, candidate version history, transactional primary-image switching, and ordered storyboard snapshots. A collection may only use the primary image of every active video.

- [ ] **Step 4: Run test to verify it passes**

Run: same command as Step 2. Expected: PASS.

- [ ] **Step 5: Commit**

Run: `git add backend/internal/storage backend/internal/batchfactoryv11; git commit -m "feat: persist comic pages and collections"`

### Task 4: 分镜推导、漫画出图与合集接口

**Files:**
- Modify: `backend/internal/batchfactoryv11/comic_prompt_service.go`
- Create: `backend/internal/batchfactoryv11/comic_page_service.go`
- Create: `backend/internal/batchfactoryv11/comic_collection_service.go`
- Modify: `backend/internal/httpapi/router.go`
- Create: `backend/internal/httpapi/batch_factory_v12_comic_pages.go`
- Test: `backend/internal/batchfactoryv11/comic_prompt_service_test.go`
- Test: `backend/internal/httpapi/batch_factory_v12_comic_pages_test.go`

**Interfaces:**

```text
POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/comic-pages/generate
POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-pages/{pageId}/images/upload
PUT  /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-pages/{pageId}/images/{imageId}/primary
GET  /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-collections
POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-collections
```

`RunComicPromptExtraction` saves editable `comicPrompt` and provider-ready `comicImagePrompt` per video.

- [ ] **Step 1: Write the failing test**

```go
func TestComicDerivationUsesImageDerivationPreset(t *testing.T) { /* no video timing; non-empty image prompt */ }
func TestComicPageGenerateRejectsVideoModeAndEmptyPrompt(t *testing.T) { /* no provider call */ }
func TestCollectionHandlerReturnsAllMissingVideoIDs(t *testing.T) { /* HTTP conflict payload */ }
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/batchfactoryv11 ./internal/httpapi -run 'TestComicDerivationUsesImageDerivationPreset|TestComicPageGenerateRejectsVideoModeAndEmptyPrompt|TestCollectionHandlerReturnsAllMissingVideoIDs' -count=1`

Expected: FAIL because the service and routes do not exist.

- [ ] **Step 3: Write minimal implementation**

Generate `comicPrompt` from comic assets and the selected style, then generate `comicImagePrompt` with the selected derivation preset. Reuse configured image-model selection and local artifacts, but write `ComicPageImage`, not `BookAssetImage`. Validate owner, mode, model, prompt, and video before provider calls. The collection service snapshots storyboard ordinal order, returns every missing video ID, and never calls merge, FFmpeg, external publish, or upload.

- [ ] **Step 4: Run test to verify it passes**

Run: same command as Step 2. Expected: PASS.

- [ ] **Step 5: Commit**

Run: `git add backend/internal/batchfactoryv11 backend/internal/httpapi; git commit -m "feat: generate comic pages and collections"`

### Task 5: 单书配置与资产提示词界面

**Files:**
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryAiReasoningModal.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js`

**Interfaces:** Both switches update one `productionMode`. The existing asset selector displays comic asset choices in comic mode; it must not render a second “漫画资产提示词” selector. The comic configuration order is toggle, comic style prompt, image-description derivation.

- [ ] **Step 1: Write the failing test**

```js
assert.match(source, /productionMode/);
assert.match(source, /图片描述词推导/);
assert.doesNotMatch(source, /<PromptSelect label="漫画资产提示词"/);
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js`

Expected: FAIL because the mode and derivation configuration are incomplete.

- [ ] **Step 3: Write minimal implementation**

Bind the two switches to one mutually exclusive, always-enabled mode and retain inactive selections. Keep the screenshot’s existing asset selector; filter it by mode. Add API functions for Task 4 endpoints to the current V12 client.

- [ ] **Step 4: Run test to verify it passes**

Run: same command as Step 2. Expected: PASS.

- [ ] **Step 5: Commit**

Run: `git add frontend/src/shared/api/batchFactoryV11.js frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryAiReasoningModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js; git commit -m "feat: configure comic production mode"`

### Task 6: 分镜卡、漫画页与漫画合集媒体库

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

**Interfaces:** Comic mode shows “生成漫画”; the media modal has “漫画页” and “漫画合集” tabs. Candidate images support retry, upload, and set-primary. Video mode keeps current video actions and media.

- [ ] **Step 1: Write the failing test**

```js
assert.match(source, /生成漫画/);
assert.match(source, /漫画页/);
assert.match(source, /漫画合集/);
assert.match(source, /缺少主图/);
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: FAIL because comic media and collection views are absent.

- [ ] **Step 3: Write minimal implementation**

Reuse candidate-card upload and primary-image interaction but read comic-page APIs. In comic mode replace video generation with comic generation and video merge with 漫画合集. Show the server-returned missing-shot list; never use a candidate image as an implicit primary.

- [ ] **Step 4: Run test to verify it passes**

Run: same command as Step 2. Expected: PASS.

- [ ] **Step 5: Commit**

Run: `git add frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js; git commit -m "feat: show comic pages and comic collections"`

### Task 7: 运行时回归与嵌入式验收

**Files:** Only modify a focused file from Tasks 1-6 if a focused regression proves a defect.

- [ ] **Step 1: Run focused suites**

Run: `go test ./internal/storage ./internal/batchfactoryv11 ./internal/httpapi -run 'Comic|BookStage|BatchFactory' -count=1`

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryAiReasoningModal.comic.test.js`

Expected: PASS.

- [ ] **Step 2: Build embedded runtime**

Run: `npm --prefix frontend run build`

Run: `go build -o qiantie-local.exe ./cmd/qiantie`

Expected: both commands exit `0`; never copy `frontend/dist` back into source.

- [ ] **Step 3: Browser acceptance without generation**

At `http://localhost:4001/batch-factory`, verify exclusive switches, comic options in the existing asset selector, comic prompt labels, and the 漫画页/漫画合集 tabs. Do not click text generation, image generation, provider-backed actions, or upload.

- [ ] **Step 4: Record scope and commit focused repairs**

Record test totals, runtime SHA, `/health`, `/api/runtime-build-info`, and browser observations. Commit only any focused repair files with `fix: cover comic production regression`.

## Plan Self-Review

- Spec coverage: Tasks 1 and 5 cover modes, style and image-derivation controls; Task 2 covers comic assets through the existing selector; Tasks 3-4 cover pages, candidates, primary images, and storage; Task 6 covers 漫画合集 and media; Task 7 covers video non-regression and embedded acceptance.
- Scope boundary: no task uploads a comic collection, stitches images, creates a video task, or calls video merge.
- Type consistency: Task 3 defines `ComicPage`, `ComicPageImage`, and `ComicCollection`; Tasks 4 and 6 consume those entities.
- Placeholder scan: no `TODO`, `TBD`, or undefined deferred implementation step.
