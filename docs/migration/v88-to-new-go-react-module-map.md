# 旧 v88 → 新 Go/React 模块迁移映射表（Task 9.1.6）

本文件把 Task 9.1.1–9.1.5 的盘点结果收敛成一份**真正用于后续实施的目标架构映射**。

目标不是把旧 v88 文件逐个复制进新仓库，而是把旧行为迁到统一的：

- Go 后端
- MySQL 事实源
- Redis Queue / Lock（需要异步执行的部分）
- React + Ant Design 用户端/管理端
- TOS / 本地产物存储

## 1. 总原则

### 1.1 后端

新主线继续坚持：

- Go 是业务/API/Worker 主语言
- MySQL 是长期事实源
- Goose 管 schema migration
- Redis 管队列、锁、短期运行态
- TOS 管长期媒体对象
- 不长期保留旧 Node `V12 -> V11 Go` 双层代理架构

### 1.2 前端

React 只负责：

- 展示
- 用户输入
- 调用 Go API
- 前端交互状态

React 不作为 Batch/Book/Run/Stage/VIDEO 的事实源。

### 1.3 迁移类型

本表使用四种迁移策略：

| 类型 | 含义 |
|---|---|
| `直接迁语义` | 保留旧业务规则，但按新包结构/数据模型重写 |
| `重写` | 旧架构问题较大，不复制实现，只保留用户可见能力 |
| `合并` | 旧版多个事实源/模块在新主线中归并成一个模块 |
| `淘汰` | 旧兼容层、重复状态、历史入口不进入新主线 |

---

# 2. 建议的新 Go 包结构

后续实现时建议在 `api/internal/` 下逐步形成：

```text
api/internal/
  app/
  httpapi/
  intake/                # 已有
  metadata/              # 已有
  pipeline/              # 已有顶层 BatchProject/Run
  provider121/           # 已有 121 正文/bookinfo

  batchfactory/          # BatchProject / BatchBook / 项目查询与工作台事实源
  workflow/              # Run / BookRun / StageRun / attempt / retry / aggregation
  assets/                # 角色/场景/道具资产与图片版本
  director/              # Hook / Director / Opening / H3 director 业务服务
  promptcompiler/        # effective settings / final prompt / compile trace
  videoproduction/       # ProductionJob / ProductionTask / VIDEO 状态机
  videoprovider/         # personal/H3/YFAI/Local provider adapters + credential resolver
  executor/              # 本地执行器配对、lease、job、artifact 接口
  merge/                 # MergeJob + ffmpeg/TOS
  publish/               # 121 / external publish / intent / audit
  configprofile/         # 版本对应配置档、统一设置、配置继承
  prompts/               # 系统预设/约束提示词事实源
  auth/                  # 后续 Task 15 统一 token/session/permission
  storage/               # TOS / local media abstraction
  queue/                 # Redis queue / lock / recovery infrastructure
```

说明：以上目录不是要求一次性全部建立。只在对应 Task 实施时创建，避免空包和过度设计。

---

# 3. 核心 Batch / Book / Intake 映射

| 旧 v88 | 旧职责 | 新目标 | 策略 | 说明 |
|---|---|---|---|---|
| `backend/internal/batchfactoryv11` Batch/Book | V11 Batch/Book durable model | `api/internal/batchfactory` | 直接迁语义 | 以新 BatchProject 为顶层，逐步增加 BatchBook 事实实体 |
| `routes/batch-factory-v12.js` Batch CRUD | 浏览器 V12 兼容/聚合层 | `api/internal/httpapi/batchfactory_handlers.go` + `batchfactory.Service` | 重写 | 不保留 Node rewrite |
| `routes/batch-factory-v11.js` Go proxy | 注入账号/模型/secret 后转 Go | Go 内部 service + credential resolver | 淘汰/合并 | Node 中间 hop 最终消失 |
| `routes/batch-factory-intake.js` | 旧小说获取交接单 | 已有 `intake` + `pipeline` | 合并 | Task 1–8 已完成新的事实链 |
| V11 Intake APIs | novel/manual intake | 已有 `intake`，必要字段继续扩展 | 合并 | 不再保留第二套 intake store |
| `novel-fetch-workshop` | 121 正文/元数据 | `provider121` + `intake` | 已迁/扩展 | 121 admin/session 能力后续按发布需要拆入 publish/provider121 |

