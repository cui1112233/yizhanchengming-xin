# V88 Production Workflows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把批量工厂、水货作品库、配音、历史和设置统一到现有 Go/MySQL/Redis 事实链，并对齐公网交互与真实失败语义。

**Architecture:** batch project 是生产聚合根；Shuihuo 和 TTS 复用其生成、媒体任务和状态事实。历史是现有事实的用户范围投影，不建第二套日志系统；设置保存在现有 workspace preferences/unified settings，执行器状态实时读取受控服务。

**Tech Stack:** Go/MySQL/Redis/TOS、React/AntD、Vitest、Go integration tests、Playwright。

## Global Constraints

- 先完成创作流程 Gate 2；不得恢复 V11/V12 双路由或 Node 兼容后端。
- 队列消费者必须复用 `taskruntime`；Redis 记录丢失后以 MySQL 状态恢复。
- Provider/TOS/执行器测试默认使用 fake；本机浏览器验收不得调用付费 Provider。

---

## Task 1: Batch Factory Project Contract

**Files:**
- Modify: `api/internal/intake/batch_project_detail_test.go`
- Modify: `api/internal/intake/mysql_store.go`
- Modify: `api/internal/httpapi/batch_project_handlers_test.go`
- Modify: `api/internal/httpapi/batch_project_handlers.go`
- Modify: `前台/src/BatchFactoryHome.jsx`
- Modify: `前台/src/BatchFactoryHome.test.jsx`
- Modify: `前台/src/BatchProjectListPage.jsx`
- Modify: `前台/src/BatchProjectListPage.test.jsx`

- [ ] 用测试锁定项目列表、详情、来源书籍、阶段状态、失败摘要、最近更新时间和 owned deep link。
- [ ] 覆盖跨用户/团队 403、删除/恢复语义、分页和空态；列表查询禁止 N+1。
- [ ] 前台补齐公网筛选/排序/进入项目/刷新/错误重试，所有动作走 `api.js`。
- [ ] 运行 `go test ./internal/intake ./internal/httpapi -run 'BatchProject'` 和两个页面测试。
- [ ] 提交：`feat(batch): complete owned project workspace [skip ci]`。

## Task 2: Shuihuo Works Library and Production Console

**Files:**
- Modify: `api/internal/shuihuo/shuihuo_test.go`
- Modify: `api/internal/shuihuo/shuihuo.go`
- Modify: `api/internal/httpapi/shuihuo_media_handlers_test.go`
- Modify: `api/internal/httpapi/shuihuo_media_handlers.go`
- Modify: `前台/src/ShuihuoProductionPage.jsx`
- Modify: `前台/src/ShuihuoProductionPage.test.jsx`
- Modify: `前台/src/shuihuo-production.css`
- Modify: `前台/src/shuihuo-media.css`

- [ ] 定义用户作品库投影：搜索、状态筛选、排序、分页、卡片/列表、打开项目和软删除；数据来自 batch/Shuihuo 事实。
- [ ] 测试生成片段、候选、媒体任务和失败状态的 ownership/capability/恢复，不新建 Shuihuo 历史事实表。
- [ ] 页面同时满足公网“作品库入口”和项目内“生产控制台”，通过路由 query/projectId 决定视图，不复用 Novel Fetch 根组件。
- [ ] 无执行器时展示 `executor_unavailable` 与原因；不显示成功 toast，不创建伪媒体。
- [ ] 运行 Go/前台定向测试并提交：`feat(shuihuo): migrate works library and console [skip ci]`。

## Task 3: TTS Public Workflow on Existing Media Tasks

**Files:**
- Modify: `api/internal/generation/service_test.go`
- Modify: `api/internal/generation/mysql_store_test.go`
- Modify: `api/internal/httpapi/generation_handlers_test.go`
- Modify: `前台/src/TtsPage.jsx`
- Modify: `前台/src/TtsPage.test.jsx`
- Modify: `前台/src/tts-history.css`

- [ ] 测试默认音色卡、上传/选择已归属文本、批量生成、单条重试、状态轮询、媒体元数据和刷新恢复。
- [ ] 上传对象必须先建 MySQL 归属记录，再写 TOS；TOS 未配置时安全失败且不得创建成功状态。
- [ ] 生成只进入现有 runtime/queue；Provider 未配置返回 `executor_unavailable`，不调用付费服务。
- [ ] 前台对齐公网卡片、进度、历史、空/加载/错误态；所有音频 URL 经过授权读取接口。
- [ ] 运行相关 Go/前台测试并提交：`feat(tts): complete durable voice workflow [skip ci]`。

## Task 4: Unified User History Projection

**Files:**
- Modify: `api/internal/httpapi/history_handlers_test.go`
- Modify: `api/internal/httpapi/history_handlers.go`
- Modify: `前台/src/HistoryPage.jsx`
- Modify: `前台/src/HistoryPage.test.jsx`

- [ ] 定义历史 DTO：`kind,projectId,entityId,title,status,safeError,requestId,updatedAt,href,restorable`。
- [ ] SQL 从 Script、Batch、Shuihuo、TTS/媒体任务等现有事实投影；测试用户/团队范围、搜索、筛选、稳定游标分页。
- [ ] 删除动作只调用所属事实的合法软删除/归档语义；不实现全局物理清空。
- [ ] 前台实现查看、进入项目、允许的删除/恢复、刷新、空态和脱敏错误。
- [ ] 运行 `go test ./internal/httpapi -run History` 与页面测试并提交：`feat(history): unify owned workflow projection [skip ci]`。

## Task 5: Server Settings and Executor Truth

**Files:**
- Modify: `api/internal/httpapi/workspace_settings_handlers_test.go`
- Modify: `api/internal/httpapi/workspace_settings_handlers.go`
- Modify: `api/internal/video/config_service.go`
- Modify: `api/internal/httpapi/local_executor_handlers_test.go`
- Modify: `前台/src/SettingsPage.jsx`
- Modify: `前台/src/SettingsPage.test.jsx`
- Modify: `前台/src/settings-page.css`

- [ ] 测试主题、提醒、存储偏好、保留期只保存允许字段；敏感 Provider/TOS/Redis/MySQL 值永不返回。
- [ ] 测试“已选择配置”与“实际生效配置”分别来自持久化偏好和运行时探测；执行器离线包含安全 reason、lastHeartbeat。
- [ ] 所有写入要求 Cookie/CSRF/capability；跨用户读取 403。
- [ ] 页面按公网信息结构显示主题、提醒、设备、本地执行器和存储偏好；无执行器明确说明原因。
- [ ] 运行 Go/前台测试并提交：`feat(settings): expose durable preferences and runtime truth [skip ci]`。

## Task 6: Production Workflow Acceptance

- [ ] 运行完整 Go、前台、后台测试和 build，取得最终退出码。
- [ ] Go embed 本机验收 `/batch-factory`、`/shuihuo-production`、`/tts`、`/history`、`/settings`。
- [ ] 使用隔离 Redis prefix 验证排队、租约失效和 MySQL 恢复；不连接生产 Redis/TOS/MySQL。
- [ ] 验证 TOS/Provider 未配置的真实失败，确认无伪成功、无浏览器业务事实。
- [ ] 更新 `TASKS.md` 的真实完成项，提交 `chore(parity): record production workflow acceptance [skip ci]` 并确认 `HEAD == origin/main`。
