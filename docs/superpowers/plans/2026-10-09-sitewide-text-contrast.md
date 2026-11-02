# 全站文字对比度修复 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让用户前台和管理后台在深色、浅色主题下都清楚显示文字、表单、按钮、表格和状态信息，并杜绝透明色隐藏业务文字。

**Architecture:** 在全局样式中建立前景、背景、边框和状态的语义颜色变量；用户前台、管理后台和 Ant Design 使用同一套深浅主题等级。按公共层、账号中心、创作页、高密度任务页和管理端分批替换硬编码颜色，只改可读性，不改变接口、布局或业务流程。

**Tech Stack:** React、Ant Design、CSS 自定义变量、Vite、Node test。

## Global Constraints

- 正式开发目录是 `f:\脚本测试\一战晟铭-v88` 的既有 `v88` 分支；不创建新分支，不修改公网运行文件。
- 本次只修文字、边框、输入框和状态显示的对比度；不得修改接口、业务流程、任务状态语义、后端配置或页面布局。
- 用户前台和管理后台均使用 React + Ant Design；主题颜色必须同时覆盖深色与浅色模式。
- 系统提示词仍只由后端管理；前端不得新增硬编码系统提示词正文。
- 普通正文、表单标签、输入值、按钮、链接和表格文字必须清晰可读；辅助文字可弱化但不得融入背景。
- 真实标题、状态和正文不得使用 `color: transparent`；仅装饰性渐变文字可使用透明填充，且必须有纯色后备显示。
- 每个任务先写断言或测试，再做最小样式改动；完成后运行对应测试和前端构建。
- 不提交 `.env.v88-review` 或 `.local-review/`。

---

## File Structure

- `frontend/src/shared/styles/global.css`：全局深浅主题语义颜色、公共导航、表单、弹窗和状态文字规则。
- `frontend/src/shared/styles/theme.js`：Ant Design 深浅主题 token，与 CSS 语义变量对应。
- `frontend/src/shared/styles/account-center-visual-rebuild.css`：账号与设置相关页面的私有样式，改为引用语义变量。
- `frontend/src/shared/styles/login-card.css`：登录页文字、输入框、提示与错误状态颜色。
- `frontend/src/shared/styles/agent-workspace.css`：Agent 页面文字、任务行和空状态颜色。
- `frontend/src/shared/styles/team-governance.css`：团队治理页面文字、标签和状态颜色。
- `frontend/src/shared/styles/member-center.css`：个人中心资料、会员、安全页的文字颜色。
- `frontend/src/user/pages/novel-fetch.css`：小说获取页文字和任务状态颜色。
- `frontend/src/user/pages/shuihuo-production.css`：水货生产页面和批量工厂入口文字颜色。
- `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-theme.css`：批量工厂状态和工作台主题颜色。
- `frontend/src/user/App.theme.test.js`：新增主题变量和 Ant Design token 的回归断言。
- `frontend/src/shared/styles/contrast-contract.test.js`：新增静态样式约束测试，禁止业务文本透明色并断言语义变量存在。

---

### Task 1: 建立全局深浅主题对比度契约

**Files:**
- Create: `frontend/src/shared/styles/contrast-contract.test.js`
- Create: `frontend/src/user/App.theme.test.js`
- Modify: `frontend/src/shared/styles/global.css:7-78`
- Modify: `frontend/src/shared/styles/theme.js:1-39`

**Interfaces:**
- Produces: CSS 变量 `--text-primary`、`--text-secondary`、`--text-tertiary`、`--text-disabled`、`--text-link`、`--text-on-accent`、`--text-danger`、`--text-warning`、`--text-success`、`--surface-base`、`--surface-panel`、`--surface-input`、`--border-default`、`--border-strong`。
- Produces: `createAntTheme('dark')` 和 `createAntTheme('light')` 都返回正文、辅助、弱提示、禁用、背景和边框 token。

- [ ] **Step 1: 写失败的样式契约测试**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';

