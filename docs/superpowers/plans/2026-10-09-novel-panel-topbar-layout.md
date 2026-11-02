# 小说面板顶部布局改造 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 删除视频画面提示词工具顶部说明，并将右侧操作整理为两行右对齐按钮。

**Architecture:** 在现有静态 HTML 中只拆分顶部操作容器，不变更任意按钮 ID、文本或事件绑定。样式只改工作台的 `topbar` 与 `top-actions` 范围，使用 Grid 保证桌面端两行右对齐，小屏端可换行。

**Tech Stack:** 原生 HTML、CSS、现有工作台 JavaScript。

## Global Constraints

- 只修改 `public/novel-panel/workbench/index.html` 和 `public/novel-panel/workbench/style.css`。
- 删除 `eyebrow` 与 `subtitle` 顶部说明，保留页面标题。
- 第一行必须包含 `aiStatus`、版本标识、模式切换；第二行必须包含历史、保存、AI 指令、AI 设置。
- 所有现有按钮 ID 和功能保持不变。
- 两行操作在宽屏右对齐；窄屏可换行但仍从右侧开始排列。
- 不提交、不发布。

---

### Task 1: 调整顶部 HTML 结构

**Files:**
- Modify: `public/novel-panel/workbench/index.html:14-33`
- Test: `test/novel-panel-topbar-layout.test.js`

- [ ] **Step 1: 写失败的结构测试**

```js
import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync('public/novel-panel/workbench/index.html', 'utf8');

test('小说面板顶部只保留标题，并将操作拆成两行', () => {
  assert.doesNotMatch(source, /class="eyebrow"/);
  assert.doesNotMatch(source, /class="subtitle"/);
  assert.match(source, /class="top-actions-row top-actions-row-primary"/);
  assert.match(source, /class="top-actions-row top-actions-row-secondary"/);
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test test/novel-panel-topbar-layout.test.js`

Expected: FAIL，顶部尚未拆成两行。

- [ ] **Step 3: 最小化修改 HTML**

保留 `h1`。删除 `eyebrow` 和 `subtitle` 段落。将 `aiStatus`、版本标识和 `outputModeToggle` 放进 `top-actions-row-primary`；将 `historyBtn`、`saveProjectBtn`、`promptSettingsBtn`、`settingsBtn` 放进 `top-actions-row-secondary`。保留所有既有 ID、按钮文字和隐藏 `projectName` 输入。

- [ ] **Step 4: 运行结构测试确认通过**

Run: `node --test test/novel-panel-topbar-layout.test.js`

Expected: PASS。

### Task 2: 实现两行右对齐样式

**Files:**
- Modify: `public/novel-panel/workbench/style.css:24-33`
- Test: `test/novel-panel-topbar-layout.test.js`

- [ ] **Step 1: 扩展失败测试**

```js
test('小说面板顶部操作区使用右对齐两行网格', () => {
  const css = fs.readFileSync('public/novel-panel/workbench/style.css', 'utf8');
  assert.match(css, /\.top-actions\s*\{[\s\S]*grid-template-columns:\s*1fr/);
  assert.match(css, /\.top-actions-row\s*\{[\s\S]*justify-content:\s*flex-end/);
  assert.match(css, /\.top-actions-row-primary\s*\{/);
  assert.match(css, /\.top-actions-row-secondary\s*\{/);
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test test/novel-panel-topbar-layout.test.js`

Expected: FAIL，新的样式选择器不存在。

- [ ] **Step 3: 编写最小样式**

```css
.top-actions {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  justify-items: end;
}
.top-actions-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}
```

补充统一的按钮高度和移动端规则；不要更改页面其他按钮的外观。

- [ ] **Step 4: 运行测试确认通过**

Run: `node --test test/novel-panel-topbar-layout.test.js`

Expected: PASS。

### Task 3: 浏览器验证

**Files:**
- Verify: `public/novel-panel/workbench/index.html`
- Verify: `public/novel-panel/workbench/style.css`

- [ ] **Step 1: 访问小说面板工作台**

确认顶部不存在两段说明，标题保持左侧。

- [ ] **Step 2: 核对宽屏右对齐**

确认右侧有两行完整操作，第一行模型/版本/模式，第二行历史/保存/AI 指令/AI 设置；两行右边缘对齐。

- [ ] **Step 3: 核对窄屏适配**

确认按钮可换行且没有水平溢出，并保持右对齐。
