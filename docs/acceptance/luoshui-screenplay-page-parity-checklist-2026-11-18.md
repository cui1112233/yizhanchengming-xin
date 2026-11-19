# 洛水「剧本创作」页面逐项对齐清单

> 目标：按原版**同一个页面、同一个操作状态**逐项补齐；模型接口一律用 V88 自己的，不重新设计后端。
> 取证源：`F:\软件\洛水2026\洛水2026\dist\assets\index-C1aA08SG.js`（2026-11-17 构建，最新版）
> + 官方更新日志原文（bundle 内 changelog）。V88 侧证据为源码文件:行号。
> 状态：`✅已对齐` / `⚠️能力有但位置不同` / `❌缺失待补`

---

## 零、页面结构对照（骨架）

| 项 | 洛水原版 | V88 现状 | 状态 |
|---|---|---|---|
| 顶部导航 | 首页/小说创作/角色库/音色库/漫剧解说/**内容创作组（短剧工作室·剧本创作·momo画布）**/AI剪辑/社区 | 左侧菜单「漫剧创作」单入口 | ⚠️ 形态不同（桌面多模块 vs Web 单页），不强行照搬外壳 |
| 项目库 tab | 个人作品/团队作品 | 个人作品/团队作品/已归档 + 归档库/普通资产库/历史/创建漫剧 | ✅ 已对齐（V88 更全） |
| 工作区三阶段 | 剧本 / 资产 / 分镜（Tab） | `剧本/资产/分镜` 三阶段 tab | ✅ 已对齐（`comic-screenplay-stages.js:2-4`） |
| 右侧 Agent 抽屉 | Agent 对话/画布，@提及素材、技能按钮 | 顶部 `Agent` 按钮 → 创作 Agent 抽屉 + 独立「CM 画布」 | ✅ 已对齐（`ComicScreenplayPage.jsx:649,656`） |

---

## 一、作品管理页（剧本主编辑页）

洛水更新日志原文：
> 「作品管理页支持左侧章节管理、右侧正文编辑，可修改章名与内容、新建章节、单章导出。」
> 「工具栏按钮（导出本章 / AI缩写 / AI生剧本 / AI分镜 / 版本择优 / 分镜脚本等）改为更紧凑的小号样式。」

| 项 | 洛水原版 | V88 现状 | 状态 |
|---|---|---|---|
| 左章节管理/右正文编辑 | ✅ | `ComicScreenplayStoryStage.jsx` + `ComicScreenplayAuthoringPanels.jsx:26-32`（新增章节/生成本章剧本/生成并保存章节计划） | ✅ |
| 改章名/内容/新建章节 | ✅ | 同上 | ✅ |
| 故事概要文案 | "研读剧本 · 故事概要" | 原「故事摘要」已改为「故事概要」（`ComicScreenplayStoryStage.jsx`）；生成数据链路 2026-11-19 补齐（见第十六节） | ✅ 已补（文案+数据链路） |
| 导出本章 | 工具栏按钮 | ✅ 已补：「章节与分集」面板工具栏「导出本章」（见第十三节） | ✅ 已补（见第十三节） |
| AI 缩写 | 工具栏按钮 | ✅ 取证结论：洛水**运行时无此功能**，不实现（见第十四节） | ✅ 已核实（不适用） |
| AI 生剧本 | 工具栏按钮 | 「生成分集剧本」（`StoryStage.jsx:30-36`）、「生成本章剧本」 | ✅ 语义对齐，位置需挪到工具栏 |
| AI 分镜 | 工具栏按钮 | 「分镜脚本」弹窗（`StoryboardStage.jsx:40`） | ✅ 语义对齐 |
| 版本择优 | 工具栏按钮 | ✅ 已补：「当前剧本」面板标题栏「版本择优」，多版本横排对照（见第十四节） | ✅ 已补（见第十四节） |
| 分镜脚本 | 工具栏按钮 | ✅ | ✅ |
| 剧本医生/分镜医生 | 可审核产出（更新日志原文） | ✅ 已补：「当前剧本」面板「剧本医生」+ 分镜页工具栏「分镜医生」，洛水 review 报告规范（见第十五节） | ✅ 已补（见第十五节） |

---

## 二、AI 分镜弹窗（本次重点，洛水 `wasb-*` DOM 实证）

洛水原版弹窗布局（bundle `wasb-*` class + 官方文案实证）：

```
┌─ AI 分镜 ────────────────────────────────┐
│ 推理模型 [选择模型 ▾]  分镜词(剧本核心)[选择提示词 ▾] │
│ ☑ 规划小节  （默认开，只出规划不进分镜JSON）        │
│   规划小节密度：更紧凑/默认/更细拆/SD2.5专项        │
│ 规划文件 [不指定(各集用自己的规划，没有则不带) ▾]   │
│ 分镜上下文 [开关]  （1-1→1-2 顺序 + 最近5条历史）   │
│ 分镜前缀 [____]  额外内容 [____]                  │
│ ─ 分镜脚本生成有两种方式 ─                        │
│  1.Agent分镜  2.内置批量(先规划再分镜)             │
│ [开始规划(N集)]   [Agent 分镜]                    │
└──────────────────────────────────────────┘
```

