# Batch Factory CM Prompt Library Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Put the runtime-confirmed CM asset prompts into the backend-owned preset catalog and display the complete comic prompt chain under `管理后台 -> 提示词库 -> 批量工厂 -> 漫画生成` without exposing internal category prompts in the user-facing unified asset selector.

**Architecture:** Keep every prompt body in Go-embedded backend JSON and use stable IDs plus protocol slots to route presets. Upgrade only recognized legacy system seeds, never administrator-created versions. The admin React page groups published, draft, and archived records by protocol metadata; the user Batch Factory continues selecting only `batch.comic-assets`, while category and three-view presets use dedicated internal slots.

**Tech Stack:** Go 1.23, embedded JSON prompt catalog, React 18, Ant Design 5, Node test runner, Vite, Go embed release path.

## Global Constraints

- Work directly on Git branch `v88`; do not create a new branch or worktree.
- Preserve all unrelated dirty files and stage only files named by each task.
- System prompt bodies live only in the Go backend prompt module; the frontend must not hardcode prompt bodies.
- User-visible Snow Manga-derived preset names use the `CM·` prefix.
- `CM·VIP 智能角色推导` is a preset name only and has no membership or payment restriction.
- The five CM preset display names are `CM·漫画资产统一推导`, `CM·VIP 智能角色推导`, `CM·主要场景推导`, `CM·关键道具推导`, and `CM·角色三视图`.
- The full comic prompt catalog appears under `提示词库 -> 批量工厂 -> 漫画生成` in five subgroups: asset derivation, comic storyboard and page, image description, style library, and comic constraints.
- Existing administrator-created prompt versions must never be overwritten by startup seeding.
- Frontend source is authoritative; do not implement by editing `frontend/dist`.
- Do not publish or deploy. Local verification uses the approved Go embed path only after source tests pass.

---

### Task 1: Add The Backend-Owned CM Prompt Set

**Files:**
- Create: `backend/internal/prompt/defaults/cm_asset_presets.json`
- Modify: `backend/internal/prompt/defaults/comic_presets.json`
- Modify: `backend/internal/prompt/comic_presets.go`
- Test: `backend/internal/prompt/comic_presets_test.go`

**Interfaces:**
- Consumes: exact runtime-observed strings from `G:/Backup/Documents/ChatGPT/漫画/雪漫画_反编译源码/RUNTIME_ADDITIONAL_PROMPTS.json` keys `cinegacha.core.visual_prompt_template.VISUAL_DESCRIPTION_PROMPT`, `SCENE_PRESET_PROMPT`, `PROP_PRESET_PROMPT`, and `cinegacha.core.image_prompt_sanwei.DEFAULT_CHARACTER_SANWEI_PROMPT`.
- Produces: five published presets addressable by stable ID and slot through `Catalog.PublishedBody(module, id)`.

- [ ] **Step 1: Write failing catalog contract tests**

Extend `TestSystemComicPresetsArePublishedAndUsable` with exact names, slots, and source phrases:

```go
tests := []struct {
    id, name, module, slot, phrase string
}{
    {"batch-comic-assets", "CM·漫画资产统一推导", "batch-factory", "batch.comic-assets", "年龄段变体拆分"},
    {"vip-smart-character-derive", "CM·VIP 智能角色推导", "batch-factory", "batch.comic.assets.character", "首句融合规则"},
    {"batch-comic-assets-cm-scene", "CM·主要场景推导", "batch-factory", "batch.comic.assets.scene", "昼夜/时段强制锚定"},
    {"batch-comic-assets-cm-prop", "CM·关键道具推导", "batch-factory", "batch.comic.assets.prop", "无则空数组"},
    {"batch-comic-character-three-view-cm", "CM·角色三视图", "batch-factory", "batch.comic.character-image-template", "正面近照、正面全身、侧面全身、背面全身"},
}
```

For every row, assert `PublishedBody` is found, `Name`, `Slot`, and body phrase match, and the body does not contain a user-visible `雪漫画·` preset name.

- [ ] **Step 2: Write failing startup-upgrade tests**

Add two tests:

```go
func TestEnsureComicPresetsUpgradesRecognizedLegacySystemSeeds(t *testing.T)
func TestEnsureComicPresetsPreservesAdministratorPublishedVersions(t *testing.T)
```

The first fixture contains the current published system records `batch-comic-assets` V2 named `漫画资产提示词` and `vip-smart-character-derive` V1 named `VIP 智能角色推导`, both with `createdBy: "system"`. After `EnsureComicPresets`, assert the old records are archived and the CM versions are published.

The second fixture contains `batch-comic-assets` V3 with `createdBy: "admin"`, `publishedBy: "admin"`, and body `管理员自定义正文`. After `EnsureComicPresets`, assert that exact record remains the only published version and its body and name are unchanged.

- [ ] **Step 3: Run the focused backend tests and verify failure**

