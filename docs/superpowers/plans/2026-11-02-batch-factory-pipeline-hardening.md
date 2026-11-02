# 批量工厂链路加固实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按设计文档落地 A 最小加固（预检降级、操作反馈、发布幂等、缺页补图）+ B 状态机收敛（阶段契约、增量状态存储、常显状态行）+ C 拆分清理。

**Architecture:** Go 后端 `backend/internal/httpapi` 改预检/发布/漫画补图/状态存储；React 前端 `frontend/src/user/pages/shuihuo` 改反馈与状态行；契约 JSON 单一来源由 Go 测试守护；每任务独立提交。

**Tech Stack:** Go 1.x（`go build ./...`、`go test ./backend/internal/httpapi/`）、React + Vite、vitest source test、MySQL 容器（测试不依赖）。

**Spec:** `docs/superpowers/specs/2026-11-02-batch-factory-pipeline-hardening-design.md`

## Global Constraints

- 目标分支 `v88`；每个任务独立 `git commit`，**只 stage 本任务相关文件**。
- 以下 WIP 文件属于其它会话，**任何任务都不得 stage 或修改**：`backend/internal/batchfactoryv11/mysql_store.go`、`frontend/src/user/pages/shuihuo/ProjectsView.jsx`、`frontend/src/user/pages/shuihuo/shuihuo-production.css`、`frontend/src/user/pages/shuihuo/batchFactoryProjects.js`、`frontend/src/user/pages/shuihuo/batchFactoryProjects.test.js`（提交前用 `git status` 核对）。
- 用户可见文案中 `121` 必须写 `视频管理系统`。
- 工作目录 `f:\脚本测试\一战晟铭-v88`，所有 Go 命令在此目录执行。
- Go 测试命令统一：`go test ./backend/internal/httpapi/ -run <TestName> -v`；每次提交前 `go build ./...`。
- 前端测试命令：`cd frontend && npx vitest run src/user/pages/shuihuo/<test-file>`。
- 不引入新依赖；不改分类算法、提示词策略、121 表单字段映射。

---

