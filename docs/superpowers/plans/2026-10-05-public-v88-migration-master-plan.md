# 公网 / v88 迁移总实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把旧公网 / v88 中仍然需要保留的核心能力，迁移并收口到 `cui1112233/yizhanchengming-xin` 新主线，使小说获取、Batch Factory V11、水货生产、剧本 / Director / Hook、音频 matchAudio、视频生成、发布与权限最终都能在统一 Go + React + MySQL + Redis 架构下稳定运行，并在 ECS 公网完成真实全流程验收后切换生产。

**Architecture:** 后端统一以 Go 服务作为业务事实入口，MySQL 作为业务状态事实源，Redis 承担异步队列 / Lock / 定时任务，TOS 管理对象产物；用户端和管理端使用 React + Ant Design。迁移过程不一次性重写全部旧系统，而是按“先事实模型与 API，再工作台 UI，再执行链，再视频 / 发布，再 ECS 全流程”的顺序逐模块迁移，每个小任务用 TDD 验证、CI 通过后立即合并 `main`。

**Tech Stack:** Go 1.23.x、React、Ant Design、MySQL、Goose、Redis、TOS、GitHub Actions、Vitest + Testing Library、Go `testing` + `go-sqlmock`。

**Spec / 事实来源:** `TASKS.md`、`docs/migration/v88-batch-factory-v11-frontend-inventory.md`、`docs/migration/v88-batch-factory-v11-backend-api-inventory.md`、`docs/migration/v88-batch-factory-v11-status-model-inventory.md`、`docs/migration/v88-batch-factory-video-provider-inventory.md`、`docs/migration/ecs-v88-capability-audit.md`、`docs/migration/v88-to-new-go-react-module-map.md`。

## Global Constraints

- 后端 / API / Worker 默认必须使用 Go；只有明确技术限制时才允许其它语言，并在计划或 PR 中说明原因。
- 用户端使用 React + Ant Design；管理端使用 React + Ant Design。
- 数据库使用 MySQL，Schema 变更使用 Goose migration。
- Redis 用于 Queue / Lock / Scheduler；不得把核心业务事实只存在内存。
- TOS 用于最终对象存储与大文件产物。
- 最终静态前端由 Go embed 打包进后端二进制，ECS 最终只需部署统一产物。
- 系统预设提示词保存在后端事实源中，运行时由后端拼装真实提示词，不把重要系统 Prompt 固化在前端。
- 所有新能力进入新仓库 `cui1112233/yizhanchengming-xin`；旧 v88 只作为迁移参考，不继续扩展旧代码。
- 不修改旧 v78。
- 同时只保留一个短期功能分支；功能完成且 CI 全绿后立即合并 `main`，不长期堆积 `feat/*` 分支。
- 每个功能必须走 RED → GREEN → 全量测试 → PR CI → 合并 main → main CI。
- 任何任务只有“代码 + 测试 + 行为对标 + main CI”都完成，才算当前阶段完成。
- “公网可用”只能在 Task 16 ECS 真实浏览器 / API 全流程验证后确认，不能用 CI 绿代替生产验收。
- 迁移旧行为时保留有效功能语义，但不原样迁移旧 bug、503、Token 闪退、状态错判等问题。

## Review Focus

1. **多书城 / 多本书同一 BatchProject**：列表和详情不得假设一个项目只有一个 source；必须从真实 books 数据聚合。
2. **部分失败**：单本失败不能拖死其它书，项目 / Run 不能被错误标记为整体成功，错误原因必须可见。
3. **重复点击与恢复**：创建 / 提交 / Worker 执行必须有幂等和 Lock，避免重复 Run、重复视频和重复发布。
4. **认证刷新**：Token 过期 / 刷新不能造成页面闪退或把正常业务错误伪装成登录失效。
5. **服务重启**：Scheduler / Queue / Worker 重启后不能丢失未来任务、处理中任务或最终产物状态。

---

# 一、最终我们要做成什么

最终公网应形成一条清晰业务链：

