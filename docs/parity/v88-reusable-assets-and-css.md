# V88 可复用 CSS 与资产迁移清单

## 使用边界

此清单是迁移候选与审查规则，不是“复制旧 `dist`”的指令。公网是旧 Express 运行时，当前新站是 React + Ant Design + Go embed；应提取项目拥有、可授权、可隔离的源文件/设计 token，再由 React 组件引用。不得引用公网 hash 文件作为长期依赖。

|类别|公网观察/旧源码候选|可否直接迁入新站|迁入方式与限制|
|---|---|---|---|
|品牌图|旧源码 `public/assets/brand-logo-black.png`、`brand-logo-white.png`；公网首页/管理端可见。|条件可用|先确认版权、文件 hash 和深浅主题用途；复制到新站受控静态资产并由 `UserShell`/品牌组件引用。|
|首页视频|旧源码 `public/assets/home-hero.mp4`；公网首页 Hero 使用。|条件可用|需确认来源、大小、移动端降级与无障碍 fallback；Go embed 后必须由同源静态资源服务，不能依赖公网旧 URL。|
|首页布局样式|`frontend/src/user/pages/HomePage.jsx`、`frontend/src/shared/styles/global.css`、`home-gradient-button.css`。|可提取，不可整份直搬|抽取 CSS variables、Hero/卡片/断点，挂在 `.user-shell` 或页面根节点；用 AntD token 覆盖，避免污染 Batch、Novel Panel 等页面。|
|用户外壳|`frontend/src/shared/layouts/UserLayout.jsx`、`account-center-visual-rebuild.css`、`member-center.css`。|可提取结构/视觉 token|迁到 `前台/src/UserShell.jsx`、`AccountCenterPage.jsx` 的局部样式；保留 Go 当前用户/登出而非旧 auth client。|
|剧本页样式|旧 `ScriptPage.jsx` 相邻 CSS、共享 global 样式。|按组件迁移|只迁移网格、间距、控件状态；请求、抽屉状态和草稿由 React 状态与 Go API 重建。|
|小说获取/改文|旧 `NovelFetchPage`、`NovelFetchWorkshop` 组件及相邻 CSS；公网出现独立 `NovelFetchPage-*.css`。|按页提取|先把三条路由拆开再迁移。保留新站 intake 数据模型，不复用 Node upload/web-submit 会话。|
|批量工厂|旧 `BatchFactoryPageV11.jsx`、`BatchFactoryWorkbenchPage.css`、`batch-factory-v11/*.css`；公网有 `BatchFactoryWorkbenchPage-*.css`。|仅样式和静态文案候选|不要复制 V88 V11/V12 业务逻辑；按当前 Go batch project、generation、video 模型重新绑定动作。|
|水货生产|旧 `ShuihuoProductionPage.jsx`、`新·批量工厂/ShuihuoProductionPage.jsx`、`shuihuo-production.css`；公网有 `shuihuo-production-*.css`。|按模块迁移|可复用作品卡、搜索/筛选栏、网格/列表视觉；卡片必须读取当前用户项目，删除/打开必须走 Go 权限校验。|
|Agent|旧 `AgentPageV2.jsx`、`agent-workspace.css`；公网有 `AgentPageV2-*.css`。|局部迁移|只迁移版式/加载占位；使用新 `/api/v1/agent/projects`、canvas/versions API，绝不恢复 `/api/agent/chat`。|
|TTS/历史|旧 `TtsPage.jsx`、`HistoryPage.jsx`、`public/css/tts.css`。|局部迁移|可复用卡片、表格、空态视觉；历史和任务来自 Go MySQL 投影，不能使用浏览器列表。|
|问题/设置|旧 `IssueLogPage.jsx`、`SettingsPage.jsx`。|局部迁移|复用 loading、筛选、设置分组表达；错误文本必须用 Go 脱敏视图，密钥不下发。|
|宠物装饰|公网请求 `/pets/stacky/spritesheet.webp`、`/pets/pixiu/spritesheet.webp`；旧 `cm-penguin-companion.css`。|默认不迁|需核实许可、动画性能、可访问性和用户偏好；若保留，应是可关闭的纯装饰，不能绑定业务状态。|
|图标|公网加载 `createLucideIcon-*.css`，旧项目使用 lucide 组件。|不复制 hash CSS|在新 React 项目通过依赖的图标组件显式 import；保持 aria-label/tooltip。|
|字体|本轮 DOM/CSS 采样未发现可确认的自定义字体文件。|不能假定可迁|在取得字体来源与许可前使用系统/AntD 字体栈；不要从压缩 bundle 猜测或提取字体。|

