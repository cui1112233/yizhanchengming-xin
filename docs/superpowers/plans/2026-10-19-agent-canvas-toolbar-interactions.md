# V88 Agent 画布工具与交互 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有 `/agent/canvas` 中交付节点、连线、缩放、分组、对齐、撤销与可靠保存，并保持 Agent 对话和 MySQL revision 保存链路。

**Architecture:** 继续使用现有 `@xyflow/react` 画布。纯画布命令和历史记录放入可单测模块，`AgentCanvas` 负责 React Flow 事件和工具 UI；Go Agent Store 用节点的可选 `groupId` 持久化分组，保存继续经过现有串行队列和 revision 冲突保护。

**Tech Stack:** React 18、Ant Design 5、`@xyflow/react` 12、Lucide React、Node 内置测试、Go、MySQL 8、Go embed。

## Global Constraints

- 直接在当前 `v88` 分支工作，不新建分支，不覆盖或清理现有脏工作区。
- 不修改迁移 `8800001` 或任何已执行迁移；分组字段使用全新 `8800002` 追加迁移。
- 前端继续使用 React + Ant Design，后端和保存接口继续使用 Go。
- 不建立第二套画布、第二套保存 API、旧 JSON 写入或 Node 发布入口。
- 工具栏采用 A 方案：左侧创建工具、顶部中央编辑工具、底部中央缩放工具、右侧可收起 Agent 对话。
- 分组、解组、对齐和等距按钮常驻显示；条件不足时禁用并显示原因。
- 本任务只做本地开发与 `4001` 已登录浏览器验收，不发布公网。
- 每项实现先运行新增测试并观察预期失败，再写最小实现。

---

## File Structure

- `backend/internal/storage/agent_workspace_schema.go`：追加 `8800002` 分组字段迁移。
- `backend/internal/agentworkspace/types.go`、`mysql_store.go`：读写节点 `groupId`。
- `frontend/src/user/pages/agent-canvas-commands.js`：连线、分组、对齐、等距和分组边界纯函数。
- `frontend/src/user/pages/agent-canvas-history.js`：不包含选择状态的撤销/重做历史。
- `frontend/src/user/pages/agent-canvas-state.js`：保存队列增加立即保存和失败重试。
- `frontend/src/user/components/AgentCanvas.jsx`：React Flow 交互和三组浮动工具栏。
- `frontend/src/user/pages/AgentPageV2.jsx`：手动保存入口和真实缩放状态。
- `frontend/src/shared/styles/agent-workspace.css`：A 方案布局和窄屏规则。

---

### Task 1: 持久化节点分组编号

**Files:**
- Modify: `backend/internal/storage/agent_workspace_schema.go`
- Modify: `backend/internal/storage/agent_workspace_schema_test.go`
- Modify: `backend/internal/storage/migrations_mysql84_test.go`
- Modify: `backend/internal/agentworkspace/types.go`
- Modify: `backend/internal/agentworkspace/mysql_store.go`
- Modify: `backend/internal/agentworkspace/store_test.go`
- Modify: `backend/internal/agentworkspace/mysql_store_test.go`

**Interfaces:**
- Consumes: `AgentWorkspaceMigrations() []Migration`、`Node`、`Canvas`、`Store.SaveCanvas`。
- Produces: `Node.GroupID string`，JSON 名称 `groupId`；迁移 `8800002`；MySQL 列 `agent_canvas_nodes.group_id VARCHAR(96) NULL`。

- [ ] **Step 1: 写迁移失败测试**

在 `agent_workspace_schema_test.go` 增加：

```go
func TestAgentWorkspaceGroupMigrationIsAppendOnly(t *testing.T) {
	migrations := AgentWorkspaceMigrations()
	if len(migrations) != 2 { t.Fatalf("len=%d, want 2", len(migrations)) }
	if migrations[0].Version != 8800001 || migrations[0].CallbackChecksum != "v88-agent-workspace-foundation-v1" {
		t.Fatalf("historical migration changed: %+v", migrations[0])
	}
	group := migrations[1]
	if group.Version != 8800002 || group.CallbackChecksum != "v88-agent-canvas-groups-v1" {
		t.Fatalf("group migration=%+v", group)
	}
	if !strings.Contains(strings.ToLower(strings.Join(group.SQL, "\n")), "add column group_id") {
		t.Fatalf("missing group_id migration: %v", group.SQL)
	}
}
```

更新最后迁移测试，断言 `AppMigrations()` 以 `8800002` 结束且历史编号均小于它。

- [ ] **Step 2: 运行测试并确认失败**

