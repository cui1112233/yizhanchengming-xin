# Novel Panel React Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 v88 中把小说面板从 iframe + 旧 workbench 逐步迁移为功能等价、可维护的 React 页面。

**Architecture:** 保留现有 Go API 和数据格式，先建立 React 侧的 API/状态边界，再按页面职责迁移项目、资产、剧本分镜、提示词、生成和历史功能。迁移期间保留旧 workbench 回退入口，完成回归后才切换默认入口并删除旧资源。

**Tech Stack:** React 18、Ant Design、Vite、Go HTTP API、Node test runner、Go test。

**Spec:** `docs/superpowers/specs/2026-10-15-novel-panel-react-migration-design.md`

## Global Constraints

- Git `v88` 是唯一维护主线；v78 只作只读参考。
- 后端继续使用当前 Go API，不新增 Node 后端依赖。
- 系统提示词由后端统一保存和渲染，前端不硬编码提示词正文。
- 当前工作区已有未提交改动，只修改本计划列出的文件，不覆盖无关改动。
- 删除旧 workbench 前必须确认无运行时引用，并保留可回滚提交。

## Review Focus

- 项目不存在、项目参数缺失或项目数据损坏时，页面必须显示可恢复错误，而不是白屏（Task 2）。
- 资产含有图片、没有图片、上传失败和旧资产格式时，预览和编辑状态必须稳定（Task 3）。
- 提示词生成返回空 JSON、非法 JSON、部分字段缺失或后端错误时，必须保留原内容并给出重试入口（Task 4）。
- 生成任务重复点击、取消、轮询超时和刷新页面时，不能重复提交或丢失状态（Task 5）。
- 旧 workbench 仍被路由、静态资源或 Go embed 引用时，不得删除资源（Task 6）。

### Task 1: 建立小说面板接口与功能基线

**Files:**
- Create: `frontend/src/user/pages/novel-panel/novelPanelContracts.js`
- Create: `frontend/src/user/pages/novel-panel/novelPanelContracts.test.js`
- Modify: `frontend/src/shared/api/novelPanel.js`
- Test: `backend/internal/httpapi/novel_panel_alias_test.go`, `backend/internal/httpapi/novel_panel_compat_test.go`

**Interfaces:**
- Consumes: 当前 `/api/novel-panel/*` Go 接口和现有 `apiRequest`。
- Produces: `normalizeNovelProject(value)`, `normalizeNovelAsset(value)`, `normalizeNovelPrompt(value)`，以及统一的错误对象 `{ code, message, retryable }`。

- [ ] **Step 1: 写失败测试**：覆盖项目、资产、提示词的最小字段归一化，以及错误响应保留 `retryable`。
- [ ] **Step 2: 运行测试确认失败**：`node --test frontend/src/user/pages/novel-panel/novelPanelContracts.test.js`。
- [ ] **Step 3: 实现归一化函数和 API 封装**：只做字段兼容，不改变后端数据结构。
- [ ] **Step 4: 运行前端和 Go 定向测试**：确认当前别名接口仍通过。
- [ ] **Step 5: 提交**：`feat: define novel panel react contracts`。

### Task 2: React 页面壳、项目加载和回退开关

**Files:**
- Create: `frontend/src/user/pages/novel-panel/NovelPanelReactPage.jsx`
- Create: `frontend/src/user/pages/novel-panel/NovelPanelReactPage.test.js`
- Create: `frontend/src/user/pages/novel-panel/novel-panel-react.css`
- Modify: `frontend/src/user/pages/NovelPanelPage.jsx`
- Modify: `frontend/src/user/App.jsx`

**Interfaces:**
- Consumes: Task 1 的归一化函数；`GET /api/novel-panel/projects/{projectId}`。
- Produces: `NovelPanelReactPage({ theme, projectId, onFallback })`，具备加载、错误、空项目和回退状态。

- [ ] **Step 1: 写失败测试**：缺少 project、项目不存在、加载成功和回退按钮四种页面状态。
- [ ] **Step 2: 运行测试确认失败**：`node --test frontend/src/user/pages/novel-panel/NovelPanelReactPage.test.js`。
- [ ] **Step 3: 实现页面壳**：保留主题同步、项目参数和错误边界；回退开关默认关闭，测试可打开。
- [ ] **Step 4: 运行测试和 `npm --prefix frontend run build`**。
- [ ] **Step 5: 提交**：`feat: add react novel panel shell`。

### Task 3: 资产、人物、场景和参考图模块

**Files:**
- Create: `frontend/src/user/pages/novel-panel/AssetWorkspace.jsx`
- Create: `frontend/src/user/pages/novel-panel/AssetWorkspace.test.js`
- Create: `frontend/src/user/pages/novel-panel/asset-workspace.css`
- Modify: `frontend/src/shared/api/novelPanel.js`

