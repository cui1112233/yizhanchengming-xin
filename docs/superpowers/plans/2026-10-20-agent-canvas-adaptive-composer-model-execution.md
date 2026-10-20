# V88 Agent Adaptive Composer and Model Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current heavy node editor with one adaptive composer and make text, image, and video model selections execute through the current account's real Go-backed model configuration.

**Architecture:** Keep React Flow as the canvas engine and extract the node composer into a focused React component. Add synchronous Agent text/image generation routes that validate the selected owner-scoped model, use existing text/image provider code, record the execution, and apply the result only when the target node version is unchanged. Preserve the existing asynchronous Agent video job lifecycle.

**Tech Stack:** Go 1.x, MySQL, TOS/local Agent object storage, React 18, Ant Design, React Flow, Lucide, Node test runner, Go embed.

## Global Constraints

- Work directly on the canonical `v88` branch; do not create another branch.
- Preserve all unrelated dirty-worktree changes; never reset, clean, or overwrite them.
- Backend code remains Go under `backend`; frontend remains React under `frontend` and is published only through Go embed.
- Model credentials remain server-side and come from the current account's unified model catalog.
- System prompts remain in `backend/internal/prompt`; React must not hardcode system prompts.
- Do not modify an applied migration. This plan does not require a schema change.
- Do not restore a Node release path, use GitHub Actions, push, or publish publicly.
- A successful build or HTTP response is not provider E2E proof. Browser acceptance must distinguish configured-model success from missing credentials or external-provider failure.

---

### Task 1: Add atomic Agent execution and node-result application

**Files:**
- Modify: `backend/internal/agentworkspace/store.go`
- Modify: `backend/internal/agentworkspace/types.go`
- Modify: `backend/internal/agentworkspace/memory_store.go`
- Modify: `backend/internal/agentworkspace/mysql_store.go`
- Test: `backend/internal/agentworkspace/store_test.go`
- Test: `backend/internal/agentworkspace/mysql_store_test.go`

**Interfaces:**
- Produces: `BeginNodeExecution(ctx, owner, projectID string, execution Execution) (Execution, error)`.
- Produces: `ApplyNodeExecution(ctx, owner, projectID string, input ApplyNodeExecutionInput) (Canvas, Execution, error)`.
- `ApplyNodeExecutionInput` contains `ExecutionID`, `NodeID`, `NodeVersion`, `NodeType`, `Summary`, `AssetRef`, `Detail`, and `Now`.
- Produces: `ErrExecutionExists` and `ErrExecutionNodeChanged`.

- [ ] **Step 1: Write failing memory-store tests**

```go
func TestMemoryStoreAppliesTextExecutionOnlyToExpectedNodeVersion(t *testing.T) {
    execution, err := store.BeginNodeExecution(ctx, "alice", "project-1", agentworkspace.Execution{
        ID: "execution-1", ToolName: "text-generation", Status: "running",
        IdempotencyKey: "text-node-1-v1", CreatedAt: now, UpdatedAt: now,
    })
    if err != nil { t.Fatal(err) }
    canvas, completed, err := store.ApplyNodeExecution(ctx, "alice", "project-1", agentworkspace.ApplyNodeExecutionInput{
        ExecutionID: execution.ID, NodeID: "node-1", NodeVersion: 1, NodeType: "text",
        Summary: "生成后的正文", Detail: map[string]any{"modelId":"text-primary"}, Now: now,
    })
    if err != nil || canvas.Nodes[0].Summary != "生成后的正文" || completed.Status != "completed" { t.Fatalf("canvas=%#v execution=%#v err=%v", canvas, completed, err) }
}
```

Add cases for duplicate idempotency, missing/wrong node type, stale node version, image `AssetRef`, and unchanged canvas on rejection.

- [ ] **Step 2: Run the focused tests and confirm the new API is missing**

Run: `cd backend; go test ./internal/agentworkspace -run 'TestMemoryStoreApplies|TestMySQLStoreApplies' -count=1`

Expected: FAIL because `BeginNodeExecution`, `ApplyNodeExecution`, and their input/errors do not exist.

- [ ] **Step 3: Implement the memory store and shared types**

```go
type ApplyNodeExecutionInput struct {
    ExecutionID string
    NodeID string
    NodeVersion int64
    NodeType string
    Summary string
    AssetRef map[string]any
    Detail map[string]any
    Now time.Time
}
```

