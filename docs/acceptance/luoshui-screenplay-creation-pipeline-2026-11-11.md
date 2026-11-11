# 洛水 v2.9.0「剧本创作」链路说明（普通用户版）

> 读法：每一节先给结论，再展开。技术细节放在括号、表格和附录里，不影响正文理解。
>
> 取证方式：只读。读取了洛水安装目录（`F:\软件\洛水2026\洛水2026`）、本机 `data` 数据、前端压缩包 `recovered\frontend\index-DszfAkKD.js`、`recovered\catalogs` 下的模块字符串与接口清单、`spec` 下的既有对照文档。**未启动洛水、未改任何文件、未调用任何模型或生成接口、未重启任何服务。**
>
> 偏移说明：文中「前端 @数字」指前端压缩包解压后的**字符偏移**（按文本读取计数），用于定位证据，不是字节偏移。

---

## 0. 一页速览

| 问题 | 结论 |
|---|---|
| 它是干什么的 | 把小说原文改写成**分集剧本**，再逐集生成**分镜词/分镜 JSON**，最后接生图、生视频、配音、剪映导出 |
| 输入是什么 | 小说原文（粘贴，或导入 `.txt` / `.docx`）；界面只要求文本非空，技术上也接受大纲或已有剧本文本，但提示词和分集规则是按小说原文设计的 |
| 最终交出来什么 | 每集一份剧本 + 每集一份分镜 + 分镜对应的图片/视频/音频文件；可导出剪映草稿。**不是直接一键出成片**，成片需要后段另行合成 |
| 入口在哪 | 左侧菜单「内容创作 → 剧本创作」；项目列表页再点「创建剧本」 |
| 谁在干活 | 文本靠大模型（HTTP），图片/视频/配音靠各家引擎（HTTP 或本机 CLI），Agent 模式靠本机 `agent.pyd` 里的技能包和 8 步批次编排 |
| 东西存在哪 | 每个作品一个文件夹，核心是 `screenplay.db`、`storyboard_table.db`、`chapters.json`、`agent_assets.json`，外加 `images/`、`videos/`、`audio/`、`assets/`、`.luoshui/` |

---

## 1. 它是干什么的

**一句话：它是一条「小说 → 分集剧本 → 分镜词」的流水线，后面顺手把分镜变成图、视频和配音。**

- **解决什么问题**：网文作者手里通常只有几万到几十万字的小说，要拍短剧/漫剧，必须先拆成分集、写成可拍的剧本，再拆成一个个镜头。这个模块把这段最耗人力的改写工作交给 AI，过程中保留人工审核与修改。
- **输入**：小说原文。可以在新建弹窗里直接粘贴，也可以点导入选择 `.txt` / `.docx`（前端 @39255397 选择文件、@39256995 解析文档；浏览器兜底只接受 `.txt,.docx`）。
- **中间产物**：分集剧本（每集正文）、角色/场景/道具资产、每集分镜。
- **最终交付**：每集剧本 + 每集分镜（可生成为图片/视频/配音文件）+ 可导出的剪映草稿。是否合成一条完整成片，取决于用户在后段是否点「合并成片」。
- **它不是什么**：
  - 不是「丢进去就出成片」的黑盒。模型、提示词、画幅、资产解析档位都要人选。
  - 不是纯本地功能。打开剧本项目时先过账号权限闸门，没开通会直接拦住（前端 @39266200 起，提示「该账号无权限，请联系小洛微信：bylx109开通」）。
  - 不是只有一套编辑界面。它同时存在「旧编辑器」和「新工作台」两种形态，项目列表会按 `editor_version` 过滤（前端 @39313981）。

---

## 2. 从进到出的完整链路

### 2.1 入口：左侧菜单，不是下拉

- 左侧导航分组「内容创作」下有「剧本创作」这一项（前端 @49630446）。
- 首页是否另有快捷卡片：本次**未取到证据**，不写推测。
- 点进去是项目库页面：标题「剧本创作」，副标题「管理和创建您的剧本作品」（@39450022），上方「个人作品 / 团队作品」两个页签（@39455960 / @39456176），右上「创建剧本」按钮（@39454848）。
- 项目库支持网格/列表两种视图、按时间/名称排序、搜索作品、右键菜单（打开作品 / 复制 / 导出作品 / 删除作品，运行中还会出现「取消挂机任务」）（前端 @39254093、@39305545、@15306085 附近）。

### 2.2 新建弹窗：填什么、选什么

点「创建剧本」后弹出「新建剧本」弹窗（前端 @39555000 起）。字段如下：

| 字段 | 说明 | 默认值 | 证据 |
|---|---|---|---|
| 作品名称 | 必填，空时提示「请输入作品名称」 | 空 | 前端 @39265676 |
| 选择模型 | 文本大模型下拉，来自本机已配置的 API 模型列表 | 列表第一项 | 前端 @39560000 附近；本机 `settings.json` 的 `apiModels` 共 113 条 |
| 选择提示词 | 剧本提示词下拉，占位文案「请选择剧本大师提示词」 | 列表第一项 | 前端 @39564389；提示词来自 `app.db.prompts` |
| 选择风格 | 画风/风格参考，弹窗选择「点击选择画风」 | 空 | 前端 @39585015 |
| 剧本比例 | 画幅比例，默认 9:16 | 9:16 | 前端 @39252613 |
| 原文内容 | 粘贴区或导入文件；导入后自动填作品名（去掉 `.txt/.docx` 后缀） | 空 | 前端 @39255397 |
| 资产解析程度 | 四档：大师模式-3D / 大师模式-写实 / 差异化 / 标准 | 大师模式-3D | 前端 @39252826 起 `Te` 数组；标签在 @39588705 |
| 剧本大师 | 开关，决定剧本文本走「剧本大师」提示词体系 | 开 | 前端 @39252500 起 `scriptMaster: !0` |
| 工作模式 | `pro`（专业模式）；剧本创作会过滤掉 `simple` | pro | 前端 @39252500 起 `workspaceMode: "pro"` |

