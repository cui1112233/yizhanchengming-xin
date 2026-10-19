# Agent Canvas Context Menu Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a working right-click menu, project-scoped material library, and TOS-backed upload to the local V88 Agent workspace.

**Architecture:** Extend the one existing React Flow canvas, its history and revision-save queue. Add project-owned asset metadata in MySQL and file bytes in TOS through Go; canvas nodes retain only an `assetRef`.

**Tech Stack:** Go, MySQL, Volcengine TOS SDK, React, Ant Design, `@xyflow/react`, Node built-in test runner.

## Global Constraints

- Work in the current `v88` checkout only; do not create a branch, reset the dirty worktree, or publish publicly.
- The app stays Go-served with an embedded React build.
- Add Agent schema only as idempotent append-only migration `8800004`, after `8800003`; never modify an applied migration.
- Do not copy Shenlang code, CSS, icons, credentials, paths, or cloud contracts.
- Canvas edits use the existing `PATCH /api/agent/tasks/{id}/canvas` revision guard.
- MySQL stores asset records, TOS stores bytes, and `assetRef` never contains a browser blob URL or disk path.
- Preserve the current V88 colors, responsive layout, conversation panel, and local-only boundary.

---

## File Structure

- `frontend/src/user/pages/agent-canvas-context-menu.js`: pure canvas menu, clipboard and edge-pruning commands.
- `frontend/src/user/pages/agent-canvas-context-menu.test.js`: command tests.
- `frontend/src/user/components/AgentCanvas.jsx`: event handlers, menu, picker and upload state.
- `frontend/src/shared/styles/agent-workspace.css`: scoped menus and responsive picker styles.
- `frontend/src/shared/api/agent.js`: Agent-asset browser requests.
- `backend/internal/agentworkspace/assets.go`: asset type, validation and TOS object-key contract.
- `backend/internal/agentworkspace/mysql_assets.go`: MySQL asset storage.
- `backend/internal/agentworkspace/memory_assets.go`: in-memory storage for HTTP tests.
- `backend/internal/httpapi/agent_assets.go`: authenticated list and upload API.
- `backend/internal/storage/agent_workspace_schema.go`: migration `8800004`.
- `backend/internal/config/config.go`, `backend/internal/app/app.go`, `backend/internal/httpapi/router.go`: TOS config and dependency wiring.

## Shared Interfaces

```go
type Asset struct {
    ID string `json:"id"`
    ProjectID string `json:"projectId"`
    Kind string `json:"kind"` // image or video
    Name string `json:"name"`
    ContentType string `json:"contentType"`
    ObjectKey string `json:"-"`
    URL string `json:"url"`
    ByteSize int64 `json:"byteSize"`
    CreatedAt time.Time `json:"createdAt"`
}
type AssetStore interface {
    ListAssets(context.Context, string, string) ([]Asset, error)
    CreateAsset(context.Context, string, Asset) error
    Asset(context.Context, string, string, string) (Asset, error)
}
type AssetObjectStore interface { Put(context.Context, Asset, io.Reader) error }
```

```js
export function createNodeAt({ id, kind, position, asset })
export function copyCanvasSelection({ nodes, edges, selectedIds })
export function pasteCanvasSelection({ clipboard, position, nextID })
export function removeCanvasNodes({ nodes, edges, ids })
export function listAgentAssets(taskId, kind = '')
export function uploadAgentAsset(taskId, file)
```

