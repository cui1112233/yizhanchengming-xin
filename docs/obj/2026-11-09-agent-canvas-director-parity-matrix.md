# 洛水 v2.9.0 → V88 第一期功能一致性台账

日期：2026-11-09

状态说明：`matched` 为完整一致；`equivalent` 为技术实现不同但用户效果等价；`partial` 为部分链路；`excluded` 为用户确认不做。没有本地运行和浏览器证据时不得标记为完整一致。

| 功能 ID | 洛水参考能力 | V88 当前位置 | 当前状态 | 已知差异 | 下一项证据 |
|---|---|---|---|---|---|
| `L29-AG-001` | Agent / Chat 模式 | `AgentPageV2.jsx`、Agent API | `equivalent` | 三种模式、上下文、历史和停止链已接通；停止通过浏览器中止请求并沿 Go request context 传递，不另建会话协议 | 不消耗额度前提下补一次真实进行中任务停止验收 |
| `L29-AG-002` | 工作区文件与历史会话 | Agent 项目、素材、生成历史 | `equivalent` | 已形成画布、故事板、素材、任务四视图，继续复用 V88 项目和生成历史数据 | 持续回归刷新恢复和窄屏布局 |
| `L29-AG-003` | 自动任务、批次、继续一步 | `agent_generation_tasks`、generation summary | `partial` | 状态机、失败隔离、取消、重试和恢复测试已完成，任务视图也已接入；当前登录项目无任务，未做真实按钮点击 | 使用无额度测试任务做浏览器取消/重试验收 |
| `L29-AG-004` | 技能选择和 `@` 引用 | skill picker、node mentions、附件 | `partial` | 节点、素材、附件已使用统一上下文和稳定 ID；技能引用的后端重读仍待浏览器核验 | 带素材和技能的刷新恢复验收 |
| `L29-CV-001` | 多媒体节点画布 | `AgentCanvas.jsx`、六类节点 | `partial` | 核心节点已有，需全链刷新和操作验收 | 六类节点 CRUD 浏览器证据 |
| `L29-CV-002` | 分组、框选、撤销重做 | Canvas 选择、历史、分组逻辑 | `partial` | 源功能存在，需刷新和边界验收 | 画布操作测试和浏览器证据 |
| `L29-CV-003` | 文件页统一素材 | Material Library、workspace assets | `equivalent` | 已有正式素材视图，并和抽屉、对话、画布复用同一份项目素材；空项目浏览器状态已核 | 带真实项目素材再验一次加入对话/画布 |
| `L29-CV-004` | 多视频卡独立生成 | video jobs、generation tasks | `partial` | 多节点失败隔离和统一状态投影已有测试；旧视频任务仍按后端契约回画布走专用链路 | 无额度多节点浏览器任务验收 |
| `L29-CV-005` | 悬浮画布助手 | conversation panel、bottom control | `partial` | 能开关，拖动/缩放/停靠一致性不足 | 交互与布局验收 |
| `L29-SB-001` | 剧本拆分分镜 | `AgentStoryboardWorkspace.jsx`、shot batch | `partial` | 结构化拆分、批建、版本化提示词契约和测试已接入 | 带真实剧本的浏览器拆分验收 |
| `L29-SB-002` | 分镜合并 | Agent 分镜合并 API 和工作区入口 | `partial` | 已有 10/15/30 秒相邻合并、资产/台词合并和旧结果保留测试 | 带真实分镜的浏览器合并验收 |
| `L29-SB-003` | 分镜与画布互通 | shot nodes、选中和工作区切换 | `partial` | 双向定位、稳定映射、幂等批建已实现并测试 | 浏览器双向定位验收 |
| `L29-DD-001` | 场景、人物、道具白模 | Director Desk `directorProject` | `partial` | 场景树、角色、道具、对象选择和变换测试已完成 | 对象变换浏览器证据 |
| `L29-DD-002` | 多机位和镜头参数 | Director Desk camera fields | `partial` | 多机位、活动机位、画幅及语义镜头参数已实现并测试 | 多机位浏览器验收 |
| `L29-DD-003` | 截图/白模回到创作链 | capture、canvas edge、video sync | `partial` | 截图和直接视频同步已有基础，完整失败链待核 | 上传/保存失败测试 |
| `L29-TH-001` | 多主题切换 | `theme.js`、User/Admin Layout | `implemented` | 已提供 13 个选项并按账号保存 | 全页面对比度持续回归 |
| `L29-TH-002` | 窗口背景关闭、预设 2～5、自定义上传 | `WindowBackgroundPicker.jsx`、成员背景接口 | `implemented` | 预设使用 V88 自有渐变图，不复制洛水图片文件 | 已登录设置页和跨设备回读 |
| `L29-PR-001` | Agent/资产/分镜/视频提示词 | `backend/internal/prompt/` | `partial` | 首批八个 v1 PromptDefinition 已建立并测试，当前接入文本节点、分镜拆分和视觉引用三个调用点 | 继续逐个接入剩余真实调用点并保存 metadata |

## 证据更新格式

每个状态变化追加：

```text
功能 ID：
V88 页面/代码：
API/数据：
Prompt ID/version：
自动测试：
本地运行：
浏览器验收：
差异与结论：
```

功能 ID：`L29-AG-001`
V88 页面/代码：`AgentPageV2.jsx`、`shared/api/agent.js`
API/数据：`POST /api/agent/chat`，AbortSignal 沿请求 context 中止，不删除既有历史
Prompt ID/version：沿用当前 Agent prompt 定义
自动测试：前端全量 1574/1574；新增停止链和请求 signal 测试
本地运行：Go embed `bfcf2b39`，`/health` 200，runtime manifest 为 `bfcf2b39`
浏览器验收：已登录页面实际展开三种模式；未发送付费请求，未人工触发进行中停止
差异与结论：用户可见效果等价，状态记为 `equivalent`

功能 ID：`L29-AG-002`、`L29-CV-003`
V88 页面/代码：`AgentWorkspaceHeader.jsx`、`AgentPageV2.jsx`
API/数据：继续使用项目 canvas、assets、executions、generations，不新增第二份数据
Prompt ID/version：不涉及
自动测试：四视图、导航、素材引用和全量前端测试通过
本地运行：Go embed `bfcf2b39`
浏览器验收：画布、故事板、素材、任务四个入口均可切换；素材和任务空状态来自当前真实项目
差异与结论：四工作区与统一素材达到等价实现，状态记为 `equivalent`

功能 ID：`L29-AG-003`、`L29-CV-004`
V88 页面/代码：`agent-generation-state.js`、`AgentPageV2.jsx`、generation task API
API/数据：统一任务状态、失败原因、取消、重试、确认回画布；旧视频任务保留专用链路
Prompt ID/version：生成任务继续记录现有 prompt metadata
自动测试：任务动作聚焦测试 47/47；Agent workspace、HTTP、prompt 后端测试通过
本地运行：Go embed `bfcf2b39`
浏览器验收：任务视图和真实空状态已核；当前项目无历史生成任务，未触发付费或额度任务
差异与结论：代码和运行链已完成，真实任务按钮尚无无额度样本，继续标 `partial`
