# v88 Task 9–15 功能复用与 Go 化审计

审计日期：2026-10-05

本文件回答一个核心问题：Task 9–15 到底哪些能力旧 v88 已经写过，哪些只是需要迁移/Go 化，哪些才是真正从零缺失。

> 重要：本审计只判断“功能资产与迁移方式”，**不代表任务已经完成**。`TASKS.md` 的 `[x]` 标准保持不变：代码进入 `main`、测试/build 通过、完成 v88 行为对标；涉及公网的能力还必须在 Task 16 完成 ECS/浏览器/真实接口验收。

## 1. 新主线技术栈硬约束

后续 Task 9–15 一律遵守：

- 后端、API、Worker、Scheduler、Provider、发布服务：**Go**。
- 用户端、管理端：**React + Ant Design**。
- 长期业务事实源：**MySQL**。
- Schema migration：**Goose**。
- Redis：只用于 Queue、延迟任务、Lock、短期 lease/heartbeat 等运行基础设施，不作为长期事实源。
- 媒体长期对象：TOS；必要时保留受控本地产物层。
- 最终前端静态资源由 Go `embed` 收口进二进制。
- 旧 v88 的 Node/JavaScript 后端只作为行为、接口和兼容逻辑参考，**不得直接成为新主线生产后端**。
- 如果旧能力只存在 Node 中，状态应标为“旧 v88 已有，但需 Go 重写”，而不是直接复制 Node。

## 2. 审计标签

| 标签 | 含义 |
|---|---|
| `♻️ 可复用` | 旧 v88 已有明确业务语义/React UI/Go 服务，可以迁移行为或组件 |
| `🔁 Go 化` | 旧 v88 已有能力，但旧实现依赖 Node、中间网关、内存事实源或多事实源，需要用 Go/MySQL 重写 |
| `🧩 新仓部分已有` | 新仓已经有数据模型/API/页面的一部分，可在现有实现上继续补齐 |
| `⬜ 真缺失` | 没有足够旧实现可直接迁，属于真正新增基础设施或新需求 |
| `🌐 待部署验收` | 源码存在或迁移完成后，仍必须用 ECS/真实账号/真实 Provider 验证 |

## 3. 总结结论

Task 9–15 **并不是七个从零开发的大任务**。

旧 v88 已经存在完整的 Batch Factory UI、Batch/Book/Settings、Hook、Director、Stages、Final Prompt Compiler、VIDEO Production、多 Provider、本地执行器、Merge、外部发布、自动化控制器和权限/认证相关链路。旧公网真实架构是：

```text
React
  -> Node V12/V11 可信网关
  -> Go V11
  -> MySQL/Provider/TOS/本地执行器
```

新主线真正要做的是：

```text
React + Ant Design
  -> Go API
  -> MySQL / Redis
  -> Go Worker / Provider / TOS
```

因此后续不能再把“旧 v88 已有功能”当成空白需求重新设计；应优先迁语义、迁 React 交互、迁旧测试表达的行为，然后把 Node 可信职责收敛进 Go。

---

# Task 9：Batch Factory V11 主工作台

## 9.2 批量项目列表

| 子任务 | 审计 | 结论 |
|---|---|---|
| 9.2.4 书城来源 | `🧩 新仓部分已有` | 新仓 Book 已有 `source/platformId`，只缺 BatchProject 聚合查询与前端列 |
| 9.2.5 小说数量 | `🧩 新仓部分已有` | 新仓已有 BatchProject ↔ Intake/Book 数据，只缺聚合 count |
| 9.2.6 男女频 | `🧩 新仓部分已有` | Task 3/8 已有 gender，列表层未聚合 |
| 9.2.7 风格 | `🧩 新仓部分已有` | Task 3/8 已有 style，列表层未聚合 |
| 9.2.8 运行状态 | `🧩 新仓部分已有 + ♻️` | 新仓已有 Run；旧 v88 有 runtime/stage/production 聚合语义，需按新模型收敛 |
| 9.2.9 新建后立即出现 | `🧩 新仓部分已有` | Task 8 已创建真实 BatchProject，缺列表刷新/一致性验收 |

结论：9.2 不是重做批量工厂，只是在现有 Go/MySQL API 上补聚合字段和 React 展示。

## 9.3 项目详情

