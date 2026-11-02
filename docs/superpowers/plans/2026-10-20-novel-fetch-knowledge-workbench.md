# 小说获取知识库正式工作台 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将知识库预览的紧凑双栏工作台迁入正式小说获取页面，并继续使用真实的知识库数据和主题。

**Architecture:** 保持 `NovelFetchPage.jsx` 作为知识库状态和 API 调用的唯一来源。`KnowledgeLibraryManager` 负责资料选择、位置导航与编辑，`NovelFetchSubpageShell` 保持页面壳，CSS 仅使用现有主题变量。

**Tech Stack:** React 18、Ant Design 5、lucide-react、Node test runner、Vite、Go embed 本机服务。

## Global Constraints

- 正式页面只能使用 React + Ant Design；不得嵌入 `55420` 或其静态预览文件。
- 知识库数据继续通过现有 Go API 读取、保存、复制、删除和优化。
- 不硬编码浅色或深色页面颜色；使用 `--legacy-*` 和 `--text-*` 主题变量。
- 不修改无关的处理、任务、配置、提交页面或已有的脏工作区文件。

---

### Task 1: 资料选择工作台与位置导航

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx:1-30,415-480`
- Modify: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: `items`, `selectedId` 和 `onSelect`，均由 `LegacyNovelFetchPage` 的真实知识库状态提供。
- Produces: 带有“上一条资料”和“下一条资料”按钮、当前 `n / total` 位置提示的 `KnowledgeLibraryManager`。

- [x] **Step 1: 写出失败的布局测试**

```js
test('知识库资料列表提供位置提示和前后切换', () => {
  assert.match(source, /aria-label="上一条资料"/);
  assert.match(source, /aria-label="下一条资料"/);
  assert.match(source, /className="novel-fetch-knowledge-list-position"/);
});
```

- [x] **Step 2: 运行测试确认失败**

Run: `node --test src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL，缺少位置导航的标记。

- [x] **Step 3: 最小实现真实资料导航**

```jsx
const selectedIndex = Math.max(0, filtered.findIndex(item => String(item.id) === String(selectedId)));
const selectRelative = delta => {
  const next = filtered[Math.max(0, Math.min(filtered.length - 1, selectedIndex + delta))];
  if (next) onSelect(String(next.id));
};
```

在圆柱列表下方使用带 Tooltip 的 lucide 上下箭头按钮调用 `selectRelative(-1)` 和 `selectRelative(1)`；禁用第一项的上一步和最后一项的下一步。位置文本显示 `0 / 0` 或 `${selectedIndex + 1} / ${filtered.length}`。

- [x] **Step 4: 运行测试确认通过**

Run: `node --test src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: PASS。

### Task 2: 正式双栏布局和主题样式

**Files:**
- Modify: `frontend/src/user/pages/novel-fetch-subpages.css:109-134,191-197`
- Modify: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: `novel-fetch-knowledge-*` 结构和全局主题变量。
- Produces: 桌面 300px 左栏、紧凑分隔线、圆柱列表导航和窄屏单栏布局。

- [x] **Step 1: 写出失败的样式测试**

```js
test('知识库左栏与圆柱导航遵循紧凑主题工作台', () => {
  assert.match(css, /\.novel-fetch-knowledge-layout\s*\{[^}]*grid-template-columns:\s*300px minmax\(0,1fr\)/s);
  assert.match(css, /\.novel-fetch-knowledge-list-position\s*\{[^}]*text-align:\s*center/s);
  assert.match(css, /\.novel-fetch-knowledge-list-actions\s*\{[^}]*display:\s*grid/s);
  assert.doesNotMatch(css, /\.novel-fetch-knowledge-(?:workbench|layout|library|editor)[^{]*\{[^}]*(?:#[0-9a-f]{3,8}|rgba?\()/is);
});
```

- [x] **Step 2: 运行测试确认失败**

Run: `node --test src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: FAIL，左栏仍为 320px，且没有导航位置样式。

- [x] **Step 3: 最小实现主题跟随样式**

```css
.novel-fetch-knowledge-layout { grid-template-columns:300px minmax(0,1fr); }
.novel-fetch-knowledge-list-navigator { display:grid; grid-template-columns:32px minmax(0,1fr) 32px; gap:8px; }
.novel-fetch-knowledge-list-position { color:var(--text-tertiary); font-size:12px; text-align:center; }
.novel-fetch-knowledge-list-actions { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:8px; }
```

让资料列表视区维持固定高度、编辑区保持底部保存操作，移动端将导航和操作按钮改为可换行单列，不添加主题专用色值。

- [x] **Step 4: 运行测试确认通过**

Run: `node --test src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: PASS。

### Task 3: 构建和真实页面验收

**Files:**
- Verify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Verify: `frontend/src/user/pages/novel-fetch-subpages.css`
- Verify: `scripts/build-go-local.ps1`
- Verify: `scripts/start-local-go-review.ps1`

**Interfaces:**
- Consumes: 前两项中的正式 React 页面、现有 Go 嵌入构建脚本和本机 MySQL 配置。
- Produces: `http://localhost:4001/novel-fetch?tab=knowledge` 的可用正式工作台。

- [x] **Step 1: 运行前端知识库和圆柱列表测试**

Run: `node --test src/user/pages/novel-fetch-subpages.layout.test.js src/user/pages/novel-fetch-cylinder.test.js`

Expected: PASS。

- [x] **Step 2: 构建嵌入式本机服务**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Expected: 前端构建和 Go embed 构建完成且退出码为 0。

- [x] **Step 3: 启动或重启本机审阅服务并检查路由**

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`

Expected: 本机服务监听 `4001`，小说获取页面可访问。

- [x] **Step 4: 在已登录页面检查真实交互**

验证资料库切换、搜索、圆柱滚动、上下一条、编辑保存、爆款开头、处理规则，以及浅色和深色主题。确认页面没有控制台错误，且操作目标仍为真实知识库条目。
