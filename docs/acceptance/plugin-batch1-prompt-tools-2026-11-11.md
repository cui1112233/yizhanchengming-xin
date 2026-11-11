# 插件第一批验收：4 条提示词工具（2026-11-11）

验收对象（同一条实现链路，4 个工具 ID）：

| 插件 ID | 名称 | 目录行 | 面板规格行 |
|---|---|---|---|
| `novel-to-drama` | 小说转剧 | `frontend/src/user/pages/pluginCatalog.js:33` | `pluginCatalog.js:56` |
| `storyboard-master` | 分镜大师 | `pluginCatalog.js:35` | `pluginCatalog.js:74` |
| `viral-opening` | 爆款开头 | `pluginCatalog.js:37` | `pluginCatalog.js:92` |
| `character-edit` | 角色修改 | `pluginCatalog.js:40` | `pluginCatalog.js:110` |

选择理由：这 4 条是洛水 v2.9.0 中**有真实整页面板**且已抓到截图的一组（见 `shots/plugin-*.png`、`*-panel.jpg`、`*-empty-toast.jpg`、`*-model-dropdown.jpg`、`*-prompt-dropdown.jpg`、`*-concurrent-*.jpg`），可以逐项做「照抄 / 推导 / 差异」判断，符合验收标准里「洛水有真实界面」的批次条件。

总体状态：**部分验证**。逻辑层与后端 handler 层已验证；界面层（真实浏览器点击、Toast、双击、取消）在构建冻结期间未验证。

---

## 1. 公共链路取证

### 1.1 入口

- 目录条目与 `href` 派生：`frontend/src/user/pages/pluginCatalog.js:2`（`promptTool` 生成 `/plugins/tool/<toolId>`）。
- 路由从目录派生，不逐条登记：`frontend/src/user/App.jsx:63-66`，命中逻辑在 `App.jsx:103-106`。
- 卡片点击：`frontend/src/user/pages/PluginCenterPage.jsx:49-53`（`target.kind === 'prompt-tool'` 时 `history.pushState` 并派发 `popstate`）。
- 面板按 pathname 反查工具：`frontend/src/user/pages/PluginToolPage.jsx:28-33`。

### 1.2 调用链

```
PluginCenterPage 卡片
  → App.jsx 路由渲染 PluginToolPage（key=pathname，切换工具重建实例）
  → PluginToolPage 初始化：GET /api/plugins/{toolID}/config
  → 点击「AI生成」：POST /api/plugins/{toolID}/run
  → backend/internal/httpapi/plugin_tools.go handler
  → backend/internal/prompt/plugin_tools.go（提示词选项 + 系统提示词）
  → backend/internal/prompt/luoshui/registry.go（内嵌语料正文 / 管理员覆盖正文）
  → AgentTextCompletionExecutor 或 options.TextProvider（文本模型通道）
  → 响应 {"result": "..."} → React state → 本地下载
```

证据：

- 组件入口：`PluginToolPage.jsx:46`。
- 配置加载：`PluginToolPage.jsx:63-82` → `frontend/src/shared/api/pluginTools.js:5-9`。
- 运行：`PluginToolPage.jsx:109-121`（`generatePart`）→ `pluginTools.js:11-19`。
- 路由装配：`backend/internal/httpapi/router.go:564`。
- config handler：`backend/internal/httpapi/plugin_tools.go:50-57`。
- run handler：`plugin_tools.go:193-235`。
- 提示词选项表：`backend/internal/prompt/plugin_tools.go:54-76`。
- 选项只暴露 ID/Name：`plugin_tools.go:92-102`。
- 系统提示词取值（默认 `default`）：`plugin_tools.go:104-119`。
- 内嵌语料：`backend/internal/prompt/luoshui/registry.go:19-20`（`go:embed prompts/manifest.json prompts/*.txt`）、加载与 SHA256 校验 `registry.go:88-120`、取正文 `registry.go:155-161`。
- 管理员覆盖优先：`plugin_tools.go:39-50`、`registry.go:176-185`。

### 1.3 网络请求

