# Agent Studio Go Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the legacy Agent creation workspace to React and Go while making MySQL/TOS the durable source of truth and reusing the existing session, capability, generation, and taskruntime boundaries.

**Architecture:** Create a focused `agentstudio` Go package with a MySQL store for free-standing, account-scoped Agent projects, messages, canvas revisions, attachments, executions, and skill versions. HTTP handlers authorize through the existing Cookie Session/capability middleware and delegate executable work only to the existing generation/taskruntime contract; the unavailable-provider path persists a truthful `executor_unavailable` execution. The React implementation adds `/agent` and `/agent/canvas` views backed solely by these APIs and dedicated CSS inspired by the old V88 layout.

**Tech Stack:** Go, MySQL/Goose, existing authn/httpapi/generation/taskruntime/TOS adapters, React 18, Ant Design, Vitest.

## Global Constraints

- Frontend is React + Ant Design; all backend production code is under `api/` and Go only.
- `agent_projects` are free-standing facts; BatchProject and Book references are optional, non-copying, and must be access-checked for the acting user.
- MySQL is authoritative for Agent facts; TOS stores only attachment bytes; browser storage carries no business facts.
- Reuse HttpOnly Cookie Session, same-origin CSRF protection, capability checks, generation, taskruntime Queue/Lease/Lock, and existing TOS abstraction.
- Do not create a Worker, Scheduler, Queue, Runtime, Auth, Bearer flow, iframe, or `/api/chat` endpoint.
- System prompts resolve in Go from the prompt module; selected skills are server-side references.
- Without a configured real Provider, persist and return `executor_unavailable`; never fabricate an answer, asset, or success.

---

### Task 1: Durable Agent Studio contract and migration

**Files:**
- Create: `api/db/migrations/00014_task16_agent_studio.sql`
- Create: `api/internal/agentstudio/model.go`
- Create: `api/internal/agentstudio/store.go`
- Create: `api/internal/agentstudio/mysql_store.go`
- Test: `api/internal/agentstudio/mysql_store_test.go`

**Interfaces:**
- Produces `agentstudio.Store`, project/message/canvas/version/attachment/execution/skill models, and additive tables scoped by `owner_user_id` and optional `team_id`.
- Consumes `auth_users`, `auth_batch_project_ownership`, `batch_projects`, and `books` without duplicating their payloads.

- [ ] **Step 1: Write failing store tests** for owner-scoped project listing, rejected foreign BatchProject/Book links, message ordering, canvas optimistic revision conflict, version snapshot restore, and attachment ownership.
- [ ] **Step 2: Run** `go test ./internal/agentstudio -run 'TestMySQLStore'` and confirm it fails because the package is absent.
- [ ] **Step 3: Add the minimal migration**: `agent_projects`, `agent_messages`, `agent_canvas_documents`, `agent_canvas_versions`, `agent_attachments`, `agent_executions`, and `agent_skill_versions`; use foreign keys to users and optional project/book references, unique version/revision keys, and indexed owner/project lookup.
- [ ] **Step 4: Implement Store and MySQLStore** using explicit actor user/team scopes in every select, insert, update, restore, and delete predicate.
- [ ] **Step 5: Re-run the focused store tests** and keep the migration additive and reversible.

### Task 2: Agent service and truthful execution boundary

**Files:**
- Create: `api/internal/agentstudio/service.go`
- Create: `api/internal/agentstudio/prompts.go`
- Create: `api/internal/agentstudio/service_test.go`
- Modify: `api/internal/app/app.go`

**Interfaces:**
- Produces `agentstudio.Service` methods `CreateProject`, `Continue`, `SaveCanvas`, `RestoreCanvasVersion`, `UploadAttachment`, `CreateExecution`, and `RetryExecution`.
- Consumes `generation.Provider` only through an adapter; consumes existing TOS uploader and `taskruntime.Queue` only when configured.

- [ ] **Step 1: Write failing service tests** proving that the system prompt is assembled server-side, unavailable execution becomes `executor_unavailable`, messages persist before execution outcome, and retry creates an auditable new execution linked to the same existing runtime identity rather than a private queue.
- [ ] **Step 2: Run** `go test ./internal/agentstudio -run 'TestService'` and confirm the expected failure.
- [ ] **Step 3: Implement the smallest service** that resolves Go-owned prompt text, validates selected skills, saves execution lifecycle facts, and calls only the injected existing execution adapter.
- [ ] **Step 4: Wire the service in app.go** with the existing generation provider/TOS configuration and taskruntime queue; leave it explicitly unavailable when dependencies are absent.
- [ ] **Step 5: Re-run focused tests** and verify no provider response is synthesized by a fake fallback.

