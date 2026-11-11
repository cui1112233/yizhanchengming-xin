# AI-CUT Agent 功能与交互对齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把当前五关键词入口重写成真正能够读取并修改 AI-CUT 项目时间线的 Agent，并对齐洛水的 Agent/Q&A、工具过程、取消、项目聊天持久化、全局导航收起和 Agent 面板独立收起交互。

**Architecture:** AI-CUT Agent 是 AI-CUT 编辑器的自然语言控制层，不是独立产品。React 只提交指令、展示事件并刷新同一个项目；Go 后端负责系统提示词、模型循环、工具参数校验和工具执行；MySQL 保存运行、消息和事件事实，Redis 只负责唤醒任务；素材继续走现有对象存储，时间线继续使用现有 revision/undo 机制。

**Tech Stack:** React 18、Ant Design、Go、MySQL、Redis、现有文本模型路由、现有 AI-CUT Store、Go embed。

## Global Constraints

- 只复现已经有证据的可观察行为，不复制洛水原代码、提示词原文、密钥、私有接口、图标或专有资源。
- 系统提示词只放在 Go 后端 `prompt` 模块，前端不得硬编码系统提示词。
- 不改变一战晟铭首页、私人定制、小说获取、水货生产和现有 Agent 工作区。
- 不把浏览器直接调用剪映；剪映草稿只能作为以后可选的 Windows 本地执行器。
- MySQL 是 Agent 运行和消息的事实来源；Redis 只排队和唤醒，不保存最终状态。
- 每一个时间线写工具都必须携带 `expectedRevision`，版本冲突不得覆盖用户的新修改。
- 不调用真实付费模型完成自动测试；模型测试一律使用可控假 Provider。
- 主题必须继续读取全站 CSS 变量，不能把浅色或深色写死。
- 当前工作区已有大量用户改动；不重置、不清理、不另建分支、不推送、不部署，也不提交，除非用户另行明确授权。

---

## 文件职责图

- `backend/internal/aicut/agent_types.go`：Agent 运行、消息、事件、工具调用的数据结构和状态机。
- `backend/internal/aicut/agent_store.go`：Agent 所需的存储接口，不掺入 HTTP 逻辑。
- `backend/internal/aicut/agent_mysql_store.go`：运行、消息和事件的 MySQL 实现。
- `backend/internal/aicut/agent_tools.go`：受控工具注册表、JSON Schema 校验和时间线工具执行。
- `backend/internal/aicut/agent_service.go`：最多 30 轮的模型→工具→结果回灌循环、取消和失败收口。
- `backend/internal/prompt/aicut_agent.go`：我们独立撰写的 Agent/Q&A 系统规则和工具输出协议。
- `backend/internal/httpapi/aicut_agent_routes.go`：创建运行、查询、事件流、取消和聊天历史 API。
- `frontend/src/shared/api/aicutAgent.js`：前端唯一的 Agent API 客户端。
- `frontend/src/user/components/AICutAgentPanel.jsx`：聊天、模式、模型、工具事件、停止和独立折叠 UI。
- `frontend/src/user/pages/AICutPage.jsx`：只负责把项目、时间线和刷新回调交给 Agent 面板。
- `frontend/src/user/pages/aicut-page.css`：对齐三栏比例、Agent 窄轨、折叠动画和主题变量。

---

### Task 1: 冻结证据规格和当前基线

**Files:**
- Modify: `docs/superpowers/specs/2026-11-09-aicut-rea-clean-room-spec.md`
- Create: `docs/acceptance/aicut-agent-parity-matrix.md`
- Test: `frontend/src/user/pages/aicut-workbench.layout.test.js`

**Interfaces:**
- Consumes: 洛水恢复前端中 Agent/Q&A、30 轮工具循环、取消、项目持久化和双层折叠的行号证据。
- Produces: 每个行为对应“原版证据、我们的接口、自动测试、浏览器证据、当前状态”的矩阵。

- [ ] **Step 1: 写入必须保持为真的源代码守卫测试**

```js
test('AI-CUT keeps global navigation and owns a separate collapsible Agent panel', () => {
  assert.match(layoutSource, /sidebarCollapsed/);
  assert.match(pageSource, /AICutAgentPanel/);
  assert.match(agentSource, /展开 Agent/);
  assert.match(agentSource, /收起 Agent/);
});
```