补充说明：

- 长篇/短篇不是手选项，是自动判断：文本里能匹配到「第N章/回/节/篇/卷/部/集」或 `chapter N` / `episode N`，**且**正文超过 3 万字，判为长篇；否则判为短篇（前端 @39251300 附近 `je()` 正则与 `Fe()` 判断）。
- 弹窗里有一行提示：「建议小说原文生成剧前往 Agent / Chat 功能，不清楚？查看链接」（前端 @39565750）。这句话很关键：**官方自己也认为长篇小说更适合走 Agent 对话模式**，而不是弹窗里的直接转换。

### 2.3 点「创建」之后发生什么

1. 前端调用 `create_screenplay`，参数包含：名称、原文、`script_mode`、`workspace_mode`、`editor_version`、风格、比例、`content_type`、`asset_parse_degree`、`script_master`（前端 @39259885 起）。
2. 后端建立项目文件夹，返回 `id` 与 `path`（前端 @39260300 附近读取返回值）；项目数据库在首次打开/初始化时建立。
3. 前端自动分集：按章节正则切分，生成「第 1 集 / 第 2 集…」列表，每集带 `title`、`chapterName`、`content`、`scriptResult: null`，然后调用 `save_screenplay_chapters` 落盘（前端 @39263500–@39264000）。
4. 进入工作台：默认停在「剧本」页签、右侧为「AI 对话」模式（前端 @39264790 设置 `R.value = "story"`、@39264802 设置 `_.value = "ai-chat"`）。
5. 工作台按内容类型自动装载不同的 Agent 步骤序列（前端 @39264463）：
   - 长篇：`text_input → synopsis → character_parse → scene_parse → prop_parse`
   - 短篇：`text_input → line_numbering → synopsis → character_parse → scene_parse → prop_parse`
   - 其中 `line_numbering` 在界面上的名字是「分镜文件处理」，作用是给原文每行加序号（`1.xxx`），后续 Agent 都用带序号的版本（前端 @33299848）。

### 2.4 工作台的三个主页面

工作台只有三个主页面：**剧本 / 资产 / 分镜**（前端 @33913780、@33913821、@33913863），顶部用一条步骤条串起来，每一步有对应的完成状态标签，作为能否进入下一步的判断依据：

| 页签 | 状态文案 | 证据 |
|---|---|---|
| 剧本 | 剧本内容整理完毕 | 前端 @33914254 |
| 资产 | 资产设定已确认 | 前端 @33914284 |
| 分镜 | 分镜内容已确认 | 前端 @33914317 |

右侧固定是 AI 对话/工具区，模型会在里面调用工具（导演审核、台词检查切分、查找对比、修改角色/场景/道具、资产批次推理/完整检查等，前端 @33899000–@33899670；分镜序号校验 @33594447）。这就是「界面直调」和「Agent 执行」两条路径在同一个页面里并存的形态。

### 2.5 剧本这一步

用户看到/要做的事：

- 左侧是分集列表，可以调整分集顺序（前端 @33915977「已调整分集顺序」）。
- 中间是当前集剧本正文/编辑区，可手动改；编辑原文有 1 秒防抖自动保存（`save_screenplay`，前端 @39271575），分集另走 `save_screenplay_chapters`（创建时 @39263963、工作台 @33626601）。
- 右侧可选模型、选提示词、让 AI 生成或修改；顶部有「开始转换」按钮（前端 @39346686），点击后走 `screenplay_ai_inference`，按 `batch_size = 15` 分段处理（前端 @39288890）。
- 生成过程中会实时回报「已完成第 X/Y 批」「已完成分镜 A–B」等进度（前端 `onScreenplayAIBatch` / `onScreenplayAIProgress` 回调，前端 @39286691、@39296522）。

点完按钮后：

- 后端把原文 + 解密后的提示词 + 模型名送进大模型，分段返回，前端合并成 `characters / scenes / props / storyboards / project_info` 五块结构化结果。
- 结果先落到「AI 结果历史」，再把结构化数据写进项目数据库；若发现有角色/场景/道具缺少描述，会自动触发「补全 Agent」（`run_completion_agent`，前端 @39292344）。
- 如果已有旧数据，会先检查数据库是否存在，存在就直接进旧编辑器（`check_screenplay_db_exists`，前端 @39267478 附近）。

### 2.6 资产这一步

- 资产分三组：**角色、场景、道具**（另有音色，属于角色资产的一部分）。
- 每组可以「批次推理」「质量检查」「完整检查」（前端 @33899000 附近：角色批次推理 / 场景批次推理 / 道具批次推理、资产质量检查、全部资产完整检查）。
- 资产解析程度决定用哪套模板：大师-3D、大师-写实、差异化、标准，共 4 档（前端 @39252826 起；模块侧见 `screenplay.pyd @0x3338BB4` 的四档归一化）。
- 资产库是**项目级**的，不按集拆分；同一角色在第 1 集和第 10 集共用一张设定和一张参考图。

### 2.7 分镜这一步

用户能看到两种按钮/状态：

- 先「规划」，再「生成分镜」；生成时可以并发生成、顺序生成、自动补充（前端 @33733500 附近：「正在生成分镜规划…」「并发生成中」「顺序生成中」「自动补充中」「补充」「生成分镜」「分镜生成完成」）。
- Agent 侧还有三条现成快捷指令：「先写本集分镜规划」「按规划生成分镜词」「继续生成未完成的分镜」；指令里明确要求先读规划体例、提示词、本集剧本、资产登记名，再 `write_file` 写 plan、`append_storyboard` 每次追加 1–2 节（前端 @6760215、@6760535、@6760785）。
- 也可以按集勾选，批量生成分镜脚本（前端 @33925842「请先勾选要生成的剧集」、@33926027「请先勾选要批量制作的剧集」、@33966074「该集还没有分镜脚本，请先点击「分镜脚本」生成」）。