Run from `backend`: `go test ./internal/storage -run 'AgentWorkspace|AppMigrations' -count=1`

Expected: FAIL，当前只有 `8800001`。

- [ ] **Step 3: 添加追加迁移和 Go 字段**

在 `AgentWorkspaceMigrations()` 末尾追加：

```go
{Version: 8800002, SQL: []string{
	`ALTER TABLE agent_canvas_nodes ADD COLUMN group_id VARCHAR(96) NULL AFTER asset_ref_json`,
}, CallbackChecksum: "v88-agent-canvas-groups-v1"}
```

在 `Node` 增加：

```go
GroupID string `json:"groupId,omitempty"`
```

- [ ] **Step 4: 写 Store 往返失败测试**

在内存和 MySQL Store 测试中保存 `GroupID: "group-1"`，重新读取并断言：

```go
if got.Nodes[0].GroupID != "group-1" {
	t.Fatalf("groupId=%q, want group-1", got.Nodes[0].GroupID)
}
```

在 MySQL84 测试中连续运行两次 `RunMigrations(ctx, db, AgentWorkspaceMigrations())`，查询 `information_schema.columns`，断言 `group_id` 只有一列。

- [ ] **Step 5: 运行 Store 测试并确认失败**

Run from `backend`: `go test ./internal/agentworkspace ./internal/storage -run 'Canvas|Group|AgentWorkspace' -count=1`

Expected: FAIL，MySQL Store 尚未读写 `group_id`。

- [ ] **Step 6: 完成 MySQL 读写**

查询列、`rows.Scan` 和插入语句全部加入 `group_id`。空值用 `sql.NullString` 读取和写入，保存、读取和 revision 更新后 `GroupID` 保持不变。

Run: `go test ./internal/agentworkspace ./internal/storage -count=1`

Expected: PASS；真实 MySQL 未配置导致的跳过必须单独记录。

- [ ] **Step 7: 提交**

暂存 Task 1 列出的文件并提交：`feat(agent): persist canvas groups`。

---

### Task 2: 实现可测试的画布命令

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-commands.js`
- Create: `frontend/src/user/pages/agent-canvas-commands.test.js`

**Interfaces:**
- Consumes: React Flow 节点和边。
- Produces: `commandAvailability`、`connectCanvasNodes`、`groupCanvasNodes`、`ungroupCanvasNodes`、`alignCanvasNodes`、`distributeCanvasNodes`、`canvasGroupBounds`。

- [ ] **Step 1: 写命令失败测试**

测试自连接、重复边、六种对齐、两种等距、分组/解组和 16px 分组外边距。核心断言：

```js
test('自连接和重复连线不会创建新边', () => {
  const edges = [{ id: 'e1', source: 'a', target: 'b', data: { relation: 'reference' } }];
  assert.deepEqual(connectCanvasNodes(edges, { source: 'a', target: 'a' }), edges);
  assert.deepEqual(connectCanvasNodes(edges, { source: 'a', target: 'b' }), edges);
});

test('分组和解组只改变选中成员', () => {
  const grouped = groupCanvasNodes(nodes, ['a', 'b'], 'group-1');
  assert.equal(grouped[0].data.groupId, 'group-1');
  assert.equal(grouped[2].data.groupId, undefined);
  assert.equal(ungroupCanvasNodes(grouped, ['a'])[0].data.groupId, undefined);
});
```

- [ ] **Step 2: 运行测试并确认模块缺失**

Run from `frontend`: `node --test src/user/pages/agent-canvas-commands.test.js`

Expected: FAIL，命令模块不存在。

- [ ] **Step 3: 实现固定接口**

```js
export function commandAvailability(nodes, selectedIds) {}
export function connectCanvasNodes(edges, connection) {}
export function groupCanvasNodes(nodes, selectedIds, groupId) {}
export function ungroupCanvasNodes(nodes, selectedIds) {}
export function alignCanvasNodes(nodes, selectedIds, mode) {}
export function distributeCanvasNodes(nodes, selectedIds, axis) {}
export function canvasGroupBounds(nodes, groupId, padding = 16) {}
```

尺寸使用 `node.measured?.width ?? node.width ?? 168` 和 `node.measured?.height ?? node.height ?? 88`。所有命令返回新数组，不修改输入节点或未选节点。

- [ ] **Step 4: 运行测试并提交**

Run: `node --test src/user/pages/agent-canvas-commands.test.js`

Expected: PASS。暂存两个命令文件并提交：`feat(agent): add canvas editing commands`。

---

### Task 3: 实现撤销历史和立即保存

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-history.js`
- Create: `frontend/src/user/pages/agent-canvas-history.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-state.js`
- Modify: `frontend/src/user/pages/agent-canvas-state.test.js`