Run:

```powershell
go test ./internal/prompt -run 'Test(SystemComicPresetsArePublishedAndUsable|EnsureComicPresetsUpgradesRecognizedLegacySystemSeeds|EnsureComicPresetsPreservesAdministratorPublishedVersions)$' -count=1
```

Working directory: `backend`

Expected: FAIL because the CM preset records and safe upgrade behavior do not exist.

- [ ] **Step 4: Add the exact CM prompt defaults**

Create `cm_asset_presets.json` with these records:

```json
[
  {"id":"batch-comic-assets","name":"CM·漫画资产统一推导","slot":"batch.comic-assets","version":3},
  {"id":"vip-smart-character-derive","name":"CM·VIP 智能角色推导","slot":"batch.comic.assets.character","version":2},
  {"id":"batch-comic-assets-cm-scene","name":"CM·主要场景推导","slot":"batch.comic.assets.scene","version":1},
  {"id":"batch-comic-assets-cm-prop","name":"CM·关键道具推导","slot":"batch.comic.assets.prop","version":1},
  {"id":"batch-comic-character-three-view-cm","name":"CM·角色三视图","slot":"batch.comic.character-image-template","version":1}
]
```

Each full record uses `module: "batch-factory"`, `kind: "base"`, `status: "published"`, a matching `protocolLock.operation`, and the exact backend prompt body. The unified body combines the character, scene, and prop rules into one JSON-object response with `characters`, `scenes`, and `props`; it preserves all runtime-confirmed constraints while adapting only variables and the outer response envelope. The dedicated character, scene, prop, and three-view bodies are copied from the four runtime keys named in `Interfaces`, replacing source placeholders only with the Batch Factory input names.

Remove the superseded `batch-comic-assets` and `vip-smart-character-derive` records from `comic_presets.json` so each built-in ID/version is owned by one embedded file.

- [ ] **Step 5: Implement version-aware safe seeding**

Embed the new file in `comic_presets.go`:

```go
//go:embed defaults/cm_asset_presets.json
var cmAssetPresetDefaults []byte
```

Decode and append it to the built-in defaults. Replace the ID-only skip rule with an `(id, version)` index. Add a narrowly scoped legacy upgrade helper that may archive and supersede only these recognized system seeds:

```go
var cmLegacySystemVersions = map[string]int{
    "batch-comic-assets":          2,
    "vip-smart-character-derive": 1,
}
```

Upgrade only when the currently published record has that exact ID/version, `createdBy == "system"`, and no later version exists. If a later or administrator-created published version exists, do not add a competing published record and do not mutate it. New IDs are seeded normally. A second call must make no changes.

- [ ] **Step 6: Run prompt tests and verify pass**

Run:

```powershell
go test ./internal/prompt -count=1
```

Working directory: `backend`

Expected: PASS.

- [ ] **Step 7: Commit the backend prompt set**

```powershell
git add -- backend/internal/prompt/defaults/cm_asset_presets.json backend/internal/prompt/defaults/comic_presets.json backend/internal/prompt/comic_presets.go backend/internal/prompt/comic_presets_test.go
git commit -m "feat(batch-factory): add CM asset prompt presets"
```

### Task 2: Group The Complete Comic Catalog In Admin

**Files:**
- Modify: `frontend/src/admin/pages/PresetLibraryPage.jsx`
- Create: `frontend/src/admin/pages/PresetLibraryPage.source.test.js`

**Interfaces:**
- Consumes: admin preset records containing `id`, `module`, and `protocolLock.slot`, `protocolLock.format`, `protocolLock.operation`, or `protocolLock.category`.
- Produces: a `漫画生成` section that owns all comic records before fallback sections consume them.

- [ ] **Step 1: Write the failing admin source contract test**

Create a Node source test that reads `PresetLibraryPage.jsx` and asserts:

```js
assert.match(source, /title:\s*'漫画生成'/);
assert.match(source, /title:\s*'资产推导'/);
assert.match(source, /title:\s*'漫画分镜与页面'/);
assert.match(source, /title:\s*'图片描述词'/);
assert.match(source, /title:\s*'画风库'/);
assert.match(source, /title:\s*'漫画约束'/);
assert.match(source, /batch\.comic\.assets\.character/);
assert.match(source, /batch\.comic\.character-image-template/);
```

Also assert the comic section is declared before the generic `约束设置` section so comic constraints cannot be consumed twice.

- [ ] **Step 2: Run the source test and verify failure**

Run:

```powershell
node --test src/admin/pages/PresetLibraryPage.source.test.js
```

Working directory: `frontend`

Expected: FAIL because `漫画生成` and its subgroups do not exist.

- [ ] **Step 3: Add the metadata-driven comic section**

Insert `漫画生成` before `约束设置` in `batchFactoryPresetSections`. Match records by stable IDs and protocol metadata, not by display-name text:

