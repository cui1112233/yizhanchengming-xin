# V88 Liblib-Like Director Desk Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 V88 当前基础 3D 导演台升级为具有真实场景对象、角色和多机位属性、画幅控制、独立底部编辑框及亮暗主题支持的三栏导演工作区。

**Architecture:** Agent Canvas 继续拥有项目、节点、素材和持久化；导演台通过现有 `DirectorDeskHostAdapter` 读写一个版本化的 `directorProject`。界面拆成外壳、对象树、素材区、Three.js 视口、属性检查器和底部工具，所有 V88 API 仍只经过 Host Adapter，Three.js 组件不直接请求后端。

**Tech Stack:** React 18、JavaScript、Three.js、`@react-three/fiber`、`@react-three/drei`、Lucide、Node `node:test`、Vite、Go embed。

## Global Constraints

- 只在当前 `v88` 分支工作，不创建新分支，不重置、不清理、不覆盖工作区里的其他修改。
- 每次提交前只暂存本任务明确列出的文件，并检查 `git diff --cached`，避免提交无关脏文件。
- React 源码是唯一实现来源；禁止手工修改 `frontend/dist` 或 `backend/web/dist` 来实现功能。
- 发布和本地运行只走 Go embed 路径；不启动或恢复旧 Node 应用服务。
- 继续使用现有 Agent Canvas 项目、节点配置、项目素材上传和截图回写流程；不新增数据库表和后端接口。
- 所有可见按钮必须连接真实能力；姿势、关节和动画时间轴在对应数据合同完成前不显示。
- 场景描述框只保存 `scene.prompt`，不调用模型生成，不消耗积分。
- 验收期间可以做本地 Three.js 截图，但不触发图片、视频、音频或文本模型生成。
- 亮色和暗色主题都必须保持文本、图标、输入框、选中态和错误信息可见。
- 三栏工作区不跟随 Agent Canvas 缩放；全局侧栏展开或收起时不得遮挡导演台。

---

## File Structure

| File | Responsibility |
|---|---|
| `frontend/src/user/pages/agent-director-desk.js` | `3d-v2` 数据规范化、`3d-v1`/2.5D 迁移、角色/摄像机/画幅更新。 |
| `frontend/src/user/pages/agent-director-desk.test.js` | 迁移、对象更新、非法数据回退和兼容层测试。 |
| `frontend/src/user/components/director-desk/director-scene-state.js` | 视角选择、画幅尺寸、Three.js 普通对象与数组数据互转。 |
| `frontend/src/user/components/director-desk/director-scene-state.test.js` | 摄像机、画幅和角色投影的纯函数测试。 |
| `frontend/src/user/components/director-desk/director-save-queue.js` | 串行保存、脏状态和关闭前刷新。 |
| `frontend/src/user/components/director-desk/director-save-queue.test.js` | 保存合并、失败重试和 `flush()` 测试。 |
| `frontend/src/user/components/director-desk/DirectorDeskLoadBoundary.jsx` | 动态模块加载失败的可见错误与整页重试。 |
| `frontend/src/user/components/director-desk/DirectorDeskShell.jsx` | 三栏全屏外壳、顶部视角切换、两侧面板收起。 |
| `frontend/src/user/components/director-desk/DirectorToolRail.jsx` | 左侧图标工具入口，只列出真实可用工具。 |
| `frontend/src/user/components/director-desk/DirectorSceneTree.jsx` | 场景、角色和摄像机对象列表。 |
| `frontend/src/user/components/director-desk/DirectorAssetPanel.jsx` | 项目图片、设为场景和加入角色。 |
| `frontend/src/user/components/director-desk/director-inspector-state.js` | 检查器数值更新的纯函数。 |
| `frontend/src/user/components/director-desk/director-inspector-state.test.js` | 角色、摄像机和场景属性更新测试。 |
| `frontend/src/user/components/director-desk/DirectorInspector.jsx` | 按选中对象显示角色、摄像机或场景属性。 |
| `frontend/src/user/components/director-desk/DirectorScene.jsx` | 3D 视口、活动摄像机、对象选择、移动支点和截图。 |
| `frontend/src/user/components/director-desk/DirectorBottomDock.jsx` | 移动、重置、截图、素材入口和场景描述保存。 |
| `frontend/src/user/components/director-desk/DirectorDeskApp.jsx` | 组合组件、加载素材、选择状态、保存队列和错误状态。 |
| `frontend/src/user/components/director-desk/DirectorDeskApp.test.js` | 组件边界、真实工具和失败状态的源码合同测试。 |
| `frontend/src/user/components/AgentCanvas.jsx` | 懒加载边界、Host Adapter 和截图回写接线。 |
| `frontend/src/user/components/AgentCanvas.director-desk.source.test.js` | 懒加载失败可见、导演台接线和截图原子性测试。 |
| `frontend/src/shared/styles/agent-workspace.css` | 导演台语义变量、三栏布局、响应式面板、视口和底部控件。 |

---

### Task 1: Make Lazy-Load Failure Visible and Recoverable

**Files:**
- Create: `frontend/src/user/components/director-desk/DirectorDeskLoadBoundary.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.director-desk.source.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: Existing lazy `DirectorDeskApp` import and `directorDeskAdapter` condition.
- Produces: `DirectorDeskLoadBoundary({ children, onRetry })`; retry performs `window.location.reload()` so a stale hashed chunk can be replaced by the current Go-embedded manifest.

- [ ] **Step 1: Add failing source-contract assertions.**

```js
test('导演台动态模块失败时显示错误和重新加载入口', () => {
  const canvas = fs.readFileSync(path.join(dirname, 'AgentCanvas.jsx'), 'utf8');
  const boundary = fs.readFileSync(path.join(dirname, 'director-desk/DirectorDeskLoadBoundary.jsx'), 'utf8');
  assert.match(canvas, /DirectorDeskLoadBoundary/);
  assert.match(boundary, /导演台资源加载失败/);
  assert.match(boundary, /重新加载页面/);
  assert.match(boundary, /window\.location\.reload/);
});
```

- [ ] **Step 2: Run the focused test and confirm the intended failure.**

Run from `frontend`:

```powershell
node --test src/user/components/AgentCanvas.director-desk.source.test.js
```

Expected: FAIL because `DirectorDeskLoadBoundary.jsx` does not exist.

- [ ] **Step 3: Implement the boundary.**

```jsx
import { Component } from 'react';

