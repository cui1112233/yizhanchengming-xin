# Agent 3D Director Desk Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current 2.5D director-desk modal with a persisted, full-screen Three.js previsualization workspace whose captures become normal image nodes on the existing Agent Canvas.

**Architecture:** The canvas remains the sole project and persistence owner. A small V88 host adapter bridges an adapted, lazily loaded open-source director-desk UI to the existing director node configuration, project-asset API, and canvas mutation functions. The workspace never creates its own project, database tables, asset store, or task queue.

**Tech Stack:** React 18, JavaScript, `three`, `@react-three/fiber`, `@react-three/drei`, Ant Design, `@xyflow/react`, Go embed, MySQL, TOS-backed project assets.

## Global Constraints

- Maintain the canonical `v88` source; do not use generated frontend output as source.
- Preserve the dirty checkout. Do not reset, clean, overwrite unrelated changes, create branches, commit, push, or deploy externally without explicit user authorization.
- Keep the current Go-embedded React release path. Rebuild with `npm --prefix frontend run build`, then `scripts/build-go-local.ps1`, then use `scripts/start-local-go-review.ps1` for local browser acceptance.
- Do not import the reference project's Tauri shell, SQLite storage, sidecar, CLI, or third-party model binaries.
- Keep Workflow and Storyboard over the same Agent Canvas nodes, edges, assets, and configuration.
- Do not start a paid or provider-backed generation task during verification. A director capture is a local render uploaded as a normal project image asset.
- First release uses current-project images as scene/actor planes inside a true Three.js scene. GLB/FBX upload, shared model library, lighting presets, and multi-camera presets are out of scope.

---

## File Structure

| File | Responsibility |
|---|---|
| `frontend/src/user/pages/agent-director-desk.js` | Versioned state normalization, old 2.5D migration, scene/actor/camera updates, capture metadata. |
| `frontend/src/user/pages/agent-director-desk.test.js` | Pure-state regression tests for migration and persistence-ready data. |
| `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.js` | Narrow V88 bridge: read/write node state, list project images, upload capture, and close notification. |
| `frontend/src/user/components/director-desk/DirectorDeskApp.jsx` | Lazily loaded full-screen director-desk entry, keyboard ownership, close action, and composition capture. |
| `frontend/src/user/components/director-desk/DirectorScene.jsx` | Three.js scene: background plane, image actor planes, camera framing, and selection/drag controls. |
| `frontend/src/user/components/director-desk/DirectorDeskApp.test.js` | UI contracts for view modes, adapter use, error handling, and close behavior. |
| `frontend/src/user/components/director-desk/DirectorScene.test.js` | Scene transform and capture contracts without a provider call. |
| `frontend/src/user/components/AgentDirectorDeskNode.jsx` | Compact director-node summary and full-screen launcher; removes the old 2.5D workspace body. |
| `frontend/src/user/components/AgentCanvas.jsx` | Owns host-adapter callbacks, closes/flushed state, uploads captures, creates image nodes, and creates the director-to-image edge. |
| `frontend/src/shared/styles/agent-workspace.css` | Scoped full-screen director-desk layout, scene viewport, panels, and canvas-event isolation. |

---

### Task 1: Define the Versioned Director Project Contract

**Files:**
- Modify: `frontend/src/user/pages/agent-director-desk.js`
- Modify: `frontend/src/user/pages/agent-director-desk.test.js`

**Consumes:** Current `config.directorDesk` objects containing `scenePrompt`, `sceneAsset`, `actors`, and `camera`.

**Produces:** `normalizeDirectorProject`, `migrateLegacyDirectorDesk`, `updateDirectorActorTransform`, and `directorProjectCaptureMetadata` for Tasks 2-4 UI and canvas code.

- [ ] **Step 1: Write the failing migration and update tests.**