点完之后：

- 每集生成一份分镜 JSON，固定六字段契约：`id / scene_id / duration_sec / characters / scene / props / video_desc`（本机既有对照文档 `spec\漫剧剧本与分镜-需求规格-评审版.md` 第 21 行）。
- 分镜同时写入新版分镜表数据库 `storyboard_table.db`，用 `chapter_index` 区分第几集（模块字符串：`每个作品项目一个独立库文件：{project_path}/storyboard_table.db`）。
- 写入前会做一次账号权限校验；无权限直接报错，不进入推理。

### 2.8 图片 / 视频 / 配音 / 导出

分镜页签里再分媒体子页：**生图 / 视频**（前端 @14955000 附近「生图」「视频」「批量操作」「范围」「起始」「结束」），配音入口存在但设置弹窗是占位（见第 7 节）。

- 生图：按分镜范围批量生成，可用角色/场景/道具的绑定图做垫图（模块字符串：`剧本创作 - 即梦官网生图服务…自动收集分镜中的角色、场景、道具绑定图片作为垫图`；另有 4/6/9/25 宫格生图）。
- 视频：按分镜范围批量生成；可选本机小云雀 CLI（`pippit-tool-cli generate-video`、`query-result`）、Grok 视频（`grok-videos`）、即梦 CLI 等（模块字符串）。本机当前选的是 `grok_video_luoshui`。
- 配音：按分镜音色绑定批量合成（桥接 `tts_batch_start`、`add_voice_to_storyboard`）。
- 导出：分镜编辑器里可直接「导出剪映」（前端 @14704847 `export_storyboard_to_jianying`；旧编辑器路径是 `export_to_jianying`，前端 @19776819）。
- 合并成片：后段是「合并分镜 → 分段生成 → 合并成片」三段流程（前端 @26351600 附近）；长分镜先按目标秒数合并，再分段生成，最后合成成片。合成本身调用本机 FFmpeg（模块字符串多处 `ffmpeg` / `未找到 FFmpeg`）。

---

## 3. 谁在干活

### 3.1 两条不同路径

**路径 A：界面直接调（专业模式 / 弹窗直转）**

- 前端点按钮 → `window.pywebview.api.screenplay_*` → `api.pyd`（桥接层）→ `screenplay.pyd`（提示词派发、分段、解析、写库）。
- 特点：一条龙、参数明确、进度条可见；适合短篇或已有明确提示词的场景。
- 典型调用：`screenplay_ai_inference`（前端 @39288890）、`screenplay_master_inference`（@33673306）、`screenplay_generate_chapter_plan`（@33854572）、`screenplay_master_storyboard_inference`（生成分镜，@33736362）。

**路径 B：转给 Agent 引擎执行（Agent / Chat 模式）**

- 前端调 `screenplay_agents_run`（前端 @34100861），把 `project_path`、`content`、`model`、`pipeline`、角色/场景设计提示词、资产解析档位等交给 `agent.pyd`。
- Agent 在自己的工作区里用工具读写文件：`read_file` / `write_file` / `read_plan_guide` / `read_storyboard_prompt` / `read_assets` / `append_storyboard` / `patch_storyboard` 等（工具名见 `screenplay.pyd` 工具表；前端快捷指令里也直接点名 `read_assets` / `write_file` / `append_storyboard`，@6760280、@6760592）。
- 流水线型任务（生图、生视频、角色主体等）再通过 `delegate_pipeline` 委派回 `ScreenplayAIDispatcher`（`screenplay.pyd`），所以 Agent 不是绕过引擎，而是「先决定做什么，再调用同一套引擎」。
- 特点：适合长篇小说、需要多轮修改、需要医生审核和批次编排的场景。

### 3.2 三层模块

| 层 | 文件 | 大小（本机） | 职责 |
|---|---|---|---|
| 桥接层 | `modules\api.pyd` | 18,226,688 B | 暴露 `window.pywebview.api.*`，参数转换、回调、合并日志 |
| 业务层 | `modules\screenplay.pyd` | 55,643,648 B | 提示词派发、分段推理、资产解析、生图/生视频/配音/字幕/导出 |
| Agent 层 | `modules\agent.pyd` | 6,875,648 B | 技能包、工具循环、8 步批次编排、状态机 |
| 技能执行 | `modules\screenplay_skill_agent.pyd` | 3,089,408 B | 剧本类技能执行 |
| 字幕 | `modules\screenplay_subtitle_asr.pyd` | 501,248 B | 剧本分镜音频 + ASR + 原文对轴 |
| 粗剪/切分 | `screenplay_rough_cut.pyd` / `screenplay_video_cut.pyd` | 596,992 / 403,968 B | 分镜视频粗剪、切分 |

### 3.3 逐步对照：谁在干活、走本地还是 HTTP

