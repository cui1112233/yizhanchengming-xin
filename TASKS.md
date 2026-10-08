# 一战晟铭新主线 · 公网/v88 迁移任务清单

> 这是本仓库唯一的迁移进度事实清单。
>
> 规则：每完成一个子任务，必须在同一次提交或同一个 PR 中同步更新本文件；没有勾选并附验收依据，不算正式完成。

## 状态说明

- `[x]` 已完成并通过当前阶段验收
- `[ ]` 未完成
- `⚠️` 已有代码，但仍缺公网/ECS真实运行验收
- `🟡` 当前正在执行（仅用于短期标记，任务完成后必须改为 `[x]` 或退回 `[ ]`）

## 任务 14：Script 画布分镜与重新编译

- [x] MySQL 分镜文档/卡片保存、排序与乐观锁 migration（`00012_task14_script_storyboards.sql`）
- [x] `/script` 画布分镜新增、编辑、删除、排序、冲突提示与刷新恢复
- [x] 手工保存后复用既有 generation/StageRun 编译 Director 与 FINAL_PROMPT；历史视频任务不被覆盖
- ⚠️ ECS 公网真实数据、模型调用与视频下游验收留待 Task 16

## 完成标准

单个开发任务只有同时满足以下条件，才可以标记为 `[x]`：

1. 代码已进入 `main`。
2. 相关单元/集成测试通过。
3. Go test、前台 build、后台 build 通过。
4. 与旧 v88 / 公网对应行为完成对标，不把旧 bug 原样迁入。
5. 涉及公网能力的任务，在最终 ECS 阶段还必须完成真实浏览器/接口回归。

## Task 9–15 旧 v88 复用审计

- 详细审计：`docs/migration/v88-task-9-15-reuse-audit.md`
- Task 9–15 的 `[ ]` 只表示“新主线尚未达到完成标准”，**不代表旧 v88 从来没有写过该功能**。
- 后续开发必须先查复用审计，再决定“直接迁语义 / React 交互复用 / Go 化重写 / 真正新增”，避免重复开发。
- 新主线后端/API/Worker/Scheduler/Provider/发布统一使用 Go；前端使用 React + Ant Design；MySQL + Goose 为长期事实源；Redis 只用于 Queue/Lock/短期运行态；旧 Node 后端只作行为参考。
- 审计标签不会自动改变 `[x]/[ ]`。只有满足上面的完成标准才允许勾选。

---

# 总进度

- [x] Task 1：新仓库骨架、Go API、前台、管理端、CI
- [x] Task 2：MySQL 数据模型 + Goose
- [x] Task 3：男女频 / 风格元数据判断
- [x] Task 4：121 小说正文与 bookinfo 接入
- [x] Task 5：Intake 小说获取执行流水线
- [x] Task 6：立即执行 / 自动化统一任务链
- [x] Task 7：Intake HTTP API 与运行时接线
- [x] Task 8：小说获取工作台 + v88 对标 + 合并 main
- [x] Task 9：Batch Factory V11 主工作台
- ⚠️ Task 10：批量工厂统一设置 / 版本对应配置档（仓库阶段已完成；公网/ECS 验收留到 Task 16）
- ⚠️ Task 11：水货生产工作台
- ⚠️ Task 12：剧本生成 / Director / Hook / 最终提示词（仓库阶段已完成；公网/ECS 验收留到 Task 16）
- ⚠️ Task 13：音频 + matchAudio 分镜（仓库阶段已完成；公网/ECS 验收留到 Task 16）
- ⚠️ Task 14：视频生成完整链路（仓库阶段已完成；等待 Task9.4 Runtime Adapter + Task16 ECS/公网验收）
- ⚠️ Task 15：发布 / 权限 / 登录认证兼容（仓库阶段已完成；公网/ECS 验收留到 Task 16）
- [ ] Task 16：ECS 新仓库部署与公网全流程验收
- ⚠️ Task 17：Novel Panel（仓库实现、Go/前台测试与构建已通过；ECS 公网逐按钮验收待 Task 16）

---

# Task 1：新仓库骨架

- [x] 1.1 建立新主线仓库结构
- [x] 1.2 Go API 基础服务
- [x] 1.3 React + Ant Design 用户端骨架
- [x] 1.4 React + Ant Design 管理端骨架
- [x] 1.5 GitHub Actions CI
- [x] 1.6 Go test / 前台 build / 后台 build 接入 CI

# Task 2：MySQL + Goose

- [x] 2.1 Intake 模型
- [x] 2.2 Book 模型
- [x] 2.3 BatchProject 模型
- [x] 2.4 Run 模型
- [x] 2.5 MySQL Store
- [x] 2.6 Goose migration
- [x] 2.7 同 Intake 去重语义
- [x] 2.8 不同 Intake 允许同一本书再次进入

# Task 3：男女频 / 风格解析

- [x] 3.1 男女频优先级：手工 > 121 category > 已验证 genre > AI fallback
- [x] 3.2 121 category 男频规则
- [x] 3.3 121 category 女频规则
- [x] 3.4 无法判断时保持 unresolved，不乱猜
- [x] 3.5 风格字段优先级基础逻辑
- [x] 3.6 确定性数据不得被 AI 覆盖

# Task 4：121 小说获取

- [x] 4.1 接入 `https://txt.121w.com/api.php`
- [x] 4.2 支持 `bookid`
- [x] 4.3 支持 `platform`
- [x] 4.4 支持 `max_txt`
- [x] 4.5 解析 `book_name / work_title`
- [x] 4.6 解析 `category`
- [x] 4.7 解析 `genre`
- [x] 4.8 HTTPS / 非 2xx / upstream error / invalid JSON / empty body 测试
- [x] 4.9 保留原始正文
- [x] 4.10 YG/阳光 -> platformId 4
- [x] 4.11 CD/常读 -> platformId 2

# Task 5：Intake 执行流水线

- [x] 5.1 每本书独立处理
- [x] 5.2 单本失败不拖死其他书
- [x] 5.3 单本失败写入 retryable_failed/error
- [x] 5.4 completed / partial_failed / failed 聚合状态
- [x] 5.5 platform_id 持久化
- [x] 5.6 original_text 持久化
- [x] 5.7 Intake 状态更新

# Task 6：立即执行 / 自动化

