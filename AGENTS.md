# AGENTS.md — 一战晟铭新主线协作规则

本文件是 `cui1112233/yizhanchengming-xin` 的长期协作规则。所有 AI / Coding Agent / 开发任务在写代码前必须先读取本文件和 `docs/collaboration/WORKSTREAM_REGISTRY.md`，再读取目标分支当前 HEAD 与现有代码。

## 1. 唯一主线与验收

- `main` 是唯一最终验收和发布主线。
- 模块成果通过 PR 汇入 `main`；禁止旧分支覆盖、替换、force-push `main`。
- 当前代码起点 `main@19df4c4ce8f446df856ec0f0eb1558fd913de51d` 带 `[skip ci]`，且该 SHA 无 workflow run 记录；它只能作为代码起点，不能作为“测试已通过”的基线。
- 部署只允许使用经过明确验收的完整 SHA；禁止使用“某分支最新版”作为部署标识。
- 本轮及未获明确授权前：不合并、不部署、不删除分支、不修改分支保护、不改公网、不删除生产数据、不调用真实付费模型、不执行真实发布。

## 2. 固定任务分支

不得为同一功能创建 `final`、`rebased`、`new`、`copy`、`v2` 等替代分支。小修改继续提交到原任务分支。

| 任务 | 固定分支 |
| --- | --- |
| A 共享基础 / 总集成 | `feat/shared-platform-foundation-20261007` |
| B 小说获取 | `feat/novel-fetch-public-parity` |
| C 剧本生成 | `feat/script-generation-parity-20261007` |
| D 小说面板 | `feat/novel-panel-module-20261007` |
| 已有水货 B1 成果 | `feat/v88-parity-shuihuo` |

`feat/v88-parity-shuihuo` 当前相对 `main` 是 18 个领先提交、0 落后，必须保留并优先评估复用；不得重写或覆盖其 B1 业务成果。

## 3. Worktree 与写入者

- 每个开发任务必须使用一个独立 worktree、一个固定分支、一个写入者。
- GitHub 远端分支不是本地 worktree。不能执行本机命令时，必须登记实际使用的云环境或远端操作方式，并把 worktree 状态写为“无法核验”。
- 不得要求用户为了补登记手工折腾 Git；能通过现有远端工具核对的内容直接核对。
- 在能执行本机 Git 的环境中，写入前必须核对：当前 worktree、当前分支、HEAD、未提交修改和其它 worktree，不得覆盖其他任务工作。

## 4. 文件归属

### A：共享基础 / 总集成

A 统一维护和最终收口：

- `AGENTS.md`、`docs/collaboration/**`、全局进度登记；
- `api/internal/app/**` 的全局依赖注入；
- `api/internal/httpapi/server.go` 的全局路由和 `Dependencies` 接线；
- 共享 Auth / team context / shared API contract；
- `api/cmd/server/**`；
- `api/go.mod`、`api/go.sum`；
- `.github/workflows/**`；
- 数据库 migration 编号分配与最终串行化；
- Docker / deploy / embed / 服务启动配置；
- 前端全局路由、共享请求层和最终模块挂载；
- `TASKS.md` 与最终集成进度。

B/C/D 或 Shuihuo 可以在模块分支中保留必要的“接线候选 hunk”，但最终修改上述全局文件由 A 统一核对、重放或收口。

### B：小说获取

B 固定维护小说获取业务，优先放在模块专属路径：

- `api/internal/provider121/**`；
- 小说获取专属 service / handler / tests；
- 前端小说获取模块专属目录；
- 121 正文与元数据、失败重试、处理记录等模块逻辑。

`api/internal/intake/**` 中属于共享 BatchProject / Run DTO、共享事实模型或全局聚合的修改需先向 A 登记，不另造并行 Intake/Run 模型。

### C：剧本生成

C 固定维护：

- `api/internal/generation/**`；
- 剧本 / EXTRACT / SCRIPT / HOOK / DIRECTOR / FINAL_PROMPT 模块逻辑；
- `前台/src/script-workbench/**`；
- 对应模块测试。

C 当前远端提交 `e1411daac834589fde31c5777ecf44f1139cfbed` 已包含对 `前台/src/BatchProjectListPage.jsx` 的集成 hunk；该提交保留，但后续该共享文件的最终挂载由 A 收口，C 不继续扩大共享文件改动范围。

