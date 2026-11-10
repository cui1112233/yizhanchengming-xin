# AI-CUT Chinese Concat-Style Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把现有 AI-CUT 重构成中文、可调整布局、可直接操作画布的专业编辑器，并交付第一批真实滤镜、特效、转场、文字预设和识别字幕链路。

**Architecture:** 前端把 1200 行页面拆为编辑器壳、资源库、画布、属性栏、时间线和字幕工作区，所有资源通过稳定目录 ID 驱动。工程数据继续由 Go 保存，浏览器预览和 FFmpeg 导出共享同一资源 ID 与参数；ASR 继续复用现有 MySQL/Redis 任务链路。

**Tech Stack:** React、Ant Design、Lucide、Go、MySQL、Redis、TOS、FFmpeg、Node test、Go test。

## Global Constraints

- 用户可见名称、状态、错误、参数和快捷键全部使用中文，工程内部使用稳定英文 ID。
- 不复制 Concat 的 AGPL 核心源码、效果资源、模型、字体或品牌；只参考公开布局与操作原则。
- 系统预设提示词统一位于 Go 后端 prompt 模块，前端不得硬编码完整系统提示词。
- 每个可见操作必须产生真实结果、明确禁用原因或配置入口。
- 浏览器预览和 Go/FFmpeg 导出必须读取同一工程参数。
- 未经明确授权不执行付费 ASR 调用，不推送远端，不公开部署。
- 保留工作区已有无关改动，只提交每个任务明确列出的文件。

## File Structure

- `frontend/src/user/pages/AICutPage.jsx`：路由级数据装配，逐步移除工作区具体渲染。
- `frontend/src/user/pages/aicut/editorCatalog.js`：滤镜、特效、转场和文字预设目录。
- `frontend/src/user/pages/aicut/editorLayout.js`：布局默认值、边界和持久化纯函数。
- `frontend/src/user/pages/aicut/EditorShell.jsx`：顶栏、三列上部区和底部时间线。
- `frontend/src/user/pages/aicut/ResourceLibrary.jsx`：六类中文资源库。
- `frontend/src/user/pages/aicut/PreviewCanvas.jsx`：播放画面和直接变换操作。
- `frontend/src/user/pages/aicut/InspectorPanel.jsx`：上下文属性编辑。
- `frontend/src/user/pages/aicut/TimelineEditor.jsx`：轨道、片段、转场和效果层。
- `frontend/src/user/pages/aicut/CaptionWorkspace.jsx`：ASR、SRT 和手动字幕。
- `frontend/src/user/pages/aicut/editor.css`：新编辑器样式，只使用全局主题变量。
- `backend/internal/aicut/types.go`：工程结构与目录 ID 校验。
- `backend/internal/httpapi/aicut_export.go`：FFmpeg 滤镜、转场、文字和字幕导出映射。

---

### Task 1: 稳定资源目录

**Files:**
- Create: `frontend/src/user/pages/aicut/editorCatalog.js`
- Create: `frontend/src/user/pages/aicut/editorCatalog.test.js`
- Modify: `frontend/src/user/pages/aicutEffects.js`

**Interfaces:**
- Produces: `editorCatalog`, `catalogById(id)`, `catalogForTab(tab)`, `normalizeCatalogParameters(id, values)`。
- Resource shape: `{ id, labelZh, category, description, parameters, supportedKinds, previewAdapter, renderAdapter }`。

- [ ] **Step 1: Write the failing catalogue tests**

验证 16 个滤镜、12 个特效、16 个转场和 12 个文字预设；所有 ID 唯一，中文名称非空，参数有边界，资源具备浏览器和服务端映射。

- [ ] **Step 2: Run the test and verify failure**

Run: `node --test src/user/pages/aicut/editorCatalog.test.js`

- [ ] **Step 3: Implement the catalogue and normalization**

使用声明式数组，参数规范为 `{ key, labelZh, type, defaultValue, min, max, step }`；`normalizeCatalogParameters` 截断数值并丢弃未知字段。

- [ ] **Step 4: Run catalogue and existing effect tests**

Run: `node --test src/user/pages/aicut/editorCatalog.test.js src/user/pages/aicutEffects.test.js`

- [ ] **Step 5: Commit only catalogue files**

Commit: `feat: add AI-CUT Chinese resource catalogue`

### Task 2: 可保存的编辑器骨架

