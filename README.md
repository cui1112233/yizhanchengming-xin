# 一战晟铭 V88

一战晟铭 V88 是一个由 Go 后端和 React 前端组成的内容生产平台，包含小说获取、小说面板、剧本生成、批量工厂、水货生产、配音、视频生成和管理后台等功能。

## 当前架构

```text
浏览器
  └─ React/Vite 前端（frontend/src）
       └─ Go 单体 HTTP 入口（backend/cmd/qiantie）
            ├─ /api/*       Go API 与业务流程
            ├─ /health      健康检查
            └─ 其他页面     Go embed.FS 提供的前端静态资源

MySQL  ← 业务数据、任务、提示词预设、迁移
Redis  ← 队列、短时运行状态、锁
TOS    ← 图片、视频和合成产物
FFmpeg ← 音频/视频标准化与合成
外部接口 ← 121 小说接口、视频管理系统 HTTP 插件、文本/图片/视频模型
```

主服务入口是 `backend/cmd/qiantie/main.go`。它将 API 和嵌入式前端放在同一个 Go 进程内，不依赖 Node 服务端转发。前端源代码仍然只在 `frontend/src` 维护；构建产物不能反向当作源码。

## 页面入口

| 页面 | 路径 | 用途 |
| --- | --- | --- |
| 首页 | `/` | 平台入口和功能导航 |
| 小说获取 | `/novel-fetch` | 获取原文、处理版本、风格/男女频判断、敏感词处理和任务查看 |
| 小说面板 | `/novel-panel` | 管理小说、人物/场景/分镜和相关资产 |
| 批量工厂 | `/batch-factory` | 批量工程列表与生产工作台 |
| 水货生产 | `/shuihuo-production` | 批量工厂与创作漫剧的生产入口 |
| 剧本生成 | `/script` | 剧本、资产和分镜生成 |
| 配音 | `/tts` | TTS 配音与音频处理 |
| 管理后台 | `/admin/*` | 账号、模型、提示词和权限管理 |

`/novel-fetch`、`/novel-panel` 和批量工厂默认使用 React 页面。`/batch-rewrite/v2` 是保留的 V78 对照/兼容入口，只用于核对历史行为，不是 React 页面源码，也不是默认入口。

## 核心业务链路

### 小说获取

1. 前端提交平台、书籍 ID/书名、截取字数和需要生成的版本。
2. Go 从 121 上游获取并保存完整原文；业务处理只使用 `maxTxt` 范围，原文不被覆盖。
3. 优先读取上游 `category`、`genre` 判断男频/女频；无法明确判断时才调用 AI 兜底。
4. 按用户勾选执行原文、AI1-AI5 和敏感词 AI 修复，并保存每个版本的状态、计数、错误和审计信息。
5. 任务列表、实时日志、历史批次和失败重试都通过 Go API 读取，日期筛选使用真实任务日期。

### 批量工厂

1. 创建批次并导入小说，可从小说获取或手工录入。
2. 应用系统、批次、单书和单视频的配置覆盖；提示词从后端预设库读取，运行时替换变量。
3. 按流程生成配音/时长、资产、故事剧本、导演分镜、视频任务、合成和上传状态。
4. V11/V12 API 保留明确的兼容边界，但默认入口和新页面使用当前 React + Go 链路；状态、失败原因和恢复入口必须可追踪。

### 视频管理系统

`backend/internal/videomanagementconnector` 通过 Go HTTP 插件连接视频管理系统，负责登录会话、组织、网站配置档、书籍和上传相关接口。它是外部系统边界，不应把外部页面编译产物混入本项目的前端源码。

## 目录说明

```text
backend/                 Go module、API、任务、提示词、连接器和 embed 入口
  cmd/qiantie/            主 Go 服务
  cmd/qiantie-executor/   本地执行器
  internal/httpapi/        HTTP 路由与接口
  internal/novelfetchworkshop/ 小说获取业务
  internal/batchfactoryv11/     批量工厂业务模型与流程
  internal/prompt/        后端提示词目录与渲染
  internal/queue/         Redis 队列和运行状态
  internal/mergeworker/   合成 Worker 与对象存储
  web/                    embed.FS 入口和构建产物目录
frontend/                 React/Vite 源码与前端构建配置
  src/user/               用户端页面
  src/admin/              管理端页面
  public/batch-rewrite/   V78 对照/兼容静态页面，禁止当作默认实现
scripts/                  本地构建、启动、检查脚本
docs/                     项目执行记忆、盘点和部署说明
```