1. 用户在“小说获取工作台”按书城添加 Book ID。
2. 系统通过 121 获取正文与 bookinfo，写入 MySQL。
3. 系统判断 / 保留书名、书城、platformId、分类、genre、男女频、风格、状态和真实错误。
4. 全部成功后创建 BatchProject + Run；可立即执行，也可设置未来 `run_at`。
5. `/batch-factory` 显示真实批量项目列表，项目进入 V11 主工作台。
6. V11 对项目内所有小说执行剧本、Hook、Director、最终提示词、音频、matchAudio、视频生成、合并和发布。
7. 每一本书、每个步骤、每个 Run 都有独立状态；失败可重试，成功不重复执行。
8. “生产统一设置 / 发布统一设置”在工作台右侧 Drawer 管理；“解析输入”被“版本对应配置档”替代。
9. 版本配置档可同步 121 网站配置与批量风格类型。
10. 视频支持 `personal_api`、`yd2.0-mini`、豆包本地执行器，并统一 VIDEO 状态。
11. 发布和权限体系统一，登录 / Token 刷新稳定。
12. 最后通过 GitHub 构建产物部署到 ECS，完成公网全流程回归并正式切换新仓库版本。

---

# 二、执行方法

每个子任务都使用同一套方法，不允许跳步：

- [ ] **Step A — 对标旧 v88 / ECS**：找到旧页面、接口、字段和正常行为；记录“存在 / 可用 / 已坏 / 未验证”。
- [ ] **Step B — 写 RED**：先写最小失败测试，测试必须只失败在当前缺失行为。
- [ ] **Step C — 验证 RED**：GitHub Actions / 本地测试日志必须证明失败原因正确。
- [ ] **Step D — 最小 GREEN**：只实现当前子任务，不提前把后续功能塞进同一提交。
- [ ] **Step E — 全量回归**：`go test ./...`、前台 `npm test`、前台 build、后台 build。
- [ ] **Step F — 更新 `TASKS.md`**：完成子项勾选，下一执行点向前移动。
- [ ] **Step G — PR 差异检查**：确认没有夹带无关模块。
- [ ] **Step H — PR CI**：全部成功后立即合并。
- [ ] **Step I — main CI**：合并后的 `main` 再跑一次完整 CI。
- [ ] **Step J — 进入下一个小任务**：不保留多个并行开发分支。

---

# 三、Task 9 — Batch Factory V11 主工作台

## 目标

先把 BatchProject 从“数据库里有记录”变成“用户能看到、能进入、能执行、能追踪状态的真实 V11 工作台”。这是后续剧本、音频、视频、发布的承载层。

## 9.1 旧系统盘点

当前已完成，继续把下列文档作为迁移事实参考：

- `docs/migration/v88-batch-factory-v11-frontend-inventory.md`
- `docs/migration/v88-batch-factory-v11-backend-api-inventory.md`
- `docs/migration/v88-batch-factory-v11-status-model-inventory.md`
- `docs/migration/v88-batch-factory-video-provider-inventory.md`
- `docs/migration/ecs-v88-capability-audit.md`
- `docs/migration/v88-to-new-go-react-module-map.md`

## 9.2 批量项目列表

### 要做什么

让 `/batch-factory` 成为真实 BatchProject 列表页，数据直接来自 MySQL，不再依赖假数据或旧 Node 状态。

### 怎么做

- [x] 9.2.1 提供 `GET /api/v1/batch-projects`。
- [x] 9.2.2 `MySQLStore.ListBatchProjects()` 读取真实 `batch_projects`。
- [x] 9.2.3 前端显示 `project.name`。
- [ ] 9.2.4 返回并展示 `sources[]`。来源从 `batch_projects.intake_id -> books.intake_id -> books.source` 去重聚合；不能假设一个项目只有一个书城。
- [ ] 9.2.5 返回 `bookCount`，由 SQL 聚合 `COUNT(books.id)`，避免前端 N+1 查询。
- [ ] 9.2.6 返回项目级男女频摘要。建议保留去重后的 `genders[]`，混合项目必须显示“男频 / 女频混合”而不是随机挑一个值。
- [ ] 9.2.7 返回项目级风格摘要。建议保留去重后的 `styles[]`，不覆盖单书真实风格。
- [ ] 9.2.8 返回最新 Run 状态。读取项目最新 Run；无 Run 时显示“未执行”。
- [ ] 9.2.9 Task 8 创建成功后项目无需人工同步数据，刷新 `/batch-factory` 即能从同一 MySQL 事实源看到。

### 主要文件

- `api/internal/intake/model.go`
- `api/internal/intake/mysql_store.go`
- `api/internal/intake/mysql_store_test.go`
- `api/internal/httpapi/batch_project_handlers.go`
- `api/internal/httpapi/batch_project_handlers_test.go`
- `前台/src/BatchProjectListPage.jsx`
- `前台/src/App.test.jsx`
- `前台/src/api.js`