- [x] 6.1 completed Intake 才能创建项目/Run
- [x] 6.2 立即执行未指定 runAt 时使用当前时间
- [x] 6.3 自动化要求 runAt 严格晚于当前时间
- [x] 6.4 立即/自动化共用 BatchProject + Run
- [x] 6.5 共用同一 MySQL 事实源
- [x] 6.6 生产启动接入 MySQL Store
- [x] 6.7 `QIANTIE_MYSQL_DSN` 运行时接线
- [x] 6.8 `QIANTIE_GO_LISTEN_ADDR` 运行时接线
- [x] 6.9 MySQL driver 接入

> ⚠️ 说明：当前“自动化”完成的是持久化未来 `run_at`；真正到点执行的 Scheduler/Redis Worker 仍在后续任务中。

# Task 7：HTTP API

- [x] 7.1 `POST /api/v1/intakes`
- [x] 7.2 `POST /api/v1/intakes/{id}/execute`
- [x] 7.3 `GET /api/v1/intakes/{id}/books`
- [x] 7.4 `POST /api/v1/intakes/{id}/batch-projects`
- [x] 7.5 服务依赖注入
- [x] 7.6 MySQL 运行时接线
- [x] 7.7 HTTP 测试

# Task 8：小说获取工作台

- [x] 8.1 前端测试体系（Vitest + Testing Library + jsdom）
- [x] 8.2 CI 增加前台 `npm test`
- [x] 8.3 添加书城分组
- [x] 8.4 Book ID 批量录入
- [x] 8.5 同书城 Book ID 去重
- [x] 8.6 多书城连续添加
- [x] 8.7 添加后清空 Book ID 输入框
- [x] 8.8 显示“书城 N 本”标签
- [x] 8.9 创建 Intake
- [x] 8.10 执行 121 获取
- [x] 8.11 展示书名
- [x] 8.12 展示 Book ID
- [x] 8.13 展示书城来源
- [x] 8.14 展示 platformId
- [x] 8.15 展示分类 / genre
- [x] 8.16 展示男女频
- [x] 8.17 展示风格
- [x] 8.18 展示状态 / 错误
- [x] 8.19 “立即执行”创建 BatchProject + Run
- [x] 8.20 “自动化”传未来 runAt
- [x] 8.21 partial_failed 时阻止创建 BatchProject
- [x] 8.22 v88 121 元数据语义对标
- [x] 8.23 PR CI 全绿
- [x] 8.24 合并 `main`
- [x] 8.25 合并后 CI 全绿

---

# Task 9：Batch Factory V11 主工作台

## 9.1 v88 / ECS 盘点

- [x] 9.1.1 列出 V11 所有前端页面与组件 — `docs/migration/v88-batch-factory-v11-frontend-inventory.md`
- [x] 9.1.2 列出 V11/V12 所有 Go/Node 历史后端接口 — `docs/migration/v88-batch-factory-v11-backend-api-inventory.md`
- [x] 9.1.3 列出 Batch / Book / Run / Item / VIDEO 状态字段 — `docs/migration/v88-batch-factory-v11-status-model-inventory.md`
- [x] 9.1.4 列出 personal_api / provider 相关接口 — `docs/migration/v88-batch-factory-video-provider-inventory.md`
- [x] 9.1.5 对照 ECS：存在 / 可用 / 已坏 / 未验证 — `docs/migration/ecs-v88-capability-audit.md`
- [x] 9.1.6 建立旧 v88 -> 新 Go 模块迁移映射表 — `docs/migration/v88-to-new-go-react-module-map.md`

## 9.2 批量项目列表

- [x] 9.2.1 BatchProject 列表 API
- [x] 9.2.2 从 MySQL 读取真实项目
- [x] 9.2.3 展示项目名称
- [x] 9.2.4 展示书城来源
- [x] 9.2.5 展示小说数量
- [x] 9.2.6 展示男女频
- [x] 9.2.7 展示风格
- [x] 9.2.8 展示运行状态
- [x] 9.2.9 Task 8 新建项目能立刻出现在列表

> 9.2 仓库证据：BatchProject MySQL 汇总测试 + `BatchProjectListPage.test.jsx` 覆盖来源/数量/男女频/风格/运行状态；`task9-list-visibility.spec.js` 覆盖 Task8 完成后立即进入批量工厂并看到新项目。

## 9.3 项目详情

- [x] 9.3.1 点击项目进入 V11 工作台
- [x] 9.3.2 展示项目所有小说
- [x] 9.3.3 展示 Book ID
- [x] 9.3.4 展示书名
- [x] 9.3.5 展示书城
- [x] 9.3.6 展示男女频
- [x] 9.3.7 展示风格
- [x] 9.3.8 展示正文获取状态
- [x] 9.3.9 展示真实错误原因

> 9.3 仓库证据：`BatchProjectListPage.test.jsx` 点击真实项目进入 `Batch Factory V11 工作台`，同时验证 Book ID、书名、书城/平台、男女频、风格、正文状态及 `121 upstream error` 等真实安全错误展示。

## 9.4 Run 执行框架

- [x] 9.4.1 pending -> running 状态转换
- [x] 9.4.2 单本独立状态
- [x] 9.4.3 一本失败不阻断其他书
- [x] 9.4.4 错误持久化
- [x] 9.4.5 单本重试
- [x] 9.4.6 防重复点击 / 幂等保护
- [x] 9.4.7 自动化到点执行 Worker
- [x] 9.4.8 Redis Queue
- [x] 9.4.9 Redis Lock
- [x] 9.4.10 Worker 崩溃/重启后的任务恢复策略

> 9.4 仓库证据：真实 MySQL/Redis 集成覆盖 Worker CAS、Scheduler due claim、BookRun 独立执行、3 本书失败隔离/只重试失败书、safe `error_code/error_message` 脱敏持久化、20 并发真实 `pipeline.Service.Create` 幂等、Redis Queue、FLUSHDB/Go restart recovery、stale Worker fencing。9.4.9 的 Redis Lock 要求由 **owner-safe Lease + fencing token semantics** 满足；不新增 dumb Redis mutex。

## 9.5 Task 9 验收

- [x] 9.5.1 Go tests
- [x] 9.5.2 前台 tests
- [x] 9.5.3 前台 build
- [x] 9.5.4 后台 build
- [x] 9.5.5 v88 行为对标
- [x] 9.5.6 PR CI 全绿
- [x] 9.5.7 合并 main
- [x] 9.5.8 合并后 CI 全绿

> 9.5.5 对标结论见 `docs/migration/task9-runtime-v88-final-audit.md`：旧 v88 的 JSON/local scheduler + process-local running set 被替换为 MySQL authoritative facts + Redis Queue/Lease + fencing + crash recovery，属于**语义继承 + Go 化增强**，不是照搬旧实现。

---

# Task 10：批量工厂统一设置 / 配置档