| 序号 | 洛水原版控件 | 洛水原文/证据 | V88 现状 | 状态 |
|---|---|---|---|---|
| 2.1 | 推理模型选择（弹窗内） | "推理模型 / 选择模型" | 弹窗顶部配置行**只读展示**当前模型 + "页头更换"提示（`Prep.jsx:73` 配置行）；模型是项目级且必须规划/分镜同源，不做弹窗内临时下拉（避免两阶段用不同模型破坏链路） | ✅ 只读对齐（不做假按钮） |
| 2.2 | 分镜词（剧本核心）选择（弹窗内） | "分镜词（剧本核心）/ 选择提示词" | 弹窗顶部配置行**只读展示**当前分镜词 + "设置里更换"提示（`Prep.jsx:73`）；后端规划请求不收 prompt_id，分镜词是项目级配置 | ✅ 只读对齐（不做假按钮） |
| 2.3 | 规划小节开关（默认开） | "默认开启。只生成详细分节规划…不会自动进入分镜 JSON" | 弹窗新增原版文案的「规划小节」真开关，与 V88 原有「只生成规划/生成分镜JSON」任务状态双向同步（`Prep.jsx`） | ✅ 已补 |
| 2.4 | 规划小节密度 | "更少/标准/更多/SD2.5专项" | `StoryboardDensity`（`Prep.jsx`，`Settings.jsx:9`） | ✅ |
| 2.5 | 规划文件下拉 | "选择已有规划后按规划生成分镜 / 不指定（各集用自己的规划，没有则不带）" | 「使用已有规划」+下拉（`Prep.jsx`） | ✅ 基本对齐 |
| 2.6 | 分镜上下文开关 | "开启后一集内按 1-1→1-2 顺序生成，并把最近5条历史分镜传给下一场"；**原版在分镜页工具栏**（见截图 `Clipboard_Screenshot-1.png`）也有该开关 | 两处真开关：① 分镜页**工具栏** `Stage.jsx` `.comic-sb-toolbar-context`；② 弹窗顶部配置行 `Prep.jsx`。均走 `settingsRef.saveConfig({storyboard_use_context_history})` 即时保存 | ✅ 已补（原版工具栏位置 + 弹窗两处） |
| 2.7 | 分镜前缀 | "分镜前缀 / 严格按照规划文档内容执行分镜脚本小节。" | 弹窗内原版位置补真文本框，默认带出账号已选前缀；`StoryboardRequest.Prefix` 承接本次覆盖值并纳入幂等指纹（`storyboard_service.go`、`screenplay_routes.go`） | ✅ 已补 |
| 2.8 | 额外内容 | "额外内容 / 加在分镜提示词末尾的补充说明…" | 弹窗内补真文本框；`StoryboardRequest.ExtraContent` 追加为 `【额外内容】` 段落，承接在 V88 自己的分镜提示词末尾（`storyboard_scenes.go`） | ✅ 已补 |
| 2.9 | 开始规划/开始分镜（N集） | "开始规划（N集）/ 开始分镜（N集）" | 按钮保留稳定文案并补 `（1集）` 提示；单集窗口按当前集处理，不伪造批量 | ✅ |
| 2.10 | Agent 按钮 | "点右下角 Agent 分镜按钮" | 「Agent 分镜 · 打开助手」（`Prep.jsx:71`） | ✅ |
| 2.11 | 规划续写 | "上一轮规划还在等确认…决定下一步" | 本次补：partial→「确认续写规划」（`Prep.jsx:68` + `is_continue/existing_plan`） | ✅ 本次已补 |

---

## 三、顶部工具栏（分镜编辑态）

洛水更新日志原文：
> 「顶部工具栏『换装』后新增『分镜合并』图标…最长合并时长 10s/15s/30s」
> 「『剧本创作·超长视频』：顶部工具栏新增『超长视频』图标…目标总时长 100s/200s/250s」
> 「片段重拍图标：20 帧时间轴拖拽框选…首尾关键帧」

| 项 | 洛水原版 | V88 现状 | 状态 |
|---|---|---|---|
| 当前文本模型显示 | 工具栏模型状态 | 分镜工具栏新增只读模型显示（`.comic-sb-toolbar-model`），更换仍走页头真实入口 | ✅ 已补 |
| 分镜合并 | 顶部工具栏图标 + 时长下拉 | ✅ 已补：工具栏「合并」（`StoryboardStage.jsx`）+ 独立弹窗 `ComicScreenplayStoryboardMerge.jsx`；时长 10/15/30 秒（默认 30）；后端 `POST …/shots/merge`（`storyboard_merge_batch.go`），`dry_run` 只预览不落库 | ✅ 已补（真功能，不调模型） |
| 超长视频 | 顶部工具栏图标 + 时长选择 | ✅ 已补：工具栏「超长视频」（`StoryboardStage.jsx`）+ 弹窗 `ComicScreenplayLongVideo.jsx`；目标 100/200/250/300 秒（默认 300），单组≤30s；后端 `POST …/long-video/plan|generate|status|combine`，逐组走账号已保存的视频模型，结果导入媒体库，ffmpeg 顺序拼接成片 | ✅ 已补（真功能，用 V88 视频通道） |
| 片段重拍 | 时间轴拖拽框选 + 首尾关键帧 | ✅ 已补：工具栏「片段重拍」+ 弹窗 `ComicScreenplayReshoot.jsx`（剧集/分镜选择、20 帧时间轴拖拽框选、生成时长 1-30 秒、提示词、进度、新片段预览、更新拼接）；后端 `POST …/reshoot/candidates|frames|generate|status|apply` | ✅ 已补（真功能，见第九节） |
| 换装 | 顶部工具栏 | ✅ 已补：工具栏「换装」（`合并` 前）+ 弹窗 `ComicScreenplayCostumeChange.jsx`（剧集、分镜范围 1-本集镜头数、筛选场景含「不限制」、原人物 → 替换为）；后端 `GET …/costume/options`、`POST …/costume/preview|apply`，保留人物顺序 | ✅ 已补（真功能，见第十节） |
| 画风选择 | — | 顶部 `ComicStylePicker`（`Page.jsx:649`） | ✅ V88 自有 |

---

## 四、分镜检查 / 技能流水线

洛水 bundle 实证：`storyboard_check` 技能
> "检查原文分镜与AI分镜的缺失/重复/越界，自动补齐缺失分镜并生成检查报告。输出 storyboard_check_report.json"

| 项 | 洛水原版 | V88 现状 | 状态 |
|---|---|---|---|
| 分镜检查（缺失/重复/越界） | `storyboard_check` 技能 | 「检查」按钮仅查遗漏人物/场景/道具（`StoryboardStage.jsx:40`；`AssetCheckDialog.jsx:43`） | ✅ 已补：工具栏「分镜检查」+ 弹窗与报告（见第十二节） |
| 检查报告 | `storyboard_check_report.json` | 自动生产里的质检记录（`ProductionPanel.jsx:8,40`） | ✅ 已补：主动检查 + 真实下载报告（见第十二节） |
| 剧本医生/分镜医生 | 审核产出 | ✅ 已补：两个主动审核弹窗，按 issues/severity/pass 出报告（见第十五节） | ✅ 已补（见第十五节） |

---

## 五、对齐优先级建议

