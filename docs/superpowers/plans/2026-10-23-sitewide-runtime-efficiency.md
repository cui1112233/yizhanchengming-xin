# 全站运行一致性与资源效率 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让本地网站的导航、页面加载和执行器能力统一，同时减少没有用户操作时的脚本下载与重复状态请求。

**Architecture:** `localExecutors.js` 作为豆包执行器状态与配对接口的唯一前端入口；设置页、剧本生成、Agent 画布和批量工厂都从它读取相同的状态。路由仍保持按需加载，导航悬停才预取目标工作台。批量工厂把对同一运行状态的并发轮询收束为单一、可见时才运行的刷新通道。

**Tech Stack:** React、Ant Design、Vite、Go embed、本地 Go 服务、Node 内置测试。

## Global Constraints

- 只处理 `localhost:4001` 网站及直接依赖；不改本机执行器软件、账号、安装包、历史目录或公网。
- 豆包执行器统一供剧本生成、Agent 画布、批量工厂使用；巨量素材执行器只用于巨量素材导入。
- 执行器离线或未配对时，保留用户操作入口，让后端返回明确状态；前端不提前把正常执行入口隐藏或禁用。
- 保留现有对象 URL 回收、事件解绑、页面隐藏暂停和任务结束补刷行为。
- 工作区存在无关未提交内容：只暂存本计划实际修改的文件；未经用户明确授权，不提交、推送或发布。
- 前端改完必须重新构建并重新生成 Go 内嵌资源，再用 `localhost:4001` 实测；仅 Vite 构建成功不算完成。

---

## File Structure

- `frontend/src/shared/api/localExecutors.js`：唯一封装豆包执行器的列表、配对和在线能力判断。
- `frontend/src/shared/api/localExecutors.test.js`：纯接口数据和能力判断测试。
- `frontend/src/shared/hooks/useLocalExecutorStatus.js`：按需读取状态、去重请求、页面隐藏暂停、前台补刷。
- `frontend/src/shared/hooks/useLocalExecutorStatus.test.js`：状态刷新生命周期测试。
- `frontend/src/user/App.jsx` 与 `frontend/src/shared/components/Link.jsx`：移除全站强制预加载，改为导航项明确预取。
- `frontend/src/user/App.performance.test.js`：验证不再在应用启动时同时下载两个工作台。
- `frontend/src/user/pages/SettingsPage.jsx`：豆包执行器唯一配对和设备状态入口；巨量素材专用区域保持独立。
- `frontend/src/user/pages/ScriptPage.jsx`、`frontend/src/user/pages/AgentPageV2.jsx`、`frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.jsx`：共用豆包状态提示和同一份在线判断。
- `frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialExecutorStatus.jsx`：只保留巨量素材执行器专用状态轮询，并在不可见时暂停。
- `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`：将自动化状态和工作台状态合并到同一单飞刷新入口。

## Task 1: 统一豆包执行器 API 与状态 Hook

**Files:**
- Modify: `frontend/src/shared/api/localExecutors.js`
- Create: `frontend/src/shared/hooks/useLocalExecutorStatus.js`
- Modify: `frontend/src/shared/api/localExecutors.test.js`
- Create: `frontend/src/shared/hooks/useLocalExecutorStatus.test.js`

**Interfaces:**
- Produces `listLocalExecutors(options?): Promise<LocalExecutor[]>`，请求 `/api/shuihuo-production/local-executors`。
- Produces `createLocalExecutorPairing(): Promise<PairingSecret>`，向 `/api/shuihuo-production/local-executors/pairings` 提交 `{ platform: 'doubao' }`。
- Produces `localExecutorStatus(executors): { online: boolean, paired: number, onlineCount: number }`。
- Produces `useLocalExecutorStatus({ enabled, intervalMs })`，返回 `{ executors, status, loading, error, refresh }`；同一 Hook 实例在上一次请求未结束时不能再次请求。

- [ ] **Step 1: 写失败测试，固定状态数据与页面隐藏行为**

```js
test('local executor status reports paired and online counts', () => {
  assert.deepEqual(localExecutorStatus([{ online: true }, { online: false }]), {
    online: true, paired: 2, onlineCount: 1
  });
});

test('status hook does not start a second request while refresh is pending', async () => {
  const pending = deferred();
  const load = () => pending.promise;
  const gate = createLocalExecutorRefreshGate(load);
  const first = gate.refresh();
  const second = gate.refresh();
  assert.equal(first, second);
});
```

- [ ] **Step 2: 运行测试，确认新增接口尚不存在**

Run: `node --test frontend/src/shared/api/localExecutors.test.js frontend/src/shared/hooks/useLocalExecutorStatus.test.js`