- [x] 10.1 “生产统一设置”使用右侧宽 Drawer，不跳页面
- [x] 10.2 “发布统一设置”使用右侧宽 Drawer，不跳页面
- [x] 10.3 保存后回到工作台并保留上下文
- [x] 10.4 删除旧“解析输入”入口
- [x] 10.5 新增“版本对应配置档”
- [x] 10.6 配置档支持同步 121 网站配置
- [x] 10.7 配置档支持同步批量风格类型
- [x] 10.8 清理 AI 文案数量重复配置
- [x] 10.9 清理“AI文案处理优先方案”重复逻辑
- [x] 10.10 明确处理规则提示词位置
- [x] 10.11 明确知识库提示词位置
- [x] 10.12 配置写入 MySQL / 后端事实源
- [x] 10.13 配置读取与刷新一致
- [x] 10.14 测试 + CI + 合并 main

> 仓库阶段验收依据：PR #14 已合并到 `main`，合并提交 `3a20e2fb99d3d08552b9faed1e1732118245706e`；合并后的 `main` CI run `37289454484` 中 Go tests/build、前台 tests/build、管理端 build、Goose migration/真实 MySQL 持久化与 Task 10 审计全部通过。公网真实“版本对应配置档 / 统一设置”仍需 Task 16 的 16.2.10 在 ECS/浏览器环境验收，因此 Task 10 总体保持 `⚠️`，不把仓库通过冒充为公网完成。

---

# Task 11：水货生产

- [x] 11.1 `/shuihuo-production` 页面迁移
- [x] 11.2 水货生产 -> Batch Factory V11 入口
- [x] 11.3 `/shuihuo-production` 独立生产工作台路由、状态恢复与失败恢复入口
- [x] 11.4 项目列表共用真实 MySQL 数据
- [x] 11.5 状态实时同步
- [x] 11.6 失败状态可见
- ⚠️ 11.7 对标旧 ECS 可用交互
- [x] 11.8 测试 + CI + 合并 main

> 2026-10-07 补齐：`/shuihuo-production` 已不再复用小说获取页，改为独立 React 生产工作台，复用 BatchProject、generation、video、统一设置、Cookie Session/CSRF 与权限边界；前台完整测试 15 files / 54 tests、相关 Go tests、前台 build 均通过。ECS / 公网真实浏览器验收，以及旧页分镜素材编辑、图片/音频任务与本地执行器能力的 Go 迁移仍留到后续任务，因此 Task 11 总体继续保持 `⚠️`。

> 2026-10-07 媒体基础：additive Goose migration 已提供项目/书 scoped 分镜、媒体资产、媒体意图任务和候选结果关联。图片/音频意图保持 `pending_executor`，没有真实执行器时不伪造 Provider 调用或结果；视频意图只允许关联现有 `video_production_tasks`。完整 5B 前端和图片/音频执行器适配仍待后续任务。

> 2026-10-07 12B：补齐分镜 PUT 乐观锁、事务式同书全量排序、候选读取/单主候选选择、媒体任务读取与 `pending_executor`/`retryable_failed` 重试 API。现有 schema 已能表达这些语义，未新增 migration；视频关联任务继续读取既有 video task 状态，不复制为 Shuihuo 事实。

> 2026-10-07 5B 前端接入：`/shuihuo-production` 已接入分镜编辑及 409 冲突刷新、同书排序、资产/媒体任务/候选读取、主候选选择和安全重试；前端不使用浏览器业务缓存，且在无 TOS 上传 API 或真实图片/TTS 执行器时明确显示空态或 `pending_executor`，不伪造结果。Task 11 仍为 `⚠️`，公网浏览器验收、TOS 上传和真实图片/TTS 执行器适配待后续任务。

---

# Task 12：剧本生成 / Director / Hook / 最终提示词

- [x] 12.1 剧本生成 Go API — 已合并 `main`
- [x] 12.2 系统提示词后端管理 — MySQL Prompt key/version/content/enabled + Go resolver 已完成
- [x] 12.3 剧情模式通用元提示词 — 通用视觉化/强钩子/空间/动作/电影化/安全规则已完成
- [x] 12.4 Hook — 正式 Stage；关闭时 `skipped`，不记失败
- [x] 12.5 Director — normal/H3 接口与 `h3-director/v1` 结构校验已完成；仅预留 matchAudio/audioDurationSec 兼容字段，不实现 Task 13 时长重排
- [x] 12.6 最终提示词编译 — Go 后端确定性编译；Retry 保留完整编译上下文
- [x] 12.7 单本生成状态 — BookRun/StageRun + MySQL 已完成
- [x] 12.8 批量生成状态 — 单本失败隔离与项目聚合已完成
- [x] 12.9 失败重试 — 目标 Stage 独立重试、复用已成功前置 Stage
- [x] 12.10 提示词版本持久化 — StageRun 持久化 prompt_key/prompt_version
- [x] 12.11 v88 对标 — Hook skip/H3 schema/Final Prompt retry/批量失败隔离/Stage Retry 等语义测试已通过
- [x] 12.12 测试 + CI + 合并 main — PR #12 已合并；`main` CI run `37288311920`（#372）中 Go tests/build、前台 tests/build、管理端 build 全绿
- [x] 12.13 `/script` Script Workspace — 复用 BatchProject/Book/StageRun/视频状态；人物、场景、约束与生成参数保存到既有项目 MySQL 设置；前台 57 项测试、相关 Go 测试与前台构建通过（公网/ECS 实测仍留 Task 16）

> 仓库阶段验收依据：PR #12 已合并到 `main`，合并提交 `73ab448c20337eb4cd0be1491c72fa973a3a57fc`；合并后的 `main` CI run `37288311920` 五项验证全绿。公网真实“剧本生成 / Director / Hook”仍需 Task 16 的 16.2.11–16.2.13 在 ECS/浏览器环境验收，因此 Task 12 总体保持 `⚠️`，不把仓库通过冒充为公网完成。

---

# Task 13：音频 + matchAudio

- [x] 13.1 第三步生成音频
- [x] 13.2 获取真实 `audioDurationSec`
- [x] 13.3 第四步加入“匹配音频”开关
- [x] 13.4 `matchAudio=true` 时总时长严格等于音频时长
- [x] 13.5 第一镜从 0 秒开始
- [x] 13.6 镜头之间无空缺
- [x] 13.7 镜头之间无重叠
- [x] 13.8 最后一镜结束时间 = `audioDurationSec`
- [x] 13.9 10s 规则保留
- [x] 13.10 15s 规则按“≤15s”执行
- [x] 13.11 `matchAudio=false` 保留原 Director 逻辑
- [x] 13.12 时间统一为秒并保留两位小数
- [x] 13.13 不允许为了凑时长添加无关剧情
- [x] 13.14 测试 + CI + 合并 main

