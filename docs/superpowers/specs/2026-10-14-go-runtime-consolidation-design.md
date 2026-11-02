# V88 Go 运行时收口设计

日期：2026-10-14  
维护主线：`v88`

## 1. 目标与边界

目标是在不改动小说获取、改文和批量工厂业务链路的前提下，将 V88 的 Node 运行时职责逐步收口到 Go。

允许 Node 的唯一用途是 React/Vite 前端构建。生产、本地运行、Windows 本地执行器、视频管理系统浏览器自动化和视频合成运行时均不得依赖 Node。

本设计只覆盖当前本地 `v88` 工作区。不得直接修改 ECS、公网运行副本，且每项变更必须先进入 Git 的 `v88` 维护线。

本轮不改动：

- 小说获取、改文及其现有 Node/Go 兼容路径；
- 批量工厂的业务编排、生产调度与任务语义；
- MySQL、Redis、FFmpeg、TOS 数据、Docker 网络；
- ECS 公网当前运行环境。

## 2. 目标架构

```text
React 源码
  ↓ Node / Vite（仅构建期）
frontend/dist（唯一前端成品源）
  ↓ Go 构建时复制
backend/web/dist（Go embed 派生产物）
  ↓
Go API 与静态资源服务
  ├─ Go 本地执行器控制面
  ├─ Go 视频合成 Worker
  ├─ Go 视频管理系统 Connector（CDP）
  └─ Go 发布任务执行器
       ↓
MySQL / Redis / FFmpeg / Chromium / TOS
```

每个业务域必须有且只有一个权威的写入、状态和调度方。兼容入口只能透明转发，不能复制状态机或业务决策。

## 3. 阶段 A：静态资源与主入口

### 当前问题

- Node 直接服务 `frontend/dist`；
- Go 使用嵌入的 `backend/web/dist`；
- 根目录 `dist` 与 `.build-verify` 是额外构建/验证副本。

### 设计

- `frontend/dist` 为唯一前端构建成品源；
- `backend/web/dist` 仅由 Go 构建步骤从 `frontend/dist` 派生，用于 `go:embed`；
- 不允许手工编辑 `backend/web/dist`；
- Go 运行时提供页面、`/assets` 和现有 `/api/*`；
- 根 `dist`、`.build-verify` 仅作为历史或验证产物，确认无运行引用后删除；
- 前端保留原有 `/api/*` 调用路径，不因入口切换改变协议。

### 验收

- 不启动 Node 时，本地 Go `:4001` 能加载登录页、用户端、管理端与静态资源；
- Go 健康检查、认证、模型配置、TTS 和本地执行器 API 正常；
- 从空的 `frontend/dist` 与 `backend/web/dist` 可重新构建完整运行产物。

## 4. 阶段 B：Windows 本地执行器 Go 化

### 当前问题

现有 Electron/Node 执行器通过 `/api/local-executor/v1/*` 接入。Go 服务端已拥有配对、心跳、领取、租约、输入下载、结果上传和产物存储；Node 主要是公网桥接与局部兼容。

### 设计

在 `giant-material-executor` 中实现 Go Windows 客户端，兼容现有 API 合约：

- 配对码注册、令牌安全持久化；
- 心跳、任务领取、租约续期；
- 输入下载、任务执行、进度与结果回传；
- 本地日志、自动更新检查；
- 与既有任务 ID、状态和值域完全兼容。

迁移期可与 Electron 客户端共存，但同一任务仅能由一个领取者执行。稳定后停止构建 Electron，旧安装包仅作过渡下载，最终移除 Node 执行器运行依赖。

### 验收

- Go 客户端可完成配对、在线、领取、取消、超时恢复和 MP4 上传；
- 现有任务状态和公开 API 对用户无破坏性变化；
- 不启动 Node 时，Go 客户端仍能直接连通 Go 服务端。

## 5. 阶段 C：视频合成统一为 Go Worker

### 当前问题

当前存在 API 进程内 `LocalMergeAdapter` 与独立 `merge-worker + Redis` 两条 Go 合成执行路径。

