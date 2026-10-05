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

Destructive load tests against production ECS are out of scope. Real paid providers remain explicit opt-in only.

## Repository baseline

- Audit branch: `test/performance-concurrency-capacity`
- Current synced main SHA: `0548a17328ed699af2aea7536d85a2d3c0496710`
- Current main commit: `docs(tasks): record Task14 repository acceptance`
- Current main CI: run `37310103359`, conclusion `success`
- Current migrations:
  - `00001_phase1_intake.sql`
  - `00002_task12_generation.sql`
  - `00003_task10_unified_settings.sql`
  - `00004_task15_auth.sql`
  - `00006_task13_match_audio.sql`
  - `00007_task14_video.sql`
- Audit MySQL migration run reached Goose version `7`.

### Task status at current main

- Task 9: **not merged/accepted as Runtime**. `TASKS.md` still has 9.4.1–9.4.10 and 9.5 unchecked. Worker, Scheduler, Redis Queue/Lock, lease/recovery and Runtime idempotency are not available as accepted mainline capability.
- Task 14: **merged and repository-accepted in main**. Video model/state machine, provider adapters, persistence, TOS/merge/ffmpeg components are present. However, application wiring explicitly does not create an async queue, Redis lease runtime or scheduler; that adapter still waits for Task 9.4.

The audit branch was re-synced with main when Task 14 landed. No unmerged Task 9/14 implementation was cherry-picked.

## Current connection/runtime configuration

### MySQL

`api/cmd/server/main.go` uses `sql.Open("mysql", dsn)` and does not call the Go `database/sql` pool setters.

Effective defaults:

- `MaxOpenConns`: `0` (unlimited)
- `MaxIdleConns`: default `2`
- `ConnMaxLifetime`: `0`
- `ConnMaxIdleTime`: `0`

No tuning was made in this audit. The first load set records `DB.Stats()` before any tuning decision.

### Redis

No Redis client dependency/configuration is present in current main, consistent with Task 9.4 still pending. Therefore Redis pool size, timeout, retry and active-connection limits are **not configured yet**, rather than configured as zero.

## Test environment

Primary measured baseline source:

- GitHub Actions run: `37310980711`
- Job: `111766038496`
- Audit test commit: `77cf9e128b3557fc68b5a80ac0d8be82ff8157f4`
- Runner: Ubuntu `24.04.5`
- Go: `1.23.12 linux/amd64`
- MySQL: `8.4.11`
- Goose database version: `7`
- no production ECS traffic
- no real paid AI/video provider load

The run completed migrations, baseline tests, controlled Task 14 probes, `go test ./...`, and `go build ./...` successfully. The Task 9 idempotency RED is intentionally executed as a diagnostic and preserved as a failing correctness assertion while the workflow continues to collect the remaining metrics.

## First workload set

### MySQL / HTTP baseline

Standard project sizes:

- Small: 10 books / 10 concurrent HTTP requests
- Medium: 50 books / 50 concurrent HTTP requests
- Large: 100 books / 100 concurrent HTTP requests

Measured paths:

- Run creation;
- BatchProject list query/API;
- BatchProject detail query/API;
- request counts, errors and throughput;
- P50/P95/P99;
- heap/RSS before, peak and post-GC;
- goroutine before/after;
- DB open connections and pool waits.

### Task 9 idempotency RED

20 goroutines simultaneously call the real pipeline Create path for the same Intake.

Measured result:

- concurrency: `20`
- successful calls: `20`
- failed calls: `0`
- BatchProjects: `1`
- logical Runs: **`20`**, expected `1`
- duration: `240.374857ms`
- goroutines: `4 -> 5`
- DB open: `2`
- DB wait count/duration: `0 / 0s`

### Task 14 controlled provider load

A controlled `httptest` provider is used; no paid provider is called. Each task performs exactly one submit + one poll.

| Concurrency | Tasks | Success / failed | Provider requests | P50 | P95 | P99 |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | 10 | 10 / 0 | 20 | 0.824 ms | 1.159 ms | 1.159 ms |
| 50 | 50 | 50 / 0 | 100 | 4.296 ms | 5.071 ms | 5.200 ms |
| 100 | 100 | 100 / 0 | 200 | 5.808 ms | 7.280 ms | 7.529 ms |

These numbers validate adapter concurrency behavior only; they are not a real provider SLA.