```js
test('legacy 2.5D director data migrates to a versioned 3D project', () => {
  const project = migrateLegacyDirectorDesk({
    sceneAsset: { id: 'scene-1', url: '/scene.png', name: '天桥' },
    actors: [{ id: 'actor-1', assetId: 'actor-image', x: 20, y: 70, scale: 1 }],
    camera: { zoom: 1.2, x: 4, y: -3 }
  });
  assert.equal(project.rendererVersion, '3d-v1');
  assert.equal(project.scene.assetId, 'scene-1');
  assert.equal(project.actors[0].assetId, 'actor-image');
  assert.deepEqual(project.camera.target, [0, 0, 0]);
});

test('actor transforms and captures keep stable asset IDs', () => {
  const project = normalizeDirectorProject({ rendererVersion: '3d-v1' });
  const moved = updateDirectorActorTransform(project, 'actor-1', {
    position: [2, 0, -1], rotation: [0, 1.2, 0], scale: 1.3
  });
  const next = directorProjectCaptureMetadata(moved, {
    id: 'capture-1', url: '/capture.png', name: '构图 1', createdAt: '2026-10-24T00:00:00Z'
  });
  assert.equal(next.captures[0].assetId, 'capture-1');
});
```

- [ ] **Step 2: Run the new tests before implementation.**

Run: `node --test src/user/pages/agent-director-desk.test.js`

Expected: FAIL because the versioned contract helpers do not exist.

- [ ] **Step 3: Implement the smallest pure contract.**

```js
export function normalizeDirectorProject(input = {}) {
  return {
    rendererVersion: '3d-v1',
    scene: normalizeAsset(input.scene),
    actors: Array.isArray(input.actors) ? input.actors.map(normalizeActor).filter(Boolean) : [],
    camera: normalizeCamera(input.camera),
    captures: Array.isArray(input.captures) ? input.captures.map(normalizeCapture).filter(Boolean) : []
  };
}

export function migrateLegacyDirectorDesk(legacy = {}) {
  return normalizeDirectorProject({
    scene: legacy.sceneAsset ? { assetId: legacy.sceneAsset.id, url: legacy.sceneAsset.url, name: legacy.sceneAsset.name } : null,
    actors: (legacy.actors || []).map(actor => ({
      id: actor.id, assetId: actor.assetId,
      position: [(Number(actor.x) - 50) / 12, 0, (Number(actor.y) - 65) / 12],
      rotation: [0, 0, Number(actor.rotation || 0) * Math.PI / 180], scale: Number(actor.scale || 1)
    })
  });
}
```

- [ ] **Step 4: Preserve existing callers during transition.**

Keep `directorDeskConfig`, `moveDirectorDeskActor`, and `directorDeskSceneStyle` as compatibility wrappers until Task 4 removes the old modal. Each wrapper must normalize through the new project contract rather than maintain a second authoritative shape.

- [ ] **Step 5: Run focused state tests.**

Run: `node --test src/user/pages/agent-director-desk.test.js`

Expected: PASS with both old compatibility tests and the two new 3D-project tests.

---

### Task 2: Add the V88 Director Desk Host Adapter

**Files:**
- Create: `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.js`
- Create: `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.test.js`
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`

**Consumes:** The Task 1 `normalizeDirectorProject` contract, `handleNodeConfigChange`, `listAgentAssets`, and `uploadAgentAsset`.

**Produces:** `createDirectorDeskHostAdapter({ nodeID, getNode, saveProject, listImages, uploadCapture, onClose })` with `readState`, `writeState`, `listProjectImages`, `storeCapture`, and `close` methods.

- [ ] **Step 1: Write failing adapter tests with fake canvas services.**

```js
test('host adapter writes the versioned project to the selected director node', async () => {
  const writes = [];
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => ({ id: 'director-1', data: { type: 'director', config: {} } }),
    saveProject: (id, patch) => writes.push([id, patch]),
    listImages: async () => [],
    uploadCapture: async () => ({ id: 'asset-1', url: '/capture.png', name: 'capture.png' }),
    onClose: () => undefined
  });
  adapter.writeState({ rendererVersion: '3d-v1', actors: [] });
  assert.deepEqual(writes, [['director-1', { directorProject: { rendererVersion: '3d-v1', scene: null, actors: [], camera: { position: [0, 2, 6], target: [0, 0, 0], fov: 45 }, captures: [] } }]]);
});