- [ ] **Step 2: 运行守卫并确认当前失败点**

Run: `cd frontend; node --test src/user/pages/aicut-workbench.layout.test.js`

Expected: Agent 独立折叠相关断言失败；全局导航断言通过。

- [ ] **Step 3: 在规格中补齐精确调用链**

必须明确写成：`instruction + mode + modelId + activeSkill + project snapshot + last messages -> model -> validated toolCalls -> tool executor -> saved timeline revision -> tool result -> next model turn -> final reply`，并列出空指令、重复发送、非法工具、版本冲突、模型失败和取消的提示语义。

- [ ] **Step 4: 建立验收矩阵，不把未知行为写成已完成**

矩阵至少包含：全局导航收起、Agent 内部收起、Agent/Q&A、模型选择、项目聊天恢复、工具开始/完成/失败、时间线刷新、停止、30 轮上限、正常/空值/重复/错误/取消。

---

### Task 2: 建立 Agent 持久状态和 API 契约

**Files:**
- Create: `backend/internal/aicut/agent_types.go`
- Create: `backend/internal/aicut/agent_store.go`
- Create: `backend/internal/aicut/agent_mysql_store.go`
- Create: `backend/internal/aicut/agent_mysql_store_test.go`
- Modify: `backend/internal/storage/aicut_schema.go`
- Modify: `backend/internal/storage/aicut_schema_test.go`

**Interfaces:**
- Produces: `CreateAgentRun(ctx, run)`, `ClaimAgentRun(ctx)`, `AppendAgentMessage(ctx, message)`, `AppendAgentEvent(ctx, event)`, `RequestAgentCancel(ctx, owner, runID)`, `AgentRun(ctx, owner, runID)`, `AgentMessages(ctx, owner, projectID, limit)`。
- State: `queued | running | cancel_requested | succeeded | failed | canceled`。

- [ ] **Step 1: 先写状态机和所有者隔离失败测试**

```go
func TestAgentRunStateAndOwnerIsolation(t *testing.T) {
    run := AgentRun{ID: "run-1", Owner: "alice", ProjectID: "cut-1", State: AgentQueued}
    // alice 可读取；bob 必须得到 ErrAgentRunNotFound。
    // queued -> running -> succeeded 合法；succeeded -> running 必须拒绝。
}
```

- [ ] **Step 2: 运行测试确认结构尚不存在**

Run: `cd backend; go test ./internal/aicut ./internal/storage -run 'AgentRun|AICutSchema'`

Expected: FAIL，缺少 Agent 类型和表结构。

- [ ] **Step 3: 增加三张 MySQL 表**

`aicut_agent_runs` 保存运行状态、mode、model_id、active_skill_id、expected_revision、turn_count、error、cancel_requested_at；`aicut_agent_messages` 保存用户/助手/工具消息；`aicut_agent_events` 保存单调递增 sequence 的 thinking/tool_start/tool_done/tool_failed/reply/usage 事件。所有表必须带 owner、project_id 和时间字段。

- [ ] **Step 4: 实现原子领取、幂等取消和顺序事件**

重复领取只能有一个 worker 成功；对终态重复取消返回当前终态；相同 `clientRequestId` 重复提交只能返回同一个 run，不得重复执行工具。

- [ ] **Step 5: 运行存储和迁移测试**

Run: `cd backend; go test ./internal/aicut ./internal/storage -run 'AgentRun|AICutSchema'`

Expected: PASS。

---

### Task 3: 实现受控时间线工具注册表

**Files:**
- Create: `backend/internal/aicut/agent_tools.go`
- Create: `backend/internal/aicut/agent_tools_test.go`
- Modify: `backend/internal/aicut/types.go`
- Reuse: `backend/internal/aicut/store.go`

**Interfaces:**
- Consumes: `ToolCall{Name string, Arguments json.RawMessage}`、owner、projectID、expectedRevision。
- Produces: `ToolResult{Name string, OK bool, Message string, Revision int64, Data any}`。
- Allowed initial tools: `get_project_snapshot`, `place_asset`, `move_clip`, `split_clip`, `delete_clip`, `set_playback_rate`, `add_bgm`, `replace_subtitle_text`, `undo_timeline`, `redo_timeline`, `start_asr`, `smart_split_asset`, `extract_audio`。

