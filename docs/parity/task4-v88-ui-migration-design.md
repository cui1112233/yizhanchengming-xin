# Task 4：V88 剧本与小说工作流 UI/交互迁移设计

## 目标

将旧 V88 的 `/script`、`/novel-fetch`、`/novel-fetch-workshop`、`/novel-panel` 的页面层级、视觉语言、文案、卡片、抽屉、表单、空态、加载态和响应式布局迁入 Go + React 前台。所有业务事实来自已有 Go HTTP API、MySQL 和 TOS；不得保留 Node API、iframe、`postMessage` 业务桥接、Bearer token、`/api/chat` 或浏览器业务存储。

## 已确认事实源

| 页面 | 已有 Go/MySQL 事实源 | 当前需接线或补齐的范围 |
| --- | --- | --- |
| `/script` | Intake/Book 原文、项目设置、StageRun、提示词目录、`script_storyboard_documents/cards` | 人物、场景、约束、时长必须以项目设置恢复；分镜卡更新、排序、删除、重编译走 `ScriptStoryboardService` |
| `/novel-fetch` | Intakes、Books、执行和重试 | 来源、平台、批量书号、进度、错误和重试全部只读/写 Go API；无前端任务缓存 |
| `/novel-fetch-workshop` | Intake 的 `workshop_settings_json`、Book 原文、服务端 prompt 目录、恢复原文 | 删除上传交接存储；若 Go 不提供实际上传任务、TOS 对象或处理执行契约，记录缺口并不显示成功状态 |
| `/novel-panel` | `novel_panel_workspaces/history`、MySQL store、项目权限中间件 | 原文、人物、关系、分镜、样式、历史恢复直接运行 React 工作台；不嵌入旧静态页面 |

## UI 迁移策略

以旧 V88 页面作为视觉与交互基线，迁移可复用的页面 DOM 分区、CSS 类名与静态资产；React 组件只承担呈现、暂存编辑和调用 Go API。短暂的未提交输入可留在 React state，但任何刷新、重新登录、历史恢复后的内容均通过服务端重新读取。服务端未具备的真实操作不得使用占位结果代替。

不改动 `前台/src/UserShell.jsx`、`前台/src/RouterApp.jsx` 或 `前台/src/api.js` 的共享结构；只使用已存在的 API 函数。新增接口时由局部页面适配器调用既有 `requestJSON` 模式，避免在页面中持有认证细节。

## 旧页面来源记录

| 页面 | DOM / CSS 参考 | 新页面事实源 |
| --- | --- | --- |
| `/script` | 旧 `frontend/src/user/pages/ScriptPage.jsx` 的 `script-workbench`、`script-left/right`、`script-chat-shell`；`frontend/src/shared/styles/global.css` 相邻局部规则 | 批量项目、Book、production settings、StageRun 与 `script_storyboard_*` MySQL 记录 |
| `/novel-fetch` | 公网加载态与旧 `NovelFetchPage` 的独立工作台分区；`NovelFetchPage-*.css` 仅作视觉参考 | Intake/Book/execute Go API |
| `/novel-fetch-workshop` | 旧 `NovelFetchWorkshopPage.jsx` 的 `legacy-panel-card`、任务表与原文抽屉分区 | Intake Workshop snapshot、设置、提示词目录和 Book 原文 |
| `/novel-panel` | 旧 `public/novel-panel/workbench/style.css` 的卡片、画布和窄屏规则 | `novel_panel_workspaces/history` Go API 与 MySQL store |

## 失败与权限语义

- 401：显示登录过期并交给全局会话恢复；不得保存业务状态到浏览器。
- 403：显示无权限；不尝试刷新权限、不调用其他写接口。
- 409：保留当前编辑内容，提示重新加载或合并。
- 422/5xx：显示服务端脱敏错误；重试仅重放对应的 Go 操作，不伪称成功。
- Provider、凭据或产品决策尚缺：以明确不可用原因呈现，并追加到缺口文档。

## 测试与证据清单

1. Go：路由权限、同源写保护、MySQL Store 持久化、历史恢复、分镜保存/排序/重编译、Intake 执行和 Workshop 原文恢复。
2. 前端单测：四页读取服务端事实；创建或保存；刷新重载；401/403；失败重试；不读取/写入 `localStorage`、`sessionStorage`、iframe 或 `postMessage`。
3. 构建：`go test ./...`（在 `api`）与 `CI=1 npm test -- --run`、`npm run build`（在 `前台`）。
4. 浏览器：局域网 `10.0.11.112:18080` 的认证会话，对四页分别保存截图；每页记录创建、保存、刷新恢复、重新登录恢复、403、失败重试。公网仅保留旧 V88 参考截图，不部署 ECS。

## 约束

- 未取得真实 API 响应、MySQL 记录或 TOS 回读前，不声明生成、上传或恢复成功。
- 大段系统提示词只留在 Go prompt 模块；前台显示名称和版本，不硬编码提示词正文。
- 此任务不修改公共导航壳、路由壳或共享 API 文件。
