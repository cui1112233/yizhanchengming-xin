# Agent Canvas Phase 3: Generation Lifecycle, History, and Audio Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give every node kind one durable generation lifecycle, multiple outputs, project generation history, and first-class audio generation and processing.

**Architecture:** Add a generic generation task/result-set domain beside the existing video jobs, then route text/image/video/audio through the same visible states. Keep provider adapters behind Go interfaces, persist assets before completion, and expose account-scoped history APIs consumed by a workspace history drawer.

**Tech Stack:** Go, MySQL, Redis task mechanism, TOS/local object storage abstraction, React 18, Ant Design 5, Node and Go tests.

## Global Constraints

- Valid states are `draft`, `queued`, `running`, `awaiting_confirmation`, `completed`, `failed`, and `cancelled`.
- A result is never marked complete until its project asset is durable.
- Every task carries owner, project ID, node ID, node version, provider/model snapshot, settings snapshot, reference IDs, and idempotency key.
- Supervised mode requires confirmation; explicit automatic mode may apply a valid result automatically.
- A stale node version never overwrites newer node edits.
- Multiple outputs remain a result set; selecting one does not silently delete the others.
- Audio processing creates a new asset version and never overwrites source bytes.
- MySQL migration changes are append-only and use the next unused migration number after confirming the live repository state.

---

## File Map

- `backend/internal/agentworkspace/generation_tasks.go`: task states, result sets, and store interface.
- `backend/internal/agentworkspace/mysql_generation_tasks.go`: MySQL implementation.
- `backend/internal/httpapi/agent_generation_tasks.go`: submit/list/get/confirm/cancel/retry endpoints.
- `backend/internal/agentworkspace/audio_jobs.go`: audio generation and processing contracts.
- `backend/internal/httpapi/agent_audio_jobs.go`: audio endpoints.
- `frontend/src/shared/api/agent.js`: generic task/history/audio client.
- `frontend/src/user/pages/agent-generation-state.js`: task-to-node presentation.
- `frontend/src/user/components/AgentGenerationHistory.jsx`: filters and asset actions.

### Task 1: Generic task and result-set persistence

**Files:**
- Create: `backend/internal/agentworkspace/generation_tasks.go`
- Create: `backend/internal/agentworkspace/generation_tasks_test.go`
- Create: `backend/internal/agentworkspace/mysql_generation_tasks.go`
- Create: `backend/internal/agentworkspace/mysql_generation_tasks_test.go`
- Modify: `backend/internal/storage/agent_workspace_schema.go`
- Modify: `backend/internal/storage/agent_workspace_schema_test.go`

**Interfaces:**
- Produces: `GenerationTask`, `GenerationResult`, `CreateGenerationTaskInput`, and `GenerationTaskStore` methods `Create`, `Get`, `List`, `Transition`, `Confirm`, `Cancel`, and `Retry`.
- Unique idempotency scope: `(owner_username, project_id, idempotency_key)`.

- [ ] **Step 1: Confirm migration number and write failing state tests**

Run: `rg -n "Version: 880000" backend/internal/storage/agent_workspace_schema.go`

Expected: `8800007` is the latest Agent workspace migration and `8800008` is unused. If it is already used, replace `8800008` in this plan with the next unused version before editing.

```go
func TestGenerationTaskTransitionsAndVersionConflict(t *testing.T) {
    store := NewMemoryStore()
    task, err := store.CreateGenerationTask(ctx, "alice", input)
    require.NoError(t, err)
    require.ErrorIs(t, store.TransitionGenerationTask(ctx, "alice", task.ID, GenerationCompleted), ErrGenerationState)
    require.NoError(t, store.TransitionGenerationTask(ctx, "alice", task.ID, GenerationRunning))
    require.ErrorIs(t, store.ConfirmGenerationTask(ctx, "alice", task.ProjectID, task.ID), ErrGenerationNodeChanged)
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/storage -run 'TestGenerationTask|TestAgentWorkspaceSchema.*Generation' -count=1`

Expected: FAIL because the generic task model and tables do not exist.

- [ ] **Step 3: Implement the state machine, migration, and MySQL store**

```go
type GenerationTask struct {
    ID, ProjectID, OwnerUsername, NodeID, Kind, Provider, ModelID string
    NodeVersion int64
    State GenerationState
    Settings map[string]any
    ReferenceAssetIDs []string
    Results []GenerationResult
    IdempotencyKey, ErrorMessage string
    CreatedAt, UpdatedAt time.Time
}
```

Append migration version `8800008` to `AgentWorkspaceMigrations()` with new task/result tables and a new callback checksum. The transaction that confirms a result locks task, node, and selected asset rows, verifies ownership and node version, updates the node asset reference, and then marks the task completed.

- [ ] **Step 4: Run store and schema tests**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/storage -run 'TestGenerationTask|TestMySQLGeneration|TestAgentWorkspaceSchema' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit persistence**