### 新事实关系

建议最终关系：

```text
Intake
  └─ IntakeBook
       ↓ completed
BatchProject
  ├─ BatchBook
  └─ Run
       └─ BookRun
            └─ StageRun
```

Intake Book 的 `fetched/retryable_failed` 只负责“获取阶段”，不能继续承担 Director/VIDEO/Publish 状态。

---

# 4. Run / Automation / Scheduler 映射

| 旧 v88 | 新目标 | 策略 |
|---|---|---|
| Node `automation-orchestrator` | `api/internal/workflow` + `api/internal/queue` | 重写 |
| Node automation preset state | `configprofile` / MySQL | 重写 |
| Node file-backed schedule router | 新 `Run.run_at` + Redis due queue | 淘汰 |
| `BookStageRun` | `workflow.StageRun` | 直接迁语义 |
| old staging `Item.status` | Stage/BookRun 聚合状态 | 淘汰旧枚举 |
| runtime-summary Node 聚合 | Go batch runtime summary query | 重写 |

### 必须保留的 Stage 语义

- assets
- director
- opening
- visual
- image
- video

状态：

- queued
- running
- succeeded
- failed
- cancelled

同时保留：

- attempt
- requestId / idempotency key
- inputRevision
- errorMessage

### 不再保留

旧 staging：

- `queued_hook`
- `hook_generating`
- `hook_review`
- `queued_director`
- `director_generating`
- `complete`

这些应由明确的 stage/status 组合表达。

---

# 5. Hook / Director / H3 映射

| 旧 v88 | 新目标 | 策略 | Task |
|---|---|---|---|
| `DirectorService` | `api/internal/director` | 直接迁语义 | Task 12 |
| working-front viral rewrite | `director.HookService` 或 workflow hook stage | 合并 | Task 12 |
| Hook draft/approve | `director` + MySQL revision | 直接迁语义 | Task 12 |
| Opening variants | `director/opening` | 直接迁语义 | Task 12 |
| H3 director | `director/h3` | 直接迁语义 | Task 12/13 |
| H3 audio measurement | `director/audio` 或 `workflow/audio` | 重写 | Task 13 |
| H3 compile | `promptcompiler` | 合并 | Task 12/13 |
| H3 trace | `promptcompiler` compile trace | 直接迁语义 | Task 12/13 |
| Node text credential injection | Go credential resolver | 重写 | Task 12/15 |

### 关键原则

- 提示词正文/系统预设在后端解析
- API Key 不返回浏览器
- Director revision 必须持久化
- Opening 失败不能回滚已成功 Director
- H3 音频/编译必须可追踪到冻结 revision

---

# 6. Assets / Images 映射

| 旧 v88 | 新目标 | 策略 |
|---|---|---|
| V11 assets CRUD | `api/internal/assets` | 直接迁语义 |
| Asset image versions | `assets.ImageVersion` + MySQL | 直接迁语义 |
| primary image | assets service | 直接迁语义 |
| Node image model credential injection | Go server-side model resolver | 重写 |
| legacy `/image-generation` | 不新建重复接口 | 淘汰/待确认 |
| local image upload | `assets` + `storage` | 重写 |

前端目标：

`前台/src/features/batch-factory/assets/`

资产 UI 必须只保存 asset/image ID，不把 provider URL 当唯一事实源。

---

# 7. Prompt / Config / 统一设置映射