test('host adapter rejects captures when upload fails', async () => {
  const adapter = createAdapterWith({ uploadCapture: async () => { throw new Error('上传失败'); } });
  await assert.rejects(() => adapter.storeCapture(new File(['x'], 'capture.png', { type: 'image/png' })), /上传失败/);
});
```

- [ ] **Step 2: Run the adapter tests before implementation.**

Run: `node --test src/user/components/director-desk/DirectorDeskHostAdapter.test.js`

Expected: FAIL because the adapter module does not exist.

- [ ] **Step 3: Add a project-image list API adapter only if the existing asset response lacks a stable image shape.**

```js
export function directorDeskProjectImages(taskId) {
  return listAgentAssets(taskId, 'image').then(result =>
    (result.assets || []).map(asset => ({ id: asset.id, url: asset.url, name: asset.name || asset.filename || '图片' }))
  );
}
```

Do not add a new backend route if `/api/agent/tasks/:taskId/assets?kind=image` already returns `id`, `url`, and a display name.

- [ ] **Step 4: Implement the adapter with no direct API calls from 3D components.**

```js
export function createDirectorDeskHostAdapter({ nodeID, getNode, saveProject, listImages, uploadCapture, onClose }) {
  return {
    readState: () => normalizeDirectorProject(getNode()?.data?.config?.directorProject || migrateLegacyDirectorDesk(getNode()?.data?.config?.directorDesk)),
    writeState: state => saveProject(nodeID, { directorProject: normalizeDirectorProject(state) }),
    listProjectImages: () => listImages(),
    storeCapture: file => uploadCapture(file),
    close: () => onClose()
  };
}
```

- [ ] **Step 5: Run focused adapter and state tests.**

Run: `node --test src/user/pages/agent-director-desk.test.js src/user/components/director-desk/DirectorDeskHostAdapter.test.js`

Expected: PASS.

---

### Task 3: Adapt the Full-Screen Three.js Scene

**Files:**
- Create: `frontend/src/user/components/director-desk/DirectorDeskApp.jsx`
- Create: `frontend/src/user/components/director-desk/DirectorScene.jsx`
- Create: `frontend/src/user/components/director-desk/DirectorDeskApp.test.js`
- Create: `frontend/src/user/components/director-desk/DirectorScene.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Consumes:** Task 1 state functions and Task 2 host adapter.

**Produces:** `DirectorDeskApp({ adapter })`, rendered only when the selected director node opens, and `DirectorScene({ project, images, onProjectChange })`.

- [ ] **Step 1: Write the failing full-screen app contract test.**

```js
test('director desk owns a full-screen workspace and saves through its host adapter', async () => {
  render(<DirectorDeskApp adapter={adapter} />);
  expect(screen.getByRole('heading', { name: '3D导演台' })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: '关闭导演台' })).toBeInTheDocument();
  await userEvent.click(screen.getByRole('button', { name: '导演视角' }));
  expect(adapter.writeState).toHaveBeenCalled();
});
```

- [ ] **Step 2: Define and test the pure scene projection helpers before rendering WebGL.**

Export the following helpers from `DirectorScene.jsx` so the scene contract can be tested without a WebGL context:

```js
export function actorPlaneProps(actor) {
  return {
    position: actor.position,
    rotation: actor.rotation,
    scale: [actor.scale, actor.scale, actor.scale]
  };
}

export function updateCamera(project, camera) {
  return normalizeDirectorProject({ ...project, camera });
}

export function cameraFromControls(camera, target) {
  return {
    position: camera.position.toArray(),
    target: target.toArray(),
    fov: camera.fov
  };
}
```

`SceneBackground({ scene })` accepts a normalized scene asset and draws its image texture on the rear plane. `ImageActorPlane({ actor, asset, onTransform })` accepts one normalized actor, its image asset, and a callback that receives `{ position, rotation, scale }` after a completed drag or transform.

```js
test('scene projects a selected image as an actor plane', () => {
  const actor = actorPlaneProps({ id: 'actor-1', assetId: 'image-1', position: [2, 0, -1], scale: 1.2 });
  expect(actor.position).toEqual([2, 0, -1]);
  expect(actor.scale).toEqual([1.2, 1.2, 1.2]);
});

test('camera framing updates the persisted camera values', () => {
  expect(updateCamera(normalizeDirectorProject(), { position: [0, 3, 8], fov: 38 }).camera.fov).toBe(38);
});
```

- [ ] **Step 3: Implement the minimal lazy-loaded 3D scene.**