### 设计

生产统一采用独立 Go `merge-worker`：

```text
Go API → MySQL 合成任务 → Redis 队列 → Go merge-worker → FFmpeg → 本地工件 / TOS
```

- API 只创建、查询与鉴权读取任务和成品；
- `merge-worker` 负责下载、FFmpeg、上传、重试和恢复；
- `LocalMergeAdapter` 仅允许在开发或无 Redis 的显式本地模式使用；
- 两条路径在过渡期共享相同任务结构、状态枚举和产物 URL 协议；
- 生产模式禁止 API 进程内执行重合成任务。

### 验收

- API 重启不丢失已入队任务；
- Worker 支持重试、恢复、查询及受鉴权成品读取；
- 使用同一输入媒体校验时长、分辨率、音视频流；
- FFmpeg 不由 Node 或 Electron 运行。

## 6. 阶段 D：视频管理系统 Go 化

### D1. Go 视频管理系统 Connector

替换 Node + Playwright Browser Worker，使用 Go 与 Chromium CDP：

- 登录和会话复用；
- 素材目录读取；
- 预签名上传与 TXT/视频提交；
- 发布回读、重试和审计；
- 会话加密保存、临时文件管理和脱敏错误。

组件边界：

- `connector`：视频管理系统业务协议；
- `session-store`：加密 Cookie/会话；
- `cdp-client`：浏览器控制；
- `publish-runner`：发布任务领取、重试、审计；
- `artifact-reader`：读取 Go 合成和执行器产物。

### D2. 替换 `shuihuo-compat`

`shuihuo-compat` 不能通过删除 Node 自动消失。迁移需要逐路由建立权威归属表：

1. Go 已实现：直接使用 Go；
2. Go 未实现：补充 Go；
3. 仅旧页面调用：前端切至稳定 Go API；
4. 无调用：通过运行审计证明后删除。

迁移顺序为影子读取比对、写入切换、路由切换、下线兼容服务。禁止双写长期存在。

### 验收

- 登录、会话复用、目录、TXT/视频上传、提交、发布回读和失败重试通过；
- 发布任务可审计、可恢复；
- 关闭 Node Browser Worker 与 `shuihuo-compat` 后，上述能力保持可用。

## 7. 阶段 E：Node 运行时退役与磁盘回收

仅在 A-D 全部完成且经验证后执行：

1. Nginx 直接反代 Go API 与 Go 静态资源；
2. 停止并删除 Node 服务、Node Browser Worker 与其运行镜像；
3. 删除 Node `local-executor` 构建链和运行依赖；
4. 删除仅为 Node 运行保留的根 `node_modules`、服务端 Node 依赖、旧桥接脚本和已证实无引用的重复产物；
5. 保留 `frontend/package.json`、`frontend/package-lock.json` 与前端构建期依赖；
6. 保留仍承载 MySQL、Redis、Nginx、Go Worker 的 Docker 组件；
7. 每次删除前校验 Git 引用、运行拓扑、健康检查与已知回滚 SHA。

## 8. 风险控制

- 不提前删除 Docker 网络、Worker、兼容服务或持久卷；
- 不复制 ECS 密钥、账号 Cookie 或生产配置到本地；
- 外部账号自动化采用隔离凭据做真实验收；
- 所有公开协议保持 URL、认证、任务 ID、状态和值域兼容，直至旧客户端停止使用；
- 每阶段完成前必须具备本地测试、健康检查和可恢复路径。

## 9. 实施顺序

1. 阶段 A：构建与 Go 静态入口；
2. 阶段 B：Go Windows 本地执行器；
3. 阶段 C：独立 Go 合成 Worker；
4. 阶段 D1：Go 视频管理系统 Connector；
5. 阶段 D2：替换 `shuihuo-compat`；
6. 阶段 E：退役 Node 运行时并清理空间。

小说获取、改文和批量工厂在上述过程保持冻结，不作为本轮迁移或清理目标。