| 旧 v88 | 新目标 | 策略 | Task |
|---|---|---|---|
| V11 batch settings | `api/internal/configprofile` | 合并 | Task 10 |
| Book override | configprofile scoped overrides | 直接迁语义 | Task 10 |
| Video override | configprofile scoped overrides | 直接迁语义 | Task 10 |
| config versions | “版本对应配置档” | 重写 | Task 10 |
| `/api/presets` | `api/internal/prompts` | 重写/迁语义 | Task 12/15 |
| `/api/script-constraint-prompts` | `prompts` + auth | 重写 | Task 12/15 |
| Node preset body enrichment | Go 后端 prompt resolver | 重写 | Task 12 |
| effective settings | `promptcompiler.ResolveEffective` | 直接迁语义 | Task 12 |
| final prompt | `promptcompiler.Compile` | 直接迁语义 | Task 12 |

### Task 10 UI 目标

`前台/src/features/batch-factory/settings/`

包含：

- ProductionSettingsDrawer
- PublishSettingsDrawer
- BookSettingsDrawer/Modal
- VideoSettingsDrawer
- ConfigProfileCard

旧“解析输入”入口直接淘汰。

---

# 8. VIDEO Provider 映射

| 旧 v88 | 新目标 | 策略 | Task |
|---|---|---|---|
| `ProductionService` | `api/internal/videoproduction` | 直接迁语义/重构 | Task 14 |
| `ProductionJob` | `videoproduction.Job` MySQL | 直接迁语义 | Task 14 |
| `ProductionTask` | `videoproduction.Task` MySQL | 直接迁语义 | Task 14 |
| `MemoryVideoProviderRegistry` | `videoprovider` credential resolver | 淘汰 | Task 14 |
| `personal_api` / Yadi | `videoprovider/yadi` | 直接迁语义 | Task 14 |
| `yd2.0-mini` | Yadi adapter model binding | 直接迁语义 | Task 14 |
| `autodl_comfyui` | `videoprovider/autodl` | 直接迁语义 | Task 14 |
| `yfai_seedance` | `videoprovider/yfai` | 直接迁语义 | Task 14 |
| `doubao_local_executor` adapter | `videoprovider/localexecutor` | 直接迁语义 | Task 14 |
| Node provider secret sync | Go account credential resolver | 重写 | Task 14/15 |
| Node V12 provider compatibility | 统一新 Go API | 淘汰 | Task 14 |

### 新 provider 接口建议

Go 内部统一：

```text
Submit(ctx, request) -> ProviderTask
Poll(ctx, providerTaskID) -> ProviderTask
Cancel(ctx, providerTaskID) -> result
Capabilities() -> ProviderCapabilities
```

状态统一：

- queued
- running
- succeeded
- failed
- cancelled

### 必须修正的旧架构问题

- 未配置 provider 不得返回笼统 503
- Go 重启不能丢失关键账号 provider 配置
- browser 不传 API Key
- model/provider 不匹配必须 409/明确错误
- provider 错误必须落库

---

# 9. Local Executor 映射

| 旧 v88 | 新目标 | 策略 |
|---|---|---|
| `backend/internal/localexecutor` | `api/internal/executor` | 直接迁语义 |
| Node local executor bridge | Go HTTP API | 淘汰中间层 |
| pairing | executor Pairing service | 直接迁语义 |
| lease/job heartbeat | executor worker protocol | 直接迁语义 |
| local artifacts | `executor` + `storage` | 重写 |
| protected artifact token | Go signed media access | 直接迁语义 |

最终前端只需要：

- 查看设备状态
- 发起配对
- 查看任务状态

执行器协议本身由 Go 管理。

---

# 10. Merge / TOS 映射

| 旧 v88 | 新目标 | 策略 |
|---|---|---|
| `MergeService` | `api/internal/merge` | 直接迁语义 |
| `MergeJob` | MySQL merge_jobs | 直接迁语义 |
| local ffmpeg adapter | merge local worker | 直接迁语义 |
| remote merge adapter | merge provider adapter（如仍需要） | 按需保留 |
| TOS object store | `api/internal/storage` | 合并 |
| merge cover | merge thumbnail service | 直接迁语义 |
| Node media proxy/compat | Go protected media endpoint | 重写 |

Merge 状态继续独立，不压进 Run.status。

---

# 11. Publish 映射