| 环节 | 谁执行 | 本地进程 / HTTP | 具体模型或接口 | 证据 |
|---|---|---|---|---|
| 新建/保存/分集 | `api.pyd` → `screenplay.pyd` | 本地 | 项目 SQLite | 前端 @39259885、@39263963 |
| 剧本文本生成 | 文本大模型 | HTTP | 用户下拉选的模型；本机 `apiModels` 113 条 | 前端 @39560000；`official_llm_router` 接口清单 |
| 长篇自动分集/分段 | `screenplay.pyd` | 本地 | `batch_size=15` | 前端 @39288890 |
| 资产提取（角色/场景/道具） | `screenplay.pyd` 模板派发 | HTTP | 4 档资产模板 | `screenplay.pyd @0x3339764`、`@0x3338BB4` |
| 资产补全 | `screenplay.pyd`（`run_completion_agent`） | HTTP | 同文本模型 | 前端 @39292344 |
| 分镜规划 | `screenplay.pyd` | HTTP | 章节规划提示词 | 前端 @33854572；Agent step5 提示词 |
| 分镜词生成 | `screenplay.pyd` / `agent.pyd` | HTTP | 小洛 v7-15s、小云雀 v3 等 | 既有对照文档第 20、46–47 行 |
| 剧本/分镜审核（医生） | `agent.pyd` 技能 | HTTP | `luoshui-screenplay-doctor-v1`、`luoshui-preset-doctor-v1`、`luoshui-storyboard-doctor-v1` | 模块字符串与既有文档 |
| 生图 | 各家图片引擎 | HTTP 或本机 CLI | 即梦官网、宫格 Nano2/Gemini 3 Pro Image、万相 3.0、Pippit CLI、Lovart、小云雀等 | 模块字符串；本机当前 `nano2`（来源 `kongyang`） |
| 生视频 | 各家视频引擎 | 多为 HTTP，小云雀为本地 CLI | Grok 视频（`grok-videos`，模块说明基于空氧 API）、小云雀 CLI（`pippit-tool-cli generate-video`）、Seedance 2.0、RunningHub MiniMax-H3、GRSAI MiniMax-H3、AutoDL ComfyUI、Lovart、即梦 CLI、LibTV、Higgsfield 等 | 模块字符串；本机当前 `grok_video_luoshui`（来源 `luoshui`） |
| 配音 | TTS 引擎 | HTTP | MiniMax / Qwen / CosyVoice | 桥接 `tts_batch_start`；本机 `current_audio_engine = minimax` |
| 字幕/对轴 | `screenplay_subtitle_asr.pyd` | 本地 + 可能云端 ASR | 剧本分镜音频 + ASR + 原文对轴 | 模块字符串 |
| 合并成片 | 本机 FFmpeg | 本地 | 无外部模型 | 模块字符串 `ffmpeg`；桥接 `merge_*` |
| 导出剪映 | `jianying_draft_io.pyd` | 本地写草稿 | 剪映草稿目录（设置项 `jianyingExportPath`，本机为空） | 前端 @14704847、@19776819；`settings.json` |

### 3.4 文本模型实际走哪里

`official_llm_router` 里能取到的文本接口地址包括：

- 官方厂商：`https://api.deepseek.com`、`https://api.moonshot.cn/v1`、`https://ark.cn-beijing.volces.com/api/v3`、`https://dashscope.aliyuncs.com/compatible-mode/v1`、`https://open.bigmodel.cn/api/paas/v4`、`https://openrouter.ai/api/v1`、`https://api.xiaomimimo.com/v1`
- 洛水中转：`https://www.xiaoyumanman.com/v1`、`https://www.xiaoyumanman.com/v1/messages`、`https://xiaoluo.site/v1`、`https://xiaoluo.site/v1/messages`
- 聚合：`https://api.aibh.site/v1`

一个具体项目到底用了哪个模型，取决于项目里的 `ai_config.json`（模块字符串：`从项目ai_config.json读取模型`）。**本机 `data\剧本创作` 是空的，没有真实项目，因此「某个剧本项目实际用了哪个模型」未取到。** 上文的本机当前引擎（nano2 / grok_video_luoshui / minimax）只能说明「这台机器默认会选什么」，不能说明某个作品用了什么。

### 3.5 Agent 模式的 8 步（0–8 共 9 个编号）

批次编排把流程拆成 9 个键，提示词里写「step N/8」，实际是 0–8 编号（既有对照文档第 17 行）：

| 编号 | 步键 | 标签 | 产物 |
|---|---|---|---|
| 0 | `0_ingest` | 整理原文 | `.luoshui/context/novel.md` |
| 1 | `1_screenplay` | 生成剧本 | `.luoshui/context/screenplay.md` → 自动同步 `screenplay.db` |
| 2 | `2_screenplay_doctor` | 剧本审核 | `screenplay_doctor_report.md` + `.luoshui/batch/reviews/screenplay_review.json` |
| 3 | `3_assets` | 资产提取 | `character_parse_result.json`、`scene_parse_result.json`、`prop_parse_result.json` → 合并 `character_tags.json` |
| 4 | `4_preset_doctor` | 预设质检 | `preset_doctor_report.md` + `.luoshui/batch/reviews/preset_review.json` |
| 5 | `5_storyboard` | 按集分镜 | 每集分镜 JSON + 规划 |
| 6 | `6_storyboard_doctor` | 分镜审核 | `storyboard_review_ep{N}.json` |
| 7 | `7_video` | 批量出视频 | `assets/dreamina_batch` 等 |
| 8 | `8_finalize` | 收尾归档 | `assets/分镜_01.mp4…` + `finalize_report.md` |

---

## 4. 生成出来的东西存在哪

### 4.1 根目录

