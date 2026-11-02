# Go 小说获取全流程修复实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让本地 `4001/novel-fetch` 使用纯 Go 后端完成可恢复的异步处理，并与可用基准页保持一致的任务状态、正文和提交语义。

**Architecture:** MySQL 保存批次、书籍阶段状态和提交回执；Redis 保存待执行队列、短期进度和分布式租约；TOS 保存完整原文、处理正文和 AI 版本。React + Ant Design 用户前台只消费 Go API，不增加 Node 路由或旧静态兼容实现。

**Tech Stack:** Go、MySQL、Redis、TOS、React、Ant Design、Go embed。

## Global Constraints

- 正式源码只改 `v88`，保留所有既有未提交改动。
- 所有新后端业务逻辑必须是 Go；不新增 Node API、Node 队列或 Node 发布入口。
- MySQL 是任务事实来源；Redis 只保存队列、租约、锁和短期进度。
- 完整原文必须保留；`maxTxt` 只限制处理文本。
- 已确认提交的版本不得重复提交。
- 本次不触发真实外部提交，不发布公网，不推送远程。

---

### Task 1: 修复模型选择与任务状态归一化

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/app/app_test.go`
- Create: `backend/internal/novelfetchworkshop/status.go`
- Create: `backend/internal/novelfetchworkshop/status_test.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`

**Interfaces:**
- Produces: `selectedTextModelID(settings map[string]any) string` 同时读取 snake_case 和 camelCase。
- Produces: `novelfetchworkshop.DeriveTaskStatus(Document) string` 返回 `input_ready`、`processing`、`pending_submit`、`completed` 或 `failed`。

- [ ] 写失败测试，证明 `text_model_id` 当前无法选中模型，已完成任务仍显示待处理。
- [ ] 运行定向测试并确认按预期失败。
- [ ] 实现最小兼容读取和统一状态推导。
- [ ] 运行定向测试并确认通过。

### Task 2: 修复完整原文与处理字数边界

**Files:**
- Modify: `backend/internal/novelfetchworkshop/upstream.go`
- Modify: `backend/internal/novelfetchworkshop/upstream_test.go`

**Interfaces:**
- Produces: 上游抓取不使用用户的处理上限截断正文；`ProcessRawText` 仍按 `maxTxt` 截取处理副本。

- [ ] 写失败测试，断言完整抓取请求与本地处理上限分离。
- [ ] 运行测试确认失败。
- [ ] 实现最小修复并保留上游协议兼容。
- [ ] 运行测试确认通过。

### Task 3: 建立 MySQL 事实 + Redis 队列的 Go 异步执行

**Files:**
- Create: `backend/internal/novelfetchworkshop/queue.go`
- Create: `backend/internal/novelfetchworkshop/redis_queue.go`
- Create: `backend/internal/novelfetchworkshop/queue_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop_test.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**
- Produces: `Queue.Enqueue`, `Queue.Claim`, `Queue.Ack`, `Queue.Fail`。
- Produces: `/process/start` 立即返回 `202` 与持久化批次 ID；Go worker 异步更新 MySQL 文档和运行记录。

- [ ] 写失败测试，证明启动接口当前阻塞且运行记录因缺少 `BookID` 无法持久化。
- [ ] 写 Redis 队列契约测试，覆盖租约、确认和失败重试。
- [ ] 运行测试确认失败。
- [ ] 实现 Go 队列和 worker，Redis 不可用时返回明确不可用状态。
- [ ] 运行定向测试与 Go 全量测试。

### Task 4: 接入 TOS 正文存储边界

**Files:**
- Create: `backend/internal/novelfetchworkshop/object_body_store.go`
- Create: `backend/internal/novelfetchworkshop/object_body_store_test.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**
- Produces: `BodyObjectStore.Put/Get/Delete`；MySQL `BodyRef` 保存对象键、哈希、字数和状态。
- Consumes: 现有 TOS SDK 配置与客户端。

- [ ] 写失败测试，证明正文写入对象存储后仍可从 `Store` 透明读取。
- [ ] 运行测试确认失败。
- [ ] 实现 TOS 适配与本地测试替身；生产配置缺失时明确报告不可用。
- [ ] 运行定向测试确认通过。

### Task 5: React 用户前台状态对齐与本地验收

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/shared/api/novelFetchWorkshop.js`
- Modify: `frontend/src/shared/api/novelFetchWorkshop.test.js`
- Create: `frontend/src/user/pages/novel-fetch-status.test.js`

**Interfaces:**
- Consumes: Go API 的统一 `status`、阶段状态、真实字数、批次进度和失败原因。
- Produces: 与基准页一致的“处理中 / 排队中 / 待提交 / 完成 / 失败原因”显示。

- [ ] 写失败测试，复现 `done + ai_done + submitted` 被显示成“待处理”。
- [ ] 运行测试确认失败。
- [ ] 实现统一状态展示并移除旧字段优先级。
- [ ] 运行前端测试、生产构建和 Go embed 构建。
- [ ] 启动本地 Go 服务，浏览器验证页面、轮询、刷新后状态一致；不执行真实外部提交。

