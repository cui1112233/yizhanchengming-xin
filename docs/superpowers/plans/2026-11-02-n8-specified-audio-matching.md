# N8 指定音频自动匹配 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在漫画发布统一配置中自动匹配、允许手动修改并提交 N8 指定音频。

**Architecture:** 后端以 N8 `POST /center-api/material/audio/select` 为唯一音频目录来源，按已选组织和音频分类过滤启用音频。前端保存指定音频 ID/名称；分类或组织变更时重新读取候选并按“已保存匹配项→唯一候选→最新更新音频”规则自动选择，用户仍可在下拉框覆盖。

**Tech Stack:** Go、React、Ant Design、N8 Admin Token、Node test、Go test。

## Global Constraints

- Git `v88` 是唯一维护源码；禁止用构建产物覆盖源码。
- N8 管理员 Token 仅从环境变量读取，不能写入前端、日志或 Git。
- 指定音频必须同时匹配 `audio_category_id` 与 `organization_id`，且 `status=ENABLE`。
- 无匹配指定音频时禁止发布，界面必须显示明确原因。
- 发布请求同时提交 `audio_category_id` 和真实指定音频 ID。

---

### Task 1: N8 指定音频目录客户端

**Files:**
- Modify: `backend/internal/n8catalog/client.go`
- Test: `backend/internal/n8catalog/client_test.go`

**Interfaces:**
- Produces: `type Audio struct { ID int64; Name string; OrganizationID string; CategoryIDs []int64; Status string; UpdatedAt string }`
- Produces: `func (c *Client) ListAudio(ctx context.Context) ([]Audio, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestListAudioMapsN8Response(t *testing.T) {
  // Test server returns SUCCESS/data.list with id, name, organization_id,
  // category_ids, status and updated_at.
  // Assert ListAudio returns the enabled audio with CategoryIDs []int64{9}.
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/n8catalog -run TestListAudioMapsN8Response -count=1`
Expected: FAIL because `ListAudio` does not exist.

- [ ] **Step 3: Write minimal implementation**

Add `Audio` and `ListAudio`. Issue one POST to `${baseURL-without-/auth}/material/audio/select` with `{"page":1,"page_size":100}` and `N8-Admin-Token`. Decode `SUCCESS/data.list`, retaining only positive IDs, non-empty names and `ENABLE` status.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/n8catalog -run TestListAudioMapsN8Response -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/n8catalog/client.go backend/internal/n8catalog/client_test.go
git commit -m "feat: read N8 specified audio directory"
```

### Task 2: Expose filtered audio candidates and publish field

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_n8_comic_publish.go`
- Modify: `backend/internal/n8comic/client.go`
- Test: `backend/internal/httpapi/batch_factory_n8_comic_publish_test.go`
- Test: `backend/internal/n8comic/client_test.go`

**Interfaces:**
- Consumes: `n8catalog.Client.ListAudio(context.Context) ([]n8catalog.Audio, error)`
- Produces: `GET /api/batch-factory/v12/comic-publish/audios?audioCategoryId=<id>&organizationId=<id>` returning `{audios:[{id,name,updatedAt}]}`.
- Produces: `n8ComicPublishInput.SpecifiedAudioID int64`.

- [ ] **Step 1: Write failing HTTP tests**

```go
func TestComicPublishAudiosFiltersCategoryAndOrganization(t *testing.T) {
  // Directory returns three audio entries; only one has category 9,
  // organization 1 and ENABLE status.
  // GET .../audios?audioCategoryId=9&organizationId=1 must return that entry.
}

func TestComicPublishRequiresSpecifiedAudioID(t *testing.T) {
  // POST publish without specifiedAudioId must return a validation error.
}
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./internal/httpapi -run 'TestComicPublishAudiosFiltersCategoryAndOrganization|TestComicPublishRequiresSpecifiedAudioID' -count=1`
Expected: FAIL because endpoint and input field do not exist.

- [ ] **Step 3: Implement the directory route and validation**

Add an audio directory capability to the existing publisher. Filter server-side by exact numeric `audioCategoryId` in `CategoryIDs` and exact `organizationId`. Return candidates sorted by parsed `updated_at` descending, then ID descending. Reject an empty category or organization parameter with HTTP 400. Add `specifiedAudioId` to `n8ComicPublishInput`; reject publish if it is not a positive ID or is not present in the filtered directory for the chosen category/organization.

- [ ] **Step 4: Extend N8 create payload**

```go
payload["audio_category_id"] = input.AudioCategoryID
payload["audio_id"] = input.SpecifiedAudioID
```

Keep existing payload fields unchanged. The exact field name must be verified against a successful N8 create request before production release; until then, integration tests assert the selected contract and runtime validation prevents an unfiltered ID from being sent.

- [ ] **Step 5: Run tests to verify pass**