export class DirectorDeskLoadBoundary extends Component {
  state = { error: null };

  static getDerivedStateFromError(error) {
    return { error };
  }

  render() {
    if (!this.state.error) return this.props.children;
    return <section className="agent-director-desk-load-error" role="alert">
      <strong>导演台资源加载失败</strong>
      <span>本地运行包可能仍在使用旧资源，请重新加载当前页面。</span>
      <button type="button" onClick={() => (this.props.onRetry || (() => window.location.reload()))()}>重新加载页面</button>
    </section>;
  }
}
```

Wrap the existing `Suspense` block:

```jsx
<DirectorDeskLoadBoundary onRetry={() => window.location.reload()}>
  <Suspense fallback={<div className="agent-director-desk-loading">正在加载 3D导演台…</div>}>
    <DirectorDeskApp adapter={directorDeskAdapter} onCapture={asset => commitDirectorCapture(directorDeskNodeID, asset)} />
  </Suspense>
</DirectorDeskLoadBoundary>
```

- [ ] **Step 4: Add a theme-safe error surface and rerun the test.**

Use only `--director-*` variables defined in Task 4, with temporary fallbacks to existing `--legacy-*` variables so this task remains independently usable.

Run the Step 2 command. Expected: PASS.

- [ ] **Step 5: Commit only this task.**

```powershell
git add -- frontend/src/user/components/director-desk/DirectorDeskLoadBoundary.jsx frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/AgentCanvas.director-desk.source.test.js frontend/src/shared/styles/agent-workspace.css
git diff --cached --check
git commit -m "fix(agent-canvas): surface director desk load failures"
```

---

### Task 2: Upgrade Director Project State to `3d-v2`

**Files:**
- Modify: `frontend/src/user/pages/agent-director-desk.js`
- Modify: `frontend/src/user/pages/agent-director-desk.test.js`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.test.js`

**Interfaces:**
- Consumes: Existing `3d-v1` state and legacy `directorDesk` state.
- Produces: `normalizeDirectorProject`, `addDirectorActor`, `updateDirectorActor`, `addDirectorCamera`, `updateDirectorCamera`, `updateDirectorViewCamera`, `setDirectorActiveCamera`, `removeDirectorObject`, `setDirectorAspectRatio`, and `directorSceneObjects`.

- [ ] **Step 1: Write failing migration and update tests.**

```js
test('3d-v1 camera and scalar actor scale migrate deterministically to 3d-v2', () => {
  const result = normalizeDirectorProject({
    rendererVersion: '3d-v1',
    actors: [{ id: 'actor-1', assetId: 'image-1', position: [1, 0, 2], rotation: [0, 0.2, 0], scale: 1.4 }],
    camera: { position: [0, 3, 8], target: [0, 1, 0], fov: 42 }
  });
  assert.equal(result.rendererVersion, '3d-v2');
  assert.deepEqual(result.actors[0].scale, [1.4, 1.4, 1.4]);
  assert.equal(result.cameras[0].id, 'camera-main');
  assert.equal(result.activeCameraId, 'camera-main');
  assert.deepEqual(result.directorCamera.position, [0, 3, 8]);
});

test('camera, aspect ratio and scene object helpers preserve valid state', () => {
  const base = normalizeDirectorProject();
  const withCamera = addDirectorCamera(base, { id: 'camera-b', name: '侧面机位' });
  const active = setDirectorActiveCamera(withCamera, 'camera-b');
  const framed = setDirectorAspectRatio(active, '9:16');
  assert.equal(framed.activeCameraId, 'camera-b');
  assert.equal(framed.aspectRatio, '9:16');
  assert.deepEqual(directorSceneObjects(framed).map(item => item.kind), ['scene', 'camera', 'camera']);
});
```

- [ ] **Step 2: Run the state tests and confirm failure.**

```powershell
node --test src/user/pages/agent-director-desk.test.js src/user/components/director-desk/DirectorDeskHostAdapter.test.js
```

Expected: FAIL because `3d-v2` helpers do not exist and the adapter still expects `3d-v1`.

- [ ] **Step 3: Implement exact defaults and normalization.**

```js
export const DIRECTOR_ASPECT_RATIOS = Object.freeze(['adaptive', '21:9', '16:9', '4:3', '1:1', '3:4', '9:16']);
const DEFAULT_DIRECTOR_CAMERA = Object.freeze({ position: [0, 3, 8], target: [0, 0, 0], fov: 42 });

function normalizeScale(value) {
  const source = Array.isArray(value) ? value : [value, value, value];
  return source.map(item => clamp(item, 0.1, 10, 1));
}

function normalizeCameraObject(camera = {}, index = 0) {
  return {
    id: text(camera.id || (index === 0 ? 'camera-main' : `camera-${index + 1}`), 200),
    name: text(camera.name || (index === 0 ? '主机位' : `机位${index + 1}`), 200),
    position: numberList(camera.position, DEFAULT_DIRECTOR_CAMERA.position),
    rotation: numberList(camera.rotation, [0, 0, 0]),
    target: numberList(camera.target, DEFAULT_DIRECTOR_CAMERA.target),
    fov: clamp(camera.fov, 20, 90, 42)
  };
}
```