- [ ] **Step 1: 为每个工具写正常、缺字段、错误 ID、锁定轨道和旧 revision 测试**

```go
func TestAgentToolDeleteClipRejectsLockedTrackAndPreservesTimeline(t *testing.T) {}
func TestAgentToolMoveClipRejectsStaleRevision(t *testing.T) {}
func TestAgentToolCallRejectsUnknownTool(t *testing.T) {}
func TestAgentToolCallReplayDoesNotApplyTwice(t *testing.T) {}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend; go test ./internal/aicut -run AgentTool`

Expected: FAIL，工具注册表尚不存在。

- [ ] **Step 3: 实现工具描述和严格参数校验**

未知字段、缺少字段、负时间、空 ID、越界 source range 和未知工具全部返回 `OK=false`，且不得调用 `SaveTimeline`。模型永远不能直接获得 SQL、对象存储密钥或任意文件路径。

- [ ] **Step 4: 实现幂等工具执行**

每个 tool call 带 `callId`；存储层记录 `(run_id, call_id)`，重复 callId 返回第一次结果，不再次修改时间线。

- [ ] **Step 5: 运行工具测试**

Run: `cd backend; go test ./internal/aicut -run AgentTool`

Expected: PASS。

---

### Task 4: 实现后端提示词和 30 轮 Agent 循环

**Files:**
- Create: `backend/internal/prompt/aicut_agent.go`
- Create: `backend/internal/prompt/aicut_agent_test.go`
- Create: `backend/internal/aicut/agent_service.go`
- Create: `backend/internal/aicut/agent_service_test.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**
- Consumes: `AgentRunInput{Instruction, Mode, ModelID, ActiveSkillID, ExpectedRevision, ClientRequestID}`。
- Produces: 持久化事件和终态；Q&A 只产生 reply，Agent 才允许 toolCalls。
- Model envelope: `{"reply":"...","toolCalls":[{"id":"call-1","name":"move_clip","arguments":{...}}],"followup":null}`。

- [ ] **Step 1: 写 Q&A 不改时间线和 Agent 工具回灌测试**

```go
func TestAgentQAModeNeverExecutesToolCalls(t *testing.T) {}
func TestAgentFeedsToolResultBackIntoNextModelTurn(t *testing.T) {}
func TestAgentStopsAfterThirtyTurns(t *testing.T) {}
func TestAgentCancelStopsBeforeNextTool(t *testing.T) {}
func TestAgentBadJSONRetriesOnceThenFails(t *testing.T) {}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend; go test ./internal/aicut ./internal/prompt -run 'Agent|AICutAgent'`

Expected: FAIL。

- [ ] **Step 3: 独立撰写后端系统规则**

规则必须要求模型只使用提供的素材 ID/轨道 ID/片段 ID，只输出约定 JSON，不得声称未执行的工具成功；Q&A 模式明确禁止工具；工具失败必须在下一轮上下文中可见。

- [ ] **Step 4: 实现循环和事件**

每轮依次写入 `thinking -> tool_start -> tool_done/tool_failed`；无工具调用且有 reply 时成功结束；瞬时模型错误最多重试一次；非法 JSON 最多修复一次；30 轮仍未结束则失败并提示“执行步骤过多，请缩小任务范围”。

- [ ] **Step 5: 接入现有用户模型路由**

必须使用用户在 API 配置中启用的文本模型；未选择且无默认文本模型时返回“请先到 API 配置启用文本模型”，不得悄悄使用假结果。

- [ ] **Step 6: 运行循环测试**

Run: `cd backend; go test ./internal/aicut ./internal/prompt -run 'Agent|AICutAgent'`

Expected: PASS，且测试不访问网络。

---

### Task 5: 暴露运行、事件和取消 API

**Files:**
- Create: `backend/internal/httpapi/aicut_agent_routes.go`
- Create: `backend/internal/httpapi/aicut_agent_routes_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**
- `POST /api/ai-cut/projects/{id}/agent-runs`
- `GET /api/ai-cut/projects/{id}/agent-runs/{runID}`
- `GET /api/ai-cut/projects/{id}/agent-runs/{runID}/events?after={sequence}`
- `POST /api/ai-cut/projects/{id}/agent-runs/{runID}/cancel`
- `GET /api/ai-cut/projects/{id}/agent-messages?limit=100`