### Task 1: A1 预检降级为「全局致命 + 逐书跳过」

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_automation.go`（`CreateJob` L686-688；`validateAutomationPreflight` L732-815；`run()` 书循环，插入点 L1681 `if len(productionLines(book)) == 0 {` 之前）
- Test: `backend/internal/httpapi/batch_factory_automation_test.go`（追加）

**Interfaces:**
- Consumes: 现有 `isDirectFirstSourceBlocked`、`comicSettingString`、`hasAutomationPublishProfile`、`reasonOrDefault`、`appendAutomationBookLog`。
- Produces:
  - `func (c *BatchAutomationCoordinator) validateAutomationGlobalPreflight(ctx context.Context, owner string, batch batchfactoryv11.Batch, runMode string) error`
  - `func (c *BatchAutomationCoordinator) automationBookPreflightReason(ctx context.Context, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, runMode string) string`
  - 书状态新值：`status=blocked, stage=preflight`（`blocked` 已在终态集合内，重试按钮可复位）。

- [ ] **Step 1: 写失败测试**

在 `batch_factory_automation_test.go` 末尾追加（复用文件内已有 fake，参考现有 `Test...Automation...` 构造方式）：

```go
func TestAutomationBookPreflightReason(t *testing.T) {
	c := &BatchAutomationCoordinator{}
	batch := batchfactoryv11.Batch{}
	book := batchfactoryv11.Book{ID: "b1", Title: "书一"}
	// 缺文本模型
	if reason := c.automationBookPreflightReason(context.Background(), "owner", batch, book, "execute_only"); reason == "" {
		t.Fatal("expected text model reason")
	}
	// storyboard_only 模式下只查文本模型
	patch := batchfactoryv11.SettingsPatch{"textModelId": json.RawMessage(`"m1"`)}
	book.SettingsState.Patch = patch
	if reason := c.automationBookPreflightReason(context.Background(), "owner", batch, book, "storyboard_only"); reason != "" {
		t.Fatalf("storyboard_only should pass, got %q", reason)
	}
	// 视频模式缺视频模型
	if reason := c.automationBookPreflightReason(context.Background(), "owner", batch, book, "execute_only"); !strings.Contains(reason, "视频模型") {
		t.Fatalf("expected video model reason, got %q", reason)
	}
}

func TestAutomationGlobalPreflightIgnoresBookConfig(t *testing.T) {
	c := &BatchAutomationCoordinator{}
	batch := batchfactoryv11.Batch{} // 无书级配置也不应整批拒绝
	if err := c.validateAutomationGlobalPreflight(context.Background(), "owner", batch, "storyboard_only"); err != nil {
		t.Fatalf("global preflight should not inspect book config: %v", err)
	}
}
```

- [ ] **Step 2: 跑测试确认编译失败**

Run: `go test ./backend/internal/httpapi/ -run 'TestAutomationBookPreflightReason|TestAutomationGlobalPreflightIgnoresBookConfig' -v`
Expected: FAIL（函数未定义）。

- [ ] **Step 3: 实现拆分**

将 `validateAutomationPreflight`（L732-815）改写为两个函数：

```go
// validateAutomationGlobalPreflight 只保留“全批不可执行”的条件：执行器整体缺失。
func (c *BatchAutomationCoordinator) validateAutomationGlobalPreflight(ctx context.Context, owner string, batch batchfactoryv11.Batch, runMode string) error {
	reasons := make([]string, 0)
	if c.stages == nil || c.stages.Director == nil || c.stages.Director.Provider == nil {
		reasons = append(reasons, "文本模型执行器未配置")
	}
	productionMode := comicSettingString(batch.SettingsState.Patch, "productionMode", "video")
	if productionMode == "comic" && runMode != "storyboard_only" {
		if c.stages == nil || c.stages.ImageReady == nil {
			reasons = append(reasons, "图片生成执行器未配置")
		} else if ready, reason := c.stages.ImageReady(ctx, owner); !ready {
			reasons = append(reasons, reasonOrDefault(reason, "图片生成执行器不可用"))
		}
	}
	if productionMode != "comic" && runMode != "storyboard_only" {
		if c.stages == nil || c.stages.VideoReady == nil {
			reasons = append(reasons, "视频生成执行器未配置")
		} else if ready, reason := c.stages.VideoReady(ctx, owner); !ready {
			reasons = append(reasons, reasonOrDefault(reason, "视频生成执行器不可用"))
		}
	}
	if runMode == "full_submit" {
		if productionMode == "comic" && c.comicPublish == nil {
			reasons = append(reasons, "漫画发布执行器未配置")
		}
		if productionMode != "comic" && c.publish121 == nil && c.external == nil {
			reasons = append(reasons, "视频发布执行器未配置")
		}
	}
	if len(reasons) == 0 {
		return nil
	}
	return fmt.Errorf("%w: 自动生产前检查失败：%s", batchfactoryv11.ErrInvalid, strings.Join(reasons, "；"))
}

// automationBookPreflightReason 逐书检查，返回空串表示可执行。只影响单书。
func (c *BatchAutomationCoordinator) automationBookPreflightReason(ctx context.Context, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, runMode string) string {
	if isDirectFirstSourceBlocked(book) {
		title := strings.TrimSpace(book.Title)
		if title == "" {
			title = "未命名作品"
		}
		return fmt.Sprintf("《%s》书城原文获取失败，等待人工处理", title)
	}
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	if comicSettingString(effective, "textModelId", "") == "" {
		return "文本模型未配置"
	}
	if runMode == "storyboard_only" {
		return ""
	}
	productionMode := comicSettingString(batch.SettingsState.Patch, "productionMode", "video")
	if productionMode == "comic" {
		if comicSettingString(effective, "imageModelId", "") == "" {
			return "图片模型未配置"
		}
	} else if comicSettingString(effective, "videoModelId", "") == "" {
		return "视频模型未配置"
	}
	if runMode == "full_submit" {
		var publish map[string]any
		if raw := effective["publishSettings"]; len(raw) > 0 {
			_ = json.Unmarshal(raw, &publish)
		}
		if publish == nil || !hasAutomationPublishProfile(publish) {
			return "发布配置未绑定网站或组织"
		}
		if productionMode == "comic" && c.comicPublishPreflight != nil {
			if err := c.comicPublishPreflight(ctx, effective); err != nil {
				return err.Error()
			}
		}
	}
	return ""
}
```

删除旧 `validateAutomationPreflight`。`CreateJob` L686 改为：

```go
if err := c.validateAutomationGlobalPreflight(ctx, owner, batch, runMode); err != nil {
	return nil, err
}
```

`run()` 书循环内，在 `if len(productionLines(book)) == 0 {`（L1681）之前插入：

```go
if reason := c.automationBookPreflightReason(ctx, owner, batch, book, automationStringValue(job["runMode"], "full_submit")); reason != "" {
	entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "blocked", "preflight", "", reason, c.now()
	appendAutomationBookLog(entry, c.now(), "preflight", reason)
	persist()
	continue
}
```

注意：该插入点位于 `waiting_submit` 分支之后，定时提交的书不会被误拦；定时任务走同一个 `run()`，自动补齐调度缺口。

- [ ] **Step 4: 跑测试 + 编译**

Run: `go build ./... ; go test ./backend/internal/httpapi/ -run 'TestAutomation' -v`
Expected: PASS。另跑全量 `go test ./backend/internal/httpapi/` 确认无回归（预检相关旧测试若断言整批拒绝，按新语义改为断言全局条件）。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/httpapi/batch_factory_automation.go backend/internal/httpapi/batch_factory_automation_test.go
git commit -m "fix(automation): 预检降级为全局致命+逐书跳过，单书配置缺陷不再整批零执行"
```

---

### Task 2: A2 自动化操作全程可见反馈

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`（`runAutomationAction` L4249-4303；在组件state区追加 `automationBusyAtRef`）
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryAutomationBusyFeedback.source.test.js`（新建）

**Interfaces:**
- Consumes: 现有 `automationBusy`/`setAutomationBusy`、`message`。
- Produces: 无新导出；行为契约 = busy 时提示、30s 看门狗、提交期间 loading。

- [ ] **Step 1: 写失败 source test**

```js
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { describe, it, expect } from 'vitest';

const here = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

describe('runAutomationAction 可见反馈', () => {
  it('不存在静默吞点击的裸 return', () => {
    expect(source).not.toContain('if (!batch?.id || automationBusy) return;');
  });
  it('busy 时给出可见提示', () => {
    expect(source).toContain('上一次操作仍在执行，请稍候');
  });
  it('有 30 秒看门狗', () => {
    expect(source).toContain('automationBusyAtRef');
    expect(source).toMatch(/30\s*\*\s*1000|30000/);
  });
  it('提交期间有 loading 常驻提示', () => {
    expect(source).toContain('正在提交自动化任务');
  });
});
```

Run: `cd frontend && npx vitest run src/user/pages/shuihuo/BatchFactoryAutomationBusyFeedback.source.test.js`
Expected: FAIL。

- [ ] **Step 2: 实现**

在组件内 `automationBusy` state 声明附近加：

```js
const automationBusyAtRef = useRef(0);
```

`runAutomationAction` 开头改为：

```js
async function runAutomationAction(key, { immediate = false, bookIds = null, presetId = '' } = {}) {
  if (!batch?.id) { message.error('页面批次尚未加载，请刷新后重试'); return; }
  if (automationBusy) {
    if (Date.now() - automationBusyAtRef.current > 30 * 1000) {
      automationBusyAtRef.current = 0;
      setAutomationBusy('');
      message.warning('上一次请求超时，已重置状态');
    } else {
      message.info('上一次操作仍在执行，请稍候');
      return;
    }
  }
```

`setAutomationBusy(key)` 之后加 `automationBusyAtRef.current = Date.now();`。try 前加：

```js
const hideLoading = message.loading('正在提交自动化任务…', 0);
```

`finally` 块改为：

```js
} finally { hideLoading(); automationBusyAtRef.current = 0; setAutomationBusy(''); }
```

- [ ] **Step 3: 跑测试**

Run: `cd frontend && npx vitest run src/user/pages/shuihuo/BatchFactoryAutomationBusyFeedback.source.test.js src/user/pages/shuihuo/BatchFactoryAutomationStartFeedback.source.test.js`
Expected: PASS（含旧反馈测试不回归）。

- [ ] **Step 4: Commit**

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/BatchFactoryAutomationBusyFeedback.source.test.js
git commit -m "fix(workbench): 自动化按钮点击全程可见反馈，busy 不再静默吞点击"
```

---

### Task 3: A3 发布幂等三态

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_v11_publish_121.go`（`Submit` L131-318；文件头部加哨兵）
- Modify: `backend/internal/httpapi/batch_factory_automation.go`（`run()` 内 video 模式发布失败处，识别哨兵写 `stage=upload_confirm`）
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryNovelListHelpers.js`（`isBookUploadedTo121` L597-599）
- Test: `backend/internal/httpapi/batch_factory_v11_publish_121_test.go`（追加）

**Interfaces:**
- Produces:
  - `var errPublishUnconfirmed = errors.New("publish uploaded but confirmation readback failed")`（httpapi 包内）
  - book metadata 新状态：`websiteSubmitStatus=upload_pending_confirm`，`websiteSubmitHeads=[]string`。
  - 前端 `isBookUploadedTo121` 对 `uploaded` 与 `upload_pending_confirm` 都返回 true。

- [ ] **Step 1: 写失败测试**

参考文件内已有 fake connector/store，追加：

```go
func TestPublish121IdempotentUploaded(t *testing.T) {
	// book.SourceMetadata["websiteSubmitStatus"]="uploaded" 时：
	// Submit 应直接返回 status=uploaded 且 connector 的上传/提交计数为 0。
}

func TestPublish121PendingConfirmReadbackFirst(t *testing.T) {
	// websiteSubmitStatus=upload_pending_confirm + websiteSubmitHeads 已持久化：
	// 1) GetSubmission 返回存在 → 补写 uploaded，不再 UploadVideo/Submit；
	// 2) GetSubmission 返回不存在 → 用已持久化 heads 直接重新 Submit，不重新 UploadVideo。
}
```

（实现时按现有 fake 结构补全断言；两条用例都必须先失败。）

- [ ] **Step 2: 实现 Submit 三态**

文件头部（import 后）加：

```go
// errPublishUnconfirmed 表示视频与正文已上传，但回读确认失败。调用方应把该
// 书标记为“已上传待确认”，重试时优先回读而不是无条件重传。
var errPublishUnconfirmed = errors.New("视频管理系统已上传但回读确认失败")
```

`Submit` 在取到 `metadata`（L160 后）加短路：

```go
submitStatus := strings.TrimSpace(toString(metadata["websiteSubmitStatus"]))
if submitStatus == "uploaded" && !automationBoolInput(input["force"]) {
	return map[string]any{"status": "uploaded", "bookId": book.BookID, "internalBookId": book.ID, "skipped": true}, nil
}
pendingHeads := []string{}
if submitStatus == "upload_pending_confirm" {
	if raw, ok := metadata["websiteSubmitHeads"].([]any); ok {
		for _, item := range raw { pendingHeads = append(pendingHeads, fmt.Sprint(item)) }
	}
	confirmed, readErr := p.Connector.GetSubmission(ctx, targetSession, book.BookID)
	if readErr == nil {
		// 补写 uploaded 并返回（复用下方成功持久化块，抽小函数 persistUploaded）
		...
	}
	// 回读失败（网络）→ 返回哨兵；确认不存在 → 走下方重新提交但跳过上传
	if !isSubmissionNotFound(readErr) {
		return nil, fmt.Errorf("%w：%v", errPublishUnconfirmed, readErr)
	}
}
```

（`targetSession` 构造上移到短路之前；`isSubmissionNotFound` 按 connector 现有“不存在”错误判定，若无现成判定则用错误消息匹配并注明。）

上传成功后、`Connector.Submit` 之前持久化：

```go
pendingMetadata := clonePublishMetadata(metadata)
pendingMetadata["websiteSubmitStatus"] = "upload_pending_confirm"
pendingMetadata["websiteSubmitHeads"] = heads
if _, err := p.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: pendingMetadata, ExpectedRevision: book.Revision}); err != nil {
	return nil, fmt.Errorf("视频管理系统上传进度保存失败：%w", err)
}
book.Revision++ // 本地 revision 跟进，避免下方成功块冲突
```

回读失败处（L270-273）改为：

```go
confirmed, err := p.Connector.GetSubmission(ctx, targetSession, book.BookID)
if err != nil {
	return nil, fmt.Errorf("%w：%v", errPublishUnconfirmed, err)
}
```

`upload_pending_confirm` 重入且确认不存在时：若 `len(pendingHeads) == len(selectedJobs)`，跳过 `UploadVideo` 循环直接用 `pendingHeads`；否则才完整重传。

- [ ] **Step 3: automation 侧识别哨兵**

`run()` 内 video 模式发布调用处（`c.publishVideo(...)` 的 fail 点），在 `fail("submit", ...)` / `fail("publish", ...)` 前加：

```go
if errors.Is(err, errPublishUnconfirmed) {
	entry["status"], entry["stage"], entry["error"], entry["message"], entry["updatedAt"] = "failed", "upload_confirm", err.Error(), "已上传待确认：回读失败，重试时将优先回读", c.now()
	appendAutomationBookLog(entry, c.now(), "upload_confirm", entry["message"].(string))
	failed.Add(1)
	persist()
	continue
}
```

（`publishVideo` 透传 err，不包裹即天然 `errors.Is` 可达；若现有代码有 `fmt.Errorf` 包裹，确认用 `%w`。）

- [ ] **Step 4: 前端识别待确认态**

`batchFactoryNovelListHelpers.js` L597：

```js
export function isBookUploadedTo121(book) {
  return ['uploaded', 'upload_pending_confirm'].includes(
    String(book?.sourceMetadata?.websiteSubmitStatus || '').trim().toLowerCase());
}
```

`batchFactoryBookState.js` `websiteSubmitLabel`/`uploadStatus` 的数组里把 `upload_pending_confirm` 归入「待回读」/running 组。

- [ ] **Step 5: 测试 + 编译 + Commit**

Run: `go build ./... ; go test ./backend/internal/httpapi/ -run 'TestPublish121' -v ; cd frontend && npx vitest run src/user/pages/shuihuo/batchFactoryBookState.test.js`
Expected: PASS。

```bash
git add backend/internal/httpapi/batch_factory_v11_publish_121.go backend/internal/httpapi/batch_factory_v11_publish_121_test.go backend/internal/httpapi/batch_factory_automation.go frontend/src/user/pages/shuihuo/batchFactoryNovelListHelpers.js frontend/src/user/pages/shuihuo/batchFactoryBookState.js
git commit -m "fix(publish): 视频管理系统发布三态幂等，重试优先回读不再重复上传"
```

---

### Task 4: A4 漫画缺页补图轮

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`（整本生成函数 L616-712）
- Test: `backend/internal/httpapi/batch_factory_comic_pages_test.go`（若无则新建，参考 `batch_factory_comic_generation_concurrency_test.go` 风格）

