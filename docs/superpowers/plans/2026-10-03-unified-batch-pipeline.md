# Unified Batch Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first clean Go-only production path for Intake → 121 fetch → deterministic metadata resolution → Batch creation, with one shared pipeline for immediate and scheduled execution.

**Architecture:** React talks only to one Go API. MySQL is the durable source of truth, Redis is only the execution queue/runtime coordination layer, and one Go worker executes pipeline stages. 121 integration is a first-class HTTP client using the supplied 121 endpoints; browser automation is not the normal path.

**Tech Stack:** Go 1.23+, MySQL, Goose migrations, Redis, React + Ant Design (frontend later), GitHub Actions.

**Spec:** Conversation-approved design on 2026-10-03: single Go mainline; 121 official HTTP API; gender precedence = manual > category > verified genre mapping > AI fallback; deterministic metadata is never overwritten by AI.

## Global Constraints

- Production backend/business/task logic uses Go.
- Node/npm is build-time only for React and must not be a production server.
- MySQL stores durable business/task state; Redis is queue/lock/transient state only.
- Immediate execution and automation use the same pipeline and differ only by `run_at`.
- 121 normal submission/fetch path uses HTTP APIs rather than Browser Worker.
- User-entered gender has highest priority and is never overwritten.
- Category mappings recognize explicit male markers: 男生/男频/男性/男向.
- Category mappings recognize explicit female markers: 女生/女频/女性/女向.
- Genre mapping is used only when it has been explicitly verified; unknown genres do not guess.
- AI classification is fallback-only when deterministic metadata remains missing.

## Review Focus

- Conflicting category text must not silently choose a gender; it should remain unresolved for fallback.
- Empty/malformed 121 responses must fail the fetch stage without creating fake metadata.
- A manually supplied gender must survive every later enrichment stage.
- Redis loss must not lose pipeline truth; unfinished MySQL jobs must be requeueable.
- Duplicate Book IDs from different source groups must have explicit duplicate/version semantics rather than accidental overwrite.

---

### Task 1: Deterministic gender resolver

**Files:**
- Create: `api/internal/novel/gender.go`
- Test: `api/internal/novel/gender_test.go`

**Interfaces:**
- Produces: `ResolveGender(Input) Result`, with source values `manual`, `121_category`, `121_genre`, `unresolved`.

- [ ] Write tests for manual precedence, explicit male/female category markers, ambiguous category, verified genre fallback, and unknown genre.
- [ ] Run tests and verify RED.
- [ ] Implement the smallest resolver that passes.
- [ ] Run package and full Go tests and verify GREEN.

### Task 2: 121 fetch client and response model

**Files:**
- Create: `api/internal/integrations/121/client.go`
- Create: `api/internal/integrations/121/types.go`
- Test: `api/internal/integrations/121/client_test.go`

**Interfaces:**
- Produces: `Client.FetchBook(ctx, bookID, platformID string, maxTxt int) (BookResponse, error)`.

- [ ] Write HTTP-server tests for query construction, JSON decoding, upstream failures, and missing正文.
- [ ] Verify RED.
- [ ] Implement HTTP client with timeout and bounded response body.
- [ ] Verify GREEN.

### Task 3: Shared pipeline state model

**Files:**
- Create: `api/internal/pipeline/types.go`
- Create: `api/internal/pipeline/runner.go`
- Test: `api/internal/pipeline/runner_test.go`

**Interfaces:**
- Produces stages `fetch_book`, `resolve_metadata`, conditional `ai_classify`, `create_batch` and a single `RunAt` field for immediate/scheduled execution.

- [ ] Test that deterministic gender skips AI fallback.
- [ ] Test unresolved metadata requests AI fallback.
- [ ] Test immediate and scheduled jobs share the same stage sequence.
- [ ] Verify RED, implement, verify GREEN.

### Task 4: Durable schema and repositories

**Files:**
- Create: `api/migrations/00001_init.sql`
- Create: `api/internal/storage/mysql.go`
- Create: repository packages for intake, books, batches, pipeline jobs.
- Tests: repository tests using sqlmock where appropriate.

**Interfaces:**
- MySQL is authoritative for intake groups, books, source metadata, batches, pipeline jobs/stages, retries, errors, and outputs.

- [ ] Write repository contract tests first.
- [ ] Implement Goose-compatible schema and repositories.
- [ ] Verify full Go suite.

### Task 5: Redis queue adapter and worker

**Files:**
- Create: `api/internal/queue/redis.go`
- Create: `api/cmd/worker/main.go`
- Tests: queue serialization/idempotency tests.

**Interfaces:**
- Worker receives only durable job IDs and reloads job truth from MySQL.

- [ ] Test idempotent enqueue payload and recovery behavior.
- [ ] Implement queue adapter and worker dispatch.
- [ ] Verify full Go suite.

### Task 6: Go API server endpoints

**Files:**
- Create: `api/cmd/server/main.go`
- Create: `api/internal/httpapi/router.go`
- Tests: HTTP handler tests.

**Interfaces:**
- Public path is `/api/batch-factory` with no V11/V12 compatibility hop.

- [ ] Test create intake, execute-now, schedule, and batch read endpoints.
- [ ] Implement handlers over the same pipeline service.
- [ ] Verify full Go suite.

### Task 7: CI and first vertical-slice verification

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] CI runs `go test ./...` from `api/`.
- [ ] Verify failing test commit is actually red before production implementation.
- [ ] Verify final branch CI is green.
- [ ] Record exact branch/commit and remaining unsupported stages (rewrite/director/video/merge/121 submit) for the next slice.
