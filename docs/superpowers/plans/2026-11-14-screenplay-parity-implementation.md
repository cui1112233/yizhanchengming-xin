# 漫剧全流程补齐 Implementation Plan

> Execute with superpowers:subagent-driven-development, task reviews and final integration review. User approved implementation by “执行”. No new branches/worktrees; preserve unrelated dirty changes. Continue through tasks without repeated permission questions.

**Goal:** 完成全流程对照报告中缺失的可操作工作流，优先可靠生产与真实交付，不增加空入口。

**Architecture:** Go 三个独立功能单元：生产运行时、内容与协作、媒体工作台；MySQL 持久保存、Redis 控制任务，复用已有 TOS/TTS/ASR/AI-CUT 能力；React 分区组件调用真实接口。

**Tech Stack:** Go、React/Antd、MySQL、Redis、TOS、FFmpeg、Go embed。

## Global Constraints

- 系统预设提示词只在后端 prompt；不复制原软件源码/专有资源。
- 不进行付费供应商测试，不部署公网、不推送 Git。
- 不清理、重置、覆盖或提交其他工作区改动。
- 分项状态必须区分实现、模拟验证、真实页面、供应商实跑与像素对照。

### Task 1: 可靠生产运行时

Files: backend/internal/screenplay/production_*.go; backend/internal/httpapi/screenplay_production_*.go; existing batch_executor.go, screenplay_media_runner.go; storage migration.

- [x] 失败测试覆盖 >64 单元、后台继续、Redis 竞争、重启接续、逐镜头查询/复用、暂停和权限。
- [x] MySQL 持久任务/镜头，Redis 调度/锁；生产 start/status/pause routes，后台推进已有 BatchExecutor。
- [x] 质量检查调用真实后端提示词；保存阻断/问题，提供明确基础检查与质检状态。
- [x] 最终完成必须有可交付结果，或明确镜头生成阶段，不能只写 completed marker 冒充成片。
- [x] 相关测试、接口契约、审查通过。

### Task 2: 章节、助手与协作

Files: backend/internal/screenplay/authoring_*.go; httpapi/screenplay_authoring_*.go; prompt/screenplay_*.go; dedicated migration.

- [x] 测试章节修订/计划、助手会话及受限编辑、显示配置、批量分镜、归档/恢复与成员权限。
- [x] 实现章节/计划、作品内助手/Agent 会话与可审查操作；模型服务从现有配置解析。
- [x] 实现显示配置、批量分镜、归档/恢复、团队读取/编辑/生成权限，默认拒绝共享模型使用。
- [x] 保留历史项目；模型或服务缺失时返回实际失败，不伪造 AI 内容。
- [x] 相关测试、接口契约、审查通过。

### Task 3: 媒体与交付工作台

Files: backend/internal/httpapi/screenplay_media_workbench*.go; screenplay/media_*.go; dedicated migration; existing service integration.

- [x] 测试资产/镜头图片历史和参考图、音色/旁白、ASR 字幕、剪辑参数、时间线与媒体包导出。
- [x] 复用当前图片、TTS、ASR、AI-CUT 时间线与渲染服务，实现真实调用与持久结果，不另造假适配器。
- [x] 把实际参考图/音色进入媒体请求，把视频/音频/字幕及裁切顺序进入时间线；生成分集媒体/工程包及可访问结果。
- [x] 逐项检查归属、配置缺失、失败恢复和媒体路径安全。
- [x] 相关测试、接口契约、审查通过。

### Task 4: 前端完整工作流

Files: shared/api/comicScreenplay*.js; user/pages/ComicScreenplayPage.jsx; new focused panels and styles.

- [x] 用可测状态/API helpers 连接后台任务，页面只展示/控制不负责生产循环。
- [x] 章节/助手、资产/媒体、分镜批量、字幕/剪辑/导出、归档/团队/显示设置分别接真接口。
- [x] 新旧工作区均使用同一实际数据与服务；没有已接能力时显示配置原因，不画假按钮。
- [x] 相关前端测试与构建通过，逐控件检查请求/保存/显示链路。

### Task 5: 集成、审查、本地交付

- [x] 迁移和模块在 app/router 注册，复用原服务及停止上下文，跨模块职责清晰。
- [x] 独立审查关键正确性、权限、重复付费风险、数据完整性；修复发现。
- [x] 相关集成测试/构建，冻结包，空闲后只替换已核实本地程序。
- [x] 本地 4001 登录界面真实验收，记录可见结果、版本与仍未实跑的供应商能力。
- [x] 逐项更新 57 项对照状态，不能将部分阶段完成标为全部复刻完成。

验收范围及未验证边界详见 `docs/audit/2026-11-14-screenplay-implementation-followup.md`。计划完成表示本轮实现与约定测试已完成，不代表原软件57项全部一比一或供应商整链已实跑。
