# Task 4 report — scheduling and batch accounting

Date: 2026-11-17 Asia/Shanghai. Status: implemented locally, ready for independent review; this is not a whole-project parity or runtime acceptance result.

- Base: `5ea766dc32ddad7db3fdcd1a8423cd91c5f87107`, existing `v88` branch.
- Implementation commit: `8c8165ec8ec15e332ef84c2e03da2e055b68a588`.
- No branch/worktree created. No provider/paid calls, live MySQL migration, 4001 restart, embed rebuild, deployment or push. Unrelated dirty files were not staged.

## Delivered behavior

Schedules default to 30 minutes / 10 daily saved chapters / 3 consecutive failures. Valid ranges are 1–1440 / 1–100 / 1–10. Batches default to 5 chapters / 3 failures, with ranges 1–50 / 1–10. Legacy invalid values remain unchanged and enter a configuration pause; no silent clamping.

All automated creation and round accounting use the existing project transaction. A normal round key derives from the persisted planned due timestamp, rather than scan time. Migration `8800313` adds `novel_automation_rounds` with owner/kind/plan/round primary key and owner/run uniqueness. The run, message, durable dispatch record, unique round and plan accounting commit together. Persistence failures roll back the outer transaction. Busy work is detected before appending a message or run. Neighbor scan errors remain aggregated, including configuration-stop persistence errors covered by Task2.

Only an actual saved `novel_chapter_versions` row for the run can increment success counts. Batch failures do not advance completed chapters. Accounting clears the active round transactionally, so another process cannot count it again. Explicit temporary errors retry within the configured failure limit; successful saves reset consecutive failures. Authentication/configuration and deterministic failures stop immediately. Unknown results retain their active round and reject ordinary resume. `ready` / `save_pending` local recovery stays on the same run.

Blueprint checkpoints do not count as saved chapters. Inspection follows `confirm:<sourceID>` even after confirmation changes the source status to succeeded, and associates the saved successor with the original round. Stop also finds and cancels that successor. Pause preserves the current run and blocks further dispatch. Cancel notifications/context signals occur only after the encompassing automation transaction commits.

Successful/explicitly failed normal rounds calculate the next due time from completion handling plus interval. Immediate rounds remain separate and never overwrite the normal due time, including on completion. Busy scans keep the due timestamp and do not create messages/runs or add failures. Only one overdue round is dispatched; there is no catch-up loop. Shanghai calendar days are calculated from UTC timestamps, including late accounting of a chapter saved before midnight. Daily eligibility is next Shanghai 00:00; it does not force an otherwise future due time. User pause does not automatically resume at midnight.

## HTTP and Task7 bindings

Existing request fields and routes remain compatible. The following fields are added to POST/PUT schedule requests:

- `dailyChapterCap` (default 10), `maxConsecutiveFailures` (default 3).
- Existing `intervalMinutes` now defaults to 30 and accepts 1–1440.

POST batch requests add `maxConsecutiveFailures` (default 3); omitted `totalChapters` now defaults to 5.

New endpoints, all owner/project scoped:

- `POST /api/novel-creation/projects/{projectID}/schedules/{scheduleID}/pause`
- `POST /api/novel-creation/projects/{projectID}/schedules/{scheduleID}/resume`
- `POST /api/novel-creation/projects/{projectID}/schedules/{scheduleID}/stop`
- `GET /api/novel-creation/projects/{projectID}/schedules/{scheduleID}/logs`
- `GET /api/novel-creation/projects/{projectID}/batches/{batchID}/logs`

Control requests need no fields and return `{schedule: ...}`. Log endpoints return `{rounds: [...]}`. Existing batch pause/resume/stop endpoints now preserve pending saved results and confirmed successor linkage.

The existing GET schedules/batches lists expose:

- Schedule `status`: `not_started`, `running`, `waiting`, `paused`, `stopped`; batch retains `running`, `paused`, `stopped`, `completed`, `failed`.
- `dailyChapterCap`, `maxConsecutiveFailures`, `consecutiveFailures`, `totalSavedChapters`, `dailySavedChapters`, `successDay` (Shanghai date).
- `stoppedReason`, `lastFailure`, `waitingReason`, optional `eligibleAt`, `activeRound`, `rounds`.
- Existing schedule `configurationError`, `enabled`, `nextRunAt`, `lastRunAt`, `lastRunId` remain available.
- Existing batch `completedChapters`, `totalChapters`, `currentRunId`, `lastError` remain authoritative. Current chapter is completed count + 1 until the target is met.

`activeRound` references a round's `key`; round records expose `plannedAt`, `immediate`, `runId`, `sourceRunId` (blueprint checkpoint), `previousRunId` (prior batch attempt/chapter), `state`, error code/message and optional completion handling timestamp. Checkpoint/ready/save_pending waiting reasons are readable Chinese messages. UI still needs to map status values to Chinese labels in Task7.

Input snapshots also persist `source_type`, `source_plan_id`, `source_round`, and `source_run_id` for schedule / schedule_immediate / batch. Existing mode/reference/note/style snapshots and continuation/version fencing remain in the shared pipeline.

