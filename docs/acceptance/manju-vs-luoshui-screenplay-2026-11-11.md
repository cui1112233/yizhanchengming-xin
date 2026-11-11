# 漫剧创作 vs 洛水「剧本创作」只读盘点与取证

- 日期：2026-11-11
- 范围：V88（`F:\脚本测试\一战晟铭-v88`）只读分析；洛水 v2.9.0 安装目录只读、bundle 只读
- 约束：未改代码、未构建、未重启 4001、未调用任何生成接口、未消耗额度
- 证据写法：V88 用「文件 + 行号」；洛水用「bundle 字符偏移」或「截图文件名」。正文用 E1～E44 编号，完整清单见第 7 节

---

## 0. 先给结论

1. V88 里用户可见的「漫剧」只有两处：
   - 菜单「漫剧创作」→ `/comic-screenplay`（新工作台）；
   - 菜单「水货生产」→ `/batch-factory`（页内标题「漫剧解说」，原批量工厂）。
2. **没有一条页面叫「漫剧创作」**。如果按“从小说/剧本一路做到成片”来定义，今天唯一能真正出成片的是 `/batch-factory`；但它产品名叫「漫剧解说」，不是洛水那种“剧本→资产→分镜”的剧本创作工作台。
3. 真正对标洛水「剧本创作」的是 `/comic-screenplay`：入口、创建表单、前三步（剧本/资产/分镜）都对得上，**但做到“分镜词”就停了**。第四步「生产与导出」只是把 9 个步骤的状态写进数据库，没有任何执行器去生成图片、视频、旁白或成片。
4. 「私人定制 → 漫画生成」和「CM 画布 → 漫画作品」都是“漫画/图片”链路，不是漫剧视频；「配音」「AI-CUT」是独立工具，没有接进漫剧工作台。
5. 因此“漫剧创作”在项目里**没有唯一指代**，需要用户拍板：是把 `/comic-screenplay` 定为漫剧创作入口（推荐），还是以 `/batch-factory` 为当前可交付生产线。详见 2.6。

---

## 1. 口径与阅读说明

- “漫剧”在本报告里指：有剧本、角色/场景/道具资产、分镜，最终能生成画面、配音并合成视频的内容形态。
- “漫画”指：按页出图、以图片合集为最终产物，不合成视频。
- 洛水 bundle 是压缩后的单行 JS，字符偏移是文件内的绝对位置；同一段文案可能出现多次，报告引用的是最靠近实际页面逻辑的那一处。
- 本文只证明“代码和界面结构上有什么”，**没有实跑生成**；涉及“能不能出片”的判断，均以现有执行器、接口和页面文字为依据，并标注“未实测”。

---

## 2. V88 盘点：所有和“漫剧”有关的入口

### 2.1 总表

| 编号 | 用户看到的名字 | 路由/入口 | 页面形态 | 最终产物 | 当前状态 |
|---|---|---|---|---|---|
| A | 漫剧创作 | `/comic-screenplay` | 项目库 + 四步工作台 | 剧本、资产、分镜 JSON；**没有图片/视频/旁白/成片** | 前三步可跑，第四步是状态壳 |
| B | 水货生产（页内标题：漫剧解说） | `/batch-factory` | 项目库 + 批量工厂 + 单本生产台 | 视频成片（自动流程含音频与合成），或漫画图片合集 | 有真实执行器，未实测额度 |
| C | 私人定制 → 漫画生成 | `/script` 的“漫画生成”页签 | 任务表单 + 漫画页结果 | 漫画页图片 | 独立于漫剧 |
| D | CM 画布 → 漫画作品 | `/agent` 项目类型 `comic` | 画布/项目 | 剧本、文本、图片、漫画提示词 | 明确不是视频链路 |
| E | 配音 / AI-CUT | `/tts`、`/ai-cut` | 独立工具页 | 音频、字幕/粗剪 | 没有接进 A 或 B |

依据：E1、E2、E20、E26、E27、E28。

### 2.2 主线 A：`/comic-screenplay`「漫剧创作」

#### 入口和页面

- 左侧菜单「漫剧创作」，路由 `/comic-screenplay`（E1、E2）。
- 页面分两层：
  1. **项目库**：标题“漫剧创作 / 管理和创建您的漫剧作品”，个人作品/团队作品、搜索、排序、新建（E2、E3）。
  2. **项目工作台**：左侧四步导航，右侧四个页签（E4）。

#### 四步与四个页签

