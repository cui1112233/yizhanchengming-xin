# 洛水最新版「剧本创作」实机取证 vs V88 剧本创作对照

- 日期：2026-11-18
- 洛水证据源（用户指定"最新"）：`F:\软件\洛水2026\洛水2026\`（安装目录，`launcher.log` 显示 2026-11-17 02:07 应用更新成功，当前 bundle `index-C1aA08SG.js` 17.4MB / `index-B5otKszC.css`，均 11-17 构建）
- 对比证据源：`F:\软件\洛水2026.zip`（2026-03-24 备份，`index-CnO3GE3B.js` 3.68MB）与旧对照文档 `docs/acceptance/luoshui-screenplay-creation-pipeline-2026-11-11.md`、`manju-vs-luoshui-screenplay-2026-11-11.md`
- 约束：全程只读。未启动洛水 GUI、未改安装目录任何文件、未调用任何生成接口、未消耗额度。
- 归档：zip 解压副本在 `_local-archive-20261028/luoshui-2026-zip/`（已排除约 300MB 浏览器缓存）。

## 0. 一句话结论

最新洛水剧本创作是一个**已经跑通真实项目、做到"先规划→分镜词→场次图片/视频绑定"的本地流水线**；V88 的 `/comic-screenplay` 已把同样的九步骨架和医生/版本恢复做进代码。2026-11-18 逐行复核后修正上轮口径：洛水的**核心能力在 V88 代码中均已落入实现**——规划续写后端（`existing_plan/continue`）已完整写好并带测试/golden（工作区未提交）、场次级图片/视频用独立 `MediaItem` 关联表打通、字幕在 AI-CUT 与剧本页媒体面板均有入口。差异集中在**前端续写参数未接线、媒体/字幕的数据归属形态不同**，属"实现形态差异"，不应定性为功能缺口。（2026-11-18 复核证据：`storyboard_plan_existing.go`、`storyboard_planning.go:39-42`、`ComicScreenplayStoryboardPrep.jsx:51`、`media_store.go`、`aicut/CaptionWorkspace.jsx`）

## 1. 本次新取证（旧文档"未取到"的东西，现已拿到）

### 1.1 真实剧本项目 `data/剧本创作/1212/`（2026-11-14 创建）

| 文件 | 说明 |
|---|---|
| `info.json` | `id=1794591482818, name=1212, script_mode=dialogue, workspace_mode=pro, editor_version=new` |
| `ai_config.json` | `model=doubao-seed-evolving-latest-version-官方`, `promptId=144/专业剧本`, 风格 `3D国漫1`, `aspectRatio=9:16`, `storyboardPlanFirst=true`, `selected_storyboard_prompt_id=307` |
| `model_config.json` | 生图 `luoshui_nano_banana`、配音 `minimax`/`index-tts2`、资产解析档 `master`、比例 `9:16`、2K、角色三视图提示词（213/289号） |
| `chapters.json` | 1 集（第1章），每集键：`id/title/content/scriptResult/storyboardFailedScenes/storyboardResult/scriptMode` |
| `chapter_0_storyboard.json` | `{storyboard:[...], plan:"# 本作品·第1章 分镜小节规划…"}` —— **规划与分镜词同文件** |
| `agent_assets.json` | `characters(4)/scenes(2)/props(0)/voices(0)`，角色字段 `name/prompt/type/identity/voice/颜值定位/出现集数`，场景字段含 `空间要点` |
| `ai_raw_responses/storyboard_master_ch1_20261115_041010.json` | `mode=storyboard_master_parallel`，并发生成 + 失败场次重试（`initial_failed_scene_count/final_failed_scene_count`） |
| `screenplay_ch0.db` | 项目级 SQLite：`characters(4)/props(1)/scenes(2)/shots(3)/voice_config(0)` |

### 1.2 分镜字段契约（黄金证据）

顶层键：`id / scene_id / duration_sec / characters / scene / props / video_desc`，外加内部序号 `_scene_index / _scene_number / _scene_name / _scene_local_order`。
`video_desc` 内每个镜头又细分：`【时间标识】【运镜】【画面】【动作/情绪】【位置承接】【OS】【画外音】【音效】【台词】`。
`duration_sec=12`（一节 3 镜），规划文本按"总节数/总时长/scene_id/资产绑定/入口承接"组织。

`screenplay_ch0.db.shots` 字段：`id,sort_order,script_text,character_ids,scene_id,scene_ids,prop_ids,location_reference,image_prompt,video_prompt,negative_prompt,image_path,image_collection,video_path` —— **场次级已有图片/视频产物字段**。
资产表字段：`bound_image_path / bound_audio_path / sora2_info`（资产绑定图/绑定音频/新模型信息）。

### 1.3 最新版 hint（bundle `index-C1aA08SG.js`）

- 前端"内容创作 → 剧本创作"、"漫剧解说"、"创建剧本"、个人/团队作品、AI 对话、工作台均在（完整客户端，`index.html` 标题"分镜编辑器"只是打包入口名）。
- 提示词分类：`剧本大师(高质量剧本创作与增强) / 中长篇转剧 / 剧本分镜(剧本到镜头的分镜输出) / 分镜(镜头拆解与分镜结构规划) / 爆款开头 / 旁白解说 / 对话剧情`。
- 桥命令（74 个剧本/分镜相关）含：`create_screenplay / save_screenplay* / screenplay_ai_inference / screenplay_master_inference / screenplay_agents_run / storyboard_ai_inference / screenplay_master_storyboard_inference`，媒体侧 `batch_generate_images / batch_generate_videos / batch_generate_narration_audio`，检查 `storyboard_check / storyboard_check_report`，确认 `screenplay_ai_confirm_split / screenplay_ai_confirm_media_generation`，字幕 `screenplay_generate_combined_subtitle`，导出 `export_storyboard_to_jianying / export_to_jianying`，定时 `save_screenplay_scheduled_tasks`。
- "配音功能开发中..." 占位仍在（旧文档 E42 结论未变）。
- app.db（11-14，9.25MB）提示词 84 条：`剧本分镜 19 / 视频推理 17 / 画面推理 10 / 中长篇转剧 7 / 小说转剧 4 / 剧本大师 2 / 爆款开头 2 …`（zip 备份同为 84 条，旧文档记录的 182 条来自更早版本）。

## 2. 与 V88 逐项对照

| 维度 | 洛水最新版（实机） | V88 现状（代码取证） | 结论 |
|---|---|---|---|
| 入口 | 内容创作→剧本创作 + 漫剧解说 | `/script`（私人定制）+ `/comic-screenplay`（漫剧创作）+ `/batch-factory`（水货生产） | 形态接近，V88 无统一"内容创作"分组 |
| 真实项目落地 | `data/剧本创作/1212/` 全套 JSON+SQLite+AI 原始响应 | MySQL 19 表 + screenplay 库，导出仅 JSON | V88 无标准作品目录，迁移/审计弱 |
| 先规划后写词 | `storyboardPlanFirst=true`，`chapter_0_storyboard.json` 内 `plan`+分镜词同文件 | `/comic-screenplay` 有 `chapters/plan` 与 GenerateStoryboardPlan；`/script` 无 | **已对齐**（`/comic-screenplay` 侧）；续写后端已写、前端未接线 |
| 规划续写 | 规划文本可续写（Agent 快捷指令 + plan 字段） | 后端 `existing_plan/continue` + continuation 生成 + 3 测试文件/golden 已完整写好（`storyboard_plan_existing.go:10`、`storyboard_planning.go:109-127`），工作区未提交；前端 `generateStoryboardPlan` 为 `JSON.stringify(input)` 透传（`comic-screenplay-storyboard-api.js:8`），只是调用点 `ComicScreenplayStoryboardPrep.jsx:51` 未传续写参数 | **代码已实现，差两步**：提交 Git + 前端传参（非重写功能） |
| 分镜字段 | 六字段 + 4 个场次序号字段；`video_desc` 内 9 类子字段 | 六字段（`id/scene_id/duration_sec/characters/scene/props/video_desc`） | V88 少场次序号/细粒度子字段 |
| 场次级媒体 | shots 表有 `image_path/image_collection/video_path` | `Shot` 无媒体字段，媒体在独立 `MediaItem`（Kind/URL/ShotID/TargetID）+ 打通的按镜头生成调用链（`screenplay_media_workbench.go`、`ComicScreenplayMediaPanel.jsx`） | **等价实现**：关联表替代内嵌字段，链路已通，仅数据形态不同 |
| 图片/视频批量 | `batch_generate_images/videos`，多引擎（即梦/小云雀/seedance/vidu/万相/nano_banana） | `/comic-screenplay` 已跑到镜头视频；`/script` 仅单镜头视频 | V88 部分具备 |
| 配音/字幕 | 批量 TTS + `screenplay_generate_combined_subtitle`；设置面板仍"开发中" | `/script` 原文 TTS；字幕在 AI-CUT `CaptionWorkspace.jsx` 精细编辑 + 剧本页媒体面板基础 ASR（提示可转 AI-CUT） | **已实现但归属不同**：洛水在剧本链内，V88 在 AI-CUT/媒体面板 |
| 医生/审核 | `storyboard_check(_report)`、剧本医生等 | `/comic-screenplay` 九步含 `2_screenplay_doctor/6_storyboard_doctor` | 已对齐 |
| 版本/多版本 | 有 AI 结果历史+原文对比 | `/comic-screenplay` 有版本历史与恢复；`/script` 仅历史列表无 diff | 部分对齐 |
| 爆款开头 | 提示词类型"爆款开头" | `mode=hook` 枚举值 | 弱对齐 |
| 权限闸门 | `check_dongman_permission` | 无 | 产品策略差异 |
| 挂机/定时 | `save_screenplay_scheduled_tasks`、取消挂机 | `/comic-screenplay` 有定时/暂停/重建 | 已对齐 |

## 3. 给用户的落地判断

1. **方向没错**：V88 的剧本创作骨架（剧本→资产→分镜→生产九步、医生、版本、定时）与洛水最新版同构，不是缺一整条链。
2. **核心能力均已落入代码，剩余是收尾而非重写**（2026-11-18 逐行复核后口径）：
   - 规划续写：后端 `existing_plan/continue` + continuation 生成 + 测试/golden 已完整写好（`storyboard_plan_existing.go`、`storyboard_planning.go:109-127`），**仅差**提交到 v88 + 前端调用点传参（`ComicScreenplayStoryboardPrep.jsx:51` 当前未传，但 API 层已透传）。
   - 场次级媒体：`MediaItem`（Kind/URL/ShotID/TargetID）+ 按镜头生成/保存的调用链已打通（`screenplay_media_workbench.go`、`ComicScreenplayMediaPanel.jsx`），是洛水 shots 内嵌字段的**等价实现**，数据形态不同。
   - 字幕/配音：AI-CUT `CaptionWorkspace.jsx` 精细编辑 + 剧本页媒体面板基础 ASR 均已实现，**归属与洛水不同**（洛水在剧本链内），非功能缺失。
3. **不要照搬**：洛水"配音功能开发中"占位、多引擎客户端 CLI 依赖、动漫权限闸门，均非必须对齐项。
4. **最有价值的下一步**：以真实项目 `1212` 为验收样本，把它的 `chapters.json` + `chapter_0_storyboard.json(含plan)` + `agent_assets.json` 作为输入，跑通 V88 的导入→规划→分镜→媒体，逐字段比对产出（用于验证形态差异是否影响结果，而非补写缺失功能）。

## 4. 证据索引

- 洛水版本：`dist/index.html`（bundle 名）、`launcher.log`（11-17 更新）、`version.json`（加密）
- 真实项目：`data/剧本创作/1212/`（上表）
- 提示词/表结构：`data/app.db`（sqlite3 只读）
- bundle 证据：`dist/assets/index-C1aA08SG.js`（关键词与桥命令用 python 提取）
- V88：`frontend/src/user/pages/ScriptPage.jsx`、`ComicScreenplayPage.jsx`、`ComicScreenplayProductionPanel.jsx:8,28-38`、`frontend/src/shared/api/{generation,comicScreenplayWorkflow}.js`、`backend/internal/screenplay/{storyboard_planning,authoring_service,model,media_store}.go`