### 验收

- 多书城项目能同时显示多个来源。
- 0 本 / 1 本 / 多本数量准确。
- 男女频、风格来自真实 Book 数据，不由前端猜测。
- 最新 Run 状态准确。
- 项目列表 API 一次请求即可渲染，不产生每个项目额外请求。

## 9.3 项目详情 / V11 工作台入口

### 要做什么

点击项目进入独立 V11 工作台，展示该 BatchProject 所属 Intake 的所有小说和每本真实状态。

### 怎么做

- [ ] 增加 `GET /api/v1/batch-projects/{id}`。
- [ ] 增加 `GET /api/v1/batch-projects/{id}/books` 或在详情 API 返回稳定嵌套模型。
- [ ] 路由 `/batch-factory/:id` 显示项目详情。
- [ ] 展示 Book ID、书名、书城、男女频、风格、正文状态、错误原因。
- [ ] 项目不存在返回 404；数据库错误返回 500，不把底层 DSN / password 泄露给浏览器。

### 验收

- 点击列表项目能进入真实详情。
- 多书城、多状态数据与 MySQL 一致。
- 单本失败能直接看到失败书和错误原因。

## 9.4 Run 执行框架

### 要做什么

把当前只“创建 Run 记录”的能力升级为真正可执行、可恢复、可重试的 Batch Factory 运行框架。

### 怎么做

- [ ] 明确定义 Run 状态：`pending -> running -> completed/failed`。
- [ ] 为单书执行建立 Item / Step 状态模型，不能只有项目级状态。
- [ ] 单本失败不阻断其它书；最终 Run 根据所有 item 聚合结果。
- [ ] 错误写入数据库，前端可读。
- [ ] 单本重试只重试失败 item，不重复成功 item。
- [ ] 创建 Run / 重试加入幂等 key 与 Redis Lock。
- [ ] 自动化 `run_at` 到点后由 Scheduler 将任务投递 Redis Queue。
- [ ] Worker 消费队列并更新 MySQL 状态。
- [ ] Worker crash 后依据 MySQL 状态 + lease / heartbeat 恢复未完成任务。

### 验收

- 重复点执行不会产生重复生产链。
- 服务重启后未来任务仍存在。
- 一个项目 10 本书，1 本失败时其它 9 本仍可成功。
- 失败原因、重试次数和最终状态可追踪。

---

# 四、Task 10 — 统一设置 / 版本对应配置档

## 目标

整理旧 Batch Factory 配置入口，把重复、分散、难理解的配置收敛成工作台内可维护的统一设置。

## 要做什么

- “生产统一设置”右侧宽 Drawer。
- “发布统一设置”右侧宽 Drawer。
- 删除“解析输入”。
- 新增“版本对应配置档”。
- 支持同步 121 网站配置。
- 支持同步批量风格类型。
- 清理 AI 文案数量和“AI 文案处理优先方案”等重复配置。
- 明确处理规则提示词、知识库提示词的所属位置。

## 怎么做

- [ ] 先从旧 v88 找出所有配置字段、默认值、真实调用点，建立配置字段映射表。
- [ ] 新建 Go 配置模型 + Goose migration；前端不作为配置事实源。
- [ ] 新建读 / 写 API，写入时做版本号或 updated_at 控制。
- [ ] Drawer 保存成功后留在当前项目工作台，不跳页面。
- [ ] 配置档同步 121 时只更新明确同步字段，不能覆盖用户手工配置。
- [ ] 清理重复设置时先写兼容迁移，避免旧项目读不到配置。

## 验收

- 页面不再出现旧“解析输入”。
- 两个统一设置均在 Drawer 完成。
- 刷新后配置不丢。
- 同步 121 / 风格类型后数据一致。
- 不存在两处互相覆盖同一个 AI 文案配置的情况。

---

# 五、Task 11 — 水货生产工作台

## 目标

迁移 `/shuihuo-production`，并让它和新 Batch Factory V11 共用同一 MySQL / Run 事实源，彻底消除旧公网“页面能开但 Go service 503 / 点击闪退”的割裂。

## 怎么做