```js
const comicSlots = new Set([
  'batch.comic-assets',
  'batch.comic.assets.character',
  'batch.comic.assets.scene',
  'batch.comic.assets.prop',
  'batch.comic.character-image-template',
  'batch.comic-page',
  'batch.comic-image-derivation',
  'batch.comic-style'
]);

const comicConstraintCategories = new Set([
  'comic_quality',
  'comic_restriction',
  'comic_negative'
]);
```

The section subgroups are:

```js
[
  { title: '资产推导', filter: preset => comic asset and character-image slots },
  { title: '漫画分镜与页面', filter: preset => preset.protocolLock?.slot === 'batch.comic-page' },
  { title: '图片描述词', filter: preset => preset.protocolLock?.slot === 'batch.comic-image-derivation' },
  { title: '画风库', filter: preset => preset.protocolLock?.slot === 'batch.comic-style' },
  { title: '漫画约束', filter: preset => comicConstraintCategories.has(preset.protocolLock?.category) }
]
```

Exclude the three comic constraint categories from the later generic `约束设置` matcher. Preserve the user's existing removal of the historical `novel-panel` entries in this already-dirty file.

- [ ] **Step 4: Run the admin source test and frontend build**

Run:

```powershell
node --test src/admin/pages/PresetLibraryPage.source.test.js
npm run build
```

Working directory: `frontend`

Expected: source test PASS and Vite build PASS.

- [ ] **Step 5: Commit only the admin source files**

```powershell
git add -- frontend/src/admin/pages/PresetLibraryPage.jsx frontend/src/admin/pages/PresetLibraryPage.source.test.js
git commit -m "feat(admin): group CM comic prompt presets"
```

### Task 3: Protect The User-Facing Asset Selector

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryAiReasoningModal.comic.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`
- Modify only if tests expose a leak: the corresponding `.jsx` file.

**Interfaces:**
- Consumes: public prompt metadata from `Catalog.Published("batch-factory")`.
- Produces: user selectors that accept only slot `batch.comic-assets`; internal CM character, scene, prop, and three-view slots remain backend/admin-only.

- [ ] **Step 1: Add selector isolation assertions**

Assert each unified asset selector filters exactly on `item.slot === 'batch.comic-assets'` and does not use prefix matching such as `startsWith('batch.comic.assets')`. Assert the internal slot strings are absent from selector option builders.

- [ ] **Step 2: Run the three focused tests**

Run:

```powershell
node --test src/user/pages/shuihuo/BatchFactoryAiReasoningModal.comic.test.js src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js
```

Working directory: `frontend`

Expected: PASS with the current exact-slot filtering. If a test fails, make the smallest matching `.jsx` correction and rerun.

- [ ] **Step 3: Commit the selector regression tests**

```powershell
git add -- frontend/src/user/pages/shuihuo/BatchFactoryAiReasoningModal.comic.test.js frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js
git commit -m "test(batch-factory): isolate internal CM prompt slots"
```

### Task 4: Integrated Verification

**Files:**
- Verify only; no compiled artifact is committed.

**Interfaces:**
- Consumes: Tasks 1-3.
- Produces: source, backend, build, and local UI evidence for the CM prompt library stage.

- [ ] **Step 1: Run focused backend and frontend tests**

```powershell
go test ./internal/prompt ./internal/httpapi ./internal/batchfactoryv11 -count=1
node --test src/admin/pages/PresetLibraryPage.source.test.js src/user/pages/shuihuo/BatchFactoryAiReasoningModal.comic.test.js src/user/pages/shuihuo/BatchFactoryBookSettingsModal.source.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js
```

Run Go from `backend` and Node from `frontend`. Expected: PASS.

- [ ] **Step 2: Build the frontend and Go binary**

```powershell
npm run build
go test ./internal/web -count=1
go build ./cmd/qiantie
```

Run npm from `frontend` and Go commands from `backend`. Expected: all commands exit 0. Do not stage `frontend/dist` or the local binary.

- [ ] **Step 3: Verify the local admin UI**

Start the local Go-embedded runtime through the repository's approved local script or reuse the verified `:4001` listener only when its runtime SHA matches the new HEAD. In an authenticated browser open `/admin/presets`, select `批量工厂`, and verify:

1. `漫画生成` is a visible top-level collapse section.
2. All five subgroups are visible.
3. The five CM asset presets appear under `资产推导` with editable bodies.
4. The 25 style presets appear under `画风库` with their existing preview behavior.
5. Comic constraints appear only under `漫画生成 -> 漫画约束`.
6. Batch Factory's user asset selector shows `CM·漫画资产统一推导` but not the four internal presets.

- [ ] **Step 4: Record final evidence without publishing**

Report the three feature commit SHAs, focused test results, frontend/Go build results, local runtime URL, and any browser limitation. Do not deploy or publish.