| 步骤 | 页签 | 一句话在干什么 | 页面上的关键动作 |
|---|---|---|---|
| 1 | 原文与剧本 | 上传原著，让模型把它改成分集剧本 | 生成剧本、保存剧本、查看原文 |
| 2 | 资产库 | 从剧本里抽出人物/场景/道具，并可编辑 | 提取资产、编辑名称/身份/描述/提示词/音色/标签 |
| 3 | 分镜工作台 | 按集生成 CM 分镜（场次+六个分镜字段） | 生成 CM 分镜、编辑/删除/排序场次 |
| 4 | 生产与导出 | 本应生成图片、视频、旁白并合成 | 启动/暂停/继续生产、导出项目 JSON |

依据：E4、E5、E6。

#### 创建项目时要填什么

- 作品名称；
- 选择提示词（内置 A1/A2/B 等）；
- 选择文本模型；
- 选择画风；
- 剧本比例（9:16 等）；
- 资产解析程度（4 档，和洛水一致）；
- 上传剧本（`.txt` / `.docx`）。
- 工作台右侧还能改：文本模型、画面比例、创作模式（小说转剧本 / 导演级剧本）。

依据：E5、E6。

#### 每一步调哪个接口、数据存哪

| 动作 | 前端调用 | 后端接口/处理 | 数据落点 |
|---|---|---|---|
| 项目列表/新建/删除 | `listComicScreenplays` / `createComicScreenplay` / `deleteComicScreenplay` | `GET/POST /api/comic-screenplays`、`GET/DELETE /api/comic-screenplays/projects/{id}` | `screenplay_projects` |
| 提示词列表 | `listComicScreenplayPrompts` | `GET /api/comic-screenplays/prompts` | 内置 prompt 语料 |
| 解析原文 | `parseComicScreenplaySource` | `POST .../source/parse` | `screenplay_projects.source_text` |
| 生成剧本 | `generateComicScreenplay` | `POST .../generate` | `screenplay_documents` |
| 手工改剧本 | `updateComicScreenplayDocument` | `PATCH .../document` | `screenplay_documents` |
| 提取资产 | `extractComicScreenplayAssets` | `POST .../assets/extract` | `screenplay_characters` / `screenplay_scenes` / `screenplay_props` |
| 编辑资产 | `create/update/deleteComicScreenplayAsset` | `.../assets` | 同上 + `screenplay_assets` |
| 生成分镜 | `generateComicStoryboard` | `POST .../storyboards/generate` | `screenplay_storyboards` / `screenplay_storyboard_shots` |
| 编辑场次 | `create/update/deleteComicScreenplayShot` | `POST/PATCH/DELETE .../shots` | `screenplay_shots` |
| 启动/暂停/继续生产 | `start/pause/resumeComicScreenplayBatch` | `POST .../batch`、`.../batch/cancel` | `screenplay_batches` / `screenplay_batch_steps` / `screenplay_batch_units` |
| 导出 | `exportComicScreenplay` | `GET .../export` | 返回 JSON 包，不生成媒体文件 |

依据：E6、E7、E8、E9、E10、E11、E12、E13、E16。

#### 现在真实能跑到哪一步

**能真实跑（代码具备、调用的是用户自己配置的文本模型）：**

1. 建项目、上传并解析 TXT/DOCX；
2. 生成/编辑分集剧本；
3. 提取/编辑人物、场景、道具；
4. 按“当前选中集”生成分镜词，分镜字段为：场次标题、剧本内容、角色、场景、道具、视频描述；
5. 手工增删改场次、导出项目 JSON。

**跑不到（占位或空壳）：**

- 资产没有参考图、没有资产衍生图、没有音色试听；
- 分镜没有“先写规划、再写分镜词”的两步，也没有分镜医生/审核；
- 分镜结果只有文字，没有每个场次的图片、视频、配音、字幕字段；
- “生产与导出”点“启动生产任务”只会创建 9 个步骤记录，全部停在 pending；
- 没有后台 worker 推进这 9 步：全仓 `NextUnit(` 只出现在 `*_test.go`；
- 页面自己也写明：“图片、视频和旁白生产需要配置可用服务；当前任务会把九步状态持久化，暂停和继续不会丢失进度。”

依据：E4、E13、E14、E15、E17、E18、E19。

#### 九步批次是哪九步

批次定义在 `state.go`，和洛水后端拆解出的九步一致：原文/剧本/剧本医生/资产/资产预设/分镜/分镜医生/批量出视频/收尾。其中“批量出视频”被标为依赖外部能力。

依据：E14。

### 2.3 主线 B：`/batch-factory`「漫剧解说 / 批量工厂」

#### 它是什么

- 菜单名「水货生产」，页面标题「漫剧解说」，按钮「创作漫剧」「批量工厂」（E20）。
- 创建项目时输入的是“小说/原文”，不是已经写好的分集剧本（E21）。
- 它有两个生产模式（E22）：
  - **视频模式**：原文 → 资产/图片 → 导演提示词 → 视频 → 上传 → 完成；
  - **漫画模式**：资产 → 漫画初始化 → 漫画图 → 合集 → 发布。