**Files:**
- Create: `frontend/src/user/pages/aicut/editorLayout.js`
- Create: `frontend/src/user/pages/aicut/editorLayout.test.js`
- Create: `frontend/src/user/pages/aicut/EditorShell.jsx`
- Create: `frontend/src/user/pages/aicut/editor.css`
- Modify: `frontend/src/user/pages/AICutPage.jsx`

**Interfaces:**
- Produces: `defaultEditorLayout`, `normalizeEditorLayout(layout, viewport)`, `toggleEditorPane(layout, pane)`。
- `EditorShell` props: `{ projectHeader, resourceLibrary, previewCanvas, inspector, timeline, agentLauncher, layout, onLayoutChange }`。

- [ ] **Step 1: Write failing layout tests**

覆盖左右面板和时间线的最小/最大尺寸、收起恢复、窄窗口属性抽屉、刷新反序列化和非法值回退。

- [ ] **Step 2: Verify the tests fail**

Run: `node --test src/user/pages/aicut/editorLayout.test.js`

- [ ] **Step 3: Implement pure layout functions**

默认比例：资源库 22%、属性栏 22%、时间线 40%；桌面最小宽度分别为 260、280、225 像素；小于 1180 像素时属性栏进入抽屉模式。

- [ ] **Step 4: Implement EditorShell and migrate top-level layout**

用 CSS 变量保存三个尺寸；拖动手柄只更新临时布局，松手写入 `aicut.editorLayout.v3`；所有面板保留中文 `aria-label`。

- [ ] **Step 5: Run layout and source-contract tests**

Run: `node --test src/user/pages/aicut/editorLayout.test.js src/user/pages/aicut-workbench.layout.test.js`

- [ ] **Step 6: Commit editor skeleton**

Commit: `feat: rebuild AI-CUT editor workspace`

### Task 3: 中文资源库直接操作

**Files:**
- Create: `frontend/src/user/pages/aicut/ResourceLibrary.jsx`
- Create: `frontend/src/user/pages/aicut/ResourceLibrary.source.test.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut/editor.css`

**Interfaces:**
- Consumes: `editorCatalog`, `catalogForTab(tab)`。
- Produces callbacks: `onPreviewResource(resource)`, `onApplyResource(resource)`, `onDragResource(resource, event)`。

- [ ] **Step 1: Write source-contract tests**

断言六个中文页签存在；资源项提供搜索、收藏、最近使用、点击预览、双击应用和拖动；旧的“功能说明卡→大弹窗”入口不再出现。

- [ ] **Step 2: Verify failure**

Run: `node --test src/user/pages/aicut/ResourceLibrary.source.test.js`

- [ ] **Step 3: Implement ResourceLibrary**

媒体显示上传素材；其他页签显示目录缩略卡。临时预览保存在页面状态，应用时才调用 `commitTimeline`。

- [ ] **Step 4: Wire existing upload, ASR and timeline callbacks**

视频/音频卡片保留“识别字幕”；无 ASR 模型时仍打开识别窗口并显示配置入口。

- [ ] **Step 5: Run tests**

Run: `node --test src/user/pages/aicut/ResourceLibrary.source.test.js src/user/pages/aicut-controls.contract.test.js src/user/pages/aicutCapabilities.test.js`

- [ ] **Step 6: Commit resource library**

Commit: `feat: add direct AI-CUT resource library`

### Task 4: 预览画布直接变换

**Files:**
- Create: `frontend/src/user/pages/aicut/canvasTransform.js`
- Create: `frontend/src/user/pages/aicut/canvasTransform.test.js`
- Create: `frontend/src/user/pages/aicut/PreviewCanvas.jsx`
- Modify: `frontend/src/user/pages/AICutPage.jsx`

**Interfaces:**
- Produces: `normalizeTransform`, `dragTransform`, `resizeTransform`, `rotateTransform`。
- Transform shape: `{ x, y, scaleX, scaleY, rotation, opacity }`，位置使用归一化画布坐标。

- [ ] **Step 1: Write transform boundary tests**

覆盖拖动、四角缩放、旋转、负尺寸防护、画布边界、安全框和取消拖动不提交。

- [ ] **Step 2: Verify failure**

Run: `node --test src/user/pages/aicut/canvasTransform.test.js`

- [ ] **Step 3: Implement transform pure functions**

