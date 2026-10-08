# V88 Creative Workflows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让首页、剧本、小说获取、改文工作台和小说面板以真实用户数据完成公网同等流程，并在刷新后从 MySQL 恢复。

**Architecture:** 复用 `intake`、`workshop`、`generation`、`novelpanel` 与 batch project 服务。新增的聚合只做读模型，不复制事实；需要保存自由文本草稿时使用带用户/团队归属的 additive MySQL 表。每个页面先稳定数据合同，再迁移隔离 CSS。

**Tech Stack:** Go/MySQL、React/AntD、Vitest、Go sqlmock/integration tests、Playwright。

## Global Constraints

- 先完成基础计划 Gate 1；不得在页面直接 `fetch` 或保存业务数据到浏览器。
- 公网只读对照；每页记录 DOM、计算样式、按钮和状态证据，不依赖旧 Node 接口。
- `/novel-fetch`、`/novel-fetch-workshop`、`/novel-panel` 是独立页面；`/novel-panel` 的项目归属必须由 Go 校验。

---

## Task 1: Home Recent-Work Projection

**Files:**
- Modify: `api/internal/intake/service.go`
- Modify: `api/internal/intake/mysql_store.go`
- Modify: `api/internal/intake/service_test.go`
- Create: `api/internal/httpapi/home_handlers.go`
- Create: `api/internal/httpapi/home_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`
- Modify: `前台/src/api.js`
- Modify: `前台/src/HomePage.jsx`
- Modify: `前台/src/HomePage.test.jsx`
- Modify: `前台/src/home.css`

- [ ] 用测试定义 `GET /api/v1/workspace/recent?limit=6`：仅返回当前用户可见的最近 intake/batch/script/novel-panel/tts 投影，含 `kind,id,title,status,updatedAt,href`。
- [ ] 测试团队隔离、稳定排序、无数据空态和 limit 上限；禁止用静态假卡片代替项目。
- [ ] 先运行 `cd api && go test ./internal/intake ./internal/httpapi -run 'Home|Recent'`，确认失败。
- [ ] 在 store 用 `UNION ALL`/明确排序构建只读投影，不创建第二套 recent 表。
- [ ] 在前台通过 `api.js` 加载，覆盖 loading、空态、错误、刷新和进入项目。
- [ ] 用公网审计值调整 `.home-page` 范围内 Hero、六入口卡片、视频和响应式样式。
- [ ] 运行 Go/前台定向测试并提交：`feat(home): project real recent workspace activity [skip ci]`。

## Task 2: Durable Script Draft and Generation Flow

**Files:**
- Create: `api/db/migrations/00019_script_workspaces.sql`
- Create: `api/internal/scriptworkspace/model.go`
- Create: `api/internal/scriptworkspace/mysql_store.go`
- Create: `api/internal/scriptworkspace/service.go`
- Create: `api/internal/scriptworkspace/service_test.go`
- Create: `api/internal/httpapi/script_workspace_handlers.go`
- Create: `api/internal/httpapi/script_workspace_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`
- Modify: `前台/src/api.js`
- Modify: `前台/src/ScriptWorkspace.jsx`
- Modify: `前台/src/ScriptWorkspace.test.jsx`
- Modify: `前台/src/script-workspace.css`

- [ ] migration contract test 定义 `script_workspaces`：`id,user_id,team_id,title,source_text,mode,duration,status,revision,created_at,updated_at,deleted_at`，并为 `(user_id,updated_at)` 建索引。
- [ ] service 测试定义创建、自动保存、乐观 revision 冲突、软删除/恢复、当前用户与团队隔离；系统 Prompt 只从后端 prompt 服务取得。
- [ ] handler 测试定义列表/详情/保存/生成 API，以及 Cookie/CSRF/capability/ownership 的 401/403。
- [ ] 前台测试覆盖公网自由文本入口、人物/场景提取、模式/时长、历史/导出、刷新恢复和 `executor_unavailable`。
- [ ] 实现最小垂直链路；生成动作复用现有 generation/runtime，不创建页面专属 Worker。
- [ ] 运行 `go test ./internal/scriptworkspace ./internal/httpapi` 与 `CI=1 npm test -- --run src/ScriptWorkspace.test.jsx`。
- [ ] 提交：`feat(script): migrate durable public script workflow [skip ci]`。

## Task 3: Novel Fetch Public Parity

**Files:**
- Modify: `api/internal/intake/service_test.go`
- Modify: `api/internal/intake/service.go`
- Modify: `api/internal/httpapi/intake_handlers_test.go`
- Modify: `前台/src/NovelFetchPage.jsx`
- Modify: `前台/src/NovelFetchPage.test.jsx`
- Modify: `前台/src/novel-fetch.css`

- [ ] 测试书城/书籍输入、正文与 `bookinfo.category` 解析、分页章节、保存 intake、恢复已保存结果和真实 Provider 不可用。
- [ ] 明确性别映射优先级：手工值 > `bookinfo.category` > 已验证 genre > AI 兜底；不得被 AI 覆盖。
- [ ] 修复服务缺口后，通过统一 API 客户端渲染 loading/空态/错误/重试/保存成功。
- [ ] 只迁移 `.novel-fetch-page` 范围内公网表单、卡片、图标、间距和移动端规则。
- [ ] 运行 `go test ./internal/intake ./internal/httpapi -run 'Intake|NovelFetch'` 和页面测试。
- [ ] 提交：`feat(novel-fetch): complete public intake workflow [skip ci]`。

## Task 4: Workshop and Novel Panel Recovery

**Files:**
- Modify: `api/internal/workshop/service_test.go`
- Modify: `api/internal/workshop/service.go`
- Modify: `api/internal/novelpanel/service_test.go`
- Modify: `api/internal/novelpanel/service.go`
- Modify: `api/internal/httpapi/workshop_handlers_test.go`
- Modify: `api/internal/httpapi/novel_panel_handlers.go`
- Modify: `api/internal/httpapi/novel_panel_handlers_test.go`
- Modify: `前台/src/NovelFetchWorkshop.jsx`
- Modify: `前台/src/NovelFetchWorkshop.test.jsx`
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.jsx`
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.test.jsx`
- Modify: `前台/src/novel-panel/NovelPanelWorkbench.css`

- [ ] Workshop 测试覆盖设置保存、表格选择、进入批量项目、刷新恢复和跨用户 403。
- [ ] Novel Panel 测试覆盖 projectId 深链、缺失 projectId 的可操作引导、乐观 revision、历史恢复和归属隔离。
- [ ] 错误态不得直接显示 Go/Provider 原文；显示安全消息、request id 和重试。
- [ ] 迁移公网抽屉、表格、空态和响应式 CSS，保持两个页面 CSS 隔离。
- [ ] 运行相关 Go/前台测试并提交：`feat(novel-panel): complete workshop recovery flow [skip ci]`。

## Task 5: Creative Workflow Acceptance

- [ ] 运行 `cd api && go test ./...`。
- [ ] 运行前台完整安装、Vitest 和 build；运行后台完整 Vitest/build 确认共享底座未回归。
- [ ] 构建 Go embed，使用隔离数据库登录后逐页验证 `/`、`/script`、`/novel-fetch`、`/novel-fetch-workshop`、`/novel-panel?projectId=<owned>`。
- [ ] 验证刷新恢复、跨用户 403、CSRF、空态、Provider unavailable、浏览器无业务 localStorage。
- [ ] 更新 `TASKS.md` 只勾选已完成项；提交 `chore(parity): record creative workflow acceptance [skip ci]` 并确认 `HEAD == origin/main`。
