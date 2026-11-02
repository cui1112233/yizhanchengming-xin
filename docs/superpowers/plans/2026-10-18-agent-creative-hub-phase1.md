# V88 Agent 智能创作中枢第一阶段实施计划

日期：2026-10-18  
范围：本地 V88，仅第一阶段，不发布公网  
目标：把现有 Agent 首页和 React Flow 画布接到正式 MySQL 数据源，完成项目切换、节点引用、画布持久化和并发冲突保护。

## 执行边界

- 直接在当前 `v88` 工作区完成，不新建分支，不拉取或覆盖远程。
- 只修改 Agent、数据库迁移、路由注入和对应测试文件，不清理现有脏工作区。
- 不修改任何已执行迁移；新增全局唯一且高于现有编号的 `8800001` 追加迁移。
- 旧 `agent-tasks.json` 只读、按账号幂等导入，导入后接口只读写 MySQL，不双写、不删除旧文件。
- 本阶段不实现智能工具执行、媒体生产和上传，这些仍按六阶段设计在后续阶段完成。

## 任务 1：Agent MySQL 表结构

文件：

- 新建 `backend/internal/storage/agent_workspace_schema.go`
- 新建 `backend/internal/storage/agent_workspace_schema_test.go`
- 修改 `backend/internal/storage/local_executor_schema.go`
- 修改 `backend/internal/storage/migrations_test.go`

步骤：

1. 先写迁移测试，断言新增迁移编号为 `8800001`，高于当前所有迁移，并包含项目、消息、画布、节点、边、节点版本、执行记录和导入记录表。
2. 运行 `go test ./internal/storage -run AgentWorkspace -count=1`，确认因迁移尚不存在而失败。
3. 增加只追加的幂等建表语句，并把迁移追加到 `AppMigrations()`。
4. 重跑测试，确认通过；再运行 `go test ./internal/storage -count=1`。

## 任务 2：正式 Agent 数据模型和存储

文件：

- 新建 `backend/internal/agentworkspace/types.go`
- 新建 `backend/internal/agentworkspace/store.go`
- 新建 `backend/internal/agentworkspace/memory_store.go`
- 新建 `backend/internal/agentworkspace/mysql_store.go`
- 新建 `backend/internal/agentworkspace/mysql_store_test.go`

步骤：

1. 先写 Store 合约和 MySQL 测试，覆盖账号隔离、项目创建与列表、消息追加、画布读取、节点与边保存、revision 冲突、执行记录读取。
2. 运行 `go test ./internal/agentworkspace -count=1`，确认测试按缺失实现失败。
3. 实现类型和 MySQL 事务：保存画布时锁定画布行、核对预期 revision、整体替换该项目的节点和边、revision 加一。
4. 实现内存 Store，供 HTTP 接口做真实行为测试，不用伪造 SQL 调用。
5. 重跑 `go test ./internal/agentworkspace -count=1`。

## 任务 3：旧 JSON 幂等导入

文件：

- 新建 `backend/internal/agentworkspace/legacy_import.go`
- 新建 `backend/internal/agentworkspace/legacy_import_test.go`

步骤：

1. 先写测试，使用临时 `agent-tasks.json` 验证项目、消息、画布和执行记录被完整导入。
2. 再写测试验证相同文件重复导入不会产生重复数据；文件不存在时只记录无数据结果；原文件保持不变。
3. 运行目标测试确认失败。
4. 实现按账号与 SHA-256 摘要记录的导入流程。导入失败不得写完成标记。
5. 重跑包内全部测试。

## 任务 4：把兼容 API 切换到 Agent Store

文件：

- 重构 `backend/internal/httpapi/agent_compat.go`
- 新建 `backend/internal/httpapi/agent_workspace_test.go`
- 修改 `backend/internal/httpapi/router.go`
- 修改 `backend/internal/app/app.go`

步骤：

1. 先写 HTTP 行为测试，覆盖创建项目、列表/详情、重命名、清空消息、删除确认后的项目删除、画布保存、旧 revision 返回 409、选中节点上下文随聊天消息保存。
2. 运行目标测试确认仍在文件存储或缺少新依赖而失败。
3. 给 `RouterOptions` 注入 `agentworkspace.Store` 和旧文件目录；应用启动时创建 MySQL Store。
4. 保留原 `/api/agent/*` 路径和响应结构，把所有任务、消息、画布、执行记录读写切换到 Store。
5. 每个账号第一次访问前触发旧 JSON 幂等导入；不再调用旧文件保存函数。
6. 聊天模型调用成功后，用 Store 原子追加用户消息和 Agent 回复。
7. 重跑 HTTP 测试和 `go test ./internal/httpapi -run Agent -count=1`。

## 任务 5：前端画布保存与节点引用

文件：

- 新建 `frontend/src/user/pages/agent-canvas-state.js`
- 新建 `frontend/src/user/pages/agent-canvas-state.test.js`
- 修改 `frontend/src/user/pages/AgentPageV2.jsx`
- 修改 `frontend/src/user/components/AgentCanvas.jsx`
- 修改 `frontend/src/shared/styles/agent-workspace.css`
- 扩展 `frontend/src/user/pages/agent-workspace.layout.test.js`

步骤：

1. 先写纯函数测试，覆盖节点引用摘要、revision 递增、保存中的后续变化不会丢失、409 冲突需要重新加载。
2. 运行 `node --test src/user/pages/agent-canvas-state.test.js src/user/pages/agent-workspace.layout.test.js`，确认新增行为失败。
3. 实现串行防抖保存：同一时间只保存一个版本，有新变化时紧接着保存最新版本；服务端返回的 revision 成为下一次基线。
4. 保存状态明确显示“未保存 / 正在保存 / 已保存 / 保存冲突 / 保存失败”，不能在请求未成功时固定显示“已保存”。
5. 节点选择继续以胶囊显示，并把有界、可公开的节点摘要送入聊天上下文。
6. 为窄屏补充底部/右侧面板不遮挡画布和输入区的稳定尺寸规则。
7. 重跑前端目标测试。

## 任务 6：完整验证与本地验收

命令：

1. `go test ./internal/storage ./internal/agentworkspace ./internal/httpapi -count=1`
2. `go test ./... -count=1`
3. 在 `frontend` 运行 Agent 目标测试，再运行现有前端测试。
4. 在 `frontend` 运行构建，确认 React 资源可由 Go embed 使用。
5. 运行 `scripts/build-go-local.ps1`。
6. 运行 `scripts/start-local-go-review.ps1 -UseYcmMysql`，保持本地 `4001` 服务。
7. 运行 `scripts/check-local-go.ps1`。
8. 使用已登录浏览器真实验证：创建项目、切换项目、发送消息、移动节点、刷新恢复、节点引用、桌面和窄屏布局。
9. 检查 `git diff --check` 和 Agent 相关差异，确保没有碰无关文件、旧迁移或公网配置。

## 第一阶段完成标准

- Agent 项目、消息、画布、节点、边、视图和 revision 的事实来源是 MySQL。
- 旧 JSON 可重复安全导入，接口不再向旧文件写入，旧文件仍保留。
- 创建与切换项目可用，节点选择会进入对话上下文。
- 移动节点后刷新仍恢复；并发旧 revision 不会静默覆盖新数据。
- 桌面和窄屏没有工具栏、画布与对话区互相遮挡。
- Go 测试、前端测试、前端构建、Go embed 构建和本地已登录浏览器验收均有本次新证据。
