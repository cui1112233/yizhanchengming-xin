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
- Main SHA at branch creation and re-check: `9f136e16deb2ea46bba4e2ececb1e63c829047eb`
- Main CI at branch creation: GitHub Actions `ci` run `37304304619`, conclusion `success`
- Migrations present and successfully applied in the first audit run:
  - `00001_phase1_intake.sql`
  - `00002_task12_generation.sql`
  - `00003_task10_unified_settings.sql`
  - `00004_task15_auth.sql`
  - `00006_task13_match_audio.sql`
- Goose reported database version `6` after migration.

### Parallel-task status at branch creation/re-check

- Task 9: **not accepted/merged as a complete task**. `TASKS.md` states Task 9.4 Runtime and Task 9.5 acceptance are still pending. Worker, Scheduler, Redis Queue/Lock, crash recovery and Runtime idempotency are therefore not treated as available capability in this audit branch.
- Task 14: **not merged**. Video task model/state machine/providers/polling/TOS/Merge/ffmpeg are blocked until Task 14 enters `main`.

When either task later enters `main`, this audit branch must sync the new `main` before enabling the corresponding tests. No Task 9/14 implementation is cherry-picked into this branch.

## Current connection/runtime configuration

### MySQL

`api/cmd/server/main.go` opens the database with `sql.Open("mysql", dsn)` and does not currently call any `database/sql` pool setters.

Current effective configuration is therefore Go `database/sql` defaults rather than repository-defined tuning:

- `MaxOpenConns`: `0` (unlimited)
- `MaxIdleConns`: default idle limit (`2`)
- `ConnMaxLifetime`: `0`
- `ConnMaxIdleTime`: `0`

No pool tuning was made. The first load run records `DB.Stats()` and establishes evidence before any tuning decision.

### Redis

Redis Runtime is not present in current `main`: Task 9.4 Queue/Lock/Worker work is still pending and `api/go.mod` does not contain a Redis client dependency. Therefore the following are **not configured in current main**, rather than zero-valued settings:

- Redis pool size;
- connection timeout;
- retry policy;
- max active connections;
- Queue backpressure;
- Redis restart/wipe recovery;
- lease/fencing contention.

These are explicit Task 9 blockers.

## Test environment

First measured audit run:

- GitHub Actions run: `37308959563`
- Job: `111759375286`
- Runner: Ubuntu `24.04.5` / `ubuntu-24.04`
- Go: `1.23.12 linux/amd64`
- MySQL image: `mysql:8.4`, runtime log reports MySQL `8.4.11`
- Goose migrations applied before the integration tests
- no production ECS traffic
- no real paid AI/video provider load

The test workload itself completed in well under one second; most workflow wall time was environment provisioning/download and is not counted as application latency.

## First workload set

### Baseline A — MySQL and HTTP reads

For each standard project size:

- Small: 10 books
- Medium: 50 books
- Large: 100 books

The fixture seeds one completed Intake, deterministic book text, one BatchProject and one pending Run. It measures:

- Run creation latency;
- `ListBatchProjects` query P50/P95/P99;
- `ListBooks`/project-detail query P50/P95/P99;
- HTTP project list at 10/50/100 simultaneous requests;
- HTTP project detail at 10/50/100 simultaneous requests;
- total/success/failure and throughput;
- heap before / load peak / post-GC;
- RSS before / load peak / post-GC;
- goroutine count before/after;
- DB open connections and pool wait delta.

No arbitrary latency SLA is imposed on this first baseline. Any request error is a test failure; latency is recorded as a regression baseline.

### RED A — concurrent duplicate Run creation

Reproduction:

1. seed one completed Intake;
2. start 20 goroutines simultaneously;
3. each calls the real pipeline `Create` path for the same Intake;
4. query MySQL for BatchProject count and logical Run count;
5. require one project and one logical Run.

Measured result:

- concurrency: `20`
- successful create calls: `20`
- failed create calls: `0`
- BatchProjects: `1`
- logical Runs: **`20`** (expected `1`)
- duration: `142.539032ms`
- goroutines: `4 -> 5`
- DB open after probe: `2`
- DB wait count/duration: `0 / 0s`

The RED failed exactly at `logical runs = 20, want 1`.

## First findings / risk register

### PERF-RUN-IDEMPOTENCY-001 — P0 CONFIRMED

**Impact:** concurrent double-click/API retry creates duplicate logical production Runs for one intended execution.

**Reproduction evidence:** 20 simultaneous calls for the same Intake produced one idempotent BatchProject but 20 Runs, with all 20 API/service calls reporting success.

**Root cause in current main:** `batch_projects.intake_id` has a uniqueness/upsert path, while `runs` has no logical idempotency key/unique constraint and current pipeline `Create` inserts a Run unconditionally.

**Boundary:** no Runtime/business fix is made in this audit PR because Task 9.4.6 owns idempotency. The RED remains as regression evidence and must turn green after Task 9 merges.

### PERF-DB-POOL-001 — P1 RISK, not confirmed saturation

`MaxOpenConns` is unlimited and the repository does not define pool limits. In the first baseline, measured post-load open connections were `2` and `WaitCount/WaitDuration` stayed `0/0s` even at 100 concurrent HTTP requests.

