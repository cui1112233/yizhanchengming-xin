# A/B/C/D 协作与共享接口登记

更新时间：2026-10-07  
仓库：`cui1112233/yizhanchengming-xin`

本文件记录当前协作事实、活动分支、环境登记、迁移编号和 B/C/D 必须复用的共享接口。它不是“完成清单”；仓库实现、自动测试、PR 合并和公网验收必须分别报告。

## 1. 当前主线基线

- `main`：`19df4c4ce8f446df856ec0f0eb1558fd913de51d`
- 该提交 message 含 `[skip ci]`。
- 已核对该 SHA 的 GitHub workflow runs：0。
- 结论：这是代码起点，不是“CI 已通过”的验收基线。

## 2. A/B/C/D 环境登记

当前会话无法执行各任务机器上的本机 Git 命令，因此不把 GitHub 分支冒充成本地 worktree。以下“当前 HEAD”仅表示已验证的 GitHub 远端状态。

| 任务 | 固定分支 | 起始 SHA | 最后已核验远端 HEAD（快照） | 实际可验证操作方式 | Worktree 状态 | 写入边界 |
| --- | --- | --- | --- | --- | --- | --- |
| A 共享基础/总集成 | `feat/shared-platform-foundation-20261007` | `19df4c4ce8f446df856ec0f0eb1558fd913de51d` | 最新登记提交前已到 `480e1590c654a67e348dc153bba674cd146b0fb5`；本文件提交后以 A 分支 HEAD 为准 | ChatGPT 云端 GitHub connector / GitHub API 远端操作 | 无法核验 | 共享基础、全局接线、迁移编号、CI/部署配置、最终集成 |
| B 小说获取 | `feat/novel-fetch-public-parity` | `19df4c4ce8f446df856ec0f0eb1558fd913de51d` | `af114e065e99a00f171ca13ec25930c0dc372858` | 当前只能通过 GitHub 远端验证分支状态；任务执行机器/worktree 未暴露 | 无法核验 | 小说获取模块 |
| C 剧本生成 | `feat/script-generation-parity-20261007` | `19df4c4ce8f446df856ec0f0eb1558fd913de51d` | `15a09e1de4c1577434dab6a1f0185094e2d3bd7e` | 当前只能通过 GitHub 远端验证分支状态；任务执行机器/worktree 未暴露 | 无法核验 | generation + script-workbench |
| D 小说面板 | `feat/novel-panel-module-20261007` | `19df4c4ce8f446df856ec0f0eb1558fd913de51d` | `19df4c4ce8f446df856ec0f0eb1558fd913de51d` | 当前只能通过 GitHub 远端验证分支状态；任务执行机器/worktree 未暴露 | 无法核验 | 小说面板模块 |

登记刷新时的远端事实：

- B：最后复核快照为 `af114e065e99a00f171ca13ec25930c0dc372858 feat(novel-fetch): add isolated six-view module UI`；相对 `main` ahead=3 / behind=0；该 SHA workflow runs=0。其前一提交 `2eea26b…` 新增的是 Novel Fetch 模块内 `GET /runs/{id}/books` 读回能力，不视为共享 Task submit/get/cancel facade；是否挂到全局路由仍由 A 集成。
- C：`15a09e1de4c1577434dab6a1f0185094e2d3bd7e fix(script): preserve existing single-book action label`；相对 `main` ahead=5 / behind=0；该 SHA workflow runs=0。C 的 `e1411da…` 是其中较早的 workbench 提交，不再称为当前 HEAD。
- D：仍为 `19df4c4ce8f446df856ec0f0eb1558fd913de51d`，ahead=0 / behind=0。
- A：本协作登记已在固定 A 分支继续提交；最终 HEAD 以本轮报告的 commit SHA 为准。

这些 HEAD 是最后一次远端核验快照；并行任务可继续推进，因此不能把快照当作分支永久 HEAD，也不代表 worktree 状态或未提交修改。后续若远端 HEAD 变化，以新的实际 SHA 更新本表，不根据聊天记忆推断。