**Interfaces:**
- Produces:
  - `func missingComicPageWorkItems(generationVideos []batchfactoryv11.Video, pages []batchfactoryv11.ComicPage) []comicPageWorkItem`
  - `var comicPageRefillDelay = 60 * time.Second`（可变量，测试可改小）。

- [ ] **Step 1: 写失败测试**

```go
func TestMissingComicPageWorkItems(t *testing.T) {
	videos := []batchfactoryv11.Video{{ID: "v1"}, {ID: "v2"}, {ID: "v3"}}
	pages := []batchfactoryv11.ComicPage{{VideoID: "v1", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}}}
	work := missingComicPageWorkItems(videos, pages)
	if len(work) != 2 || work[0].page.ID != "v2" || work[1].page.ID != "v3" {
		t.Fatalf("unexpected missing work: %+v", work)
	}
}
```

（`ComicPageImage` 类型名以 `batchfactoryv11` 实际定义为准，先 `Grep "IsPrimary" backend/internal/batchfactoryv11` 核对。）

- [ ] **Step 2: 实现**

在整本生成函数中，首轮 `runComicPageWork` 结束且 `err == nil` 后、合集创建前插入：

```go
existingPages, err = comicStore.ListComicPages(ctx, owner, batchID, bookID)
if err != nil {
	return batchfactoryv11.ComicCollection{}, err
}
missingWork := missingComicPageWorkItems(generationVideos, existingPages)
if len(missingWork) > 0 {
	if progress != nil {
		progress("comic_images", fmt.Sprintf("有 %d 页因在途锁暂未生成，等待 60 秒后自动补齐", len(missingWork)))
	}
	timer := time.NewTimer(comicPageRefillDelay)
	select {
	case <-ctx.Done():
		timer.Stop()
		return batchfactoryv11.ComicCollection{}, ctx.Err()
	case <-timer.C:
	}
	err = runComicPageWork(ctx, missingWork, pageWorkers, func(generateCtx context.Context, item comicPageWorkItem) error {
		_, _, generateErr := g.generatePage(generateCtx, owner, batchID, bookID, item.page.ID, batchfactoryv11.UpsertComicPageInput{
			ComicPrompt: item.page.ComicPrompt, ImagePrompt: item.page.ComicImagePrompt, Provider: provider,
			ModelID: modelID, AspectRatio: aspectRatio,
		})
		if generateErr != nil && (errors.Is(generateErr, errShuihuoImageContentPolicy) || comicPageGenerationInFlightError(generateErr)) {
			return generateErr
		}
		if generateErr != nil {
			return fmt.Errorf("%s 补图失败: %w", item.page.Label, generateErr)
		}
		return nil
	})
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
	existingPages, err = comicStore.ListComicPages(ctx, owner, batchID, bookID)
	if err != nil {
		return batchfactoryv11.ComicCollection{}, err
	}
}
```

