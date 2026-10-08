# V88 可复用 CSS 与资产迁移清单

> 2026-10-09 范围决策 superseded 本文中 Agent 和问题日志的复用建议：旧 Agent 前端/HTTP 实现与 Issues 页面/API 已退役，不得作为当前代码入口。Agent Go 领域包、migration 与已有数据仅作未来重新设计的可恢复基础；下文资产记录仍保留为历史审计证据。

## 使用边界

此清单是迁移候选与审查规则，不是“复制旧 `dist`”的指令。公网是旧 Express 运行时，当前新站是 React + Ant Design + Go embed；应提取项目拥有、可授权、可隔离的源文件/设计 token，再由 React 组件引用。不得引用公网 hash 文件作为长期依赖。

|类别|公网观察/旧源码候选|可否直接迁入新站|迁入方式与限制|
|---|---|---|---|
|品牌图|公网首页 PNG；管理端特权 DOM 未获访问。|已确认文件一致，许可仍待核实|现有前台同名文件与公网 SHA-256 一致，无需重复制；由同源 Go 静态资产引用。|
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

## 2026-10-09 Foundation 主题证据与实现边界

本轮只读审计窗口为 2026-10-09 02:28–02:32 Asia/Singapore。来源为
`http://115.190.156.223:3000` 的首页、剧本、设置 DOM/computed-style 及当前公开静态 CSS，
旧 Express 的 `/api/build-info` SHA 为 `2011ffb533eba2aa1041a2f67e40ac81caa0916c`。
该身份只用于追溯旧站证据，不能证明新 Go 运行时。公开 hash CSS/JS 仅是审计来源，项目未引用或复制它们。

共享值由 `ui/theme-contract.js` 维护，两端 `src/theme.js` 仅连接各自 AntD 算法。
CSS variables 由同一 JS 合约安装到 html，以供组件及 portal 继承；没有第二份手写颜色表。

|角色|深色|浅色|证据层级|
|---|---|---|---|
|背景/面板/卡片|#070b11 / rgba(11,17,26,.96) / rgba(14,22,33,.96)|#f4f6fa / rgba(255,255,255,.9) / #fff|深色含 computed corroboration；浅色仅静态 CSS|
|悬停卡片/输入|rgba(20,31,45,.98) / #0b111a|#f5edf0 / #fff|公开 CSS 语义表|
|正文/次要/弱文字|#f5f7fb / #8391a4 / #586679|#202330 / #667085 / #98a2b3|公开 CSS，深色采样|
|边框/强边框|rgba(151,168,190,.12) / rgba(176,195,219,.2)|#e4e7ee / #cfd5e1|公开 CSS|
|悬停阴影|0 8px 32px rgba(0,0,0,.22)|0 8px 24px rgba(31,41,55,.08)|公开 CSS|
|主色/悬停/淡底|#f07167 / #f08a7d / rgba(240,113,103,.14)|同深色|公开声明及设置主按钮 computed|

字体采用公开注入 AntD 的 `-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif`，
基础字号 14px。公开 body/home computed 为 sans-serif，实际 glyph fallback 未测；
统一系统栈是批准的两端一致性选择，不是所有元素 computed 字符串逐字相同的证明。
公开 shell 未观察到字体二进制或 @font-face，Inter 只出现在无关页面；本次无字体引入、下载或许可推断。

外壳角色为 expanded220 / collapsed60 / rail64 / topbar56 / gutter24 / contentMax1480px；
控件角色为 small24 / default32 / homePill42px，圆角为 form8 / nav10 / card14px。
这些值按组件使用，不覆盖所有控件；6px 小按钮圆角、999px 首页 pill 等局部角色例外保留。
工作区顶栏 56px 不等于公开 flex 压缩后的品牌行 32px，也不等于窄屏首页导航实测约 83.94px。
900px rail 和 720px Drawer 阈值必须作为 CSS media literal 保留，var() 不能作为 media-query 阈值。

本地用户 Drawer 在 <=720px 保留，作为可用性/无障碍差异明确记录。
721–900px 侧栏使用 64px 宽度，但本地导航只有文本，没有旧站 icon markup；
文字保持可见并换行以保证可达，不能宣称旧站的图标轨道逐像素一致。
本地首页视频 Hero、登录渐变/48px 控件/24px 卡片、头像和装饰渐变、
首页 1180px 两列布局等局部视觉仍为现有设计差异；其它业务页面 CSS 后续逐页迁移。
这次没有扩大路由、导航 JSX 或业务修改范围。

管理端保留 light 行为及现有 Sider 几何/breakpoint/collapsedWidth；
两端 Card.headerHeight52、Drawer.footerPaddingBlock16/footerPaddingInline24、
Table.cellPaddingBlock14 是本地兼容值，不是公网 Admin 测量值。没有两端 token/component 差异。
公开 /admin 重定向至普通 /profile，特权 DOM、密度、行为及截图均未验证。
dark 管理主题只验证合约函数，不代表 dark 管理 UI 验收。

现有品牌图均为 1920×1080；黑图 SHA-256
`e269d71008daafff7dfb812870d18d102152b0866c10ed00fa8377eee7692e51`，
白图 SHA-256 `32682771272a85ff66ff6846042dee025a4a3b5bbe4dca449379fad45ed9003d`。
前台文件逐字节相同，文件来源/版权许可仍是独立 gate。

html/body/#root 的高度、body margin/min-width、AntD reset 是窄 canvas/reset 例外；
body overscroll 只消费合约背景变量。视觉规则归属前台/后台 root，移动 Drawer 用已有专属类，
未增加全局 AntD selector 或 universal *。用户主题仍由 Go 偏好恢复，localStorage 只保留既有非敏感 UI fallback；
管理端没有新增偏好机制。

尚需独立运行时验收：本地 Go 精确 SHA/served bundle、真实认证/API/capability、桌面与窄屏、
light/dark computed/screenshot、服务端主题恢复与刷新、键盘/focus/Drawer 关闭导航、
loading/empty/error/retry/无权限。Vitest 和 Vite production build 不关闭这些 gate，
production build 也不证明 Vite dev-server 对根目录相对 import 的 fs 许可。

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