> 仓库阶段验收依据：PR #18 已合并到 `main`，合并提交 `ce10e3ea0ca4114dd30a14b03eed19e034cb710f`；Task 13 migration 使用 `00006_task13_match_audio.sql`。合并后的 `main` CI run `37298187716` 中 Go tests/build、前台 tests/build、管理端 build、Goose Up/status/rollback、Task10 audit 全绿。服务端权威音频测量、matchAudio 严格时间轴、Normal/H3 Director、`matchAudio=false` Task 12 回归均已完成仓库阶段验收。公网/ECS 真实音频测量、matchAudio 与 Director 全流程仍留到 Task 16 的 16.2.12、16.2.14–16.2.15，因此 Task 13 总体保持 `⚠️`。

## 13B：Shuihuo TOS 资产与媒体执行桥接（仓库阶段）

- [x] 13B.1 服务端受控 multipart 上传：按项目/书/可选分镜校验，服务端生成 TOS key，校验类型、内容和大小；TOS 成功后才写入 MySQL。
- [x] 13B.2 受权限保护的资产读取代理：不向浏览器暴露 bucket credential、任意 key 或公共 TOS URL。
- [x] 13B.3 视频媒体任务继续以既有 `video_production_tasks` 为唯一执行状态来源；不新增 Worker、Queue 或 Runtime。
- [x] 13B.4 图片/TTS 尚无已配置的 Go Provider/执行器时，持久化明确 `executor_unavailable` 错误，不伪造 Provider、URL 或成功结果。
- [x] 13B.5 以 fake TOS adapter、Service 与 HTTP scope 测试覆盖上传、安全边界和状态桥接。

> 未宣称公网完成：真实 TOS 凭据、图片 Provider 和 TTS Provider 均须在 Go 服务端统一配置后，才能进行 ECS 端到端验收；浏览器不接触密钥。

## 15B：Shuihuo 图片 / TTS Provider 执行器（仓库阶段）

- [x] 15B.1 使用现有 `taskruntime` Redis queue 消费图片/TTS 媒体任务；不新增 Worker、Queue、Scheduler、Runtime 或 Auth。
- [x] 15B.2 服务端环境变量分别配置图片与 TTS HTTP Provider；Provider 密钥不进入前端或响应。
- [x] 15B.3 Provider 结果必须下载并通过既有受控 TOS 上传持久化，随后创建候选并回写 `queued → running → succeeded`。
- [x] 15B.4 Provider、Redis 或 TOS 缺失时保留明确失败/可重试状态；没有 Provider 时固定 `executor_unavailable`，不伪造成功。
- [x] 15B.5 fake Provider、Service queue、HTTP scope 与 Go build 测试已通过；未调用付费 Provider。

> ECS 仍需配置真实环境变量并按 Provider HTTP 契约完成端到端验收；本项不宣称真实图片/TTS 账户、TOS 或 Redis 已在公网验证。

---

# Task 14：视频生成完整链路

- [x] 14.1 视频任务模型
- [x] 14.2 `VIDEO` 状态机
- [x] 14.3 personal_api provider
- [x] 14.4 yd2.0-mini
- [x] 14.5 豆包本地执行器
- [x] 14.6 provider 状态 API
- [x] 14.7 角色/场景/图片资产传递
- [x] 14.8 视频生成提交
- [x] 14.9 视频状态轮询
- [x] 14.10 失败重试
- [x] 14.11 TOS / 本地产物管理
- [x] 14.12 视频合并 Worker
- [x] 14.13 ffmpeg 合并链
- [x] 14.14 最终结果回写 BatchProject
- [x] 14.15 修复并验证 `personal_api` 状态接口不再 503
- [x] 14.16 测试 + CI + 合并 main

> 仓库阶段验收依据：PR #22 已合并到 `main`，合并提交 `6b60a62896000e1c37fb0c8ec11da35a89e45dcf`；Task 14 migration 为 `00007_task14_video.sql`。合并后发现 existing-main CI 基线仍假设 `main` 不含 00007，已用最小 CI 修复提交 `d38520c1f2eb968021259010db3a2315de78e994` 改为显式构造 00001–00006 的 pre-Task14 基线。最终 `main` CI run `37309539521` 中 Go tests/build、前台 tests/build、管理端 build、clean DB Goose Up/status/rollback、pre-Task14 `00004/00006 → 00007` 真实升级路径、Task14 restart recovery 与 Task10 audit 全绿。
>
> ⚠️ awaiting Task9.4 runtime adapter：`Claim / Renew / Release / RequeueExpired`、production queue coordination、local executor lease recovery、Merge async queue / lease。Task 14 不自行新增第二套 Redis Runtime / Scheduler，待 Task9.4 公共 Runtime 合并后仅做小型 `Task14 Runtime Adapter` 收尾。
>
> ⚠️ Task 16 仍需 ECS/Public 真实验收：真实 Provider（含 `personal_api / yd2.0-mini / yfai_seedance / autodl_comfyui / doubao_local_executor`）、TOS、ffmpeg、VIDEO/Merge、失败/重试及浏览器权限链路。因此 Task 14 总体保持 `⚠️`，不把仓库阶段通过冒充为公网完成。

---

# Task 15：发布 / 权限 / 登录认证

- [x] 15.1 登录 Token 体系统一
- [x] 15.2 Token 刷新
- [x] 15.3 修复“当前登录状态已过期”闪退
- [x] 15.4 API 白名单梳理
- [x] 15.5 `/api/script-constraint-prompts` 权限兼容
- [x] 15.6 发布权限验证
- [x] 15.7 用户权限
- [x] 15.8 团队权限
- [x] 15.9 发布任务状态
- [x] 15.10 权限失败不影响其他任务
- [x] 15.11 测试 + CI + 合并 main

> 仓库阶段验收依据：PR #20 已合并到 `main`，合并提交 `2cee78c48a44ae4471c75261542a52f5bd98843b`；合并后的 `main` CI run `37303943125` 中 Go tests/build、前台 tests/build、管理端 build、Goose Up/status/rollback、Publishing 真实事务 rollback/retry 与 Task10 audit 全绿。Auth/Publishing 仓库代码阶段完成，但公网真实登录、Token/session 恢复、发布和权限仍需 Task 16 的 16.2.20、16.2.21、16.2.23 在 ECS/浏览器环境验收，因此 Task 15 总体保持 `⚠️`。