#### 真实执行链

前端自动化阶段（E22）：

```
视频：preflight → fetch_original → assets → director → video → upload → complete
      运行期：preflight → fetch_original → storyboards → openings → audio → merge → submit
漫画：assets → comic_bootstrap → comic_images → comic_collection → comic_publish
```

后端确有执行器（E23、E24、E25）：

- `batch_factory_automation.go` 有自动化协调器，按阶段执行、可暂停/继续/重试；
- `audio` 阶段会逐句走 TTS，并把音频拼起来；
- `merge` 阶段调用合并服务，合并 worker 用 ffmpeg 拼视频并替换音轨；
- 视频通道已有 `minimax-h3-video`、seedance 等（E44）。

#### 它和“漫剧创作”的区别

| 维度 | `/comic-screenplay` 漫剧创作 | `/batch-factory` 漫剧解说 |
|---|---|---|
| 起点 | 上传剧本 | 小说/原文 |
| 页面 | 剧本/资产/分镜/生产四页签 | 项目库 + 批量工厂 + 单本生产台 |
| 人工编辑 | 剧本、资产、场次都能逐条改 | 以批次和自动流程为主 |
| 图片 | 没有 | 有资产图、漫画图 |
| 配音 | 只有文字音色描述 | 有 TTS 音频阶段 |
| 成片 | 没有 | 有音频合并、视频合并、上传 |
| 定位 | 洛水式剧本创作工作台 | 可出片的批量生产线（解说向） |

**结论：** 如果用户要“今天就能出成片”，B 是唯一选择；如果用户要“像洛水一样从剧本开始创作”，A 的入口和前三步已经就位，但必须把 B 的生产能力接进来。

### 2.4 相邻链路：私人定制「漫画生成」

- `/script` 页面有“剧本创作 / 漫画生成”两个页签（E26）。
- 漫画生成调用私人漫画任务，产物是漫画页/图片，不是漫剧视频。
- 它和 `/comic-screenplay` 不共享项目数据，也不是同一套页面。

### 2.5 相邻链路：CM 画布「漫画作品」

- Agent 项目类型里有 `comic`，用户看到的名字是「漫画作品」；
- 项目类型说明写的是“剧本、文本、图片和漫画提示词生成链路”（E27）；
- 没有“成片/配音/合成”的承诺，因此不算漫剧创作。

### 2.6 「漫剧创作」到底指哪条链：歧义说明

**代码事实：**

- 唯一带“漫剧创作”字样的是 A；
- 唯一带“漫剧解说/创作漫剧”字样的是 B；
- 没有第三个“漫剧创作”页面。

**建议口径（需用户确认）：**

- **口径一（推荐）**：把 A `/comic-screenplay` 定义为“漫剧创作”的产品入口，对标洛水剧本创作；把 B 降为“生产基础设施/漫剧解说”，只复用它的图片、视频、TTS、合并能力。
- **口径二**：把 B 定义为“漫剧创作”，因为它今天能出成片；A 只作为未来的剧本工作台。
- 两个口径会直接改变差距清单的优先级：口径一优先补 A 的生产执行器；口径二优先补 B 的剧本/资产/分镜编辑体验。

---

## 3. 洛水「剧本创作」拆解

### 3.1 入口与工作模式

- 工作模式：洛水有“短剧创作 / Agent”两种模式（E30）；「剧本创作」属于“短剧创作”模式。
- 入口：左侧「内容创作」分组下的「剧本创作」，内部 key 为 `screenplay`（E29）。
- 首页也有一张「剧本创作」大卡片，和菜单指向同一入口（E29）。
- 另有权限门槛：进入剧本创作会检查“动漫权限”，无权限提示联系客服（E41）。V88 目前没有这个门槛。

截图证据：
- 左侧导航与工作模式：`F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\shots\novel-to-drama-panel.jpg`
- 剧本创作项目库：`C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-list.png`

### 3.2 项目库与创建弹窗

项目库页面（截图 `luoshui-screenplay-list.png`）：

- 标题「剧本创作」，副标题「管理和创建您的剧本作品」；
- 页签「个人作品 / 团队作品」；
- 搜索「搜索作品...」、排序「按时间降序」；
- 右上「创建剧本」；
- 空态提示：点击右上角“创建剧本”按钮创建你的第一个作品。

创建弹窗（截图 `luoshui-screenplay-create-dialog.png`）字段：

| 字段 | 可见值/选项 |
|---|---|
| 作品名称 | 必填 |
| 选择模型 | 例如「氛围版-gemini-3.1」（洛水氛围版，High） |
| 选择风格 | 点击选择画风 |
| 剧本比例 | 9:16（竖屏）等 |
| 资产解析程度 | 大师模式-3D / 大师模式-写实 / 差异化 / 标准 |
| 上传剧本 | 支持 TXT / DOCX |
| 底部 | 取消 / 创建 |