| 旧 v88 | 新目标 | 策略 | Task |
|---|---|---|---|
| external credential/intents/audits | `api/internal/publish` | 直接迁语义 | Task 15 |
| Node 121 session publisher | `publish/provider121` | 重写 | Task 15 |
| 121 organizations | publish/provider121 | 直接迁语义 | Task 15 |
| classify publish metadata | publish metadata service | 合并 | Task 15 |
| Node session/cookie fallback chain | 服务端 credential/session abstraction | 重写 | Task 15 |
| publish permission | auth/permission + publish | 重写 | Task 15 |

浏览器永远不接触 121 Cookie/API secret。

---

# 12. Auth / Node Bridge 映射

| 旧 v88 | 新目标 | 策略 |
|---|---|---|
| Node `apiAuth` | `api/internal/auth` middleware | 重写 |
| Node accountStore/memberStore | MySQL user/team/permission repositories | 重写 |
| signed Node→Go BridgeAuth | 单体/统一 Go 后不再需要同进程业务 bridge | 淘汰 |
| Node model-catalog-runtime | Go account model resolver | 重写 |
| Node token/session compatibility | Go auth/session | 重写 |

### 例外

如果过渡期仍必须与旧服务通讯，可以保留**边界 adapter**，但不能让 Node 继续成为新主线事实源。

---

# 13. 前端页面映射

建议建立：

```text
前台/src/
  api/
    intake.js
    batchFactory.js
    videoProduction.js
    publish.js

  pages/
    NovelIntakePage.jsx
    BatchFactoryPage.jsx
    ShuihuoProductionPage.jsx

  features/batch-factory/
    project-list/
    workbench/
    books/
    settings/
    assets/
    director/
    production/
    merge/
    publish/
    automation/
```

### 旧入口映射

| 旧 v88 | 新 React |
|---|---|
| `BatchFactoryWorkbenchPage.jsx` | `pages/BatchFactoryPage.jsx` + feature modules |
| `BatchFactoryV11UiPage.jsx` | 不再作为第二套权威页面；能力拆入新 feature |
| `BatchFactoryV11Workbench.jsx` | `features/batch-factory/workbench` |
| `BatchFactoryNovelList.jsx` | `features/batch-factory/books` |
| `BatchFactoryCreateModal.jsx` | Task 8 intake + BatchFactory project actions；避免再造第二套小说获取 |
| `BatchFactoryUnifiedSettingsModal.jsx` | `features/batch-factory/settings/ProductionSettingsDrawer` |
| `BatchFactoryPublishSettings.jsx` | `settings/PublishSettingsDrawer` |
| `DirectorPanel.jsx` | `features/batch-factory/director` |
| `HookReviewPanel.jsx` | `features/batch-factory/director/HookReview` |
| `FinalPromptPreviewDrawer.jsx` | `features/batch-factory/director/FinalPromptDrawer` |
| `ExternalPublishPanel.jsx` | `features/batch-factory/publish` |
| `ProductionMediaBoundary.jsx` | `features/batch-factory/production` 的受保护媒体组件 |

### 明确淘汰

- 两套 Batch Factory 权威页面并存
- V11 preview 作为生产事实入口
- 浏览器端业务状态作为事实源
- 旧“解析输入”

---

# 14. 管理端映射

建议建立：

```text
后台/src/features/
  prompts/
  config-profiles/
  models/
  permissions/
  executors/
  diagnostics/
```

管理端负责：

- 系统预设/提示词
- 配置档模板
- 模型目录
- 团队/发布权限
- 执行器/Provider 诊断

不直接编辑用户具体 Batch 的运行事实。

---

# 15. API Namespace 收敛

旧 v88 同时存在：

- `/api/batch-factory`
- `/api/batch-factory/v11`
- `/api/batch-factory/v12`

新主线当前已有 `/api/v1/intakes`。

后续新业务统一建议：

