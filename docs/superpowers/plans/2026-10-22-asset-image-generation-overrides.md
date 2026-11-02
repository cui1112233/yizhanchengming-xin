# 资产图片生成独立设置 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make each non-voice asset persist an optional image model and aspect-ratio override, and make asset image generation use that override before the project defaults.

**Architecture:** Add two nullable properties to the durable `BookAsset` record and return them in the Shuihuo project read model. The image-generation route continues to receive project defaults, but resolves a model and ratio independently for each selected asset before it calls either the configured provider or the Doubao local executor.

**Tech Stack:** Go, MySQL, React, Ant Design, Node source tests, Go unit and HTTP-route tests.

## Global Constraints

- Preserve existing user data: existing assets have no override and continue to inherit project defaults.
- `imageModelId` and `imageAspectRatio` are optional per-asset values; clearing either field restores inheritance only for that field.
- Allow aspect ratios only `16:9`, `9:16`, and `1:1`.
- Show neither setting for voice assets.
- Do not submit real provider or Doubao generation requests during verification.

---

### Task 1: Persist and expose optional asset-level image settings

**Files:**
- Modify: `backend/internal/batchfactoryv11/types.go`
- Modify: `backend/internal/batchfactoryv11/book_assets_mysql.go`
- Modify: `backend/internal/batchfactoryv11/memory_store.go`
- Modify: `backend/internal/storage/batch_factory_v11_schema.go`
- Modify: `backend/internal/storage/batch_factory_v11_book_assets_schema_test.go`
- Modify: `backend/internal/batchfactoryv11/book_assets_test.go` or create `backend/internal/batchfactoryv11/book_asset_image_overrides_test.go`

**Interfaces:**
- Produces `BookAsset.ImageModelID string` as JSON `imageModelId` and `BookAsset.ImageAspectRatio string` as JSON `imageAspectRatio`.
- Extends `CreateBookAssetInput` and `UpdateBookAssetInput` with the same optional JSON fields.
- Stores them in nullable `image_model_id` and `image_aspect_ratio` columns of `batch_factory_v11_book_assets`.

- [ ] **Step 1: Write failing memory-store tests for a saved override and an inherited asset.**

```go
saved, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{
    Kind: "character", Name: "林默", Prompt: "青年男子",
    ImageModelID: "seedream-4.5", ImageAspectRatio: "1:1",
})
if err != nil || saved.ImageModelID != "seedream-4.5" || saved.ImageAspectRatio != "1:1" {
    t.Fatalf("saved override=%+v err=%v", saved, err)
}
```

- [ ] **Step 2: Run the focused test to verify it fails.**

Run: `go test ./backend/internal/batchfactoryv11 -run TestBookAssetImageOverrides -count=1`

Expected: the test does not compile because the input and asset fields do not exist.

- [ ] **Step 3: Add optional fields and validation.**

```go
ImageModelID     string `json:"imageModelId,omitempty"`
ImageAspectRatio string `json:"imageAspectRatio,omitempty"`
```

Trim the model identifier. Accept an empty ratio for inheritance and reject any non-empty ratio outside `16:9`, `9:16`, and `1:1` in both create and update normalization.

- [ ] **Step 4: Add an append-only schema helper for the two nullable columns.**

```go
func V11BookAssetImageOverrideStatements() []string {
    return []string{
        `ALTER TABLE batch_factory_v11_book_assets ADD COLUMN image_model_id VARCHAR(255) NULL AFTER relationships`,
        `ALTER TABLE batch_factory_v11_book_assets ADD COLUMN image_aspect_ratio VARCHAR(8) NULL AFTER image_model_id`,
    }
}
```

Register it beside the existing `V11BookAssetDetailsStatements` migration path. Do not change the checksum-sensitive historical book-assets migration.

- [ ] **Step 5: Thread the columns through MySQL and memory storage.**

Every `SELECT`, `INSERT`, `UPDATE`, and row scan for `BookAsset` must include the two columns. Use `COALESCE(image_model_id,'')` and `COALESCE(image_aspect_ratio,'')` when reading. A normal asset update must retain the empty value when the user chooses inheritance.

- [ ] **Step 6: Run focused storage tests.**

Run: `go test ./backend/internal/batchfactoryv11 ./backend/internal/storage -run "TestBookAssetImageOverrides|TestV11BookAssets" -count=1`

Expected: PASS.

### Task 2: Carry the override from the editor through the Shuihuo HTTP API

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Modify: `backend/internal/httpapi/shuihuo_project_compat_test.go`

**Interfaces:**
- `POST /api/shuihuo-production/projects/{projectId}/assets` accepts `imageModelId` and `imageAspectRatio`.
- `PUT /api/shuihuo-production/assets/{assetId}` accepts and returns both fields.
- `GET /api/shuihuo-production/projects/{projectId}` includes the two fields under each returned asset.

- [ ] **Step 1: Write a failing route test that creates and then updates an asset with both settings.**

```go
created := signedJSONRequest(t, api, now, "alice", http.MethodPost, projectURL+"/assets", map[string]any{
    "category": "character", "name": "林默", "prompt": "青年男子",
    "imageModelId": "seedream-4.5", "imageAspectRatio": "1:1",
})
if created.Code != http.StatusCreated { t.Fatalf("create=%d body=%s", created.Code, created.Body.String()) }
```

Assert the project response contains both properties, then send the update with empty strings and assert both properties are empty after reloading the project.

- [ ] **Step 2: Run the focused HTTP test to verify it fails.**

Run: `go test ./backend/internal/httpapi -run TestShuihuoAssetImageOverrides -count=1`

