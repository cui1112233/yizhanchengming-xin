# 全项目 Go 运行时迁移设计

日期：2026-10-13  
状态：已确认，待实施  
维护分支：`v88`

## 目标

将 ECS、公网和用户电脑上的应用运行时迁移为 Go：

`浏览器 → Go Web/API → MySQL / Redis / TOS / FFmpeg / Chromium / 外部模型 HTTP API`

最终不得运行 Node.js、npm、Node 容器、Node 版 `local-executor`、Node API bridge 或 Node/Playwright 121 Worker。React 只作为构建时前端源码；生产环境仅部署其静态构建产物，由 Go 提供。

## 约束

- Git `v88` 是唯一源码；所有改动先入 Git，再通过既定 Stage → Cutover 发布。
- 不直接修改 ECS；不使用旧 Docker/GHCR 发布路径作为常规入口。
- 前六个收口阶段保留 Node 源码与历史工作流作为只读回滚证据；只有第七阶段才能删除。
- 用户可见名称必须为“视频管理系统”，不显示“121”。
- MySQL、Redis、Nginx、TOS、FFmpeg、Chrome/Chromium 可保留；它们不是 Node 运行时。
- 不迁移或复制公网密钥。所有真实外部验收使用单独提供的本地/生产配置。
- 切换前每个模块必须具备 Go 回归测试、影子验证和回滚开关。

## 当前基线

Go 已具备批量工厂领域模型、MySQL 任务持久化、部分视频供应商适配、合成模型、`merge-worker` 和发布 intent/审计框架。当前残余 Node 职责主要是：

1. `app.js` 和 Node 路由作为页面/API/兼容网关；
2. `routes/script-video.js` 中 H3、Seedance、Yadi 的提交与轮询；
3. `lib/batch-factory-v11/121-publisher.js` 中的视频管理系统预签名上传、TXT 提交与回读；
4. `services/121-browser-worker` 的登录、会话和 Playwright 自动化；
5. Windows `local-executor` 的 Node 任务执行与更新机制；
6. 尚未迁移页面的 Node 业务接口。

## 目标组件

### 1. Go 主服务 `qiantie`

负责认证、所有 `/api/*`、嵌入式静态前端、SPA 回退、配置和运行时依赖检查。它不得代理到 Node。`/api/runtime-build-info` 增加 `runtime: "go"`、迁移版本和 Node bridge 命中计数，切换前后均可审计。

### 2. Go 视频供应商层

在 `backend/internal/batchfactoryv11` 使用统一接口：

```go
type VideoProvider interface {
    Submit(context.Context, FrozenVideoModel, FinalPrompt) (ProviderTaskRef, error)
    Poll(context.Context, FrozenVideoModel, ProviderTaskRef) (ProviderTaskRef, error)
    Cancel(context.Context, FrozenVideoModel, ProviderTaskRef) error
}
```

H3、Seedance、Yadi、豆包本地执行器都实现该接口。提交前持久化任务，轮询写入状态事件；最终媒体 URL 必须经过现有受控 URL/对象存储校验。

### 3. Go 合成服务 `merge-worker`

独立 Go 进程通过 Redis 获取任务，下载受控源素材，执行 FFmpeg，上传 TOS，写回 MySQL 状态。合成 API 只创建幂等任务；Worker 不直接暴露给浏览器。合成失败保存安全错误摘要并允许受控重试。

### 4. Go 视频管理系统连接器 `video-management-connector`

独立 Go 二进制，使用 Chrome DevTools Protocol 驱动受管 Chromium，不使用 Playwright 或 Node。职责：登录、加密会话存储、会话刷新、组织/风格目录读取、预签名上传、TXT multipart 提交、后台回读。

主服务通过版本化内部 HTTP 合约调用它：

```text
POST /v1/sessions/login
POST /v1/catalogs/refresh
POST /v1/submissions
GET  /v1/submissions/{id}
```

连接器不能读取主站数据库；主服务保留凭据加密、发布 intent、确认、审计和最终业务状态。

### 5. Go Windows 本地执行器 `qiantie-executor`

以 Go Windows 二进制替换 Node `local-executor`，保留配对、签名验证、心跳、领取任务、下载输入、执行浏览器/模型动作、上传结果、自动更新检查。协议版本先兼容现有服务端，稳定后才删除旧协议分支。

## 迁移阶段

### 阶段 A：Go-only 基线与可观测性

前端嵌入 Go；所有 API 路由先记录是否使用旧 bridge；启动检查报告 MySQL、Redis、TOS、FFmpeg、Chromium、供应商与连接器状态。此阶段不改变生产流量。

### 阶段 B：视频供应商与合成

补齐 Go 的 H3/Seedance/Yadi API 行为，启动并部署 Go `merge-worker`，以同一批测试媒体与 Node 现有结果进行影子比对。通过后，视频生产和合成写路径切到 Go。

### 阶段 C：视频管理系统发布

实现 Go CDP 连接器并与 Node 连接器对同一非生产账号做目录、会话、上传、提交、回读比对。Go 连接器成功完成真实发布并回读后，主服务切换发布写路径。

### 阶段 D：本地执行器

发布 Go Windows 执行器，支持旧客户端升级；在真实配对、任务领取、浏览器动作、结果上传和离线恢复验证后停止 Node 执行器发行。

### 阶段 E：其余 Node 业务模块

以 API 清单逐项迁移小说获取、小说面板、剧本、配音、设置和其余历史页面。每项完成 Go API 合约测试、浏览器验收与 bridge 零命中验证。

### 阶段 F：生产切换与移除运行时

以精确 Git SHA 部署 Go 主服务、Go Workers、Go 连接器和 Go 执行器；停止 Node 服务和 Node Worker；验证进程、容器、Nginx upstream、运行时 build-info 与零 bridge 命中。保留 Node 源码和旧部署文件至 V88 第七阶段。

## 数据与安全

- 发布凭据及视频管理系统会话一律 AES-GCM 加密保存，密钥只在运行时配置中出现。
- 发布动作必须沿用 intent → 显式确认 → 提交 → 审计的不可替换媒体来源规则。
- 上传仅接受主服务签发或对象存储受控的媒体 URL；连接器不得接受浏览器任意 URL。
- Chrome 用户目录为专用受限目录；日志中不得输出 Cookie、Authorization、预签名 URL 查询参数或密钥。
- 任务幂等键维持 owner、batch、book、request 组合；重启后的恢复依据 MySQL/Redis，不依据进程内存。

## 验收与回退

每阶段必须包含单元测试、HTTP 合约测试、真实但隔离的外部 smoke、浏览器回归和重启恢复测试。发布阶段还必须证明：

1. `ps`/服务清单中没有 Node/npm/Node Worker；
2. 没有 Node 容器，Nginx 无 Node upstream；
3. Go build-info 显示当前精确 SHA 和 `runtime=go`；
4. Node bridge 命中计数为零；
5. 出图/视频/合成/视频管理系统发布均可完成并回读。

任一阶段失败，仅回退到上一个已验证 Git SHA 和既有运行配置；不得通过线上临时编辑绕过失败。

## 非目标

- 当前迁移不改变产品流程、数据语义或 React 页面设计。
- 当前迁移不在第七阶段之前删除历史 Node 源码、分支或工作流。
- 当前迁移不承诺第三方模型、对象存储或视频管理系统账号未经真实配置即可成功。