> 结论：V88 的创建表单与洛水高度一致，只多了“选择提示词”，少了一点洛水模型品牌包装。

### 3.3 项目内页签与顺序

洛水项目工作台是 **3 个主页签**，顺序固定（E33）：

1. **剧本**
2. **资产**
3. **分镜**

页签切换时有完成提示，明确要求“上一步确认后再进入下一步”（E33）：

- 剧本内容整理完毕 → 进入资产；
- 资产设定已确认 → 进入分镜；
- 分镜内容已确认 → 进入生成。

> 这正是“分几步、停在哪”的核心：洛水把人工确认点放在三个主页签之间。

### 3.4 第 1 步：剧本

**输入：**

- 小说章节（来自作品库）或导入的剧本；
- 选择的文本模型、提示词/模式；
- 分集方式。

**页面上的工具（E34）：**

- AI 生剧本；
- 文稿对比；
- 版本；
- 剧本医生；
- 爆款开头；
- AI 分析；
- 导入剧本。

**输出：**

- 分集剧本（可多版本）；
- `screenplay.md`。

**调用的后端方法（E38）：**

- `screenplay_ai_inference`：普通 AI 生剧本；
- `screenplay_master_inference`：大师模式剧本；
- `screenplay_agents_run`：Agent 流程；
- `screenplay_generate_chapter_plan`：章节规划。

### 3.5 第 2 步：资产

**输入：**

- 已确认的剧本；
- 资产解析程度（4 档）；
- 文本模型/资产提示词。

**做什么：**

- 资产分析：抽出人物、场景、道具；
- 生成/编辑资产提示词；
- 生成资产参考图；
- 资产衍生、重生、完整性检查；
- 为角色配置音色描述。

**输出：**

- 角色/场景/道具条目；
- 参考图；
- `agent_assets.json`。

**调用方法（E38）：**

- 资产解析 Agent 命令（`screenplay_character_parse_agent` 等）；
- `generate_ai_character_scene`：角色/场景参考图；
- `appearance_generate_images`：形象图；
- `canvas_save_asset_image`：保存资产图。

### 3.6 第 3 步：分镜（先规划，再写词）

洛水的分镜不是“一步出词”，而是两步（E35）：

1. **先写本集分镜规划**（`先写本集分镜规划`）；
2. **按规划生成分镜词**（`按规划生成分镜词`）；
3. 支持 **继续生成未完成的分镜**。

**输入：**

- 已确认剧本；
- 资产库（角色/场景/道具名称与描述）；
- 分镜模型/提示词；
- 画风与画面比例。

**输出：**

- 每集一份分镜 JSON：`chapter_{N}_storyboard.json`；
- 每个场次六个字段：场次标题、剧本内容、角色、场景、道具、视频描述。

**调用方法（E38）：**

- `storyboard_ai_inference`：分镜词；
- `screenplay_master_storyboard_inference`：大师模式分镜；
- `screenplay_agents_run`：Agent 分镜流程。

### 3.7 分镜词之后：怎么接图像、配音、合成

洛水把这一段放在分镜页签内部，分成 3 个媒体页签（E36）：

**① 生图**

- 每个场次一条图片任务；
- 支持单条生成、批量生图、取消批量；
- 图像会参考资产图（角色/场景/道具）；
- 生图引擎包括：即梦画布 CLI、小云雀 CLI（`pippit-tool-cli generate-image`）、Seedream、GRSAI nano-banana 等（E43）。

**② 视频**

- 用场次图片（首帧/参考图）+ 视频描述 + 音频生成视频；
- 支持单条、批量生视频、取消批量、生成日志；
- 视频引擎包括：即梦画布、小云雀（`generate-video`）、可灵、Vidu、SD、Seedance 等（E43）；
- 另有“白模视频”作为参考。

**③ 配音**

- 每个场次/角色一条配音；
- 支持批量配音、逐条更新音色；
- 音色来源：MiniMax、Qwen-TTS、CosyVoice、音色库、多人配音（E43）；
- **注意**：洛水的“配音设置”弹窗里仍写着“配音功能开发中...”（E42），说明它自己也还没完全收口；批量配音和逐条音色更新是有的。

**④ 字幕与音频合并**

- 生成字幕：ASR 识别 + 与原文对轴（E39）；
- 合并音频：把所有场次配音合并成一条，含气口处理（E39）。

**⑤ 合并成片与导出**