### Task 1: Pure context-menu commands

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-context-menu.js`
- Create: `frontend/src/user/pages/agent-canvas-context-menu.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-commands.js`

**Consumes:** Current React Flow nodes, edges and `createCanvasHistory` snapshots.

**Produces:** Create, copy, paste and delete operations which later UI code can commit as one history entry.

- [ ] **Step 1: Write the failing test**

```js
test('pasting a copied pair recreates only its internal edge at the click position', () => {
  const clipboard = copyCanvasSelection({ nodes: [imageNode, videoNode], edges: [edge], selectedIds: ['image-1', 'video-1'] });
  const pasted = pasteCanvasSelection({ clipboard, position: { x: 500, y: 240 }, nextID: value => `copy-${value}` });
  assert.deepEqual(pasted.nodes.map(node => node.id), ['copy-image-1', 'copy-video-1']);
  assert.equal(pasted.edges[0].source, 'copy-image-1');
  assert.equal(pasted.edges[0].target, 'copy-video-1');
});
test('removing a node also removes every attached edge', () => {
  const result = removeCanvasNodes({ nodes: [textNode, imageNode], edges: [edge], ids: ['text-1'] });
  assert.equal(result.nodes.length, 1); assert.equal(result.edges.length, 0);
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js`  
Expected: FAIL because the helper module is absent.

- [ ] **Step 3: Implement minimal commands**

```js
export function createNodeAt({ id, kind, position, asset = null }) {
  const media = kind === 'image' || kind === 'video';
  return { id, type: 'agent', position: { x: Number(position.x) || 0, y: Number(position.y) || 0 }, selected: true,
    data: { type: kind, title: media ? asset?.name || `新${kind === 'image' ? '图片' : '视频'}` : '新文本',
      summary: '', version: 1, sourceNodeIds: [], status: 'draft', assetRef: media ? asset : null } };
}
```

Copy only edges where both endpoints are selected, create new IDs with a source-to-copy map, append `副本` once to copied titles, and remove every edge attached to deleted nodes.

- [ ] **Step 4: Run the command tests**

Run: `node --test frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`  
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/user/pages/agent-canvas-commands.js frontend/src/user/pages/agent-canvas-context-menu.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "feat(agent): add canvas context commands"
```

### Task 2: Right-click canvas UI

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`
- Test: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Consumes:** Task 1 helpers plus existing `commit`, `applyHistory`, selection and `screenToFlowPosition` behavior.

**Produces:** Empty-canvas, node and multi-select menus for creation, copy/paste, deletion, disconnection, grouping, alignment, undo and redo.

- [ ] **Step 1: Write the failing test**

```js
test('canvas registers pane and node context-menu handlers', () => {
  assert.match(canvasSource, /onPaneContextMenu=\{handlePaneContextMenu\}/);
  assert.match(canvasSource, /onNodeContextMenu=\{handleNodeContextMenu\}/);
  assert.match(canvasSource, /handleContextPaste/);
  assert.match(canvasSource, /handleContextDelete/);
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`  
Expected: FAIL because handlers and menu markup are absent.

- [ ] **Step 3: Implement menu state and handlers**

```jsx
function openContextMenu(event, target, nodeIds = []) {
  event.preventDefault();
  const position = flowInstance?.screenToFlowPosition({ x: event.clientX, y: event.clientY }) || { x: 0, y: 0 };
  setContextMenu({ clientX: event.clientX, clientY: event.clientY, position, target, nodeIds });
}
function handleContextDelete() {
  const result = removeCanvasNodes({ nodes: nodesRef.current, edges: edgesRef.current, ids: contextMenu.nodeIds });
  setSelectedIds([]); commit(result.nodes, result.edges); setContextMenu(null);
}
```

Use an Ant Design menu inside `.agent-canvas`, clamp it to canvas bounds, and close it on `Escape`, pane click, drag, project reload or command completion. Reuse existing group/ungroup/align/distribute functions; do not duplicate them.

- [ ] **Step 4: Add scoped styles**

```css
.agent-page .agent-canvas-context-menu { position: absolute; z-index: 30; min-width: 184px; }
.agent-page .agent-canvas-context-menu .ant-dropdown-menu { border-radius: 8px; }
@media (max-width: 560px) { .agent-page .agent-canvas-context-menu { max-width: calc(100% - 16px); } }
```

- [ ] **Step 5: Run the tests and commit**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-workspace.layout.test.js`  
Expected: PASS.

```bash
git add frontend/src/user/components/AgentCanvas.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-workspace.layout.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "feat(agent): add canvas context menu"
```

### Task 3: Agent asset domain and schema

**Files:**
- Create: `backend/internal/agentworkspace/assets.go`
- Create: `backend/internal/agentworkspace/mysql_assets.go`
- Create: `backend/internal/agentworkspace/memory_assets.go`
- Create: `backend/internal/agentworkspace/asset_store_test.go`
- Modify: `backend/internal/storage/agent_workspace_schema.go`
- Modify: `backend/internal/storage/agent_workspace_schema_test.go`

**Consumes:** Existing owner/project MySQL conventions.

**Produces:** Validated asset records owned by one Agent project and migration `8800004`.

- [ ] **Step 1: Write failing store and schema tests**

```go
func TestAssetStoreIsOwnerAndProjectScoped(t *testing.T) {
  store := NewMemoryAssetStore()
  asset := Asset{ID: "asset-1", ProjectID: "agent-1", Kind: "image", Name: "cover.png", ContentType: "image/png", ObjectKey: "agent-assets/a/agent-1/asset-1.png", ByteSize: 12}
  require.NoError(t, store.CreateAsset(ctx, "alice", asset))
  _, err := store.Asset(ctx, "bob", "agent-1", "asset-1")
  require.ErrorIs(t, err, ErrAssetNotFound)
}
```

Add a migration test for table `agent_assets` and `idx_agent_assets_owner_project_created`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/agentworkspace ./internal/storage -run 'TestAsset|TestAgentWorkspace' -count=1` from `backend`  
Expected: FAIL because Asset and migration `8800004` are absent.

- [ ] **Step 3: Implement stores and migration**

```go
func AgentAssetObjectKey(owner, projectID, assetID, extension string) (string, error) {
  if !safeID.MatchString(projectID) || !safeID.MatchString(assetID) { return "", ErrInvalidAsset }
  sum := sha256.Sum256([]byte(strings.TrimSpace(owner)))
  return "agent-assets/" + hex.EncodeToString(sum[:12]) + "/" + projectID + "/" + assetID + "." + extension, nil
}
```

Append `CREATE TABLE IF NOT EXISTS agent_assets`, primary key `(project_id, id)` and owner/project/created index. Allow only allowlisted `image/*` and `video/*` types. Another owner's asset always returns `ErrAssetNotFound`.

- [ ] **Step 4: Run tests and commit**

Run: `go test ./internal/agentworkspace ./internal/storage -count=1` from `backend`  
Expected: PASS.

```bash
git add backend/internal/agentworkspace backend/internal/storage/agent_workspace_schema.go backend/internal/storage/agent_workspace_schema_test.go
git commit -m "feat(agent): add project asset storage"
```

### Task 4: TOS-backed asset HTTP API

**Files:**
- Create: `backend/internal/httpapi/agent_assets.go`
- Create: `backend/internal/httpapi/agent_assets_test.go`
- Modify: `backend/internal/httpapi/agent_workspace_compat.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/app/app.go`

**Consumes:** Task 3 stores and current Volcengine TOS SDK.

**Produces:** Authenticated project-asset list/upload endpoints and explicit unconfigured-storage errors.

- [ ] **Step 1: Write failing HTTP tests**

```go
func TestAgentAssetUploadCreatesProjectScopedRecord(t *testing.T) {
  body := multipartBody(t, "file", "cover.png", "image/png", []byte("png"))
  response := agentRequest(t, mux, http.MethodPost, "/api/agent/tasks/agent-1/assets", body)
  require.Equal(t, http.StatusCreated, response.Code)
  require.Contains(t, response.Body.String(), `"kind":"image"`)
}
```

Also test an invalid media type returns 400, another project returns 404, and unavailable TOS returns 503 without creating a database row.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/httpapi -run TestAgentAsset -count=1` from `backend`  
Expected: FAIL because the routes do not exist.

- [ ] **Step 3: Implement routes, object store and configuration**

```text
GET  /api/agent/tasks/{id}/assets?kind=image|video
POST /api/agent/tasks/{id}/assets
```

Read one bounded multipart file, sniff and validate its media type, generate an asset ID/key, put bytes to TOS, then insert metadata. If MySQL insert fails after upload, delete the TOS object. Add five explicit Agent TOS config values (endpoint, region, bucket, access key, secret key); create the object store only when all are present. Return an authenticated Agent URL, never a raw bucket path.

- [ ] **Step 4: Run tests and commit**

Run: `go test ./internal/httpapi ./internal/config ./internal/app -count=1` from `backend`  
Expected: PASS.

```bash
git add backend/internal/httpapi backend/internal/config backend/internal/app backend/internal/agentworkspace/assets.go
git commit -m "feat(agent): add TOS asset upload API"
```

### Task 5: Asset picker and material nodes

**Files:**
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Consumes:** Tasks 2 and 4.

**Produces:** Image/video drawer and upload flow which create or replace nodes only after server confirmation.

- [ ] **Step 1: Write a failing frontend test**

```js
test('canvas asset actions load and upload project-scoped assets before committing a node', () => {
  assert.match(canvasSource, /listAgentAssets\(taskId, assetKind\)/);
  assert.match(canvasSource, /uploadAgentAsset\(taskId, file/);
  assert.match(canvasSource, /createNodeAt\(\{[\s\S]*asset/);
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-workspace.layout.test.js`  
Expected: FAIL because the asset picker is absent.

- [ ] **Step 3: Implement API, drawer and upload**

```js
export function uploadAgentAsset(taskId, file) {
  const body = new FormData(); body.append('file', file);
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/assets`, { method: 'POST', body });
}
```

Use an Ant Design drawer with image/video filters and multi-select. The hidden upload input accepts `image/*,video/*`; show inline errors, create no node until an asset response succeeds, and record the add/replace as one history step. Replacing media preserves node ID, coordinates, edges and title. Render only `assetRef.url`; show a labeled placeholder on load failure.

- [ ] **Step 4: Run tests and build, then commit**

Run: `node --test src/user/pages/agent-canvas-context-menu.test.js src/user/pages/agent-workspace.layout.test.js && npm run build` from `frontend`  
Expected: PASS.

```bash
git add frontend/src/shared/api/agent.js frontend/src/user/components/AgentCanvas.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-workspace.layout.test.js
git commit -m "feat(agent): add canvas asset picker"
```

### Task 6: Embedded local runtime acceptance

**Files:**
- Modify only verification tests required by observed browser behavior.

**Consumes:** Tasks 1-5.

**Produces:** Local embedded Go runtime and authenticated browser evidence for all completed operations.

- [ ] **Step 1: Run focused regression suites**

```powershell
Push-Location backend; go test ./internal/agentworkspace ./internal/storage ./internal/httpapi ./internal/app -count=1; Pop-Location
Push-Location frontend; node --test src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-context-menu.test.js src/user/pages/agent-workspace.layout.test.js; npm run build; Pop-Location
```

Expected: PASS; record unrelated existing failures without changing their code.

- [ ] **Step 2: Rebuild and start the local Go review runtime**

```powershell
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql
.\scripts\check-local-go.ps1
```

Expected: the rebuilt Go binary serves `http://localhost:4001`.

- [ ] **Step 3: Perform authenticated browser acceptance**

Verify empty-canvas creation, copy/paste/delete/undo/redo and refresh restoration; upload one image and one video; add assets from the picker; replace an asset without moving or disconnecting its node; verify a second project cannot list or bind the first project's assets; verify a narrow viewport keeps menus and the composer visible.

- [ ] **Step 4: Commit only a verification-driven test correction**

```bash
git add frontend/src/user/pages/agent-workspace.layout.test.js
git commit -m "test(agent): cover context menu acceptance"
```

Do not create an empty commit when browser acceptance needs no test adjustment.

## Plan Self-Review

- Tasks 1-2 cover menus and reversible canvas operations; Tasks 3-5 cover MySQL/TOS assets, upload, picker and replacement; Task 6 covers Go embedding and authenticated browser acceptance.
- AI image/video/voice generation task execution is deliberately excluded, matching the approved design's later phase C.
- Every consumer uses the shared interfaces above, and every task has a focused failing-test, passing-test and commit step.