**第一批（弹窗内，改动小、收益直接）：**
- 2.1/2.2 模型与分镜词 → 2026-11-18 已落地为弹窗顶部只读展示（项目级、不做假按钮）
- 2.6 分镜上下文开关 → 2026-11-18 已落地为真开关（走 settingsRef.saveConfig，即时保存）
- 2.7/2.8 分镜前缀与额外内容：2026-11-18 已落到弹窗原版位置，并由 V88 自己的 `StoryboardRequest` 承接，未改模型接口、未做假按钮

**第二批（工具栏，改动中等）：**
- 1.导出本章 ✅ 2026-11-19 已落地（见第十三节）；1.AI缩写 ✅ 2026-11-19 取证为洛水运行时不存在（见第十四节）；1.版本择优 ✅ 2026-11-19 已落地（见第十四节）；本节全部闭环
- 3.分镜合并 ✅ 2026-11-18 已落地（真功能，见第三节）；3.超长视频 ✅ 2026-11-18 已落地（见第八节）；3.片段重拍 ✅ 2026-11-18 已落地（见第九节）；3.换装 ✅ 2026-11-18 已落地（见第十节）

**第三批（大功能，需单独设计）：**
- 剧本医生/分镜医生、分镜检查报告 → 分镜检查与报告 ✅ 2026-11-19 已落地（见第十二节）；剧本医生/分镜医生 ✅ 2026-11-19 已落地（见第十五节）

至此洛水「漫剧创作」页面工具批全部闭环：分镜检查报告 / 导出本章 / 版本择优 / 片段重拍 / 换装 / 批量翻译 / 合并 / 超长视频 / 剧本医生·分镜医生均已按原版规格落地；AI 缩写取证为洛水运行时不存在（不实现）。

**不照搬项**（洛水本机 CLI 引擎、动漫权限闸门、"配音开发中"占位）——非对齐必需。

---

## 六、2026-11-18 晚间整页可见元素核对（bundle 逐词取证）

> 目的：用户反馈"页面好像没变化"，对剧本页、分镜页**整页可见元素**逐词核对洛水最新版 bundle，确认真实剩余差异。

### 剧本页（洛水截图对照 + bundle 取证）

| 元素 | 洛水原文 | V88 现状 | 结论 |
|---|---|---|---|
| 分集剧本按钮 | 「新增一集」（bundle 3 hits，另「新一集」仅作无集数默认标题） | 「新增一集」 | ✅ 一致，不改 |
| 展开控制 | 「展开全部 / 收起全部」 | 同款 | ✅ |
| 导入剧本 / 分集剧本 / 故事概要 / AI 推理分析 / 下一步 | 均有 | 均有同款文案与样式（卡片、大字"共 N 集"、悬浮胶囊） | ✅ |
| 「结论全文」「帮我看懂啦」「剧本整理员」「30图集」 | **bundle 0 hits，最新版不存在** | 无 | ❎ 无需对齐（此前为截图 OCR 误读，从待办中划除） |
| 「补光」 | bundle 0 hits，不存在 | 无 | ❎ 无需对齐 |
| 「翻译」 | 实为「批量翻译」（分镜提示词 zh2en，44 hits） | ✅ 已补：工具栏「批量翻译」+ 弹窗 `ComicScreenplayBatchTranslate.jsx`；后端 `POST …/storyboards/{engine}/{episode}/translate/prepare|item` | ✅ 已补（真功能，见第十一节） |

### 分镜页工具栏（洛水更新日志原文实证）

洛水官方更新日志（bundle changelog）：
> 「分镜顶栏文案缩短：**预设前缀**改成**前缀**，**分镜上下文**改成**上下文**，**规划密度**改成**规划设置**，**资产检查**改成**检查**，**分镜失败？**改成**失败**」

| 文案 | V88 工具栏 | 结论 |
|---|---|---|
| 上下文 / 规划设置 / 检查 / 失败 | 已用同款短文案（`Stage.jsx`，上下文开关 `.comic-sb-toolbar-context`） | ✅ 已对齐 |

### 真实剩余可见差异（全部需后端能力，按"不做假按钮"不放）

- 分镜工具栏：换装已于 2026-11-18 落地（见第十节）；片段重拍见第九节
- 分镜批量翻译 ✅ 2026-11-18 已落地（见第十一节）
- 分镜检查完整报告（storyboard_check）✅ 2026-11-19 已落地（见第十二节）

### 结论

整页可见层面（文案+布局骨架）V88 与洛水已基本一致；上一轮改动集中在分镜弹窗内部，因此整页"看起来没变化"符合事实。剩余差异均为洛水特有功能按钮，需要后端能力承接，不做假按钮。

---

## 七、分镜合并落地记录（2026-11-18）

洛水更新日志原文（bundle changelog，2026-10-29）：
> 「顶部工具栏『换装』后新增 **分镜合并** 图标：按分镜顺序把相邻分镜打包成组，每组推荐时长之和不超过所选最长时长，每组只保留第一个分镜，其余并入后删除。」
> 「最长合并时长下拉可选 **10s / 15s / 30s**，默认 **30s**。例如分镜推荐时长为 15s、5s、9s、15s，选 30s 时前三个（共 29s）合并为一个分镜，第四个单独保留。」
> 「合并后保留分镜的推荐时长为各分镜之和，视频提示词、剧本台词按顺序拼接；绑定的角色、场景、道具、站位、白模视频取并集；已生成的图片 / 视频素材也一并保留到合并后的分镜。」
> 「弹窗实时预览合并结果（合并后剩余分镜数、减少个数）…全部分镜都没有推荐时长时不可合并。」

V88 落地：

| 层 | 实现 |
|---|---|
| 分组算法 | `PlanStoryboardMerge`（`storyboard_merge_batch.go`）贪心打包，0 时长分镜作为断点不跨越；全部无时长返回 `ErrInvalidState` |
| 执行 | `MergeStoryboardShotsByDuration`：时长求和、提示词/台词按序拼接、角色/场景/道具求并集、`merged_shots` 记原始成员、媒体 `ShotID` 重挂到保留分镜 |
| 接口 | `POST /api/comic-screenplays/projects/{id}/storyboards/{engine}/{episode}/shots/merge`，body `{max_seconds,dry_run}`；`dry_run` 只返回 plan 不落库 |
| 前端 | 工具栏「合并」按钮 + `ComicScreenplayStoryboardMerge.jsx`（剧集选择、10/15/30 秒、实时预览、确认执行） |
| 模型 | **不调用任何模型接口**，纯本地整理 |

