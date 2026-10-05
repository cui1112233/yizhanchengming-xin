# 一战晟铭 · 新主线

这是从现有公网/v88 功能重新收敛出来的新主线。

## 迁移任务清单

- [查看 TASKS.md：公网/v88 → 新主线完整迁移进度](./TASKS.md)
- 后续每完成一个子任务，必须在同一次提交或 PR 中同步更新 `TASKS.md`；没有勾选并附验收依据，不算正式完成。

## 核心原则

- 生产后端统一使用 Go。
- React + Ant Design 仅负责前端；Node/npm 只允许用于前端构建，不作为生产服务。
- 业务主链统一为：React → Go API → MySQL/Redis → Go Worker → 外部 API/TOS。
- MySQL 是任务与业务状态的唯一事实来源；Redis 只用于队列、延迟任务、锁和短期运行态。
- 立即执行与自动化执行共用同一个 Pipeline，仅 `run_at` 不同。
- 121 优先使用官方 HTTP API，不以 Browser Worker 作为正常业务路径。
- 男女频优先级：用户手工填写 > 121 `bookinfo.category` > 已验证的 `genre` 映射 > AI 兜底。
- 确定性数据不得被 AI 覆盖。

## 第一阶段目标

跑通：添加书城/书籍 → 创建 Intake → 121 获取正文与 bookinfo → 男女频/风格元数据解析 → 创建批量项目 → MySQL 持久化 → 前端读取。
