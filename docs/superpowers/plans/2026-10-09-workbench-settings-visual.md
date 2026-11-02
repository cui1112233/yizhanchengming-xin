# 工作台设置视觉改造 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将工作台设置调整为冷静专业的深蓝黑界面，统一上排配置、执行器状态区和按钮的视觉层级及对齐方式。

**Architecture:** 仅调整 `SettingsPage.jsx` 中已有控件的容器和类名，不触碰保存、下载、配对、恢复或读取逻辑。把布局规则集中在 `global.css` 的 settings 样式区，复用现有 Ant Design 的表单、按钮和开关组件。

**Tech Stack:** React、Ant Design、Lucide 图标、CSS Grid、Flexbox。

## Global Constraints

- 仅改工作台设置页面；不修改接口、保存逻辑、配对逻辑、下载链接或业务数据。
- 基调是深蓝黑；蓝紫只用于主操作和当前选中；青蓝只用于下载；红色只用于删除。
- 上排两个区块必须使用同一网格基线、内边距、标题高度、字段间距和底部操作区高度。
- 宽屏保持上排两列；窄屏自动改为单列，不压扁控件或按钮。
- 保持现有 Ant Design 控件的可访问名称与功能；交互元素必须保留清晰的 `:focus-visible` 焦点反馈。
- 不创建分支、不提交、不发布，除非用户另外明确要求。

---

### Task 1: 为工作台设置添加布局定位类

**Files:**
- Modify: `frontend/src/user/pages/SettingsPage.jsx:242-420`
- Test: `frontend/src/user/pages/settings-workbench-layout.source.test.js`

**Interfaces:**
- Consumes: 已有 `saveSection`、执行器状态、配对状态及下载链接。
- Produces: 仅增加布局 className；所有按钮的 `onClick`、`href`、`loading`、`disabled` 保持原值。

- [ ] **Step 1: 写出失败的源码结构测试**

```js
import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';

const source = fs.readFileSync(new URL('./SettingsPage.jsx', import.meta.url), 'utf8');

test('工作台设置为上排配置和执行器操作提供对齐容器', () => {
  assert.match(source, /className="settings-top-grid"/);
  assert.match(source, /className="settings-section-footer settings-workspace-actions"/);
  assert.match(source, /className="settings-section-footer settings-storage-actions"/);
  assert.match(source, /className="settings-executor-actions settings-executor-actions-doubao"/);
  assert.match(source, /className="settings-executor-actions settings-executor-actions-giant"/);
});
```

- [ ] **Step 2: 运行测试，确认当前版本失败**

Run: `node --test frontend/src/user/pages/settings-workbench-layout.source.test.js`

Expected: FAIL，提示 `settings-top-grid` 尚不存在。

- [ ] **Step 3: 最小化调整 JSX 结构**

```jsx
<div className="settings-top-grid">
  <section className="settings-section settings-model-section">…</section>
  <section className="settings-section settings-storage-section">…</section>
</div>
```

将宠物保存按钮包进 `settings-section-footer settings-workspace-actions`；将归档操作组保留在 `settings-section-footer settings-storage-actions`；分别为两个执行器操作组添加 `settings-executor-actions-doubao`、`settings-executor-actions-giant`。不改变按钮的文案、事件或链接。

- [ ] **Step 4: 运行结构测试，确认通过**

Run: `node --test frontend/src/user/pages/settings-workbench-layout.source.test.js`

Expected: PASS。

### Task 2: 统一深色调、字段和按钮对齐样式

**Files:**
- Modify: `frontend/src/shared/styles/global.css:2186-2284`
- Test: `frontend/src/user/pages/settings-workbench-layout.source.test.js`

**Interfaces:**
- Consumes: Task 1 新增的布局 className。
- Produces: 上排配置对齐、按钮层级、执行器操作顺序及小屏适配。

- [ ] **Step 1: 扩展失败测试，声明视觉样式钩子**

```js
test('工作台设置样式包含统一对齐、按钮层级和窄屏规则', () => {
  const css = fs.readFileSync(new URL('../../shared/styles/global.css', import.meta.url), 'utf8');
  assert.match(css, /\.settings-top-grid\s*\{/);
  assert.match(css, /\.settings-section-footer\s*\{/);
  assert.match(css, /\.settings-page \.ant-btn:focus-visible\s*\{/);
  assert.match(css, /\.settings-executor-actions-doubao\s*\{/);
  assert.match(css, /\.settings-executor-actions-giant\s*\{/);
});
```

- [ ] **Step 2: 运行测试，确认样式钩子当前失败**

Run: `node --test frontend/src/user/pages/settings-workbench-layout.source.test.js`

Expected: FAIL，提示新增选择器不存在。

- [ ] **Step 3: 只在 settings 样式区实现对齐和色彩规则**

```css
.settings-top-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(280px, .85fr);
  gap: 18px;
  align-items: stretch;
}

.settings-top-grid > .settings-section {
  display: flex;
  min-height: 100%;
  flex-direction: column;
}

.settings-section-footer {
  display: flex;
  min-height: 36px;
  margin-top: auto;
  padding-top: 18px;
  align-items: center;
}
```

在同一 settings 样式区补齐：深蓝黑背景、弱分隔线、蓝紫主按钮、青蓝下载按钮、普通深色描边按钮、统一 36px 按钮高度和 `:focus-visible` 焦点环。豆包和巨量执行器操作组用 CSS `order` 保证下载、生成配对码、刷新状态的展示顺序；删除按钮不改变现有危险色。不得改动全局其他页面的 Ant Design 主题。

- [ ] **Step 4: 添加窄屏覆盖规则**

```css
@container (max-width: 700px) {
  .settings-top-grid { grid-template-columns: minmax(0, 1fr); }
  .settings-section-footer,
  .settings-executor-actions { justify-content: flex-start; }
}
```

- [ ] **Step 5: 运行源码结构测试，确认通过**

Run: `node --test frontend/src/user/pages/settings-workbench-layout.source.test.js`

Expected: PASS。

### Task 3: 构建并在浏览器确认实际对齐

**Files:**
- Verify: `frontend/src/user/pages/SettingsPage.jsx`
- Verify: `frontend/src/shared/styles/global.css`

**Interfaces:**
- Consumes: Task 1 的结构类和 Task 2 的样式规则。
- Produces: 可在 `/settings` 中验证的工作台设置页面。

- [ ] **Step 1: 运行前端构建**

Run: `npm run build`

Expected: exit code 0，生成前端构建产物。

- [ ] **Step 2: 在本地 `/settings` 核对宽屏布局**

确认：标题、字段、按钮底线对齐；宠物和归档区高度一致；主按钮为蓝紫、下载按钮为青蓝、刷新/恢复/查看文件为次级按钮。

- [ ] **Step 3: 在窄宽度核对自适应布局**

确认：上排改为单列，输入框与按钮不溢出，执行器操作仍完整可点击。

- [ ] **Step 4: 运行相关源码测试**

Run: `node --test frontend/src/user/pages/settings-workbench-layout.source.test.js frontend/src/user/pages/settings-retention-source.test.js frontend/src/user/pages/settings-giant-material-executor-source.test.js`

Expected: PASS。