9.3.1–9.3.9：`♻️ 可复用 + 🔁 Go 化`

旧 v88 已有：

- `BatchFactoryWorkbenchPage.jsx`
- `BatchFactoryV11UiPage.jsx`
- `BatchFactoryV11Workbench.jsx`
- `BatchFactoryNovelList.jsx`
- Novel source / status / error / settings / production 相关工作台组件
- Batch/Book 查询与 runtime API

新主线应复用 UI/交互语义，但详情事实源统一从新 Go + MySQL 查询，不把旧前端状态或 Node 聚合当事实源。

## 9.4 Run 执行框架

| 子任务 | 审计 | 结论 |
|---|---|---|
| 9.4.1 pending -> running | `♻️ + 🧩` | 新仓已有 Run 状态；旧 v88 有 Stage/Production 状态机 |
| 9.4.2 单本独立状态 | `♻️` | 旧 v88 有 BookStageRun/ProductionTask 语义，需要迁为 BookRun/StageRun |
| 9.4.3 一本失败不阻断其他书 | `♻️` | 旧批处理/production 已有失败隔离语义 |
| 9.4.4 错误持久化 | `♻️ + 🔁` | 旧模型保留 errorMessage；新主线落 MySQL |
| 9.4.5 单本重试 | `♻️` | 旧 stage retry/attempt 语义可迁 |
| 9.4.6 幂等保护 | `♻️ + 🔁` | 旧 requestId/attempt 可复用，需新 Go/MySQL 唯一约束 |
| 9.4.7 到点 Worker | `⬜ 真缺失` | 新仓现在仅保存未来 `run_at`，真正 Scheduler/Worker 尚未完成 |
| 9.4.8 Redis Queue | `⬜ 真缺失` | 新主线新增运行基础设施 |
| 9.4.9 Redis Lock | `⬜ 真缺失` | 新主线新增运行基础设施 |
| 9.4.10 崩溃恢复 | `⬜ 真缺失` | 需要基于 MySQL 状态 + Redis lease/recovery 设计 |

**真正从零的主要是 9.4.7–9.4.10，不是整个 Task 9。**

---

# Task 10：批量工厂统一设置 / 版本对应配置档

| 子任务 | 审计 | 结论 |
|---|---|---|
| 10.1–10.3 生产/发布 Drawer 与保存上下文 | `♻️ 可复用` | 旧 v88 已有 `BatchFactoryV11SettingsDrawers`、`BatchFactoryV11PublishSettings`、`BatchFactoryUnifiedSettingsModal` 等 React 资产 |
| 10.4 删除“解析输入” | `♻️/淘汰旧入口` | 这是明确迁移规则，不需重新设计旧功能 |
| 10.5 版本对应配置档 | `♻️ + 🔁` | 旧 v88 已有 `config-versions` API/配置版本语义，但新需求要收敛成正式配置档 |
| 10.6 同步 121 配置 | `♻️ + 🔁` | 旧 121/Batch 设置链可参考，事实源改 Go/MySQL |
| 10.7 同步批量风格类型 | `♻️ + 🔁` | 可复用旧 settings/config 行为，写入新配置档 |
| 10.8–10.11 清理重复配置/明确提示词位置 | `♻️ + 重构` | 旧 UI/行为已存在，本轮是去重和信息架构收敛 |
| 10.12 配置 MySQL 事实源 | `🔁 Go 化` | 不保留 Node/local/memory 多事实源 |
| 10.13 保存刷新一致 | `🔁 Go 化` | 必须以后端持久化结果为准 |
| 10.14 测试/CI | 待实施 | 迁移旧测试表达的行为并补新测试 |

结论：Task 10 大部分是**旧设置体系的收敛/Go 化**，不是空白功能。

---

# Task 11：水货生产工作台

| 子任务 | 审计 | 结论 |
|---|---|---|
| 11.1 `/shuihuo-production` 页面 | `♻️ 可复用` | 旧 v88 页面和相关 React/CSS 已存在 |
| 11.2 进入 Batch Factory | `♻️ 可复用` | 旧页面/批量工厂已有互通行为 |
| 11.3 路由不闪退 | `♻️ + 修复` | 迁移现有路由，不能带回旧登录/状态闪退问题 |
| 11.4 真实项目列表 | `🔁 Go 化` | 新仓统一 MySQL BatchProject，不再依赖旧页面态 |
| 11.5 状态同步 | `♻️ + 🔁` | 旧 runtime summary/production status 可迁，查询改 Go |
| 11.6 失败可见 | `♻️` | 旧 stage/production error 语义已有 |
| 11.7 ECS 对标 | `🌐 待部署验收` | 必须真实浏览器回归 |
| 11.8 测试/CI | 待实施 | 可参考旧页面 source tests |