**Interfaces:**
- Consumes: 快照 `{ nodes, edges }` 和现有 `createCanvasSaveQueue`。
- Produces: `createCanvasHistory` 返回 `current/canUndo/canRedo/push/undo/redo/reset`；保存队列新增 `flush()` 和 `retry()`。

- [ ] **Step 1: 写历史失败测试**

```js
test('撤销和重做以一次命令为一个步骤', () => {
  const history = createCanvasHistory({ nodes: [{ id: 'a', position: { x: 0, y: 0 } }], edges: [] });
  history.push({ nodes: [{ id: 'a', position: { x: 40, y: 0 } }], edges: [] });
  assert.equal(history.undo().nodes[0].position.x, 0);
  assert.equal(history.redo().nodes[0].position.x, 40);
});
```

再测试 `reset` 清空旧项目历史，历史快照不包含 `selected` 和临时尺寸。

- [ ] **Step 2: 写立即保存失败测试**

```js
test('flush skips debounce and saves newest canvas', async () => {
  const calls = [];
  const queue = createCanvasSaveQueue({ revision: 0, delay: 5000, save: async canvas => {
    calls.push(canvas); return { canvas: { ...canvas, revision: 1 } };
  }});
  queue.enqueue({ version: 1, revision: 0, nodes: [{ id: 'latest' }], edges: [] });
  await queue.flush();
  assert.equal(calls[0].nodes[0].id, 'latest');
});
```

再验证普通失败后 `retry()` 重试同一份画布，409 仍重新加载服务端版本。

- [ ] **Step 3: 运行测试并确认失败**

Run from `frontend`: `node --test src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-state.test.js`

Expected: FAIL，历史模块、`flush` 和 `retry` 不存在。

- [ ] **Step 4: 实现并验证**

历史内部维护 `past/current/future`；`push` 清空 `future`，`reset` 清空两侧。`flush()` 清除计时器并返回本轮保存 Promise；`retry()` 仅重试最后一次普通失败保留的画布；`dispose()` 保持项目隔离。

Run: `node --test src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-state.test.js`

Expected: PASS。暂存四个文件并提交：`feat(agent): add canvas history and manual save`。

---

### Task 4: 接入 React Flow 真实交互和工具栏

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Create: `frontend/src/user/pages/agent-canvas-ui.source.test.js`

**Interfaces:**
- Consumes: Task 2 命令、Task 3 历史、`canvas/onChange/onSelectionChange`。
- Produces: `AgentCanvas({ canvas, onChange, onSelectionChange, onViewportChange })` 和真实画布操作。

- [ ] **Step 1: 写 UI 接线失败测试**

源码测试必须断言 `onConnect`、`onNodeDragStop`、`connectCanvasNodes`、`createCanvasHistory`、快捷键判断，以及类名 `agent-canvas-toolrail`、`agent-canvas-editbar`、`agent-canvas-zoombar`。按钮必须使用 Lucide 图标、`Tooltip` 和 `aria-label`。

- [ ] **Step 2: 运行测试并确认失败**

Run from `frontend`: `node --test src/user/pages/agent-canvas-ui.source.test.js`

Expected: FAIL，当前没有三组工具栏、连线和历史。

- [ ] **Step 3: 接入 React Flow 实例和左侧工具栏**

通过 `onInit={setFlowInstance}` 保存实例。选择、平移、连线和新增节点使用 `MousePointer2`、`Hand`、`GitBranch`、`Plus` 图标。新增节点放在视图中央，类型为 `copy`，标题为“新节点”。`onConnect` 调用 `connectCanvasNodes`，自连接和重复边不触发保存。

- [ ] **Step 4: 接入历史边界**

节点拖动中只更新显示，`onNodeDragStop` 只推入一次历史并保存一次。新增/删除节点、边、分组、解组、对齐和等距各形成一个步骤；选择和尺寸变化不入历史。项目切换或冲突重载调用 `history.reset`。画布聚焦且输入框未聚焦时支持 `Ctrl+Z`、`Ctrl+Shift+Z`、`Ctrl+Y`。

- [ ] **Step 5: 接入分组显示和拖动**

按 `groupId` 汇总成员并使用 `canvasGroupBounds` 绘制 16px 外边距分组框。标题固定“节点分组”。拖动分组标题同步移动所有成员，结束时形成一个历史和保存步骤；解组保留绝对位置。

- [ ] **Step 6: 接入编辑栏和缩放栏**