## Red → green evidence

Initial real behavior test command:

`go test ./internal/novelcreation -run 'TestAutomation(ScheduleBusy|ScheduleImmediate|BatchOnly)' -count=1`

Observed failures before implementation:

```text
--- FAIL: TestAutomationScheduleBusyRetainsPlannedDue
    busy advanced due: 2026-11-16 19:17:28... -> 2026-11-16 20:47:28...
--- FAIL: TestAutomationBatchOnlySavedSuccessCounts
    status-only success counted without saved chapter
FAIL qiantie/backend/internal/novelcreation
```

Expanded busy assertion later exposed an additional real failure:

```text
--- FAIL: TestAutomationScheduleBusyRetainsPlannedDue
    automation_state_test.go:28: busy created extra runs/messages: 2/2
FAIL qiantie/backend/internal/novelcreation
```

Fixed by transaction-scoped busy preflight before message/run creation. This test now verifies no extra run or message, as well as unchanged planned due.

The checkpoint/saving test also initially failed with `wrong successor` and a source checkpoint incorrectly treated as a succeeded-but-unsaved chapter. Following the persisted confirmation successor fixes that case; the test now reaches the next real chapter, with source linkage intact.

Final focused command:

`go test ./internal/novelcreation ./internal/httpapi -run 'Schedule|Batch|Automation' -count=1`

```text
ok qiantie/backend/internal/novelcreation 0.050s
ok qiantie/backend/internal/httpapi 27.491s
```

Final full domain regression after the final read-time day-display adjustment:

`go test ./internal/novelcreation -count=1`

```text
ok qiantie/backend/internal/novelcreation 0.132s
```

Focused storage migration checks:

`go test ./internal/storage -run 'Novel|AutomationInputMigration' -count=1`

```text
ok qiantie/backend/internal/storage 0.018s
```

The full storage package was also checked. Its remaining failure is inherited and outside Task4:

```text
--- FAIL: TestAppMigrationsIncludesV11AndV12Migrations
    migrations_test.go:81: AppMigrations length=116, want 115
FAIL qiantie/backend/internal/storage
```

Verified against base/HEAD sources: `AppMigrations` already includes `ProjectRetentionMigrations()`, while the pre-existing expected list in `migrations_test.go` omits it. Task4's own migration-count expectation was updated from 6 to 7 and its focused checks pass. The unrelated expected-list mismatch was left for Task8.

`git diff --check` passed for the scoped files. Additional current-HEAD signed HTTP control/log ownership verification: `go test ./internal/httpapi -run '^TestNovelCreationAutomationControlsCapsAndLogsHTTP$' -count=1` → `ok qiantie/backend/internal/httpapi 0.034s`.

## Test boundaries

Domain behavior uses MemoryStore and injected automation time / explicit `now`; no minute/day sleeps. The existing DB lease clock remains unchanged. Two distinct Service instances concurrently claim and account one shared persisted plan in the test store. Tests cover caps, Shanghai midnight, future due eligibility, user pause across midnight, immediate completion, bounded failures, auth/unknown stops, source confirmation, save_pending, stop cancellation, actual-success preservation, and duplicate accounting.

SQLMock verifies project `FOR UPDATE`, successful round insert/commit, duplicate-key failure/rollback, owner-scoped saved-version reads, and automation-state round trip. The migration checks assert the unique primary/run keys and run foreign key. This is SQLMock plus schema evidence, **not live MySQL multi-process acceptance**. Task8 must apply/verify migration 8800313 on the authorized acceptance database and exercise competing processes. No current 4001 database was changed.

No UI, Task5 Agent settings, Task6 materials expansion, real provider behavior, production persistence, or full parity is claimed by this report.

## Changed files

- `backend/internal/novelcreation/automation_state.go`, `automation_state_test.go`, `automation_state_mysql_test.go`: common state/outcome logic, source/confirmation/cancel handling, round stores and behavioral tests.
- `backend/internal/novelcreation/scheduler.go`, `batch.go`: transactional orchestration, caps, controls and round accounting.
- `backend/internal/novelcreation/types.go`, `validation.go`, `store.go`, `service.go`: inputs/defaults/ranges, store interfaces, automation clock and source snapshots.
- `backend/internal/novelcreation/automation_input.go`, `memory_store.go`, `mysql_store.go`, `worker_store.go`: plan JSON persistence, rollback copies, active scans and saved chapter truth. No lease/fencing implementation changes.
- `backend/internal/novelcreation/service_test.go`: former status-only automation fixtures now save real chapter/version/memory bundles and finish the durable journal.
- `backend/internal/novelcreation/automation_input_mysql_test.go`, `automation_input_test.go`: preserve Task2 material contract tests and preserve fault injection through project transactions.
- `backend/internal/httpapi/novel_creation_routes.go`, `novel_creation_automation_state_test.go`: control/log routes, added request fields, signed request/ownership verification.
- `backend/internal/storage/novel_creation_schema.go`, `novel_creation_schema_test.go`: migration 8800313 and focused schema checks.