结论：水货生产页面不是从零；主要工作是**迁 React + 接新 Go/MySQL 事实源**。

---

# Task 12：剧本生成 / Director / Hook / 最终提示词

这是“旧 v88 已经写得最多”的任务之一。

旧 v88 已确认存在：

- 单书 Hook 生成
- Hook approve
- working-front viral rewrite
- 单书 Director
- 批量 Director
- opening variants
- Book Stage run/retry
- H3 Director
- H3 compile/trace
- Effective Settings
- Final Prompt Compiler
- Director/Hook/FinalPrompt 对应 React UI

| 子任务 | 审计 | 结论 |
|---|---|---|
| 12.1 剧本生成 Go API | `♻️ + 🔁` | 旧 V11 Go/Node 已有生成链；新 API namespace/服务需收敛到 Go |
| 12.2 系统提示词后端管理 | `♻️ + 🔁` | 旧 presets/script-constraint-prompts 已有，Node enrichment 改 Go prompt resolver |
| 12.3 剧情模式通用元提示词 | `⬜/需落实新需求` | 需求已经明确，但不能仅凭旧 Director 存在就宣称已完成 |
| 12.4 Hook | `♻️ 可复用` | 旧 Go API + React UI 已有 |
| 12.5 Director | `♻️ 可复用` | 旧 DirectorService/单书/批量/H3 已有 |
| 12.6 最终提示词编译 | `♻️ 可复用` | 旧 PromptCompilerService 已有 |
| 12.7 单本生成状态 | `♻️ + 🔁` | 旧 StageRun 可迁入新 workflow |
| 12.8 批量生成状态 | `♻️ + 🔁` | 旧批量 Director/runtime 聚合可迁 |
| 12.9 失败重试 | `♻️ 可复用` | 旧 stage retry/attempt 语义明确 |
| 12.10 提示词版本持久化 | `♻️ + 🔁` | 旧 revision/config/compile trace 可复用，持久化统一 MySQL |
| 12.11 v88 对标 | `♻️` | 有大量旧 API/UI/tests 可作对标基线 |
| 12.12 测试/CI | 待实施 | 不能因为旧代码存在就勾选 |

结论：**Hook、Director、最终提示词绝不是从零写。** 应优先迁旧 Go 业务语义和 React 交互，只重写 Node 注入/凭据/预设解析部分。

---

# Task 13：音频 + matchAudio

这里必须与 Task 12 区分。

旧 v88 可以确认存在：

- `h3LineAudio.js` 等前端音频/H3 辅助资产
- 原生 V12 Go `h3/audio-measurement`
- H3 Director / compile / trace
- 旧 Director 时长/视频约束语义

但目前没有足够证据证明后来定义的完整 `matchAudio=true` 精确规则已经全部进入旧生产代码。因此不能为了“看起来完成度高”而误标。

| 子任务 | 审计 | 结论 |
|---|---|---|
| 13.1 第三步生成音频 | `♻️/需继续核验` | 有音频/H3资产，但正式新链仍需接入 |
| 13.2 audioDurationSec | `♻️ 可复用` | 旧 Go H3 audio-measurement 已存在 |
| 13.3 匹配音频开关 | `⬜ 待落实` | 按最新产品要求实现，不假设旧版完整存在 |
| 13.4–13.8 精确总时长/连续镜头/末镜结束 | `⬜ 待落实` | 属于明确的新硬约束，需 Go/Director 测试证明 |
| 13.9 10s 规则 | `♻️` | 旧时长约束可参考 |
| 13.10 15s 为 ≤15s | `⬜/规则修正` | 以最新需求覆盖旧行为 |
| 13.11 matchAudio=false 保留旧 Director | `♻️ + 新开关` | Director 可复用，分支控制需新增 |
| 13.12 两位小数 | `⬜ 待落实` | 新输出约束 |
| 13.13 不凑无关剧情 | `⬜ 待落实` | 新 prompt/validator 约束 |
| 13.14 测试/CI | 待实施 | 精确时间规则必须有边界测试 |