顶部编辑栏包含撤销、重做、分组、解组、六种对齐和两种等距；对齐项使用 Ant Design `Dropdown` 收纳。底部调用 React Flow 的 `zoomOut`、`zoomIn`、`fitView`；聚焦选中调用 `fitView({ nodes, padding: 0.3 })`，显示四舍五入缩放百分比。

- [ ] **Step 7: 运行测试并提交**

Run from `frontend`: `node --test src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-ui.source.test.js`

Expected: PASS。暂存组件和 UI 测试并提交：`feat(agent): add professional canvas controls`。

---

### Task 5: 接入手动保存和 A 方案响应式样式

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Interfaces:**
- Consumes: `flush()`、`retry()` 和 `AgentCanvas.onViewportChange`。
- Produces: 手动保存按钮、真实缩放状态、桌面和窄屏稳定布局。

- [ ] **Step 1: 写布局失败测试**

断言页面调用 `canvasSaveQueueRef.current?.flush()`，失败态调用 `retry()`，保存按钮使用 Lucide `Save`。CSS 测试断言三组工具栏、36px 固定按钮、最大 8px 圆角、禁用态、顶部横向滚动和 560px 窄屏避让规则。

- [ ] **Step 2: 运行测试并确认失败**

Run from `frontend`: `node --test src/user/pages/agent-workspace.layout.test.js`

Expected: FAIL，页面没有手动保存和三组工具栏样式。

- [ ] **Step 3: 接入保存和缩放状态**

顶部加入 `Save` 图标按钮：`dirty/saving` 调用 `flush()`，`error` 显示“重试保存”并调用 `retry()`，`saved/idle` 禁用并提示“当前内容已保存”。`onViewportChange` 更新真实缩放，替换写死的“缩放 100%”。

- [ ] **Step 4: 完成 A 方案样式**

左侧工具栏距画布左上 12px、宽 44px；顶部编辑栏居中且最大宽度 `calc(100% - 136px)`；底部缩放栏距底部 12px；图标按钮 36px × 36px、圆角不超过 8px。560px 以下沿用有界底部 Agent 面板，工具栏不得覆盖输入区。

- [ ] **Step 5: 运行测试、构建并提交**

Run from `frontend`:

`node --test src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-ui.source.test.js src/user/pages/agent-workspace.layout.test.js`

Run: `npm run build`

Expected: 全部 PASS 且 Vite 构建无错误。暂存 Task 5 文件并提交：`feat(agent): finish canvas toolbar layout`。

---

### Task 6: 全链路验证和本地已登录验收

**Files:**
- Modify only when verification exposes a scoped defect in Tasks 1-5 files.

**Interfaces:**
- Consumes: Go 分组合同、画布命令、保存队列和 A 方案 UI。
- Produces: 测试结果、本地 `4001` 服务和真实浏览器操作证据。

- [ ] **Step 1: 运行 Go 目标测试和全量测试**

Run from `backend`:

`go test ./internal/storage ./internal/agentworkspace ./internal/httpapi -count=1`

`go test ./... -count=1`

Expected: PASS。真实 MySQL 被跳过时记录原因；现有无关失败保留包名和错误，不修改无关代码掩盖。

- [ ] **Step 2: 运行前端目标测试和构建**

Run from `frontend`:

`node --test src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-canvas-history.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-ui.source.test.js src/user/pages/agent-workspace.layout.test.js`

`npm run build`

Expected: PASS。

- [ ] **Step 3: 构建并启动本地 Go embed**

Run from repository root:

`.\scripts\build-go-local.ps1`

`.\scripts\start-local-go-review.ps1 -UseYcmMysql`

`.\scripts\check-local-go.ps1`

Expected: `4001` 由本次 Go 二进制监听，健康检查通过，不启动旧 Node 页面。

- [ ] **Step 4: 已登录浏览器验收**

在真实项目中验证：新建三个节点和两条边；拒绝自连接和重复边；两个节点分组后整体拖动并刷新恢复；对齐两个节点和等距三个节点；逐步撤销和重做；缩放、适合屏幕和聚焦选中；自动保存和手动保存；收起/打开 Agent；桌面与 390px 窄屏无重叠。

- [ ] **Step 5: 检查最终差异和证据**

运行 `git diff --check`、`git status --short`、`git log -5 --oneline`。只暂存本计划文件。若验收触发 scoped 修复，提交信息使用 `fix(agent): verify canvas interactions`。

最终记录 HEAD、目标测试、全量测试、构建、监听和浏览器验收结果。不得把 HTTP 200、构建成功或未登录页面描述成完成。