- 合并成片：把所有场次视频拼成一条，可后台执行，统一分辨率/帧率，补静音，重新抽封面（E39）；
- 超长视频：按 100/200/250/300 秒分组，解决平台时长限制（E39）；
- 片段重拍：首尾帧 + 视频参考重生成（E39）；
- 分镜合并：相邻场次合并，最长 30 秒（E39）；
- 导出剪映：两种模式——“直接导出原视频”和“智能剪辑导出”（E39）。

**关键桥命令（E38）：**

| 环节 | 桥命令 |
|---|---|
| 批量生图 | `batch_generate_images`、`generate_storyboard_images_batch`、`generate_single_image` |
| 批量生视频 | `batch_generate_videos`、`generate_single_video`、`add_video_task` |
| 批量配音 | `batch_generate_narration_audio`、`tts_batch_start`、`update_storyboard_voice` |
| 字幕 | `generate_combined_subtitle`、`process_subtitle_alignment_async` |
| 合并音频 | `merge_subtitle_audio` |
| 合并成片 | `merge_storyboards`、`render_timeline_video` |
| 剪映导出 | `export_to_jianying`、`export_storyboard_to_jianying` |

### 3.8 最终产物与落盘

洛水是“本地项目目录 + JSON 文件”的模式（E40）：

- `screenplay.md`：剧本；
- `chapter_{N}_storyboard.json`：每集分镜；
- `agent_assets.json`：资产库；
- 项目目录下还有 storyboards、图片、视频、配音等文件；
- 剪映草稿导出到本地。

V88 则全部存 MySQL，导出时只给一个 JSON 包，不落标准作品目录（E16）。

### 3.9 洛水自身的“未完成项”

- “配音设置”弹窗写着“配音功能开发中...”（E42）；
- 生图/视频依赖本机 CLI 或外部服务（即梦画布、小云雀、可灵、GRSAI、Seedream 等），不是纯云端；
- 因此“洛水有”不等于“V88 必须一比一照搬”，尤其是配音设置面板。

---

## 4. 对照表

| 维度 | V88 `/comic-screenplay` 漫剧创作 | 洛水「剧本创作」 | 差异结论 |
|---|---|---|---|
| 入口 | 左侧菜单「漫剧创作」 | 工作模式「短剧创作」→ 内容创作 → 剧本创作 | 入口层级不同，V88 少“工作模式” |
| 工作模式 | 无 | 短剧创作 / Agent 两种 | V88 无模式切换 |
| 项目库 | 个人/团队、搜索、排序、新建 | 个人/团队、搜索、排序、创建剧本 | 基本对齐 |
| 创建表单 | 名称、提示词、模型、画风、比例、资产程度、上传 | 名称、模型、风格、比例、资产程度、上传 | 基本对齐，V88 多提示词 |
| 步骤数 | 4 页签（剧本/资产/分镜/生产） | 3 主页签（剧本/资产/分镜）+ 分镜内媒体子步骤 | 主步骤同序，V88 多一个“生产”空页签 |
| 第 1 步输入 | 上传 TXT/DOCX | 小说章节或导入剧本 | 基本一致 |
| 第 1 步输出 | 分集剧本（单版本） | 分集剧本（多版本）+ screenplay.md | V88 缺版本/对比/医生/爆款开头 |
| 第 2 步输入 | 剧本 + 4 档资产程度 | 剧本 + 4 档资产程度 + 模型 | 基本一致 |
| 第 2 步输出 | 人物/场景/道具文本 | 资产文本 + 参考图 + 音色描述 + agent_assets.json | V88 缺资产图/衍生/完整性检查 |
| 第 3 步输入 | 剧本 + 资产文本 + 固定分镜引擎 | 剧本 + 资产 + 规划 | V88 缺“先规划后写词” |
| 第 3 步输出 | 分镜词（六字段） | 分镜规划 + 分镜词（六字段）+ chapter_N_storyboard.json | 字段对齐，流程少一步 |
| 第 4 步 | 生产与导出：只写 9 步状态 | 生图→视频→配音→字幕→合并音频→合并成片→导出 | **最大缺口** |
| 图像 | 无 | 单条/批量生图，参考资产图 | V88 完全缺 |
| 视频 | 无 | 单条/批量生视频，多引擎 | V88 完全缺 |
| 配音 | 只有文字音色描述 | 批量 TTS + 音频合并（设置面板仍开发中） | V88 缺 |
| 字幕 | 无 | ASR + 对轴 | V88 缺 |
| 合成 | 无 | 合并成片、超长视频、剪映导出 | V88 缺 |
| 中间数据 | MySQL 19 张表 | 本地项目目录 + JSON | 存储模型不同 |
| 最终产物 | 项目 JSON | 剧本、资产、分镜 JSON、图片、视频、音频、成片 | V88 只有前半段 |
| 停在哪一步 | 分镜词 | 成片/剪映草稿 | V88 停在约 40% |
| 人工介入 | 每步手动；生产点了也没用 | 三步确认 + 媒体批量；可选自动停点 | V88 人工点更多、产出更少 |
| 权限 | 无动漫权限门槛 | 有动漫权限门槛 | 产品策略差异 |
| 引擎选择 | UI 固定 xiaoluo v7 15s | 多引擎可选（小云雀 v2/v3、即梦、可灵等） | V88 后端有 v3，UI 未暴露 |

