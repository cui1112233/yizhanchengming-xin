# Agent Canvas Production Chain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a durable script-to-shot production chain to the existing Agent Canvas, including selected-shot batch generation and traceable result versions.

**Architecture:** Reuse the existing `agent_generation_tasks`, canvas nodes, project assets, storyboard projection, and provider/model APIs. Add a small durable shot domain that links a script node to ordered shots and links each output node/result to one shot; the active node asset remains the active result version used by the storyboard and downstream references.

**Tech Stack:** Go, MySQL, Redis-backed existing generation tasks, TOS project assets, React 18, Ant Design 5, `@xyflow/react`, Node tests, Go tests.

## Global Constraints

- Do not replace the workflow, storyboard, project conversation, account permissions, or Go-embedded release path.
- MySQL changes are append-only; use migration `8800011` only after confirming it remains unused at implementation time.
- Script splitting, shots, asset lineage, and active-version selection are owner/project scoped.
- A submitted generation is not complete until the current generation task reports a persisted result asset.
- Existing image, video, audio, text nodes and provider-model capability selection remain compatible.
- Do not commit, push, or deploy without explicit user instruction.

---

## File Map

- `backend/internal/agentworkspace/script_shots.go`: script, shot, and output-lineage contracts plus store interface.
- `backend/internal/agentworkspace/mysql_script_shots.go`: MySQL implementation for script shots and selected version links.
- `backend/internal/storage/agent_workspace_schema.go`: migration `8800011`.
- `backend/internal/httpapi/agent_script_shots.go`: split, list, update, and batch-create routes.
- `frontend/src/shared/api/agent.js`: script-shot client functions.
- `frontend/src/user/pages/agent-script-shots.js`: pure projection, selection, and batch payload helpers.
- `frontend/src/user/components/AgentScriptNodeComposer.jsx`: script node editing, splitting status, and shot selection.
- `frontend/src/user/components/AgentStoryboardWorkspace.jsx`: shot section and locate/edit controls over the same project records.
- `frontend/src/user/components/AgentWorkspaceDrawer.jsx`: result-library filters and version/source labels.
- `frontend/src/user/components/AgentCanvas.jsx`: script node rendering and create-node/menu integration.

### Task 1: Durable script, shot, and lineage records

**Files:**
- Create: `backend/internal/agentworkspace/script_shots.go`
- Create: `backend/internal/agentworkspace/script_shots_test.go`
- Create: `backend/internal/agentworkspace/mysql_script_shots.go`
- Create: `backend/internal/agentworkspace/mysql_script_shots_test.go`
- Modify: `backend/internal/storage/agent_workspace_schema.go`
- Modify: `backend/internal/storage/agent_workspace_schema_test.go`

**Interfaces:**
- Produces `ScriptShot`, `ScriptSplit`, `ShotGenerationLink`, and `ScriptShotStore`.
- `ScriptShotStore` methods: `ReplaceShots`, `ListShots`, `UpdateShot`, `AttachGeneration`, `ListGenerationLinks`.

- [ ] **Step 1: Write the failing ownership, ordering, and lineage tests**

```go
func TestScriptShotStoreKeepsOrderedOwnerScopedShotsAndLineage(t *testing.T) {
    err := store.ReplaceShots(ctx, "alice", "project-1", "script-1", []ScriptShot{{ID: "shot-1", Sequence: 1}})
    require.NoError(t, err)
    require.Empty(t, must(store.ListShots(ctx, "bob", "project-1", "script-1")))
    require.NoError(t, store.AttachGeneration(ctx, "alice", ShotGenerationLink{
        ProjectID: "project-1", ScriptNodeID: "script-1", ShotID: "shot-1", NodeID: "image-1", GenerationID: "gen-1", Attempt: 1,
    }))
    require.Equal(t, "gen-1", must(store.ListGenerationLinks(ctx, "alice", "project-1", "shot-1"))[0].GenerationID)
}
```

- [ ] **Step 2: Run the new test to verify it fails**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/storage -run 'TestScriptShot|TestAgentWorkspaceSchema.*Script' -count=1`

Expected: FAIL because the script-shot contracts and migration do not exist.

- [ ] **Step 3: Implement the records and append migration `8800011`**

```go
type ScriptShot struct {
    ID, ProjectID, ScriptNodeID, Scene, Characters, Dialogue, Narration string
    Sequence int
    ImagePrompt, VideoPrompt, TTSText string
    Revision int64
    CreatedAt, UpdatedAt time.Time
}