## 不可直接搬运的运行时内容

|内容|为何不能迁|新站正确边界|
|---|---|---|
|`frontend/dist`、公网 `assets/*-HASH.{js,css}`|哈希构建产物绑定旧 Express 路由、Node API 和当时依赖；直接引用会绕过 Go embed 构建。|从 React 源文件重建，执行前台 build 后交给 Go `embed`。|
|旧 Node API client 与 `/api/chat`|请求模型、鉴权和错误语义不兼容；`/api/chat` 被本次迁移规则禁止。|在 `前台/src/api.js` 对接 `/api/v1/*`，服务端 prompt 留在 Go prompt 模块。|
|Bearer token、`getToken()`、登录/上传 session|会泄露会话或创建第二套认证；部分依赖旧浏览器 Cookie。|仅用现有 HttpOnly Cookie Session、CSRF same-origin 与 capability。|
|localStorage/sessionStorage 中的项目、历史、任务、设置|它们不是可审计的业务事实，跨设备/刷新会失真。|项目、历史、问题、偏好和执行器状态以 MySQL/现有任务运行时为准；localStorage 仅可保存非敏感 UI 偏好。|
|用户头像 `/user-content/avatars/...`、项目封面、用户素材|归属特定账号，可能是私密/TOS 媒体，不是公共设计资产。|按当前用户授权经 API/TOS 读取；审计文档不复制媒体。|
|provider 密钥、发布 credential、MFA/恢复码、执行器 pairing secret|高敏感信息，且 V88 的配置格式不应成为迁移接口。|服务端密封保存；前端只见脱敏状态和有效性。|
|旧自动化 scheduler/worker/队列实现|迁入会制造第二套 Runtime/Queue/Worker，违背架构约束。|复用既有 Redis Queue、Lease、Lock、taskruntime；无执行器明确显示不可用。|

## 建议的迁移顺序与验收

1. 先建立每页的 React DOM 骨架和真实 Go 数据绑定，再逐段移植上述局部 CSS；不先复制整份全局样式。
2. 每次迁移用桌面和窄屏分别验收：导航折叠、顶栏、内容滚动、弹层焦点、空态、loading、错误、无权限和刷新恢复。
3. 所有静态媒体以构建后的 Go embed 服务地址验收；确认不是开发 Vite 地址，也不是公网旧 Express 资源。
4. 视觉评估以公网每页真实状态为基准；用户登录数据造成的作品/空态差异应记录为数据差异，不伪造相同卡片。

## 可追溯文件索引

### 当前新站

- 路由与外壳：`前台/src/RouterApp.jsx`、`前台/src/UserShell.jsx`。
- 首页/剧本/改文/面板：`HomePage.jsx`、`ScriptWorkspace.jsx`、`NovelFetchWorkshop.jsx`、`novel-panel/NovelPanelWorkbench.jsx`。
- 生产页：`App.jsx`、`BatchFactoryHome.jsx`、`BatchProjectListPage.jsx`、`BatchProjectVideo.jsx`、`ShuihuoProductionPage.jsx`。
- 账户与工作区：`TtsPage.jsx`、`HistoryPage.jsx`、`IssuesPage.jsx`、`SettingsPage.jsx`、`AccountCenterPage.jsx`、`AgentStudioPage.jsx`、`AgentCanvasPage.jsx`。
- API/权限：`前台/src/api.js`、`api/internal/httpapi/server.go`、`api/internal/httpapi/issues_handlers.go`、`api/internal/httpapi/history_handlers.go`。

### V88 源码候选（只读参考）

- 用户页与布局：`frontend/src/user/pages/*`、`frontend/src/shared/layouts/UserLayout.jsx`、`frontend/src/shared/styles/*`。
- API 语义：`frontend/src/shared/api/{auth,generation,novelFetch,novelFetchWorkshop,batchFactoryV11,shuihuoProduction,agent,tts,history,member,admin}.js`。
- 静态资产：`public/assets/brand-logo-black.png`、`public/assets/brand-logo-white.png`、`public/assets/home-hero.mp4`。
