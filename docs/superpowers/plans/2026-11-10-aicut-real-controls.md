# AI-CUT Real Controls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 AI-CUT 页面只显示真实可用的操作，并打通首批画面效果从预览、保存到 FFmpeg 导出的完整链路。

**Architecture:** 前端使用统一能力注册表决定工具入口是否显示以及如何执行，删除散落的字符串判断和提示型假入口。片段画面效果使用结构化字段贯穿 React 播放器、Go 校验、MySQL 和 FFmpeg；未接入的外部能力隐藏。

**Tech Stack:** React、Ant Design、Vitest、Go、MySQL、FFmpeg、Go embed。

## Global Constraints

- 不复制洛水或其他项目的私有源码、密钥、私有接口、提示词和专有资源。
- 不新建分支，不提交、不推送、不部署，不覆盖当前脏工作区的无关修改。
- 系统提示词只允许放在 Go 后端 prompt 模块，前端不得硬编码。
- 页面只显示具有真实动作和明确失败处理的入口。
- 任何完成结论必须同时给出源码测试、构建、嵌入运行时和 4001 可见行为证据。

---

### Task 1: 建立能力注册表并清除假入口

**Files:**
- Create: `frontend/src/user/pages/aicutCapabilities.js`
- Create: `frontend/src/user/pages/aicutCapabilities.test.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-page.css`

**Interfaces:**
- Produces: `visibleAICutCategories(context)`, `visibleAICutFeatures(category, context)` 和稳定能力 ID。
- Consumes: 页面已有项目、素材、选中片段和模型配置状态。

- [ ] **Step 1: 写失败测试**

```js
it('never exposes a feature without a real action', () => {
  const features = visibleAICutFeatures('特效', { handlers: {} });
  expect(features).toEqual([]);
});

it('shows supported visual effects and hides unconfigured AI effects', () => {
  const features = visibleAICutFeatures('特效', {
    handlers: { openVisualEffects() {} },
  });
  expect(features.map(item => item.id)).toEqual(['visual-effects']);
});
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- --run src/user/pages/aicutCapabilities.test.js`

Expected: FAIL，因为能力注册表尚不存在。

- [ ] **Step 3: 实现最小注册表**

注册表每项包含 `id/category/title/detail/source/handler/visible/disabledReason`，过滤时要求 handler 存在。把未接入类别从工具轨隐藏；删除 `message.info('需要配置服务')` 兜底。

- [ ] **Step 4: 接入收藏和最近使用**

使用 `aicut.favoriteFeatures` 和 `aicut.recentFeatures` 两个 localStorage 键。只有 handler 成功返回后才写最近使用；收藏按钮必须可切换。

- [ ] **Step 5: 运行测试**

Run: `npm test -- --run src/user/pages/aicutCapabilities.test.js src/user/pages/aicut-workbench.layout.test.js`

Expected: PASS。

### Task 2: 顶栏和工具栏状态真实化

**Files:**
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicutWorkbench.js`
- Test: `frontend/src/user/pages/aicutWorkbench.test.js`

**Interfaces:**
- Produces: `saveState: 'saved' | 'dirty' | 'saving' | 'failed'`。
- Consumes: 所有 `commitTimeline` 调用和 `saveTimeline` 返回值。

- [ ] **Step 1: 添加保存状态测试**

验证本地修改后显示“有未保存修改”，保存请求中显示“保存中”，成功后显示“云端已保存”，失败后显示“保存失败”。

- [ ] **Step 2: 实现保存状态机**

所有时间线本地变更统一标记 dirty；保存开始、成功、失败分别更新状态。切换项目时从 saved 开始。

- [ ] **Step 3: 删除禁用假按钮**

从时间线工具栏移除关键帧、定格和多时间线；从属性栏移除抠像、蒙版、美化、调色、动画和跟踪占位标签。快捷键弹窗只列出 `resolveAICutShortcut` 真正支持的快捷键。

- [ ] **Step 4: 运行测试**

Run: `npm test -- --run src/user/pages/aicutWorkbench.test.js src/user/pages/aicut-workbench.layout.test.js`

Expected: PASS，源码中不再出现三个禁用按钮和六个占位标签。

### Task 3: 画面效果数据模型和校验

**Files:**
- Modify: `backend/internal/aicut/types.go`
- Modify: `backend/internal/aicut/validation.go`
- Modify: `backend/internal/aicut/mysql_store.go`
- Modify: `backend/internal/storage/aicut_schema.go`
- Test: `backend/internal/aicut/store_test.go`
- Test: `backend/internal/storage/aicut_schema_test.go`

**Interfaces:**
- Produces: `VisualEffect`、`NormalizeVisualEffect(VisualEffect) (VisualEffect, error)`。
- Consumes: 时间线保存和读取。

- [ ] **Step 1: 写失败的边界测试**

覆盖默认值、合法预设、亮度/对比度/饱和度/模糊边界、NaN、无穷值和未知预设。

- [ ] **Step 2: 运行 Go 测试并确认失败**

Run: `go test ./internal/aicut ./internal/storage`

Expected: FAIL，因为字段和迁移尚不存在。

- [ ] **Step 3: 添加结构和迁移**

`Clip` 新增 `VisualEffect VisualEffect \`json:"visualEffect,omitempty"\``；迁移新增 `visual_effect_json JSON NULL`。MySQL SELECT/INSERT 必须对称编码解码。

