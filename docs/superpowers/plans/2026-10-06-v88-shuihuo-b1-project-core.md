# V88 Shuihuo B1 Project Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the ordinary v88 Shuihuo project/source/segmentation/segment core on the new Go/MySQL mainline without creating duplicate assets, providers, storage or runtimes.

**Architecture:** Add a focused `api/internal/shuihuo` domain with project/segment models, a MySQL store, service-level ownership checks, pure segmentation logic and an injected smart-segmentation consumer. Mount it behind the existing `httpapi` Auth/capability/same-origin middleware and wire it from `internal/app` using the already-wired shared text provider only as a consumer.

**Tech Stack:** Go 1.23, MySQL, Goose, existing `httpapi` auth/capability middleware.

**Spec:** `docs/superpowers/specs/2026-10-06-v88-shuihuo-b1-project-core-design.md`

## Global Constraints

- Branch: `feat/v88-parity-shuihuo` from the true latest `origin/main` at task start.
- Do not modify global Router/UserLayout/Sidebar/Header/Account Center/AuthBoundary architecture.
- Do not implement Batch Factory V11; task C owns it.
- Do not create Shuihuo-private assets/media/model/provider/credential/runtime systems.
- Do not create a second Scheduler, Redis queue, VIDEO runtime or provider registry.
- MySQL/Go is the project/source/segment fact source; LocalStorage is forbidden for these facts.
- Do not trigger GitHub Actions; commit messages use `[skip ci]`.
- Do not merge.

## Review Focus

- Cross-user/team project access must not leak or mutate another owner's project.
- Replacing source text must invalidate stale segmentation atomically.
- Confirm/reorder/delete must preserve contiguous 1-based segment positions.
- Smart segmentation unavailability must be explicit and must not create a private provider fallback.
- Imported v88 segmentation rows such as `1<TAB>text` must round-trip without keeping the numeric prefix.

---

### Task 1: RED domain contract

**Files:**
- Create: `api/internal/shuihuo/service_test.go`

**Interfaces:**
- Consumes: none.
- Produces: executable expectations for project lifecycle, ownership, four segmentation modes, confirm and segment CRUD/reorder.

- [x] Write failing tests for project lifecycle/ownership.
- [x] Write failing tests for paragraph/fixed/import/smart segmentation and confirm.
- [x] Write failing tests for segment CRUD/reorder.
- [x] Run isolated `go test ./internal/shuihuo` and verify failure is caused by missing B1 implementation.

### Task 2: Project/segment domain and service

**Files:**
- Create: `api/internal/shuihuo/model.go`
- Create: `api/internal/shuihuo/store.go`
- Create: `api/internal/shuihuo/service.go`
- Create: `api/internal/shuihuo/segmentation.go`

**Interfaces:**
- Consumes: injected `Store` and optional `SmartSegmenter`.
- Produces: `Service` methods for create/list/read/delete/source, segmentation, confirm and segment CRUD/reorder.

- [ ] Implement models/errors/ownership policy.
- [ ] Implement source selection and paragraph/fixed/import segmentation.
- [ ] Implement service authorization and project read models.
- [ ] Implement smart segmentation through the narrow injected interface.
- [ ] Run isolated core tests to GREEN.

### Task 3: MySQL persistence and Goose schema

**Files:**
- Create: `api/db/migrations/00009_shuihuo_project_core.sql`
- Create: `api/internal/shuihuo/mysql_store.go`
- Create: `api/internal/shuihuo/migration_contract_test.go`

**Interfaces:**
- Consumes: `Store` contract from Task 2.
- Produces: durable project/segment persistence and transactional source reset/confirm/reorder/delete semantics.

- [ ] Add `shuihuo_projects` and `shuihuo_segments` only.
- [ ] Include ownership/team indexes and cascade segment deletion.
- [ ] Implement MySQL store with transactional segment replacement/reorder/compaction.
- [ ] Add migration contract tests.
- [ ] Run isolated Shuihuo tests to GREEN.

### Task 4: Shared text provider adapter for smart segmentation

**Files:**
- Create: `api/internal/shuihuo/smart_segmenter.go`
- Create: `api/internal/shuihuo/smart_segmenter_test.go`

**Interfaces:**
- Consumes: existing `generation.Provider`-compatible `Complete` contract.
- Produces: `SmartSegmenter` only; no model/provider ownership.

- [ ] Write RED parsing/unavailable tests.
- [ ] Implement JSON-first parsing with safe line fallback.
- [ ] Map shared generation unavailability to `ErrSmartUnavailable`.
- [ ] Run tests to GREEN.

### Task 5: HTTP API and existing Auth boundary

**Files:**
- Create: `api/internal/httpapi/shuihuo_handlers.go`
- Modify: `api/internal/httpapi/server.go`
- Create: `api/internal/httpapi/shuihuo_handlers_test.go`

**Interfaces:**
- Consumes: `shuihuo.Service`, existing `authn.CurrentUser`, `requireCapability`, `requireSameOrigin`, `parsePositiveID`, `decodeJSON`.
- Produces: `/api/v1/shuihuo-production/*` B1 routes.

- [ ] Add handler service interface/dependency.
- [ ] Mount read routes behind existing authenticated capability middleware.
- [ ] Mount mutations behind existing same-origin + capability middleware.
- [ ] Map invalid/not-found/forbidden/smart-unavailable errors to stable HTTP statuses.
- [ ] Add ownership and basic contract tests.

### Task 6: Application wiring and verification

**Files:**
- Modify: `api/internal/app/app.go`

**Interfaces:**
- Consumes: `shuihuo.NewMySQLStore`, `shuihuo.NewService`, shared text `generation.Provider` adapter.
- Produces: production B1 service in the existing Go server.

- [ ] Wire store/service using the existing DB.
- [ ] Reuse the existing shared text provider for smart segmentation; add no provider config.
- [ ] Run `go test ./internal/shuihuo`.
- [ ] Run repository-level relevant Go tests when a full repository execution environment is available.
- [ ] Do not merge or run Actions.