`BeginNodeExecution` validates the project and inserts a `running` execution with owner-scoped idempotency. `ApplyNodeExecution` checks the exact node type/version, updates only `Summary` or `AssetRef`, increments node and canvas revisions, sets node status to `confirmed`, and marks the execution `completed` in the same lock.

- [ ] **Step 4: Implement the MySQL transaction**

Use `SELECT ... FOR UPDATE` on `agent_executions`, `agent_projects`, `agent_canvases`, and the target `agent_canvas_nodes`. Update the node and canvas revision and mark the execution completed in one transaction. Convert duplicate key `1062` into `ErrExecutionExists`.

- [ ] **Step 5: Run all Agent store tests**

Run: `cd backend; go test ./internal/agentworkspace -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 1 files only**

```powershell
git add -- backend/internal/agentworkspace/store.go backend/internal/agentworkspace/types.go backend/internal/agentworkspace/memory_store.go backend/internal/agentworkspace/mysql_store.go backend/internal/agentworkspace/store_test.go backend/internal/agentworkspace/mysql_store_test.go
git commit -m "feat: apply agent generation results atomically"
```

### Task 2: Resolve exact owner-scoped text models and generate text nodes

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/httpapi/router.go`
- Create: `backend/internal/httpapi/agent_node_generation.go`
- Create: `backend/internal/httpapi/agent_node_generation_test.go`
- Modify: `backend/internal/prompt/agent_workspace.go`
- Test: `backend/internal/app/text_provider_test.go`

**Interfaces:**
- Produces: `type AgentTextModelResolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)`.
- Produces: `POST /api/agent/tasks/{id}/text-generations`.
- Request: `{nodeId,nodeVersion,modelId,prompt,idempotencyKey}`.
- Response: `{canvas,execution,result}`.

- [ ] **Step 1: Write failing resolver and route tests**

```go
func TestAgentTextGenerationUsesRequestedOwnerModelAndAppliesResult(t *testing.T) {
    resolver := func(_ context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
        if owner != "alice" || modelID != "text-secondary" { t.Fatalf("owner=%s model=%s", owner, modelID) }
        return fixedAgentTextProvider{result: "生成后的正文"}, nil
    }
    // POST a version-1 text node and assert response canvas has summary and version 2.
}
```

Add tests for a video node, empty prompt, disabled/unknown model, duplicate idempotency, provider error, and stale node version.

- [ ] **Step 2: Run tests and confirm the route/resolver is missing**

Run: `cd backend; go test ./internal/httpapi ./internal/app -run 'TestAgentTextGeneration|TestConfiguredTextProvider' -count=1`

Expected: FAIL because the exact-model Agent resolver and route do not exist.

- [ ] **Step 3: Expose exact text model resolution from app assembly**

Reuse `configuredTextProvider(settings, selected)` with the requested catalog ID. Do not fall back to the account default when a non-empty requested ID is unknown or invalid. Pass the resolver through `RouterOptions` and into the Agent route registration.

- [ ] **Step 4: Implement the Go text-generation route**

Validate project/node/version before calling the provider. Begin the execution, call the requested text provider with the Agent text-node system prompt from `backend/internal/prompt`, then call `ApplyNodeExecution`. On provider failure, persist a failed execution with a readable detail and return `502` without changing the node.

- [ ] **Step 5: Run focused and application assembly tests**