结论：Task 13 不是完全从零，但 **matchAudio 的最终硬规则仍是真正待实现项**。

---

# Task 14：视频生成完整链路

这是另一个旧 v88 已有大量实现的任务。

旧 v88 已确认存在：

- ProductionJob / ProductionTask / VIDEO 状态机
- `personal_api`
- `yd2.0-mini`
- `doubao_local_executor`
- `autodl_comfyui` / H3
- `yfai_seedance`
- provider/model 强绑定
- Submit / Poll / Cancel
- providerTaskId / mediaUrl / errorMessage
- 角色/场景参考图传递
- Local Executor pairing/lease/artifact
- MergeService / MergeJob
- ffmpeg 本地合并
- TOS 产物
- merge cover
- 外部发布前产物链

| 子任务 | 审计 | 结论 |
|---|---|---|
| 14.1 视频任务模型 | `♻️ 可复用` | 旧 ProductionJob/Task 字段非常完整，迁 MySQL 模型 |
| 14.2 VIDEO 状态机 | `♻️ 可复用` | queued/running/succeeded/failed/cancelled 已存在 |
| 14.3 personal_api | `♻️ + 🔁` | Adapter 可迁；凭据解析/持久化必须 Go 化 |
| 14.4 yd2.0-mini | `♻️ 可复用` | 旧模型映射已存在 |
| 14.5 豆包本地执行器 | `♻️ 可复用` | pairing/adapter/artifact 协议已有 |
| 14.6 provider 状态 API | `♻️ + 🔁` | 接口已有，但诊断与持久化需修复 |
| 14.7 资产传递 | `♻️ 可复用` | 旧 asset/referenceImageUrls 语义已有 |
| 14.8 视频提交 | `♻️ 可复用` | ProductionService + adapters 已有 |
| 14.9 状态轮询 | `♻️ 可复用` | Yadi/AutoDL/YFAI/Local 均有轮询/状态映射 |
| 14.10 失败重试 | `♻️ + 🔁` | attempt/error 语义已有，接新 workflow |
| 14.11 TOS/本地产物 | `♻️ + 🔁` | 旧能力存在，收敛 storage 包 |
| 14.12 视频合并 Worker | `♻️ + 🔁` | MergeService 已有；新 Worker/Queue 接法需 Go 化 |
| 14.13 ffmpeg 合并 | `♻️ 可复用` | 旧本地合并链已存在 |
| 14.14 回写 BatchProject | `♻️ + 🔁` | 旧状态/产物回写语义可迁到新 MySQL |
| 14.15 personal_api 不再 503 | `🔁 + 🌐` | 必须修复旧架构缺陷后真实 ECS/provider 验证 |
| 14.16 测试/CI | 待实施 | provider adapter/状态/错误需回归 |

## 14.x 必须新增的修复：Provider 持久化

旧 Go 使用 `MemoryVideoProviderRegistry`，以 `owner + provider` 为 key，但只存在内存：

- Go 重启后配置丢失
- 多实例状态不共享
- 旧 Node 被迫在生产/status 边界重新同步 secret

这正是“旧功能存在但不能原样迁”的典型。

新主线必须：

- Provider/credential owner 隔离
- Go 服务端解析 secret
- 敏感凭据加密落库或接统一凭据事实源
- Go 重启后仍可恢复配置
- 浏览器永远不接触 API Key
- “未配置”与“服务不可达”返回不同诊断
- model/provider 不匹配返回明确冲突

因此 Task 14 的主体不是重新写视频功能，而是**迁已有 Production/Provider/Merge 能力 + 修掉旧 registry/Node bridge 架构问题**。

---

# Task 15：发布 / 权限 / 登录认证

旧 v88 已有：

- Go external publish credential/intents/audits
- Node 121 organization/session/publish bridge
- publish metadata classify
- 发布权限校验
- `apiAuth`
- Node -> Go `BridgeAuth`
- account/member/team 相关逻辑

但历史公网也明确出现过：

- 登录后“当前登录状态已过期”跳回登录页
- `/api/script-constraint-prompts` 从 404 修到路由存在后又出现 invalid/expired token
- V11 Go bridge/provider 相关 503