### Task 14 ffmpeg executor concurrency boundary

A fake command runner is used to measure executor-level concurrency without consuming real ffmpeg CPU/RAM.

| Requested concurrency | Success / failed | Maximum simultaneously active executor calls |
| ---: | ---: | ---: |
| 1 | 1 / 0 | 1 |
| 2 | 2 / 0 | 2 |
| 4 | 4 / 0 | 4 |

This proves that the current executor does not impose an internal ffmpeg concurrency limit. Real CPU/RAM/disk capacity must still be measured in an approved isolated environment.

## Baseline results

### Latency / errors

| Workload | Run create | MySQL list P50 / P95 / P99 | MySQL detail P50 / P95 / P99 | HTTP list P50 / P95 / P99 | HTTP detail P50 / P95 / P99 | HTTP errors |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 books / 10 concurrent | 0.916 ms | 0.242 / 0.374 / 0.567 ms | 0.323 / 0.592 / 0.683 ms | 3.634 / 4.669 / 4.669 ms | 3.648 / 5.238 / 5.238 ms | 0 / 20 |
| 50 books / 50 concurrent | 0.807 ms | 0.278 / 0.338 / 0.432 ms | 0.769 / 1.021 / 1.141 ms | 13.202 / 21.904 / 22.721 ms | 35.996 / 44.704 / 46.441 ms | 0 / 100 |
| 100 books / 100 concurrent | 0.923 ms | 0.375 / 0.579 / 0.597 ms | 1.283 / 1.656 / 1.949 ms | 46.078 / 65.410 / 66.066 ms | 101.869 / 116.894 / 119.217 ms | 0 / 200 |

### Throughput

| Workload | Project list throughput | Project detail throughput |
| --- | ---: | ---: |
| 10 | 2086.07 req/s | 1875.62 req/s |
| 50 | 2096.81 req/s | 1067.75 req/s |
| 100 | 1487.06 req/s | 829.28 req/s |

These are short isolated-runner measurements for regression comparison, not production SLA commitments.

### Memory / goroutines / DB waits

| Workload | Heap before | Heap peak | Heap post-GC | RSS before | RSS peak | RSS post-GC | Goroutines before -> after | DB open after | DB wait count / duration |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | 0.33 MiB | 0.99 MiB | 0.44 MiB | 12.74 MiB | 15.86 MiB | 15.05 MiB | 6 -> 13 | 2 | 0 / 0s |
| 50 | 0.42 MiB | 7.10 MiB | 0.59 MiB | 13.16 MiB | 24.81 MiB | 21.75 MiB | 13 -> 13 | 2 | 0 / 0s |
| 100 | 0.59 MiB | 46.59 MiB | 1.27 MiB | 16.68 MiB | 78.91 MiB | 78.91 MiB | 13 -> 13 | 2 | 0 / 0s |

## First findings / risk register

### PERF-RUN-IDEMPOTENCY-001 — P0 CONFIRMED

Concurrent double-click/API retry can create duplicate logical Runs. Twenty simultaneous calls for one Intake produced one BatchProject but twenty Runs, with all twenty calls reporting success.

Current main has no Runtime idempotency guarantee for this path. This audit does not fix it because Task 9.4.6 owns that business behavior. The RED remains as regression evidence and must become green after Task 9 Runtime merges.

### PERF-FFMPEG-001 — P1 CONFIRMED CAPACITY RISK

The current ffmpeg executor permits all tested merge calls to run simultaneously: requested concurrency 4 produced `max_active=4`.

This is not itself a functional failure, but on an ECS it can allow CPU/RAM/disk saturation if many merge jobs are admitted concurrently. Per audit boundary, no new scheduler/semaphore is added here. Total-control should decide the production concurrency policy after real 1/2/4+ ffmpeg resource measurement.

### PERF-DB-POOL-001 — P1 RISK, saturation not reproduced

The application leaves `MaxOpenConns` unlimited and uses default idle-pool behavior. First-stage loads showed `WaitCount=0` and `WaitDuration=0`, so there is no evidence yet for tuning values.

The probe samples pool stats before/after the workload and does not continuously capture peak open connections. A peak sampler plus larger controlled load is required before pool changes.

### PERF-HTTP-SERVER-001 — P1 RISK, not yet reproduced as a failure

Server startup uses direct `http.ListenAndServe`. Explicit read/write/idle/header timeouts and graceful `Shutdown` are not visible in the startup path. Slow-client cancellation and SIGTERM/recovery remain pending tests.

