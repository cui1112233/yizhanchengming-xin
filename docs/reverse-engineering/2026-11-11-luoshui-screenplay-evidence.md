# 洛水 2026「剧本创作」证据报告

日期：2026-11-11  
目标版本：洛水 v2.9.0（本机已登录桌面应用）  
目标目录：`F:\软件\洛水2026\洛水2026`

## 1. 分析边界与文件身份

本报告只记录本机安装包的静态证据、已登录页面的非破坏性交互，以及本地数据库结构。没有复制洛水源码、密钥、账户数据、私有接口或专有素材，也没有发起付费生成。

| 证据对象 | SHA-256 | 用途 |
| --- | --- | --- |
| `dist/assets/index-DszfAkKD.js` | `2E5C09685FF5BC10160480F963936B56BAC38A57A9E9B6BD199492F9437DF1E8` | 前端入口、界面文案、pywebview 调用点 |
| `modules/screenplay.pyd` | `EF7F7CE0F01CE9AD2F226C302B7C1E822D912C52CF5CE48C8EEE7B2FF1EF515E` | 剧本领域、本地文件和数据库能力 |
| `data/app.db` | `29DAF4634B964836FD4B9A5C357BD8C719358DEB2AFB8551A48E6340D570E32D` | 只读检查数据表结构；未抄录用户数据 |

REA 对前端包的分析证据号为 `ev_ff6783f59bb6161196a44249197b96b800f9f793a79ace4710a56fee92b6589c`。该次分析读取 5 个相关文件、24,084,486 字节、3 个 JavaScript 文件和 1,806,461 个 AST 节点；主包过大导致部分调用图受 1GB 内存上限影响。因此，本报告把“找到了什么”作为证据，不把“没有找到”说成不存在。

## 2. 用户入口与页面结构

### 2.1 入口

观察结论：入口位于左侧导航的“内容创作”分组中，菜单名为“剧本创作”。页面标题仍为“剧本创作”，副标题是“管理和创建您的剧本作品”。

证据：

- 前端包中的“剧本创作”：`dist/assets/index-DszfAkKD.js:29:1479088`（字符偏移 `1587526`）。
- 已登录页面截图：`C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-list.png`。

### 2.2 项目库

观察结论：项目库包含个人作品/团队作品切换、搜索、时间排序、网格视图、新版/旧版切换、历史入口和“创建剧本”按钮。空库显示“暂无项目，在上方输入框输入内容开始创作”。

证据：

- “个人作品”：`dist/assets/index-DszfAkKD.js:35:1479958`（偏移 `6243044`）。
- “团队作品”：`dist/assets/index-DszfAkKD.js:56:3443078`（偏移 `12819859`）。
- “搜索作品...”：`dist/assets/index-DszfAkKD.js:35:1480064`（偏移 `6243150`）。
- “按时间降序”：`dist/assets/index-DszfAkKD.js:29:4046797`（偏移 `4155235`）。
- “新版/旧版”：`dist/assets/index-DszfAkKD.js:29:2039642` 与 `:29:1566659`。
- “创建剧本”：`dist/assets/index-DszfAkKD.js:29:4127002`（偏移 `4235440`）。
- 页面截图：`C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-list.png`。

### 2.3 创建项目弹窗

观察结论：创建项目要求填写作品名称和上传源文件；可选择模型、画面风格、剧本比例、资产解析程度。源文件入口明确接受 `txt/docx`。默认比例为 9:16；解析程度可选“大师模式-3D / 大师模式-写实 / 差异化 / 标准”。

证据：

- “作品名称”：`dist/assets/index-DszfAkKD.js:29:4072351`（偏移 `4180789`）。
- “选择模型”：`dist/assets/index-DszfAkKD.js:29:1502174`（偏移 `1610612`）。
- “选择风格”：`dist/assets/index-DszfAkKD.js:29:3967269`（偏移 `4075707`）。
- “剧本比例”：`dist/assets/index-DszfAkKD.js:56:2750493`（偏移 `12127274`）。
- “资产解析程度”：`dist/assets/index-DszfAkKD.js:56:3455017`（偏移 `12831798`）。
- “上传剧本格式”：`dist/assets/index-DszfAkKD.js:56:3455553`（偏移 `12832334`）。
- “点击或拖动文件 (txt / docx)”：`dist/assets/index-DszfAkKD.js:56:3457596`（偏移 `12834377`）。
- 弹窗截图：`C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-create-dialog.png`。

## 3. 前端到本地后端的调用链

### 3.1 总链路

派生结论：洛水桌面端不是浏览器直接请求剧本服务，而是前端通过 `window.pywebview.api` 调 Python 本地接口；Python 模块再负责数据库、文件、模型与供应商网络调用。

证据：