Expected: the response does not expose the settings or the input is ignored.

- [ ] **Step 3: Decode, pass through, and expose the fields.**

Add the two fields to the asset create/update request structs, pass them to `CreateBookAssetInput` and `UpdateBookAssetInput`, and add them to `shuihuoAssets`. Keep the existing revision and ownership checks unchanged.

- [ ] **Step 4: Run the focused HTTP test.**

Run: `go test ./backend/internal/httpapi -run TestShuihuoAssetImageOverrides -count=1`

Expected: PASS.

### Task 3: Resolve each selected asset independently during generation

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Modify: `backend/internal/httpapi/shuihuo_project_compat_test.go` or create `backend/internal/httpapi/shuihuo_asset_generation_test.go`

**Interfaces:**
- Add `resolveShuihuoAssetImageSettings(asset, fallbackModelID, fallbackAspectRatio) (modelID, aspectRatio string)`.
- `generateShuihuoAssetImages` obtains the fallback pair once, then resolves a pair per selected asset.

- [ ] **Step 1: Write a failing pure-function test for precedence and clearing.**

```go
asset := batchfactoryv11.BookAsset{ImageModelID: "seedream-4.5", ImageAspectRatio: "1:1"}
model, ratio := resolveShuihuoAssetImageSettings(asset, "default-image", "9:16")
if model != "seedream-4.5" || ratio != "1:1" { t.Fatalf("got %q %q", model, ratio) }

inheritedModel, inheritedRatio := resolveShuihuoAssetImageSettings(batchfactoryv11.BookAsset{}, "default-image", "9:16")
if inheritedModel != "default-image" || inheritedRatio != "9:16" { t.Fatalf("got %q %q", inheritedModel, inheritedRatio) }
```

- [ ] **Step 2: Run the focused test to verify it fails.**

Run: `go test ./backend/internal/httpapi -run TestResolveShuihuoAssetImageSettings -count=1`

Expected: the resolver is undefined.

- [ ] **Step 3: Implement per-asset resolution before provider dispatch.**

Inside the selected-asset loop, resolve `assetModelID` and `assetAspectRatio`. Use that pair for catalog lookup, the external provider payload, and `waitForLocalExecutorImage`. Do not use a single model lookup before the loop because two selected assets may use different models.

- [ ] **Step 4: Add a dry-run/provider-recording test.**

Use an in-process HTTP test server or recording local-executor store to select two assets: one with `1:1`/its own model and one with no override. Assert the captured requests use `1024x1024` and the fallback `9:16` size respectively, without calling any real provider.

- [ ] **Step 5: Run focused generation tests.**

Run: `go test ./backend/internal/httpapi -run "TestResolveShuihuoAssetImageSettings|TestGenerateShuihuoAssetImages" -count=1`

Expected: PASS.

### Task 4: Add the controls to the asset editor and verify the real interaction

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/AssetsView.jsx`
- Create: `frontend/src/user/pages/shuihuo/AssetsView.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo-production.css` only if the existing two-column editor needs spacing adjustment.

**Interfaces:**
- The `editing` asset draft carries optional `imageModelId` and `imageAspectRatio` fields.
- The save payload retains empty fields to clear an existing override.

- [ ] **Step 1: Write a failing source test for editor controls and the voice exclusion.**

```js
assert.match(source, /图片模型/);
assert.match(source, /图片画幅/);
assert.match(source, /editing\?\.category !== 'voice'/);
assert.match(source, /使用项目默认/);
```

- [ ] **Step 2: Run the source test to verify it fails.**

Run: `node --test frontend/src/user/pages/shuihuo/AssetsView.source.test.js`

Expected: FAIL because the asset editor has no controls or inheritance label.

- [ ] **Step 3: Render per-asset controls in the editor.**

For non-voice assets, render an allow-clear `Select` using the already-loaded `imageModels`, and an allow-clear `Select` whose empty placeholder is `使用项目默认`. Store `null` for clear actions so the JSON payload expresses inheritance. Do not change the top-toolbar settings because they remain project defaults.

- [ ] **Step 4: Prevent invalid asset-level model choices.**

When saving, reject a non-empty `imageModelId` that is not in the current enabled image-model list. Leave inherited assets valid when no project model is selected; generation will retain its existing project-default validation.

- [ ] **Step 5: Run frontend checks and build.**

Run: `node --test frontend/src/user/pages/shuihuo/AssetsView.source.test.js`

Run: `npm --prefix frontend run build`

Expected: both PASS.

- [ ] **Step 6: Verify with the authenticated local page without generating.**

Open `/batch-factory`, enter an existing project’s assets screen, edit a non-voice asset, choose a model and `1:1`, save, close, and reopen it. Verify both values remain. Clear both, save, reopen, and verify `使用项目默认`. Open an audio asset and verify neither image control appears. Do not click the generation action.

### Task 5: Run the affected regression set

**Files:**
- No source files beyond any test corrections exposed by this task.

- [ ] **Step 1: Run backend and frontend focused regressions.**

Run: `go test ./backend/internal/batchfactoryv11 ./backend/internal/storage ./backend/internal/httpapi -run "TestBookAssetImageOverrides|TestV11BookAssets|TestShuihuoAssetImageOverrides|TestResolveShuihuoAssetImageSettings|TestGenerateShuihuoAssetImages" -count=1`

Run: `node --test frontend/src/user/pages/shuihuo/AssetsView.source.test.js`

- [ ] **Step 2: Run the local Go review health check after rebuilding the embedded frontend and Go binary.**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Run: `powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1`

Expected: the targeted test suite and local review health check pass. Record any pre-existing unrelated failure separately rather than masking it.