`normalizeDirectorProject()` must always return `rendererVersion: '3d-v2'`, a valid `camera-main`, `activeCameraId`, `aspectRatio`, `directorCamera`, array actor scale, and existing captures.

- [ ] **Step 4: Implement immutable state helpers.**

```js
export function updateDirectorCamera(project, cameraID, patch) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({
    ...current,
    cameras: current.cameras.map(camera => camera.id === cameraID ? { ...camera, ...patch, id: camera.id } : camera)
  });
}

export function updateDirectorActor(project, actorID, patch) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({
    ...current,
    actors: current.actors.map(actor => actor.id === actorID ? { ...actor, ...patch, id: actor.id, assetId: actor.assetId } : actor)
  });
}

export function updateDirectorViewCamera(project, patch) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({ ...current, directorCamera: { ...current.directorCamera, ...patch } });
}

export function setDirectorAspectRatio(project, ratio) {
  const current = normalizeDirectorProject(project);
  return { ...current, aspectRatio: DIRECTOR_ASPECT_RATIOS.includes(ratio) ? ratio : 'adaptive' };
}
```

`addDirectorActor()` keeps the asset ID as the persistent reference and assigns a collision-safe actor ID. `removeDirectorObject()` must reject removal of `scene`, `camera-main`, and the last camera. When removing another active camera, choose the first remaining camera.

- [ ] **Step 5: Keep legacy wrappers and update the adapter expectation.**

`directorDeskConfig`, `moveDirectorDeskActor`, and `directorDeskSceneStyle` continue reading normalized state until the compact canvas card no longer needs legacy projections. `DirectorDeskHostAdapter.writeState()` now returns and saves `3d-v2`.

Run the Step 2 command. Expected: PASS.

- [ ] **Step 6: Commit only state-contract files.**

```powershell
git add -- frontend/src/user/pages/agent-director-desk.js frontend/src/user/pages/agent-director-desk.test.js frontend/src/user/components/director-desk/DirectorDeskHostAdapter.test.js
git diff --cached --check
git commit -m "feat(agent-canvas): add director project v2 state"
```

---

### Task 3: Add Camera, Aspect-Ratio, and Serialized-Save Helpers

**Files:**
- Modify: `frontend/src/user/components/director-desk/director-scene-state.js`
- Modify: `frontend/src/user/components/director-desk/director-scene-state.test.js`
- Create: `frontend/src/user/components/director-desk/director-save-queue.js`
- Create: `frontend/src/user/components/director-desk/director-save-queue.test.js`

**Interfaces:**
- Consumes: Task 2 `3d-v2` project.
- Produces: `actorPlaneProps`, `cameraForView`, `cameraFromControls`, `aspectRatioValue`, `captureViewportSize`, and `createDirectorSaveQueue(write, { onStatus })`.

- [ ] **Step 1: Write failing pure-function tests.**

```js
test('画幅映射为稳定比例和截图尺寸', () => {
  assert.equal(aspectRatioValue('21:9'), 21 / 9);
  assert.equal(aspectRatioValue('adaptive'), null);
  assert.deepEqual(captureViewportSize('9:16', 1280), { width: 720, height: 1280 });
  assert.deepEqual(captureViewportSize('16:9', 1280), { width: 1280, height: 720 });
});

test('机位视角使用活动摄像机，导演视角使用自由摄像机', () => {
  const project = normalizeDirectorProject();
  assert.equal(cameraForView(project, 'camera').id, 'camera-main');
  assert.equal(cameraForView(project, 'director').kind, 'director');
});
```

```js
test('保存队列合并未开始的状态并在 flush 后完成', async () => {
  const writes = [];
  const queue = createDirectorSaveQueue(value => writes.push(value));
  queue.schedule({ revision: 1 });
  queue.schedule({ revision: 2 });
  await queue.flush();
  assert.deepEqual(writes, [{ revision: 2 }]);
  assert.equal(queue.getStatus(), 'saved');
});

test('保存失败保留 dirty 状态并允许后续重试', async () => {
  let fail = true;
  const queue = createDirectorSaveQueue(async () => { if (fail) throw new Error('保存失败'); });
  queue.schedule({ revision: 1 });
  await assert.rejects(queue.flush(), /保存失败/);
  assert.equal(queue.getStatus(), 'error');
  fail = false;
  await queue.flush();
  assert.equal(queue.getStatus(), 'saved');
});
```

- [ ] **Step 2: Run both test files and confirm failure.**

```powershell
node --test src/user/components/director-desk/director-scene-state.test.js src/user/components/director-desk/director-save-queue.test.js
```

Expected: FAIL for missing helpers and missing save-queue module.

- [ ] **Step 3: Implement aspect and camera helpers.**

```js
const RATIOS = Object.freeze({ '21:9': 21 / 9, '16:9': 16 / 9, '4:3': 4 / 3, '1:1': 1, '3:4': 3 / 4, '9:16': 9 / 16 });

export function aspectRatioValue(value) {
  return RATIOS[value] || null;
}

export function captureViewportSize(value, longEdge = 1280) {
  const ratio = aspectRatioValue(value);
  if (!ratio) return { width: longEdge, height: Math.round(longEdge * 9 / 16) };
  return ratio >= 1
    ? { width: longEdge, height: Math.round(longEdge / ratio) }
    : { width: Math.round(longEdge * ratio), height: longEdge };
}
```

`actorPlaneProps()` must return array scale unchanged. `cameraFromControls()` must return plain arrays. `cameraForView()` must fall back to `camera-main` when `activeCameraId` is stale.

- [ ] **Step 4: Implement a serial, retryable save queue.**

