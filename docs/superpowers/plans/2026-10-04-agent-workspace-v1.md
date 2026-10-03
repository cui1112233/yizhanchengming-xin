# Agent Workspace V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship an openable `/agent` workspace with durable MySQL threads/messages/tasks/tool calls, system-help/navigation tools, contextual task/media cards, and no special local media storage.

**Architecture:** The React workspace talks only to the Go HTTP API. The Go Agent service persists all conversational/task/tool state in MySQL and interprets a small deterministic V1 command set through a server-side allow-listed tool registry. The browser executes returned `ui.navigate` tool calls; future creative tools plug into the same registry and reuse existing product services, Redis workers, and TOS media assets.

**Tech Stack:** Go 1.23, net/http, database/sql + MySQL, React 18.3.1, Ant Design 5.21.6, Vite 6.4.3.

**Spec:** `docs/superpowers/specs/2026-10-04-agent-workspace-v1-design.md`

## Global Constraints

- Production backend/business logic remains Go; Node/npm are frontend build tools only.
- MySQL is durable source of truth; Redis is transport only.
- Formal image/video bytes live in TOS; Agent stores only `media_asset_id` references.
- Agent tools must reuse product capabilities rather than implementing second copies of novel/script/video pipelines.
- Existing Batch Factory route must remain available.
- Owner identity is server-side; clients never submit an owner field.
- Work is performed on branch `agent-workspace-v1`, intentionally outside the existing `feat/**` push trigger.

## Review Focus

- Cross-owner thread IDs must return 404 and never leak messages/tasks/tool calls.
- Navigation tool must reject arbitrary/external URLs and only return allow-listed product paths.
- Empty/oversized chat content must be rejected without partial writes.
- A tool-call write failure must not claim that navigation/execution completed.
- Frontend must tolerate an empty database and API failures without rendering a blank page.

---

### Task 1: Agent schema and SQL repository

**Files:**
- Create: `api/migrations/00003_agent_workspace.sql`
- Create: `api/internal/agent/types.go`
- Create: `api/internal/agent/store.go`
- Test: `api/internal/agent/store_test.go`

**Interfaces:**
- Produces: `agent.Store` with `CreateThread`, `ListThreads`, `GetThread`, `AppendMessage`, `CreateTask`, `ListTasks`, `CreateToolCall`, `UpdateThreadTitle`, and `AttachMessageMedia`.

- [ ] **Step 1: Write failing SQL-store tests** for owner scoping, persisted history, task/tool-call ordering, and media references.
- [ ] **Step 2: Verify the tests fail** because the Agent package/store does not exist.
- [ ] **Step 3: Add Goose migration** creating `agent_threads`, `agent_messages`, `agent_tasks`, `agent_tool_calls`, `agent_message_media` with owner/thread indexes and foreign keys.
- [ ] **Step 4: Implement `SQLStore`** with crypto-random IDs and transactional append operations where needed.
- [ ] **Step 5: Run `go test ./internal/agent/...`** and require PASS.

### Task 2: Tool registry and V1 responder

**Files:**
- Create: `api/internal/agent/tools.go`
- Create: `api/internal/agent/responder.go`
- Test: `api/internal/agent/responder_test.go`

**Interfaces:**
- Consumes: Agent domain types from Task 1.
- Produces: `Responder.Respond(ctx, owner, threadID, content) (Response, error)` and allow-listed `ToolCall` records.

- [ ] **Step 1: Write failing tests** for “打开小说获取”, “打开批量工厂”, “图片模型在哪里设置”, generic help, and an attempted external URL/navigation injection.
- [ ] **Step 2: Verify RED.**
- [ ] **Step 3: Implement server-side registry** for product destinations and help aliases.
- [ ] **Step 4: Implement deterministic V1 responder** returning assistant copy plus optional `ui.navigate` tool call.
- [ ] **Step 5: Run `go test ./internal/agent/...`** and require PASS.

### Task 3: Agent service and HTTP API

**Files:**
- Create: `api/internal/agent/service.go`
- Create: `api/internal/httpapi/agent.go`
- Test: `api/internal/httpapi/agent_test.go`
- Modify: `api/internal/httpapi/router.go`

**Interfaces:**
- Consumes: `agent.Store`, `agent.Responder`, server owner resolver.
- Produces REST endpoints under `/api/agent/*` from the spec.