- [ ] **Step 1: 写 HTTP 合约测试**

覆盖：正常创建 202、空指令 400、项目不存在 404、模型缺失 422、重复 `clientRequestId` 返回同一 run、跨用户 404、取消 200、终态重复取消幂等、events 按 sequence 增量返回。

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend; go test ./internal/httpapi -run AICutAgent`

Expected: FAIL，路由不存在。

- [ ] **Step 3: 实现路由和 Redis 唤醒**

创建运行只写 MySQL 后发送 Redis 唤醒；Redis 不可用时将运行标记为 failed 并返回明确错误，不把 queued 永久留在页面。取消只写取消请求；worker 在模型调用前后和每个工具前检查取消状态。

- [ ] **Step 4: 运行 HTTP 测试**

Run: `cd backend; go test ./internal/httpapi -run AICutAgent`

Expected: PASS。

---

### Task 6: 重写前端 Agent 面板并增加独立折叠

**Files:**
- Create: `frontend/src/shared/api/aicutAgent.js`
- Create: `frontend/src/shared/api/aicutAgent.test.js`
- Create: `frontend/src/user/components/AICutAgentPanel.jsx`
- Create: `frontend/src/user/components/AICutAgentPanel.source.test.js`
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-page.css`
- Modify: `frontend/src/user/pages/aicut-workbench.layout.test.js`
- Remove usage from page: `resolveAICutAgentTool` in `frontend/src/user/pages/aicutTimeline.js`

**Interfaces:**
- Props: `projectId`, `timelineRevision`, `models`, `activeSkill`, `onTimelineChanged(revision)`。
- UI state: `mode='agent'|'ask'`, `collapsed`, `running`, `events`, `messages`, `selectedModelId`。

- [ ] **Step 1: 写前端合约和源代码测试**

```js
test('collapsed Agent keeps a narrow rail and can expand again', () => {});
test('ask mode explains that it will not change the timeline', () => {});
test('running request shows stop and disables duplicate send', () => {});
test('tool_done with a newer revision asks the page to reload timeline', () => {});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend; node --test src/shared/api/aicutAgent.test.js src/user/components/AICutAgentPanel.source.test.js src/user/pages/aicut-workbench.layout.test.js`

Expected: FAIL。

- [ ] **Step 3: 用真实 API 替换五关键词路由**

发送时生成稳定 `clientRequestId`；运行中重复点击不创建第二个 run；轮询事件时携带最后 sequence；tool_done 返回新 revision 时调用 `onTimelineChanged` 重新加载服务器时间线，而不是在浏览器伪造成功。

- [ ] **Step 4: 实现原版语义的面板内容**

完整态包含 AI-CUT Agent 标题、就绪/运行状态、Agent/Q&A 切换、模型选择、四个快捷卡片、消息历史、工具步骤、输入框和停止按钮；折叠态只保留窄轨、状态点和“展开 Agent”。折叠选择保存到 `localStorage['aicut.agentCollapsed']`，模型选择保存到 `localStorage['aicut.agentModel']`。

- [ ] **Step 5: 保留全局导航折叠并调整编辑区网格**

展开 Agent 时第一列 350px；折叠时第一列 56px；中间素材列和右侧预览自动扩展。全局 `UserLayout` 的“收起导航/展开导航”不得删除，也不得和 Agent 折叠共用一个状态。

- [ ] **Step 6: 运行前端测试和构建**

Run: `cd frontend; node --test src/shared/api/aicutAgent.test.js src/user/components/AICutAgentPanel.source.test.js src/user/pages/aicut-workbench.layout.test.js`

Run: `cd frontend; npm run build`

Expected: 全部 PASS，Vite 构建成功。

---

### Task 7: 同一组输入做原版规则与新实现对照

**Files:**
- Create: `backend/internal/httpapi/aicut_agent_parity_test.go`
- Modify: `docs/acceptance/aicut-agent-parity-matrix.md`

**Interfaces:**
- Consumes: 固定项目 `cut-parity-1`、两段视频片段、一段音频、两条字幕和可控假模型响应。
- Produces: 每个场景的运行状态、事件序列、最终 revision 和时间线差异。

- [ ] **Step 1: 正常输入**

输入：“把 clip-b 从 5000ms 移到 3000ms，然后删除 subtitle-2。”期望两个 tool_done、revision 连续增加、最终 reply 成功、刷新后时间线与工具结果一致。