Expected: FAIL，缺少 `localExecutorStatus`、`createLocalExecutorRefreshGate` 或 Hook 文件。

- [ ] **Step 3: 实现最小共享 API 和刷新门闩**

```js
export async function listLocalExecutors(options = {}) {
  const result = await apiRequest('/api/shuihuo-production/local-executors', {
    suppressGlobalError: true,
    ...options
  });
  return parseLocalExecutorsResponse(result);
}

export function localExecutorStatus(executors) {
  const items = Array.isArray(executors) ? executors : [];
  const onlineCount = items.filter(item => item?.online === true).length;
  return { online: onlineCount > 0, paired: items.length, onlineCount };
}
```

Hook 中使用 `document.visibilityState === 'visible'` 才启动定时刷新；监听 `visibilitychange`，重新可见时调用一次 `refresh()`；清理时取消计时器与监听器。`refresh()` 必须复用未完成 Promise，避免同页多个触发重叠请求。

- [ ] **Step 4: 运行共享模块测试**

Run: `node --test frontend/src/shared/api/localExecutors.test.js frontend/src/shared/hooks/useLocalExecutorStatus.test.js`

Expected: PASS。

## Task 2: 移除启动时全量预加载，改为导航预取

**Files:**
- Modify: `frontend/src/user/App.jsx`
- Modify: `frontend/src/shared/components/Link.jsx`
- Modify: `frontend/src/shared/layouts/UserLayout.jsx`
- Modify: `frontend/src/user/App.performance.test.js`
- Test: `frontend/src/user/NovelPanelRemoval.source.test.js`

**Interfaces:**
- Produces `preloadRoute(href)`：只接受 `/novel-fetch` 与 `/batch-factory`，返回对应动态导入 Promise；其他地址返回 `undefined`。
- `Link` 新增可选 `onIntent` 回调；鼠标悬停或键盘获得焦点时调用一次，不改变现有点击、修饰键或 `reload` 语义。

- [ ] **Step 1: 改写性能测试为按意图预取**

```js
test('UserApp does not preload workbench pages on application mount', async () => {
  const source = await fs.readFile(appPath, 'utf8');
  assert.doesNotMatch(source, /useEffect\(\(\) => scheduleWorkbenchPreload/);
  assert.match(source, /export function preloadRoute\(href\)/);
});

test('navigation preloads only workbench destinations on user intent', async () => {
  const source = await fs.readFile(new URL('../shared/layouts/UserLayout.jsx', appPath), 'utf8');
  assert.match(source, /onIntent=\{\(\) => preloadRoute\(item\.href\)\}/);
  assert.doesNotMatch(source, /小说面板/);
});
```

- [ ] **Step 2: 运行测试，确认旧的启动预加载断言失败**

Run: `node --test frontend/src/user/App.performance.test.js frontend/src/user/NovelPanelRemoval.source.test.js`

Expected: FAIL，因为应用仍调用 `scheduleWorkbenchPreload()`。

- [ ] **Step 3: 将预取绑定到导航意图**

```jsx
export function preloadRoute(href) {
  if (href === '/novel-fetch') return loadNovelFetchPage();
  if (href === '/batch-factory') return loadShuihuoProductionPage();
  return undefined;
}

<Link href={item.href} onIntent={() => { void preloadRoute(item.href); }}>
  {item.label}
</Link>
```

`Link` 内部用一个 ref 防止同一链接多次触发加载；`onMouseEnter` 和 `onFocus` 只调用一次 `onIntent`。删除 `scheduleWorkbenchPreload()` 和 `UserApp` 中对应的挂载 effect。

- [ ] **Step 4: 运行路由与性能测试**

Run: `node --test frontend/src/user/App.performance.test.js frontend/src/user/NovelPanelRemoval.source.test.js`

Expected: PASS。

## Task 3: 设置页成为唯一配对入口，并接入剧本生成

**Files:**
- Modify: `frontend/src/user/pages/SettingsPage.jsx`
- Modify: `frontend/src/user/pages/ScriptPage.jsx`
- Create: `frontend/src/user/pages/SettingsPage.local-executor.source.test.js`
- Modify: `frontend/src/user/pages/scriptVideoTasks.test.js`

**Interfaces:**
- Consumes `useLocalExecutorStatus`、`createLocalExecutorPairing`、`localExecutorStatus`。
- `SettingsPage` 只在“本机图片与视频执行通道”区域生成豆包配对码；不向批量工厂下发单独配对状态。
- `ScriptPage.generateVideoForShot(card, index, options)` 用 Hook 的 `status.online` 校验 `local-doubao-executor-video`，且沿用现有 `createScriptVideo(videoPayload)` 提交路径。