- 数据根：本机 `settings.json` 的 `dataPath = F:\软件\洛水2026\洛水2026\data`。
- 剧本创作根：`data\剧本创作`（本机存在，但**是空的**，0 个文件）。
- 模块文档给出的项目文件夹样例是 `E:\外部作品\剧本创作\`（`初始化项目数据库 :param project_path: 项目文件夹路径，如 E:\外部作品\剧本创作\`）。这说明作品根不强制在安装目录下；本机存在空目录 `data\剧本创作`，**它是否为新项目默认根，本次未取到配置项、未验证**。
- 每个作品文件夹下有自己的数据库：`{项目路径}/screenplay.db`（模块字符串原文）。

### 4.2 一个作品的目录树（推导自模块字符串与前端，逐项标注来源）

```text
<作品根>\<作品名>\
├─ screenplay.db                     # 旧版/兼容版项目库，5 张表（模块：项目级SQLite数据库）
├─ storyboard_table.db               # 新版分镜表库，按 chapter_index 区分集（模块原文）
├─ chapters.json                     # 分集列表：标题、原文、剧本结果（模块/前端）
├─ agent_assets.json                 # 人物/场景/道具/音色 的唯一真源（模块：加载/保存/增量合并）
├─ character_tags.json               # 三份资产解析结果合并后的标签库（模块：批次 step3 收尾）
├─ ai_config.json                    # 模型、promptId、风格、比例、剧本模式（模块：从项目ai_config.json读取模型）
├─ workflow.json                     # 项目级 UI 布局/工作流（模块：保存项目级UI布局到workflow.json）
├─ positioning.json                  # 多人站位卡（模块：站位卡永久写入 agent_assets.json）
├─ chapter_{N}_storyboard.json       # 每集分镜 JSON（模块 step6 用 chapter_+_storyboard.json；前端更新日志也有同名）
├─ input\                            # Agent 模式上传的原文（模块/前端「上传文件到 input/」）
├─ images\                           # 分镜图、资产图（模块：下载 image artifact → images/）
├─ videos\                           # 分镜视频（模块：输出到 images/videos）
├─ audio\                            # 配音、合并音频（模块：查找 audio/combined_audio.wav）
├─ assets\                           # Agent 收尾归档（模块：assets/分镜_01.mp4）
└─ .luoshui\                         # Agent 工作区元数据（本机 q 工作区已实机确认该目录存在）
   ├─ workspace.json                 # 工作区名/版本/启用技能/阶段（实机样本 85 B）
   ├─ context\
   │  ├─ novel.md                    # 整理后的原文
   │  ├─ screenplay.md               # 剧本文本（可自动同步到 screenplay.db）
   │  └─ finalize_report.md          # 收尾报告
   ├─ batch\
   │  ├─ state.json                  # 批次状态机（9 步骨架）
   │  ├─ batch.log                   # 批次日志
   │  └─ reviews\                    # 各医生审核 JSON
   ├─ checkpoints\<skill_id>\state.json  # 技能断点续接
   ├─ sessions\                      # 会话
   ├─ skills\                        # 工作区技能
   └─ runs\                          # 运行记录