```js
export function createDirectorSaveQueue(write, { onStatus = () => {} } = {}) {
  let pending = null;
  let inFlight = null;
  let status = 'saved';
  let lastError = null;

  async function drain() {
    if (inFlight) return inFlight;
    inFlight = (async () => {
      while (pending) {
        const value = pending;
        pending = null;
        status = 'saving';
        onStatus(status);
        try { await write(value); status = pending ? 'dirty' : 'saved'; lastError = null; onStatus(status); }
        catch (error) { pending = value; status = 'error'; lastError = error; onStatus(status, error); throw error; }
      }
    })().finally(() => { inFlight = null; });
    return inFlight;
  }

  return {
    schedule(value) {
      pending = value;
      status = 'dirty';
      onStatus(status);
      queueMicrotask(() => { void drain().catch(() => {}); });
    },
    flush: drain,
    getStatus: () => status,
    getError: () => lastError
  };
}
```

Scheduling two states in the same synchronous turn keeps only the newest pending value. Every schedule starts a microtask drain, so inspector and viewport edits reach `adapter.writeState()` without waiting for close. `flush()` waits for the active write and every newer pending value; after a failure it retries the preserved pending value.

- [ ] **Step 5: Run tests and commit.**

Run the Step 2 command. Expected: PASS.

```powershell
git add -- frontend/src/user/components/director-desk/director-scene-state.js frontend/src/user/components/director-desk/director-scene-state.test.js frontend/src/user/components/director-desk/director-save-queue.js frontend/src/user/components/director-desk/director-save-queue.test.js
git diff --cached --check
git commit -m "feat(agent-canvas): add director camera and save helpers"
```

---

### Task 4: Build the Theme-Safe Three-Column Shell

**Files:**
- Create: `frontend/src/user/components/director-desk/DirectorDeskShell.jsx`
- Create: `frontend/src/user/components/director-desk/DirectorToolRail.jsx`
- Create: `frontend/src/user/components/director-desk/DirectorSceneTree.jsx`
- Create: `frontend/src/user/components/director-desk/DirectorAssetPanel.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Produces:
  - `DirectorDeskShell({ left, viewport, right, top, bottom, leftCollapsed, rightCollapsed, onToggleLeft, onToggleRight, onClose })`
  - `DirectorToolRail({ activeTool, onToolChange })`
  - `DirectorSceneTree({ project, selectedObject, onSelect, onLocate, onRemove })`
  - `DirectorAssetPanel({ images, onSetScene, onAddActor })`

- [ ] **Step 1: Add failing source-contract tests.**

```js
test('导演台外壳包含真实工具、三栏插槽和面板收起操作', () => {
  const shell = read('DirectorDeskShell.jsx');
  const rail = read('DirectorToolRail.jsx');
  assert.match(shell, /agent-director-shell__left/);
  assert.match(shell, /agent-director-shell__viewport/);
  assert.match(shell, /agent-director-shell__right/);
  for (const label of ['场景对象', '项目图片', '添加角色', '添加机位', '画幅比例']) assert.match(rail, new RegExp(label));
  assert.doesNotMatch(rail, /姿势|动画时间轴/);
});

test('场景树只来自实际项目对象', () => {
  const tree = read('DirectorSceneTree.jsx');
  assert.match(tree, /directorSceneObjects/);
  assert.match(tree, /onSelect/);
  assert.match(tree, /onLocate/);
});
```

- [ ] **Step 2: Run the app test and confirm failure.**

```powershell
node --test src/user/components/director-desk/DirectorDeskApp.test.js
```

Expected: FAIL because the shell components do not exist.

- [ ] **Step 3: Implement the shell and tool IDs.**

```js
export const DIRECTOR_TOOLS = Object.freeze([
  { id: 'scene', label: '场景对象', icon: Layers3 },
  { id: 'assets', label: '项目图片', icon: Images },
  { id: 'actors', label: '添加角色', icon: UserRoundPlus },
  { id: 'cameras', label: '添加机位', icon: Video },
  { id: 'ratio', label: '画幅比例', icon: Ratio }
]);
```

Use icon-only rail buttons with `aria-label`, `title`, and `aria-pressed`. The left content area is 228px and the icon rail is 44px. The right inspector is 288px.

- [ ] **Step 4: Implement scene tree and asset panel with real callbacks.**

```jsx
{directorSceneObjects(project).map(item => <button
  key={`${item.kind}:${item.id}`}
  type="button"
  aria-pressed={selectedObject?.kind === item.kind && selectedObject?.id === item.id}
  onClick={() => onSelect({ kind: item.kind, id: item.id })}
>{item.name}</button>)}
```

The asset panel must call only `onSetScene(asset)` and `onAddActor(asset)`. It must not import API functions.

- [ ] **Step 5: Define semantic theme variables and responsive tracks.**

```css
.agent-director-desk-fullscreen {
  --director-bg: var(--legacy-bg);
  --director-panel: var(--legacy-card);
  --director-panel-strong: var(--legacy-input);
  --director-border: var(--legacy-border);
  --director-text: var(--legacy-text);
  --director-muted: var(--legacy-muted);
  --director-accent: var(--legacy-accent);
  color: var(--director-text);
  background: var(--director-bg);
}

.agent-director-shell {
  display: grid;
  grid-template-columns: 272px minmax(0, 1fr) 288px;
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;
}

