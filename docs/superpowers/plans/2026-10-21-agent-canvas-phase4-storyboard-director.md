# Agent Canvas Phase 4: Storyboard and Director Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a three-column storyboard over the same project assets and give Director Agent supervised and explicitly automatic workflow actions.

**Architecture:** Storyboard state is a projection of canvas nodes/assets plus persisted ordering and category metadata; it never duplicates source media. Director requests compile into typed, reviewable canvas commands. The Go command endpoint enforces ownership, expected revision, mode, deletion limits, and task permissions before applying commands through the existing project store.

**Tech Stack:** React 18, Ant Design 5, Go, MySQL, existing Agent conversation API, Node and Go tests.

## Global Constraints

- Workflow and storyboard are two views of one project and share durable assets, node configs, and revisions.
- Storyboard columns are Director conversation, categorized materials, and selected-material preview/editor.
- `添加到对话` sends durable IDs and safe metadata, never private object URLs or credentials.
- Supervised mode is the default and requires approval for prompts, structure changes, generation, and result application.
- Automatic mode begins only after an explicit user request such as `全自动化`.
- Automatic mode still stops for missing configuration, unsafe durable-asset deletion, cost/limit failures, and contradictory instructions.
- Ambiguous user intent produces one focused question.
- Director may remove canvas nodes/edges but may not silently delete durable project assets.

---

## File Map

- `frontend/src/user/pages/agent-storyboard-state.js`: category projection, ordering, selection, and shared updates.
- `frontend/src/user/components/AgentStoryboard.jsx`: resizable three-column shell.
- `frontend/src/user/components/AgentStoryboardMaterials.jsx`: categorized material explorer.
- `frontend/src/user/components/AgentStoryboardInspector.jsx`: selected material preview/editor.
- `frontend/src/user/pages/agent-director-commands.js`: frontend command preview and labels.
- `backend/internal/agentworkspace/director_commands.go`: typed command validation/application.
- `backend/internal/httpapi/agent_director.go`: propose/approve/execute endpoints.

### Task 1: Storyboard projection and persistence

**Files:**
- Create: `frontend/src/user/pages/agent-storyboard-state.js`
- Create: `frontend/src/user/pages/agent-storyboard-state.test.js`
- Modify: `frontend/src/user/pages/agent-workspace-view.js`
- Modify: `frontend/src/user/pages/agent-workspace-view.test.js`

**Interfaces:**
- Produces: `buildStoryboardMaterials({ nodes, assets, storyboard })`, `patchStoryboardMaterial(canvas, materialId, patch)`, and `normalizeStoryboardState(value)`.
- Material shape: `{ id, sourceType, sourceId, kind, category, title, thumbnailAssetId, order, selected }`.

- [ ] **Step 1: Write failing shared-identity tests**

```js
test('storyboard material edits update the owning node without creating a duplicate', () => {
  const materials = buildStoryboardMaterials({ nodes, assets, storyboard: {} });
  const updated = patchStoryboardMaterial(canvas, materials[0].id, { title: '雨夜开场' });
  assert.equal(updated.nodes[0].title, '雨夜开场');
  assert.equal(updated.nodes.length, canvas.nodes.length);
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-storyboard-state.test.js frontend/src/user/pages/agent-workspace-view.test.js`

Expected: FAIL because storyboard projection and persistence do not exist.

- [ ] **Step 3: Implement normalized storyboard state**

```js
export const STORYBOARD_CATEGORIES = Object.freeze({
  text: ['script', 'other_text'], image: ['character', 'scene', 'object', 'other_image'],
  video: ['clip', 'final_video'], audio: ['voice', 'music', 'effect'],
});
```

Persist only ordering, category overrides, selected material, and column sizes. Resolve title/preview/config from the owning node or asset each time.

- [ ] **Step 4: Run state tests**

Run: `node --test frontend/src/user/pages/agent-storyboard-state.test.js frontend/src/user/pages/agent-workspace-view.test.js`

Expected: PASS.

- [ ] **Step 5: Commit storyboard state**

```powershell
git add frontend/src/user/pages/agent-storyboard-state.js frontend/src/user/pages/agent-storyboard-state.test.js frontend/src/user/pages/agent-workspace-view.js frontend/src/user/pages/agent-workspace-view.test.js
git commit -m "feat(agent): persist shared storyboard state"
```

### Task 2: Three-column storyboard UI

