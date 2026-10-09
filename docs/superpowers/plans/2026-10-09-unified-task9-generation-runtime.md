# Unified Task9 Generation Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move text Generation admission and execution onto the existing MySQL + Redis `task9runtime` so HTTP never calls the paid Provider, queued work survives Redis/process loss, and every new BookRun belongs to a durable Run.

**Architecture:** HTTP validates Cookie Session, CSRF, capability, ownership, archived state, Provider availability and request shape, then creates one idempotent MySQL Run plus queued BookRuns and dispatches the existing Redis queue. One `task9runtime.Worker` executes a Generation adapter under Redis lease plus MySQL fencing; recovery rebuilds Redis from MySQL. This plan does not add a second runtime, queue, worker, scheduler, Node service, browser business storage, or a generic mirror job table.

**Tech Stack:** Go 1.26, MySQL 8/Goose, Redis/go-redis, existing `taskruntime` and `task9runtime`, React + Ant Design, Vitest.

## Global Constraints

- Work directly on `main`; no branch, force push, deployment, ECS change, GitHub Actions run, or `.github/workflows` edit.
- Backend code stays under `api`; user UI stays under `前台`.
- MySQL is the business fact source. Redis is delivery/lease/lock state only; Redis loss must be recoverable from MySQL.
- Reuse current Cookie Session, same-origin CSRF, effective capability and BatchProject ownership/archived guards.
- System Prompt content remains server-side in `generation` Prompt resolution; snapshots store only whitelisted request options and prompt keys/versions, never Prompt content or secrets.
- Provider calls occur only inside the runtime adapter, never in HTTP admission or poll GET.
- TOS is not required for text Generation. Missing Provider/Redis/runtime returns truthful unavailable state and performs zero paid calls.
- External Provider side effects are at-least-once unless the Provider supports an idempotency key; durable MySQL finalize is fenced exactly once.
- All new errors use safe code/message/requestId; never expose raw Provider body, Authorization, Cookie, password, DSN or key.
- Agent remains a placeholder and `/issues` remains removed.
- Each task uses TDD, a local `[skip ci]` commit, independent review, then fetch/rebase/test/push before the next task.

---

### Task 1: Redis Processing Receipts and Reclaim

**Files:**
- Modify: `api/internal/taskruntime/contracts.go`
- Modify: `api/internal/taskruntime/redis.go`
- Modify: `api/internal/taskruntime/contracts_test.go`
- Modify: `api/internal/task9runtime/runtime.go`
- Modify: `api/internal/task9runtime/runtime_contract_test.go`

**Interfaces:**
- Consumes: existing `taskruntime.Queue`, `taskruntime.Message`, and `task9runtime.Worker.RunOnce`.
- Produces: `type ProcessingReclaimer interface { ReclaimExpired(context.Context, time.Time, int) (int, error) }` and delivery metadata that lets the existing Redis queue recover abandoned processing receipts without changing business facts.

- [ ] **Step 1: Add failing real-Redis contract tests**

```go
func TestRedisQueueReclaimsExpiredProcessingReceipt(t *testing.T) {
    q := newRealRedisQueue(t)
    require.NoError(t, q.Enqueue(ctx, Message{TaskKey: "bookrun:41:book:9", Attempt: 1}))
    delivery, err := q.Claim(ctx, "worker-a", 10*time.Millisecond)
    require.NoError(t, err)
    reclaimed, err := q.ReclaimExpired(ctx, time.Now().Add(time.Minute), 10)
    require.NoError(t, err)
    require.Equal(t, 1, reclaimed)
    redelivered, err := q.Claim(ctx, "worker-b", 10*time.Millisecond)
    require.NoError(t, err)
    require.Equal(t, delivery.Message, redelivered.Message)
}
```

Also cover ACK removing processing metadata, NACK moving metadata with the envelope, duplicate reclaim being harmless, and upgrade recovery of a processing envelope with missing metadata.