| 子任务 | 审计 | 结论 |
|---|---|---|
| 15.1 Token 体系统一 | `🔁 Go 化` | 旧 auth 存在但多层，需统一 Go auth/session |
| 15.2 Token 刷新 | `♻️ + 🔁` | 旧会话逻辑可参考，按统一 token 体系重做 |
| 15.3 登录闪退 | `🔁 + 🌐` | 属于旧公网明确缺陷，必须修复并实测 |
| 15.4 API 白名单 | `🔁 Go 化` | 不能简单全局放开，要按路由/auth policy 梳理 |
| 15.5 script-constraint-prompts 权限 | `♻️ + 🔁 + 🌐` | 旧接口存在，认证兼容需修复 |
| 15.6 发布权限验证 | `♻️ + 🔁` | 旧逻辑存在，迁到 auth/permission + publish |
| 15.7 用户权限 | `♻️ + 🔁` | 旧 account/member 逻辑可参考，MySQL/Go 统一 |
| 15.8 团队权限 | `♻️ + 🔁` | 同上 |
| 15.9 发布任务状态 | `♻️ + 🔁` | 旧 publish intent/audit 可迁为 MySQL publish job |
| 15.10 权限失败隔离 | `♻️ + 🔁` | 保留单任务失败不拖死其它任务语义 |
| 15.11 测试/CI | 待实施 | auth/permission/publish 必须补集成测试 |

结论：Task 15 也不是从零；主要是**把旧 Node + Go 双层认证/发布链收敛为统一 Go 服务，并修复历史会话问题**。

---

# 4. 真正需要从零优先补的能力

经过 Task 9–15 重新盘点，真正不能靠“迁旧代码”直接解决的重点主要是：

1. `Run.run_at` 到点执行 Scheduler / Go Worker。
2. Redis Queue。
3. Redis distributed Lock。
4. Worker 崩溃/重启后的 lease/recovery。
5. 新 MySQL `BookRun/StageRun` 等工作流事实模型（可迁旧语义，但新表/仓储仍要实现）。
6. Provider credential 的持久化/加密/owner 隔离，用来替代旧 `MemoryVideoProviderRegistry`。
7. 最新版 `matchAudio` 精确时长规则和对应 validator/tests。
8. 剧情模式通用元提示词的正式后端版本化落地。
9. 统一 Go auth/session/token 刷新体系。
10. 最终 Task 16 的 Go embed、自动构建、ECS 部署和公网真实回归。

其它大量工作应优先定义为“迁移/重构/Go 化”，而不是重新开发。

# 5. 建议执行顺序

为了最快恢复到旧 v88 已有能力，同时不把旧架构问题带回来：

1. 先完成 Task 9.2–9.3：让 Task 8 创建的真实项目完整进入 Batch Factory。
2. 紧接 Task 12：迁 Hook / Director / Prompt Compiler，因为旧实现和测试资产最多。
3. Task 10：迁统一设置 + 版本对应配置档，并把事实源落 MySQL。
4. Task 11：迁水货生产页面，接同一 BatchProject/Run 事实源。
5. Task 13：复用 H3 audio measurement，再补完整 matchAudio 硬规则。
6. Task 14：迁 Production/Provider/Merge，并先修持久化 Provider registry。
7. Task 15：统一 Go Auth/Publish/Permission。
8. 期间补 Task 9.4.7–9.4.10 的 Scheduler/Redis/Recovery 基础设施；最终进入 Task 16 公网切换。

# 6. 审计依据

本结论基于：

- `docs/migration/v88-batch-factory-v11-frontend-inventory.md`
- `docs/migration/v88-batch-factory-v11-backend-api-inventory.md`
- `docs/migration/v88-batch-factory-v11-status-model-inventory.md`
- `docs/migration/v88-batch-factory-video-provider-inventory.md`
- `docs/migration/ecs-v88-capability-audit.md`
- `docs/migration/v88-to-new-go-react-module-map.md`
- 旧仓库 `cui1112233/-` 的 `v88` 分支

## 最终原则

**以后看到 Task 9–15 的 `[ ]`，必须先查本审计，不能直接解释成“功能从来没写过”。**

`[ ]` 只表示“新主线尚未达到正式完成标准”。它可能是：旧功能待迁、旧 Node 待 Go 化、新仓部分已写、真缺失，或仅剩部署验收。