**Files:**
- Create: `frontend/src/user/components/AgentStoryboard.jsx`
- Create: `frontend/src/user/components/AgentStoryboardMaterials.jsx`
- Create: `frontend/src/user/components/AgentStoryboardInspector.jsx`
- Create: `frontend/src/user/pages/agent-storyboard.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: shared task conversation, projected materials, selected material, and node-config updates.
- Produces: persisted column sizes, category/search/filter selection, `onAttachToConversation(materialId)`, and `onMaterialPatch(materialId, patch)`.

- [ ] **Step 1: Write failing structure test**

```js
test('storyboard has director, material, and inspector columns', () => {
  const source = read('frontend/src/user/components/AgentStoryboard.jsx');
  assert.match(source, /AgentDirectorConversation/);
  assert.match(source, /AgentStoryboardMaterials/);
  assert.match(source, /AgentStoryboardInspector/);
});
```

- [ ] **Step 2: Run source test and verify failure**

Run: `node --test frontend/src/user/pages/agent-storyboard.source.test.js`

Expected: FAIL because storyboard components are absent.

- [ ] **Step 3: Implement the resizable shell and material inspector**

```jsx
<section className="agent-storyboard" style={{ '--director-width': `${directorWidth}px`, '--materials-width': `${materialsWidth}px` }}>
  <AgentDirectorConversation />
  <AgentStoryboardMaterials />
  <AgentStoryboardInspector />
</section>
```

The inspector reuses the same normalized config helpers and model/settings components as the canvas editor. At widths below 900 px, material and inspector panels become switchable tabs while Director remains reachable without overlap.

- [ ] **Step 4: Run source/state tests and build**

Run: `node --test frontend/src/user/pages/agent-storyboard.source.test.js frontend/src/user/pages/agent-storyboard-state.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit storyboard UI**

```powershell
git add frontend/src/user/components/AgentStoryboard.jsx frontend/src/user/components/AgentStoryboardMaterials.jsx frontend/src/user/components/AgentStoryboardInspector.jsx frontend/src/user/pages/agent-storyboard.source.test.js frontend/src/user/pages/AgentPageV2.jsx frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): add shared project storyboard"
```

### Task 3: Typed Director command model

**Files:**
- Create: `backend/internal/agentworkspace/director_commands.go`
- Create: `backend/internal/agentworkspace/director_commands_test.go`
- Create: `frontend/src/user/pages/agent-director-commands.js`
- Create: `frontend/src/user/pages/agent-director-commands.test.js`

**Interfaces:**
- Command types: `create_node`, `update_node`, `connect_nodes`, `disconnect_nodes`, `move_nodes`, `remove_nodes`, `start_generation`, and `apply_result`.
- Produces: `ValidateDirectorPlan(project, plan) error`, `ApplyDirectorPlan(ctx, owner, projectID, expectedRevision, plan)`, and `directorPlanPresentation(plan)`.

- [ ] **Step 1: Write failing permission and revision tests**

```go
func TestDirectorPlanRejectsDurableAssetDeleteAndRevisionConflict(t *testing.T) {
    require.ErrorIs(t, ValidateDirectorPlan(project, planWith("delete_asset")), ErrDirectorCommandForbidden)
    _, err := ApplyDirectorPlan(ctx, store, "alice", project.ID, project.Canvas.Revision-1, validPlan)
    require.ErrorIs(t, err, ErrRevisionConflict)
}
```

- [ ] **Step 2: Run Go and frontend tests and verify failure**

Run: `Set-Location backend; go test ./internal/agentworkspace -run TestDirector -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-director-commands.test.js`

Expected: FAIL because typed Director plans are absent.

- [ ] **Step 3: Implement typed validation and readable previews**

```go
type DirectorCommand struct {
    Type string `json:"type"`
    NodeID string `json:"nodeId,omitempty"`
    SourceNodeID string `json:"sourceNodeId,omitempty"`
    TargetNodeID string `json:"targetNodeId,omitempty"`
    Patch map[string]any `json:"patch,omitempty"`
}
```

Reject unknown commands, cross-project IDs, durable asset deletion, unsupported node kinds, invalid edges, and generation without a configured compatible model. Apply the full approved plan in one expected-revision transaction.

- [ ] **Step 4: Run command tests**

Run: `Set-Location backend; go test ./internal/agentworkspace -run TestDirector -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-director-commands.test.js`

Expected: PASS.

- [ ] **Step 5: Commit Director commands**

```powershell
git add backend/internal/agentworkspace/director_commands.go backend/internal/agentworkspace/director_commands_test.go frontend/src/user/pages/agent-director-commands.js frontend/src/user/pages/agent-director-commands.test.js
git commit -m "feat(agent): define director canvas commands"
```

