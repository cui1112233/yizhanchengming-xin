# Module Boundary Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish enforceable feature-module boundaries and migrate Novel Fetch as the first complete module without changing public behavior.

**Architecture:** Shared infrastructure lives behind service interfaces. Each feature owns its HTTP registration, task flow, state, persistence adapter, frontend page, and tests; the application shell only assembles modules.

**Tech Stack:** Go 1.23, `net/http`, React, Ant Design, Node test runner, MySQL, Redis, Go embed.

## Global Constraints

- Only service code may be imported by multiple feature modules.
- Feature modules must not import another feature module's internal code.
- Public URLs, JSON contracts, database tables, and exact Chinese UI text stay unchanged during migration.
- System prompts remain backend-only.
- Work directly on `v88`; do not create extra branches.
- Preserve unrelated dirty files and commit only task paths.

---

### Task 1: Add executable dependency-boundary checks

**Files:**
- Create: `scripts/check-module-boundaries.mjs`
- Create: `tests/module-boundaries.test.js`
- Create: `docs/architecture/module-boundaries.json`

**Interfaces:**
- Consumes: Go imports and JavaScript imports under declared module roots.
- Produces: non-zero exit with source file, forbidden import, and owning module.

- [ ] Write a fixture test with `module-a` importing `module-b/internal` and verify failure.
- [ ] Write a fixture where both modules import `service/media` and verify success.
- [ ] Implement import parsing for Go, JS, and JSX literals.
- [ ] Add an explicit migration baseline for current violations; new violations fail immediately.
- [ ] Run the fixture tests and repository scan.
- [ ] Commit.

### Task 2: Create the Novel Fetch frontend module entry

**Files:**
- Create: `frontend/src/modules/novel-fetch/index.jsx`
- Create: `frontend/src/modules/novel-fetch/NovelFetchModule.jsx`
- Modify: the existing user application route that currently imports `user/pages/NovelFetchPage.jsx`
- Test: `frontend/src/modules/novel-fetch/NovelFetchModule.test.js`

**Interfaces:**
- Consumes: Novel Fetch service client and application-shell auth context.
- Produces: default `NovelFetchModule` route component for `/novel-fetch`.

- [ ] Write a route test proving `/novel-fetch` resolves through the module entry.
- [ ] Verify it fails while the shell imports the legacy page directly.
- [ ] Add the module entry as a temporary adapter around the existing page.
- [ ] Change only the shell import; preserve rendered behavior.
- [ ] Run Novel Fetch focused tests and frontend build.
- [ ] Commit.

### Task 3: Move Novel Fetch API clients into shared service

**Files:**
- Create: `frontend/src/service/novel-fetch/workshop.js`
- Create: `frontend/src/service/novel-fetch/submission.js`
- Modify: `frontend/src/shared/api/novelFetchWorkshop.js`
- Modify: `frontend/src/shared/api/novelFetch.js`
- Test: `frontend/src/service/novel-fetch/novelFetchService.test.js`

**Interfaces:**
- Produces: the existing exported API functions with unchanged paths and payloads.

- [ ] Copy contract tests to the new service entry and verify they fail before exports exist.
- [ ] Move implementations; leave old paths as temporary re-export adapters.
- [ ] Update the new feature module to import only from `src/service`.
- [ ] Run API contract and account-scope tests.
- [ ] Commit.

### Task 4: Extract Novel Fetch route assembly from the global router

**Files:**
- Create: `backend/internal/httpapi/novelfetch_module_routes.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/httpapi/novelfetch_module_routes_test.go`

**Interfaces:**
- Consumes: a `NovelFetchModuleOptions` value containing store, queue, model resolvers, fetcher, auth, user data path, and video-management connector.
- Produces: `registerNovelFetchModule(root *http.ServeMux, options NovelFetchModuleOptions)`.

- [ ] Write a route-table test against the existing canonical and compatibility URLs.
- [ ] Verify failure before the module registrar exists.
- [ ] Move only registration and alias wiring out of `router.go`; keep handlers unchanged.
- [ ] Replace the global inline block with one registrar call.
- [ ] Run all `httpapi` Novel Fetch tests and full Go tests.
- [ ] Commit.

### Task 5: Move backend handlers behind the Novel Fetch module package

**Files:**
- Create: `api/modules/novel-fetch/transport/`
- Create: `api/modules/novel-fetch/application/`
- Create: `api/service/` interfaces needed by Novel Fetch
- Modify: Novel Fetch handler files currently under `backend/internal/httpapi`
- Modify: `backend/go.mod` only if the final Go module root changes in a separately reviewed step.

**Interfaces:**
- Produces: `novelfetch.Module.RegisterRoutes`, application commands/queries, and service ports.

- [ ] Inventory handler files and group them by route contract.
- [ ] Add characterization tests for each route group before moving it.
- [ ] Move one route group at a time; no cross-feature handler moves in the same commit.
- [ ] Replace direct imports of other feature internals with service interfaces.
- [ ] Run focused and full Go tests after every route group.
- [ ] Commit each independently testable route group.

### Task 6: Split the Novel Fetch page

**Files:**
- Create focused components under `frontend/src/modules/novel-fetch/` for input, queue, task list, task detail, configuration, and knowledge library.
- Reduce: `frontend/src/user/pages/NovelFetchPage.jsx` to a temporary re-export, then remove it after all imports migrate.

- [ ] Characterize current visible states and button rules.
- [ ] Extract one component with its state hook and tests per commit.
- [ ] Keep state ownership inside the Novel Fetch module.
- [ ] Run focused browser/source tests and frontend build after each extraction.
- [ ] Verify local Go-embedded runtime on port 4001 before any production release.