- [ ] **Step 1: Write failing handler tests** for list/create thread, get thread, post message, missing owner, invalid JSON, cross-owner 404, and navigation-tool response.
- [ ] **Step 2: Verify RED.**
- [ ] **Step 3: Implement `agent.Service`** to persist user message -> generate response -> persist assistant/tool call -> return updated records.
- [ ] **Step 4: Implement HTTP handler** with bounded JSON bodies and route parsing.
- [ ] **Step 5: Extend router constructor** without breaking existing Batch Factory constructors/tests.
- [ ] **Step 6: Run `go test ./internal/httpapi/... ./internal/agent/...`** and require PASS.

### Task 4: Wire Agent into Go server

**Files:**
- Modify: `api/cmd/server/main.go`
- Test: `api/cmd/server/main_test.go`

**Interfaces:**
- Consumes: DB connection and owner resolver already used by Intake.
- Produces: live Agent API from the same Go server.

- [ ] **Step 1: Add failing server wiring test** showing Agent handler/service is included.
- [ ] **Step 2: Verify RED.**
- [ ] **Step 3: Instantiate `agent.SQLStore`, `agent.Responder`, `agent.Service` and pass them to the router.**
- [ ] **Step 4: Run `go test ./cmd/server ./internal/httpapi/... ./internal/agent/...`** and require PASS.

### Task 5: Frontend Agent data client and intent UI model

**Files:**
- Create: `前台/src/agentApi.js`
- Create: `前台/src/agentViewModel.js`
- Test: `前台/src/agentViewModel.test.js`

**Interfaces:**
- Produces: API functions for thread/message calls plus pure helpers for status labels, task grouping, tool-card labels, and safe navigation execution.

- [ ] **Step 1: Write failing Node tests** for tool label/status mapping, task grouping, allowed navigation handling, and media-reference normalization.
- [ ] **Step 2: Verify RED.**
- [ ] **Step 3: Implement API client and pure view helpers.**
- [ ] **Step 4: Run `npm test`** and require PASS.

### Task 6: Full Agent workspace UI and routing

**Files:**
- Create: `前台/src/AgentWorkspace.jsx`
- Create: `前台/src/AgentWorkspace.css`
- Create: `前台/src/BatchFactoryPage.jsx` by extracting current Batch Factory page unchanged in behavior
- Modify: `前台/src/App.jsx`
- Modify: `前台/src/styles.css`

**Interfaces:**
- Consumes: Task 5 API/helpers and existing Batch Factory intake logic.
- Produces: `/agent` and `/batch-factory` pages in one embedded SPA; `/` defaults to Agent workspace for V1.

- [ ] **Step 1: Extract current Batch Factory UI** so it remains reachable at `/batch-factory`.
- [ ] **Step 2: Implement Agent shell** with thread sidebar, new chat, header/status, message stream, task/tool/media cards, contextual inspector, responsive drawers, and composer.
- [ ] **Step 3: Execute returned `ui.navigate` only through safe frontend helper.**
- [ ] **Step 4: Add robust loading/error/empty states and optimistic send lock.**
- [ ] **Step 5: Run `npm test && npm run build`** and require PASS.

### Task 7: Embedded SPA route verification

**Files:**
- Modify/Test: `api/internal/webui/handler_test.go` only if existing SPA fallback tests do not cover `/agent`.
- Test: `api/cmd/server/main_test.go`

**Interfaces:**
- Confirms `/agent` is served by the embedded SPA while `/api/agent/*` stays API-only.

- [ ] **Step 1: Add route-isolation test** for `/agent` vs `/api/agent/threads`.
- [ ] **Step 2: Verify RED if coverage is absent.**
- [ ] **Step 3: Make minimal routing fix only if needed.**
- [ ] **Step 4: Run full Go suite and frontend test/build.**

### Task 8: Whole-branch verification

**Files:** none unless fixes are required.

- [ ] **Step 1: Re-read spec and compare all requirements to implementation.**
- [ ] **Step 2: Verify no Agent table stores local media paths/provider URLs as formal assets.**
- [ ] **Step 3: Verify branch remains `agent-workspace-v1` and has not modified the workflow trigger.**
- [ ] **Step 4: Run final `go test ./...` and `npm test && npm run build` when an execution environment is available; otherwise explicitly report the verification gap rather than claiming green.**
