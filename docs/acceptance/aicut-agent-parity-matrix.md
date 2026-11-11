# AI-CUT Agent 对照验收矩阵

日期：2026-11-10

规则：只有同时存在源码、自动测试和 4001 可见运行证据时才标记完成。当前先记录基线，不把计划或页面骨架当成功能完成。

| 行为 | 洛水静态证据 | 一战晟铭目标 | 自动测试 | 4001证据 | 状态 |
|---|---|---|---|---|---|
| 全局导航收起 | 左上角导航折叠按钮 | 保留 UserLayout 独立状态 | `aicut-workbench.layout.test.js` 通过 | 浏览器确认 Agent 收起时全局导航仍展开 | 完成 |
| Agent 内部收起 | `AiCutChatPanel.collapsed`、`toggle-collapse`、窄轨“展开 Agent” | 350px 完整态/56px 窄轨态 | 组件与布局守卫通过 | 实测列宽 `56px 450px 747px`，可再次展开 | 完成 |
| Agent/Q&A | Q&A 明确不改时间轴 | Q&A 禁止执行任何工具 | `TestAgentQAModeNeverExecutesToolCalls` 通过 | 4001 可切换问答并显示不改时间线 | 完成 |
| 模型选择 | 项目 Agent 保存所选模型 | 使用 API 配置中的文本模型 | API/组件测试通过 | 4001 显示已配置文本模型并持久本地选择 | 完成 |
| 项目聊天恢复 | agentChat 与 tracks/items/captions 一起保存恢复 | MySQL 按 owner/project 保存消息 | Agent store/MySQL store 测试通过 | 页面通过 `/agent-messages` 加载 | 完成 |
| 工具开始/完成/失败 | thinking/tool-start/tool-done/usage 事件 | 按 sequence 持久化并增量读取 | Agent service/store 测试通过 | 面板已连接增量事件列表 | 完成 |
| 时间线真实修改 | 工具结果回灌模型并修改同一项目 | 每个写工具使用 expectedRevision | move/delete/split/speed/replace/place/undo/redo 测试通过 | 页面收到 tool_done 后重载同一工程 | 部分：未消耗额度做线上模型实调 |
| Agent 识别字幕 | Agent 调用字幕识别能力并回灌任务结果 | `transcribe_asset` 提交现有 ASR 队列，返回 job，不冒充已完成 | `TestAgentCompositeToolsCallConfiguredServices`、ASR service 测试通过 | 需有 Redis、素材存储和 ASR 模型；缺配置会明确失败 | 完成（未消耗真实额度） |
| Agent 智能切镜 | 洛水说明与 `capcut-cli detect-scenes` 参数一致 | `smart_split_asset` 复用页面 FFmpeg 场景检测与时间线保存 | 复合工具、路由、切点/连续性测试通过 | 默认 threshold=0.4、min gap=2秒 | 完成 |
| Agent 分离音频 | Agent 可调用编辑器媒体工具 | `extract_audio` 复用页面 FFmpeg 音频提取并写入素材库 | 复合工具重放与路由相关测试通过 | 缺 FFmpeg/对象存储会明确提示 | 完成 |
| Agent 去口误停顿 | Agent 基于字幕与模型删除口误/停顿 | `clean_oral_pauses` 复用页面后端 prompt 和时间线裁切链 | 复合工具与 `TestAICutOralCleanUsesSeparateTimelineCutRoute` 通过 | 需要已配置文本模型，不在前端硬编码提示词 | 完成（未消耗真实额度） |
| 停止运行 | 原版有取消检查和停止动作 | cancel_requested -> canceled | 取消测试通过 | 运行中显示“停止”并调用取消接口 | 完成 |
| 30轮上限 | 原版 Agent 循环上限 30 | 第30轮仍无最终回复则失败 | `TestAgentStopsAfterThirtyTurns` 通过 | 非视觉规则 | 完成 |
| 正常输入 | 执行工具并返回结果 | 修改时间线、revision 增加 | 服务与工具集成测试通过 | 建议卡填入、发送可用；未点付费模型 | 部分 |
| 空值输入 | 不执行 | HTTP 400、无运行、无修改 | HTTP 空值测试通过 | 空输入发送按钮禁用 | 完成 |
| 重复操作 | 不应重复应用同一工具 | clientRequestId/callId 幂等 | HTTP 与工具重放测试通过 | 非视觉规则 | 完成 |
| 错误操作 | 工具失败反馈模型 | tool_failed、时间线不变 | 锁轨、边界、版本冲突测试通过 | 错误会显示在事件流 | 完成 |
| 取消操作 | 停止后续执行 | canceled、没有后续工具 | 取消前检查测试通过 | 停止按钮已接入 | 完成 |

## 当前基线结论

旧的 `resolveAICutAgentTool` 五关键词分发已从页面删除。当前链路为：页面创建持久 Agent run → Redis/本地队列 → 最多 30 轮文本模型循环 → 白名单工具修改同一项目时间线 → 工具结果回灌模型 → 事件和消息持久化 → 页面增量读取并刷新时间线。

未做的最后一项是使用真实付费文本模型发送一次指令；本轮遵守“不主动消耗额度”约束，因此用确定性模型替身完成了相同输入的正常、空值、重复、错误和取消测试。不能把这一项写成真实模型线上已验证。