- [ ] **Step 2: Run the focused tests and record RED**

Run: `cd api && go test ./internal/taskruntime ./internal/task9runtime -run 'RedisQueue|Processing|RunOnce' -count=1`

Expected: failure because `ReclaimExpired` and processing receipt metadata do not exist.

- [ ] **Step 3: Implement receipt metadata in the existing RedisQueue**

Use the same namespace and keys:

```go
type ProcessingReclaimer interface {
    ReclaimExpired(context.Context, time.Time, int) (int, error)
}

func (q *RedisQueue) processingDeadlineKey() string { return q.prefix + ":queue:processing:deadline" }
func (q *RedisQueue) processingEnvelopeKey() string { return q.prefix + ":queue:processing:envelope" }
```

On claim, persist `receipt -> raw envelope` and a deadline. ACK/NACK remove both records. `ReclaimExpired` atomically removes expired metadata, removes the exact envelope from processing, and returns it to ready. A bounded upgrade pass requeues processing envelopes that have no metadata. Preserve the current Queue interface so existing fakes do not need business changes; Worker may use the optional `ProcessingReclaimer` capability.

- [ ] **Step 4: Make worker transient errors retry instead of permanently stopping**

Add a bounded-backoff supervisor helper around existing `Worker.RunOnce`; context cancellation remains the only normal shutdown. Queue/lease transient errors log safe metadata at the app layer and retry; malformed envelopes are ACKed once.

- [ ] **Step 5: Run tests and commit**

Run:

```bash
cd api
go test ./internal/taskruntime ./internal/task9runtime -count=1
go test ./... -count=1
```

Commit: `feat(runtime): reclaim abandoned queue deliveries [skip ci]`

---

### Task 2: Typed Generation Run Admission and Atomic Materialization

**Files:**
- Create: `api/db/migrations/00021_generation_runtime.sql` (re-read live main first and use the next free number)
- Modify: `api/internal/task9runtime/mysql_store.go`
- Create: `api/internal/task9runtime/mysql_store_test.go`
- Modify: `api/internal/task9runtime/mysql_integration_test.go`
- Modify: `api/internal/task9runtime/runtime.go`
- Modify: `api/internal/task9runtime/runtime_contract_test.go`
- Create: `api/internal/task9runtime/generation_runtime_migration_contract_test.go`
- Create: `api/internal/task9runtime/generation_admission.go`
- Create: `api/internal/task9runtime/generation_admission_test.go`

**Interfaces:**
- Consumes: BatchProject ID, one or all owned Book IDs, a stable idempotency key, whitelisted Generation options and requested actor ID.
- Produces:

```go
type GenerationRequest struct {
    BatchProjectID int64
    BookIDs []int64
    HookEnabled bool
    PlotMode bool
    DirectorMode string
    MatchAudio bool
    ShotDurationLimitSec int64
    RequestID string
    RequestedByUserID int64
}

type AdmissionResult struct {
    Run RunRecord
    Items []WorkItem
    Created bool
    Dispatch string // queued | pending
}

func (s *MySQLStore) AdmitGeneration(context.Context, GenerationRequest) (AdmissionResult, error)
```

- [ ] **Step 1: Add failing migration and admission tests**

The additive migration extends `runs` with `run_kind`, `target_book_id`, `request_schema_version`, `request_snapshot`, `request_hash`, `requested_by_user_id`, `cancel_requested_at`, `cancelled_at`; extends `book_runs` with cancel timestamps; adds `(run_kind,status,run_at,id)` index. `run_kind` defaults to `legacy`, so historical pending rows are never executed as Generation work. Down must explicitly `SIGNAL SQLSTATE '45000'` because rollback would discard execution metadata.

Tests cover: 20 concurrent identical requests create one Run and one BookRun per selected Book; same idempotency key with different hash returns `ErrIdempotencyConflict`; Book IDs must belong to the BatchProject intake; archived project returns `ErrProjectArchived`; no new `run_id=NULL`; snapshot contains no Prompt, secret, Cookie or raw original text. Also verify migration defaults old rows to `legacy`, Down signals before any DROP, and no generic runtime fact table is created.