```jsx
export function DirectorScene({ project, images, onProjectChange }) {
  return <Canvas camera={{ position: project.camera.position, fov: project.camera.fov }}>
    <ambientLight intensity={1.1} />
    <SceneBackground scene={project.scene} />
    {project.actors.map(actor => <ImageActorPlane key={actor.id} actor={actor} asset={images.find(image => image.id === actor.assetId)} onTransform={next => onProjectChange(updateDirectorActorTransform(project, actor.id, next))} />)}
    <OrbitControls target={project.camera.target} onEnd={event => onProjectChange(updateCamera(project, cameraFromControls(event.target.object, event.target.target)))} />
  </Canvas>;
}
```

`SceneBackground` and `ImageActorPlane` use project images as textured planes. They must tolerate a missing or deleted asset by rendering an in-scene unavailable asset card rather than crashing the canvas.

- [ ] **Step 4: Implement the workspace shell and event isolation.**

```jsx
export default function DirectorDeskApp({ adapter }) {
  const [project, setProject] = useState(() => adapter.readState());
  const save = next => { const normalized = normalizeDirectorProject(next); setProject(normalized); adapter.writeState(normalized); };
  return <section className="agent-director-desk-fullscreen" data-director-desk-open onPointerDown={event => event.stopPropagation()} onWheel={event => event.stopPropagation()} onContextMenu={event => event.stopPropagation()}>
    <header><h1>3D导演台</h1><button aria-label="关闭导演台" onClick={adapter.close}>关闭</button></header>
    <DirectorScene project={project} images={images} onProjectChange={save} />
  </section>;
}
```

The shell must expose director/camera view buttons and a project-image picker. It must register `Ctrl/Cmd+Z`, copy, and paste only while `data-director-desk-open` exists.

- [ ] **Step 5: Add scoped styling and verify tests.**

Run: `node --test src/user/components/director-desk/DirectorDeskApp.test.js src/user/components/director-desk/DirectorScene.test.js`

Expected: PASS with no dependency on a provider or backend generation request.

---

### Task 4: Replace the Old Modal and Deliver Captures to Canvas

**Files:**
- Modify: `frontend/src/user/components/AgentDirectorDeskNode.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentCanvasInteraction.source.test.js`
- Create: `frontend/src/user/components/AgentCanvas.director-desk.source.test.js`

**Consumes:** Tasks 1-3 and existing `uploadAgentAsset`, `createNodeAt`, `createTypedCanvasEdge`, `commit`, and canvas durable-save functions.

**Produces:** Full-screen desk launch, legacy-modal removal, and one durable image node plus edge for each successful `Capture to Canvas` action.

- [ ] **Step 1: Write failing source and pure-flow tests.**

```js
test('canvas opens the full-screen desk instead of the 2.5D modal', async () => {
  const source = await readFile(new URL('./AgentCanvas.jsx', import.meta.url), 'utf8');
  assert.match(source, /DirectorDeskApp/);
  assert.match(source, /createDirectorDeskHostAdapter/);
  assert.doesNotMatch(source, /AgentDirectorDeskWorkspace/);
});

test('a successful director capture creates an image node and one director edge', () => {
  const { nodes, edges } = insertDirectorCapture({ nodes: [directorNode], edges: [], directorNodeID: directorNode.id, asset: { id: 'capture-1', url: '/capture.png', kind: 'image' } });
  assert.equal(nodes.filter(node => node.data.type === 'image').length, 1);
  assert.equal(edges.length, 1);
  assert.equal(edges[0].source, directorNode.id);
});
```

- [ ] **Step 2: Run the tests before implementation.**

Run: `node --test src/user/components/AgentCanvas.director-desk.source.test.js src/user/components/AgentCanvasInteraction.source.test.js`

Expected: FAIL because the full-screen adapter and capture insertion helper are absent.

- [ ] **Step 3: Replace the modal ownership with lazy full-screen ownership.**

```jsx
const DirectorDeskApp = lazy(() => import('./director-desk/DirectorDeskApp'));

{directorDeskNode ? <Suspense fallback={<div className="agent-director-desk-loading">正在加载 3D导演台…</div>}>
  <DirectorDeskApp adapter={directorDeskAdapter} />
</Suspense> : null}
```

Keep only the compact `AgentDirectorDeskNode` card on the React Flow canvas. Remove `AgentDirectorDeskWorkspace` after its replacement has browser acceptance.