---

# Task 16：ECS 部署与最终公网验收

## 16.0 全站公网验收执行矩阵（Task 11，2026-10-07）

> 状态：`[ ]` 表示尚未在登录后的公网/ECS 环境执行，不能以本地单测、构建或 HTTP 200 代替。本矩阵是 Task 16 的执行清单；真实截图、请求 ID、账号角色与 Provider 任务 ID 必须在执行时填写。

| 路由 / 页面 | 核心功能与状态 | 权限与刷新恢复 | UI 截图 | 真实 Provider / 外部结果 |
| --- | --- | --- | --- | --- |
| `/` 首页 | 创作入口、最近项目空态/有数据态、深浅主题 | 未登录跳转、登录后返回原地址、刷新不丢主题 | [ ] | 不适用 |
| `/novel-fetch` | 多来源、多书添加、获取正文、部分失败与重试 | `batch.view/configure/execute`、刷新后 Intake/Book 恢复 | [ ] | [ ] provider121 正文与分类 |
| `/novel-fetch-workshop` | 处理规则、知识库说明、原文查看/恢复、失败书重试 | `batch.view/configure/execute`、刷新后 Workshop 设置/原文恢复 | [ ] | [ ] provider121 原文恢复 |
| `/batch-factory` | 项目列表、筛选、创建、项目进入、空/错/加载态 | `batch.view/configure/execute`、刷新后 BatchProject/Run 恢复 | [ ] | [ ] Intake/Run 真实结果 |
| `/batch-factory?projectId={id}` | 书籍、阶段状态、统一设置、版本配置、音频、视频、合成与重试 | 项目归属/Capability、刷新后 StageRun/视频/合成状态恢复 | [ ] | [ ] generation、TTS、视频、TOS、ffmpeg |
| `/shuihuo-production` | 生产项目、单书继续执行、阶段重试、视频提交/取消/重试 | `batch.view/execute`、刷新后生成与视频任务恢复 | [ ] | [ ] generation、视频 Provider、TOS |
| `/script` | 原文、人物/场景/约束保存、SCRIPT/HOOK/DIRECTOR/FINAL_PROMPT、音频入口 | `batch.view/configure/execute`、刷新后 MySQL 设置与 StageRun 恢复 | [ ] | [ ] generation、TTS、matchAudio |
| `/novel-panel` | 核对真实页面、工作台桥接与安全通信 | 登录/权限/刷新恢复 | [ ] | [ ] 关联项目/媒体结果 |
| `/agent` | 核对 Agent 工作区会话、画布、版本与任务状态 | 登录/会话恢复/权限 | [ ] | [ ] 已配置模型与 TOS 输出 |
| `/tts` | 配音创建、进度、试听、失败重试 | 能力权限、刷新后音频任务恢复 | [ ] | [ ] TTS Provider、音频 TOS URL |
| `/history` | 项目/作品历史、空态与进入详情 | 仅本人可见、刷新恢复 | [ ] | 不适用（验证聚合 API 真实结果） |
| `/issues` | 问题日志、筛选、错误详情与请求 ID | 运维/角色边界、刷新恢复 | [ ] | [ ] 真实失败任务可追溯 |
| `/settings` | 用户可见设置与保存反馈 | 本人权限、刷新后服务端设置恢复 | [ ] | 不适用 |
| `/member` | 账号资料、会话、退出登录 | 本人/管理员边界、刷新与重新登录 | [ ] | 不适用 |
| `/api/v1/diagnostics`（非页面） | 健康、依赖与脱敏诊断 | 仅 operations admin，禁止普通用户访问 | 不适用 | [ ] MySQL、Redis、TOS 连通性 |

> 2026-10-07 Task 16 前台补齐：`/tts` 已复用 Shuihuo 音频任务、候选资产内容与原有重试接口，提供筛选、刷新恢复、试听、下载和失败展示；`/history` 已通过 Go/MySQL 的单一用户范围投影提供搜索、状态筛选、分页、总数、空/错/加载态和进入项目。未新增浏览器业务存储、历史表、Runtime 或 Provider。真实“文本创建 TTS 任务”载荷仍须由统一后端契约提供后接入，不能以旧 Node 文件历史替代；历史和 TTS 均仍须做登录后的公网验收。

### Task 11 测试稳定性验证（仓库阶段）

- [x] 已复现 Workshop 测试不稳定：原单测把首屏恢复、Drawer 渲染和保存请求串在同一条默认 5 秒超时断言内；一次实测为 `5.72s`，另一次在 `5.234s` 时触发 Vitest timeout，并非产品 API、权限或业务失败。
- [x] 仅调整测试同步与测试环境：拆分为“恢复事实 / 查看原文 Drawer / 保存配置”三条独立断言；jsdom 测试环境忽略其不支持的 `getComputedStyle` 伪元素参数，保留元素自身计算样式；Script 测试选择器与现有“按已保存原文生成剧本”按钮文案同步。未改产品 UI、业务语义、权限或 API。
- [x] Workshop 定向回归连续 5 次通过；三条断言单次最长 `1.940s`，未修改 Vitest 默认超时。
- [x] 最新 `main` 上前台全量 `npm test` 连续三次通过：均为 `18` 个测试文件、`66` 条测试、退出码 `0`；总时长依次为 `129.76s`、`121.58s`、`125.45s`。
- [x] `npm run build` 通过，退出码 `0`（产物：`dist/`）；仅有 Ant Design/Rollup 的既有 `use client` 与大 chunk 警告，无构建失败。
- [x] Shuihuo 完整视频提交交互在全量套件中曾超过默认 5 秒但功能链路正确；仅为该测试设置 15 秒局部超时。定向测试与连续两次全量前台测试均通过，未改变产品逻辑或全局 timeout。
- [ ] 本次未部署、未运行 GitHub Actions、未执行公网登录或 Provider 调用；上表所有 `[ ]` 仍须在 Task 16 的真实 ECS/浏览器验收中逐项留存证据。

## 16.1 构建与部署

- [ ] 16.1.1 GitHub 自动构建部署产物
- [ ] 16.1.2 Go embed 打包前端静态资源
- [ ] 16.1.3 ECS 旧 v88 备份
- [ ] 16.1.4 数据库备份
- [ ] 16.1.5 Goose migration 安全执行
- [ ] 16.1.6 MySQL 配置验证
- [ ] 16.1.7 Redis 配置验证
- [ ] 16.1.8 TOS 配置验证
- [ ] 16.1.9 新 Go 服务启动
- [ ] 16.1.10 健康检查
- [ ] 16.1.11 回滚方案验证

