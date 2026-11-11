# 漫剧创作（`/comic-screenplay`）第一批实施与验证报告

- 日期：2026-11-11
- 范围：需求规格第 9 节第 1–4 项 + 剧本创作链路提示词归位
- 约束遵守情况：
  - 未 commit、未部署、未调用生成模型、未调用 TTS、未消耗额度；
  - 未启动开发服务器、未构建前端、未重启 4001；
  - 第 5 件「单镜头配音试听」本轮未做。

**结论：代码改动、Go 测试和前端语法检查已完成，可以构建了。**

---

## 1. 本轮改了什么

### 1.1 项目设置真保存

- 前端在右侧「项目设置」下新增「保存设置」按钮，点击后调用已有的 `PATCH /api/comic-screenplays/projects/{id}`。
- 保存字段：文本模型、画面比例、创作模式；同时回传现有标题、风格、梗概，避免后端把未传字段清空。
- 保存成功后刷新项目列表；刷新页面再进入项目，设置从数据库恢复。
- 代码位置：
  - `frontend/src/user/pages/ComicScreenplayPage.jsx` 的 `saveSettings`（123–146 行）
  - 按钮在 337 行
  - `frontend/src/shared/api/comicScreenplay.js` 的 `updateComicScreenplay`（47–50 行）

### 1.2 原文落库，刷新可恢复

- `screenplay_projects` 新增 `source_text LONGTEXT NULL` 列：
  - 迁移：`backend/internal/storage/screenplay_schema.go` 8800104（331–337 行）
  - 模型字段：`backend/internal/screenplay/model.go` 的 `Project.SourceText`（87 行）
- 新增只写原文、不参与 revision 乐观锁的存储方法：
  - 接口：`backend/internal/screenplay/store.go` 27–28 行
  - MySQL：`backend/internal/screenplay/mysql_store.go` 123–129 行
  - 内存实现：`backend/internal/screenplay/memory_store.go` 346 行起
- 新增接口 `PUT /api/comic-screenplays/projects/{id}/source`：
  - `backend/internal/httpapi/screenplay_routes.go` 154–171 行
- 前端的保存时机：
  - 打开项目时从 `project.source_text` 回填，不再把「已生成剧本」当原文；
  - 上传 TXT/DOCX 解析成功后立即保存；
  - 原文输入框失焦时静默保存；
  - 点「保存原文」显式保存；
  - 点「生成分集剧本」前先保存原文，保存失败不会继续调用模型。
- 前端位置：`ComicScreenplayPage.jsx` 的 `openProject`（93–106 行）、`persistSource`（108–121 行）、原文页签（314 行）。

### 1.3 分镜按集生成

- 从当前剧本正文按 `# 第N集` 标题切分集数（`splitDocumentEpisodes`，`ComicScreenplayPage.jsx` 23–37 行）。
- 分镜工作台新增集数下拉：
  - 单集：提交该集正文 + `episodeNo`；
  - 「全部集」：提交 `episodes` 映射，逐集提交；
  - 没有 `# 第N集` 标题时：页面明确提示「未检测到标题，将按第 1 集提交全文」。
- 生成按钮改为调用 `startStoryboard`（185–200 行），不再把整篇文档塞进第 1 集。
- 下拉与提示：316–322 行。

### 1.4 warnings 显示，不再无条件「处理完成」

- `runTask` 现在读取接口返回的 `warnings` 和 `failed_chunks`：
  - 有警告：顶部弹出「处理完成，但有 N 条警告」，并在页面内容区显示可关闭的黄色警告条，逐条列出；
  - 无警告：才显示「处理完成」；
  - 取消：显示「已取消，本次内容不会保存」。
- 位置：`ComicScreenplayPage.jsx` 的 `collectWarnings`（39–44 行）、`runTask`（164–183 行）、警告条（312 行）。
- 覆盖接口：资产提取（`warnings` + `failed_chunks`）、分镜生成（`warnings`）。

### 1.5 其他顺带修正

- 原文超过 60 万字时，前端在保存/生成前直接拦截并提示「原文超过 60 万字，请按章拆分后分批提交」，不发起请求（109 行）。
- 原文空、标题空、模型失败等原有校验保持不变，错误仍显示服务端原文。

---

## 2. 提示词导入与归位

### 2.1 结论先说

本轮需要的 12 条剧本创作链路提示词，**在本轮开始前已经以洛水原文形式内嵌在仓库**（`backend/internal/prompt/luoshui/prompts/*.txt`），并带 SHA-16 校验。本轮没有改写任何措辞，也没有重复导入副本；实际做的是：

1. 按运行时真实调用位置逐条核对；
2. 在管理后台新增「漫剧创作」分区，把这 12 条从「未归属」移到对应分区；
3. 增加测试，锁定分区数量、条目顺序，以及名称不得出现「洛水」「小洛」。