注意：原 L697-700 的 re-read 被本次插入合并（不重复读）。合集缺页报错处（L709-711）补充页码：

```go
collection, missing, err := comicStore.CreateComicCollection(ctx, owner, batchID, bookID, requestedVideoIDs)
if err != nil && len(missing) > 0 {
	return batchfactoryv11.ComicCollection{}, fmt.Errorf("%w: %d 个漫画页缺少主图（缺页序号 %v）", err, len(missing), missing)
}
```

- [ ] **Step 3: 测试 + 编译 + Commit**

Run: `go build ./... ; go test ./backend/internal/httpapi/ -run 'TestMissingComicPageWorkItems|TestRunComicPageWork' -v`

```bash
git add backend/internal/httpapi/batch_factory_comic_pages.go backend/internal/httpapi/batch_factory_comic_pages_test.go
git commit -m "fix(comic): 整本出图内增加缺页补图轮，单页在途锁不再整本失败"
```

---

### Task 5: B1 阶段契约单一来源

**Files:**
- Create: `frontend/src/user/pages/shuihuo/automationStages.json`
- Create: `backend/internal/httpapi/batch_factory_automation_stages.go`
- Test: `backend/internal/httpapi/batch_factory_automation_stages_test.go`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryBookState.js`（timeline 标签改从契约取，带兜底）

**Interfaces:**
- Produces:
  - JSON 结构：`{"stages":[{"id":"source","label":"原文已就绪","modes":["video","comic"],"terminal":false}, ...]}`
  - Go: `var automationStageContract = []automationStageDefinition{...}`（同字段）；测试跨目录读 JSON 比对。

- [ ] **Step 1: 写契约 JSON**

以 `batchFactoryBookState.js` 现有两条 timeline 为准提取阶段（video 线：source/assets/director/video/upload/complete；comic 线：source/assets/director/video/complete），外加自动化运行阶段（preflight/fetch_original/comic_assets/comic_bootstrap/comic_images/comic_publish/submit/published/upload_confirm）。每个阶段含 `id/label/modes/terminal`。

- [ ] **Step 2: 写 Go 常量 + 失败测试**

```go
func TestAutomationStagesContract(t *testing.T) {
	raw, err := os.ReadFile("../../../frontend/src/user/pages/shuihuo/automationStages.json")
	// 解析后与 automationStageContract 逐项比对 id/label/modes/terminal，不一致即 t.Errorf
}
```

Run: `go test ./backend/internal/httpapi/ -run TestAutomationStagesContract -v` → 先 FAIL（常量未定义），再实现 `batch_factory_automation_stages.go` 后 PASS。

- [ ] **Step 3: 前端接契约（带兜底）**

`batchFactoryBookState.js` 顶部：

```js
import automationStagesContract from './automationStages.json';