**Interfaces:**
- Consumes: `GET/POST /api/novel-panel/reference-assets/*`、项目中的人物/场景/道具数据。
- Produces: `AssetWorkspace({ project, onChange })`，输出带 `assetType`、`assetId`、`imageUrl`、`revision` 的资产状态。

- [ ] **Step 1: 写失败测试**：有图资产显示缩略图、无图资产显示占位、上传/生成失败可重试、保存后版本号更新。
- [ ] **Step 2: 运行测试确认失败**。
- [ ] **Step 3: 实现资产列表、编辑、上传、生成和主图切换**；不在组件内硬编码提示词。
- [ ] **Step 4: 运行前端测试和 Go 资产接口测试**。
- [ ] **Step 5: 提交**：`feat: migrate novel panel asset workspace`。

### Task 4: 故事剧本、分镜和提示词模块

**Files:**
- Create: `frontend/src/user/pages/novel-panel/StoryOutlineWorkspace.jsx`
- Create: `frontend/src/user/pages/novel-panel/StoryOutlineWorkspace.test.js`
- Create: `frontend/src/user/pages/novel-panel/story-outline-workspace.css`
- Modify: `frontend/src/shared/api/novelPanel.js`

**Interfaces:**
- Consumes: `/api/novel-panel/analyze`、`/outline-scenes`、`/regenerate-scene-outline`、`/generate-scene-prompts`。
- Produces: `StoryOutlineWorkspace({ project, constraints, onSave })`，维护原文、故事剧本、场景、镜头和最终提示词版本。

- [ ] **Step 1: 写失败测试**：显示分镜列表、编辑后未保存提示、生成中禁用重复提交、非法响应保留旧结果。
- [ ] **Step 2: 运行测试确认失败**。
- [ ] **Step 3: 实现故事/分镜/提示词编辑器**；统一通过 API 获取后端渲染后的提示词。
- [ ] **Step 4: 运行前端测试和 `go test ./backend/internal/httpapi/...`**。
- [ ] **Step 5: 提交**：`feat: migrate novel panel story and prompt workspace`。

### Task 5: 生成、重试、取消、历史和审计状态

**Files:**
- Create: `frontend/src/user/pages/novel-panel/GenerationPanel.jsx`
- Create: `frontend/src/user/pages/novel-panel/GenerationPanel.test.js`
- Create: `frontend/src/user/pages/novel-panel/HistoryPanel.jsx`
- Create: `frontend/src/user/pages/novel-panel/HistoryPanel.test.js`
- Modify: `frontend/src/shared/api/novelPanel.js`

**Interfaces:**
- Consumes: 当前生成、诊断、历史接口及 Task 1 的错误对象。
- Produces: 可取消的 `runGeneration()`、`retryGeneration()`、`loadHistory()` 状态机，状态包括 `idle/running/succeeded/failed/cancelled`。

- [ ] **Step 1: 写失败测试**：重复点击只产生一次请求；取消后不再更新旧任务；刷新可恢复历史状态；失败显示重试原因。
- [ ] **Step 2: 运行测试确认失败**。
- [ ] **Step 3: 实现生成状态机、轮询、取消、历史恢复和审计信息展示**。
- [ ] **Step 4: 运行前端测试、Go 接口测试和构建**。
- [ ] **Step 5: 提交**：`feat: add novel panel generation and history state`。

### Task 6: 入口切换、回归和旧 workbench 清理

**Files:**
- Modify: `frontend/src/user/pages/NovelPanelPage.jsx`
- Modify: `frontend/src/user/App.jsx`
- Modify: `backend/internal/web/static.go`
- Modify: `backend/internal/web/static_test.go`
- Delete only after verification: `frontend/dist/novel-panel/workbench/*` and related unused bridge files
- Test: `frontend/src/user/App.performance.test.js`, `backend/internal/web/static_test.go`

**Interfaces:**
- Consumes: Tasks 2-5 的 React 页面和状态模块。
- Produces: `/novel-panel` 默认进入 React 页面；旧 workbench 仅在明确回退开关下可访问，最终无运行时引用。

- [ ] **Step 1: 写入口回归测试**：默认路由使用 React 页面；回退开关仍能加载旧 workbench；Go 静态路由行为不误伤其他页面。
- [ ] **Step 2: 运行测试确认入口切换前后差异可观测**。
- [ ] **Step 3: 将默认入口切到 React，并保留一次可回退提交**。
- [ ] **Step 4: 全量运行 `npm --prefix frontend test`、`npm --prefix frontend run build`、`go test ./backend/...`；再做浏览器真实流程检查**。
- [ ] **Step 5: 代码搜索确认无旧 workbench 运行时引用后，单独提交删除旧资源**：`chore: remove legacy novel panel workbench`。

## Final Verification

- [ ] 确认所有改动都在 v88 Git 记录中，未修改公网文件。
- [ ] 记录构建产物和测试结果；失败时保留回退入口，不删除旧资源。
- [ ] 只有在 React 页面完成真实浏览器流程后，才允许进入旧 workbench 删除步骤。
