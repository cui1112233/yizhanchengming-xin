# Performance / Concurrency / Capacity Audit

## Audit scope

This audit validates concurrency correctness and capacity of the new mainline without changing Task 9 Runtime or Task 14 Video/Merge business architecture merely to make load tests pass.

Correctness priority:

1. no duplicate execution;
2. no lost tasks;
3. no permanently stuck tasks;
4. no unbounded retry;
5. no resource leak;
6. throughput/latency after correctness.

Destructive load tests against the production ECS are explicitly out of scope. Paid providers are opt-in only.

## Repository baseline

- Audit branch: `test/performance-concurrency-capacity`
- Main SHA at branch creation: `9f136e16deb2ea46bba4e2ececb1e63c829047eb`
- Main CI at branch creation: GitHub Actions `ci` run `37304304619`, conclusion `success`
- Migrations present at branch creation:
  - `00001_phase1_intake.sql`
  - `00002_task12_generation.sql`
  - `00003_task10_unified_settings.sql`
  - `00004_task15_auth.sql`
  - `00006_task13_match_audio.sql`

### Parallel-task status at branch creation

- Task 9: **not accepted/merged as a complete task**. The current `TASKS.md` states Task 9.4 Runtime and Task 9.5 acceptance are still pending. Worker, Scheduler, Redis Queue/Lock, crash recovery and idempotency are therefore not treated as available Runtime capability in this audit branch.
- Task 14: **not merged**. Video task model/state machine/providers/polling/TOS/Merge/ffmpeg are blocked until Task 14 enters `main`.

When either task later enters `main`, this audit branch must sync the new `main` before enabling the corresponding tests. No Task 9/14 implementation is cherry-picked into this branch.

## Current connection/runtime configuration

### MySQL

`api/cmd/server/main.go` opens the database with `sql.Open("mysql", dsn)` and does not currently call any of the `database/sql` pool setters.

Current effective configuration is therefore the Go `database/sql` defaults rather than repository-defined tuning:

- `MaxOpenConns`: `0` (unlimited)
- `MaxIdleConns`: default idle limit (`2`)
- `ConnMaxLifetime`: `0`
- `ConnMaxIdleTime`: `0`

This audit does **not** tune these values based on intuition. Load tests record `DB.Stats()` open/in-use/idle connections, `WaitCount`, and `WaitDuration` first.

### Redis

Redis Runtime is not present in current `main`: Task 9.4 Queue/Lock/Worker work is still pending and `api/go.mod` does not yet contain a Redis client dependency. Therefore the following cannot yet be truthfully reported or tested:

- Redis pool size;
- connection timeout;
- retry policy;
- max active connections;
- Queue backpressure;
- Redis restart/wipe recovery;
- lease/fencing contention.

These are explicit **Task 9 blockers**, not zero-valued settings.

## Test environment

First-stage execution environment:

- GitHub Actions isolated Ubuntu runner;
- MySQL `8.4` service container;
- Go `1.23.x`;
- Goose migrations applied before integration load tests;
- no production ECS traffic;
- no real paid AI/video provider load.

## First workload set

### Baseline A — MySQL and HTTP reads

For each standard project size:

- Small: 10 books
- Medium: 50 books
- Large: 100 books

Seed a completed Intake, books with deterministic medium-size text, BatchProject and one pending Run. Measure:

- Run creation latency;
- `ListBatchProjects` query P50/P95/P99;
- `ListBooks`/project-detail query P50/P95/P99;
- HTTP project list at 10/50/100 concurrent requests;
- HTTP project detail at 10/50/100 concurrent requests;
- success/error counts and throughput;
- heap before load / after load / post-GC;
- RSS before load / after load / post-GC;
- goroutine count before/after;
- DB open connections and pool wait delta.

No millisecond pass/fail SLA is imposed in this first baseline. A request failure is a test failure; latency is recorded for comparison.

### RED A — concurrent duplicate Run creation

Reproduction:

1. seed one completed Intake;
2. start 20 goroutines simultaneously;
3. each calls the real pipeline `Create` path for the same Intake;
4. query MySQL for BatchProject count and logical Run count;
5. require one project and one logical Run.

The current schema makes `batch_projects.intake_id` unique, but `runs` has no logical idempotency key and `CreateRun` performs a plain INSERT. The probe is therefore expected to expose Task 9.4.6 until Task 9 Runtime merges.

Metrics emitted by the RED include successes/failures, project count, logical Run count, duration, goroutine counts, open/in-use/idle DB connections, pool waits and wait duration.

