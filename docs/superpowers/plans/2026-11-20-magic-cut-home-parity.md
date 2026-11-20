# 魔剪无登录入口与原版首页 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 从一战晟铭切换到 4040 后不再出现独立登录表单，并按用户提供的 1920×1080 原版截图恢复魔剪首页的主要结构、文案、入口和近期工程行为。

**Architecture:** 保留现有 Go 会话桥和工程 API；React 的未认证分支改成只读“返回一战晟铭”提示。首页拆成纯数据模型与真实 React 页面，页面继续从 `/api/magic-cut/projects` 读取近期工程，“原片顺剪”和“新建自由创作”共用现有创建工程弹窗。

**Tech Stack:** React 18、Ant Design 5、lucide-react、Node `node:test`、Vite、Go embed。

## Global Constraints

- 不新建分支，直接在 `v88` 提交；不 push，不使用 Actions。
- 不运行原版程序，不登录原版账号，不调用付费或外部接口。
- 4040 不提供用户名、密码或通行密钥输入框；会话失败只允许返回一战晟铭。
- 页面以用户提供的 1920×1080 截图和原版 `Home-B-_BdFFe.js` 为基准。
- 每个任务都遵循 RED -> GREEN -> REFACTOR，并更新 `magic-cut-workbench/docs/parity/feature-ledger.md`。
- 不提交既存的 `magic-cut-workbench/api/magic-cut-4040.exe`。

---

### Task 1: 删除魔剪独立登录入口

**Files:**
- Create: `magic-cut-workbench/前台/src/app/sessionGateModel.js`
- Test: `magic-cut-workbench/前台/src/app/sessionGateModel.test.js`
- Modify: `magic-cut-workbench/前台/src/app/App.jsx`
- Modify: `magic-cut-workbench/前台/src/app/app.css`
- Delete: `magic-cut-workbench/前台/src/features/account-login/AccountLogin.jsx`

**Interfaces:**
- Consumes: `loadSession(): Promise<{authenticated, account, errorCode, message}>`
- Produces: `createSessionGateModel(session, locationLike): {title, message, returnLabel, returnHref}`

- [ ] **Step 1: Write the failing test**

```js
test('unauthenticated magic cut only offers returning to yizhanchengming', () => {
  assert.deepEqual(createSessionGateModel(
    { errorCode: 'AUTH_REQUIRED', message: 'ignored' },
    { protocol: 'http:', hostname: '192.168.1.20' },
  ), {
    title: '请从一战晟铭进入魔剪',
    message: '当前魔剪会话已失效，请返回一战晟铭后重新切换到魔剪。',
    returnLabel: '返回一战晟铭',
    returnHref: 'http://192.168.1.20:4001/',
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test -- src/app/sessionGateModel.test.js`

Expected: FAIL because `sessionGateModel.js` does not exist.

- [ ] **Step 3: Write minimal implementation**

Create the pure model, remove the `AccountLogin` import and component branch from `App.jsx`, and render only the message plus one return button. Keep the loading state and authenticated content unchanged.

- [ ] **Step 4: Run focused and full frontend tests**

Run: `pnpm test -- src/app/sessionGateModel.test.js`

Run: `pnpm test`

Expected: focused test and all frontend tests PASS.

- [ ] **Step 5: Commit**

```powershell
git add -- magic-cut-workbench/前台/src/app magic-cut-workbench/前台/src/features/account-login/AccountLogin.jsx
git commit -m "fix(magic-cut): remove standalone login page"
```

### Task 2: 建立原版首页内容模型

**Files:**
- Modify: `magic-cut-workbench/前台/src/features/home/homeModel.js`
- Modify: `magic-cut-workbench/前台/src/features/home/homeModel.test.js`

**Interfaces:**
- Produces: `createHomeModel()`，包含 `primaryTools`、`secondaryTools`、`sidebarGroups`、`stats` 和 `recentProjects` 文案。

- [ ] **Step 1: Replace the existing model test with literal original-page expectations**

```js
assert.deepEqual(model.primaryTools.map(item => item.title), ['原片顺剪', 'AI前贴', '选剧助手'])
assert.deepEqual(model.secondaryTools.map(item => item.title), [
  '批量顺剪', '新剧抢发', '爆剧仿剪', '高光短切', '视频裂变', '封面素材', '漫剧转译',
])
assert.deepEqual(model.stats.map(item => item.label), ['魔剪素材数', '已使用素材数', '有消耗素材数', '消耗'])
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test -- src/features/home/homeModel.test.js`

Expected: FAIL because the current model does not expose the original-page collections.

- [ ] **Step 3: Implement the literal model**