1. 前端调用点 `get_screenplays`：`dist/assets/index-DszfAkKD.js:29:4123327`（偏移 `4231765`）。
2. 前端调用点 `create_screenplay`：`dist/assets/index-DszfAkKD.js:29:4124027`（偏移 `4232465`）。
3. 前端调用点 `save_screenplay_chapters`：`dist/assets/index-DszfAkKD.js:29:4124066`（偏移 `4232504`）。
4. 前端可见失败文案“剧本创作接口不可用”：`dist/assets/index-DszfAkKD.js:29:4124113`（偏移 `4232551`）。
5. 原生模块中存在 `ScreenplayAPI.delete_screenplay` 字符串：`modules/screenplay.pyd` 文件偏移 `54308979`。

可验证链路为：

`剧本创作菜单 -> 项目库 get_screenplays -> 创建弹窗 -> create_screenplay -> save_screenplay_chapters -> 导入/生成资产 -> 进入工作台`

前端包还明确包含一条导入进度链：

`准备剧本与分镜数据 -> 新建剧本创作项目 -> 写入分集剧本 -> 导入 JSON 分镜 -> 导入作品资产 -> 跳转到剧本创作`

这些步骤位于 `dist/assets/index-DszfAkKD.js:29`，字符偏移范围 `4232631..4232874`。

### 3.2 项目、章节和配置

观察结论：前端暴露了项目增删改查、章节加载/保存、AI 配置和显示配置等接口。

| 能力 | 关键接口 | 证据 |
| --- | --- | --- |
| 项目列表 | `get_screenplays` | `index-DszfAkKD.js:29:4123327` |
| 新建项目 | `create_screenplay` | `index-DszfAkKD.js:29:4124027` |
| 保存项目 | `save_screenplay` | `index-DszfAkKD.js:29:3860461` |
| 删除项目 | `delete_screenplay` | `index-DszfAkKD.js:50:234426` |
| 保存章节 | `save_screenplay_chapters` | `index-DszfAkKD.js:29:4124066` |
| 加载 AI 配置 | `load_screenplay_ai_config` | `index-DszfAkKD.js:50:275088` |
| 保存 AI 配置 | `save_screenplay_ai_config` | `index-DszfAkKD.js:29:3860461` |

### 3.3 资产与镜头

观察结论：剧本项目不是单一文本框，它至少管理角色、场景、道具、音色、镜头和分镜。

| 数据 | 关键接口 | 证据 |
| --- | --- | --- |
| 角色 | `create_screenplay_character` | `index-DszfAkKD.js:50:236072`（偏移 `7671792`） |
| 场景 | `create_screenplay_scene` | `index-DszfAkKD.js:50:236525`（偏移 `7672245`） |
| 道具 | `create_screenplay_prop` | `index-DszfAkKD.js:50:236974`（偏移 `7672694`） |
| 镜头 | `create_screenplay_shot` | `index-DszfAkKD.js:50:233691`（偏移 `7669411`） |
| 修改镜头 | `update_screenplay_shot` | `index-DszfAkKD.js:50:234195`（偏移 `7669915`） |
| 移动镜头 | `move_screenplay_shot` | `index-DszfAkKD.js:50:235084`（偏移 `7670804`） |
| 删除镜头 | `delete_screenplay_shot` | `index-DszfAkKD.js:50:234426`（偏移 `7670146`） |

### 3.4 AI 与批处理

观察结论：前端包含普通推理、大师推理、分镜推理、对话、Agent、章节计划、批量图片/视频、旁白和取消任务等能力。

| 能力 | 关键接口 | 证据 |
| --- | --- | --- |
| 剧本推理 | `screenplay_ai_inference` | `index-DszfAkKD.js:29:4197620`（偏移 `4306058`） |
| AI 对话 | `screenplay_ai_chat` | `index-DszfAkKD.js:56:1653712`（偏移 `11030493`） |
| Agent 执行 | `screenplay_agents_run` | `index-DszfAkKD.js:56:2435901`（偏移 `11812682`） |
| 取消推理 | `cancel_screenplay_ai_inference` | `index-DszfAkKD.js:29:4197613`（偏移 `4306051`） |
| 章节计划 | 章节计划接口族 | `index-DszfAkKD.js:56:2230539`（偏移 `11607320`） |
| 批量分镜图 | 分镜图片批处理接口族 | `index-DszfAkKD.js:50:254168`（偏移 `7689888`） |
| 即梦批量生成 | `screenplay_jimeng_generate_batch` | `index-DszfAkKD.js:50:252152`（偏移 `7687872`） |
| 旁白音频 | 旁白音频接口族 | `index-DszfAkKD.js:35:1316062`（偏移 `6079148`） |

### 3.5 导出与定时任务

观察结论：剧本工作流包含任务状态、取消任务和导出到剪映等收尾能力。

- 任务状态：`dist/assets/index-DszfAkKD.js:56:3376277`（偏移 `12753058`）。
- 取消定时任务：`dist/assets/index-DszfAkKD.js:56:3377783`（偏移 `12754564`）。
- 导出分镜到剪映：`dist/assets/index-DszfAkKD.js:35:1323636`（偏移 `6086722`）。
- `export_to_jianying`：`dist/assets/index-DszfAkKD.js:50:271755`（偏移 `7707475`）。