@media (max-width: 1180px) {
  .agent-director-shell { grid-template-columns: auto minmax(0, 1fr) auto; }
  .agent-director-shell__left.is-collapsed,
  .agent-director-shell__right.is-collapsed { width: 44px; }
}
```

Add explicit light-theme overrides using the project's existing theme selector; do not use a white hard-coded popover on a dark surface.

- [ ] **Step 6: Run tests and commit.**

Run the Step 2 command. Expected: PASS.

```powershell
git add -- frontend/src/user/components/director-desk/DirectorDeskShell.jsx frontend/src/user/components/director-desk/DirectorToolRail.jsx frontend/src/user/components/director-desk/DirectorSceneTree.jsx frontend/src/user/components/director-desk/DirectorAssetPanel.jsx frontend/src/user/components/director-desk/DirectorDeskApp.test.js frontend/src/shared/styles/agent-workspace.css
git diff --cached --check
git commit -m "feat(agent-canvas): add director desk workspace shell"
```

---

### Task 5: Implement the Real Object Inspector

**Files:**
- Create: `frontend/src/user/components/director-desk/director-inspector-state.js`
- Create: `frontend/src/user/components/director-desk/director-inspector-state.test.js`
- Create: `frontend/src/user/components/director-desk/DirectorInspector.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `{ kind: 'scene'|'actor'|'camera', id: string }` selection and Task 2 project helpers.
- Produces: `applyDirectorInspectorChange(project, selection, field, value)` and `DirectorInspector({ project, selection, onProjectChange, onSetActiveCamera, onChangeScene })`.

- [ ] **Step 1: Write failing inspector-state tests.**

```js
test('角色检查器更新单轴位置、旋转和缩放', () => {
  const base = normalizeDirectorProject({ actors: [{ id: 'actor-1', assetId: 'image-1' }] });
  const moved = applyDirectorInspectorChange(base, { kind: 'actor', id: 'actor-1' }, 'position.0', '2.5');
  const scaled = applyDirectorInspectorChange(moved, { kind: 'actor', id: 'actor-1' }, 'uniformScale', '1.8');
  assert.equal(scaled.actors[0].position[0], 2.5);
  assert.deepEqual(scaled.actors[0].scale, [1.8, 1.8, 1.8]);
});

test('摄像机检查器限制 FOV 并拒绝无效数字', () => {
  const base = normalizeDirectorProject();
  const wide = applyDirectorInspectorChange(base, { kind: 'camera', id: 'camera-main' }, 'fov', '120');
  const invalid = applyDirectorInspectorChange(wide, { kind: 'camera', id: 'camera-main' }, 'position.1', 'abc');
  assert.equal(wide.cameras[0].fov, 90);
  assert.deepEqual(invalid, wide);
});
```

- [ ] **Step 2: Run the test and confirm failure.**

```powershell
node --test src/user/components/director-desk/director-inspector-state.test.js
```

Expected: FAIL because the module does not exist.

- [ ] **Step 3: Implement field-path updates through Task 2 helpers.**

```js
export function applyDirectorInspectorChange(project, selection, field, rawValue) {
  const value = Number(rawValue);
  if (!Number.isFinite(value)) return normalizeDirectorProject(project);
  if (selection.kind === 'actor') {
    const patch = actorPatchForField(project, selection.id, field, value);
    return patch ? updateDirectorActor(project, selection.id, patch) : normalizeDirectorProject(project);
  }
  if (selection.kind === 'camera') {
    const patch = cameraPatchForField(project, selection.id, field, value);
    return patch ? updateDirectorCamera(project, selection.id, patch) : normalizeDirectorProject(project);
  }
  return normalizeDirectorProject(project);
}
```

`actorPatchForField()` and `cameraPatchForField()` are private immutable path-to-patch helpers in `director-inspector-state.js`. The allowed actor paths are `position.0..2`, `rotation.0..2`, `scale.0..2`, and `uniformScale`. The allowed camera paths are `position.0..2`, `rotation.0..2`, `target.0..2`, and `fov`. Reject every other path by returning `null`, and skip the update when the patch is `null`.

- [ ] **Step 4: Implement accessible controls.**

`DirectorInspector` must render:

- Actor: name, position XYZ, rotation XYZ, scale XYZ, uniform scale.
- Camera: name, position XYZ, rotation XYZ, target XYZ, FOV, “设为活动机位”.
- Scene: scene name, current ratio, “更换场景图片”.

Each input uses a unique label such as `角色位置 X` or `机位 FOV`. `onBlur` commits a valid number; `Enter` blurs; invalid text restores the last valid value and shows `请输入有效数值` beside that field.

- [ ] **Step 5: Run inspector and app tests, then commit.**

```powershell
node --test src/user/components/director-desk/director-inspector-state.test.js src/user/components/director-desk/DirectorDeskApp.test.js
```

Expected: PASS.

```powershell
git add -- frontend/src/user/components/director-desk/director-inspector-state.js frontend/src/user/components/director-desk/director-inspector-state.test.js frontend/src/user/components/director-desk/DirectorInspector.jsx frontend/src/user/components/director-desk/DirectorDeskApp.test.js frontend/src/shared/styles/agent-workspace.css
git diff --cached --check
git commit -m "feat(agent-canvas): add director object inspector"
```

---

### Task 6: Add Multi-Camera View, Selection, Transform Control, and Aspect Frame

**Files:**
- Modify: `frontend/src/user/components/director-desk/DirectorScene.jsx`
- Modify: `frontend/src/user/components/director-desk/director-scene-state.js`
- Modify: `frontend/src/user/components/director-desk/director-scene-state.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: Task 2 project, Task 3 camera/aspect helpers, selection, and `viewMode: 'director'|'camera'`.
- Produces: `DirectorScene({ project, images, selection, viewMode, transformMode, onSelectionChange, onProjectChange, onReadyChange })` plus `capture()`, `resetDirectorView()`, and `focusObject(selection)` ref methods.

- [ ] **Step 1: Add failing camera and capture-ratio tests.**

```js
test('stale active camera falls back to camera-main', () => {
  const project = normalizeDirectorProject({ activeCameraId: 'missing' });
  assert.equal(cameraForView(project, 'camera').id, 'camera-main');
});