## First findings / risk register

### PERF-RUN-IDEMPOTENCY-001 — P0 candidate

**Risk:** concurrent double-click/API retry can create multiple logical Runs for one intended execution.

**Evidence before execution:** the project row is idempotent by unique Intake + upsert, while the Run row has neither an idempotency key nor uniqueness constraint and is inserted unconditionally by the current pipeline.

**Status:** automated real-MySQL RED added. Final reproduction count is recorded from the workflow run; no Runtime fix is made in this audit branch because Task 9.4.6 owns that business behavior.

### PERF-DB-POOL-001 — P1 risk, not yet a confirmed defect

**Risk:** unlimited `MaxOpenConns` can fan out MySQL connections under high concurrency while the default idle pool is very small.

**Status:** observe `DB.Stats()` under 10/50/100 first. Do not tune until saturation/wait evidence exists.

### PERF-HTTP-SERVER-001 — P1 risk, not yet a confirmed defect

**Risk:** current server startup uses `http.ListenAndServe` directly. Explicit read/write/idle/header timeouts and graceful `Shutdown` are not visible in the current startup path.

**Status:** slow-client and SIGTERM recovery tests are scheduled for the audit after the baseline. No server refactor is made solely for this audit.

### PERF-BATCH-LIST-001 — P2 risk, measurement pending

`ListBatchProjects` aggregates book metadata and uses a correlated latest-Run subquery. The 10/50/100 baseline records query latency before any SQL/index change is considered.

## Baseline results

The first GitHub Actions baseline run is the source of truth for the numbers below. Do not fill cells from estimates.

| Workload | Run create | MySQL list P50/P95/P99 | Detail P50/P95/P99 | HTTP list P50/P95/P99 | HTTP detail P50/P95/P99 | Errors | Peak/post-GC memory | Goroutines | DB waits |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 10 books | pending first run | pending | pending | pending | pending | pending | pending | pending | pending |
| 50 books | pending first run | pending | pending | pending | pending | pending | pending | pending | pending |
| 100 books | pending first run | pending | pending | pending | pending | pending | pending | pending | pending |

## Goroutine / memory status

No goroutine or memory leak is declared from code inspection alone. The first baseline records before/after/post-GC values; repeated create/query/retry leak loops will be added separately after the current-main primitives have a stable baseline.

## Task 9 blocked tests

Blocked until Task 9 Runtime enters `main`:

- 1/2/5/10 Worker claim distribution;
- duplicate claim and lease contention;
- lease renew/expiry/reclaim/fencing;
- stale complete/fail/release;
- Worker crash storm and stale-running recovery;
- Redis restart, wipe, temporary unavailability;
- Scheduler 100 simultaneous `run_at` Runs and multi-Scheduler ownership;
- Redis connection-pool capacity;
- Queue backpressure at 100/500 pending BookRuns;
- graceful Worker shutdown and startup recovery;
- Runtime-level 50/100-book partial failure + failed-only retry;
- multi-Run 1/5/10 execution isolation at Worker level.

The concurrent Run-create RED is intentionally executable before Task 9 merges because it exercises current MySQL/pipeline idempotency rather than inventing a Runtime.

## Task 14 blocked tests

Blocked until Task 14 enters `main`:

- controlled video-provider submit/poll/retry/cancel/rate-limit/slow-response tests;
- real provider smoke (`ENABLE_REAL_PROVIDER_LOAD_TEST=true` only);
- 429 backoff/attempt-limit validation;
- 100-video poll-storm request-rate measurement;
- concurrent TOS upload / retry / temp cleanup / memory use;
- ffmpeg 1/2/4+ concurrency and `PERF-FFMPEG-001` decision;
- 10/20/50-fragment merge ordering/disk/duration/output/cleanup;
- disk-full behavior;
- retry after merge failure without regenerating provider VIDEO;
- final BatchProject video result writeback.

## CI layering

- Normal repo CI remains the existing lightweight test/build workflow.
- `performance-audit.yml` runs isolated MySQL diagnostic workloads on this audit branch and can also be triggered manually.
- Known RED behavior is emitted explicitly as a diagnostic warning rather than hidden as fake success.
- Heavy 100/500 Runtime load, long leak tests, Video/TOS/ffmpeg/disk tests will remain manual; they must not make every normal commit run a long capacity suite.

## ECS recommendation

No minimum ECS specification is recommended yet. Task 16 capacity guidance will only be added after measured CPU, memory, DB connections, disk/temp usage and ffmpeg concurrency data exist.
