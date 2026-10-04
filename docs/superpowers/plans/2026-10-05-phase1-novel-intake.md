# Phase 1 Novel Intake Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first clean, production-oriented slice of the new mainline: add books grouped by bookstore/source, create an intake, fetch 121 content/bookinfo, resolve gender/style metadata, create a batch project, persist authoritative state in MySQL, and expose it to the React frontend.

**Architecture:** The new repository is a clean Go-first mainline, not a wholesale copy of V88. The Go API owns business state and pipeline orchestration; MySQL is the source of truth, Redis is reserved for queue/lock runtime, and React + Ant Design is build-time UI only. V88 is used as a behavior/reference source, including its current Go packages (`backend/internal/novelfetchworkshop`, `backend/internal/batchfactoryv11`) and the latest verified fixes, but legacy Node runtime code is not copied into the new production path.

**Tech Stack:** Go, MySQL, Goose migrations, Redis, React, Ant Design, Vite, Go embed (packaging stage), GitHub Actions.

**Spec:** `README.md`

## Global Constraints

- Production backend is Go.
- React + Ant Design is frontend only; Node/npm is build-time tooling, not a production service.
- Business path is React → Go API → MySQL/Redis → Go Worker → external API/TOS.
- MySQL is the only business/task source of truth; Redis is queue/lock/short-lived runtime only.
- Immediate execution and automation use the same pipeline; only `run_at` differs.
- 121 uses official HTTP API as the normal path; Browser Worker is not the default business path.
- Gender priority is: manual user value > 121 `bookinfo.category` > verified `genre` mapping > AI fallback.
- Deterministic metadata must never be overwritten by AI.
- Migration source baseline is V88, currently newer than the new repository; do not assume the October 3 README commit contains current application code.

## Review Focus

- 121 timeouts/non-2xx/invalid responses must not leave an intake falsely marked successful; retryable state must be explicit.
- Duplicate submission of the same source + book identifier must be idempotent rather than create duplicate books/projects.
- Mixed bookstore batches must preserve each book's source label through intake and batch-project creation.
- Metadata conflict resolution must preserve manual/category values and must not allow lower-priority AI output to overwrite them.
- Scheduled and immediate execution must call the same pipeline and differ only in `run_at` scheduling state.

---

### Task 1: Establish the clean repository skeleton and CI

**Files:**
- Create: `.gitignore`
- Create: `.github/workflows/ci.yml`
- Create: `api/go.mod`
- Create: `api/cmd/server/main.go`
- Create: `api/internal/httpapi/server.go`
- Create: `api/internal/httpapi/server_test.go`
- Create: `后台/package.json`
- Create: `后台/src/main.jsx`
- Create: `后台/index.html`
- Create: `前台/package.json`
- Create: `前台/src/main.jsx`
- Create: `前台/index.html`

**Interfaces:**
- Produces: Go API entrypoint and `GET /healthz`; buildable admin/user React shells.

- [ ] Write `api/internal/httpapi/server_test.go` asserting `GET /healthz` returns HTTP 200 and JSON `{ "status": "ok" }`.
- [ ] Run `go test ./...` in `api`; expect RED because the HTTP server package/handler is not implemented yet.
- [ ] Implement the minimum `httpapi.NewHandler()` and `cmd/server` entrypoint.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Add minimal React + Ant Design shells for `后台/` and `前台/` and a CI workflow that runs Go tests plus both frontend builds.
- [ ] Commit: `chore: scaffold unified Go mainline`.

### Task 2: Create authoritative MySQL schema and store contracts

**Files:**
- Create: `api/db/migrations/00001_phase1_intake.sql`
- Create: `api/internal/intake/model.go`
- Create: `api/internal/intake/store.go`
- Create: `api/internal/intake/mysql_store.go`
- Create: `api/internal/intake/mysql_store_test.go`

**Interfaces:**
- Produces: `Intake`, `Book`, `BatchProject`, `Run` persistence contracts and MySQL implementation.

- [ ] Write store tests for create/read intake, idempotent source+book key, metadata columns, batch project creation, and scheduled `run_at`.
- [ ] Run the tests; expect RED because schema/store implementation is absent.
- [ ] Add Goose migration for `intakes`, `books`, `batch_projects`, and `runs`, including unique `(source, external_book_id)` identity.
- [ ] Implement MySQL store methods required by the tests.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Commit: `feat: add intake persistence model`.

### Task 3: Implement deterministic metadata priority

**Files:**
- Create: `api/internal/metadata/resolver.go`
- Create: `api/internal/metadata/resolver_test.go`

**Interfaces:**
- Produces: `ResolveGender(manual, category, genre, ai string) (value string, source string)` and style metadata merge helpers.

