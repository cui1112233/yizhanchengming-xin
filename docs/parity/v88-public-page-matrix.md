# V88 公网逐页对照矩阵

> 2026-10-09 范围决策 superseded 本文中 Agent 和 `/issues` 的历史迁移建议：当前 `/agent` 与 `/agent/canvas` 只保留占位，旧 Agent React/HTTP adapter 已退役但领域包、migration 和数据保留；`/issues` 页面与 HTTP API 已退役。下表对公网 V88 的观察仍作为历史证据，不再是当前实施目标。

## 审计边界与证据

- 审计日期：2026-10-08；公网观察地址为 `http://115.190.156.223:3000`，使用已登录的普通用户会话，仅作只读浏览。
- 新站基线：`main` 的 `d2c715942e73f44c3658a66cc2ee587477227d2b`；对应本机 Go embed 站点及其源码。公网目前仍返回旧 Express 前端，而不是这份 Go embed 产物。因此“本机能打开”不等于公网已迁移。
- 旧源码参考目录在另一工作区且存在未提交生成物，不能作为可发布版本；本表把它作为 V88 路由/API 语义参考，公网 DOM 才是视觉事实源。
- 观察到的“正在加载”是实际采样时的状态，不推断其最终成功态。`/agent/canvas` 被浏览器内容拦截（`ERR_BLOCKED_BY_CLIENT`），该页只能做源码/路由对照，不能声称完成视觉验收。

## 全局外壳

|项|公网 V88 观察|当前 main 对应|差异与迁移结论|
|---|---|---|---|
|导航|折叠式左侧导航：首页、剧本生成、小说获取、小说面板、水货生产、Agent 工作区、历史、问题日志、配音；顶栏有主题切换、个人中心、用户名和退出。|`前台/src/UserShell.jsx`、`前台/src/home.css`。|已有 React 外壳，但导航信息架构、图标、间距和移动端行为需逐项对照；不可把旧全局 CSS 无隔离复制。|
|会话与权限|Cookie 登录态决定用户名、项目与管理入口；未观察到前端直出凭据。|`前台/src/api.js`，Go `api/internal/httpapi/server.go` 的 Cookie Session、CSRF same-origin、capability、项目访问校验。|必须保留 Go 的 Cookie/CSRF/capability 边界；旧 Bearer/token 辅助逻辑不可迁入。|
|个人中心|点击个人中心后打开“个人中心”层，含关闭、折叠和退出动作。|`前台/src/AccountCenterPage.jsx`。|当前 `/profile`、`/member` 是独立页；需对齐为菜单/抽屉式体验或明确产品差异。|

## 路由矩阵