type ShotGenerationLink struct {
    ProjectID, ScriptNodeID, ShotID, NodeID, GenerationID string
    Attempt int
    Active bool
}
```

Create `agent_script_shots` with `(project_id, id)` primary key and unique `(project_id, script_node_id, sequence)`. Create `agent_shot_generation_links` with unique `(project_id, shot_id, node_id, generation_id)`. All reads and writes include `owner_username`; `UpdateShot` increments `revision` only when the caller supplies the current revision.

- [ ] **Step 4: Run store/schema tests**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/storage -run 'TestScriptShot|TestMySQLScriptShot|TestAgentWorkspaceSchema' -count=1`

Expected: PASS.

### Task 2: Script splitting and selected-shot batch API

**Files:**
- Create: `backend/internal/httpapi/agent_script_shots.go`
- Create: `backend/internal/httpapi/agent_script_shots_test.go`
- Modify: `backend/internal/httpapi/agent_workspace_compat.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/agentworkspace/script_shots.go`

**Interfaces:**
- Produces `POST /api/agent/tasks/{id}/script-splits`, `GET /script-nodes/{nodeId}/shots`, `PATCH /shots/{shotId}`, and `POST /shots/batch-nodes`.
- Batch response: `{ canvas, nodes, generationLinks, skipped: [{ shotId, reason }] }`.

- [ ] **Step 1: Write failing route tests**

```go
func TestScriptSplitAndBatchNodesStayProjectScoped(t *testing.T) {
    split := postJSON(t, alice, "/api/agent/tasks/project-1/script-splits", map[string]any{"nodeId":"script-1", "nodeVersion":1, "modelId":"text-model"})
    require.Equal(t, http.StatusAccepted, split.Code)
    shots := getJSON(t, alice, "/api/agent/tasks/project-1/script-nodes/script-1/shots")
    require.Equal(t, http.StatusOK, shots.Code)
    require.Equal(t, http.StatusNotFound, getJSON(t, bob, "/api/agent/tasks/project-1/script-nodes/script-1/shots").Code)
}
```

- [ ] **Step 2: Run route tests to verify failure**

Run: `Set-Location backend; go test ./internal/httpapi -run 'TestScriptSplit|TestShotBatch' -count=1`

Expected: FAIL because the script-shot routes do not exist.

- [ ] **Step 3: Implement split and batch behavior**

```go
type batchShotNodesRequest struct {
    ScriptNodeID string   `json:"scriptNodeId"`
    ShotIDs      []string `json:"shotIds"`
    Kind         string   `json:"kind"` // image, video, audio
    ModelID      string   `json:"modelId"`
}
```

Validate project ownership, source script node type/version, configured model capability, and non-empty unique shot IDs. The split task persists structured shots only after its text result succeeds. Batch creation creates one node per shot with `config.source = { scriptNodeId, shotId, operation, attempt }`, plus a `ShotGenerationLink`; it reuses the existing image/video/audio submission path instead of calling providers directly from this route.

- [ ] **Step 4: Run API regression tests**

Run: `Set-Location backend; go test ./internal/httpapi ./internal/agentworkspace -run 'Test(ScriptSplit|ShotBatch|AgentGeneration|AgentVideo)' -count=1`

Expected: PASS, including existing generation routes.

### Task 3: Frontend script node and shot selection

**Files:**
- Create: `frontend/src/user/pages/agent-script-shots.js`
- Create: `frontend/src/user/pages/agent-script-shots.test.js`
- Create: `frontend/src/user/components/AgentScriptNodeComposer.jsx`
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-node-presentation.js`

**Interfaces:**
- Produces `normalizeScriptShots(response)`, `selectedShotIDs(shots, selection)`, and `buildShotBatchPayload(input)`.
- Adds canvas node type `script` with `config.sourceText`, `config.splitModelId`, and `config.splitState`.

- [ ] **Step 1: Write the failing pure UI-state tests**

```js
test('selected shots produce source-aware batch payloads', () => {
  const payload = buildShotBatchPayload({ scriptNodeId: 'script-1', selectedIDs: ['shot-2'], kind: 'video', modelId: 'video-1' });
  assert.deepEqual(payload, { scriptNodeId: 'script-1', shotIds: ['shot-2'], kind: 'video', modelId: 'video-1' });
});

test('normalization retains provider failure text without marking split complete', () => {
  assert.equal(normalizeScriptShots({ state: 'failed', errorMessage: '上游失败' }).state, 'failed');
});
```

- [ ] **Step 2: Run frontend test to verify failure**

Run: `node --test frontend/src/user/pages/agent-script-shots.test.js`

Expected: FAIL because script-shot helpers and the script node do not exist.

- [ ] **Step 3: Implement API client, node editor, and menu entry**

```js
export function createAgentScriptSplit(taskId, payload) {
  return apiRequest(`/api/agent/tasks/${encodeURIComponent(taskId)}/script-splits`, { method: 'POST', body: JSON.stringify(payload), suppressGlobalError: true });
}