## 4. 数据变化与文件

### 4.1 本地数据库

只读 `PRAGMA table_info` 证明 `data/app.db` 至少存储以下结构：

- `projects`：项目名、描述、项目路径、字幕/输出/音频路径及时间戳。
- `screenplay_characters / screenplay_scenes / screenplay_props`：项目 ID、名称、描述、图片提示信息、参考图、排序和时间戳。
- `screenplay_shots`：台词、角色/场景/道具引用、图片/视频描述、媒体路径、时长、配音、状态、元数据和时间戳。
- `screenplay_voices`：角色、音色名、音色 ID 和音色设置。
- `storyboards`：字幕时间、字幕文本、媒体描述、图片/视频/音频路径、角色标签、状态和时长。

原生模块还包含建表和文件名字符串：

- `shots` 建表：`modules/screenplay.pyd` 偏移 `54323936`。
- `characters`：偏移 `54325324`。
- `scenes`：偏移 `54326227`。
- `props`：偏移 `54326919`。
- `st_chapters`：偏移 `55076582`。
- `screenplay.db`：偏移 `51911704`。
- `project.db`：偏移 `51849072`。
- `engine_config.json`：偏移 `51892702`。
- 定时任务 JSON：偏移 `53693857`。

### 4.2 取消操作

观察结论：打开“新建剧本”后点取消，会返回项目库，且不会写入项目数据库或剧本目录。

证据：

- 操作前后 `data/app.db` SHA-256 均为 `29DAF4634B964836FD4B9A5C357BD8C719358DEB2AFB8551A48E6340D570E32D`，修改时间未变化。
- 操作前后 `data/剧本创作` 文件数均为 0。
- 取消后截图：`C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-after-cancel.png`。

## 5. 网络请求边界

观察结论：剧本前端首先调用本机 pywebview 桥接接口，不直接把业务请求写成浏览器 `fetch`。原生模块中存在 OpenAI、Anthropic、阿里云百炼、火山引擎等兼容供应商地址，也存在洛水私有服务地址。

安全边界：独立重写只接一战晟铭现有、由用户配置的官方或兼容模型通道；不复制洛水私有域名、令牌、签名算法和账户接口。静态字符串只能证明模块包含候选地址，不能证明本次页面实际请求了哪个供应商。

运行时仅观察到 Python 进程监听本机 `127.0.0.1:59175`（pywebview 桥）和 `127.0.0.1:60624`（本地 Web 服务），以及若干无法归因到本功能的后台 TLS 连接。没有为了取证发起付费生成，所以模型请求体和响应体仍是未知项。

## 6. 已知限制

1. `screenplay.pyd` 的 REA/Ghidra 总览与字符串搜索在 330 秒后超时，因此原生函数内部伪代码、精确调用者和异常分支未被完整恢复。
2. 前端主包是 16,803,812 字节的压缩单文件，REA 的完整图受内存上限影响；本报告用精确行/列/字符偏移补足关键入口证据。
3. 没有创建真实项目、消耗模型额度或删除用户数据，所以真实生成返回结构、供应商错误正文和项目卡片的运行时状态需要在一战晟铭独立实现后用受控测试数据验证。
4. “全部功能”应以本报告列出的接口族和后续工作台逐项走查为清单，不能把存在字符串直接等同于已经运行验证。

## 7. 对一战晟铭的直接结论

1. “漫剧创作”必须是独立模块，不改 `/script` 和“私人定制”。本项目现状证据：`frontend/src/shared/layouts/UserLayout.jsx:30` 把 `/script` 明确定义为“私人定制”，`frontend/src/user/App.jsx:71` 把 `/script` 映射到 `ScriptPage`。
2. 本项目已有可复用但尚未接入 HTTP/前端的剧本底座：`backend/internal/screenplay/model.go:77-323` 定义项目、文档、镜头、角色、场景、道具、音色、批次与版本；`backend/internal/screenplay/screenplay_service.go:60-183` 已有生成、解析、持久化和错误处理；`backend/internal/storage/screenplay_schema.go:16` 起定义 MySQL 表。
3. 系统提示词已经位于后端：`backend/internal/screenplay/screenplay_service.go:72-75` 通过后端注册表读取 `a1_novel_to_screenplay`，符合“前端不硬编码提示词”的项目规则。
4. 当前 `backend/internal/httpapi/router.go` 没有检出 screenplay 路由，所以现有能力仍不是用户可用产品。重写必须补齐 HTTP、权限、取消、任务状态和前端工作台，并通过真实 API 验收，不能只放按钮。

## 8. 按钮级对照与实现状态

按钮、输入、输出、边界、失败提示及一战晟铭当前真实实现状态，见同目录的
`2026-11-11-luoshui-screenplay-button-parity.md`。该表明确区分“已打通”“部分打通”“未实现”，图标或空弹窗不计为功能完成。