### Task 4: Supervised and explicit automatic execution

**Files:**
- Create: `backend/internal/httpapi/agent_director.go`
- Create: `backend/internal/httpapi/agent_director_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `frontend/src/shared/api/agent.js`
- Create: `frontend/src/user/components/AgentDirectorConversation.jsx`
- Create: `frontend/src/user/pages/agent-director-mode.test.js`

**Interfaces:**
- Produces routes `POST /director/propose`, `POST /director/plans/{id}/approve`, `POST /director/plans/{id}/reject`, and `POST /director/mode`.
- Mode values: `supervised` and `automatic`; stored with each execution record.

- [ ] **Step 1: Write failing mode and stop-condition tests**

```go
func TestDirectorAutomaticModeRequiresExplicitRequestAndStopsOnMissingModel(t *testing.T) {
    require.Equal(t, SupervisedMode, DetectDirectorMode("帮我做这本书"))
    require.Equal(t, AutomaticMode, DetectDirectorMode("全自动化完成这本书"))
    require.Equal(t, StopMissingConfiguration, EvaluateAutomaticPlan(planWithoutModel).Reason)
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/httpapi -run TestDirector -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-director-mode.test.js`

Expected: FAIL because mode persistence and approval routes are absent.

- [ ] **Step 3: Implement proposal, approval, and automatic stops**

```jsx
<Segmented
  value={mode}
  options={[{ label: '逐步确认', value: 'supervised' }, { label: '自动执行', value: 'automatic' }]}
  onChange={requestModeChange}
/>
```

Do not switch to automatic merely because the segmented control was rendered; persist the switch as an explicit user action. In supervised mode, render command cards with changes, cost/status, approve, reject, and feedback. For ambiguous requests, return one `clarification_required` response with one question and no commands.

- [ ] **Step 4: Run Director and conversation tests**

Run: `Set-Location backend; go test ./internal/httpapi ./internal/agentworkspace -run TestDirector -count=1; Set-Location ..; $directorTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-director-*.test.js').FullName; node --test $directorTests; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit Director execution**

```powershell
git add backend/internal/httpapi/agent_director.go backend/internal/httpapi/agent_director_test.go backend/internal/httpapi/router.go frontend/src/shared/api/agent.js frontend/src/user/components/AgentDirectorConversation.jsx frontend/src/user/pages/agent-director-mode.test.js
git commit -m "feat(agent): add supervised director execution"
```

### Task 5: Storyboard and Director browser acceptance

**Files:**
- Create: `test/agent-storyboard-director.source.test.js`
- Modify: `docs/superpowers/specs/2026-10-21-agent-canvas-liblib-workspace-design.md`

**Interfaces:**
- Consumes: complete Phase 4 feature set.
- Produces: authenticated local evidence that workflow, storyboard, and Director share one durable project.

- [ ] **Step 1: Add source boundary guards**

```js
test('storyboard reuses project assets and director cannot delete durable assets', () => {
  assert.match(read('frontend/src/user/pages/agent-storyboard-state.js'), /sourceId/);
  assert.match(read('backend/internal/agentworkspace/director_commands.go'), /delete_asset/);
});
```

- [ ] **Step 2: Run the full Phase 4 suite**

Run: `$storyboardTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-storyboard*.test.js').FullName; $directorTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-director-*.test.js').FullName; node --test $storyboardTests $directorTests test/agent-storyboard-director.source.test.js; npm --prefix frontend run build; Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run 'TestDirector|TestGeneration' -count=1; go build ./cmd/api`

Expected: all commands succeed.

- [ ] **Step 3: Rebuild/start Go runtime and verify real interactions**

Switch one project between workflow and storyboard; categorize and reorder materials; edit an image model setting from storyboard and verify the canvas node changes; attach a material to Director; ask for a plan, reject with feedback, approve the revision, and verify nodes/edges change. Request `全自动化`, verify the visible mode changes, then remove a required model and verify execution stops with a focused configuration message. Save/reload and confirm conversation, mode, material selection, ordering, and canvas revision persist.

- [ ] **Step 4: Verify responsive layout screenshots**

Capture desktop and mobile-width screenshots. Confirm all three desktop columns resize without overlay; on narrow width, material/inspector tabs remain usable and Director messages, approval buttons, model names, and settings text are not clipped.

- [ ] **Step 5: Commit acceptance guard**

```powershell
git add test/agent-storyboard-director.source.test.js docs/superpowers/specs/2026-10-21-agent-canvas-liblib-workspace-design.md
git commit -m "test(agent): verify storyboard director flow"
```