测试：`storyboard_merge_batch_test.go` 8 项（含洛水 15/5/9/15→29s 举例）、`screenplay_routes_test.go` HTTP 1 项、`comic-screenplay-storyboard-merge.browser.test.js` 2 项，全部通过。

---

## 八、超长视频落地记录（2026-11-18）

洛水更新日志原文（bundle changelog，2026-10-29）：
> 「顶部工具栏新增 **超长视频** 图标：打开左右分栏的新窗口，目标总时长可选 **100s / 200s / 250s / 300s**。」
> 「按推荐时长累积分镜自动分组，标注为 **超长视频分镜\*1、\*2…**：每组不超过单次生成上限（约 30s）…全部组累计不超过所选目标（选 300s 也可能实际生成 280s，不会超）。」
> 「分组内自动合并各分镜的角色、场景、道具、站位、白模视频绑定，视频提示词按顺序拼接，每组提示词可在窗口中单独修改。」
> 「支持『批量生成』一键跑全部组，也可每组单独生成 / 重新生成；左侧列表带筛选（全部 / 待生成 / 生成中 / 已完成 / 失败），每组实时进度条；右侧『过程视频』可播放任一组的生成结果。」
> 「全部完成后自动（也可手动点『合并成片』）由后台把所有过程视频按顺序拼接成一个完整视频…保存到本地并写回分镜数据库。」

V88 落地：

| 层 | 实现 |
|---|---|
| 分组算法 | `PlanLongVideoSections`（`long_video.go`）：按推荐时长贪心累积，单组 ≤30s、累计 ≤目标；跳过已标注 `超长视频分镜*N` 的分镜；组内提示词按序拼接，角色/场景/道具去重合并 |
| 持久化 | `MediaState.longVideos`（按集存 `LongVideoBatch`）：分组、提示词、供应商任务号、结果素材、成片状态，关窗或重启后可按任务号继续 |
| 生成 | `POST …/long-video/generate`：复用账号在设置中保存的视频模型（`scriptVideoService`），每组一次提交，参考图取组内各分镜已选图（≤4 张），结果导入媒体库（`kind=video`） |
| 状态 | `GET …/long-video/status?episode=`：返回分组进度、过程视频地址、成片状态；失败组保留原因，可单组重试 |
| 合并成片 | `POST …/long-video/combine`：按分组顺序读取素材、`mergeworker.FFmpegRunner.Merge` 统一分辨率帧率拼接，成片流式存入素材库并回填 `final_item_id` |
| 前端 | 工具栏「超长视频」+ `ComicScreenplayLongVideo.jsx`（左右分栏、目标时长、筛选、每组提示词可改、进度条、过程视频、合并成片） |

测试：`long_video_test.go` 6 项（洛水 15/5/9/15→29s、目标上限、超长单镜裁剪、跳过已分组、字段合并、选项归一）、`screenplay_long_video_test.go` 3 项（分组预览、提示词与模型下传、重启保留改写）、`comic-screenplay-long-video.browser.test.js` 1 项，全部通过。

差异说明：洛水把成片回写分镜数据库的 `video_path`；V88 先落到媒体库（可在「媒体与剪辑」直接使用并合成），暂不回写分镜主视频，避免覆盖已生成的镜头视频。

---

## 说明

- 本清单严格按洛水最新版 bundle/更新日志原文取证，V88 侧逐条核到源码行号。
- 所有补齐一律复用 V88 现有后端接口与数据模型，遵守"源码优先、本地优先、可追溯 SHA"项目规则。
- "形态不同/位置不同"项：能力已在 V88 代码里，是否要把控件物理挪到洛水位置，按用户"同一个页面、同一个操作状态"要求逐项决定。


---

## 九、片段重拍落地记录（2026-11-18）

洛水更新日志原文（bundle changelog，2026-10-29，`v2026-10-29` →「剧本创作 · 片段重拍」）：
> 「播放器下方工具栏、字幕图标前新增 **片段重拍** 图标：打开剪辑风格窗口，在 20 帧时间轴上拖拽框选当前视频中的任意片段进行重拍。」
> 「提示词支持 @ 绑定人物 / 场景 / 道具 / 位置参考；生成时长默认为截取片段时长取整，可自行修改。」
> 「两种重拍类型可选：**首尾关键帧** 自动抽取片段第一帧 / 最后一帧，并按『图片1是第一帧起始画面、图片2是最后一帧画面』前缀组合提示词发给 AI；**视频参考** 不抽帧，把截取片段作为白模视频发给 AI 做视频参考。」
> 「生成过程显示进度条，关掉窗口后台继续跑；完成后可在窗口内直接播放结果…点击『更新』把新片段拼接回原位置，生成完整新视频保存到本地，同步更新当前分镜数据库并重新抽取封面地址。」

V88 落地：

| 层 | 实现 |
|---|---|
| 时间轴 | `POST …/reshoot/frames`：ffmpeg `fps=20/时长` 抽 20 帧（480px JPEG，base64 返回），返回真实时长；前端按帧铺满时间轴，两侧手柄拖拽框选，点击定位播放头 |
| 生成 | `POST …/reshoot/generate`：ffmpeg 在 `start` 与 `end-0.05s` 各抽一帧，写入对象存储并签名成模型可访问地址；提示词按洛水前缀 `@图片1是视频的第一帧起始画面，@图片2是视频的最后一帧画面` 组合，`image_to_video` 提交到账号已保存的视频模型（复用 `scriptVideoService`），时长默认取框选长度四舍五入、限制 1-30 秒 |
| 状态 | `GET …/reshoot/status?shot_id=`：按供应商任务号续查并导入素材库；关窗/重启不重复提交，失败保留原因 |
| 更新 | `POST …/reshoot/apply`：ffmpeg 把「原前段 + 新片段 + 原后段」拼成新视频（统一画布 30fps、缺音轨补静音），流式保存为这条分镜的 `kind=video` 素材，不覆盖原视频 |
| 前端 | 工具栏「片段重拍」+ `ComicScreenplayReshoot.jsx`：剧集/分镜选择（只列已生成镜头视频）、20 帧时间轴、拖拽框选、生成时长、提示词、进度条、新片段预览、重新生成/更新 |
| 不做假按钮 | 洛水的「视频参考」需要视频模型接收参考视频，V88 视频通道只支持图片参考，因此只提供「首尾帧参考」并明示原因；@ 引用未接入时不放入输入框占位文案 |
| 入口差异 | 洛水入口在镜头视频播放器工具栏；V88 漫剧页没有该播放器工具栏，入口放在分镜工具栏，打开后按「剧集 + 分镜」选择，实际操作对象与洛水一致 |