限制 scale 为 `0.05..20`、opacity 为 `0..1`、rotation 规范到 `-180..180`，按画布实际尺寸换算指针移动。

- [ ] **Step 4: Implement PreviewCanvas**

显示选择框、四角缩放柄和旋转柄；拖动中只更新 echo，松手调用一次 `onCommitTransform`；Escape 取消。

- [ ] **Step 5: Run preview and transform tests**

Run: `node --test src/user/pages/aicut/canvasTransform.test.js src/user/pages/aicutPreview.test.js`

- [ ] **Step 6: Commit canvas direct manipulation**

Commit: `feat: add direct manipulation to AI-CUT canvas`

### Task 5: 上下文中文属性栏

**Files:**
- Create: `frontend/src/user/pages/aicut/InspectorPanel.jsx`
- Create: `frontend/src/user/pages/aicut/InspectorPanel.source.test.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`

**Interfaces:**
- Consumes: `{ selection, resource, projectSettings }`。
- Emits: `onPatchClip(patch)`, `onPatchTrack(patch)`, `onPatchEffect(patch)`, `onPatchTransition(patch)`。

- [ ] **Step 1: Write context tests**

验证未选中、视频、图片、文字、字幕、音频、转场和效果层分别显示正确中文分组；窄屏渲染抽屉入口。

- [ ] **Step 2: Verify failure**

Run: `node --test src/user/pages/aicut/InspectorPanel.source.test.js`

- [ ] **Step 3: Implement context panels**

数字字段在 `onBlur` 或 Enter 时提交，滑杆拖动用 echo 预览，松手形成一次历史；未知目录 ID 只读显示，不破坏工程。

- [ ] **Step 4: Run inspector tests**

Run: `node --test src/user/pages/aicut/InspectorPanel.source.test.js src/user/pages/aicut-controls.contract.test.js`

- [ ] **Step 5: Commit inspector**

Commit: `feat: add contextual AI-CUT inspector`

### Task 6: 真实效果层和转场时间线

**Files:**
- Create: `frontend/src/user/pages/aicut/TimelineEditor.jsx`
- Modify: `frontend/src/user/pages/aicutTimeline.js`
- Modify: `frontend/src/user/pages/aicutTimeline.test.js`
- Modify: `backend/internal/aicut/types.go`
- Modify: `backend/internal/aicut/visual_effect_test.go`

**Interfaces:**
- Produces pure operations: `addEffectLayer`, `moveEffectLayer`, `resizeEffectLayer`, `setClipTransition`, `removeTimelineObject`。
- Effect layer shape: `{ id, kind:'effect', resourceId, startMs, durationMs, parameters }`。

- [ ] **Step 1: Add failing timeline tests**

覆盖效果层添加/移动/裁剪、锁定轨拒绝、同轨相邻转场、空隙拒绝、过长转场截断和撤销粒度。

- [ ] **Step 2: Verify failure**

Run: `node --test src/user/pages/aicutTimeline.test.js`

- [ ] **Step 3: Implement pure timeline operations**

所有函数返回新对象和 `{ ok, reasonZh }`，不得原地修改；转场持续时间限制为相邻两个片段可用时长的一半以内。

- [ ] **Step 4: Implement TimelineEditor UI**

渲染文字、效果、视频、音频、字幕轨；拖动资源到目标轨；转场显示在连接点；效果显示独立区间。

- [ ] **Step 5: Extend Go validation**

Go 接受同一稳定 ID 集合和参数边界，拒绝未知 ID、负持续时间和非法转场位置。

- [ ] **Step 6: Run frontend and Go tests**

Run: `node --test src/user/pages/aicutTimeline.test.js`

Run: `go test ./internal/aicut`

- [ ] **Step 7: Commit timeline model and UI**

Commit: `feat: add AI-CUT effect layers and transitions`

### Task 7: 字幕识别工作区