const automationStageLabels = (() => {
  try {
    const map = {};
    for (const stage of automationStagesContract?.stages || []) map[stage.id] = stage.label;
    return map;
  } catch (_) { return {}; }
})();
```

timeline 里 `label:` 改为 `automationStageLabels.<id> || '<原硬编码>'`（保留兜底文案不变）。

- [ ] **Step 4: 测试 + Commit**

Run: `go build ./... ; go test ./backend/internal/httpapi/ -run TestAutomationStagesContract -v ; cd frontend && npx vitest run src/user/pages/shuihuo/batchFactoryBookState.test.js`

```bash
git add frontend/src/user/pages/shuihuo/automationStages.json backend/internal/httpapi/batch_factory_automation_stages.go backend/internal/httpapi/batch_factory_automation_stages_test.go frontend/src/user/pages/shuihuo/batchFactoryBookState.js
git commit -m "refactor(automation): 阶段定义单一来源契约 JSON，Go 测试跨端守护"
```

---

### Task 6: B2 状态存储改增量

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_automation.go`（`loadLocked` L370-394、`saveLocked` L396-412、`setJob` L414-423、`appendAutomationBookLog` 加 200 条上限）
- Test: `backend/internal/httpapi/batch_factory_automation_test.go`（追加迁移测试）