- [ ] Write table tests proving priority `manual > category > verified genre mapping > AI` and proving lower-priority values never overwrite a resolved higher-priority value.
- [ ] Run metadata tests; expect RED because resolver does not exist.
- [ ] Implement minimal resolver and an explicit verified genre map.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Commit: `feat: add deterministic metadata resolution`.

### Task 4: Add the 121 HTTP gateway behind an interface

**Files:**
- Create: `api/internal/provider121/client.go`
- Create: `api/internal/provider121/client_test.go`

**Interfaces:**
- Produces: `Client.Fetch(ctx, Request) (Result, error)` where `Result` contains normalized body and bookinfo/category/genre fields.
- Consumes: deployment-supplied HTTPS endpoint/API key; no credential logging.

- [ ] Write HTTP-server-backed tests for success, timeout, non-2xx, invalid JSON, and missing required fields.
- [ ] Run provider tests; expect RED because client is absent.
- [ ] Implement an HTTPS-only, timeout-bounded client modeled on the audited V88 external provider transport contract.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Commit: `feat: add 121 HTTP provider`.

### Task 5: Build the intake pipeline service

**Files:**
- Create: `api/internal/intake/service.go`
- Create: `api/internal/intake/service_test.go`

**Interfaces:**
- Consumes: intake store, 121 client, metadata resolver.
- Produces: `CreateIntake`, `ExecuteIntake`, and normalized per-book status transitions.

- [ ] Write service tests covering multiple bookstore groups, duplicate books, partial provider failure, successful metadata resolution, and retryable failure state.
- [ ] Run service tests; expect RED.
- [ ] Implement the minimum orchestration that stores intake first, fetches 121 data, resolves metadata, and commits per-book status without marking failed work successful.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Commit: `feat: implement novel intake pipeline`.

### Task 6: Create batch projects and unify immediate/scheduled execution

**Files:**
- Create: `api/internal/pipeline/service.go`
- Create: `api/internal/pipeline/service_test.go`

**Interfaces:**
- Consumes: completed intake/books and store.
- Produces: batch project plus `Run{RunAt}`; immediate execution is `run_at <= now`, automation is future `run_at`.

- [ ] Write tests proving immediate and scheduled requests call the same creation path and only differ in `run_at`.
- [ ] Run pipeline tests; expect RED.
- [ ] Implement the shared pipeline service and persistence calls.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Commit: `feat: unify batch project execution scheduling`.

### Task 7: Expose Phase 1 HTTP API

**Files:**
- Create: `api/internal/httpapi/intake_handlers.go`
- Create: `api/internal/httpapi/intake_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`

**Interfaces:**
- Produces: endpoints for creating/listing intakes, executing intake, listing books/results, and creating immediate/scheduled batch projects.

- [ ] Write handler tests for validation, idempotent book submission, provider failure mapping, and successful result payloads.
- [ ] Run handler tests; expect RED.
- [ ] Implement handlers and route registration without adding Node middleware/services.
- [ ] Run `go test ./...`; expect PASS.
- [ ] Commit: `feat: expose phase1 intake API`.

### Task 8: Implement the user-facing bookstore batching flow

**Files:**
- Create: `前台/src/api/intake.js`
- Create: `前台/src/pages/NovelIntakePage.jsx`
- Create: `前台/src/pages/NovelIntakePage.test.jsx`
- Modify: `前台/src/main.jsx`

**Interfaces:**
- Consumes: Phase 1 Go HTTP API.
- Produces: bookstore/source tabs, book input, add-bookstore behavior, selected-source counts, immediate execution and automation execution entry points.

- [ ] Write UI tests for adding books under separate bookstore/source labels, clearing the input after adding a group, preserving counts, and sending the correct execution mode.
- [ ] Run frontend tests; expect RED.
- [ ] Implement the minimum React + Ant Design flow matching the current product behavior without importing legacy `dist` assets.
- [ ] Run frontend tests and build; expect PASS.
- [ ] Commit: `feat: add bookstore intake workbench`.

### Task 9: Final verification and migration record

**Files:**
- Create: `docs/migration/phase1-v88-source-map.md`
- Modify: `README.md`

**Interfaces:**
- Produces: auditable record of which V88 packages/behaviors were migrated and which remain reference-only.

- [ ] Run `go test ./...` in `api`.
- [ ] Run test/build commands for `后台` and `前台`.
- [ ] Verify CI is green on the feature branch.
- [ ] Record V88 source baseline and migrated behavior, explicitly excluding legacy Node runtime and `frontend/dist`.
- [ ] Commit: `docs: record phase1 migration baseline`.