### 2.2 逐条映射

| 链路步骤 | 运行时读取位置 | prompt ID | 后台显示名 | 语料来源 | 原文文件 | SHA-16 |
|---|---|---|---|---|---|---|
| 剧本 · 小说转剧本 A1 | `backend/internal/screenplay/screenplay_service.go:73` | `a1_novel_to_screenplay` | 小说转漫剧 A1 | `data\app.db` prompts#228（§3.1，name=小说转漫剧A1） | `a1_novel_to_screenplay.txt` | `2c7af473b54a612d` |
| 剧本 · 导演级剧本 B | `screenplay_service.go:117` | `b_director_screenplay` | 网文→漫剧剧本 B 档精品 | `data\app.db` prompts#260（§3.3，name=网文→导演级分镜剧本B挡） | `b_director_screenplay.txt` | `04f24c456bf509bc` |
| 资产 · 角色（大师 3D） | `backend/internal/screenplay/assets_service.go:285` | `asset_character_v6_3d` | AI小说角色提取器 V3（3D/V6） | `modules\screenplay.pyd` @0x1FEA0F9（§3.10） | `asset_character_v6_3d.txt` | `37fc9e6c3d950d83` |
| 资产 · 角色（大师写实） | `assets_service.go:287` | `asset_character_real_v7` | AI角色提取器（真人写实版） | `modules\screenplay.pyd` @0x1FEEA03（§3.12） | `asset_character_real_v7.txt` | `00ff11dbbf49746f` |
| 资产 · 角色（标准） | `assets_service.go:281` | `asset_character_standard_v5` | AI角色提取器（标准版） | `modules\screenplay.pyd` @0x1FF14BA（§3.13） | `asset_character_standard_v5.txt` | `0c5ea75fef8b3aae` |
| 资产 · 角色（差异化） | `assets_service.go:283` | `asset_character_lite` | AI角色提取器（精简版·按剧本推理） | `modules\screenplay.pyd` @0x1FED657（§3.11） | `asset_character_lite.txt` | `a6304fb76c94646d` |
| 资产 · 场景（默认） | `assets_service.go:293` | `asset_scene_v4` | AI小说场景提取器 V4 | `modules\screenplay.pyd` @0x1FF3CBB（§3.15） | `asset_scene_v4.txt` | `0cbae13c98d5a420` |
| 资产 · 场景（标准） | `assets_service.go:291` | `asset_scene_v3` | AI小说场景提取器 V3 | `modules\screenplay.pyd` @0x1FE4116（§3.14） | `asset_scene_v3.txt` | `8ae05af03f16ceac` |
| 资产 · 道具（默认） | `assets_service.go:298` | `asset_prop_v4` | AI小说特殊道具提取器 V4 | `modules\screenplay.pyd` @0x1FFAB80（§3.17） | `asset_prop_v4.txt` | `d51aa036b4237a23` |
| 资产 · 道具（标准） | `assets_service.go:296` | `asset_prop_v3` | AI小说特殊道具提取器 V3 | `modules\screenplay.pyd` @0x1FFD1D2（§3.16） | `asset_prop_v3.txt` | `1478ca0d88ecba13` |
| 分镜 · CM 分镜（v7-15s） | `backend/internal/screenplay/storyboard_service.go:332` | `storyboard_xiaoluo_v7_15s` | 小Y v7-15s 分镜词 | `data\app.db` prompts#225（§3.5，name=小洛v7-15s版-913最新版） | `storyboard_xiaoluo_v7_15s.txt` | `094f98b7c49e829d` |
| 分镜 · 小云雀 v3（默认引擎） | `storyboard_service.go:330` | `storyboard_xiaoyunque_v3` | 小云雀分镜词 v3 | `data\app.db` prompts#268（§3.6，name=小云雀v3） | `storyboard_xiaoyunque_v3.txt` | `630a00ee06267fc7` |

说明：

- 后台显示名没有「洛水」「小洛」字样；测试会扫描全部条目名称并在出现时失败。
- 管理后台分区定义：`backend/internal/prompt/luoshui_admin.go` 21–48 行（`comic-screenplay`，标签「漫剧创作」）。条目顺序固定为「剧本 → 资产 → 分镜」。
- 后台条目说明文案为「漫剧创作链路在用，编辑发布后生效」。
- 对应测试：`backend/internal/prompt/luoshui_admin_test.go` 的 `TestLuoshuiAdminPresetsGroupedByFunction`（分区 7 个、漫剧创作 12 条、未归属 23 条、顺序固定、名称无旧产品名）；`TestMergeLuoshuiAdminPresetsLocksComicScreenplaySlot` 保证历史覆盖版本也显示在新分区，不会被旧分区名带走。