- [ ] 重建 `/shuihuo-production` React 页面。
- [ ] 点击 Batch Factory 进入新 `/batch-factory`，不跳旧服务。
- [ ] 共用项目列表 API 和状态模型。
- [ ] 处理加载失败时显示错误，不黑屏、不全页闪退。
- [ ] 对照旧 ECS 可用交互，把旧 bug 排除在迁移外。

## 验收

- 页面直接打开不闪退。
- 可看到新仓库真实项目。
- V11 入口稳定。
- 后端不可用时显示明确错误，不把用户踢回登录页。

---

# 六、Task 12 — 剧本生成 / Hook / Director / 最终提示词

## 目标

恢复旧 v88 的内容生产核心，并把 Prompt、版本和执行状态都纳入 Go / MySQL 管理。

## 怎么做

- [ ] 定义脚本生成请求 / 响应模型。
- [ ] 系统 Prompt 存后端，按版本读取。
- [ ] 接入“剧情模式”通用元提示词：保持稳定结构，剧情、空间、动作和情绪由模型根据小说变化。
- [ ] Hook 独立步骤。
- [ ] Director 独立步骤。
- [ ] 最终提示词编译步骤，将角色 / Hook / Director / 约束组合成最终模型输入。
- [ ] 每本书每一步持久化状态、版本、错误、重试次数。
- [ ] 批量项目聚合状态来自单书步骤状态。

## 验收

- 系统 Prompt 不依赖前端硬编码。
- 同一 Prompt 版本可追溯。
- 单书脚本失败不影响其它书。
- Director / Hook / final prompt 都能看到真实执行状态。

---

# 七、Task 13 — 音频 + matchAudio

## 目标

把第四步原“快速导演分镜”替换为可开关的“匹配音频”逻辑，使最终镜头总时长严格匹配真实音频。

## 硬规则

当 `matchAudio=true`：

- 总时长必须严格等于 `audioDurationSec`。
- 第一镜必须从 `0.00s` 开始。
- 所有镜头时间必须落在 `[0, audioDurationSec]`。
- 镜头之间不得有空缺。
- 镜头之间不得重叠。
- 最后一镜结束时间必须严格等于音频时长。
- 根据对白 / 旁白 / 停顿 / 节奏分配时间。
- 不得为了凑时长添加无关剧情。
- 秒数统一保留两位小数。
- 10s 规则继续保留。
- 选择 15s 的含义是“单镜 ≤15s”，不是每镜必须 15s。

当 `matchAudio=false`：

- 保留原 Director 逻辑，不强制匹配音频总长。

## 怎么做

- [ ] 音频步骤持久化真实 `audioDurationSec`。
- [ ] Director 输入加入 `matchAudio` 和 `audioDurationSec`。
- [ ] 后端 Prompt 模板加入硬时长约束。
- [ ] 生成后增加 Go 校验器二次验证时间轴；不只相信模型输出。
- [ ] 校验失败进入 retryable 状态，不能把非法时间轴继续送视频生成。

## 验收

例如音频 `28.00s`：首镜 `0.00`，末镜严格 `28.00`，中间无空档、无重叠，任何一镜不违反用户选择的 10s / ≤15s 上限。

---

# 八、Task 14 — 视频生成完整链路

## 目标

迁移并统一旧公网 personal_api / yd2.0-mini / 豆包本地执行器，建立稳定的视频任务状态机、轮询、重试、合并和最终产物回写。

## 怎么做

- [ ] 建立 `video_tasks` / 对应模型和 Goose migration。
- [ ] 统一 `VIDEO` 状态，例如 `pending / submitted / processing / succeeded / failed / retryable_failed`。
- [ ] Provider 抽象统一提交和查询接口。
- [ ] 实现 `personal_api` provider。
- [ ] 实现 `yd2.0-mini` provider。
- [ ] 接入豆包本地执行器，并明确本地执行结果如何回传公网 Go 服务。
- [ ] 角色 / 场景 / 图片资产通过结构化引用传递，避免仅靠提示词文本猜资产。
- [ ] 视频提交使用幂等 key。
- [ ] 状态轮询写回 MySQL。
- [ ] 失败按可重试 / 不可重试分类。
- [ ] 成功产物写 TOS 或稳定对象位置。
- [ ] ffmpeg 合并由 Worker 执行，不阻塞 HTTP 请求。
- [ ] 最终视频写回 BatchProject / Book。

## 关键旧问题必须修复

