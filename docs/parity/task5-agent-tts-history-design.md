# Task 5：Agent、配音与历史 UI 迁移设计

## 目标与边界

将 `/agent`、`/agent/canvas`、`/tts` 与 `/history` 的 V88 工作台信息架构迁入 Go/React 用户端。迁移只覆盖各页面、其私有 CSS 与私有测试；不修改 `UserShell.jsx`、`RouterApp.jsx`、`api.js`，不调用旧 Node API、`/api/chat`、Token 流程或浏览器业务 LocalStorage。

页面状态只来自现有 Cookie Session 保护的 Go API：Agent 的项目、消息、画布、附件、技能、执行记录来自 MySQL/TOS；TTS 只读取并创建既有 Shuihuo 媒体任务与候选资产；History 只读取既有 BatchProject/Generation 投影。Provider、TOS 或执行器不可用时，UI 必须明确展示实际 `executor_unavailable`、`storage_unavailable`、等待或失败状态，不能合成成功卡片或可播放音频。

## 对照矩阵与现有接口

| 页面 | V88 源码对照 | 保留的 UI 结构 | Go API / 事实源 | 已知差异与处理 |
| --- | --- | --- | --- | --- |
| `/agent` | `frontend/src/user/pages/AgentPageV2.jsx` | 开始创作首屏、最近项目、项目搜索、进入/删除、加载/空/失败状态 | `GET/POST/DELETE /api/v1/agent/projects`；发送首条消息为 `POST /messages` | 不复制旧任务、Pet、连接器或 `/api/chat`；项目是新的 MySQL Agent Project。 |
| `/agent/canvas` | `frontend/src/user/pages/AgentPageV2.jsx`、`components/AgentCanvas.jsx`、`shared/styles/agent-workspace.css` | 三栏：版本/项目侧栏、画布中心、消息/技能/附件/执行记录右栏；移动端折叠为纵向区块 | Agent canvas、versions、messages、executions、skills、attachments endpoints；附件字节由 TOS 受权读取 | 不引入 React Flow 或旧任务模型；现有 JSON 画布是 Go revision 合约的编辑器，保留版本恢复和冲突提示。 |
| `/tts` | `frontend/src/user/pages/TtsPage.jsx` 与其卡片/工具条 | 页面标题、项目和状态筛选、创建抽屉、任务卡、试听/下载、刷新、失败重试、空态 | BatchProject/books/segments、Shuihuo media tasks/candidates/assets/retry、TOS content URL | 旧页用 Node `textToSpeech` 和本地临时卡片；新页不暴露模型/密钥，不创建无 Provider 的成功音频。 |
| `/history` | `frontend/src/user/pages/HistoryPage.jsx` | 标题、搜索、状态筛选、分页表、更新时间、打开项目、空/错/加载态 | `GET /api/v1/history` 的 MySQL 用户范围投影 | 不保留旧 Node script-history 的预览/删除/清空，因为 Go 投影不拥有第二套历史事实表。 |

## 实施方案

### Agent 首页

保留“描述需求 → 新建项目 → 写入首条服务端消息 → 进入画布”的顺序；任意一步失败显示真实错误。项目列表每次进入或删除后从服务端刷新。视觉上使用 V88 的深色工作台层次、标题区、项目卡和紧凑搜索，但不引入全局样式或共享壳层依赖。

### Agent 画布

保持三个独立的服务端刷新面：画布与版本、对话与执行记录、附件与技能。保存携带 revision；冲突只提示刷新而不覆盖服务器版本。附件列表使用受权内容 URL，权限错误显示明确不可访问状态。执行记录必须原样显示 `executor_unavailable` 等真实状态；消息发送后重新读取服务端会话。窄屏时三栏依次堆叠，避免横向截断。

### TTS

创建抽屉仅选择已保存的项目/小说/分镜；浏览器不提交 Provider、模型或密钥。任务卡只在 `succeeded` 且选中候选资产仍存在时提供音频控件及下载；失败或可重试失败显示服务端错误与重试，`pending_executor`/`executor_unavailable`/`storage_unavailable` 显示实际不可用原因。运行状态按现有轮询刷新；其他状态允许手动刷新。

### History

输入关键词、状态和页码作为服务端查询参数，数据不在浏览器缓存为历史事实。表格保留项目/小说、状态、更新时间、打开项目；空态解释当前账号没有可读项目，而不是制造样例记录。链接保留现有项目页路径。

## 测试先行计划

1. 为每页先补一个可观察的失败测试：Agent 项目刷新后的服务端恢复；画布三栏的执行不可用状态与附件访问；TTS 对不可用执行器绝不渲染播放器或下载；History 改变状态过滤后发送服务端筛选参数。
2. 逐个运行聚焦 Vitest，确认测试因尚未实现的行为失败，再仅修改任务 5 页面/CSS使其通过。
3. 运行所有前台测试与构建；运行与 Agent、History、Shuihuo media 相关的 Go HTTP/服务测试，确认本次未改变接口契约。

## 验收截图清单（局域网、登录后）

- `/agent`：最近项目加载、空态/错误态、创建项目并刷新后仍可进入。
- `/agent/canvas?projectId=…`：三栏桌面布局；窄屏纵向布局；画布保存、版本恢复、附件上传/删除、技能选择、执行不可用、刷新恢复与权限拒绝。
- `/tts`：筛选、创建抽屉、`pending_executor` 或 `executor_unavailable` 卡片、失败重试、真实候选资产的试听和下载。
- `/history`：关键词、状态筛选、分页、空态、跳转项目。

`/agent/canvas` 的公网内容策略曾阻断浏览器访问，因此本轮只做 V88 源码对照和局域网认证验收；在获得可控公网浏览器截图前，不宣称像素级公网视觉一致。

## 不在本任务内

- 公网/ECS 部署、Provider 或 TOS 凭据配置和真实付费 Provider 调用。
- 共享导航、路由和 API 基础设施的修改。
- 旧 V88 工作树、Node 运行时、旧 LocalStorage 历史或旧 dist 文件。