- [ ] **Step 2: Run focused RED tests**

Run: `cd api && go test ./internal/task9runtime -run 'Admission|Generation|Migration|Idempotency' -count=1`

- [ ] **Step 3: Implement one transaction for Run plus first BookRuns**

Lock in the global order BatchProject → Run → BookRun, reject `archived_at IS NOT NULL`, validate every selected Book through `books.intake_id=batch_projects.intake_id`, insert/read the idempotent Run, compare `request_hash`, and materialize queued BookRuns in the same transaction. Do not use the old `ClaimDueRun` then `EnsureQueuedBooks` two-transaction window for immediate Generation.

Persist `request_snapshot` as schema-versioned JSON containing IDs and the boolean/enum/duration settings only. Resolve actual Prompt versions during execution and persist them on StageRun. Normalize, sort and deduplicate explicit Book IDs. An empty selection means all current project books and freezes their concrete IDs in the snapshot; a later replay cannot absorb newly added books. Compare the original idempotency key bytes after a case-insensitive unique-index hit, and determine `Created` by an explicit locked lookup rather than `RowsAffected`, which changes under `clientFoundRows`.

- [ ] **Step 4: Repair scheduled claim/materialize crash window**

Replace the split scheduler store calls with:

```go
type SchedulerStore interface {
    ClaimAndMaterializeDueRun(context.Context, time.Time) (RunClaim, []WorkItem, bool, error)
}
```

The transaction claims only exact `run_kind='generation'`, validates active project and supported snapshot/hash, inserts BookRuns, and commits once. It must follow the same BatchProject → Run → BookRun lock order as admission. Invalid, archived or legacy candidates must not block later valid due work. Recovery is invoked from `Recovery.Rebuild` and marks `running` Runs with zero BookRuns back to pending only when `run_kind='generation'`, the snapshot/hash is valid, the frozen book list is non-empty and the project is active.

- [ ] **Step 5: Run full Go tests and commit**

Commit: `feat(runtime): admit generation runs atomically [skip ci]`

---

### Task 3: Fenced Generation Executor

**Files:**
- Create: `api/internal/generation/runtime_executor.go`
- Create: `api/internal/generation/runtime_executor_test.go`
- Modify: `api/internal/generation/service.go`
- Modify: `api/internal/generation/store.go`
- Modify: `api/internal/generation/mysql_store.go`
- Modify: `api/internal/task9runtime/runtime.go`
- Modify: `api/internal/task9runtime/mysql_store.go`
- Modify: `api/internal/task9runtime/error_persistence_integration_test.go`

**Interfaces:**
- Consumes: `task9runtime.Execution` for an already-created queued BookRun.
- Produces:

```go
type RuntimeBookRunStore interface {
    RuntimeBookRun(context.Context, int64, int, uint64, string) (BookRun, RunBookRequest, error)
    CreateStageRunFenced(context.Context, task9runtime.Execution, StageRun) (StageRun, error)
    UpdateStageRunFenced(context.Context, task9runtime.Execution, StageRun) (StageRun, error)
}

type RuntimeExecutor struct {
    store RuntimeBookRunStore
    provider Provider
    resolver *PromptResolver
    compiler FinalPromptCompiler
    now Clock
}
func (e *RuntimeExecutor) Execute(context.Context, task9runtime.Execution) error
```

- [ ] **Step 1: Add failing executor tests**

Cover SCRIPT → optional HOOK → DIRECTOR → local FINAL_PROMPT, disabled hook as skipped, H3 validation, authoritative audio measurement, safe Provider failure, Prompt version persistence, context cancellation, and stale token rejection before every StageRun create/update and before BookRun finalization.

Use a fake Provider counting calls. Admission tests must assert Provider calls remain zero; executor tests assert calls happen only inside `Execute`.