## 3. Shuihuo 18 个领先提交保护登记

分支：`feat/v88-parity-shuihuo`  
HEAD：`7aaecb1c7960e46c0b8b8a931e616197f5aba35b`  
相对 `main`：ahead=18 / behind=0。  
该 HEAD workflow runs：0。

该成果链包含：

- `api/internal/shuihuo/**` Project / Source / Segment / segmentation / MySQL Store；
- `api/internal/httpapi/shuihuo_handlers.go`；
- `api/internal/app/shuihuo_smart_segmenter.go`；
- `api/db/migrations/00009_shuihuo_project_core.sql`；
- B1 design / plan；
- `api/internal/app/app.go`、`api/internal/httpapi/server.go` 的接线候选。

复用原则：

- Shuihuo domain/store/migration/handler 作为已有成果保留；
- shared text provider 只作为注入消费，不创建 Shuihuo 私有 provider registry；
- app/server 的最终接线由 A 收口；
- 不让 B/C/D 重写这套 B1 核心。

## 4. Migration 活动分支扫描

分配新 migration 编号前必须重新扫描所有仍有领先提交的活动分支。

本次扫描已确认：

- `main`：`api/db/migrations/00001,00002,00003,00004,00006,00007,00008`；
- `feat/v88-parity-shuihuo`：在上述基础上增加 `00009_shuihuo_project_core.sql`；
- `feat/script-generation-parity-20261007`：当前仍使用 main 的 00001–00008，无新 migration；
- `audit/security-permission-final`、`test/performance-concurrency-capacity`：基于较早快照，目前到 `00007`；
- `feat/task9-batch-factory-workbench`：历史 `00002_batch_run_execution.sql`；
- `tmp-task14-migration-rename`：历史 `00005_task14_video.sql`；
- `agent-workspace-v1` 使用另一历史路径 `api/migrations/00001–00005`；
- `feat/unified-batch-pipeline*` 使用另一历史路径 `api/migrations/**`。

因此：

- `00009` 永久保留给 Shuihuo B1；
- 暂不在本登记中预分配 `00010`；
- 下一次分配前再次扫描所有 ahead>0 活动分支，并核对“路径 + 语义 + 是否待吸收”。

## 5. PR / 分支处置事实

### PR #11 与 #16

PR #16 的说明明确写明其复用了 PR #11 的聚合实现，且 #16 已合入 main。#11 不整支强行合并；剩余测试/文档先逐项核对，再决定是否移植。

### Task11 旧分支表述修正

`feat/task11-shuihuo-production` 当前 compare graph 相对 `main` 显示 ahead=7、behind=409。

这**只能确认当前提交图中有 7 个 main 未包含提交**。不能据此写“PR #13 合并后又新增了 7 个提交”。是否属于后续新功能，必须对照：

1. PR #13 的 head/base/merge 方式；
2. 分支 merge-base；
3. 这 7 个提交的补丁是否已通过 squash/cherry-pick/等价变更进入 main。

结论核对前，分支保留，不删除、不整支合并。

### PR #21 / #25

- #21 `audit/security-permission-final`：保留，按仍有效的安全测试、修复和文档逐项吸收。
- #25 `test/performance-concurrency-capacity`：保留，在 Runtime 最终结构稳定后重新同步和验证。
- 两者均不得因“分支落后”而强行整支合并。

## 6. 共享接口契约

### 6.1 身份与团队上下文 — 已存在

**权威用户类型**

路径：`api/internal/authn/session.go`

```go
type User struct {
    ID           int64    `json:"id"`
    Username     string   `json:"username"`
    DisplayName  string   `json:"name"`
    Role         string   `json:"role"`
    TeamID       int64    `json:"teamId,omitempty"`
    Capabilities []string `json:"capabilities,omitempty"`
}
```

**从 Context 读取当前用户**

路径：`api/internal/authn/context.go`