测试：`reshoot_test.go` 5 项（范围校验、时长取整、提示词前缀、前后段计划、状态读写）；`ffmpeg_reshape_test.go` 3 项（拼接 filter 段数、真实 ffmpeg 抽 20 帧 + 前段/新片段/后段拼接时长、单帧抽取）；`screenplay_reshoot_test.go` 4 项（候选只列有视频的分镜、20 帧抽取、提交首尾帧+模型+时长、拼接保存为分镜视频）；`comic-screenplay-reshoot.browser.test.js` 1 项（20 帧渲染、拖拽框选 3.0→6.0 秒、提交参数、更新与成片预览）。


---

## 十、换装落地记录（2026-11-18）

洛水原文（bundle changelog `v2026-08-18` →「剧本分镜 · 换装」，2026-09-05 版补充范围与场景）：
> 「左侧资产管理『复制资产』旁新增 **换装**：按场景筛选分镜，把绑定的人物 A 批量换成 A-睡衣或其他人物资产。」
> 「例如筛选『卧室』，将『张三』换成『张三-睡衣』：只改该场景且已绑定张三的分镜；学校分镜、未绑定张三的分镜、已经是张三-睡衣的分镜都不会动。」
> 「替换时保留人物在分镜中的顺序，方便 `@图片N` 仍对得上；进度条显示更新过程，完成后写回数据库。」
> 2026-09-05：「换装窗口新增 **分镜范围**；筛选场景可选手动指定或 **不限制**，例如只改 1-30 镜、不限场景。」

V88 落地：

| 层 | 实现 |
|---|---|
| 规则 | `backend/internal/screenplay/costume_change.go`：`PlanCostumeChange` 先按 seq 闭区间圈范围，再按场景名筛选，命中条件为「已绑定原人物且还不是目标人物」；`ReplaceCostumeName` 按词元整体替换并保留顺序（`、，；/\|` 均识别，输出统一 `、`），已在目标人物里的分镜整条不动 |
| 选项 | `GET …/costume/options`：人物、场景取自资产库（`character` / `scene`），再补上本集分镜里实际出现的名字 |
| 预览 | `POST …/costume/preview`：返回命中条数与逐条 `before → after`，不写库；无命中时返回洛水同款文案 |
| 执行 | `POST …/costume/apply`：整集一次性保存（`saveEditedStoryboard`），返回 `message` 与命中明细 |
| 前端 | 工具栏「换装」（位于「合并」前，与洛水顶栏顺序一致）+ `ComicScreenplayCostumeChange.jsx`：剧集、分镜范围（默认 1-本集镜头数）、筛选场景（含「不限制场景」）、原人物、替换为、命中预览列表、执行提示 |
| 差异说明 | 洛水逐条写库带进度条与取消；V88 分镜整集是一个聚合，一次保存即可完成，因此不摆「取消」假按钮，执行中显示真实等待状态；洛水改 `character_ids`，V88 分镜存人物名，按词元替换等价于「换人且保持 @图片N 顺序」 |

测试：`costume_change_test.go` 8 项（洛水卧室例子、不限制场景、已是目标人物跳过、顺序与分隔符、输入校验、范围收敛、提示文案、落库与无命中不写库）；`screenplay_costume_routes_test.go` 1 项（options / preview 只读 / apply 写库 / 同名 400 / 无命中非 200 / 越权 404）；`comic-screenplay-costume-change.browser.test.js` 2 项（默认范围 1-6、选人物与场景、预览参数与命中列表、执行调用与关闭；只读禁止执行）。


---

## 十一、批量翻译落地记录（2026-11-18）

洛水原文（bundle「批量操作」弹窗翻译页 + 运行时代码取证）：
> 批量操作卡有 生图 / 视频 / 配音 / 翻译 四个标签页；翻译页控件为「类型」（画面提示词 / 视频提示词）、「翻译方向」（中文 → 英文 / 英文 → 中文）、「范围」（全部 / 未完成 / 未锁定 / 已锁定 / 自定义范围 + 起止分镜）。
> 文本取所选类型提示词（positive_prompt / video_prompt），为空时回退字幕文本；逐条调用 `batch_translate_prompts` 翻译后回写原字段；进度提示「正在翻译 画面提示词（中→英）：X/Y」；完成后「翻译完成：成功 X/Y」（有失败时警告）。

V88 落地：

| 层 | 实现 |
|---|---|
| 范围 | `screenplay/batch_translate.go`：`PlanBatchTranslate` 按洛水规则筛选——全部 / 未完成（= 该字段提示词非空，有文本才可翻译）/ 自定义 seq 闭区间；文本取 `ImagePrompt` / `VideoDesc`，为空回退 `Line`（字幕），并标记 `from_subtitle` |
| 翻译 | `TranslateStoryboardShot` 逐条调用当前作品文本模型（`resolver` → `OpenAICompatibleProvider`），系统提示词要求逐句对应、保留换行与括号结构、专有名词不译、只输出译文；写回原字段整集保存 |
| 路由 | `POST …/storyboards/{engine}/{episode}/translate/prepare`（只读，返回待翻译条目与「没有需要翻译的分镜」）、`POST …/translate/item`（翻译一条并写库，前端逐条调用，还原洛水的逐条进度） |
| 前端 | 工具栏「批量翻译」（「批量」之后）+ `ComicScreenplayBatchTranslate.jsx`：剧集、类型、翻译方向（中文 → 英文 / 英文 → 中文）、范围（全部 / 未完成 / 自定义范围 + 起止分镜 + 共 N 个分镜）、进度「正在翻译 画面提示词（中→英）：X/Y」、失败明细、成功汇总 |
| 差异说明 | ① 洛水范围含「未锁定 / 已锁定」，V88 分镜没有锁定概念，不做假选项（弹窗内明示）；② 洛水翻译在批量操作弹窗的翻译页，V88 批量弹窗是集级规划/生成，语义不同，入口放分镜工具栏；③ 洛水的翻译模型为其后端配置，V88 用页头当前作品文本模型，未选模型时禁止执行并提示 |

