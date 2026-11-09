# 洛水 v2.9.0 → V88 第一期功能一致性台账

日期：2026-11-09

状态说明：`matched` 为完整一致；`equivalent` 为技术实现不同但用户效果等价；`partial` 为部分链路；`excluded` 为用户确认不做。没有本地运行和浏览器证据时不得标记为完整一致。

| 功能 ID | 洛水参考能力 | V88 当前位置 | 当前状态 | 已知差异 | 下一项证据 |
|---|---|---|---|---|---|
| `L29-AG-001` | Agent / Chat 模式 | `AgentPageV2.jsx`、Agent API | `partial` | 已有创作助手和对话，完整权限/停止链待核 | 模式与停止测试、浏览器验收 |
| `L29-AG-002` | 工作区文件与历史会话 | Agent 项目、素材、生成历史 | `partial` | 数据存在但未形成四工作区统一入口 | 四视图导航和刷新恢复 |
| `L29-AG-003` | 自动任务、批次、继续一步 | `agent_generation_tasks`、generation summary | `partial` | 队列协调器和重启恢复尚未全部完成 | Go 状态/恢复测试 |
| `L29-AG-004` | 技能选择和 `@` 引用 | skill picker、node mentions、附件 | `partial` | 节点/附件已有，文件和技能引用契约待统一 | 稳定 ID 和后端重读测试 |
| `L29-CV-001` | 多媒体节点画布 | `AgentCanvas.jsx`、六类节点 | `partial` | 核心节点已有，需全链刷新和操作验收 | 六类节点 CRUD 浏览器证据 |
| `L29-CV-002` | 分组、框选、撤销重做 | Canvas 选择、历史、分组逻辑 | `partial` | 源功能存在，需刷新和边界验收 | 画布操作测试和浏览器证据 |
| `L29-CV-003` | 文件页统一素材 | Material Library、workspace assets | `partial` | 当前以抽屉为主，缺正式素材工作区 | 素材视图和引用检查 |
| `L29-CV-004` | 多视频卡独立生成 | video jobs、generation tasks | `partial` | 可生成但批量失败隔离和状态投影待核 | 多节点任务测试 |
| `L29-CV-005` | 悬浮画布助手 | conversation panel、bottom control | `partial` | 能开关，拖动/缩放/停靠一致性不足 | 交互与布局验收 |
| `L29-SB-001` | 剧本拆分分镜 | `AgentStoryboardWorkspace.jsx`、shot batch | `partial` | 拆分/批建基础已有，提示词契约待统一 | 结构化拆分测试 |
| `L29-SB-002` | 分镜合并 | 批量工厂已有相邻分镜语义 | `partial` | Agent 分镜工作区尚无正式合并入口 | 10/15/30 秒合并测试 |
| `L29-SB-003` | 分镜与画布互通 | shot nodes、选中和工作区切换 | `partial` | 批建已有，双向定位和稳定映射待补 | 幂等和定位测试 |
| `L29-DD-001` | 场景、人物、道具白模 | Director Desk `directorProject` | `partial` | 对象基础已有，完整对象树需验收 | 对象变换测试/浏览器证据 |
| `L29-DD-002` | 多机位和镜头参数 | Director Desk camera fields | `partial` | 语义机位已有部分实现，多机位/画幅待核 | 状态迁移和机位验收 |
| `L29-DD-003` | 截图/白模回到创作链 | capture、canvas edge、video sync | `partial` | 截图和直接视频同步已有基础，完整失败链待核 | 上传/保存失败测试 |
| `L29-TH-001` | 12 个主题切换 | `theme.js`、User/Admin Layout | `partial` | 当前仅深色/浅色 | 主题目录、账号偏好、全页面对比度 |
| `L29-PR-001` | Agent/资产/分镜/视频提示词 | 分散在现有服务和前端契约 | `partial` | 尚无首批八个统一 PromptDefinition | Prompt ID/版本/解析测试 |

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
