# Batch Factory V12 Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the active Batch Factory and Agent integration use V12 names and public routes while keeping existing V12 data readable and retaining only a narrow internal compatibility layer.

**Architecture:** Add canonical V12 frontend API naming over the already-V12 HTTP contract, move the Agent binding endpoint to V12 with a temporary old-route adapter, and remove old public exposure from the normal runtime. Keep `batchfactoryv11` internal types/tables until a separately verified database migration exists.

**Tech Stack:** React, Go, `net/http`, MySQL-backed Go stores, Go `embed` frontend runtime.

**Spec:** This user request: migrate the active Batch Factory surface into V12 without breaking existing batches.

## Global Constraints

- V12 is the only normal public Batch Factory API namespace.
- Existing batches must remain readable.
- No database table deletion or destructive migration in this pass.
- Legacy adapters may remain internal and must not be loaded by the normal public route.
- Verify with focused tests, frontend build, embedded Go rebuild, and local health check.

## Review Focus

- Existing V12 batch read/write operations continue to resolve through the same store.
- Agent binding uses V12 first and does not require the V11 public endpoint.
- Old V11 endpoint remains available only where an existing compatibility test or adapter explicitly requires it.
- The active `/batch-factory` page does not import the retired V11 workbench page.
- A rebuilt local Go runtime serves the new frontend assets.

### Task 1: Canonical V12 API naming

**Files:**
- Create: `frontend/src/shared/api/batchFactory.js`
- Modify: active Batch Factory imports that currently import `batchFactoryV11.js`
- Test: `frontend/src/shared/api/batchFactory.test.js`

- [ ] Add a V12-named API facade that exports the existing V12-backed functions without changing request payloads.
- [ ] Update active page imports to use the facade.
- [ ] Verify the facade and active imports do not create V11 HTTP paths.

### Task 2: Agent binding V12 route

**Files:**
- Modify: `backend/internal/httpapi/agent_batch_factory_bridge.go`
- Modify: `frontend/src/shared/api/batchFactory.js`
- Test: `backend/internal/httpapi/agent_batch_factory_bridge_test.go`

- [ ] Register `POST /api/batch-factory/v12/batches/{batchId}/agent-project` as the canonical route.
- [ ] Keep the V11 route only as a narrow compatibility adapter during this pass.
- [ ] Verify both paths resolve to the same owner-scoped binding behavior.

### Task 3: Public runtime cleanup

**Files:**
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/httpapi/legacy_v11_visibility.go`
- Test: `backend/internal/httpapi/legacy_v11_visibility_test.go`

- [ ] Ensure normal local and production runtime registration exposes V12 only.
- [ ] Keep internal V11 mux construction only where V12 adapters still depend on it.
- [ ] Verify the public V11 namespace is not enabled by default.

### Task 4: Build and runtime verification

**Files:**
- Modify: generated embedded frontend artifacts through the existing build script only.

- [ ] Run focused frontend and backend tests.
- [ ] Build frontend and refresh Go embedded assets.
- [ ] Restart the confirmed local Go runtime.
- [ ] Verify `/health` and `/api/runtime-build-info`.