const root = path.resolve(import.meta.dirname, '..', '..', '..');
const globalCSS = fs.readFileSync(path.join(root, 'shared/styles/global.css'), 'utf8');

for (const name of [
  '--text-primary', '--text-secondary', '--text-tertiary', '--text-disabled',
  '--text-link', '--text-on-accent', '--text-danger', '--text-warning',
  '--text-success', '--surface-base', '--surface-panel', '--surface-input',
  '--border-default', '--border-strong'
]) {
  test(`global theme defines ${name}`, () => {
    assert.match(globalCSS, new RegExp(`${name}\\s*:`));
  });
}

test('account center does not hide its page title with transparent text', () => {
  const css = fs.readFileSync(path.join(root, 'shared/styles/account-center-visual-rebuild.css'), 'utf8');
  assert.doesNotMatch(css, /\.account-center-workspace \.legacy-page-title\s*\{[^}]*color:\s*transparent/s);
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Expected: FAIL，因为语义变量和透明标题修复尚不存在。

- [ ] **Step 3: 写 Ant Design 主题断言**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import { createAntTheme } from '../shared/styles/theme.js';

test('dark theme exposes readable text levels', () => {
  const { token } = createAntTheme('dark');
  assert.equal(token.colorText, '#f5f7fb');
  assert.equal(token.colorTextSecondary, '#c7d0dd');
  assert.equal(token.colorTextTertiary, '#a7b4c5');
  assert.equal(token.colorTextDisabled, 'rgba(199, 208, 221, 0.62)');
});

test('light theme exposes readable text levels', () => {
  const { token } = createAntTheme('light');
  assert.equal(token.colorText, '#172033');
  assert.equal(token.colorTextSecondary, '#3f4e63');
  assert.equal(token.colorTextTertiary, '#596a80');
  assert.equal(token.colorTextDisabled, '#748196');
});
```

- [ ] **Step 4: 运行主题测试确认失败**

Run: `npm --prefix frontend test -- src/user/App.theme.test.js`

Expected: FAIL，因为当前 `theme.js` 还使用旧颜色值。

- [ ] **Step 5: 实现全局语义变量和 Ant Design token**

在 `:root` 添加以下深色变量：

```css
--text-primary: #f5f7fb;
--text-secondary: #c7d0dd;
--text-tertiary: #a7b4c5;
--text-disabled: rgba(199, 208, 221, 0.62);
--text-link: #93b4ff;
--text-on-accent: #ffffff;
--text-danger: #ff9b9d;
--text-warning: #ffd08a;
--text-success: #75ddb4;
--surface-base: #070b11;
--surface-panel: #0e1621;
--surface-input: #0b111a;
--border-default: rgba(151, 168, 190, 0.22);
--border-strong: rgba(176, 195, 219, 0.36);
```

在 `[data-theme='light']` 添加以下浅色变量：

```css
--text-primary: #172033;
--text-secondary: #3f4e63;
--text-tertiary: #596a80;
--text-disabled: #748196;
--text-link: #185bb8;
--text-on-accent: #ffffff;
--text-danger: #b4232d;
--text-warning: #8a4b00;
--text-success: #087a4b;
--surface-base: #f4f6fa;
--surface-panel: #ffffff;
--surface-input: #ffffff;
--border-default: #c8d2df;
--border-strong: #9daabd;
```

随后让已有 `--legacy-*` 和 `--admin-*` 映射到语义变量；将 `theme.js` 中的文字、背景、输入框和边框 token 同步为测试中的值。

- [ ] **Step 6: 修复透明业务标题**

删除 `account-center-visual-rebuild.css` 中 `.account-center-workspace .legacy-page-title { color: transparent; ... }` 规则，只保留 `color: var(--text-secondary)`。

- [ ] **Step 7: 运行测试和构建**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js src/user/App.theme.test.js`

Run: `npm --prefix frontend run build`

Expected: 全部 PASS。

- [ ] **Step 8: 提交**

```bash
git add frontend/src/shared/styles/global.css frontend/src/shared/styles/theme.js frontend/src/shared/styles/contrast-contract.test.js frontend/src/user/App.theme.test.js frontend/src/shared/styles/account-center-visual-rebuild.css
git commit -m "fix(ui): establish readable theme contrast tokens"
```

### Task 2: 修复登录、公共壳和账号中心文字对比度

**Files:**
- Modify: `frontend/src/shared/styles/login-card.css`
- Modify: `frontend/src/shared/styles/account-center-visual-rebuild.css`
- Modify: `frontend/src/shared/styles/member-center.css`
- Modify: `frontend/src/shared/styles/team-governance.css`
- Modify: `frontend/src/shared/styles/team-collaboration.css`
- Modify: `frontend/src/shared/styles/global.css`
- Test: `frontend/src/shared/styles/contrast-contract.test.js`

**Interfaces:**
- Consumes: Task 1 的语义颜色变量。
- Produces: 登录、侧边栏、顶部栏、个人中心、安全、会员、团队、模型配置页面的文字颜色只使用语义变量或业务状态色。

- [ ] **Step 1: 扩展失败的静态样式测试**

```js
test('account and login styles do not hide business text with transparent color', () => {
  for (const file of [
    'shared/styles/login-card.css',
    'shared/styles/account-center-visual-rebuild.css',
    'shared/styles/member-center.css',
    'shared/styles/team-governance.css',
    'shared/styles/team-collaboration.css'
  ]) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.doesNotMatch(css, /(?:color|-webkit-text-fill-color):\s*transparent/);
  }
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Expected: FAIL，因为账号中心目前仍存在透明标题规则，或登录/账户样式仍使用不符合约束的透明业务字色。

- [ ] **Step 3: 最小替换账号和登录文字色**

- 登录标题、说明、标签、复选框、输入值、占位符、错误提示分别使用 `--text-primary`、`--text-secondary`、`--text-tertiary`、`--text-danger`。
- 个人中心侧边栏、顶部栏、页面标题、卡片正文、说明、表格、输入框和标签分别使用 `--text-primary`、`--text-secondary`、`--text-tertiary`、状态变量。
- 禁用按钮、不可用功能和空状态使用 `--text-disabled`，不得用低于 0.5 的文字透明度。
- 保留红、黄、绿业务状态，但替换为 `--text-danger`、`--text-warning`、`--text-success` 或可读的标签专用状态色。

- [ ] **Step 4: 运行测试与构建**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Run: `npm --prefix frontend run build`

Expected: PASS。

- [ ] **Step 5: 浏览器核对深浅主题**

依次访问 `/settings`、`/profile`、`/member-center`、`/security`、`/team`、`/api-config`；在深色和浅色主题下检查页面标题、表单标签、输入值、占位符、保存按钮、禁用文字、表格文字和状态标签。

- [ ] **Step 6: 提交**

```bash
git add frontend/src/shared/styles/login-card.css frontend/src/shared/styles/account-center-visual-rebuild.css frontend/src/shared/styles/member-center.css frontend/src/shared/styles/team-governance.css frontend/src/shared/styles/team-collaboration.css frontend/src/shared/styles/global.css frontend/src/shared/styles/contrast-contract.test.js
git commit -m "fix(ui): improve account and login text contrast"
```

### Task 3: 修复创作、小说与通用工作台页面文字对比度

**Files:**
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Modify: `frontend/src/user/pages/giant-material-test.css`
- Modify: `frontend/src/user/pages/GiantMaterialStatusCard.css`
- Modify: `frontend/src/shared/styles/global.css`
- Test: `frontend/src/shared/styles/contrast-contract.test.js`

**Interfaces:**
- Consumes: Task 1 的语义颜色变量。
- Produces: Agent、小说获取、小说面板入口、巨量素材状态卡和公共创作工作台的正文、辅助说明、状态和空状态具备深浅主题可读性。

- [ ] **Step 1: 写创作页语义变量使用测试**

```js
test('creative-workspace style files use global readable text tokens', () => {
  for (const file of [
    'shared/styles/agent-workspace.css',
    'user/pages/novel-fetch.css',
    'user/pages/giant-material-test.css',
    'user/pages/GiantMaterialStatusCard.css'
  ]) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.match(css, /var\(--text-(?:primary|secondary|tertiary|disabled|danger|warning|success)\)/);
  }
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Expected: FAIL，因为至少一个页面文件尚未引用全局文字变量。

- [ ] **Step 3: 最小替换创作页文字层级**

- 页面标题、任务名称、主要内容使用 `--text-primary`。
- 表单标签、普通说明、卡片正文使用 `--text-secondary`。
- 时间、辅助信息、次要元数据使用 `--text-tertiary`。
- 不可操作任务、未配置提示使用 `--text-disabled`。
- 失败、警告、成功分别使用状态变量，不得用接近背景的低透明灰。

- [ ] **Step 4: 运行测试与构建**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Run: `npm --prefix frontend run build`

Expected: PASS。

- [ ] **Step 5: 浏览器核对深浅主题**

依次访问 `/script`、`/novel-fetch`、`/novel-panel`、`/agent`、`/history`、`/issues`、`/tts`；切换两种主题，检查任务状态、空状态、加载提示、失败提示、表格文字和抽屉/弹窗文字。

- [ ] **Step 6: 提交**

```bash
git add frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/novel-fetch.css frontend/src/user/pages/giant-material-test.css frontend/src/user/pages/GiantMaterialStatusCard.css frontend/src/shared/styles/global.css frontend/src/shared/styles/contrast-contract.test.js
git commit -m "fix(ui): improve creative workspace text contrast"
```

### Task 4: 修复水货生产、批量工厂与巨量素材工作台对比度

**Files:**
- Modify: `frontend/src/user/pages/shuihuo-production.css`
- Modify: `frontend/src/user/pages/batch-factory-workbench.css`
- Modify: `frontend/src/user/pages/BatchFactoryWorkbenchPage.css`
- Modify: `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-theme.css`
- Modify: `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-workbench.css`
- Modify: `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-settings.css`
- Modify: `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-scoped.css`
- Modify: `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-detail.css`
- Modify: `frontend/src/user/pages/batch-factory-v11/batch-factory-v11-batch-manager.css`
- Test: `frontend/src/shared/styles/contrast-contract.test.js`

**Interfaces:**
- Consumes: Task 1 的语义颜色变量。
- Produces: 水货生产和批量工厂的密集表格、抽屉、弹窗、状态徽章、任务阶段、媒体卡片和空状态文字在两种主题下可读。

- [ ] **Step 1: 写批量工厂状态色契约测试**

```js
test('batch factory and shuihuo styles use semantic text tokens for status text', () => {
  for (const file of [
    'user/pages/shuihuo-production.css',
    'user/pages/batch-factory-v11/batch-factory-v11-theme.css'
  ]) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.match(css, /var\(--text-(?:primary|secondary|tertiary|disabled|danger|warning|success)\)/);
  }
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Expected: FAIL，因为高密度任务页尚未使用统一状态文字变量。

- [ ] **Step 3: 最小替换密集任务页颜色**

- 所有表头、单元格、任务名、阶段名、抽屉标题和表单标签映射到 `--text-primary` 或 `--text-secondary`。
- 次要统计、时间、任务 ID、媒体元数据映射到 `--text-tertiary`。
- 错误、重试、暂停、完成等状态文字映射到语义状态变量，并确保状态底色不会降低文字可读性。
- 只保留必要的状态色硬编码；禁止用透明度低于 0.5 的文字作为业务信息。

- [ ] **Step 4: 运行测试与构建**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Run: `npm --prefix frontend run build`

Expected: PASS。

- [ ] **Step 5: 浏览器核对深浅主题**

访问 `/shuihuo-production`、`/batch-factory` 及其项目详情、设置抽屉、媒体弹窗和巨量素材弹窗；在深色、浅色主题下检查表格、标签、状态卡、按钮、媒体说明、错误提示和禁用操作。

- [ ] **Step 6: 提交**

```bash
git add frontend/src/user/pages/shuihuo-production.css frontend/src/user/pages/batch-factory-workbench.css frontend/src/user/pages/BatchFactoryWorkbenchPage.css frontend/src/user/pages/batch-factory-v11 frontend/src/shared/styles/contrast-contract.test.js
git commit -m "fix(ui): improve batch factory text contrast"
```

### Task 5: 管理后台与最终主题回归

**Files:**
- Modify: `frontend/src/shared/styles/global.css`
- Modify: `frontend/src/admin/App.jsx`
- Modify: `frontend/src/admin/pages/DashboardPage.jsx`
- Modify: `frontend/src/admin/pages/ErrorLogPage.jsx`
- Modify: `frontend/src/admin/pages/PresetLibraryPage.jsx`
- Modify: `frontend/src/admin/pages/PromptStrategyPage.jsx`
- Modify: `frontend/src/shared/styles/contrast-contract.test.js`
- Modify: `frontend/src/user/App.jsx`

**Interfaces:**
- Consumes: Task 1 至 Task 4 的语义变量。
- Produces: 管理后台使用同一深浅文字色阶；用户前台和后台主题切换都保留可读文本。

- [ ] **Step 1: 写后台和透明文字全局回归测试**

```js
test('shared CSS does not hide known business text selectors', () => {
  const css = fs.readFileSync(path.join(root, 'shared/styles/global.css'), 'utf8');
  assert.doesNotMatch(css, /\.legacy-global-status--idle\s*\{[^}]*color:\s*transparent/s);
  assert.doesNotMatch(css, /\.legacy-page-title\s*\{[^}]*color:\s*transparent/s);
});

test('admin application renders with theme-aware root class', async () => {
  const source = fs.readFileSync(path.join(root, 'admin/App.jsx'), 'utf8');
  assert.match(source, /createAntTheme|data-theme|ConfigProvider/);
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `npm --prefix frontend test -- src/shared/styles/contrast-contract.test.js`

Expected: FAIL，因为 `legacy-global-status--idle` 仍隐藏文字，或管理端未明确使用主题 token。

- [ ] **Step 3: 修复后台和剩余透明业务文字**

- 将 `.legacy-global-status--idle` 改为 `color: var(--text-tertiary)`。
- 保留真正用于装饰的渐变标题时，为不支持裁剪文字的情况添加 `color: var(--text-primary)` 后备值。
- 管理端卡片、表格、表单、空状态、错误日志、提示词正文和按钮使用全局语义变量或同步的 Ant Design token。
- 用户端与管理端根元素均使用一致的 `data-theme` 值，以确保全局变量切换生效。

- [ ] **Step 4: 运行完整前端验证**

Run: `npm --prefix frontend test`

Run: `npm --prefix frontend run build`

Expected: PASS。

- [ ] **Step 5: 最终浏览器回归**

在两种主题下检查用户前台全部导航入口和管理后台全部菜单入口。每页至少确认页面标题、一个正文区域、一个表单/按钮、一个状态或辅助文字均清晰可读。

- [ ] **Step 6: 提交**

```bash
git add frontend/src/shared/styles/global.css frontend/src/admin frontend/src/user/App.jsx frontend/src/shared/styles/contrast-contract.test.js
git commit -m "fix(ui): complete sitewide text contrast audit"
```

## Final Acceptance Checklist

- [ ] 深色和浅色主题下，用户前台和管理后台文字都清晰可读。
- [ ] 普通正文、表单、按钮、链接、表格、标签、提示、禁用状态和状态色均符合设计要求。
- [ ] 真实业务标题、状态和正文没有 `color: transparent`。
- [ ] 公共样式、账号中心、创作页、高密度任务页和管理端都使用统一语义变量。
- [ ] `npm --prefix frontend test` 和 `npm --prefix frontend run build` 均通过。
- [ ] 本地浏览器检查未发现新的深浅主题可读性问题。