- [ ] **Step 4: Define, test, and implement atomic capture insertion after upload succeeds.**

Export `insertDirectorCapture` from `AgentCanvas.jsx` with this exact contract:

```js
export function insertDirectorCapture({ nodes, edges, directorNodeID, asset }) {
  // reject a missing director node or a non-image asset
  // create one standard image node carrying asset.id, asset.url, and asset.name
  // append one typed edge from directorNodeID to the new image node
  // return { nodes: nextNodes, edges: nextEdges }
}
```

Add a focused pure test in `AgentCanvas.director-desk.source.test.js` for a missing director node and a non-image asset, asserting that each throws before mutating either collection.

```js
async function saveDirectorCapture(nodeID, file) {
  const asset = await uploadAgentAsset(taskIdRef.current, file);
  const next = insertDirectorCapture({
    nodes: nodesRef.current,
    edges: edgesRef.current,
    directorNodeID: nodeID,
    asset
  });
  commit(next.nodes, next.edges);
  await persistCurrentCanvas();
  return asset;
}
```

`insertDirectorCapture` must only run after upload resolves. If upload or persistence fails, leave the existing nodes and edges unchanged and surface the returned error inside the desk.

- [ ] **Step 5: Run focused canvas tests.**

Run: `node --test src/user/components/AgentCanvas.director-desk.source.test.js src/user/components/AgentCanvasInteraction.source.test.js src/user/pages/agent-director-desk.test.js`

Expected: PASS.

---

### Task 5: Embedded Runtime and Browser Acceptance

**Files:**
- Modify only when acceptance exposes a defect: files from Tasks 1-4.
- Do not modify generated `frontend/dist` or `backend/web/dist` by hand.

**Consumes:** Completed Tasks 1-4.

**Produces:** A locally rebuilt Go-embedded runtime on `http://localhost:4001` with browser evidence for the first-release acceptance criteria.

- [ ] **Step 1: Run all focused frontend tests.**

Run:

```powershell
node --test `
  src/user/pages/agent-director-desk.test.js `
  src/user/components/director-desk/DirectorDeskHostAdapter.test.js `
  src/user/components/director-desk/DirectorDeskApp.test.js `
  src/user/components/director-desk/DirectorScene.test.js `
  src/user/components/AgentCanvas.director-desk.source.test.js
```

Expected: all tests pass.

- [ ] **Step 2: Run relevant Go storage tests.**

Run: `go test ./internal/agentworkspace ./internal/httpapi -count=1`

Expected: all tests pass, including `TestValidateCanvasForStorageAllowsDirectorNode`.

- [ ] **Step 3: Build and start the embedded local runtime.**

Run:

```powershell
npm --prefix frontend run build
scripts/build-go-local.ps1
scripts/start-local-go-review.ps1
```

Expected: `GET /api/runtime-build-info` reports `runtime: "go"` and the new hashed frontend entry.

- [ ] **Step 4: Perform browser acceptance without model generation.**

1. Open an existing project and select its director node.
2. Open `3D导演台`; verify it occupies the viewport rather than appearing as the old 2.5D modal.
3. Choose a project image as background and add another image as an actor plane.
4. Move the actor and camera, close, reopen, then refresh the browser.
5. Confirm scene, actor, and camera position are restored.
6. Use `Capture to Canvas`; confirm one project image asset, one image node, and one director-to-image edge exist after refresh.
7. Do not click any image or video generation action.

- [ ] **Step 5: Record any non-director baseline failure separately.**

If a broad pre-existing test such as `AgentCanvas.generation-history.source.test.js` fails, report its file, assertion, and relationship to this work. Do not alter it unless the failure is caused by the director-desk implementation.

## Coverage Check

- Full-screen independent workspace: Task 3 and Task 4.
- Same canvas/project persistence: Tasks 1, 2, and 4.
- Project image scene and actor planes: Tasks 2 and 3.
- Camera, transforms, save/reload: Tasks 1, 3, and 5.
- Capture becomes a normal connected image node: Tasks 2, 4, and 5.
- Existing Go-embedded local release path: Task 5.
- No model generation, model library, GLB/FBX import, lighting presets, or multi-camera work: global constraints and Task 5 acceptance scope.