**Files:**
- Create: `frontend/src/user/pages/aicut/CaptionWorkspace.jsx`
- Create: `frontend/src/user/pages/aicut/CaptionWorkspace.source.test.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `backend/internal/aicut/asr_service.go`
- Modify: `backend/internal/aicut/asr_service_test.go`

**Interfaces:**
- Consumes existing ASR API functions and `asrStateLabels`。
- Emits: `onOpenASR`, `onImportSRT`, `onExportSRT`, `onAddManualCaption`, `onApplyCaptionStyle`。

- [ ] **Step 1: Write ASR workspace tests**

验证素材入口、无模型配置入口、语言选择、排队/识别/取消/成功/失败状态、重复保护、失败重试和不覆盖字幕。

- [ ] **Step 2: Verify failure**

Run: `node --test src/user/pages/aicut/CaptionWorkspace.source.test.js`

- [ ] **Step 3: Implement CaptionWorkspace**

把识别字幕从通用弹窗迁入左侧文字工作区；任务列表常驻；只有模型、素材和语言有效时允许开始识别。

- [ ] **Step 4: Complete backend empty-result and conflict messages**

空 cues 返回“没有识别到有效语音”；工程版本变化返回可识别冲突错误；取消后不写入轨道。

- [ ] **Step 5: Run frontend and backend tests**

Run: `node --test src/user/pages/aicut/CaptionWorkspace.source.test.js src/user/pages/aicut-controls.contract.test.js`

Run: `go test ./internal/aicut ./internal/httpapi`

- [ ] **Step 6: Commit caption workspace**

Commit: `feat: integrate caption recognition workspace`

### Task 8: FFmpeg 导出目录映射

**Files:**
- Modify: `backend/internal/httpapi/aicut_export.go`
- Modify: `backend/internal/httpapi/aicut_export_test.go`
- Create: `docs/licenses/aicut-editor-third-party.md`

**Interfaces:**
- Consumes the stable IDs from Task 1 mirrored in Go constants.
- Produces FFmpeg filters and xfade names for all first-batch usable resources.

- [ ] **Step 1: Write table-driven export tests**

每个滤镜、特效和转场至少验证一次 filter fragment；未知 ID 必须失败；参数必须被截断到目录范围。

- [ ] **Step 2: Verify failure**

Run: `go test ./internal/httpapi -run 'TestAICut.*(Filter|Effect|Transition|Export)'`

- [ ] **Step 3: Implement mappings**

用小函数按类别生成 FFmpeg 片段，不在 HTTP handler 内拼接巨型条件；文字和字幕继续生成 ASS 样式。

- [ ] **Step 4: Validate every generated filter with local FFmpeg**

对一段本地测试色条逐一运行生成的滤镜，要求全部退出码为 0；不执行网络请求。

- [ ] **Step 5: Record third-party inventory**

记录 FFmpeg、字体、图标和未来可能采用的依赖许可证；注明未复制 Concat 代码和资源。

- [ ] **Step 6: Run backend tests**

Run: `go test ./internal/aicut ./internal/httpapi`

- [ ] **Step 7: Commit export mappings**

Commit: `feat: export AI-CUT Chinese editor resources`

### Task 9: 主题、构建和浏览器验收

**Files:**
- Modify: `frontend/src/user/pages/aicut/editor.css`
- Modify: `frontend/src/user/pages/aicut-workbench.layout.test.js`
- Create: `docs/acceptance/aicut-concat-chinese-editor.md`

**Interfaces:**
- Consumes all prior tasks.
- Produces final local 4001 acceptance evidence.

- [ ] **Step 1: Add theme and responsive tests**

检查所有编辑器颜色来自主题变量；AntD 浮层继承主题；窄屏属性抽屉存在；禁止硬编码大面积白色弹窗。

- [ ] **Step 2: Run complete focused frontend tests**

Run: `node --test src/user/pages/aicut/*.test.js src/user/pages/aicut*.test.js`

- [ ] **Step 3: Run frontend production build**

Run: `npm run build`

- [ ] **Step 4: Run backend tests and build embedded runtime**

Run: `go test ./internal/aicut ./internal/httpapi`

Run: `powershell -File scripts/build-go-local.ps1`

- [ ] **Step 5: Restart only the confirmed local runtime**

确认 4001 监听进程路径为当前仓库 `backend/qiantie-local.exe` 后停止并运行 `scripts/start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached`。

- [ ] **Step 6: Browser acceptance**

逐项验证中文页签、资源直接预览、应用、属性修改、时间线拖动、收起恢复、主题切换、识别字幕配置入口、取消操作和刷新持久化。不得用源代码字符串代替可见结果。

- [ ] **Step 7: Record evidence and commit acceptance document**

Commit: `docs: record AI-CUT Chinese editor acceptance`
