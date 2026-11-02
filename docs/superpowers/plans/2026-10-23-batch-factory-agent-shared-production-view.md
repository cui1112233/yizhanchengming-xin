# 批量工厂与 Agent 共用生产链 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (\`- [ ]\`) syntax for tracking.

**Goal:** 让批量工厂和 Agent 画布共用一份作品、书、素材、版本、合成和视频管理系统发布记录。

**Architecture:** 批量工厂继续是书、素材、分镜、生产、合成和发布的真实来源；Agent 保存作品绑定、画布位置、分组、连线和编辑界面状态。投影服务把批量工厂书籍映射为 Agent 现有的五种节点，并让 Agent 对映射节点调用原有批量工厂记录，不复制媒体文件或创建第二条任务链。

**Tech Stack:** Go、MySQL、Redis、React、Ant Design、React Flow、Go embed。

## Global Constraints

- 维护分支为 \`v88\`；不得新建分支或带入现有未提交改动。
- 节点类型只能是 \`script\`、\`text\`、\`image\`、\`video\`、\`audio\`。
- 人物、场景、道具、分镜和合成视频是用途标签，不是新节点类型。
- 媒体、生产、合成和发布继续走批量工厂已有记录；画布节点不保存媒体字节。
- 默认删除只删除画布引用和连线；永久删除继续走批量工厂明确确认的删除能力。
- 数据库变更采用追加式 \`AgentWorkspaceMigrations()\` migration，不能改已应用的 migration。
- 最终必须验证 Go 内嵌前端后的 \`http://localhost:4001\`，不能只验源码或开发服务。

## File Structure

- \`backend/internal/storage/agent_workspace_schema.go\`：作品绑定与节点来源映射表。
- \`backend/internal/agentworkspace/batch_factory_binding.go\`：绑定类型和存储接口。
- \`backend/internal/agentworkspace/batch_factory_projection.go\`：书籍到节点/连线的投影与时长计算。
- \`backend/internal/agentworkspace/mysql_batch_factory_binding.go\`：MySQL 映射存储。
- \`backend/internal/httpapi/agent_batch_factory_bridge.go\`：绑定、同步、删除画布引用的 API。
- \`frontend/src/shared/api/agent.js\`、\`frontend/src/shared/api/batchFactoryV11.js\`：桥接 API 客户端。
- \`frontend/src/user/pages/AgentPageV2.jsx\`、\`frontend/src/user/components/AgentCanvas.jsx\`：书籍分组定位和展示。
- \`frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx\`：从单本书进入 Agent。
- \`frontend/src/user/pages/agent-canvas-batch-factory.js\`：前端纯数据适配和时长/状态呈现。

### Task 1: 持久化作品和节点来源绑定

**Files:**
- Modify: \`backend/internal/storage/agent_workspace_schema.go\`
- Modify: \`backend/internal/storage/agent_workspace_schema_test.go\`
- Modify: \`backend/internal/agentworkspace/store.go\`
- Create: \`backend/internal/agentworkspace/batch_factory_binding.go\`
- Create: \`backend/internal/agentworkspace/mysql_batch_factory_binding.go\`
- Create: \`backend/internal/agentworkspace/memory_batch_factory_binding.go\`
- Test: \`backend/internal/agentworkspace/batch_factory_binding_test.go\`
- Test: \`backend/internal/agentworkspace/mysql_batch_factory_binding_test.go\`

**Interfaces:** Consumes batch, book, Agent project and node IDs. Produces an owner-scoped project binding and source-origin rows for later projection.

- [ ] **Step 1: Write the failing binding tests**

~~~go
func TestValidateBatchFactoryBindingRejectsEmptyIDs(t *testing.T) {
  assert.Error(t, ValidateBatchFactoryBinding(BatchFactoryBinding{Owner: "alice", ProjectID: "agent-1"}))
  assert.NoError(t, ValidateBatchFactoryBinding(BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: "agent-1"}))
}

func TestNodeOriginNeverContainsMediaBytes(t *testing.T) {
  origin := BatchFactoryNodeOrigin{BatchID: "batch-1", BookID: "book-1", ResourceType: "book_asset", ResourceID: "asset-1"}
  assert.NoError(t, ValidateBatchFactoryNodeOrigin(origin))
}
~~~

- [ ] **Step 2: Run test to verify it fails**

Run: \`go test ./backend/internal/agentworkspace -run 'TestValidateBatchFactoryBinding|TestNodeOriginNeverContainsMediaBytes' -count=1\`

Expected: FAIL because the binding contract does not exist.

- [ ] **Step 3: Implement the minimum binding contract and migration**

~~~go
type BatchFactoryBinding struct {
  Owner string
  BatchID string
  ProjectID string
  CreatedAt time.Time
  UpdatedAt time.Time
}

type BatchFactoryNodeOrigin struct {
  ProjectID string
  NodeID string
  BatchID string
  BookID string
  ResourceType string // book_script, book_asset, audio, video, merge
  ResourceID string
  VersionID string
}

type BatchFactoryBindingStore interface {
  BindBatchFactoryProject(context.Context, BatchFactoryBinding) (BatchFactoryBinding, error)
  BatchFactoryBinding(context.Context, string, string) (BatchFactoryBinding, error)
  BindBatchFactoryNode(context.Context, string, BatchFactoryNodeOrigin) error
  BatchFactoryNodeOrigins(context.Context, string, string, string) ([]BatchFactoryNodeOrigin, error)
  RemoveBatchFactoryNodeBinding(context.Context, string, string, string) error
}
~~~

Append one migration after \`8800011\`: create \`agent_batch_factory_bindings\` with unique owner/batch and owner/project keys; create \`agent_batch_factory_node_origins\` with primary project/node and owner/book index. Neither table may contain a media byte, object key or provider credential column. Implement it for MySQL and memory stores.

- [ ] **Step 4: Run storage tests**

Run: \`go test ./backend/internal/agentworkspace ./backend/internal/storage -run 'Test.*BatchFactoryBinding|TestMySQL84AgentWorkspace|TestAgentWorkspace' -count=1\`

Expected: PASS; bindings are account-isolated and append-only.

- [ ] **Step 5: Commit**

~~~powershell
git add backend/internal/storage/agent_workspace_schema.go backend/internal/storage/agent_workspace_schema_test.go backend/internal/agentworkspace/store.go backend/internal/agentworkspace/batch_factory_binding.go backend/internal/agentworkspace/mysql_batch_factory_binding.go backend/internal/agentworkspace/memory_batch_factory_binding.go backend/internal/agentworkspace/batch_factory_binding_test.go backend/internal/agentworkspace/mysql_batch_factory_binding_test.go
git commit -m "feat(agent): bind batch factory projects to canvas"
~~~

### Task 2: 投影批量工厂书籍为五种既有节点

**Files:**
- Create: \`backend/internal/agentworkspace/batch_factory_projection.go\`
- Create: \`backend/internal/agentworkspace/batch_factory_projection_test.go\`
- Modify: \`backend/internal/agentworkspace/mysql_store.go\`
- Modify: \`backend/internal/agentworkspace/memory_store.go\`

**Interfaces:** Consumes Task 1 bindings and \`batchfactoryv11.Store\` book, asset, video, production and merge records. Produces \`SyncBatchFactoryBook(ctx, owner, binding, bookID) (Canvas, error)\`.

- [ ] **Step 1: Write the failing projection test**

~~~go
func TestSyncBatchFactoryBookProjectsOnlyExistingNodeTypes(t *testing.T) {
  canvas, err := service.SyncBatchFactoryBook(ctx, "alice", binding, "book-1")
  require.NoError(t, err)
  assertNodeTypes(t, canvas.Nodes, "script", "image", "audio", "video")
  assert.All(t, canvas.Nodes, func(node Node) bool { return node.GroupID == "book:book-1" })
  assert.NotContains(t, nodeTypes(canvas.Nodes), "asset")
}
~~~

- [ ] **Step 2: Run test to verify it fails**

Run: \`go test ./backend/internal/agentworkspace -run TestSyncBatchFactoryBookProjectsOnlyExistingNodeTypes -count=1\`

Expected: FAIL because the projection service does not exist.

- [ ] **Step 3: Implement deterministic reconciliation**

~~~go
func (s *BatchFactoryProjectionService) SyncBatchFactoryBook(
  ctx context.Context, owner string, binding BatchFactoryBinding, bookID string,
) (Canvas, error)
~~~

Resolve the bound book and upsert stable source-bound nodes: one script node for source text; one image node per \`BookAsset\`, with prompt and primary image; one audio node only for a persisted audio record; one video node per \`Book.Videos\`; one video node named “最终合成视频” for a persisted merge. All use \`GroupID = "book:<bookID>"\`. Create edges only from stored source references. Preserve manual text nodes, positions and unrelated edges. Save origin rows without copying TOS or executor bytes into \`agent_assets\`.

- [ ] **Step 4: Run projection regressions**

Run: \`go test ./backend/internal/agentworkspace -run 'TestSyncBatchFactoryBook|Test.*Canvas|Test.*Node' -count=1\`

Expected: PASS; repeated synchronization keeps IDs stable and creates no duplicate nodes.

- [ ] **Step 5: Commit**

~~~powershell
git add backend/internal/agentworkspace/batch_factory_projection.go backend/internal/agentworkspace/batch_factory_projection_test.go backend/internal/agentworkspace/mysql_store.go backend/internal/agentworkspace/memory_store.go
git commit -m "feat(agent): project batch factory books onto canvas"
~~~

### Task 3: 新增绑定、同步和安全删除 API

**Files:**
- Create: \`backend/internal/httpapi/agent_batch_factory_bridge.go\`
- Create: \`backend/internal/httpapi/agent_batch_factory_bridge_test.go\`
- Modify: \`backend/internal/app/app.go\`
- Modify: \`backend/internal/httpapi/agent_workspace_compat.go\`

**Interfaces:** Produces:
- \`POST /api/batch-factory/v11/batches/{batchId}/agent-project\`
- \`GET /api/agent/tasks/{id}/batch-factory\`
- \`POST /api/agent/tasks/{id}/batch-factory/books/{bookId}/sync\`
- \`DELETE /api/agent/tasks/{id}/batch-factory/nodes/{nodeId}\`

- [ ] **Step 1: Write the failing HTTP tests**

~~~go
func TestBindBatchFactoryReturnsSameProjectForSameOwner(t *testing.T) { /* POST twice; project ID matches */ }
func TestSyncBookRejectsAnotherOwnersBatch(t *testing.T) { /* expect 404; do not leak title */ }
func TestDeleteBoundNodeOnlyRemovesCanvasReference(t *testing.T) { /* DeleteBookAsset must never be called */ }
~~~

- [ ] **Step 2: Run test to verify it fails**

Run: \`go test ./backend/internal/httpapi -run 'TestBindBatchFactory|TestSyncBookRejects|TestDeleteBoundNode' -count=1\`

Expected: FAIL because bridge routes are absent.

- [ ] **Step 3: Implement authenticated bridge handlers**

Use current \`BridgeAuth\`. Binding creates one Agent project only once for the owner. Synchronizing validates project, batch and book ownership before projection. Default deletion removes only origin row, canvas node and related canvas edges; it does not call batch-factory asset, generation, merge or publication deletion. Responses include safe source, version and state fields only.

- [ ] **Step 4: Run API and app wiring tests**

Run: \`go test ./backend/internal/httpapi ./backend/internal/app -run 'Test.*BatchFactory.*Agent|TestAgentWorkspace|TestApp' -count=1\`

Expected: PASS; repeated binds are idempotent and default deletion cannot erase real media.

- [ ] **Step 5: Commit**

~~~powershell
git add backend/internal/httpapi/agent_batch_factory_bridge.go backend/internal/httpapi/agent_batch_factory_bridge_test.go backend/internal/app/app.go backend/internal/httpapi/agent_workspace_compat.go
git commit -m "feat(agent): bridge batch factory records into canvas"
~~~

### Task 4: 接入批量工厂书籍入口和画布分组

**Files:**
- Modify: \`frontend/src/shared/api/agent.js\`
- Modify: \`frontend/src/shared/api/batchFactoryV11.js\`
- Modify: \`frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx\`
- Modify: \`frontend/src/user/pages/AgentPageV2.jsx\`
- Modify: \`frontend/src/user/components/AgentCanvas.jsx\`
- Create: \`frontend/src/user/pages/agent-canvas-batch-factory.js\`
- Create: \`frontend/src/user/pages/agent-canvas-batch-factory.test.js\`

**Interfaces:** Consumes Task 3 response \`{ projectId, batchId, bookId, canvas }\`. Produces \`/agent/canvas?taskId=<projectId>&batchId=<batchId>&bookId=<bookId>&view=workflow\` and focuses group \`book:<bookId>\`.

- [ ] **Step 1: Write failing frontend adapter tests**

~~~js
it('opens the bound project at the requested book group', () => {
  expect(agentCanvasURL({ projectId: 'agent-1', batchId: 'batch-1', bookId: 'book-1' }))
    .toBe('/agent/canvas?taskId=agent-1&batchId=batch-1&bookId=book-1&view=workflow');
});
it('uses existing node types inside a book group', () => {
  expect(projectedNodePresentation({ type: 'image', groupId: 'book:book-1' }).groupLabel).toBe('书籍素材');
});
~~~

- [ ] **Step 2: Run test to verify it fails**

Run: \`npm test -- --runInBand frontend/src/user/pages/agent-canvas-batch-factory.test.js\`

Expected: FAIL because the adapter does not exist.

- [ ] **Step 3: Implement entry and grouped view**

Add “Agent 优化” and “编辑内容” on each batch-factory book; bind first, then navigate by the URL contract. In \`AgentPageV2.jsx\`, synchronize the requested book before editing and retain existing debounced saving. In \`AgentCanvas.jsx\`, use existing \`groupId\` for headers, collapse/expand and locate behavior; show source tags and “从画布移除” for mapped nodes; keep the add menu limited to the five allowed types.

- [ ] **Step 4: Run focused frontend tests**

Run: \`npm test -- --runInBand frontend/src/user/pages/agent-canvas-batch-factory.test.js frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/pages/agent-canvas-state.test.js\`

Expected: PASS; routing is stable and automatic saves remain batched.

- [ ] **Step 5: Commit**

~~~powershell
git add frontend/src/shared/api/agent.js frontend/src/shared/api/batchFactoryV11.js frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/AgentPageV2.jsx frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-batch-factory.js frontend/src/user/pages/agent-canvas-batch-factory.test.js
git commit -m "feat(canvas): open batch factory books as canvas groups"
~~~

### Task 5: 统一版本、音频时长、状态和映射节点操作

**Files:**
- Modify: \`backend/internal/agentworkspace/batch_factory_projection.go\`
- Modify: \`backend/internal/httpapi/agent_batch_factory_bridge.go\`
- Modify: \`frontend/src/user/components/AgentNodeComposer.jsx\`
- Modify: \`frontend/src/user/components/AgentCanvas.jsx\`
- Modify: \`frontend/src/user/pages/agent-canvas-batch-factory.js\`
- Test: \`backend/internal/agentworkspace/batch_factory_projection_test.go\`
- Test: \`backend/internal/httpapi/agent_batch_factory_bridge_test.go\`
- Test: \`frontend/src/user/pages/agent-canvas-batch-factory.test.js\`

**Interfaces:** Produces \`ResolveVideoDuration(audioSeconds, allowedSeconds, overrideSeconds) DurationResolution\` and normalized mapped-node state.

- [ ] **Step 1: Write failing timing and version tests**

~~~go
func TestResolveVideoDurationUsesNearestAllowedAudioDuration(t *testing.T) {
  got := ResolveVideoDuration(7.4, []float64{5, 10, 15}, 0)
  assert.Equal(t, 10.0, got.RequestedSeconds)
  assert.True(t, got.Adjusted)
}
func TestResolveVideoDurationPreservesManualOverride(t *testing.T) {
  got := ResolveVideoDuration(7.4, []float64{5, 10, 15}, 15)
  assert.Equal(t, 15.0, got.RequestedSeconds)
  assert.True(t, got.HasMismatchWarning)
}
~~~

~~~js
it('shows a timing mismatch after manual override', () => {
  expect(videoNodeTiming({ audioSeconds: 7.4, selectedSeconds: 15, allowedSeconds: [5, 10, 15] }).warning)
    .toContain('与已连接音频时长不一致');
});
~~~

- [ ] **Step 2: Run test to verify it fails**

Run: \`go test ./backend/internal/agentworkspace -run TestResolveVideoDuration -count=1\`

Run: \`npm test -- --runInBand frontend/src/user/pages/agent-canvas-batch-factory.test.js\`

Expected: FAIL because timing resolution is absent.

- [ ] **Step 3: Implement shared operations and truthful state**

Connected audio sets the default target duration. Fixed model durations use the nearest option. A manual override stays selected and carries the mismatch warning. For mapped nodes, composer actions invoke existing batch-factory image/video/audio generation, primary version, merge and publication operations, then synchronize projection. They cannot create a duplicate Agent asset or generation task. Normalize mapped status to \`queued\`, \`running\`, \`succeeded\`, \`failed\`, \`cancelled\`, preserving safe error and retry data. Unbound user-created Agent nodes retain their existing Agent APIs.

- [ ] **Step 4: Run shared-state tests**

Run: \`go test ./backend/internal/agentworkspace ./backend/internal/httpapi -run 'TestResolveVideoDuration|Test.*BatchFactory.*(Version|Status|Delete|Sync)' -count=1\`

Run: \`npm test -- --runInBand frontend/src/user/pages/agent-canvas-batch-factory.test.js frontend/src/user/pages/agent-canvas-video-jobs.test.js frontend/src/user/pages/agent-canvas-audio-jobs.test.js\`

Expected: PASS; mapped nodes create no duplicate task and state matches batch factory.

- [ ] **Step 5: Commit**

~~~powershell
git add backend/internal/agentworkspace/batch_factory_projection.go backend/internal/httpapi/agent_batch_factory_bridge.go frontend/src/user/components/AgentNodeComposer.jsx frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-canvas-batch-factory.js frontend/src/user/pages/agent-canvas-batch-factory.test.js backend/internal/agentworkspace/batch_factory_projection_test.go backend/internal/httpapi/agent_batch_factory_bridge_test.go
git commit -m "feat(agent): reuse batch factory media state on canvas"
~~~

### Task 6: 真实运行验收和 Go embed 构建

**Files:**
- Create: \`frontend/e2e/batch-factory-agent-shared-production.spec.js\`
- Modify: \`backend/internal/app/app_test.go\`
- Modify: \`scripts/build-go-local.test.ps1\` only if the existing smoke script lacks the needed route check.

**Interfaces:** Consumes Tasks 1-5 and the authenticated local runtime. Produces browser proof that a book has one shared production record visible from both interfaces.

- [ ] **Step 1: Write the failing browser acceptance test**

~~~js
test('batch factory and Agent use the same book media and merge state', async ({ page }) => {
  await openBatchFactoryBook(page, 'book-1');
  await page.getByRole('button', { name: 'Agent 优化' }).click();
  await expect(page).toHaveURL(/\/agent\/canvas\?.*bookId=book-1/);
  await expect(page.getByLabel('书籍分组 书一')).toBeVisible();
  await selectMappedImageVersion(page, '角色一', 'version-2');
  await returnToBatchFactory(page);
  await expect(bookAssetPrimaryVersion(page, '角色一')).toHaveText('version-2');
});
~~~

- [ ] **Step 2: Run test to verify it fails**

Run: \`npm run test:e2e -- --grep "batch factory and Agent use the same book media"\`

Expected: FAIL until Tasks 1-5 are done.

- [ ] **Step 3: Complete acceptance fixture assertions**

The fixture contains two books; one has five image assets, one audio record, two video records and one merge. Assert requested-book focus, only five types, actual-only edges, both-direction version visibility, safe default deletion, duration warning, matching merge/publication state, and refresh persistence.

- [ ] **Step 4: Run all verification**

Run: \`go test ./backend/internal/agentworkspace ./backend/internal/httpapi ./backend/internal/storage ./backend/internal/app -count=1\`

Run: \`npm test -- --runInBand frontend/src/user/pages/agent-canvas-batch-factory.test.js frontend/src/user/pages/agent-canvas-state.test.js frontend/src/user/pages/agent-workspace-navigation.test.js\`

Run: \`npm run build\`

Run: \`powershell -ExecutionPolicy Bypass -File scripts/build-go-local.test.ps1\`

Run: \`npm run test:e2e -- --grep "batch factory and Agent use the same book media"\`

Expected: every command PASS and Go-embedded localhost:4001 shows the same book state in both views.

- [ ] **Step 5: Commit**

~~~powershell
git add frontend/e2e/batch-factory-agent-shared-production.spec.js backend/internal/app/app_test.go scripts/build-go-local.test.ps1
git commit -m "test: verify shared batch factory agent production"
~~~

## Plan Self-Review

- Spec coverage: Tasks 1-3 establish one account-scoped source mapping and safe deletion; Tasks 2 and 4 preserve the five-node model and book grouping; Task 5 handles versions, truthful status and audio timing; Task 6 proves bidirectional merge/publication behavior in the final runtime.
- Placeholder scan: every task names files, interfaces, test command, expected result and minimal behavior.
- Type consistency: later tasks use BatchFactoryBinding, BatchFactoryNodeOrigin, SyncBatchFactoryBook and DurationResolution introduced earlier.