- [ ] **Step 2: Extract execution from synchronous admission**

Keep pure stage-building/compiler helpers in `generation.Service`, but make runtime execution accept an existing fenced BookRun. New production HTTP paths must not call `RunBook`, `RunBatch`, or `RetryStage` synchronously. Legacy methods may remain temporarily for focused unit compatibility, but app wiring must inject the async admission service instead of exposing them to HTTP.

- [ ] **Step 3: Fence every durable write**

Every StageRun INSERT/UPDATE verifies the parent BookRun tuple:

```sql
id = ? AND attempt = ? AND status = 'running'
AND execution_token = ? AND execution_owner = ?
```

A worker that lost its lease may finish a Provider call, but all later StageRun and BookRun writes return `task9runtime.ErrStaleExecution`.

- [ ] **Step 4: Aggregate parent Run after finalization**

After successful or failed fenced BookRun finalize, call `AggregateRunStatus`. Preserve `partial_failed`; no HTTP request performs this aggregation.

- [ ] **Step 5: Run Generation + runtime tests and commit**

Commit: `feat(generation): execute book runs through task9 runtime [skip ci]`

---

### Task 4: Async HTTP Contract and Frontend Polling

**Files:**
- Modify: `api/internal/httpapi/generation_handlers.go`
- Modify: `api/internal/httpapi/generation_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`
- Modify: `api/internal/app/app.go`
- Modify: `前台/src/api.js`
- Modify: `前台/src/BatchProjectListPage.jsx`
- Modify: `前台/src/BatchProjectListPage.test.jsx`
- Modify: `前台/src/ScriptWorkspace.jsx`
- Modify: `前台/src/ScriptWorkspace.test.jsx`

**Interfaces:**
- Consumes: `GenerationAdmission.AdmitBook`, `AdmitBatch`, `RetryStage` plus existing read-only Generation summaries.
- Produces 202 responses:

```json
{
  "created": true,
  "status": "queued",
  "runId": 123,
  "taskIds": [456],
  "pollUrl": "/api/v1/batch-projects/12/generation",
  "dispatch": "queued"
}
```

- [ ] **Step 1: Add failing handler tests**

Assert POST returns 202 before any Provider call; repeated same idempotency key and body returns the same Run; same key/different body returns safe 409; queue/runtime/provider unavailable returns safe 503 and zero Provider calls; ownership/capability/CSRF/archive behavior remains enforced.

- [ ] **Step 2: Replace synchronous handlers with admission**

`Idempotency-Key` is preferred; existing body `requestId` is accepted as the compatibility key. MySQL commit followed by transient Redis enqueue failure returns 202 with `dispatch=pending`; Recovery will dispatch it. A runtime that was never configured returns 503 and persists a safe terminal failure fact rather than accepted success.

- [ ] **Step 3: Make poll GET read MySQL only**

Keep existing summary GETs and add an exact Run GET if the UI cannot identify one run safely. GET must perform no Provider/TOS/Redis work.

- [ ] **Step 4: Update frontend state machine**

Treat 202 as queued/running, poll server state with bounded backoff, stop on terminal status/unmount/project switch, preserve requestId on failure, and never write task facts to LocalStorage. Reuse the existing stale-response guards and safe error presentation.

- [ ] **Step 5: Run Go tests, foreground full frontend tests, build and commit**

Commit: `feat(generation): expose async runtime workflow [skip ci]`

---

### Task 5: Production Lifecycle, Recovery and Readiness

**Files:**
- Modify: `api/internal/app/app.go`
- Modify: `api/internal/app/app_test.go`
- Modify: `api/cmd/server/main.go`
- Modify: `api/internal/task9runtime/runtime.go`
- Modify: `api/internal/task9runtime/recovery_integration_test.go`
- Modify: `api/internal/httpapi/observability.go`
- Modify: `api/internal/httpapi/observability_test.go`
- Modify: `TASKS.md`