Run: `cd backend; go test ./internal/httpapi ./internal/app -run 'TestAgentTextGeneration|TestConfiguredTextProvider|TestApplicationRegistersAgent' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 2 files only**

```powershell
git add -- backend/internal/app/app.go backend/internal/httpapi/router.go backend/internal/httpapi/agent_node_generation.go backend/internal/httpapi/agent_node_generation_test.go backend/internal/prompt
git commit -m "feat: generate agent text nodes with configured models"
```

### Task 3: Generate, ingest, and apply image-node results

**Files:**
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Modify: `backend/internal/httpapi/agent_node_generation_test.go`
- Modify: `backend/internal/httpapi/agent_video_asset_ingest.go`
- Modify: `backend/internal/httpapi/agent_video_asset_ingest_test.go`
- Modify: `backend/internal/httpapi/router.go`

**Interfaces:**
- Produces: `POST /api/agent/tasks/{id}/image-generations`.
- Request: `{nodeId,nodeVersion,modelId,prompt,size,referenceAssetIds,idempotencyKey}`.
- Response: `{canvas,execution,resultAsset}`.
- Produces a media ingestor that accepts `kind=image|video`, validates HTTPS/data results, enforces content type and size, writes the object before `agent_assets`, and deletes the object if database persistence fails.

- [ ] **Step 1: Write failing image route and media-ingest tests**

```go
func TestAgentImageGenerationPersistsAssetBeforeApplyingNode(t *testing.T) {
    // Configure image-demo, return a valid PNG data URL, submit version-1 image node.
    // Assert one image asset exists and the returned node references that asset at version 2.
}
```

Add cases for wrong model kind, disabled/unknown model, invalid result media, oversized payload, cross-project reference IDs, object-store failure, asset-store failure rollback, and stale node version after provider completion.

- [ ] **Step 2: Run focused tests and confirm the route is missing**

Run: `cd backend; go test ./internal/httpapi -run 'TestAgentImageGeneration|TestAgentMediaAssetIngest' -count=1`

Expected: FAIL because the Agent image route and generalized media ingestor do not exist.

- [ ] **Step 3: Generalize the existing safe result ingestor**

Retain the current private-IP, unsafe redirect, timeout, media sniffing, size-limit, object-key, and rollback protections. Existing video tests must continue to pass unchanged.

- [ ] **Step 4: Implement the Agent image-generation route**

Validate the selected model by invoking the existing `configuredBatchImage` path with the requested catalog ID. Convert the returned URL/data URL into an Agent image asset, then atomically apply the asset with `ApplyNodeExecution`. Do not place provider URLs directly in `AssetRef`.

- [ ] **Step 5: Run image, video-ingest, and configured-image regression tests**

Run: `cd backend; go test ./internal/httpapi -run 'TestAgentImageGeneration|TestAgentMediaAssetIngest|TestAgentVideoAssetIngest|TestConfiguredImage|TestBatchFactoryV12' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 3 files only**

```powershell
git add -- backend/internal/httpapi/agent_node_generation.go backend/internal/httpapi/agent_node_generation_test.go backend/internal/httpapi/agent_video_asset_ingest.go backend/internal/httpapi/agent_video_asset_ingest_test.go backend/internal/httpapi/router.go
git commit -m "feat: generate and store agent image results"
```

### Task 4: Add frontend generation clients and pure composer state

**Files:**
- Modify: `frontend/src/shared/api/agent.js`
- Create: `frontend/src/user/pages/agent-canvas-node-generation.js`
- Create: `frontend/src/user/pages/agent-canvas-node-generation.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Produces: `createAgentTextGeneration(taskId, payload)` and `createAgentImageGeneration(taskId, payload)`.
- Produces: `nodeGenerationViewState({kind,busy,error,execution})`.
- Produces: `composerModelSelection(kind, config, options)` and `composerShouldClose(eventState)`.

- [ ] **Step 1: Write failing pure-function and API-contract tests**

```js
test('text and image generation clients use Agent project routes', async () => {
  await createAgentTextGeneration('project-1', payload, request)
  assert.equal(request.calls[0].path, '/api/agent/tasks/project-1/text-generations')
  await createAgentImageGeneration('project-1', payload, request)
  assert.equal(request.calls[1].path, '/api/agent/tasks/project-1/image-generations')
})
```

Cover preserved model selection, first valid model fallback, no-model disabled state, busy state, readable failure, and idempotency-key changes when node version changes.

- [ ] **Step 2: Run tests and confirm the module/functions are missing**

Run: `node --test frontend/src/user/pages/agent-canvas-node-generation.test.js`

Expected: FAIL because the module does not exist.

- [ ] **Step 3: Implement API clients and pure helpers**

Keep model IDs unchanged between UI and API. Do not map display names back to model IDs. Derive button availability solely from prompt, selected valid model, node kind, and busy state.

- [ ] **Step 4: Run the focused frontend tests**

Run: `node --test frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: PASS.

- [ ] **Step 5: Commit Task 4 files only**

```powershell
git add -- frontend/src/shared/api/agent.js frontend/src/user/pages/agent-canvas-node-generation.js frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "feat: add agent node generation clients"
```

### Task 5: Build the shared adaptive composer and connect all node types

