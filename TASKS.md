# 一战晟铭新主线 · 公网/v88 迁移任务清单

> 这是本仓库唯一的迁移进度事实清单。
>
> 规则：每完成一个子任务，必须在同一次提交或同一个 PR 中同步更新本文件；没有勾选并附验收依据，不算正式完成。

## 状态说明

- `[x]` 已完成并通过当前阶段验收
- `[ ]` 未完成
- `⚠️` 已有代码，但仍缺公网/ECS真实运行验收
- `🟡` 当前正在执行（仅用于短期标记，任务完成后必须改为 `[x]` 或退回 `[ ]`）

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
- [ ] Task 9：Batch Factory V11 主工作台
- [ ] Task 10：批量工厂统一设置 / 版本对应配置档
- ⚠️ Task 11：水货生产工作台
- [ ] Task 12：剧本生成 / Director / Hook / 最终提示词
- [ ] Task 13：音频 + matchAudio 分镜
- [ ] Task 14：视频生成完整链路
- [ ] Task 15：发布 / 权限 / 登录认证兼容
- [ ] Task 16：ECS 新仓库部署与公网全流程验收

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
- [ ] 9.2.4 展示书城来源
- [ ] 9.2.5 展示小说数量
- [ ] 9.2.6 展示男女频
- [ ] 9.2.7 展示风格
- [ ] 9.2.8 展示运行状态
- [ ] 9.2.9 Task 8 新建项目能立刻出现在列表

## 9.3 项目详情

- [ ] 9.3.1 点击项目进入 V11 工作台
- [ ] 9.3.2 展示项目所有小说
- [ ] 9.3.3 展示 Book ID
- [ ] 9.3.4 展示书名
- [ ] 9.3.5 展示书城
- [ ] 9.3.6 展示男女频
- [ ] 9.3.7 展示风格
- [ ] 9.3.8 展示正文获取状态
- [ ] 9.3.9 展示真实错误原因

## 9.4 Run 执行框架

- [ ] 9.4.1 pending -> running 状态转换
- [ ] 9.4.2 单本独立状态
- [ ] 9.4.3 一本失败不阻断其他书
- [ ] 9.4.4 错误持久化
- [ ] 9.4.5 单本重试
- [ ] 9.4.6 防重复点击 / 幂等保护
- [ ] 9.4.7 自动化到点执行 Worker
- [ ] 9.4.8 Redis Queue
- [ ] 9.4.9 Redis Lock
- [ ] 9.4.10 Worker 崩溃/重启后的任务恢复策略

## 9.5 Task 9 验收

- [ ] 9.5.1 Go tests
- [ ] 9.5.2 前台 tests
- [ ] 9.5.3 前台 build
- [ ] 9.5.4 后台 build
- [ ] 9.5.5 v88 行为对标
- [ ] 9.5.6 PR CI 全绿
- [ ] 9.5.7 合并 main
- [ ] 9.5.8 合并后 CI 全绿

---

# Task 10：批量工厂统一设置 / 配置档

- [ ] 10.1 “生产统一设置”使用右侧宽 Drawer，不跳页面
- [ ] 10.2 “发布统一设置”使用右侧宽 Drawer，不跳页面
- [ ] 10.3 保存后回到工作台并保留上下文
- [ ] 10.4 删除旧“解析输入”入口
- [ ] 10.5 新增“版本对应配置档”
- [ ] 10.6 配置档支持同步 121 网站配置
- [ ] 10.7 配置档支持同步批量风格类型
- [ ] 10.8 清理 AI 文案数量重复配置
- [ ] 10.9 清理“AI文案处理优先方案”重复逻辑
- [ ] 10.10 明确处理规则提示词位置
- [ ] 10.11 明确知识库提示词位置
- [ ] 10.12 配置写入 MySQL / 后端事实源
- [ ] 10.13 配置读取与刷新一致
- [ ] 10.14 测试 + CI + 合并 main