测试：`batch_translate_test.go` 8 项（全部/未完成/视频字段/自定义范围/参数校验/系统提示词方向/写回与 revision/空翻译与越界报错）；`screenplay_translate_routes_test.go` 1 项（prepare 只读、字幕回退标记、item 按条翻译写回且不串字段、空范围提示、非法字段 4xx、越权 404）；`comic-screenplay-batch-translate.browser.test.js` 2 项（类型/方向/范围选择、prepare 与逐条 item 调用顺序、成功汇总、onChanged；未选模型与只读禁止执行）。

---

## 十二、分镜检查落地记录（2026-11-19）

洛水 bundle 原文（`storyboard_check` 技能，与「台词检查」等并列于技能流水线）：
> 「检查原文分镜与AI分镜的缺失/重复/越界，自动补齐缺失分镜并生成检查报告。」
> details：「输入：原文序号行 + 当前storyboard分镜列表」「检查：缺失分镜、重复分镜、越界分镜、异常ID」「修复：缺失分镜自动调用AI补全并插入/更新」「输出：storyboard_check_report.json 检查报告」

洛水运行时 `Bt()` 取证：原文按 `/^(\d+)[.、．]\s*/gm` 提取序号行，与 AI 结果 `scene_id` 对照得 missing / extra（越界）/ duplicates，报告结构 `{total, originalTotal, errors, missing, extra, duplicates, minNum, maxNum, passed}`；弹窗展示【缺失的分镜】原文有但AI结果没有、【多余的分镜】AI结果有但原文没有、【重复的分镜】AI结果中重复出现，缺失时提示「发现 N 个缺失分镜，建议重新处理对应批次」。

V88 落地：

| 层 | 实现 |
|---|---|
| 检查算法 | `backend/internal/screenplay/storyboard_check.go`：`CheckStoryboardShots` 检查序号完整性——missing = 1..最大序号 中缺席（跳号）、duplicates = 同一 seq 出现多次、invalid = seq ≤ 0（异常ID）、passed = 三者皆空；报告结构对齐 `storyboard_check_report.json`（snake_case：total / expected_total / min_seq / max_seq / missing / extra / duplicates / invalid / passed） |
| 差异明示 | 洛水按「原文序号行」对照，extra = AI 有但原文没有（越界）；V88 分镜由规划链生成、seq 即镜头序号，无原文序号行对照基准，因此 `extra` 恒为空（保留字段对齐报告结构），检查基准为 1..最大序号 连续性，弹窗 hint 原文明示 |
| AI 补齐 | `FillMissingStoryboardShot`：只允许补当前报告确实缺失的序号（已存在/超范围序号拒绝且不调模型）；用户提示词带前/后分镜上下文（视频/画面提示词、人物、场景、台词），系统提示词要求只输出 JSON；解析容忍代码块包裹（首 `{` 到末 `}`）；新分镜 `newScreenplayID("sb_shot")` 按 seq 插入排序位置，时长取前一分镜推荐时长（缺省 5 秒），整集一次保存 |
| 路由 | `POST …/storyboards/{engine}/{episode}/check`（只读报告）、`POST …/check/fill`（body `{seq, model}`，补一条写库，前端逐条调用还原洛水逐批进度） |
| 前端 | 工具栏「分镜检查」（「检查」（资产检查）前）+ `ComicScreenplayStoryboardCheck.jsx`：剧集选择、开始检查、统计条（分镜数/序号范围/缺失/重复/异常ID）、洛水同款【缺失的分镜】【重复的分镜】【异常ID】分区与「发现 N 个缺失分镜」提示、AI 补齐逐条进度与「分镜补全完成：已补全 X/Y」汇总、**下载报告 storyboard_check_report.json**（Blob 真实下载） |
| 报告下载 | 洛水把 `storyboard_check_report.json` 写到作品文件夹；V88 弹窗提供真实 Blob 下载同名文件，内容为 `{storyboard_check_report: 报告}` |

测试：`storyboard_check_test.go` 8 项（缺失/重复/异常统计、连续通过、空分镜报错、补全提示词上下文、系统提示词、JSON 解析含代码块包裹、补齐插入中间并写库且补后通过、非缺失序号拒绝且不调模型、缺省时长与去空白）；`screenplay_check_routes_test.go` 1 项（check 只读不返回整集、fill 模型输入与写库顺序 1/2/3/4、补后 passed、已存在序号拒绝、缺 seq 4xx、越权 404）；`comic-screenplay-storyboard-check.browser.test.js` 2 项（检查调用与报告、缺失分区与补齐逐条调用、汇总、onChanged、补后重查通过；未选模型禁止补齐），全部通过。

本地 4001 验证：entry `/assets/user-Dfmbfaf4.js` → `ComicScreenplayPage-jKTCbDO0.js`（含 缺失的分镜、AI 补齐缺失分镜、分镜补全完成、storyboard_check_report.json、检查通过：序号连续、异常ID）；`check` POST 401 / GET 405、`check/fill` POST 401 / GET 405（已注册非 404）；ffmpeg 可用。启动前清掉了占用 4001 的外部遗留进程 `qiantie-toolbar-bottom.exe`（novel-toolbar-bottom-20261119）。

---

## 十三、导出本章落地记录（2026-11-19）

洛水取证（bundle 运行时 + `modules/works.pyd` 字符串）：
> 更新日志原文：「作品管理页支持左侧章节管理、右侧正文编辑，可修改章名与内容、新建章节、**单章导出**。」「工具栏按钮（**导出本章** / AI缩写 / AI生剧本 / AI分镜 / 版本择优 / 分镜脚本等）改为更紧凑的小号样式。」