**Interfaces:**
- 保持内存结构 `automationStateFile` 不变（所有调用方零改动）；只改持久层：
  - 新布局：`<dir>/automation/index.json` + `<dir>/automation/jobs/<jobId>.json`
  - 旧文件 `c.path` 启动时迁移，成功后改名 `.migrated`。
- Produces: `func (c *BatchAutomationCoordinator) automationStorePaths() (indexPath, jobsDir string)`。

- [ ] **Step 1: 写失败迁移测试**

```go
func TestAutomationStateStoreMigration(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, "automation-state.json")
	// 写含 2 个 job 的旧格式文件 → NewBatchAutomationCoordinator(... path=legacy ...)
	// loadLocked 后：automation/index.json 与 automation/jobs/<id>.json 存在，
	// 旧文件已改名 .migrated，重建的 state 内容与原文件等价。
}
```

- [ ] **Step 2: 实现持久层**

- `loadLocked`：先读新布局（index + 遍历 jobsDir 重组 `state.Jobs`）；新布局不存在且旧文件存在 → 读旧文件、拆分写出、`os.Rename(c.path, c.path+".migrated")`；迁移写失败仅记录、继续用旧数据。
- `setJob`：tmp+rename 写 `jobs/<jobId>.json`；仅当 job `state`/`updatedAt` 变化才重写 `index.json`（索引存 `{owner,batchId,state,updatedAt}`）。
- `saveLocked` 保留给全量写场景（迁移用），改为写新布局。
- `appendAutomationBookLog`：追加后 `if len(logs) > 200 { logs = logs[len(logs)-200:] }`（先 Grep 确认 logs 字段名）。

