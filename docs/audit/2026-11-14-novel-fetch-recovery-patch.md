# Novel Fetch recovery patch — 2026-11-14

Branch: `fix/novel-fetch-recovery-20261114`.
Source: local Git `v88` at `75c45576` after rebasing onto the latest local V88, without unrelated working-tree edits.

## Changes

- Migration 1100002 verifies existing foreign-key definitions before resuming a missing ledger checkpoint. It only adds missing keys. A conflicting key on another table fails without changing that table.
- Migration 1200105 inventories all source/target table pairs and atomically renames only the remaining source tables. Existing targets are preserved; ambiguous or missing pairs fail explicitly. Historical migration checksums remain unchanged.
- Startup permits five minutes for durable MySQL DDL instead of interrupting it after thirty seconds.
- Deferred Novel Fetch schedules follow the state of their own books after automation retries them under a different process job. Successful processing remains pending until required upload confirmation is present.
- Reconciliation never enqueues the schedule again. It preserves cancellation and concurrent completion and skips unchanged status writes.
- Schedule cards display the recorded error and next retry time.

## Verification

- Storage package tests passed.
- Schedule retry, schedule controls, automation and full-chain/readback recovery tests passed.
- A disposable MySQL 8.4.11 database reproduced an interrupted foreign-key migration and a partial V11-to-V12 table rename. Recovery and repeated startup passed; the original book row was preserved. Runtime was 81.828 seconds. The generated test databases were removed and their absence verified.
- All 119 existing Novel Fetch frontend tests passed with the installed project dependencies. JSX parsing completed without warnings.
- `git diff --check` passed.

## Not yet verified or applied

The user's existing 11:27 schedule and its first upstream-fetch error were not readable in the live business database. No live book was generated or uploaded during this verification.

This patch is prepared in the writable G: workspace. The actual project on F: is outside this session's writable roots, and the Docker API denied access. The patch has not been applied to that project or deployed to port 4001. A later independent restart made `/health` return 200; the runtime reported `git_sha=bc16b9a5` and `runtime_manifest_sha=75c45576`. These reports do not prove this patch is running or that the original task has recovered.

Before deployment, inspect the live schema and migration ledger. If both V11 and V12 tables exist, do not drop tables or mark migrations complete without checking their contents. The patch deliberately refuses ambiguous table/key ownership instead of silently discarding data.

After applying to the maintained V88 branch, rebuild the React source and embedded Go binary, start local 4001 with its existing business database/configuration, verify `/health` and `/api/runtime-build-info`, then inspect and resume the existing schedule. Local live validation remains required before any production deployment.