| 请求 | 方法/URL | 请求体字段 | 响应字段 | 超时 | 大小限制 |
|---|---|---|---|---|---|
| 读提示词选项 | `GET /api/plugins/{toolID}/config` | 无 | `prompts: [{id,name}]` | 8s | 无 |
| 执行生成 | `POST /api/plugins/{toolID}/run` | `input`、`modelId`、`promptId`、`raw` | `result` | 10min | 请求体 2MB（`plugin_tools.go:199`） |
| 读 Word/TXT | `POST /api/plugins/extract-text` | multipart `file` | `text`、`name` | 2min | 20MB（`plugin_tools.go:153`） |

`raw: true` 时后端不再包 `## 需要处理的内容` 前缀，直接发送拼装后的分段提示词：`plugin_tools.go:215-218`；前端在 `PluginToolPage.jsx:113` 传入 `raw: true`。

### 1.4 数据变化

- 服务端：**无任何写入**。run handler 只调用模型并返回字符串（`plugin_tools.go:219-234`）。
- 前端：输入/输出都在 React state（`PluginToolPage.jsx:52-53`），不写 localStorage。
- 落盘只发生在用户点「保存」时，由浏览器下载完成：`PluginToolPage.jsx:35-42`（`downloadFile`）、`:202-209`（TXT）、`:211-222`（JSON）。
- 因此：不存在脏数据、重复记录、半成品写入数据库的问题；这是本批的确定性结论。

### 1.5 关键函数

| 函数 | 位置 | 作用 |
|---|---|---|
| `currentPromptTool` | `PluginToolPage.jsx:28` | pathname → 目录条目 + 面板规格 |
| `readUpload` | `PluginToolPage.jsx:84-106` | TXT/MD 本地读，DOC/DOCX 走 `extract-text` |
| `splitInputByLength` | `frontend/src/user/pages/pluginToolLogic.js:85-113` | 按 `progressChunkSize` 分段，超长行硬切 |
| `buildPartPrompt` | `pluginToolLogic.js:117-124` | 多段时带完整原文与历史记录 |
| `generatePart` | `PluginToolPage.jsx:109-121` | 单段调用，空结果按失败处理，最多 2 次 |
| `generate` | `PluginToolPage.jsx:123-166` | 串行/并发分段，汇总输出 |
| `extractScript` | `pluginToolLogic.js:8-21` | 小说转剧正文提取 |
| `formatScriptText` | `pluginToolLogic.js:22-33` | 剧本格式化 |
| `buildDialogueList` | `pluginToolLogic.js:53-83` | 保存 JSON 的台词结构 |

---

## 2. 需求规格（逐工具）

公共部分：

- 输入：`input`（string，必填，trim 后非空）、`modelId`（string，必填，来自 V88 文本模型库）、`promptId`（string，必填，来自 `/config`）、`concurrent`（boolean，默认 `false`）、可选文档上传（`.txt/.md/.doc/.docx`，≤20MB）。
- 输出：结果字符串写入结果区；不落库；「保存」触发浏览器下载。
- 边界：输入 2MB 上限；分段阈值见各工具；空输入/空模型/空提示词均在点击后提示拦截，主按钮点击前不禁用。
- 失败：`404 插件工具不存在`、`400 请输入需要处理的内容`、`400 请选择有效的提示词`、`503 文本模型执行器未配置`、`502 文本模型请求失败`（`plugin_tools.go:194-233`）。

### 2.1 小说转剧 `novel-to-drama`

- 输入：textarea 6 行 / 高 148（`pluginCatalog.js:56-73`）；placeholder `请输入中长篇小说内容，或上传txt/word文档...`；分段阈值 2000（`pluginCatalog.js:33`）。
- 输出：结果标题「生成的剧本」，10 行；显示字数；格式化按剧本模式；保存支持 TXT/JSON；文件名「中长篇转剧结果」。
- 边界：空输入提示「请输入或上传小说内容」；仅空格、空模型、空提示词分别在点击时拦截（`PluginToolPage.jsx:125-128`）；重复点击已加运行中拦截（`:125`）。
- 失败：后端文案原样透出（`pluginTools.js:13-18` → `apiRequest` 抛出）。
- 自检：通过（规格项与 `pluginCatalog.js:56-73`、`PluginToolPage.jsx:239-330` 一一对应）。

### 2.2 分镜大师 `storyboard-master`

- 输入：8 行 / 高 193；placeholder `请输入剧本或小说内容，或上传txt/word文档...`；阈值 1500。
- 输出：标题「AI分镜大师」；结果标题「生成的分镜」，8 行；不显示字数；按行格式化；仅 TXT；文件名「AI分镜结果」。
- 边界：空输入「请输入或上传内容」；无「已授权」标记。
- 自检：通过（`pluginCatalog.js:74-91`）。

