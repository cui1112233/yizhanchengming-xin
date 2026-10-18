# 批量工厂漫画模式实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不影响现有视频流程的前提下，使用现有下拉配置新增一条漫画资产预设、三条漫画约束预设与逐页漫画提示词直接出图。

**Architecture:** 漫画选择写入 `aiPromptConfig.comic`，漫画资产以 `source=comic` 保存，与 `source=director/manual` 的视频资产隔离。漫画提示词写入视频级 `comicPrompt`，并由独立漫画阶段生成；图片生成沿用已配置的图片模型和本地成品存储。

**Tech Stack:** Go、MySQL、React、Ant Design、Node test、Go test。

## 全局约束

- 不修改或伪造已执行的 MySQL 迁移记录。
- 不改变既有 `visualPrompt`、视频资产或视频生产行为。
- 提示词正文仅由后端解析已发布预设后使用，浏览器只保存预设标识。
- 漫画页由图片模型一次直接生成，不做后期拼图或文字叠加。

---

### Task 1: 漫画配置与资产隔离

**Files:**
- Modify: `backend/internal/batchfactoryv11/director_contract.go`
- Modify: `backend/internal/batchfactoryv11/book_assets.go`
- Modify: `backend/internal/batchfactoryv11/book_assets_mysql.go`
- Modify: `backend/internal/batchfactoryv11/director_service.go`
- Test: `backend/internal/batchfactoryv11/director_service_test.go`

**Interfaces:**
- Produces `AIReasoningPromptConfig.Comic`，包含 `assets`、`constraints` 和 `prompt` 三项预设快照。
- Produces `RunComicAssetExtraction(ctx, owner, batchID, bookID)`，只替换 `source=comic` 的自动资产。

- [ ] **Step 1: 写入失败测试**

新增测试：漫画资产提取只写入 `source=comic` 的人物、场景、道具，不删除原有 `source=director` 资产。

- [ ] **Step 2: 运行失败测试**

Run: `go test ./internal/batchfactoryv11 -run TestComicAssetExtraction -count=1`

Expected: FAIL，因为漫画提取入口和独立来源尚不存在。

- [ ] **Step 3: 最小实现**

解析 `aiPromptConfig.comic.assets`，调用与现有资产提取相同的文本模型协议，保存漫画资产时固定使用 `source=comic`；视频资产读取继续排除漫画来源。

- [ ] **Step 4: 运行通过测试**

Run: `go test ./internal/batchfactoryv11 -run TestComicAssetExtraction -count=1`

Expected: PASS。

### Task 2: 漫画提示词与漫画页生成

**Files:**
- Create: `backend/internal/batchfactoryv11/comic_prompt_service.go`
- Modify: `backend/internal/batchfactoryv11/types.go`
- Modify: `backend/internal/batchfactoryv11/prompt_override.go`
- Modify: `backend/internal/batchfactoryv11/stage_runs.go`
- Modify: `backend/internal/batchfactoryv11/stage_service.go`
- Test: `backend/internal/batchfactoryv11/comic_prompt_service_test.go`

**Interfaces:**
- Produces `Video.ComicPrompt string` and `BookStageComic`.
- `RunComicPromptExtraction` reads only漫画资产、漫画画质约束、漫画画面限制、漫画负面提示词。

- [ ] **Step 1: 写入失败测试**

测试一条漫画提示词包含漫画资产、三条漫画约束、完整 `3:4` 页面规则，且不含视频时长或运镜规则。

- [ ] **Step 2: 运行失败测试**

Run: `go test ./internal/batchfactoryv11 -run TestComicPrompt -count=1`

Expected: FAIL，因为漫画提示词字段和生成合约尚不存在。

- [ ] **Step 3: 最小实现**

新增漫画提示词字段与阶段。每个现有分镜只生成一张完整漫画页提示词，并保存到该分镜独立设置中。图片阶段接收漫画提示词，不调用视频生产。

- [ ] **Step 4: 运行通过测试**

Run: `go test ./internal/batchfactoryv11 -run TestComicPrompt -count=1`

Expected: PASS。

### Task 3: API 和现有下拉框接入

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_v11_stages.go`
- Modify: `backend/internal/httpapi/batch_factory_v12_image.go`
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryAiReasoningModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

**Interfaces:**
- `comicAssets`、`comic` 与 `comicImage` 阶段使用现有书级阶段 API。
- 下拉框新增漫画资产、漫画画质约束、漫画画面限制、漫画负面提示词和漫画提示词选择，不新增模式切换。

- [ ] **Step 1: 写入失败测试**

断言配置抽屉显示一个“漫画资产提示词”下拉框、三条漫画约束下拉框和一条“漫画提示词”下拉框。

- [ ] **Step 2: 运行失败测试**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: FAIL，因为这些标签和漫画阶段尚不存在。

- [ ] **Step 3: 最小实现**

扩展已有配置读取、保存和书级阶段调用；所有漫画选项仍从现有系统预设目录读取，不把预设正文下发到浏览器。

- [ ] **Step 4: 运行通过测试**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: PASS。

### Task 4: 分镜弹窗与本地浏览器验证

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

**Interfaces:**
- 弹窗显示“视频提示词”和“漫画提示词”；漫画标签支持提取、编辑、保存和“生成漫画页”。

- [ ] **Step 1: 写入失败测试**

断言漫画标签和“生成漫画页”按钮存在，且视频标签继续调用原视频生产入口。

- [ ] **Step 2: 运行失败测试**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: FAIL，因为漫画标签和按钮尚不存在。

- [ ] **Step 3: 最小实现**

复用弹窗编辑、保存和图片模型调用路径；漫画操作只读取/写入 `comicPrompt` 与漫画成图结果。

- [ ] **Step 4: 回归验证**

Run: `go test ./internal/batchfactoryv11 -count=1` and `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: PASS。随后在 `http://localhost:4001/shuihuo-production` 检查视频提示词仍存在，并能看到漫画提示词入口。