test('actor plane preserves non-uniform v2 scale', () => {
  assert.deepEqual(actorPlaneProps({ position: [0, 0, 0], rotation: [0, 0, 0], scale: [1, 2, 1] }).scale, [1, 2, 1]);
});
```

- [ ] **Step 2: Run scene-state tests and confirm failure.**

```powershell
node --test src/user/components/director-desk/director-scene-state.test.js
```

Expected: FAIL until the v2 helpers are used consistently.

- [ ] **Step 3: Make the viewport frame own the real ratio.**

```jsx
<div className={`agent-director-viewport-frame ${project.aspectRatio === 'adaptive' ? 'is-adaptive' : ''}`}
  style={aspectRatioValue(project.aspectRatio) ? { aspectRatio: String(aspectRatioValue(project.aspectRatio)) } : undefined}>
  <Canvas>{sceneContents}</Canvas>
</div>
```

The Canvas must fill this frame. `capture()` reads the WebGL canvas inside the frame, so the PNG ratio equals the selected ratio without post-capture stretching.

- [ ] **Step 4: Switch between director and active-camera state.**

```jsx
const cameraState = cameraForView(project, viewMode);
<CameraFrame cameraState={cameraState} />
{viewMode === 'director' ? <OrbitControls
  ref={controlsRef}
  enabled={!transformDragging}
  target={project.directorCamera.target}
  onEnd={event => onProjectChange(updateDirectorViewCamera(project, cameraFromControls(event.target.object, event.target.target)))}
/> : null}
```

Camera view is disabled in the app when there is no normalized camera. Camera markers are visible and selectable only in director view.

- [ ] **Step 5: Attach a real move gizmo to the selected object.**

Use Drei `TransformControls` with `mode="translate"`. Wrap only the selected actor or camera marker. While dragging, disable OrbitControls. On pointer release, convert the Three.js object transform to plain arrays and call the appropriate Task 2 update helper.

```jsx
<TransformControls mode="translate" onMouseDown={() => setTransformDragging(true)} onMouseUp={finishTransform}>
  <ActorPlaneMesh actor={actor} asset={asset} onClick={() => onSelectionChange({ kind: 'actor', id: actor.id })} />
</TransformControls>
```

Textured actor planes keep direct selection. Image controls and right-panel inputs remain clickable because they are outside the React Flow gesture layer.

`focusObject(selection)` updates only the free director camera: actor focus targets the actor position, camera focus targets the selected camera marker, and scene focus targets `[0, 0, 0]`. It must not modify the active production camera or Agent Canvas viewport.

- [ ] **Step 6: Expose readiness and nonblank capture.**

Call `onReadyChange(true)` after the WebGL renderer and first frame exist. If `capture()` runs before readiness, reject with `3D 构图尚未准备完成`. Preserve the current PNG `File` return contract.

- [ ] **Step 7: Run tests and commit.**

Run the Step 2 command. Expected: PASS.

```powershell
git add -- frontend/src/user/components/director-desk/DirectorScene.jsx frontend/src/user/components/director-desk/director-scene-state.js frontend/src/user/components/director-desk/director-scene-state.test.js frontend/src/shared/styles/agent-workspace.css
git diff --cached --check
git commit -m "feat(agent-canvas): add director camera views and transform control"
```

---

### Task 7: Add the Bottom Dock, Prompt Save, and Capture Actions

**Files:**
- Create: `frontend/src/user/components/director-desk/DirectorBottomDock.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Produces: `DirectorBottomDock({ prompt, saveStatus, canCapture, transformMode, onTransformModeChange, onPromptChange, onSavePrompt, onOpenAssets, onUploadReference, onResetView, onCapture })`.

- [ ] **Step 1: Add failing source assertions for real actions only.**

```js
test('底部导演工具只暴露当前真实能力', () => {
  const dock = read('DirectorBottomDock.jsx');
  for (const label of ['移动', '重置视角', '截图到画布', '项目图片', '上传参考图', '保存场景描述']) assert.match(dock, new RegExp(label));
  assert.doesNotMatch(dock, /生成场景|动画时间轴|姿势/);
  assert.match(dock, /onSavePrompt/);
  assert.match(dock, /onCapture/);
});
```

- [ ] **Step 2: Run the app test and confirm failure.**

```powershell
node --test src/user/components/director-desk/DirectorDeskApp.test.js
```

Expected: FAIL because `DirectorBottomDock.jsx` does not exist.

- [ ] **Step 3: Implement the dock with stable dimensions.**

```jsx
<section className="agent-director-bottom-dock" aria-label="导演台工具">
  <div className="agent-director-bottom-dock__tools">
    <button type="button" aria-pressed={transformMode === 'translate'} onClick={() => onTransformModeChange('translate')} title="移动"><Move3D size={17} /></button>
    <button type="button" onClick={onResetView} title="重置视角"><RotateCcw size={17} /></button>
    <button type="button" onClick={onCapture} disabled={!canCapture} title="截图到画布"><Camera size={17} /></button>
  </div>
  <div className="agent-director-composer">
    <button type="button" onClick={onOpenAssets} title="项目图片"><ImagePlus size={17} /></button>
    <label title="上传参考图"><Upload size={17} /><input type="file" accept="image/*" onChange={event => onUploadReference(event.target.files?.[0])} /></label>
    <textarea value={prompt} onChange={event => onPromptChange(event.target.value)} aria-label="场景描述" />
    <button type="button" onClick={onSavePrompt} disabled={!prompt.trim()} title="保存场景描述"><Send size={17} /></button>
  </div>
  <span role="status">{saveStatus === 'saving' ? '正在保存' : saveStatus === 'error' ? '尚未同步' : '已保存'}</span>
</section>
```