### 2.3 爆款开头 `viral-opening`

- 输入：8 行 / 高 193；placeholder `请输入小说原文，或上传txt/word文档...`；阈值 0（整段一次发送）。
- 输出：标题「爆款开头」；结果标题「生成结果」，11 行；不显示字数；按行格式化；仅 TXT；文件名「爆款开头结果」。
- 边界：空输入「请输入或上传内容」；阈值 0 时 `splitInputByLength` 返回单段（`pluginToolLogic.js:88`）。
- 自检：通过（`pluginCatalog.js:92-109`）。

### 2.4 角色修改 `character-edit`

- 输入：8 行 / 高 193；placeholder `请输入需要修改角色的内容，或上传txt/word文档...`；阈值 1500。
- 输出：结果标题「修改结果」，8 行；不显示字数；按行格式化；仅 TXT；文件名「角色修改结果」。
- 边界：空输入「请输入或上传内容」。
- 自检：通过（`pluginCatalog.js:110-127`）。

4 条共同：都有「并发生成」开关（`PluginToolPage.jsx:321-328`），运行中禁用；模型/提示词下拉默认空、可搜索（`:297-318`）；主按钮 `loading` 而非 `disabled`（`:327`）。

---

## 3. 五场景实测

### 3.1 已验证（可复现）

命令与结果（2026-11-11，本机，未构建）：

```
cd frontend
node --test src/user/pages/pluginToolLogic.test.js src/user/pages/pluginCatalog.layout.test.js
# tests 16 / pass 16 / fail 0

cd backend
go test ./internal/prompt/ -run "PluginTool|Luoshui" -v -count=1
# 7 项 PASS（Luoshui 预设/覆盖 5 项 + 提示词选项/语料 2 项）
go test ./internal/httpapi/ -run "PluginTool|Docx|CoverDesign" -v -count=1
# 6 项 PASS（路由自持提示词、config 不泄漏正文、未知工具/空输入、Docx 提取、封面）
```

新增测试文件已登记到 `frontend/package.json:7` 的 `test` 脚本，后续 `npm test` 会一并执行。

| 场景 | 实测输入 | 实际结果 | 结论 |
|---|---|---|---|
| 正常输入 | `第一场\n第二场`，limit 12 | 分 2 段且 `parts.join('\n')` 还原原文；提示词含「完整原文」「第1/2部分」 | 通过 |
| 空值 | `''` / `'   '` / `'\n\n\t '` | `extractScript` 返回 `内容为空`；格式化与台词列表均为 `[]`；分段返回单段交由上层拦截 | 通过（逻辑层） |
| 重复操作 | 同输入连续调用分段/拼装/格式化 | 结果逐字节一致；同角色相邻台词 ≤150 字只合并成 1 条 | 通过（逻辑层）；页面已加运行中拦截（`PluginToolPage.jsx:125`），浏览器连点未验证 |
| 错误操作 | 5000 字单行 limit 2000；`$& $1 \\ " 😀` | 切片长度 2000/2000/1000 且拼回等于原文；特殊字符原样保留；非台词行返回 `null` | 通过 |
| 取消操作 | 关闭面板/切路由（2026-11-11 补测） | 已补 `AbortController`：`PluginToolPage` 在卸载时 `abort()`，请求用带 `signal` 的 `apiRequest`；`apiRequest` 收到 caller abort 时抛 `AbortError` 且不上报、不弹全局错误。单测断言 `signal.aborted === true` 且取消后不返回任何结果 | 通过（逻辑层 + 请求层）；浏览器点击级仍未验证，见 §3.2 |

后端 handler 已实测：未知工具 404、空输入 400（`plugin_tools_test.go:93-112`）；5 条工具均使用后端自持提示词，用户正文进入 `UserPrompt`（`plugin_tools_test.go:25-46`）；`/config` 不泄漏提示词正文（`:48-61`）。

### 3.2 未验证（必须重测）