```go
func CurrentUser(ctx context.Context) (User, bool)
func WithCurrentUser(ctx context.Context, user User) context.Context
```

**鉴权与 capability**

路径：

- `api/internal/httpapi/auth_handlers.go`：`requireAuth`
- `api/internal/httpapi/permissions.go`：`requireCapability`

现有 capability：

- `batch.view`
- `batch.execute`
- `batch.configure`
- `publish.configure`
- `publish.account.configure`
- `publish.execute`
- `publish.audit.view`

规则：

- B/C/D handler 不定义第二套 User/Session/Role。
- 所有 user/team ownership 都从 `authn.CurrentUser(r.Context())` 派生。
- `admin` / `owner` 当前由共享权限层视为 capability bypass；模块不得自行添加另一种 superuser 语义。
- 模块若只需 owner/team 隔离，使用 `User.ID` 和 `User.TeamID`。
- **缺失**：新主线尚无统一的团队成员枚举 / 邀请 / 团队治理 service contract。B/C/D 如需此能力，登记给 A；不得自建 team 表或第二套成员系统。

### 6.2 文本模型调用

#### 已存在：generation Provider

Provider 接口路径：`api/internal/generation/store.go`

```go
type Provider interface {
    Complete(context.Context, TextRequest) (string, error)
}
```

请求类型路径：`api/internal/generation/model.go`

```go
type TextRequest struct {
    BookID               int64
    Stage                Stage
    SystemPrompt         string
    UserPrompt           string
    DirectorMode         DirectorMode
    MatchAudio           bool
    AudioDurationSec     float64
    ShotDurationLimitSec int64
}
```

HTTP provider 实现：`api/internal/generation/provider.go`，`HTTPProvider.Complete`。

当前生产接线：`api/internal/app/app.go`，读取：

- `QIANTIE_TEXT_API_BASE_URL`
- `QIANTIE_TEXT_API_KEY`
- `QIANTIE_TEXT_MODEL`

规则：

- C 继续复用 `generation.Provider`；不得在 C 新增 Provider Registry / API Key 表。
- Shuihuo B1 已通过窄 adapter 消费同一 provider。
- B/D 如果只是调用现有 generation 语义，可通过 A 注入适配器；不得直接读取环境变量或保存 API Key。

**待 A 实现：模块无关的共享 Text Model / Model Catalog contract。**

原因：`generation.TextRequest` 带有 Stage/Director 语义，不适合作为 B/D 所有文本推理的长期公共 DTO。A 后续应提供模块无关的文本调用接口和服务端凭据/模型目录；在此之前 B/D 只能声明依赖，不能自行造第二套。

### 6.3 任务提交 / 查询 / 取消

#### 已存在：Pipeline 创建 Run

路径：`api/internal/pipeline/service.go`

```go
type CreateRequest struct {
    IntakeID       int64
    Name           string
    RunAt          time.Time
    IdempotencyKey string
}

type CreateResult struct {
    Project intake.BatchProject
    Run     intake.Run
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (CreateResult, error)
```

它只适用于 Intake -> BatchProject + Run，不是模块无关 Task API。

#### 已存在：低层 Runtime 协调

路径：`api/internal/taskruntime/contracts.go`

```go
type Queue interface {
    Enqueue(context.Context, Message) error
    Claim(context.Context, string, time.Duration) (Delivery, error)
    Ack(context.Context, Delivery) error
    Nack(context.Context, Delivery, time.Duration) error
}
```

路径：`api/internal/task9runtime/runtime.go`

- `WorkItem`
- `RuntimeCoordinator.Enqueue`
- Worker / Scheduler / Recovery

这些是服务端运行基础设施，不是浏览器或业务模块可以各自包装成第二套任务系统的理由。

#### 已存在：VIDEO 专用任务

类型：`api/internal/video/types.go`

- `StartRequest`
- `StartResult`
- `ProductionJob`
- `ProductionTask`
- `TaskStatus` / `JobStatus`

HTTP 路由集中在 `api/internal/httpapi/server.go`：

