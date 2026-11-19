# 洛水真实项目 1212 × V88 字段级对比执行报告

- 日期：2026-11-18 15:30
- 性质：**只读实测**。未启动洛水 GUI、未调用任何生成接口、未消耗额度、未改产品代码、未发布/未重启 4001。
- 洛水样本：`F:\软件\洛水2026\洛水2026\data\剧本创作\1212\`（用户指定最新版本里的真实项目，2026-11-14 创建）
- V88 证据：`backend/internal/screenplay/{model,authoring_service,media_store,storyboard_service}.go`、`backend/internal/httpapi/screenplay_authoring_routes.go`
- 执行方式：把洛水真实分镜 JSON 喂给 V88 原生解析器 `parseStoryboardJSON`（`storyboard_service.go:579`），逐字段断言，跑 `go test`。

## 0. 一句话结论

**洛水 1212 的分镜词能被 V88 原生解析器无损解析，核心字段全部对上**；差异只有两处且都不丢数据（`image_prompt` 洛水本就没有；`_scene_local_order` 没映射到专用字段但完整保存在 ProtocolFields）。这从执行层面印证了上一份报告的修正口径：**能力已落地，剩下的是字段形态与映射收尾，不是缺一整条链**。

## 1. 执行验证：洛水真实分镜 → V88 原生解析器

测试：`backend/internal/screenplay/storyboard_luoshui1212_parity_test.go`
样本：`backend/internal/screenplay/testdata/luoshui1212_storyboard.json`（从 1212 的 `chapter_0_storyboard.json` 提取，3 条真实分镜）

```
=== RUN   TestLuoshui1212StoryboardNativeParse
形态差异1: 洛水分镜无 image_prompt 字段（307 提示词仅用 video_desc），V88 解析为空串，语义等价
形态差异2: _scene_local_order 洛水=1，V88 专用字段 SourceSceneLocalOrder=0（未映射），但原值保存在 ProtocolFields 未丢失
结论: 洛水 1212 真实分镜 3 条核心字段全部通过 V88 原生解析且无损；扩展字段经 ProtocolFields 完整保留
--- PASS
EXIT:0
```

### 1.1 逐字段映射结果（洛水分镜 → V88 `StoryboardShot`）

| 洛水字段（`chapter_0_storyboard.json.storyboard[]`） | V88 字段（`model.go:243-264`） | 执行结果 |
|---|---|---|
| `id` | `Seq`（`storyboard_service.go:621`） | ✅ 无损（数值一致） |
| `scene_id` | `SceneID` | ✅ 无损 |
| `duration_sec` | `Duration` | ✅ 无损 |
| `characters` | `Characters` | ✅ 无损 |
| `scene` | `Scene` | ✅ 无损 |
| `props` | `Props` | ✅ 无损 |
| `video_desc` | `VideoDesc` | ✅ 无损（含九类子字段的完整原文） |
| `image_prompt` | `ImagePrompt` | ⚠️ 洛水 307 提示词版**无此字段**，V88 解析为 `""`，语义等价、非丢失 |
| `_scene_index` | `SourceSceneIndex` | ✅ 无损 |
| `_scene_local_order` | `SourceSceneLocalOrder` | ⚠️ 洛水有值（1/2/3），V88 专用字段**未映射**（=0），但原值完整保存在 `ProtocolFields`，未丢失 |
| 任意扩展字段 | `ProtocolFields`（json.RawMessage） | ✅ 整组原始 JSON 原样保留，零字段丢失 |

**判定**：10 个核心字段里 8 个直接无损、2 个是形态差异且都不丢数据。V88 的 `ProtocolFields` 兜底设计保证洛水未来新增任何字段都不会解析失败。

## 2. 全项目字段映射（1212 所有 JSON → V88 结构）

### 2.1 项目元信息 `info.json`

| 洛水字段 | 样本值 | V88 对应 | 结论 |
|---|---|---|---|
| `name` | 1212 | `Project.Title` | ✅ |
| `script_mode=dialogue` | dialogue | `Project.ScriptMode` / 创作配置 | ✅ 枚举对应 |
| `workspace_mode=pro` | pro | 无直接字段 | ⚠️ V88 无"页面工作台档位"概念，产品差异 |
| `editor_version=new` | new | 无直接字段（V88 只有新版编辑器） | ✅ 形态不同，V88 无需该开关 |
| `created_at/updated_at` | 2026/11/14 | `Project.CreatedAt/UpdatedAt` | ✅ |

### 2.2 章节 `chapters.json` → `Chapter`（`authoring_service.go:15-36`）

| 洛水字段 | V88 字段 | 结论 |
|---|---|---|
| `id`(int 0) | `Chapter.ID`(string) | ✅ 类型不同（int→string），语义对应 |
| `title` | `Chapter.Title` | ✅ |
| `content`（原文） | `Chapter.SourceText` | ✅ |
| `scriptResult`（剧本，None） | `Chapter.ScriptText` | ✅ 本样本为 None，V88 空串 |
| `storyboardResult`（含 storyboard+plan 的 JSON 串） | `Chapter.PlanningText`（plan）+ 独立 Storyboard 记录（shots） | ✅ **拆分存储**：洛水一个串装两样，V88 规划与分镜分表；内容都能落 |
| `storyboardFailedScenes`（失败场次） | `ai_raw_responses` 等价的重试状态 / `Storyboard.Status` | ⚠️ 洛水在章节内记失败场次，V88 用生成尝试表，形态不同 |
| `scriptMode=master` | 项目级脚本模式 | ✅ |

### 2.3 规划 `plan`（4830 字）→ `Chapter.PlanningText`

| 洛水 | V88 | 结论 |
|---|---|---|
| `storyboard/plan` 内的规划正文 | `Chapter.PlanningText`（`authoring_service.go:24`） | ✅ 直接落同一个字符串字段 |
| `storyboardPlanFirst=true` | 提交优先走 GenerateStoryboardPlan，续写字段 `Continue/ExistingPlan`（`storyboard_planning.go:39-42`） | ✅ 方向一致；续写后端已写好待提交+前端接线 |

### 2.4 资产 `agent_assets.json` → `Asset`（`model.go:187-206`）

| 洛水字段 | V88 字段 | 结论 |
|---|---|---|
| `characters[].name` | `Asset.Name` | ✅ |
| `characters[].prompt` | `Asset.Prompt` | ✅ |
| `characters[].identity` | `Asset.Identity` | ✅ 专用字段 |
| `characters[].voice` | `Asset.Voice` | ✅ 专用字段 |
| `characters[].颜值定位` | `Asset.Extra["颜值定位"]` | ✅ 扩展字段兜底 |
| `characters[].出现集数` | `Asset.Episodes` | ✅ 专用字段 |
| `scenes[].description` | `Asset.Description` | ✅ |
| `scenes[].空间要点` | `Asset.Extra["空间要点"]` | ✅ 扩展字段兜底 |
| `scenes[].prompt` | `Asset.Prompt` | ✅ |
| `props[]`（本样本空） | `Asset.Type=prop` | ✅ |

**判定**：洛水资产字段在 V88 全部有落点——4 个专用字段 + 中文扩展字段走 `Extra` 兜底，零字段无处可放。

### 2.5 配置 `ai_config.json` / `model_config.json` → V88 模型/创作配置

| 洛水字段 | 样本值 | V88 对应 | 结论 |
|---|---|---|---|
| `model` | doubao-seed-... | `Project.ModelID` | ✅（用自有模型替代） |
| `promptId=144/专业剧本` | 144 | V88 预设库 | ✅ 概念对应，ID 体系不同 |
| `storyboardPlanFirst` | true | `Continue/ExistingPlan` 规划流程 | ✅ |
| `aspectRatio=9:16` | 9:16 | `Project.AspectRatio` | ✅ |
| `imageGenModel/voiceDesignModel/dubbingModel` | 自有模型 | `CreativeConfig` 各模型 ID | ✅ 形态一致，值用自有 |
| `assetParseDegree=master` | master | `Asset.Degree` 档位 | ✅ |
| `characterThreeView/四宫格提示词` | 213/289 号提示词 | V88 预设提示词 | ✅ 概念对应 |

## 3. 实测发现的三个"执行层"待办（都非功能缺失）

1. **`_scene_local_order` 未映射到专用字段**：`parseStoryboardJSON`（`storyboard_service.go:625-639`）没有把该字段赋给 `StoryboardShot.SourceSceneLocalOrder`。数据没丢（在 ProtocolFields），但下游若读专用字段会拿到 0。**这是一行赋值的改动**，属于可顺手补的映射收尾。
2. **`image_prompt` 缺省**：洛水 307 分镜词只有 `video_desc`，V88 解析为空串——这是正确行为（与 `storyboard_139_optional_image` 约定一致），不算 bug。
3. **无一键导入接口**：把 `1212` 整套数据导进 V88 需要走多个写入接口（章节、资产、分镜、媒体）逐项写入，目前没有"洛水项目 JSON 一键导入"的入口。这是工具链缺口，不是功能缺口。

## 4. 结论与建议

- **执行验证通过**：洛水真实分镜数据能被 V88 原生链路接受，字段级对照全部有落点。上一份报告的"实现形态差异"判断得到实测确认。
- **可顺手补的一步**：在 `parseStoryboardJSON` 里补一行 `_scene_local_order → SourceSceneLocalOrder` 的映射（并纳入测试），让这个字段从"存在 ProtocolFields"升级为"专用字段可读"。
- **不建议做的**：为对齐洛水去改数据模型（把 MediaItem 关联表改成 shots 内嵌字段、把字幕挪回剧本链）——V88 的关联表/AI-CUT 归属是等价实现，改动收益低、风险高。
- **本次未做**：真实导入 1212 到运行库、真实 LLM 生成、付费媒体调用（均需登录额度，超出只读边界）。

## 5. 证据索引

- 测试：`backend/internal/screenplay/storyboard_luoshui1212_parity_test.go`（新增，独立测试名）
- 样本：`backend/internal/screenplay/testdata/luoshui1212_storyboard.json`（13KB，真实分镜）、`luoshui1212_plan.txt`（4830 字规划）
- 洛水原件：`F:\软件\洛水2026\洛水2026\data\剧本创作\1212\`（info/chapters/chapter_0_storyboard/agent_assets/ai_config/model_config JSON + screenplay_ch0.db）
- V88 结构：`model.go:121-264`、`authoring_service.go:15-36`、`media_store.go:12-39`、`storyboard_service.go:579-645`
- 前序报告：`docs/acceptance/luoshui-latest-vs-v88-screenplay-2026-11-18.md`