| 项目 | 未验证原因 |
|---|---|
| 浏览器内空输入 Toast 文案、双击第二次提示 | 构建冻结：禁止 `npm run build`、禁止改 `dist`；4001 仍是旧构建 |
| 并发生成开关切换后的实际分段顺序与进度文案 | 同上 |
| 上传 TXT/Word 的实际读入与错误提示 | 同上（`extract-text` 后端单测通过，前端交互未测） |
| 取消/切路由后的请求终止与结果丢弃 | 2026-11-11 已补 AbortController 并通过单测；浏览器内「关闭面板后 Network 面板显示 canceled」仍待重测 |

---

## 4. 与洛水对照

洛水证据（本轮直接打开 bundle 复核）：

- 小说转剧处理器：`index-DszfAkKD.js:464402-464442`（`long-novel-to-script`，先查 `check_dongman_permission`，再 `get_all_prompts("中长篇转剧",1,100,"")`）。
- AI改文：`464444` 起。
- 分镜大师：`464524-464548`（`get_all_prompts("分镜",...)`）。
- 角色修改：`464549-464573`（`get_all_prompts("角色修改",...)`）。
- 爆款开头：`464574-464597`（`get_all_prompts("爆款开头",...)`）。

截图证据（`F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\shots\`）：

- 面板：`novel-to-drama-panel.jpg`、`plugin-storyboard-master.png`、`plugin-viral-opening.png`、`character-edit-panel.jpg`。
- 空态提示：`novel-to-drama-empty-toast.jpg`、`storyboard-master-empty-toast.jpg`、`viral-opening-empty-toast.jpg`、`character-edit-empty-toast.jpg`。
- 下拉：`*-model-dropdown.jpg`、`*-prompt-dropdown.jpg`。
- 并发：`novel-to-drama-concurrent-on.jpg`、`novel-to-drama-concurrent-off.jpg`。

本轮重看 `novel-to-drama-panel.jpg` 的实测结论：整页面板（不是弹窗），面包屑 `返回 / 插件市场 > 小说转剧`，标题左侧彩色图标 + 「已授权」+ 右上关闭，顶部「上传文档」，下面是输入框、两个下拉、「并发生成」开关、整宽绿色「AI生成」、结果区「生成的剧本」带 4 个工具图标（复制/清空/格式化/保存），空态文案「选择模型和提示词，输入内容后点击AI生成」。

| 项目 | 结论 |
|---|---|
| 整页面板而非弹窗 | 照抄（`PluginPanelShell.jsx:16-55`；`PluginToolPage.jsx:239`） |
| 面包屑 + 标题图标 + 已授权 + 关闭 | 照抄（`PluginPanelShell.jsx:20-47`） |
| 输入区 → 下拉 → 并发生成 → 整宽主按钮 → 结果区 | 照抄（`PluginToolPage.jsx:284-330`） |
| 结果区 4 个图标与空态文案 | 照抄（`PluginToolPage.jsx:249-280`） |
| 模型/提示词来源 | **差异**：洛水走 `pywebview.api.get_all_prompts(category)`；V88 走自己的文本模型库 + `/api/plugins/{id}/config` |
| 权限 | **差异**：洛水有 `check_dongman_permission`；V88 无此权限门，只有小说转剧显示「已授权」标记 |
| 分段/历史提示词 | **推导**：分段阈值来自目录条目（2000/1500/0），拼装文案由洛水 `build_part_prompt` 行为推导，未在 bundle 内逐字复核 |
| 保存 TXT/JSON | **推导 + 部分照抄**：`extractScript` 等 3 个函数注释标明照搬洛水 `fc/pu/bc`，但本轮未重新在 bundle 定位这 3 个函数 |

**重要事实更正**：`shots/v88-panel-novel-to-drama.png` 与 `shots/panel6-compare-novel-to-drama.png` 里 V88 仍是**旧弹窗**（左右两栏、无面包屑、无并发生成）。当前源码已改为整页面板，但构建冻结期间没有新截图，因此这两个文件**不能作为当前实现的对照证据**。

---

## 5. 本轮代码改动

| 改动 | 文件 | 目的 |
|---|---|---|
| 抽出纯逻辑模块 | 新增 `frontend/src/user/pages/pluginToolLogic.js` | 让分段/拼装/剧本解析可被 node 直接测试 |
| 页面改为引用模块 | `frontend/src/user/pages/PluginToolPage.jsx:8-15` | 行为不变，消除重复实现 |
| 运行中重复提交拦截 | `PluginToolPage.jsx:125` | 对应「重复操作」场景 |
| 五场景逻辑测试 | 新增 `frontend/src/user/pages/pluginToolLogic.test.js` | 正常/空值/重复/错误/取消五类 |
| 修正静态断言 | `frontend/src/user/pages/pluginCatalog.layout.test.js:28,139-140,156-157,219-220` | 适配逻辑模块拆分；补 `local-tool`/`turnaround-tool` 两种 launch 类型（原断言已过期） |

已知缺口（不通过）：

1. **取消（2026-11-11 已修复，逻辑层通过）**：新增 `frontend/src/user/pages/pluginGeneration.js`（`createGenerationGate` / `executeGeneration` / `validateGenerationInput`）与单测；`PluginToolPage.jsx` 卸载时 `gateRef.abort()`，请求经 `apiRequest(..., { signal })` 发出；`client.js` 对 caller abort 抛 `AbortError`、不上报、不弹全局错误。浏览器点击级验证仍缺（构建冻结）。
2. 关闭面板不会清空结果，但组件卸载后状态自然丢失；不影响数据一致性。

---

## 6. 26 项插件状态登记

| 插件 | 入口实现 | 状态 |
|---|---|---|
| 小说转剧 / 分镜大师 / 爆款开头 / 角色修改 | `PluginToolPage` 整页面板 | 部分验证（本文件） |
| AI改文 | `route('/plugins/ai-rewrite')`，独立页 | 未纳入本批，未验证 |
| 资产参考 | `AssetReferenceWorkbench` | 已实现，本批未验收 |
| AI-分段 | `AISegmentWorkbench` | 已实现，本批未验收 |
| AI-Subtitle / AI-Filler / AI-Split | `DraftToolWorkbench`，写回 AI-CUT 时间轴 | 已实现，本批未验收 |
| AI-Upscale | `UpscaleWorkbench` | 已实现，本批未验收 |
| 封面生成 / 角色三视图 | 图片通道工作台 | 已实现，本批未验收（角色三视图有 `mock=1` 演练） |
| 格式化文本 / 字幕合并 / 文档合并 / 分镜整理 | `LocalToolWorkbench` 纯本地 | 已实现，单元测试在 `localToolLogic.test.js`、`storyboardOrganizer.test.js` |
| 字幕对轴 | `desktop-tool` 占位 | **未实现**；洛水侧有真实面板（bundle `464641-464649` 注册 `subtitle-alignment-progress/complete` 事件），证据线索见 §7 |
| AI-双语字幕 / AI-BGM / AI-Shot | `unavailable` | 未实现，界面已标「未接入」 |
| 声音克隆 / 音色设计 | `desktop-tool` | 依赖桌面执行器，V88 无链路 |
| 草稿合并 / 剪映关键帧 | `desktop-tool`（剪映） | 依赖桌面执行器 |
| AI-CUT插件版 | `route('/ai-cut')` | 已有 AI-CUT 工作台，本批未验收 |

---

## 7. 下一批入口

推荐批次：**字幕对轴**。

洛水侧证据（本轮复核）：

- 处理器 `subtitle-align`：`index-DszfAkKD.js:464641-464649`，初始化 5 个状态与 `{message, percent}` 进度，注册 `subtitle-alignment-progress`、`subtitle-alignment-complete` 两个事件。
- 文案「字幕对轴」出现在 `459664`、`459666`、`464226`、`482541`（`482541` 疑为面板模板，下一批需逐行展开确认控件顺序）。
- 当前 V88 只有 `desktopTool('音视频与字幕对轴执行器')` 占位（`pluginCatalog.js:38` 附近）。

下一批必须先完成：面板模板展开（控件顺序/默认值/选项集合）→ 输入输出规格 → 纯本地对齐算法（可 node 测试）→ 五场景实测 → 洛水对照。

---

## 8. 未验证清单（汇总）

- 界面层全部交互（点击、Toast、双击、上传、并发开关）：构建冻结，未验证。
- 真实模型输出质量与超时：不消耗额度，未验证。
- 取消请求：2026-11-11 已补 AbortController 并通过单测（`pluginGeneration.test.js`）；浏览器里请求是否真的 canceled 未验证。
- 洛水 `build_part_prompt` / `fc/pu/bc` 原始函数：本轮未在 bundle 重新定位，标注为「推导 + 继承证据」。