- [ ] **Step 3: 测试 + 编译 + Commit**

Run: `go build ./... ; go test ./backend/internal/httpapi/ -run 'TestAutomation' -v`

```bash
git add backend/internal/httpapi/batch_factory_automation.go backend/internal/httpapi/batch_factory_automation_test.go
git commit -m "refactor(automation): 状态文件拆分索引+单任务文件，日志上限 200 条"
```

---

### Task 7: B3 工作台常显状态行

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`（工作台工具栏区域）
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryAutomationStatusLine.source.test.js`（新建）

**Interfaces:**
- Consumes: 现有 `automationStatus`（来自 `loadRuntimeStatus` 的 automation 快照）、Task 5 的契约标签。
- Produces: JSX 元素 `data-testid="automation-status-line"`。

- [ ] **Step 1: 写失败 source test**

```js
it('工具栏含常显状态行', () => {
  expect(source).toContain('data-testid="automation-status-line"');
  expect(source).toMatch(/automationStatus\?\.(state|books)/);
});
```

- [ ] **Step 2: 实现**

在工具栏（自动化按钮所在行，Grep `automationStartBlockedLabel` 附近的 JSX）下方加：

```jsx
<div data-testid="automation-status-line" className="batch-factory-automation-status-line">
  {automationStatus?.state
    ? `自动化：${({ running: '运行中', paused: '已暂停', scheduled: '定时待执行', needs_attention: '待处理', stopped: '已停止', completed: '已完成', failed: '失败' })[String(automationStatus.state).toLowerCase()] || automationStatus.state}`
      + (Number(automationStatus?.counts?.failed || 0) + Number(automationStatus?.counts?.blocked || 0) > 0
        ? `，失败/阻塞 ${Number(automationStatus?.counts?.failed || 0) + Number(automationStatus?.counts?.blocked || 0)} 本` : '')
      + (automationStatus?.lastError ? `，最近错误：${automationStatus.lastError}` : '')
    : '自动化：未启动'}
</div>
```

