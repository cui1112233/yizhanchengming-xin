# Task 9 Batch Factory V11 Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Task 9.2, then 9.3, then 9.4 on the new Go/React/MySQL mainline while selectively reusing validated PR #11 work instead of reimplementing it.

**Architecture:** BatchProject and book facts come from MySQL through Go stores/API. React consumes deterministic API contracts and renders Ant Design components. Redis is introduced only in Task 9.4 for queue/lock/lease infrastructure; MySQL remains durable truth.

**Tech Stack:** Go, React, Ant Design, MySQL, Goose, Redis, Vitest/Testing Library, Go sqlmock/store tests.

**Spec:** `docs/superpowers/specs/2026-10-05-task9-batch-factory-workbench-design.md`

## Global Constraints

- Work only on `feat/task9-batch-factory-workbench`, created from latest `main`.
- Audit PR #11 (`task9-batch-project-source-ui`) and reuse valid file/commit-level Task 9.2 changes; do not wholesale merge its divergent branch.
- Backend/API/Worker/Scheduler must be Go.
- Frontend must remain React + Ant Design.
- MySQL is the durable business source of truth; Goose owns schema migrations.
- Redis is only Queue/Lock/short-lived runtime infrastructure.
- Do not implement Task 10–15.
- Do not use LocalStorage or frontend state as the BatchProject business fact source.
- Follow RED -> verify expected failure -> GREEN -> full regression for every new behavior.
- Do not mark unfinished acceptance items `[x]`; use `🟡`/`⚠️` until repository completion rules are met.

## Review Focus

- Multi-source / multi-gender / multi-style projects must remain deterministic and never collapse to the first book.
- Latest run tie-breaking must use `run_at DESC, id DESC`.
- A Task 8-created project must appear on the next `/batch-factory` load without cache participation.
- Testing Library cleanup must prevent stale DOM from producing false duplicate rows or false positives.
- A single failed book in Task 9.4 must persist its error without blocking unrelated books.

---

### Task 1: Port validated PR #11 summary behavior onto latest main

**Files:**
- Modify: `api/internal/intake/model.go`
- Modify: `api/internal/intake/mysql_store.go`
- Modify: `api/internal/httpapi/intake_handlers.go`
- Modify: `api/internal/httpapi/batch_project_handlers.go`
- Modify: `api/internal/app/app_test.go`
- Modify: `api/internal/httpapi/batch_project_handlers_test.go`
- Create/port: `api/internal/intake/batch_project_summary_test.go`
- Modify: `前台/src/BatchProjectListPage.jsx`
- Modify: `前台/src/App.test.jsx`

**Interfaces:**
- Produces: `intake.BatchProject` summary fields `Sources []string`, `BookCount int`, `Genders []string`, `Styles []string`, `RunStatus RunStatus`.
- Produces: `GET /api/v1/batch-projects` JSON fields `sources`, `bookCount`, `genders`, `styles`, `runStatus`.

- [ ] **Step 1: Port/add the PR #11 store/API/frontend tests first, adjusted only for current-main file structure.**

- [ ] **Step 2: Run the focused Go/frontend tests and verify RED is caused by missing summary fields/UI, not syntax or stale test harness state.**

- [ ] **Step 3: Port the minimal compatible production diffs from PR #11.**

Store query requirements: aggregate distinct trimmed source/gender/style values, count books, and select latest run by `ORDER BY r.run_at DESC, r.id DESC LIMIT 1`.

- [ ] **Step 4: Run focused tests and then the Go/frontend suites to verify GREEN.**

- [ ] **Step 5: Commit the validated 9.2.4–9.2.8 port.**

### Task 2: Task 9.2.9 real-time MySQL visibility

**Files:**
- Modify: `api/internal/intake/mysql_store_test.go` or add a focused batch-project visibility test beside the store.
- Modify: `api/internal/httpapi/batch_project_handlers_test.go` and/or app wiring test.
- Modify: `前台/src/App.test.jsx` or create `前台/src/BatchProjectListPage.test.jsx` if isolation is clearer.
- Modify production only if a failing test proves a real gap.

**Interfaces:**
- Consumes: Task 1 `GET /api/v1/batch-projects` contract.
- Produces: `/batch-factory` performs a fresh list request on mount and renders the newly persisted project.

- [ ] **Step 1: Add a frontend test whose test-local fetch/mock sequence represents Task 8 create success followed by a fresh `/batch-factory` mount; assert a new `GET /api/v1/batch-projects` is issued and the returned new project renders. Ensure explicit Testing Library cleanup or isolated render lifecycle.**

- [ ] **Step 2: Add/confirm Go store/API coverage showing each list call executes against MySQL and returns the newly present project; no process cache layer is involved.**

- [ ] **Step 3: Run focused tests and verify the intended RED. If RED is only stale DOM cleanup, use systematic-debugging and fix the test harness rather than rewriting working business logic.**

- [ ] **Step 4: Make the smallest production/test-harness change required for GREEN.**