- [ ] **Step 2: 空值输入**

输入仅空格。期望 HTTP 400，不创建 run、不调用模型、不改变时间线。

- [ ] **Step 3: 重复操作**

相同 `clientRequestId` 连续提交两次。期望返回相同 runID；相同 callId 不得二次移动或二次删除。

- [ ] **Step 4: 错误操作**

模型请求删除不存在的 clip。期望 tool_failed 进入下一轮上下文；最终时间线不变，助手不得声称删除成功。

- [ ] **Step 5: 取消操作**

假模型阻塞时请求取消。期望 canceled 终态、没有后续工具、时间线不变、页面显示“已取消”。

- [ ] **Step 6: Q&A 输入**

Q&A 模式询问“帮我删掉第二段”。即便假模型返回 toolCalls，也必须丢弃工具并回复“问答模式不会改时间轴”。

- [ ] **Step 7: 运行整组对照测试**

Run: `cd backend; go test ./internal/aicut ./internal/httpapi -run 'AICutAgent|AgentTool|AgentParity'`

Expected: PASS；矩阵逐项附测试名和结果，不以截图代替数据结果。

---

### Task 8: 浏览器 UI、主题和运行时验收

**Files:**
- Modify: `docs/acceptance/aicut-agent-parity-matrix.md`
- Generated through existing build only: embedded frontend assets and local Go binary。

**Interfaces:**
- Runtime: `http://127.0.0.1:4001/ai-cut`。
- Viewports: 1920×1080、1440×900、1280×720。

- [ ] **Step 1: 跑完整的相关自动测试**

Run: `cd backend; go test ./internal/aicut ./internal/httpapi ./internal/storage ./internal/prompt`

Run: `cd frontend; npm test`

Run: `cd frontend; npm run build`

Expected: 新增测试全部通过；如全量套件存在既有失败，单独记录且不得冒充本次通过。

- [ ] **Step 2: 同步 Go embed 并重建本地二进制**

使用仓库已有构建流程，不手工复制零散文件；只重启确认位于本项目路径的 `qiantie-local.exe`。

- [ ] **Step 3: 验证运行身份**

检查 `/health` 和 `/api/runtime-build-info`，确认 4001 提供的是刚构建的本地版本。

- [ ] **Step 4: 逐项点击验收双层折叠**

先点击全局导航收起并截图，再恢复；再点击 Agent 收起，确认素材与预览区扩展且窄轨仍可点击恢复；两个状态互不影响。

- [ ] **Step 5: 验收主题不是写死颜色**

依次切换深夜黑、明亮白和任一彩色主题；Agent、素材区、预览工具栏和时间线必须跟随全站 CSS 变量，视频预览画布保持黑色。

- [ ] **Step 6: 浏览器执行五组对照**

使用 Task 7 的正常、空值、重复、错误、取消输入；记录页面消息、工具事件、最终时间线 revision 和刷新后的数据。没有真实模型配置时使用本地假 Provider，不消耗付费额度。

- [ ] **Step 7: 更新验收矩阵**

只有同时具备源码、测试和 4001 可见运行证据的项目才标记“完成”；其余标记“部分”或“未完成”。

---

## 执行顺序和停检点

1. Task 1–5 完成后停检：后端 Agent 能通过假模型真实修改时间线，五组边界测试通过。
2. Task 6 完成后停检：4001 不再使用关键词壳，Agent 内部收起和全局导航收起互不干扰。
3. Task 7–8 完成后停检：自动测试、同输入对照、主题切换和浏览器截图都有证据。
4. 未通过任一停检点，不进入“已一致”或“已完成”结论。

## 自检结果

- 规格覆盖：Agent/Q&A、模型、Skill、项目历史、工具回灌、取消、失败、30 轮、双层折叠、主题和五类输入均有对应任务。
- 步骤完整性：每项都给出了文件、接口、输入输出、失败条件和验证命令，没有空白实施步骤。
- 类型一致性：run 状态、事件 sequence、clientRequestId、callId 和 timeline revision 从存储到 API、前端及验收保持同名。
- 边界说明：剪映桌面草稿不属于本批网页 Agent 交付；本批目标是服务器可执行的 AI-CUT 和 Agent 时间线工具。