- `/api/batch-factory/v11/video-provider/status?provider=personal_api` 不得再返回旧公网 503。
- VIDEO 状态必须来自统一 Go 服务，不依赖旧 Node/Go 双服务互相猜状态。

## 验收

- 至少一个 provider 能从提交到成功完整闭环。
- Provider 失败不会卡死整个 BatchProject。
- 合并失败有真实错误、可重试。
- 服务重启后状态可恢复。

---

# 九、Task 15 — 发布 / 权限 / 登录认证

## 目标

统一登录 Token、API 权限和发布权限，解决旧公网“登录状态已过期闪退”和接口权限不一致问题。

## 怎么做

- [ ] 明确 access token / refresh token 生命周期和刷新流程。
- [ ] 前端只在真实认证失败时进入登录流程，业务 500/503 不得被误判为登录失效。
- [ ] 后端统一鉴权 middleware。
- [ ] 梳理公开白名单和必须鉴权 API。
- [ ] 修复 `/api/script-constraint-prompts?category=prefix` 的旧 404 / Unauthorized 兼容问题。
- [ ] 发布前校验用户 / 团队权限。
- [ ] 单个发布任务权限失败不得影响其它任务。
- [ ] 发布状态持久化。

## 验收

- 刷新页面不会随机退出登录。
- Token 过期能无感刷新或明确重新登录。
- 业务 503 不会触发“登录过期”。
- 发布权限失败有明确原因。

---

# 十、Task 16 — ECS 部署与最终公网验收

## 目标

只有这一阶段结束，才能宣布新仓库“公网可用”。

## 16.1 构建 / 部署

- [ ] GitHub Actions 生成可部署产物。
- [ ] 前台 / 管理端 production build。
- [ ] Go embed 打包静态资源。
- [ ] 产出单一 Go 二进制或明确版本化部署包。
- [ ] ECS 旧 v88 文件和服务配置备份。
- [ ] MySQL 完整备份。
- [ ] Goose migration 预演 / dry check。
- [ ] 验证 `QIANTIE_MYSQL_DSN`。
- [ ] 验证 Redis 配置。
- [ ] 验证 TOS 配置。
- [ ] 启动新 Go 服务。
- [ ] `/healthz` / 关键 API 健康检查。
- [ ] 明确一条命令可回滚旧版本。

## 16.2 公网真实浏览器 / API 验收

依次验证，不跳项：

- [ ] 登录与刷新。
- [ ] 小说获取工作台。
- [ ] 多书城添加。
- [ ] 121 正文获取。
- [ ] 书名 / platformId / 男女频 / 风格。
- [ ] 立即执行。
- [ ] 自动化执行。
- [ ] Batch Factory V11 列表与详情。
- [ ] 水货生产。
- [ ] 版本对应配置档。
- [ ] 剧本生成。
- [ ] Hook。
- [ ] Director。
- [ ] 音频。
- [ ] matchAudio。
- [ ] personal_api 视频。
- [ ] yd2.0-mini。
- [ ] 豆包本地执行器。
- [ ] VIDEO 状态。
- [ ] ffmpeg 合并。
- [ ] 发布。
- [ ] 用户 / 团队权限。
- [ ] 单本失败和重试。
- [ ] 服务重启恢复。
- [ ] 登录 Token 刷新后的页面状态保持。
- [ ] 与旧 ECS 核心功能逐项核对。

## 16.3 正式切换

只有以上验收全部通过后：

- [ ] 将公网流量 / 服务入口切到新仓库产物。
- [ ] 保留旧 v88 作为短期回滚包，不再继续开发。
- [ ] 记录正式上线 commit SHA、镜像 / 二进制版本、DB migration 版本。
- [ ] 观察错误日志、队列积压、视频失败率和认证错误。

---

# 十一、分支、提交和 CI 规则

## 分支

- 任意时间只开发一个短分支。
- 分支从最新 `main` 创建。
- 当前小任务通过后立即合并。
- 不为了“以后可能会用”长期保留大量 feat 分支。

## 提交

推荐保持以下粒度：

1. `test: define ...` — RED。
2. `feat: implement ...` / `fix: ...` — GREEN。
3. `docs: update migration progress` — TASKS / 文档。

## CI 最低门槛

每个 PR 必须全部通过：

- `go test ./...`
- 用户端 `npm test`
- 用户端 `npm run build`
- 管理端 `npm run build`