Run: `go test ./internal/httpapi ./internal/n8comic -count=1`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/httpapi/batch_factory_n8_comic_publish.go backend/internal/httpapi/batch_factory_n8_comic_publish_test.go backend/internal/n8comic/client.go backend/internal/n8comic/client_test.go
git commit -m "feat: validate specified N8 audio for comic publishing"
```

### Task 3: Add specified-audio API client and matching helper

**Files:**
- Modify: `frontend/src/shared/api/batchFactoryV12.js`
- Create: `frontend/src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.js`
- Test: `frontend/src/shared/api/batchFactoryV12.test.js`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.test.js`

**Interfaces:**
- Produces: `getN8ComicPublishAudios(audioCategoryId, organizationId, options)`.
- Produces: `pickSpecifiedAudio(savedAudioId, candidates)` returning a matching saved candidate, the sole candidate, or the latest candidate.

- [ ] **Step 1: Write failing tests**

```js
test('specified audio request includes category and organization', async () => {
  // Mock request and assert query contains audioCategoryId=9 and organizationId=1.
});

test('pickSpecifiedAudio preserves matching saved selection then latest candidate', () => {
  const candidates = [{ id: 2, updatedAt: '2026-10-01' }, { id: 3, updatedAt: '2026-11-01' }];
  assert.equal(pickSpecifiedAudio(2, candidates).id, 2);
  assert.equal(pickSpecifiedAudio(0, candidates).id, 3);
});
```

- [ ] **Step 2: Run failing tests**

Run: `node --test src/shared/api/batchFactoryV12.test.js src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.test.js`
Expected: FAIL because API and helper do not exist.

- [ ] **Step 3: Implement API and helper**

Use `URLSearchParams` to create the query. The helper accepts normalized candidate objects and sorts a copy by `updatedAt` descending and numeric `id` descending; it must not mutate caller data.

- [ ] **Step 4: Run tests**

Run: `node --test src/shared/api/batchFactoryV12.test.js src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.test.js`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/shared/api/batchFactoryV12.js frontend/src/shared/api/batchFactoryV12.test.js frontend/src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.js frontend/src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.test.js
git commit -m "feat: add specified audio matching client"
```

### Task 4: Render selector and persist selected audio

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

**Interfaces:**
- Consumes: `getN8ComicPublishAudios()` and `pickSpecifiedAudio()`.
- Persists: `comicPublishSettings.specifiedAudioId` and `comicPublishSettings.specifiedAudioName`.
- Sends: `specifiedAudioId` with the existing comic publish request.

- [ ] **Step 1: Write failing source assertions**

```js
assert.match(source, /指定音频/);
assert.match(source, /getN8ComicPublishAudios/);
assert.match(source, /specifiedAudioId/);
assert.match(source, /已自动匹配，可更改/);
```

- [ ] **Step 2: Run source tests to verify failure**

Run: `node --test src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`
Expected: FAIL because selector and payload field do not exist.

- [ ] **Step 3: Implement unified configuration selector**

Place a `指定音频` Select below the two audio category controls. Load candidates whenever `organizationId` or `audioCategoryId` changes. Show the automatic-match helper text only when selection was computed, not after user change. On manual change, persist both selected ID and name. On no candidates, clear both saved fields and show a warning.

- [ ] **Step 4: Implement publish dialog parity**

In `ComicMediaPanel`, load candidates using its selected organization/category, auto-select with the helper, let users override, require `specifiedAudioId > 0` in `canPublish`, and include `specifiedAudioId` in `publishComicCollection` payload. Preserve existing work and category controls.

- [ ] **Step 5: Run source tests**

Run: `node --test src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js
git commit -m "feat: select specified audio for comic publishing"
```

### Task 5: Full verification

**Files:**
- Modify only if a failing test proves an implementation bug.

- [ ] **Step 1: Run focused frontend tests**

Run: `node --test src/shared/api/batchFactoryV12.test.js src/user/pages/shuihuo/batchFactoryComicSpecifiedAudio.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`
Expected: PASS.

- [ ] **Step 2: Run backend tests**

Run: `go test ./internal/n8catalog ./internal/n8comic ./internal/httpapi -count=1`
Expected: PASS.

- [ ] **Step 3: Build frontend**

Run: `npm run build` from `frontend/`
Expected: successful Vite build.

- [ ] **Step 4: Confirm manual scenario**

Use category `9` and organization `1`: the verified N8 directory must show only audio records whose `category_ids` contains `9`, `organization_id` is `1`, and `status` is `ENABLE`. Confirm an automatic selection appears, manually change it when multiple candidates exist, save, reopen, and verify the saved matching ID remains selected.

- [ ] **Step 5: Commit only requested source changes**

```bash
git status --short
git add <only files changed for specified audio>
git commit -m "fix: complete N8 specified audio matching"
```