---

## 5. 差距清单（按影响排序）

### P0：不补就无法“出片”

**G1. 没有生产执行器，9 步批次是状态壳**

- 现象：点“启动生产任务”只创建 batch + step 记录，状态全 pending；没有 worker 推进。
- 依据：E13、E14、E15。
- 类型：**新做后端**（可复用 B 的自动化协调器）。
- 影响：这是“漫剧创作”断掉的第一刀；不补，后面所有媒体能力都没有入口。

**G2. 没有场次级媒体数据模型和接口**

- 现象：分镜场次只有六字段文字，没有图片、视频、配音、字幕字段；后端没有对应接口。
- 依据：E12、E18。
- 类型：**新做后端**（新增表/字段 + 接口）。
- 影响：即使接了生成服务，也没有地方存结果。

**G3. 资产没有参考图/衍生图**

- 现象：资产只有名称、身份、描述、提示词、音色、标签；没有图片生成、没有资产衍生、没有完整性检查。
- 依据：E4、E11。
- 类型：**接一下 + 新做少量后端**（复用 B 的资产图/图片通道）。
- 影响：分镜生图没有角色/场景参考，画面一致性无法保证。

### P1：决定“像不像洛水”

**G4. 分镜缺“先规划、再写词”**

- 现象：V88 直接生成分镜词；洛水先出本集规划，再按规划写词，并支持续写未完成分镜。
- 依据：E35、E17。
- 类型：**新做后端编排 + 前端按钮**（提示词已有）。
- 影响：长集/复杂剧情的分镜稳定性差距明显。

**G5. 缺剧本医生、分镜医生、预设质检、爆款开头、版本对比**

- 现象：V88 只有生成/编辑；洛水有剧本医生、分镜医生、预设审核、爆款开头、文稿对比、版本。
- 依据：E34、E19。
- 类型：**新做后端 + 前端**（医生/预设提示词已经在 V88 的 prompt 语料里）。
- 影响：这是“可用的初稿”和“能交付的剧本”的差距。

**G6. 生图/视频/配音/字幕/音频合并/成片全缺**

- 现象：V88 没有任何媒体生成接口；洛水有完整的单条/批量媒体链。
- 依据：E36、E37、E38、E39。
- 类型：**新做后端为主，部分可复用 B**。
- 影响：这是“漫剧”与“剧本工具”的分界线。

**G7. 分镜引擎只有一个，且 UI 写死**

- 现象：V88 后端支持小云雀 v3 与小洛 v7，但前端固定传 `xiaoluo_v7_15s`；没有引擎下拉。
- 依据：E17、E4。
- 类型：**接一下就能补齐**（后端能力已有）。
- 影响：画面风格/时长适配受限。

**G8. 不能按“全部集”批量生成分镜**

- 现象：V88 只对当前选中集生成；洛水支持逐集、续写未完成。
- 依据：E4、E12。
- 类型：**接一下 + 少量后端**。
- 影响：多集项目操作成本高。

### P2：体验与工程化

**G9. 没有标准作品目录和产物包**

- 现象：V88 全存 MySQL，导出只有 JSON；洛水有 `screenplay.md`、`chapter_N_storyboard.json`、`agent_assets.json`、媒体目录。
- 依据：E40、E16。
- 类型：**新做导出格式**。
- 影响：迁移、备份、与外部工具协作不方便。

**G10. 没有动漫权限门槛/团队权限**

- 现象：洛水进入剧本创作要检查动漫权限；V88 只有个人/团队作品筛选，没有对应权限控制。
- 依据：E41。
- 类型：**新做后端权限**。
- 影响：商业化/内测控制缺失。

**G11. 缺超长视频、片段重拍、分镜合并、剪映导出**

- 现象：洛水有这些成片后的进阶能力；V88 没有。
- 依据：E39。
- 类型：**依赖外部 CLI/本机环境 + 新做后端**。
- 影响：不影响“能出片”，影响“能交付”。

**G12. 外部引擎依赖**

- 洛水依赖：即梦画布 CLI、小云雀 CLI、可灵 CLI、GRSAI nano-banana、Seedream、MiniMax/Qwen-TTS/CosyVoice、ASR、ffmpeg（E43）。
- V88 已有：seedance、minimax-h3-video、ffmpeg 合并、TTS 页面、AI-CUT 字幕能力（E24、E25、E28、E44）。
- 类型：**外部依赖**；需要逐项确认账号、额度和安装环境。
- 影响：不是代码写不写的问题，是“有没有通道”的问题。