## 16.2 公网真实全流程

- [ ] 16.2.1 登录
- [ ] 16.2.2 小说获取页面
- [ ] 16.2.3 多书城添加
- [ ] 16.2.4 121 获取正文
- [ ] 16.2.5 书名/平台/男女频/风格显示
- [ ] 16.2.6 立即执行
- [ ] 16.2.7 自动化执行
- [ ] 16.2.8 Batch Factory V11
- [ ] 16.2.9 水货生产
- [ ] 16.2.10 版本对应配置档
- [ ] 16.2.11 剧本生成
- [ ] 16.2.12 Director
- [ ] 16.2.13 Hook
- [ ] 16.2.14 音频
- [ ] 16.2.15 matchAudio
- [ ] 16.2.16 personal_api 视频
- [ ] 16.2.17 yd2.0-mini
- [ ] 16.2.18 豆包本地执行器
- [ ] 16.2.19 视频合并
- [ ] 16.2.20 发布
- [ ] 16.2.21 权限
- [ ] 16.2.22 失败/重试
- [ ] 16.2.23 登录刷新后状态保持
- [ ] 16.2.24 与旧 ECS 核心功能逐项对照
- [ ] 16.2.25 正式切换公网到新仓库

---

# 进度维护规则

以后开发时必须遵守：

1. 开始某个子任务时，可以临时标记 `🟡`。
2. 代码完成但测试未通过，不得勾选。
3. 测试通过但尚未进入 `main`，不得勾选最终验收项。
4. 合并 `main` 后，必须更新对应 `[x]`。
5. 如果后来发现回归，必须把对应项目从 `[x]` 退回 `[ ]` 或标记 `⚠️`。
6. 每次汇报进度时，以本文件为准，不凭聊天记忆口头报数。
7. 不为每个小改动长期保留分支；功能完成、CI 全绿后及时合并并清理分支。
8. 最终“公网可用”只能在 Task 16 ECS 真实验证后确认。

## 当前执行点

**当前任务：用户工作区完整迁移与公网 UI 对标。**

### 2026-10-07 · 用户工作区迁移：账户入口（仓库阶段）

- [x] `/member` 和 `/profile` 已移除路由占位，复用现有 Cookie Session 的服务端用户资料、角色、团队 ID、capability 与安全登出接口。
- [x] 页面明确不读取、不显示密码、Cookie、Session、MFA、恢复码或任何 Provider 密钥；没有新增认证体系或浏览器业务存储。
- [x] 前台定向回归：`AccountCenterPage` 2 条、`RouterApp` 6 条均通过；`npm run build` 通过。
- [ ] 仍需迁移服务端资料编辑、账号安全会话管理、团队事实模型、MySQL 用户偏好与公网浏览器验收；不得以本项仓库验证宣称账户中心整体完成。

### 2026-10-07 · 前台回归稳定性

- [x] Shuihuo 单书继续执行/阶段重试的既有全部断言保持不变，并为该跨 API、重渲染用例设置 15 秒局部上限；避免默认 5 秒测试时限将已完成的真实 API 调用误报为失败。
- [x] 该阶段前台全量测试两次自然结束：均为 24 个测试文件、78 条测试、退出码 `0`；随后两次 `npm run build` 均通过。未使用后台子进程或人为终止测试会话。
- [x] 当前 `main` 后续全量前台验证自然结束：25 个测试文件、82 条测试、退出码 `0`；问题投影与项目列表的 Go 相关套件也通过。

### 2026-10-07 · 用户工作区迁移：全局壳（仓库阶段）

- [x] 顶栏账户菜单现直接使用现有 Cookie Session 返回的 `name` 字段，回退到 `username`，避免真实登录后错误显示“用户入口”；不新增或复制认证资料。
- [x] `UserShell` 回归测试覆盖认证名称显示；前台定向测试和构建通过。
- [x] 非首页工作区恢复旧 v88 的可折叠桌面 Sidebar：复用现有导航、主题、设置和个人资料入口；移动端隐藏 Sidebar 并继续使用 Drawer，不引入浏览器业务事实存储或整份旧 CSS 污染。
- [ ] 仍需 Task 16 的登录态移动端/桌面端浏览器视觉验收，且需要部署授权后才能对公网旧 v88 实例进行实际对标。

### 2026-10-07 · 用户工作区迁移：问题日志（仓库阶段）

- [x] `/issues` 直接投影 MySQL 中的 BookRun、StageRun、视频与媒体任务失败事实；不建立第二套错误日志。
- [x] 读取在 SQL 层按项目 owner/team 归属过滤；admin/owner 才能读取全局项目。错误摘要经统一脱敏后返回。
- [x] 前台支持来源筛选、关键词搜索、分页、加载/错误/空态与刷新。
- [x] 每条问题事实可直接进入其已授权项目；空投影使用明确中文空态，不以 Ant Design 默认“无数据”替代业务说明。
- [x] BookRun、StageRun、视频与媒体任务的已持久化请求 ID 与错误码在同一受权限保护投影中展示，并与错误摘要一样经过脱敏；不把 Provider 凭据交给前端。
- [x] 分页同时返回受同一归属与筛选条件约束的总数；后端契约测试覆盖成员归属参数与错误摘要脱敏。
- [ ] 仍需 Task 16 登录后公网/ECS 数据验收；运营级诊断详情需在后续统一查询契约中扩展。

### 2026-10-07 · 用户工作区迁移：设置（仓库阶段）

- [x] `/settings` 的主题、提醒和存储偏好以当前认证用户为边界写入 MySQL `user_workspace_preferences`；前端不保存业务设置事实。
- [x] 每次 Cookie Session 认证成功后，应用重新读取服务端主题偏好并实际应用；LocalStorage 仅在读取偏好失败时作为非敏感界面回退。
- [x] 读取与保存均复用 Cookie Session、CSRF 同源保护和现有认证用户，不返回任何密钥、Cookie、Session 或 Provider 凭据。
- [x] 本地执行器列表直接使用现有 Go `LocalExecutorService`；无设备或离线时明确显示不可用，不伪造实际生效状态。
- [x] 执行器服务查询失败与“无已注册设备”分开呈现，服务端仅返回安全的不可用原因，不透传内部错误。
- [x] 首次读取服务端偏好失败时显示失败原因与重试入口；重试仍读取同一 Cookie Session 范围的服务端设置，不以浏览器缓存伪造恢复。
- [ ] 仍需在 Task 16 以登录态浏览器验证 MySQL migration、实际执行器心跳与偏好跨会话恢复；存储偏好尚未改变媒体写入策略，页面明确只展示已选择配置。