export function buildShotBatchPayload({ scriptNodeId, selectedIDs, kind, modelId }) {
  return { scriptNodeId, shotIds: [...new Set(selectedIDs)].filter(Boolean), kind, modelId };
}
```

`AgentScriptNodeComposer` displays source text, model picker, actual split task state, structured shot rows, and selection controls. `AgentCanvas` renders script nodes, supports creation from its existing add-node menu, and keeps node selection/editing behavior consistent with existing node types.

- [ ] **Step 4: Run focused frontend tests and build**

Run: `node --test frontend/src/user/pages/agent-script-shots.test.js frontend/src/user/pages/agent-canvas-node-presentation.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

### Task 4: Storyboard projection, result library, and active-version visibility

**Files:**
- Create: `frontend/src/user/pages/agent-result-lineage.js`
- Create: `frontend/src/user/pages/agent-result-lineage.test.js`
- Modify: `frontend/src/user/pages/agent-storyboard-model.js`
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Modify: `frontend/src/user/components/AgentWorkspaceDrawer.jsx`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Produces `lineageLabel(node, generation)`, `groupResultLibrary({ shots, nodes, generations, assets })`, and `locateShotNode(nodes, shotId)`.
- Consumes existing active node `assetRef`, generic generation task results, and `config.source` from Task 2.

- [ ] **Step 1: Write failing projection tests**

```js
test('result library groups the active version beneath its source shot', () => {
  const groups = groupResultLibrary({ shots: [{ id: 'shot-1', sequence: 1 }], nodes: [node], generations: [generation], assets: [asset] });
  assert.equal(groups[0].items[0].active, true);
  assert.equal(groups[0].items[0].source.shotId, 'shot-1');
});

test('locateShotNode returns the source script node before derived media nodes', () => {
  assert.equal(locateShotNode(nodes, 'shot-1').id, 'script-1');
});
```

- [ ] **Step 2: Run projection tests to verify failure**

Run: `node --test frontend/src/user/pages/agent-result-lineage.test.js`

Expected: FAIL because the lineaged grouping is absent.

- [ ] **Step 3: Implement shared storyboard/library projection**

Extend `buildStoryboardLibrary` with a `shots` section sourced from the API, not duplicated into `canvas.viewState`. The selected shot inspector exposes locate-source, edit-shot, selected-shot batch actions, and a compact list of active/older result versions. The workspace drawer shows source labels such as `镜头 03 · 视频生成 · 第 2 次` and locates the workflow node on selection.

- [ ] **Step 4: Run storyboard/drawer tests and build**

Run: `node --test frontend/src/user/pages/agent-result-lineage.test.js frontend/src/user/pages/agent-storyboard-model.test.js frontend/src/user/pages/agent-workspace-drawer.source.test.js frontend/src/user/pages/agent-storyboard-workspace.source.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

### Task 5: Local Go-embedded acceptance

**Files:**
- Create: `test/agent-canvas-production-chain.source.test.js`
- Modify: `docs/V88_PROJECT_EXECUTION_MEMORY.md` only after all checks have passed.

**Interfaces:**
- Consumes the APIs and UI from Tasks 1-4.
- Produces focused source guards and recorded local verification evidence.

- [ ] **Step 1: Add the route and UI source guards**

```js
test('production chain keeps one source lineage across script, canvas, and storyboard', () => {
  assert.match(read('frontend/src/user/components/AgentCanvas.jsx'), /script/);
  assert.match(read('frontend/src/user/components/AgentStoryboardWorkspace.jsx'), /shot/);
  assert.match(read('backend/internal/httpapi/agent_script_shots.go'), /script-splits/);
});
```

- [ ] **Step 2: Run focused tests, build, and local Go checks**

Run: `node --test frontend/src/user/pages/agent-script-shots.test.js frontend/src/user/pages/agent-result-lineage.test.js test/agent-canvas-production-chain.source.test.js; npm --prefix frontend run build; Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi ./internal/storage -count=1; go build ./cmd/api`

Expected: PASS. Record unrelated failures separately rather than treating them as this feature's acceptance.

- [ ] **Step 3: Verify the real rebuilt local page**

Run: `Set-Location F:\脚本测试\一战晟铭-v88; .\scripts\build-go-local.ps1; .\scripts\start-local-go-review.ps1 -UseYcmMysql; .\scripts\check-local-go.ps1`

In an authenticated browser at `http://localhost:4001/agent/canvas`, create a script node, split a small script, edit one shot, select two shots, batch-create TTS and image nodes, force/observe one failed task, retry it, activate a result version, reload, locate the asset from the storyboard and result library, and confirm the same source labels remain visible.

- [ ] **Step 4: Update the local execution record only after acceptance**

Record the actual local build SHA, commands/results, screenshots, providers exercised, and any unavailable provider separately in `docs/V88_PROJECT_EXECUTION_MEMORY.md`. Do not claim successful provider generation when the configured upstream fails.
