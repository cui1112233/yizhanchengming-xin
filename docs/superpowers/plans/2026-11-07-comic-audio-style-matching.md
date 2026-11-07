# 漫画音频按小说风格匹配 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 漫画发布时按每本书已保存的男女频和风格匹配 N8 音频分类，从该分类随机选择启用音频并持久化，重试时复用同一结果。

**Architecture:** 后端是最终裁决者：新增独立的音频匹配单元，根据书籍元数据和实时 N8 目录得到分类，再把随机结果保存到书级 `comicAudioSelection`。前端复用同一张明确的对照表用于配置预览，并停止把混合批次强行收敛成一个分类；手动选择继续通过 `comicPublishSettings` 覆盖自动选择。

**Tech Stack:** Go、MySQL-backed `batchfactoryv11.Store`、React、Ant Design、Node test runner、Go testing

## Global Constraints

- 每本书单独匹配，混合风格批次不得因分类不同而失败。
- 分类名称和分类 ID 必须来自 N8 实时目录的同一条启用记录。
- 自动随机结果保存后，刷新、自动重试、发布重试和断点续跑必须复用。
- 手动指定优先，自动逻辑不得覆盖手动结果。
- 不重新调用 AI，不修改男女频和风格识别方式。
- 不修改发布账号、组织和作品 ID 的匹配规则。
- 保留工作区内所有无关未提交改动，每次提交只包含本任务文件。

---

## File Structure

- Create `backend/internal/httpapi/batch_factory_comic_audio_match.go`: 风格映射、实时分类解析、随机选择、书级持久化与冲突复用。
- Create `backend/internal/httpapi/batch_factory_comic_audio_match_test.go`: 后端映射、随机保存、重试复用、手动优先和错误隔离测试。
- Modify `backend/internal/httpapi/batch_factory_n8_comic_publish.go`: 发布入口接入书级音频解析，支持现有 `specifiedAudioId` 字段。
- Modify `backend/internal/batchfactoryv11/settings.go`: 深合并 `comicPublishSettings`，让书级手动覆盖不丢失批次公共字段。
- Modify `backend/internal/batchfactoryv11/settings_v23_test.go`: 固化漫画发布设置的批次/单书合并行为。
- Modify `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.js`: 完整风格映射和逐书推断接口。
- Modify `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`: 覆盖全部风格、男女频兜底和混合批次。
- Modify `frontend/src/user/pages/shuihuo/batchFactoryComicPublishAccounts.js`: 有已保存音频时复用，否则从启用候选中随机选择。
- Modify `frontend/src/user/pages/shuihuo/batchFactoryComicPublishAccounts.test.js`: 固化随机、复用与禁用音频过滤。
- Modify `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`: 批次自动模式显示“逐本匹配”，不保存一个覆盖全批次的随机音频；单书仍允许手动覆盖。
- Modify `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`: 验证逐本匹配提示和手动覆盖入口。

---