### D：小说面板

D 固定维护小说面板专属 domain / service / handler / tests 与前端模块目录。若新建正式包，使用唯一正式目录（例如 `api/internal/novelpanel/**`、`前台/src/novel-panel/**`），禁止创建 `novel-panel-v2/new/final` 等重复实现。

### Shuihuo B1

`feat/v88-parity-shuihuo` 已拥有：

- `api/internal/shuihuo/**`；
- `api/db/migrations/00009_shuihuo_project_core.sql`；
- Shuihuo B1 handler / tests / design / plan。

其中 `api/internal/app/app.go` 与 `api/internal/httpapi/server.go` 的改动属于 A 最终集成区；Shuihuo domain / store / migration / module handler 本身按已有成果保留。

## 5. Migration 编号

- `00009_shuihuo_project_core.sql` 已由 `feat/v88-parity-shuihuo` 占用，永久保留给该成果，禁止改号或挪作它用。
- 分配任何新编号前，A 必须扫描所有仍有领先提交的活动分支，不能只看 `main`。
- 旧活动分支可能包含不同历史路径（例如 `api/migrations/**`）和旧编号；必须同时核对“路径 + 语义 + 是否仍待吸收”，不能仅按最大数字递增。
- B/C/D 不自行决定 migration 编号；先向 A 登记用途、依赖和 rollback 语义，由 A 串行分配。

## 6. PR 与旧分支处置

- PR #11 的主要 BatchProject 列表/聚合功能已由 PR #16 明确复用吸收；#11 不整支强行合并。先核对剩余测试/文档差异，再决定是否还有需要移植的内容。
- `feat/task11-shuihuo-production` 相对当前 `main` 的提交图显示有 7 个 main 未包含提交。**这只能说明提交图未包含，不能表述为“PR 合并后新增了 7 个提交”**；必须对照 PR 合并方式、merge/base/head 和实际补丁后，才能判断是否有新功能。
- PR #21 安全审计和 PR #25 性能审计保留；后续按仍有效的测试、修复和文档逐项吸收，不强行整支合并。
- 已完全进入 `main` 的旧分支也不得直接删除。删除前必须同时确认：远端 ahead=0、本地无未推送 commit、没有 worktree 正在使用、用户明确批准。
- 其它分支逐项核对，禁止批量删除。

## 7. 共享接口规则

B/C/D 必须优先使用 `docs/collaboration/WORKSTREAM_REGISTRY.md` 中登记的共享接口：

- 身份与团队上下文；
- 文本模型调用；
- 任务提交 / 查询 / 取消；
- 资产引用；
- API 错误格式；
- 模块路由接入方式。

已有接口直接复用准确路径和类型；登记为“待 A 实现”的接口，任何模块不得自行另造第二套事实源、Provider Registry、Task Runtime、Asset Store、Auth/Team 模型或错误协议。

## 8. 安全与秘密

- 原账号、配置、作品、历史数据必须保留。
- 密码、Token、API Key、Cookie、TOS 凭据、DSN、发布凭据不得进入聊天、Git、测试固定值或日志。
- 浏览器不得接触服务端 Provider Secret、TOS Secret 或发布凭据。
- MySQL 是长期业务事实源；Redis 只做 Queue / delayed work / Lease / Fencing / 短期运行协调；TOS 只做长期对象存储。
- Redis 丢失后必须可由 MySQL 事实恢复；不得把 Redis 或 localStorage 作为核心业务事实源。

## 9. 每次交付报告

每次交付必须准确报告：

- 任务、实际执行环境 / 远端操作方式、worktree 状态；
- 固定分支、起始 SHA、当前 HEAD；
- 本次 commit SHA 与 message；
- 修改文件；
- migration（如有）；
- 实际执行过的测试命令及 PASS / FAIL；未执行必须明确写“未执行”；
- PR 编号和状态；
- 是否已进入 `main`，若已进入则给 merge SHA；
- 未提交修改状态（若无法执行本机命令则写“无法核验”）；
- 公网验收状态。

禁止用“应该通过”“已存在类似功能”“以前 PR 绿过”替代当前提交的真实验证结果。
