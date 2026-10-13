# SD One-Call Asset Preset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将人物名单、人物事实、人物外形、人物关系合并为一次 AI 调用，输出完整、可审计且可供 SD 分镜复用的资产预设。

**Architecture:** 保留结构化人物字段和关系字段作为系统事实，同一次响应直接返回每个人物、场景、道具的完整 `prompt`，并冻结供后续分镜复用。批量工厂不再分别触发四个独立人物阶段。

**Tech Stack:** Node.js/React、系统预设目录、现有文本 AI 调用协议、Node test。

**Spec:** `G:/Backup/Documents/WXWork/1688857688816333/Cache/File/2026-09/SD_AI完整处理流程与新系统接入指令.md`

## Global Constraints

- 原文是剧情事实和顺序的最高权威。
- 完整配音和逐行时长由系统负责，不能交给 AI 决定。
- 资产 `prompt` 不能替代结构化人物事实、外形和关系字段。
- 一次业务调用只选择一种 AI 协议，只发送一次，不自动切换模型或协议。
- 系统提示词放在后端预设目录，前端只选择预设，不硬编码完整提示词。

## Review Focus

- 一次响应必须同时包含人物、事实、外形、关系和可渲染资产提示词；测试覆盖字段完整性。
- 同一人物的 `slot_id`、别名和外形锚点必须稳定；测试覆盖别名绑定。
- AI 返回合法 JSON 但缺字段时必须阻止进入分镜阶段；测试覆盖业务校验失败。
- 原文中的提示注入文字不能覆盖系统协议；测试覆盖数据边界。
- 历史四阶段任务不能被默认静默改写；测试覆盖旧任务版本保留。

### Task 1: 定义并发布一次调用的 SD 资产预设

**Files:**
- Modify: `lib/system-preset-catalog.js`
- Test: `lib/system-preset-catalog.test.js`

**Interfaces:**
- Produces: 一个资产预设，固定返回 `characters`、`relationships`、`asset_prompt` 及版本化协议字段。

- [x] 现有 SD 预设已声明一次调用、保留结构化字段，并禁止把结果定义为摘要。
- [x] 现有预设正文明确一次请求同时处理名单、事实、外形、关系。
- [x] 运行系统预设测试并确认通过。

### Task 2: 将批量工厂资产阶段改为一次调用

**Files:**
- Modify: `routes/batch-factory-v11.js` 或实际资产阶段编排文件
- Modify: `lib/batch-factory/*` 中负责人物阶段的调用和校验模块
- Test: 对应批量工厂源代码测试和流程测试

**Interfaces:**
- Consumes: Task 1 的资产预设 ID。
- Produces: 冻结资产对象，包含结构化人物数据、关系图和 `asset_prompt`，供故事剧本与导演阶段使用。

- [x] 增加失败回归测试：同一任务的 SD 人物资产阶段只产生一次逻辑 AI 请求。
- [x] 合并输入上下文，保留完整原文、风格和人工人物槽位。
- [x] 对一次响应执行 JSON 解析、字段校验、别名归一和人物锚点冻结。
- [x] 复用现有审计/持久化链路保存调用结果。
- [x] 运行流程测试，确认 SD 预设不会进入逐人物编译。

### Task 3: 让故事剧本、导演分镜和最终 Prompt 使用冻结资产

**Files:**
- Modify: 故事剧本/导演分镜编排模块
- Modify: 最终 SD Prompt 渲染模块
- Test: 分镜 Prompt 结构测试

**Interfaces:**
- Consumes: Task 2 的冻结资产对象。
- Produces: 含统一风格、人物锚点、场景和镜头的最终 SD 分段 Prompt。

- [x] 现有冻结资产和最终 Prompt 链路保持资产锚点，不重新调用人物 AI。
- [x] 故事剧本保持一行原文一行故事剧本。
- [x] 导演只填写固定镜头画面，不改变行数、顺序和时间。
- [x] 最终 Prompt 保留多场景多镜头结构。
- [x] 运行相关 Go 流程测试。

### Task 4: 前端显示调用结果和资产预设

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Test: 对应 source tests

- [x] 界面显示 SD 资产预设，并使用同一个资产选择入口。
- [x] 删除会误导用户的独立人物、场景提示词选择。
- [x] 保留结构化结果查看入口；旧配置会在保存时清除退休选择。
- [x] 运行前端定向测试和构建。

### Task 5: 集成验证

- [x] 对照 8081 真实任务确认最终 Prompt 的资产引用结构。
- [x] 验证一次资产 AI 调用和现有故事剧本/导演/最终 Prompt 链路。
- [x] 验证历史配置只在批量工厂资产选择处迁移，不改写历史业务数据。
- [x] 运行定向测试、前端构建和 Go 资产流程测试。
- [ ] 记录提交 SHA 和运行环境问题（待提交后完成）。