Use the screenshot copy exactly. Mark unfinished tool destinations with `available: false`; set `原片顺剪` and `新建自由创作` to the existing create-project flow.

- [ ] **Step 4: Run the model test**

Run: `pnpm test -- src/features/home/homeModel.test.js`

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add -- magic-cut-workbench/前台/src/features/home/homeModel.js magic-cut-workbench/前台/src/features/home/homeModel.test.js
git commit -m "feat(magic-cut): define original home content"
```

### Task 3: 复刻首页结构和交互

**Files:**
- Modify: `magic-cut-workbench/前台/src/features/home/HomePage.jsx`
- Modify: `magic-cut-workbench/前台/src/app/App.jsx`
- Modify: `magic-cut-workbench/前台/src/app/app.css`
- Create: `magic-cut-workbench/前台/src/features/home/homeInteractionModel.js`
- Test: `magic-cut-workbench/前台/src/features/home/homeInteractionModel.test.js`

**Interfaces:**
- Produces: `resolveHomeToolAction(tool): 'create-project' | 'unavailable' | 'navigate'`
- Preserves: `projectCreated(project)` route `/projects/:id/edit`

- [ ] **Step 1: Write the failing interaction test**

```js
assert.equal(resolveHomeToolAction({ action: 'create-project', available: true }), 'create-project')
assert.equal(resolveHomeToolAction({ action: 'navigate', available: false }), 'unavailable')
assert.equal(resolveHomeToolAction({ action: 'navigate', available: true }), 'navigate')
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm test -- src/features/home/homeInteractionModel.test.js`

Expected: FAIL because the resolver does not exist.

- [ ] **Step 3: Implement the page**

Render a 36px top bar, 210px sidebar, three primary cards, seven secondary cards, overview panel, banner, and recent-project row. Use real buttons for clickable cards, show an Ant Design info message for unavailable cards, and open `CreateProjectDialog` for 原片顺剪/新建自由创作. Existing projects must still open `/projects/:id/edit`.

- [ ] **Step 4: Add fixed screenshot-aligned CSS**

At 1920×1080 use the measured proportions from the supplied screenshot: sidebar 210px, content left 42px, primary grid 3 columns, secondary grid 7 columns, overview/banner 1fr/0.95fr, dark background `#17171a`, card border `#303036`, purple active color `#b667ff`.

- [ ] **Step 5: Run all frontend tests and build**

Run: `pnpm test`

Run: `pnpm build`

Expected: all tests PASS and Vite production build succeeds.

- [ ] **Step 6: Commit**

```powershell
git add -- magic-cut-workbench/前台/src/app magic-cut-workbench/前台/src/features/home
git commit -m "feat(magic-cut): rebuild original home layout"
```

### Task 4: Embed、运行验收与台账登记

**Files:**
- Replace generated output: `magic-cut-workbench/api/web/dist/`
- Modify: `magic-cut-workbench/docs/parity/feature-ledger.md`

**Interfaces:**
- Consumes: Vite `dist/`
- Produces: Go-embedded 4040 frontend and MC-002/MC-003 acceptance record.

- [ ] **Step 1: Sync the production build into Go embed assets**

Remove only the contents of `magic-cut-workbench/api/web/dist`, verify that resolved path stays inside `magic-cut-workbench/api/web`, then copy the new Vite `dist` files into it.

- [ ] **Step 2: Run backend and frontend verification**

Run: `go test ./...` from `magic-cut-workbench/api`.

Run: `pnpm test` and `pnpm build` from `magic-cut-workbench/前台`.

Expected: all commands PASS.

- [ ] **Step 3: Rebuild and restart only the confirmed 4040 runtime**

Build `./cmd/magic-cut-4040`, replace the isolated runtime binary after confirming its exact process path, restart it, then verify `/health` and the root page.

- [ ] **Step 4: Perform visible 1920×1080 acceptance**

Open 4040 through the existing authenticated browser, confirm no login form appears, capture the full homepage, click 原片顺剪, create/open a project, return to the homepage, and confirm the recent project card persists.

- [ ] **Step 5: Update the parity ledger honestly**

Set MC-002 to complete only if no-login entry is proven in the running browser. Keep MC-003 partial unless the screenshot comparison shows the declared regions and interactions. Add commit hashes, tests, runtime result, screenshot path, and remaining gaps.

- [ ] **Step 6: Commit docs and generated embed assets**

```powershell
git add -- magic-cut-workbench/api/web/dist magic-cut-workbench/docs/parity/feature-ledger.md
git commit -m "docs(magic-cut): register home parity acceptance"
```