```powershell
git add backend/internal/agentworkspace/generation_tasks.go backend/internal/agentworkspace/generation_tasks_test.go backend/internal/agentworkspace/mysql_generation_tasks.go backend/internal/agentworkspace/mysql_generation_tasks_test.go backend/internal/storage/agent_workspace_schema.go backend/internal/storage/agent_workspace_schema_test.go
git commit -m "feat(agent): persist generation task results"
```

### Task 2: Unified task HTTP lifecycle

**Files:**
- Create: `backend/internal/httpapi/agent_generation_tasks.go`
- Create: `backend/internal/httpapi/agent_generation_tasks_test.go`
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Modify: `backend/internal/httpapi/agent_video_jobs.go`
- Modify: `backend/internal/httpapi/router.go`

**Interfaces:**
- Produces: `POST /api/agent/tasks/{projectId}/generations`, `GET /generations`, `GET /generations/{id}`, and action routes `/confirm`, `/cancel`, `/retry`.
- Existing video job routes remain compatible adapters until all callers move to generic routes.

- [ ] **Step 1: Write failing owner, idempotency, and confirmation tests**

```go
func TestAgentGenerationLifecycleIsOwnerScopedAndIdempotent(t *testing.T) {
    first := postGeneration(t, alice, requestWithKey("same-key"))
    second := postGeneration(t, alice, requestWithKey("same-key"))
    require.Equal(t, decodeID(first), decodeID(second))
    require.Equal(t, http.StatusNotFound, getGeneration(t, bob, decodeID(first)).Code)
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/httpapi -run 'TestAgentGeneration' -count=1`

Expected: FAIL because generic routes are absent.

- [ ] **Step 3: Implement routes and compatibility adapters**

```go
type createGenerationRequest struct {
    NodeID string `json:"nodeId"`
    NodeVersion int64 `json:"nodeVersion"`
    Kind string `json:"kind"`
    Provider string `json:"provider"`
    ModelID string `json:"modelId"`
    Mode string `json:"mode"`
    Parameters map[string]any `json:"parameters"`
    Advanced map[string]any `json:"advanced"`
    ReferenceAssetIDs []string `json:"referenceAssetIds"`
    IdempotencyKey string `json:"idempotencyKey"`
}
```

Validate owner/project/node/model before task creation. Provider failure transitions to `failed`; cancellation is accepted only for queued/running tasks; retry creates a new task linked to the prior one.

- [ ] **Step 4: Run HTTP and existing video tests**

Run: `Set-Location backend; go test ./internal/httpapi -run 'TestAgent(Generation|Video)' -count=1`

Expected: PASS, including old video route coverage.

- [ ] **Step 5: Commit the unified routes**

```powershell
git add backend/internal/httpapi/agent_generation_tasks.go backend/internal/httpapi/agent_generation_tasks_test.go backend/internal/httpapi/agent_node_generation.go backend/internal/httpapi/agent_video_jobs.go backend/internal/httpapi/router.go
git commit -m "feat(agent): unify generation lifecycle api"
```

### Task 3: Frontend task state and multiple-output confirmation