- `POST /api/v1/batch-projects/{projectId}/books/{bookId}/video`
- `POST /api/v1/video-tasks/{taskId}/poll`
- `POST /api/v1/video-tasks/{taskId}/cancel`
- `POST /api/v1/video-tasks/{taskId}/retry`

这些仅供 VIDEO 链使用。

**待 A 实现：模块无关 Task submit/get/cancel facade。**

最低契约要求：

- submit 必须带稳定 module/operation/resource identity 与 idempotency key；
- get 返回 durable MySQL 状态，不以 Redis 为事实源；
- cancel 必须是显式、可鉴权、幂等的状态转换；
- 状态至少统一表达 queued/running/succeeded/failed/cancelled；
- 错误使用本文件 6.5 的安全错误格式；
- 底层复用现有 MySQL Run/BookRun 或模块已有 durable job，不允许 B/C/D 各自新建第二套 Redis queue/lease/lock。

在 A 实现公共 facade 前，B/C/D 继续使用各自**已经存在**的业务 API；需要新异步任务时先登记给 A，不自行创建 `tasks_v2` / 新 Queue / 新 Scheduler。

### 6.4 资产引用

#### 已存在：VIDEO Artifact

路径：`api/internal/video/types.go`

```go
type Artifact struct {
    Bucket    string
    ObjectKey string
    URL       string
}

type SubmitRequest struct {
    // ...
    ReferenceImageURLs []string
}
```

存储接口：`api/internal/video/artifact_store.go`

```go
type ArtifactStore interface {
    Persist(context.Context, string, string) (Artifact, error)
}
```

并有 `HTTPArtifactStore.PersistFile` 用于已落地本地文件上传 TOS。

规则：

- 任何模块不得接触 TOS AccessKey / SecretKey。
- `video.ArtifactStore` 是现有 VIDEO 专属实现，不把它直接扩散成所有模块的 domain 类型。
- C 当前分支 `e1411da…` 的 `generation.ScriptEntity.ReferenceImages []string` 是模块 DTO，不是共享资产事实源。

**待 A 实现：模块无关 AssetRef / AssetService。**

在共享资产 contract 落地前：

- B/C/D 只在已有接口明确允许 URL 的字段上传递引用；
- 不新建模块私有 TOS client、asset table、credential table 或 storage registry；
- 需要人物/场景/道具资产持久化时先向 A 登记，由 A 定义共享引用和 ownership。

### 6.5 错误格式

新模块 API 的目标格式以 `api/internal/httpapi/safe_errors.go` 为准：

```json
{
  "code": "STABLE_MACHINE_CODE",
  "message": "可理解且不含秘密的用户消息",
  "request_id": "req_..."
}
```

辅助方法：

```go
func (h handler) writeServiceError(
    w http.ResponseWriter,
    r *http.Request,
    status int,
    code, message, subsystem, operation string,
    err error,
)
```

Request ID 由 `api/internal/observability/observability.go` 的 `X-Request-ID` / context 机制统一生成和传播。

规则：

- 401 = 未认证 / Session 无效；
- 403 = 已认证但无权限；
- 业务 500/503 不得转成“登录过期”；
- raw SQL、DSN、Authorization、Cookie、Token、API Key、Secret 不得返回浏览器；
- 新模块优先使用 `writeServiceError`，不要复制旧接口中历史遗留的 `{"error": ...}` 变体作为新规范。

### 6.6 模块路由接入方式

#### 后端

全局路由与 dependency container：

- `api/internal/httpapi/server.go`
  - `type Dependencies struct`
  - `NewHandler(...)`
  - `http.ServeMux` 路由注册
- `api/internal/app/app.go`
  - MySQL Store / Service / Provider / Runtime 实例化和依赖注入
- `api/cmd/server/main.go`
  - 进程启动和服务端运行配置

这些文件最终由 A 修改。

B/C/D 的做法：