- [ ] **Step 5: Run Go tests, frontend tests, frontend build, admin build.**

- [ ] **Step 6: Commit Task 9.2.9.**

### Task 3: Task 9.3 project detail API/store

**Files:**
- Modify: `api/internal/intake/store.go`
- Modify: `api/internal/intake/mysql_store.go`
- Add tests under `api/internal/intake/`
- Modify/create `api/internal/httpapi/batch_project_handlers.go` detail handler
- Modify `api/internal/httpapi/server.go` route registration if needed
- Add API contract tests under `api/internal/httpapi/`

**Interfaces:**
- Produces: `GET /api/v1/batch-projects/{id}` returning real project + all books.
- Book fields: internal id, external Book ID, title, source, platformId, gender, style, status, errorMessage.

- [ ] **Step 1: Write MySQL store tests for project lookup and complete book list, including failed book with persisted error.**
- [ ] **Step 2: Run and verify RED.**
- [ ] **Step 3: Implement minimal store methods.**
- [ ] **Step 4: Write API contract test for detail response.**
- [ ] **Step 5: Run and verify RED for handler/route.**
- [ ] **Step 6: Implement minimal Go handler/route.**
- [ ] **Step 7: Run focused + full Go tests.**
- [ ] **Step 8: Commit detail backend.**

### Task 4: Task 9.3 React V11 project detail / novel list

**Files:**
- Modify: `前台/src/api.js`
- Modify: `前台/src/BatchProjectListPage.jsx`
- Create focused detail component/test if that keeps responsibilities smaller.
- Modify routing in `前台/src/App.jsx` only as required.

**Interfaces:**
- Consumes: Task 3 detail API.
- Produces: clicking a project enters the workbench/detail view and renders every real book with Book ID, title, source, platformId, gender, style, fetch status, and real error.

- [ ] **Step 1: Write frontend tests for click -> detail fetch -> all required fields, including a failed book error.**
- [ ] **Step 2: Run and verify RED.**
- [ ] **Step 3: Implement minimal React/Ant Design detail view without mock business data.**
- [ ] **Step 4: Run focused and full frontend tests/build.**
- [ ] **Step 5: Commit Task 9.3 frontend.**

### Task 5: Task 9.4 durable per-book execution model

**Files:**
- Add Goose migration(s) for per-book run state/idempotency fields/tables as dictated by failing tests.
- Add focused Go domain/store files under `api/internal/` following existing package conventions.
- Add MySQL store tests first.

**Interfaces:**
- Produces: durable run/book-run transitions, per-book attempts/errors, idempotency key/constraint.

- [ ] **Step 1: Write failing tests for pending->running, independent book states, failure isolation, persisted errors, single-book retry, and duplicate-start rejection/idempotency.**
- [ ] **Step 2: Verify RED.**
- [ ] **Step 3: Add minimal Goose schema + Go store/service implementation.**
- [ ] **Step 4: Run focused + full Go tests.**
- [ ] **Step 5: Commit durable execution model.**

### Task 6: Task 9.4 Scheduler + Redis Queue/Lock + recovery

**Files:**
- Add Go scheduler/worker/queue/lock packages under `api/internal/` using repository conventions.
- Modify app/server wiring minimally.
- Add unit/integration tests with Redis abstraction/fakes where appropriate; preserve real behavior assertions at service boundaries.

**Interfaces:**
- Consumes: Task 5 durable run/book-run state.
- Produces: due-run scheduling, queue claim, distributed lease, worker execution, expired-lease recovery.

- [ ] **Step 1: Write failing tests for due-run enqueue, lock exclusion, crash/lease expiry recovery, and restart-safe requeue from MySQL state.**
- [ ] **Step 2: Verify RED.**
- [ ] **Step 3: Implement minimal Go Scheduler/Queue/Lock/Worker.**
- [ ] **Step 4: Run focused + full Go tests.**
- [ ] **Step 5: Commit Task 9.4 infrastructure.**

### Task 7: Acceptance, TASKS.md status, PR

**Files:**
- Modify: `TASKS.md`
- PR description only for shared-interface notes and verification evidence.

**Interfaces:**
- Consumes: Tasks 1–6.
- Produces: auditable Task 9 branch/PR without premature completion claims.

- [ ] **Step 1: Run all Go tests, frontend tests, frontend build, admin build, API contract tests, MySQL store tests.**
- [ ] **Step 2: Load `superpowers:verification-before-completion` and verify evidence before any completion claim.**
- [ ] **Step 3: Update Task 9 entries using `🟡`/`⚠️` where CI/main/ECS acceptance is still pending; do not mark `[x]` contrary to TASKS.md completion rules.**
- [ ] **Step 4: Open an independent PR from `feat/task9-batch-factory-workbench` to `main`, documenting PR #11 file-level reuse and any minimal shared-interface changes.**
- [ ] **Step 5: Verify PR CI and report any remaining acceptance items explicitly.**