### 2.3 本轮明确没有纳入的提示词

| ID | 原因 |
|---|---|
| `doctor_screenplay_*`、`doctor_super_drama` | 当前漫剧链路没有运行时代码读取审校提示词，纳入会造成「后台可编辑但实际不生效」的误解 |
| `asset_sd_triplet` | 当前运行时未引用 |
| `batch_step_0`–`batch_step_8` | 只在批次步骤表里作为 `prompt_id` 出现，没有代码读取正文 |

它们仍留在「未归属」分区，后台说明为「历史存档，当前无运行时引用」。

---

## 3. 五种边界逐条结论

> 说明：本轮按要求没有构建前端、没有启动或操作 4001，因此所有需要真实点击页面的项统一记为「未验证」，不写「通过」。静态结论指读代码可确认的行为。

| 情况 | 结论 | 依据 |
|---|---|---|
| 空值 | **代码路径已覆盖，页面未验证** | 标题空：`createComicScreenplay` 抛「作品名称不能为空」；原文空：`generateComicScreenplay` 抛「请先上传或输入原文」；生成前会先保存原文，保存空内容允许（用于清空原文）。未实际点击。 |
| 超长 | **代码路径已覆盖，页面未验证** | 前端 `persistSource` 在超过 600000 字时直接抛「原文超过 60 万字，请按章拆分后分批提交」；后端同一接口返回 400 同文案；生成服务原有 60 万字上限未改动。未实际粘贴 60 万字验证。 |
| 重复点击 | **代码路径已覆盖，页面未验证** | `runTask` 首行 `if (working) return`；`saveSource` 用 `sourceSaving` 拦截；`saveSettings` 用 `settingsSaving` 拦截。未实际连点验证。 |
| 取消 | **提示已补，页面未验证** | `AbortError` 分支现在提示「已取消，本次内容不会保存」，按钮恢复。取消时服务端是否已落库取决于取消瞬间任务是否已完成，本轮未实机验证。 |
| 模型失败 | **错误透传保持，页面未验证** | `runTask` 捕获后显示服务端原文（如「API Key未配置…」「API请求失败…」）；分镜/资产的警告走黄色警告条。未实机制造模型失败。 |

---

## 4. 已执行的验证命令与结果

| 命令 | 结果 |
|---|---|
| `go test ./internal/screenplay/... ./internal/prompt/... ./internal/storage/...` | 通过；新增 `TestSaveProjectSourcePersistsWithoutBumpingRevision` 验证原文持久化、revision 不变、更新设置不清空原文 |
| `go build ./...` | 通过 |
| `node --test src/shared/api/comicScreenplay.test.js` | 9 项全部通过，含新增「项目设置和原文都调用已有项目接口」 |
| `node -e "esbuild.transformSync(...)"` 检查 `ComicScreenplayPage.jsx`、`comicScreenplay.js` | 语法通过 |
| `go test ./internal/httpapi/` | 未通过，但与本轮改动无关：`TestModelCatalogH3ConfigRequiresKeyAndSavesWorkflowTogether`、`TestGPTImageConcurrencyIsPerTaskNotProcessWide` 失败；之后并行会话正在编辑 `shuihuo_*` 测试文件，出现过 `undefined: errors` 等编译错误。本轮未修改这些文件。 |

---

## 5. 构建后建议的人工点击清单

构建完成、页面可用后，按下面顺序点一遍即可覆盖五种边界：

1. **空值**：新建项目不填名称点创建，应看到「作品名称不能为空」；进入项目，原文为空时点「生成分集剧本」，应看到「请先上传或输入原文」。
2. **超长**：粘贴超过 60 万字的文本后失焦或点「保存原文」，应看到「原文超过 60 万字，请按章拆分后分批提交」，且请求未发出。
3. **重复点击**：生成剧本或分镜时连续点击按钮，只应有一个请求，按钮保持 loading。
4. **取消**：生成中点「取消任务」，应看到「已取消，本次内容不会保存」，按钮恢复。
5. **模型失败**：把模型改成不可用配置后生成，应显示服务端错误原文，不得出现「处理完成」。

另需确认：

6. 改模型/比例/创作模式 → 点「保存设置」→ 刷新页面，三项仍在。
7. 粘贴原文 → 失焦 → 刷新页面，原文仍在；上传 TXT/DOCX 后刷新也在。
8. 剧本含 `# 第1集`、`# 第2集` 时，分镜页下拉出现两集；只选第 2 集生成，接口请求里 `episodeNo=2` 且 `content` 只含第 2 集。
9. 资产提取故意让某一批失败，页面应出现黄色警告条，而不是只提示「处理完成」。

以上 1–9 项目前均为「未验证」，需在构建后的真实页面上确认。