**Files:**
- Create: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Create: `frontend/src/user/pages/agent-node-composer.source.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- `AgentNodeComposer` consumes `node`, `modelsByKind`, `generationState`, reference callbacks, config callback, and text/image/video submit/cancel/retry/confirm callbacks.
- `AgentCanvas` owns one `activeEditorNodeID` and one child-popover state. Closing the editor also clears `assetPanel` and composer popovers.
- Text/image/video submit handlers call their matching real API and replace local canvas with the server-returned canvas through `onServerCanvasChange`.

- [ ] **Step 1: Write failing source and interaction assertions**

Assert that all three node kinds render `AgentNodeComposer`, the canvas loads `text`, `image`, and `video` models, a pane click/Escape clears both the active editor and asset panel, and text/image submit handlers call real Agent APIs.

- [ ] **Step 2: Run focused tests and confirm the composer is missing**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: FAIL because `AgentNodeComposer` and the new handlers do not exist.

- [ ] **Step 3: Extract and implement `AgentNodeComposer`**

Use one 680px responsive shell. Render the capability row only for implemented controls, numbered references above the prompt, a borderless autosizing prompt, and one-line bottom toolbar. Model/mode/parameter panels anchor above the toolbar; advanced settings expand within the shell. Long model names remain readable in the opened selector.

- [ ] **Step 4: Connect model lists and real generation handlers**

Load all three model kinds together. Text success applies the returned canvas and keeps the generated summary. Image success applies the returned asset/canvas. Video keeps the existing job state and confirmation behavior. Missing models disable submission with an API configuration link.

- [ ] **Step 5: Fix unified close and positioning behavior**

Pane click, Escape, project change, node deletion, and switching nodes must close every composer-owned surface. Clicking the composer or its Ant Design portal must not trigger the canvas close handler. Opening near an edge calls the existing React Flow viewport controls to keep the editor visible.

- [ ] **Step 6: Apply scoped visual styling**

Keep V88 variables and palette. Remove the fixed 800px panel rules and wrapped toolbar rules. Do not introduce a new theme, gradients, decorative cards, or nested cards.

- [ ] **Step 7: Run Agent frontend tests and build**

Run: `node --test frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-history.test.js frontend/src/user/pages/agent-canvas-video-jobs.test.js frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/pages/agent-node-composer.source.test.js`

Run: `npm --prefix frontend run build`

Expected: PASS.

- [ ] **Step 8: Commit Task 5 files only**

```powershell
git add -- frontend/src/user/components/AgentNodeComposer.jsx frontend/src/user/components/AgentCanvas.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "feat: unify agent node generation composer"
```

### Task 6: Rebuild Go embed and perform real browser acceptance

**Files:**
- Modify only files whose failing checks reveal a task-scoped defect.

**Interfaces:**
- Consumes all prior backend routes, frontend APIs, and composer behavior.
- Produces a verified local Go runtime at `http://localhost:4001`.

- [ ] **Step 1: Run the complete relevant backend suites**

Run: `cd backend; go test ./internal/agentworkspace ./internal/httpapi ./internal/app ./internal/storage ./cmd/qiantie -count=1`

Expected: PASS.

- [ ] **Step 2: Build and restart the approved local Go runtime**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`

Run: `powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1`

Expected: the embedded runtime reports the new Git SHA and local health checks pass.

- [ ] **Step 3: Verify composer interaction in the authenticated browser**

Open the existing Agent project. Verify image, video, and text previews each open the correct composer; switching nodes closes the previous composer; pane click and Escape close the composer plus model/settings/material popovers; model labels are fully readable; reference selection and automatic edge references remain usable.

- [ ] **Step 4: Verify real text, image, and video execution**

Use one enabled configured model for each kind. Generate text and confirm the current text node updates and survives reload. Generate an image and confirm it appears in project assets and the image node after reload. Submit a video and confirm the existing job can recover after reload and reach its supported confirmation/apply state.

If a kind lacks valid credentials, pairing, quota, or provider availability, capture the exact visible/API error and report that kind as not provider-E2E verified; do not call it successful.

- [ ] **Step 5: Check final Git scope and commit any test-discovered fixes**

Run: `git diff --check`

Run: `git status --short`

Stage and commit only files changed for this feature. Do not push or deploy publicly.

## Plan Self-Review

- Spec coverage: Tasks 1-3 cover owner-scoped model validation, atomic result application, text generation, image ingestion, execution records, errors, and persistence. Tasks 4-5 cover the shared adaptive composer, all three model lists, unified close behavior, references, and real submit handlers. Task 6 covers Go embed and provider E2E boundaries.
- Placeholder scan: no TBD/TODO or unspecified implementation steps remain.
- Type consistency: `BeginNodeExecution`, `ApplyNodeExecution`, `ApplyNodeExecutionInput`, `AgentTextModelResolver`, frontend generation clients, and `AgentNodeComposer` are introduced before their consumers.
- Scope: audio generation and unrelated Agent chat changes are excluded. Existing video jobs remain the video source of truth.