**Interfaces:**
- Consumes: one RedisQueue, one RedisLeaseStore, one Scheduler, one Recovery and one Worker with the Generation adapter.
- Produces:

```go
type RuntimeStatus struct {
    State string // available | degraded | unavailable
    ReasonCode string
}

type RuntimeLifecycle interface {
    Start(context.Context) error
    Ready() RuntimeStatus
    Wait() error
    Close() error
}

type Application struct {
    Handler http.Handler
    Runtime RuntimeLifecycle
}

func NewApplication(*sql.DB, intake.Fetcher, intake.Classifier, pipeline.Clock) (*Application, error)
```

Keep `NewHandler` as a compatibility constructor for tests; production `cmd/server` uses `NewApplication` and owns the lifecycle.

- [ ] **Step 1: Add failing lifecycle tests**

Cover startup Recovery before readiness, Redis FLUSHDB queued reconstruction, expired running recovery, worker transient error supervision, clean signal cancellation, HTTP shutdown, queue/lease close, no goroutine leak, and diagnostics reflecting actual configured/ready/degraded state.

- [ ] **Step 2: Build one Task9 runtime in app wiring**

Use one namespace `{QIANTIE_REDIS_PREFIX}:task9`. Construct Queue and LeaseStore together, register the Generation executor, run initial Recovery, then start Scheduler/periodic Recovery/fixed workers. Do not modify or remove the existing Shuihuo worker in this plan; its later cutover is a separate reviewed slice after Generation is proven.

- [ ] **Step 3: Add graceful server lifecycle**

Use `signal.NotifyContext`, `http.Server`, bounded `Shutdown`, runtime cancel/join, Redis Close, then DB Close. Replace `context.Background()` only for the new Task9 goroutines in this plan.

- [ ] **Step 4: Make diagnostics truthful**

Remove hard-coded `pending_task9_runtime` for Generation. Report configured, ready, degraded or unavailable from actual Queue/Lease/worker state without exposing Redis address or credentials.

- [ ] **Step 5: Verify the vertical gate and commit**

Run:

```bash
cd api
go test ./... -count=1
go test -race ./internal/taskruntime ./internal/task9runtime ./internal/generation ./internal/app
```

The gate passes only when: HTTP Provider call count is zero; every new Generation BookRun has non-NULL `run_id`; a Redis flush is rebuilt from MySQL; a stale worker cannot persist StageRun or final state; and shutdown joins all Task9 goroutines.

Commit: `feat(runtime): run generation lifecycle in the embedded server [skip ci]`

---

### Task 6: Generation Runtime Acceptance

**Files:**
- Modify: `docs/deploy/local-staging-acceptance.md`
- Modify: `TASKS.md`

**Interfaces:**
- Consumes: Tasks 1–5.
- Produces: evidence for the Generation vertical slice; it does not claim Video/Merge/Shuihuo/TTS cutover.

- [ ] **Step 1: Run migration only on local `ycm_staging`**

Record Goose before/after status and verify live schema columns/indexes. Never connect to production.

- [ ] **Step 2: Run no-Provider acceptance**

Start the embedded binary on `127.0.0.1:18080` with local MySQL, staging Redis prefix, no text Provider and no TOS. Verify login/CSRF/ownership and that generation returns truthful `executor_unavailable` with zero queued fake success.

- [ ] **Step 3: Run fake local Provider recovery acceptance**

Use a non-paid local fixture Provider. Submit one run, interrupt/restart the Go process, flush only the staging Redis prefix, and verify MySQL rebuild plus one durable terminal result. Confirm exact `/api/build-info` SHA and refresh recovery in the browser.

- [ ] **Step 4: Run full tests and embedded build**

Run full Go tests, foreground frontend tests to natural exit, both frontend builds, and `./scripts/build-embedded-ui.sh`. Record all exits and counts.

- [ ] **Step 5: Commit evidence**

Commit: `test(runtime): record local generation recovery acceptance [skip ci]`