| 入口/行为 | 洛水证据 | 说明 |
|---|---|---|
| 正文工具栏按钮 | 小号按钮「导出本章」`disabled:!ge.value`，onClick `ju()` | ge=当前章节有内容，无内容禁用 |
| 章节菜单项 | 章节下拉 `{重命名 / 导出本章(key=export) / 删除}`，`"export"===e&&await ju(t)` | 与「重命名 / 删除」并列 |
| 导出实现 | `window.pywebview.api.export_work_chapter(workId, chapterId)` → `{success, path}`，前端提示「已导出：${path}」 | `works.pyd` `WorksApi.export_chapter`：另存为对话框 `Text files (*.txt)`、文件名默认章节标题（非法路径字符 `[\\/:*?\"<>|]` 替换）、utf-8 写入、`「导出单章到用户选择的位置。」` |
| 剧本边界 | `if(io.value) return o.info("剧本导出请使用复制内容或后续导出功能")` | 本章导出**只导章节原文**；剧本另有 `export_scripts_txt`（选集数合并导出一个 TXT，merged/separate 两种模式） |

V88 落地：

| 层 | 实现 |
|---|---|
| 入口 | 「章节与分集」面板（`ComicChaptersPanel`，对应洛水作品管理页）工具栏新增「导出本章」按钮（下载图标，置于「新增章节」前），导出当前选中章节 |
| 导出 | `comic-screenplay-chapter-export.js`：内容 = 章节**原文**（`source_text`，与洛水「正文编辑」同位；本章剧本不走本章导出，按钮 title 明示「剧本用复制或剧本合集导出」），文件名 = 章节标题（空标题回退 `第N章`，非法字符替换为 `_`）+ `.txt`，Blob UTF-8 下载；空原文禁用按钮并提示 |
| 成功提示 | 洛水同款「已导出：文件名」 |
| 等价差异 | 洛水弹「另存为」系统对话框；V88 是 Web 端，浏览器下载即等价（真实下载非假按钮） |

测试：`comic-screenplay-chapter-export.test.js` 2 项（文件名规则：标题/空标题回退/非法字符替换；内容与注入保存回调；空原文拒绝并提示）；`comic-screenplay-chapter-export.browser.test.js` 1 项（面板渲染、有内容章节导出按钮可用、点击后 Blob 内容为章节原文、成功提示「已导出：第一章 相遇.txt」、切到空章节按钮禁用）。

---

## 十四、AI 缩写（不适用）与版本择优（2026-11-19 落地）

### AI 缩写：洛水运行时不存在，不做假按钮

取证过程（2026-11-19，`index-C1aA08SG.js` / `index-DgTjggsH.js` / `index-DszfAkKD.js` 三份 bundle + `modules/*.pyd`）：

- 「AI缩写」在所有 bundle 中**只出现 1 次**，位于 v2026-07-22 更新日志文本（「工具栏按钮（导出本章 / AI缩写 / AI生剧本 / AI分镜 / 版本择优 / 分镜脚本等）改为更紧凑的小号样式」），不是运行时按钮。
- 当前运行时作品页工具栏实际渲染：小说 Tab = `导出本章 + AI生剧本`；剧本 Tab = `文稿对比 / 分镜脚本 / AI分镜 / 导入 / 导出`。均无「AI缩写」。
- `works.pyd` 完整 API 名清单已核对：`export_chapter / export_scripts_txt / create_chapter / create_script / get_script_polish / get_script_version_merge` 等，**无任何 abridge / 缩写 / shorten 类 API**。
- 结论：AI 缩写为更新日志残留（v2026-07-22 期间存在或计划，随后被移除/未发布）。按「真支持才做」原则不实现，避免做出与洛水行为对不上的假按钮。

### 版本择优：洛水规格与 V88 落地

洛水运行时取证（`wsvp-modal` 组件，@2418202 起）：

- 弹窗标题「版本择优」副题「按集分组横排对照，选用更好的一稿」；集分组 chips（`N版`）。
- 每列 `wsvp-col`：`首选` 标记（preferred）、`版本 N`、`字数 · updatedAt`、标题、**完整正文**、按钮 `选用此版`（primary，把所选版 preferred 置 true 其余 false，成功提示「已选用：<标题>」）与 `删除该版`（确认删除，不可恢复）。
- 数据源：`get_work_script(workId, scriptId)` 逐版取正文；`update_work_script({preferred})` 写首选。
- 空态：「还没有可对照的剧本版本，先 AI 生成或导入后再来」。

V88 落地（真功能）：

| 层 | 实现 |
|---|---|
| 后端领域 | `screenplay.DeleteDocument`：当前活动版本拒绝（「当前版本不可删除，请先选用其他版本」）；历史版本连带删除其 shots/characters/scenes |
| 后端 Store | `Store.DeleteDocument` 接口 + MemoryStore（内存清理）+ MySQLStore（事务删 screenplay_shots/characters/scenes/documents，0 行回滚报 ErrNotFound） |
| 后端路由 | `GET /api/comic-screenplays/projects/{id}/documents/{documentID}`（单版完整正文，列表接口不返回正文）；`DELETE …/documents/{documentID}`（204；越权 404） |
| 前端 | `ComicScreenplayVersionCompare.jsx`：「当前剧本」面板标题栏「版本择优」按钮 → 弹窗横排对照列（`wsvp-col` 对齐洛水 class）：`当前版本` 标记（对应洛水「首选」）、`字数 · 时间`、完整正文滚动预览、`选用此版`（= 恢复为新当前版本，成功提示「已选用：版本 vN」）、`删除该版`（Popconfirm 确认）；空态文案同洛水 |
| 差异明示 | 洛水按集分组（每集一份剧本）；V88 剧本是整个项目一份文档，按版本横排对照（弹窗副题相应改写） |

测试：`document_delete_test.go` 5 项（active 拒绝 / 空 ID / 未知 ID / MySQL 删除 SQL 序列含回滚）；`screenplay_document_compare_routes_test.go` 1 项（GET 单版含正文、DELETE active 400、DELETE 历史 204 且连带清理、越权 404）；`comic-screenplay-version-compare.browser.test.js` 1 项（弹窗渲染、横排列、当前版本标记与按钮禁用、删除 Popconfirm → DELETE 调用、选用 → restore 调用 + onChanged）。

---

## 十五、剧本医生 / 分镜医生（2026-11-19 落地，工具批收官）