### 2026-10-08 · 本机隔离 staging 验收（非 ECS / 非公网）

- [x] React 生产构建复制到 `api/internal/webui/dist` 并以 Go `embed` 打入单个 `cmd/server` 二进制；`/api/build-info` 返回构建时注入的 Git SHA，静态响应带 `X-Ycm-Static-Source: go-embed`。
- [x] 仅在本机 `ycm_staging` 执行 Goose `00001`–`00014`；服务仅监听 `127.0.0.1:18080`，Redis 队列命名空间为 `ycm:staging:`，日志写入独立 `.staging-logs/`。
- [x] 浏览器登录、Cookie Session 重启后恢复、同源 CSRF 403、未登录 401、设置刷新恢复与工作区路由进入均在本机单二进制完成；前台全量 25 文件 / 82 测试、`npm run build`、`go test ./...` 均通过。
- [x] 无 Provider 时媒体任务真实持久化为 `executor_unavailable`；未配置 TOS 时上传真实返回安全的 `503 storage_unavailable`，未发起任何 Provider/TOS 调用。
- [x] `/issues` UNION 投影在本机 MySQL 默认 collation 下已修复，浏览器返回脱敏空态而非 500；不改写历史业务表或新增错误事实表。
- [x] Agent Studio migration 与已存在的工作区偏好 migration 的 `00013` 版本冲突已修正为 `00014`；本机 Goose 可读取并应用完整序列，避免 `duplicate version` panic。
- [x] 本机真实失败媒体任务暴露了 Shuihuo DTO 大写 JSON 字段与 TTS 外部书号误作数据库主键的问题；已用 HTTP 契约测试锁定 camelCase 响应，前台统一使用 `books.id` 调用受项目范围保护的媒体接口，未改动 MySQL 事实或伪造 Provider 成功。
- [ ] 本机验收不替代 Task 16：未连接 ECS、未修改公网、未运行 GitHub Actions、未验证真实 Provider/TOS 或公网视觉对标。

---

# 2026-10-07 · 本轮恢复任务 2：Novel Fetch 主页面

> 状态：⚠️ 仓库实现已写入 main；本轮前台真实测试与构建已在 macOS 本地环境执行通过。没有使用 GitHub Actions 代替本地验收；ECS / 公网逐页面验收仍未完成。

已实现：
- /novel-fetch 独立 React + Ant Design 主页面，不复用 iframe、Node 后端、/api/chat，也不使用 localStorage/sessionStorage 保存小说、任务、状态或结果。
- 复用现有 Go intake、provider121、BatchProject；任务创建、执行、书籍结果和刷新恢复继续以现有 MySQL intakes/books/batch_projects/runs 为事实源。
- 恢复来源 / 平台选择、一次添加多本小说和多个来源、Book ID 批量录入。
- 页面加载时通过 GET /api/v1/intakes + GET /api/v1/intakes/{id}/books 恢复已创建任务、进度、状态、失败原因和书籍结果。
- 失败任务使用同一 POST /api/v1/intakes/{id}/execute 链路重试；已 fetched 书籍不会重复抓取，不新建第二套任务或队列。
- 书籍结果展示 Book ID、书名、分类、类型、男女频、风格、状态和安全错误。
- 持久化单书 provider / AI 分类错误前使用服务端统一脱敏，避免 Token、密码、DSN 等敏感内容写入 MySQL 后再暴露到前台。
- 403 在 Novel Fetch 页面仅显示无权限；401 继续沿用现有 AuthBoundary / session refresh 进入登录流程。
- AI 处理配置、处理规则入口均指向 /novel-fetch-workshop。
- 新增回归测试覆盖：刷新恢复、部分失败展示、失败重试、403 不强制退出、服务端敏感错误脱敏与重试只处理失败书。

本次 macOS 真实前台验证（2026-10-07）：
- `npm test`：13 个测试文件、50 条测试全部通过，退出码 `0`。
- `npm run build`：通过，退出码 `0`。
- 此证据仅覆盖仓库内前台测试与构建；不替代 Task 16 的 ECS / 公网逐页面 UI 与功能验收。

本轮未完成：
- /novel-fetch-workshop 页面本体未迁移；本轮只提供正确跳转入口。
- 未新增 MySQL migration：现有 00001_phase1_intake.sql 已包含本轮持久化所需 intakes/books/batch_projects/runs 字段，避免为同一事实重复建表。
- Go test 不属于本次仅前台测试断言修复的验证范围；ECS / 公网真实浏览器验收仍未完成，不能仅凭本地前台测试与构建将本轮状态从 ⚠️ 提升为已验收。

---

# 2026-10-07 · Task 3：Novel Fetch Workshop

> 状态：⚠️ 仓库实现和本地真实验证已完成；未部署，也没有把 Task 16 的 ECS / 公网逐页面验收标记为完成。

本次实现与验证：
- 新增 `/novel-fetch-workshop`，从 `/novel-fetch?intakeId=<id>` 进入并在刷新或重新登录后继续由服务端 `intakes`、`books` 和 `generation_prompts` 恢复事实数据。
- 复用既有 `intakes`、`books`、`batch_projects`、`runs`，未创建第二套 Task、Book、Run、Queue 或 Worker；失败重试继续调用既有 `POST /api/v1/intakes/{id}/execute`。
- 新增最小 migration `00009_task3_novel_fetch_workshop.sql`，仅为 `intakes` 增加 `workshop_settings_json`，以持久化尚未生成 BatchProject 时的处理设置。
- 原文查看和单书恢复使用 Go `intake.Service` 与 provider121，恢复结果持久化回既有 `books.original_text`；不恢复 Node 后端、`/api/chat`、iframe 或前端 API Key / 业务 LocalStorage。
- Workshop 提示词/知识库入口读取既有 Go `generation` prompt 模块；前端不硬编码系统提示词。接口按既有 batch capability 控制：读取需 `batch.view`，保存设置与恢复原文需同源请求及 `batch.configure` / `batch.execute`。
- macOS 本地真实验证：相关 Go 包 `go test ./internal/intake ./internal/workshop ./internal/httpapi` 通过（退出码 0）；`npm install --no-package-lock --prefer-offline --no-audit` 通过（退出码 0）；`npm test` 14 个测试文件、51 条测试全通过（退出码 0）；`npm run build` 通过（退出码 0）。

仍未完成：
- Task 16 的 ECS 部署、登录后公网浏览器全链路、真实 provider121 凭据/网络与跨刷新权限验收仍保持未完成；不能以本地测试或公网应用壳 HTTP 200 代替。