1. 在模块包实现 domain/service/store；
2. 在 `api/internal/httpapi/<module>_handlers.go` 实现 handler 和模块窄 Service interface；
3. 不自行长期编辑 `server.go` / `app.go`；
4. 向 A 提交 route manifest，至少包含：
   - HTTP method；
   - path；
   - handler；
   - dependency/interface；
   - required capability；
   - mutation 是否要求 same-origin；
   - resource ownership check；
5. A 统一把 route 和 dependency 注入最终主线。

已有 Shuihuo / C 分支里的 server/app 或共享页面 hunk 作为“接线候选”保留，由 A 集成时核对，不要求丢弃已有提交。

#### 前端

共享请求层：`前台/src/api.js`

- `requestJSON(path, options)` 统一 `credentials: 'include'`；
- 统一 401 refresh；
- 保持 403 业务权限语义。

规则：

- B/C/D 页面可以创建模块专属组件和模块专属 API wrapper，但底层必须复用 `requestJSON`，不得再造 token/localStorage auth client。
- 全局页面路由/导航和共享页面挂载由 A 收口。
- 模块 PR 应告诉 A：期望 URL、页面组件 export、所需 query/path params、所需 capability。

## 7. 24 条远端已完全被 main 包含的旧分支

以下分支本次 compare 结果均为 ahead=0，只能列为“删除候选”，当前不删除：

1. `audit/ui-layout-interaction-parity`
2. `chore/task9-closeout`
3. `feat/phase1-novel-intake-20261005`
4. `feat/task8-novel-intake-workbench-20261005`
5. `feat/task9-list-detail-final`
6. `feat/task9-runtime-infrastructure`
7. `feat/task10-unified-settings`
8. `feat/task12-script-director`
9. `feat/task13-match-audio`
10. `feat/task14-video-provider-chain`
11. `feat/task15-auth-publish-permission`
12. `fix/task16-known-contract-regressions`
13. `ops/observability-logging-runbook`
14. `task9-batch-project-list-api`
15. `task9-batch-project-mysql-list`
16. `task9-batch-project-name-ui`
17. `task9-ecs-capability-audit`
18. `task9-v11-backend-api-inventory`
19. `task9-v11-inventory`
20. `task9-v11-status-inventory`
21. `task9-v88-new-module-mapping`
22. `task9-video-provider-inventory`
23. `test/task16-browser-e2e`
24. `test/task16-postmerge-verify`

删除前仍需确认本地未推送工作和 worktree 使用情况；当前云端环境无法核验这两项，因此本轮不申请删除。

## 8. 仍需逐项核对的活动/分叉分支

当前已知仍有 main 未包含提交的分支包括：

- `agent-workspace-v1`
- `audit/security-permission-final`
- `feat/task9-batch-factory-workbench`
- `feat/task9-list-detail-closure`
- `feat/task9-list-detail-closure-rebased`
- `feat/task9-list-detail-pr`
- `feat/task11-shuihuo-production`
- `feat/unified-batch-pipeline`
- `feat/unified-batch-pipeline-red2`
- `feat/unified-batch-pipeline-stage2`
- `feat/v88-parity-shuihuo`
- `task9-batch-project-source-ui` / PR #11
- `test/performance-concurrency-capacity` / PR #25
- `tmp-task14-migration-rename`
- `feat/script-generation-parity-20261007`

此列表表示“需要审计”，不表示“都应该合并”。

## 9. 标准交付登记

每个模块交付按以下字段报告：

```text
任务：
执行环境 / 远端操作方式：
Worktree：
分支：
起始 SHA：
当前 HEAD：

本次提交：
- <sha> <message>

修改文件：
- ...

Migration：
- 无 / 000xx_xxx.sql

测试：
- <command> -> PASS/FAIL
- 未执行的项目必须明确写未执行

PR：
- #xx / 尚未创建
- Draft/Open/Merged

main：
- 未进入 main
或
- 已进入 main，merge SHA = ...

未提交修改：
- clean / 具体文件 / 无法核验

公网验收：
- 未执行 / 已执行及证据
```