```text
/api/v1/batch-projects
/api/v1/batch-projects/{id}/books
/api/v1/batch-projects/{id}/runs
/api/v1/runs/{id}
/api/v1/runs/{id}/books/{bookId}/stages
/api/v1/video-jobs
/api/v1/video-providers
/api/v1/merge-jobs
/api/v1/publish-jobs
/api/v1/config-profiles
```

原则：

- 新主线不继续制造 V13/V14 命名空间
- 版本放 HTTP API 大版本，不把业务代号 V11/V12 固化成长期 URL
- 过渡兼容只在 ECS 切换期提供 adapter，最终删除

---

# 16. 数据库事实源映射

## 已有

- intakes
- books/intake books
- batch_projects
- runs

## Task 9/后续建议新增

- batch_project_books
- run_books / book_runs
- stage_runs
- stage_attempts（或 attempt 直接在 stage_runs）
- assets
- asset_images
- director_revisions
- hooks / hook_revisions
- video_jobs
- video_tasks
- merge_jobs
- publish_jobs / publish_intents / publish_audits
- config_profiles / scoped overrides

所有新增必须使用 Goose migration，并为幂等/owner/batch/book 查询建立必要唯一索引。

---

# 17. Redis 职责映射

Redis 只承担运行基础设施，不承担长期事实：

- due Run queue
- stage work queue
- distributed lock
- provider polling schedule
- retry delay
- worker lease/heartbeat（如需要）

任务完成/失败、attempt、error、providerTaskId 等最终结果必须写回 MySQL。

---

# 18. 旧模块迁移处理总表

| 旧域 | 新域 | 处理 |
|---|---|---|
| V11/V12 Node proxy | Go httpapi/service | 淘汰 Node 中转 |
| Batch/Book | batchfactory | 迁语义 |
| Intake | intake/provider121 | 已迁大部分 |
| staging Item | workflow StageRun | 淘汰旧粗状态 |
| Automation/Schedule | workflow + queue | 重写 |
| Assets | assets | 迁语义 |
| Hook/Director/H3 | director | 迁语义/重构 |
| Prompt compile | promptcompiler | 迁语义 |
| Settings/Config versions | configprofile | 重写 |
| VIDEO Production | videoproduction | 迁语义/重构 |
| Provider registry | videoprovider credential resolver | 淘汰内存事实源 |
| Local executor | executor | 迁语义 |
| Merge | merge | 迁语义 |
| TOS/local artifact | storage | 合并 |
| Publish | publish | 迁语义/重写 121 session |
| apiAuth/BridgeAuth | auth | 重写并收敛 |
| 两套 Batch Factory UI | 单一 BatchFactoryPage + features | 合并 |

---

# 19. Task 9 后续的实际落地顺序

映射完成后，Task 9 不应该一次把所有旧 V11 功能复制回来。

按当前 checklist：

1. **9.2 BatchProject 列表**
   - 先建立 `batchfactory` 查询能力
   - Task 8 创建的项目必须立即可见
2. **9.3 项目详情**
   - 建立 BatchProject ↔ IntakeBook/BatchBook 关系
   - 展示书名、Book ID、书城、男女频、风格、正文状态、错误
3. **9.4 Run 执行框架**
   - Run -> BookRun -> StageRun
   - Redis queue/lock
   - retry/idempotency/recovery
4. Task 10 再做配置档/统一设置
5. Task 12 再接 Director/Hook/Prompt
6. Task 14 再接 VIDEO provider

这样可以做到“一块完成、一块合并”，而不是重新形成一个巨大 V11 分支。

---

# 20. Task 9.1.6 验收结论

已完成：

- 旧 Node / Go / React 模块逐域映射
- 新 Go 包职责划分
- 新 React 页面/feature 职责划分
- 旧 V11/V12 API namespace 收敛原则
- MySQL / Redis / TOS 职责边界
- 明确哪些旧模块迁语义、重写、合并或淘汰
- 明确 Task 9.2 起的实际开发顺序

本映射是后续迁移的架构约束；若实际开发发现必须调整，应先更新本文件和 `TASKS.md`，再改代码，避免再次出现旧 v88 多层事实源和多套入口并存。