---

## 6. 需要用户拍板的问题

1. **“漫剧创作”以哪条链为准？**
   - A：`/comic-screenplay` 漫剧创作（推荐，和洛水剧本创作对齐）；
   - B：`/batch-factory` 漫剧解说（今天唯一能出片）。
2. **V88 的“生产”是复用 `/batch-factory` 的自动化，还是为 `/comic-screenplay` 新写一套？**
   - 复用：快，但要把 screenplay 的数据映射到 batch factory 的资产/导演/视频/音频模型；
   - 新写：干净，但工作量大。
3. **洛水的哪些外部引擎必须对齐？**
   - 只要求“能出片”，还是必须支持即梦画布/小云雀/可灵/GRSAI 等具体通道？
4. **剧本医生、版本对比、爆款开头是否列入第一版？**
   - 这些不影响“出片”，但决定“像不像洛水”。

---

## 7. 证据附录

### 7.1 V88 证据

| 编号 | 文件/位置 | 证明什么 |
|---|---|---|
| E1 | [UserLayout.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/layouts/UserLayout.jsx#L30-L38) | 菜单含“漫剧创作”“水货生产” |
| E2 | [App.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/App.jsx#L79-L100) | `/comic-screenplay`、`/batch-factory` 等路由 |
| E3 | [ComicScreenplayPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ComicScreenplayPage.jsx#L259-L263) | 项目库标题、个人/团队页签 |
| E4 | [ComicScreenplayPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ComicScreenplayPage.jsx#L198-L228) | 四步导航、四个页签、生产页文字 |
| E5 | [ComicScreenplayPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ComicScreenplayPage.jsx#L264-L280) | 创建表单字段 |
| E6 | [ComicScreenplayPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ComicScreenplayPage.jsx#L205-L227) | 各步骤调用的前端 API |
| E7 | [comicScreenplay.js](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/api/comicScreenplay.js#L1-L80) | 前端 API 封装 |
| E8 | [screenplay_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/screenplay_routes.go#L70-L120) | 项目/提示词接口 |
| E9 | [screenplay_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/screenplay_routes.go#L256-L330) | 原文解析、剧本生成、文档保存 |
| E10 | [screenplay_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/screenplay_routes.go#L332-L407) | 资产提取/编辑 |
| E11 | [screenplay_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/screenplay_routes.go#L408-L429) | 分镜生成/编辑 |
| E12 | [screenplay_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/screenplay_routes.go#L431-L478) | 批次创建/状态/取消/导出 |
| E13 | [screenplay_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/screenplay_routes.go#L445-L477) | 启动生产只写 batch/step，不执行 |
| E14 | [state.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/screenplay/state.go#L20-L35) | 九步定义，视频步依赖外部 |
| E15 | 全仓搜索 `NextUnit(`，仅 `*_test.go` | 没有运行时执行器推进批次 |
| E16 | [screenplay_schema.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/storage/screenplay_schema.go#L10-L190) | 19 张 MySQL 表 |
| E17 | [storyboard_service.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/screenplay/storyboard_service.go#L258-L336) | 分镜提示词、资产占位、双引擎 |
| E18 | [model.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/screenplay/model.go#L120-L300) | 项目/资产/场次字段（无图片/视频/配音） |
| E19 | [prompts 目录](file:///f:/脚本测试/一战晟铭-v88/backend/internal/prompt/luoshui/prompts) | 已有 doctor/batch/分镜语料 |
| E20 | [ProjectsView.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/shuihuo/ProjectsView.jsx#L180-L196) | “漫剧解说”“创作漫剧”“批量工厂” |
| E21 | [CreateProjectModal.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/shuihuo/CreateProjectModal.jsx#L36-L70) | 从小说创建漫剧项目 |
| E22 | [automationStages.json](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/shuihuo/automationStages.json#L1-L40) | 视频/漫画两套真实阶段 |
| E23 | [stage_runs.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/batchfactoryv11/stage_runs.go#L10-L35) | 资产/导演/开场/视觉/漫画/图片/视频阶段 |
| E24 | [batch_factory_automation.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/batch_factory_automation.go#L118-L135) | 自动化协调器、阶段执行、TTS/合并 |
| E25 | [ffmpeg.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/mergeworker/ffmpeg.go#L140-L215) | ffmpeg 合并/换音轨 |
| E26 | [ScriptPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ScriptPage.jsx#L2218-L2226) | 私人定制“剧本创作/漫画生成” |
| E27 | [agent-project-type.js](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/agent-project-type.js#L28-L36) | CM 画布“漫画作品”定义 |
| E28 | [UserLayout.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/layouts/UserLayout.jsx#L36-L39) | 配音、AI-CUT 为独立菜单 |
| E29 | 洛水 bundle `index-DszfAkKD.js` @49630354、@2877128 | 内容创作→剧本创作入口、首页卡片 |
| E30 | 洛水 bundle @15475925 | 工作模式：短剧 / Agent |
| E31 | 截图 `luoshui-screenplay-list.png` | 剧本创作项目库页面 |
| E32 | 截图 `luoshui-screenplay-create-dialog.png` | 创建剧本弹窗字段与默认值 |
| E33 | 洛水 bundle @33913782、@33914255 | 剧本/资产/分镜三页签与完成提示 |
| E34 | 洛水 bundle @7492754、@7983830、@8441888 等 | 资产分析、爆款开头、文稿对比等工具 |
| E35 | 洛水 bundle @6760215、@6760535、@6760785 | 先规划→按规划写词→续写 |
| E36 | 洛水 bundle @14956799、@14963922、@14971056 | 生图/视频/配音三个媒体页签 |
| E37 | 洛水 bundle @14684014、@14685822、@14688875 | 批量生图/生视频/配音 |
| E38 | 洛水 bundle 方法名偏移见 7.2 | 各环节桥命令 |
| E39 | 洛水 bundle @14888401、@14898152、@14997535、@15002328、@2698240 | 字幕、合并音频、导出模式、合并成片 |
| E40 | 洛水 bundle @2747744、@2747898、@2770427 | `screenplay.md`、`chapter_N_storyboard.json`、`agent_assets.json` |
| E41 | 洛水 bundle @14599959 | 动漫权限检查 |
| E42 | 洛水 bundle @14868497 | “配音功能开发中...” |
| E43 | 洛水 bundle @2701828、@2708188、@2701953、@2709773、@2874221、@2874245、@2744155 | 即梦/小云雀/Seedream/nano-banana/Qwen-TTS/CosyVoice/MiniMax |
| E44 | [autodl_h3_video_adapter.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/batchfactoryv11/autodl_h3_video_adapter.go#L15-L25) | V88 已有 `minimax-h3-video` 图+音→视频通道 |

### 7.2 洛水桥命令与偏移

| 命令 | 偏移 | 用途 |
|---|---|---|
| `screenplay_ai_inference` | 8238790 | AI 生剧本 |
| `screenplay_master_inference` | 33673310 | 大师模式剧本 |
| `screenplay_master_storyboard_inference` | 8203623 | 大师模式分镜 |
| `screenplay_agents_run` | 34100865 | Agent 流程 |
| `storyboard_ai_inference` | 14552677 | 分镜词 |
| `screenplay_generate_chapter_plan` | 33672757 | 章节规划 |
| `batch_generate_images` | 14684575 | 批量生图 |
| `generate_storyboard_images_batch` | 14684613 | 分镜批量图 |
| `generate_single_image` | 6713689 | 单图 |
| `batch_generate_videos` | 14686374 | 批量生视频 |
| `generate_single_video` | 14686589 | 单视频 |
| `add_video_task` | 14686472 | 视频任务 |
| `batch_generate_narration_audio` | 14689544 | 批量配音 |
| `tts_batch_start` | 16758878 | TTS 批量 |
| `update_storyboard_voice` | 14859031 | 更新音色 |
| `generate_combined_subtitle` | 14888401 | 生成字幕 |
| `merge_subtitle_audio` | 14898152 | 合并音频 |
| `merge_storyboards` | 26352271 | 合并成片 |
| `render_timeline_video` | 22431712 | 时间线渲染 |
| `export_to_jianying` | 14997535 | 导出剪映 |
| `export_storyboard_to_jianying` | 14993448 | 分镜导出剪映 |

> 说明：偏移是 bundle 中该字符串首次/最靠近实际调用的位置；同一命令可能出现多次。部分数字为对象属性/调用参数位置，用于定位，不改变结论。

### 7.3 截图清单

| 文件 | 内容 |
|---|---|
| `F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\shots\novel-to-drama-panel.jpg` | 洛水左侧导航：短剧创作、内容创作→剧本创作 |
| `C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-list.png` | 洛水剧本创作项目库 |
| `C:\Users\Administrator\AppData\Local\Temp\luoshui-screenplay-create-dialog.png` | 洛水创建剧本弹窗 |

---

## 8. 一句话总结

V88 的「漫剧创作」已经有一个很像洛水的壳：入口、创建表单、剧本/资产/分镜三步都对得上；但它只做到了“分镜词”，后面的生图、视频、配音、字幕、合成、成片全部没有接。真正能出片的是另一条叫「漫剧解说」的批量生产线。下一步不是再画页面，而是决定用哪条生产链，把“分镜词 → 画面 → 声音 → 成片”这段接上。