- [ ] **Step 1: 写失败测试，声明统一入口和剧本页面复用状态**

```js
test('settings is the only user-facing Doubao pairing entry', () => {
  assert.match(settings, /createLocalExecutorPairing/);
  assert.match(settings, /useLocalExecutorStatus/);
});

test('script video checks shared executor status before local submission', () => {
  assert.match(script, /const \{ status: localExecutorStatus/);
  assert.match(script, /localExecutorStatus\.online/);
  assert.doesNotMatch(script, /apiRequest\('\/api\/shuihuo-production\/local-executors'/);
});
```

- [ ] **Step 2: 运行测试，确认剧本页仍直接请求执行器列表**

Run: `node --test frontend/src/user/pages/SettingsPage.local-executor.source.test.js frontend/src/user/pages/scriptVideoTasks.test.js`

Expected: FAIL，因为 `ScriptPage` 仍在 `generateVideoForShot` 中直接调用 `/local-executors`。

- [ ] **Step 3: 改为共享状态**

```jsx
const { executors, status: localExecutorStatus, refresh: refreshLocalExecutors } =
  useLocalExecutorStatus({ enabled: true, intervalMs: 30000 });

if (scriptVideoModelKey === 'local-doubao-executor-video' && !localExecutorStatus.online) {
  await refreshLocalExecutors();
  throw new Error('豆包执行器未在线；请到设置完成一次配对或检查执行器。');
}
```

设置页删除本地重复 `apiRequest` 列表读取，改用 Hook；保留“刷新状态”“生成配对码”“测试图片/视频生成”按钮。巨量素材区域继续使用自己的 `loadGiantMaterialExecutors()`，标题改为“巨量素材专用执行器”。

- [ ] **Step 4: 运行页面测试**

Run: `node --test frontend/src/user/pages/SettingsPage.local-executor.source.test.js frontend/src/user/pages/scriptVideoTasks.test.js`

Expected: PASS。

## Task 4: Agent 与批量工厂复用豆包状态，不再提供分散配对入口

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.jsx`
- Modify: `frontend/src/user/pages/batch-factory-v11/BatchFactoryV11SettingsDrawers.jsx`
- Modify: `frontend/src/user/pages/agent-model-capabilities.test.js`
- Modify: `frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.source.test.js`

**Interfaces:**
- Consumes `useLocalExecutorStatus({ enabled: isCanvasWorkspace, intervalMs: 30000 })`。
- Produces统一提示文本：`豆包执行器未在线；请到设置完成一次配对或检查执行器。`
- 批量工厂只显示“前往设置”链接，不在其设置抽屉生成配对码。

- [ ] **Step 1: 写失败测试，禁止批量工厂内配对**

```js
test('Batch Factory reads shared Doubao status and does not create a pairing code', async () => {
  assert.match(page, /useLocalExecutorStatus/);
  assert.doesNotMatch(drawer, /生成豆包配对码/);
  assert.match(drawer, /前往设置/);
});

test('Agent local image and video capabilities preserve the local provider', () => {
  assert.ok(compatibleModels(models, { kind: 'video', provider: 'doubao_local_executor' }).length > 0);
});
```

- [ ] **Step 2: 运行测试，确认旧抽屉仍能生成配对码**

Run: `node --test frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.source.test.js`

Expected: FAIL，因为 `BatchFactoryV11SettingsDrawers.jsx` 仍包含“生成豆包配对码”。

- [ ] **Step 3: 接入统一状态与后端失败回传**

```jsx
const { status: localExecutorStatus, refresh: refreshLocalExecutors } =
  useLocalExecutorStatus({ enabled: true, intervalMs: 30000 });