**Files:**
- Create: `frontend/src/user/pages/agent-generation-state.js`
- Create: `frontend/src/user/pages/agent-generation-state.test.js`
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`

**Interfaces:**
- Produces: `generationPresentation(task)`, `mergeGenerationTask(node, task)`, and API helpers `createAgentGeneration`, `listAgentGenerations`, `confirmAgentGeneration`, `cancelAgentGeneration`, `retryAgentGeneration`.

- [ ] **Step 1: Write failing state and result selection tests**

```js
test('awaiting confirmation keeps all outputs and the selected candidate', () => {
  const state = mergeGenerationTask(node, { state: 'awaiting_confirmation', results, selectedResultId: 'r2' });
  assert.equal(state.previewAssetId, 'r2');
  assert.equal(state.resultOptions.length, 4);
  assert.equal(generationPresentation(state.task).primaryAction, '确认使用');
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-generation-state.test.js frontend/src/shared/api/agent-video-jobs.test.js`

Expected: FAIL because generic generation state helpers do not exist.

- [ ] **Step 3: Implement polling and result selection**

```js
export function generationPresentation(task) {
  const map = {
    queued: ['排队中', 'cancel'], running: ['生成中', 'cancel'],
    awaiting_confirmation: ['等待确认', 'confirm'], failed: ['生成失败', 'retry'],
    completed: ['已完成', 'none'], cancelled: ['已取消', 'retry'],
  };
  const [label, action] = map[task?.state] || ['未生成', 'generate'];
  return { label, primaryAction: action };
}
```

Poll only nonterminal tasks, stop on unmount/task switch, and keep output selection local until confirmation succeeds.

- [ ] **Step 4: Run frontend tests and build**

Run: `node --test frontend/src/user/pages/agent-generation-state.test.js frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/shared/api/agent-video-jobs.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit the visible lifecycle**

```powershell
git add frontend/src/user/pages/agent-generation-state.js frontend/src/user/pages/agent-generation-state.test.js frontend/src/shared/api/agent.js frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/AgentNodeComposer.jsx
git commit -m "feat(agent): show unified generation lifecycle"
```

### Task 4: Audio generation and nondestructive processing

**Files:**
- Create: `backend/internal/agentworkspace/audio_jobs.go`
- Create: `backend/internal/agentworkspace/audio_jobs_test.go`
- Create: `backend/internal/httpapi/agent_audio_jobs.go`
- Create: `backend/internal/httpapi/agent_audio_jobs_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`

**Interfaces:**
- Produces audio operations `speech`, `music`, `sound_effect`, `trim`, `speed`, `smart_split`, and `custom_split`.
- Processing request references `sourceAssetId` plus operation settings and returns a generic generation task.

- [ ] **Step 1: Write failing source-preservation and owner tests**

```go
func TestAudioTrimCreatesNewAssetWithoutChangingSource(t *testing.T) {
    source := createAudioAsset(t, "source.wav")
    result := runAudioProcess(t, source.ID, map[string]any{"operation":"trim","startMs":1000,"endMs":3000})
    require.NotEqual(t, source.ID, result.AssetID)
    require.Equal(t, source.ObjectKey, getAsset(t, source.ID).ObjectKey)
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run 'TestAudio' -count=1`

Expected: FAIL because audio jobs and routes are absent.

- [ ] **Step 3: Implement audio adapters and UI controls**

```go
type AudioProcessRequest struct {
    SourceAssetID string `json:"sourceAssetId"`
    Operation string `json:"operation"`
    Settings map[string]any `json:"settings"`
}
```

Validate source ownership/content type/ranges before queueing. Persist produced bytes as a new project asset, then attach that asset as a task result. The composer shows only operations supported by the selected audio model or installed local processor.

- [ ] **Step 4: Run audio, task, and build checks**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run 'Test(Audio|AgentGeneration)' -count=1; Set-Location ..; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit audio support**

```powershell
git add backend/internal/agentworkspace/audio_jobs.go backend/internal/agentworkspace/audio_jobs_test.go backend/internal/httpapi/agent_audio_jobs.go backend/internal/httpapi/agent_audio_jobs_test.go backend/internal/httpapi/router.go frontend/src/shared/api/agent.js frontend/src/user/components/AgentNodeComposer.jsx
git commit -m "feat(agent): add audio generation and processing"
```

### Task 5: Generation history and asset actions

**Files:**
- Create: `frontend/src/user/pages/agent-generation-history.js`
- Create: `frontend/src/user/pages/agent-generation-history.test.js`
- Create: `frontend/src/user/components/AgentGenerationHistory.jsx`
- Create: `frontend/src/user/pages/agent-generation-history.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: generic generation list API and project assets.
- Produces filters `{ scope, projectId, kinds, states, ratings, query, sort }` and actions add-to-canvas, save-to-assets, retry, download, and confirmed delete.

- [ ] **Step 1: Write failing filter and action-permission tests**

```js
test('history filters current project and exposes only valid actions', () => {
  const rows = filterGenerationHistory(history, { scope: 'project', projectId: 'p1', kinds: ['image'], states: ['completed'] });
  assert.deepEqual(rows.map(row => row.id), ['g2']);
  assert.deepEqual(historyActions(rows[0]), ['add_to_canvas', 'save_asset', 'download', 'delete']);
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-generation-history.test.js frontend/src/user/pages/agent-generation-history.source.test.js`

Expected: FAIL because history helpers and UI are absent.

- [ ] **Step 3: Implement history UI and explicit deletion**

```jsx
<AgentGenerationHistory
  taskId={activeTask.id}
  filters={filters}
  onAddToCanvas={createNodeFromHistoryResult}
  onRetry={retryAgentGeneration}
/>
```

Deletion always opens a confirmation naming the result count; deleting a history record does not delete a reusable project asset unless that asset is selected in a separate asset deletion action.

- [ ] **Step 4: Run all Phase 3 checks**

Run: `$generationTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-generation-*.test.js').FullName; node --test $generationTests frontend/src/user/pages/agent-canvas-node-generation.test.js; npm --prefix frontend run build; Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi ./internal/storage -count=1; go build ./cmd/api`

Expected: all commands succeed.

- [ ] **Step 5: Browser acceptance and commit**

In the rebuilt authenticated local Go runtime, generate a two-image result set, select and confirm one result, retry one forced failure, cancel a queued task, produce and trim audio, add a history result to the canvas, save/reload, and verify every task/result/reference remains attached to the correct project and node.

```powershell
git add frontend/src/user/pages/agent-generation-history.js frontend/src/user/pages/agent-generation-history.test.js frontend/src/user/components/AgentGenerationHistory.jsx frontend/src/user/pages/agent-generation-history.source.test.js frontend/src/user/pages/AgentPageV2.jsx frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): add generation history workspace"
```