## 本地运行

### 依赖

- Go 1.23 或兼容版本
- Node.js 仅用于本地构建 React；运行已构建服务不需要 Node 服务端
- MySQL
- Redis
- FFmpeg（需要音视频处理时）

### 配置

复制 `.env.v88-review.example` 为本地配置文件，并填写测试环境的 MySQL、Redis、模型和桥接密钥。密钥只放本地环境变量，不得写入源码或提交。

### 构建并启动 Go 服务

```powershell
Set-Location F:\脚本测试\一战晟铭-v88
npm --prefix frontend install
npm --prefix frontend run build
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql
```

只想重新整理前端静态文件、不重新编译 Go 时，可以使用：

```powershell
.\scripts\build-go-local.ps1 -SkipGoBuild
```

默认本地检查服务监听 `4001`（具体地址以环境变量和启动脚本输出为准）。检查是否误启动了旧版主应用 Node 服务，可运行下面的 Go 启动保护检查；独立浏览器 Worker 和桌面执行器不属于主应用 API 进程：

> **如果页面里“作品全没了”，先看启动日志里的 `本地启动配置：slice=...`。** 代码里 `QIANTIE_BATCH_FACTORY_V11_SLICE` 默认是 `0`，而 Slice 小于 `1` 时 V12 批次路由不会注册，`GET /api/batch-factory/v12/batches/summary` 返回 404，前端就表现为作品列表为空。`scripts/start-local-go.ps1` 会强制把 Slice 兜底到 `3`（`0`/空/非数字都会提升并警告），用临时端口验证可加 `-SkipBuild -ListenAddr ':4002'` 复用现有二进制直接启动。

```powershell
.\scripts\verify-no-node-runtime.ps1
```

## 测试与验证

前端测试和构建：

```powershell
npm --prefix frontend test -- --runInBand
npm --prefix frontend run build
```

Go 测试：

```powershell
go test ./...
```

完成后至少检查：

1. `/health` 返回正常；
2. `/novel-fetch`、`/batch-factory`、`/shuihuo-production`、`/admin/presets` 能打开对应 React 页面；
3. 关键 API 的实际返回、任务状态和错误信息正确；
4. 生产或公网验证必须绑定明确 Git SHA。构建成功、单个 HTTP 200 或静态测试通过，都不能单独证明公网完整链路成功。

## 源码与发布规则

- 维护分支是 `v88`，历史 V78 和旧功能分支只读参考。
- 所有功能先改 React/Go 源码，再本地构建；禁止把 `frontend/dist`、服务器静态文件或其他编译产物复制回来当实现。
- 禁止使用 GitHub Actions 编译、发布或部署；发布必须由受控脚本完成并记录源码提交 SHA、构建结果、目标环境和验证结果。
- 日常本地运行以 Go 二进制为入口。线上是否仍有混合运行时，必须以实际部署记录和公网检查为准，不能仅凭本地代码推断。
- 未经用户明确指令，不向 GitHub 远端提交或推送。

## 相关文档

- [AGENT.md](AGENT.md)：协作禁区、源码优先原则和当前代码基线。
- [docs/GO_EMBED_DEPLOYMENT.md](docs/GO_EMBED_DEPLOYMENT.md)：Go embed 单二进制构建、启动和发布规范。
- [CURRENT_VERSION.md](CURRENT_VERSION.md)：V88 合并冻结和历史生产运行时说明。
- [docs/V88_PROJECT_EXECUTION_MEMORY.md](docs/V88_PROJECT_EXECUTION_MEMORY.md)：V88 项目执行记忆。
- [docs/obj/2026-09-08-v88-consolidation-stage1-inventory-freeze.md](docs/obj/2026-09-08-v88-consolidation-stage1-inventory-freeze.md)：合并冻结盘点基线。
