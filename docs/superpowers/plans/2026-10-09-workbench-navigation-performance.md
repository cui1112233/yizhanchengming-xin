# 工作台首次进入性能优化实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让小说获取、小说面板、批量工厂首次点击时更快显示页面外壳，并在首页空闲时预加载对应页面代码。

**Architecture:** 保留现有懒加载路由和业务接口，只给三个目标页面增加可复用的模块预加载入口，并在用户应用首次渲染后用浏览器空闲时间触发。iframe 和批量列表继续异步加载，页面先显示现有外壳与加载状态，不改变业务协议和视觉样式。

**Tech Stack:** React 18、React.lazy、Suspense、Vite、Node test runner。

## Global Constraints

- 当前唯一维护分支是 `v88`，不要创建新分支，不要推送 GitHub。
- 不读取、输出或提交 `.env.v88-review`。
- 不提交 `.local-review/`、`frontend/dist/downloads/`、`docs/superpowers/`。
- 不修改 CSS、主题文件、Go 后端、Node 后端接口或公网配置。
- 不执行生产数据库、Redis、TOS 或公网服务操作。
- 必须先写失败测试，再写实现代码。
- 最后运行前端相关测试和 `npm --prefix frontend run build`。

---

### Task 1: 建立页面预加载函数

**Files:**
- Modify: `frontend/src/user/App.jsx`
- Create: `frontend/src/user/App.performance.test.js`

**Interfaces:**
- Produces `preloadWorkbenchPages()`，用于安排三个目标页面模块的动态导入。
- 预加载失败必须被内部捕获，不能产生未处理 Promise 拒绝。

- [ ] **Step 1: Write the failing test**

在 `frontend/src/user/App.performance.test.js` 中测试：调用预加载函数后，三个页面加载器都会被安排执行；如果其中一个加载器失败，函数仍然正常完成。

测试通过导出的预加载函数接收可注入的 loaders，避免真的加载完整页面：

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import { preloadWorkbenchPages } from './App.jsx';

test('preloadWorkbenchPages schedules the three slow workbench modules', async () => {
  const called = [];
  await preloadWorkbenchPages({
    novelFetch: async () => called.push('novel-fetch'),
    novelPanel: async () => called.push('novel-panel'),
    batchFactory: async () => called.push('batch-factory')
  });
  assert.deepEqual(called.sort(), ['batch-factory', 'novel-fetch', 'novel-panel']);
});

test('preloadWorkbenchPages ignores a failed module preload', async () => {
  await assert.doesNotReject(() => preloadWorkbenchPages({
    novelFetch: async () => { throw new Error('chunk unavailable'); },
    novelPanel: async () => {},
    batchFactory: async () => {}
  }));
});
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
node --test frontend/src/user/App.performance.test.js
```

Expected: FAIL because `preloadWorkbenchPages` is not exported yet.

- [ ] **Step 3: Write minimal implementation**

在 `App.jsx` 中把三个动态导入保存为命名加载器，并增加：

```js
export async function preloadWorkbenchPages(loaders = {
  novelFetch: () => import('./pages/NovelFetchPage'),
  novelPanel: () => import('./pages/NovelPanelPage'),
  batchFactory: () => import('./pages/BatchFactoryWorkbenchPage')
}) {
  await Promise.allSettled(Object.values(loaders).map(load => load()));
}
```

在 `UserApp` 的首次渲染 effect 中用 `requestIdleCallback`（没有该 API 时使用短 `setTimeout`）调用预加载函数。预加载只执行动态导入，不创建组件、不请求业务接口。

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
node --test frontend/src/user/App.performance.test.js
```

Expected: PASS。

- [ ] **Step 5: Run existing frontend tests**

Run:

```powershell
npm --prefix frontend test
```

Expected: 现有前端测试通过；若出现与本次无关的既有失败，记录完整测试名和错误，不修改测试绕过。

---

### Task 2: 验证三个页面的首屏行为不被阻塞

**Files:**
- Test: `frontend/src/user/App.performance.test.js`
- Read-only verification: `frontend/src/user/pages/NovelFetchPage.jsx`, `frontend/src/user/pages/NovelPanelPage.jsx`, `frontend/src/user/pages/BatchFactoryWorkbenchPage.jsx`

**Interfaces:**
- 保持三个页面原有导出、路由和接口调用方式。
- 不新增同步网络请求，不改变 iframe 地址、消息协议或批量接口。

- [ ] **Step 1: Write the failing source contract tests**

在测试文件增加源码契约检查：目标页面仍在 `App.jsx` 中由同一批动态导入加载；小说获取和小说面板仍保留 iframe 外壳；批量工厂仍然通过 `listBatchSummaries()` 异步读取列表。

```js
import fs from 'node:fs/promises';

test('workbench pages keep async loading boundaries', async () => {
  const app = await fs.readFile(new URL('./App.jsx', import.meta.url), 'utf8');
  const novelFetch = await fs.readFile(new URL('./pages/NovelFetchPage.jsx', import.meta.url), 'utf8');
  const novelPanel = await fs.readFile(new URL('./pages/NovelPanelPage.jsx', import.meta.url), 'utf8');
  const batchFactory = await fs.readFile(new URL('./pages/BatchFactoryWorkbenchPage.jsx', import.meta.url), 'utf8');
  assert.match(app, /import\('\.\/pages\/NovelFetchPage'\)/);
  assert.match(app, /import\('\.\/pages\/NovelPanelPage'\)/);
  assert.match(app, /import\('\.\/pages\/BatchFactoryWorkbenchPage'\)/);
  assert.match(novelFetch, /iframe[^>]+novel-fetch-original-workbench/);
  assert.match(novelPanel, /workbenchSrc/);
  assert.match(batchFactory, /listBatchSummaries\(\)/);
});
```

- [ ] **Step 2: Run test to verify it fails if the contract is absent**

Run:

```powershell
node --test frontend/src/user/App.performance.test.js
```

Expected: PASS for the existing boundaries; if the new preload refactor accidentally removes one, the test must fail and identify the missing boundary.

- [ ] **Step 3: Make only necessary implementation adjustments**

不调整三个页面的 CSS、主题或业务接口。如果检查发现页面外壳被条件渲染挡住，只做最小 JSX 调整，让外壳和加载状态先渲染，不能把接口请求改成同步等待。

- [ ] **Step 4: Run focused tests**

Run:

```powershell
node --test frontend/src/user/App.performance.test.js
```

Expected: PASS。

---

### Task 3: 完整验证并检查改动范围

**Files:**
- Verify only: `frontend/src/user/App.jsx`, `frontend/src/user/App.performance.test.js`

- [ ] **Step 1: Run all frontend tests**

Run:

```powershell
npm --prefix frontend test
```

Expected: 本次新增性能测试和既有相关测试通过；既有无关失败要单独记录。

- [ ] **Step 2: Run production build**

Run:

```powershell
npm --prefix frontend run build
```

Expected: 构建成功；不得因为测试而修改 CSS、主题或后端文件。

- [ ] **Step 3: Check the diff scope**

Run:

```powershell
git diff --check -- frontend/src/user/App.jsx frontend/src/user/App.performance.test.js
git status --short
```

确认不包含 `.env.v88-review`、`.local-review/`、`frontend/dist/downloads/`，也不把构建产物加入提交范围。

- [ ] **Step 4: Report results**

报告改动文件、预加载测试结果、前端测试结果、构建结果，以及仍可能受 iframe 或后端响应速度影响的等待部分。未明确要求时不创建 Git commit。