样式复用现有 class 约定，不新增 css 文件改动（内联或沿用工具栏 class）。

- [ ] **Step 3: 测试 + Commit**

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/BatchFactoryAutomationStatusLine.source.test.js
git commit -m "feat(workbench): 工具栏常显自动化状态行，在跑/卡点/最近错误一眼可见"
```

---

### Task 8: C1 大文件拆分

**Files:**
- Split: `backend/internal/httpapi/shuihuo_projects.go`（4557 行）→ 保留路由与通用 + `shuihuo_projects_comic.go`、`shuihuo_projects_publish.go`、`shuihuo_projects_assets.go`
- Split: `backend/internal/httpapi/novel_fetch_workshop.go`（3614 行）→ 保留抓取 + `novel_fetch_workshop_classify.go`（分类/工作台操作）

**Steps:**
- [ ] **Step 1:** `Grep "^func " shuihuo_projects.go` 导出函数清单，按名称前缀归类（comic/publish/assets/其它），确认无跨文件循环依赖。
- [ ] **Step 2:** 用整块移动（不改函数体一字）把归类的函数搬到新文件，package 与 imports 保持一致（imports 按新文件实际引用精简）。
- [ ] **Step 3:** `go build ./... ; go test ./backend/internal/httpapi/` 全量通过；导出符号集合与拆前一致（`go doc` 对比）。
- [ ] **Step 4:** Commit：`refactor(httpapi): 拆分 shuihuo_projects 与 novel_fetch_workshop 大文件`。

---

### Task 9: C2 磁盘清理

- [ ] **Step 1:** 列目录抽样确认：
  - `Get-ChildItem data/local-executor-artifacts | Select-Object -First 10`（确认仅执行器中间产物）
  - `Grep -r "local-executor-artifacts" --include="*.go" backend/` 确认无运行时代码引用该路径写死（配置注入除外）。
  - `Get-ChildItem _local-archive-20261028/.direct-staging | Select-Object -First 10`（确认历史 Node 发布残留）。
- [ ] **Step 2:** 删除两目录（均在仓库运行数据/归档区，不删代码）。`data/` 若在 `.gitignore` 内则无提交；否则提交删除。
- [ ] **Step 3:** 若有提交：`chore: 清理执行器中间产物与历史 Node 发布残留（约 6.5GB）`。

---

### Task 10: C3 Git 分叉处理 + 最终验收

- [ ] **Step 1:** `git branch v88-backup-20261102` 建备份。
- [ ] **Step 2:** `git fetch origin`，`git log --oneline v88..origin/v88` 逐条核对落后提交是否已被本地覆盖。
- [ ] **Step 3:** `git merge origin/v88`（不用 rebase）；只在文档/锁文件冲突时手工解，源码冲突立即停止回报。
- [ ] **Step 4:** `git gc`。
- [ ] **Step 5:** 最终验收：`go build ./... && go test ./backend/internal/httpapi/ && cd frontend && npm run build`；随后 `scripts/deploy-go-embedded.ps1` 重建二进制，重启本地 4001，按设计文档 §8 手工验收三条（提交 loading 即时可见 / 缺视频模型的书 blocked 且原因可见 / 重复点击有提示）。
- [ ] **Step 6:** 验收通过后推送（推送前向用户确认）。

---

## Self-Review 记录

- Spec §4 A1-A4 → Task 1-4；§5 B1-B3 → Task 5-7；§6 C1-C3 → Task 8-10；§8 测试策略逐条落在各任务 Step；§9 验收标准映射到 Task 10 Step 5。
- 类型一致性：`automationBookPreflightReason` / `errPublishUnconfirmed` / `missingComicPageWorkItems` / `automationStageContract` 在 Produce 与使用处命名一致。
- Task 3 中 `isSubmissionNotFound` 依赖 connector 现有错误形态，实施时先 Grep `GetSubmission` 实现确认；若无法可靠区分“不存在/网络错误”，退化为「回读失败一律返回哨兵，重试走完整回读优先路径」，并在提交说明中记录。