|路由|公网 DOM、布局、动作、状态|当前 main 对应文件|功能/视觉结论|旧 Node API → Go API 对应|不可直接搬运|
|---|---|---|---|---|---|
|`/`|视频 Hero（“让小说章节直接进入可视化剧本工作流”）、开始生成/进入配音/查看全部历史记录/打开设置；六个入口卡（新建剧本、小说面板、批量工厂、水货生产、Agent、配音）及“最近创作项目”列表和帮助 CTA。|`前台/src/HomePage.jsx`、`home.css`、`UserShell.jsx`。|需以真实最近项目/创作聚合驱动；不能以静态假卡片替代。视觉需对齐 Hero、入口密度和项目列表。|旧 `platformProjects`、`history`、`auth` → 当前 `GET /api/v1/history`、`GET /api/v1/batch-projects`、当前用户 API。|旧 token、用户头像 URL、浏览器缓存的项目状态。|
|`/script`|三栏剧本台：原文输入（“粘贴小说原文，开始构思…”）、人物/场景计数及提取/添加/重提取、持续/爆点/分段、10s/15s、画布模式、撤销/导出；有生成剧本、复制、查找替换、配音原文和匹配音频等操作。采样未打开抽屉。|`前台/src/ScriptWorkspace.jsx`、`script-workspace.css`。|核心 UI 已有入口；需逐按钮核对真实后端链路、抽屉、空态与失败态，不能只比页面骨架。|旧 `generation`、`scriptVideo`、`tts`、`history`、`config` → Go 批量项目 generation、音频、提示词/设置端点（见 `server.go`）。|旧 `/api/chat`、浏览器临时草稿、前端 token。|
|`/novel-fetch`|采样为全屏工作台 loading：“小说获取 / 正在加载小说获取… / 正在准备工作台，请稍候”。有独立 `NovelFetchPage` CSS chunk。|`前台/src/App.jsx`（`IntakeWorkbench`）、`RouterApp.jsx`。|**功能与视觉缺口**：当前与批量工厂/水货共用同一页，未复刻独立小说获取工作台和其 loading/错误/恢复。|旧 `novelFetch`: `/api/novel-fetch`、`/process`、`/save`、上传会话；→ Go `POST/GET /api/v1/intakes`、`POST /api/v1/intakes/{id}/execute`、书目/工作台接口。|旧上传登录/会话、Web submit 配置及任何网站 Cookie。|
|`/novel-fetch-workshop`|“改文工作台”，处理/任务/配置分区；开始处理、刷新任务、返回；三处搜索/输入，平台、输入格式、字段顺序、截取字数和 AI 文案数量；任务表为空时 `No data`。|`前台/src/NovelFetchWorkshop.jsx`；`api.js` 的 intake/workshop 调用。|有专页，需以真实 intake/任务数据验收所有筛选、分页、失败重试。公网存在明确表格和空态。|旧 `novelFetchWorkshop` → `GET/PUT /api/v1/intakes/{id}/workshop`、intake/书目 API。|旧浏览器任务筛选缓存、Web 平台登录信息。|
|`/novel-panel`|采样为“小说面板 / 正在连接小说面板 / 工作台加载完成后会显示在这里”。|`前台/src/novel-panel/NovelPanelWorkbench.jsx`、`NovelPanelWorkbench.css`，`RouterApp.jsx`。|当前要求 `?projectId=`，缺参时仅 RouteFoundation；公网呈现连接工作台的体验。需从项目入口带入项目上下文。|旧 `novelPanel`、`tts`、`models` → `GET/PUT /api/v1/batch-projects/{id}/novel-panel`、history/restore 端点。|旧 `getToken()`、模型密钥、媒体 data URL。|
|`/batch-factory`|“Batch Factory / Unified batch, single book configs, storyboard production/video submission”，创作漫剧、刷新、新建批量；采样状态“正在读取批量工程”。有独立 workbench CSS。|`前台/src/App.jsx`（错误地复用 `IntakeWorkbench`）；另有未路由的 `BatchFactoryHome.jsx`、`BatchProjectListPage.jsx`、`BatchProjectVideo.jsx`。|**重大缺口**：公网批量工程 UI 未由该路由承载；现有组件与路由未接通。|旧 `batchFactoryV11`（`/api/batch-factory/v12/*`）→ Go `/api/v1/batch-projects`、books、generation、video、merge、settings、publishing。|V88 provider 凭据、发布 credential、自动化调度与本地执行器 pairing secret。|
|`/shuihuo-production`|作品库而非 intake：标题“漫剧解说”，创作漫剧/批量工厂，个人作品范围、搜索、合集筛选、时间排序、网格视图；真实作品卡有打开、打开工作台、删除。|`RouterApp.jsx` 把它指向 `App.jsx`；另有 `ShuihuoProductionPage.jsx`、`shuihuo-production.css`、`shuihuo-media.css`，但未被路由使用。|**重大功能和视觉缺口**：当前路由没有 V88 作品库、筛选、卡片操作或媒体任务台。|旧 `shuihuoProduction`、`batchFactoryV11`、`localExecutors` → Go batch project、`/shuihuo/*`、video/local-executor API。|用户项目封面、素材 URL、TOS/本地执行器凭据、旧队列状态。|
|`/agent`|独立 `AgentPageV2` 样式；采样为“Agent 工作区 / 正在加载工作台”。|`前台/src/AgentStudioPage.jsx`、`agent-studio.css`，`api.js` agent 项目/消息/执行 API。|有新专页；需与公网项目列表、加载、动作、空态逐项验收。|旧 `/api/agent/tasks*`、`/api/agent/skills` → Go `/api/v1/agent/projects*`、skills、executions。|旧 `/api/agent/chat`（禁止迁入）、模型/系统提示词、草稿缓存。|
|`/agent/canvas`|浏览器被内容拦截，未取得可靠公网 DOM。|`前台/src/AgentCanvasPage.jsx`、`api.js` canvas API。|不可标为视觉完成；应在可控浏览器重新验收画布加载、保存、版本恢复及授权错误。|旧 `/api/agent/tasks/{id}/canvas` → Go `GET/PUT /api/v1/agent/projects/{id}/canvas`、版本恢复。|画布 localStorage 草稿、旧 task ID/token。|
|`/tts`|配音卡片工作台：添加卡片、保存为默认配音、上传小说、全部生成；声音“晓晓（女声·温柔）”、通用、语速/音调；空态“还没有配音卡片”。|`前台/src/TtsPage.jsx`、`tts-history.css`。|已有页；需验证卡片和任务事实源、上传/生成失败态，而非仅静态空态。|旧 `tts` → Go TTS/媒体任务相关 API（应以当前 `api.js` 调用为准）。|语音 provider 密钥、上传原文、浏览器已选声音偏好。|
|`/history`|“项目与生成历史”，清空剧本记录；表头来源/项目或记录/说明/最后更新/操作；空态 `No data`。|`前台/src/HistoryPage.jsx`、`tts-history.css`。|Go 历史投影已接入，仍需把 V88 的清理语义、过滤、跳转和空态对齐，且不得创建第二份历史事实。|旧 `history`、`platformProjects` → `GET /api/v1/history`（从项目/书目/生成事实投影）。|旧 localStorage 历史、跨账号记录、原始异常详情。|
|`/issues`|采样为“问题日志 / 正在加载工作台”，无独立 route CSS chunk。|`前台/src/IssuesPage.jsx`、`api.js`，Go `issues_handlers.go`。|已有 MySQL 事实投影入口；还需对齐 V88 loading/筛选/刷新、脱敏信息和分页操作。|旧错误日志、小说面板诊断、改文问题 API → `GET /api/v1/issues?page=…`（Run/StageRun/媒体事实投影）。|堆栈、provider 响应体、token、用户未授权项目错误。|
|`/settings`|采样为“正在加载工作台”；外壳仍是用户站。|`前台/src/SettingsPage.jsx`、`api.js`，Go workspace settings handler。|已有服务端偏好入口；需对齐主题、提醒、存储、设备/执行器“已选/实际生效”及 unavailable 表述。|旧 `config`、local executor、模型目录 → `GET/PUT /api/v1/workspace/settings`、video provider/executor status。|密钥、provider 原始配置、旧 localStorage 设置。|
|`/member`|实际打开“个人中心”层，含关闭、折叠资料菜单、退出；未将 MFA/恢复码呈现为可抄录文本。|`前台/src/AccountCenterPage.jsx`（`mode="member"`）。|当前独立路由与公网菜单层不一致；成员/团队信息需由 capability 限制。|旧 `member`、`teamAdmin` → current-user/profile/security/team API（Go 对应实现）。|Session、Cookie、MFA secret/恢复码、成员重置密码参数。|
|`/profile`|与 `/member` 同样打开“个人中心”层。|`AccountCenterPage.jsx`。|同上；当前可分别路由，但必须避免把安全材料回显给前端。|旧 `member`、`accountRecovery` → 个人资料/安全 API。|密码、MFA、恢复码、登录 token。|
|`/admin`|独立管理端：标题“一战晟铭管理端”；卡片入口提示词库、Prompt 策略、CM 平台技能、水货生产模型、错误日志；有管理/查看操作。|`RouterApp.jsx` 无 `/admin` 分支，当前落到 404；有无对应前台管理组件需另行实施。|**明确未迁移**：不能把 V88 管理页当作普通用户页搬入；需要 operations-admin capability、独立 API 和审计。|旧 `admin`、`accountAdmin`、`modelCatalog` → 当前仅诊断管理能力可见；完整 admin API/UI 尚无等价路由。|账号管理、重置密码、prompt 内容、provider/model 凭据、管理 Bearer token。|

## 路由/API 迁移判断规则

1. “对应”表示可用 Go 模型/端点可以承接同一业务事实，不表示请求路径、字段或 UI 已经等价。
2. Go 的写请求必须继续使用 Cookie Session、same-origin CSRF、防越权 capability 和项目归属校验；不得用旧 Node Bearer token 兼容替换。
3. 上传和真实媒体仅在现有 TOS 边界可用时写入；没有 Provider/执行器时应返回 `executor_unavailable` 或安全错误，不得伪造成功。
4. 管理端、跨团队操作、密钥和恢复材料均不属于普通用户页面可迁入内容。

## 当前优先级（审计结论，不是实施承诺）

1. 先拆开并接通 `/novel-fetch`、`/batch-factory`、`/shuihuo-production` 的路由与各自事实源；当前三者共用 `IntakeWorkbench` 是最大断层。
2. 把首页最近项目、历史、问题日志、设置和个人中心接到已存在的 Go 数据/权限边界后，再做 V88 视觉外壳的隔离式 CSS 迁移。
3. `/admin` 需要单独授权模型和产品范围；在此之前保持 404/拒绝访问优于伪造管理功能。
