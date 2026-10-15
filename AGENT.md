# 仓库协作规则

## GitHub Actions 与编译禁用规则

- **源码优先，禁止编译产物覆盖源码或页面**：不得直接把 `frontend/dist`、`frontend/public`、服务器现有静态文件或其他编译产物复制回来，作为页面实现或修复方式。必须先修改可维护的 React/前端源码，经过本地构建验证后，再由 Go 发布新的构建结果。
- 禁止创建、恢复或提交 `.github/workflows/` 目录及其中的任何 workflow 文件。
- 禁止通过 GitHub Actions 编译、打包、构建、发布或部署本项目。
- 不得新增 workflow 配置、Actions job、CI 编译脚本或以 GitHub Actions 为前提的发布流程。
- 测试、构建和镜像制作必须在本地开发环境、受控构建主机或 ECS 上通过明确的人工/受控脚本执行。
- 需要发布时，必须先记录源码提交、构建产物或镜像摘要、目标 ECS、服务状态和实际公网验证结果；构建成功不等于部署成功。
- 发现分支或补丁重新引入 `.github/workflows/` 时，必须在合并前移除，并检查 Git diff 确认该目录不存在。

## 数据与运行环境

- 不得使用构建或发布操作删除生产容器、MySQL 数据、Docker 数据卷或业务输出文件。
- 不得把密钥、密码、令牌或供应商凭据写入源码、镜像、配置文件或 Git 提交。
- 线上状态以 ECS 上实际运行的容器、镜像、配置和公网行为为准，不以本地构建、静态检查或单个 HTTP 200 作为完整发布证明。

## 第一性原理工作准则

所有需求、排查、设计、修改和回复都遵守以下准则：

1. **从原始需求和问题本质出发**：不从惯例、模板或旧实现直接推断方案。
2. **先确认目标再行动**：不要假设用户已经明确知道自己想要什么；当动机或目标不清晰且无法从现有上下文确认时，先用中文讨论清楚。
3. **优先选择最短路径**：目标明确但当前路径不是最短时，直接指出问题，并给出更简单、风险更低的方案。
4. **追查根因，不做表面补丁**：遇到故障必须沿数据流、调用链和实际运行结果定位原因；每个改动都要能回答“为什么这样改”。
5. **只输出影响决策的信息**：回复、进度和结论说重点，删除不改变决策的背景和重复说明。
6. **统一使用中文**：所有回复、进度汇报、结论、表格、提问和总结一律使用中文；代码标识符、命令、文件路径和报错原文保持原样。

## 当前代码基线（以仓库实际代码为准）

编码、注释、测试和 commit 规范统一遵循根目录 [CODING_STYLE.md](CODING_STYLE.md)。

以下内容是当前 V88 代码的事实基线。发生冲突时，先检查代码、测试和启动脚本，再修改文档；不要用线上编译产物反向覆盖源码。

- **后端入口**：Go module 位于 `backend/go.mod`，主服务入口是 `backend/cmd/qiantie/main.go`。它同时处理 `/health`、`/api/*` 和嵌入的前端静态资源；未知 API 路径由 Go 明确返回 404/405，不再通过 Node 兜底。
- **前端源码**：可维护的 React 源码位于 `frontend/src`，使用 Vite 构建；用户端和管理端都由 React + Ant Design 组成。路由包括 `/novel-fetch`、`/novel-panel`、`/batch-factory`、`/shuihuo-production`、`/script`、`/tts` 和 `/admin/*` 等。
- **发布方式**：先构建 `frontend/dist`，再由 `scripts/build-go-local.ps1` 复制到 `backend/web/dist`，通过 `backend/web/assets.go` 的 `embed.FS` 编进 Go 二进制。`frontend/dist` 和 `backend/web/dist` 都是构建结果，不是功能源码，禁止直接编辑或当作源码恢复。
- **数据职责**：MySQL 保存业务数据和迁移；Redis 负责队列、短时运行状态和分布式锁；TOS 保存对象及合成产物；FFmpeg 参与音视频标准化、配音和合成。提示词正文由后端 `backend/internal/prompt` 管理并在运行时渲染，前端只选择预设和展示结果。
- **主要业务模块**：`backend/internal/novelfetchworkshop` 负责小说获取、原文保留、截取、风格/男女频判断和敏感词处理；`backend/internal/batchfactoryv11` 与 `backend/internal/httpapi` 负责批量工厂 V11/V12、资产、分镜、视频、合成、上传和恢复接口；`backend/internal/videomanagementconnector` 是视频管理系统的 Go HTTP 插件边界。
- **旧页面边界**：`frontend/public/batch-rewrite` 只作为明确的 V78 对照/兼容页面，由 `/batch-rewrite/v2` 显式访问；`/novel-fetch` 必须指向 React 页面。`frontend/public/novel-panel/workbench` 已被 Go 静态路由拒绝，不能重新接回线上默认入口。
- **本地启动**：使用 `scripts/start-local-go-review.ps1`（可按需加 `-UseYcmMysql`、`-EnableBatchProduction`、`-EnableLocalMerge`），服务由 Go 二进制提供；`scripts/verify-no-node-runtime.ps1` 用于检查运行时是否误启动 Node。不得把“浏览器能打开”或“HTTP 200”当成完整链路验证。
- **文档职责**：`README.md` 说明项目结构、启动、构建、测试和业务链路；本文件只记录协作禁区、源码优先原则和当前代码基线。改动功能时必须同步检查两份文档是否仍准确。

## 公网部署确认规则

- 任何涉及公网服务器、生产环境、正式域名或对外服务的部署，都必须先完成本地构建和检查，再向用户明确说明目标环境、Git SHA、发布内容和回滚点。
- 用户第一次说“部署”“发布”或“执行”不等于公网授权。必须再次向用户确认“是否部署到公网”，得到明确肯定后才能执行上传、替换二进制、重启服务或切换流量。
- 没有第二次明确确认时，只允许本地构建、打包、静态检查和本地运行验证；严禁调用 `scp`、`ssh`、公网发布接口或任何生产切换命令。
- 公网部署完成后必须回报实际目标、服务状态、外部访问结果和对应 Git SHA；不能用“构建成功”代替公网部署成功。
