# Task 9 Runtime Infrastructure implementation plan

Spec authority: user-approved Task 9.4 + 9.5 constraints on 2026-10-05.

## Global constraints

- MySQL is the durable source of truth. Redis is coordination only.
- Reuse `runs`, `book_runs`, `stage_runs`; do not add `runtime_jobs` or `runtime_book_runs`.
- `00007` is reserved for Task 14. Task 9 uses `00008_task9_runtime.sql`.
- Queue claim is never execution authority; MySQL CAS is required before execution.
- Every execution uses a durable MySQL execution epoch/fencing token. Complete/fail/renew reject stale workers.
- Redis lease renew/release are atomic owner+token checked operations.
- Redis wipe recovery is driven by MySQL stale deadlines and bounded attempts, never by missing Redis keys alone.
- Scheduler claim and API idempotency are guarded by MySQL atomic transitions/UNIQUE constraints.
- Runtime package stays domain-agnostic; Task 9 owns the adapter.
- No Task 14 video-provider, ffmpeg, MergeJob, or doubao business changes.

## Tasks

1. RED: runtime queue/lease contracts, fencing, stale-owner protection, Task 9 lifecycle/aggregation/retry/idempotency/scheduler/recovery/redaction/migration contracts.
2. GREEN: `00008_task9_runtime.sql`, generic `internal/taskruntime`, Task 9 MySQL adapter and scheduler/worker.
3. Integrate existing generation pipeline through the Task 9 adapter without duplicating BookRun or stage-attempt models.
4. Wire browser APIs through existing Task 15 auth/same-origin/capability/ownership boundaries; internal worker/scheduler use service boundaries.
5. Add real Redis + MySQL integration coverage to CI; verify Goose up/status/down and existing-main -> 00008 upgrade path.
6. Re-sync latest main before PR, rerun Go/frontend/admin/migration/integration verification, then update TASKS.md only with verified facts.