### Task 3: Authenticated Go HTTP API

**Files:**
- Create: `api/internal/httpapi/agent_handlers.go`
- Create: `api/internal/httpapi/agent_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`
- Modify: `api/internal/httpapi/permissions.go`
- Modify: `api/internal/app/app.go`

**Interfaces:**
- Produces `/api/v1/agent/projects`, nested messages/canvas/versions/attachments/executions/skills endpoints.
- Uses existing `requireAuth`, `requireSameOrigin`, and new narrow `agent.view`, `agent.create`, `agent.execute`, `agent.configure` capabilities.

- [ ] **Step 1: Write failing handler tests** for 401 without Cookie Session, 403 without capability, 403 for foreign project or foreign BatchProject/Book link, Same-Origin rejection for mutations, 409 canvas revision conflict, and explicit unavailable executor response.
- [ ] **Step 2: Run** `go test ./internal/httpapi -run 'TestAgent'` and confirm the handler routes do not yet exist.
- [ ] **Step 3: Implement route registration and handlers** with actor extraction from the existing auth context, bounded request parsing, response-safe errors, multipart upload to the Agent service, and no bearer-token path.
- [ ] **Step 4: Re-run focused HTTP tests** and inspect that all mutations pass through same-origin checks.

### Task 4: React Agent home, canvas, and API client

**Files:**
- Create: `前台/src/AgentStudioPage.jsx`
- Create: `前台/src/AgentCanvasPage.jsx`
- Create: `前台/src/agent-studio.css`
- Create: `前台/src/AgentStudioPage.test.jsx`
- Create: `前台/src/AgentCanvasPage.test.jsx`
- Modify: `前台/src/api.js`
- Modify: `前台/src/App.jsx`

**Interfaces:**
- Consumes only `/api/v1/agent/*` through `requestJSON` with the existing Cookie Session recovery flow.
- Provides `/agent` creation/recent-project UI and `/agent/canvas?projectId=` three-pane project/canvas/conversation UI.

- [ ] **Step 1: Write failing UI tests** covering loading/empty/error/forbidden states, create-and-continue flow, recent projects after refresh, skill selection, upload/delete attachment, execution unavailable/retry display, canvas save conflict refresh, history restore confirmation, and deep-link routing.
- [ ] **Step 2: Run** `CI=1 npm test -- --run AgentStudioPage AgentCanvasPage` in `前台` and confirm the pages are absent.
- [ ] **Step 3: Add API helpers and pages** using Ant Design components, server-reloaded project state, multipart upload with Cookie credentials, and no localStorage/sessionStorage facts.
- [ ] **Step 4: Add dedicated responsive CSS** that carries forward the V88 start screen, project drawer, conversation panel, cards, controls, and mobile collapse behavior without importing legacy bundles.
- [ ] **Step 5: Re-run focused UI tests** and confirm all state shown after navigation is fetched from Go.

### Task 5: End-to-end repository verification and delivery

**Files:**
- Modify: `TASKS.md`
- Test: `api/...`, `前台/...`

- [ ] **Step 1: Add only verified Task 16 Agent Studio items** to `TASKS.md`, explicitly retaining Provider/public-browser/deployment gaps.
- [ ] **Step 2: Run Go package tests**, then `go test ./...` from `api`.
- [ ] **Step 3: Run** `npm install --no-package-lock --prefer-offline --no-audit`, `CI=1 npm test -- --run`, and `npm run build` from `前台`.
- [ ] **Step 4: Inspect diff/status**, commit the implementation with a `[skip ci]` message, and record the SHA.
- [ ] **Step 5: Run** `git pull --rebase origin main`; on conflict stop and preserve unrelated work; otherwise push `main` and verify `HEAD` equals `origin/main`.

## Self-review

- The plan has a dedicated task for every requested durable fact: projects/messages, skills, TOS attachments, canvas/version restore, executions/retry, authorization, and responsive UI.
- BatchProject/Book are references only; the migration never copies their business data.
- Runtime execution is explicitly constrained to existing generation/taskruntime and unavailable dependencies are recorded truthfully.
- No plan task creates a Node backend, iframe, bearer token, browser business storage, second queue/worker, or `/api/chat` route.