---

# Task 11：水货生产

- [x] 11.1 `/shuihuo-production` 页面迁移
- [x] 11.2 水货生产 -> Batch Factory V11 入口
- ⚠️ 11.3 页面路由不闪退
- [x] 11.4 项目列表共用真实 MySQL 数据
- [x] 11.5 状态实时同步
- [x] 11.6 失败状态可见
- ⚠️ 11.7 对标旧 ECS 可用交互
- [x] 11.8 测试 + CI + 合并 main

> 仓库验收依据：PR #13 已合并 `main`；`main` CI run `37285192021` 中 Go tests、前台 tests、前台 build、管理端 build 全绿。ECS / 公网真实浏览器验收仍留到 Task 16，因此 Task 11 总体继续保持 `⚠️`。

---

# Task 12：剧本生成 / Director / Hook / 最终提示词

- [ ] 12.1 剧本生成 Go API
- [ ] 12.2 系统提示词后端管理
- [ ] 12.3 剧情模式通用元提示词
- [ ] 12.4 Hook
- [ ] 12.5 Director
- [ ] 12.6 最终提示词编译
- [ ] 12.7 单本生成状态
- [ ] 12.8 批量生成状态
- [ ] 12.9 失败重试
- [ ] 12.10 提示词版本持久化
- [ ] 12.11 v88 对标
- [ ] 12.12 测试 + CI + 合并 main

---

# Task 13：音频 + matchAudio

- [ ] 13.1 第三步生成音频
- [ ] 13.2 获取真实 `audioDurationSec`
- [ ] 13.3 第四步加入“匹配音频”开关
- [ ] 13.4 `matchAudio=true` 时总时长严格等于音频时长
- [ ] 13.5 第一镜从 0 秒开始
- [ ] 13.6 镜头之间无空缺
- [ ] 13.7 镜头之间无重叠
- [ ] 13.8 最后一镜结束时间 = `audioDurationSec`
- [ ] 13.9 10s 规则保留
- [ ] 13.10 15s 规则按“≤15s”执行
- [ ] 13.11 `matchAudio=false` 保留原 Director 逻辑
- [ ] 13.12 时间统一为秒并保留两位小数
- [ ] 13.13 不允许为了凑时长添加无关剧情
- [ ] 13.14 测试 + CI + 合并 main

---

# Task 14：视频生成完整链路

- [ ] 14.1 视频任务模型
- [ ] 14.2 `VIDEO` 状态机
- [ ] 14.3 personal_api provider
- [ ] 14.4 yd2.0-mini
- [ ] 14.5 豆包本地执行器
- [ ] 14.6 provider 状态 API
- [ ] 14.7 角色/场景/图片资产传递
- [ ] 14.8 视频生成提交
- [ ] 14.9 视频状态轮询
- [ ] 14.10 失败重试
- [ ] 14.11 TOS / 本地产物管理
- [ ] 14.12 视频合并 Worker
- [ ] 14.13 ffmpeg 合并链
- [ ] 14.14 最终结果回写 BatchProject
- [ ] 14.15 修复并验证 `personal_api` 状态接口不再 503
- [ ] 14.16 测试 + CI + 合并 main

---

# Task 15：发布 / 权限 / 登录认证

- [ ] 15.1 登录 Token 体系统一
- [ ] 15.2 Token 刷新
- [ ] 15.3 修复“当前登录状态已过期”闪退
- [ ] 15.4 API 白名单梳理
- [ ] 15.5 `/api/script-constraint-prompts` 权限兼容
- [ ] 15.6 发布权限验证
- [ ] 15.7 用户权限
- [ ] 15.8 团队权限
- [ ] 15.9 发布任务状态
- [ ] 15.10 权限失败不影响其他任务
- [ ] 15.11 测试 + CI + 合并 main

---

# Task 16：ECS 部署与最终公网验收

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

**下一任务：Task 9.2.4 — 展示书城来源。**