The dock uses viewport-fixed CSS size, not React Flow scale. The composer width is `min(620px, calc(100% - 32px))`; button sizes are fixed at 34px.

The hidden file input must reset its value after each attempt so the same image can be selected again. `onUploadReference(file)` uses the Host Adapter's existing Agent Canvas image-upload path, refreshes project images, and never calls a generation API.

- [ ] **Step 4: Add light/dark theme styles and text-overflow rules.**

The textarea must remain readable in both themes, grow from one to four lines without moving the Three.js frame, and scroll after four lines. The status line must not overlap the send button.

- [ ] **Step 5: Run tests and commit.**

Run the Step 2 command. Expected: PASS.

```powershell
git add -- frontend/src/user/components/director-desk/DirectorBottomDock.jsx frontend/src/user/components/director-desk/DirectorDeskApp.test.js frontend/src/shared/styles/agent-workspace.css
git diff --cached --check
git commit -m "feat(agent-canvas): add director desk bottom dock"
```

---

### Task 8: Integrate the Workspace in `DirectorDeskApp`

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.test.js`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.js`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: Tasks 2-7.
- Produces: Complete staged-parity workspace using only current V88 capabilities; Host Adapter adds `uploadProjectImage(file)` by delegating to the existing Agent Canvas `uploadAgentAsset(taskId, file)` path.

- [ ] **Step 1: Add failing integration-contract assertions.**

```js
test('导演台组合真实对象树、视口、检查器和底部工具', () => {
  const source = read('DirectorDeskApp.jsx');
  for (const component of ['DirectorDeskShell', 'DirectorToolRail', 'DirectorSceneTree', 'DirectorAssetPanel', 'DirectorScene', 'DirectorInspector', 'DirectorBottomDock']) {
    assert.match(source, new RegExp(component));
  }
  assert.match(source, /createDirectorSaveQueue/);
  assert.match(source, /viewMode/);
  assert.match(source, /selectedObject/);
});

test('关闭导演台前刷新最后一次保存', () => {
  const source = read('DirectorDeskApp.jsx');
  assert.match(source, /await saveQueue\.flush\(\)/);
  assert.match(source, /adapter\.close\(\)/);
});
```

- [ ] **Step 2: Run app and adapter tests and confirm failure.**

```powershell
node --test src/user/components/director-desk/DirectorDeskApp.test.js src/user/components/director-desk/DirectorDeskHostAdapter.test.js
```

Expected: FAIL until the new components and save queue are wired.

- [ ] **Step 3: Compose state and derived panels.**

`DirectorDeskApp` owns these UI states and stable refs:

```js
const [project, setProject] = useState(() => normalizeDirectorProject(adapter.readState()));
const [images, setImages] = useState([]);
const [activeTool, setActiveTool] = useState('scene');
const [selectedObject, setSelectedObject] = useState({ kind: 'scene', id: 'scene' });
const [viewMode, setViewMode] = useState('director');
const [transformMode, setTransformMode] = useState('translate');
const [promptDraft, setPromptDraft] = useState(project.scene.prompt);
const [saveStatus, setSaveStatus] = useState('saved');
const [error, setError] = useState('');
const [sceneReady, setSceneReady] = useState(false);
const [busy, setBusy] = useState(false);
const sceneRef = useRef(null);
const projectRef = useRef(project);
const saveQueueRef = useRef(null);
```

Create the queue once with `adapter.writeState` and its `onStatus` callback. `applyProject(next)` normalizes, updates both `projectRef.current` and React state, then schedules the save queue. It must not call canvas APIs directly. The status callback updates `dirty/saving/error/saved` and preserves the latest save error on screen.

- [ ] **Step 4: Wire actual tool behavior.**

- `scene`: render `DirectorSceneTree`.
- `assets` and `actors`: render `DirectorAssetPanel`; `actors` focuses its “加入角色” actions.
- `cameras`: render camera presets defined locally as current view, front medium, front close-up, side follow, overhead, low angle, over-shoulder and Dutch angle. Choosing one calls `addDirectorCamera()` with real coordinates.
- `ratio`: render buttons from `DIRECTOR_ASPECT_RATIOS`; choosing one calls `setDirectorAspectRatio()`.
- Scene-tree selection updates `selectedObject`; locate calls `sceneRef.current.focusObject(selection)`; removal calls `removeDirectorObject()` and falls back to selecting the scene when the selected object was removed.
- “上传参考图” calls `adapter.uploadProjectImage(file)`, refreshes `adapter.listProjectImages()`, and selects the asset tool. `DirectorDeskHostAdapter` validates the returned asset as an image just as it validates captures.

No tool may call a generation API.

- [ ] **Step 5: Wire save, close, screenshot and error states.**

```js
const close = async () => {
  try { await saveQueue.flush(); adapter.close(); }
  catch (cause) { setError(cause?.message || '导演台保存失败'); setSaveStatus('error'); }
};

const capture = async () => {
  if (!sceneReady || busy) return;
  setBusy(true);
  try {
    await saveQueue.flush();
    const file = await sceneRef.current.capture();
    const asset = await adapter.storeCapture(file);
    const next = directorProjectCaptureMetadata(projectRef.current, asset);
    await adapter.writeState(next);
    projectRef.current = next;
    setProject(next);
    await onCapture?.(asset);
  } catch (cause) { setError(cause?.message || '3D 构图保存失败'); }
  finally { setBusy(false); }
};
```

Do not silently discard a known save error. Close remains blocked and the visible error remains until `flush()` succeeds; the user can retry close after connectivity or storage recovers. Unmount must not start a new write or suppress an already reported failure.

- [ ] **Step 6: Verify old functionality is still present.**

Run:

```powershell
node --test src/user/components/director-desk/*.test.js src/user/pages/agent-director-desk.test.js src/user/components/AgentCanvas.director-desk.source.test.js
```