### Task 1: 完整的风格分类对照表

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`
- Create: `backend/internal/httpapi/batch_factory_comic_audio_match.go`
- Create: `backend/internal/httpapi/batch_factory_comic_audio_match_test.go`

**Interfaces:**
- Produces frontend `inferBookN8AudioCategory(book, categories) -> { category, categoryName, reason }`。
- Produces backend `comicAudioCategoryName(book batchfactoryv11.Book) (string, error)`。

- [ ] **Step 1: 写前后端失败测试**

前端用表驱动测试覆盖 17 种已知风格，并明确断言：`现代悬疑 -> 现代悬疑`、`男频都市 -> 男频都市`、`古风虐文 -> 古风虐文`、`职场打脸 + 女频 -> 现代通用`、未知风格按男女频兜底。

后端加入对应表驱动测试：

```go
for _, tc := range []struct{ style, gender, want string }{
    {"现代悬疑", "女频", "现代悬疑"},
    {"男频都市", "男频", "男频都市"},
    {"古风虐文", "女频", "古风虐文"},
    {"职场打脸", "女频", "现代通用"},
    {"未知", "男频", "男频都市"},
} {
    book := batchfactoryv11.Book{SourceMetadata: map[string]any{"style": tc.style, "gender": tc.gender, "classifyStatus": "classified"}}
    got, err := comicAudioCategoryName(book)
    if err != nil || got != tc.want { t.Fatalf("got %q, %v; want %q", got, err, tc.want) }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend; node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`

Run: `cd backend; go test ./internal/httpapi -run ComicAudioCategory -count=1`

Expected: 新接口或新映射尚不存在而失败。

- [ ] **Step 3: 实现最小映射函数**

使用一个明确映射表，不调用 AI；先匹配风格，只有 `职场打脸` 和未知风格读取男女频。实时目录只接受 `status=ENABLE` 且名称唯一的记录。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd frontend; node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`

Run: `cd backend; go test ./internal/httpapi -run ComicAudioCategory -count=1`

Expected: PASS。

- [ ] **Step 5: 提交本任务文件**

```powershell
git add -- frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.js frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js backend/internal/httpapi/batch_factory_comic_audio_match.go backend/internal/httpapi/batch_factory_comic_audio_match_test.go
git commit -m "feat(batch-factory): 按小说风格匹配漫画音频分类"
```

### Task 2: 随机选择、保存和重试复用

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_comic_audio_match.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_audio_match_test.go`
- Modify: `backend/internal/httpapi/batch_factory_n8_comic_publish.go`

**Interfaces:**
- Consumes: `comicAudioCategoryName(book) (string, error)`。
- Produces: `resolveBookComicAudio(ctx, owner, batchID, bookID, config) (comicAudioSelection, error)`。
- Persists top-level book setting `comicAudioSelection` with `categoryName`, `categoryId`, `audioName`, `audioId`, `mode`, `gender`, `style`, `selectedAt`。

- [ ] **Step 1: 写失败测试**

测试使用两个启用音频和可控 `RandomIndex`：第一次固定选择第二首并断言写入书级设置；第二次把随机函数改成第一首，仍必须复用已保存的第二首。另测分类不存在、分类下无音频、手动 `specifiedAudioId` 优先及保存冲突后读取赢家。自动保存的音频仍启用时直接复用；被目录明确标记为不可用时才在原分类重新随机。手动音频不可用时返回明确错误，不得自动覆盖。

同时增加预检测试：自动逐书模式只检查组织和 N8 音频目录可访问，不得在缺少书籍上下文时擅自选择目录第一项；手动模式继续校验用户指定的分类和音频。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend; go test ./internal/httpapi -run 'ComicAudioSelection|PublishFromSettingsUsesBookAudio' -count=1`

Expected: 缺少书级解析与持久化而失败。

- [ ] **Step 3: 实现后端书级选择**

给 `BatchFactoryN8ComicPublisher` 增加测试可注入的：

```go
RandomIndex func(upper int) (int, error)
```

生产默认使用 `crypto/rand.Int`。选择前过滤启用音频；保存时使用当前书 revision。遇到 `ErrConflict` 重新读取书级 `comicAudioSelection`，以数据库中先保存的结果为准。`PublishFromSettings` 解析 `specifiedAudioId`/`specifiedAudioName`，并在真正发布前调用书级解析。`ValidateSettings` 区分自动与手动模式：自动模式不提前替某本书选音频，手动模式仍验证已指定值。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend; go test ./internal/httpapi -run 'ComicAudioSelection|PublishFromSettings' -count=1`

Expected: PASS。

- [ ] **Step 5: 提交本任务文件**

```powershell
git add -- backend/internal/httpapi/batch_factory_comic_audio_match.go backend/internal/httpapi/batch_factory_comic_audio_match_test.go backend/internal/httpapi/batch_factory_n8_comic_publish.go
git commit -m "feat(batch-factory): 保存并复用逐书随机音频"
```

### Task 3: 手动覆盖与前端逐书提示

**Files:**
- Modify: `backend/internal/batchfactoryv11/settings.go`
- Modify: `backend/internal/batchfactoryv11/settings_v23_test.go`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicPublishAccounts.js`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryComicPublishAccounts.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

**Interfaces:**
- Consumes: `inferBookN8AudioCategory` 和后端 `comicAudioSelection`。
- Produces: `chooseN8Audio(audios, savedID, random = Math.random)`，保存值存在时复用，否则随机启用候选。

- [ ] **Step 1: 写失败测试**

后端断言 `ResolveSettings(batch, book)` 对 `comicPublishSettings` 深合并，书级 `specifiedAudioId` 覆盖批次值，同时保留批次 `organizationId`。

前端断言保存 ID 永远优先、禁用音频永不入选、注入 `() => 0.99` 时选中最后一首，并验证统一配置出现“发布时按每本书的男女频和风格自动匹配并随机选择音频”。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend; go test ./internal/batchfactoryv11 -run ComicPublishSettings -count=1`

Run: `cd frontend; node --test src/user/pages/shuihuo/batchFactoryComicPublishAccounts.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

Expected: 深合并、随机接口或新提示尚不存在而失败。

- [ ] **Step 3: 实现手动优先和界面提示**

`ResolveSettings` 对 `comicPublishSettings` 使用现有 `mergeJSONObject`。统一配置在自动模式不再把某一本或整个批次推断出的分类、音频写回批次；单书手动选择仍保存到 `comicPublishSettings`，后端发布时优先使用。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend; go test ./internal/batchfactoryv11 -run ComicPublishSettings -count=1`

Run: `cd frontend; node --test src/user/pages/shuihuo/batchFactoryComicPublishAccounts.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

Expected: PASS。

- [ ] **Step 5: 提交本任务文件**

```powershell
git add -- backend/internal/batchfactoryv11/settings.go backend/internal/batchfactoryv11/settings_v23_test.go frontend/src/user/pages/shuihuo/batchFactoryComicPublishAccounts.js frontend/src/user/pages/shuihuo/batchFactoryComicPublishAccounts.test.js frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js
git commit -m "fix(batch-factory): 漫画音频手动覆盖优先于自动匹配"
```

### Task 4: 全链路回归与本地可见验证

**Files:**
- Modify only if a regression test exposes a task-related defect.

**Interfaces:**
- Consumes all prior task interfaces.
- Produces verified source, embedded frontend build, and local runtime behavior.

- [ ] **Step 1: 运行后端相关测试**

Run: `cd backend; go test ./internal/httpapi ./internal/batchfactoryv11 -count=1`

Expected: PASS。

- [ ] **Step 2: 运行前端完整测试与构建**

Run: `cd frontend; npm test`

Run: `cd frontend; npm run build`

Expected: 全部通过；构建生成可嵌入资源。

- [ ] **Step 3: 构建并重启确认的本地二进制**

使用项目现有本地构建脚本生成包含前端资源的 `backend/qiantie-local.exe`，只替换已确认监听 `localhost:4001` 的本地进程。

- [ ] **Step 4: 检查真实运行结果**

访问 `/health`、`/api/runtime-build-info` 和当前批量工厂页面。确认统一配置显示逐本自动匹配说明；选择单书时仍可手动覆盖；混合风格批次不再显示“批次分类不一致”。

- [ ] **Step 5: 最终提交任务修复（如有）**

只暂存本任务验证中新增的修复文件，不包含任何无关脏文件。