This does **not** prove peak connections stayed at 2 because the first probe samples `DB.Stats()` before/after the workload, not continuously during it. A peak sampler is required before deciding whether to tune the pool.

### PERF-HTTP-SERVER-001 — P1 RISK, not yet a confirmed defect

Current startup uses `http.ListenAndServe` directly. Explicit read/write/idle/header timeouts and graceful `Shutdown` are not visible in that startup path.

Slow-client cancellation and SIGTERM/recovery tests remain scheduled. No server refactor is made solely for this audit.

### PERF-BATCH-DETAIL-SCALE-001 — P2 scaling trend

At the tested payload size, 100 concurrent project-detail requests remained error-free, but detail P95 increased from `4.576ms` at 10 concurrency to `43.700ms` at 50 and `135.494ms` at 100. This is not a correctness failure, but it is the clearest first-stage latency scaling trend and should be retained as a regression baseline.

The direct MySQL detail query remained much lower: P95 `0.556ms`, `1.018ms`, `1.822ms` for 10/50/100-book projects respectively. The larger HTTP cost therefore includes response serialization/copying/concurrent delivery, not just the SQL query.

## Baseline results

### Latency / errors

| Workload | Run create | MySQL list P50 / P95 / P99 | MySQL detail P50 / P95 / P99 | HTTP list P50 / P95 / P99 | HTTP detail P50 / P95 / P99 | HTTP errors |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 books / 10 concurrent | 0.859 ms | 0.277 / 0.433 / 0.737 ms | 0.436 / 0.556 / 0.708 ms | 4.568 / 4.736 / 4.736 ms | 4.001 / 4.576 / 4.576 ms | 0 / 20 |
| 50 books / 50 concurrent | 0.885 ms | 0.303 / 0.356 / 0.550 ms | 0.744 / 1.018 / 1.078 ms | 30.051 / 38.324 / 39.262 ms | 30.121 / 43.700 / 44.399 ms | 0 / 100 |
| 100 books / 100 concurrent | 1.012 ms | 0.387 / 0.408 / 0.665 ms | 1.383 / 1.822 / 1.967 ms | 49.113 / 69.494 / 71.112 ms | 83.632 / 135.494 / 137.030 ms | 0 / 200 |

HTTP error denominator counts list + detail requests for the workload.

### Throughput

| Workload | Project list throughput | Project detail throughput |
| --- | ---: | ---: |
| 10 | 2052.84 req/s | 2117.85 req/s |
| 50 | 1262.02 req/s | 1098.25 req/s |
| 100 | 1395.33 req/s | 722.72 req/s |

These are short isolated runner measurements, not production SLA claims.

### Memory / goroutines / DB waits

| Workload | Heap before | Heap peak | Heap post-GC | RSS before | RSS peak | RSS post-GC | Goroutines before -> after | DB open after | DB wait count / duration |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | 0.30 MiB | 2.59 MiB | 0.52 MiB | 12.49 MiB | 15.74 MiB | 15.83 MiB | 6 -> 13 | 2 | 0 / 0s |
| 50 | 0.40 MiB | 3.00 MiB | 0.70 MiB | 15.20 MiB | 22.83 MiB | 20.14 MiB | 13 -> 13 | 2 | 0 / 0s |
| 100 | 0.56 MiB | 11.14 MiB | 1.08 MiB | 16.20 MiB | 53.25 MiB | 46.25 MiB | 13 -> 13 | 2 | 0 / 0s |

Raw byte values remain available in the Actions log for exact regression comparisons.

## Goroutine / memory status

### Goroutines

No goroutine leak is confirmed by the first baseline.

- the first HTTP workload initializes reusable server/client goroutines (`6 -> 13`);
- subsequent 50 and 100 concurrency workloads remain stable at `13 -> 13`.

This is only a first-pass signal. The required repeated create/query/retry cycles still need a longer leak test.

### Memory

No heap leak is confirmed by the first baseline.

For 100 books / 100 concurrent detail requests:

- heap rose to about `11.14 MiB` during load;
- post-GC heap returned to about `1.08 MiB`.

RSS rose to about `53.25 MiB` and remained about `46.25 MiB` after GC. Go/OS RSS is not expected to track heap release immediately, so a single run is insufficient to call this a leak. Repeated-cycle and profile evidence are required before filing a memory-leak bug.

## Verification evidence

Audit run `37308959563` / job `111759375286`:

- MySQL 8.4 service: healthy
- all migrations: PASS
- idempotency concurrency probe: RED reproduced (`20 Runs`, want `1`)
- 10/50/100 baseline: PASS, zero HTTP errors
- repository `go test ./...`: PASS

The RED step is intentionally diagnostic and preserves the real defect rather than changing Task 9 business architecture in this branch.

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
- The audit workflow also publishes compact commit statuses for the RED/baselines so numeric outcomes remain machine-readable.
- Heavy 100/500 Runtime load, long leak tests, Video/TOS/ffmpeg/disk tests remain manual and must not make every normal commit run a long capacity suite.

## ECS recommendation

No minimum ECS specification is recommended yet. The current GitHub runner baseline is useful for regression and correctness, but Task 16 capacity guidance must wait for measured CPU, memory, DB connections, disk/temp usage and ffmpeg concurrency in an approved isolated capacity environment.