PR 绿后合并 `main`，合并后的 push CI 还要再绿一次。

---

# 十二、数据事实源规则

- `Intake`：一次小说获取批次。
- `Book`：每一本真实小说；保存 source、platformId、Book ID、书名、正文、分类、genre、男女频、风格、状态和错误。
- `BatchProject`：生产项目，关联唯一 Intake。
- `Run`：一次立即 / 自动化执行。
- 后续 `Item / Step / VideoTask / PublishTask` 必须指向明确 Book / Run，不使用模糊全局状态。
- 项目级展示字段（书城、男女频、风格、数量、运行状态）优先通过 SQL 聚合或稳定后端 DTO 返回，不在前端进行大量 N+1 请求。
- Redis 不是最终事实源；Redis 数据丢失后应可根据 MySQL 重建 / 恢复任务。

---

# 十三、错误处理规则

- 数据库错误：HTTP 500，前端显示通用可理解错误，不泄露 DSN / password / SQL 机密。
- 依赖尚未接线：HTTP 503，明确“服务暂不可用”，不能伪装成空列表或成功。
- 121 单书失败：该 Book `retryable_failed`，其它书继续。
- Provider 单视频失败：该 VideoTask 失败 / 可重试，其它视频继续。
- 权限失败：明确 401/403，与业务 500/503 区分。
- 不允许前端看到“成功”但后台实际没有落库。
- 不允许后台失败后前端自动把用户踢到登录页，除非认证确实无效。

---

# 十四、当前执行位置

截至 2026-10-05：

- Task 1–8 已进入 `main` 并完成当前阶段 CI 验收。
- Task 9.1 盘点已完成。
- Task 9.2.1 BatchProject 列表 API 已完成。
- Task 9.2.2 MySQL 真实项目读取已完成。
- Task 9.2.3 项目名称展示已完成。
- **当前正在执行：Task 9.2.4 — 展示书城来源。**
- 当前短分支：`task9-batch-project-source-ui`。
- 当前 PR：#11。
- 9.2.4 已完成 RED：测试已经证明旧 `ListBatchProjects()` 尚未从 `books.source` 聚合项目来源。
- 下一步 GREEN：`BatchProject` 增加只读 `Sources []string`，MySQL 使用一次 JOIN + 去重聚合来源，然后 HTTP / 前端展示 `sources[]`。

---

# 十五、每个大 Task 的完成定义

一个大 Task（例如 Task 9）只有全部满足以下条件才允许在 `TASKS.md` 标为完成：

- [ ] 所有子任务完成。
- [ ] Go tests 全绿。
- [ ] 用户端 tests 全绿。
- [ ] 用户端 build 全绿。
- [ ] 管理端 build 全绿。
- [ ] 与旧 v88 正常行为对标完成。
- [ ] 相关 PR 已合并 `main`。
- [ ] 合并后 `main` CI 全绿。
- [ ] 如果属于公网能力，Task 16 最终还需真实 ECS 验收；在此之前只能说“代码阶段完成”，不能说“公网已经可用”。

---

# 十六、执行顺序（禁止随意跳跃）

默认按以下顺序推进：

1. **Task 9** — 先把 Batch Factory V11 主工作台和执行框架打稳。
2. **Task 10** — 收口生产 / 发布设置和版本对应配置档。
3. **Task 11** — 迁移水货生产入口并共用 V11 事实源。
4. **Task 12** — 剧本 / Hook / Director / final prompt。
5. **Task 13** — 音频和 matchAudio。
6. **Task 14** — 视频 provider / VIDEO / 合并链。
7. **Task 15** — 发布 / 权限 / 登录认证统一。
8. **Task 16** — 构建、ECS 部署、真实公网全流程、正式切换。

除非出现会阻断当前 Task 的基础设施问题，否则不提前跳到后续模块，以免形成“每块都有一点、没有一块真正能用”的状态。

---

# 十七、执行原则

**我们不是在做“把页面搬过去”，而是在做“把旧公网真正能用的业务能力迁成一个可维护的新主线”。**

判断任何工作是否值得做，只看四个问题：

1. 它是不是最终公网需要的真实能力？
2. 它有没有明确事实源和状态模型？
3. 它能不能用测试证明正确？
4. 它最终能不能在 ECS 浏览器 / API 中真实验收？

如果答案不是“是”，就不为了表面进度提前堆代码。