### 洛水取证

- bundle 与 `modules/*.pyd` 取证：医生是 Agent 审核技能（`doctor_prompt_name:"洛水超级短剧医生"`），审核产出并按 `luoshui_doctor_shared` review JSON 规范返回。
- review 规范还原：`{pass, issues:[{id,dimension,location,episode,severity,problem,suggestion,evidence}], suggested_actions:[...], summary, reviewed_at}`；severity 枚举 `blocker/major/minor`；`pass` 由是否含 blocker 决定；`suggested_actions` 为枚举动作列表。
- 剧本医生审核整份剧本全文的剧情逻辑/人物动机/资产引用/节奏推进/镜头缺失/台词违和；分镜医生对照剧本审核人物错位/场景不连贯/资产引用错误/视频提示词结构缺失；两者均**只出报告不改稿**。

### V88 落地（真功能）

| 层 | 实现 |
|---|---|
| 后端领域 | `screenplay/doctor.go`：`ReviewScreenplayDoctor`（当前活动文档全文）/ `ReviewStoryboardDoctor`（按集分镜对照剧本），复用 V88 文本模型通道；`NormalizeDoctorReview` 归一化模型输出：severity 白名单（未知降级 minor）、空 problem 丢弃、issue id 去重与补号、**pass 由服务端按 blocker 重新计算**（不信任模型自报） |
| 后端路由 | `POST /api/comic-screenplays/projects/{id}/doctor/screenplay`、`POST …/storyboards/{engine}/{episode}/doctor/storyboard`；入参 model，缺模型 4xx，越权 404 |
| 前端 | `ComicScreenplayDoctor.jsx` 双模式共用组件：「当前剧本」面板标题栏「剧本医生」按钮 + 分镜页工具栏「分镜医生」按钮（「分镜检查」后）；报告 = pass 成功/失败横幅 + 问题数分级统计 + issues 按 severity 着色（阻断/重大/优化）+ dimension/location/建议/原文 + suggested_actions + 审核时间 |
| 差异明示 | 洛水医生产出含 rewrite 建议并可回写改写；V88 严格按「只出报告不改稿」，不提供自动改写入口，避免越权改动用户稿件 |

测试：`doctor_test.go`（归一化：severity 白名单/去重/空丢弃/pass 重算；剧本医生与分镜医生空内容拒绝、正常 review）；`screenplay_doctor_routes_test.go`（两路由 POST 报告、缺模型 4xx、越权 404）；`comic-screenplay-doctor.browser.test.js` 2 项（剧本医生报告横幅/分级/建议/POST 调用；分镜医生剧集选择/POST 调用/无内容或只读禁用）。

---

## 十六、故事概要数据链路补齐（2026-11-19）

### 问题（用户对照截图发现）

洛水「AI 推理分析」剧本页顶部有「标题 + 题材标签（女频/现代言情/校园…）+ 故事摘要卡片」，V88 页面直接从「共 N 集」开始。根因不是缺 UI：`ComicScreenplayStoryStage.jsx` 早已照搬该展示位（`result?.title` / `result.audience` / `result.genre` / `result.tags` / `result.synopsis`），但全项目无任何地方生成/写入 `workspace.synopsis_result`——洛水这一整块是 `synopsis` Agent 的产物，V88 当初只做了展示壳，没做生成链路，`result` 恒为 null 故整块不渲染。第四节第 31 行「故事概要文案 ✅」仅指文案/UI，不代表数据链路。

### 洛水取证

- `screenplay.pyd` 的 `SYNOPSIS_AGENT_PROMPT`（完整提取）：「你是一个专业的故事摘要与项目信息提取助手……」；摘要 300-500 字，涵盖起因→发展→高潮→结局，保留核心人物姓名与关键事件；标题 2-10 字；`audience` 男频/女频二选一；`genre` 14 类枚举；`tags` 3-6 个、每个 2-4 字。
- 输出 JSON `{title, synopsis, audience, genre, tags}`；完成时写入 `Se.value.synopsis`（`agent_key==="synopsis"`），保存为 `synopsis_result.json`，剧本页据此渲染。

### V88 落地（真功能）

| 层 | 实现 |
|---|---|
| 后端领域 | `screenplay/synopsis.go`：`SynopsisResult` 结构 + 洛水同款 system prompt（原文提取）；`GenerateSynopsis` 读原文（超 60000 字截断）调文本模型；`NormalizeSynopsisResult` 归一化：audience 二选一（非法回退女频）、genre 白名单（非法回退其他）、tags 3-6 个 2-4 字去重、synopsis 空则报错（长度容错）；空原文/空模型拒绝 |
| 持久化 | `CreativeConfig.SynopsisResult` 字段（JSON envelope 存于 `screenplay_project_creative_configs.config`，免改 DB schema）；生成后经 `SaveCreativeConfig` 带 revision 乐观锁写回 |
| 后端路由 | `POST …/projects/{id}/synopsis/generate`（model 入参，max 4000 tokens，SingleAttempt）；workspace 顶层新增 `synopsis_result` 输出，前端直接可用 |
| 前端 | 「剧本设置」→「故事概要」label 旁「AI 生成」按钮（有原文可用、只读禁用）；调 API 后同步概要文本到表单 + 刷新 workspace；`ComicScreenplayStoryStage` 既有展示位直接吃到结构化结果：标题替换项目名、题材标签行、故事概要卡片；无结果时回退手动填的纯文本概要、不渲染标签行 |

测试：`synopsis_test.go` 6 项（结构归一化/回退/拒绝、空原文与空模型拒绝、长原文截断与模型调用、模型错误透传）；`screenplay_synopsis_routes_test.go` 1 项（生成成功+provider 收到洛水 prompt 与原文+写回 config+workspace 顶层输出、空原文拒绝、越权 404）；`comic-screenplay-story-stage.browser.test.js` 1 项（结构化标题/标签/摘要渲染、无结果回退）；`comicScreenplay.test.js` +1（生成 API 路径）。

差异明示：洛水由「AI 推理分析」流水线自动跑 synopsis Agent；V88 提供手动触发的「AI 生成」按钮（走 V88 自己的文本模型通道），不强制随剧本生成自动运行，避免额外模型费用；两者产物结构与展示一致。