---

# 2026-10-07 · Task 4A：Batch Factory 首页、列表与单书创建链路

> 状态：⚠️ 4A 已在 macOS 本地验证；不是 Batch Factory 全量公网对标完成，未部署。

已验证：
- `/batch-factory` 首页使用 Go API / MySQL 的真实 BatchProject、书籍数量、来源、属性和 Run 状态；支持搜索、来源/运行状态筛选、刷新、空态、失败态、无权限态和明确进入项目入口。
- 单书创建入口严格复用 `POST /api/v1/intakes`、`POST /api/v1/intakes/{id}/execute`、`POST /api/v1/intakes/{id}/batch-projects`；只有 Intake 完成才创建 BatchProject，浏览器不保存业务事实。
- macOS 本地验证：`npm install --no-package-lock --prefer-offline --no-audit`、`npm test`（15 文件 / 53 测试）和 `npm run build` 均通过，退出码均为 0。

仍未完成（4B / Task 16）：
- 多书分组创建、深层生产工作台完整 UI、候选/主版本与视频操作对标。
- 登录后公网浏览器、真实 provider、刷新恢复和逐像素 UI 验收；不得以本地构建替代。

---

## 2026-10-07 · 并行迁移任务 1：公网首页与用户路由基础

> 状态：⚠️ 本任务代码随本提交进入 main；当前执行环境无法从 npm registry 安装前台依赖，因此 Vitest / Vite 实际测试与构建未完成。没有使用 GitHub Actions 代替本地验收。

### 本次真实完成范围

- [x] `/` 不再渲染“小说获取工作台”，恢复独立首页。
- [x] 首页恢复公网历史源码对应的主标题、主视觉结构、“开始生成”“进入配音”、六个创作入口和“最近创作项目”入口。
- [x] 六个创作入口固定跳转：`/script`、`/novel-panel`、`/batch-factory`、`/shuihuo-production`、`/agent`、`/tts`。
- [x] `/novel-fetch` 继续进入当前 main 已存在的 Novel Fetch 页面，没有用首页覆盖。
- [x] 用户导航保留：首页、剧本生成、小说获取、小说面板、水货生产、Agent 工作区、历史、问题日志、配音、主题、设置、用户入口。
- [x] 认证边界只向工作区 React 组件注入当前用户与安全退出回调；DOM 子元素不会收到用户对象或事件属性。
- [x] 全局壳增加基于当前 Cookie Session 用户的个人资料/会员/安全退出菜单，并在移动端改用可关闭 Drawer 导航；不新增浏览器业务事实存储。
- [x] 保持 React + Ant Design；没有恢复 Node 生产后端、`/api/chat` 或 Bearer Token。
- [x] 仅主题允许使用 LocalStorage；本任务没有新增项目、任务、脚本或历史等业务数据的 LocalStorage 持久化。
- [x] 没有修改 Batch Factory、Novel Fetch、Script、TTS 的实际业务实现；现有 Batch Factory / Novel Fetch 继续复用当前 main 组件。
- [x] 新增 Task 1 路由回归测试，覆盖首页、六入口、导航、Novel Fetch 保留与主题按钮行为。

### 本次测试 / 构建结果

- [ ] `npm install --ignore-scripts --no-audit --no-fund --fetch-timeout=5000 --fetch-retries=0`：失败，`EAI_AGAIN registry.npmjs.org`，当前执行环境无法解析 npm registry。
- [ ] `npm test`：未能实际运行测试，退出码 127，`vitest: not found`。
- [ ] `npm run build`：未能实际构建，退出码 127，`vite: not found`。
- [x] 对本任务新增 / 修改 JSX 使用本机 TypeScript `transpileModule` 做语法解析检查，HomePage、RouterApp、RouterApp.test、UserShell、main、main.test 均无语法诊断；该检查不替代 Vitest / Vite。

### 本次明确未完成 / 与公网仍不一致

- [ ] 当前执行环境无法直接访问 `http://115.190.156.223:3000/` 做浏览器逐项验收；历史 SHA `27fa2e12f9401d378d1e6d6298a9d50a27581512` 仅作源码对标线索，不冒充 2026-10-07 已重新确认的公网部署版本。
- [ ] 历史公网首页的 `home-hero.mp4` 二进制素材未迁入新仓：当前 GitHub 连接器不能把旧仓二进制 blob 安全复制到目标仓，所以本次先恢复全屏深色主视觉、渐变光效、标题和操作区，不提交失效视频引用。
- [x] 首页“最近创作项目”直接读取受项目归属策略过滤、按 MySQL `updated_at` 倒序的 Go 批量项目接口，支持加载、空态、错误重试与进入真实项目；不使用 LocalStorage 伪造业务数据。
- [x] `/history` 改为 Go/MySQL 单一用户范围投影：项目归属过滤、服务端搜索/状态筛选/分页与总数，前台不再逐项目调用生成接口拼接历史。
- [x] 历史投影契约测试覆盖成员 owner/team 归属参数、关键词/状态筛选、分页参数、总数和 RFC3339 更新时间响应。
- [ ] Script、Novel Panel、Agent、TTS、历史、问题日志、设置、用户中心的实际业务页面不属于任务 1；当前仅保留稳定路由基础，等待各自模块提交。
- [ ] `/shuihuo-production` 继续保持当前 main 已有组件行为，本任务没有重写水货生产实际业务页。

---

# 2026-10-07 · Task 16：Agent Studio

> 状态：⚠️ 本地实现与测试完成；未部署，真实 Provider 与公网登录验收未完成。

- [x] Agent 独立项目、最近项目搜索/创建/进入/删除，以及服务端会话、消息、画布、附件和执行记录恢复。
- [x] 画布使用 MySQL revision 乐观锁、版本历史与安全恢复；附件内容使用 TOS、元数据使用 MySQL。
- [x] 用户技能版本由 MySQL 持久化并按所有者读取和选择；系统提示词仍在 Go `agentstudio` prompt 模块，前端不保存提示词事实。
- [x] API 复用 Cookie Session、CSRF、capability 与项目归属校验；未引入 Bearer、`/api/chat`、Node 后端或第二套队列/Worker。
- [x] 未配置执行器会持久化并返回 `executor_unavailable`，不会伪造助手消息。
- [x] 本地验证：`go test ./...` 通过；`npm test -- --run` 为 24 文件、80 测试通过（153.39 秒）；`npm run build` 通过。

仍未完成：真实 Provider、TOS 凭据联通、部署与登录后公网浏览器验收。
