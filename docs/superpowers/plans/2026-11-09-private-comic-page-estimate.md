# Private Comic Page Estimate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在私人定制漫画正文输入区实时显示与批量工厂一致的预计漫画张数。

**Architecture:** 复用 `batchFactoryContentRange.js` 已有的 `comicEstimatedPageRange(characterCount)`，由 `PrivateComicWorkspace.jsx` 根据现有 `sourceLength` 计算并展示。后端真实分镜页数仍由 `comicReadiness.pagesTotal` 决定，预计值只用于生成前提示。

**Tech Stack:** React 18、Ant Design、Node.js 内置测试、Vite、Go embed

## Global Constraints

- 每张目标 90～120 字，最多 35 张。
- 正文为空时不显示预计张数。
- 已生成后的真实页数继续以后端数据为准。
- 不调用 AI，不要求资产图，不改变生成、发布或重试链路。

---

### Task 1: 私人定制漫画预计张数

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Test: `frontend/src/user/pages/PrivateComicWorkspace.source.test.js`

**Interfaces:**
- Consumes: `comicEstimatedPageRange(characterCount: number): { min: number, max: number }`
- Produces: 正文存在时的 `已读取 N 字 · 预计约 X～Y 张` 可见文案

- [ ] **Step 1: 写失败测试**

在 `PrivateComicWorkspace.source.test.js` 添加断言，要求页面导入并调用 `comicEstimatedPageRange(sourceLength)`，且把结果显示在正文计数行。

```js
test('private comic workspace shows a live page estimate before generation', () => {
  assert.match(source, /comicEstimatedPageRange/);
  assert.match(source, /预计约/);
});
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `node --test src/user/pages/PrivateComicWorkspace.source.test.js`

Expected: FAIL，因为页面尚未导入和展示 `comicEstimatedPageRange`。

- [ ] **Step 3: 最小实现**

在 `PrivateComicWorkspace.jsx` 导入共享函数，基于现有 `sourceLength` 得到范围，并把计数行改为：

```jsx
<span>{sourceLength ? `已读取 ${sourceLength} 字 · 预计约 ${pageEstimate.min}～${pageEstimate.max} 张` : '尚未读取'}</span>
```

真实页数继续使用现有 `pagesTotal`，不修改任何生成按钮或请求。

- [ ] **Step 4: 验证测试与构建**

Run: `node --test src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/shuihuo/batchFactoryContentRange.test.js`

Expected: 所有测试 PASS。

Run: `npm run build`

Expected: Vite 构建退出码 0。

- [ ] **Step 5: 提交并更新本地运行时**

```powershell
git add frontend/src/user/pages/PrivateComicWorkspace.jsx frontend/src/user/pages/PrivateComicWorkspace.source.test.js
git commit -m "feat: show private comic page estimate"
scripts/build-go-local.ps1
scripts/start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached
```

验证 `/health` 为 200、`/api/runtime-build-info` 的 SHA 与本地提交一致，并确认运行中的 `/script` 资源包含“预计约”。