- [ ] **Step 4: 添加校验**

只接受 `original/bright/cool-cinema/warm/grayscale/soft-focus/custom`，数值超界返回 `ErrInvalidTimeline`，不静默写入坏数据。

- [ ] **Step 5: 运行 Go 测试**

Run: `go test ./internal/aicut ./internal/storage`

Expected: PASS。

### Task 4: 浏览器预览和画面效果面板

**Files:**
- Create: `frontend/src/user/pages/aicutEffects.js`
- Create: `frontend/src/user/pages/aicutEffects.test.js`
- Modify: `frontend/src/user/pages/aicutPreview.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-page.css`

**Interfaces:**
- Produces: `normalizeVisualEffect`, `visualEffectCSS`, `visualEffectPreset`。
- Consumes: 后端 `visualEffect` JSON。

- [ ] **Step 1: 写预设和 CSS 映射失败测试**

```js
expect(visualEffectCSS({ brightness: .2, contrast: 1.1, saturation: .8, blur: 2, grayscale: false }))
  .toContain('brightness(1.2)');
```

同时覆盖原片、重复选择同一预设和非法值回退。

- [ ] **Step 2: 实现纯函数并运行测试**

Run: `npm test -- --run src/user/pages/aicutEffects.test.js src/user/pages/aicutPreview.test.js`

Expected: PASS。

- [ ] **Step 3: 实现真实效果面板**

“画面特效”仅在选择图片/视频片段时启用。面板提供六个预设及亮度、对比度、饱和度、模糊、黑白控件；取消关闭不修改，应用更新当前片段并进入撤销栈。

- [ ] **Step 4: 接入播放器**

`PreviewMedia` 对对应片段应用 `visualEffectCSS`。锁定轨道时面板只读并给出“轨道已锁定”。

- [ ] **Step 5: 运行前端测试与构建**

Run: `npm test -- --run src/user/pages/aicutEffects.test.js src/user/pages/aicutPreview.test.js src/user/pages/aicutTimeline.test.js`

Run: `npm run build`

Expected: 全部 PASS，构建成功。

### Task 5: FFmpeg 导出一致性

**Files:**
- Modify: `backend/internal/httpapi/aicut_export.go`
- Test: `backend/internal/httpapi/aicut_export_test.go`

**Interfaces:**
- Produces: `buildAICutVisualFilter(effect aicut.VisualEffect) []string`。
- Consumes: Task 3 的 `VisualEffect`。

- [ ] **Step 1: 写失败测试**

同一片段设置亮度、对比度、饱和度、模糊和黑白后，断言 FFmpeg filter graph 包含 `eq`、`gblur` 和 `hue=s=0`；原片不得增加多余滤镜。

- [ ] **Step 2: 运行测试并确认失败**

Run: `go test ./internal/httpapi -run AICutExport`

Expected: FAIL，因为导出尚未读取效果。

- [ ] **Step 3: 实现过滤器构建**

在每个视觉片段 scale/pad 前插入由结构化参数生成的白名单过滤器，不允许把用户字符串直接拼进 FFmpeg 命令。

- [ ] **Step 4: 运行测试**

Run: `go test ./internal/httpapi -run AICutExport`

Expected: PASS。

### Task 6: 全量按钮合同测试

**Files:**
- Create: `frontend/src/user/pages/aicut-controls.contract.test.js`
- Modify: `docs/acceptance/aicut-controls-matrix.md`

**Interfaces:**
- Consumes: 能力注册表和页面源码。
- Produces: 可见按钮到 handler、数据变化、保存和导出的验收矩阵。

- [ ] **Step 1: 添加合同测试**

遍历所有可见能力，断言存在 handler、空值提示和成功结果；扫描页面源码，拒绝 `disabled title="尚未配置` 和 `需要先配置对应的在线处理服务`。

- [ ] **Step 2: 执行正常、空值、重复、错误和取消测试**

Run: `npm test -- --run src/user/pages/aicutCapabilities.test.js src/user/pages/aicutEffects.test.js src/user/pages/aicut-controls.contract.test.js`

Run: `go test ./internal/aicut ./internal/httpapi ./internal/storage`

Expected: 全部 PASS。

### Task 7: 嵌入构建和 4001 可见验收

**Files:**
- Modify: `backend/current-release.json`（只由现有构建流程生成）
- Update: `docs/acceptance/aicut-controls-matrix.md`

**Interfaces:**
- Consumes: 前端 dist 和 Go embed。
- Produces: 本地 4001 可见运行证据。

- [ ] **Step 1: 构建前端并嵌入 Go**

使用项目既有本地构建脚本，不调用 GitHub Actions，不上传部署。

- [ ] **Step 2: 重启确认过路径的本地 `qiantie-local.exe`**

确认监听 4001 的进程路径、`/health`、`/api/runtime-build-info`、前端入口文件和二进制 SHA256。

- [ ] **Step 3: 浏览器同组输入验收**

对画面效果执行正常、无选择、重复应用、锁轨错误和取消；保存后刷新确认恢复；导出测试文件后确认任务成功。逐一点击所有可见按钮，不能出现无动作或通用占位提示。

- [ ] **Step 4: 记录未验证项**

真实付费模型、未配置外部服务和公网部署不得标记完成。