```

说明：

- `.luoshui` 的目录骨架在本机 `F:\软件\洛水2026\洛水2026\q` 工作区实机可见：`.luoshui/` 下有 `context/`、`runs/`、`sessions/`、`skills/`、`workspace.json`（均为空目录/空壳）；同级的 `input/`、`assets/` 在工作区根目录。`batch/`、`checkpoints/` 本机尚未生成，结构来自模块字符串。
- `chapter_{N}_plan` 的具体文件名**未取到**：Agent 工具只约定 `target=plan`，磁盘文件名在 8 步提示词里没有写死。
- `images/`、`videos/`、`audio/` 的具体子目录层级**未取到**：模块字符串只给出「输出到 images/videos」「audio/combined_audio.wav」这类片段。

### 4.3 数据库表结构（旧版 `screenplay.db`）

本机 `app.db` 中保留了同构的 5 张剧本表（`screenplay_shots`、`screenplay_characters`、`screenplay_scenes`、`screenplay_props`、`screenplay_voices`），字段如下：

| 表 | 关键字段 | 说明 |
|---|---|---|
| `screenplay_shots` | `id / screenplay_id / sort_order / script_text / character_ids / scene_id / prop_ids / image_prompt / video_prompt / negative_prompt / image_path / image_collection / video_path / video_cover / duration / voice_text / voice_path / status / metadata / location_reference` | 一个镜头一行；资产用 id 外键引用，不是文本复制 |
| `screenplay_characters` | `id / screenplay_id / name / description / image_prompt / reference_image / sort_order` | 角色 |
| `screenplay_scenes` | `id / screenplay_id / name / description / image_prompt / reference_image / sort_order` | 场景 |
| `screenplay_props` | `id / screenplay_id / name / description / image_prompt / reference_image / sort_order` | 道具 |
| `screenplay_voices` | `id / screenplay_id / character_id / voice_name / voice_id / voice_settings` | 角色音色 |

新版分镜表 `storyboard_table.db` 的设计原则（模块原文）：

- 不修改 `screenplay.db`，不挂旧版合并链路；
- 推理 JSON 与 `agent_assets.json` 仍是「来源层」，`storyboard_table.db` 是「工作层」；
- 全局资产变更时只刷新 DB 中的引用绑定，不重写分镜内容。

### 4.4 从建项目到出成品，会产生哪些文件（按顺序）

| 阶段 | 新增/更新 | 说明 |
|---|---|---|
| 创建项目 | 项目文件夹、`screenplay.db`（首次建立/打开时初始化）、`workflow.json`、`ai_config.json` | 分集结果进 `chapters.json` |
| 整理原文（Agent） | `input/*` → `.luoshui/context/novel.md` | 长篇建议走这步 |
| 生成剧本 | `.luoshui/context/screenplay.md`、`screenplay.db`（5 表） | 有 md 会自动同步 DB |
| 剧本审核 | `screenplay_doctor_report.md`、`.luoshui/batch/reviews/screenplay_review.json` | 医生可阻断下一步 |
| 资产提取 | 三份 `*_parse_result.json`、`character_tags.json`、`agent_assets.json` | 角色/场景/道具/音色 |
| 预设质检 | `preset_doctor_report.md`、`preset_review.json` | 检查资产与剧本一致性 |
| 分镜 | 每集 `chapter_{N}_storyboard.json`、`storyboard_table.db` | 每集一份，按 `chapter_index` 区分 |
| 分镜审核 | `storyboard_review_ep{N}.json` | 医生可阻断 |
| 生图 | `images/*`（分镜图、资产图） | 路径写回分镜/资产 |
| 生视频 | `videos/*`（或引擎临时目录 `assets/dreamina_batch`） | 路径写回分镜 |
| 配音 | `audio/*`；合并音频 `audio/combined_audio.wav` | 按角色音色 |
| 收尾 | `assets/分镜_01.mp4…`、`.luoshui/context/finalize_report.md` | Agent step8 |
| 导出剪映 | 剪映草稿目录下的草稿文件夹 | 由 `jianyingExportPath` 决定 |

### 4.5 体量参考（能取到的真实数字）

| 对象 | 大小/数量 | 来源 |
|---|---|---|
| 本机 `data\剧本创作` | 0 个文件（空） | 实机只读 |
| 相邻模块「漫剧解说」样本项目 | `project.db` 49,152 B；`workflow.json` 413 B；原文 `新建文本文档.txt` 1,192 B；`audio/`、`images/`、`videos/` 为空目录 | 实机只读，**不是剧本创作项目**，仅用于说明量级 |
| 本机 `app.db` 内一套剧本分镜样本 | 44 个分镜、9 个角色、44 个场景、0 个道具、0 个音色；分镜文本每条 100 字；44 条状态全部是 `pending`；图片/视频/配音路径均为空、`image_collection` 为空数组 | 实机 SQLite 只读查询 |
| 核心模块 | `screenplay.pyd` 55.6 MB，`api.pyd` 18.2 MB，`agent.pyd` 6.9 MB | 实机文件大小 |
| 提示词库 | 本机 `prompts` 表 182 条；其中「剧本核心」46 条、「剧本分镜」40 条、「剧本大师」3 条、「中长篇转剧」8 条、「小说转剧」4 条、「短篇转剧」1 条 | 实机 SQLite 只读查询 |

结论：**文字阶段的数据量很小**（几十个分镜、几百 KB 级），真正占空间的是 `images/`、`videos/`、`audio/` 和剪映草稿。本机没有剧本创作实机项目，所以无法给出真实项目的磁盘占用。

---

## 5. 它的「工作树 / 项目树」长什么样

### 5.1 界面上看到的层级

```text
剧本创作（项目库）
└─ 项目 / 作品
   ├─ 剧本（页签）
   │  └─ 分集 第1集、第2集…（可排序、可勾选批量）
   │     └─ 本集剧本正文
   ├─ 资产（页签）
   │  ├─ 角色
   │  ├─ 场景
   │  └─ 道具（+ 角色音色）
   └─ 分镜（页签）
      └─ 分集
         └─ 分镜列表（镜头行）
            └─ 媒体：生图 / 视频 / 配音（字幕、导出剪映在导出区）
```

### 5.2 界面层级与磁盘层级是否一一对应

| 界面层级 | 磁盘对应 | 是否一一对应 | 说明 |
|---|---|---|---|
| 项目/作品 | 一个作品文件夹 | 是 | 文件夹名通常等于作品名 |
| 分集 | `chapters.json` 里的一条 + 每集分镜文件 | 是 | 分集顺序写在 `chapters.json` |
| 分镜 | 每集 `chapter_{N}_storyboard.json` + `storyboard_table.db` 的行 | 是 | `chapter_index` = 第几集 |
| 角色/场景/道具 | `agent_assets.json`（+ 旧库的 3 张表） | **否** | 资产是项目级共享库，不按集分目录 |
| 「场」 | `screenplay_scenes` 是**场景资产**，分镜用 `scene_id` 引用 | **界面上没有独立的「场」节点** | 没有「项目→集→场→镜头」的四层树；场是资产，不是文件夹 |

### 5.3 每一层能做什么

| 层 | 能做什么 | 证据 |
|---|---|---|
| 项目库 | 新建、打开、复制、导出作品、删除、取消挂机任务；个人/团队切换；网格/列表；搜索、排序 | 前端 @39254093、@39450022 |
| 分集 | 查看/编辑本集剧本；调整顺序；勾选批量生成；单独生成分镜脚本 | 前端 @33915977、@33925842、@33966074 |
| 剧本 | 选模型/提示词/风格；开始转换；AI 对话修改；状态「剧本内容整理完毕」 | 前端 @39346686、@33914254 |
| 资产 | 角色/场景/道具分组；批次推理；质量检查/完整检查；状态「资产设定已确认」 | 前端 @33899000、@33914284 |
| 分镜 | 规划；并发生成/顺序生成/自动补充；查看/修改分镜；状态「分镜内容已确认」 | 前端 @33733500、@33914317 |
| 媒体 | 按范围批量生图/生视频；配音；字幕与导出剪映在导出区 | 前端 @14955000、@14704847 |

### 5.4 一句话解释「场」

在界面上你不会看到「第 1 集 → 第 3 场 → 第 5 镜」这种四层树。洛水的做法是：**场景属于资产库，镜头属于分集**；一个镜头通过 `scene_id` 指向某个场景资产。所以「场」在磁盘上没有对应文件夹，也不会单独出现在左侧导航里（表结构：`screenplay_shots.scene_id`；资产页签：角色/场景/道具）。

---

## 6. 跟别的模块怎么接

| 去向 | 怎么触发 | 接口/桥命令 | 交接物 |
|---|---|---|---|
| 画布创作（精修图片） | 分镜媒体区按钮 | 前端事件 `open-canvas-editor`（前端 @14866107、@15300722） | 分镜图路径 |
| AI 剪辑 | 分镜编辑器按钮 | 前端事件 `open-aicut`（前端 @39246351、@39328652） | 项目路径、标题、剧本模式、视图模式 |
| 剪映 | 分镜编辑器「导出剪映」 | `export_storyboard_to_jianying`（@14704847）、`export_to_jianying`（@19776819） | 剪映草稿（写到设置里的剪映导出目录；本机该目录为空，需先设置） |
| AI 对话 | 工作台右侧 | 前端事件 `navigate-to-chat`（@39246331） | 当前项目与提示词 |
| 作品库 / 视频编辑 | 打开作品、切「视频编辑」页签 | `open_work_folder`、`export_work`；前端「视频编辑」页签（@19691740） | 整个作品文件夹 |
| 字幕 | 分镜/成片字幕 | `screenplay_subtitle_asr.pyd`；桥接 `generate_combined_subtitle`、`import_srt_subtitle` | 分镜音频 + 原文对轴 |
| 合并成片 | 后段「合并成片」 | 本机 FFmpeg | 一条完整成片 + 封面 |
| 批量挂机 | 项目右键「取消挂机任务」/定时任务 | `data/json/scheduled_screenplay_video_tasks.json`（模块字符串） | 任务队列 |
| 漫剧解说 | 不属于剧本创作内部链路 | 独立模块 | 共享部分视频/图片引擎 |

---

## 7. 它自己没做完的地方

### 7.1 明确写了「开发中」的地方

| 位置 | 原文 | 性质 | 证据 |
|---|---|---|---|
| 分镜编辑器的「配音设置」弹窗 | 「配音功能开发中...」整块占位，没有实际设置项 | **整块未完成** | 前端 @14868501 |
| 分镜编辑器的「情绪设置」按钮 | 点击只弹提示「情绪设置功能开发中」 | **未实现** | 前端 @14729294 |
| 字幕擦除渠道里的「牛马」 | 「牛马（暂未接入）」置灰；旁边文案「牛马渠道暂未接入，后续补充对应配置。」 | **渠道未接入**（属于字幕擦除，不是剧本文本链路） | 前端 @15478984、@15853963 |

### 7.2 不是「未完成」、但容易误会的半成品/双轨

| 现象 | 说明 |
|---|---|
| 同时有「旧编辑器」和「新工作台」 | 项目列表按 `editor_version` 过滤（前端 @39313981）；旧项目打开会进 `script-editor`（前端 @39267646），新项目进「剧本/资产/分镜」工作台 |
| 同时有「界面直调」和「Agent 模式」 | 两套入口都能改同一份项目文件；官方在新建弹窗里提示长篇小说优先走 Agent / Chat（前端 @39565750） |
| 同时有 `screenplay.db` 和 `storyboard_table.db` | 旧库兼容 + 新分镜工作层双写；模块明确写「不修改 screenplay.db」 |
| 「8 步」实际是 9 个编号（0–8） | 提示词写 step N/8，状态机是 0–8 共 9 键（既有对照文档第 17 行） |

### 7.3 会挡住用户的闸门（不是 bug，但要知道）

- **账号权限**：打开剧本项目时会先调用 `check_dongman_permission`，没权限直接提示「该账号无权限，请联系小洛微信：bylx109开通」并停住（前端 @39266200 起的打开流程）；分镜媒体/插件入口还会再查一次（@14599959、@14702784、@16444444）。这意味着「买了软件」不等于「能跑全部分镜与媒体功能」。
- **提示词必须先解密成功**：`get_prompt_content_decrypted` 失败会直接「获取提示词内容失败」，不会继续（前端 @39286313）。
- **空输入**：作品名为空提示「请输入作品名称」（@39265676），原文为空提示「原文内容为空」（@39271958），都不发起请求。
- **长篇但没分集标记**：会被判成短篇，只有一集；不会自动按字数硬切（前端 @39251300 `Fe()`）。

### 7.4 已知工程风险（来自既有对照文档，标注来源）

| 风险 | 说明 | 来源 |
|---|---|---|
| 批次 `state.json` 无文件锁 | 并发 `run_next` 可能丢更新 | 既有对照文档第 22 行；`agent.pyd` 中 `flock`/`.lock`/`asyncio.Lock` 均为 0 |
| A1 与 B 档共用同一个 `screenplay.md` | 两条剧本路径可能互相覆盖 | 既有对照文档第 62 行（部分为推断） |
| 分镜输出靠提示词约束 JSON | 模型不按 schema 返回时需要容错解析 | `api.pyd @0xA49477` |

### 7.5 本次「未取到」的东西（不猜）

1. 剧本创作页面的实机截图：`shots\` 下 132 张截图按文件名筛选，未见剧本创作列表/工作台相关命名（有分镜大师、小说转剧、资产参考、角色编辑等）；**未逐张打开肉眼确认**。
2. 真实剧本创作项目目录与文件大小：本机 `data\剧本创作` 为空。
3. 某个具体作品的 `ai_config.json` 内容（实际模型、promptId、风格）：无项目样本。
4. `chapter_{N}_plan` 的磁盘文件名：工具只用 `target=plan`，没有写死文件名。
5. `images/`、`videos/`、`audio/` 的内部子目录规则：模块字符串只给出片段。
6. 小云雀 CLI 的可执行文件绝对路径：只取到命令名 `pippit-tool-cli`、安装命令 `npx @pippit-dev/cli@latest install`、配置键 `engines.screenplay_xiaoyunque_cli`；绝对路径未取到。
7. 剧本创作「团队作品」的服务端/协作实现：前端有页签，但后端协作细节未取到。

---

## 附录 A：证据索引

### A.1 前端压缩包（`recovered\frontend\index-DszfAkKD.js`）

| 偏移 | 内容 |
|---|---|
| @49630446 | 左侧菜单「内容创作」分组下的「剧本创作」 |
| @39246087 | 组件名 `ScreenplayCreation` |
| @39450022 / @39454848 / @39455960 | 页面副标题 / 「创建剧本」/ 「个人作品」与「团队作品」页签 |
| @39555000–@39589568 | 新建弹窗字段：作品名称、选择模型、选择提示词（@39564389）、选择风格（@39585015）、剧本比例（@39252613）、资产解析程度（@39252826、标签 @39588705） |
| @39255397 / @39256995 | 选择文件 / 解析 `.txt`、`.docx` |
| @39251300 | 分集正则与长短篇判断（3 万字门槛） |
| @39259885 | `create_screenplay` 调用与参数 |
| @39263963 | `save_screenplay_chapters` 自动保存分集 |
| @39264463 | 长篇/短篇的两套 Agent 步骤序列 |
| @39264790 / @39264802 | 进入工作台：`R = "story"`、`_ = "ai-chat"` |
| @33913780 / @33913821 / @33913863 | 工作台三页签：剧本 / 资产 / 分镜 |
| @33914254 / @33914284 / @33914317 | 三个状态文案 |
| @33915977 | 「已调整分集顺序」 |
| @33925842 / @33926027 / @33966074 | 分集勾选与「分镜脚本」提示 |
| @33899000 起 | 导演审核、台词检查切分、查找对比、修改角色/场景/道具、资产批次推理/完整检查 |
| @33733500 起 | 规划 / 并发生成 / 顺序生成 / 自动补充 / 生成分镜 |
| @39288890 | `screenplay_ai_inference` 调用（`batch_size`、进度回调） |
| @39286691 / @39296522 | `onScreenplayAIBatch` / `onScreenplayAIProgress` 回调 |
| @39292344 | `run_completion_agent` 补全 Agent |
| @33673306 | `screenplay_master_inference` |
| @33854572 | `screenplay_generate_chapter_plan` |
| @33736362 | `screenplay_master_storyboard_inference` |
| @34100861 | `screenplay_agents_run`（Agent 路径） |
| @14704847 | `export_storyboard_to_jianying` |
| @19776819 | `export_to_jianying` |
| @14866107 / @15300722 | `open-canvas-editor` |
| @39246351 / @39328652 | `open-aicut` |
| @14868501 | 「配音功能开发中...」 |
| @14729294 | 「情绪设置功能开发中」 |
| @15478984 / @15853963 | 「牛马（暂未接入）」与说明 |
| @49978477 | `onOpenAicut` 路由处理 |
| @39267478 | `check_screenplay_db_exists`（旧库存在则进旧编辑器） |
| @39313981 | 项目列表按 `editor_version` 过滤 |
| @39266200 / @14599959 / @14702784 / @16444444 | `check_dongman_permission` 权限闸门（打开项目 + 分镜媒体/插件入口） |

### A.2 模块与文档

| 来源 | 内容 |
|---|---|
| `modules\api.pyd @0xA72FEC` | `uApi.screenplay_character_parse_agent`（资产解析桥接） |
| `modules\api.pyd @0xA76BEE / @0xA76C4F / @0xA76CA2` | 三份资产结果合并日志 |
| `modules\api.pyd @0xA66779` | 分镜推理三回调 |
| `modules\api.pyd @0xA6D956` | 剧本推理分段 |
| `modules\screenplay.pyd @0x3338BB4` | 资产解析 4 档归一化 |
| `modules\screenplay.pyd @0x3339764` | 资产模板派发表 |
| `modules\agent.pyd @0x543ABD / @0x543B90` | 批次 `new_state` / `default_steps` / `.luoshui` 目录与 `state.json` |
| `modules\agent.pyd @0x37657F` | 「短剧批次 8 步」表（标签/产物/校验） |
| `modules\agent.pyd @0x53FD84` | `ScreenplayBatchOrchestrator` |
| `modules\agent.pyd @0x545192 / @0x54586A` | step5/step6 提示词与 `chapter_{N}_storyboard.json` |
| 模块字符串 | `project_path` 示例 `E:\外部作品\剧本创作\`；`{项目路径}/screenplay.db`；`{project_path}/storyboard_table.db`；`audio/combined_audio.wav`；`assets/分镜_01.mp4`；`data/json/scheduled_screenplay_video_tasks.json` |
| 既有对照文档 | `F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\spec\漫剧剧本与分镜-需求规格-评审版.md`（第 17、20、21、22、46–47、62 行等） |

### A.3 实机只读检查

| 检查 | 结果 |
|---|---|
| `data\剧本创作` | 存在，0 个文件 |
| `data\agent\workspaces.json` | 2 个工作区：默认工作区、`q` |
| `q\` 工作区 | 实机存在 `input/`、`assets/`、`.luoshui/`（`.luoshui` 内有 `context/`、`runs/`、`sessions/`、`skills/`、`workspace.json`，空壳；`batch/`、`checkpoints/` 尚未创建） |
| `data\engine_config.json` | 当前图片引擎 `nano2`（来源 `kongyang`）、视频引擎 `grok_video_luoshui`（来源 `luoshui`）、音频引擎 `minimax` |
| `data\settings.json` | `dataPath`、`jianyingExportPath`、`apiModels`（113 条）等 |
| `data\app.db` | 剧本 5 表结构、44 分镜/9 角色/44 场景样本、`prompts` 182 条 |
| `data\漫剧解说\新建文本文档` | 相邻模块样本：`project.db` 49,152 B、`workflow.json` 413 B、`.txt` 1,192 B |
| `shots\` | 132 张截图；按文件名筛选未见剧本创作页面（未逐张肉眼确认） |

---

## 附录 B：适合谁看哪一节

| 读者 | 建议阅读 |
|---|---|
| 普通用户/运营 | 第 1、2、5、7 节；技术细节可跳过 |
| 产品/需求 | 第 0、2、5、6、7 节 + 附录 A.1 |
| 开发/逆向 | 第 3、4 节 + 附录 A 全部 |
| 排障 | 第 3.3、7.3、7.4 节 |

---

## 附录 C：术语表

| 术语 | 白话解释 |
|---|---|
| 分集 | 把小说切成第 1 集、第 2 集……每集对应一段正文 |
| 分镜 | 把一集剧本拆成一个个镜头；每个镜头有画面描述、台词、时长、角色/场景/道具绑定 |
| 资产 | 角色、场景、道具、音色这些「跨镜头复用」的设定和参考图 |
| 剧本大师 | 一套专门写剧本的提示词体系（本机 3 条） |
| 剧本分镜 | 专门写分镜词的提示词体系（本机 40 条） |
| 医生 | 审核角色：剧本医生、预设医生、分镜医生；不通过会卡住下一步 |
| 批次 | Agent 模式下的自动化流水线，按 0–8 共 9 个编号推进 |
| 界面直调 | 点按钮后前端直接调后端推理接口，不走 Agent 对话 |
| Agent | 会自己读写项目文件、调用工具、按步骤推进的执行体 |

---

**只读声明**：本次分析未修改洛水安装目录、`data\` 下任何文件，未启动洛水，未调用模型或生成接口，未重启任何服务。唯一写入是本报告文件。