if (provider === 'doubao_local_executor' && !localExecutorStatus.online) {
  message.warning('豆包执行器未在线；任务仍会提交并显示执行结果。');
}
```

删除 `ProductionSettingsDrawer` 的 `onCreateLocalExecutorPairing`、`pairingBusy` 和 `pairingSecret` 参数与页面块。保留模型选择；不在线时不返回 `false`，仍提交到后端并以任务错误状态呈现。Agent 画布只显示统一状态提示，不改变现有 `doubao_local_executor` 图片/视频模型选择及节点任务提交。

- [ ] **Step 4: 运行聚焦测试**

Run: `node --test frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.source.test.js`

Expected: PASS。

## Task 5: 收束轮询并保留前后台语义

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialExecutorStatus.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialExecutorStatus.test.js`
- Create: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.polling.test.js`

**Interfaces:**
- `BatchFactoryGiantMaterialExecutorStatus` 对可见组件只使用一个 `refreshHealth()` 调度器；它依次读取本机健康、后台在线设备、设备偏好，而不是三个 `setInterval`。
- `BatchFactoryNovelList` 提供 `refreshRuntimeStatus({ reason })`；同一时刻只有一个请求，下一次触发复用进行中的 Promise。

- [ ] **Step 1: 写失败测试，限制同页计时器数量并保护在途请求**

```js
test('giant material status uses one visible-page refresh scheduler', () => {
  assert.match(source, /function refreshHealth/);
  assert.doesNotMatch(source, /setInterval\(poll, 3000\)[\s\S]*setInterval\(pollBackend, 6000/);
});

test('batch factory runtime polling uses one in-flight request', async () => {
  const gate = createRuntimeRefreshGate(load);
  assert.equal(gate.refresh({ reason: 'automation' }), gate.refresh({ reason: 'media' }));
});
```

- [ ] **Step 2: 运行测试，确认现有多个定时器会触发失败**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialExecutorStatus.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.polling.test.js`

Expected: FAIL，因为组件存在三个独立 `setInterval`，书籍页存在两条独立状态轮询。

- [ ] **Step 3: 实现单一刷新调度**

```js
function createRuntimeRefreshGate(load) {
  let pending = null;
  return {
    refresh(options) {
      if (!pending) pending = Promise.resolve(load(options)).finally(() => { pending = null; });
      return pending;
    }
  };
}
```

巨量素材组件在可见时每 10 秒调度一次 `refreshHealth()`，并行读取三类独立数据后一次写入状态；隐藏时停止，恢复可见时立即补刷。书籍页由一个 effect 决定刷新间隔：自动化运行或生产任务运行时 15 秒，相关弹窗打开但没有执行任务时 30 秒，否则停止；恢复可见时通过同一 gate 补刷，并保留自动化从运行到结束的最后一次 `onBatchChanged()`。

- [ ] **Step 4: 运行轮询测试**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialExecutorStatus.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.polling.test.js`

Expected: PASS。

## Task 6: 构建、重启并按真实页面验收

**Files:**
- Modify: 仅由构建生成的 `frontend/dist/**` 和 Go embed 输入；不要反向编辑生成文件。
- Verify: `backend/internal/web/static.go`
- Verify: `scripts/build-go-local.ps1`
- Verify: `scripts/start-local-go.ps1`

- [ ] **Step 1: 运行所有聚焦测试**

Run: `node --test frontend/src/user/App.performance.test.js frontend/src/user/NovelPanelRemoval.source.test.js frontend/src/shared/api/localExecutors.test.js frontend/src/shared/hooks/useLocalExecutorStatus.test.js frontend/src/user/pages/SettingsPage.local-executor.source.test.js frontend/src/user/pages/scriptVideoTasks.test.js frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.source.test.js frontend/src/user/pages/shuihuo/BatchFactoryGiantMaterialExecutorStatus.test.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.polling.test.js`

Expected: PASS。

- [ ] **Step 2: 运行前端生产构建与 Go 本地构建**

Run: `npm run build`（在 `frontend`）随后运行 `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`。

Expected: 两个命令退出码为 0，且 Go 构建包含新生成的前端资源。

- [ ] **Step 3: 重启当前本地 Go 服务**

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go.ps1`。

Expected: `http://localhost:4001` 返回 200；运行页面加载新资源，而非旧内嵌 bundle。

- [ ] **Step 4: 浏览器验收**

验证以下可见结果：

1. 设置、批量工厂、剧本生成、Agent 画布侧栏均没有“小说面板”。
2. 初始打开设置页时不自动下载小说获取和批量工厂两个工作台；鼠标移到各自导航项后才开始下载。
3. 设置页可显示豆包设备状态并生成配对码；剧本生成、Agent、批量工厂显示相同的豆包在线/离线含义。
4. 巨量素材状态仍只出现在巨量素材导入流程，且页面隐藏后不再刷新。
5. 批量工厂任务、剧本视频生成和 Agent 节点提交仍显示排队、执行、成功、失败等现有状态；控制台没有新增错误。

## Plan Self-Review

- 覆盖：导航与运行版本由 Task 2、6 覆盖；页面加载由 Task 2 覆盖；统一执行器由 Task 1、3、4 覆盖；轮询由 Task 5 覆盖；实际 Go embed 验收由 Task 6 覆盖。
- 边界：未列入执行器软件、历史模块、安装包、公网和无关工作区内容。
- 完整性：已检查本文，没有悬而未决的步骤或未定义的后续任务。
- 一致性：所有页面共用 `localExecutorStatus.online` 与 `useLocalExecutorStatus`；后端任务接口和本机执行器协议不变。
