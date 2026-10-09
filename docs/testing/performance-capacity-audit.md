# Performance / Concurrency / Capacity Audit

## Scope and guardrails

This audit validates concurrency correctness and capacity without changing Task 9 Runtime or Task 14 business architecture merely to make load tests pass.

Priority: no duplicate execution -> no lost/stuck tasks -> bounded retry -> no resource leak -> throughput/latency.

- Branch: `test/performance-concurrency-capacity`
- Draft PR: #25; keep Draft until Task 9 Runtime merges and the final Runtime sweep is complete.
- No destructive production ECS load.
- No large real paid AI/video provider load.
- Real provider load remains explicit opt-in only.

## Repository baseline

- Synced main SHA: `0548a17328ed699af2aea7536d85a2d3c0496710`
- Main commit: `docs(tasks): record Task14 repository acceptance`
- Migrations through `00007_task14_video.sql`; Goose version `7`.
- Task 9 Runtime: not accepted in main; 9.4/9.5 remain blockers.
- Task 14: repository-accepted in main; Video/Provider/TOS/Merge/ffmpeg components are testable, while async Queue/Lease/Scheduler integration still waits for Task 9.4.

## Connection/runtime configuration

### MySQL

Server startup uses `sql.Open("mysql", dsn)` without pool setters. Effective Go defaults remain:

- `MaxOpenConns=0` (unlimited)
- `MaxIdleConns=2` default
- `ConnMaxLifetime=0`
- `ConnMaxIdleTime=0`

No pool tuning was made. Current measured workloads show no pool waits, so there is not yet evidence for choosing 20/50/100 or another limit.

### Redis

No accepted Task 9 Redis Runtime/client configuration exists in current main. Pool size, timeout, retry, active-connection limits, queue depth, lease/fencing and recovery remain `BLOCKED_BY_TASK9_RUNTIME` rather than zero-valued configuration.

### HTTP server