### PERF-BATCH-DETAIL-SCALE-001 — P2 scaling trend

Project-detail HTTP P95 increased from `5.238ms` at 10 concurrency to `44.704ms` at 50 and `116.894ms` at 100, with zero request errors.

Direct MySQL detail P95 remained `0.592ms`, `1.021ms`, and `1.656ms`, so the observed HTTP scaling cost is not explained by the SQL query alone and includes response construction/serialization/concurrent delivery.

## Goroutine / memory status

No goroutine leak is confirmed in the first baseline. After reusable HTTP/server goroutines were initialized (`6 -> 13`), the 50- and 100-concurrency runs remained `13 -> 13`.

No heap leak is confirmed. At the 100-book workload, heap rose to about `46.59 MiB` and returned to about `1.27 MiB` after GC.

RSS rose to about `78.91 MiB` and remained at that level immediately after GC. A single Go process sample is insufficient to call this a leak because RSS does not necessarily return to the OS with heap GC. This remains a retention signal requiring repeated cycles and profile evidence before filing a leak defect.

## Generation status

Generation code is in main and the repository-wide tests pass, but a dedicated 10/50/100 controlled-provider Generation throughput workload has not yet been completed in this first batch. No real AI load is enabled by default.

## Task 9 blocked tests

Blocked until Task 9 Runtime enters `main`:

- 1/2/5/10 Worker claim distribution;
- duplicate claim and claim contention;
- lease renew / expiry / reclaim / fencing;
- stale complete / fail / release;
- Worker crash storm and stale-running recovery;
- Redis restart, wipe and temporary unavailability;
- Redis connection pool capacity;
- Scheduler 100 simultaneous/near-simultaneous `run_at` Runs and multi-Scheduler ownership;
- Queue backpressure at 100/500 pending BookRuns;
- graceful Worker shutdown and startup recovery;
- Runtime-level 50/100-book partial failure isolation and failed-only Retry;
- true 1/5/10 concurrent Run execution isolation;
- production async Video scheduling/polling/lease recovery that depends on the shared Runtime.

## Task 14 status / remaining capacity work

Task 14 itself is no longer a merge blocker. Controlled Personal API concurrency and executor-level ffmpeg concurrency are now directly tested.

Still pending as manual/isolated capacity work rather than Task 14 merge blockers:

- 429/transient-5xx/slow-response rate-limit behavior under the final shared Runtime retry policy;
- 100-running-video poll-storm request-rate measurement after Task 9 scheduler/worker exists;
- TOS concurrent upload/retry/temp cleanup capacity;
- real ffmpeg CPU/RAM/duration at 1/2/4+ concurrency;
- 10/20/50-fragment real merge disk/duration/cleanup;
- disk-full behavior;
- real provider smoke, only with explicit `ENABLE_REAL_PROVIDER_LOAD_TEST=true`.

## CI layering

- Existing normal repository CI remains unchanged.
- `.github/workflows/performance-audit.yml` provides an isolated MySQL audit workflow for this branch and `workflow_dispatch`.
- Known Task 9 idempotency RED is reported explicitly rather than converted to fake success; the workflow continues so the rest of the metrics can still be collected.
- 10/50/100 MySQL/HTTP baseline and controlled Task 14 probes are lightweight.
- 100/500 Runtime, long leak, real ffmpeg/TOS/disk and real-provider tests remain manual.

## Verification evidence

Audit run `37310980711` / job `111766038496` on test commit `77cf9e128b3557fc68b5a80ac0d8be82ff8157f4`:

- MySQL service healthy;
- migrations through `00007_task14_video.sql`: PASS;
- Task 9 idempotency RED: reproduced, `20` logical Runs instead of `1`;
- 10/50/100 MySQL + HTTP baseline: PASS, zero request errors;
- controlled Task 14 provider 10/50/100: PASS;
- ffmpeg executor 1/2/4 probe: PASS, maximum active matched requested concurrency;
- `go test ./...`: PASS;
- `go build ./...`: PASS.

## ECS recommendation

No minimum ECS specification is recommended yet. The GitHub runner results establish correctness/regression baselines but do not supply real ffmpeg CPU/RAM/disk measurements. Task 16 ECS sizing must wait for approved isolated capacity measurement rather than guessing a machine size.