Expected: all Director Desk tests pass, including screenshot persistence-before-canvas insertion.

- [ ] **Step 7: Commit the integrated workspace.**

```powershell
git add -- frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/director-desk/DirectorDeskApp.jsx frontend/src/user/components/director-desk/DirectorDeskApp.test.js frontend/src/user/components/director-desk/DirectorDeskHostAdapter.js frontend/src/user/components/director-desk/DirectorDeskHostAdapter.test.js frontend/src/shared/styles/agent-workspace.css
git diff --cached --check
git commit -m "feat(agent-canvas): integrate director desk parity workspace"
```

---

### Task 9: Build, Embed, Restart, and Perform Browser Acceptance

**Files:**
- Modify only when a failure is proven to come from Tasks 1-8.
- Generated by build, never hand-edited: `frontend/dist/**`, `backend/web/dist/**`, `backend/qiantie-local.exe`.

**Interfaces:**
- Consumes: Complete Director Desk implementation.
- Produces: Verified Go-embedded runtime and visible browser evidence.

- [ ] **Step 1: Run the full focused Director Desk suite.**

From `frontend`:

```powershell
node --test src/user/pages/agent-director-desk.test.js src/user/components/director-desk/*.test.js src/user/components/AgentCanvas.director-desk.source.test.js src/user/components/AgentDirectorDeskNode.source.test.js
```

Expected: all focused tests pass.

- [ ] **Step 2: Run nearby Agent Canvas interaction tests.**

```powershell
node --test src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-viewport.test.js src/user/components/AgentCanvasInteraction.source.test.js
```

Expected: all current tests pass. Record unrelated historical failures separately and do not weaken current interaction behavior to satisfy stale assertions.

- [ ] **Step 3: Build React from source.**

```powershell
npm.cmd --prefix frontend run build
```

Expected: Vite completes successfully and emits a current `DirectorDeskApp-*.js` chunk referenced by the current `AgentPageV2-*.js` chunk.

- [ ] **Step 4: Build the Go-embedded runtime.**

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
```

Expected: `backend/web/dist` receives the new frontend and `backend/qiantie-local.exe` builds successfully.

- [ ] **Step 5: Restart only the Go runtime.**

Use the repository-approved local launcher. Stop only the current `backend/qiantie-local.exe` process bound to port `4001`, then run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/start-local-go.ps1
```

Do not start a Node application server. Keep the process running for browser acceptance.

- [ ] **Step 6: Verify runtime identity and dynamic assets.**

```powershell
(Invoke-WebRequest -UseBasicParsing http://localhost:4001/health).StatusCode
(Invoke-WebRequest -UseBasicParsing http://localhost:4001/api/runtime-build-info).Content
```

Expected:

- `/health` returns `200`;
- `runtime` is `go`;
- `git_sha` and `runtime_manifest_sha` equal the implementation HEAD;
- the current Director Desk chunk returns `200` and no old chunk name is requested after refresh.

- [ ] **Step 7: Browser-check desktop dark and light themes.**

At `http://localhost:4001/agent/canvas?taskId=agent-bf-dbba32e978f13f1909e3b928&view=workflow`:

1. Open an existing Director Desk node.
2. Confirm the 3D canvas is nonblank by screenshot and canvas-pixel sampling.
3. Confirm left panel, viewport and right inspector do not overlap.
4. Switch dark/light theme; verify every label, icon, input, selected row and error surface remains visible.
5. Collapse and expand both side panels.
6. Confirm the bottom dock remains viewport-sized when Agent Canvas zoom changes.

- [ ] **Step 8: Browser-check actual scene operations without generation.**

1. Select a project image as the scene.
2. Add one project image as an actor.
3. Select the actor from both the viewport and scene tree.
4. Move it with the transform control and change position, rotation and scale in the inspector.
5. Add a second camera, set it active, change FOV, and switch Director/Camera view twice.
6. Test `16:9`, `1:1`, and `9:16`; confirm the viewport and screenshot use the same ratio.
7. Save a scene description, close, reopen, and refresh; confirm state persists.
8. Capture once to Canvas; confirm one image node and one Director-to-image edge appear and remain after refresh.
9. Do not click any provider-backed generation action.

- [ ] **Step 9: Browser-check a narrow viewport.**

Use a viewport no wider than 900px. Confirm side panels collapse, text does not overlap, icon buttons stay fixed-size, and the 3D canvas still renders nonblank.

- [ ] **Step 10: Commit only acceptance-driven fixes, if any.**

Before committing, inspect each touched file and stage only Director Desk fixes. Use a message describing the actual defect, for example:

```powershell
git commit -m "fix(agent-canvas): correct director desk narrow layout"
```

If no defect is found, do not create an empty commit.

## Coverage Check

- Three-column shell and responsive collapse: Task 4 and Task 9.
- Theme visibility: Task 4, Task 7, and Task 9.
- Scene object tree and project images: Task 4 and Task 8.
- `3d-v1`/legacy migration to `3d-v2`: Task 2.
- Actor position, rotation and scale: Task 5 and Task 6.
- Multi-camera, active camera, FOV and view switch: Task 2, Task 5, Task 6, and Task 8.
- Aspect ratios and matching screenshots: Task 3, Task 6, and Task 9.
- Independent bottom editor and real prompt save: Task 7 and Task 8.
- Visible dynamic-import recovery: Task 1 and Task 9.
- Serialized saves and close-before-flush protection: Task 3 and Task 8.
- Existing screenshot-to-image-node flow: Task 8 and Task 9.
- No fake pose/timeline/generation controls: global constraints, Task 4, Task 7, and Task 8.
- React build, Go embed, runtime identity, desktop/mobile pixel checks: Task 9.