Startup still uses direct `http.ListenAndServe` with no visible explicit `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, or graceful `Shutdown` path. This audit probes the risk but does not refactor server startup.

## Primary measured environment

Latest full metric source for the expanded workload before the slow-client/report-only follow-up commits:

- Performance workflow run: `37313261537`
- Job: `111773618137`
- Commit: `879f102f56d911e1b9c19e676f4c00ed2e1e1fbd`
- Ubuntu 24.04.5 GitHub runner
- Go 1.23.12 linux/amd64
- MySQL 8.4.11
- Goose database version 7
- no production ECS traffic
- no real paid provider load

That run completed migrations, all expanded capacity probes, `go test ./...`, and `go build ./...` successfully. The Task 9 idempotency RED remains an intentional diagnostic known-failure while the workflow continues to collect other metrics.

## PERF-RUN-IDEMPOTENCY-001 — P0 CONFIRMED / BLOCKED_BY_TASK9_4

Reproduction is deliberately preserved unchanged:

1. create one completed Intake;
2. issue 20 simultaneous calls to the real pipeline Create path for the same logical execution;
3. query MySQL for BatchProject and logical Run counts.

Latest measured reproduction:

- concurrent requests: `20`
- successful calls: `20`
- failed calls: `0`
- BatchProjects: `1`
- logical Runs: **`20`**
- expected: **`1` logical Run**
- duration: `73.532132ms`
- DB open: `2`
- DB waits: `0 / 0s`

This branch does not fix the defect. After Task 9.4 idempotency merges into main, sync main and run the same test. Acceptance then becomes a hard CI gate: `20 concurrent requests -> 1 logical Run`.

## MySQL / HTTP 10 / 50 / 100 baseline

Latest expanded-run snapshot:

| Workload | Run create | MySQL detail P95 | HTTP list P95 | HTTP detail P95 | HTTP detail P99 | HTTP errors | DB waits |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 books / 10 concurrent | 0.711 ms | 0.510 ms | 4.567 ms | 4.995 ms | 4.995 ms | 0 | 0 |
| 50 books / 50 concurrent | 0.849 ms | 1.190 ms | 28.167 ms | 50.155 ms | 51.938 ms | 0 | 0 |
| 100 books / 100 concurrent | 1.037 ms | 1.794 ms | 67.323 ms | 150.881 ms | 227.969 ms | 0 | 0 |

At 100 concurrency the detail endpoint still completes without errors, but latency scales materially and remains a P2 regression signal rather than an SQL-tuning instruction.

## Generation controlled load — 10 / 50 / 100

Generation is tested against real MySQL plus a controlled in-process provider. No AI credits are consumed. Current `RunBatch` is synchronous and sequential; the tests explicitly record `provider_max_active=1` instead of pretending the current model is concurrent.

| BookRuns | Success / failed | Duration | Throughput | P50 | P95 | P99 | Provider calls | Provider max active | DB waits |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | 10 / 0 | 195.235 ms | 51.22 books/s | 14.771 ms | 48.588 ms | 48.588 ms | 20 | 1 | 0 |
| 50 | 50 / 0 | 1.107 s | 45.15 books/s | 14.514 ms | 50.610 ms | 123.390 ms | 100 | 1 | 0 |
| 100 | 100 / 0 | 3.467 s | 28.84 books/s | 17.653 ms | 102.104 ms | 138.574 ms | 200 | 1 | 0 |

100-BookRun resource snapshot:

- heap before / peak / post-GC: `241,888 / 1,258,160 / 217,304` bytes
- RSS before / peak / post-GC: `14,163,968 / 15,634,432 / 14,979,072` bytes
- goroutines: `5 -> 5`
- DB: open `1`, in-use `0`, idle `1`, WaitCount `0`, WaitDuration `0s`

The throughput drop at larger batch size is recorded as a scaling characteristic of the current synchronous path, not a justification to rewrite Generation concurrency in this audit.

### Generation failure isolation

Controlled 10-book batch with book 5 failing at Script stage:

- total: `10`
- completed: `9`
- failed: `1`
- provider calls: `19`
- duration: `203.646ms`
- a later book after the failed book still completed successfully.

Result: current synchronous Generation batch correctly isolates one failed book and continues later books.

## 10-round RSS / goroutine retention

Workload: warm server, then 10 rounds of 100 concurrent 100-book project-detail requests; each round performs GC and an idle interval before sampling.

| Round | HeapAlloc | HeapInuse | Sys | RSS | Goroutines | DB waits |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 986,928 | 2,482,176 | 72,701,192 | 18,350,080 | 12 | 0 |
| 2 | 963,848 | 2,605,056 | 89,809,176 | 33,185,792 | 12 | 0 |
| 3 | 1,101,952 | 2,736,128 | 94,265,624 | 47,919,104 | 12 | 0 |
| 4 | 1,057,608 | 2,768,896 | 94,265,624 | 40,361,984 | 12 | 0 |
| 5 | 1,357,184 | 3,252,224 | 94,265,624 | 66,363,392 | 12 | 0 |
| 6 | 966,640 | 2,605,056 | 94,265,624 | 63,823,872 | 12 | 0 |
| 7 | 923,080 | 2,531,328 | 94,265,624 | 71,143,424 | 12 | 0 |
| 8 | 893,200 | 2,449,408 | 94,265,624 | 24,580,096 | 12 | 0 |
| 9 | 1,680,832 | 3,514,368 | 98,459,928 | 101,904,384 | 12 | 0 |
| 10 | 1,343,344 | 3,104,768 | 98,591,000 | 85,557,248 | 12 | 0 |

Summary sample immediately after the series:

- base RSS: `28,028,928`
- final RSS: `84,312,064`
- delta: `+56,283,136`
- strict RSS increase steps: `5/9`
- goroutines: `12 -> 12`

Interpretation: this does **not** prove a memory leak. RSS is strongly non-monotonic (for example round 8 falls to ~23.4 MiB), post-GC heap remains low, and Go `Sys` largely reaches a plateau around 94–99 MiB. Keep this as an allocator/runtime/OS retention signal. Escalate to heap/profile investigation only if longer repeated workloads show sustained growth rather than a plateau/oscillation.

### Goroutine leak status

No goroutine leak was reproduced in the 10-round loop: every measured round stayed at `12`, with summary delta `0`.

This covers the HTTP project-detail loop. Task 9 Worker/Scheduler ticker lifecycle remains blocked until the Runtime exists in main.

## Project Detail P95 decomposition

A 100-book response serializes to approximately `1,262,465` JSON bytes (~1.20 MiB). Under 100 simultaneous operations:

| Layer | P50 | P95 | P99 | Failures |
| --- | ---: | ---: | ---: | ---: |
| direct store / row materialization | 93.156 ms | 138.813 ms | 274.206 ms | 0 |
| store + JSON marshal | 182.479 ms | 229.682 ms | 234.433 ms | 0 |
| in-process HTTP handler | 93.446 ms | 128.550 ms | 130.331 ms | 0 |
| httptest network HTTP | 109.151 ms | 177.396 ms | 180.377 ms | 0 |

`auth/session` is not part of this benchmark handler and therefore is not a measured contributor here.

Conclusion: the earlier low single-query latency does not mean SQL is the only or main 100-way-concurrency cost. At high concurrency, DB/row materialization itself rises substantially; serializing a ~1.20 MiB object adds meaningful CPU/allocation cost; network delivery adds more overhead. Do not optimize the SQL statement blindly from this evidence.

## Task 14 controlled Provider failure behavior

20 simultaneous Submit calls per controlled failure mode:

| Case | Requests | Hidden retries | Failed | Error classification | Duration |
| --- | ---: | ---: | ---: | --- | ---: |
| HTTP 429 | 20 | 0 | 20 | `provider_request_failed` | 2.113 ms |
| HTTP 500 | 20 | 0 | 20 | `provider_request_failed` | 1.064 ms |
| HTTP 401 | 20 | 0 | 20 | `provider_auth_failed` | 0.937 ms |
| client timeout | 20 | 0 | 20 | `provider_unavailable` | 16.641 ms |
| slow success | 20 | 0 | 0 | success | 21.417 ms |
| unavailable endpoint | 1 | 0 | 1 | `provider_unavailable` | 0.241 ms |

The provider adapter itself does not create a retry storm. This does **not** validate final Runtime retry/backoff/lease policy; those tests remain blocked by Task 9. Poll storm testing also remains blocked until the shared Runtime exists.

## Merge fragment capacity — controlled executor

These are deterministic executor tests using a fake downloader/command runner/artifact store; they validate ordering, workdir cleanup and control-flow scaling, not real ffmpeg CPU/RAM.

| Fragments | Result | Duration | Controlled input bytes | Heap delta | Temp work dirs after | Ordered |
| ---: | --- | ---: | ---: | ---: | ---: | --- |
| 10 | PASS | 0.570 ms | 10,240 | 51,048 | 0 | yes |
| 20 | PASS | 0.985 ms | 20,480 | 79,672 | 0 | yes |
| 50 | PASS | 2.110 ms | 51,200 | 192,184 | 0 | yes |

Merge retry probe:

- attempt 1: controlled failure
- retry attempt 2: succeeded
- merge executor calls: `2`
- `MergeService` has no Video Provider dependency; therefore retry has no path to resubmit an already-successful VIDEO production task (`video_submit_calls=0_by_service_boundary`).

## Controlled artifact persistence / TOS adapter boundary

No real TOS capacity traffic is used.

| Concurrent PersistFile calls | Success / failed | Uploader calls | Hidden retries | Max active | Duration | Heap delta |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | 10 / 0 | 10 | 0 | 10 | 2.208 ms | 16,752 |
| 50 | 50 / 0 | 50 | 0 | 50 | 2.324 ms | 81,184 |

A controlled upload failure produced exactly one uploader call and zero hidden retry. Source files remain caller-owned; merge work-directory cleanup is tested separately and removes its temp directory.

This shows the storage adapter does not impose a concurrency gate or hidden retry. Real TOS network throughput, throttling and large-file behavior remain manual/Task16 capacity work.

## PERF-FFMPEG-001 — P1 CONFIRMED CAPACITY RISK

Executor-level fake-runner evidence remains:

- requested concurrency 1 -> `max_active=1`
- requested concurrency 2 -> `max_active=2`
- requested concurrency 4 -> `max_active=4`

There is no internal ffmpeg concurrency gate. No semaphore/scheduler is added in this audit branch.

A manual-only workflow, `.github/workflows/performance-ffmpeg-manual.yml`, is prepared to generate controlled synthetic fragments and measure real ffmpeg at concurrency 1/2/4:

- wall time
- max RSS
- user/system CPU time
- temp disk usage
- output bytes
- failure count
- cleanup

It is intentionally `workflow_dispatch` only. It has **not yet been executed in this audit session**, so there are no real ffmpeg CPU/RAM numbers and no concurrency ceiling recommendation yet.

## PERF-HTTP-SERVER-001 — P1 RISK

A slow incomplete-header client probe is included to demonstrate connection retention when server timeouts are zero-valued. It is diagnostic only; server startup is not refactored in this PR.

This risk should also be carried into Observability/Operations and final Release/Deployment closure for an explicit timeout/graceful-shutdown decision.

## MySQL pool status

Do not tune yet.

Measured latest workloads:

- 100-concurrent HTTP baseline: DB open `2`, WaitCount delta `0`, WaitDuration delta `0s`.
- 10-round 100-concurrent detail loop: DB open `2`, in-use `0` at end-of-round samples, WaitCount `0`, WaitDuration `0s`.
- Generation 100: DB open `1`, WaitCount delta `0`, WaitDuration delta `0s`.

These end-of-workload samples do not establish the true instantaneous peak connection count, so the unlimited `MaxOpenConns=0` remains a capacity risk, but there is still no saturation evidence that justifies arbitrary tuning.

## Risk register

### P0

- `PERF-RUN-IDEMPOTENCY-001` — CONFIRMED, `BLOCKED_BY_TASK9_4`. 20 concurrent creates -> 20 logical Runs instead of 1.

### P1

- `PERF-FFMPEG-001` — CONFIRMED capacity risk: no executor concurrency gate; real resource ceiling still unmeasured.
- `PERF-DB-POOL-001` — risk only, saturation not reproduced; unlimited MaxOpenConns but zero waits in measured loads.
- `PERF-HTTP-SERVER-001` — risk: no explicit production server timeouts/graceful-shutdown wiring; diagnostic slow-client probe added.

### P2

- `PERF-BATCH-DETAIL-SCALE-001` — high-concurrency response/materialization/serialization scaling trend; 100-book object is ~1.20 MiB and 100-way P95 is materially above single-query latency.
- Generation current synchronous throughput declines at larger batch sizes; correctness remains intact and no business concurrency change is made in this PR.

No new P0/P1 was created by the second workload set. RSS remains a retention signal, not a confirmed leak.

## Task 9 blocked tests

Unchanged until Task 9 Runtime enters main:

- 1/2/5/10 Worker claim distribution;
- duplicate claim / contention;
- lease renew, expiry, reclaim and fencing;
- stale complete/fail/release;
- Worker crash storm and stale-running recovery;
- Redis restart, wipe, temporary unavailability and connection-pool capacity;
- 100 simultaneous/near-simultaneous scheduled Runs and multi-Scheduler ownership;
- 100/500 queue-depth backpressure using the real Runtime;
- graceful Worker shutdown and startup recovery;
- Runtime-level 50/100 partial failure + failed-only retry;
- true 1/5/10 concurrent Run execution isolation;
- final Runtime retry/backoff behavior for Provider failures;
- 100-running-video poll storm / lease coordination.

The audit explicitly does not invent a fake Redis Runtime and call those scenarios accepted.

## CI layering

- Normal repository CI remains unchanged.
- `performance-audit.yml` runs isolated MySQL + controlled provider/storage/merge workloads on the audit branch and via manual dispatch.
- `PERF-RUN-IDEMPOTENCY-001` remains a diagnostic known-failure with explicit P0/BLOCKED status so other metrics can run. After Task 9 merges, the same assertion becomes a hard GREEN gate.
- `.github/workflows/performance-ffmpeg-manual.yml` is manual-only for real ffmpeg resource measurement.
- 100/500 real Runtime queue tests, long-duration profiling, real TOS, disk-full and real-provider smoke remain manual/blocked as appropriate.

## ECS recommendation

No ECS machine size is recommended yet. Minimum/recommended/headroom tiers require actual Task 9 Worker concurrency plus real ffmpeg CPU/RAM/temp-disk and real storage throughput measurements. Do not infer 2C4G/4C8G/8C16G from the current GitHub-runner control-plane tests.
